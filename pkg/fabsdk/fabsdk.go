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
	"github.com/hyperledger/fabric-sdk-go/pkg/core/cryptosuite"
	sdkApi "github.com/hyperledger/fabric-sdk-go/pkg/fabsdk/api"
	"github.com/hyperledger/fabric-sdk-go/pkg/fabsdk/provider/chpvdr"
	"github.com/pkg/errors"
)

// FabricSDK provides access (and context) to clients being managed by the SDK.
type FabricSDK struct {
	opts     options
	provider *context.Provider
}

type options struct {
	Core    sdkApi.CoreProviderFactory
	MSP     sdkApi.MSPProviderFactory
	Service sdkApi.ServiceProviderFactory
	Logger  api.LoggerProvider
}

// Option configures the SDK.
type Option func(opts *options) error

type closeable interface {
	Close()
}

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
func fromPkgSuite(config core.Config, pkgSuite pkgSuite, opts ...Option) (*FabricSDK, error) {
	core, err := pkgSuite.Core()
	if err != nil {
		return nil, errors.WithMessage(err, "Unable to initialize core pkg")
	}

	msp, err := pkgSuite.MSP()
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
			Core:    core,
			MSP:     msp,
			Service: svc,
			Logger:  lg,
		},
	}

	err = initSDK(&sdk, config, opts)
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

func initSDK(sdk *FabricSDK, config core.Config, opts []Option) error {
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

	// Initialize crypto provider
	cryptoSuite, err := sdk.opts.Core.CreateCryptoSuiteProvider(config)
	if err != nil {
		return errors.WithMessage(err, "failed to initialize crypto suite")
	}

	// Initialize rand (TODO: should probably be optional)
	rand.Seed(time.Now().UnixNano())

	// Setting this cryptosuite as the factory default
	cryptosuite.SetDefault(cryptoSuite)

	// Initialize state store
	userStore, err := sdk.opts.MSP.CreateUserStore(config)
	if err != nil {
		return errors.WithMessage(err, "failed to initialize state store")
	}

	// Initialize Signing Manager
	signingManager, err := sdk.opts.Core.CreateSigningManager(cryptoSuite, config)
	if err != nil {
		return errors.WithMessage(err, "failed to initialize signing manager")
	}

	// Initialize IdentityManagerProvider
	identityManagerProvider, err := sdk.opts.MSP.CreateIdentityManagerProvider(config, cryptoSuite, userStore)
	if err != nil {
		return errors.WithMessage(err, "failed to initialize identity manager provider")
	}

	// Initialize Fabric provider
	infraProvider, err := sdk.opts.Core.CreateInfraProvider(config)
	if err != nil {
		return errors.WithMessage(err, "failed to initialize infra provider")
	}

	// Initialize discovery provider
	discoveryProvider, err := sdk.opts.Service.CreateDiscoveryProvider(config, infraProvider)
	if err != nil {
		return errors.WithMessage(err, "failed to initialize discovery provider")
	}

	// Initialize selection provider (for selecting endorsing peers)
	selectionProvider, err := sdk.opts.Service.CreateSelectionProvider(config)
	if err != nil {
		return errors.WithMessage(err, "failed to initialize selection provider")
	}

	channelProvider, err := chpvdr.New(infraProvider)
	if err != nil {
		return errors.WithMessage(err, "failed to initialize channel provider")
	}

	//update sdk providers list since all required providers are initialized
	sdk.provider = context.NewProvider(context.WithConfig(config),
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
		pi.Initialize(sdk.provider)
	}

	if pi, ok := discoveryProvider.(providerInit); ok {
		pi.Initialize(sdk.provider)
	}

	if pi, ok := selectionProvider.(providerInit); ok {
		pi.Initialize(sdk.provider)
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

// Config returns the SDK's configuration.
func (sdk *FabricSDK) Config() core.Config {
	return sdk.provider.Config()
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
