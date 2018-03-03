/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package defclient

import (
	"github.com/hyperledger/fabric-sdk-go/pkg/client/channel"
	"github.com/hyperledger/fabric-sdk-go/pkg/client/common/discovery"
	"github.com/hyperledger/fabric-sdk-go/pkg/context"
	"github.com/hyperledger/fabric-sdk-go/pkg/context/api/fab"
	"github.com/pkg/errors"
)

// SessionClientFactory represents the default implementation of a session client.
type SessionClientFactory struct{}

// NewSessionClientFactory creates a new default session client factory.
func NewSessionClientFactory() *SessionClientFactory {
	f := SessionClientFactory{}
	return &f
}

// CreateChannelClient returns a client that can execute transactions on specified channel
func (f *SessionClientFactory) CreateChannelClient(providers context.Providers, session context.Session, channelID string, targetFilter fab.TargetFilter) (*channel.Client, error) {

	chProvider := providers.ChannelProvider()
	chService, err := chProvider.ChannelService(session, channelID)
	if err != nil {
		return &channel.Client{}, errors.WithMessage(err, "create channel service failed")
	}

	discoveryService, err := providers.DiscoveryProvider().NewDiscoveryService(channelID)
	if err != nil {
		return &channel.Client{}, errors.WithMessage(err, "create discovery service failed")
	}

	discoveryService = discovery.NewDiscoveryFilterService(discoveryService, targetFilter)

	selection, err := providers.SelectionProvider().NewSelectionService(channelID)
	if err != nil {
		return &channel.Client{}, errors.WithMessage(err, "create selection service failed")
	}

	ctx := channel.Context{
		Providers:        providers,
		DiscoveryService: discoveryService,
		SelectionService: selection,
		ChannelService:   chService,
	}
	return channel.New(ctx)
}
