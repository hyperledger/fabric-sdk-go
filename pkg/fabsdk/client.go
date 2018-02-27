/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package fabsdk

import (
	"github.com/hyperledger/fabric-sdk-go/pkg/client/channel"
	"github.com/hyperledger/fabric-sdk-go/pkg/client/ledger"
	"github.com/hyperledger/fabric-sdk-go/pkg/client/resmgmt"
	"github.com/hyperledger/fabric-sdk-go/pkg/context"
	"github.com/hyperledger/fabric-sdk-go/pkg/context/api/core"
	"github.com/hyperledger/fabric-sdk-go/pkg/context/api/fab"
	"github.com/hyperledger/fabric-sdk-go/pkg/fabsdk/api"
	"github.com/pkg/errors"
)

// ClientContext  represents the fabric transaction clients
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
	opts          *contextOptions
	identity      context.IdentityContext
	providers     providers
	clientFactory api.SessionClientFactory
}

type providers interface {
	api.Providers
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
func (sdk *FabricSDK) NewClient(identityOpt IdentityOption, opts ...ContextOption) *ClientContext {
	// delay execution of the following logic to avoid error return from this function.
	// this is done to allow a cleaner API - i.e., client, err := sdk.NewClient(args).<Desired Interface>(extra args)
	provider := func() (*clientContext, error) {
		o, err := newContextOptions(sdk.config, opts)
		if err != nil {
			return nil, errors.WithMessage(err, "unable to retrieve configuration from SDK")
		}

		identity, err := sdk.newIdentity(o.orgID, identityOpt)
		if err != nil {
			return nil, errors.WithMessage(err, "unable to create client context")
		}

		cc := clientContext{
			opts:          o,
			identity:      identity,
			providers:     sdk.context(),
			clientFactory: sdk.opts.Session,
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

	fabProvider := p.providers.FabricProvider()
	resource, err := fabProvider.CreateResourceClient(session)
	if err != nil {
		return nil, err
	}

	discovery := p.providers.DiscoveryProvider()
	chProvider := p.providers.ChannelProvider()

	ctx := resmgmt.Context{
		ProviderContext:   p.providers,
		IdentityContext:   session,
		Resource:          resource,
		DiscoveryProvider: discovery,
		ChannelProvider:   chProvider,
		FabricProvider:    fabProvider,
	}

	return resmgmt.New(ctx, resmgmt.WithDefaultTargetFilter(o.targetFilter))

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

	discovery := p.providers.DiscoveryProvider()
	discService, err := discovery.NewDiscoveryService(id)
	if err != nil {
		return nil, err
	}

	ctx := ledger.Context{
		ProviderContext:  p.providers,
		IdentityContext:  session,
		DiscoveryService: discService,
	}

	return ledger.New(ctx, id, ledger.WithDefaultTargetFilter(o.targetFilter))

}

// Channel returns a client API for transacting on a channel.
func (c *ClientContext) Channel(id string, opts ...ClientOption) (*channel.Client, error) {
	p, err := c.provider()
	if err != nil {
		return &channel.Client{}, errors.WithMessage(err, "unable to get client provider context")
	}
	o, err := newClientOptions(opts)
	if err != nil {
		return &channel.Client{}, errors.WithMessage(err, "unable to retrieve client options")
	}
	session := newSession(p.identity, p.providers.ChannelProvider())
	client, err := p.clientFactory.CreateChannelClient(p.providers, session, id, o.targetFilter)
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

// Session returns the underlying identity of the client.
//
// Deprecated: this method is temporary.
func (c *ClientContext) Session() (context.SessionContext, error) {
	p, err := c.provider()
	if err != nil {
		return nil, errors.WithMessage(err, "unable to get client provider context")
	}

	return newSession(p.identity, p.providers.ChannelProvider()), nil
}
