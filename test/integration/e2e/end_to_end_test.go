/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package e2e

import (
	"path"
	"strconv"
	"testing"
	"time"

	"github.com/hyperledger/fabric-sdk-go/api/apitxn"
	"github.com/hyperledger/fabric-sdk-go/def/fabapi"
	"github.com/hyperledger/fabric-sdk-go/test/integration"
	"github.com/hyperledger/fabric-sdk-go/test/metadata"
	"github.com/hyperledger/fabric-sdk-go/third_party/github.com/hyperledger/fabric/common/cauthdsl"

	chmgmt "github.com/hyperledger/fabric-sdk-go/api/apitxn/chmgmtclient"
	resmgmt "github.com/hyperledger/fabric-sdk-go/api/apitxn/resmgmtclient"

	packager "github.com/hyperledger/fabric-sdk-go/pkg/fabric-client/ccpackager/gopackager"
)

const (
	channelID = "mychannel"
	orgName   = "Org1"
	orgAdmin  = "Admin"
	ccID      = "e2eExampleCC"
)

func TestE2E(t *testing.T) {

	// Create SDK setup for the integration tests
	sdkOptions := fabapi.Options{
		ConfigFile: "../" + integration.ConfigTestFile,
	}

	sdk, err := fabapi.NewSDK(sdkOptions)
	if err != nil {
		t.Fatalf("Failed to create new SDK: %s", err)
	}

	// Channel management client is responsible for managing channels (create/update channel)
	// Supply user that has privileges to create channel (in this case orderer admin)
	chMgmtClient, err := sdk.NewChannelMgmtClientWithOpts("Admin", &fabapi.ChannelMgmtClientOpts{OrgName: "ordererorg"})
	if err != nil {
		t.Fatalf("Failed to create channel management client: %s", err)
	}

	// Org admin user is signing user for creating channel
	orgAdminUser, err := sdk.NewPreEnrolledUser(orgName, orgAdmin)
	if err != nil {
		t.Fatalf("NewPreEnrolledUser failed for %s, %s: %s", orgName, orgAdmin, err)
	}

	// Create channel
	req := chmgmt.SaveChannelRequest{ChannelID: channelID, ChannelConfig: path.Join("../../../", metadata.ChannelConfigPath, "mychannel.tx"), SigningUser: orgAdminUser}
	if err = chMgmtClient.SaveChannel(req); err != nil {
		t.Fatal(err)
	}

	// Allow orderer to process channel creation
	time.Sleep(time.Second * 3)

	// Org resource management client (Org1 is default org)
	orgResMgmt, err := sdk.NewResourceMgmtClient(orgAdmin)
	if err != nil {
		t.Fatalf("Failed to create new resource management client: %s", err)
	}

	// Org peers join channel
	if err = orgResMgmt.JoinChannel(channelID); err != nil {
		t.Fatalf("Org peers failed to JoinChannel: %s", err)
	}

	// Create chaincode package for example cc
	ccPkg, err := packager.NewCCPackage("github.com/example_cc", "../../fixtures/testdata")
	if err != nil {
		t.Fatal(err)
	}

	// Install example cc to org peers
	installCCReq := resmgmt.InstallCCRequest{Name: ccID, Path: "github.com/example_cc", Version: "0", Package: ccPkg}
	_, err = orgResMgmt.InstallCC(installCCReq)
	if err != nil {
		t.Fatal(err)
	}

	// Set up chaincode policy
	ccPolicy := cauthdsl.SignedByAnyMember([]string{"Org1MSP"})

	// Org resource manager will instantiate 'example_cc' on channel
	err = orgResMgmt.InstantiateCC(channelID, resmgmt.InstantiateCCRequest{Name: ccID, Path: "github.com/example_cc", Version: "0", Args: integration.ExampleCCInitArgs(), Policy: ccPolicy})
	if err != nil {
		t.Fatal(err)
	}

	// ************ Test setup complete ************** //

	// Channel client is used to query and execute transactions
	chClient, err := sdk.NewChannelClient(channelID, "User1")
	if err != nil {
		t.Fatalf("Failed to create new channel client: %s", err)
	}

	value, err := chClient.Query(apitxn.QueryRequest{ChaincodeID: ccID, Fcn: "invoke", Args: integration.ExampleCCQueryArgs()})
	if err != nil {
		t.Fatalf("Failed to query funds: %s", err)
	}

	eventID := "test([a-zA-Z]+)"

	// Register chaincode event (pass in channel which receives event details when the event is complete)
	notifier := make(chan *apitxn.CCEvent)
	rce := chClient.RegisterChaincodeEvent(notifier, ccID, eventID)

	// Move funds
	_, err = chClient.ExecuteTx(apitxn.ExecuteTxRequest{ChaincodeID: ccID, Fcn: "invoke", Args: integration.ExampleCCTxArgs()})
	if err != nil {
		t.Fatalf("Failed to move funds: %s", err)
	}

	select {
	case ccEvent := <-notifier:
		t.Logf("Received CC event: %s\n", ccEvent)
	case <-time.After(time.Second * 20):
		t.Fatalf("Did NOT receive CC event for eventId(%s)\n", eventID)
	}

	// Unregister chain code event using registration handle
	err = chClient.UnregisterChaincodeEvent(rce)
	if err != nil {
		t.Fatalf("Unregister cc event failed: %s", err)
	}

	// Verify move funds transaction result
	valueAfterInvoke, err := chClient.Query(apitxn.QueryRequest{ChaincodeID: ccID, Fcn: "invoke", Args: integration.ExampleCCQueryArgs()})
	if err != nil {
		t.Fatalf("Failed to query funds after transaction: %s", err)
	}

	valueInt, _ := strconv.Atoi(string(value))
	valueAfterInvokeInt, _ := strconv.Atoi(string(valueAfterInvoke))
	if valueInt+1 != valueAfterInvokeInt {
		t.Fatalf("ExecuteTx failed. Before: %s, after: %s", value, valueAfterInvoke)
	}

	// Release all channel client resources
	chClient.Close()

}
