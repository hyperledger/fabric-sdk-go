/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package chpvdr

import (
	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/context"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/fab"
)

// ChannelProvider keeps context across ChannelService instances.
//
// TODO: add listener for channel config changes. Upon channel config change,
// underlying channel services need to recreate their channel clients.
type ChannelProvider struct {
	infraProvider fab.InfraProvider
}

// New creates a ChannelProvider based on a context
func New(infraProvider fab.InfraProvider) (*ChannelProvider, error) {
	cp := ChannelProvider{infraProvider: infraProvider}
	return &cp, nil
}

// ChannelService creates a ChannelService for an identity
func (cp *ChannelProvider) ChannelService(ctx fab.ClientContext, channelID string) (fab.ChannelService, error) {
	cs := ChannelService{
		provider:      cp,
		infraProvider: cp.infraProvider,
		context:       ctx,
		channelID:     channelID,
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
	channelID     string
}

// Config returns the Config for the named channel
func (cs *ChannelService) Config() (fab.ChannelConfig, error) {
	return cs.infraProvider.CreateChannelConfig(cs.channelID)
}

// EventService returns the EventService.
func (cs *ChannelService) EventService() (fab.EventService, error) {
	return cs.infraProvider.CreateEventService(cs.context, cs.channelID)
}

// Membership returns the member identifier for this channel
func (cs *ChannelService) Membership() (fab.ChannelMembership, error) {
	return cs.infraProvider.CreateChannelMembership(cs.context, cs.channelID)
}

// ChannelConfig returns the channel config for this channel
func (cs *ChannelService) ChannelConfig() (fab.ChannelCfg, error) {
	return cs.infraProvider.CreateChannelCfg(cs.context, cs.channelID)
}
