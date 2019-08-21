/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/
package txn

import (
	"fmt"
	"reflect"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/golang/mock/gomock"
	"github.com/golang/protobuf/proto"
	"github.com/stretchr/testify/assert"

	"time"

	"github.com/hyperledger/fabric-sdk-go/pkg/common/errors/multi"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/fab"
	mock_context "github.com/hyperledger/fabric-sdk-go/pkg/common/providers/test/mockfab"
	"github.com/hyperledger/fabric-sdk-go/pkg/context"
	"github.com/hyperledger/fabric-sdk-go/pkg/fab/mocks"
	mspmocks "github.com/hyperledger/fabric-sdk-go/pkg/msp/test/mockmsp"
	pb "github.com/hyperledger/fabric-protos-go/peer"
)

const (
	testChannel = "testchannel"
)

func TestNewHeader(t *testing.T) {
	user := mspmocks.NewMockSigningIdentity("test", "1234")
	ctx := mocks.NewMockContext(user)

	creator, err := ctx.Serialize()
	require.NoError(t, err)

	txh, err := NewHeader(ctx, testChannel)
	require.NoError(t, err)
	require.NotEmptyf(t, txh.nonce, "Expecting nonce")
	require.Equal(t, creator, txh.creator)
	require.NotEmpty(t, txh.id)

	creator = []byte("someothercreator")
	nonce := []byte("123456")

	txh, err = NewHeader(ctx, testChannel, fab.WithCreator(creator), fab.WithNonce(nonce))
	require.NoError(t, err)
	require.Equal(t, nonce, txh.nonce)
	require.Equal(t, creator, txh.creator)
	require.NotEmpty(t, txh.id)
}

func TestNewTransactionProposal(t *testing.T) {
	user := mspmocks.NewMockSigningIdentity("test", "1234")
	ctx := mocks.NewMockContext(user)

	request := fab.ChaincodeInvokeRequest{
		ChaincodeID: "qscc",
		Fcn:         "Hello",
	}

	txh, err := NewHeader(ctx, testChannel)
	if err != nil {
		t.Fatalf("create transaction ID failed: %s", err)
	}

	tp, err := CreateChaincodeInvokeProposal(txh, request)
	if err != nil {
		t.Fatalf("Create Transaction Proposal Failed: %s", err)
	}

	signedProposal, err := signProposal(ctx, tp.Proposal)
	if err != nil {
		t.Fatalf("signProposal failed: %s", err)
	}

	_, err = proto.Marshal(signedProposal)
	if err != nil {
		t.Fatalf("Call to proposal bytes failed: %s", err)
	}
}

func TestSendTransactionProposal(t *testing.T) {
	user := mspmocks.NewMockSigningIdentity("test", "1234")
	ctx := mocks.NewMockContext(user)
	responseMessage := "success"

	peer := mocks.MockPeer{MockName: "Peer1", MockURL: "http://peer1.com",
		MockRoles: []string{}, MockCert: nil, Status: 200, Payload: []byte("A"),
		ResponseMessage: responseMessage}

	request := fab.ChaincodeInvokeRequest{
		ChaincodeID: "cc",
		Fcn:         "Hello",
		Args:        [][]byte{{1, 2, 3}},
	}

	txh, err := NewHeader(ctx, testChannel)
	if err != nil {
		t.Fatalf("create transaction ID failed: %s", err)
	}

	tp, err := CreateChaincodeInvokeProposal(txh, request)
	if err != nil {
		t.Fatalf("new transaction proposal failed: %s", err)
	}

	reqCtx, cancel := context.NewRequest(ctx, context.WithTimeout(10*time.Second))
	defer cancel()

	_, err = SendProposal(reqCtx, tp, []fab.ProposalProcessor{nil})
	if err == nil || !strings.Contains(err.Error(), "target is nil") {
		t.Fatalf("Should have failed due to nil target")
	}

	tpr, err := SendProposal(reqCtx, tp, []fab.ProposalProcessor{&peer})
	if err != nil {
		t.Fatalf("send transaction proposal failed: %s", err)
	}

	expectedTpr := &pb.ProposalResponse{Response: &pb.Response{Message: responseMessage, Status: 200, Payload: []byte("A")}}

	if !reflect.DeepEqual(tpr[0].ProposalResponse.Response, expectedTpr.Response) {
		t.Fatalf("Unexpected transaction proposal response: %v, %v", tpr, tp.TxnID)
	}
}

