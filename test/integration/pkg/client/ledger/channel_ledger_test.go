/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package ledger

import (
	"fmt"
	"strconv"
	"strings"
	"testing"

	"github.com/hyperledger/fabric-sdk-go/pkg/common/errors/retry"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/errors/status"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/fab"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/msp"
	"github.com/hyperledger/fabric-sdk-go/test/integration"
	"github.com/hyperledger/fabric-sdk-go/test/metadata"

	"github.com/hyperledger/fabric-sdk-go/pkg/client/channel"
	"github.com/hyperledger/fabric-sdk-go/pkg/client/ledger"
	mspclient "github.com/hyperledger/fabric-sdk-go/pkg/client/msp"
	"github.com/hyperledger/fabric-sdk-go/pkg/client/resmgmt"
	"github.com/hyperledger/fabric-sdk-go/pkg/fabsdk"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/require"
)

const (
	channelConfigTxFile = "mychannel.tx"
	channelID           = "mychannel"
	orgName             = org1Name
)

func initializeLedgerTests(t *testing.T) (*fabsdk.FabricSDK, []string) {
	// Using shared SDK instance to increase test speed.
	sdk := mainSDK

	// Get signing identity that is used to sign create channel request
	orgMspClient, err := mspclient.New(sdk.Context(), mspclient.WithOrg(orgName))
	if err != nil {
		t.Fatalf("failed to create org2MspClient, err : %s", err)
	}

	adminIdentity, err := orgMspClient.GetSigningIdentity("Admin")
	if err != nil {
		t.Fatalf("failed to load signing identity: %s", err)
	}

	configBackend, err := sdk.Config()
	if err != nil {
		t.Fatalf("failed to get config backend from SDK: %s", err)
	}

	targets, err := integration.OrgTargetPeers([]string{orgName}, configBackend)
	if err != nil {
		t.Fatalf("creating peers failed: %s", err)
	}

	req := resmgmt.SaveChannelRequest{ChannelID: channelID, ChannelConfigPath: integration.GetChannelConfigTxPath(channelConfigTxFile), SigningIdentities: []msp.SigningIdentity{adminIdentity}}
	err = integration.InitializeChannel(sdk, orgName, req, targets)
	if err != nil {
		t.Fatalf("failed to ensure channel has been initialized: %s", err)
	}
	return sdk, targets
}

func TestLedgerQueries(t *testing.T) {
	testSetup := mainTestSetup

	aKey := integration.GetKeyName(t)
	bKey := integration.GetKeyName(t)
	moveTxArg := integration.ExampleCCTxArgs(aKey, bKey, "1")
	queryArg := integration.ExampleCCQueryArgs(bKey)

	// Setup tests with a random chaincode ID.
	sdk, targets := initializeLedgerTests(t)

	// Using shared SDK instance to increase test speed.
	//defer client.Close()

	chaincodeID := integration.GenerateExampleID(false)

	if metadata.CCMode == "lscc" {
		err := integration.PrepareExampleCC(sdk, fabsdk.WithUser("Admin"), testSetup.OrgID, chaincodeID)
		require.Nil(t, err, "InstallAndInstantiateExampleCC return error")
	} else {
		err := integration.PrepareExampleCCLc(sdk, fabsdk.WithUser("Admin"), testSetup.OrgID, chaincodeID)
		require.Nil(t, err, "InstallAndInstantiateExampleCC return error")
	}

	//prepare required contexts
	channelClientCtx := sdk.ChannelContext(channelID, fabsdk.WithUser("Admin"), fabsdk.WithOrg(orgName))

	//Reset example cc keys
	integration.ResetKeys(t, channelClientCtx, chaincodeID, "200", aKey, bKey)

	// Get a ledger client.
	ledgerClient, err := ledger.New(channelClientCtx)
	require.Nil(t, err, "ledger new return error")

	// Test Query Info - retrieve values before transaction
	testTargets := targets[0:1]
	bciBeforeTx, err := ledgerClient.QueryInfo(ledger.WithTargetEndpoints(testTargets...))
	if err != nil {
		t.Fatalf("QueryInfo return error: %s", err)
	}

	// Invoke transaction that changes block state
	channelClient, err := channel.New(channelClientCtx)
	if err != nil {
		t.Fatalf("creating channel failed: %s", err)
	}

	txID, expectedQueryValue, err := changeBlockState(t, channelClient, queryArg, moveTxArg, chaincodeID)
	if err != nil {
		t.Fatalf("Failed to change block state (invoke transaction). Return error: %s", err)
	}

	verifyTargetsChangedBlockState(t, channelClient, chaincodeID, targets, queryArg, expectedQueryValue)

	// Test Query Info - retrieve values after transaction
	bciAfterTx, err := ledgerClient.QueryInfo(ledger.WithTargetEndpoints(testTargets...))
	if err != nil {
		t.Fatalf("QueryInfo return error: %s", err)
	}

	// Test Query Info -- verify block size changed after transaction
	if (bciAfterTx.BCI.Height - bciBeforeTx.BCI.Height) <= 0 {
		t.Fatal("Block size did not increase after transaction")
	}

	testQueryTransaction(t, ledgerClient, txID, targets)

	testQueryBlock(t, ledgerClient, targets)

	testQueryBlockByTxID(t, ledgerClient, txID, targets)

	//prepare context
	clientCtx := sdk.Context(fabsdk.WithUser("Admin"), fabsdk.WithOrg(orgName))

	resmgmtClient, err := resmgmt.New(clientCtx)

	require.Nil(t, err, "resmgmt new return error")

	if metadata.CCMode == "lscc" {
		testInstantiatedChaincodes(t, chaincodeID, channelID, resmgmtClient, targets)
	}

	testQueryConfigBlock(t, ledgerClient, targets)
}

