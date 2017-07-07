/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/
package channel

import (
	"reflect"
	"testing"

	"github.com/hyperledger/fabric-sdk-go/api/apitxn"
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

	badRequest1 := apitxn.ChaincodeInvokeRequest{
		Fcn:  "method",
		Args: []string{"arg"},
	}
	_, err = channel.QueryByChaincode(badRequest1)
	if err == nil {
		t.Fatalf("QueryByChannelcode: name cannot be empty")
	}

	badRequest2 := apitxn.ChaincodeInvokeRequest{
		ChaincodeID: "qscc",
	}
	_, err = channel.QueryByChaincode(badRequest2)
	if err == nil {
		t.Fatalf("QueryByChannelcode: arguments cannot be empty")
	}

	badRequest3 := apitxn.ChaincodeInvokeRequest{
		ChaincodeID: "qscc",
		Fcn:         "method",
		Args:        []string{"arg"},
	}
	_, err = channel.QueryByChaincode(badRequest3)
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

	request := apitxn.ChaincodeInvokeRequest{
		ChaincodeID: "cc",
		Fcn:         "Hello",
	}
	_, err := channel.QueryByChaincode(request)
	if err == nil {
		t.Fatalf("Expected error")
	}
	_, err = queryByChaincode("", request, channel.ClientContext())
	if err == nil {
		t.Fatalf("Expected error")
	}

	peer := mocks.MockPeer{MockName: "Peer1", MockURL: "http://peer1.com", MockRoles: []string{}, MockCert: nil, Payload: []byte("A")}
	channel.AddPeer(&peer)

	request = apitxn.ChaincodeInvokeRequest{
		Fcn: "Hello",
	}
	_, err = channel.QueryByChaincode(request)
	if err == nil {
		t.Fatalf("Expected error")
	}

	request = apitxn.ChaincodeInvokeRequest{
		ChaincodeID: "cc",
	}
	_, err = channel.QueryByChaincode(request)
	if err == nil {
		t.Fatalf("Expected error")
	}

	request = apitxn.ChaincodeInvokeRequest{
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

	request := apitxn.ChaincodeInvokeRequest{
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
