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
	"net"
	"testing"

	"google.golang.org/grpc"

	mocks "github.com/hyperledger/fabric-sdk-go/fabric-client/mocks"
	"github.com/hyperledger/fabric-sdk-go/fabric-client/util"
	pb "github.com/hyperledger/fabric/protos/peer"
)

var testAddress = "0.0.0.0:5244"

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

func TestJoinChannel(t *testing.T) {
	var peers []Peer
	endorserServer := startEndorserServer(t)
	chain, _ := setupTestChain()
	peer, _ := NewPeer(testAddress, "", "")
	peers = append(peers, peer)
	orderer := &mockOrderer{DeliverResponse: NewMockDeliverResponse(mocks.NewSimpleMockBlock())}
	nonce, _ := util.GenerateRandomNonce()
	txID, _ := util.ComputeTxID(nonce, []byte("testID"))
	request := &JoinChannelRequest{Targets: peers, Nonce: nonce, TxID: txID}
	chain.AddOrderer(orderer)
	chain.AddPeer(peer)
	// Test join channel with valid arguments
	err := chain.JoinChannel(request)
	if err != nil {
		t.Fatalf("Did not expect error from join channel. Got: %s", err)
	}
	// Test join channel without request
	err = chain.JoinChannel(nil)
	if err.Error() != "JoinChannelRequest argument is required to join channel" {
		t.Fatalf("Expected error without join channel request")
	}
	// Test join channel without target peers
	request = &JoinChannelRequest{Targets: nil, Nonce: nonce, TxID: txID}
	err = chain.JoinChannel(request)
	if err.Error() != "Atleast one target peer is required to join channel" {
		t.Fatalf("Expected error without target peers")
	}
	// Test join channel without nonce
	request = &JoinChannelRequest{Targets: peers, Nonce: nil, TxID: txID}
	err = chain.JoinChannel(request)
	if err.Error() != "Nonce is required to join channel" {
		t.Fatalf("Expected error without nonce")
	}
	// Test join channel without TxID
	request = &JoinChannelRequest{Targets: peers, Nonce: nonce, TxID: ""}
	err = chain.JoinChannel(request)
	if err.Error() != "Transaction ID is required to join channel" {
		t.Fatalf("Expected error without transaction ID")
	}
	// Test failed proposal error handling
	endorserServer.ProposalError = fmt.Errorf("Test Error")
	request = &JoinChannelRequest{Targets: peers, Nonce: nonce, TxID: txID}
	err = chain.JoinChannel(request)
	if err == nil {
		t.Fatalf("Expected error")
	}
}

func TestChainInitializeFromOrderer(t *testing.T) {
	org1MSPID := "ORG1MSP"
	org2MSPID := "ORG2MSP"

	chain, _ := setupTestChain()
	builder := &mocks.MockConfigBlockBuilder{
		MockConfigGroupBuilder: mocks.MockConfigGroupBuilder{
			ModPolicy: "Admins",
			MSPNames: []string{
				org1MSPID,
				org2MSPID,
			},
			OrdererAddress: "localhost:7054",
		},
		Index:           0,
		LastConfigIndex: 0,
	}
	orderer := &mockOrderer{DeliverResponse: NewMockDeliverResponse(builder.Build())}
	chain.AddOrderer(orderer)

	err := chain.Initialize(nil)
	if err != nil {
		t.Fatalf("channel Initialize failed : %v", err)
	}
	mspManager := chain.GetMSPManager()
	if mspManager == nil {
		t.Fatalf("nil MSPManager on new chain")
	}
	msps, err := mspManager.GetMSPs()
	if err != nil || len(msps) == 0 {
		t.Fatalf("At least one MSP expected in MSPManager")
	}
	msp, ok := msps[org1MSPID]
	if !ok {
		t.Fatalf("Could not find %s", org1MSPID)
	}
	if identifier, _ := msp.GetIdentifier(); identifier != org1MSPID {
		t.Fatalf("Expecting MSP identifier to be %s but got %s", org1MSPID, identifier)
	}
	msp, ok = msps[org2MSPID]
	if !ok {
		t.Fatalf("Could not find %s", org2MSPID)
	}
	if identifier, _ := msp.GetIdentifier(); identifier != org2MSPID {
		t.Fatalf("Expecting MSP identifier to be %s but got %s", org2MSPID, identifier)
	}
}

