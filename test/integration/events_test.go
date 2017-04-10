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
	"github.com/hyperledger/fabric-sdk-go/fabric-client/events"
)

func TestEvents(t *testing.T) {

	testSetup := BaseSetupImpl{}

	testSetup.InitConfig()

	eventHub, err := testSetup.GetEventHubAndConnect()
	if err != nil {
		t.Fatalf("GetEventHub return error: %v", err)
	}
	chain, err := testSetup.GetChain()
	if err != nil {
		t.Fatalf("GetChain return error: %v", err)
	}
	// Create and join channel represented by 'chain'
	testSetup.CreateAndJoinChannel(t, chain)

	// Install and Instantiate Events CC
	err = testSetup.InstallCC(chain, chainCodeID, "github.com/events_cc", chainCodeVersion, nil, nil)
	if err != nil {
		t.Fatalf("installCC return error: %v", err)
	}
	err = testSetup.InstantiateCC(chain, eventHub)
	if err != nil {
		t.Fatalf("instantiateCC return error: %v", err)
	}

	testFailedTx(t, chain, eventHub)

}

func testFailedTx(t *testing.T, chain fabricClient.Chain, eventHub events.EventHub) {

	// Arguments for events CC
	var args []string
	args = append(args, "invoke")
	args = append(args, "invoke")
	args = append(args, "SEVERE")

	tpResponses1, tx1, err := CreateAndSendTransactionProposal(chain, chainCodeID, chainID, args, []fabricClient.Peer{chain.GetPrimaryPeer()})
	if err != nil {
		t.Fatalf("CreateAndSendTransactionProposal return error: %v \n", err)
	}

	tpResponses2, tx2, err := CreateAndSendTransactionProposal(chain, chainCodeID, chainID, args, []fabricClient.Peer{chain.GetPrimaryPeer()})
	if err != nil {
		t.Fatalf("CreateAndSendTransactionProposal return error: %v \n", err)
	}

	// Register tx1 and tx2 for commit/block event(s)
	done1, fail1 := RegisterEvent(tx1, eventHub)
	defer eventHub.UnregisterTxEvent(tx1)

	done2, fail2 := RegisterEvent(tx2, eventHub)
	defer eventHub.UnregisterTxEvent(tx2)

	// Test invalid transaction: create 2 invoke requests in quick succession that modify
	// the same state variable which should cause one invoke to be invalid
	_, err = CreateAndSendTransaction(chain, tpResponses1)
	if err != nil {
		t.Fatalf("First invoke failed err: %v", err)
	}
	_, err = CreateAndSendTransaction(chain, tpResponses2)
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
