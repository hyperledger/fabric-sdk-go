// +build deprecated

/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/
package channel

import (
	"fmt"
	"net"
	"reflect"
	"strings"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/hyperledger/fabric-sdk-go/pkg/context/api/fab"
	mock_fab "github.com/hyperledger/fabric-sdk-go/pkg/context/api/fab/mocks"
	mocks "github.com/hyperledger/fabric-sdk-go/pkg/fab/mocks"
	"github.com/hyperledger/fabric-sdk-go/third_party/github.com/hyperledger/fabric/common/cauthdsl"
	"github.com/hyperledger/fabric-sdk-go/third_party/github.com/hyperledger/fabric/protos/common"
	pb "github.com/hyperledger/fabric-sdk-go/third_party/github.com/hyperledger/fabric/protos/peer"
	"github.com/pkg/errors"
	"google.golang.org/grpc"
)

var testAddress = "127.0.0.1:0"

func TestChannelMethods(t *testing.T) {
	user := mocks.NewMockUser("test")
	ctx := mocks.NewMockContext(user)
	channel, err := New(ctx, mocks.NewMockChannelCfg("testChannel"))
	if err != nil {
		t.Fatalf("New return error[%s]", err)
	}
	if channel.Name() != "testChannel" {
		t.Fatalf("New create wrong channel")
	}

	_, err = New(ctx, mocks.NewMockChannelCfg(""))
	if err != nil {
		t.Fatalf("Got error creating channel with empty channel ID: %s", err)
	}

	_, err = New(nil, mocks.NewMockChannelCfg("testChannel"))
	if err == nil {
		t.Fatalf("NewChannel didn't return error")
	}
	if err.Error() != "client is required" {
		t.Fatalf("NewChannel didn't return right error")
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

func TestAddAndRemovePeers(t *testing.T) {
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
	if len(channel.Peers()) != 1 {
		t.Fatal("Add Peer failed")
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
	if primary.URL() != peer2.URL() {
		t.Fatalf("Primary and our choice are not equal")
	}

}

func TestQueryOnSystemChannel(t *testing.T) {
	channel, _ := setupChannel(fab.SystemChannel)
	peer := mocks.MockPeer{MockName: "Peer1", MockURL: "http://peer1.com", MockRoles: []string{}, MockCert: nil, Status: 200}
	err := channel.AddPeer(&peer)
	if err != nil {
		t.Fatalf("Error adding peer to channel: %s", err)
	}

	request := fab.ChaincodeInvokeRequest{
		ChaincodeID: "ccID",
		Fcn:         "method",
		Args:        [][]byte{[]byte("arg")},
	}
	if _, err := channel.QueryByChaincode(request); err != nil {
		t.Fatalf("Error invoking chaincode on system channel: %s", err)
	}
}

func TestQueryBySystemChaincode(t *testing.T) {
	channel, _ := setupTestChannel()

	peer := mocks.MockPeer{MockName: "Peer1", MockURL: "http://peer1.com", MockRoles: []string{}, MockCert: nil, Payload: []byte("A"), Status: 200}
	channel.AddPeer(&peer)

	request := fab.ChaincodeInvokeRequest{
		ChaincodeID: "cc",
		Fcn:         "Hello",
	}
	resp, err := channel.QueryBySystemChaincode(request)
	if err != nil {
		t.Fatalf("Failed to query: %s", err)
	}
	expectedResp := []byte("A")

	if !reflect.DeepEqual(resp[0], expectedResp) {
		t.Fatalf("Unexpected transaction proposal response: %v", resp)
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

func setupTestChannel() (*Channel, error) {
	return setupChannel("testChannel")
}

func setupChannel(channelID string) (*Channel, error) {
	user := mocks.NewMockUser("test")
	ctx := mocks.NewMockContext(user)
	return New(ctx, mocks.NewMockChannelCfg(channelID))
}

func setupMassiveTestChannel(numberOfPeers int, numberOfOrderers int) (*Channel, error) {
	channel, error := setupTestChannel()
	if error != nil {
		return channel, error
	}

	for i := 0; i < numberOfPeers; i++ {
		peer := mocks.MockPeer{MockName: fmt.Sprintf("MockPeer%d", i), MockURL: fmt.Sprintf("http://mock%d.peers.r.us", i),
			MockRoles: []string{}, MockCert: nil}
		err := channel.AddPeer(&peer)
		if err != nil {
			return nil, errors.WithMessage(err, "failed to add peer")
		}
	}

	for i := 0; i < numberOfOrderers; i++ {
		orderer := mocks.NewMockOrderer(fmt.Sprintf("http://mock%d.orderers.r.us", i), nil)
		err := channel.AddOrderer(orderer)
		if err != nil {
			return nil, errors.WithMessage(err, "failed to add orderer")
		}
	}

	return channel, error
}

func TestAddPeerDuplicateCheck(t *testing.T) {
	channel, _ := setupTestChannel()

	peer := mocks.MockPeer{MockName: "Peer1", MockURL: "http://peer1.com", MockRoles: []string{}, MockCert: nil}
	channel.AddPeer(&peer)

	err := channel.AddPeer(&peer)

	if err == nil || !strings.Contains(err.Error(), "http://peer1.com already exists") {
		t.Fatal("Duplicate Peer check is not working as expected")
	}
}

func TestChannelConfigs(t *testing.T) {

	user := mocks.NewMockUser("test")
	ctx := mocks.NewMockContext(user)

	channel, _ := New(ctx, mocks.NewMockChannelCfg("testChannel"))

	if channel.IsReadonly() {
		//TODO: Rightnow it is returning false always, need to revisit test once actual implementation is provided
		t.Fatal("Is Readonly test failed")
	}

	if channel.UpdateChannel() {
		//TODO: Rightnow it is returning false always, need to revisit test once actual implementation is provided
		t.Fatal("UpdateChannel test failed")
	}

	channel.SetMSPManager(nil)

}

func TestSendInstantiateProposal(t *testing.T) {
	//Setup channel
	user := mocks.NewMockUserWithMSPID("test", "1234")
	ctx := mocks.NewMockContext(user)
	channel, _ := New(ctx, mocks.NewMockChannelCfg("testChannel"))

	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()
	proc := mock_fab.NewMockProposalProcessor(mockCtrl)

	tpr := fab.TransactionProposalResponse{Endorser: "example.com", Status: 99}

	proc.EXPECT().ProcessTransactionProposal(gomock.Any()).Return(&tpr, nil)
	proc.EXPECT().ProcessTransactionProposal(gomock.Any()).Return(&tpr, nil)
	targets := []fab.ProposalProcessor{proc}

	//Add a Peer
	peer := mocks.MockPeer{MockName: "Peer1", MockURL: "http://peer1.com", MockRoles: []string{}}
	channel.AddPeer(&peer)

	tresponse, _, err := channel.SendInstantiateProposal("", nil, "",
		"", cauthdsl.SignedByMspMember("Org1MSP"), nil, targets)

	if err == nil || err.Error() != "chaincodeName is required" {
		t.Fatal("Validation for chain code name parameter for send Instantiate Proposal failed")
	}

	tresponse, _, err = channel.SendInstantiateProposal("qscc", nil, "",
		"", cauthdsl.SignedByMspMember("Org1MSP"), nil, targets)

	tresponse, _, err = channel.SendInstantiateProposal("qscc", nil, "",
		"", cauthdsl.SignedByMspMember("Org1MSP"), nil, targets)

	if err == nil || err.Error() != "chaincodePath is required" {
		t.Fatal("Validation for chain code path for send Instantiate Proposal failed")
	}

	tresponse, _, err = channel.SendInstantiateProposal("qscc", nil, "test",
		"", cauthdsl.SignedByMspMember("Org1MSP"), nil, targets)

	if err == nil || err.Error() != "chaincodeVersion is required" {
		t.Fatal("Validation for chain code version for send Instantiate Proposal failed")
	}

	tresponse, _, err = channel.SendInstantiateProposal("qscc", nil, "test",
		"1", nil, nil, nil)
	if err == nil || err.Error() != "chaincodePolicy is required" {
		t.Fatal("Validation for chain code policy for send Instantiate Proposal failed")
	}

	tresponse, _, err = channel.SendInstantiateProposal("qscc", nil, "test",
		"1", cauthdsl.SignedByMspMember("Org1MSP"), nil, targets)

	if err != nil || len(tresponse) == 0 {
		t.Fatal("Send Instantiate Proposal Test failed")
	}

	tresponse, _, err = channel.SendInstantiateProposal("qscc", nil, "test",
		"1", cauthdsl.SignedByMspMember("Org1MSP"), nil, nil)

	if err == nil || err.Error() != "missing peer objects for chaincode proposal" {
		t.Fatal("Missing peer objects validation is not working as expected")
	}

	// Define the private data collection policy config
	collConfig := []*common.CollectionConfig{
		newCollectionConfig("somecollection", 1, 3, cauthdsl.SignedByAnyMember([]string{"Org1MSP", "Org2MSP"})),
	}
	tresponse, _, err = channel.SendInstantiateProposal("qscc", nil, "test",
		"1", cauthdsl.SignedByMspMember("Org1MSP"), collConfig, targets)
	if err != nil || len(tresponse) == 0 {
		t.Fatal("Send Instantiate Proposal Test failed")
	}
}

func TestSendUpgradeProposal(t *testing.T) {
	//Setup channel
	user := mocks.NewMockUserWithMSPID("test", "1234")
	ctx := mocks.NewMockContext(user)
	channel, _ := New(ctx, mocks.NewMockChannelCfg("testChannel"))

	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()
	proc := mock_fab.NewMockProposalProcessor(mockCtrl)

	tpr := fab.TransactionProposalResponse{Endorser: "example.com", Status: 99, ProposalResponse: nil}

	proc.EXPECT().ProcessTransactionProposal(gomock.Any()).Return(&tpr, nil)
	targets := []fab.ProposalProcessor{proc}

	//Add a Peer
	peer := mocks.MockPeer{MockName: "Peer1", MockURL: "http://peer1.com", MockRoles: []string{}, MockCert: nil}
	channel.AddPeer(&peer)

	tresponse, _, err := channel.SendUpgradeProposal("", nil, "",
		"", cauthdsl.SignedByMspMember("Org1MSP"), targets)

	if err == nil || err.Error() != "chaincodeName is required" {
		t.Fatal("Validation for chain code name parameter for send Upgrade Proposal failed")
	}

	tresponse, _, err = channel.SendUpgradeProposal("qscc", nil, "",
		"", cauthdsl.SignedByMspMember("Org1MSP"), targets)

	tresponse, _, err = channel.SendUpgradeProposal("qscc", nil, "",
		"", cauthdsl.SignedByMspMember("Org1MSP"), targets)

	if err == nil || err.Error() != "chaincodePath is required" {
		t.Fatal("Validation for chain code path for send Upgrade Proposal failed")
	}

	tresponse, _, err = channel.SendUpgradeProposal("qscc", nil, "test",
		"", cauthdsl.SignedByMspMember("Org1MSP"), targets)

	if err == nil || err.Error() != "chaincodeVersion is required" {
		t.Fatal("Validation for chain code version for send Upgrade Proposal failed")
	}

	tresponse, _, err = channel.SendUpgradeProposal("qscc", nil, "test",
		"2", nil, nil)
	if err == nil || err.Error() != "chaincodePolicy is required" {
		t.Fatal("Validation for chain code policy for send Upgrade Proposal failed")
	}

	tresponse, _, err = channel.SendUpgradeProposal("qscc", nil, "test",
		"2", cauthdsl.SignedByMspMember("Org1MSP"), targets)

	if err != nil || len(tresponse) == 0 {
		t.Fatal("Send Upgrade Proposal Test failed")
	}

	tresponse, _, err = channel.SendUpgradeProposal("qscc", nil, "test",
		"2", cauthdsl.SignedByMspMember("Org1MSP"), nil)
	if err == nil || err.Error() != "missing peer objects for chaincode proposal" {
		t.Fatal("Missing peer objects validation is not working as expected")
	}
}

func startEndorserServer(t *testing.T, grpcServer *grpc.Server) (*mocks.MockEndorserServer, string) {
	lis, err := net.Listen("tcp", testAddress)
	addr := lis.Addr().String()

	endorserServer := &mocks.MockEndorserServer{}
	pb.RegisterEndorserServer(grpcServer, endorserServer)
	if err != nil {
		t.Logf("Error starting test server %s", err)
		t.FailNow()
	}
	t.Logf("Starting test server on %s\n", addr)
	go grpcServer.Serve(lis)
	return endorserServer, addr
}

func newCollectionConfig(collName string, requiredPeerCount, maxPeerCount int32, policy *common.SignaturePolicyEnvelope) *common.CollectionConfig {
	return &common.CollectionConfig{
		Payload: &common.CollectionConfig_StaticCollectionConfig{
			StaticCollectionConfig: &common.StaticCollectionConfig{
				Name:              collName,
				RequiredPeerCount: requiredPeerCount,
				MaximumPeerCount:  maxPeerCount,
				MemberOrgsPolicy: &common.CollectionPolicyConfig{
					Payload: &common.CollectionPolicyConfig_SignaturePolicy{
						SignaturePolicy: policy,
					},
				},
			},
		},
	}
}
