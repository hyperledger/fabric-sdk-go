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
	"strconv"
	"testing"
)

func TestChainQueries(t *testing.T) {

	InitConfigForEndToEnd()
	testSetup := BaseSetupImpl{}

	eventHub := testSetup.GetEventHub(t, nil)
	queryChain, invokeChain, deployChain := testSetup.GetChains(t)
	testSetup.SetupChaincodeDeploy()
	err := installCC(deployChain)
	if err != nil {
		t.Fatalf("installCC return error: %v", err)
	}
	err = instantiateCC(deployChain, eventHub)
	if err != nil {
		t.Fatalf("instantiateCC return error: %v", err)
	}

	// Test Query Info - retrieve values before transaction
	bciBeforeTx, err := queryChain.QueryInfo()
	if err != nil {
		t.Fatalf("QueryInfo return error: %v", err)
	}

	// Start transaction that will change block state
	value, err := getQueryValue(t, queryChain)
	if err != nil {
		t.Fatalf("getQueryValue return error: %v", err)
	}

	txID, err := invoke(t, invokeChain, eventHub)
	if err != nil {
		t.Fatalf("invoke return error: %v", err)
	}

	valueAfterInvoke, err := getQueryValue(t, queryChain)
	if err != nil {
		t.Fatalf("getQueryValue return error: %v", err)
	}

	// Verify that transaction changed block state
	valueInt, _ := strconv.Atoi(value)
	valueInt = valueInt + 1
	valueAfterInvokeInt, _ := strconv.Atoi(valueAfterInvoke)
	if valueInt != valueAfterInvokeInt {
		t.Fatalf("SendTransaction didn't change the QueryValue %s", value)
	}

	// End transaction that changed block state

	// Test Query Info - retrieve values after transaction
	bciAfterTx, err := queryChain.QueryInfo()
	if err != nil {
		t.Fatalf("QueryInfo return error: %v", err)
	}

	// Test Query Info -- verify block size changed after transaction
	if (bciAfterTx.Height - bciBeforeTx.Height) <= 0 {
		t.Fatalf("Block size did not increase after transaction")
	}

	// Test Query Transaction -- verify that transaction has been processed
	processedTransaction, err := queryChain.QueryTransaction(txID)
	if err != nil {
		t.Fatalf("QueryTransaction return error: %v", err)
	}

	if processedTransaction.TransactionEnvelope == nil {
		t.Fatalf("QueryTransaction failed to return transaction envelope")
	}

	// Test Query Transaction -- Retrieve non existing transaction
	processedTransaction, err = queryChain.QueryTransaction("123ABC")
	if err == nil {
		t.Fatalf("QueryTransaction non-existing didn't return an error")
	}

	// Test Query Block by Hash - retrieve current block by hash
	block, err := queryChain.QueryBlockByHash(bciAfterTx.CurrentBlockHash)
	if err != nil {
		t.Fatalf("QueryBlockByHash return error: %v", err)
	}

	if block.Data == nil {
		t.Fatalf("QueryBlockByHash block data is nil")
	}

	// Test Query Block by Hash - retrieve block by non-existent hash
	block, err = queryChain.QueryBlockByHash([]byte("non-existent"))
	if err == nil {
		t.Fatalf("QueryBlockByHash non-existent didn't return an error")
	}

	// Test Query Block - retrieve block by number
	block, err = queryChain.QueryBlock(1)
	if err != nil {
		t.Fatalf("QueryBlock return error: %v", err)
	}

	if block.Data == nil {
		t.Fatalf("QueryBlock block data is nil")
	}

	// Test Query Block - retrieve block by non-existent number
	block, err = queryChain.QueryBlock(2147483647)
	if err == nil {
		t.Fatalf("QueryBlock return error: %v", err)
	}
}
