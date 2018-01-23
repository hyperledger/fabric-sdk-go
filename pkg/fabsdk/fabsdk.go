/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

// Package fabsdk enables client usage of a Hyperledger Fabric network.
package fabsdk

import (
	"github.com/hyperledger/fabric-sdk-go/api/apiconfig"
	"github.com/hyperledger/fabric-sdk-go/api/apicore"
	"github.com/hyperledger/fabric-sdk-go/api/apifabclient"
	"github.com/hyperledger/fabric-sdk-go/api/apilogging"

	"github.com/hyperledger/fabric-sdk-go/api/apicryptosuite"
	"github.com/hyperledger/fabric-sdk-go/pkg/cryptosuite"
	"github.com/hyperledger/fabric-sdk-go/pkg/errors"
	apisdk "github.com/hyperledger/fabric-sdk-go/pkg/fabsdk/api"
	"github.com/hyperledger/fabric-sdk-go/pkg/logging"
)

// FabricSDK provides access (and context) to clients being managed by the SDK.
type FabricSDK struct {
	opts options

	configProvider    apiconfig.Config
	stateStore        apifabclient.KeyValueStore
	cryptoSuite       apicryptosuite.CryptoSuite
	discoveryProvider apifabclient.DiscoveryProvider
	selectionProvider apifabclient.SelectionProvider
	signingManager    apifabclient.SigningManager
	fabricProvider    apicore.FabricProvider
}

type options struct {
	Core    apisdk.CoreProviderFactory
	Service apisdk.ServiceProviderFactory
	Context apisdk.OrgClientFactory
	Session apisdk.SessionClientFactory
	Logger  apilogging.LoggerProvider
}

// Option configures the SDK.
type Option func(opts *options) error

// New initializes the SDK based on the set of options provided.
// configProvider provides the application configuration.
func New(cp apiconfig.ConfigProvider, opts ...Option) (*FabricSDK, error) {
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
func WithConfig(config apiconfig.Config) apiconfig.ConfigProvider {
	return func() (apiconfig.Config, error) {
		return config, nil
	}
}

// fromPkgSuite creates an SDK based on the implementations in the provided pkg suite.
// TODO: For now leaving this method as private until we have more usage.
func fromPkgSuite(configProvider apiconfig.Config, pkgSuite apisdk.PkgSuite, opts ...Option) (*FabricSDK, error) {
	core, err := pkgSuite.Core()
	if err != nil {
		return nil, errors.WithMessage(err, "Unable to initialize core pkg")
	}

	svc, err := pkgSuite.Service()
	if err != nil {
		return nil, errors.WithMessage(err, "Unable to initialize service pkg")
	}

	ctx, err := pkgSuite.Context()
	if err != nil {
		return nil, errors.WithMessage(err, "Unable to initialize context pkg")
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
			Context: ctx,
			Session: sess,
			Logger:  lg,
		},
		configProvider: configProvider,
	}

	err = initSDK(&sdk, opts)
	if err != nil {
		return nil, err
	}

	return &sdk, err
}

// WithCorePkg injects the core implementation into the SDK.
func WithCorePkg(core apisdk.CoreProviderFactory) Option {
	return func(opts *options) error {
		opts.Core = core
		return nil
	}
}

// WithServicePkg injects the service implementation into the SDK.
func WithServicePkg(service apisdk.ServiceProviderFactory) Option {
	return func(opts *options) error {
		opts.Service = service
		return nil
	}
}

// WithContextPkg injects the context implementation into the SDK.
func WithContextPkg(context apisdk.OrgClientFactory) Option {
	return func(opts *options) error {
		opts.Context = context
		return nil
	}
}

// WithSessionPkg injects the session implementation into the SDK.
func WithSessionPkg(session apisdk.SessionClientFactory) Option {
	return func(opts *options) error {
		opts.Session = session
		return nil
	}
}

// WithLoggerPkg injects the logger implementation into the SDK.
func WithLoggerPkg(logger apilogging.LoggerProvider) Option {
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
	cs, err := sdk.opts.Core.NewCryptoSuiteProvider(sdk.configProvider)
	if err != nil {
		return errors.WithMessage(err, "failed to initialize crypto suite")
	}

	sdk.cryptoSuite = cs

	// Setting this cryptosuite as the factory default
	cryptosuite.SetDefault(cs)

	// Initialize state store
	store, err := sdk.opts.Core.NewStateStoreProvider(sdk.configProvider)
	if err != nil {
		return errors.WithMessage(err, "failed to initialize state store")
	}
	sdk.stateStore = store

	// Initialize Signing Manager
	signingMgr, err := sdk.opts.Core.NewSigningManager(sdk.cryptoSuite, sdk.configProvider)
	if err != nil {
		return errors.WithMessage(err, "failed to initialize signing manager")
	}
	sdk.signingManager = signingMgr

	// Initialize Fabric Provider
	fabricProvider, err := sdk.opts.Core.NewFabricProvider(sdk.configProvider, sdk.stateStore, sdk.cryptoSuite, sdk.signingManager)
	if err != nil {
		return errors.WithMessage(err, "failed to initialize core fabric provider")
	}
	sdk.fabricProvider = fabricProvider

	// Initialize discovery provider
	discoveryProvider, err := sdk.opts.Service.NewDiscoveryProvider(sdk.configProvider)
	if err != nil {
		return errors.WithMessage(err, "failed to initialize discovery provider")
	}
	if pi, ok := discoveryProvider.(providerInit); ok {
		pi.Initialize(sdk)
	}
	sdk.discoveryProvider = discoveryProvider

	// Initialize selection provider (for selecting endorsing peers)
	selectionProvider, err := sdk.opts.Service.NewSelectionProvider(sdk.configProvider)
	if err != nil {
		return errors.WithMessage(err, "failed to initialize selection provider")
	}
	if pi, ok := selectionProvider.(providerInit); ok {
		pi.Initialize(sdk)
	}
	sdk.selectionProvider = selectionProvider

	return nil
}

// ConfigProvider returns the SDK's configuration.
// TODO rename to Config
func (sdk *FabricSDK) ConfigProvider() apiconfig.Config {
	return sdk.configProvider
}

func (sdk *FabricSDK) context() *sdkContext {
	c := sdkContext{
		sdk: sdk,
	}
	return &c
}

func (sdk *FabricSDK) newUser(orgID string, userName string) (apifabclient.IdentityContext, error) {

	credentialMgr, err := sdk.opts.Context.NewCredentialManager(orgID, sdk.configProvider, sdk.cryptoSuite)
	if err != nil {
		return nil, errors.WithMessage(err, "failed to get credential manager")
	}

	signingIdentity, err := credentialMgr.GetSigningIdentity(userName)
	if err != nil {
		return nil, errors.WithMessage(err, "failed to get signing identity")
	}

	user, err := sdk.fabricProvider.NewUser(userName, signingIdentity)
	if err != nil {
		return nil, errors.WithMessage(err, "NewPreEnrolledUser returned error")
	}

	return user, nil
}
