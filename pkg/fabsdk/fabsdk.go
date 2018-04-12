/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

// Package fabsdk enables client usage of a Hyperledger Fabric network.
package fabsdk

import (
	"math/rand"
	"time"

	contextApi "github.com/hyperledger/fabric-sdk-go/pkg/common/providers/context"
	"github.com/hyperledger/fabric-sdk-go/pkg/context"
	"github.com/hyperledger/fabric-sdk-go/pkg/core/logging/api"

	"github.com/hyperledger/fabric-sdk-go/pkg/common/logging"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/core"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/fab"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/msp"
	"github.com/hyperledger/fabric-sdk-go/pkg/core/cryptosuite"
	fabImpl "github.com/hyperledger/fabric-sdk-go/pkg/fab"
	sdkApi "github.com/hyperledger/fabric-sdk-go/pkg/fabsdk/api"
	"github.com/hyperledger/fabric-sdk-go/pkg/fabsdk/provider/chpvdr"
	mspImpl "github.com/hyperledger/fabric-sdk-go/pkg/msp"
	"github.com/pkg/errors"
)

var logger = logging.NewLogger("fabsdk")

// FabricSDK provides access (and context) to clients being managed by the SDK.
type FabricSDK struct {
	opts     options
	provider *context.Provider
}

type options struct {
	Core              sdkApi.CoreProviderFactory
	MSP               sdkApi.MSPProviderFactory
	Service           sdkApi.ServiceProviderFactory
	Logger            api.LoggerProvider
	CryptoSuiteConfig core.CryptoSuiteConfig
	endpointConfig    fab.EndpointConfig
	IdentityConfig    msp.IdentityConfig
	ConfigBackend     core.ConfigBackend
}

// Option configures the SDK.
type Option func(opts *options) error

type closeable interface {
	Close()
}

// New initializes the SDK based on the set of options provided.
// ConfigOptions provides the application configuration.
func New(configProvider core.ConfigProvider, opts ...Option) (*FabricSDK, error) {
	pkgSuite := defPkgSuite{}
	return fromPkgSuite(configProvider, &pkgSuite, opts...)
}

// fromPkgSuite creates an SDK based on the implementations in the provided pkg suite.
// TODO: For now leaving this method as private until we have more usage.
func fromPkgSuite(configProvider core.ConfigProvider, pkgSuite pkgSuite, opts ...Option) (*FabricSDK, error) {
	coreProv, err := pkgSuite.Core()
	if err != nil {
		return nil, errors.WithMessage(err, "Unable to initialize core pkg")
	}

	mspProv, err := pkgSuite.MSP()
	if err != nil {
		return nil, errors.WithMessage(err, "Unable to initialize core pkg")
	}

	svc, err := pkgSuite.Service()
	if err != nil {
		return nil, errors.WithMessage(err, "Unable to initialize service pkg")
	}

	lg, err := pkgSuite.Logger()
	if err != nil {
		return nil, errors.WithMessage(err, "Unable to initialize logger pkg")
	}

	sdk := FabricSDK{
		opts: options{
			Core:    coreProv,
			MSP:     mspProv,
			Service: svc,
			Logger:  lg,
		},
	}

	err = initSDK(&sdk, configProvider, opts)
	if err != nil {
		return nil, err
	}

	return &sdk, err
}

// WithConfigCryptoSuite injects a CryptoSuiteConfig interface to the SDK
func WithConfigCryptoSuite(cryptoConfig core.CryptoSuiteConfig) Option {
	return func(opts *options) error {
		opts.CryptoSuiteConfig = cryptoConfig
		return nil
	}
}

// WithConfigEndpoint injects a EndpointConfig interface to the SDK
func WithConfigEndpoint(endpointConfig fab.EndpointConfig) Option {
	return func(opts *options) error {
		opts.endpointConfig = endpointConfig
		return nil
	}
}

// WithConfigIdentity injects a IdentityConfig interface to the SDK
func WithConfigIdentity(identityConfig msp.IdentityConfig) Option {
	return func(opts *options) error {
		opts.IdentityConfig = identityConfig
		return nil
	}
}

// WithCorePkg injects the core implementation into the SDK.
func WithCorePkg(core sdkApi.CoreProviderFactory) Option {
	return func(opts *options) error {
		opts.Core = core
		return nil
	}
}

