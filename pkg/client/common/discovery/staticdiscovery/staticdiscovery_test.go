/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package staticdiscovery

import (
	"testing"

	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/fab"
	"github.com/hyperledger/fabric-sdk-go/pkg/core/config"
	fabImpl "github.com/hyperledger/fabric-sdk-go/pkg/fab"
	"github.com/hyperledger/fabric-sdk-go/pkg/fab/peer"
)

func TestStaticDiscovery(t *testing.T) {

	configBackend, err := config.FromFile("../../../../../test/fixtures/config/config_test.yaml")()
	if err != nil {
		t.Fatalf(err.Error())
	}

	config1, err := fabImpl.ConfigFromBackend(configBackend)
	if err != nil {
		t.Fatalf(err.Error())
	}

	peerCreator := defPeerCreator{config: config1}
	discoveryProvider, err := New(config1, &peerCreator)
	if err != nil {
		t.Fatalf("Failed to  setup discovery provider: %s", err)
	}

	discoveryService, err := discoveryProvider.CreateDiscoveryService("mychannel")
	if err != nil {
		t.Fatalf("Failed to setup discovery service: %s", err)
	}

	peers, err := discoveryService.GetPeers()
	if err != nil {
		t.Fatalf("Failed to get peers from discovery service: %s", err)
	}

	// One peer is configured for "mychannel"
	expectedNumOfPeeers := 1
	if len(peers) != expectedNumOfPeeers {
		t.Fatalf("Expecting %d, got %d peers", expectedNumOfPeeers, len(peers))
	}

}

func TestStaticDiscoveryWhenChannelIsEmpty(t *testing.T) {
	configBackend, err := config.FromFile("../../../../../test/fixtures/config/config_test.yaml")()
	if err != nil {
		t.Fatalf(err.Error())
	}

	config1, err := fabImpl.ConfigFromBackend(configBackend)
	if err != nil {
		t.Fatalf(err.Error())
	}

	peerCreator := defPeerCreator{config: config1}
	discoveryProvider, _ := New(config1, &peerCreator)
	// If channel is empty discovery service will return all configured network peers
	discoveryService, err := discoveryProvider.CreateDiscoveryService("")
	if err != nil {
		t.Fatalf("Failed to setup discovery service: %s", err)
	}

	peers, err := discoveryService.GetPeers()
	if err != nil {
		t.Fatalf("Failed to get peers from discovery service: %s", err)
	}

	// Two peers are configured at network level
	expectedNumOfPeeers := 2
	if len(peers) != expectedNumOfPeeers {
		t.Fatalf("Expecting %d, got %d peers", expectedNumOfPeeers, len(peers))
	}
}

type defPeerCreator struct {
	config fab.EndpointConfig
}

func (pc *defPeerCreator) CreatePeerFromConfig(peerCfg *fab.NetworkPeer) (fab.Peer, error) {
	return peer.New(pc.config, peer.FromPeerConfig(peerCfg))
}