func TestOrganizationUnits(t *testing.T) {
	org1MSPID := "ORG1MSP"
	org2MSPID := "ORG2MSP"

	chain, _ := setupTestChain()
	orgUnits, err := chain.GetOrganizationUnits()
	if len(orgUnits) > 0 {
		t.Fatalf("Returned non configured organizational unit : %v", err)
	}
	builder := &mocks.MockConfigBlockBuilder{
		MockConfigGroupBuilder: mocks.MockConfigGroupBuilder{
			ModPolicy: "Admins",
			MSPNames: []string{
				chain.GetName(),
				org1MSPID,
				org2MSPID,
			},
			OrdererAddress: "localhost:7054",
		},
		Index:           0,
		LastConfigIndex: 0,
	}
	orderer := &mockOrderer{DeliverResponse: NewMockDeliverResponse(builder.Build())}
	chain.AddOrderer(orderer)

	err = chain.Initialize(nil)
	if err != nil {
		t.Fatalf("channel Initialize failed : %v", err)
	}
	orgUnits, err = chain.GetOrganizationUnits()
	if err != nil {
		t.Fatalf("CANNOT retrieve organizational units : %v", err)
	}
	if !isValueInList(chain.GetName(), orgUnits) {
		t.Fatalf("Could not find %s in the list of organizations", chain.GetName())
	}
	if !isValueInList(org1MSPID, orgUnits) {
		t.Fatalf("Could not find %s in the list of organizations", org1MSPID)
	}
	if !isValueInList(org2MSPID, orgUnits) {
		t.Fatalf("Could not find %s in the list of organizations", org2MSPID)
	}

}

func isValueInList(value string, list []string) bool {
	for _, v := range list {
		if v == value {
			return true
		}
	}
	return false
}

func TestChainInitializeFromUpdate(t *testing.T) {
	org1MSPID := "ORG1MSP"
	org2MSPID := "ORG2MSP"

	chain, _ := setupTestChain()
	builder := &mocks.MockConfigUpdateEnvelopeBuilder{
		ChannelID: "testchannel",
		MockConfigGroupBuilder: mocks.MockConfigGroupBuilder{
			ModPolicy: "Admins",
			MSPNames: []string{
				org1MSPID,
				org2MSPID,
			},
			OrdererAddress: "localhost:7054",
		},
	}

	err := chain.Initialize(builder.BuildBytes())
	if err != nil {
		t.Fatalf("channel Initialize failed : %v", err)
	}
	mspManager := chain.GetMSPManager()
	if mspManager == nil {
		t.Fatalf("nil MSPManager on new chain")
	}
	msps, err := mspManager.GetMSPs()
	if err != nil || len(msps) == 0 {
		t.Fatalf("At least one MSP expected in MSPManager")
	}
	msp, ok := msps[org1MSPID]
	if !ok {
		t.Fatalf("Could not find %s", org1MSPID)
	}
	if identifier, _ := msp.GetIdentifier(); identifier != org1MSPID {
		t.Fatalf("Expecting MSP identifier to be %s but got %s", org1MSPID, identifier)
	}
	msp, ok = msps[org2MSPID]
	if !ok {
		t.Fatalf("Could not find %s", org2MSPID)
	}
	if identifier, _ := msp.GetIdentifier(); identifier != org2MSPID {
		t.Fatalf("Expecting MSP identifier to be %s but got %s", org2MSPID, identifier)
	}
}

func setupTestChain() (Chain, error) {
	client := NewClient()
	user := NewUser("test")
	cryptoSuite := &mocks.MockCryptoSuite{}
	client.SaveUserToStateStore(user, true)
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
		orderer := mockOrderer{MockURL: fmt.Sprintf("http://mock%d.orderers.r.us", i)}
		chain.AddOrderer(&orderer)
	}

	return chain, error
}

func startEndorserServer(t *testing.T) *mocks.MockEndorserServer {
	grpcServer := grpc.NewServer()
	lis, err := net.Listen("tcp", testAddress)
	endorserServer := &mocks.MockEndorserServer{}
	pb.RegisterEndorserServer(grpcServer, endorserServer)
	if err != nil {
		fmt.Printf("Error starting test server %s", err)
		t.FailNow()
	}
	fmt.Printf("Starting test server\n")
	go grpcServer.Serve(lis)
	return endorserServer
}
