/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package fab

import (
	"path"
	"strconv"
	"testing"

	"github.com/hyperledger/fabric-sdk-go/pkg/common/errors/retry"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/fab"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/msp"
	"github.com/hyperledger/fabric-sdk-go/test/integration"
	"github.com/hyperledger/fabric-sdk-go/test/metadata"

	"github.com/hyperledger/fabric-sdk-go/pkg/client/channel"
	"github.com/hyperledger/fabric-sdk-go/pkg/client/ledger"
	"github.com/hyperledger/fabric-sdk-go/pkg/client/resmgmt"
	"github.com/hyperledger/fabric-sdk-go/pkg/fabsdk"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/require"
)

const (
	channelConfigFile = "mychannel.tx"
	channelID         = "mychannel"
	orgName           = org1Name
)

func initializeLedgerTests(t *testing.T) (*fabsdk.FabricSDK, []string) {
	// Using shared SDK instance to increase test speed.
	sdk := mainSDK

	//var sdkConfigFile = "../" + integration.ConfigTestFile
	//	sdk, err := fabsdk.New(config.FromFile(sdkConfigFile))
	//	if err != nil {
	//		t.Fatalf("SDK init failed: %v", err)
	//	}
	// Get signing identity that is used to sign create channel request

	adminIdentity, err := integration.GetSigningIdentity(sdk, "Admin", orgName)
	if err != nil {
		t.Fatalf("failed to load signing identity: %s", err)
	}

	configBackend, err := sdk.Config()
	if err != nil {
		t.Fatalf("failed to get config backend from SDK: %v", err)
	}

	targets, err := integration.OrgTargetPeers(configBackend, []string{orgName})
	if err != nil {
		t.Fatalf("creating peers failed: %v", err)
	}

	req := resmgmt.SaveChannelRequest{ChannelID: channelID, ChannelConfigPath: path.Join("../../../", metadata.ChannelConfigPath, channelConfigFile), SigningIdentities: []msp.SigningIdentity{adminIdentity}}
	err = integration.InitializeChannel(sdk, orgName, req, targets)
	if err != nil {
		t.Fatalf("failed to ensure channel has been initialized: %s", err)
	}
	return sdk, targets
}

func TestLedgerQueries(t *testing.T) {

	// Setup tests with a random chaincode ID.
	sdk, targets := initializeLedgerTests(t)

	// Using shared SDK instance to increase test speed.
	//defer sdk.Close()

	chaincodeID := integration.GenerateRandomID()
	resp, err := integration.InstallAndInstantiateExampleCC(sdk, fabsdk.WithUser("Admin"), orgName, chaincodeID)
	require.Nil(t, err, "InstallAndInstantiateExampleCC return error")
	require.NotEmpty(t, resp, "instantiate response should be populated")

	//prepare required contexts

	channelClientCtx := sdk.ChannelContext(channelID, fabsdk.WithUser("Admin"), fabsdk.WithOrg(orgName))

	// Get a ledger client.
	ledgerClient, err := ledger.New(channelClientCtx)
	require.Nil(t, err, "ledger new return error")

	// Test Query Info - retrieve values before transaction
	testTargets := targets[0:1]
	bciBeforeTx, err := ledgerClient.QueryInfo(ledger.WithTargetURLs(testTargets...))
	if err != nil {
		t.Fatalf("QueryInfo return error: %v", err)
	}

	// Invoke transaction that changes block state
	channelClient, err := channel.New(channelClientCtx)
	if err != nil {
		t.Fatalf("creating channel failed: %v", err)
	}

	txID, err := changeBlockState(t, channelClient, chaincodeID)
	if err != nil {
		t.Fatalf("Failed to change block state (invoke transaction). Return error: %v", err)
	}

	// Test Query Info - retrieve values after transaction
	bciAfterTx, err := ledgerClient.QueryInfo(ledger.WithTargetURLs(testTargets...))
	if err != nil {
		t.Fatalf("QueryInfo return error: %v", err)
	}

	// Test Query Info -- verify block size changed after transaction
	if (bciAfterTx.BCI.Height - bciBeforeTx.BCI.Height) <= 0 {
		t.Fatalf("Block size did not increase after transaction")
	}

	testQueryTransaction(t, ledgerClient, txID, targets)

	testQueryBlock(t, ledgerClient, targets)

	testQueryBlockByTxID(t, ledgerClient, txID, targets)

	//prepare context
	clientCtx := sdk.Context(fabsdk.WithUser("Admin"), fabsdk.WithOrg(orgName))

	resmgmtClient, err := resmgmt.New(clientCtx)

	require.Nil(t, err, "resmgmt new return error")

	testInstantiatedChaincodes(t, chaincodeID, channelID, resmgmtClient, targets)

	testQueryConfigBlock(t, ledgerClient, targets)
}

func changeBlockState(t *testing.T, client *channel.Client, chaincodeID string) (fab.TransactionID, error) {

	req := channel.Request{
		ChaincodeID: chaincodeID,
		Fcn:         "invoke",
		Args:        integration.ExampleCCQueryArgs(),
	}
	resp, err := client.Query(req)
	if err != nil {
		return "", errors.WithMessage(err, "query funds failed")
	}
	value := resp.Payload

	// Start transaction that will change block state
	txID, err := moveFundsAndGetTxID(t, client, chaincodeID)
	if err != nil {
		return "", errors.WithMessage(err, "move funds failed")
	}

	resp, err = client.Query(req)
	if err != nil {
		return "", errors.WithMessage(err, "query funds failed")
	}
	valueAfterInvoke := resp.Payload

	// Verify that transaction changed block state
	valueInt, _ := strconv.Atoi(string(value))
	valueInt = valueInt + 1
	valueAfterInvokeInt, _ := strconv.Atoi(string(valueAfterInvoke))
	if valueInt != valueAfterInvokeInt {
		return "", errors.Errorf("SendTransaction didn't change the QueryValue %s", value)
	}

	return txID, nil
}

