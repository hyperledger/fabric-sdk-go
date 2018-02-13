/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/
package txn

import (
	"fmt"
	"reflect"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/golang/protobuf/proto"
	"github.com/stretchr/testify/assert"

	fab "github.com/hyperledger/fabric-sdk-go/api/apifabclient"
	"github.com/hyperledger/fabric-sdk-go/api/apifabclient/mocks"

	"github.com/hyperledger/fabric-sdk-go/pkg/errors/multi"
	"github.com/hyperledger/fabric-sdk-go/pkg/fabric-client/mocks"
	pb "github.com/hyperledger/fabric-sdk-go/third_party/github.com/hyperledger/fabric/protos/peer"
)

const (
	testChannel = "testchannel"
	testAddress = "127.0.0.1:0"
)

func TestNewTransactionProposal(t *testing.T) {
	user := mocks.NewMockUserWithMSPID("test", "1234")
	ctx := mocks.NewMockContext(user)

	request := fab.ChaincodeInvokeRequest{
		ChaincodeID: "qscc",
		Fcn:         "Hello",
	}

	tp, err := NewProposal(ctx, testChannel, request)
	if err != nil {
		t.Fatalf("Create Transaction Proposal Failed: %s", err)
	}

	_, err = proto.Marshal(tp.SignedProposal)

	if err != nil {
		t.Fatalf("Call to proposal bytes failed: %s", err)
	}

}

func TestSendTransactionProposal(t *testing.T) {
	user := mocks.NewMockUserWithMSPID("test", "1234")
	ctx := mocks.NewMockContext(user)
	responseMessage := "success"

	peer := mocks.MockPeer{MockName: "Peer1", MockURL: "http://peer1.com",
		MockRoles: []string{}, MockCert: nil, Status: 200, Payload: []byte("A"),
		ResponseMessage: responseMessage}

	request := fab.ChaincodeInvokeRequest{
		ChaincodeID: "cc",
		Fcn:         "Hello",
		Args:        [][]byte{[]byte{1, 2, 3}},
	}

	tp, err := NewProposal(ctx, testChannel, request)
	if err != nil {
		t.Fatalf("new transaction proposal failed: %s", err)
	}

	tpr, err := SendProposal(tp, []fab.ProposalProcessor{&peer})
	if err != nil {
		t.Fatalf("send transaction proposal failed: %s", err)
	}

	expectedTpr := &pb.ProposalResponse{Response: &pb.Response{Message: responseMessage, Status: 200, Payload: []byte("A")}}

	if !reflect.DeepEqual(tpr[0].ProposalResponse.Response, expectedTpr.Response) {
		t.Fatalf("Unexpected transaction proposal response: %v, %v", tpr, tp.TxnID)
	}
}

func TestNewTransactionProposalParams(t *testing.T) {
	user := mocks.NewMockUserWithMSPID("test", "1234")
	ctx := mocks.NewMockContext(user)

	request := fab.ChaincodeInvokeRequest{
		ChaincodeID: "cc",
		Fcn:         "Hello",
	}

	tp, err := NewProposal(ctx, testChannel, request)
	if err != nil {
		t.Fatalf("new transaction proposal failed: %s", err)
	}

	_, err = SendProposal(tp, nil)
	if err == nil {
		t.Fatalf("Expected error")
	}

	request = fab.ChaincodeInvokeRequest{
		Fcn: "Hello",
	}

	tp, err = NewProposal(ctx, testChannel, request)
	if err == nil {
		t.Fatalf("Expected error")
	}

	request = fab.ChaincodeInvokeRequest{
		ChaincodeID: "cc",
	}

	tp, err = NewProposal(ctx, testChannel, request)
	if err == nil {
		t.Fatalf("Expected error")
	}

	request = fab.ChaincodeInvokeRequest{
		ChaincodeID: "cc",
		Fcn:         "Hello",
	}
	tp, err = NewProposal(ctx, testChannel, request)
	if err != nil {
		t.Fatalf("new transaction proposal failed: %s", err)
	}
}

