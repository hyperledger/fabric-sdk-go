/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package chpvdr

import (
	reqContext "context"

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
	providerContext   context.Providers
	eventServiceCache cache
	chCfgCache        cache
	membershipCache   cache
}

// New creates a ChannelProvider based on a context
func New(config fab.EndpointConfig) (*ChannelProvider, error) {
	eventIdleTime := config.Timeout(fab.EventServiceIdle)
	chConfigRefresh := config.Timeout(fab.ChannelConfigRefresh)
	membershipRefresh := config.Timeout(fab.ChannelMembershipRefresh)

	eventServiceCache := lazycache.New(
		"Event_Service_Cache",
		func(key lazycache.Key) (interface{}, error) {
			ck := key.(cacheKey)
			return NewEventClientRef(
				eventIdleTime,
				func() (fab.EventClient, error) {
					return getEventClient(ck.Context(), ck.ChannelConfig(), ck.Opts()...)
				},
			), nil
		},
	)

	cp := ChannelProvider{
		eventServiceCache: eventServiceCache,
		chCfgCache:        chconfig.NewRefCache(chConfigRefresh),
		membershipCache:   membership.NewRefCache(membershipRefresh),
	}
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
	key, err := NewCacheKey(cs.context, chnlCfg, opts...)
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
	if cs.channelID == "" {
		// System channel
		return chconfig.NewChannelCfg(""), nil
	}
	chCfgRef, err := cs.loadChannelCfgRef()
	if err != nil {
		return nil, err
	}
	chCfg, err := chCfgRef.Get()
	if err != nil {
		return nil, errors.WithMessage(err, "could not get chConfig cache reference")
	}
	return chCfg.(fab.ChannelCfg), nil
}

// Transactor returns the transactor
func (cs *ChannelService) Transactor(reqCtx reqContext.Context) (fab.Transactor, error) {
	cfg, err := cs.ChannelConfig()
	if err != nil {
		return nil, err
	}
	return channelImpl.NewTransactor(reqCtx, cfg)
}

func (cs *ChannelService) loadChannelCfgRef() (*chconfig.Ref, error) {
	key, err := chconfig.NewCacheKey(cs.context, func(string) (fab.ChannelConfig, error) { return cs.Config() }, cs.channelID)
	if err != nil {
		return nil, err
	}
	c, err := cs.provider.chCfgCache.Get(key)
	if err != nil {
		return nil, err
	}

	return c.(*chconfig.Ref), nil
}

func getEventClient(ctx context.Client, chConfig fab.ChannelCfg, opts ...options.Opt) (fab.EventClient, error) {
	useDeliver, err := useDeliverEvents(ctx, chConfig)
	if err != nil {
		return nil, err
	}

	if useDeliver {
		logger.Debugf("Using deliver events for channel [%s]", chConfig.ID())
		return deliverclient.New(ctx, chConfig, opts...)
	}

	logger.Debugf("Using event hub events for channel [%s]", chConfig.ID())
	return eventhubclient.New(ctx, chConfig, opts...)
}

func useDeliverEvents(ctx context.Client, chConfig fab.ChannelCfg) (bool, error) {
	switch ctx.EndpointConfig().EventServiceType() {
	case fab.DeliverEventServiceType:
		return true, nil
	case fab.EventHubEventServiceType:
		return false, nil
	case fab.AutoDetectEventServiceType:
		logger.Debugf("Determining event service type from channel capabilities...")
		return chConfig.HasCapability(fab.ApplicationGroupKey, fab.V1_1Capability), nil
	default:
		return false, errors.Errorf("unsupported event service type: %d", ctx.EndpointConfig().EventServiceType())
	}
}

type cacheKey interface {
	lazycache.Key
	Context() fab.ClientContext
	ChannelConfig() fab.ChannelCfg
	Opts() []options.Opt
}
