/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/
package channel

import (
	"fmt"
	"net"
	"testing"

	"google.golang.org/grpc"

	"github.com/golang/mock/gomock"
	pb "github.com/hyperledger/fabric/protos/peer"

	fab "github.com/hyperledger/fabric-sdk-go/api/apifabclient"
	"github.com/hyperledger/fabric-sdk-go/api/apitxn"
	"github.com/hyperledger/fabric-sdk-go/api/apitxn/mocks"
	fc "github.com/hyperledger/fabric-sdk-go/pkg/fabric-client/internal"
	"github.com/hyperledger/fabric-sdk-go/pkg/fabric-client/mocks"
	"github.com/hyperledger/fabric-sdk-go/pkg/fabric-client/peer"
)

func TestCreateTransactionProposal(t *testing.T) {

	channel, _ := setupTestChannel()

	tProposal, err := channel.CreateTransactionProposal("qscc", nil, true, nil)

	if err != nil {
		t.Fatal("Create Transaction Proposal Failed", err)
	}

	_, errx := channel.QueryExtensionInterface().ProposalBytes(tProposal)

	if errx != nil {
		t.Fatal("Call to proposal bytes from channel extension failed")
	}

}

func TestJoinChannel(t *testing.T) {
	var peers []fab.Peer

	grpcServer := grpc.NewServer()
	defer grpcServer.Stop()

	endorserServer, addr := startEndorserServer(t, grpcServer)
	channel, _ := setupTestChannel()
	peer, _ := peer.NewPeer(addr, mocks.NewMockConfig())
	peers = append(peers, peer)
	orderer := mocks.NewMockOrderer("", nil)
	orderer.(mocks.MockOrderer).EnqueueForSendDeliver(mocks.NewSimpleMockBlock())
	nonce, _ := fc.GenerateRandomNonce()
	txID, _ := fc.ComputeTxID(nonce, []byte("testID"))

	genesisBlockReqeust := &fab.GenesisBlockRequest{
		TxID:  txID,
		Nonce: nonce,
	}
	genesisBlock, err := channel.GenesisBlock(genesisBlockReqeust)
	if err == nil {
		t.Fatalf("Should not have been able to get genesis block because of orderer missing")
	}

	err = channel.AddOrderer(orderer)
	if err != nil {
		t.Fatalf("Error adding orderer: %v", err)
	}
	genesisBlock, err = channel.GenesisBlock(genesisBlockReqeust)
	if err != nil {
		t.Fatalf("Error getting genesis block: %v", err)
	}

	err = channel.JoinChannel(nil)
	if err == nil {
		t.Fatalf("Should not have been able to join channel because of missing request parameter")
	}

	request := &fab.JoinChannelRequest{
		Targets:      peers,
		GenesisBlock: genesisBlock,
		Nonce:        nonce,
		//TxID:         txID,
	}
	err = channel.JoinChannel(request)
	if err == nil {
		t.Fatalf("Should not have been able to join channel because of missing TxID parameter")
	}

	request = &fab.JoinChannelRequest{
		Targets:      peers,
		GenesisBlock: genesisBlock,
		//Nonce:        nonce,
		TxID: txID,
	}
	err = channel.JoinChannel(request)
	if err == nil {
		t.Fatalf("Should not have been able to join channel because of missing Nonce parameter")
	}

	request = &fab.JoinChannelRequest{
		Targets: peers,
		//GenesisBlock: genesisBlock,
		Nonce: nonce,
		TxID:  txID,
	}
	err = channel.JoinChannel(request)
	if err == nil {
		t.Fatalf("Should not have been able to join channel because of missing GenesisBlock parameter")
	}

	request = &fab.JoinChannelRequest{
		//Targets: peers,
		GenesisBlock: genesisBlock,
		Nonce:        nonce,
		TxID:         txID,
	}
	err = channel.JoinChannel(request)
	if err == nil {
		t.Fatalf("Should not have been able to join channel because of missing Targets parameter")
	}

	request = &fab.JoinChannelRequest{
		Targets:      peers,
		GenesisBlock: genesisBlock,
		Nonce:        nonce,
		TxID:         txID,
	}
	if err == nil {
		t.Fatalf("Should not have been able to join channel because of invalid targets")
	}

	err = channel.AddPeer(peer)
	if err != nil {
		t.Fatalf("Error adding peer: %v", err)
	}
	// Test join channel with valid arguments
	err = channel.JoinChannel(request)
	if err != nil {
		t.Fatalf("Did not expect error from join channel. Got: %s", err)
	}

	// Test failed proposal error handling
	endorserServer.ProposalError = fmt.Errorf("Test Error")
	request = &fab.JoinChannelRequest{Targets: peers, Nonce: nonce, TxID: txID}
	err = channel.JoinChannel(request)
	if err == nil {
		t.Fatalf("Expected error")
	}
}

