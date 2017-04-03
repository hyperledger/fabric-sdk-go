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
	"fmt"
	"strconv"
	"testing"

	fabricClient "github.com/hyperledger/fabric-sdk-go/fabric-client"
	"github.com/hyperledger/fabric-sdk-go/fabric-client/events"
)

func TestChainQueries(t *testing.T) {

	testSetup := BaseSetupImpl{}
	testSetup.InitConfig()

	eventHub, err := testSetup.GetEventHub(nil)
	if err != nil {
		t.Fatalf("GetEventHub return error: %v", err)
	}
	chain, err := testSetup.GetChain()
	if err != nil {
		t.Fatalf("GetChain return error: %v", err)
	}
	err = testSetup.InstallCC(chain, chainCodeID, chainCodePath, chainCodeVersion, nil, nil)
	if err != nil {
		t.Fatalf("installCC return error: %v", err)
	}
	err = testSetup.InstantiateCC(chain, eventHub)
	if err != nil {
		t.Fatalf("instantiateCC return error: %v", err)
	}

	// Test Query Info - retrieve values before transaction
	bciBeforeTx, err := chain.QueryInfo()
	if err != nil {
		t.Fatalf("QueryInfo return error: %v", err)
	}

	// Invoke transaction that changes block state
	txID, err := changeBlockState(testSetup, chain, eventHub)
	if err != nil {
		t.Fatalf("Failed to change block state (invoke transaction). Return error: %v", err)
	}

	// Test Query Info - retrieve values after transaction
	bciAfterTx, err := chain.QueryInfo()
	if err != nil {
		t.Fatalf("QueryInfo return error: %v", err)
	}

	// Test Query Info -- verify block size changed after transaction
	if (bciAfterTx.Height - bciBeforeTx.Height) <= 0 {
		t.Fatalf("Block size did not increase after transaction")
	}

	testQueryTransaction(t, chain, txID)

	testQueryBlock(t, chain)

	testQueryChannels(t, chain)

	testInstalledChaincodes(t, chain)

	testQueryByChaincode(t, chain)

	// TODO: Synch with test in node SDK when it becomes available
	// testInstantiatedChaincodes(t, chain)

}

func changeBlockState(testSetup BaseSetupImpl, chain fabricClient.Chain, eventHub events.EventHub) (string, error) {

	value, err := testSetup.GetQueryValue(chain)
	if err != nil {
		return "", fmt.Errorf("getQueryValue return error: %v", err)
	}

	// Start transaction that will change block state
	txID, err := testSetup.Invoke(chain, eventHub)
	if err != nil {
		return "", fmt.Errorf("invoke return error: %v", err)
	}

	valueAfterInvoke, err := testSetup.GetQueryValue(chain)
	if err != nil {
		return "", fmt.Errorf("getQueryValue return error: %v", err)
	}

	// Verify that transaction changed block state
	valueInt, _ := strconv.Atoi(value)
	valueInt = valueInt + 1
	valueAfterInvokeInt, _ := strconv.Atoi(valueAfterInvoke)
	if valueInt != valueAfterInvokeInt {
		return "", fmt.Errorf("SendTransaction didn't change the QueryValue %s", value)
	}

	return txID, nil
}

func testQueryTransaction(t *testing.T, chain fabricClient.Chain, txID string) {

	// Test Query Transaction -- verify that valid transaction has been processed
	processedTransaction, err := chain.QueryTransaction(txID)
	if err != nil {
		t.Fatalf("QueryTransaction return error: %v", err)
	}

	if processedTransaction.TransactionEnvelope == nil {
		t.Fatalf("QueryTransaction failed to return transaction envelope")
	}

	// Test Query Transaction -- Retrieve non existing transaction
	processedTransaction, err = chain.QueryTransaction("123ABC")
	if err == nil {
		t.Fatalf("QueryTransaction non-existing didn't return an error")
	}
}

func testQueryBlock(t *testing.T, chain fabricClient.Chain) {

	// Retrieve current blockchain info
	bci, err := chain.QueryInfo()
	if err != nil {
		t.Fatalf("QueryInfo return error: %v", err)
	}

	// Test Query Block by Hash - retrieve current block by hash
	block, err := chain.QueryBlockByHash(bci.CurrentBlockHash)
	if err != nil {
		t.Fatalf("QueryBlockByHash return error: %v", err)
	}

	if block.Data == nil {
		t.Fatalf("QueryBlockByHash block data is nil")
	}

	// Test Query Block by Hash - retrieve block by non-existent hash
	block, err = chain.QueryBlockByHash([]byte("non-existent"))
	if err == nil {
		t.Fatalf("QueryBlockByHash non-existent didn't return an error")
	}

	// Test Query Block - retrieve block by number
	block, err = chain.QueryBlock(1)
	if err != nil {
		t.Fatalf("QueryBlock return error: %v", err)
	}

	if block.Data == nil {
		t.Fatalf("QueryBlock block data is nil")
	}

	// Test Query Block - retrieve block by non-existent number
	block, err = chain.QueryBlock(2147483647)
	if err == nil {
		t.Fatalf("QueryBlock non-existent didn't return an error")
	}

}

