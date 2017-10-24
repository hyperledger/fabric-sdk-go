/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package chclient

import (
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

func TestQuery(t *testing.T) {

	chClient := setupChannelClient(t)

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

	chClient := setupChannelClientWithError(errors.New("Test Error"), nil, t)

	_, err := chClient.Query(apitxn.QueryRequest{ChaincodeID: "testCC", Fcn: "invoke", Args: [][]byte{[]byte("query"), []byte("b")}})
	if err == nil {
		t.Fatalf("Should have failed to query with error in discovery.GetPeers()")
	}

}

func TestQuerySelectionError(t *testing.T) {

	chClient := setupChannelClientWithError(nil, errors.New("Test Error"), t)

	_, err := chClient.Query(apitxn.QueryRequest{ChaincodeID: "testCC", Fcn: "invoke", Args: [][]byte{[]byte("query"), []byte("b")}})
	if err == nil {
		t.Fatalf("Should have failed to query with error in selection.GetEndorsersFor ...")
	}

}

func TestQueryWithOptSync(t *testing.T) {

	chClient := setupChannelClient(t)

	result, err := chClient.QueryWithOpts(apitxn.QueryRequest{ChaincodeID: "testCC", Fcn: "invoke", Args: [][]byte{[]byte("query"), []byte("b")}}, apitxn.QueryOpts{})
	if err != nil {
		t.Fatalf("Failed to invoke test cc: %s", err)
	}

	if result != nil {
		t.Fatalf("Expecting nil, got %s", result)
	}
}

func TestQueryWithOptAsync(t *testing.T) {

	chClient := setupChannelClient(t)

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

	chClient := setupChannelClient(t)

	testPeer := fcmocks.MockPeer{MockName: "Peer1", MockURL: "http://peer1.com", MockRoles: []string{}, MockCert: nil}

	peers := []apifabclient.Peer{&testPeer}

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

	chClient := setupChannelClient(t)

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

	chClient := setupChannelClientWithError(errors.New("Test Error"), nil, t)

	_, err := chClient.ExecuteTx(apitxn.ExecuteTxRequest{ChaincodeID: "testCC", Fcn: "invoke", Args: [][]byte{[]byte("move"), []byte("a"), []byte("b"), []byte("1")}})
	if err == nil {
		t.Fatalf("Should have failed to execute tx with error in discovery.GetPeers()")
	}

}

func TestExecuteTxSelectionError(t *testing.T) {

	chClient := setupChannelClientWithError(nil, errors.New("Test Error"), t)

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

func setupChannelClient(t *testing.T) *ChannelClient {

	return setupChannelClientWithError(nil, nil, t)
}

func setupChannelClientWithError(discErr error, selectionErr error, t *testing.T) *ChannelClient {

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

	selectionService, err := setupTestSelection(selectionErr, nil)
	if err != nil {
		t.Fatalf("Failed to setup discovery service: %s", err)
	}

	ch, err := NewChannelClient(fcClient, testChannel, discoveryService, selectionService, nil)
	if err != nil {
		t.Fatalf("Failed to create new channel client: %s", err)
	}

	return ch
}
