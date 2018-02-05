/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package fabsdk

import (
	"github.com/hyperledger/fabric-sdk-go/api/apiconfig"
	"github.com/hyperledger/fabric-sdk-go/api/apifabclient"
	"github.com/hyperledger/fabric-sdk-go/api/apitxn/chclient"
	chmgmt "github.com/hyperledger/fabric-sdk-go/api/apitxn/chmgmtclient"
	resmgmt "github.com/hyperledger/fabric-sdk-go/api/apitxn/resmgmtclient"
	apisdk "github.com/hyperledger/fabric-sdk-go/pkg/fabsdk/api"
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
	config apiconfig.Config
}

// ClientOption configures the clients created by the SDK.
type ClientOption func(opts *clientOptions) error

type clientOptions struct {
	targetFilter apifabclient.TargetFilter
}

type clientProvider func() (*clientContext, error)

type clientContext struct {
	opts          *contextOptions
	identity      apifabclient.IdentityContext
	providers     providers
	clientFactory apisdk.SessionClientFactory
}

type providers interface {
	apisdk.Providers
}

// WithOrg uses the configuration and users from the named organization.
func WithOrg(name string) ContextOption {
	return func(opts *contextOptions) error {
		opts.orgID = name
		return nil
	}
}

// WithTargetFilter allows for filtering target peers.
func WithTargetFilter(targetFilter apifabclient.TargetFilter) ClientOption {
	return func(opts *clientOptions) error {
		opts.targetFilter = targetFilter
		return nil
	}
}

// withConfig allows for overriding the configuration of the client.
// TODO: This should be removed once the depreacted functions are removed.
func withConfig(config apiconfig.Config) ContextOption {
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

func newContextOptions(config apiconfig.Config, options []ContextOption) (*contextOptions, error) {
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

// ChannelMgmt returns a client API for managing channels.
func (c *ClientContext) ChannelMgmt() (chmgmt.ChannelMgmtClient, error) {
	p, err := c.provider()
	if err != nil {
		return nil, errors.WithMessage(err, "unable to get client provider context")
	}

	session := newSession(p.identity, p.providers.ChannelProvider())
	client, err := p.clientFactory.NewChannelMgmtClient(p.providers, session)
	if err != nil {
		return nil, errors.WithMessage(err, "failed to create new channel management client")
	}

	return client, nil
}

// ResourceMgmt returns a client API for managing system resources.
func (c *ClientContext) ResourceMgmt(opts ...ClientOption) (resmgmt.ResourceMgmtClient, error) {
	p, err := c.provider()
	if err != nil {
		return nil, errors.WithMessage(err, "unable to get client provider context")
	}
	o, err := newClientOptions(opts)
	if err != nil {
		return nil, errors.WithMessage(err, "unable to retrieve client options")
	}

	session := newSession(p.identity, p.providers.ChannelProvider())
	client, err := p.clientFactory.NewResourceMgmtClient(p.providers, session, o.targetFilter)
	if err != nil {
		return nil, errors.WithMessage(err, "failed to created new resource management client")
	}

	return client, nil
}

// Channel returns a client API for transacting on a channel.
func (c *ClientContext) Channel(id string, opts ...ClientOption) (chclient.ChannelClient, error) {
	p, err := c.provider()
	if err != nil {
		return nil, errors.WithMessage(err, "unable to get client provider context")
	}
	o, err := newClientOptions(opts)
	if err != nil {
		return nil, errors.WithMessage(err, "unable to retrieve client options")
	}
	session := newSession(p.identity, p.providers.ChannelProvider())
	client, err := p.clientFactory.NewChannelClient(p.providers, session, id, o.targetFilter)
	if err != nil {
		return nil, errors.WithMessage(err, "failed to created new channel client")
	}

	return client, nil
}

// ChannelService returns a client API for interacting with a channel.
func (c *ClientContext) ChannelService(id string) (apifabclient.ChannelService, error) {
	p, err := c.provider()
	if err != nil {
		return nil, errors.WithMessage(err, "unable to get client provider context")
	}

	channelProvider := p.providers.ChannelProvider()
	return channelProvider.NewChannelService(p.identity, id)
}

// Session returns the underlying identity of the client.
//
// Deprecated: this method is temporary.
func (c *ClientContext) Session() (apisdk.SessionContext, error) {
	p, err := c.provider()
	if err != nil {
		return nil, errors.WithMessage(err, "unable to get client provider context")
	}

	return newSession(p.identity, p.providers.ChannelProvider()), nil
}