func testQueryChannels(t *testing.T, chain fabricClient.Chain) {

	// Our target will be primary peer on this channel
	target := chain.GetPrimaryPeer()

	fmt.Printf("****QueryChannels for %s\n", target.GetURL())

	// Test Query Channels for target
	channelQueryResponse, err := chain.QueryChannels(target)
	if err != nil {
		t.Fatalf("QueryChannels return error: %v", err)
	}

	for _, channel := range channelQueryResponse.Channels {
		fmt.Printf("**Channel: %s\n", channel)
	}

}

func testInstalledChaincodes(t *testing.T, chain fabricClient.Chain) {

	// Our target will be primary peer on this channel
	target := chain.GetPrimaryPeer()

	fmt.Printf("****QueryInstalledChaincodes for %s\n", target.GetURL())

	// Test Query Installed chaincodes for target (primary)
	chaincodeQueryResponse, err := chain.QueryInstalledChaincodes(target)
	if err != nil {
		t.Fatalf("QueryInstalledChaincodes return error: %v", err)
	}

	for _, chaincode := range chaincodeQueryResponse.Chaincodes {
		fmt.Printf("**InstalledCC: %s\n", chaincode)
	}

}

func testInstantiatedChaincodes(t *testing.T, chain fabricClient.Chain) {

	// Our target will indirectly be primary peer on this channel
	target := chain.GetPrimaryPeer()

	fmt.Printf("QueryInstantiatedChaincodes for primary %s\n", target.GetURL())

	// Test Query Instantiated chaincodes
	chaincodeQueryResponse, err := chain.QueryInstantiatedChaincodes()
	if err != nil {
		t.Fatalf("QueryInstantiatedChaincodes return error: %v", err)
	}

	for _, chaincode := range chaincodeQueryResponse.Chaincodes {
		fmt.Printf("**InstantiatedCC: %s\n", chaincode)
	}

}

func testQueryByChaincode(t *testing.T, chain fabricClient.Chain) {

	// Test valid targets
	targets := chain.GetPeers()

	queryResponses, err := chain.QueryByChaincode("lccc", []string{"getinstalledchaincodes"}, targets)
	if err != nil {
		t.Fatalf("QueryByChaincode failed %s", err)
	}

	// Number of responses should be the same as number of targets
	if len(queryResponses) != len(targets) {
		t.Fatalf("QueryByChaincode number of results mismatch. Expected: %d Got: %d", len(targets), len(queryResponses))
	}

	// Create invalid target
	firstInvalidTarget, err := fabricClient.CreateNewPeer("test:1111", "", "")
	if err != nil {
		t.Fatalf("CreateNewPeer error(%v)", err)
	}

	// Create second invalid target
	secondInvalidTarget, err := fabricClient.CreateNewPeer("test:2222", "", "")
	if err != nil {
		t.Fatalf("CreateNewPeer error(%v)", err)
	}

	// Add invalid targets to targets
	invalidTargets := append(targets, firstInvalidTarget)
	invalidTargets = append(invalidTargets, secondInvalidTarget)

	// Add invalid targets to chain otherwise validation will fail
	chain.AddPeer(firstInvalidTarget)
	chain.AddPeer(secondInvalidTarget)

	// Test valid + invalid targets
	queryResponses, err = chain.QueryByChaincode("lccc", []string{"getinstalledchaincodes"}, invalidTargets)
	if err == nil {
		t.Fatalf("QueryByChaincode failed to return error for non-existing target")
	}

	fmt.Printf("QueryByChaincode invalid target error: %s\n", err)

	// Verify that valid targets returned response
	if len(queryResponses) != len(targets) {
		t.Fatalf("QueryByChaincode number of results mismatch. Expected: %d Got: %d", len(targets), len(queryResponses))
	}

	chain.RemovePeer(firstInvalidTarget)
	chain.RemovePeer(secondInvalidTarget)

}
