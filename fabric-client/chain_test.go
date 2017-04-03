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
	"fmt"
	"io/ioutil"
	"testing"

	mocks "github.com/hyperledger/fabric-sdk-go/fabric-client/mocks"
	pb "github.com/hyperledger/fabric/protos/peer"
)

func TestChainMethods(t *testing.T) {
	client := NewClient()
	chain, err := NewChain("testChain", client)
	if err != nil {
		t.Fatalf("NewChain return error[%s]", err)
	}
	if chain.GetName() != "testChain" {
		t.Fatalf("NewChain create wrong chain")
	}

	_, err = NewChain("", client)
	if err == nil {
		t.Fatalf("NewChain didn't return error")
	}
	if err.Error() != "failed to create Chain. Missing required 'name' parameter" {
		t.Fatalf("NewChain didn't return right error")
	}

	_, err = NewChain("testChain", nil)
	if err == nil {
		t.Fatalf("NewChain didn't return error")
	}
	if err.Error() != "failed to create Chain. Missing required 'clientContext' parameter" {
		t.Fatalf("NewChain didn't return right error")
	}

}

func TestQueryMethods(t *testing.T) {
	chain, _ := setupTestChain()

	_, err := chain.QueryBlock(-1)
	if err == nil {
		t.Fatalf("Query block cannot be negative number")
	}

	_, err = chain.QueryBlockByHash(nil)
	if err == nil {
		t.Fatalf("Query hash cannot be nil")
	}

	_, err = chain.QueryChannels(nil)
	if err == nil {
		t.Fatalf("QueryChanels: peer cannot be nil")
	}

	_, err = chain.QueryInstalledChaincodes(nil)
	if err == nil {
		t.Fatalf("QueryInstalledChaincodes: peer cannot be nil")
	}

	_, err = chain.QueryByChaincode("", []string{"method"}, nil)
	if err == nil {
		t.Fatalf("QueryByChaincode: name cannot be empty")
	}

	_, err = chain.QueryByChaincode("qscc", nil, nil)
	if err == nil {
		t.Fatalf("QueryByChaincode: arguments cannot be empty")
	}

	_, err = chain.QueryByChaincode("qscc", []string{"method"}, nil)
	if err == nil {
		t.Fatalf("QueryByChaincode: targets cannot be empty")
	}

}

func TestTargetPeers(t *testing.T) {

	p := make(map[string]Peer)
	chain := &chain{name: "targetChain", peers: p}

	// Chain has two peers
	peer1 := mockPeer{"Peer1", "http://peer1.com", []string{}, nil}
	chain.AddPeer(&peer1)
	peer2 := mockPeer{"Peer2", "http://peer2.com", []string{}, nil}
	chain.AddPeer(&peer2)

	// Set target to invalid URL
	invalidChoice := mockPeer{"", "http://xyz.com", []string{}, nil}
	targetPeers, err := chain.getTargetPeers([]Peer{&invalidChoice})
	if err == nil {
		t.Fatalf("Target peer didn't fail for an invalid peer")
	}

	// Test target peers default to chain peers if target peers are not provided
	targetPeers, err = chain.getTargetPeers(nil)

	if err != nil || targetPeers == nil || len(targetPeers) != 2 {
		t.Fatalf("Target Peers failed to default")
	}

	// Set target to valid peer 2 URL
	choice := mockPeer{"", "http://peer2.com", []string{}, nil}
	targetPeers, err = chain.getTargetPeers([]Peer{&choice})
	if err != nil {
		t.Fatalf("Failed to get valid target peer")
	}

	// Test target equals our choice
	if len(targetPeers) != 1 || targetPeers[0].GetURL() != peer2.GetURL() || targetPeers[0].GetName() != peer2.GetName() {
		t.Fatalf("Primary and our choice are not equal")
	}

}

