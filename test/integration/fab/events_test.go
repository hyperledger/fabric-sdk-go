/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package fab

import (
	"sync"
	"testing"
	"time"

	"github.com/hyperledger/fabric-sdk-go/pkg/context/api/fab"
	"github.com/hyperledger/fabric-sdk-go/pkg/fab/resource"
	"github.com/hyperledger/fabric-sdk-go/pkg/fabsdk"
	"github.com/pkg/errors"

	"github.com/hyperledger/fabric-sdk-go/test/integration"

	"github.com/hyperledger/fabric-sdk-go/third_party/github.com/hyperledger/fabric/protos/common"
	pb "github.com/hyperledger/fabric-sdk-go/third_party/github.com/hyperledger/fabric/protos/peer"
)

const (
	eventTimeout = time.Second * 30
)

// Arguments for events CC
var eventCCArgs = [][]byte{[]byte("invoke"), []byte("SEVERE")}

func TestEvents(t *testing.T) {
	chainCodeID := integration.GenerateRandomID()
	testSetup, sdk := initializeTests(t, chainCodeID)
	defer sdk.Close()

	transactor, err := getTransactor(sdk, testSetup.ChannelID, "Admin", testSetup.OrgID)
	if err != nil {
		t.Fatalf("Failed to get channel transactor: %s", err)
	}

	eventHub, err := getEventHub(sdk, testSetup.ChannelID, "Admin", testSetup.OrgID)
	if err != nil {
		t.Fatalf("Failed to get channel event hub: %s", err)
	}

	testReconnectEventHub(t, eventHub)
	testFailedTx(t, transactor, eventHub, testSetup, chainCodeID)
	testFailedTxErrorCode(t, transactor, eventHub, testSetup, chainCodeID)
	testMultipleBlockEventCallbacks(t, transactor, eventHub, testSetup, chainCodeID)
}

func testFailedTx(t *testing.T, transactor fab.Transactor, eventHub fab.EventHub, testSetup integration.BaseSetupImpl, chainCodeID string) {
	fcn := "invoke"

	// Arguments for events CC
	var args [][]byte
	args = append(args, []byte("invoke"))
	args = append(args, []byte("SEVERE"))

	tpResponses1, prop1, err := createAndSendTransactionProposal(transactor, chainCodeID, fcn, args, testSetup.Targets[:1], nil)
	if err != nil {
		t.Fatalf("CreateAndSendTransactionProposal return error: %v", err)
	}

	tpResponses2, prop2, err := createAndSendTransactionProposal(transactor, chainCodeID, fcn, args, testSetup.Targets[:1], nil)
	if err != nil {
		t.Fatalf("CreateAndSendTransactionProposal return error: %v", err)
	}

	// Register tx1 and tx2 for commit/block event(s)
	done1, fail1 := registerTxEvent(t, prop1.TxnID, eventHub)
	defer eventHub.UnregisterTxEvent(prop1.TxnID)

	done2, fail2 := registerTxEvent(t, prop2.TxnID, eventHub)
	defer eventHub.UnregisterTxEvent(prop2.TxnID)

	// Setup monitoring of events
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		monitorFailedTx(t, testSetup, done1, fail1, done2, fail2)
	}()

	// Test invalid transaction: create 2 invoke requests in quick succession that modify
	// the same state variable which should cause one invoke to be invalid
	_, err = createAndSendTransaction(transactor, prop1, tpResponses1)
	if err != nil {
		t.Fatalf("First invoke failed err: %v", err)
	}
	_, err = createAndSendTransaction(transactor, prop2, tpResponses2)
	if err != nil {
		t.Fatalf("Second invoke failed err: %v", err)
	}

	wg.Wait()
}

func monitorFailedTx(t *testing.T, testSetup integration.BaseSetupImpl, done1 chan bool, fail1 chan error, done2 chan bool, fail2 chan error) {
	rcvDone := false
	rcvFail := false
	timeout := time.After(eventTimeout)

Loop:
	for !rcvDone || !rcvFail {
		select {
		case <-done1:
			rcvDone = true
		case <-fail1:
			t.Fatalf("Received fail for first invoke")
		case <-done2:
			t.Fatalf("Received success for second invoke")
		case <-fail2:
			rcvFail = true
		case <-timeout:
			t.Logf("Timeout: Didn't receive events")
			break Loop
		}
	}

	if !rcvDone || !rcvFail {
		t.Fatalf("Didn't receive events (done: %t; fail %t)", rcvDone, rcvFail)
	}
}

