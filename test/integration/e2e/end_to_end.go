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

	"github.com/hyperledger/fabric-sdk-go/pkg/common/errors/retry"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/msp"
	"github.com/stretchr/testify/require"

	"github.com/hyperledger/fabric-sdk-go/test/integration"
	"github.com/hyperledger/fabric-sdk-go/test/metadata"
	"github.com/hyperledger/fabric-sdk-go/third_party/github.com/hyperledger/fabric/common/cauthdsl"

	"github.com/hyperledger/fabric-sdk-go/pkg/client/channel"
	"github.com/hyperledger/fabric-sdk-go/pkg/client/resmgmt"

	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/core"
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

// Run enables testing an end-to-end scenario against the supplied SDK options
func Run(t *testing.T, configOpt core.ConfigProvider, sdkOpts ...fabsdk.Option) {

	if integration.IsLocal() {
		//If it is a local test then add entity mapping to config backend to parse URLs
		configOpt = integration.AddLocalEntityMapping(configOpt)
	}

	sdk, err := fabsdk.New(configOpt, sdkOpts...)
	if err != nil {
		t.Fatalf("Failed to create new SDK: %s", err)
	}
	defer sdk.Close()

	// Delete all private keys from the crypto suite store
	// and users from the user store at the end
	integration.CleanupUserData(t, sdk)
	defer integration.CleanupUserData(t, sdk)

	//clientContext allows creation of transactions using the supplied identity as the credential.
	clientContext := sdk.Context(fabsdk.WithUser(orgAdmin), fabsdk.WithOrg(ordererOrgName))

	// Resource management client is responsible for managing channels (create/update channel)
	// Supply user that has privileges to create channel (in this case orderer admin)
	resMgmtClient, err := resmgmt.New(clientContext)
	if err != nil {
		t.Fatalf("Failed to create channel management client: %s", err)
	}

	// Create channel

	// Org admin user is signing user for creating channel

	createChannel(sdk, t, resMgmtClient)

	//prepare context
	adminContext := sdk.Context(fabsdk.WithUser(orgAdmin), fabsdk.WithOrg(orgName))

	// Org resource management client
	orgResMgmt, err := resmgmt.New(adminContext)
	if err != nil {
		t.Fatalf("Failed to create new resource management client: %s", err)
	}

	// Org peers join channel
	if err = orgResMgmt.JoinChannel(channelID, resmgmt.WithRetry(retry.DefaultResMgmtOpts)); err != nil {
		t.Fatalf("Org peers failed to JoinChannel: %s", err)
	}

	// Create chaincode package for example cc
	createCC(t, orgResMgmt)

	// ************ Test setup complete ************** //

	//prepare channel client context using client context
	clientChannelContext := sdk.ChannelContext(channelID, fabsdk.WithUser("User1"), fabsdk.WithOrg(orgName))
	// Channel client is used to query and execute transactions (Org1 is default org)
	client, err := channel.New(clientChannelContext)
	if err != nil {
		t.Fatalf("Failed to create new channel client: %s", err)
	}

	value := queryCC(client, t)

	eventID := "test([a-zA-Z]+)"

	// Register chaincode event (pass in channel which receives event details when the event is complete)
	reg, notifier, err := client.RegisterChaincodeEvent(ccID, eventID)
	if err != nil {
		t.Fatalf("Failed to register cc event: %s", err)
	}
	defer client.UnregisterChaincodeEvent(reg)

	// Move funds
	executeCC(client, t)

	select {
	case ccEvent := <-notifier:
		t.Logf("Received CC event: %#v\n", ccEvent)
	case <-time.After(time.Second * 20):
		t.Fatalf("Did NOT receive CC event for eventId(%s)\n", eventID)
	}

	// Verify move funds transaction result
	verifyFundsIsMoved(client, t, value)

}

func verifyFundsIsMoved(client *channel.Client, t *testing.T, value []byte) {
	newValue := queryCC(client, t)
	valueInt, err := strconv.Atoi(string(value))
	if err != nil {
		t.Fatal(err.Error())
	}
	valueAfterInvokeInt, err := strconv.Atoi(string(newValue))
	if err != nil {
		t.Fatal(err.Error())
	}
	if valueInt+1 != valueAfterInvokeInt {
		t.Fatalf("Execute failed. Before: %s, after: %s", value, newValue)
	}
}

func executeCC(client *channel.Client, t *testing.T) {
	_, err := client.Execute(channel.Request{ChaincodeID: ccID, Fcn: "invoke", Args: integration.ExampleCCTxArgs()},
		channel.WithRetry(retry.DefaultChannelOpts))
	if err != nil {
		t.Fatalf("Failed to move funds: %s", err)
	}
}

func queryCC(client *channel.Client, t *testing.T) []byte {
	response, err := client.Query(channel.Request{ChaincodeID: ccID, Fcn: "invoke", Args: integration.ExampleCCQueryArgs()},
		channel.WithRetry(retry.DefaultChannelOpts))
	if err != nil {
		t.Fatalf("Failed to query funds: %s", err)
	}
	return response.Payload
}

func createCC(t *testing.T, orgResMgmt *resmgmt.Client) {
	ccPkg, err := packager.NewCCPackage("github.com/example_cc", "../../fixtures/testdata")
	if err != nil {
		t.Fatal(err)
	}
	// Install example cc to org peers
	installCCReq := resmgmt.InstallCCRequest{Name: ccID, Path: "github.com/example_cc", Version: "0", Package: ccPkg}
	_, err = orgResMgmt.InstallCC(installCCReq, resmgmt.WithRetry(retry.DefaultResMgmtOpts))
	if err != nil {
		t.Fatal(err)
	}
	// Set up chaincode policy
	ccPolicy := cauthdsl.SignedByAnyMember([]string{"Org1MSP"})
	// Org resource manager will instantiate 'example_cc' on channel
	resp, err := orgResMgmt.InstantiateCC(
		channelID,
		resmgmt.InstantiateCCRequest{Name: ccID, Path: "github.com/example_cc", Version: "0", Args: integration.ExampleCCInitArgs(), Policy: ccPolicy},
		resmgmt.WithRetry(retry.DefaultResMgmtOpts),
	)
	require.Nil(t, err, "error should be nil")
	require.NotEmpty(t, resp, "transaction response should be populated")
}

func createChannel(sdk *fabsdk.FabricSDK, t *testing.T, resMgmtClient *resmgmt.Client) {
	adminIdentity, err := integration.GetSigningIdentity(sdk, orgAdmin, orgName)
	if err != nil {
		t.Fatal(err)
	}
	req := resmgmt.SaveChannelRequest{ChannelID: channelID,
		ChannelConfigPath: path.Join("../../../", metadata.ChannelConfigPath, "mychannel.tx"),
		SigningIdentities: []msp.SigningIdentity{adminIdentity}}
	txID, err := resMgmtClient.SaveChannel(req, resmgmt.WithRetry(retry.DefaultResMgmtOpts))
	require.Nil(t, err, "error should be nil")
	require.NotEmpty(t, txID, "transaction ID should be populated")
}
