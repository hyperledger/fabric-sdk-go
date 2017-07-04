/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package channel

import (
	"fmt"
	"net"
	"os"
	"strconv"
	"testing"

	"google.golang.org/grpc"

	fab "github.com/hyperledger/fabric-sdk-go/api/apifabclient"
	"github.com/hyperledger/fabric-sdk-go/api/apitxn"
	fc "github.com/hyperledger/fabric-sdk-go/pkg/fabric-client/internal"
	mocks "github.com/hyperledger/fabric-sdk-go/pkg/fabric-client/mocks"
	peer "github.com/hyperledger/fabric-sdk-go/pkg/fabric-client/peer"

	"io/ioutil"
	"time"

	"strings"

	"github.com/golang/mock/gomock"
	"github.com/hyperledger/fabric-sdk-go/api/apitxn/mocks"
	"github.com/hyperledger/fabric/protos/common"
	pb "github.com/hyperledger/fabric/protos/peer"
)

var testAddress = "0.0.0.0:5244"

var validRootCA = `-----BEGIN CERTIFICATE-----
MIICYjCCAgmgAwIBAgIUB3CTDOU47sUC5K4kn/Caqnh114YwCgYIKoZIzj0EAwIw
fzELMAkGA1UEBhMCVVMxEzARBgNVBAgTCkNhbGlmb3JuaWExFjAUBgNVBAcTDVNh
biBGcmFuY2lzY28xHzAdBgNVBAoTFkludGVybmV0IFdpZGdldHMsIEluYy4xDDAK
BgNVBAsTA1dXVzEUMBIGA1UEAxMLZXhhbXBsZS5jb20wHhcNMTYxMDEyMTkzMTAw
WhcNMjExMDExMTkzMTAwWjB/MQswCQYDVQQGEwJVUzETMBEGA1UECBMKQ2FsaWZv
cm5pYTEWMBQGA1UEBxMNU2FuIEZyYW5jaXNjbzEfMB0GA1UEChMWSW50ZXJuZXQg
V2lkZ2V0cywgSW5jLjEMMAoGA1UECxMDV1dXMRQwEgYDVQQDEwtleGFtcGxlLmNv
bTBZMBMGByqGSM49AgEGCCqGSM49AwEHA0IABKIH5b2JaSmqiQXHyqC+cmknICcF
i5AddVjsQizDV6uZ4v6s+PWiJyzfA/rTtMvYAPq/yeEHpBUB1j053mxnpMujYzBh
MA4GA1UdDwEB/wQEAwIBBjAPBgNVHRMBAf8EBTADAQH/MB0GA1UdDgQWBBQXZ0I9
qp6CP8TFHZ9bw5nRtZxIEDAfBgNVHSMEGDAWgBQXZ0I9qp6CP8TFHZ9bw5nRtZxI
EDAKBggqhkjOPQQDAgNHADBEAiAHp5Rbp9Em1G/UmKn8WsCbqDfWecVbZPQj3RK4
oG5kQQIgQAe4OOKYhJdh3f7URaKfGTf492/nmRmtK+ySKjpHSrU=
-----END CERTIFICATE-----
`

func TestChannelMethods(t *testing.T) {
	client := mocks.NewMockClient()
	channel, err := NewChannel("testChannel", client)
	if err != nil {
		t.Fatalf("NewChannel return error[%s]", err)
	}
	if channel.Name() != "testChannel" {
		t.Fatalf("NewChannel create wrong channel")
	}

	_, err = NewChannel("", client)
	if err == nil {
		t.Fatalf("NewChannel didn't return error")
	}
	if err.Error() != "failed to create Channel. Missing required 'name' parameter" {
		t.Fatalf("NewChannel didn't return right error")
	}

	_, err = NewChannel("testChannel", nil)
	if err == nil {
		t.Fatalf("NewChannel didn't return error")
	}
	if err.Error() != "failed to create Channel. Missing required 'clientContext' parameter" {
		t.Fatalf("NewChannel didn't return right error")
	}

}

