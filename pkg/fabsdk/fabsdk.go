/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

// Package fabsdk enables client usage of a Hyperledger Fabric network.
package fabsdk

import (
	"math/rand"
	"time"

	"github.com/hyperledger/fabric-sdk-go/internal/github.com/hyperledger/fabric/core/operations"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/logging"
	coptions "github.com/hyperledger/fabric-sdk-go/pkg/common/options"
	contextApi "github.com/hyperledger/fabric-sdk-go/pkg/common/providers/context"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/core"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/fab"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/msp"
	"github.com/hyperledger/fabric-sdk-go/pkg/context"
	"github.com/hyperledger/fabric-sdk-go/pkg/core/config/lookup"
	"github.com/hyperledger/fabric-sdk-go/pkg/core/cryptosuite"
	"github.com/hyperledger/fabric-sdk-go/pkg/core/logging/api"
	fabImpl "github.com/hyperledger/fabric-sdk-go/pkg/fab"
	sdkApi "github.com/hyperledger/fabric-sdk-go/pkg/fabsdk/api"
	"github.com/hyperledger/fabric-sdk-go/pkg/fabsdk/metrics"
	metricsCfg "github.com/hyperledger/fabric-sdk-go/pkg/fabsdk/metrics/cfg"
	mspImpl "github.com/hyperledger/fabric-sdk-go/pkg/msp"
	"github.com/pkg/errors"
)

var logger = logging.NewLogger("fabsdk")

// FabricSDK provides access (and context) to clients being managed by the SDK.
type FabricSDK struct {
	opts          options
	provider      *context.Provider
	cryptoSuite   core.CryptoSuite
	system        *operations.System
	clientMetrics *metrics.ClientMetrics
}

type configs struct {
	cryptoSuiteConfig core.CryptoSuiteConfig
	endpointConfig    fab.EndpointConfig
	identityConfig    msp.IdentityConfig
	metricsConfig     metricsCfg.MetricsConfig
}

type options struct {
	Core              sdkApi.CoreProviderFactory
	MSP               sdkApi.MSPProviderFactory
	Service           sdkApi.ServiceProviderFactory
	Logger            api.LoggerProvider
	CryptoSuiteConfig core.CryptoSuiteConfig
	endpointConfig    fab.EndpointConfig
	IdentityConfig    msp.IdentityConfig
	ConfigBackend     []core.ConfigBackend
	ProviderOpts      []coptions.Opt // Provider options are passed along to the various providers
	metricsConfig     metricsCfg.MetricsConfig
}

// Option configures the SDK.
type Option func(opts *options) error

type closeable interface {
	Close()
}

