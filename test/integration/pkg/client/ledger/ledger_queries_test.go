/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package ledger

import (
	"fmt"
	"strings"
	"testing"

	"github.com/golang/protobuf/proto"
	"github.com/hyperledger/fabric-sdk-go/pkg/client/ledger"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/errors/retry"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/errors/status"
	providersFab "github.com/hyperledger/fabric-sdk-go/pkg/common/providers/fab"
	"github.com/hyperledger/fabric-sdk-go/pkg/core/config"
	"github.com/hyperledger/fabric-sdk-go/pkg/fab"
	"github.com/hyperledger/fabric-sdk-go/pkg/fabsdk"
	"github.com/hyperledger/fabric-sdk-go/test/integration"
	"github.com/stretchr/testify/require"
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
		t.Fatalf("QueryInfo return error: %s", err)
	}

	testPeerConfig(ledgerInfo, t)

	// Same query with target
	target := testSetup.Targets[0]
	//ledgerInfoFromTarget, err := client.QueryInfo(ledger.WithTargetEndpoints(target))
	ledgerInfoFromTarget, err := retry.NewInvoker(retry.New(retry.TestRetryOpts)).Invoke(
		func() (interface{}, error) {
			response, e := client.QueryInfo(ledger.WithTargetEndpoints(target))

			if err != nil && (strings.Contains(e.Error(), "QueryInfo failed") || strings.Contains(e.Error(), "Number of responses")) {
				return nil, status.New(status.TestStatus, status.GenericTransient.ToInt32(), "query funds failed", nil)
			}
			if !proto.Equal(response.BCI, ledgerInfo.BCI) {
				return nil, status.New(status.TestStatus, status.GenericTransient.ToInt32(), "results mismatch between default and target peers", nil)
			}
			return response, nil
		},
	)
	if err != nil {
		t.Fatalf("QueryInfo return error: %s", err)
	}

	if !proto.Equal(ledgerInfoFromTarget.(*providersFab.BlockchainInfoResponse).BCI, ledgerInfo.BCI) {
		t.Fatal("Expecting same result from default peer and target peer")
	}

	testQueryBlockNumberByHash(client, ledgerInfo, t)

	testQueryBlockNumber(client, t, 0)

}
func testPeerConfig(ledgerInfo *providersFab.BlockchainInfoResponse, t *testing.T) {
	sdk := mainSDK
	configBackend, err := sdk.Config()
	if err != nil {
		t.Fatalf("failed to get config backend, error: %s", err)
	}

	endpointConfig, err := fab.ConfigFromBackend(configBackend)
	if err != nil {
		t.Fatalf("failed to get endpoint config, error: %s", err)
	}

	expectedPeerConfig1, ok := endpointConfig.PeerConfig("peer0.org1.example.com")
	require.Truef(t, ok, "Unable to fetch Peer config for %s", "peer0.org1.example.com")
	expectedPeerConfig2, ok := endpointConfig.PeerConfig("peer1.org1.example.com")
	require.Truef(t, ok, "Unable to fetch Peer config for %s", "peer1.org1.example.com")

	if !strings.Contains(ledgerInfo.Endorser, expectedPeerConfig1.URL) && !strings.Contains(ledgerInfo.Endorser, expectedPeerConfig2.URL) {
		t.Fatalf("Expecting %s or %s, got %s", expectedPeerConfig1.URL, expectedPeerConfig2.URL, ledgerInfo.Endorser)
	}
}

func testQueryBlockNumberByHash(client *ledger.Client, ledgerInfo *providersFab.BlockchainInfoResponse, t *testing.T) {
	// Test Query Block by Hash - retrieve current block by hash
	block, err := client.QueryBlockByHash(ledgerInfo.BCI.CurrentBlockHash)
	if err != nil {
		t.Fatalf("QueryBlockByHash return error: %s", err)
	}
	if block.Data == nil {
		t.Fatal("QueryBlockByHash block data is nil")
	}
	// Test Query Block by Hash - retrieve block by non-existent hash
	_, err = client.QueryBlockByHash([]byte("non-existent"))
	if err == nil {
		t.Fatal("QueryBlockByHash non-existent didn't return an error")
	}
}

func testQueryBlockNumber(client *ledger.Client, t *testing.T, blockNumber uint64) {
	// Test Query Block by Number
	block, err := client.QueryBlock(blockNumber)
	if err != nil {
		t.Fatalf("QueryBlockByHash return error: %s", err)
	}
	if block.Data == nil {
		t.Fatal("QueryBlockByHash block data is nil")
	}
	// Test Query Block by Number (non-existent)
	_, err = client.QueryBlock(12345678)
	if err == nil {
		t.Fatal("QueryBlock should have failed to query non-existent block")
	}
}

func TestNoLedgerEndpoints(t *testing.T) {

	// Using shared SDK instance to increase test speed.
	testSetup := mainTestSetup

	configProvider := config.FromFile(integration.GetConfigPath("config_test_endpoints.yaml"))
	//Add entity matchers if local test
	if integration.IsLocal() {
		configProvider = integration.AddLocalEntityMapping(configProvider)
	}

	sdk, err := fabsdk.New(configProvider)
	if err != nil {
		panic(fmt.Sprintf("Failed to create new SDK: %s", err))
	}
	defer sdk.Close()

	//prepare contexts
	org1AdminChannelContext := sdk.ChannelContext(testSetup.ChannelID, fabsdk.WithUser(org1AdminUser), fabsdk.WithOrg(org1Name))

	// Ledger client
	client, err := ledger.New(org1AdminChannelContext)
	if err != nil {
		t.Fatalf("Failed to create new resource management client: %s", err)
	}

	_, err = client.QueryInfo()
	if err == nil {
		t.Fatal("Should have failed due to no ledger query peers")
	}

}
