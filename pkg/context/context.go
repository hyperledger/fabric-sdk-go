/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package context

import (
	reqContext "context"

	"github.com/pkg/errors"

	"time"

	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/context"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/core"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/fab"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/msp"
	"github.com/hyperledger/fabric-sdk-go/pkg/fabsdk/metrics"
)

// Client supplies the configuration and signing identity to client objects.
type Client struct {
	context.Providers
	msp.SigningIdentity
}

// Local supplies the configuration and signing identity to
// clients that will be invoking the peer outside of a channel
// context using an identity in the peer's local MSP.
type Local struct {
	context.Client
	localDiscovery fab.DiscoveryService
}

//LocalDiscoveryService returns core discovery service
func (c *Local) LocalDiscoveryService() fab.DiscoveryService {
	return c.localDiscovery
}

//Channel supplies the configuration for channel context client
type Channel struct {
	context.Client
	channelService fab.ChannelService
	channelID      string
	metrics        *metrics.ClientMetrics
}

//Providers returns core providers
func (c *Channel) Providers() context.Client {
	return c
}

//ChannelService returns channel service
func (c *Channel) ChannelService() fab.ChannelService {
	return c.channelService
}

//ChannelID returns channel id
func (c *Channel) ChannelID() string {
	return c.channelID
}

//Provider implementation of Providers interface
type Provider struct {
	cryptoSuiteConfig      core.CryptoSuiteConfig
	endpointConfig         fab.EndpointConfig
	identityConfig         msp.IdentityConfig
	userStore              msp.UserStore
	cryptoSuite            core.CryptoSuite
	localDiscoveryProvider fab.LocalDiscoveryProvider
	signingManager         core.SigningManager
	idMgmtProvider         msp.IdentityManagerProvider
	infraProvider          fab.InfraProvider
	channelProvider        fab.ChannelProvider
	clientMetrics          *metrics.ClientMetrics
}

// CryptoSuite returns the BCCSP provider of sdk.
func (c *Provider) CryptoSuite() core.CryptoSuite {
	return c.cryptoSuite
}

// IdentityManager returns identity manager for organization
func (c *Provider) IdentityManager(orgName string) (msp.IdentityManager, bool) {
	return c.idMgmtProvider.IdentityManager(orgName)
}

// SigningManager returns signing manager
func (c *Provider) SigningManager() core.SigningManager {
	return c.signingManager
}

// UserStore returns state store
func (c *Provider) UserStore() msp.UserStore {
	return c.userStore
}

//IdentityConfig returns the Identity config
func (c *Provider) IdentityConfig() msp.IdentityConfig {
	return c.identityConfig
}

// LocalDiscoveryProvider returns the local discovery provider
func (c *Provider) LocalDiscoveryProvider() fab.LocalDiscoveryProvider {
	return c.localDiscoveryProvider
}

// ChannelProvider provides channel services.
func (c *Provider) ChannelProvider() fab.ChannelProvider {
	return c.channelProvider
}

// InfraProvider provides fabric objects such as peer and user
func (c *Provider) InfraProvider() fab.InfraProvider {
	return c.infraProvider
}

//EndpointConfig returns end point network config
func (c *Provider) EndpointConfig() fab.EndpointConfig {
	return c.endpointConfig
}

// GetMetrics will return the SDK's metrics instance
func (c *Provider) GetMetrics() *metrics.ClientMetrics {
	return c.clientMetrics
}

//SDKContextParams parameter for creating FabContext
type SDKContextParams func(opts *Provider)

//WithCryptoSuiteConfig sets core cryptoSuite config to Context Provider
func WithCryptoSuiteConfig(cryptoSuiteConfig core.CryptoSuiteConfig) SDKContextParams {
	return func(ctx *Provider) {
		ctx.cryptoSuiteConfig = cryptoSuiteConfig
	}
}

