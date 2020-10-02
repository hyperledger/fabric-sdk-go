/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package e2e

import (
	"strconv"
	"testing"
	"time"

	"github.com/hyperledger/fabric-sdk-go/pkg/common/errors/retry"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/fab"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/msp"
	"github.com/hyperledger/fabric-sdk-go/test/metadata"
	"github.com/stretchr/testify/require"

	pb "github.com/hyperledger/fabric-protos-go/peer"
	"github.com/hyperledger/fabric-sdk-go/test/integration"
	"github.com/hyperledger/fabric-sdk-go/third_party/github.com/hyperledger/fabric/common/policydsl"

	"github.com/hyperledger/fabric-sdk-go/pkg/client/channel"
	"github.com/hyperledger/fabric-sdk-go/pkg/client/resmgmt"

	mspclient "github.com/hyperledger/fabric-sdk-go/pkg/client/msp"

	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/core"
	packager "github.com/hyperledger/fabric-sdk-go/pkg/fab/ccpackager/gopackager"
	lcpackager "github.com/hyperledger/fabric-sdk-go/pkg/fab/ccpackager/lifecycle"
	"github.com/hyperledger/fabric-sdk-go/pkg/fabsdk"
)

const (
	channelID      = "mychannel"
	orgName        = "Org1"
	orgAdmin       = "Admin"
	ordererOrgName = "OrdererOrg"
	peer1          = "peer0.org1.example.com"
)

var (
	ccID = "example_cc_fabtest_e2e" + metadata.TestRunID
)

// Run enables testing an end-to-end scenario against the supplied SDK options
func Run(t *testing.T, configOpt core.ConfigProvider, sdkOpts ...fabsdk.Option) {
	setupAndRun(t, true, configOpt, e2eTest, sdkOpts...)
}

// RunWithoutSetup will execute the same way as Run but without creating a new channel and registering a new CC
func RunWithoutSetup(t *testing.T, configOpt core.ConfigProvider, sdkOpts ...fabsdk.Option) {
	setupAndRun(t, false, configOpt, e2eTest, sdkOpts...)
}

type testSDKFunc func(t *testing.T, sdk *fabsdk.FabricSDK)

// setupAndRun enables testing an end-to-end scenario against the supplied SDK options
// the createChannel flag will be used to either create a channel and the example CC or not(ie run the tests with existing ch and CC)
func setupAndRun(t *testing.T, createChannel bool, configOpt core.ConfigProvider, test testSDKFunc, sdkOpts ...fabsdk.Option) {

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

	if createChannel {
		createChannelAndCC(t, sdk)
	}

	test(t, sdk)
}

func e2eTest(t *testing.T, sdk *fabsdk.FabricSDK) {
	//prepare channel client context using client context
	clientChannelContext := sdk.ChannelContext(channelID, fabsdk.WithUser("User1"), fabsdk.WithOrg(orgName))
	// Channel client is used to query and execute transactions (Org1 is default org)
	client, err := channel.New(clientChannelContext)
	if err != nil {
		t.Fatalf("Failed to create new channel client: %s", err)
	}

	existingValue := queryCC(t, client)
	ccEvent := moveFunds(t, client)

	// Verify move funds transaction result on the same peer where the event came from.
	verifyFundsIsMoved(t, client, existingValue, ccEvent)
}

