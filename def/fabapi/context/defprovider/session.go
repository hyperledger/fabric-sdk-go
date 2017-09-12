/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package defprovider

import (
	"fmt"

	"github.com/hyperledger/fabric-sdk-go/api/apiconfig"
	fab "github.com/hyperledger/fabric-sdk-go/api/apifabclient"
	"github.com/hyperledger/fabric-sdk-go/api/apitxn"
	"github.com/hyperledger/fabric-sdk-go/def/fabapi/context"
	clientImpl "github.com/hyperledger/fabric-sdk-go/pkg/fabric-client"
	"github.com/hyperledger/fabric-sdk-go/pkg/fabric-client/events"
	"github.com/hyperledger/fabric-sdk-go/pkg/fabric-client/orderer"
	chImpl "github.com/hyperledger/fabric-sdk-go/pkg/fabric-txn/chclient"
)

// SessionClientFactory represents the default implementation of a session client.
type SessionClientFactory struct{}

// NewSessionClientFactory creates a new default session client factory.
func NewSessionClientFactory() *SessionClientFactory {
	f := SessionClientFactory{}
	return &f
}

// NewSystemClient returns a new FabricClient.
func (f *SessionClientFactory) NewSystemClient(sdk context.SDK, session context.Session, config apiconfig.Config) (fab.FabricClient, error) {
	client := clientImpl.NewClient(config)

	client.SetCryptoSuite(sdk.CryptoSuiteProvider())
	client.SetStateStore(sdk.StateStoreProvider())
	client.SetUserContext(session.Identity())
	client.SetSigningManager(sdk.SigningManager())

	return client, nil
}

// NewChannelClient returns a client that can execute transactions on specified channel
func (f *SessionClientFactory) NewChannelClient(sdk context.SDK, session context.Session, config apiconfig.Config, channelName string) (apitxn.ChannelClient, error) {

	client := clientImpl.NewClient(sdk.ConfigProvider())
	client.SetCryptoSuite(sdk.CryptoSuiteProvider())
	client.SetStateStore(sdk.StateStoreProvider())
	client.SetUserContext(session.Identity())
	client.SetSigningManager(sdk.SigningManager())

	channel, err := getChannel(client, channelName)
	if err != nil {
		return nil, fmt.Errorf("Unable to create channel:%v", err)
	}

	discovery, err := sdk.DiscoveryProvider().NewDiscoveryService(channel)
	if err != nil {
		return nil, fmt.Errorf("Unable to create discovery service:%v", err)
	}

	eventHub, err := getEventHub(client, channelName)
	if err != nil {
		return nil, err
	}

	return chImpl.NewChannelClient(client, channel, discovery, eventHub)
}

// getChannel is helper method to initializes and returns a channel based on config
func getChannel(client fab.FabricClient, channelID string) (fab.Channel, error) {

	channel, err := client.NewChannel(channelID)
	if err != nil {
		return nil, fmt.Errorf("NewChannel return error: %v", err)
	}

	chConfig, err := client.Config().ChannelConfig(channel.Name())
	if err != nil {
		return nil, fmt.Errorf("Error reading channel config: %v", err)
	}

	for _, name := range chConfig.Orderers {
		ordererConfig, err := client.Config().OrdererConfig(name)
		if err != nil {
			return nil, fmt.Errorf("Unable to retrieve configuration for orderer(%s): %s", name, err)
		}

		serverHostOverride := ""
		if str, ok := ordererConfig.GRPCOptions["ssl-target-name-override"].(string); ok {
			serverHostOverride = str
		}
		orderer, err := orderer.NewOrderer(ordererConfig.URL, ordererConfig.TLSCACerts.Path, serverHostOverride, client.Config())
		if err != nil {
			return nil, fmt.Errorf("NewOrderer return error: %v", err)
		}
		err = channel.AddOrderer(orderer)
		if err != nil {
			return nil, fmt.Errorf("Error adding orderer: %v", err)
		}
	}

	return channel, nil
}

func getEventHub(client fab.FabricClient, channelID string) (*events.EventHub, error) {

	peerConfig, err := client.Config().ChannelPeers(channelID)
	if err != nil {
		return nil, fmt.Errorf("Unable to read configuration for channel(%s) peers: %s", channelID, err)
	}

	serverHostOverride := ""
	var eventSource *apiconfig.PeerConfig

	for _, p := range peerConfig {

		if p.EventSource {
			serverHostOverride = ""
			if str, ok := p.GRPCOptions["ssl-target-name-override"].(string); ok {
				serverHostOverride = str
			}
			eventSource = &p.PeerConfig
			break
		}
	}

	if eventSource == nil {
		return nil, fmt.Errorf("Unable to find peer event source for channel: %s", channelID)
	}

	// Event source found create event hub
	eventHub, err := events.NewEventHub(client)
	if err != nil {
		return nil, err
	}

	eventHub.SetPeerAddr(eventSource.EventURL, eventSource.TLSCACerts.Path, serverHostOverride)

	return eventHub, nil
}
