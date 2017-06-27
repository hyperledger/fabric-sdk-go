/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package fabricclient

import (
	"testing"

	mocks "github.com/hyperledger/fabric-sdk-go/pkg/fabric-client/mocks"
	"github.com/hyperledger/fabric-sdk-go/pkg/fabric-client/peer"
)

const (
	ecertPath  = "../../../test/fixtures/tls/fabricca/client/client_client1.pem"
	peer1URL   = "localhost:7050"
	peer2URL   = "localhost:7054"
	peerURLBad = "localhost:9999"
)

//
// Peer via channel setPeer/getPeer
//
// Set the Peer URL through the channel setPeer method. Verify that the
// Peer URL was set correctly through the getPeer method. Repeat the
// process by updating the Peer URL to a different address.
//
func TestPeerViaChannel(t *testing.T) {
	config := mocks.NewMockConfig()

	client := NewClient(config)
	channel, err := client.NewChannel("testChannel-peer")
	if err != nil {
		t.Fatalf("error from NewChannel %v", err)
	}
	peer1, err := peer.NewPeer(peer1URL, config)
	if err != nil {
		t.Fatalf("Failed to create NewPeer error(%v)", err)
	}
	err = channel.AddPeer(peer1)
	if err != nil {
		t.Fatalf("Error adding peer: %v", err)
	}

	peers := channel.GetPeers()
	if peers == nil || len(peers) != 1 || peers[0].URL() != peer1URL {
		t.Fatalf("Failed to retieve the new peers URL from the channel")
	}
	channel.RemovePeer(peer1)
	peer2, err := peer.NewPeer(peer2URL, config)
	if err != nil {
		t.Fatalf("Failed to create NewPeer error(%v)", err)
	}
	err = channel.AddPeer(peer2)
	if err != nil {
		t.Fatalf("Error adding peer: %v", err)
	}
	peers = channel.GetPeers()

	if peers == nil || len(peers) != 1 || peers[0].URL() != peer2URL {
		t.Fatalf("Failed to retieve the new peers URL from the channel")
	}
}

//
// Peer via channel missing peer
//
// Attempt to send a request to the peer with the SendTransactionProposal method
// before the peer was set. Verify that an error is reported when tying
// to send the request.
//
func TestOrdererViaChannelMissingOrderer(t *testing.T) {
	config := mocks.NewMockConfig()

	client := NewClient(config)
	channel, err := client.NewChannel("testChannel-peer")
	if err != nil {
		t.Fatalf("error from NewChannel %v", err)
	}
	_, err = channel.SendTransactionProposal(nil, 0, nil)
	if err == nil {
		t.Fatalf("SendTransactionProposal didn't return error")
	}
}

//
// Peer via channel nil data
//
// Attempt to send a request to the peers with the SendTransactionProposal method
// with the data set to null. Verify that an error is reported when tying
// to send null data.
//
func TestPeerViaChannelNilData(t *testing.T) {
	config := mocks.NewMockConfig()

	client := NewClient(config)
	channel, err := client.NewChannel("testChannel-peer")
	if err != nil {
		t.Fatalf("error from NewChannel %v", err)
	}
	peer, err := peer.NewPeer(peer1URL, config)
	if err != nil {
		t.Fatalf("Failed to create NewPeer error(%v)", err)
	}
	err = channel.AddPeer(peer)
	if err != nil {
		t.Fatalf("Error adding peer: %v", err)
	}
	_, err = channel.SendTransactionProposal(nil, 0, nil)
	if err == nil {
		t.Fatalf("SendTransaction didn't return error")
	}
	if err.Error() != "signedProposal is nil" {
		t.Fatalf("SendTransactionProposal didn't return right error")
	}
}
