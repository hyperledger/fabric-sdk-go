/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package defclient

import (
	apichclient "github.com/hyperledger/fabric-sdk-go/api/apitxn/chclient"
	apisdk "github.com/hyperledger/fabric-sdk-go/pkg/fabsdk/api"

	"github.com/hyperledger/fabric-sdk-go/api/apifabclient"
	"github.com/hyperledger/fabric-sdk-go/pkg/fabric-txn/chclient"
	"github.com/hyperledger/fabric-sdk-go/pkg/fabric-txn/discovery"
	"github.com/pkg/errors"
)

// SessionClientFactory represents the default implementation of a session client.
type SessionClientFactory struct{}

// NewSessionClientFactory creates a new default session client factory.
func NewSessionClientFactory() *SessionClientFactory {
	f := SessionClientFactory{}
	return &f
}

// NewChannelClient returns a client that can execute transactions on specified channel
func (f *SessionClientFactory) NewChannelClient(providers apisdk.Providers, session apisdk.SessionContext, channelID string, targetFilter apifabclient.TargetFilter) (apichclient.ChannelClient, error) {

	chProvider := providers.ChannelProvider()
	chService, err := chProvider.NewChannelService(session, channelID)
	if err != nil {
		return nil, errors.WithMessage(err, "create channel service failed")
	}

	discoveryService, err := providers.DiscoveryProvider().NewDiscoveryService(channelID)
	if err != nil {
		return nil, errors.WithMessage(err, "create discovery service failed")
	}

	discoveryService = discovery.NewDiscoveryFilterService(discoveryService, targetFilter)

	selection, err := providers.SelectionProvider().NewSelectionService(channelID)
	if err != nil {
		return nil, errors.WithMessage(err, "create selection service failed")
	}

	ctx := chclient.Context{
		ProviderContext:  providers,
		DiscoveryService: discoveryService,
		SelectionService: selection,
		ChannelService:   chService,
	}
	return chclient.New(ctx)
}
