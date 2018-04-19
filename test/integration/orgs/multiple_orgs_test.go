/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package orgs

import (
	"math"
	"path"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/hyperledger/fabric-sdk-go/pkg/common/errors/multi"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/errors/retry"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/errors/status"
	contextAPI "github.com/hyperledger/fabric-sdk-go/pkg/common/providers/context"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/msp"
	packager "github.com/hyperledger/fabric-sdk-go/pkg/fab/ccpackager/gopackager"
	"github.com/hyperledger/fabric-sdk-go/pkg/fabsdk"
	"github.com/stretchr/testify/require"

	"github.com/hyperledger/fabric-sdk-go/pkg/client/ledger"
	"github.com/hyperledger/fabric-sdk-go/pkg/client/resmgmt"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/fab"

	"github.com/hyperledger/fabric-sdk-go/pkg/fabsdk/factory/defsvc"
	"github.com/hyperledger/fabric-sdk-go/test/integration"
	"github.com/hyperledger/fabric-sdk-go/test/metadata"

	selection "github.com/hyperledger/fabric-sdk-go/pkg/client/common/selection/dynamicselection"

	"github.com/hyperledger/fabric-sdk-go/pkg/client/channel"
	"github.com/hyperledger/fabric-sdk-go/pkg/fab/resource/api"
	"github.com/hyperledger/fabric-sdk-go/third_party/github.com/hyperledger/fabric/common/cauthdsl"
)

const (
	pollRetries      = 5
	org1             = "Org1"
	org2             = "Org2"
	ordererAdminUser = "Admin"
	ordererOrgName   = "ordererorg"
	org1AdminUser    = "Admin"
	org2AdminUser    = "Admin"
	org1User         = "User1"
	org2User         = "User1"
	channelID        = "orgchannel"
)

// Peers
var orgTestPeer0 fab.Peer
var orgTestPeer1 fab.Peer

// TestOrgsEndToEnd creates a channel with two organisations, installs chaincode
// on each of them, and finally invokes a transaction on an org2 peer and queries
// the result from an org1 peer
func TestOrgsEndToEnd(t *testing.T) {
	// Create SDK setup for the integration tests
	sdk, err := fabsdk.New(integration.ConfigBackend)
	if err != nil {
		t.Fatalf("Failed to create new SDK: %s", err)
	}
	defer sdk.Close()

	// Delete all private keys from the crypto suite store
	// and users from the user store at the end
	integration.CleanupUserData(t, sdk)
	defer integration.CleanupUserData(t, sdk)

	// Load specific targets for move funds test
	loadOrgPeers(t, sdk.Context(fabsdk.WithUser(org1AdminUser), fabsdk.WithOrg(org1)))

	expectedValue := testWithOrg1(t, sdk)
	expectedValue = testWithOrg2(t, expectedValue)
	verifyWithOrg1(t, sdk, expectedValue)
}

