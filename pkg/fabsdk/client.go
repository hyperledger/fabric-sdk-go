/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package fabsdk

import (
	"github.com/hyperledger/fabric-sdk-go/pkg/client/channel"
	"github.com/hyperledger/fabric-sdk-go/pkg/client/ledger"
	"github.com/hyperledger/fabric-sdk-go/pkg/client/resmgmt"
	contextApi "github.com/hyperledger/fabric-sdk-go/pkg/common/context"
	"github.com/hyperledger/fabric-sdk-go/pkg/context"
	"github.com/hyperledger/fabric-sdk-go/pkg/context/api/core"
	"github.com/hyperledger/fabric-sdk-go/pkg/context/api/fab"
	"github.com/pkg/errors"
)

// ClientContext  represents the fabric transaction clients
//Deprecated: use context.Client or context.Channel instead
type ClientContext struct {
	provider clientProvider
}

// ContextOption configures the client context created by the SDK.
type ContextOption func(opts *contextOptions) error

type contextOptions struct {
	orgID  string
	config core.Config
}

// ClientOption configures the clients created by the SDK.
type ClientOption func(opts *clientOptions) error

type clientOptions struct {
	targetFilter fab.TargetFilter
}

type clientProvider func() (*clientContext, error)

type clientContext struct {
	opts      *contextOptions
	identity  contextApi.Identity
	providers providers
}

type providers interface {
	contextApi.Providers
}

// WithOrg uses the configuration and users from the named organization.
func WithOrg(name string) ContextOption {
	return func(opts *contextOptions) error {
		opts.orgID = name
		return nil
	}
}

// WithTargetFilter allows for filtering target peers.
func WithTargetFilter(targetFilter fab.TargetFilter) ClientOption {
	return func(opts *clientOptions) error {
		opts.targetFilter = targetFilter
		return nil
	}
}

// withConfig allows for overriding the configuration of the client.
// TODO: This should be removed once the depreacted functions are removed.
func withConfig(config core.Config) ContextOption {
	return func(opts *contextOptions) error {
		opts.config = config
		return nil
	}
}

// NewClient allows creation of transactions using the supplied identity as the credential.
//Deprecated: use sdk.Context() or sdk.ChannelContext() instead
func (sdk *FabricSDK) NewClient(identityOpt IdentityOption, opts ...ContextOption) *ClientContext {
	// delay execution of the following logic to avoid error return from this function.
	// this is done to allow a cleaner API - i.e., client, err := sdk.NewClient(args).<Desired Interface>(extra args)
	provider := func() (*clientContext, error) {
		o, err := newContextOptions(sdk.provider.Config(), opts)
		if err != nil {
			return nil, errors.WithMessage(err, "unable to retrieve configuration from SDK")
		}

		identity, err := sdk.newIdentity(identityOpt, WithOrgName(o.orgID))
		if err != nil {
			return nil, errors.WithMessage(err, "unable to create client context")
		}

		cc := clientContext{
			opts:      o,
			identity:  identity,
			providers: &context.Client{Providers: &sdk.provider, Identity: identity},
		}
		return &cc, nil
	}
	client := ClientContext{
		provider: provider,
	}
	return &client
}

func newContextOptions(config core.Config, options []ContextOption) (*contextOptions, error) {
	// Read default org name from configuration
	client, err := config.Client()
	if err != nil {
		return nil, errors.WithMessage(err, "unable to retrieve client from network config")
	}

	opts := contextOptions{
		orgID:  client.Organization,
		config: config,
	}

	for _, option := range options {
		err := option(&opts)
		if err != nil {
			return nil, errors.WithMessage(err, "error in option passed to client")
		}
	}

	if opts.orgID == "" {
		return nil, errors.New("must provide default organisation name in configuration")
	}

	return &opts, nil
}

func newClientOptions(options []ClientOption) (*clientOptions, error) {
	opts := clientOptions{}

	for _, option := range options {
		err := option(&opts)
		if err != nil {
			return nil, errors.WithMessage(err, "error in option passed to client")
		}
	}

	return &opts, nil
}

// ResourceMgmt returns a client API for managing system resources.
func (c *ClientContext) ResourceMgmt(opts ...ClientOption) (*resmgmt.Client, error) {
	p, err := c.provider()
	if err != nil {
		return nil, errors.WithMessage(err, "unable to get client provider context")
	}

	o, err := newClientOptions(opts)
	if err != nil {
		return nil, errors.WithMessage(err, "unable to retrieve client options")
	}

	session := newSession(p.identity, p.providers.ChannelProvider())

	ctxProvider := c.createClientContext(p.providers, session)

	return resmgmt.New(ctxProvider, resmgmt.WithDefaultTargetFilter(o.targetFilter))

}