//WithEndpointConfig sets fab endpoint network config to Context Provider
func WithEndpointConfig(endpointConfig fab.EndpointConfig) SDKContextParams {
	return func(ctx *Provider) {
		ctx.endpointConfig = endpointConfig
	}
}

//WithIdentityConfig sets msp identity config to Context Provider
func WithIdentityConfig(identityConfig msp.IdentityConfig) SDKContextParams {
	return func(ctx *Provider) {
		ctx.identityConfig = identityConfig
	}
}

// WithUserStore sets user store to Context Provider
func WithUserStore(userStore msp.UserStore) SDKContextParams {
	return func(ctx *Provider) {
		ctx.userStore = userStore
	}
}

//WithCryptoSuite sets cryptosuite parameter to Context Provider
func WithCryptoSuite(cryptoSuite core.CryptoSuite) SDKContextParams {
	return func(ctx *Provider) {
		ctx.cryptoSuite = cryptoSuite
	}
}

//WithLocalDiscoveryProvider sets the local discovery provider
func WithLocalDiscoveryProvider(discoveryProvider fab.LocalDiscoveryProvider) SDKContextParams {
	return func(ctx *Provider) {
		ctx.localDiscoveryProvider = discoveryProvider
	}
}

//WithSigningManager sets signingManager to Context Provider
func WithSigningManager(signingManager core.SigningManager) SDKContextParams {
	return func(ctx *Provider) {
		ctx.signingManager = signingManager
	}
}

//WithIdentityManagerProvider sets IdentityManagerProvider maps to context
func WithIdentityManagerProvider(provider msp.IdentityManagerProvider) SDKContextParams {
	return func(ctx *Provider) {
		ctx.idMgmtProvider = provider
	}
}

//WithInfraProvider sets infraProvider maps to Context Provider
func WithInfraProvider(infraProvider fab.InfraProvider) SDKContextParams {
	return func(ctx *Provider) {
		ctx.infraProvider = infraProvider
	}
}

//WithChannelProvider sets channelProvider to Context Provider
func WithChannelProvider(channelProvider fab.ChannelProvider) SDKContextParams {
	return func(ctx *Provider) {
		ctx.channelProvider = channelProvider
	}
}

//WithClientMetrics sets clientMetrics to Context Provider
func WithClientMetrics(cm *metrics.ClientMetrics) SDKContextParams {
	return func(ctx *Provider) {
		ctx.clientMetrics = cm
	}
}

//NewProvider creates new context client provider
// Not be used by end developers, fabsdk package use only
func NewProvider(params ...SDKContextParams) *Provider {
	ctxProvider := Provider{}
	for _, param := range params {
		param(&ctxProvider)
	}
	return &ctxProvider
}

// localServiceInit interface allows for initializing services
// with the provided local context
type localServiceInit interface {
	Initialize(context context.Local) error
}

//NewLocal returns a new local context
func NewLocal(clientProvider context.ClientProvider) (*Local, error) {
	client, err := clientProvider()
	if err != nil {
		return nil, errors.WithMessage(err, "failed to get client context to create local context")
	}

	discoveryService, err := client.LocalDiscoveryProvider().CreateLocalDiscoveryService(client.Identifier().MSPID)
	if err != nil {
		return nil, errors.WithMessage(err, "failed to create local discovery service")
	}

	local := &Local{
		Client:         client,
		localDiscovery: discoveryService,
	}

	if ci, ok := discoveryService.(localServiceInit); ok {
		if err := ci.Initialize(local); err != nil {
			return nil, err
		}
	}

	return local, nil
}

// serviceInit interface allows for initializing services
// with the provided channel context
type serviceInit interface {
	Initialize(context context.Channel) error
}

