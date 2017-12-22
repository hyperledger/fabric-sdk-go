/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package fab

import (
	"path"
	"strconv"
	"testing"
	"time"

	config "github.com/hyperledger/fabric-sdk-go/api/apiconfig"
	fab "github.com/hyperledger/fabric-sdk-go/api/apifabclient"
	"github.com/hyperledger/fabric-sdk-go/api/apitxn"
	"github.com/hyperledger/fabric-sdk-go/test/integration"
	"github.com/hyperledger/fabric-sdk-go/test/metadata"

	"github.com/hyperledger/fabric-sdk-go/pkg/errors"
	peer "github.com/hyperledger/fabric-sdk-go/pkg/fabric-client/peer"
)

func TestChannelQueries(t *testing.T) {

	testSetup := &integration.BaseSetupImpl{
		ConfigFile:      "../" + integration.ConfigTestFile,
		ChannelID:       "mychannel",
		OrgID:           org1Name,
		ChannelConfig:   path.Join("../../../", metadata.ChannelConfigPath, "mychannel.tx"),
		ConnectEventHub: true,
	}

	if err := testSetup.Initialize(t); err != nil {
		t.Fatalf(err.Error())
	}

	channel := testSetup.Channel
	client := testSetup.Client

	if err := testSetup.InstallAndInstantiateExampleCC(); err != nil {
		t.Fatalf("InstallAndInstantiateExampleCC return error: %v", err)
	}

	// Test Query Info - retrieve values before transaction
	bciBeforeTx, err := channel.QueryInfo()
	if err != nil {
		t.Fatalf("QueryInfo return error: %v", err)
	}

	// Invoke transaction that changes block state
	txID, err := changeBlockState(t, testSetup)
	if err != nil {
		t.Fatalf("Failed to change block state (invoke transaction). Return error: %v", err)
	}

	// Test Query Info - retrieve values after transaction
	bciAfterTx, err := channel.QueryInfo()
	if err != nil {
		t.Fatalf("QueryInfo return error: %v", err)
	}

	// Test Query Info -- verify block size changed after transaction
	if (bciAfterTx.Height - bciBeforeTx.Height) <= 0 {
		t.Fatalf("Block size did not increase after transaction")
	}

	testQueryTransaction(t, channel, txID)

	testQueryBlock(t, channel)

	testQueryChannels(t, channel, client)

	testInstalledChaincodes(t, channel, client, testSetup)

	testQueryByChaincode(t, channel, client.Config(), testSetup)

	// TODO: Synch with test in node SDK when it becomes available
	// testInstantiatedChaincodes(t, channel)

}

func changeBlockState(t *testing.T, testSetup *integration.BaseSetupImpl) (string, error) {

	tpResponses, _, err := testSetup.CreateAndSendTransactionProposal(testSetup.Channel, testSetup.ChainCodeID, "invoke", integration.ExampleCCQueryArgs(), []apitxn.ProposalProcessor{testSetup.Channel.PrimaryPeer()}, nil)
	if err != nil {
		return "", errors.WithMessage(err, "CreateAndSendTransactionProposal failed")
	}

	value := tpResponses[0].ProposalResponse.GetResponse().Payload

	// Start transaction that will change block state
	txID, err := moveFundsAndGetTxID(t, testSetup)
	if err != nil {
		return "", errors.WithMessage(err, "move funds failed")
	}

	tpResponses, _, err = testSetup.CreateAndSendTransactionProposal(testSetup.Channel, testSetup.ChainCodeID, "invoke", integration.ExampleCCQueryArgs(), []apitxn.ProposalProcessor{testSetup.Channel.PrimaryPeer()}, nil)
	if err != nil {
		return "", errors.WithMessage(err, "CreateAndSendTransactionProposal failed")
	}

	valueAfterInvoke := tpResponses[0].ProposalResponse.GetResponse().Payload

	// Verify that transaction changed block state
	valueInt, _ := strconv.Atoi(string(value))
	valueInt = valueInt + 1
	valueAfterInvokeInt, _ := strconv.Atoi(string(valueAfterInvoke))
	if valueInt != valueAfterInvokeInt {
		return "", errors.Errorf("SendTransaction didn't change the QueryValue %s", value)
	}

	return txID, nil
}

