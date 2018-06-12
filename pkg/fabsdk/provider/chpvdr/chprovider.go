/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package chpvdr

import (
	reqContext "context"

	"github.com/hyperledger/fabric-sdk-go/pkg/client/common/discovery/dynamicdiscovery"
	"github.com/hyperledger/fabric-sdk-go/pkg/client/common/discovery/staticdiscovery"
	"github.com/hyperledger/fabric-sdk-go/pkg/client/common/selection/dynamicselection"
	"github.com/hyperledger/fabric-sdk-go/pkg/client/common/selection/fabricselection"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/logging"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/options"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/context"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/fab"
	channelImpl "github.com/hyperledger/fabric-sdk-go/pkg/fab/channel"
	"github.com/hyperledger/fabric-sdk-go/pkg/fab/channel/membership"
	"github.com/hyperledger/fabric-sdk-go/pkg/fab/chconfig"
	"github.com/hyperledger/fabric-sdk-go/pkg/fab/events/deliverclient"
	"github.com/hyperledger/fabric-sdk-go/pkg/fab/events/eventhubclient"
	"github.com/hyperledger/fabric-sdk-go/pkg/util/concurrent/lazycache"
	"github.com/pkg/errors"
)

var logger = logging.NewLogger("fabsdk")

type cache interface {
	Get(lazycache.Key) (interface{}, error)
	Close()
}

// ChannelProvider keeps context across ChannelService instances.
//
// TODO: add listener for channel config changes. Upon channel config change,
// underlying channel services need to recreate their channel clients.
type ChannelProvider struct {
	providerContext       context.Providers
	eventServiceCache     cache
	discoveryServiceCache cache
	selectionServiceCache cache
	chCfgCache            cache
	membershipCache       cache
}

// New creates a ChannelProvider based on a context
func New(config fab.EndpointConfig) (*ChannelProvider, error) {
	eventIdleTime := config.Timeout(fab.EventServiceIdle)
	chConfigRefresh := config.Timeout(fab.ChannelConfigRefresh)
	membershipRefresh := config.Timeout(fab.ChannelMembershipRefresh)

	cp := ChannelProvider{
		chCfgCache:      chconfig.NewRefCache(chConfigRefresh),
		membershipCache: membership.NewRefCache(membershipRefresh),
	}

	cp.discoveryServiceCache = lazycache.New(
		"Discovery_Service_Cache",
		func(key lazycache.Key) (interface{}, error) {
			ck := key.(*cacheKey)
			return cp.createDiscoveryService(ck.context, ck.channelConfig)
		},
	)

	cp.selectionServiceCache = lazycache.New(
		"Selection_Service_Cache",
		func(key lazycache.Key) (interface{}, error) {
			ck := key.(*cacheKey)
			return cp.createSelectionService(ck.context, ck.channelConfig)
		},
	)

	cp.eventServiceCache = lazycache.New(
		"Event_Service_Cache",
		func(key lazycache.Key) (interface{}, error) {
			ck := key.(*eventCacheKey)
			return NewEventClientRef(
				eventIdleTime,
				func() (fab.EventClient, error) {
					return cp.createEventClient(ck.context, ck.channelConfig, ck.opts...)
				},
			), nil
		},
	)

	return &cp, nil
}

// Initialize sets the provider context
func (cp *ChannelProvider) Initialize(providers context.Providers) error {
	cp.providerContext = providers
	return nil
}

// Close frees resources and caches.
func (cp *ChannelProvider) Close() {
	logger.Debug("Closing event service cache...")
	cp.eventServiceCache.Close()

	logger.Debug("Closing membership cache...")
	cp.membershipCache.Close()

	logger.Debug("Closing channel configuration cache...")
	cp.chCfgCache.Close()

	logger.Debug("Closing selection service cache...")
	cp.selectionServiceCache.Close()

	logger.Debug("Closing discovery service cache...")
	cp.discoveryServiceCache.Close()
}

// ChannelService creates a ChannelService for an identity
func (cp *ChannelProvider) ChannelService(ctx fab.ClientContext, channelID string) (fab.ChannelService, error) {
	cs := ChannelService{
		provider:  cp,
		context:   ctx,
		channelID: channelID,
	}

	return &cs, nil
}

func (cp *ChannelProvider) createEventClient(ctx context.Client, chConfig fab.ChannelCfg, opts ...options.Opt) (fab.EventClient, error) {
	useDeliver, err := useDeliverEvents(ctx, chConfig)
	if err != nil {
		return nil, err
	}

	discovery, err := cp.getDiscoveryService(ctx, chConfig.ID())
	if err != nil {
		return nil, errors.WithMessage(err, "could not get discovery service")
	}

	if useDeliver {
		logger.Debugf("Using deliver events for channel [%s]", chConfig.ID())
		return deliverclient.New(ctx, chConfig, discovery, opts...)
	}

	logger.Debugf("Using event hub events for channel [%s]", chConfig.ID())
	return eventhubclient.New(ctx, chConfig, discovery, opts...)
}

func (cp *ChannelProvider) createDiscoveryService(ctx context.Client, chConfig fab.ChannelCfg) (fab.DiscoveryService, error) {
	if chConfig.HasCapability(fab.ApplicationGroupKey, fab.V1_2Capability) {
		logger.Debugf("Using Dynamic Discovery based on V1_2 capability.")
		return dynamicdiscovery.NewChannelService(ctx, chConfig.ID())
	}
	return staticdiscovery.NewService(ctx.EndpointConfig(), ctx.InfraProvider(), chConfig.ID())
}

