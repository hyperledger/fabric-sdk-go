/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package channel

import (
	"crypto/rand"
	"os"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/golang/mock/gomock"

	fab "github.com/hyperledger/fabric-sdk-go/api/apifabclient"
	"github.com/hyperledger/fabric-sdk-go/api/apitxn"
	"github.com/hyperledger/fabric-sdk-go/api/apitxn/mocks"
	"github.com/hyperledger/fabric-sdk-go/third_party/github.com/hyperledger/fabric/protos/common"
	pb "github.com/hyperledger/fabric-sdk-go/third_party/github.com/hyperledger/fabric/protos/peer"

	"github.com/hyperledger/fabric-sdk-go/pkg/errors"
	"github.com/hyperledger/fabric-sdk-go/pkg/fabric-client/mocks"
	"github.com/hyperledger/fabric-sdk-go/third_party/github.com/hyperledger/fabric/common/cauthdsl"
)

func TestCreateTransaction(t *testing.T) {
	channel, _ := setupTestChannel()

	peer := mocks.MockPeer{MockName: "Peer1", MockURL: "http://peer1.com", MockRoles: []string{}, MockCert: nil}
	channel.AddPeer(&peer)

	//Test Empty proposal response scenario
	_, err := channel.CreateTransaction([]*apitxn.TransactionProposalResponse{})

	if err == nil || err.Error() != "at least one proposal response is necessary" {
		t.Fatal("Proposal response was supposed to fail in Create Transaction, for empty proposal response scenario")
	}

	//Test invalid proposal header scenario

	txid := apitxn.TransactionID{
		ID: "1234",
	}

	test := &apitxn.TransactionProposalResponse{
		TransactionProposalResult: apitxn.TransactionProposalResult{
			Endorser: "http://peer1.com",
			Proposal: apitxn.TransactionProposal{
				TxnID:          txid,
				Proposal:       &pb.Proposal{Header: []byte("TEST"), Extension: []byte(""), Payload: []byte("")},
				SignedProposal: &pb.SignedProposal{Signature: []byte(""), ProposalBytes: []byte("")},
			},
			ProposalResponse: &pb.ProposalResponse{Response: &pb.Response{Message: "success", Status: 99, Payload: []byte("")}},
		},
	}

	input := []*apitxn.TransactionProposalResponse{test}

	_, err = channel.CreateTransaction(input)

	if err == nil || !strings.Contains(err.Error(), "unmarshal") {
		t.Fatal("Proposal response was supposed to fail in Create Transaction, invalid proposal header scenario")
	}

	//Test invalid proposal payload scenario
	test = &apitxn.TransactionProposalResponse{
		TransactionProposalResult: apitxn.TransactionProposalResult{
			Endorser: "http://peer1.com",
			Proposal: apitxn.TransactionProposal{
				TxnID:          txid,
				Proposal:       &pb.Proposal{Header: []byte(""), Extension: []byte(""), Payload: []byte("TEST")},
				SignedProposal: &pb.SignedProposal{Signature: []byte(""), ProposalBytes: []byte("")},
			},
			ProposalResponse: &pb.ProposalResponse{Response: &pb.Response{Message: "success", Status: 99, Payload: []byte("")}},
		},
	}

	input = []*apitxn.TransactionProposalResponse{test}

	_, err = channel.CreateTransaction(input)
	if err == nil || !strings.Contains(err.Error(), "unmarshal") {
		t.Fatal("Proposal response was supposed to fail in Create Transaction, invalid proposal payload scenario")
	}

	//Test proposal response
	test = &apitxn.TransactionProposalResponse{
		TransactionProposalResult: apitxn.TransactionProposalResult{
			Endorser: "http://peer1.com",
			Proposal: apitxn.TransactionProposal{
				Proposal:       &pb.Proposal{Header: []byte(""), Extension: []byte(""), Payload: []byte("")},
				SignedProposal: &pb.SignedProposal{Signature: []byte(""), ProposalBytes: []byte("")}, TxnID: txid,
			},
			ProposalResponse: &pb.ProposalResponse{Response: &pb.Response{Message: "success", Status: 99, Payload: []byte("")}},
		},
	}

	input = []*apitxn.TransactionProposalResponse{test}
	_, err = channel.CreateTransaction(input)

	if err == nil || err.Error() != "proposal response was not successful, error code 99, msg success" {
		t.Fatal("Proposal response was supposed to fail in Create Transaction")
	}

	//Test repeated field header nil scenario

	test = &apitxn.TransactionProposalResponse{
		TransactionProposalResult: apitxn.TransactionProposalResult{
			Endorser: "http://peer1.com",
			Proposal: apitxn.TransactionProposal{
				Proposal:       &pb.Proposal{Header: []byte(""), Extension: []byte(""), Payload: []byte("")},
				SignedProposal: &pb.SignedProposal{Signature: []byte(""), ProposalBytes: []byte("")}, TxnID: txid,
			},
			ProposalResponse: &pb.ProposalResponse{Response: &pb.Response{Message: "success", Status: 200, Payload: []byte("")}},
		},
	}

	_, err = channel.CreateTransaction([]*apitxn.TransactionProposalResponse{test})

	if err == nil || err.Error() != "repeated field endorsements has nil element" {
		t.Fatal("Proposal response was supposed to fail in Create Transaction")
	}

	//TODO: Need actual sample payload for success case

}