func changeBlockState(t *testing.T, client *channel.Client, queryArg [][]byte, moveArg [][]byte, chaincodeID string) (fab.TransactionID, int, error) {

	req := channel.Request{
		ChaincodeID: chaincodeID,
		Fcn:         "invoke",
		Args:        queryArg,
	}
	//resp, err := client.Query(req, channel.WithRetry(retry.DefaultChannelOpts))
	resp, err := retry.NewInvoker(retry.New(retry.TestRetryOpts)).Invoke(
		func() (interface{}, error) {
			response, err := client.Query(req, channel.WithRetry(retry.DefaultChannelOpts))

			if err != nil && strings.Contains(err.Error(), "Nil amount") {
				return nil, status.New(status.TestStatus, status.GenericTransient.ToInt32(), "query funds failed", nil)
			}
			return response, nil
		},
	)
	if err != nil {
		return "", 0, errors.WithMessage(err, "query funds failed")
	}
	value := resp.(channel.Response).Payload

	// Start transaction that will change block state
	txID, err := moveFundsAndGetTxID(t, client, moveArg, chaincodeID)
	if err != nil {
		return "", 0, errors.WithMessage(err, "move funds failed")
	}

	valueInt, _ := strconv.Atoi(string(value))
	valueInt = valueInt + 1

	return txID, valueInt, nil
}

func verifyTargetsChangedBlockState(t *testing.T, client *channel.Client, chaincodeID string, targets []string, queryArg [][]byte, expectedValue int) {
	for _, target := range targets {
		verifyTargetChangedBlockState(t, client, chaincodeID, target, queryArg, expectedValue)
	}
}

func verifyTargetChangedBlockState(t *testing.T, client *channel.Client, chaincodeID string, target string, queryArg [][]byte, expectedValue int) {
	req := channel.Request{
		ChaincodeID: chaincodeID,
		Fcn:         "invoke",
		Args:        queryArg,
	}

	_, err := retry.NewInvoker(retry.New(retry.TestRetryOpts)).Invoke(
		func() (interface{}, error) {
			resp, err := client.Query(req, channel.WithTargetEndpoints(target), channel.WithRetry(retry.DefaultChannelOpts))
			require.NoError(t, err, "query funds failed")

			// Verify that transaction changed block state
			actualValue, _ := strconv.Atoi(string(resp.Payload))
			if expectedValue != actualValue {
				return nil, status.New(status.TestStatus, status.GenericTransient.ToInt32(), fmt.Sprintf("ledger value didn't match expectation [%d, %d]", expectedValue, actualValue), nil)
			}
			return &actualValue, nil
		},
	)
	require.NoError(t, err)
}

func testQueryTransaction(t *testing.T, ledgerClient *ledger.Client, txID fab.TransactionID, targets []string) {

	// Test Query Transaction -- verify that valid transaction has been processed
	processedTransaction, err := ledgerClient.QueryTransaction(txID, ledger.WithTargetEndpoints(targets...))
	if err != nil {
		t.Fatalf("QueryTransaction return error: %s", err)
	}

	if processedTransaction.TransactionEnvelope == nil {
		t.Fatal("QueryTransaction failed to return transaction envelope")
	}

	// Test Query Transaction -- Retrieve non existing transaction
	_, err = ledgerClient.QueryTransaction("123ABC", ledger.WithTargetEndpoints(targets...))
	if err == nil {
		t.Fatal("QueryTransaction non-existing didn't return an error")
	}
}

