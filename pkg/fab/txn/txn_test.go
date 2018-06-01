/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package txn

import (
	reqContext "context"
	"fmt"
	"os"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"

	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/fab"
	"github.com/hyperledger/fabric-sdk-go/pkg/context"
	"github.com/hyperledger/fabric-sdk-go/pkg/fab/mocks"
	mspmocks "github.com/hyperledger/fabric-sdk-go/pkg/msp/test/mockmsp"
	"github.com/hyperledger/fabric-sdk-go/third_party/github.com/hyperledger/fabric/protos/common"
	pb "github.com/hyperledger/fabric-sdk-go/third_party/github.com/hyperledger/fabric/protos/peer"
)

func TestNewTransaction(t *testing.T) {

	txnReq := fab.TransactionRequest{
		Proposal:          &fab.TransactionProposal{Proposal: &pb.Proposal{}},
		ProposalResponses: []*fab.TransactionProposalResponse{},
	}
	//Test Empty proposal response scenario
	_, err := New(txnReq)

	if err == nil || err.Error() != "at least one proposal response is necessary" {
		t.Fatal("Proposal response was supposed to fail in Create Transaction, for empty proposal response scenario")
	}

	//Test invalid proposal header scenario

	th := TransactionHeader{
		id: "1234",
	}

	proposal := fab.TransactionProposal{
		TxnID:    fab.TransactionID(th.id),
		Proposal: &pb.Proposal{Header: []byte("TEST"), Extension: []byte(""), Payload: []byte("")},
	}

	proposalResp := fab.TransactionProposalResponse{
		Endorser:         "http://peer1.com",
		ProposalResponse: &pb.ProposalResponse{Response: &pb.Response{Message: "success", Status: 99, Payload: []byte("")}},
	}

	txnReq = fab.TransactionRequest{
		Proposal:          &proposal,
		ProposalResponses: []*fab.TransactionProposalResponse{&proposalResp},
	}
	_, err = New(txnReq)

	if err == nil || !strings.Contains(err.Error(), "unmarshal") {
		t.Fatal("Proposal response was supposed to fail in Create Transaction, invalid proposal header scenario")
	}

	//Test invalid proposal payload scenario
	proposal = fab.TransactionProposal{
		TxnID:    fab.TransactionID(th.id),
		Proposal: &pb.Proposal{Header: []byte(""), Extension: []byte(""), Payload: []byte("TEST")},
	}

	proposalResp = fab.TransactionProposalResponse{
		Endorser:         "http://peer1.com",
		ProposalResponse: &pb.ProposalResponse{Response: &pb.Response{Message: "success", Status: 99, Payload: []byte("")}},
	}

	txnReq = fab.TransactionRequest{
		Proposal:          &proposal,
		ProposalResponses: []*fab.TransactionProposalResponse{&proposalResp},
	}
	_, err = New(txnReq)
	if err == nil || !strings.Contains(err.Error(), "unmarshal") {
		t.Fatal("Proposal response was supposed to fail in Create Transaction, invalid proposal payload scenario")
	}

	//Test proposal response
	proposal = fab.TransactionProposal{
		TxnID:    fab.TransactionID(th.id),
		Proposal: &pb.Proposal{Header: []byte(""), Extension: []byte(""), Payload: []byte("")},
	}

	proposalResp = fab.TransactionProposalResponse{
		Endorser:         "http://peer1.com",
		ProposalResponse: &pb.ProposalResponse{Response: &pb.Response{Message: "success", Status: 99, Payload: []byte("")}},
	}

	txnReq = fab.TransactionRequest{
		Proposal:          &proposal,
		ProposalResponses: []*fab.TransactionProposalResponse{&proposalResp},
	}
	_, err = New(txnReq)
	if err == nil || err.Error() != "proposal response was not successful, error code 99, msg success" {
		t.Fatal("Proposal response was supposed to fail in Create Transaction")
	}

	//Test repeated field header nil scenario
	checkRepeatedFieldHeader(proposal, th, proposalResp, txnReq, t)

	//TODO: Need actual sample payload for success case

}