func testQueryTransaction(t *testing.T, channel fab.Channel, txID string) {

	// Test Query Transaction -- verify that valid transaction has been processed
	processedTransaction, err := channel.QueryTransaction(txID)
	if err != nil {
		t.Fatalf("QueryTransaction return error: %v", err)
	}

	if processedTransaction.TransactionEnvelope == nil {
		t.Fatalf("QueryTransaction failed to return transaction envelope")
	}

	// Test Query Transaction -- Retrieve non existing transaction
	processedTransaction, err = channel.QueryTransaction("123ABC")
	if err == nil {
		t.Fatalf("QueryTransaction non-existing didn't return an error")
	}

}

func testQueryBlock(t *testing.T, channel fab.Channel) {

	// Retrieve current blockchain info
	bci, err := channel.QueryInfo()
	if err != nil {
		t.Fatalf("QueryInfo return error: %v", err)
	}

	// Test Query Block by Hash - retrieve current block by hash
	block, err := channel.QueryBlockByHash(bci.CurrentBlockHash)
	if err != nil {
		t.Fatalf("QueryBlockByHash return error: %v", err)
	}

	if block.Data == nil {
		t.Fatalf("QueryBlockByHash block data is nil")
	}

	// Test Query Block by Hash - retrieve block by non-existent hash
	block, err = channel.QueryBlockByHash([]byte("non-existent"))
	if err == nil {
		t.Fatalf("QueryBlockByHash non-existent didn't return an error")
	}

	// Test Query Block - retrieve block by number
	block, err = channel.QueryBlock(1)
	if err != nil {
		t.Fatalf("QueryBlock return error: %v", err)
	}

	if block.Data == nil {
		t.Fatalf("QueryBlock block data is nil")
	}

	// Test Query Block - retrieve block by non-existent number
	block, err = channel.QueryBlock(2147483647)
	if err == nil {
		t.Fatalf("QueryBlock non-existent didn't return an error")
	}

}

func testQueryChannels(t *testing.T, channel fab.Channel, client fab.FabricClient) {

	// Our target will be primary peer on this channel
	target := channel.PrimaryPeer()
	t.Logf("****QueryChannels for %s", target.URL())
	channelQueryResponse, err := client.QueryChannels(target)
	if err != nil {
		t.Fatalf("QueryChannels return error: %v", err)
	}

	for _, channel := range channelQueryResponse.Channels {
		t.Logf("**Channel: %s", channel)
	}

}

func testInstalledChaincodes(t *testing.T, channel fab.Channel, client fab.FabricClient, testSetup *integration.BaseSetupImpl) {

	// Our target will be primary peer on this channel
	target := channel.PrimaryPeer()
	t.Logf("****QueryInstalledChaincodes for %s", target.URL())

	chaincodeQueryResponse, err := client.QueryInstalledChaincodes(target)
	if err != nil {
		t.Fatalf("QueryInstalledChaincodes return error: %v", err)
	}

	for _, chaincode := range chaincodeQueryResponse.Chaincodes {
		t.Logf("**InstalledCC: %s", chaincode)
	}

}

func testInstantiatedChaincodes(t *testing.T, channel fab.Channel) {

	// Our target will indirectly be primary peer on this channel
	target := channel.PrimaryPeer()

	t.Logf("QueryInstantiatedChaincodes for primary %s", target.URL())

	// Test Query Instantiated chaincodes
	chaincodeQueryResponse, err := channel.QueryInstantiatedChaincodes()
	if err != nil {
		t.Fatalf("QueryInstantiatedChaincodes return error: %v", err)
	}

	for _, chaincode := range chaincodeQueryResponse.Chaincodes {
		t.Logf("**InstantiatedCC: %s", chaincode)
	}

}

