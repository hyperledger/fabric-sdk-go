/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package fab

import (
	"path"
	"strconv"
	"testing"

	fab "github.com/hyperledger/fabric-sdk-go/api/apifabclient"
	"github.com/hyperledger/fabric-sdk-go/api/apitxn/chclient"
	chmgmt "github.com/hyperledger/fabric-sdk-go/api/apitxn/chmgmtclient"
	"github.com/hyperledger/fabric-sdk-go/test/integration"
	"github.com/hyperledger/fabric-sdk-go/test/metadata"

	"github.com/hyperledger/fabric-sdk-go/pkg/config"
	"github.com/hyperledger/fabric-sdk-go/pkg/fabsdk"
	"github.com/pkg/errors"
)

const (
	sdkConfigFile     = "../" + integration.ConfigTestFile
	channelConfigFile = "mychannel.tx"
	channelID         = "mychannel"
	orgName           = org1Name
)

func initializeLedgerTests(t *testing.T) (*fabsdk.FabricSDK, []fab.ProposalProcessor) {
	sdk, err := fabsdk.New(config.FromFile(sdkConfigFile))
	if err != nil {
		t.Fatalf("SDK init failed: %v", err)
	}

	session, err := sdk.NewClient(fabsdk.WithUser("Admin"), fabsdk.WithOrg(orgName)).Session()
	if err != nil {
		t.Fatalf("failed getting admin user session for org: %s", err)
	}

	targets, err := integration.CreateProposalProcessors(sdk.Config(), []string{orgName})
	if err != nil {
		t.Fatalf("creating peers failed: %v", err)
	}

	channelConfig := path.Join("../../../", metadata.ChannelConfigPath, channelConfigFile)
	req := chmgmt.SaveChannelRequest{ChannelID: channelID, ChannelConfig: channelConfig, SigningIdentity: session}
	err = integration.InitializeChannel(sdk, orgName, req, targets)
	if err != nil {
		t.Fatalf("failed to ensure channel has been initialized: %s", err)
	}
	return sdk, targets
}

func TestLedgerQueries(t *testing.T) {

	// Setup tests with a random chaincode ID.
	sdk, targets := initializeLedgerTests(t)
	chaincodeID := integration.GenerateRandomID()
	if err := integration.InstallAndInstantiateExampleCC(sdk, fabsdk.WithUser("Admin"), orgName, chaincodeID); err != nil {
		t.Fatalf("InstallAndInstantiateExampleCC return error: %v", err)
	}

	// Get a ledger client.
	client := sdk.NewClient(fabsdk.WithUser("Admin"), fabsdk.WithOrg(orgName))
	channelSvc, err := client.ChannelService(channelID)
	if err != nil {
		t.Fatalf("creating channel service failed: %v", err)
	}
	ledger, err := channelSvc.Ledger()
	if err != nil {
		t.Fatalf("creating channel ledger client failed: %v", err)
	}

	// Test Query Info - retrieve values before transaction
	bciBeforeTx, err := ledger.QueryInfo(targets[0:1])
	if err != nil {
		t.Fatalf("QueryInfo return error: %v", err)
	}

	// Invoke transaction that changes block state
	channel, err := client.Channel(channelID)
	if err != nil {
		t.Fatalf("creating channel failed: %v", err)
	}

	txID, err := changeBlockState(t, channel, chaincodeID)
	if err != nil {
		t.Fatalf("Failed to change block state (invoke transaction). Return error: %v", err)
	}

	// Test Query Info - retrieve values after transaction
	bciAfterTx, err := ledger.QueryInfo(targets[0:1])
	if err != nil {
		t.Fatalf("QueryInfo return error: %v", err)
	}

	// Test Query Info -- verify block size changed after transaction
	if (bciAfterTx[0].Height - bciBeforeTx[0].Height) <= 0 {
		t.Fatalf("Block size did not increase after transaction")
	}

	testQueryTransaction(t, ledger, txID, targets)

	testQueryBlock(t, ledger, targets)

	testInstantiatedChaincodes(t, chaincodeID, ledger, targets)

}