func TestSendTransactionProposal(t *testing.T) {

	channel, _ := setupTestChannel()

	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()
	proc := mock_apitxn.NewMockProposalProcessor(mockCtrl)

	tp := apitxn.TransactionProposal{
		SignedProposal: &pb.SignedProposal{},
	}
	tpr := apitxn.TransactionProposalResult{Endorser: "example.com", Status: 99, Proposal: tp, ProposalResponse: nil}
	proc.EXPECT().ProcessTransactionProposal(tp).Return(tpr, nil)
	targets := []apitxn.ProposalProcessor{proc}

	result, err := channel.SendTransactionProposal(&apitxn.TransactionProposal{
		SignedProposal: &pb.SignedProposal{},
	}, 1, nil)

	if result != nil || err == nil || err.Error() != "peers and target peers is nil or empty" {
		t.Fatal("Test SendTransactionProposal failed, validation on peer is nil is not working as expected")
	}

	result, err = SendTransactionProposal(&apitxn.TransactionProposal{
		SignedProposal: &pb.SignedProposal{},
	}, 1, []apitxn.ProposalProcessor{})

	if result != nil || err == nil || err.Error() != "Missing peer objects for sending transaction proposal" {
		t.Fatal("Test SendTransactionProposal failed, validation on missing peer objects is not working")
	}

	peer := mocks.MockPeer{MockName: "Peer1", MockURL: "http://peer1.com", MockRoles: []string{}, MockCert: nil}
	channel.AddPeer(&peer)

	err = channel.AddPeer(&peer)

	if err == nil || err.Error() != "Peer with URL http://peer1.com already exists" {
		t.Fatal("Duplicate Peer check is not working as expected")
	}

	result, err = channel.SendTransactionProposal(&apitxn.TransactionProposal{
		SignedProposal: nil,
	}, 1, nil)

	if result != nil || err == nil || err.Error() != "signedProposal is nil" {
		t.Fatal("Test SendTransactionProposal failed, validation on signedProposal is nil is not working as expected")
	}

	result, err = SendTransactionProposal(&apitxn.TransactionProposal{
		SignedProposal: nil,
	}, 1, nil)

	if result != nil || err == nil || err.Error() != "signedProposal is nil" {
		t.Fatal("Test SendTransactionProposal failed, validation on signedProposal is nil is not working as expected")
	}

	targetPeer := mocks.MockPeer{MockName: "Peer2", MockURL: "http://peer2.com", MockRoles: []string{}, MockCert: nil}

	channel.AddPeer(&targetPeer)
	result, err = channel.SendTransactionProposal(&apitxn.TransactionProposal{
		SignedProposal: &pb.SignedProposal{},
	}, 1, targets)

	if result == nil || err != nil {
		t.Fatalf("Test SendTransactionProposal failed, with error '%s'", err.Error())
	}

}

func TestConcurrentPeers(t *testing.T) {
	const numPeers = 10000
	channel, err := setupMassiveTestChannel(numPeers, 0)
	if err != nil {
		t.Fatalf("Failed to create massive channel: %s", err)
	}

	result, err := channel.SendTransactionProposal(&apitxn.TransactionProposal{
		SignedProposal: &pb.SignedProposal{},
	}, 1, nil)
	if err != nil {
		t.Fatalf("SendTransactionProposal return error: %s", err)
	}

	if len(result) != numPeers {
		t.Error("SendTransactionProposal returned an unexpected amount of responses")
	}

	//Negative scenarios
	_, err = channel.SendTransactionProposal(nil, 1, nil)

	if err == nil || err.Error() != "signedProposal is nil" {
		t.Fatal("nil signedProposal validation check not working as expected")
	}

}

func startEndorserServer(t *testing.T, grpcServer *grpc.Server) (*mocks.MockEndorserServer, string) {
	lis, err := net.Listen("tcp", testAddress)
	addr := lis.Addr().String()

	endorserServer := &mocks.MockEndorserServer{}
	pb.RegisterEndorserServer(grpcServer, endorserServer)
	if err != nil {
		fmt.Printf("Error starting test server %s", err)
		t.FailNow()
	}
	fmt.Printf("Starting test server on %s\n", addr)
	go grpcServer.Serve(lis)
	return endorserServer, addr
}