func testQueryTransaction(t *testing.T, ledgerClient *ledger.Client, txID fab.TransactionID, targets []string) {

	// Test Query Transaction -- verify that valid transaction has been processed
	processedTransaction, err := ledgerClient.QueryTransaction(txID, ledger.WithTargetURLs(targets...))
	if err != nil {
		t.Fatalf("QueryTransaction return error: %v", err)
	}

	if processedTransaction.TransactionEnvelope == nil {
		t.Fatalf("QueryTransaction failed to return transaction envelope")
	}

	// Test Query Transaction -- Retrieve non existing transaction
	_, err = ledgerClient.QueryTransaction("123ABC", ledger.WithTargetURLs(targets...))
	if err == nil {
		t.Fatalf("QueryTransaction non-existing didn't return an error")
	}
}

func testQueryBlock(t *testing.T, ledgerClient *ledger.Client, targets []string) {

	// Retrieve current blockchain info
	bci, err := ledgerClient.QueryInfo(ledger.WithTargetURLs(targets...))
	if err != nil {
		t.Fatalf("QueryInfo return error: %v", err)
	}

	// Test Query Block by Hash - retrieve current block by hash
	block, err := ledgerClient.QueryBlockByHash(bci.BCI.CurrentBlockHash, ledger.WithTargetURLs(targets...))
	if err != nil {
		t.Fatalf("QueryBlockByHash return error: %v", err)
	}

	if block.Data == nil {
		t.Fatalf("QueryBlockByHash block data is nil")
	}

	// Test Query Block by Hash - retrieve block by non-existent hash
	_, err = ledgerClient.QueryBlockByHash([]byte("non-existent"), ledger.WithTargetURLs(targets...))
	if err == nil {
		t.Fatalf("QueryBlockByHash non-existent didn't return an error")
	}

	// Test Query Block - retrieve block by number
	block, err = ledgerClient.QueryBlock(1, ledger.WithTargetURLs(targets...))
	if err != nil {
		t.Fatalf("QueryBlock return error: %v", err)
	}
	if block.Data == nil {
		t.Fatalf("QueryBlock block data is nil")
	}

	// Test Query Block - retrieve block by non-existent number
	_, err = ledgerClient.QueryBlock(2147483647, ledger.WithTargetURLs(targets...))
	if err == nil {
		t.Fatalf("QueryBlock non-existent didn't return an error")
	}
}

func testQueryBlockByTxID(t *testing.T, ledgerClient *ledger.Client, txID fab.TransactionID, targets []string) {

	// Test Query Block- retrieve block by non-existent tx ID
	_, err := ledgerClient.QueryBlockByTxID("non-existent", ledger.WithTargetURLs(targets...))
	if err == nil {
		t.Fatal("QueryBlockByTxID non-existent didn't return an error")
	}

	// Test Query Block - retrieve block by valid tx ID
	block, err := ledgerClient.QueryBlockByTxID(txID, ledger.WithTargetURLs(targets...))
	if err != nil {
		t.Fatalf("QueryBlockByTxID return error: %v", err)
	}
	if block.Data == nil {
		t.Fatal("QueryBlockByTxID block data is nil")
	}

}

func testInstantiatedChaincodes(t *testing.T, ccID string, channelID string, resmgmtClient *resmgmt.Client, targets []string) {

	found := false

	// Test Query Instantiated chaincodes
	chaincodeQueryResponse, err := resmgmtClient.QueryInstantiatedChaincodes(channelID, resmgmt.WithTargetURLs(targets...), resmgmt.WithRetry(retry.DefaultResMgmtOpts))
	if err != nil {
		t.Fatalf("QueryInstantiatedChaincodes return error: %v", err)
	}

	for _, chaincode := range chaincodeQueryResponse.Chaincodes {
		t.Logf("**InstantiatedCC: %s", chaincode)
		if chaincode.Name == ccID {
			found = true
		}
	}

	if !found {
		t.Fatalf("QueryInstantiatedChaincodes failed to find instantiated %s chaincode", ccID)
	}
}

// MoveFundsAndGetTxID ...
func moveFundsAndGetTxID(t *testing.T, client *channel.Client, chaincodeID string) (fab.TransactionID, error) {

	transientDataMap := make(map[string][]byte)
	transientDataMap["result"] = []byte("Transient data in move funds...")

	req := channel.Request{
		ChaincodeID:  chaincodeID,
		Fcn:          "invoke",
		Args:         integration.ExampleCCTxArgs(),
		TransientMap: transientDataMap,
	}
	resp, err := client.Execute(req)
	if err != nil {
		return "", errors.WithMessage(err, "execute move funds failed")
	}

	return resp.TransactionID, nil
}

func testQueryConfigBlock(t *testing.T, ledgerClient *ledger.Client, targets []string) {

	// Retrieve current channel configuration
	cfgEnvelope, err := ledgerClient.QueryConfig(ledger.WithTargetURLs(targets...))
	if err != nil {
		t.Fatalf("QueryConfig return error: %v", err)
	}

	if cfgEnvelope == nil {
		t.Fatalf("QueryConfig config data is nil")
	}

}
