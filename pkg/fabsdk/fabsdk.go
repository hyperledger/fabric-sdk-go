/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

// Package fabsdk enables client usage of a Hyperledger Fabric network.
package fabsdk

import (
	"math/rand"
	"time"

	"github.com/hyperledger/fabric-sdk-go/pkg/context"
	"github.com/hyperledger/fabric-sdk-go/pkg/context/api/fab"
	"github.com/hyperledger/fabric-sdk-go/pkg/logging/api"

	"github.com/hyperledger/fabric-sdk-go/pkg/context/api/core"
	"github.com/hyperledger/fabric-sdk-go/pkg/core/cryptosuite"
	sdkApi "github.com/hyperledger/fabric-sdk-go/pkg/fabsdk/api"
	"github.com/hyperledger/fabric-sdk-go/pkg/fabsdk/provider/chpvdr"
	"github.com/hyperledger/fabric-sdk-go/pkg/logging"
	"github.com/pkg/errors"
)

// FabricSDK provides access (and context) to clients being managed by the SDK.
type FabricSDK struct {
	opts options

	config            core.Config
	stateStore        core.KVStore
	cryptoSuite       core.CryptoSuite
	discoveryProvider fab.DiscoveryProvider
	selectionProvider fab.SelectionProvider
	signingManager    core.SigningManager
	identityManager   map[string]core.IdentityManager
	fabricProvider    fab.InfraProvider
	channelProvider   fab.ChannelProvider
}

type options struct {
	Core    sdkApi.CoreProviderFactory
	Service sdkApi.ServiceProviderFactory
	Session sdkApi.SessionClientFactory
	Logger  api.LoggerProvider
}

// Option configures the SDK.
type Option func(opts *options) error

// New initializes the SDK based on the set of options provided.
// configProvider provides the application configuration.
func New(cp core.ConfigProvider, opts ...Option) (*FabricSDK, error) {
	pkgSuite := defPkgSuite{}
	config, err := cp()
	if err != nil {
		return nil, errors.WithMessage(err, "unable to load configuration")
	}
	return fromPkgSuite(config, &pkgSuite, opts...)
}

// WithConfig converts a Config interface to a ConfigProvider.
// This is a helper function for those who already loaded the config
// prior to instantiating the SDK.
func WithConfig(config core.Config) core.ConfigProvider {
	return func() (core.Config, error) {
		return config, nil
	}
}

// fromPkgSuite creates an SDK based on the implementations in the provided pkg suite.
// TODO: For now leaving this method as private until we have more usage.
func fromPkgSuite(config core.Config, pkgSuite PkgSuite, opts ...Option) (*FabricSDK, error) {
	core, err := pkgSuite.Core()
	if err != nil {
		return nil, errors.WithMessage(err, "Unable to initialize core pkg")
	}

	svc, err := pkgSuite.Service()
	if err != nil {
		return nil, errors.WithMessage(err, "Unable to initialize service pkg")
	}

	sess, err := pkgSuite.Session()
	if err != nil {
		return nil, errors.WithMessage(err, "Unable to initialize session pkg")
	}

	lg, err := pkgSuite.Logger()
	if err != nil {
		return nil, errors.WithMessage(err, "Unable to initialize logger pkg")
	}

	sdk := FabricSDK{
		opts: options{
			Core:    core,
			Service: svc,
			Session: sess,
			Logger:  lg,
		},
		config: config,
	}

	err = initSDK(&sdk, opts)
	if err != nil {
		return nil, err
	}

	return &sdk, err
}

// WithCorePkg injects the core implementation into the SDK.
func WithCorePkg(core sdkApi.CoreProviderFactory) Option {
	return func(opts *options) error {
		opts.Core = core
		return nil
	}
}

// WithServicePkg injects the service implementation into the SDK.
func WithServicePkg(service sdkApi.ServiceProviderFactory) Option {
	return func(opts *options) error {
		opts.Service = service
		return nil
	}
}

// WithSessionPkg injects the session implementation into the SDK.
func WithSessionPkg(session sdkApi.SessionClientFactory) Option {
	return func(opts *options) error {
		opts.Session = session
		return nil
	}
}

// WithLoggerPkg injects the logger implementation into the SDK.
func WithLoggerPkg(logger api.LoggerProvider) Option {
	return func(opts *options) error {
		opts.Logger = logger
		return nil
	}
}

// providerInit interface allows for initializing providers
// TODO: minimize interface
type providerInit interface {
	Initialize(sdk *FabricSDK) error
}