func testWithOrg1(t *testing.T, sdk *fabsdk.FabricSDK) int {

	//prepare contexts
	ordererClientContext := sdk.Context(fabsdk.WithUser(ordererAdminUser), fabsdk.WithOrg(ordererOrgName))
	org1AdminClientContext := sdk.Context(fabsdk.WithUser(org1AdminUser), fabsdk.WithOrg(org1))
	org2AdminClientContext := sdk.Context(fabsdk.WithUser(org2AdminUser), fabsdk.WithOrg(org2))
	org1AdminChannelContext := sdk.ChannelContext(channelID, fabsdk.WithUser(org1AdminUser), fabsdk.WithOrg(org1))
	org1ChannelClientContext := sdk.ChannelContext(channelID, fabsdk.WithUser(org1User), fabsdk.WithOrg(org1))
	org2ChannelClientContext := sdk.ChannelContext(channelID, fabsdk.WithUser(org2User), fabsdk.WithOrg(org2))

	// Channel management client is responsible for managing channels (create/update channel)
	chMgmtClient, err := resmgmt.New(ordererClientContext)
	if err != nil {
		t.Fatal(err)
	}

	// Get signing identity that is used to sign create channel request
	org1AdminUser, err := integration.GetSigningIdentity(sdk, org1AdminUser, org1)
	if err != nil {
		t.Fatalf("failed to get org1AdminUser, err : %v", err)
	}

	org2AdminUser, err := integration.GetSigningIdentity(sdk, org2AdminUser, org2)
	if err != nil {
		t.Fatalf("failed to get org2AdminUser, err : %v", err)
	}

	createChannel(org1AdminUser, org2AdminUser, chMgmtClient, t)

	// Org1 resource management client (Org1 is default org)
	org1ResMgmt, err := resmgmt.New(org1AdminClientContext)
	if err != nil {
		t.Fatalf("Failed to create new resource management client: %s", err)
	}

	// Org1 peers join channel
	if err = org1ResMgmt.JoinChannel("orgchannel", resmgmt.WithRetry(retry.DefaultResMgmtOpts)); err != nil {
		t.Fatalf("Org1 peers failed to JoinChannel: %s", err)
	}

	// Org2 resource management client
	org2ResMgmt, err := resmgmt.New(org2AdminClientContext)
	if err != nil {
		t.Fatal(err)
	}

	// Org2 peers join channel
	if err = org2ResMgmt.JoinChannel("orgchannel", resmgmt.WithRetry(retry.DefaultResMgmtOpts)); err != nil {
		t.Fatalf("Org2 peers failed to JoinChannel: %s", err)
	}

	ccPkg, err := packager.NewCCPackage("github.com/example_cc", "../../fixtures/testdata")
	if err != nil {
		t.Fatal(err)
	}

	// Create chaincode package for example cc
	createCC(t, org1ResMgmt, org2ResMgmt, ccPkg)

	chClientOrg1User, chClientOrg2User := connectUserToOrgChannel(org1ChannelClientContext, t, org2ChannelClientContext)

	// Call with a dummy function and expect a fail with multiple errors
	verifyErrorFromCC(chClientOrg1User, t)

	// Org1 user queries initial value on both peers
	value := queryCC(chClientOrg1User, t)
	initial, _ := strconv.Atoi(string(value))

	ledgerClient, err := ledger.New(org1AdminChannelContext)
	if err != nil {
		t.Fatalf("Failed to create new ledger client: %s", err)
	}

	// Ledger client will verify blockchain info
	ledgerInfoBefore := getBlockchainInfo(ledgerClient, t)

	// Org2 user moves funds on org2 peer
	transactionID := moveFunds(chClientOrg2User, t)

	// Assert that funds have changed value on org1 peer
	verifyValue(t, chClientOrg1User, initial+1)

	// Get latest block chain info
	checkLedgerInfo(ledgerClient, t, ledgerInfoBefore, transactionID)

	// Start chaincode upgrade process (install and instantiate new version of exampleCC)
	upgradeCC(ccPkg, org1ResMgmt, t, org2ResMgmt)

	// Org2 user moves funds on org2 peer (cc policy fails since both Org1 and Org2 peers should participate)
	testCCPolicy(chClientOrg2User, t)

	// Assert that funds have changed value on org1 peer
	beforeTxValue, _ := strconv.Atoi(integration.ExampleCCUpgradeB)
	expectedValue := beforeTxValue + 1
	verifyValue(t, chClientOrg1User, expectedValue)

	return expectedValue
}

func connectUserToOrgChannel(org1ChannelClientContext contextAPI.ChannelProvider, t *testing.T, org2ChannelClientContext contextAPI.ChannelProvider) (*channel.Client, *channel.Client) {
	// Org1 user connects to 'orgchannel'
	chClientOrg1User, err := channel.New(org1ChannelClientContext)
	if err != nil {
		t.Fatalf("Failed to create new channel client for Org1 user: %s", err)
	}
	// Org2 user connects to 'orgchannel'
	chClientOrg2User, err := channel.New(org2ChannelClientContext)
	if err != nil {
		t.Fatalf("Failed to create new channel client for Org2 user: %s", err)
	}
	return chClientOrg1User, chClientOrg2User
}

