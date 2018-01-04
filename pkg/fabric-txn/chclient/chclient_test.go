/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package chclient

import (
	"strings"
	"testing"
	"time"

	"github.com/hyperledger/fabric-sdk-go/api/apifabclient"
	"github.com/hyperledger/fabric-sdk-go/api/apitxn"

	"github.com/hyperledger/fabric-sdk-go/pkg/errors"
	"github.com/hyperledger/fabric-sdk-go/pkg/fabric-client/channel"
	fcmocks "github.com/hyperledger/fabric-sdk-go/pkg/fabric-client/mocks"
	"github.com/hyperledger/fabric-sdk-go/pkg/fabric-client/peer"
	txnmocks "github.com/hyperledger/fabric-sdk-go/pkg/fabric-txn/mocks"
)

func TestTxProposalResponseFilter(t *testing.T) {
	// failed if status not 200
	testPeer1 := fcmocks.NewMockPeer("Peer1", "http://peer1.com")
	testPeer2 := fcmocks.NewMockPeer("Peer2", "http://peer2.com")
	testPeer2.Status = 300
	peers := []apifabclient.Peer{testPeer1, testPeer2}
	chClient := setupChannelClient(peers, t)

	_, err := chClient.Query(apitxn.QueryRequest{ChaincodeID: "testCC", Fcn: "invoke", Args: [][]byte{[]byte("query"), []byte("b")}})
	if err == nil {
		t.Fatalf("Should have failed for not success status")
	}
	if !strings.Contains(err.Error(), "proposal response was not successful, error code 300") {
		t.Fatalf("Return wrong error message %v", err)
	}

	testPeer2.Payload = []byte("wrongPayload")
	testPeer2.Status = 200
	peers = []apifabclient.Peer{testPeer1, testPeer2}
	chClient = setupChannelClient(peers, t)
	_, err = chClient.Query(apitxn.QueryRequest{ChaincodeID: "testCC", Fcn: "invoke", Args: [][]byte{[]byte("query"), []byte("b")}})
	if err == nil {
		t.Fatalf("Should have failed for not success status")
	}
	if !strings.Contains(err.Error(), "ProposalResponsePayloads do not match") {
		t.Fatalf("Return wrong error message %v", err)
	}

}

func TestQuery(t *testing.T) {

	chClient := setupChannelClient(nil, t)

	result, err := chClient.Query(apitxn.QueryRequest{})
	if err == nil {
		t.Fatalf("Should have failed for empty query request")
	}

	result, err = chClient.Query(apitxn.QueryRequest{Fcn: "invoke", Args: [][]byte{[]byte("query"), []byte("b")}})
	if err == nil {
		t.Fatalf("Should have failed for empty chaincode ID")
	}

	result, err = chClient.Query(apitxn.QueryRequest{ChaincodeID: "testCC", Args: [][]byte{[]byte("query"), []byte("b")}})
	if err == nil {
		t.Fatalf("Should have failed for empty function")
	}

	result, err = chClient.Query(apitxn.QueryRequest{ChaincodeID: "testCC", Fcn: "invoke", Args: [][]byte{[]byte("query"), []byte("b")}})
	if err != nil {
		t.Fatalf("Failed to invoke test cc: %s", err)
	}

	if result != nil {
		t.Fatalf("Expecting nil, got %s", result)
	}

}

func TestQueryDiscoveryError(t *testing.T) {

	chClient := setupChannelClientWithError(errors.New("Test Error"), nil, nil, t)

	_, err := chClient.Query(apitxn.QueryRequest{ChaincodeID: "testCC", Fcn: "invoke", Args: [][]byte{[]byte("query"), []byte("b")}})
	if err == nil {
		t.Fatalf("Should have failed to query with error in discovery.GetPeers()")
	}

}

func TestQuerySelectionError(t *testing.T) {

	chClient := setupChannelClientWithError(nil, errors.New("Test Error"), nil, t)

	_, err := chClient.Query(apitxn.QueryRequest{ChaincodeID: "testCC", Fcn: "invoke", Args: [][]byte{[]byte("query"), []byte("b")}})
	if err == nil {
		t.Fatalf("Should have failed to query with error in selection.GetEndorsersFor ...")
	}

}

func TestQueryWithOptSync(t *testing.T) {

	chClient := setupChannelClient(nil, t)

	result, err := chClient.QueryWithOpts(apitxn.QueryRequest{ChaincodeID: "testCC", Fcn: "invoke", Args: [][]byte{[]byte("query"), []byte("b")}}, apitxn.QueryOpts{})
	if err != nil {
		t.Fatalf("Failed to invoke test cc: %s", err)
	}

	if result != nil {
		t.Fatalf("Expecting nil, got %s", result)
	}
}

func TestQueryWithOptAsync(t *testing.T) {

	chClient := setupChannelClient(nil, t)

	notifier := make(chan apitxn.QueryResponse)

	result, err := chClient.QueryWithOpts(apitxn.QueryRequest{ChaincodeID: "testCC", Fcn: "invoke", Args: [][]byte{[]byte("query"), []byte("b")}}, apitxn.QueryOpts{Notifier: notifier})
	if err != nil {
		t.Fatalf("Failed to invoke test cc: %s", err)
	}

	if result != nil {
		t.Fatalf("Expecting nil, got %s", result)
	}

	select {
	case response := <-notifier:
		if response.Error != nil {
			t.Fatalf("Query returned error: %s", response.Error)
		}
		if string(response.Response) != "" {
			t.Fatalf("Expecting empty, got %s", response.Response)
		}
	case <-time.After(time.Second * 20):
		t.Fatalf("Query Request timed out")
	}

}

