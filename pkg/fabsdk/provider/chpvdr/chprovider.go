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
	"github.com/hyperledger/fabric-sdk-go/pkg/fab/chconfig"
	"github.com/hyperledger/fabric-sdk-go/pkg/util/concurrent/lazycache"
)

var logger = logging.NewLogger("fabsdk")

// ChannelProvider keeps context across ChannelService instances.
//
// TODO: add listener for channel config changes. Upon channel config change,
// underlying channel services need to recreate their channel clients.
type ChannelProvider struct {
	providerContext context.Providers
	ctxtCaches      *lazycache.Cache
}

// New creates a ChannelProvider based on a context
func New(config fab.EndpointConfig, opts ...options.Opt) (*ChannelProvider, error) {
	return &ChannelProvider{
		ctxtCaches: lazycache.New(
			"Client_Context_Cache",
			func(key lazycache.Key) (interface{}, error) {
				ck := key.(*ctxtCacheKey)
				return newContextCache(ck.context, opts), nil
			},
		),
	}, nil
}

// Initialize sets the provider context
func (cp *ChannelProvider) Initialize(providers context.Providers) error {
	cp.providerContext = providers
	return nil
}

// Close frees resources and caches.
func (cp *ChannelProvider) Close() {
	cp.ctxtCaches.Close()
}

// CloseContext frees resources and caches for the given context.
func (cp *ChannelProvider) CloseContext(ctx fab.ClientContext) {
	key, err := newCtxtCacheKey(ctx)
	if err != nil {
		logger.Warnf("Unable to close context: %s", err)
		return
	}
	logger.Warnf("Deleting context cache...")
	cp.ctxtCaches.Delete(key)
}

// ChannelService creates a ChannelService for an identity
func (cp *ChannelProvider) ChannelService(ctx fab.ClientContext, channelID string) (fab.ChannelService, error) {
	key, err := newCtxtCacheKey(ctx)
	if err != nil {
		return nil, err
	}

	ctxtCache, err := cp.ctxtCaches.Get(key)
	if err != nil {
		// This should never happen since the creation of a cache never returns an error
		return nil, err
	}

	return &ChannelService{
		provider:  cp,
		context:   ctx,
		channelID: channelID,
		ctxtCache: ctxtCache.(*contextCache),
	}, nil
}

// ChannelService provides Channel clients and maintains contexts for them.
// the identity context is used
type ChannelService struct {
	provider  *ChannelProvider
	context   context.Client
	channelID string
	ctxtCache *contextCache
}

// Config returns the Config for the named channel
func (cs *ChannelService) Config() (fab.ChannelConfig, error) {
	return chconfig.New(cs.channelID)
}

// EventService returns the EventService.
func (cs *ChannelService) EventService(opts ...options.Opt) (fab.EventService, error) {
	return cs.ctxtCache.GetEventService(cs.channelID, opts...)
}

// Membership returns and caches a channel member identifier
// A membership reference is returned that refreshes with the configured interval
func (cs *ChannelService) Membership() (fab.ChannelMembership, error) {
	return cs.ctxtCache.GetMembership(cs.channelID)
}

// ChannelConfig returns the channel config for this channel
func (cs *ChannelService) ChannelConfig() (fab.ChannelCfg, error) {
	return cs.ctxtCache.GetChannelConfig(cs.channelID)
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
	return cs.ctxtCache.GetDiscoveryService(cs.channelID)
}

// Selection returns a SelectionService for the given channel
func (cs *ChannelService) Selection() (fab.SelectionService, error) {
	return cs.ctxtCache.GetSelectionService(cs.channelID)
}