func TestPrimaryPeer(t *testing.T) {
	chain, _ := setupTestChain()

	// Chain had one peer
	peer1 := mockPeer{"Peer1", "http://peer1.com", []string{}, nil}
	chain.AddPeer(&peer1)

	// Test primary defaults to chain peer
	primary := chain.GetPrimaryPeer()
	if primary.GetURL() != peer1.GetURL() {
		t.Fatalf("Primary Peer failed to default")
	}

	// Chain has two peers
	peer2 := mockPeer{"Peer2", "http://peer2.com", []string{}, nil}
	chain.AddPeer(&peer2)

	// Set primary to invalid URL
	invalidChoice := mockPeer{"", "http://xyz.com", []string{}, nil}
	err := chain.SetPrimaryPeer(&invalidChoice)
	if err == nil {
		t.Fatalf("Primary Peer was set to an invalid peer")
	}

	// Set primary to valid peer 2 URL
	choice := mockPeer{"", "http://peer2.com", []string{}, nil}
	err = chain.SetPrimaryPeer(&choice)
	if err != nil {
		t.Fatalf("Failed to set valid primary peer")
	}

	// Test primary equals our choice
	primary = chain.GetPrimaryPeer()
	if primary.GetURL() != peer2.GetURL() || primary.GetName() != peer2.GetName() {
		t.Fatalf("Primary and our choice are not equal")
	}

}

func TestCreateChain(t *testing.T) {
	client := NewClient()
	chain, err := NewChain("testChain", client)
	if err != nil {
		t.Fatalf(err.Error())
	}
	// Create channel without configuration
	err = chain.CreateChannel(CreateChannelRequest{})
	if err == nil {
		t.Fatalf("Expected error creating channel without config tx")
	}

	configTx, err := ioutil.ReadFile("../test/fixtures/channel/testchannel.tx")
	if err != nil {
		t.Fatalf(err.Error())
	}
	// Setup mock orderer
	orderer := mockOrderer{fmt.Sprintf("0.0.0.0:1234"), nil}
	chain.AddOrderer(&orderer)
	// Test with valid cofiguration
	err = chain.CreateChannel(CreateChannelRequest{ConfigData: configTx})
	if err != nil {
		t.Fatalf("Did not expect error from create channel. Got error: %s", err.Error())
	}
}

func TestConcurrentPeers(t *testing.T) {
	const numPeers = 10000
	chain, err := setupMassiveTestChain(numPeers, 0)
	if err != nil {
		t.Fatalf("Failed to create massive chain: %s", err)
	}

	result, err := chain.SendTransactionProposal(&TransactionProposal{
		signedProposal: &pb.SignedProposal{},
	}, 1, nil)
	if err != nil {
		t.Fatalf("SendTransactionProposal return error: %s", err)
	}

	if len(result) != numPeers {
		t.Error("SendTransactionProposal returned an unexpected amount of responses")
	}
}

func TestConcurrentOrderers(t *testing.T) {
	const numOrderers = 10000
	chain, err := setupMassiveTestChain(0, numOrderers)
	if err != nil {
		t.Fatalf("Failed to create massive chain: %s", err)
	}

	txn := Transaction{
		proposal: &TransactionProposal{
			proposal: &pb.Proposal{},
		},
		transaction: &pb.Transaction{},
	}
	_, err = chain.SendTransaction(&txn)
	if err != nil {
		t.Fatalf("SendTransaction returned error: %s", err)
	}
}

func setupTestChain() (Chain, error) {
	client := NewClient()
	user := NewUser("test")
	cryptoSuite := &mocks.MockCryptoSuite{}
	client.SetUserContext(user, true)
	client.SetCryptoSuite(cryptoSuite)
	return NewChain("testChain", client)
}

func setupMassiveTestChain(numberOfPeers int, numberOfOrderers int) (Chain, error) {
	chain, error := setupTestChain()
	if error != nil {
		return chain, error
	}

	for i := 0; i < numberOfPeers; i++ {
		peer := mockPeer{fmt.Sprintf("MockPeer%d", i), fmt.Sprintf("http://mock%d.peers.r.us", i), []string{}, nil}
		chain.AddPeer(&peer)
	}

	for i := 0; i < numberOfOrderers; i++ {
		orderer := mockOrderer{fmt.Sprintf("http://mock%d.orderers.r.us", i), nil}
		chain.AddOrderer(&orderer)
	}

	return chain, error
}