func TestSendInstantiateProposal(t *testing.T) {
	//Setup channel
	client := mocks.NewMockClient()
	user := mocks.NewMockUserWithMSPID("test", "1234")
	cryptoSuite := &mocks.MockCryptoSuite{}
	client.SaveUserToStateStore(user, true)
	client.SetCryptoSuite(cryptoSuite)
	client.SetUserContext(user)
	channel, _ := NewChannel("testChannel", client)

	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()
	proc := mock_apitxn.NewMockProposalProcessor(mockCtrl)

	tp := apitxn.TransactionProposal{SignedProposal: &pb.SignedProposal{}}
	tpr := apitxn.TransactionProposalResult{Endorser: "example.com", Status: 99, Proposal: tp, ProposalResponse: nil}

	proc.EXPECT().ProcessTransactionProposal(gomock.Any()).Return(tpr, nil)
	targets := []apitxn.ProposalProcessor{proc}

	//Add a Peer
	peer := mocks.MockPeer{MockName: "Peer1", MockURL: "http://peer1.com", MockRoles: []string{}, MockCert: nil}
	channel.AddPeer(&peer)

	tresponse, txnid, err := channel.SendInstantiateProposal("", nil, "",
		"", cauthdsl.SignedByMspMember("Org1MSP"), targets)

	if err == nil || err.Error() != "chaincodeName is required" {
		t.Fatal("Validation for chain code name parameter for send Instantiate Proposal failed")
	}

	tresponse, txnid, err = channel.SendInstantiateProposal("qscc", nil, "",
		"", cauthdsl.SignedByMspMember("Org1MSP"), targets)

	tresponse, txnid, err = channel.SendInstantiateProposal("qscc", nil, "",
		"", cauthdsl.SignedByMspMember("Org1MSP"), targets)

	if err == nil || err.Error() != "chaincodePath is required" {
		t.Fatal("Validation for chain code path for send Instantiate Proposal failed")
	}

	tresponse, txnid, err = channel.SendInstantiateProposal("qscc", nil, "test",
		"", cauthdsl.SignedByMspMember("Org1MSP"), targets)

	if err == nil || err.Error() != "chaincodeVersion is required" {
		t.Fatal("Validation for chain code version for send Instantiate Proposal failed")
	}

	tresponse, txnid, err = channel.SendInstantiateProposal("qscc", nil, "test",
		"1", nil, nil)
	if err == nil || err.Error() != "chaincodePolicy is required" {
		t.Fatal("Validation for chain code policy for send Instantiate Proposal failed")
	}

	tresponse, txnid, err = channel.SendInstantiateProposal("qscc", nil, "test",
		"1", cauthdsl.SignedByMspMember("Org1MSP"), targets)

	if err != nil || len(tresponse) == 0 || txnid.ID == "" {
		t.Fatal("Send Instantiate Proposal Test failed")
	}

	tresponse, txnid, err = channel.SendInstantiateProposal("qscc", nil, "test",
		"1", cauthdsl.SignedByMspMember("Org1MSP"), nil)

	if err == nil || err.Error() != "missing peer objects for chaincode proposal" {
		t.Fatal("Missing peer objects validation is not working as expected")
	}
}

