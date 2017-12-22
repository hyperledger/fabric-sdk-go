/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package sdk

import (
	"path"
	"strconv"
	"testing"

	"github.com/hyperledger/fabric-sdk-go/api/apitxn"
	"github.com/hyperledger/fabric-sdk-go/def/fabapi"
	"github.com/hyperledger/fabric-sdk-go/test/integration"
	"github.com/hyperledger/fabric-sdk-go/test/metadata"

	"github.com/hyperledger/fabric-sdk-go/api/apiconfig"
	fab "github.com/hyperledger/fabric-sdk-go/api/apifabclient"

	"github.com/hyperledger/fabric-sdk-go/def/fabapi/context/defprovider"

	selection "github.com/hyperledger/fabric-sdk-go/pkg/fabric-txn/selection/dynamicselection"
)

func TestDynamicSelection(t *testing.T) {

	testSetup := integration.BaseSetupImpl{
		ConfigFile:      "../" + integration.ConfigTestFile,
		ChannelID:       "mychannel",
		OrgID:           org1Name,
		ChannelConfig:   path.Join("../../", metadata.ChannelConfigPath, "mychannel.tx"),
		ConnectEventHub: true,
	}

	if err := testSetup.Initialize(t); err != nil {
		t.Fatalf(err.Error())
	}

	if err := testSetup.InstallAndInstantiateExampleCC(); err != nil {
		t.Fatalf("InstallAndInstantiateExampleCC return error: %v", err)
	}

	// Specify user that will be used by dynamic selection service (to retrieve chanincode policy information)
	// This user has to have privileges to query lscc for chaincode data
	mychannelUser := selection.ChannelUser{ChannelID: testSetup.ChannelID, UserName: "User1", OrgName: "Org1"}

	// Create SDK setup for channel client with dynamic selection
	sdkOptions := fabapi.Options{
		ConfigFile:      testSetup.ConfigFile,
		ProviderFactory: &DynamicSelectionProviderFactory{ChannelUsers: []selection.ChannelUser{mychannelUser}},
	}

	sdk, err := fabapi.NewSDK(sdkOptions)
	if err != nil {
		t.Fatalf("Failed to create new SDK: %s", err)
	}

	chClient, err := sdk.NewChannelClient(testSetup.ChannelID, "User1")
	if err != nil {
		t.Fatalf("Failed to create new channel client: %s", err)
	}

	// Release all channel client resources
	defer chClient.Close()

	value, err := chClient.Query(apitxn.QueryRequest{ChaincodeID: testSetup.ChainCodeID, Fcn: "invoke", Args: integration.ExampleCCQueryArgs()})
	if err != nil {
		t.Fatalf("Failed to query funds: %s", err)
	}

	// Move funds
	_, err = chClient.ExecuteTx(apitxn.ExecuteTxRequest{ChaincodeID: testSetup.ChainCodeID, Fcn: "invoke", Args: integration.ExampleCCTxArgs()})
	if err != nil {
		t.Fatalf("Failed to move funds: %s", err)
	}

	// Verify move funds transaction result
	valueAfterInvoke, err := chClient.Query(apitxn.QueryRequest{ChaincodeID: testSetup.ChainCodeID, Fcn: "invoke", Args: integration.ExampleCCQueryArgs()})
	if err != nil {
		t.Fatalf("Failed to query funds after transaction: %s", err)
	}

	valueInt, _ := strconv.Atoi(string(value))
	valueAfterInvokeInt, _ := strconv.Atoi(string(valueAfterInvoke))
	if valueInt+1 != valueAfterInvokeInt {
		t.Fatalf("ExecuteTx failed. Before: %s, after: %s", value, valueAfterInvoke)
	}

}

// DynamicSelectionProviderFactory is configured with dynamic (endorser) selection provider
type DynamicSelectionProviderFactory struct {
	defprovider.DefaultProviderFactory
	ChannelUsers []selection.ChannelUser
}

// NewSelectionProvider returns a new implementation of dynamic selection provider
func (f *DynamicSelectionProviderFactory) NewSelectionProvider(config apiconfig.Config) (fab.SelectionProvider, error) {
	return selection.NewSelectionProvider(config, f.ChannelUsers, nil)
}
