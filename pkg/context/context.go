/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package context

import (
	reqContext "context"
	"strings"

	"github.com/pkg/errors"

	"github.com/hyperledger/fabric-sdk-go/pkg/common/context"
	"github.com/hyperledger/fabric-sdk-go/pkg/context/api/core"
	"github.com/hyperledger/fabric-sdk-go/pkg/context/api/fab"
	"github.com/hyperledger/fabric-sdk-go/pkg/context/api/msp"
)

// Client supplies the configuration and signing identity to client objects.
type Client struct {
	context.Providers
	msp.Identity
}

//Channel supplies the configuration for channel context client
type Channel struct {
	context.Client
	discovery      fab.DiscoveryService
	selection      fab.SelectionService
	channelService fab.ChannelService
}

//Providers returns core providers
func (c *Channel) Providers() context.Client {
	return c
}

//DiscoveryService returns core discovery service
func (c *Channel) DiscoveryService() fab.DiscoveryService {
	return c.discovery
}

//SelectionService returns selection service
func (c *Channel) SelectionService() fab.SelectionService {
	return c.selection
}

//ChannelService returns channel service
func (c *Channel) ChannelService() fab.ChannelService {
	return c.channelService
}

//Provider implementation for Providers interface
type Provider struct {
	config            core.Config
	stateStore        core.KVStore
	cryptoSuite       core.CryptoSuite
	discoveryProvider fab.DiscoveryProvider
	selectionProvider fab.SelectionProvider
	signingManager    core.SigningManager
	identityManager   map[string]msp.IdentityManager
	infraProvider     fab.InfraProvider
	channelProvider   fab.ChannelProvider
}

// Config returns the Config provider of sdk.
func (c *Provider) Config() core.Config {
	return c.config
}

// CryptoSuite returns the BCCSP provider of sdk.
func (c *Provider) CryptoSuite() core.CryptoSuite {
	return c.cryptoSuite
}

// IdentityManager returns identity manager for organization
func (c *Provider) IdentityManager(orgName string) (msp.IdentityManager, bool) {
	mgr, ok := c.identityManager[strings.ToLower(orgName)]
	return mgr, ok
}

// SigningManager returns signing manager
func (c *Provider) SigningManager() core.SigningManager {
	return c.signingManager
}

// StateStore returns state store
func (c *Provider) StateStore() core.KVStore {
	return c.stateStore
}

// DiscoveryProvider returns discovery provider
func (c *Provider) DiscoveryProvider() fab.DiscoveryProvider {
	return c.discoveryProvider
}

// SelectionProvider returns selection provider
func (c *Provider) SelectionProvider() fab.SelectionProvider {
	return c.selectionProvider
}

// ChannelProvider provides channel services.
func (c *Provider) ChannelProvider() fab.ChannelProvider {
	return c.channelProvider
}

// InfraProvider provides fabric objects such as peer and user
func (c *Provider) InfraProvider() fab.InfraProvider {
	return c.infraProvider
}

//SDKContextParams parameter for creating FabContext
type SDKContextParams func(opts *Provider)

//WithConfig sets config to FabContext
func WithConfig(config core.Config) SDKContextParams {
	return func(ctx *Provider) {
		ctx.config = config
	}
}

//WithStateStore sets state store to FabContext
func WithStateStore(stateStore core.KVStore) SDKContextParams {
	return func(ctx *Provider) {
		ctx.stateStore = stateStore
	}
}

//WithCryptoSuite sets cryptosuite parameter to FabContext
func WithCryptoSuite(cryptoSuite core.CryptoSuite) SDKContextParams {
	return func(ctx *Provider) {
		ctx.cryptoSuite = cryptoSuite
	}
}

//WithDiscoveryProvider sets discoveryProvider to FabContext
func WithDiscoveryProvider(discoveryProvider fab.DiscoveryProvider) SDKContextParams {
	return func(ctx *Provider) {
		ctx.discoveryProvider = discoveryProvider
	}
}

//WithSelectionProvider sets selectionProvider to FabContext
func WithSelectionProvider(selectionProvider fab.SelectionProvider) SDKContextParams {
	return func(ctx *Provider) {
		ctx.selectionProvider = selectionProvider
	}
}

//WithSigningManager sets signingManager to FabContext
func WithSigningManager(signingManager core.SigningManager) SDKContextParams {
	return func(ctx *Provider) {
		ctx.signingManager = signingManager
	}
}

//WithIdentityManager sets identityManagers maps to context
func WithIdentityManager(identityManagers map[string]msp.IdentityManager) SDKContextParams {
	return func(ctx *Provider) {
		ctx.identityManager = identityManagers
	}
}

//WithInfraProvider sets infraProvider maps to FabContext
func WithInfraProvider(infraProvider fab.InfraProvider) SDKContextParams {
	return func(ctx *Provider) {
		ctx.infraProvider = infraProvider
	}
}

//WithChannelProvider sets channelProvider to FabContext
func WithChannelProvider(channelProvider fab.ChannelProvider) SDKContextParams {
	return func(ctx *Provider) {
		ctx.channelProvider = channelProvider
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

	discoveryService, err := client.DiscoveryProvider().CreateDiscoveryService(channelID)
	if err != nil {
		return nil, errors.WithMessage(err, "failed to get discovery service to create channel client")
	}

	selectionService, err := client.SelectionProvider().CreateSelectionService(channelID)
	if err != nil {
		return nil, errors.WithMessage(err, "failed to get selection service to create channel client")
	}

	return &Channel{
		Client:         client,
		selection:      selectionService,
		discovery:      discoveryService,
		channelService: channelService,
	}, nil
}

type reqContextKey string

var reqContextCommManager = reqContextKey("commManager")

// NewRequest creates a request-scoped context.
func NewRequest(client context.Client) reqContext.Context {
	ctx := reqContext.WithValue(reqContext.Background(), reqContextCommManager, client.InfraProvider().CommManager())
	return ctx
}

// RequestCommManager extracts the CommManager from the request-scoped context.
func RequestCommManager(ctx reqContext.Context) (fab.CommManager, bool) {
	commManager, ok := ctx.Value(reqContextCommManager).(fab.CommManager)
	return commManager, ok
}
