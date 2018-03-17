/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package fabpvdr

import (
	reqContext "context"

	"github.com/hyperledger/fabric-sdk-go/pkg/common/logging"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/options"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/context"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/core"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/fab"
	channelImpl "github.com/hyperledger/fabric-sdk-go/pkg/fab/channel"
	"github.com/hyperledger/fabric-sdk-go/pkg/fab/channel/membership"
	"github.com/hyperledger/fabric-sdk-go/pkg/fab/chconfig"
	"github.com/hyperledger/fabric-sdk-go/pkg/fab/comm"
	"github.com/hyperledger/fabric-sdk-go/pkg/fab/events/deliverclient"
	"github.com/hyperledger/fabric-sdk-go/pkg/fab/events/eventhubclient"
	"github.com/hyperledger/fabric-sdk-go/pkg/fab/orderer"
	peerImpl "github.com/hyperledger/fabric-sdk-go/pkg/fab/peer"
	"github.com/hyperledger/fabric-sdk-go/pkg/util/concurrent/lazycache"
	"github.com/pkg/errors"
)

var logger = logging.NewLogger("fabsdk")

type cacheKey interface {
	lazycache.Key
	Context() fab.ClientContext
	ChannelConfig() fab.ChannelCfg
}

type cache interface {
	Get(lazycache.Key) (interface{}, error)
	Close()
}

// InfraProvider represents the default implementation of Fabric objects.
type InfraProvider struct {
	providerContext   context.Providers
	commManager       *comm.CachingConnector
	eventServiceCache cache
	chCfgCache        cache
	membershipCache   cache
}

// New creates a InfraProvider enabling access to core Fabric objects and functionality.
func New(config core.Config, opts ...options.Opt) *InfraProvider {
	idleTime := config.TimeoutOrDefault(core.ConnectionIdle)
	sweepTime := config.TimeoutOrDefault(core.CacheSweepInterval)
	eventIdleTime := config.TimeoutOrDefault(core.EventServiceIdle)
	chConfigRefresh := config.TimeoutOrDefault(core.ChannelConfigRefresh)
	membershipRefresh := config.TimeoutOrDefault(core.ChannelMembershipRefresh)

	eventServiceCache := lazycache.New(
		"Event_Service_Cache",
		func(key lazycache.Key) (interface{}, error) {
			ck := key.(cacheKey)
			return NewEventClientRef(
				eventIdleTime,
				func() (fab.EventClient, error) {
					return getEventClient(ck.Context(), ck.ChannelConfig(), opts...)
				},
			), nil
		},
	)

	return &InfraProvider{
		commManager:       comm.NewCachingConnector(sweepTime, idleTime),
		eventServiceCache: eventServiceCache,
		chCfgCache:        chconfig.NewRefCache(chConfigRefresh),
		membershipCache:   membership.NewRefCache(membershipRefresh),
	}
}

// Initialize sets the provider context
func (f *InfraProvider) Initialize(providers context.Providers) error {
	f.providerContext = providers
	return nil
}

// Close frees resources and caches.
func (f *InfraProvider) Close() {
	logger.Debug("Closing event service cache...")
	f.eventServiceCache.Close()

	logger.Debug("Closing membership cache...")
	f.membershipCache.Close()

	logger.Debug("Closing channel configuration cache...")
	f.chCfgCache.Close()

	// Comm Manager must be closed last since other resources
	// may still be using it.
	logger.Debug("Closing comm manager...")
	f.commManager.Close()
}

// CommManager provides comm support such as GRPC onnections
func (f *InfraProvider) CommManager() fab.CommManager {
	return f.commManager
}

// CreateEventService creates the event service.
func (f *InfraProvider) CreateEventService(ctx fab.ClientContext, channelID string) (fab.EventService, error) {
	chnlCfg, err := f.CreateChannelCfg(ctx, channelID)
	if err != nil {
		return nil, err
	}
	key, err := NewCacheKey(ctx, chnlCfg)
	if err != nil {
		return nil, err
	}
	eventService, err := f.eventServiceCache.Get(key)
	if err != nil {
		return nil, err
	}
	return eventService.(fab.EventService), nil
}