func createChannelAndCC(t *testing.T, sdk *fabsdk.FabricSDK) {
	//clientContext allows creation of transactions using the supplied identity as the credential.
	clientContext := sdk.Context(fabsdk.WithUser(orgAdmin), fabsdk.WithOrg(ordererOrgName))

	// Resource management client is responsible for managing channels (create/update channel)
	// Supply user that has privileges to create channel (in this case orderer admin)
	resMgmtClient, err := resmgmt.New(clientContext)
	if err != nil {
		t.Fatalf("Failed to create channel management client: %s", err)
	}

	// Create channel
	createChannel(t, sdk, resMgmtClient)

	//prepare context
	adminContext := sdk.Context(fabsdk.WithUser(orgAdmin), fabsdk.WithOrg(orgName))

	// Org resource management client
	orgResMgmt, err := resmgmt.New(adminContext)
	if err != nil {
		t.Fatalf("Failed to create new resource management client: %s", err)
	}

	// Org peers join channel
	if err = orgResMgmt.JoinChannel(channelID, resmgmt.WithRetry(retry.DefaultResMgmtOpts), resmgmt.WithOrdererEndpoint("orderer.example.com")); err != nil {
		t.Fatalf("Org peers failed to JoinChannel: %s", err)
	}

	// Create chaincode package for example cc
	if metadata.CCMode == "lscc" {
		createCC(t, orgResMgmt)
	} else {
		createCCLifecycle(t, orgResMgmt, sdk)
	}
}

func moveFunds(t *testing.T, client *channel.Client) *fab.CCEvent {

	eventID := "test([a-zA-Z]+)"

	// Register chaincode event (pass in channel which receives event details when the event is complete)
	reg, notifier, err := client.RegisterChaincodeEvent(ccID, eventID)
	if err != nil {
		t.Fatalf("Failed to register cc event: %s", err)
	}
	defer client.UnregisterChaincodeEvent(reg)

	// Move funds
	executeCC(t, client)

	var ccEvent *fab.CCEvent
	select {
	case ccEvent = <-notifier:
		t.Logf("Received CC event: %#v\n", ccEvent)
	case <-time.After(time.Second * 20):
		t.Fatalf("Did NOT receive CC event for eventId(%s)\n", eventID)
	}

	return ccEvent
}