func (cp *ChannelProvider) getDiscoveryService(context fab.ClientContext, channelID string) (fab.DiscoveryService, error) {
	chnlCfg, err := cp.channelConfig(context, channelID)
	if err != nil {
		return nil, err
	}
	key, err := newCacheKey(context, chnlCfg)
	if err != nil {
		return nil, err
	}
	discoveryService, err := cp.discoveryServiceCache.Get(key)
	if err != nil {
		return nil, err
	}
	return discoveryService.(fab.DiscoveryService), nil
}

func (cp *ChannelProvider) createSelectionService(ctx context.Client, chConfig fab.ChannelCfg) (fab.SelectionService, error) {
	discovery, err := cp.getDiscoveryService(ctx, chConfig.ID())
	if err != nil {
		return nil, err
	}

	if chConfig.HasCapability(fab.ApplicationGroupKey, fab.V1_2Capability) {
		logger.Debugf("Using Fabric Selection based on V1_2 capability.")
		return fabricselection.New(ctx, chConfig.ID(), discovery)
	}
	return dynamicselection.NewService(ctx, chConfig.ID(), discovery)
}

func (cp *ChannelProvider) getSelectionService(context fab.ClientContext, channelID string) (fab.SelectionService, error) {
	chnlCfg, err := cp.channelConfig(context, channelID)
	if err != nil {
		return nil, err
	}
	key, err := newCacheKey(context, chnlCfg)
	if err != nil {
		return nil, err
	}
	selectionService, err := cp.selectionServiceCache.Get(key)
	if err != nil {
		return nil, err
	}
	return selectionService.(fab.SelectionService), nil
}

func (cp *ChannelProvider) channelConfig(context fab.ClientContext, channelID string) (fab.ChannelCfg, error) {
	if channelID == "" {
		// System channel
		return chconfig.NewChannelCfg(""), nil
	}
	chCfgRef, err := cp.loadChannelCfgRef(context, channelID)
	if err != nil {
		return nil, err
	}
	chCfg, err := chCfgRef.Get()
	if err != nil {
		return nil, errors.WithMessage(err, "could not get chConfig cache reference")
	}
	return chCfg.(fab.ChannelCfg), nil
}

func (cp *ChannelProvider) loadChannelCfgRef(context fab.ClientContext, channelID string) (*chconfig.Ref, error) {
	key, err := chconfig.NewCacheKey(context, func(string) (fab.ChannelConfig, error) { return chconfig.New(channelID) }, channelID)
	if err != nil {
		return nil, err
	}
	c, err := cp.chCfgCache.Get(key)
	if err != nil {
		return nil, err
	}

	return c.(*chconfig.Ref), nil
}

// ChannelService provides Channel clients and maintains contexts for them.
// the identity context is used
type ChannelService struct {
	provider  *ChannelProvider
	context   context.Client
	channelID string
}

// Config returns the Config for the named channel
func (cs *ChannelService) Config() (fab.ChannelConfig, error) {
	return chconfig.New(cs.channelID)
}

// EventService returns the EventService.
func (cs *ChannelService) EventService(opts ...options.Opt) (fab.EventService, error) {
	chnlCfg, err := cs.ChannelConfig()
	if err != nil {
		return nil, err
	}
	key, err := newEventCacheKey(cs.context, chnlCfg, opts...)
	if err != nil {
		return nil, err
	}
	eventService, err := cs.provider.eventServiceCache.Get(key)
	if err != nil {
		return nil, err
	}

	return eventService.(fab.EventService), nil
}

// Membership returns and caches a channel member identifier
// A membership reference is returned that refreshes with the configured interval
func (cs *ChannelService) Membership() (fab.ChannelMembership, error) {
	chCfgRef, err := cs.loadChannelCfgRef()
	if err != nil {
		return nil, err
	}
	key, err := membership.NewCacheKey(membership.Context{Providers: cs.provider.providerContext, EndpointConfig: cs.context.EndpointConfig()},
		chCfgRef.Reference, cs.channelID)
	if err != nil {
		return nil, err
	}
	ref, err := cs.provider.membershipCache.Get(key)
	if err != nil {
		return nil, err
	}

	return ref.(*membership.Ref), nil
}

// ChannelConfig returns the channel config for this channel
func (cs *ChannelService) ChannelConfig() (fab.ChannelCfg, error) {
	return cs.provider.channelConfig(cs.context, cs.channelID)
}

// Transactor returns the transactor
func (cs *ChannelService) Transactor(reqCtx reqContext.Context) (fab.Transactor, error) {
	cfg, err := cs.ChannelConfig()
	if err != nil {
		return nil, err
	}
	return channelImpl.NewTransactor(reqCtx, cfg)
}

// Discovery returns a DiscoveryService for the given channel
func (cs *ChannelService) Discovery() (fab.DiscoveryService, error) {
	return cs.provider.getDiscoveryService(cs.context, cs.channelID)
}

// Selection returns a SelectionService for the given channel
func (cs *ChannelService) Selection() (fab.SelectionService, error) {
	return cs.provider.getSelectionService(cs.context, cs.channelID)
}

func (cs *ChannelService) loadChannelCfgRef() (*chconfig.Ref, error) {
	return cs.provider.loadChannelCfgRef(cs.context, cs.channelID)
}

func useDeliverEvents(ctx context.Client, chConfig fab.ChannelCfg) (bool, error) {
	switch ctx.EndpointConfig().EventServiceType() {
	case fab.DeliverEventServiceType:
		return true, nil
	case fab.EventHubEventServiceType:
		return false, nil
	case fab.AutoDetectEventServiceType:
		logger.Debug("Determining event service type from channel capabilities...")
		return chConfig.HasCapability(fab.ApplicationGroupKey, fab.V1_1Capability), nil
	default:
		return false, errors.Errorf("unsupported event service type: %d", ctx.EndpointConfig().EventServiceType())
	}
}
