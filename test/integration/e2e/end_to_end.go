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

	"github.com/hyperledger/fabric-sdk-go/pkg/context/api/core"

	"github.com/hyperledger/fabric-sdk-go/test/integration"
	"github.com/hyperledger/fabric-sdk-go/test/metadata"
	"github.com/hyperledger/fabric-sdk-go/third_party/github.com/hyperledger/fabric/common/cauthdsl"

	"github.com/hyperledger/fabric-sdk-go/pkg/client/channel"
	"github.com/hyperledger/fabric-sdk-go/pkg/client/resmgmt"

	"github.com/hyperledger/fabric-sdk-go/pkg/context/api/fab"
	"github.com/hyperledger/fabric-sdk-go/pkg/core/config"
	packager "github.com/hyperledger/fabric-sdk-go/pkg/fab/ccpackager/gopackager"
	"github.com/hyperledger/fabric-sdk-go/pkg/fabsdk"
)

const (
	channelID      = "mychannel"
	orgName        = "Org1"
	orgAdmin       = "Admin"
	ordererOrgName = "ordererorg"
	ccID           = "e2eExampleCC"
)

func runWithConfigFixture(t *testing.T) {
	Run(t, config.FromFile("../"+integration.ConfigTestFile))
}

// Run enables testing an end-to-end scenario against the supplied SDK options
func Run(t *testing.T, configOpt core.ConfigProvider, sdkOpts ...fabsdk.Option) {

	sdk, err := fabsdk.New(configOpt, sdkOpts...)
	if err != nil {
		t.Fatalf("Failed to create new SDK: %s", err)
	}
	defer sdk.Close()

	//clientContext allows creation of transactions using the supplied identity as the credential.
	clientContext := sdk.Context(fabsdk.WithUser(orgAdmin), fabsdk.WithOrg(ordererOrgName))

	// Channel management client is responsible for managing channels (create/update channel)
	// Supply user that has privileges to create channel (in this case orderer admin)
	chMgmtClient, err := resmgmt.New(clientContext)
	if err != nil {
		t.Fatalf("Failed to create channel management client: %s", err)
	}

	// Create channel

	// Org admin user is signing user for creating channel

	adminIdentity, err := integration.GetSigningIdentity(sdk, orgAdmin, orgName)
	if err != nil {
		t.Fatal(err)
	}

	req := resmgmt.SaveChannelRequest{ChannelID: channelID,
		ChannelConfig:     path.Join("../../../", metadata.ChannelConfigPath, "mychannel.tx"),
		SigningIdentities: []fab.IdentityContext{adminIdentity}}
	if err = chMgmtClient.SaveChannel(req); err != nil {
		t.Fatal(err)
	}

	// Allow orderer to process channel creation
	time.Sleep(time.Second * 5)

	//prepare context
	adminContext := sdk.Context(fabsdk.WithUser(orgAdmin), fabsdk.WithOrg(orgName))

	// Org resource management client
	orgResMgmt, err := resmgmt.New(adminContext)
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

	//prepare channel client context using client context
	clientChannelContext := sdk.ChannelContext(channelID, fabsdk.WithUser("User1"), fabsdk.WithOrg(orgName))
	// Channel client is used to query and execute transactions (Org1 is default org)
	client, err := channel.New(clientChannelContext)
	if err != nil {
		t.Fatalf("Failed to create new channel client: %s", err)
	}

	// Release all channel client resources
	defer client.Close()

	response, err := client.Query(channel.Request{ChaincodeID: ccID, Fcn: "invoke", Args: integration.ExampleCCQueryArgs()})
	if err != nil {
		t.Fatalf("Failed to query funds: %s", err)
	}
	value := response.Payload

	eventID := "test([a-zA-Z]+)"

	// Register chaincode event (pass in channel which receives event details when the event is complete)
	notifier := make(chan *channel.CCEvent)
	rce, err := client.RegisterChaincodeEvent(notifier, ccID, eventID)
	if err != nil {
		t.Fatalf("Failed to register cc event: %s", err)
	}

	// Move funds
	response, err = client.Execute(channel.Request{ChaincodeID: ccID, Fcn: "invoke", Args: integration.ExampleCCTxArgs()})
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
	err = client.UnregisterChaincodeEvent(rce)
	if err != nil {
		t.Fatalf("Unregister cc event failed: %s", err)
	}

	// Verify move funds transaction result
	response, err = client.Query(channel.Request{ChaincodeID: ccID, Fcn: "invoke", Args: integration.ExampleCCQueryArgs()})
	if err != nil {
		t.Fatalf("Failed to query funds after transaction: %s", err)
	}

	valueInt, _ := strconv.Atoi(string(value))
	valueAfterInvokeInt, _ := strconv.Atoi(string(response.Payload))
	if valueInt+1 != valueAfterInvokeInt {
		t.Fatalf("Execute failed. Before: %s, after: %s", value, response.Payload)
	}

}
