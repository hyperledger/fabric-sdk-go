/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package staticselection

import (
	"testing"

	"github.com/hyperledger/fabric-sdk-go/pkg/client/common/selection/options"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/fab"

	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/context"
	"github.com/hyperledger/fabric-sdk-go/pkg/core/config"
	fabImpl "github.com/hyperledger/fabric-sdk-go/pkg/fab"
	fabmocks "github.com/hyperledger/fabric-sdk-go/pkg/fab/mocks"
	mspmocks "github.com/hyperledger/fabric-sdk-go/pkg/msp/test/mockmsp"
)

type serviceInit interface {
	Initialize(context context.Channel) error
}

func TestStaticSelection(t *testing.T) {

	configBackend, err := config.FromFile("../../../../../test/fixtures/config/config_test.yaml")()
	if err != nil {
		t.Fatalf(err.Error())
	}

	config, err := fabImpl.ConfigFromBackend(configBackend...)
	if err != nil {
		t.Fatalf(err.Error())
	}

	peer1 := fabmocks.NewMockPeer("p1", "localhost:7051")
	peer2 := fabmocks.NewMockPeer("p2", "localhost:8051")

	selectionProvider, err := New(config)
	if err != nil {
		t.Fatalf("Failed to setup selection provider: %s", err)
	}

	selectionService, err := selectionProvider.CreateSelectionService("")
	if err != nil {
		t.Fatalf("Failed to setup selection service: %s", err)
	}

	ctx := fabmocks.NewMockContext(mspmocks.NewMockSigningIdentity("User1", ""))
	chctx := fabmocks.NewMockChannelContext(ctx, "testchannel")
	chctx.Discovery = fabmocks.NewMockDiscoveryService(nil, []fab.Peer{peer1, peer2})

	selectionService.(serviceInit).Initialize(chctx)

	peers, err := selectionService.GetEndorsersForChaincode(nil)
	if err != nil {
		t.Fatalf("Failed to get endorsers: %s", err)
	}

	expectedNumOfPeeers := 2
	if len(peers) != expectedNumOfPeeers {
		t.Fatalf("Expecting %d, got %d peers", expectedNumOfPeeers, len(peers))
	}

	peers, err = selectionService.GetEndorsersForChaincode(nil,
		options.WithPeerFilter(
			func(peer fab.Peer) bool {
				return peer.URL() == peer2.URL()
			},
		),
	)
	if err != nil {
		t.Fatalf("Failed to get endorsers: %s", err)
	}

	expectedNumOfPeeers = 1
	if len(peers) != expectedNumOfPeeers {
		t.Fatalf("Expecting %d, got %d peers", expectedNumOfPeeers, len(peers))
	}
	if peers[0].URL() != peer2.URL() {
		t.Fatalf("Expecting peer %s but got %s", peer2.URL(), peers[0].URL())
	}
}
