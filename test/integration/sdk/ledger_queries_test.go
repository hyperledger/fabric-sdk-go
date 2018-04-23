/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package sdk

import (
	"fmt"
	"strings"
	"testing"

	"github.com/golang/protobuf/proto"
	"github.com/hyperledger/fabric-sdk-go/pkg/client/ledger"
	providersFab "github.com/hyperledger/fabric-sdk-go/pkg/common/providers/fab"
	"github.com/hyperledger/fabric-sdk-go/pkg/core/config"
	"github.com/hyperledger/fabric-sdk-go/pkg/fab"
	"github.com/hyperledger/fabric-sdk-go/pkg/fabsdk"
)

func TestLedgerClientQueries(t *testing.T) {

	// Using shared SDK instance to increase test speed.
	sdk := mainSDK
	testSetup := mainTestSetup

	//prepare contexts
	org1AdminChannelContext := sdk.ChannelContext(testSetup.ChannelID, fabsdk.WithUser(org1AdminUser), fabsdk.WithOrg(org1Name))

	// Ledger client
	client, err := ledger.New(org1AdminChannelContext)
	if err != nil {
		t.Fatalf("Failed to create new resource management client: %s", err)
	}

	ledgerInfo, err := client.QueryInfo()
	if err != nil {
		t.Fatalf("QueryInfo return error: %v", err)
	}

	configBackend, err := sdk.Config()
	if err != nil {
		t.Fatalf("failed to get config backend, error: %v", err)
	}

	endpointConfig, err := fab.ConfigFromBackend(configBackend)
	if err != nil {
		t.Fatalf("failed to get endpoint config, error: %v", err)
	}

	expectedPeerConfig, err := endpointConfig.PeerConfig("peer0.org1.example.com")
	if err != nil {
		t.Fatalf("Unable to fetch Peer config for %s", "peer0.org1.example.com")
	}

	if !strings.Contains(ledgerInfo.Endorser, expectedPeerConfig.URL) {
		t.Fatalf("Expecting %s, got %s", expectedPeerConfig.URL, ledgerInfo.Endorser)
	}

	// Same query with target
	target := testSetup.Targets[0]
	ledgerInfoFromTarget, err := client.QueryInfo(ledger.WithTargetURLs(target))
	if err != nil {
		t.Fatalf("QueryInfo return error: %v", err)
	}

	if !proto.Equal(ledgerInfoFromTarget.BCI, ledgerInfo.BCI) {
		t.Fatalf("Expecting same result from default peer and target peer")
	}

	testQueryBlockNumberByHash(client, ledgerInfo, t)

	testQueryBlockNumber(client, t, 0)

}

func testQueryBlockNumberByHash(client *ledger.Client, ledgerInfo *providersFab.BlockchainInfoResponse, t *testing.T) {
	// Test Query Block by Hash - retrieve current block by hash
	block, err := client.QueryBlockByHash(ledgerInfo.BCI.CurrentBlockHash)
	if err != nil {
		t.Fatalf("QueryBlockByHash return error: %v", err)
	}
	if block.Data == nil {
		t.Fatalf("QueryBlockByHash block data is nil")
	}
	// Test Query Block by Hash - retrieve block by non-existent hash
	_, err = client.QueryBlockByHash([]byte("non-existent"))
	if err == nil {
		t.Fatalf("QueryBlockByHash non-existent didn't return an error")
	}
}

func testQueryBlockNumber(client *ledger.Client, t *testing.T, blockNumber uint64) {
	// Test Query Block by Number
	block, err := client.QueryBlock(blockNumber)
	if err != nil {
		t.Fatalf("QueryBlockByHash return error: %v", err)
	}
	if block.Data == nil {
		t.Fatalf("QueryBlockByHash block data is nil")
	}
	// Test Query Block by Number (non-existent)
	_, err = client.QueryBlock(12345678)
	if err == nil {
		t.Fatalf("QueryBlock should have failed to query non-existent block")
	}
}

func TestNoLedgerEndpoints(t *testing.T) {

	// Using shared SDK instance to increase test speed.
	testSetup := mainTestSetup

	sdk, err := fabsdk.New(config.FromFile("../../fixtures/config/config_test_endpoints.yaml"))
	if err != nil {
		panic(fmt.Sprintf("Failed to create new SDK: %s", err))
	}

	//prepare contexts
	org1AdminChannelContext := sdk.ChannelContext(testSetup.ChannelID, fabsdk.WithUser(org1AdminUser), fabsdk.WithOrg(org1Name))

	// Ledger client
	client, err := ledger.New(org1AdminChannelContext)
	if err != nil {
		t.Fatalf("Failed to create new resource management client: %s", err)
	}

	_, err = client.QueryInfo()
	if err == nil {
		t.Fatalf("Should have failed due to no ledger query peers")
	}

}