func testQueryBlock(t *testing.T, ledgerClient *ledger.Client, targets []string) {

	// Retrieve current blockchain info
	bci, err := ledgerClient.QueryInfo(ledger.WithTargetEndpoints(targets...))
	if err != nil {
		t.Fatalf("QueryInfo return error: %s", err)
	}

	// Test Query Block by Hash - retrieve current block by hash
	block, err := ledgerClient.QueryBlockByHash(bci.BCI.CurrentBlockHash, ledger.WithTargetEndpoints(targets...))
	if err != nil {
		t.Fatalf("QueryBlockByHash return error: %s", err)
	}

	if block.Data == nil {
		t.Fatal("QueryBlockByHash block data is nil")
	}

	// Test Query Block by Hash - retrieve block by non-existent hash
	_, err = ledgerClient.QueryBlockByHash([]byte("non-existent"), ledger.WithTargetEndpoints(targets...))
	if err == nil {
		t.Fatal("QueryBlockByHash non-existent didn't return an error")
	}

	// Test Query Block - retrieve block by number
	block, err = ledgerClient.QueryBlock(1, ledger.WithTargetEndpoints(targets...))
	if err != nil {
		t.Fatalf("QueryBlock return error: %s", err)
	}
	if block.Data == nil {
		t.Fatal("QueryBlock block data is nil")
	}

	// Test Query Block - retrieve block by non-existent number
	_, err = ledgerClient.QueryBlock(2147483647, ledger.WithTargetEndpoints(targets...))
	if err == nil {
		t.Fatal("QueryBlock non-existent didn't return an error")
	}
}

func testQueryBlockByTxID(t *testing.T, ledgerClient *ledger.Client, txID fab.TransactionID, targets []string) {

	// Test Query Block- retrieve block by non-existent tx ID
	_, err := ledgerClient.QueryBlockByTxID("non-existent", ledger.WithTargetEndpoints(targets...))
	if err == nil {
		t.Fatal("QueryBlockByTxID non-existent didn't return an error")
	}

	// Test Query Block - retrieve block by valid tx ID
	block, err := ledgerClient.QueryBlockByTxID(txID, ledger.WithTargetEndpoints(targets...))
	if err != nil {
		t.Fatalf("QueryBlockByTxID return error: %s", err)
	}
	if block.Data == nil {
		t.Fatal("QueryBlockByTxID block data is nil")
	}

}

func testInstantiatedChaincodes(t *testing.T, ccID string, channelID string, resmgmtClient *resmgmt.Client, targets []string) {

	found := false

	// Test Query Instantiated chaincodes
	chaincodeQueryResponse, err := resmgmtClient.QueryInstantiatedChaincodes(channelID, resmgmt.WithTargetEndpoints(targets...), resmgmt.WithRetry(retry.DefaultResMgmtOpts))
	if err != nil {
		t.Fatalf("QueryInstantiatedChaincodes return error: %s", err)
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
func moveFundsAndGetTxID(t *testing.T, client *channel.Client, moveArg [][]byte, chaincodeID string) (fab.TransactionID, error) {

	transientDataMap := make(map[string][]byte)
	transientDataMap["result"] = []byte("Transient data in move funds...")

	req := channel.Request{
		ChaincodeID:  chaincodeID,
		Fcn:          "invoke",
		Args:         moveArg,
		TransientMap: transientDataMap,
	}
	resp, err := client.Execute(req, channel.WithRetry(retry.DefaultChannelOpts))
	if err != nil {
		return "", errors.WithMessage(err, "execute move funds failed")
	}

	return resp.TransactionID, nil
}

func testQueryConfigBlock(t *testing.T, ledgerClient *ledger.Client, targets []string) {

	// Retrieve current channel configuration
	cfgEnvelope, err := ledgerClient.QueryConfig(ledger.WithTargetEndpoints(targets...))
	if err != nil {
		t.Fatalf("QueryConfig return error: %s", err)
	}

	if cfgEnvelope == nil {
		t.Fatal("QueryConfig config data is nil")
	}

	block, err := ledgerClient.QueryConfigBlock(ledger.WithTargetEndpoints(targets...))
	if err != nil {
		t.Fatalf("QueryConfigBlock return error: %s", err)
	}

	if block == nil {
		t.Fatal("QueryConfigBlock block is nil")
	}
}
