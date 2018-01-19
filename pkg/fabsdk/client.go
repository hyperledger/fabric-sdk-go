/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package fabsdk

import (
	"github.com/hyperledger/fabric-sdk-go/api/apiconfig"
	"github.com/hyperledger/fabric-sdk-go/api/apifabclient"
	"github.com/hyperledger/fabric-sdk-go/api/apitxn"
	chmgmt "github.com/hyperledger/fabric-sdk-go/api/apitxn/chmgmtclient"
	resmgmt "github.com/hyperledger/fabric-sdk-go/api/apitxn/resmgmtclient"
	"github.com/hyperledger/fabric-sdk-go/pkg/errors"
	apisdk "github.com/hyperledger/fabric-sdk-go/pkg/fabsdk/api"
)

// Client represents the fabric transaction clients
type Client struct {
	opts          *clientOptions
	identity      apifabclient.User
	providers     apisdk.SDK
	clientFactory apisdk.SessionClientFactory
}

// ClientOption configures the clients created by the SDK.
type ClientOption func(opts *clientOptions) error

type clientOptions struct {
	orgID          string
	configProvider apiconfig.Config
	targetFilter   resmgmt.TargetFilter
}

// WithOrg uses the configuration and users from the named organization.
func WithOrg(name string) ClientOption {
	return func(opts *clientOptions) error {
		opts.orgID = name
		return nil
	}
}

// WithTargetFilter allows for filtering target peers.
func WithTargetFilter(targetFilter resmgmt.TargetFilter) ClientOption {
	return func(opts *clientOptions) error {
		opts.targetFilter = targetFilter
		return nil
	}
}

// withConfig allows for overriding the configuration of the client.
// TODO: This should be removed once the depreacted functions are removed.
func withConfig(configProvider apiconfig.Config) ClientOption {
	return func(opts *clientOptions) error {
		opts.configProvider = configProvider
		return nil
	}
}

// NewClient allows creation of transactions using the supplied identity as the credential.
func (sdk *FabricSDK) NewClient(identityOpt IdentityOption, opts ...ClientOption) (*Client, error) {
	o, err := newClientOptions(sdk.ConfigProvider(), opts)
	if err != nil {
		return nil, errors.WithMessage(err, "unable to retrieve configuration from SDK")
	}

	identity, err := sdk.newIdentity(o.orgID, identityOpt)
	if err != nil {
		return nil, errors.WithMessage(err, "unable to create client context")
	}

	client := Client{
		opts:          o,
		identity:      identity,
		providers:     sdk,
		clientFactory: sdk.opts.Session,
	}
	return &client, nil
}

func newClientOptions(config apiconfig.Config, options []ClientOption) (*clientOptions, error) {
	// Read default org name from configuration
	client, err := config.Client()
	if err != nil {
		return nil, errors.WithMessage(err, "unable to retrieve client from network config")
	}

	opts := clientOptions{
		orgID:          client.Organization,
		configProvider: config,
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

// ChannelMgmt returns a client API for managing channels
func (c *Client) ChannelMgmt() (chmgmt.ChannelMgmtClient, error) {
	session := newSession(c.identity)
	client, err := c.clientFactory.NewChannelMgmtClient(c.providers, session, c.opts.configProvider)
	if err != nil {
		return nil, errors.WithMessage(err, "failed to create new channel management client")
	}

	return client, nil
}

// ResourceMgmt returns a client API for managing system resources
func (c *Client) ResourceMgmt() (resmgmt.ResourceMgmtClient, error) {
	session := newSession(c.identity)
	client, err := c.clientFactory.NewResourceMgmtClient(c.providers, session, c.opts.configProvider, c.opts.targetFilter)
	if err != nil {
		return nil, errors.WithMessage(err, "failed to created new resource management client")
	}

	return client, nil
}

// Channel returns a client API for transacting on a channel
func (c *Client) Channel(id string) (apitxn.ChannelClient, error) {
	session := newSession(c.identity)
	client, err := c.clientFactory.NewChannelClient(c.providers, session, c.opts.configProvider, id)
	if err != nil {
		return nil, errors.WithMessage(err, "failed to created new resource management client")
	}

	return client, nil
}

// NewClientChannelMgmt returns a new client for managing channels
func (sdk *FabricSDK) NewClientChannelMgmt(identity IdentityOption, opts ...ClientOption) (chmgmt.ChannelMgmtClient, error) {
	c, err := sdk.NewClient(identity, opts...)
	if err != nil {
		return nil, errors.WithMessage(err, "error creating client from SDK")
	}

	return c.ChannelMgmt()
}

// NewClientResourceMgmt returns a new client for managing system resources
func (sdk *FabricSDK) NewClientResourceMgmt(identity IdentityOption, opts ...ClientOption) (resmgmt.ResourceMgmtClient, error) {
	c, err := sdk.NewClient(identity, opts...)
	if err != nil {
		return nil, errors.WithMessage(err, "error creating client from SDK")
	}

	return c.ResourceMgmt()
}

// NewClientChannel returns a new client for a channel
func (sdk *FabricSDK) NewClientChannel(identity IdentityOption, channelID string, opts ...ClientOption) (apitxn.ChannelClient, error) {
	c, err := sdk.NewClient(identity, opts...)
	if err != nil {
		return nil, errors.WithMessage(err, "error creating client from SDK")
	}

	return c.Channel(channelID)
}
