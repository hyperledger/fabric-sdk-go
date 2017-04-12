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
	fcUtil "github.com/hyperledger/fabric-sdk-go/fabric-client/helpers"
)

func TestEvents(t *testing.T) {
	testSetup := BaseSetupImpl{
		ConfigFile:      "../fixtures/config/config_test.yaml",
		ChainID:         "testchannel",
		ChannelConfig:   "../fixtures/channel/testchannel.tx",
		ConnectEventHub: true,
	}

	if err := testSetup.Initialize(); err != nil {
		t.Fatalf(err.Error())
	}

	testSetup.ChainCodeID = fcUtil.GenerateRandomID()

	// Install and Instantiate Events CC
	if err := testSetup.InstallCC(testSetup.ChainCodeID, "github.com/events_cc", "v0", nil); err != nil {
		t.Fatalf("installCC return error: %v", err)
	}

	if err := testSetup.InstantiateCC(testSetup.ChainCodeID, testSetup.ChainID, "github.com/events_cc", "v0", nil); err != nil {
		t.Fatalf("instantiateCC return error: %v", err)
	}

	testFailedTx(t, testSetup)

}

func testFailedTx(t *testing.T, testSetup BaseSetupImpl) {

	// Arguments for events CC
	var args []string
	args = append(args, "invoke")
	args = append(args, "invoke")
	args = append(args, "SEVERE")

	tpResponses1, tx1, err := fcUtil.CreateAndSendTransactionProposal(testSetup.Chain, testSetup.ChainCodeID, testSetup.ChainID, args, []fabricClient.Peer{testSetup.Chain.GetPrimaryPeer()})
	if err != nil {
		t.Fatalf("CreateAndSendTransactionProposal return error: %v \n", err)
	}

	tpResponses2, tx2, err := fcUtil.CreateAndSendTransactionProposal(testSetup.Chain, testSetup.ChainCodeID, testSetup.ChainID, args, []fabricClient.Peer{testSetup.Chain.GetPrimaryPeer()})
	if err != nil {
		t.Fatalf("CreateAndSendTransactionProposal return error: %v \n", err)
	}

	// Register tx1 and tx2 for commit/block event(s)
	done1, fail1 := fcUtil.RegisterTxEvent(tx1, testSetup.EventHub)
	defer testSetup.EventHub.UnregisterTxEvent(tx1)

	done2, fail2 := fcUtil.RegisterTxEvent(tx2, testSetup.EventHub)
	defer testSetup.EventHub.UnregisterTxEvent(tx2)

	// Test invalid transaction: create 2 invoke requests in quick succession that modify
	// the same state variable which should cause one invoke to be invalid
	_, err = fcUtil.CreateAndSendTransaction(testSetup.Chain, tpResponses1)
	if err != nil {
		t.Fatalf("First invoke failed err: %v", err)
	}
	_, err = fcUtil.CreateAndSendTransaction(testSetup.Chain, tpResponses2)
	if err != nil {
		t.Fatalf("Second invoke failed err: %v", err)
	}

	for i := 0; i < 2; i++ {
		select {
		case <-done1:
		case <-fail1:
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