// Ledger returns a client API for querying ledger
func (c *ClientContext) Ledger(id string, opts ...ClientOption) (*ledger.Client, error) {
	p, err := c.provider()
	if err != nil {
		return &ledger.Client{}, errors.WithMessage(err, "unable to get client provider context")
	}
	o, err := newClientOptions(opts)
	if err != nil {
		return &ledger.Client{}, errors.WithMessage(err, "unable to retrieve client options")
	}
	session := newSession(p.identity, p.providers.ChannelProvider())

	ctxProvider := c.createClientContext(p.providers, session)

	return ledger.New(ctxProvider, id, ledger.WithDefaultTargetFilter(o.targetFilter))

}

// Channel returns a client API for transacting on a channel.
func (c *ClientContext) Channel(id string, opts ...ClientOption) (*channel.Client, error) {
	p, err := c.provider()
	if err != nil {
		return &channel.Client{}, errors.WithMessage(err, "unable to get client provider context")
	}

	session := newSession(p.identity, p.providers.ChannelProvider())

	clientCtx := c.createClientContext(p.providers, session)

	chCtxProvider := c.createChannelContext(clientCtx, id)

	client, err := channel.New(chCtxProvider)
	if err != nil {
		return &channel.Client{}, errors.WithMessage(err, "failed to created new channel client")
	}

	return client, nil
}

// ChannelService returns a client API for interacting with a channel.
func (c *ClientContext) ChannelService(id string) (fab.ChannelService, error) {
	p, err := c.provider()
	if err != nil {
		return nil, errors.WithMessage(err, "unable to get client provider context")
	}

	channelProvider := p.providers.ChannelProvider()
	return channelProvider.ChannelService(p.identity, id)
}

type clientCtx struct {
	identity  contextApi.Identity
	providers contextApi.Providers
}

// Config returns the Config provider of sdk.
func (c *clientCtx) Config() core.Config {
	return c.providers.Config()
}

// CryptoSuite returns the BCCSP provider of sdk.
func (c *clientCtx) CryptoSuite() core.CryptoSuite {
	return c.providers.CryptoSuite()
}

// IdentityManager returns identity manager for organization
func (c *clientCtx) IdentityManager(orgName string) (core.IdentityManager, bool) {
	return c.providers.IdentityManager(orgName)
}

// SigningManager returns signing manager
func (c *clientCtx) SigningManager() core.SigningManager {
	return c.providers.SigningManager()
}

// StateStore returns state store
func (c *clientCtx) StateStore() core.KVStore {
	return c.providers.StateStore()
}

// DiscoveryProvider returns discovery provider
func (c *clientCtx) DiscoveryProvider() fab.DiscoveryProvider {
	return c.providers.DiscoveryProvider()
}

// SelectionProvider returns selection provider
func (c *clientCtx) SelectionProvider() fab.SelectionProvider {
	return c.providers.SelectionProvider()
}

// ChannelProvider provides channel services.
func (c *clientCtx) ChannelProvider() fab.ChannelProvider {
	return c.providers.ChannelProvider()
}

// FabricProvider provides fabric objects such as peer and user
func (c *clientCtx) FabricProvider() fab.InfraProvider {
	return c.providers.FabricProvider()
}

//MspID returns MSPID
func (c *clientCtx) MspID() string {
	return c.identity.MspID()
}

//SerializedIdentity returns serialized identity
func (c *clientCtx) SerializedIdentity() ([]byte, error) {
	return c.identity.SerializedIdentity()
}

//PrivateKey returns private key
func (c *clientCtx) PrivateKey() core.Key {
	return c.identity.PrivateKey()
}

func (c *ClientContext) createClientContext(providers contextApi.Providers, identity contextApi.Identity) contextApi.ClientProvider {
	return func() (contextApi.Client, error) {
		return &clientCtx{providers: providers, identity: identity}, nil
	}
}

func (c *ClientContext) createChannelContext(clientProvider contextApi.ClientProvider, channelID string) contextApi.ChannelProvider {
	return func() (contextApi.Channel, error) {
		return context.NewChannel(clientProvider, channelID)
	}
}
