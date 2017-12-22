/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package fab

import (
	"path"
	"sync"
	"testing"
	"time"

	"github.com/hyperledger/fabric-sdk-go/api/apitxn"
	"github.com/hyperledger/fabric-sdk-go/test/integration"
	"github.com/hyperledger/fabric-sdk-go/test/metadata"

	"github.com/hyperledger/fabric-sdk-go/third_party/github.com/hyperledger/fabric/protos/common"
	pb "github.com/hyperledger/fabric-sdk-go/third_party/github.com/hyperledger/fabric/protos/peer"
)

const (
	eventTimeout = time.Second * 30
)

// Arguments for events CC
var eventCCArgs = [][]byte{[]byte("invoke"), []byte("SEVERE")}

func TestEvents(t *testing.T) {
	testSetup := initializeTests(t)

	testReconnectEventHub(t, testSetup)
	testFailedTx(t, testSetup)
	testFailedTxErrorCode(t, testSetup)
	testMultipleBlockEventCallbacks(t, testSetup)
}

func initializeTests(t *testing.T) integration.BaseSetupImpl {
	testSetup := integration.BaseSetupImpl{
		ConfigFile: "../" + integration.ConfigTestFile,

		ChannelID:       "mychannel",
		OrgID:           org1Name,
		ChannelConfig:   path.Join("../../", metadata.ChannelConfigPath, "mychannel.tx"),
		ConnectEventHub: true,
	}

	if err := testSetup.Initialize(t); err != nil {
		t.Fatalf(err.Error())
	}

	testSetup.ChainCodeID = integration.GenerateRandomID()

	if err := testSetup.InstallAndInstantiateCC(testSetup.ChainCodeID, "github.com/events_cc", "v0", testSetup.GetDeployPath(), nil); err != nil {
		t.Fatalf("InstallAndInstantiateCC return error: %v", err)
	}

	return testSetup
}

func testFailedTx(t *testing.T, testSetup integration.BaseSetupImpl) {
	fcn := "invoke"

	// Arguments for events CC
	var args [][]byte
	args = append(args, []byte("invoke"))
	args = append(args, []byte("SEVERE"))

	tpResponses1, tx1, err := testSetup.CreateAndSendTransactionProposal(testSetup.Channel, testSetup.ChainCodeID, fcn, args, []apitxn.ProposalProcessor{testSetup.Channel.PrimaryPeer()}, nil)
	if err != nil {
		t.Fatalf("CreateAndSendTransactionProposal return error: %v", err)
	}

	tpResponses2, tx2, err := testSetup.CreateAndSendTransactionProposal(testSetup.Channel, testSetup.ChainCodeID, fcn, args, []apitxn.ProposalProcessor{testSetup.Channel.PrimaryPeer()}, nil)
	if err != nil {
		t.Fatalf("CreateAndSendTransactionProposal return error: %v", err)
	}

	// Register tx1 and tx2 for commit/block event(s)
	done1, fail1 := testSetup.RegisterTxEvent(t, tx1, testSetup.EventHub)
	defer testSetup.EventHub.UnregisterTxEvent(tx1)

	done2, fail2 := testSetup.RegisterTxEvent(t, tx2, testSetup.EventHub)
	defer testSetup.EventHub.UnregisterTxEvent(tx2)

	// Setup monitoring of events
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		monitorFailedTx(t, testSetup, done1, fail1, done2, fail2)
	}()

	// Test invalid transaction: create 2 invoke requests in quick succession that modify
	// the same state variable which should cause one invoke to be invalid
	_, err = testSetup.CreateAndSendTransaction(testSetup.Channel, tpResponses1)
	if err != nil {
		t.Fatalf("First invoke failed err: %v", err)
	}
	_, err = testSetup.CreateAndSendTransaction(testSetup.Channel, tpResponses2)
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

func testFailedTxErrorCode(t *testing.T, testSetup integration.BaseSetupImpl) {
	fcn := "invoke"

	tpResponses1, tx1, err := testSetup.CreateAndSendTransactionProposal(testSetup.Channel, testSetup.ChainCodeID, fcn, eventCCArgs, []apitxn.ProposalProcessor{testSetup.Channel.PrimaryPeer()}, nil)

	if err != nil {
		t.Fatalf("CreateAndSendTransactionProposal return error: %v", err)
	}

	tpResponses2, tx2, err := testSetup.CreateAndSendTransactionProposal(testSetup.Channel, testSetup.ChainCodeID, fcn, eventCCArgs, []apitxn.ProposalProcessor{testSetup.Channel.PrimaryPeer()}, nil)
	if err != nil {
		t.Fatalf("CreateAndSendTransactionProposal return error: %v", err)
	}

	done := make(chan bool)
	fail := make(chan pb.TxValidationCode)

	testSetup.EventHub.RegisterTxEvent(tx1, func(txId string, errorCode pb.TxValidationCode, err error) {
		if err != nil {
			fail <- errorCode
		} else {
			done <- true
		}
	})

	defer testSetup.EventHub.UnregisterTxEvent(tx1)

	done2 := make(chan bool)
	fail2 := make(chan pb.TxValidationCode)

	testSetup.EventHub.RegisterTxEvent(tx2, func(txId string, errorCode pb.TxValidationCode, err error) {
		if err != nil {
			fail2 <- errorCode
		} else {
			done2 <- true
		}
	})

	defer testSetup.EventHub.UnregisterTxEvent(tx2)

	// Setup monitoring of events
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		monitorFailedTxErrorCode(t, testSetup, done, fail, done2, fail2)
	}()

	// Test invalid transaction: create 2 invoke requests in quick succession that modify
	// the same state variable which should cause one invoke to be invalid
	_, err = testSetup.CreateAndSendTransaction(testSetup.Channel, tpResponses1)
	if err != nil {
		t.Fatalf("First invoke failed err: %v", err)
	}
	_, err = testSetup.CreateAndSendTransaction(testSetup.Channel, tpResponses2)
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

func testReconnectEventHub(t *testing.T, testSetup integration.BaseSetupImpl) {
	// Test disconnect event hub
	err := testSetup.EventHub.Disconnect()
	if err != nil {
		t.Fatalf("Error disconnecting event hub: %s", err)
	}
	if testSetup.EventHub.IsConnected() {
		t.Fatalf("Failed to disconnect event hub")
	}
	// Reconnect event hub
	if err := testSetup.EventHub.Connect(); err != nil {
		t.Fatalf("Failed to connect event hub")
	}
}

func testMultipleBlockEventCallbacks(t *testing.T, testSetup integration.BaseSetupImpl) {
	fcn := "invoke"

	// Create and register test callback that will be invoked upon block event
	test := make(chan bool)
	testSetup.EventHub.RegisterBlockEvent(func(block *common.Block) {
		t.Logf("Received test callback on block event")
		test <- true
	})

	tpResponses, tx, err := testSetup.CreateAndSendTransactionProposal(testSetup.Channel, testSetup.ChainCodeID, fcn, eventCCArgs, []apitxn.ProposalProcessor{testSetup.Channel.PrimaryPeer()}, nil)
	if err != nil {
		t.Fatalf("CreateAndSendTransactionProposal returned error: %v", err)
	}

	// Register tx for commit/block event(s)
	done, fail := testSetup.RegisterTxEvent(t, tx, testSetup.EventHub)
	defer testSetup.EventHub.UnregisterTxEvent(tx)

	// Setup monitoring of events
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		monitorMultipleBlockEventCallbacks(t, testSetup, done, fail, test)
	}()

	_, err = testSetup.CreateAndSendTransaction(testSetup.Channel, tpResponses)
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
