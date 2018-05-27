/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package mocks

import (
	"github.com/pkg/errors"

	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/context"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/fab"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/msp"
	mspmocks "github.com/hyperledger/fabric-sdk-go/pkg/msp/test/mockmsp"
)

//Channel supplies the configuration for channel context client
type Channel struct {
	context.Client
	channelService fab.ChannelService
	channelID      string
}

//Providers returns core providers
func (c *Channel) Providers() context.Client {
	return c
}

//ChannelService returns channel service
func (c *Channel) ChannelService() fab.ChannelService {
	return c.channelService
}

//ChannelID returns channel ID
func (c *Channel) ChannelID() string {
	return c.channelID
}

type mockClientContext struct {
	context.Providers
	msp.SigningIdentity
}

//NewMockChannel creates new mock channel
func NewMockChannel(channelID string) (*Channel, error) {

	ctx := &mockClientContext{
		Providers:       NewMockProviderContext(),
		SigningIdentity: mspmocks.NewMockSigningIdentity("user", "Org1MSP"),
	}

	// Set up mock channel service
	chProvider, err := NewMockChannelProvider(ctx)
	if err != nil {
		return nil, errors.WithMessage(err, "new mock channel provider failed")
	}
	channelService, err := chProvider.ChannelService(ctx, channelID)
	if err != nil {
		return nil, errors.WithMessage(err, "failed to create mock channel service")
	}

	peer := NewMockPeer("Peer1", "example.com")

	channelService.(*MockChannelService).SetDiscovery(NewMockDiscoveryService(nil, peer))
	channelService.(*MockChannelService).SetSelection(NewMockSelectionService(nil, peer))

	channel := &Channel{
		Client:         ctx,
		channelService: channelService,
		channelID:      channelID,
	}

	return channel, nil
}