func TestQueryMethods(t *testing.T) {
	channel, _ := setupTestChannel()

	_, err := channel.QueryBlock(-1)
	if err == nil {
		t.Fatalf("Query block cannot be negative number")
	}

	_, err = channel.QueryBlockByHash(nil)
	if err == nil {
		t.Fatalf("Query hash cannot be nil")
	}
	_, err = channel.QueryByChaincode("", []string{"method"}, nil)
	if err == nil {
		t.Fatalf("QueryByChannelcode: name cannot be empty")
	}

	_, err = channel.QueryByChaincode("qscc", nil, nil)
	if err == nil {
		t.Fatalf("QueryByChannelcode: arguments cannot be empty")
	}

	_, err = channel.QueryByChaincode("qscc", []string{"method"}, nil)
	if err == nil {
		t.Fatalf("QueryByChannelcode: targets cannot be empty")
	}

}

func TestChannelQueryBlock(t *testing.T) {

	channel, _ := setupTestChannel()

	peer := mocks.MockPeer{MockName: "Peer1", MockURL: "http://peer1.com", MockRoles: []string{}, MockCert: nil}
	err := channel.AddPeer(&peer)

	_, err = channel.QueryBlock(1)

	if err != nil {
		t.Fatal("Test channel query block failed,")
	}

	_, err = channel.QueryBlockByHash([]byte(""))

	if err != nil {
		t.Fatal("Test channel query block by hash failed,")
	}

}

func TestChannelConfigs(t *testing.T) {

	client := mocks.NewMockClient()
	user := mocks.NewMockUser("test")
	cryptoSuite := &mocks.MockCryptoSuite{}
	client.SaveUserToStateStore(user, true)
	client.SetCryptoSuite(cryptoSuite)

	channel, _ := NewChannel("testChannel", client)

	if client.GetConfig().IsSecurityEnabled() != channel.IsSecurityEnabled() {
		t.Fatal("Is Security Enabled flag is incorrect in channel")
	}

	if client.GetConfig().TcertBatchSize() != channel.TCertBatchSize() {
		t.Fatal("Tcert batch size is incorrect")
	}

	channel.SetTCertBatchSize(22)

	if channel.TCertBatchSize() != 22 {
		t.Fatal("TCert batch size update on channel is not working")
	}

	if channel.IsReadonly() {
		//TODO: Rightnow it is returning false always, need to revisit test once actual implementation is provided
		t.Fatal("Is Readonly test failed")
	}

	if channel.UpdateChannel() {
		//TODO: Rightnow it is returning false always, need to revisit test once actual implementation is provided
		t.Fatal("UpdateChannel test failed")
	}

	if channel.QueryExtensionInterface().ClientContext() != client {
		t.Fatal("Client context not matching with client")
	}

	channel.SetMSPManager(nil)

}

func TestCreateTransactionProposal(t *testing.T) {

	channel, _ := setupTestChannel()

	tProposal, err := channel.CreateTransactionProposal("qscc", "testChannel", nil, true, nil)

	if err != nil {
		t.Fatal("Create Transaction Proposal Failed", err)
	}

	_, errx := channel.QueryExtensionInterface().ProposalBytes(tProposal)

	if errx != nil {
		t.Fatal("Call to proposal bytes from channel extension failed")
	}

}

func TestAddRemoveOrderer(t *testing.T) {

	//Setup channel
	channel, _ := setupTestChannel()

	//Create mock orderer
	orderer := mocks.NewMockOrderer("", nil)

	//Add an orderer
	channel.AddOrderer(orderer)

	//Check if orderer is being added successfully
	if len(channel.Orderers()) != 1 {
		t.Fatal("Adding orderers to channel failed")
	}

	//Remove the orderer now
	channel.RemoveOrderer(orderer)

	//Check if list of orderers is empty now
	if len(channel.Orderers()) != 0 {
		t.Fatal("Removing orderers from channel failed")
	}
}

