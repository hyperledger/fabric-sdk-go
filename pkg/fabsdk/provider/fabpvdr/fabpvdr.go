/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package fabpvdr

import (
	"github.com/hyperledger/fabric-sdk-go/pkg/context"
	contextApi "github.com/hyperledger/fabric-sdk-go/pkg/context/api"
	"github.com/hyperledger/fabric-sdk-go/pkg/context/api/core"
	"github.com/hyperledger/fabric-sdk-go/pkg/context/api/fab"
	fabricCAClient "github.com/hyperledger/fabric-sdk-go/pkg/fab/ca"
	channelImpl "github.com/hyperledger/fabric-sdk-go/pkg/fab/channel"
	"github.com/hyperledger/fabric-sdk-go/pkg/fab/chconfig"
	"github.com/hyperledger/fabric-sdk-go/pkg/fab/events"
	identityImpl "github.com/hyperledger/fabric-sdk-go/pkg/fab/identity"
	"github.com/hyperledger/fabric-sdk-go/pkg/fab/orderer"
	peerImpl "github.com/hyperledger/fabric-sdk-go/pkg/fab/peer"
	clientImpl "github.com/hyperledger/fabric-sdk-go/pkg/fab/resource"
	"github.com/hyperledger/fabric-sdk-go/pkg/fab/resource/api"
	"github.com/pkg/errors"
)

// FabricProvider represents the default implementation of Fabric objects.
type FabricProvider struct {
	providerContext context.ProviderContext
}

type fabContext struct {
	context.ProviderContext
	context.IdentityContext
}

// New creates a FabricProvider enabling access to core Fabric objects and functionality.
func New(ctx context.ProviderContext) *FabricProvider {
	f := FabricProvider{
		providerContext: ctx,
	}
	return &f
}

// CreateResourceClient returns a new client initialized for the current instance of the SDK.
func (f *FabricProvider) CreateResourceClient(ic context.IdentityContext) (api.Resource, error) {
	ctx := &fabContext{
		ProviderContext: f.providerContext,
		IdentityContext: ic,
	}
	client := clientImpl.New(ctx)

	return client, nil
}

// CreateChannelClient returns a new client initialized for the current instance of the SDK.
func (f *FabricProvider) CreateChannelClient(ic context.IdentityContext, cfg fab.ChannelCfg) (fab.Channel, error) {
	ctx := &fabContext{
		ProviderContext: f.providerContext,
		IdentityContext: ic,
	}
	channel, err := channelImpl.New(ctx, cfg)
	if err != nil {
		return nil, errors.WithMessage(err, "NewChannel failed")
	}

	return channel, nil
}

// CreateChannelLedger returns a new client initialized for the current instance of the SDK.
func (f *FabricProvider) CreateChannelLedger(ic context.IdentityContext, channelName string) (fab.ChannelLedger, error) {
	ctx := &fabContext{
		ProviderContext: f.providerContext,
		IdentityContext: ic,
	}
	ledger, err := channelImpl.NewLedger(ctx, channelName)
	if err != nil {
		return nil, errors.WithMessage(err, "NewLedger failed")
	}

	return ledger, nil
}

// CreateEventHub initilizes the event hub.
func (f *FabricProvider) CreateEventHub(ic context.IdentityContext, channelID string) (fab.EventHub, error) {
	peerConfig, err := f.providerContext.Config().ChannelPeers(channelID)
	if err != nil {
		return nil, errors.WithMessage(err, "read configuration for channel peers failed")
	}

	var eventSource *core.ChannelPeer
	for _, p := range peerConfig {
		if p.EventSource && p.MspID == ic.MspID() {
			eventSource = &p
			break
		}
	}

	if eventSource == nil {
		return nil, errors.New("unable to find event source for channel")
	}

	// Event source found, create event hub
	eventCtx := events.Context{
		ProviderContext: f.providerContext,
		IdentityContext: ic,
	}
	return events.FromConfig(eventCtx, &eventSource.PeerConfig)
}

// CreateChannelConfig initializes the channel config
func (f *FabricProvider) CreateChannelConfig(ic context.IdentityContext, channelID string) (fab.ChannelConfig, error) {

	ctx := chconfig.Context{
		ProviderContext: f.providerContext,
		IdentityContext: ic,
	}

	return chconfig.New(ctx, channelID)
}

// CreateChannelTransactor initializes the transactor
func (f *FabricProvider) CreateChannelTransactor(ic context.IdentityContext, cfg fab.ChannelCfg) (fab.Transactor, error) {

	ctx := chconfig.Context{
		ProviderContext: f.providerContext,
		IdentityContext: ic,
	}

	return channelImpl.NewTransactor(ctx, cfg)
}

// CreateCAClient returns a new FabricCAClient initialized for the current instance of the SDK.
func (f *FabricProvider) CreateCAClient(orgID string) (fab.FabricCAClient, error) {
	return fabricCAClient.New(orgID, f.providerContext.Config(), f.providerContext.CryptoSuite())
}

// CreateUser returns a new default implementation of a User.
func (f *FabricProvider) CreateUser(name string, signingIdentity *contextApi.SigningIdentity) (contextApi.User, error) {

	user := identityImpl.NewUser(signingIdentity.MspID, name)

	user.SetPrivateKey(signingIdentity.PrivateKey)
	user.SetEnrollmentCertificate(signingIdentity.EnrollmentCert)

	return user, nil
}

// CreatePeerFromConfig returns a new default implementation of Peer based configuration
func (f *FabricProvider) CreatePeerFromConfig(peerCfg *core.NetworkPeer) (fab.Peer, error) {
	return peerImpl.New(f.providerContext.Config(), peerImpl.FromPeerConfig(peerCfg))
}

// CreateOrdererFromConfig creates a default implementation of Orderer based on configuration.
func (f *FabricProvider) CreateOrdererFromConfig(cfg *core.OrdererConfig) (fab.Orderer, error) {
	orderer, err := orderer.New(f.providerContext.Config(), orderer.FromOrdererConfig(cfg))
	if err != nil {
		return nil, errors.WithMessage(err, "creating orderer failed")
	}
	return orderer, nil
}
