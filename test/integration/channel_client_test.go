/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package integration

import (
	"fmt"
	"testing"
	"time"

	"github.com/hyperledger/fabric-sdk-go/api/apitxn"
	"github.com/hyperledger/fabric-sdk-go/def/fabapi"
	"github.com/hyperledger/fabric-sdk-go/def/fabapi/opt"
)

var queryArgs = [][]byte{[]byte("query"), []byte("b")}
var txArgs = [][]byte{[]byte("move"), []byte("a"), []byte("b"), []byte("1")}

func TestChannelClient(t *testing.T) {

	testSetup := BaseSetupImpl{
		ConfigFile:      ConfigTestFile,
		ChannelID:       "mychannel",
		OrgID:           org1Name,
		ChannelConfig:   "../fixtures/channel/mychannel.tx",
		ConnectEventHub: true,
	}

	if err := testSetup.Initialize(); err != nil {
		t.Fatalf(err.Error())
	}

	if err := testSetup.InstallAndInstantiateExampleCC(); err != nil {
		t.Fatalf("InstallAndInstantiateExampleCC return error: %v", err)
	}

	// Create SDK setup for the integration tests
	sdkOptions := fabapi.Options{
		ConfigFile: testSetup.ConfigFile,
		StateStoreOpts: opt.StateStoreOpts{
			Path: "/tmp/enroll_user",
		},
	}

	sdk, err := fabapi.NewSDK(sdkOptions)
	if err != nil {
		t.Fatalf("Failed to create new SDK: %s", err)
	}

	chClient, err := sdk.NewChannelClient(testSetup.ChannelID, "User1")
	if err != nil {
		t.Fatalf("Failed to create new channel client: %s", err)
	}

	// Synchronous query
	testQuery("200", testSetup.ChainCodeID, chClient, t)

	// Synchronous transaction
	_, err = chClient.ExecuteTx(apitxn.ExecuteTxRequest{ChaincodeID: testSetup.ChainCodeID, Fcn: "invoke", Args: txArgs})
	if err != nil {
		t.Fatalf("Failed to move funds: %s", err)
	}

	// Verify transaction using asynchronous query
	testQueryWithOpts("201", testSetup.ChainCodeID, chClient, t)

	// Asynchronous transaction
	testAsyncTransaction(testSetup.ChainCodeID, chClient, t)

	// Verify asynchronous transaction
	testQuery("202", testSetup.ChainCodeID, chClient, t)

	// Test transaction filter error
	testFilterError(testSetup.ChainCodeID, chClient, t)

	// Test commit error
	testCommitError(testSetup.ChainCodeID, chClient, t)

	// Verify that filter error and commit error did not modify value
	testQuery("202", testSetup.ChainCodeID, chClient, t)

	// Test register and receive chaincode event
	testChaincodeEvent(testSetup.ChainCodeID, chClient, t)

	// Verify transaction with chain code event completed
	testQuery("203", testSetup.ChainCodeID, chClient, t)

	// Release channel client resources
	err = chClient.Close()
	if err != nil {
		t.Fatalf("Failed to close channel client: %v", err)
	}

}

func testQuery(expected string, ccID string, chClient apitxn.ChannelClient, t *testing.T) {

	result, err := chClient.Query(apitxn.QueryRequest{ChaincodeID: ccID, Fcn: "invoke", Args: queryArgs})
	if err != nil {
		t.Fatalf("Failed to invoke example cc: %s", err)
	}

	if string(result) != expected {
		t.Fatalf("Expecting %s, got %s", expected, result)
	}
}

func testQueryWithOpts(expected string, ccID string, chClient apitxn.ChannelClient, t *testing.T) {

	notifier := make(chan apitxn.QueryResponse)
	result, err := chClient.QueryWithOpts(apitxn.QueryRequest{ChaincodeID: ccID, Fcn: "invoke", Args: queryArgs}, apitxn.QueryOpts{Notifier: notifier})
	if err != nil {
		t.Fatalf("Failed to invoke example cc asynchronously: %s", err)
	}
	if result != nil {
		t.Fatalf("Expecting empty, got %s", result)
	}

	select {
	case response := <-notifier:
		if response.Error != nil {
			t.Fatalf("Query returned error: %s", response.Error)
		}
		if string(response.Response) != expected {
			t.Fatalf("Expecting %s, got %s", expected, response.Response)
		}
	case <-time.After(time.Second * 20):
		t.Fatalf("Query Request timed out")
	}

}