func initSDK(sdk *FabricSDK, opts []Option) error {
	for _, option := range opts {
		err := option(&sdk.opts)
		if err != nil {
			return errors.WithMessage(err, "Error in option passed to New")
		}
	}

	// Initialize logging provider with default logging provider (if needed)
	if sdk.opts.Logger == nil {
		return errors.New("Missing logger from pkg suite")
	}
	logging.InitLogger(sdk.opts.Logger)

	// Initialize crypto provider
	cs, err := sdk.opts.Core.CreateCryptoSuiteProvider(sdk.config)
	if err != nil {
		return errors.WithMessage(err, "failed to initialize crypto suite")
	}

	sdk.cryptoSuite = cs

	// Initialize rand (TODO: should probably be optional)
	rand.Seed(time.Now().UnixNano())

	// Setting this cryptosuite as the factory default
	cryptosuite.SetDefault(cs)

	// Initialize state store
	store, err := sdk.opts.Core.CreateStateStoreProvider(sdk.config)
	if err != nil {
		return errors.WithMessage(err, "failed to initialize state store")
	}
	sdk.stateStore = store

	// Initialize Signing Manager
	signingMgr, err := sdk.opts.Core.CreateSigningManager(sdk.cryptoSuite, sdk.config)
	if err != nil {
		return errors.WithMessage(err, "failed to initialize signing manager")
	}
	sdk.signingManager = signingMgr

	// Initialize Identity Managers
	sdk.identityManager = make(map[string]core.IdentityManager)
	netConfig, err := sdk.config.NetworkConfig()
	if err != nil {
		return errors.Wrapf(err, "failed to retrieve network config")
	}
	for orgName := range netConfig.Organizations {
		mgr, err := sdk.opts.Core.CreateIdentityManager(orgName, sdk.stateStore, sdk.cryptoSuite, sdk.config)
		if err != nil {
			return errors.Wrapf(err, "failed to initialize identity manager for organization: %s", orgName)
		}
		sdk.identityManager[orgName] = mgr
	}

	// Initialize Fabric Provider
	fabricProvider, err := sdk.opts.Core.CreateFabricProvider(sdk.fabContext())
	if err != nil {
		return errors.WithMessage(err, "failed to initialize core fabric provider")
	}
	sdk.fabricProvider = fabricProvider

	// Initialize discovery provider
	discoveryProvider, err := sdk.opts.Service.CreateDiscoveryProvider(sdk.config, fabricProvider)
	if err != nil {
		return errors.WithMessage(err, "failed to initialize discovery provider")
	}
	if pi, ok := discoveryProvider.(providerInit); ok {
		pi.Initialize(sdk)
	}
	sdk.discoveryProvider = discoveryProvider

	// Initialize selection provider (for selecting endorsing peers)
	selectionProvider, err := sdk.opts.Service.CreateSelectionProvider(sdk.config)
	if err != nil {
		return errors.WithMessage(err, "failed to initialize selection provider")
	}
	if pi, ok := selectionProvider.(providerInit); ok {
		pi.Initialize(sdk)
	}
	sdk.selectionProvider = selectionProvider

	channelProvider, err := chpvdr.New(fabricProvider)
	if err != nil {
		return errors.WithMessage(err, "failed to initialize channel provider")
	}
	sdk.channelProvider = channelProvider

	return nil
}

// Close frees up caches and connections being maintained by the SDK
func (sdk *FabricSDK) Close() {
	// TODO: upcoming changes will have Close funcs being called from here.
}

// Config returns the SDK's configuration.
func (sdk *FabricSDK) Config() core.Config {
	return sdk.config
}

func (sdk *FabricSDK) fabContext() core.Providers {
	return context.CreateFabContext(context.WithConfig(sdk.config),
		context.WithCryptoSuite(sdk.cryptoSuite),
		context.WithSigningManager(sdk.signingManager),
		context.WithStateStore(sdk.stateStore),
		context.WithDiscoveryProvider(sdk.discoveryProvider),
		context.WithSelectionProvider(sdk.selectionProvider),
		context.WithIdentityManager(sdk.identityManager),
		context.WithFabricProvider(sdk.fabricProvider),
		context.WithChannelProvider(sdk.channelProvider))
}

func (sdk *FabricSDK) context() context.Providers {
	fabContext := context.CreateFabContext(context.WithConfig(sdk.config),
		context.WithCryptoSuite(sdk.cryptoSuite),
		context.WithSigningManager(sdk.signingManager),
		context.WithStateStore(sdk.stateStore),
		context.WithDiscoveryProvider(sdk.discoveryProvider),
		context.WithSelectionProvider(sdk.selectionProvider),
		context.WithIdentityManager(sdk.identityManager),
		context.WithFabricProvider(sdk.fabricProvider),
		context.WithChannelProvider(sdk.channelProvider))
	c := context.SDKContext{
		*fabContext,
	}
	return &c
}