func checkLedgerInfo(ledgerClient *ledger.Client, t *testing.T, ledgerInfoBefore *fab.BlockchainInfoResponse, transactionID fab.TransactionID) {
	ledgerInfoAfter, err := ledgerClient.QueryInfo(ledger.WithTargets(orgTestPeer0.(fab.Peer), orgTestPeer1.(fab.Peer)), ledger.WithMinTargets(2))
	if err != nil {
		t.Fatalf("QueryInfo return error: %v", err)
	}
	if ledgerInfoAfter.BCI.Height-ledgerInfoBefore.BCI.Height <= 0 {
		t.Fatalf("Block size did not increase after transaction")
	}
	// Test Query Block by Hash - retrieve current block by number
	block, err := ledgerClient.QueryBlock(ledgerInfoAfter.BCI.Height-1, ledger.WithTargets(orgTestPeer0.(fab.Peer), orgTestPeer1.(fab.Peer)), ledger.WithMinTargets(2))
	if err != nil {
		t.Fatalf("QueryBlock return error: %v", err)
	}
	if block == nil {
		t.Fatalf("Block info not available")
	}
	// Get transaction info
	transactionInfo, err := ledgerClient.QueryTransaction(transactionID, ledger.WithTargets(orgTestPeer0.(fab.Peer), orgTestPeer1.(fab.Peer)), ledger.WithMinTargets(2))
	if err != nil {
		t.Fatalf("QueryTransaction return error: %v", err)
	}
	if transactionInfo.TransactionEnvelope == nil {
		t.Fatalf("Transaction info missing")
	}
}

func createChannel(org1AdminUser msp.SigningIdentity, org2AdminUser msp.SigningIdentity, chMgmtClient *resmgmt.Client, t *testing.T) {
	req := resmgmt.SaveChannelRequest{ChannelID: "orgchannel",
		ChannelConfigPath: path.Join("../../../", metadata.ChannelConfigPath, "orgchannel.tx"),
		SigningIdentities: []msp.SigningIdentity{org1AdminUser, org2AdminUser}}
	txID, err := chMgmtClient.SaveChannel(req, resmgmt.WithRetry(retry.DefaultResMgmtOpts))
	require.Nil(t, err, "error should be nil")
	require.NotEmpty(t, txID, "transaction ID should be populated")
}

func testCCPolicy(chClientOrg2User *channel.Client, t *testing.T) {
	_, err := chClientOrg2User.Execute(channel.Request{ChaincodeID: "exampleCC", Fcn: "invoke", Args: integration.ExampleCCTxArgs()}, channel.WithTargets(orgTestPeer1),
		channel.WithRetry(retry.DefaultChannelOpts))
	if err == nil {
		t.Fatalf("Should have failed to move funds due to cc policy")
	}
	// Org2 user moves funds (cc policy ok since we have provided peers for both Orgs)
	_, err = chClientOrg2User.Execute(channel.Request{ChaincodeID: "exampleCC", Fcn: "invoke", Args: integration.ExampleCCTxArgs()}, channel.WithTargets(orgTestPeer0, orgTestPeer1),
		channel.WithRetry(retry.DefaultChannelOpts))
	if err != nil {
		t.Fatalf("Failed to move funds: %s", err)
	}
}