func TestConcurrentPeers(t *testing.T) {
	const numPeers = 10000
	peers := setupMassiveTestPeers(numPeers)

	result, err := SendProposal(&fab.TransactionProposal{
		SignedProposal: &pb.SignedProposal{},
	}, peers)
	if err != nil {
		t.Fatalf("SendProposal return error: %s", err)
	}

	if len(result) != numPeers {
		t.Error("SendTransactionProposal returned an unexpected amount of responses")
	}

	//Negative scenarios
	_, err = SendProposal(nil, nil)

	if err == nil || err.Error() != "signedProposal is required" {
		t.Fatal("nil signedProposal validation check not working as expected")
	}
}

func TestSendTransactionProposalToProcessors(t *testing.T) {

	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()
	proc := mock_apifabclient.NewMockProposalProcessor(mockCtrl)

	tp := fab.TransactionProposal{
		SignedProposal: &pb.SignedProposal{},
	}
	tpr := fab.TransactionProposalResponse{Endorser: "example.com", Status: 99, Proposal: tp, ProposalResponse: nil}
	proc.EXPECT().ProcessTransactionProposal(tp).Return(tpr, nil)
	targets := []fab.ProposalProcessor{proc}

	result, err := SendProposal(&fab.TransactionProposal{
		SignedProposal: &pb.SignedProposal{},
	}, nil)

	if result != nil || err == nil || err.Error() != "targets is required" {
		t.Fatalf("Test SendTransactionProposal failed, validation on peer is nil is not working as expected: %v", err)
	}

	result, err = SendProposal(&fab.TransactionProposal{
		SignedProposal: &pb.SignedProposal{},
	}, []fab.ProposalProcessor{})

	if result != nil || err == nil || err.Error() != "targets is required" {
		t.Fatalf("Test SendTransactionProposal failed, validation on missing peer objects is not working: %v", err)
	}

	result, err = SendProposal(&fab.TransactionProposal{
		SignedProposal: nil,
	}, nil)

	if result != nil || err == nil || err.Error() != "signedProposal is required" {
		t.Fatal("Test SendTransactionProposal failed, validation on signedProposal is nil is not working as expected")
	}

	result, err = SendProposal(&fab.TransactionProposal{
		SignedProposal: &pb.SignedProposal{},
	}, targets)

	if result == nil || err != nil {
		t.Fatalf("Test SendTransactionProposal failed, with error '%s'", err.Error())
	}
}

func TestProposalResponseError(t *testing.T) {
	testError := fmt.Errorf("Test Error")

	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()
	proc := mock_apifabclient.NewMockProposalProcessor(mockCtrl)
	proc2 := mock_apifabclient.NewMockProposalProcessor(mockCtrl)

	tp := fab.TransactionProposal{
		SignedProposal: &pb.SignedProposal{},
	}

	// Test with error from lower layer
	tpr := fab.TransactionProposalResponse{Endorser: "example.com", Status: 200,
		Proposal: tp, ProposalResponse: nil}
	proc.EXPECT().ProcessTransactionProposal(tp).Return(tpr, testError)
	proc2.EXPECT().ProcessTransactionProposal(tp).Return(tpr, testError)

	targets := []fab.ProposalProcessor{proc, proc2}
	_, err := SendProposal(&fab.TransactionProposal{
		SignedProposal: &pb.SignedProposal{},
	}, targets)
	errs, ok := err.(multi.Errors)
	assert.True(t, ok, "expected multi errors object")
	assert.Equal(t, testError, errs[0])
}

func setupMassiveTestPeers(numberOfPeers int) []fab.ProposalProcessor {
	peers := []fab.ProposalProcessor{}

	for i := 0; i < numberOfPeers; i++ {
		peer := mocks.MockPeer{MockName: fmt.Sprintf("MockPeer%d", i), MockURL: fmt.Sprintf("http://mock%d.peers.r.us", i),
			MockRoles: []string{}, MockCert: nil}
		peers = append(peers, &peer)
	}

	return peers
}
