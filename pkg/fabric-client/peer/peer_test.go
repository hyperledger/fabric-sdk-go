/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package peer

import (
	"testing"

	client "github.com/hyperledger/fabric-sdk-go/pkg/fabric-client/client"
	mocks "github.com/hyperledger/fabric-sdk-go/pkg/fabric-client/mocks"
)

//
// Peer via chain setPeer/getPeer
//
// Set the Perr URL through the chain setPeer method. Verify that the
// Peer URL was set correctly through the getPeer method. Repeat the
// process by updating the Peer URL to a different address.
//
func TestPeerViaChain(t *testing.T) {
	client := client.NewClient(mocks.NewMockConfig())
	chain, err := client.NewChannel("testChain-peer")
	if err != nil {
		t.Fatalf("error from NewChain %v", err)
	}
	peer, err := NewPeer("localhost:7050", "", "", mocks.NewMockConfig())
	if err != nil {
		t.Fatalf("Failed to create NewPeer error(%v)", err)
	}
	err = chain.AddPeer(peer)
	if err != nil {
		t.Fatalf("Error adding peer: %v", err)
	}

	peers := chain.GetPeers()
	if peers == nil || len(peers) != 1 || peers[0].GetURL() != "localhost:7050" {
		t.Fatalf("Failed to retieve the new peers URL from the chain")
	}
	chain.RemovePeer(peer)
	peer2, err := NewPeer("localhost:7054", "", "", mocks.NewMockConfig())
	if err != nil {
		t.Fatalf("Failed to create NewPeer error(%v)", err)
	}
	err = chain.AddPeer(peer2)
	if err != nil {
		t.Fatalf("Error adding peer: %v", err)
	}
	peers = chain.GetPeers()

	if peers == nil || len(peers) != 1 || peers[0].GetURL() != "localhost:7054" {
		t.Fatalf("Failed to retieve the new peers URL from the chain")
	}
}

//
// Peer via chain missing peer
//
// Attempt to send a request to the peer with the SendTransactionProposal method
// before the peer was set. Verify that an error is reported when tying
// to send the request.
//
func TestOrdererViaChainMissingOrderer(t *testing.T) {
	client := client.NewClient(mocks.NewMockConfig())
	chain, err := client.NewChannel("testChain-peer")
	if err != nil {
		t.Fatalf("error from NewChain %v", err)
	}
	_, err = chain.SendTransactionProposal(nil, 0, nil)
	if err == nil {
		t.Fatalf("SendTransactionProposal didn't return error")
	}
	if err.Error() != "peers is nil" {
		t.Fatalf("SendTransactionProposal didn't return right error")
	}
}

//
// Peer via chain nil data
//
// Attempt to send a request to the peers with the SendTransactionProposal method
// with the data set to null. Verify that an error is reported when tying
// to send null data.
//
func TestPeerViaChainNilData(t *testing.T) {
	client := client.NewClient(mocks.NewMockConfig())
	chain, err := client.NewChannel("testChain-peer")
	if err != nil {
		t.Fatalf("error from NewChain %v", err)
	}
	peer, err := NewPeer("localhost:7050", "", "", mocks.NewMockConfig())
	if err != nil {
		t.Fatalf("Failed to create NewPeer error(%v)", err)
	}
	err = chain.AddPeer(peer)
	if err != nil {
		t.Fatalf("Error adding peer: %v", err)
	}
	_, err = chain.SendTransactionProposal(nil, 0, nil)
	if err == nil {
		t.Fatalf("SendTransaction didn't return error")
	}
	if err.Error() != "signedProposal is nil" {
		t.Fatalf("SendTransactionProposal didn't return right error")
	}
}
