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
	"testing"
	"time"

	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/fab"
	"github.com/hyperledger/fabric-sdk-go/pkg/context"
	"github.com/hyperledger/fabric-sdk-go/pkg/fab/mocks"
	mspmocks "github.com/hyperledger/fabric-sdk-go/pkg/msp/test/mockmsp"
	"github.com/hyperledger/fabric-protos-go/common"
	pb "github.com/hyperledger/fabric-protos-go/peer"
)

func TestNewTransaction(t *testing.T) {

	txnReq := fab.TransactionRequest{
		Proposal:          &fab.TransactionProposal{Proposal: &pb.Proposal{}},
		ProposalResponses: []*fab.TransactionProposalResponse{},
	}
	//Test Empty proposal response scenario
	_, err := New(txnReq)
	require.Error(t, err, "Proposal response was supposed to fail in Create Transaction, for empty proposal response scenario")
	require.Equalf(t, "at least one proposal response is necessary", err.Error(), "Proposal response was supposed to fail in Create Transaction, for empty proposal response scenario, got: \n \"%s\"", err.Error())

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
	require.Error(t, err, "Proposal response was supposed to fail in Create Transaction, invalid proposal header scenario")
	require.Containsf(t, err.Error(), "unmarshal", "Proposal response was supposed to fail in Create Transaction, invalid proposal header scenario - %s", err.Error())

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
	require.Error(t, err, "Proposal response was supposed to fail in Create Transaction, invalid proposal payload scenario")
	require.Containsf(t, err.Error(), "unmarshal", "Proposal response was supposed to fail in Create Transaction, invalid proposal payload scenario - %s", err.Error())

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
	require.Error(t, err, "Proposal response was supposed to fail in Create Transaction")
	require.Equalf(t, "proposal response was not successful, error code 99, msg success", err.Error(), "Proposal response was supposed to fail in Create Transaction, got: \n \"%s\"", err.Error())

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
	require.Error(t, err, "Proposal response was supposed to fail in Create Transaction")
	require.Containsf(t, err.Error(), "proto: repeated field Endorsements has nil element", "Proposal response was supposed to fail in Create Transaction - %s", err.Error())
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
	require.NoErrorf(t, err, "Test Broadcast Envelope Failed, resp: %+v", res)

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
	require.Equalf(t, 1, firstSelected+secondSelected, "Both or none orderers were selected for broadcast: %d", firstSelected+secondSelected)

	// Now make 1 of them fail and repeatedly broadcast
	broadcastCount := 50
	for i := 0; i < broadcastCount; i++ {
		orderer1.EnqueueSendBroadcastError(errors.New("Service Unavailable"))
	}
	// It should always succeed even though one of them has failed
	for i := 0; i < broadcastCount; i++ {
		resp, err1 := broadcastEnvelope(reqCtx, sigEnvelope, orderers)
		require.NoErrorf(t, err1, "Test Broadcast Envelope Failed, resp: %+v", resp)
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
		require.Contains(t, err1.Error(), "Service Unavailable", "Test Broadcast failed but didn't return the correct reason")
	}
	emptyOrderers := []fab.Orderer{}
	_, err := broadcastEnvelope(reqCtx, sigEnvelope, emptyOrderers)
	require.Error(t, err, "Test empty orderers slice validation on broadcast envelope is not working as expected")
	require.Equalf(t, "orderers not set", err.Error(), "Test empty orderers slice validation on broadcast envelope is not working as expected, got: \n \"%s\"", err.Error())
}

func TestBroadcastPayloadWithOrdererDialFailure(t *testing.T) {
	ordererAddr := "127.0.0.1:0"
	//Create mock orderers
	orderer1 := mocks.NewMockGrpcOrderer(ordererAddr, nil)
	orderer1.Start()
	orderer2 := mocks.NewMockGrpcOrderer(ordererAddr, nil)
	orderer2.Start()
	orderer3 := mocks.NewMockGrpcOrderer(ordererAddr, nil)
	orderer3.Start()

	orderers := []fab.Orderer{orderer1, orderer2, orderer3}

	sigEnvelope := &fab.SignedEnvelope{
		Signature: []byte(""),
		Payload:   []byte(""),
	}
	user := mspmocks.NewMockSigningIdentity("test", "1234")
	ctx := mocks.NewMockContext(user)
	parentCtx, cancel := context.NewRequest(ctx, context.WithTimeout(5*time.Second)) // parentContext has 5 sec timeout
	defer cancel()

	_, err := broadcastEnvelope(parentCtx, sigEnvelope, orderers)
	require.NoError(t, err, "BroadCastEnvelope to running orderers returned a connection error")

	// stop orderer2 and try again (orderer1 and orderer3 should successfully connect)
	orderer2.Stop()
	_, err = broadcastEnvelope(parentCtx, sigEnvelope, orderers)
	require.NoError(t, err, "BroadCastEnvelope to running orderer1 and orderer3 returned a connection error")

	// stop orderer1 and try again (only orderer3 should successfully connect)
	orderer1.Stop()
	_, err = broadcastEnvelope(parentCtx, sigEnvelope, orderers)
	require.NoError(t, err, "BroadCastEnvelope to running orderer3 returned a connection error")

	// now try a new parent context using 1 nano second timeout to force 'context deadline exceeded'
	orderer1.Start()
	orderer2.Start()
	parentCtx, cancel2 := context.NewRequest(ctx, context.WithTimeout(1*time.Nanosecond))
	defer cancel2()
	_, err = broadcastEnvelope(parentCtx, sigEnvelope, orderers)
	require.Error(t, err, "BroadCastEnvelope to running orderers returned no error with 1 nano second context deadline")

	orderer1.Stop()
	orderer2.Stop()
	orderer3.Stop()
}