func checkRepeatedFieldHeader(proposal fab.TransactionProposal, th TransactionHeader, proposalResp fab.TransactionProposalResponse, txnReq fab.TransactionRequest, t *testing.T) {
	proposal = fab.TransactionProposal{
		TxnID:    fab.TransactionID(th.id),
		Proposal: &pb.Proposal{Header: []byte(""), Extension: []byte(""), Payload: []byte("")},
	}
	proposalResp = fab.TransactionProposalResponse{
		Endorser:         "http://peer1.com",
		ProposalResponse: &pb.ProposalResponse{Response: &pb.Response{Message: "success", Status: 200, Payload: []byte("")}},
	}
	txnReq = fab.TransactionRequest{
		Proposal:          &proposal,
		ProposalResponses: []*fab.TransactionProposalResponse{&proposalResp},
	}
	_, err := New(txnReq)
	if err == nil || err.Error() != "repeated field endorsements has nil element" {
		t.Fatal("Proposal response was supposed to fail in Create Transaction")
	}
}

func TestBroadcastEnvelope(t *testing.T) {
	user := mspmocks.NewMockSigningIdentity("test", "1234")
	ctx := mocks.NewMockContext(user)

	lsnr1 := make(chan *fab.SignedEnvelope)
	lsnr2 := make(chan *fab.SignedEnvelope)
	//Create mock orderers
	orderer1 := mocks.NewMockOrderer("1", lsnr1)
	orderer2 := mocks.NewMockOrderer("2", lsnr2)

	orderers := []fab.Orderer{orderer1, orderer2}

	sigEnvelope := &fab.SignedEnvelope{
		Signature: []byte(""),
		Payload:   []byte(""),
	}

	reqCtx, cancel := context.NewRequest(ctx, context.WithTimeout(10*time.Second))
	defer cancel()

	res, err := broadcastEnvelope(reqCtx, sigEnvelope, orderers)

	if err != nil {
		t.Fatalf("Test Broadcast Envelope Failed, cause %s %+v", err, res)
	}

	// Ensure only 1 orderer was selected for broadcast
	firstSelected := 0
	secondSelected := 0
	for i := 0; i < 2; i++ {
		select {
		case <-lsnr1:
			firstSelected = 1
		case <-lsnr2:
			secondSelected = 1
		case <-time.After(time.Second):
		}
	}

	if firstSelected+secondSelected != 1 {
		t.Fatalf("Both or none orderers were selected for broadcast: %d", firstSelected+secondSelected)
	}

	// Now make 1 of them fail and repeatedly broadcast
	broadcastCount := 50
	for i := 0; i < broadcastCount; i++ {
		orderer1.EnqueueSendBroadcastError(errors.New("Service Unavailable"))
	}
	// It should always succeed even though one of them has failed
	for i := 0; i < broadcastCount; i++ {
		if res, err1 := broadcastEnvelope(reqCtx, sigEnvelope, orderers); err1 != nil {
			t.Fatalf("Test Broadcast Envelope Failed, cause %s %+v", err1, res)
		}
	}

	// Now, fail both and ensure any attempt fails
	checkBroadcastCount(broadcastCount, orderer1, orderer2, reqCtx, sigEnvelope, orderers, t)
}

func checkBroadcastCount(broadcastCount int, orderer1 *mocks.MockOrderer, orderer2 *mocks.MockOrderer, reqCtx reqContext.Context, sigEnvelope *fab.SignedEnvelope, orderers []fab.Orderer, t *testing.T) {
	for i := 0; i < broadcastCount; i++ {
		orderer1.EnqueueSendBroadcastError(errors.New("Service Unavailable"))
		orderer2.EnqueueSendBroadcastError(errors.New("Service Unavailable"))
	}
	for i := 0; i < broadcastCount; i++ {
		_, err1 := broadcastEnvelope(reqCtx, sigEnvelope, orderers)
		if !strings.Contains(err1.Error(), "Service Unavailable") {
			t.Fatal("Test Broadcast failed but didn't return the correct reason(should contain 'Service Unavailable')")
		}
	}
	emptyOrderers := []fab.Orderer{}
	_, err := broadcastEnvelope(reqCtx, sigEnvelope, emptyOrderers)
	if err == nil || err.Error() != "orderers not set" {
		t.Fatal("orderers not set validation on broadcast envelope is not working as expected")
	}
}

func TestSendTransaction(t *testing.T) {
	//Setup channel
	user := mspmocks.NewMockSigningIdentity("test", "1234")
	ctx := mocks.NewMockContext(user)

	reqCtx, cancel := context.NewRequest(ctx, context.WithTimeout(10*time.Second))
	defer cancel()

	response, err := Send(reqCtx, nil, nil)

	//Expect orderer is nil error
	if response != nil || err == nil || err.Error() != "orderers is nil" {
		t.Fatal("Test SendTransaction failed, it was supposed to fail with 'orderers is nil' error")
	}

	//Create mock orderer
	orderer := mocks.NewMockOrderer("", nil)
	orderers := []fab.Orderer{orderer}

	//Call Send Transaction with nil tx
	response, err = Send(reqCtx, nil, orderers)

	//Expect tx is nil error
	if response != nil || err == nil || err.Error() != "transaction is nil" {
		t.Fatal("Test SendTransaction failed, it was supposed to fail with 'transaction is nil' error")
	}

	testSendTransaction(reqCtx, orderers, t)
}