func changeBlockState(t *testing.T, channel chclient.ChannelClient, chaincodeID string) (string, error) {

	req := chclient.Request{
		ChaincodeID: chaincodeID,
		Fcn:         "invoke",
		Args:        integration.ExampleCCQueryArgs(),
	}
	resp, err := channel.Query(req)
	if err != nil {
		return "", errors.WithMessage(err, "query funds failed")
	}
	value := resp.Payload

	// Start transaction that will change block state
	txID, err := moveFundsAndGetTxID(t, channel, chaincodeID)
	if err != nil {
		return "", errors.WithMessage(err, "move funds failed")
	}

	resp, err = channel.Query(req)
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

func testQueryTransaction(t *testing.T, ledger fab.ChannelLedger, txID string, targets []fab.ProposalProcessor) {

	// Test Query Transaction -- verify that valid transaction has been processed
	processedTransactions, err := ledger.QueryTransaction(txID, targets)
	if err != nil {
		t.Fatalf("QueryTransaction return error: %v", err)
	}

	for _, processedTransaction := range processedTransactions {
		if processedTransaction.TransactionEnvelope == nil {
			t.Fatalf("QueryTransaction failed to return transaction envelope")
		}
	}

	// Test Query Transaction -- Retrieve non existing transaction
	_, err = ledger.QueryTransaction("123ABC", targets)
	if err == nil {
		t.Fatalf("QueryTransaction non-existing didn't return an error")
	}
}

func testQueryBlock(t *testing.T, ledger fab.ChannelLedger, targets []fab.ProposalProcessor) {

	// Retrieve current blockchain info
	bcis, err := ledger.QueryInfo(targets)
	if err != nil {
		t.Fatalf("QueryInfo return error: %v", err)
	}

	for i, bci := range bcis {
		// Test Query Block by Hash - retrieve current block by hash
		block, err := ledger.QueryBlockByHash(bci.CurrentBlockHash, targets[i:i+1])
		if err != nil {
			t.Fatalf("QueryBlockByHash return error: %v", err)
		}

		if block[0].Data == nil {
			t.Fatalf("QueryBlockByHash block data is nil")
		}
	}

	// Test Query Block by Hash - retrieve block by non-existent hash
	_, err = ledger.QueryBlockByHash([]byte("non-existent"), targets)
	if err == nil {
		t.Fatalf("QueryBlockByHash non-existent didn't return an error")
	}

	// Test Query Block - retrieve block by number
	blocks, err := ledger.QueryBlock(1, targets)
	if err != nil {
		t.Fatalf("QueryBlock return error: %v", err)
	}
	for _, block := range blocks {
		if block.Data == nil {
			t.Fatalf("QueryBlock block data is nil")
		}
	}

	// Test Query Block - retrieve block by non-existent number
	_, err = ledger.QueryBlock(2147483647, targets)
	if err == nil {
		t.Fatalf("QueryBlock non-existent didn't return an error")
	}
}

func testInstantiatedChaincodes(t *testing.T, ccID string, ledger fab.ChannelLedger, targets []fab.ProposalProcessor) {

	// Test Query Instantiated chaincodes
	chaincodeQueryResponses, err := ledger.QueryInstantiatedChaincodes(targets)
	if err != nil {
		t.Fatalf("QueryInstantiatedChaincodes return error: %v", err)
	}

	found := false
	for _, chaincodeQueryResponse := range chaincodeQueryResponses {
		for _, chaincode := range chaincodeQueryResponse.Chaincodes {
			t.Logf("**InstantiatedCC: %s", chaincode)
			if chaincode.Name == ccID {
				found = true
			}
		}
	}

	if !found {
		t.Fatalf("QueryInstantiatedChaincodes failed to find instantiated %s chaincode", ccID)
	}
}

// MoveFundsAndGetTxID ...
func moveFundsAndGetTxID(t *testing.T, channel chclient.ChannelClient, chaincodeID string) (string, error) {

	transientDataMap := make(map[string][]byte)
	transientDataMap["result"] = []byte("Transient data in move funds...")

	req := chclient.Request{
		ChaincodeID:  chaincodeID,
		Fcn:          "invoke",
		Args:         integration.ExampleCCTxArgs(),
		TransientMap: transientDataMap,
	}
	resp, err := channel.Execute(req)
	if err != nil {
		return "", errors.WithMessage(err, "execute move funds failed")
	}

	return resp.TransactionID.ID, nil
}
