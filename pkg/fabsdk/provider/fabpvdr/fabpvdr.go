/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package fabpvdr

import (
	"github.com/hyperledger/fabric-sdk-go/pkg/common/context"
	"github.com/hyperledger/fabric-sdk-go/pkg/context/api/core"
	"github.com/hyperledger/fabric-sdk-go/pkg/context/api/fab"
	channelImpl "github.com/hyperledger/fabric-sdk-go/pkg/fab/channel"
	"github.com/hyperledger/fabric-sdk-go/pkg/fab/channel/membership"
	"github.com/hyperledger/fabric-sdk-go/pkg/fab/chconfig"
	"github.com/hyperledger/fabric-sdk-go/pkg/fab/comm"
	"github.com/hyperledger/fabric-sdk-go/pkg/fab/events"
	"github.com/hyperledger/fabric-sdk-go/pkg/fab/orderer"
	peerImpl "github.com/hyperledger/fabric-sdk-go/pkg/fab/peer"
	clientImpl "github.com/hyperledger/fabric-sdk-go/pkg/fab/resource"
	"github.com/hyperledger/fabric-sdk-go/pkg/fab/resource/api"
	"github.com/pkg/errors"
)

// FabricProvider represents the default implementation of Fabric objects.
type FabricProvider struct {
	providerContext context.Providers
	connector       *comm.CachingConnector
}

type fabContext struct {
	context.Providers
	context.Identity
}

// New creates a FabricProvider enabling access to core Fabric objects and functionality.
func New(ctx context.Providers) *FabricProvider {
	idleTime := ctx.Config().TimeoutOrDefault(core.ConnectionIdle)
	sweepTime := ctx.Config().TimeoutOrDefault(core.CacheSweepInterval)

	cc := comm.NewCachingConnector(sweepTime, idleTime)

	f := FabricProvider{
		providerContext: ctx,
		connector:       cc,
	}
	return &f
}

// Close frees resources and caches.
func (f *FabricProvider) Close() {
	f.connector.Close()
}

// CreateResourceClient returns a new client initialized for the current instance of the SDK.
func (f *FabricProvider) CreateResourceClient(ic fab.IdentityContext) (api.Resource, error) {
	ctx := &fabContext{
		Providers: f.providerContext,
		Identity:  ic,
	}
	client := clientImpl.New(ctx)

	return client, nil
}

// CreateChannelLedger returns a new client initialized for the current instance of the SDK.
func (f *FabricProvider) CreateChannelLedger(ic fab.IdentityContext, channelName string) (fab.ChannelLedger, error) {
	ctx := &fabContext{
		Providers: f.providerContext,
		Identity:  ic,
	}
	ledger, err := channelImpl.NewLedger(ctx, channelName)
	if err != nil {
		return nil, errors.WithMessage(err, "NewLedger failed")
	}

	return ledger, nil
}

// CreateEventHub initilizes the event hub.
func (f *FabricProvider) CreateEventHub(ic fab.IdentityContext, channelID string) (fab.EventHub, error) {
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
		Providers: f.providerContext,
		Identity:  ic,
	}
	return events.FromConfig(eventCtx, &eventSource.PeerConfig)
}

// CreateChannelConfig initializes the channel config
func (f *FabricProvider) CreateChannelConfig(ic fab.IdentityContext, channelID string) (fab.ChannelConfig, error) {

	ctx := chconfig.Context{
		Providers: f.providerContext,
		Identity:  ic,
	}

	return chconfig.New(ctx, channelID)
}

// CreateChannelMembership returns a channel member identifier
func (f *FabricProvider) CreateChannelMembership(cfg fab.ChannelCfg) (fab.ChannelMembership, error) {
	return membership.New(membership.Context{Providers: f.providerContext}, cfg)
}

// CreateChannelTransactor initializes the transactor
func (f *FabricProvider) CreateChannelTransactor(ic fab.IdentityContext, cfg fab.ChannelCfg) (fab.Transactor, error) {

	ctx := chconfig.Context{
		Providers: f.providerContext,
		Identity:  ic,
	}

	return channelImpl.NewTransactor(ctx, cfg)
}

// CreatePeerFromConfig returns a new default implementation of Peer based configuration
func (f *FabricProvider) CreatePeerFromConfig(peerCfg *core.NetworkPeer) (fab.Peer, error) {
	return peerImpl.New(f.providerContext.Config(), peerImpl.FromPeerConfig(peerCfg), peerImpl.WithConnProvider(f.connector))
}

// CreateOrdererFromConfig creates a default implementation of Orderer based on configuration.
func (f *FabricProvider) CreateOrdererFromConfig(cfg *core.OrdererConfig) (fab.Orderer, error) {
	newOrderer, err := orderer.New(f.providerContext.Config(), orderer.FromOrdererConfig(cfg))
	if err != nil {
		return nil, errors.WithMessage(err, "creating orderer failed")
	}
	return newOrderer, nil
}
