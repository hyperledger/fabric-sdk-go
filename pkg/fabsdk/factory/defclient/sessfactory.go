/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package defclient

import (
	apichclient "github.com/hyperledger/fabric-sdk-go/api/apitxn/chclient"
	apichmgmt "github.com/hyperledger/fabric-sdk-go/api/apitxn/chmgmtclient"
	apiresmgmt "github.com/hyperledger/fabric-sdk-go/api/apitxn/resmgmtclient"
	apisdk "github.com/hyperledger/fabric-sdk-go/pkg/fabsdk/api"

	"github.com/hyperledger/fabric-sdk-go/api/apifabclient"
	"github.com/hyperledger/fabric-sdk-go/pkg/fabric-txn/chclient"
	"github.com/hyperledger/fabric-sdk-go/pkg/fabric-txn/chmgmtclient"
	"github.com/hyperledger/fabric-sdk-go/pkg/fabric-txn/discovery"
	"github.com/hyperledger/fabric-sdk-go/pkg/fabric-txn/resmgmtclient"
	"github.com/pkg/errors"
)

// SessionClientFactory represents the default implementation of a session client.
type SessionClientFactory struct{}

// NewSessionClientFactory creates a new default session client factory.
func NewSessionClientFactory() *SessionClientFactory {
	f := SessionClientFactory{}
	return &f
}

// NewChannelMgmtClient returns a client that manages channels (create/join channel)
func (f *SessionClientFactory) NewChannelMgmtClient(providers apisdk.Providers, session apisdk.SessionContext) (apichmgmt.ChannelMgmtClient, error) {
	// For now settings are the same as for system client
	resource, err := providers.FabricProvider().CreateResourceClient(session)
	if err != nil {
		return nil, err
	}
	ctx := chmgmtclient.Context{
		ProviderContext: providers,
		IdentityContext: session,
		Resource:        resource,
	}
	return chmgmtclient.New(ctx)
}

// NewResourceMgmtClient returns a client that manages resources
func (f *SessionClientFactory) NewResourceMgmtClient(providers apisdk.Providers, session apisdk.SessionContext, filter apiresmgmt.TargetFilter) (apiresmgmt.ResourceMgmtClient, error) {

	fabProvider := providers.FabricProvider()
	resource, err := fabProvider.CreateResourceClient(session)
	if err != nil {
		return nil, err
	}

	discovery := providers.DiscoveryProvider()
	chProvider := providers.ChannelProvider()

	ctx := resmgmtclient.Context{
		ProviderContext:   providers,
		IdentityContext:   session,
		Resource:          resource,
		DiscoveryProvider: discovery,
		ChannelProvider:   chProvider,
		FabricProvider:    fabProvider,
	}
	return resmgmtclient.New(ctx, filter)
}

// NewChannelClient returns a client that can execute transactions on specified channel
func (f *SessionClientFactory) NewChannelClient(providers apisdk.Providers, session apisdk.SessionContext, channelID string, targetFilter apifabclient.TargetFilter) (apichclient.ChannelClient, error) {

	chProvider := providers.ChannelProvider()
	chService, err := chProvider.NewChannelService(session, channelID)
	if err != nil {
		return nil, errors.WithMessage(err, "create channel service failed")
	}

	channel, err := chService.Channel()
	if err != nil {
		return nil, errors.WithMessage(err, "create channel failed")
	}

	eventHub, err := chService.EventHub()
	if err != nil {
		return nil, errors.WithMessage(err, "getEventHub failed")
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
		Channel:          channel,
		DiscoveryService: discoveryService,
		SelectionService: selection,
		EventHub:         eventHub,
	}
	return chclient.New(ctx)
}
