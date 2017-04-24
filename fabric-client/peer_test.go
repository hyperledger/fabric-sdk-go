/*
Copyright SecureKey Technologies Inc. All Rights Reserved.


Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at


      http://www.apache.org/licenses/LICENSE-2.0


Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package fabricclient

import (
	"testing"
)

//
// Peer via chain setPeer/getPeer
//
// Set the Perr URL through the chain setPeer method. Verify that the
// Peer URL was set correctly through the getPeer method. Repeat the
// process by updating the Peer URL to a different address.
//
func TestPeerViaChain(t *testing.T) {
	client := NewClient()
	chain, err := client.NewChain("testChain-peer")
	if err != nil {
		t.Fatalf("error from NewChain %v", err)
	}
	peer, err := NewPeer("localhost:7050", "", "")
	if err != nil {
		t.Fatalf("Failed to create NewPeer error(%v)", err)
	}
	chain.AddPeer(peer)

	peers := chain.GetPeers()
	if peers == nil || len(peers) != 1 || peers[0].GetURL() != "localhost:7050" {
		t.Fatalf("Failed to retieve the new peers URL from the chain")
	}
	chain.RemovePeer(peer)
	peer2, err := NewPeer("localhost:7054", "", "")
	if err != nil {
		t.Fatalf("Failed to create NewPeer error(%v)", err)
	}
	chain.AddPeer(peer2)
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
	client := NewClient()
	chain, err := client.NewChain("testChain-peer")
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
	client := NewClient()
	chain, err := client.NewChain("testChain-peer")
	if err != nil {
		t.Fatalf("error from NewChain %v", err)
	}
	peer, err := NewPeer("localhost:7050", "", "")
	if err != nil {
		t.Fatalf("Failed to create NewPeer error(%v)", err)
	}
	chain.AddPeer(peer)
	_, err = chain.SendTransactionProposal(nil, 0, nil)
	if err == nil {
		t.Fatalf("SendTransaction didn't return error")
	}
	if err.Error() != "signedProposal is nil" {
		t.Fatalf("SendTransactionProposal didn't return right error")
	}
}