func TestAnchorAndRemovePeers(t *testing.T) {
	//Setup channel
	channel, _ := setupTestChannel()

	//Add a Peer
	peer := mocks.MockPeer{MockName: "Peer1", MockURL: "http://peer1.com", MockRoles: []string{}, MockCert: nil}
	channel.AddPeer(&peer)

	//Remove and Test
	channel.RemovePeer(&peer)
	if len(channel.Peers()) != 0 {
		t.Fatal("Remove Peer failed")
	}

	//Add the Peer again
	channel.AddPeer(&peer)

	channel.Initialize(nil)
	if len(channel.AnchorPeers()) != 0 {
		//Currently testing only for empty anchor list
		t.Fatal("Anchor peer list is incorrect")
	}
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

	tresponse, str, err := channel.SendInstantiateProposal("", "testChannel", nil, "",
		"", targets)

	if err == nil || err.Error() != "Missing 'chaincodeName' parameter" {
		t.Fatal("Validation for chain code name parameter for send Instantiate Proposal failed")
	}

	tresponse, str, err = channel.SendInstantiateProposal("qscc", "", nil, "",
		"", targets)

	if err == nil || err.Error() != "Missing 'channelID' parameter" {
		t.Fatal("Validation for chain id parameter for send Instantiate Proposal failed")
	}

	tresponse, str, err = channel.SendInstantiateProposal("qscc", "1234", nil, "",
		"", targets)

	if err == nil || err.Error() != "Missing 'chaincodePath' parameter" {
		t.Fatal("Validation for chain code path for send Instantiate Proposal failed")
	}

	tresponse, str, err = channel.SendInstantiateProposal("qscc", "1234", nil, "test",
		"", targets)

	if err == nil || err.Error() != "Missing 'chaincodeVersion' parameter" {
		t.Fatal("Validation for chain code version for send Instantiate Proposal failed")
	}

	tresponse, str, err = channel.SendInstantiateProposal("qscc", "1234", nil, "test",
		"1", targets)

	if err != nil || len(tresponse) == 0 || str == "" {
		t.Fatal("Send Instantiate Proposal Test failed")
	}

	tresponse, str, err = channel.SendInstantiateProposal("qscc", "1234", nil, "test",
		"1", nil)
	if err == nil || err.Error() != "Missing peer objects for instantiate CC proposal" {
		t.Fatal("Missing peer objects validation is not working as expected")
	}

}

func TestQueryInstantiatedChaincodes(t *testing.T) {
	channel, _ := setupTestChannel()

	peer := mocks.MockPeer{MockName: "Peer1", MockURL: "http://peer1.com", MockRoles: []string{}, MockCert: nil}
	err := channel.AddPeer(&peer)

	res, err := channel.QueryInstantiatedChaincodes()

	if err != nil || res == nil {
		t.Fatal("Test QueryInstatiated chaincode failed")
	}

}

func TestQueryTransaction(t *testing.T) {
	channel, _ := setupTestChannel()

	peer := mocks.MockPeer{MockName: "Peer1", MockURL: "http://peer1.com", MockRoles: []string{}, MockCert: nil}
	err := channel.AddPeer(&peer)

	res, err := channel.QueryTransaction("txid")

	if err != nil || res == nil {
		t.Fatal("Test QueryTransaction failed")
	}
}

func TestQueryInfo(t *testing.T) {
	channel, _ := setupTestChannel()

	peer := mocks.MockPeer{MockName: "Peer1", MockURL: "http://peer1.com", MockRoles: []string{}, MockCert: nil}
	err := channel.AddPeer(&peer)

	res, err := channel.QueryInfo()

	if err != nil || res == nil {
		t.Fatal("Test QueryInfo failed")
	}
}