func testSendTransaction(reqCtx reqContext.Context, orderers []fab.Orderer, t *testing.T) {
	//Create tx with nil proposal
	txn := fab.Transaction{
		Proposal: &fab.TransactionProposal{
			Proposal: nil,
		},
		Transaction: &pb.Transaction{},
	}
	//Call Send Transaction with nil proposal
	response, err := Send(reqCtx, &txn, orderers)
	//Expect proposal is nil error
	if response != nil || err == nil || err.Error() != "proposal is nil" {
		t.Fatal("Test SendTransaction failed, it was supposed to fail with 'proposal is nil' error")
	}
	//Create tx with improper proposal header
	txn = fab.Transaction{
		Proposal: &fab.TransactionProposal{
			Proposal: &pb.Proposal{Header: []byte("TEST")},
		},
		Transaction: &pb.Transaction{},
	}
	//Call Send Transaction
	response, err = Send(reqCtx, &txn, orderers)
	//Expect header unmarshal error
	if response != nil || err == nil || !strings.Contains(err.Error(), "unmarshal") {
		t.Fatal("Test SendTransaction failed, it was supposed to fail with '...unmarshal...' error")
	}
	//Create tx with proper proposal header
	txn = fab.Transaction{
		Proposal: &fab.TransactionProposal{
			Proposal: &pb.Proposal{Header: []byte(""), Payload: []byte(""), Extension: []byte("")},
		},
		Transaction: &pb.Transaction{},
	}
	//Call Send Transaction
	response, err = Send(reqCtx, &txn, orderers)
	if response == nil || err != nil {
		t.Fatalf("Test SendTransaction failed, reason : '%s'", err)
	}
}

func TestBuildChannelHeader(t *testing.T) {
	user := mspmocks.NewMockSigningIdentity("test", "1234")
	ctx := mocks.NewMockContext(user)

	txnid, err := NewHeader(ctx, "test")
	assert.Nil(t, err, "NewID failed")

	o := ChannelHeaderOpts{
		Epoch:       1,
		ChaincodeID: "1234",
		TxnHeader:   txnid,
	}
	header, err := CreateChannelHeader(common.HeaderType_CHAINCODE_PACKAGE, o)

	if err != nil || header == nil {
		t.Fatalf("Test Build Channel Header failed, cause : '%s'", err)
	}

}

func TestSignPayload(t *testing.T) {
	user := mspmocks.NewMockSigningIdentity("test", "1234")
	ctx := mocks.NewMockContext(user)

	payload := common.Payload{}

	signedEnv, err := signPayload(ctx, &payload)

	if err != nil || signedEnv == nil {
		t.Fatal("Test Sign Payload Failed")
	}
}

func TestConcurrentOrderers(t *testing.T) {
	user := mspmocks.NewMockSigningIdentity("test", "1234")
	ctx := mocks.NewMockContext(user)

	// Determine number of orderers to use - environment can override
	const numOrderersDefault = 2000
	numOrderersEnv := os.Getenv("TEST_MASSIVE_ORDERER_COUNT")
	numOrderers, err := strconv.Atoi(numOrderersEnv)
	if err != nil {
		numOrderers = numOrderersDefault
	}

	orderers := setupMassiveTestOrderer(numOrderers)

	txn := fab.Transaction{
		Proposal: &fab.TransactionProposal{
			Proposal: &pb.Proposal{},
		},
		Transaction: &pb.Transaction{},
	}

	reqCtx, cancel := context.NewRequest(ctx, context.WithTimeout(10*time.Second))
	defer cancel()

	_, err = Send(reqCtx, &txn, orderers)
	if err != nil {
		t.Fatalf("SendTransaction returned error: %s", err)
	}
}

func setupMassiveTestOrderer(numberOfOrderers int) []fab.Orderer {
	orderers := []fab.Orderer{}

	for i := 0; i < numberOfOrderers; i++ {
		orderer := mocks.NewMockOrderer(fmt.Sprintf("http://mock%d.orderers.r.us", i), nil)
		orderers = append(orderers, orderer)
	}

	return orderers
}
