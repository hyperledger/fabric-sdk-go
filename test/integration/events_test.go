/*
Copyright SecureKey Technologies Inc. All Rights Reserved.


Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at


      http://www.apache.org/licenses/LICENSE-2.0


Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package integration

import (
	"testing"
	"time"

	fabricClient "github.com/hyperledger/fabric-sdk-go/fabric-client"
	"github.com/hyperledger/fabric-sdk-go/fabric-client/util"
	"github.com/hyperledger/fabric/protos/common"
	pb "github.com/hyperledger/fabric/protos/peer"
)

func TestEvents(t *testing.T) {
	testSetup := initializeTests(t)

	testFailedTx(t, testSetup)
	testFailedTxErrorCode(t, testSetup)
	testReconnectEventHub(t, testSetup)
	testMultipleBlockEventCallbacks(t, testSetup)
}

func initializeTests(t *testing.T) BaseSetupImpl {
	testSetup := BaseSetupImpl{
		ConfigFile:      "../fixtures/config/config_test.yaml",
		ChainID:         "mychannel",
		ChannelConfig:   "../fixtures/channel/mychannel.tx",
		ConnectEventHub: true,
	}

	if err := testSetup.Initialize(); err != nil {
		t.Fatalf(err.Error())
	}

	testSetup.ChainCodeID = util.GenerateRandomID()

	// Install and Instantiate Events CC
	if err := testSetup.InstallCC(testSetup.ChainCodeID, "github.com/events_cc", "v0", nil); err != nil {
		t.Fatalf("installCC return error: %v", err)
	}

	if err := testSetup.InstantiateCC(testSetup.ChainCodeID, testSetup.ChainID, "github.com/events_cc", "v0", nil); err != nil {
		t.Fatalf("instantiateCC return error: %v", err)
	}

	return testSetup
}

func testFailedTx(t *testing.T, testSetup BaseSetupImpl) {
	// Arguments for events CC
	var args []string
	args = append(args, "invoke")
	args = append(args, "invoke")
	args = append(args, "SEVERE")

	tpResponses1, tx1, err := util.CreateAndSendTransactionProposal(testSetup.Chain, testSetup.ChainCodeID, testSetup.ChainID, args, []fabricClient.Peer{testSetup.Chain.GetPrimaryPeer()}, nil)
	if err != nil {
		t.Fatalf("CreateAndSendTransactionProposal return error: %v \n", err)
	}

	tpResponses2, tx2, err := util.CreateAndSendTransactionProposal(testSetup.Chain, testSetup.ChainCodeID, testSetup.ChainID, args, []fabricClient.Peer{testSetup.Chain.GetPrimaryPeer()}, nil)
	if err != nil {
		t.Fatalf("CreateAndSendTransactionProposal return error: %v \n", err)
	}

	// Register tx1 and tx2 for commit/block event(s)
	done1, fail1 := util.RegisterTxEvent(tx1, testSetup.EventHub)
	defer testSetup.EventHub.UnregisterTxEvent(tx1)

	done2, fail2 := util.RegisterTxEvent(tx2, testSetup.EventHub)
	defer testSetup.EventHub.UnregisterTxEvent(tx2)

	// Test invalid transaction: create 2 invoke requests in quick succession that modify
	// the same state variable which should cause one invoke to be invalid
	_, err = util.CreateAndSendTransaction(testSetup.Chain, tpResponses1)
	if err != nil {
		t.Fatalf("First invoke failed err: %v", err)
	}
	_, err = util.CreateAndSendTransaction(testSetup.Chain, tpResponses2)
	if err != nil {
		t.Fatalf("Second invoke failed err: %v", err)
	}

	for i := 0; i < 2; i++ {
		select {
		case <-done1:
		case <-fail1:
			t.Fatalf("Received fail  for second invoke")
		case <-done2:
			t.Fatalf("Received success for second invoke")
		case <-fail2:
			// success
			return
		case <-time.After(time.Second * 30):
			t.Fatalf("invoke Didn't receive block event for txid1(%s) or txid1(%s)", tx1, tx2)
		}
	}
}

func testFailedTxErrorCode(t *testing.T, testSetup BaseSetupImpl) {
	// Arguments for events CC
	var args []string
	args = append(args, "invoke")
	args = append(args, "invoke")
	args = append(args, "SEVERE")

	tpResponses1, tx1, err := util.CreateAndSendTransactionProposal(testSetup.Chain, testSetup.ChainCodeID, testSetup.ChainID, args, []fabricClient.Peer{testSetup.Chain.GetPrimaryPeer()}, nil)
	if err != nil {
		t.Fatalf("CreateAndSendTransactionProposal return error: %v \n", err)
	}

	tpResponses2, tx2, err := util.CreateAndSendTransactionProposal(testSetup.Chain, testSetup.ChainCodeID, testSetup.ChainID, args, []fabricClient.Peer{testSetup.Chain.GetPrimaryPeer()}, nil)
	if err != nil {
		t.Fatalf("CreateAndSendTransactionProposal return error: %v \n", err)
	}

	done := make(chan bool)
	fail := make(chan error)
	var errorValidationCode pb.TxValidationCode
	testSetup.EventHub.RegisterTxEvent(tx1, func(txId string, errorCode pb.TxValidationCode, err error) {
		if err != nil {
			errorValidationCode = errorCode
			fail <- err
		} else {
			done <- true
		}
	})

	defer testSetup.EventHub.UnregisterTxEvent(tx1)

	done2 := make(chan bool)
	fail2 := make(chan error)

	testSetup.EventHub.RegisterTxEvent(tx2, func(txId string, errorCode pb.TxValidationCode, err error) {
		if err != nil {
			errorValidationCode = errorCode
			fail2 <- err
		} else {
			done2 <- true
		}
	})

	defer testSetup.EventHub.UnregisterTxEvent(tx2)

	// Test invalid transaction: create 2 invoke requests in quick succession that modify
	// the same state variable which should cause one invoke to be invalid
	_, err = util.CreateAndSendTransaction(testSetup.Chain, tpResponses1)
	if err != nil {
		t.Fatalf("First invoke failed err: %v", err)
	}
	_, err = util.CreateAndSendTransaction(testSetup.Chain, tpResponses2)
	if err != nil {
		t.Fatalf("Second invoke failed err: %v", err)
	}

	for i := 0; i < 2; i++ {
		select {
		case <-done:
		case <-fail:
			t.Fatalf("Received fail  for second invoke")
		case <-done2:
			t.Fatalf("Received success for second invoke")
		case <-fail2:
			// success
			t.Logf("Received error validation code %s", errorValidationCode.String())
			if errorValidationCode.String() != "MVCC_READ_CONFLICT" {
				t.Fatalf("Expected error code MVCC_READ_CONFLICT")
			}
			return
		case <-time.After(time.Second * 30):
			t.Fatalf("invoke Didn't receive block event for txid1(%s) or txid1(%s)", tx1, tx2)
		}
	}

}

func testReconnectEventHub(t *testing.T, testSetup BaseSetupImpl) {
	// Test disconnect event hub
	testSetup.EventHub.Disconnect()
	if testSetup.EventHub.IsConnected() {
		t.Fatalf("Failed to disconnect event hub")
	}

	// Reconnect event hub
	if err := testSetup.EventHub.Connect(); err != nil {
		t.Fatalf("Failed to connect event hub")
	}
}

func testMultipleBlockEventCallbacks(t *testing.T, testSetup BaseSetupImpl) {
	// Arguments for events CC
	var args []string
	args = append(args, "invoke")
	args = append(args, "invoke")
	args = append(args, "SEVERE")

	// Create and register test callback that will be invoked upon block event
	test := make(chan bool)
	testSetup.EventHub.RegisterBlockEvent(func(block *common.Block) {
		t.Logf("Received test callback on block event")
		test <- true
	})

	tpResponses, tx, err := util.CreateAndSendTransactionProposal(testSetup.Chain, testSetup.ChainCodeID, testSetup.ChainID, args, []fabricClient.Peer{testSetup.Chain.GetPrimaryPeer()}, nil)
	if err != nil {
		t.Fatalf("CreateAndSendTransactionProposal returned error: %v \n", err)
	}

	// Register tx for commit/block event(s)
	done, fail := util.RegisterTxEvent(tx, testSetup.EventHub)
	defer testSetup.EventHub.UnregisterTxEvent(tx)

	_, err = util.CreateAndSendTransaction(testSetup.Chain, tpResponses)
	if err != nil {
		t.Fatalf("CreateAndSendTransaction failed with error: %v", err)
	}

	for i := 0; i < 2; i++ {
		select {
		case <-done:
		case <-fail:
		case <-test:
		case <-time.After(time.Second * 30):
			t.Fatalf("Didn't receive test callback event")
		}
	}

}