func testFailedTxErrorCode(t *testing.T, transactor fab.Transactor, eventHub fab.EventHub, testSetup integration.BaseSetupImpl, chainCodeID string) {
	fcn := "invoke"

	tpResponses1, prop1, err := createAndSendTransactionProposal(transactor, chainCodeID, fcn, eventCCArgs, testSetup.Targets[:1], nil)

	if err != nil {
		t.Fatalf("CreateAndSendTransactionProposal return error: %v", err)
	}

	tpResponses2, prop2, err := createAndSendTransactionProposal(transactor, chainCodeID, fcn, eventCCArgs, testSetup.Targets[:1], nil)
	if err != nil {
		t.Fatalf("CreateAndSendTransactionProposal return error: %v", err)
	}

	done := make(chan bool)
	fail := make(chan pb.TxValidationCode)

	eventHub.RegisterTxEvent(prop1.TxnID, func(txId fab.TransactionID, errorCode pb.TxValidationCode, err error) {
		if err != nil {
			fail <- errorCode
		} else {
			done <- true
		}
	})

	defer eventHub.UnregisterTxEvent(prop1.TxnID)

	done2 := make(chan bool)
	fail2 := make(chan pb.TxValidationCode)

	eventHub.RegisterTxEvent(prop2.TxnID, func(txId fab.TransactionID, errorCode pb.TxValidationCode, err error) {
		if err != nil {
			fail2 <- errorCode
		} else {
			done2 <- true
		}
	})

	defer eventHub.UnregisterTxEvent(prop2.TxnID)

	// Setup monitoring of events
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		monitorFailedTxErrorCode(t, testSetup, done, fail, done2, fail2)
	}()

	// Test invalid transaction: create 2 invoke requests in quick succession that modify
	// the same state variable which should cause one invoke to be invalid
	_, err = createAndSendTransaction(transactor, prop1, tpResponses1)
	if err != nil {
		t.Fatalf("First invoke failed err: %v", err)
	}
	_, err = createAndSendTransaction(transactor, prop2, tpResponses2)
	if err != nil {
		t.Fatalf("Second invoke failed err: %v", err)
	}

	wg.Wait()
}

func monitorFailedTxErrorCode(t *testing.T, testSetup integration.BaseSetupImpl, done chan bool, fail chan pb.TxValidationCode, done2 chan bool, fail2 chan pb.TxValidationCode) {
	rcvDone := false
	rcvFail := false
	timeout := time.After(eventTimeout)

Loop:
	for !rcvDone || !rcvFail {
		select {
		case <-done:
			rcvDone = true
		case <-fail:
			t.Fatalf("Received fail for first invoke")
		case <-done2:
			t.Fatalf("Received success for second invoke")
		case errorValidationCode := <-fail2:
			if errorValidationCode.String() != "MVCC_READ_CONFLICT" {
				t.Fatalf("Expected error code MVCC_READ_CONFLICT. Got %s", errorValidationCode.String())
			}
			rcvFail = true
		case <-timeout:
			t.Logf("Timeout: Didn't receive events")
			break Loop
		}
	}

	if !rcvDone || !rcvFail {
		t.Fatalf("Didn't receive events (done: %t; fail %t)", rcvDone, rcvFail)
	}
}

func testReconnectEventHub(t *testing.T, eventHub fab.EventHub) {
	// Test disconnect event hub
	err := eventHub.Disconnect()
	if err != nil {
		t.Fatalf("Error disconnecting event hub: %s", err)
	}
	if eventHub.IsConnected() {
		t.Fatalf("Failed to disconnect event hub")
	}
	// Reconnect event hub
	if err := eventHub.Connect(); err != nil {
		t.Fatalf("Failed to connect event hub")
	}
}