//NewChannel creates new channel context client
// Not be used by end developers, fabsdk package use only
func NewChannel(clientProvider context.ClientProvider, channelID string) (*Channel, error) {

	client, err := clientProvider()
	if err != nil {
		return nil, errors.WithMessage(err, "failed to get client context to create channel client")
	}

	channelService, err := client.ChannelProvider().ChannelService(client, channelID)
	if err != nil {
		return nil, errors.WithMessage(err, "failed to get channel service to create channel client")
	}

	channel := &Channel{
		Client:         client,
		channelService: channelService,
		channelID:      channelID,
		metrics:        client.GetMetrics(),
	}
	if pi, ok := channelService.(serviceInit); ok {
		if err := pi.Initialize(channel); err != nil {
			return nil, err
		}
	}
	return channel, nil
}

type reqContextKey string

//ReqContextTimeoutOverrides key for grpc context value of timeout overrides
var ReqContextTimeoutOverrides = reqContextKey("timeout-overrides")
var reqContextCommManager = reqContextKey("commManager")
var reqContextClient = reqContextKey("clientContext")

//WithTimeoutType sets timeout by type defined in config to request context
func WithTimeoutType(timeoutType fab.TimeoutType) ReqContextOptions {
	return func(ctx *requestContextOpts) {
		ctx.timeoutType = timeoutType
	}
}

//WithTimeout sets timeout time duration to request context
func WithTimeout(timeout time.Duration) ReqContextOptions {
	return func(ctx *requestContextOpts) {
		ctx.timeout = timeout
	}
}

//WithParent sets existing reqContext as a parent ReqContext
func WithParent(context reqContext.Context) ReqContextOptions {
	return func(ctx *requestContextOpts) {
		ctx.parentContext = context
	}
}

//ReqContextOptions parameter for creating requestContext
type ReqContextOptions func(opts *requestContextOpts)

type requestContextOpts struct {
	timeoutType   fab.TimeoutType
	timeout       time.Duration
	parentContext reqContext.Context
}

// NewRequest creates a request-scoped context.
func NewRequest(client context.Client, options ...ReqContextOptions) (reqContext.Context, reqContext.CancelFunc) {

	//'-1' to get default config timeout when timeout options not passed
	reqCtxOpts := requestContextOpts{timeoutType: -1}
	for _, option := range options {
		option(&reqCtxOpts)
	}

	parentContext := reqCtxOpts.parentContext
	if parentContext == nil {
		//when parent request context not set, use background context
		parentContext = reqContext.Background()
	}

	var timeout time.Duration
	if reqCtxOpts.timeout > 0 {
		timeout = reqCtxOpts.timeout
	} else if timeoutOverride := requestTimeoutOverride(parentContext, reqCtxOpts.timeoutType); timeoutOverride > 0 {
		timeout = timeoutOverride
	} else {
		timeout = client.EndpointConfig().Timeout(reqCtxOpts.timeoutType)
	}

	ctx := reqContext.WithValue(parentContext, reqContextCommManager, client.InfraProvider().CommManager())
	ctx = reqContext.WithValue(ctx, reqContextClient, client)
	ctx, cancel := reqContext.WithTimeout(ctx, timeout)

	return ctx, cancel
}

// RequestCommManager extracts the CommManager from the request-scoped context.
func RequestCommManager(ctx reqContext.Context) (fab.CommManager, bool) {
	commManager, ok := ctx.Value(reqContextCommManager).(fab.CommManager)
	return commManager, ok
}

// RequestClientContext extracts the Client Context from the request-scoped context.
func RequestClientContext(ctx reqContext.Context) (context.Client, bool) {
	clientContext, ok := ctx.Value(reqContextClient).(context.Client)
	return clientContext, ok
}

// requestTimeoutOverrides extracts the timeout from timeout override map from the request-scoped context.
func requestTimeoutOverride(ctx reqContext.Context, timeoutType fab.TimeoutType) time.Duration {
	timeoutOverrides, ok := ctx.Value(ReqContextTimeoutOverrides).(map[fab.TimeoutType]time.Duration)
	if !ok {
		return 0
	}
	return timeoutOverrides[timeoutType]
}