func verifyFundsIsMoved(t *testing.T, client *channel.Client, value []byte, ccEvent *fab.CCEvent) {

	newValue := queryCC(t, client, ccEvent.SourceURL)
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

func executeCC(t *testing.T, client *channel.Client) {
	_, err := client.Execute(channel.Request{ChaincodeID: ccID, Fcn: "invoke", Args: integration.ExampleCCDefaultTxArgs()},
		channel.WithRetry(retry.DefaultChannelOpts))
	if err != nil {
		t.Fatalf("Failed to move funds: %s", err)
	}
}

func queryCC(t *testing.T, client *channel.Client, targetEndpoints ...string) []byte {
	response, err := client.Query(channel.Request{ChaincodeID: ccID, Fcn: "invoke", Args: integration.ExampleCCDefaultQueryArgs()},
		channel.WithRetry(retry.DefaultChannelOpts),
		channel.WithTargetEndpoints(targetEndpoints...),
	)
	if err != nil {
		t.Fatalf("Failed to query funds: %s", err)
	}
	return response.Payload
}

func createCC(t *testing.T, orgResMgmt *resmgmt.Client) {
	ccPkg, err := packager.NewCCPackage("github.com/example_cc", integration.GetDeployPath())
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
	ccPolicy := policydsl.SignedByAnyMember([]string{"Org1MSP"})
	// Org resource manager will instantiate 'example_cc' on channel
	resp, err := orgResMgmt.InstantiateCC(
		channelID,
		resmgmt.InstantiateCCRequest{Name: ccID, Path: "github.com/example_cc", Version: "0", Args: integration.ExampleCCInitArgs(), Policy: ccPolicy},
		resmgmt.WithRetry(retry.DefaultResMgmtOpts),
	)
	require.Nil(t, err, "error should be nil")
	require.NotEmpty(t, resp, "transaction response should be populated")
}

func createChannel(t *testing.T, sdk *fabsdk.FabricSDK, resMgmtClient *resmgmt.Client) {
	mspClient, err := mspclient.New(sdk.Context(), mspclient.WithOrg(orgName))
	if err != nil {
		t.Fatal(err)
	}
	adminIdentity, err := mspClient.GetSigningIdentity(orgAdmin)
	if err != nil {
		t.Fatal(err)
	}
	req := resmgmt.SaveChannelRequest{ChannelID: channelID,
		ChannelConfigPath: integration.GetChannelConfigTxPath(channelID + ".tx"),
		SigningIdentities: []msp.SigningIdentity{adminIdentity}}
	txID, err := resMgmtClient.SaveChannel(req, resmgmt.WithRetry(retry.DefaultResMgmtOpts), resmgmt.WithOrdererEndpoint("orderer.example.com"))
	require.Nil(t, err, "error should be nil")
	require.NotEmpty(t, txID, "transaction ID should be populated")
}

// createCCLifecycle package cc, install cc, get installed cc package, query installed cc
// approve cc, query approve cc, check commit readiness, commit cc, query committed cc
func createCCLifecycle(t *testing.T, orgResMgmt *resmgmt.Client, sdk *fabsdk.FabricSDK) {
	// Package cc
	label, ccPkg := packageCC(t)
	packageID := lcpackager.ComputePackageID(label, ccPkg)

	// Install cc
	installCC(t, label, ccPkg, orgResMgmt)

	// Get installed cc package
	getInstalledCCPackage(t, packageID, ccPkg, orgResMgmt)

	// Query installed cc
	queryInstalled(t, label, packageID, orgResMgmt)

	// Approve cc
	approveCC(t, packageID, orgResMgmt)

	// Query approve cc
	queryApprovedCC(t, orgResMgmt)

	// Check commit readiness
	checkCCCommitReadiness(t, orgResMgmt)

	// Commit cc
	commitCC(t, orgResMgmt)

	// Query committed cc
	queryCommittedCC(t, orgResMgmt)

	// Init cc
	initCC(t, sdk)

}

func packageCC(t *testing.T) (string, []byte) {
	desc := &lcpackager.Descriptor{
		Path:  integration.GetLcDeployPath(),
		Type:  pb.ChaincodeSpec_GOLANG,
		Label: "example_cc_fabtest_e2e_0",
	}
	ccPkg, err := lcpackager.NewCCPackage(desc)
	if err != nil {
		t.Fatal(err)
	}
	return desc.Label, ccPkg
}

func installCC(t *testing.T, label string, ccPkg []byte, orgResMgmt *resmgmt.Client) {
	installCCReq := resmgmt.LifecycleInstallCCRequest{
		Label:   label,
		Package: ccPkg,
	}

	packageID := lcpackager.ComputePackageID(installCCReq.Label, installCCReq.Package)

	resp, err := orgResMgmt.LifecycleInstallCC(installCCReq, resmgmt.WithRetry(retry.DefaultResMgmtOpts))
	if err != nil {
		t.Fatal(err)
	}
	require.Equal(t, packageID, resp[0].PackageID)
}

func getInstalledCCPackage(t *testing.T, packageID string, ccPkg []byte, orgResMgmt *resmgmt.Client) {
	resp, err := orgResMgmt.LifecycleGetInstalledCCPackage(packageID, resmgmt.WithTargetEndpoints(peer1), resmgmt.WithRetry(retry.DefaultResMgmtOpts))
	if err != nil {
		t.Fatal(err)
	}
	require.Equal(t, ccPkg, resp)
}

func queryInstalled(t *testing.T, label string, packageID string, orgResMgmt *resmgmt.Client) {
	resp, err := orgResMgmt.LifecycleQueryInstalledCC(resmgmt.WithTargetEndpoints(peer1), resmgmt.WithRetry(retry.DefaultResMgmtOpts))
	if err != nil {
		t.Fatal(err)
	}
	require.Equal(t, packageID, resp[0].PackageID)
	require.Equal(t, label, resp[0].Label)
}

func approveCC(t *testing.T, packageID string, orgResMgmt *resmgmt.Client) {
	ccPolicy := policydsl.SignedByAnyMember([]string{"Org1MSP"})
	approveCCReq := resmgmt.LifecycleApproveCCRequest{
		Name:              ccID,
		Version:           "0",
		PackageID:         packageID,
		Sequence:          1,
		EndorsementPlugin: "escc",
		ValidationPlugin:  "vscc",
		SignaturePolicy:   ccPolicy,
		InitRequired:      true,
	}

	txnID, err := orgResMgmt.LifecycleApproveCC(channelID, approveCCReq, resmgmt.WithTargetEndpoints(peer1), resmgmt.WithOrdererEndpoint("orderer.example.com"), resmgmt.WithRetry(retry.DefaultResMgmtOpts))
	if err != nil {
		t.Fatal(err)
	}
	require.NotEmpty(t, txnID)
}

func queryApprovedCC(t *testing.T, orgResMgmt *resmgmt.Client) {
	queryApprovedCCReq := resmgmt.LifecycleQueryApprovedCCRequest{
		Name:     ccID,
		Sequence: 1,
	}
	resp, err := orgResMgmt.LifecycleQueryApprovedCC(channelID, queryApprovedCCReq, resmgmt.WithTargetEndpoints(peer1), resmgmt.WithRetry(retry.DefaultResMgmtOpts))
	if err != nil {
		t.Fatal(err)
	}
	require.NotNil(t, resp)
}

func checkCCCommitReadiness(t *testing.T, orgResMgmt *resmgmt.Client) {
	ccPolicy := policydsl.SignedByAnyMember([]string{"Org1MSP"})
	req := resmgmt.LifecycleCheckCCCommitReadinessRequest{
		Name:              ccID,
		Version:           "0",
		EndorsementPlugin: "escc",
		ValidationPlugin:  "vscc",
		SignaturePolicy:   ccPolicy,
		Sequence:          1,
		InitRequired:      true,
	}
	resp, err := orgResMgmt.LifecycleCheckCCCommitReadiness(channelID, req, resmgmt.WithTargetEndpoints(peer1), resmgmt.WithRetry(retry.DefaultResMgmtOpts))
	if err != nil {
		t.Fatal(err)
	}
	require.NotNil(t, resp)
}

func commitCC(t *testing.T, orgResMgmt *resmgmt.Client) {
	ccPolicy := policydsl.SignedByAnyMember([]string{"Org1MSP"})
	req := resmgmt.LifecycleCommitCCRequest{
		Name:              ccID,
		Version:           "0",
		Sequence:          1,
		EndorsementPlugin: "escc",
		ValidationPlugin:  "vscc",
		SignaturePolicy:   ccPolicy,
		InitRequired:      true,
	}
	txnID, err := orgResMgmt.LifecycleCommitCC(channelID, req, resmgmt.WithRetry(retry.DefaultResMgmtOpts), resmgmt.WithTargetEndpoints(peer1), resmgmt.WithOrdererEndpoint("orderer.example.com"))
	if err != nil {
		t.Fatal(err)
	}
	require.NotEmpty(t, txnID)
}

func queryCommittedCC(t *testing.T, orgResMgmt *resmgmt.Client) {
	req := resmgmt.LifecycleQueryCommittedCCRequest{
		Name: ccID,
	}
	resp, err := orgResMgmt.LifecycleQueryCommittedCC(channelID, req, resmgmt.WithTargetEndpoints(peer1), resmgmt.WithRetry(retry.DefaultResMgmtOpts))
	if err != nil {
		t.Fatal(err)
	}
	require.Equal(t, ccID, resp[0].Name)
}

func initCC(t *testing.T, sdk *fabsdk.FabricSDK) {
	//prepare channel client context using client context
	clientChannelContext := sdk.ChannelContext(channelID, fabsdk.WithUser("User1"), fabsdk.WithOrg(orgName))
	// Channel client is used to query and execute transactions (Org1 is default org)
	client, err := channel.New(clientChannelContext)
	if err != nil {
		t.Fatalf("Failed to create new channel client: %s", err)
	}

	// init
	_, err = client.Execute(channel.Request{ChaincodeID: ccID, Fcn: "init", Args: integration.ExampleCCInitArgsLc(), IsInit: true},
		channel.WithRetry(retry.DefaultChannelOpts))
	if err != nil {
		t.Fatalf("Failed to init: %s", err)
	}

}