func upgradeCC(ccPkg *api.CCPackage, org1ResMgmt *resmgmt.Client, t *testing.T, org2ResMgmt *resmgmt.Client) {
	installCCReq := resmgmt.InstallCCRequest{Name: "exampleCC", Path: "github.com/example_cc", Version: "1", Package: ccPkg}
	// Install example cc version '1' to Org1 peers
	_, err := org1ResMgmt.InstallCC(installCCReq, resmgmt.WithRetry(retry.DefaultResMgmtOpts))
	if err != nil {
		t.Fatal(err)
	}
	// Install example cc version '1' to Org2 peers
	_, err = org2ResMgmt.InstallCC(installCCReq, resmgmt.WithRetry(retry.DefaultResMgmtOpts))
	if err != nil {
		t.Fatal(err)
	}
	// New chaincode policy (both orgs have to approve)
	org1Andorg2Policy, err := cauthdsl.FromString("AND ('Org1MSP.member','Org2MSP.member')")
	if err != nil {
		t.Fatal(err)
	}
	// Org1 resource manager will instantiate 'example_cc' version 1 on 'orgchannel'
	upgradeResp, err := org1ResMgmt.UpgradeCC("orgchannel", resmgmt.UpgradeCCRequest{Name: "exampleCC", Path: "github.com/example_cc", Version: "1", Args: integration.ExampleCCUpgradeArgs(), Policy: org1Andorg2Policy})
	require.Nil(t, err, "error should be nil")
	require.NotEmpty(t, upgradeResp, "transaction response should be populated")
}

func moveFunds(chClientOrgUser *channel.Client, t *testing.T) fab.TransactionID {
	response, err := chClientOrgUser.Execute(channel.Request{ChaincodeID: "exampleCC", Fcn: "invoke", Args: integration.ExampleCCTxArgs()}, channel.WithTargets(orgTestPeer1),
		channel.WithRetry(retry.DefaultChannelOpts))
	if err != nil {
		t.Fatalf("Failed to move funds: %s", err)
	}
	if response.ChaincodeStatus == 0 {
		t.Fatalf("Expected ChaincodeStatus")
	}
	if response.Responses[0].ChaincodeStatus != response.ChaincodeStatus {
		t.Fatalf("Expected the chaincode status returned by successful Peer Endorsement to be same as Chaincode status for client response")
	}
	return response.TransactionID
}

func getBlockchainInfo(ledgerClient *ledger.Client, t *testing.T) *fab.BlockchainInfoResponse {
	channelCfg, err := ledgerClient.QueryConfig(ledger.WithTargets(orgTestPeer0, orgTestPeer1), ledger.WithMinTargets(2))
	if err != nil {
		t.Fatalf("QueryConfig return error: %v", err)
	}
	if len(channelCfg.Orderers()) == 0 {
		t.Fatalf("Failed to retrieve channel orderers")
	}
	expectedOrderer := "orderer.example.com"
	if !strings.Contains(channelCfg.Orderers()[0], expectedOrderer) {
		t.Fatalf("Expecting %s, got %s", expectedOrderer, channelCfg.Orderers()[0])
	}
	ledgerInfoBefore, err := ledgerClient.QueryInfo(ledger.WithTargets(orgTestPeer0, orgTestPeer1), ledger.WithMinTargets(2), ledger.WithMaxTargets(3))
	if err != nil {
		t.Fatalf("QueryInfo return error: %v", err)
	}
	// Test Query Block by Hash - retrieve current block by hash
	block, err := ledgerClient.QueryBlockByHash(ledgerInfoBefore.BCI.CurrentBlockHash, ledger.WithTargets(orgTestPeer0.(fab.Peer), orgTestPeer1.(fab.Peer)), ledger.WithMinTargets(2))
	if err != nil {
		t.Fatalf("QueryBlockByHash return error: %v", err)
	}
	if block == nil {
		t.Fatalf("Block info not available")
	}
	return ledgerInfoBefore
}

func queryCC(chClientOrg1User *channel.Client, t *testing.T) []byte {
	response, err := chClientOrg1User.Query(channel.Request{ChaincodeID: "exampleCC", Fcn: "invoke", Args: integration.ExampleCCQueryArgs()},
		channel.WithRetry(retry.DefaultChannelOpts))
	if err != nil {
		t.Fatalf("Failed to query funds: %s", err)
	}
	if response.ChaincodeStatus == 0 {
		t.Fatalf("Expected ChaincodeStatus")
	}
	if response.Responses[0].ChaincodeStatus != response.ChaincodeStatus {
		t.Fatalf("Expected the chaincode status returned by successful Peer Endorsement to be same as Chaincode status for client response")
	}
	return response.Payload
}