func TestSendTransaction(t *testing.T) {
	//Setup channel
	user := mspmocks.NewMockSigningIdentity("test", "1234")
	ctx := mocks.NewMockContext(user)

	reqCtx, cancel := context.NewRequest(ctx, context.WithTimeout(10*time.Second))
	defer cancel()

	response, err := Send(reqCtx, nil, nil)

	//Expect orderer is nil error
	require.Nil(t, response, "Test SendTransaction failed, it was supposed to fail with 'orderers is nil' error")
	require.Error(t, err, "Test SendTransaction failed, it was supposed to fail with 'orderers is nil' error")
	require.Equalf(t, "orderers is nil", err.Error(), "Test SendTransaction failed, it was supposed to fail with 'orderers is nil' error, got: \n \"%s\"", err.Error())

	//Create mock orderer
	orderer := mocks.NewMockOrderer("", nil)
	orderers := []fab.Orderer{orderer}

	//Call Send Transaction with nil tx
	response, err = Send(reqCtx, nil, orderers)

	//Expect tx is nil error
	require.Nil(t, response, "Test SendTransaction failed, it was supposed to fail with 'transaction is nil' error")
	require.Error(t, err, "Test SendTransaction failed, it was supposed to fail with 'transaction is nil' error")
	require.Equalf(t, "transaction is nil", err.Error(), "Test SendTransaction failed, it was supposed to fail with 'transaction is nil' error, got: \n \"%s\"", err.Error())

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
	require.Nil(t, response, "Test SendTransaction failed, it was supposed to fail with 'proposal is nil' error")
	require.Error(t, err, "Test SendTransaction failed, it was supposed to fail with 'proposal is nil' error")
	require.Equalf(t, "proposal is nil", err.Error(), "Test SendTransaction failed, it was supposed to fail with 'proposal is nil' error, got: \n \"%s\"", err.Error())

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
	require.Nil(t, response, "Test SendTransaction failed, it was supposed to fail with '...unmarshal...' error")
	require.Error(t, err, "Test SendTransaction failed, it was supposed to fail with '...unmarshal...' error")
	require.Containsf(t, err.Error(), "unmarshal", "Test SendTransaction failed with a wrong error, got: \n \"%s\"", err.Error())

	//Create tx with proper proposal header
	txn = fab.Transaction{
		Proposal: &fab.TransactionProposal{
			Proposal: &pb.Proposal{Header: []byte(""), Payload: []byte(""), Extension: []byte("")},
		},
		Transaction: &pb.Transaction{},
	}
	//Call Send Transaction
	response, err = Send(reqCtx, &txn, orderers)
	require.NotNil(t, response, "Test valid SendTransaction did not return a valid response")
	require.NoError(t, err, "Test valid SendTransaction failed")
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
	require.NotNil(t, header, "Test Build Channel returned an empty Header")
	require.NoError(t, err, "Test Build Channel Header failed")
}

func TestSignPayload(t *testing.T) {
	user := mspmocks.NewMockSigningIdentity("test", "1234")
	ctx := mocks.NewMockContext(user)

	payload := common.Payload{}

	signedEnv, err := signPayload(ctx, &payload)
	require.NotNil(t, signedEnv, "Test Sign Payload returned an empty signature")
	require.NoError(t, err, "Test Sign Payload failed")
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
	require.NoError(t, err, "SendTransaction returned error")
}

func setupMassiveTestOrderer(numberOfOrderers int) []fab.Orderer {
	orderers := []fab.Orderer{}

	for i := 0; i < numberOfOrderers; i++ {
		orderer := mocks.NewMockOrderer(fmt.Sprintf("http://mock%d.orderers.r.us", i), nil)
		orderers = append(orderers, orderer)
	}

	return orderers
}