// WithMSPPkg injects the MSP implementation into the SDK.
func WithMSPPkg(msp sdkApi.MSPProviderFactory) Option {
	return func(opts *options) error {
		opts.MSP = msp
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
	Initialize(providers contextApi.Providers) error
}

func initSDK(sdk *FabricSDK, configProvider core.ConfigProvider, opts []Option) error { //nolint
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
	logging.Initialize(sdk.opts.Logger)

	//Initialize config if not passed through options
	err := sdk.loadConfig(configProvider)
	if err != nil {
		return errors.WithMessage(err, "failed to initialize configuration")
	}

	// Initialize crypto provider
	cryptoSuite, err := sdk.opts.Core.CreateCryptoSuiteProvider(sdk.opts.CryptoSuiteConfig)
	if err != nil {
		return errors.WithMessage(err, "failed to initialize crypto suite")
	}

	// Initialize rand (TODO: should probably be optional)
	rand.Seed(time.Now().UnixNano())

	// Setting this cryptosuite as the factory default
	if !cryptosuite.DefaultInitialized() {
		err = cryptosuite.SetDefault(cryptoSuite)
		if err != nil {
			return errors.WithMessage(err, "failed to set default crypto suite")
		}
	} else {
		logger.Debug("default cryptosuite already initialized")
	}

	// Initialize state store
	userStore, err := sdk.opts.MSP.CreateUserStore(sdk.opts.IdentityConfig)
	if err != nil {
		return errors.WithMessage(err, "failed to create state store")
	}

	// Initialize Signing Manager
	signingManager, err := sdk.opts.Core.CreateSigningManager(cryptoSuite)
	if err != nil {
		return errors.WithMessage(err, "failed to create signing manager")
	}

	// Initialize IdentityManagerProvider
	identityManagerProvider, err := sdk.opts.MSP.CreateIdentityManagerProvider(sdk.opts.endpointConfig, cryptoSuite, userStore)
	if err != nil {
		return errors.WithMessage(err, "failed to create identity manager provider")
	}

	// Initialize Fabric provider
	infraProvider, err := sdk.opts.Core.CreateInfraProvider(sdk.opts.endpointConfig)
	if err != nil {
		return errors.WithMessage(err, "failed to create infra provider")
	}

	// Initialize discovery provider
	discoveryProvider, err := sdk.opts.Service.CreateDiscoveryProvider(sdk.opts.endpointConfig)
	if err != nil {
		return errors.WithMessage(err, "failed to create discovery provider")
	}

	// Initialize selection provider (for selecting endorsing peers)
	selectionProvider, err := sdk.opts.Service.CreateSelectionProvider(sdk.opts.endpointConfig)
	if err != nil {
		return errors.WithMessage(err, "failed to create selection provider")
	}

	channelProvider, err := chpvdr.New(infraProvider)
	if err != nil {
		return errors.WithMessage(err, "failed to create channel provider")
	}

	//update sdk providers list since all required providers are initialized
	sdk.provider = context.NewProvider(context.WithCryptoSuiteConfig(sdk.opts.CryptoSuiteConfig),
		context.WithEndpointConfig(sdk.opts.endpointConfig),
		context.WithIdentityConfig(sdk.opts.IdentityConfig),
		context.WithCryptoSuite(cryptoSuite),
		context.WithSigningManager(signingManager),
		context.WithUserStore(userStore),
		context.WithDiscoveryProvider(discoveryProvider),
		context.WithSelectionProvider(selectionProvider),
		context.WithIdentityManagerProvider(identityManagerProvider),
		context.WithInfraProvider(infraProvider),
		context.WithChannelProvider(channelProvider))

	//initialize
	if pi, ok := infraProvider.(providerInit); ok {
		err = pi.Initialize(sdk.provider)
		if err != nil {
			return errors.WithMessage(err, "failed to initialize infra provider")
		}
	}

	if pi, ok := discoveryProvider.(providerInit); ok {
		err = pi.Initialize(sdk.provider)
		if err != nil {
			return errors.WithMessage(err, "failed to initialize discovery provider")
		}
	}

	if pi, ok := selectionProvider.(providerInit); ok {
		err = pi.Initialize(sdk.provider)
		if err != nil {
			return errors.WithMessage(err, "failed to initialize selection provider")
		}
	}

	return nil
}

// Close frees up caches and connections being maintained by the SDK
func (sdk *FabricSDK) Close() {
	if pvdr, ok := sdk.provider.DiscoveryProvider().(closeable); ok {
		pvdr.Close()
	}
	if pvdr, ok := sdk.provider.SelectionProvider().(closeable); ok {
		pvdr.Close()
	}
	sdk.provider.InfraProvider().Close()
}

//Config returns config backend used by all SDK config types
func (sdk *FabricSDK) Config() (core.ConfigBackend, error) {
	if sdk.opts.ConfigBackend == nil {
		return nil, errors.New("unable to find config backend")
	}
	return sdk.opts.ConfigBackend, nil
}

//Context creates and returns context client which has all the necessary providers
func (sdk *FabricSDK) Context(options ...ContextOption) contextApi.ClientProvider {

	clientProvider := func() (contextApi.Client, error) {
		identity, err := sdk.newIdentity(options...)
		if err == ErrAnonymousIdentity {
			identity = nil
			err = nil
		}
		return &context.Client{Providers: sdk.provider, SigningIdentity: identity}, err
	}

	return clientProvider
}

//ChannelContext creates and returns channel context
func (sdk *FabricSDK) ChannelContext(channelID string, options ...ContextOption) contextApi.ChannelProvider {

	channelProvider := func() (contextApi.Channel, error) {

		clientCtxProvider := sdk.Context(options...)
		return context.NewChannel(clientCtxProvider, channelID)

	}

	return channelProvider
}

//loadConfig load config from config backend when configs are not provided through opts
func (sdk *FabricSDK) loadConfig(configProvider core.ConfigProvider) error {
	if sdk.opts.CryptoSuiteConfig == nil || sdk.opts.endpointConfig == nil || sdk.opts.IdentityConfig == nil {
		configBackend, err := configProvider()
		if err != nil {
			return errors.WithMessage(err, "unable to load config backend")
		}

		//configs passed through opts takes priority
		if sdk.opts.CryptoSuiteConfig == nil {
			sdk.opts.CryptoSuiteConfig = cryptosuite.ConfigFromBackend(configBackend)
		}

		if sdk.opts.endpointConfig == nil {
			sdk.opts.endpointConfig, err = fabImpl.ConfigFromBackend(configBackend)
			if err != nil {
				return errors.WithMessage(err, "failed to initialize endpoint config from config backend")
			}
		}

		if sdk.opts.IdentityConfig == nil {
			sdk.opts.IdentityConfig, err = mspImpl.ConfigFromBackend(configBackend)
			if err != nil {
				return errors.WithMessage(err, "failed to initialize identity config from config backend")
			}
		}

		sdk.opts.ConfigBackend = configBackend
	}
	return nil
}