func verifyErrorFromCC(chClientOrg1User *channel.Client, t *testing.T) {
	_, err := chClientOrg1User.Query(channel.Request{ChaincodeID: "exampleCC", Fcn: "DUMMY_FUNCTION", Args: integration.ExampleCCQueryArgs()},
		channel.WithRetry(retry.DefaultChannelOpts))
	require.Error(t, err, "Should have failed with dummy function")
	s, ok := status.FromError(err)
	require.True(t, ok, "expected status error")
	require.Equal(t, s.Code, int32(status.MultipleErrors))
	for _, err := range err.(multi.Errors) {
		s, ok := status.FromError(err)
		require.True(t, ok, "expected status error")
		require.EqualValues(t, int32(500), s.Code)
		require.Equal(t, status.ChaincodeStatus, s.Group)
	}
}

func createCC(t *testing.T, org1ResMgmt *resmgmt.Client, org2ResMgmt *resmgmt.Client, ccPkg *api.CCPackage) {
	installCCReq := resmgmt.InstallCCRequest{Name: "exampleCC", Path: "github.com/example_cc", Version: "0", Package: ccPkg}
	// Install example cc to Org1 peers
	_, err := org1ResMgmt.InstallCC(installCCReq, resmgmt.WithRetry(retry.DefaultResMgmtOpts))
	if err != nil {
		t.Fatal(err)
	}
	// Install example cc to Org2 peers
	_, err = org2ResMgmt.InstallCC(installCCReq, resmgmt.WithRetry(retry.DefaultResMgmtOpts))
	if err != nil {
		t.Fatal(err)
	}
	// Set up chaincode policy to 'any of two msps'
	ccPolicy := cauthdsl.SignedByAnyMember([]string{"Org1MSP", "Org2MSP"})
	// Org1 resource manager will instantiate 'example_cc' on 'orgchannel'
	instantiateResp, err := org1ResMgmt.InstantiateCC("orgchannel", resmgmt.InstantiateCCRequest{Name: "exampleCC", Path: "github.com/example_cc", Version: "0", Args: integration.ExampleCCInitArgs(), Policy: ccPolicy}, resmgmt.WithRetry(retry.DefaultResMgmtOpts))
	require.Nil(t, err, "error should be nil")
	require.NotEmpty(t, instantiateResp, "transaction response should be populated")
	// Verify that example CC is instantiated on Org1 peer
	chaincodeQueryResponse, err := org1ResMgmt.QueryInstantiatedChaincodes("orgchannel", resmgmt.WithRetry(retry.DefaultResMgmtOpts))
	if err != nil {
		t.Fatalf("QueryInstantiatedChaincodes return error: %v", err)
	}
	found := false
	for _, chaincode := range chaincodeQueryResponse.Chaincodes {
		if chaincode.Name == "exampleCC" {
			found = true
		}
	}
	if !found {
		t.Fatalf("QueryInstantiatedChaincodes failed to find instantiated exampleCC chaincode")
	}

}