func TestSendUpgradeProposal(t *testing.T) {
	//Setup channel
	client := mocks.NewMockClient()
	user := mocks.NewMockUserWithMSPID("test", "1234")
	cryptoSuite := &mocks.MockCryptoSuite{}
	client.SaveUserToStateStore(user, true)
	client.SetCryptoSuite(cryptoSuite)
	client.SetUserContext(user)
	channel, _ := NewChannel("testChannel", client)

	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()
	proc := mock_apitxn.NewMockProposalProcessor(mockCtrl)

	tp := apitxn.TransactionProposal{SignedProposal: &pb.SignedProposal{}}
	tpr := apitxn.TransactionProposalResult{Endorser: "example.com", Status: 99, Proposal: tp, ProposalResponse: nil}

	proc.EXPECT().ProcessTransactionProposal(gomock.Any()).Return(tpr, nil)
	targets := []apitxn.ProposalProcessor{proc}

	//Add a Peer
	peer := mocks.MockPeer{MockName: "Peer1", MockURL: "http://peer1.com", MockRoles: []string{}, MockCert: nil}
	channel.AddPeer(&peer)

	tresponse, txnid, err := channel.SendUpgradeProposal("", nil, "",
		"", cauthdsl.SignedByMspMember("Org1MSP"), targets)

	if err == nil || err.Error() != "chaincodeName is required" {
		t.Fatal("Validation for chain code name parameter for send Upgrade Proposal failed")
	}

	tresponse, txnid, err = channel.SendUpgradeProposal("qscc", nil, "",
		"", cauthdsl.SignedByMspMember("Org1MSP"), targets)

	tresponse, txnid, err = channel.SendUpgradeProposal("qscc", nil, "",
		"", cauthdsl.SignedByMspMember("Org1MSP"), targets)

	if err == nil || err.Error() != "chaincodePath is required" {
		t.Fatal("Validation for chain code path for send Upgrade Proposal failed")
	}

	tresponse, txnid, err = channel.SendUpgradeProposal("qscc", nil, "test",
		"", cauthdsl.SignedByMspMember("Org1MSP"), targets)

	if err == nil || err.Error() != "chaincodeVersion is required" {
		t.Fatal("Validation for chain code version for send Upgrade Proposal failed")
	}

	tresponse, txnid, err = channel.SendUpgradeProposal("qscc", nil, "test",
		"2", nil, nil)
	if err == nil || err.Error() != "chaincodePolicy is required" {
		t.Fatal("Validation for chain code policy for send Upgrade Proposal failed")
	}

	tresponse, txnid, err = channel.SendUpgradeProposal("qscc", nil, "test",
		"2", cauthdsl.SignedByMspMember("Org1MSP"), targets)

	if err != nil || len(tresponse) == 0 || txnid.ID == "" {
		t.Fatal("Send Upgrade Proposal Test failed")
	}

	tresponse, txnid, err = channel.SendUpgradeProposal("qscc", nil, "test",
		"2", cauthdsl.SignedByMspMember("Org1MSP"), nil)
	if err == nil || err.Error() != "missing peer objects for chaincode proposal" {
		t.Fatal("Missing peer objects validation is not working as expected")
	}
}

type mockReader struct {
	err error
}

func (r *mockReader) Read(p []byte) (int, error) {
	if r.err != nil {
		return 0, r.err
	}
	n, _ := rand.Read(p)
	return n, nil
}

func TestBroadcastEnvelope(t *testing.T) {
	//Setup channel
	channel, _ := setupTestChannel()

	lsnr1 := make(chan *fab.SignedEnvelope)
	lsnr2 := make(chan *fab.SignedEnvelope)
	//Create mock orderers
	orderer1 := mocks.NewMockOrderer("1", lsnr1)
	orderer2 := mocks.NewMockOrderer("2", lsnr2)

	//Add the orderers
	channel.AddOrderer(orderer1)
	channel.AddOrderer(orderer2)

	peer := mocks.MockPeer{MockName: "Peer1", MockURL: "http://peer1.com", MockRoles: []string{}, MockCert: nil}
	channel.AddPeer(&peer)

	sigEnvelope := &fab.SignedEnvelope{
		Signature: []byte(""),
		Payload:   []byte(""),
	}
	res, err := channel.BroadcastEnvelope(sigEnvelope)

	if err != nil || res.Err != nil {
		t.Fatalf("Test Broadcast Envelope Failed, cause %v %v", err, res)
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
		t.Fatal("Both or none orderers were selected for broadcast:", firstSelected+secondSelected)
	}

	// Now make 1 of them fail and repeatedly broadcast
	broadcastCount := 50
	for i := 0; i < broadcastCount; i++ {
		orderer1.(mocks.MockOrderer).EnqueueSendBroadcastError(errors.New("Service Unavailable"))
	}
	// It should always succeed even though one of them has failed
	for i := 0; i < broadcastCount; i++ {
		if res, err := channel.BroadcastEnvelope(sigEnvelope); err != nil || res.Err != nil {
			t.Fatalf("Test Broadcast Envelope Failed, cause %v %v", err, res)
		}
	}

	// Now, fail both and ensure any attempt fails
	for i := 0; i < broadcastCount; i++ {
		orderer1.(mocks.MockOrderer).EnqueueSendBroadcastError(errors.New("Service Unavailable"))
		orderer2.(mocks.MockOrderer).EnqueueSendBroadcastError(errors.New("Service Unavailable"))
	}

	for i := 0; i < broadcastCount; i++ {
		res, err := channel.BroadcastEnvelope(sigEnvelope)
		if err != nil {
			t.Fatalf("Test Broadcast sending failed, cause %v", err)
		}
		if res.Err == nil {
			t.Fatal("Test Broadcast succeeded, but it should have failed")
		}
		if !strings.Contains(res.Err.Error(), "Service Unavailable") {
			t.Fatal("Test Broadcast failed but didn't return the correct reason(should contain 'Service Unavailable')")
		}
	}

	channel.RemoveOrderer(orderer1)
	channel.RemoveOrderer(orderer2)
	_, err = channel.BroadcastEnvelope(sigEnvelope)

	if err == nil || err.Error() != "orderers not set" {
		t.Fatal("orderers not set validation on broadcast envelope is not working as expected")
	}
}