func TestNewTransactionProposalParams(t *testing.T) {
	user := mspmocks.NewMockSigningIdentity("test", "1234")
	ctx := mocks.NewMockContext(user)

	request := fab.ChaincodeInvokeRequest{
		ChaincodeID: "cc",
		Fcn:         "Hello",
	}

	txh, err := NewHeader(ctx, testChannel)
	if err != nil {
		t.Fatalf("create transaction ID failed: %s", err)
	}

	tp, err := CreateChaincodeInvokeProposal(txh, request)
	if err != nil {
		t.Fatalf("new transaction proposal failed: %s", err)
	}

	reqCtx, cancel := context.NewRequest(ctx, context.WithTimeout(10*time.Second))
	defer cancel()

	_, err = SendProposal(reqCtx, tp, nil)
	if err == nil {
		t.Fatal("Expected error")
	}

	request = fab.ChaincodeInvokeRequest{
		Fcn: "Hello",
	}

	_, err = CreateChaincodeInvokeProposal(txh, request)
	if err == nil {
		t.Fatal("Expected error")
	}

	request = fab.ChaincodeInvokeRequest{
		ChaincodeID: "cc",
	}

	_, err = CreateChaincodeInvokeProposal(txh, request)
	if err == nil {
		t.Fatal("Expected error")
	}

	request = fab.ChaincodeInvokeRequest{
		ChaincodeID: "cc",
		Fcn:         "Hello",
	}
	_, err = CreateChaincodeInvokeProposal(txh, request)
	if err != nil {
		t.Fatalf("new transaction proposal failed: %s", err)
	}
}

func TestConcurrentPeers(t *testing.T) {
	const numPeers = 10000
	peers := setupMassiveTestPeers(numPeers)

	user := mspmocks.NewMockSigningIdentity("test", "1234")
	ctx := mocks.NewMockContext(user)

	reqCtx, cancel := context.NewRequest(ctx, context.WithTimeout(10*time.Second))
	defer cancel()

	result, err := SendProposal(reqCtx, &fab.TransactionProposal{
		Proposal: &pb.Proposal{},
	}, peers)
	if err != nil {
		t.Fatalf("SendProposal return error: %s", err)
	}

	if len(result) != numPeers {
		t.Error("SendTransactionProposal returned an unexpected amount of responses")
	}
}

func TestSendTransactionProposalToProcessors(t *testing.T) {

	user := mspmocks.NewMockSigningIdentity("test", "1234")
	ctx := mocks.NewMockContext(user)

	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()
	proc := mock_context.NewMockProposalProcessor(mockCtrl)

	stp, err := signProposal(ctx, &pb.Proposal{})
	if err != nil {
		t.Fatalf("signProposal returned error: %s", err)
	}
	tp := fab.ProcessProposalRequest{
		SignedProposal: stp,
	}

	tpr := fab.TransactionProposalResponse{Endorser: "example.com", Status: 99}
	proc.EXPECT().ProcessTransactionProposal(gomock.Any(), tp).Return(&tpr, nil)
	targets := []fab.ProposalProcessor{proc}

	reqCtx, cancel := context.NewRequest(ctx, context.WithTimeout(10*time.Second))
	defer cancel()

	result, err := SendProposal(reqCtx, &fab.TransactionProposal{
		Proposal: &pb.Proposal{},
	}, nil)

	if result != nil || err == nil || err.Error() != "targets is required" {
		t.Fatalf("Test SendTransactionProposal failed, validation on peer is nil is not working as expected: %s", err)
	}

	result, err = SendProposal(reqCtx, &fab.TransactionProposal{
		Proposal: &pb.Proposal{},
	}, []fab.ProposalProcessor{})

	if result != nil || err == nil || err.Error() != "targets is required" {
		t.Fatalf("Test SendTransactionProposal failed, validation on missing peer objects is not working: %s", err)
	}

	result, err = SendProposal(reqCtx, &fab.TransactionProposal{
		Proposal: &pb.Proposal{}}, targets)

	if result == nil || err != nil {
		t.Fatalf("Test SendTransactionProposal failed, with error '%s'", err)
	}
}

func TestProposalResponseError(t *testing.T) {
	testError := fmt.Errorf("Test Error")

	user := mspmocks.NewMockSigningIdentity("test", "1234")
	ctx := mocks.NewMockContext(user)

	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()
	proc := mock_context.NewMockProposalProcessor(mockCtrl)
	proc2 := mock_context.NewMockProposalProcessor(mockCtrl)

	stp, err := signProposal(ctx, &pb.Proposal{})
	if err != nil {
		t.Fatalf("signProposal returned error: %s", err)
	}
	tp := fab.ProcessProposalRequest{
		SignedProposal: stp,
	}

	// Test with error from lower layer
	tpr := fab.TransactionProposalResponse{Endorser: "example.com", Status: 200}
	proc.EXPECT().ProcessTransactionProposal(gomock.Any(), tp).Return(&tpr, testError)
	proc2.EXPECT().ProcessTransactionProposal(gomock.Any(), tp).Return(&tpr, testError)

	reqCtx, cancel := context.NewRequest(ctx, context.WithTimeout(10*time.Second))
	defer cancel()

	targets := []fab.ProposalProcessor{proc, proc2}
	_, err = SendProposal(reqCtx, &fab.TransactionProposal{
		Proposal: &pb.Proposal{},
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
