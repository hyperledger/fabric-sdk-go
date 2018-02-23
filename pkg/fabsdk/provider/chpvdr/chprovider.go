/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package chpvdr

import (
	"sync"

	"github.com/hyperledger/fabric-sdk-go/pkg/context"
	"github.com/hyperledger/fabric-sdk-go/pkg/context/api/fab"
	"github.com/hyperledger/fabric-sdk-go/pkg/fab/chconfig"
	"github.com/hyperledger/fabric-sdk-go/pkg/fabsdk/api"
)

// ChannelProvider keeps context across ChannelService instances.
//
// TODO: add cache for dynamic channel configuration. This cache is updated
// by channel services, as only channel service have an identity context.
// TODO: add listener for channel config changes. Upon channel config change,
// underlying channel services need to recreate their channel clients.
type ChannelProvider struct {
	fabricProvider api.FabricProvider
	chCfgMap       sync.Map
}

// New creates a ChannelProvider based on a context
func New(fabricProvider api.FabricProvider) (*ChannelProvider, error) {
	cp := ChannelProvider{fabricProvider: fabricProvider}
	return &cp, nil
}

// NewChannelService creates a ChannelService for an identity
func (cp *ChannelProvider) NewChannelService(ic context.IdentityContext, channelID string) (fab.ChannelService, error) {

	var cfg fab.ChannelCfg
	if channelID != "" {
		v, ok := cp.chCfgMap.Load(channelID)
		if !ok {
			p, err := cp.fabricProvider.CreateChannelConfig(ic, channelID)
			if err != nil {
				return nil, err
			}

			cfg, err = p.Query()
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
		provider:        cp,
		fabricProvider:  cp.fabricProvider,
		identityContext: ic,
		cfg:             cfg,
	}

	return &cs, nil
}

// ChannelService provides Channel clients and maintains contexts for them.
// the identity context is used
//
// TODO: add cache for channel rather than reconstructing each time.
type ChannelService struct {
	provider        *ChannelProvider
	fabricProvider  api.FabricProvider
	identityContext context.IdentityContext
	cfg             fab.ChannelCfg
}

// Channel returns the named Channel client.
func (cs *ChannelService) Channel() (fab.Channel, error) {
	return cs.fabricProvider.CreateChannelClient(cs.identityContext, cs.cfg)
}

// EventHub returns the EventHub for the named channel.
func (cs *ChannelService) EventHub() (fab.EventHub, error) {
	return cs.fabricProvider.CreateEventHub(cs.identityContext, cs.cfg.Name())
}

// Config returns the Config for the named channel
func (cs *ChannelService) Config() (fab.ChannelConfig, error) {
	return cs.fabricProvider.CreateChannelConfig(cs.identityContext, cs.cfg.Name())
}

// Ledger returns a ChannelLedger client for the current context and named channel.
func (cs *ChannelService) Ledger() (fab.ChannelLedger, error) {
	return cs.fabricProvider.CreateChannelLedger(cs.identityContext, cs.cfg.Name())
}

// Transactor returns a transaction client for the current context and named channel.
func (cs *ChannelService) Transactor() (fab.Transactor, error) {
	return cs.fabricProvider.CreateChannelTransactor(cs.identityContext, cs.cfg)
}
