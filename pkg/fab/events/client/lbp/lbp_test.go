/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package lbp

import (
	"fmt"
	"testing"

	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/fab"
	fabmocks "github.com/hyperledger/fabric-sdk-go/pkg/fab/mocks"
)

func TestRandom(t *testing.T) {
	lbp := NewRandom()

	// Test with an empty set of peers
	peer, err := lbp.Choose([]fab.Peer{})
	if err != nil {
		t.Fatalf("error choosing peer with random load-balance policy: %s", err)
	}
	if peer != nil {
		t.Fatal("expecting chosen peer to be nil with empty set of peers")
	}

	peers := newMockPeers(10)

	// Invoke a number of times and make sure it doesn't choose the same peer each time
	numInvocations := 10
	var lastPeerChosen fab.Peer
	differentPeerChosen := false

	for numInvocations > 0 {
		numInvocations--
		peer, err := lbp.Choose(peers)
		if err != nil {
			t.Fatalf("error choosing peer with random load-balance policy: %s", err)
		}
		if lastPeerChosen != nil && peer != lastPeerChosen {
			differentPeerChosen = true
			break
		}
		lastPeerChosen = peer
	}

	if !differentPeerChosen {
		t.Fatal("the same peer was chosen every time")
	}
}

func TestRoundRobin(t *testing.T) {
	lbp := NewRoundRobin()

	// Test with an empty set of peers
	peer, err := lbp.Choose([]fab.Peer{})
	if err != nil {
		t.Fatalf("error choosing peer with random load-balance policy: %s", err)
	}
	if peer != nil {
		t.Fatal("expecting chosen peer to be nil with empty set of peers")
	}

	peers := newMockPeers(10)

	lastIndexChosen := -1

	// Invoke a number of times and make sure it chooses each one consecutively
	for i := 0; i < len(peers); i++ {
		peer, err := lbp.Choose(peers)
		if err != nil {
			t.Fatalf("error choosing peer with round-robin load-balance policy: %s", err)
		}

		chosenIndex := findIndex(peers, peer)
		if lastIndexChosen >= 0 {
			if lastIndexChosen == (len(peers) - 1) {
				if chosenIndex != 0 {
					t.Fatalf("expecting chosen index to be 0 but got index %d", chosenIndex)
				}
			} else {
				if chosenIndex != lastIndexChosen+1 {
					t.Fatalf("expecting chosen index to be % but got index %d", lastIndexChosen+1, chosenIndex)
				}
			}
		}
		lastIndexChosen = chosenIndex
	}
}

func findIndex(peers []fab.Peer, peer fab.Peer) int {
	for i, p := range peers {
		if peer == p {
			return i
		}
	}
	panic("peer does not exist in list of peers")
}

func newMockPeers(numPeers int) []fab.Peer {
	var peers []fab.Peer
	for i := 0; i < numPeers; i++ {
		peers = append(peers, fabmocks.NewMockPeer(fmt.Sprintf("peer_%d", i), ""))
	}
	return peers
}
