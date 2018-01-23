/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package chpvdr

import (
	"github.com/hyperledger/fabric-sdk-go/api/apicore"
	"github.com/hyperledger/fabric-sdk-go/api/apifabclient"
)

// ChannelProvider keeps context across ChannelService instances.
//
// TODO: add cache for dynamic channel configuration. This cache is updated
// by channel services, as only channel service have an identity context.
// TODO: add listener for channel config changes. Upon channel config change,
// underlying channel services need to recreate their channel clients.
type ChannelProvider struct {
	fabricProvider apicore.FabricProvider
}

// New creates a ChannelProvider based on a context
func New(fabricProvider apicore.FabricProvider) (*ChannelProvider, error) {
	cp := ChannelProvider{fabricProvider}
	return &cp, nil
}

// NewChannelService creates a ChannelService for an identity
func (cp *ChannelProvider) NewChannelService(ic apifabclient.IdentityContext, channelID string) (apifabclient.ChannelService, error) {
	cs := ChannelService{
		fabricProvider:  cp.fabricProvider,
		identityContext: ic,
		channelID:       channelID,
	}
	return &cs, nil
}

// ChannelService provides Channel clients and maintains contexts for them.
// the identity context is used
//
// TODO: add cache for channel rather than reconstructing each time.
type ChannelService struct {
	fabricProvider  apicore.FabricProvider
	identityContext apifabclient.IdentityContext
	channelID       string
}

// Channel returns the named Channel client.
func (cs *ChannelService) Channel() (apifabclient.Channel, error) {
	return cs.fabricProvider.NewChannelClient(cs.identityContext, cs.channelID)
}

// EventHub returns the EventHub for the named channel.
func (cs *ChannelService) EventHub() (apifabclient.EventHub, error) {
	return cs.fabricProvider.NewEventHub(cs.identityContext, cs.channelID)
}
