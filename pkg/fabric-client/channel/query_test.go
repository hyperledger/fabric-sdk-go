/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/
package channel

import (
	"reflect"
	"testing"

	"github.com/golang/protobuf/proto"
	fab "github.com/hyperledger/fabric-sdk-go/api/apifabclient"
	"github.com/hyperledger/fabric-sdk-go/pkg/fabric-client/mocks"
)

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

	badRequest1 := fab.ChaincodeInvokeRequest{
		Fcn:  "method",
		Args: [][]byte{[]byte("arg")},
	}
	_, err = channel.QueryByChaincode(badRequest1)
	if err == nil {
		t.Fatalf("QueryByChannelcode: name cannot be empty")
	}

	badRequest2 := fab.ChaincodeInvokeRequest{
		ChaincodeID: "qscc",
	}
	_, err = channel.QueryByChaincode(badRequest2)
	if err == nil {
		t.Fatalf("QueryByChannelcode: arguments cannot be empty")
	}

	badRequest3 := fab.ChaincodeInvokeRequest{
		ChaincodeID: "qscc",
		Fcn:         "method",
		Args:        [][]byte{[]byte("arg")},
	}
	_, err = channel.QueryByChaincode(badRequest3)
	if err == nil {
		t.Fatalf("QueryByChannelcode: targets cannot be empty")
	}

}

func TestQueryOnSystemChannel(t *testing.T) {
	channel, _ := setupChannel(systemChannel)
	peer := mocks.MockPeer{MockName: "Peer1", MockURL: "http://peer1.com", MockRoles: []string{}, MockCert: nil}
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

func TestQueryInstantiatedChaincodes(t *testing.T) {
	channel, _ := setupTestChannel()

	peer := mocks.MockPeer{MockName: "Peer1", MockURL: "http://peer1.com", MockRoles: []string{}, MockCert: nil}
	err := channel.AddPeer(&peer)

	res, err := channel.QueryInstantiatedChaincodes()

	if err != nil || res == nil {
		t.Fatalf("Test QueryInstatiated chaincode failed: %v", err)
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
		t.Fatalf("Test QueryInfo failed: %v", err)
	}
}

func TestQueryMissingParams(t *testing.T) {
	channel, _ := setupTestChannel()

	request := fab.ChaincodeInvokeRequest{
		ChaincodeID: "cc",
		Fcn:         "Hello",
	}
	_, err := channel.QueryByChaincode(request)
	if err == nil {
		t.Fatalf("Expected error")
	}
	_, err = queryByChaincode(channel.clientContext, "", request, request.Targets)
	if err == nil {
		t.Fatalf("Expected error")
	}

	peer := mocks.MockPeer{MockName: "Peer1", MockURL: "http://peer1.com", MockRoles: []string{}, MockCert: nil, Payload: []byte("A")}
	channel.AddPeer(&peer)

	request = fab.ChaincodeInvokeRequest{
		Fcn: "Hello",
	}
	_, err = channel.QueryByChaincode(request)
	if err == nil {
		t.Fatalf("Expected error")
	}

	request = fab.ChaincodeInvokeRequest{
		ChaincodeID: "cc",
	}
	_, err = channel.QueryByChaincode(request)
	if err == nil {
		t.Fatalf("Expected error")
	}

	request = fab.ChaincodeInvokeRequest{
		ChaincodeID: "cc",
		Fcn:         "Hello",
	}
	_, err = channel.QueryByChaincode(request)
	if err != nil {
		t.Fatalf("Expected success")
	}
}

func TestQueryBySystemChaincode(t *testing.T) {
	channel, _ := setupTestChannel()

	peer := mocks.MockPeer{MockName: "Peer1", MockURL: "http://peer1.com", MockRoles: []string{}, MockCert: nil, Payload: []byte("A")}
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

func TestQueryConfig(t *testing.T) {
	channel, _ := setupTestChannel()

	// empty targets
	_, err := channel.QueryConfigBlock([]fab.Peer{}, 1)
	if err == nil {
		t.Fatalf("Should have failed due to empty targets")
	}

	// min endorsers <= 0
	_, err = channel.QueryConfigBlock([]fab.Peer{mocks.NewMockPeer("Peer1", "http://peer1.com")}, 0)
	if err == nil {
		t.Fatalf("Should have failed due to empty targets")
	}

	// peer without payload
	_, err = channel.QueryConfigBlock([]fab.Peer{mocks.NewMockPeer("Peer1", "http://peer1.com")}, 1)
	if err == nil {
		t.Fatalf("Should have failed due to nil block metadata")
	}

	// create config block builder in order to create valid payload
	builder := &mocks.MockConfigBlockBuilder{
		MockConfigGroupBuilder: mocks.MockConfigGroupBuilder{
			ModPolicy: "Admins",
			MSPNames: []string{
				"Org1MSP",
				"Org2MSP",
			},
			OrdererAddress: "localhost:7054",
			RootCA:         validRootCA,
		},
		Index:           0,
		LastConfigIndex: 0,
	}

	payload, err := proto.Marshal(builder.Build())
	if err != nil {
		t.Fatalf("Failed to marshal mock block")
	}

	// peer with valid config block payload
	peer := mocks.MockPeer{MockName: "Peer1", MockURL: "http://peer1.com", MockRoles: []string{}, MockCert: nil, Payload: payload}

	// fail with min endorsers
	res, err := channel.QueryConfigBlock([]fab.Peer{&peer}, 2)
	if err == nil {
		t.Fatalf("Should have failed with since there's one endorser and at least two are required")
	}

	// success with one endorser
	res, err = channel.QueryConfigBlock([]fab.Peer{&peer}, 1)
	if err != nil || res == nil {
		t.Fatalf("Test QueryConfig failed: %v", err)
	}

	// create second endorser with same payload
	peer2 := mocks.MockPeer{MockName: "Peer2", MockURL: "http://peer2.com", MockRoles: []string{}, MockCert: nil, Payload: payload}

	// success with two endorsers
	res, err = channel.QueryConfigBlock([]fab.Peer{&peer, &peer2}, 2)
	if err != nil || res == nil {
		t.Fatalf("Test QueryConfig failed: %v", err)
	}

	// Create different config block payload
	builder2 := &mocks.MockConfigBlockBuilder{
		MockConfigGroupBuilder: mocks.MockConfigGroupBuilder{
			ModPolicy: "Admins",
			MSPNames: []string{
				"Org1MSP",
			},
			OrdererAddress: "builder2:7054",
			RootCA:         validRootCA,
		},
		Index:           0,
		LastConfigIndex: 0,
	}

	payload2, err := proto.Marshal(builder2.Build())
	if err != nil {
		t.Fatalf("Failed to marshal mock block 2")
	}

	// peer 2 now had different payload; query config block should fail
	peer2.Payload = payload2
	res, err = channel.QueryConfigBlock([]fab.Peer{&peer, &peer2}, 2)
	if err == nil {
		t.Fatalf("Should have failed for different block payloads")
	}

}