func testQueryByChaincode(t *testing.T, channel fab.Channel, config config.Config, testSetup *integration.BaseSetupImpl) {

	// Test valid targets
	targets := peer.PeersToTxnProcessors(channel.Peers())

	// set Client User Context to Admin before calling QueryByChaincode
	testSetup.Client.SetUserContext(testSetup.AdminUser)

	request := apitxn.ChaincodeInvokeRequest{
		Targets:     targets,
		ChaincodeID: "lscc",
		Fcn:         "getinstalledchaincodes",
	}
	queryResponses, err := channel.QueryBySystemChaincode(request)
	if err != nil {
		t.Fatalf("QueryByChaincode failed %s", err)
	}

	// Number of responses should be the same as number of targets
	if len(queryResponses) != len(targets) {
		t.Fatalf("QueryByChaincode number of results mismatch. Expected: %d Got: %d", len(targets), len(queryResponses))
	}

	// Configured cert for cert pool
	cert, err := config.CAClientCertPath(org1Name)
	if err != nil {
		t.Fatal(err)
	}

	// Create invalid target
	firstInvalidTarget, err := peer.NewPeerTLSFromCert("test:1111", cert, "", config)
	if err != nil {
		t.Fatalf("Create NewPeer error(%v)", err)
	}

	// Create second invalid target
	secondInvalidTarget, err := peer.NewPeerTLSFromCert("test:2222", cert, "", config)
	if err != nil {
		t.Fatalf("Create NewPeer error(%v)", err)
	}

	// Add invalid targets to targets
	invalidTargets := append(targets, firstInvalidTarget)
	invalidTargets = append(invalidTargets, secondInvalidTarget)

	// Add invalid targets to channel otherwise validation will fail
	err = channel.AddPeer(firstInvalidTarget)
	if err != nil {
		t.Fatalf("Error adding peer: %v", err)
	}
	err = channel.AddPeer(secondInvalidTarget)
	if err != nil {
		t.Fatalf("Error adding peer: %v", err)
	}

	// Test valid + invalid targets
	request = apitxn.ChaincodeInvokeRequest{
		ChaincodeID: "lscc",
		Fcn:         "getinstalledchaincodes",
		Targets:     invalidTargets,
	}
	queryResponses, err = channel.QueryBySystemChaincode(request)
	if err == nil {
		t.Fatalf("QueryByChaincode failed to return error for non-existing target")
	}

	// Verify that valid targets returned response
	if len(queryResponses) != len(targets) {
		t.Fatalf("QueryByChaincode number of results mismatch. Expected: %d Got: %d", len(targets), len(queryResponses))
	}

	channel.RemovePeer(firstInvalidTarget)
	channel.RemovePeer(secondInvalidTarget)
}

// MoveFundsAndGetTxID ...
func moveFundsAndGetTxID(t *testing.T, setup *integration.BaseSetupImpl) (string, error) {

	transientDataMap := make(map[string][]byte)
	transientDataMap["result"] = []byte("Transient data in move funds...")

	transactionProposalResponse, txID, err := setup.CreateAndSendTransactionProposal(setup.Channel, setup.ChainCodeID, "invoke", integration.ExampleCCTxArgs(), []apitxn.ProposalProcessor{setup.Channel.PrimaryPeer()}, transientDataMap)
	if err != nil {
		return "", errors.WithMessage(err, "CreateAndSendTransactionProposal failed")
	}
	// Register for commit event
	done, fail := setup.RegisterTxEvent(t, txID, setup.EventHub)

	txResponse, err := setup.CreateAndSendTransaction(setup.Channel, transactionProposalResponse)
	if err != nil {
		return "", errors.WithMessage(err, "CreateAndSendTransaction failed")
	}
	t.Logf("txResponse: %v", txResponse)
	select {
	case <-done:
	case cerr := <-fail:
		return "", errors.Wrapf(cerr, "invoke failed for txid %s", txID)
	case <-time.After(time.Second * 30):
		return "", errors.Errorf("invoke didn't receive block event for txid %s", txID)
	}
	return txID.ID, nil
}
