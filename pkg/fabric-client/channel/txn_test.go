/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/
package channel

import (
	"net"
	"strings"
	"testing"

	"google.golang.org/grpc"

	"github.com/golang/mock/gomock"
	fab "github.com/hyperledger/fabric-sdk-go/api/apifabclient"
	mock_fab "github.com/hyperledger/fabric-sdk-go/api/apifabclient/mocks"
	"github.com/hyperledger/fabric-sdk-go/third_party/github.com/hyperledger/fabric/common/cauthdsl"
	"github.com/hyperledger/fabric-sdk-go/third_party/github.com/hyperledger/fabric/protos/common"
	pb "github.com/hyperledger/fabric-sdk-go/third_party/github.com/hyperledger/fabric/protos/peer"

	"github.com/hyperledger/fabric-sdk-go/pkg/fabric-client/mocks"
)

func TestAddPeerDuplicateCheck(t *testing.T) {
	channel, _ := setupTestChannel()

	peer := mocks.MockPeer{MockName: "Peer1", MockURL: "http://peer1.com", MockRoles: []string{}, MockCert: nil}
	channel.AddPeer(&peer)

	err := channel.AddPeer(&peer)

	if err == nil || !strings.Contains(err.Error(), "http://peer1.com already exists") {
		t.Fatal("Duplicate Peer check is not working as expected")
	}
}

func TestSendInstantiateProposal(t *testing.T) {
	//Setup channel
	user := mocks.NewMockUserWithMSPID("test", "1234")
	ctx := mocks.NewMockContext(user)
	channel, _ := New(ctx, mocks.NewMockChannelCfg("testChannel"))

	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()
	proc := mock_fab.NewMockProposalProcessor(mockCtrl)

	tp := fab.TransactionProposal{SignedProposal: &pb.SignedProposal{}}
	tpr := fab.TransactionProposalResponse{Endorser: "example.com", Status: 99, Proposal: tp, ProposalResponse: nil}

	proc.EXPECT().ProcessTransactionProposal(gomock.Any()).Return(tpr, nil)
	proc.EXPECT().ProcessTransactionProposal(gomock.Any()).Return(tpr, nil)
	targets := []fab.ProposalProcessor{proc}

	//Add a Peer
	peer := mocks.MockPeer{MockName: "Peer1", MockURL: "http://peer1.com", MockRoles: []string{}, MockCert: nil}
	channel.AddPeer(&peer)

	tresponse, txnid, err := channel.SendInstantiateProposal("", nil, "",
		"", cauthdsl.SignedByMspMember("Org1MSP"), nil, targets)

	if err == nil || err.Error() != "chaincodeName is required" {
		t.Fatal("Validation for chain code name parameter for send Instantiate Proposal failed")
	}

	tresponse, txnid, err = channel.SendInstantiateProposal("qscc", nil, "",
		"", cauthdsl.SignedByMspMember("Org1MSP"), nil, targets)

	tresponse, txnid, err = channel.SendInstantiateProposal("qscc", nil, "",
		"", cauthdsl.SignedByMspMember("Org1MSP"), nil, targets)

	if err == nil || err.Error() != "chaincodePath is required" {
		t.Fatal("Validation for chain code path for send Instantiate Proposal failed")
	}

	tresponse, txnid, err = channel.SendInstantiateProposal("qscc", nil, "test",
		"", cauthdsl.SignedByMspMember("Org1MSP"), nil, targets)

	if err == nil || err.Error() != "chaincodeVersion is required" {
		t.Fatal("Validation for chain code version for send Instantiate Proposal failed")
	}

	tresponse, txnid, err = channel.SendInstantiateProposal("qscc", nil, "test",
		"1", nil, nil, nil)
	if err == nil || err.Error() != "chaincodePolicy is required" {
		t.Fatal("Validation for chain code policy for send Instantiate Proposal failed")
	}

	tresponse, txnid, err = channel.SendInstantiateProposal("qscc", nil, "test",
		"1", cauthdsl.SignedByMspMember("Org1MSP"), nil, targets)

	if err != nil || len(tresponse) == 0 || txnid.ID == "" {
		t.Fatal("Send Instantiate Proposal Test failed")
	}

	tresponse, txnid, err = channel.SendInstantiateProposal("qscc", nil, "test",
		"1", cauthdsl.SignedByMspMember("Org1MSP"), nil, nil)

	if err == nil || err.Error() != "missing peer objects for chaincode proposal" {
		t.Fatal("Missing peer objects validation is not working as expected")
	}

	// Define the private data collection policy config
	collConfig := []*common.CollectionConfig{
		newCollectionConfig("somecollection", 1, 3, cauthdsl.SignedByAnyMember([]string{"Org1MSP", "Org2MSP"})),
	}
	tresponse, txnid, err = channel.SendInstantiateProposal("qscc", nil, "test",
		"1", cauthdsl.SignedByMspMember("Org1MSP"), collConfig, targets)
	if err != nil || len(tresponse) == 0 || txnid.ID == "" {
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

	tp := fab.TransactionProposal{SignedProposal: &pb.SignedProposal{}}
	tpr := fab.TransactionProposalResponse{Endorser: "example.com", Status: 99, Proposal: tp, ProposalResponse: nil}

	proc.EXPECT().ProcessTransactionProposal(gomock.Any()).Return(tpr, nil)
	targets := []fab.ProposalProcessor{proc}

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