type contextCloseable interface {
	CloseContext(ctxt fab.ClientContext)
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

// WithCryptoSuiteConfig injects a CryptoSuiteConfig interface to the SDK
// it accepts either a full interface of CryptoSuiteConfig or a list
// of sub interfaces each implementing one (or more) function(s) of CryptoSuiteConfig
func WithCryptoSuiteConfig(cryptoConfigs ...interface{}) Option {
	return func(opts *options) error {
		c, err := cryptosuite.BuildCryptoSuiteConfigFromOptions(cryptoConfigs...)
		if err != nil {
			return err
		}
		opts.CryptoSuiteConfig = c
		return nil
	}
}

// WithEndpointConfig injects a EndpointConfig interface to the SDK
// it accepts either a full interface of EndpointConfig or a list
// of sub interfaces each implementing one (or more) function(s) of EndpointConfig
func WithEndpointConfig(endpointConfigs ...interface{}) Option {
	return func(opts *options) error {
		c, err := fabImpl.BuildConfigEndpointFromOptions(endpointConfigs...)
		if err != nil {
			return err
		}
		opts.endpointConfig = c
		return nil
	}
}

// WithIdentityConfig injects a IdentityConfig interface to the SDK
// it accepts either a full interface of IdentityConfig or a list
// of sub interfaces each implementing one (or more) function(s) of IdentityConfig
func WithIdentityConfig(identityConfigs ...interface{}) Option {
	return func(opts *options) error {
		c, err := mspImpl.BuildIdentityConfigFromOptions(identityConfigs...)
		if err != nil {
			return err
		}
		opts.IdentityConfig = c
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

// WithMetricsConfig injects a MetricsConfig interface to the SDK
// it accepts either a full interface of MetricsConfig or a list
// of sub interfaces each implementing one (or more) function(s) of MetricsConfig
func WithMetricsConfig(metricsConfigs ...interface{}) Option {
	return func(opts *options) error {
		c, err := metricsCfg.BuildConfigMetricsFromOptions(metricsConfigs...)
		if err != nil {
			return err
		}
		opts.metricsConfig = c
		return nil
	}
}

// WithProviderOpts adds options which are propagated to the various providers.
func WithProviderOpts(sopts ...coptions.Opt) Option {
	return func(opts *options) error {
		opts.ProviderOpts = append(opts.ProviderOpts, sopts...)
		return nil
	}
}

// WithErrorHandler sets an error handler that will be invoked when a service error is experienced.
// This allows the client to take a decision of whether to ignore the error, shut down the client context,
// or shut down the entire SDK.
func WithErrorHandler(value fab.ErrorHandler) Option {
	return WithProviderOpts(withErrorHandlerProviderOpt(value))
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

	//Initialize configs if not passed through options
	cfg, err := sdk.loadConfigs(configProvider)
	if err != nil {
		return errors.WithMessage(err, "failed to initialize configuration")
	}

	// Initialize rand (TODO: should probably be optional)
	rand.Seed(time.Now().UnixNano())

	// Initialize state store
	userStore, err := sdk.opts.MSP.CreateUserStore(cfg.identityConfig)
	if err != nil {
		return errors.WithMessage(err, "failed to create state store")
	}

	// Initialize Signing Manager
	signingManager, err := sdk.opts.Core.CreateSigningManager(sdk.cryptoSuite)
	if err != nil {
		return errors.WithMessage(err, "failed to create signing manager")
	}

	// Initialize IdentityManagerProvider
	identityManagerProvider, err := sdk.opts.MSP.CreateIdentityManagerProvider(cfg.endpointConfig, sdk.cryptoSuite, userStore)
	if err != nil {
		return errors.WithMessage(err, "failed to create identity manager provider")
	}

	// Initialize Fabric provider
	infraProvider, err := sdk.opts.Core.CreateInfraProvider(cfg.endpointConfig)
	if err != nil {
		return errors.WithMessage(err, "failed to create infra provider")
	}

	// Initialize local discovery provider
	localDiscoveryProvider, err := sdk.opts.Service.CreateLocalDiscoveryProvider(cfg.endpointConfig)
	if err != nil {
		return errors.WithMessage(err, "failed to create local discovery provider")
	}

	channelProvider, err := sdk.opts.Service.CreateChannelProvider(cfg.endpointConfig, sdk.opts.ProviderOpts...)
	if err != nil {
		return errors.WithMessage(err, "failed to create channel provider")
	}

	sdk.initMetrics(cfg)

	//update sdk providers list since all required providers are initialized
	sdk.provider = context.NewProvider(context.WithCryptoSuiteConfig(cfg.cryptoSuiteConfig),
		context.WithEndpointConfig(cfg.endpointConfig),
		context.WithIdentityConfig(cfg.identityConfig),
		context.WithCryptoSuite(sdk.cryptoSuite),
		context.WithSigningManager(signingManager),
		context.WithUserStore(userStore),
		context.WithLocalDiscoveryProvider(localDiscoveryProvider),
		context.WithIdentityManagerProvider(identityManagerProvider),
		context.WithInfraProvider(infraProvider),
		context.WithChannelProvider(channelProvider),
		context.WithClientMetrics(sdk.clientMetrics),
	)

	//initialize
	if pi, ok := infraProvider.(providerInit); ok {
		err = pi.Initialize(sdk.provider)
		if err != nil {
			return errors.WithMessage(err, "failed to initialize infra provider")
		}
	}

	if pi, ok := localDiscoveryProvider.(providerInit); ok {
		err = pi.Initialize(sdk.provider)
		if err != nil {
			return errors.WithMessage(err, "failed to initialize local discovery provider")
		}
	}

	if pi, ok := channelProvider.(providerInit); ok {
		err = pi.Initialize(sdk.provider)
		if err != nil {
			return errors.WithMessage(err, "failed to initialize channel provider")
		}
	}

	logger.Debug("SDK initialized successfully")
	return nil
}

// Close frees up caches and connections being maintained by the SDK
func (sdk *FabricSDK) Close() {
	logger.Debug("SDK closing")
	if pvdr, ok := sdk.provider.LocalDiscoveryProvider().(closeable); ok {
		pvdr.Close()
	}
	if pvdr, ok := sdk.provider.ChannelProvider().(closeable); ok {
		pvdr.Close()
	}
	sdk.provider.InfraProvider().Close()
}

// CloseContext frees up caches being maintained by the SDK for the given context
func (sdk *FabricSDK) CloseContext(ctxt fab.ClientContext) {
	logger.Debug("Closing clients for the given context...")
	if pvdr, ok := sdk.provider.LocalDiscoveryProvider().(contextCloseable); ok {
		logger.Debugf("Closing local discovery provider...")
		pvdr.CloseContext(ctxt)
	}
	if pvdr, ok := sdk.provider.ChannelProvider().(contextCloseable); ok {
		logger.Debugf("Closing context in channel client...")
		pvdr.CloseContext(ctxt)
	}
}

//Config returns config backend used by all SDK config types
func (sdk *FabricSDK) Config() (core.ConfigBackend, error) {
	if sdk.opts.ConfigBackend == nil {
		return nil, errors.New("unable to find config backend")
	}
	return lookup.New(sdk.opts.ConfigBackend...), nil
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

// initializeCryptoSuite Initializes crypto provider
func (sdk *FabricSDK) initializeCryptoSuite(cryptoSuiteConfig core.CryptoSuiteConfig) error {
	var err error
	sdk.cryptoSuite, err = sdk.opts.Core.CreateCryptoSuiteProvider(cryptoSuiteConfig)
	if err != nil {
		return errors.WithMessage(err, "failed to initialize crypto suite")
	}

	// Setting this cryptosuite as the factory default
	if !cryptosuite.DefaultInitialized() {
		err = cryptosuite.SetDefault(sdk.cryptoSuite)
		if err != nil {
			return errors.WithMessage(err, "failed to set default crypto suite")
		}
	} else {
		logger.Debug("default cryptosuite already initialized")
	}
	return nil
}

//loadConfigs load config from config backend when configs are not provided through opts
func (sdk *FabricSDK) loadConfigs(configProvider core.ConfigProvider) (*configs, error) {
	c := &configs{
		identityConfig:    sdk.opts.IdentityConfig,
		endpointConfig:    sdk.opts.endpointConfig,
		cryptoSuiteConfig: sdk.opts.CryptoSuiteConfig,
		metricsConfig:     sdk.opts.metricsConfig,
	}

	var configBackend []core.ConfigBackend
	var err error

	if configProvider != nil {
		configBackend, err = configProvider()
		if err != nil {
			return nil, errors.WithMessage(err, "unable to load config backend")
		}
	}

	//configs passed through opts take priority over default ones
	// load crypto suite config
	c.cryptoSuiteConfig, err = sdk.loadCryptoConfig(configBackend...)
	if err != nil {
		return nil, errors.WithMessage(err, "unable to load crypto suite config")
	}

	//Initialize cryptosuite once crypto Suite config is available
	err = sdk.initializeCryptoSuite(c.cryptoSuiteConfig)
	if err != nil {
		return nil, errors.WithMessage(err, "unable to initialize cryptosuite using crypto suite config")
	}

	// load endpoint config
	c.endpointConfig, err = sdk.loadEndpointConfig(configBackend...)
	if err != nil {
		return nil, errors.WithMessage(err, "unable to load endpoint config")
	}

	// load identity config
	c.identityConfig, err = sdk.loadIdentityConfig(configBackend...)
	if err != nil {
		return nil, errors.WithMessage(err, "unable to load identity config")
	}

	// load metrics config
	c.metricsConfig, err = sdk.loadMetricsConfig(configBackend...)
	if err != nil {
		return nil, errors.WithMessage(err, "unable to load metrics config")
	}

	sdk.opts.ConfigBackend = configBackend

	return c, nil
}

//loadEndpointConfig loads config from config backend when configs are not provided through opts or override missing interfaces from opts with config backend
func (sdk *FabricSDK) loadEndpointConfig(configBackend ...core.ConfigBackend) (fab.EndpointConfig, error) {
	endpointConfigOpt, ok := sdk.opts.endpointConfig.(*fabImpl.EndpointConfigOptions)

	// if optional endpoint was nil or not all of its sub interface functions were overridden,
	// then get default endpoint config and override the functions that were not overridden by opts
	if sdk.opts.endpointConfig == nil || (ok && !fabImpl.IsEndpointConfigFullyOverridden(endpointConfigOpt)) {
		defEndpointConfig, err := fabImpl.ConfigFromBackend(configBackend...)
		if err != nil {
			return nil, errors.WithMessage(err, "failed to initialize endpoint config from config backend")
		}

		// if opts.endpointConfig was not provided during WithEndpointConfig(opts...) call, then return default endpointConfig
		if sdk.opts.endpointConfig == nil {
			return defEndpointConfig, nil
		}
		// else fill any empty interface from opts with defEndpointConfig interface (set default function for ones not provided by WithEndpointConfig() call) and return
		return fabImpl.UpdateMissingOptsWithDefaultConfig(endpointConfigOpt, defEndpointConfig), nil
	}
	// if optional endpoint config was completely overridden but !ok, then return an error
	if !ok {
		return nil, errors.New("failed to retrieve config endpoint from opts")
	}
	return endpointConfigOpt, nil
}

func (sdk *FabricSDK) loadCryptoConfig(configBackend ...core.ConfigBackend) (core.CryptoSuiteConfig, error) {
	cryptoConfigOpt, ok := sdk.opts.CryptoSuiteConfig.(*cryptosuite.CryptoConfigOptions)

	if sdk.opts.CryptoSuiteConfig == nil || (ok && !cryptosuite.IsCryptoConfigFullyOverridden(cryptoConfigOpt)) {
		defCryptoConfig := cryptosuite.ConfigFromBackend(configBackend...)

		if sdk.opts.CryptoSuiteConfig == nil {
			return defCryptoConfig, nil
		}

		return cryptosuite.UpdateMissingOptsWithDefaultConfig(cryptoConfigOpt, defCryptoConfig), nil
	}

	if !ok {
		return nil, errors.New("failed to retrieve crypto suite configs from opts")
	}

	return cryptoConfigOpt, nil
}

func (sdk *FabricSDK) loadIdentityConfig(configBackend ...core.ConfigBackend) (msp.IdentityConfig, error) {
	identityConfigOpt, ok := sdk.opts.IdentityConfig.(*mspImpl.IdentityConfigOptions)

	if sdk.opts.IdentityConfig == nil || (ok && !mspImpl.IsIdentityConfigFullyOverridden(identityConfigOpt)) {
		defIdentityConfig, err := mspImpl.ConfigFromBackend(configBackend...)
		if err != nil {
			return nil, errors.WithMessage(err, "failed to initialize identity config from config backend")
		}

		if sdk.opts.IdentityConfig == nil {
			return defIdentityConfig, nil
		}

		return mspImpl.UpdateMissingOptsWithDefaultConfig(identityConfigOpt, defIdentityConfig), nil
	}

	if !ok {
		return nil, errors.New("failed to retrieve identity configs from opts")
	}

	return identityConfigOpt, nil
}

func (sdk *FabricSDK) loadMetricsConfig(configBackend ...core.ConfigBackend) (metricsCfg.MetricsConfig, error) {
	metricsConfigOpt, ok := sdk.opts.metricsConfig.(*metricsCfg.OperationsConfigOptions)

	if sdk.opts.metricsConfig == nil || (ok && !metricsCfg.IsMetricsConfigFullyOverridden(metricsConfigOpt)) {
		defMetricsConfig, err := metricsCfg.ConfigFromBackend(configBackend...)
		if err != nil {
			return nil, errors.WithMessage(err, "failed to initialize metrics config from config backend")
		}
		if sdk.opts.metricsConfig == nil {
			return defMetricsConfig, nil
		}

		return metricsCfg.UpdateMissingOptsWithDefaultConfig(metricsConfigOpt, defMetricsConfig), nil
	}

	if !ok {
		return nil, errors.New("failed to retrieve metrics configs from opts")
	}

	return metricsConfigOpt, nil
}

func withErrorHandlerProviderOpt(value fab.ErrorHandler) coptions.Opt {
	return func(p coptions.Params) {
		if setter, ok := p.(errHandlerSetter); ok {
			logger.Debugf("... setting error handler")
			setter.SetErrorHandler(value)
		}
	}
}

type errHandlerSetter interface {
	SetErrorHandler(value fab.ErrorHandler)
}