func TestQueryWithOptTarget(t *testing.T) {

	chClient := setupChannelClient(nil, t)

	testPeer := fcmocks.NewMockPeer("Peer1", "http://peer1.com")

	peers := []apifabclient.Peer{testPeer}

	targets := peer.PeersToTxnProcessors(peers)

	result, err := chClient.QueryWithOpts(apitxn.QueryRequest{ChaincodeID: "testCC", Fcn: "invoke", Args: [][]byte{[]byte("query"), []byte("b")}}, apitxn.QueryOpts{ProposalProcessors: targets})
	if err != nil {
		t.Fatalf("Failed to invoke test cc: %s", err)
	}

	if result != nil {
		t.Fatalf("Expecting nil, got %s", result)
	}
}

func TestExecuteTx(t *testing.T) {

	chClient := setupChannelClient(nil, t)

	_, err := chClient.ExecuteTx(apitxn.ExecuteTxRequest{})
	if err == nil {
		t.Fatalf("Should have failed for empty invoke request")
	}

	_, err = chClient.ExecuteTx(apitxn.ExecuteTxRequest{Fcn: "invoke", Args: [][]byte{[]byte("move"), []byte("a"), []byte("b"), []byte("1")}})
	if err == nil {
		t.Fatalf("Should have failed for empty chaincode ID")
	}

	_, err = chClient.ExecuteTx(apitxn.ExecuteTxRequest{ChaincodeID: "testCC", Args: [][]byte{[]byte("move"), []byte("a"), []byte("b"), []byte("1")}})
	if err == nil {
		t.Fatalf("Should have failed for empty function")
	}

	// TODO: Test Valid Scenario with mocks
}

func TestExecuteTxDiscoveryError(t *testing.T) {

	chClient := setupChannelClientWithError(errors.New("Test Error"), nil, nil, t)

	_, err := chClient.ExecuteTx(apitxn.ExecuteTxRequest{ChaincodeID: "testCC", Fcn: "invoke", Args: [][]byte{[]byte("move"), []byte("a"), []byte("b"), []byte("1")}})
	if err == nil {
		t.Fatalf("Should have failed to execute tx with error in discovery.GetPeers()")
	}

}

func TestExecuteTxSelectionError(t *testing.T) {

	chClient := setupChannelClientWithError(nil, errors.New("Test Error"), nil, t)

	_, err := chClient.ExecuteTx(apitxn.ExecuteTxRequest{ChaincodeID: "testCC", Fcn: "invoke", Args: [][]byte{[]byte("move"), []byte("a"), []byte("b"), []byte("1")}})
	if err == nil {
		t.Fatalf("Should have failed to execute tx with error in selection.GetEndorserrsFor ...")
	}

}

func setupTestChannel() (*channel.Channel, error) {
	client := setupTestClient()
	return channel.NewChannel("testChannel", client)
}

func setupTestClient() *fcmocks.MockClient {
	client := fcmocks.NewMockClient()
	user := fcmocks.NewMockUser("test")
	cryptoSuite := &fcmocks.MockCryptoSuite{}
	client.SaveUserToStateStore(user, true)
	client.SetUserContext(user)
	client.SetCryptoSuite(cryptoSuite)
	return client
}

func setupTestDiscovery(discErr error, peers []apifabclient.Peer) (apifabclient.DiscoveryService, error) {

	mockDiscovery, err := txnmocks.NewMockDiscoveryProvider(discErr, peers)
	if err != nil {
		return nil, errors.WithMessage(err, "NewMockDiscoveryProvider failed")
	}

	return mockDiscovery.NewDiscoveryService("mychannel")
}

func setupTestSelection(discErr error, peers []apifabclient.Peer) (apifabclient.SelectionService, error) {

	mockSelection, err := txnmocks.NewMockSelectionProvider(discErr, peers)
	if err != nil {
		return nil, errors.WithMessage(err, "NewMockSelectinProvider failed")
	}

	return mockSelection.NewSelectionService("mychannel")
}

func setupChannelClient(peers []apifabclient.Peer, t *testing.T) *ChannelClient {

	return setupChannelClientWithError(nil, nil, peers, t)
}

func setupChannelClientWithError(discErr error, selectionErr error, peers []apifabclient.Peer, t *testing.T) *ChannelClient {

	fcClient := setupTestClient()

	testChannel, err := setupTestChannel()
	if err != nil {
		t.Fatalf("Failed to setup test channel: %s", err)
	}

	orderer := fcmocks.NewMockOrderer("", nil)
	testChannel.AddOrderer(orderer)

	discoveryService, err := setupTestDiscovery(discErr, nil)
	if err != nil {
		t.Fatalf("Failed to setup discovery service: %s", err)
	}

	selectionService, err := setupTestSelection(selectionErr, peers)
	if err != nil {
		t.Fatalf("Failed to setup discovery service: %s", err)
	}

	ch, err := NewChannelClient(fcClient, testChannel, discoveryService, selectionService, nil)
	if err != nil {
		t.Fatalf("Failed to create new channel client: %s", err)
	}

	return ch
}