func TestSendTransaction(t *testing.T) {

	channel, _ := setupTestChannel()

	response, err := channel.SendTransaction(nil)

	//Expect orderer is nil error
	if response != nil || err == nil || err.Error() != "orderers is nil" {
		t.Fatal("Test SendTransaction failed, it was supposed to fail with 'orderers is nil' error")
	}

	//Create mock orderer
	orderer := mocks.NewMockOrderer("", nil)

	//Add an orderer
	channel.AddOrderer(orderer)

	//Call Send Transaction with nil tx
	response, err = channel.SendTransaction(nil)

	//Expect tx is nil error
	if response != nil || err == nil || err.Error() != "transaction is nil" {
		t.Fatal("Test SendTransaction failed, it was supposed to fail with 'transaction is nil' error")
	}

	//Create tx with nil proposal
	txn := apitxn.Transaction{
		Proposal: &apitxn.TransactionProposal{
			Proposal: nil,
		},
		Transaction: &pb.Transaction{},
	}

	//Call Send Transaction with nil proposal
	response, err = channel.SendTransaction(&txn)

	//Expect proposal is nil error
	if response != nil || err == nil || err.Error() != "proposal is nil" {
		t.Fatal("Test SendTransaction failed, it was supposed to fail with 'proposal is nil' error")
	}

	//Create tx with improper proposal header
	txn = apitxn.Transaction{
		Proposal: &apitxn.TransactionProposal{
			Proposal: &pb.Proposal{Header: []byte("TEST")},
		},
		Transaction: &pb.Transaction{},
	}
	//Call Send Transaction
	response, err = channel.SendTransaction(&txn)

	//Expect header unmarshal error
	if response != nil || err == nil || !strings.Contains(err.Error(), "unmarshal") {
		t.Fatal("Test SendTransaction failed, it was supposed to fail with '...unmarshal...' error")
	}

	//Create tx with proper proposal header
	txn = apitxn.Transaction{
		Proposal: &apitxn.TransactionProposal{
			Proposal: &pb.Proposal{Header: []byte(""), Payload: []byte(""), Extension: []byte("")},
		},
		Transaction: &pb.Transaction{},
	}

	//Call Send Transaction
	response, err = channel.SendTransaction(&txn)

	if response == nil || err != nil {
		t.Fatalf("Test SendTransaction failed, reason : '%s'", err.Error())
	}
}

func TestBuildChannelHeader(t *testing.T) {

	header, err := BuildChannelHeader(common.HeaderType_CHAINCODE_PACKAGE, "test", "", 1, "1234", time.Time{}, []byte{})

	if err != nil || header == nil {
		t.Fatalf("Test Build Channel Header failed, cause : '%s'", err.Error())
	}

}

func TestSignPayload(t *testing.T) {

	client := mocks.NewMockInvalidClient()
	user := mocks.NewMockUser("test")
	cryptoSuite := &mocks.MockCryptoSuite{}
	client.SaveUserToStateStore(user, true)
	client.SetCryptoSuite(cryptoSuite)
	channel, _ := NewChannel("testChannel", client)

	signedEnv, err := channel.SignPayload([]byte(""))

	if err == nil {
		t.Fatal("Test Sign Payload was supposed to fail")
	}

	channel, _ = setupTestChannel()
	signedEnv, err = channel.SignPayload([]byte(""))

	if err != nil || signedEnv == nil {
		t.Fatal("Test Sign Payload Failed")
	}

}

func TestConcurrentOrderers(t *testing.T) {
	// Determine number of orderers to use - environment can override
	const numOrderersDefault = 2000
	numOrderersEnv := os.Getenv("TEST_MASSIVE_ORDERER_COUNT")
	numOrderers, err := strconv.Atoi(numOrderersEnv)
	if err != nil {
		numOrderers = numOrderersDefault
	}

	channel, err := setupMassiveTestChannel(0, numOrderers)
	if err != nil {
		t.Fatalf("Failed to create massive channel: %s", err)
	}

	txn := apitxn.Transaction{
		Proposal: &apitxn.TransactionProposal{
			Proposal: &pb.Proposal{},
		},
		Transaction: &pb.Transaction{},
	}
	_, err = channel.SendTransaction(&txn)
	if err != nil {
		t.Fatalf("SendTransaction returned error: %s", err)
	}
}