func testMultipleBlockEventCallbacks(t *testing.T, transactor fab.Transactor, eventHub fab.EventHub, testSetup integration.BaseSetupImpl, chainCodeID string) {
	fcn := "invoke"

	// Create and register test callback that will be invoked upon block event
	test := make(chan bool)
	eventHub.RegisterBlockEvent(func(block *common.Block) {
		t.Logf("Received test callback on block event")
		test <- true
	})

	tpResponses, prop, err := createAndSendTransactionProposal(transactor, chainCodeID, fcn, eventCCArgs, testSetup.Targets[:1], nil)
	if err != nil {
		t.Fatalf("CreateAndSendTransactionProposal returned error: %v", err)
	}

	// Register tx for commit/block event(s)
	done, fail := registerTxEvent(t, prop.TxnID, eventHub)
	defer eventHub.UnregisterTxEvent(prop.TxnID)

	// Setup monitoring of events
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		monitorMultipleBlockEventCallbacks(t, testSetup, done, fail, test)
	}()

	_, err = createAndSendTransaction(transactor, prop, tpResponses)
	if err != nil {
		t.Fatalf("CreateAndSendTransaction failed with error: %v", err)
	}

	wg.Wait()
}

func monitorMultipleBlockEventCallbacks(t *testing.T, testSetup integration.BaseSetupImpl, done chan bool, fail chan error, test chan bool) {
	rcvTxDone := false
	rcvTxEvent := false
	timeout := time.After(eventTimeout)

Loop:
	for !rcvTxDone || !rcvTxEvent {
		select {
		case <-done:
			rcvTxDone = true
		case <-fail:
			t.Fatalf("Received tx failure")
		case <-test:
			rcvTxEvent = true
		case <-timeout:
			t.Logf("Timeout while waiting for events")
			break Loop
		}
	}

	if !rcvTxDone || !rcvTxEvent {
		t.Fatalf("Didn't receive events (tx event: %t; tx done %t)", rcvTxEvent, rcvTxDone)
	}
}

// createAndSendTransaction uses transactor to create and send transaction
func createAndSendTransaction(transactor fab.Sender, proposal *fab.TransactionProposal, resps []*fab.TransactionProposalResponse) (*fab.TransactionResponse, error) {

	txRequest := fab.TransactionRequest{
		Proposal:          proposal,
		ProposalResponses: resps,
	}
	tx, err := transactor.CreateTransaction(txRequest)
	if err != nil {
		return nil, errors.WithMessage(err, "CreateTransaction failed")
	}

	transactionResponse, err := transactor.SendTransaction(tx)
	if err != nil {
		return nil, errors.WithMessage(err, "SendTransaction failed")

	}

	return transactionResponse, nil
}

// registerTxEvent registers on the given eventhub for the give transaction
// returns a boolean channel which receives true when the event is complete
// and an error channel for errors
func registerTxEvent(t *testing.T, txID fab.TransactionID, eventHub fab.EventHub) (chan bool, chan error) {
	done := make(chan bool)
	fail := make(chan error)

	eventHub.RegisterTxEvent(txID, func(txId fab.TransactionID, errorCode pb.TxValidationCode, err error) {
		if err != nil {
			t.Logf("Received error event for txid(%s)", txId)
			fail <- err
		} else {
			t.Logf("Received success event for txid(%s)", txId)
			done <- true
		}
	})

	return done, fail
}

func getTransactor(sdk *fabsdk.FabricSDK, channelID string, user string, orgName string) (fab.Transactor, error) {

	clientChannelContextProvider := sdk.ChannelContext(channelID, fabsdk.WithUser(user), fabsdk.WithOrg(orgName))

	channelContext, err := clientChannelContextProvider()
	if err != nil {
		return nil, errors.WithMessage(err, "channel service creation failed")
	}
	chService := channelContext.ChannelService()

	return chService.Transactor()
}

func getEventHub(sdk *fabsdk.FabricSDK, channelID string, user string, orgName string) (fab.EventHub, error) {

	clientChannelContextProvider := sdk.ChannelContext(channelID, fabsdk.WithUser(user), fabsdk.WithOrg(orgName))

	channelContext, err := clientChannelContextProvider()
	if err != nil {
		return nil, errors.WithMessage(err, "channel service creation failed")
	}
	chService := channelContext.ChannelService()

	eventHub, err := chService.EventHub()
	if err != nil {
		return nil, errors.WithMessage(err, "eventhub client creation failed")
	}

	if err := eventHub.Connect(); err != nil {
		return nil, errors.WithMessage(err, "eventHub connect failed")
	}

	return eventHub, nil
}

func getResource(sdk *fabsdk.FabricSDK, user string, orgName string) (*resource.Resource, error) {

	ctx := sdk.Context(fabsdk.WithUser(user), fabsdk.WithOrg(orgName))

	clientContext, err := ctx()
	if err != nil {
		return nil, errors.WithMessage(err, "create context failed")
	}

	return resource.New(clientContext), nil

}
