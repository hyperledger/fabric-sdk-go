/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package integration

import (
	"fmt"
	"strconv"
	"testing"

	fabricClient "github.com/hyperledger/fabric-sdk-go/fabric-client"
)

func TestChainQueries(t *testing.T) {

	testSetup := &BaseSetupImpl{
		ConfigFile:      "../fixtures/config/config_test.yaml",
		ChainID:         "mychannel",
		ChannelConfig:   "../fixtures/channel/mychannel.tx",
		ConnectEventHub: true,
	}

	if err := testSetup.Initialize(); err != nil {
		t.Fatalf(err.Error())
	}

	chain := testSetup.Chain
	client := testSetup.Client

	if err := testSetup.InstallAndInstantiateExampleCC(); err != nil {
		t.Fatalf("InstallAndInstantiateExampleCC return error: %v", err)
	}

	// Test Query Info - retrieve values before transaction
	bciBeforeTx, err := chain.QueryInfo()
	if err != nil {
		t.Fatalf("QueryInfo return error: %v", err)
	}

	// Invoke transaction that changes block state
	txID, err := changeBlockState(testSetup)
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

	testQueryChannels(t, chain, client)

	testInstalledChaincodes(t, chain, client)

	testQueryByChaincode(t, chain)

	// TODO: Synch with test in node SDK when it becomes available
	// testInstantiatedChaincodes(t, chain)

}

func changeBlockState(testSetup *BaseSetupImpl) (string, error) {

	value, err := testSetup.QueryAsset()
	if err != nil {
		return "", fmt.Errorf("getQueryValue return error: %v", err)
	}

	// Start transaction that will change block state
	txID, err := testSetup.MoveFunds()
	if err != nil {
		return "", fmt.Errorf("Move funds return error: %v", err)
	}

	valueAfterInvoke, err := testSetup.QueryAsset()
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

func testQueryChannels(t *testing.T, chain fabricClient.Chain, client fabricClient.Client) {

	// Our target will be primary peer on this channel
	target := chain.GetPrimaryPeer()
	fmt.Printf("****QueryChannels for %s\n", target.GetURL())
	channelQueryResponse, err := client.QueryChannels(target)
	if err != nil {
		t.Fatalf("QueryChannels return error: %v", err)
	}

	for _, channel := range channelQueryResponse.Channels {
		fmt.Printf("**Channel: %s\n", channel)
	}

}

func testInstalledChaincodes(t *testing.T, chain fabricClient.Chain, client fabricClient.Client) {

	// Our target will be primary peer on this channel
	target := chain.GetPrimaryPeer()

	fmt.Printf("****QueryInstalledChaincodes for %s\n", target.GetURL())
	// Test Query Installed chaincodes for target (primary)
	chaincodeQueryResponse, err := client.QueryInstalledChaincodes(target)
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

	queryResponses, err := chain.QueryByChaincode("lscc", []string{"getinstalledchaincodes"}, targets)
	if err != nil {
		t.Fatalf("QueryByChaincode failed %s", err)
	}

	// Number of responses should be the same as number of targets
	if len(queryResponses) != len(targets) {
		t.Fatalf("QueryByChaincode number of results mismatch. Expected: %d Got: %d", len(targets), len(queryResponses))
	}

	// Create invalid target
	firstInvalidTarget, err := fabricClient.NewPeer("test:1111", "", "")
	if err != nil {
		t.Fatalf("Create NewPeer error(%v)", err)
	}

	// Create second invalid target
	secondInvalidTarget, err := fabricClient.NewPeer("test:2222", "", "")
	if err != nil {
		t.Fatalf("Create NewPeer error(%v)", err)
	}

	// Add invalid targets to targets
	invalidTargets := append(targets, firstInvalidTarget)
	invalidTargets = append(invalidTargets, secondInvalidTarget)

	// Add invalid targets to chain otherwise validation will fail
	err = chain.AddPeer(firstInvalidTarget)
	if err != nil {
		t.Fatalf("Error adding peer: %v", err)
	}
	err = chain.AddPeer(secondInvalidTarget)
	if err != nil {
		t.Fatalf("Error adding peer: %v", err)
	}

	// Test valid + invalid targets
	queryResponses, err = chain.QueryByChaincode("lscc", []string{"getinstalledchaincodes"}, invalidTargets)
	if err == nil {
		t.Fatalf("QueryByChaincode failed to return error for non-existing target")
	}

	// Verify that valid targets returned response
	if len(queryResponses) != len(targets) {
		t.Fatalf("QueryByChaincode number of results mismatch. Expected: %d Got: %d", len(targets), len(queryResponses))
	}

	chain.RemovePeer(firstInvalidTarget)
	chain.RemovePeer(secondInvalidTarget)
}
