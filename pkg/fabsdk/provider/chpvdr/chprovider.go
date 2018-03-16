/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package chpvdr

import (
	"sync"

	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/context"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/core"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/fab"
	contextImpl "github.com/hyperledger/fabric-sdk-go/pkg/context"
	"github.com/hyperledger/fabric-sdk-go/pkg/fab/chconfig"
)

// ChannelProvider keeps context across ChannelService instances.
//
// TODO: add cache for dynamic channel configuration. This cache is updated
// by channel services, as only channel service have an identity context.
// TODO: add listener for channel config changes. Upon channel config change,
// underlying channel services need to recreate their channel clients.
type ChannelProvider struct {
	infraProvider fab.InfraProvider
	chCfgMap      sync.Map
}

// New creates a ChannelProvider based on a context
func New(infraProvider fab.InfraProvider) (*ChannelProvider, error) {
	cp := ChannelProvider{infraProvider: infraProvider}
	return &cp, nil
}

// ChannelService creates a ChannelService for an identity
func (cp *ChannelProvider) ChannelService(ctx fab.ClientContext, channelID string) (fab.ChannelService, error) {

	var cfg fab.ChannelCfg
	if channelID != "" {
		v, ok := cp.chCfgMap.Load(channelID)
		if !ok {
			p, err := cp.infraProvider.CreateChannelConfig(channelID)
			if err != nil {
				return nil, err
			}

			reqCtx, cancel := contextImpl.NewRequest(ctx, contextImpl.WithTimeoutType(core.PeerResponse))
			defer cancel()

			cfg, err = p.Query(reqCtx)
			if err != nil {
				return nil, err
			}

			cp.chCfgMap.Store(channelID, cfg)
		} else {
			cfg = v.(fab.ChannelCfg)
		}
	} else {
		// System channel
		cfg = chconfig.NewChannelCfg("")
	}

	cs := ChannelService{
		provider:      cp,
		infraProvider: cp.infraProvider,
		context:       ctx,
		cfg:           cfg,
	}

	return &cs, nil
}

// ChannelService provides Channel clients and maintains contexts for them.
// the identity context is used
//
// TODO: add cache for channel rather than reconstructing each time.
type ChannelService struct {
	provider      *ChannelProvider
	infraProvider fab.InfraProvider
	context       context.Client
	cfg           fab.ChannelCfg
}

// EventService returns the EventService.
func (cs *ChannelService) EventService() (fab.EventService, error) {
	return cs.infraProvider.CreateEventService(cs.context, cs.cfg)
}

// Config returns the Config for the named channel
func (cs *ChannelService) Config() (fab.ChannelConfig, error) {
	return cs.infraProvider.CreateChannelConfig(cs.cfg.ID())
}

// Membership returns the member identifier for this channel
func (cs *ChannelService) Membership() (fab.ChannelMembership, error) {
	return cs.infraProvider.CreateChannelMembership(cs.cfg)
}

// ChannelConfig returns the channel config for this channel
func (cs *ChannelService) ChannelConfig() fab.ChannelCfg {
	return cs.cfg
}
