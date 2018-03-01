/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package defsvc

import (
	"testing"

	discovery "github.com/hyperledger/fabric-sdk-go/pkg/client/common/discovery/staticdiscovery"
	selection "github.com/hyperledger/fabric-sdk-go/pkg/client/common/selection/staticselection"
	"github.com/hyperledger/fabric-sdk-go/pkg/context/api/core"
	"github.com/hyperledger/fabric-sdk-go/pkg/context/api/fab"
	"github.com/hyperledger/fabric-sdk-go/pkg/fab/mocks"
	"github.com/hyperledger/fabric-sdk-go/pkg/fab/peer"
	"github.com/hyperledger/fabric-sdk-go/pkg/fabsdk/provider/fabpvdr"
)

func TestCreateDiscoveryProvider(t *testing.T) {
	ctx := mocks.NewMockContext(mocks.NewMockUser("testuser"))
	fabPvdr := fabpvdr.New(ctx)

	factory := NewProviderFactory()
	config := mocks.NewMockConfig()

	dp, err := factory.CreateDiscoveryProvider(config, fabPvdr)
	if err != nil {
		t.Fatalf("Unexpected error creating discovery provider %v", err)
	}

	_, ok := dp.(*discovery.DiscoveryProvider)
	if !ok {
		t.Fatalf("Unexpected discovery provider created")
	}
}

func TestCreateSelectionProvider(t *testing.T) {
	factory := NewProviderFactory()

	config := mocks.NewMockConfig()

	dp, err := factory.CreateSelectionProvider(config)
	if err != nil {
		t.Fatalf("Unexpected error creating discovery provider %v", err)
	}

	_, ok := dp.(*selection.SelectionProvider)
	if !ok {
		t.Fatalf("Unexpected selection provider created")
	}
}

type defPeerCreator struct {
	config core.Config
}

func (pc *defPeerCreator) CreatePeerFromConfig(peerCfg *core.NetworkPeer) (fab.Peer, error) {
	return peer.New(pc.config, peer.FromPeerConfig(peerCfg))
}