func testAsyncTransaction(ccID string, chClient apitxn.ChannelClient, t *testing.T) {

	txNotifier := make(chan apitxn.ExecuteTxResponse)
	txFilter := &TestTxFilter{}
	txOpts := apitxn.ExecuteTxOpts{Notifier: txNotifier, TxFilter: txFilter}

	_, err := chClient.ExecuteTxWithOpts(apitxn.ExecuteTxRequest{ChaincodeID: ccID, Fcn: "invoke", Args: txArgs}, txOpts)
	if err != nil {
		t.Fatalf("Failed to move funds: %s", err)
	}

	select {
	case response := <-txNotifier:
		if response.Error != nil {
			t.Fatalf("ExecuteTx returned error: %s", response.Error)
		}
	case <-time.After(time.Second * 20):
		t.Fatalf("ExecuteTx timed out")
	}
}

func testCommitError(ccID string, chClient apitxn.ChannelClient, t *testing.T) {

	txNotifier := make(chan apitxn.ExecuteTxResponse)

	txFilter := &TestTxFilter{errResponses: fmt.Errorf("Error")}
	txOpts := apitxn.ExecuteTxOpts{Notifier: txNotifier, TxFilter: txFilter}

	_, err := chClient.ExecuteTxWithOpts(apitxn.ExecuteTxRequest{ChaincodeID: ccID, Fcn: "invoke", Args: txArgs}, txOpts)
	if err != nil {
		t.Fatalf("Failed to move funds: %s", err)
	}

	select {
	case response := <-txNotifier:
		if response.Error == nil {
			t.Fatalf("ExecuteTx should have returned an error")
		}
	case <-time.After(time.Second * 20):
		t.Fatalf("ExecuteTx timed out")
	}
}

func testFilterError(ccID string, chClient apitxn.ChannelClient, t *testing.T) {

	txFilter := &TestTxFilter{err: fmt.Errorf("Error")}
	txOpts := apitxn.ExecuteTxOpts{TxFilter: txFilter}

	_, err := chClient.ExecuteTxWithOpts(apitxn.ExecuteTxRequest{ChaincodeID: ccID, Fcn: "invoke", Args: txArgs}, txOpts)
	if err == nil {
		t.Fatalf("Should have failed with filter error")
	}

}

type TestTxFilter struct {
	err          error
	errResponses error
}

func (tf *TestTxFilter) ProcessTxProposalResponse(txProposalResponse []*apitxn.TransactionProposalResponse) ([]*apitxn.TransactionProposalResponse, error) {
	if tf.err != nil {
		return nil, tf.err
	}

	var newResponses []*apitxn.TransactionProposalResponse

	if tf.errResponses != nil {
		// 404 will cause transaction commit error
		txProposalResponse[0].ProposalResponse.Response.Status = 404
	}

	newResponses = append(newResponses, txProposalResponse[0])
	return newResponses, nil
}

func testChaincodeEvent(ccID string, chClient apitxn.ChannelClient, t *testing.T) {

	eventID := "test([a-zA-Z]+)"

	// Register chaincode event (pass in channel which receives event details when the event is complete)
	notifier := make(chan *apitxn.CCEvent)
	rce := chClient.RegisterChaincodeEvent(notifier, ccID, eventID)

	// Synchronous transaction
	txID, err := chClient.ExecuteTx(apitxn.ExecuteTxRequest{ChaincodeID: ccID, Fcn: "invoke", Args: txArgs})
	if err != nil {
		t.Fatalf("Failed to move funds: %s", err)
	}

	select {
	case ccEvent := <-notifier:
		fmt.Printf("Received cc event: %s\n", ccEvent)
		if ccEvent.TxID != txID.ID {
			t.Fatalf("CCEvent(%s) and ExecuteTx(%s) transaction IDs don't match", ccEvent.TxID, txID.ID)
		}
	case <-time.After(time.Second * 20):
		t.Fatalf("Did NOT receive CC for eventId(%s)\n", eventID)
	}

	// Unregister chain code event using registration handle
	err = chClient.UnregisterChaincodeEvent(rce)
	if err != nil {
		t.Fatalf("Unregister cc event failed: %s", err)
	}

}