func TestCreateTransaction(t *testing.T) {
	channel, _ := setupTestChannel()

	peer := mocks.MockPeer{MockName: "Peer1", MockURL: "http://peer1.com", MockRoles: []string{}, MockCert: nil}
	channel.AddPeer(&peer)

	//Test Empty proposal response scenario
	_, err := channel.CreateTransaction([]*apitxn.TransactionProposalResponse{})

	if err == nil || err.Error() != "At least one proposal response is necessary" {
		t.Fatal("Proposal response was supposed to fail in Create Transaction, for empty proposal response scenario")
	}

	//Test invalid proposal header scenario

	test := &apitxn.TransactionProposalResponse{
		TransactionProposalResult: apitxn.TransactionProposalResult{
			Endorser: "http://peer1.com",
			Proposal: apitxn.TransactionProposal{
				TransactionID:  "1234",
				Proposal:       &pb.Proposal{Header: []byte("TEST"), Extension: []byte(""), Payload: []byte("")},
				SignedProposal: &pb.SignedProposal{Signature: []byte(""), ProposalBytes: []byte("")},
			},
			ProposalResponse: &pb.ProposalResponse{Response: &pb.Response{Message: "success", Status: 99, Payload: []byte("")}},
		},
	}

	input := []*apitxn.TransactionProposalResponse{test}

	_, err = channel.CreateTransaction(input)

	if err == nil || err.Error() != "Could not unmarshal the proposal header" {
		t.Fatal("Proposal response was supposed to fail in Create Transaction, invalid proposal header scenario")
	}

	//Test invalid proposal payload scenario
	test = &apitxn.TransactionProposalResponse{
		TransactionProposalResult: apitxn.TransactionProposalResult{
			Endorser: "http://peer1.com",
			Proposal: apitxn.TransactionProposal{
				TransactionID:  "1234",
				Proposal:       &pb.Proposal{Header: []byte(""), Extension: []byte(""), Payload: []byte("TEST")},
				SignedProposal: &pb.SignedProposal{Signature: []byte(""), ProposalBytes: []byte("")},
			},
			ProposalResponse: &pb.ProposalResponse{Response: &pb.Response{Message: "success", Status: 99, Payload: []byte("")}},
		},
	}

	input = []*apitxn.TransactionProposalResponse{test}

	_, err = channel.CreateTransaction(input)
	if err == nil || err.Error() != "Could not unmarshal the proposal payload" {
		t.Fatal("Proposal response was supposed to fail in Create Transaction, invalid proposal payload scenario")
	}

	//Test proposal response
	test = &apitxn.TransactionProposalResponse{
		TransactionProposalResult: apitxn.TransactionProposalResult{
			Endorser: "http://peer1.com",
			Proposal: apitxn.TransactionProposal{
				Proposal:       &pb.Proposal{Header: []byte(""), Extension: []byte(""), Payload: []byte("")},
				SignedProposal: &pb.SignedProposal{Signature: []byte(""), ProposalBytes: []byte("")}, TransactionID: "1234",
			},
			ProposalResponse: &pb.ProposalResponse{Response: &pb.Response{Message: "success", Status: 99, Payload: []byte("")}},
		},
	}

	input = []*apitxn.TransactionProposalResponse{test}
	_, err = channel.CreateTransaction(input)

	if err == nil || err.Error() != "Proposal response was not successful, error code 99, msg success" {
		t.Fatal("Proposal response was supposed to fail in Create Transaction")
	}

	//Test repeated field header nil scenario

	test = &apitxn.TransactionProposalResponse{
		TransactionProposalResult: apitxn.TransactionProposalResult{
			Endorser: "http://peer1.com",
			Proposal: apitxn.TransactionProposal{
				Proposal:       &pb.Proposal{Header: []byte(""), Extension: []byte(""), Payload: []byte("")},
				SignedProposal: &pb.SignedProposal{Signature: []byte(""), ProposalBytes: []byte("")}, TransactionID: "1234",
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

func TestLoadConfigUpdateEnvelope(t *testing.T) {
	//Get Channel
	channel, _ := setupTestChannel()

	//Read config file from test directory
	fileLoc := "../../../test/fixtures/channel/mychanneltx.tx"
	res, err := ioutil.ReadFile(fileLoc)

	//Pass config to LoadConfigUpdateEnvelope and test
	err = channel.LoadConfigUpdateEnvelope(res)

	if err != nil {
		t.Fatalf("LoadConfigUpdateEnvelope Test Failed with, Cause '%s'", err.Error())
	}

	err = channel.Initialize(res)

	if err == nil {
		t.Fatalf("Initialize Negative Test Failed with, Cause '%s'", err.Error())
	}

	org1MSPID := "ORG1MSP"
	org2MSPID := "ORG2MSP"

	builder := &mocks.MockConfigUpdateEnvelopeBuilder{}

	err = channel.LoadConfigUpdateEnvelope(builder.BuildBytes())

	if err == nil {
		t.Fatal("Expected error was : channel initialization error: unable to load MSPs from config")
	}

	builder = &mocks.MockConfigUpdateEnvelopeBuilder{
		ChannelID: "mychannel",
		MockConfigGroupBuilder: mocks.MockConfigGroupBuilder{
			ModPolicy: "Admins",
			MSPNames: []string{
				org1MSPID,
				org2MSPID,
			},
			OrdererAddress: "localhost:7054",
			RootCA:         validRootCA,
		},
	}

	//Create mock orderer
	configBuilder := &mocks.MockConfigBlockBuilder{
		MockConfigGroupBuilder: mocks.MockConfigGroupBuilder{
			ModPolicy: "Admins",
			MSPNames: []string{
				org1MSPID,
				org2MSPID,
			},
			OrdererAddress: "localhost:7054",
			//RootCA:         validRootCA,
		},
		Index:           0,
		LastConfigIndex: 0,
	}
	orderer := mocks.NewMockOrderer("", nil)
	orderer.(mocks.MockOrderer).EnqueueForSendDeliver(configBuilder.Build())
	orderer.(mocks.MockOrderer).EnqueueForSendDeliver(configBuilder.Build())
	channel.AddOrderer(orderer)

	//Add a second orderer
	configBuilder = &mocks.MockConfigBlockBuilder{
		MockConfigGroupBuilder: mocks.MockConfigGroupBuilder{
			ModPolicy: "Admins",
			MSPNames: []string{
				org1MSPID,
				org2MSPID,
			},
			OrdererAddress: "localhost:7054",
			//RootCA:         validRootCA,
		},
		Index:           0,
		LastConfigIndex: 0,
	}
	orderer = mocks.NewMockOrderer("", nil)
	orderer.(mocks.MockOrderer).EnqueueForSendDeliver(configBuilder.Build())
	orderer.(mocks.MockOrderer).EnqueueForSendDeliver(configBuilder.Build())
	channel.AddOrderer(orderer)
	err = channel.Initialize(nil)

	if err == nil {
		t.Fatal("Initialize on orderers config supposed to fail with 'could not decode pem bytes'")
	}

}

func TestBroadcastEnvelope(t *testing.T) {

	//Setup channel
	channel, _ := setupTestChannel()

	//Create mock orderer
	orderer := mocks.NewMockOrderer("", nil)

	//Add an orderer
	channel.AddOrderer(orderer)

	peer := mocks.MockPeer{MockName: "Peer1", MockURL: "http://peer1.com", MockRoles: []string{}, MockCert: nil}
	channel.AddPeer(&peer)

	sigEnvelope := &fab.SignedEnvelope{
		Signature: []byte(""),
		Payload:   []byte(""),
	}
	res, err := channel.QueryExtensionInterface().BroadcastEnvelope(sigEnvelope)

	if err != nil || res == nil {
		t.Fatalf("Test Broadcast Envelope Failed, cause %s", err.Error())
	}

	channel.RemoveOrderer(orderer)
	_, err = channel.QueryExtensionInterface().BroadcastEnvelope(sigEnvelope)

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
	if response != nil || err == nil || err.Error() != "Transaction is nil" {
		t.Fatal("Test SendTransaction failed, it was supposed to fail with 'Transaction is nil' error")
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
	if response != nil || err == nil || err.Error() != "Could not unmarshal the proposal header" {
		t.Fatal("Test SendTransaction failed, it was supposed to fail with 'Could not unmarshal the proposal header' error")
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

func TestBuildChannelHeader(t *testing.T) {

	header, err := BuildChannelHeader(common.HeaderType_CHAINCODE_PACKAGE, "test", "", 1, "1234", time.Time{})

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

	signedEnv, err := channel.QueryExtensionInterface().SignPayload([]byte(""))

	if err == nil {
		t.Fatal("Test Sign Payload was supposed to fail")
	}

	channel, _ = setupTestChannel()
	signedEnv, err = channel.QueryExtensionInterface().SignPayload([]byte(""))

	if err != nil || signedEnv == nil {
		t.Fatal("Test Sign Payload Failed")
	}

}

func TestGenesisBlock(t *testing.T) {
	var peers []fab.Peer
	channel, _ := setupTestChannel()
	peer, _ := peer.NewPeer(testAddress, mocks.NewMockConfig())
	peers = append(peers, peer)
	orderer := mocks.NewMockOrderer("", nil)
	orderer.(mocks.MockOrderer).EnqueueForSendDeliver(mocks.NewSimpleMockError())
	nonce, _ := fc.GenerateRandomNonce()
	txID, _ := fc.ComputeTxID(nonce, []byte("testID"))

	genesisBlockReq := &fab.GenesisBlockRequest{
		TxID:  txID,
		Nonce: nonce,
	}

	channel.AddOrderer(orderer)

	//Call get Genesis block
	_, err := channel.GenesisBlock(genesisBlockReq)

	//Expecting error
	if err == nil {
		t.Fatal("GenesisBlock Test supposed to fail with error")
	}

	//Remove existing orderer
	channel.RemoveOrderer(orderer)

	//Create new orderer
	orderer = mocks.NewMockOrderer("", nil)

	channel.AddOrderer(orderer)

	//Call get Genesis block
	_, err = channel.GenesisBlock(genesisBlockReq)

	//It should fail with timeout
	if err == nil || !strings.HasSuffix(err.Error(), "Timeout waiting for response from orderer") {
		t.Fatal("GenesisBlock Test supposed to fail with timeout error")
	}

	//Validation test
	genesisBlockReq = &fab.GenesisBlockRequest{}
	_, err = channel.GenesisBlock(genesisBlockReq)

	if err == nil || err.Error() != "GenesisBlock - error: Missing txId input parameter with the required transaction identifier" {
		t.Fatal("validation on missing txID input parameter is not working as expected")
	}

	genesisBlockReq = &fab.GenesisBlockRequest{
		TxID: txID,
	}
	_, err = channel.GenesisBlock(genesisBlockReq)

	if err == nil || err.Error() != "GenesisBlock - error: Missing nonce input parameter with the required single use number" {
		t.Fatal("validation on missing nonce input parameter is not working as expected")
	}

	channel.RemoveOrderer(orderer)

	genesisBlockReq = &fab.GenesisBlockRequest{
		TxID:  txID,
		Nonce: nonce,
	}

	_, err = channel.GenesisBlock(genesisBlockReq)

	if err == nil || err.Error() != "GenesisBlock - error: Missing orderer assigned to this channel for the GenesisBlock request" {
		t.Fatal("validation on no ordererds on channel is not working as expected")
	}

}

func TestPrimaryPeer(t *testing.T) {
	channel, _ := setupTestChannel()

	if channel.PrimaryPeer() != nil {
		t.Fatal("Call to Primary peer on empty channel should always return nil")
	}

	// Channel had one peer
	peer1 := mocks.MockPeer{MockName: "Peer1", MockURL: "http://peer1.com", MockRoles: []string{}, MockCert: nil}
	err := channel.AddPeer(&peer1)
	if err != nil {
		t.Fatalf("Error adding peer: %v", err)
	}

	// Test primary defaults to channel peer
	primary := channel.PrimaryPeer()
	if primary.URL() != peer1.URL() {
		t.Fatalf("Primary Peer failed to default")
	}

	// Channel has two peers
	peer2 := mocks.MockPeer{MockName: "Peer2", MockURL: "http://peer2.com", MockRoles: []string{}, MockCert: nil}
	err = channel.AddPeer(&peer2)
	if err != nil {
		t.Fatalf("Error adding peer: %v", err)
	}

	// Set primary to invalid URL
	invalidChoice := mocks.MockPeer{MockName: "", MockURL: "http://xyz.com", MockRoles: []string{}, MockCert: nil}
	err = channel.SetPrimaryPeer(&invalidChoice)
	if err == nil {
		t.Fatalf("Primary Peer was set to an invalid peer")
	}

	// Set primary to valid peer 2 URL
	choice := mocks.MockPeer{MockName: "", MockURL: "http://peer2.com", MockRoles: []string{}, MockCert: nil}
	err = channel.SetPrimaryPeer(&choice)
	if err != nil {
		t.Fatalf("Failed to set valid primary peer")
	}

	// Test primary equals our choice
	primary = channel.PrimaryPeer()
	if primary.URL() != peer2.URL() || primary.Name() != peer2.Name() {
		t.Fatalf("Primary and our choice are not equal")
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

func TestConcurrentOrderers(t *testing.T) {
	// Determine number of orderers to use - environment can override
	const numOrderersDefault = 10000
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

func TestJoinChannel(t *testing.T) {
	var peers []fab.Peer
	endorserServer := startEndorserServer(t)
	channel, _ := setupTestChannel()
	peer, _ := peer.NewPeer(testAddress, mocks.NewMockConfig())
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

func TestChannelInitializeFromOrderer(t *testing.T) {
	org1MSPID := "ORG1MSP"
	org2MSPID := "ORG2MSP"

	channel, _ := setupTestChannel()
	builder := &mocks.MockConfigBlockBuilder{
		MockConfigGroupBuilder: mocks.MockConfigGroupBuilder{
			ModPolicy: "Admins",
			MSPNames: []string{
				org1MSPID,
				org2MSPID,
			},
			OrdererAddress: "localhost:7054",
			RootCA:         validRootCA,
		},
		Index:           0,
		LastConfigIndex: 0,
	}
	orderer := mocks.NewMockOrderer("", nil)
	orderer.(mocks.MockOrderer).EnqueueForSendDeliver(builder.Build())
	orderer.(mocks.MockOrderer).EnqueueForSendDeliver(builder.Build())
	err := channel.AddOrderer(orderer)
	if err != nil {
		t.Fatalf("Error adding orderer: %v", err)
	}

	err = channel.Initialize(nil)
	if err != nil {
		t.Fatalf("channel Initialize failed : %v", err)
	}
	if !channel.IsInitialized() {
		t.Fatalf("channel Initialize failed : channel initialized flag not set")
	}

	mspManager := channel.MSPManager()
	if mspManager == nil {
		t.Fatalf("nil MSPManager on new channel")
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

	channel.SetMSPManager(nil)
	if channel.MSPManager() != nil {
		t.Fatal("Set MSPManager is not working as expected")
	}

}

func TestOrganizationUnits(t *testing.T) {
	org1MSPID := "ORG1MSP"
	org2MSPID := "ORG2MSP"

	channel, _ := setupTestChannel()
	orgUnits, err := channel.OrganizationUnits()

	if len(orgUnits) > 0 {
		t.Fatalf("Returned non configured organizational unit : %v", err)
	}
	builder := &mocks.MockConfigBlockBuilder{
		MockConfigGroupBuilder: mocks.MockConfigGroupBuilder{
			ModPolicy: "Admins",
			MSPNames: []string{
				channel.Name(),
				org1MSPID,
				org2MSPID,
			},
			OrdererAddress: "localhost:7054",
			RootCA:         validRootCA,
		},
		Index:           0,
		LastConfigIndex: 0,
	}
	orderer := mocks.NewMockOrderer("", nil)
	orderer.(mocks.MockOrderer).EnqueueForSendDeliver(builder.Build())
	orderer.(mocks.MockOrderer).EnqueueForSendDeliver(builder.Build())
	err = channel.AddOrderer(orderer)
	if err != nil {
		t.Fatalf("Error adding orderer: %v", err)
	}

	err = channel.Initialize(nil)
	if err != nil {
		t.Fatalf("channel Initialize failed : %v", err)
	}
	orgUnits, err = channel.OrganizationUnits()
	if err != nil {
		t.Fatalf("CANNOT retrieve organizational units : %v", err)
	}
	if !isValueInList(channel.Name(), orgUnits) {
		t.Fatalf("Could not find %s in the list of organizations", channel.Name())
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

func TestChannelInitialize(t *testing.T) {
	org1MSPID := "ORG1MSP"
	org2MSPID := "ORG2MSP"

	channel, _ := setupTestChannel()
	builder := &mocks.MockConfigUpdateEnvelopeBuilder{
		ChannelID: "mychannel",
		MockConfigGroupBuilder: mocks.MockConfigGroupBuilder{
			ModPolicy: "Admins",
			MSPNames: []string{
				org1MSPID,
				org2MSPID,
			},
			OrdererAddress: "localhost:7054",
			RootCA:         validRootCA,
		},
	}

	err := channel.Initialize(builder.BuildBytes())
	if err != nil {
		t.Fatalf("channel Initialize failed : %v", err)
	}

	mspManager := channel.MSPManager()
	if mspManager == nil {
		t.Fatalf("nil MSPManager on new channel")
	}

}

//func TestChannelInitializeFromUpdate(t *testing.T) {
//	org1MSPID := "ORG1MSP"
//	org2MSPID := "ORG2MSP"
//
//	client := mocks.NewMockClient()
//	user := mocks.NewMockUser("test", )
//	cryptoSuite := &mocks.MockCryptoSuite{}
//	client.SaveUserToStateStore(user, true)
//	client.SetCryptoSuite(cryptoSuite)
//	channel, _ := NewChannel("testChannel", client)
//
//	builder := &mocks.MockConfigUpdateEnvelopeBuilder{
//		ChannelID: "mychannel",
//		MockConfigGroupBuilder: mocks.MockConfigGroupBuilder{
//			ModPolicy: "Admins",
//			MSPNames: []string{
//				org1MSPID,
//				org2MSPID,
//			},
//			OrdererAddress: "localhost:7054",
//			RootCA:         validRootCA,
//		},
//	}
//
//	err := channel.Initialize(builder.BuildBytes())
//	if err != nil {
//		t.Fatalf("channel Initialize failed : %v", err)
//	}
//
//	mspManager := channel.MSPManager()
//	if mspManager == nil {
//		t.Fatalf("nil MSPManager on new channel")
//	}
//
//}

func setupTestChannel() (fab.Channel, error) {
	client := mocks.NewMockClient()
	user := mocks.NewMockUser("test")
	cryptoSuite := &mocks.MockCryptoSuite{}
	client.SaveUserToStateStore(user, true)
	client.SetCryptoSuite(cryptoSuite)
	return NewChannel("testChannel", client)
}

func setupMassiveTestChannel(numberOfPeers int, numberOfOrderers int) (fab.Channel, error) {
	channel, error := setupTestChannel()
	if error != nil {
		return channel, error
	}

	for i := 0; i < numberOfPeers; i++ {
		peer := mocks.MockPeer{MockName: fmt.Sprintf("MockPeer%d", i), MockURL: fmt.Sprintf("http://mock%d.peers.r.us", i),
			MockRoles: []string{}, MockCert: nil}
		err := channel.AddPeer(&peer)
		if err != nil {
			return nil, fmt.Errorf("Error adding peer: %v", err)
		}
	}

	for i := 0; i < numberOfOrderers; i++ {
		orderer := mocks.NewMockOrderer(fmt.Sprintf("http://mock%d.orderers.r.us", i), nil)
		err := channel.AddOrderer(orderer)
		if err != nil {
			return nil, fmt.Errorf("Error adding orderer: %v", err)
		}
	}

	return channel, error
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