func testWithOrg2(t *testing.T, expectedValue int) int {

	// Specify user that will be used by dynamic selection service (to retrieve chanincode policy information)
	// This user has to have privileges to query lscc for chaincode data
	mychannelUser := selection.ChannelUser{ChannelID: "orgchannel", Username: "User1", OrgName: "Org1"}

	// Create SDK setup for channel client with dynamic selection
	sdk, err := fabsdk.New(integration.ConfigBackend,
		fabsdk.WithServicePkg(&DynamicSelectionProviderFactory{ChannelUsers: []selection.ChannelUser{mychannelUser}}))
	if err != nil {
		t.Fatalf("Failed to create new SDK: %s", err)
	}
	defer sdk.Close()

	//prepare contexts
	org2ChannelClientContext := sdk.ChannelContext(channelID, fabsdk.WithUser(org2User), fabsdk.WithOrg(org2))

	// Create new client that will use dynamic selection
	chClientOrg2User, err := channel.New(org2ChannelClientContext)
	if err != nil {
		t.Fatalf("Failed to create new channel client for Org2 user: %s", err)
	}

	// Org2 user moves funds (dynamic selection will inspect chaincode policy to determine endorsers)
	_, err = chClientOrg2User.Execute(channel.Request{ChaincodeID: "exampleCC", Fcn: "invoke", Args: integration.ExampleCCTxArgs()},
		channel.WithRetry(retry.DefaultChannelOpts))
	if err != nil {
		t.Fatalf("Failed to move funds: %s", err)
	}

	expectedValue++
	return expectedValue
}

func verifyWithOrg1(t *testing.T, sdk *fabsdk.FabricSDK, expectedValue int) {
	//prepare context
	org1ChannelClientContext := sdk.ChannelContext(channelID, fabsdk.WithUser(org1User), fabsdk.WithOrg(org1))

	// Org1 user connects to 'orgchannel'
	chClientOrg1User, err := channel.New(org1ChannelClientContext)
	if err != nil {
		t.Fatalf("Failed to create new channel client for Org1 user: %s", err)
	}

	verifyValue(t, chClientOrg1User, expectedValue)
}

func verifyValue(t *testing.T, chClient *channel.Client, expected int) {

	// Assert that funds have changed value on org1 peer
	var valueInt int
	for i := 0; i < pollRetries; i++ {
		// Query final value on org1 peer
		response, err := chClient.Query(channel.Request{ChaincodeID: "exampleCC", Fcn: "invoke", Args: integration.ExampleCCQueryArgs()}, channel.WithTargets(orgTestPeer0),
			channel.WithRetry(retry.DefaultChannelOpts))
		if err != nil {
			t.Fatalf("Failed to query funds after transaction: %s", err)
		}
		// If value has not propogated sleep with exponential backoff
		valueInt, _ = strconv.Atoi(string(response.Payload))
		if expected != valueInt {
			backoffFactor := math.Pow(2, float64(i))
			time.Sleep(time.Millisecond * 50 * time.Duration(backoffFactor))
		} else {
			break
		}
	}
	if expected != valueInt {
		t.Fatalf("Org2 'move funds' transaction result was not propagated to Org1. Expected %d, got: %d",
			(expected), valueInt)
	}

}

func loadOrgPeers(t *testing.T, ctxProvider contextAPI.ClientProvider) {

	ctx, err := ctxProvider()
	if err != nil {
		t.Fatalf("context creation failed: %s", err)
	}

	org1Peers, err := ctx.EndpointConfig().PeersConfig(org1)
	if err != nil {
		t.Fatal(err)
	}

	org2Peers, err := ctx.EndpointConfig().PeersConfig(org2)
	if err != nil {
		t.Fatal(err)
	}

	orgTestPeer0, err = ctx.InfraProvider().CreatePeerFromConfig(&fab.NetworkPeer{PeerConfig: org1Peers[0]})
	if err != nil {
		t.Fatal(err)
	}

	orgTestPeer1, err = ctx.InfraProvider().CreatePeerFromConfig(&fab.NetworkPeer{PeerConfig: org2Peers[0]})
	if err != nil {
		t.Fatal(err)
	}
}

// DynamicSelectionProviderFactory is configured with dynamic (endorser) selection provider
type DynamicSelectionProviderFactory struct {
	defsvc.ProviderFactory
	ChannelUsers []selection.ChannelUser
}

// CreateSelectionProvider returns a new implementation of dynamic selection provider
func (f *DynamicSelectionProviderFactory) CreateSelectionProvider(config fab.EndpointConfig) (fab.SelectionProvider, error) {
	return selection.New(config, f.ChannelUsers)
}
