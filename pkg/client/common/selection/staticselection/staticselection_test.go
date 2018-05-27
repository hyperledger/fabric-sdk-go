/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package staticselection

import (
	"testing"

	"github.com/hyperledger/fabric-sdk-go/pkg/client/common/selection/options"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/fab"

	fabmocks "github.com/hyperledger/fabric-sdk-go/pkg/fab/mocks"
)

func TestStaticSelection(t *testing.T) {
	peer1 := fabmocks.NewMockPeer("p1", "localhost:7051")
	peer2 := fabmocks.NewMockPeer("p2", "localhost:8051")

	selectionService, err := NewService(fabmocks.NewMockDiscoveryService(nil, peer1, peer2))
	if err != nil {
		t.Fatalf("Failed to setup selection provider: %s", err)
	}

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
