/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package sdk

import (
	"path"
	"strings"
	"testing"

	"github.com/golang/protobuf/proto"
	"github.com/hyperledger/fabric-sdk-go/pkg/client/ledger"
	"github.com/hyperledger/fabric-sdk-go/pkg/context/api/fab"
	"github.com/hyperledger/fabric-sdk-go/pkg/core/config"
	"github.com/hyperledger/fabric-sdk-go/pkg/fabsdk"
	"github.com/hyperledger/fabric-sdk-go/test/integration"
	"github.com/hyperledger/fabric-sdk-go/test/metadata"
)

func TestLedgerClientQueries(t *testing.T) {

	testSetup := integration.BaseSetupImpl{
		ConfigFile:    "../" + integration.ConfigTestFile,
		ChannelID:     "mychannel",
		OrgID:         org1Name,
		ChannelConfig: path.Join("../../../", metadata.ChannelConfigPath, "mychannel.tx"),
	}

	if err := testSetup.Initialize(); err != nil {
		t.Fatalf(err.Error())
	}

	// Create SDK setup for the integration tests
	sdk, err := fabsdk.New(config.FromFile(testSetup.ConfigFile))
	if err != nil {
		t.Fatalf("Failed to create new SDK: %s", err)
	}
	defer sdk.Close()

	//prepare contexts
	org1AdminClientContext := sdk.Context(fabsdk.WithUser(org1AdminUser), fabsdk.WithOrg(org1Name))

	// Ledger client
	client, err := ledger.New(org1AdminClientContext, testSetup.ChannelID)
	if err != nil {
		t.Fatalf("Failed to create new resource management client: %s", err)
	}

	ledgerInfo, err := client.QueryInfo()
	if err != nil {
		t.Fatalf("QueryInfo return error: %v", err)
	}

	expectedPeerConfig, err := sdk.Config().PeerConfig(org1Name, "peer0.org1.example.com")
	if err != nil {
		t.Fatalf("Unable to fetch Peer config for %s", "peer0.org1.example.com")
	}

	if !strings.Contains(ledgerInfo.Endorser, expectedPeerConfig.URL) {
		t.Fatalf("Expecting %s, got %s", expectedPeerConfig.URL, ledgerInfo.Endorser)
	}

	// Same query with target
	target := testSetup.Targets[0]
	ledgerInfoFromTarget, err := client.QueryInfo(ledger.WithTargets(target.(fab.Peer)))
	if err != nil {
		t.Fatalf("QueryInfo return error: %v", err)
	}

	if !proto.Equal(ledgerInfoFromTarget.BCI, ledgerInfo.BCI) {
		t.Fatalf("Expecting same result from default peer and target peer")
	}

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

	// Test Query Block by Number
	block, err = client.QueryBlock(0)
	if err != nil {
		t.Fatalf("QueryBlockByHash return error: %v", err)
	}
	if block.Data == nil {
		t.Fatalf("QueryBlockByHash block data is nil")
	}

	// Test Query Block by Number (non-existent)
	block, err = client.QueryBlock(12345678)
	if err == nil {
		t.Fatalf("QueryBlock should have failed to query non-existent block")
	}

}