// CreateChannelConfig initializes the channel config
func (f *InfraProvider) CreateChannelConfig(channelID string) (fab.ChannelConfig, error) {
	return chconfig.New(channelID)
}

// CreateChannelCfg creates and caches the channel configuration
func (f *InfraProvider) CreateChannelCfg(ctx fab.ClientContext, channelID string) (fab.ChannelCfg, error) {
	if channelID == "" {
		// System channel
		return chconfig.NewChannelCfg(""), nil
	}
	chCfgRef, err := f.loadChannelCfgRef(ctx, channelID)
	if err != nil {
		return nil, err
	}
	chCfg, err := chCfgRef.Get()
	if err != nil {
		return nil, errors.WithMessage(err, "could not get chConfig cache reference")
	}
	return chCfg.(fab.ChannelCfg), nil
}

// CreateChannelMembership returns and caches a channel member identifier
// A membership reference is returned that refreshes with the configured interval
func (f *InfraProvider) CreateChannelMembership(ctx fab.ClientContext, channelID string) (fab.ChannelMembership, error) {
	chCfgRef, err := f.loadChannelCfgRef(ctx, channelID)
	if err != nil {
		return nil, err
	}
	key, err := membership.NewCacheKey(membership.Context{Providers: f.providerContext},
		chCfgRef.Reference, channelID)
	if err != nil {
		return nil, err
	}
	ref, err := f.membershipCache.Get(key)
	if err != nil {
		return nil, err
	}

	return ref.(*membership.Ref), nil
}

// CreateChannelTransactor initializes the transactor
func (f *InfraProvider) CreateChannelTransactor(reqCtx reqContext.Context, cfg fab.ChannelCfg) (fab.Transactor, error) {
	return channelImpl.NewTransactor(reqCtx, cfg)
}

// CreatePeerFromConfig returns a new default implementation of Peer based configuration
func (f *InfraProvider) CreatePeerFromConfig(peerCfg *core.NetworkPeer) (fab.Peer, error) {
	return peerImpl.New(f.providerContext.Config(), peerImpl.FromPeerConfig(peerCfg))
}

// CreateOrdererFromConfig creates a default implementation of Orderer based on configuration.
func (f *InfraProvider) CreateOrdererFromConfig(cfg *core.OrdererConfig) (fab.Orderer, error) {
	newOrderer, err := orderer.New(f.providerContext.Config(), orderer.FromOrdererConfig(cfg))
	if err != nil {
		return nil, errors.WithMessage(err, "creating orderer failed")
	}
	return newOrderer, nil
}

func (f *InfraProvider) loadChannelCfgRef(ctx fab.ClientContext, channelID string) (*chconfig.Ref, error) {
	key, err := chconfig.NewCacheKey(ctx, f.CreateChannelConfig, channelID)
	if err != nil {
		return nil, err
	}
	c, err := f.chCfgCache.Get(key)
	if err != nil {
		return nil, err
	}

	return c.(*chconfig.Ref), nil
}

func getEventClient(ctx context.Client, chConfig fab.ChannelCfg, opts ...options.Opt) (fab.EventClient, error) {
	// TODO: This logic should be based on the channel capabilities. For now,
	// look at the EventServiceType specified in the config file.
	switch ctx.Config().EventServiceType() {
	case core.DeliverEventServiceType:
		return deliverclient.New(ctx, chConfig, opts...)
	case core.EventHubEventServiceType:
		logger.Debugf("Using event hub events")
		return eventhubclient.New(ctx, chConfig, opts...)
	default:
		return nil, errors.Errorf("unsupported event service type: %d", ctx.Config().EventServiceType())
	}
}
