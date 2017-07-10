/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/
package channel

import (
	"fmt"
	"net"
	"reflect"
	"testing"

	"google.golang.org/grpc"

	pb "github.com/hyperledger/fabric/protos/peer"

	fab "github.com/hyperledger/fabric-sdk-go/api/apifabclient"
	"github.com/hyperledger/fabric-sdk-go/api/apitxn"
	"github.com/hyperledger/fabric-sdk-go/pkg/fabric-client/internal/txnproc"
	"github.com/hyperledger/fabric-sdk-go/pkg/fabric-client/mocks"
	"github.com/hyperledger/fabric-sdk-go/pkg/fabric-client/peer"
)

func TestCreateTransactionProposal(t *testing.T) {
	channel, _ := setupTestChannel()

	request := apitxn.ChaincodeInvokeRequest{
		ChaincodeID: "qscc",
	}

	tProposal, err := newTransactionProposal(channel.Name(), request, channel.clientContext)
	if err != nil {
		t.Fatal("Create Transaction Proposal Failed", err)
	}

	_, errx := channel.ProposalBytes(tProposal)

	if errx != nil {
		t.Fatalf("Call to proposal bytes failed: %v", errx)
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
	txid, _ := channel.ClientContext().NewTxnID()

	badtxid1, _ := channel.ClientContext().NewTxnID()
	badtxid2, _ := channel.ClientContext().NewTxnID()

	badtxid1.ID = ""
	badtxid2.Nonce = nil

	genesisBlockReqeust := &fab.GenesisBlockRequest{
		TxnID: txid,
	}
	fmt.Printf("TxnID: %v", txid)

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
		TxnID:        badtxid1,
	}
	err = channel.JoinChannel(request)
	if err == nil {
		t.Fatalf("Should not have been able to join channel because of missing TxID parameter")
	}

	request = &fab.JoinChannelRequest{
		Targets:      peers,
		GenesisBlock: genesisBlock,
		TxnID:        badtxid2,
	}
	err = channel.JoinChannel(request)
	if err == nil {
		t.Fatalf("Should not have been able to join channel because of missing Nonce parameter")
	}

	request = &fab.JoinChannelRequest{
		Targets: peers,
		//GenesisBlock: genesisBlock,
		TxnID: txid,
	}
	err = channel.JoinChannel(request)
	if err == nil {
		t.Fatalf("Should not have been able to join channel because of missing GenesisBlock parameter")
	}

	request = &fab.JoinChannelRequest{
		//Targets: peers,
		GenesisBlock: genesisBlock,
		TxnID:        txid,
	}
	err = channel.JoinChannel(request)
	if err == nil {
		t.Fatalf("Should not have been able to join channel because of missing Targets parameter")
	}

	request = &fab.JoinChannelRequest{
		Targets:      peers,
		GenesisBlock: genesisBlock,
		TxnID:        txid,
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
	request = &fab.JoinChannelRequest{
		Targets: peers,
		TxnID:   txid,
	}
	err = channel.JoinChannel(request)
	if err == nil {
		t.Fatalf("Expected error")
	}
}

func TestAddPeerDuplicateCheck(t *testing.T) {
	channel, _ := setupTestChannel()

	peer := mocks.MockPeer{MockName: "Peer1", MockURL: "http://peer1.com", MockRoles: []string{}, MockCert: nil}
	channel.AddPeer(&peer)

	err := channel.AddPeer(&peer)

	if err == nil || err.Error() != "Peer with URL http://peer1.com already exists" {
		t.Fatal("Duplicate Peer check is not working as expected")
	}
}

func TestSendTransactionProposal(t *testing.T) {
	channel, _ := setupTestChannel()

	peer := mocks.MockPeer{MockName: "Peer1", MockURL: "http://peer1.com", MockRoles: []string{}, MockCert: nil, Payload: []byte("A")}
	channel.AddPeer(&peer)

	request := apitxn.ChaincodeInvokeRequest{
		ChaincodeID: "cc",
		Fcn:         "Hello",
	}
	tpr, txnid, err := channel.SendTransactionProposal(request)
	if err != nil {
		t.Fatalf("Failed to send transaction proposal: %s", err)
	}
	expectedTpr := &pb.ProposalResponse{Response: &pb.Response{Message: "success", Status: 99, Payload: []byte("A")}}

	if txnid.ID != "1234" || !reflect.DeepEqual(tpr[0].ProposalResponse, expectedTpr) {
		t.Fatalf("Unexpected transaction proposal response: %v, %v", tpr, txnid)
	}
}

func TestSendTransactionProposalMissingParams(t *testing.T) {
	channel, _ := setupTestChannel()

	request := apitxn.ChaincodeInvokeRequest{
		ChaincodeID: "cc",
		Fcn:         "Hello",
	}
	_, _, err := channel.SendTransactionProposal(request)
	if err == nil {
		t.Fatalf("Expected error")
	}

	peer := mocks.MockPeer{MockName: "Peer1", MockURL: "http://peer1.com", MockRoles: []string{}, MockCert: nil, Payload: []byte("A")}
	channel.AddPeer(&peer)

	request = apitxn.ChaincodeInvokeRequest{
		Fcn: "Hello",
	}
	_, _, err = channel.SendTransactionProposal(request)
	if err == nil {
		t.Fatalf("Expected error")
	}

	request = apitxn.ChaincodeInvokeRequest{
		ChaincodeID: "cc",
	}
	_, _, err = channel.SendTransactionProposal(request)
	if err == nil {
		t.Fatalf("Expected error")
	}

	request = apitxn.ChaincodeInvokeRequest{
		ChaincodeID: "cc",
		Fcn:         "Hello",
	}
	_, _, err = channel.SendTransactionProposal(request)
	if err != nil {
		t.Fatalf("Expected success")
	}
}

// TODO: Move test
func TestConcurrentPeers(t *testing.T) {
	const numPeers = 10000
	channel, err := setupMassiveTestChannel(numPeers, 0)
	if err != nil {
		t.Fatalf("Failed to create massive channel: %s", err)
	}

	result, err := txnproc.SendTransactionProposalToProcessors(&apitxn.TransactionProposal{
		SignedProposal: &pb.SignedProposal{},
	}, channel.txnProcessors())
	if err != nil {
		t.Fatalf("SendTransactionProposal return error: %s", err)
	}

	if len(result) != numPeers {
		t.Error("SendTransactionProposal returned an unexpected amount of responses")
	}

	//Negative scenarios
	_, err = txnproc.SendTransactionProposalToProcessors(nil, nil)

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
