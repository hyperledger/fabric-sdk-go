/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package orgs

import (
	"math"
	"path"
	"strconv"
	"testing"
	"time"

	"github.com/hyperledger/fabric-sdk-go/pkg/config"
	packager "github.com/hyperledger/fabric-sdk-go/pkg/fabric-client/ccpackager/gopackager"
	"github.com/hyperledger/fabric-sdk-go/pkg/fabric-client/peer"
	"github.com/hyperledger/fabric-sdk-go/pkg/fabsdk"
	"github.com/pkg/errors"

	"github.com/hyperledger/fabric-sdk-go/api/apiconfig"
	fab "github.com/hyperledger/fabric-sdk-go/api/apifabclient"
	"github.com/hyperledger/fabric-sdk-go/api/apitxn/chclient"
	chmgmt "github.com/hyperledger/fabric-sdk-go/api/apitxn/chmgmtclient"
	resmgmt "github.com/hyperledger/fabric-sdk-go/api/apitxn/resmgmtclient"

	"github.com/hyperledger/fabric-sdk-go/pkg/fabsdk/factory/defsvc"
	"github.com/hyperledger/fabric-sdk-go/test/integration"
	"github.com/hyperledger/fabric-sdk-go/test/metadata"

	selection "github.com/hyperledger/fabric-sdk-go/pkg/fabric-txn/selection/dynamicselection"

	"github.com/hyperledger/fabric-sdk-go/third_party/github.com/hyperledger/fabric/common/cauthdsl"
)

const (
	pollRetries = 5
	org1        = "Org1"
	org2        = "Org2"
)

// Peers
var orgTestPeer0 fab.Peer
var orgTestPeer1 fab.Peer

// TestOrgsEndToEnd creates a channel with two organisations, installs chaincode
// on each of them, and finally invokes a transaction on an org2 peer and queries
// the result from an org1 peer
func TestOrgsEndToEnd(t *testing.T) {

	// Create SDK setup for the integration tests
	sdk, err := fabsdk.New(config.FromFile("../" + integration.ConfigTestFile))
	if err != nil {
		t.Fatalf("Failed to create new SDK: %s", err)
	}

	// Channel management client is responsible for managing channels (create/update channel)
	chMgmtClient, err := sdk.NewClient(fabsdk.WithUser("Admin"), fabsdk.WithOrg("ordererorg")).ChannelMgmt()
	if err != nil {
		t.Fatal(err)
	}

	// Create channel (or update if it already exists)
	org1AdminUser := loadOrgUser(t, sdk, org1, "Admin")
	req := chmgmt.SaveChannelRequest{ChannelID: "orgchannel", ChannelConfig: path.Join("../../../", metadata.ChannelConfigPath, "orgchannel.tx"), SigningIdentity: org1AdminUser}
	if err = chMgmtClient.SaveChannel(req); err != nil {
		t.Fatal(err)
	}

	// Allow orderer to process channel creation
	time.Sleep(time.Second * 5)

	// Org1 resource management client (Org1 is default org)
	org1ResMgmt, err := sdk.NewClient(fabsdk.WithUser("Admin")).ResourceMgmt()
	if err != nil {
		t.Fatalf("Failed to create new resource management client: %s", err)
	}

	// Org1 peers join channel
	if err = org1ResMgmt.JoinChannel("orgchannel"); err != nil {
		t.Fatalf("Org1 peers failed to JoinChannel: %s", err)
	}

	// Org2 resource management client
	org2ResMgmt, err := sdk.NewClient(fabsdk.WithUser("Admin"), fabsdk.WithOrg(org2)).ResourceMgmt()
	if err != nil {
		t.Fatal(err)
	}

	// Org2 peers join channel
	if err = org2ResMgmt.JoinChannel("orgchannel"); err != nil {
		t.Fatalf("Org2 peers failed to JoinChannel: %s", err)
	}

	// Create chaincode package for example cc
	ccPkg, err := packager.NewCCPackage("github.com/example_cc", "../../fixtures/testdata")
	if err != nil {
		t.Fatal(err)
	}

	installCCReq := resmgmt.InstallCCRequest{Name: "exampleCC", Path: "github.com/example_cc", Version: "0", Package: ccPkg}

	// Install example cc to Org1 peers
	_, err = org1ResMgmt.InstallCC(installCCReq)
	if err != nil {
		t.Fatal(err)
	}

	// Install example cc to Org2 peers
	_, err = org2ResMgmt.InstallCC(installCCReq)
	if err != nil {
		t.Fatal(err)
	}

	// Set up chaincode policy to 'any of two msps'
	ccPolicy := cauthdsl.SignedByAnyMember([]string{"Org1MSP", "Org2MSP"})

	// Org1 resource manager will instantiate 'example_cc' on 'orgchannel'
	err = org1ResMgmt.InstantiateCC("orgchannel", resmgmt.InstantiateCCRequest{Name: "exampleCC", Path: "github.com/example_cc", Version: "0", Args: integration.ExampleCCInitArgs(), Policy: ccPolicy})
	if err != nil {
		t.Fatal(err)
	}

	// Load specific targets for move funds test
	loadOrgPeers(t, sdk)

	// Org1 user connects to 'orgchannel'
	chClientOrg1User, err := sdk.NewClient(fabsdk.WithUser("User1"), fabsdk.WithOrg(org1)).Channel("orgchannel")
	if err != nil {
		t.Fatalf("Failed to create new channel client for Org1 user: %s", err)
	}

	// Org2 user connects to 'orgchannel'
	chClientOrg2User, err := sdk.NewClient(fabsdk.WithUser("User1"), fabsdk.WithOrg(org2)).Channel("orgchannel")
	if err != nil {
		t.Fatalf("Failed to create new channel client for Org2 user: %s", err)
	}

	// Org1 user queries initial value on both peers
	response, err := chClientOrg1User.Query(chclient.Request{ChaincodeID: "exampleCC", Fcn: "invoke", Args: integration.ExampleCCQueryArgs()})
	if err != nil {
		t.Fatalf("Failed to query funds: %s", err)
	}
	initial, _ := strconv.Atoi(string(response.Payload))

	// Org2 user moves funds on org2 peer
	response, err = chClientOrg2User.Execute(chclient.Request{ChaincodeID: "exampleCC", Fcn: "invoke", Args: integration.ExampleCCTxArgs()}, chclient.WithProposalProcessor(orgTestPeer1))
	if err != nil {
		t.Fatalf("Failed to move funds: %s", err)
	}

	// Assert that funds have changed value on org1 peer
	verifyValue(t, chClientOrg1User, initial+1)

	// Start chaincode upgrade process (install and instantiate new version of exampleCC)
	installCCReq = resmgmt.InstallCCRequest{Name: "exampleCC", Path: "github.com/example_cc", Version: "1", Package: ccPkg}

	// Install example cc version '1' to Org1 peers
	_, err = org1ResMgmt.InstallCC(installCCReq)
	if err != nil {
		t.Fatal(err)
	}

	// Install example cc version '1' to Org2 peers
	_, err = org2ResMgmt.InstallCC(installCCReq)
	if err != nil {
		t.Fatal(err)
	}

	// New chaincode policy (both orgs have to approve)
	org1Andorg2Policy, err := cauthdsl.FromString("AND ('Org1MSP.member','Org2MSP.member')")
	if err != nil {
		t.Fatal(err)
	}

	// Org1 resource manager will instantiate 'example_cc' version 1 on 'orgchannel'
	err = org1ResMgmt.UpgradeCC("orgchannel", resmgmt.UpgradeCCRequest{Name: "exampleCC", Path: "github.com/example_cc", Version: "1", Args: integration.ExampleCCUpgradeArgs(), Policy: org1Andorg2Policy})
	if err != nil {
		t.Fatal(err)
	}

	// Org2 user moves funds on org2 peer (cc policy fails since both Org1 and Org2 peers should participate)
	response, err = chClientOrg2User.Execute(chclient.Request{ChaincodeID: "exampleCC", Fcn: "invoke", Args: integration.ExampleCCTxArgs()}, chclient.WithProposalProcessor(orgTestPeer1))
	if err == nil {
		t.Fatalf("Should have failed to move funds due to cc policy")
	}

	// Org2 user moves funds (cc policy ok since we have provided peers for both Orgs)
	response, err = chClientOrg2User.Execute(chclient.Request{ChaincodeID: "exampleCC", Fcn: "invoke", Args: integration.ExampleCCTxArgs()}, chclient.WithProposalProcessor(orgTestPeer0, orgTestPeer1))
	if err != nil {
		t.Fatalf("Failed to move funds: %s", err)
	}

	// Assert that funds have changed value on org1 peer
	beforeTxValue, _ := strconv.Atoi(integration.ExampleCCUpgradeB)
	expectedValue := beforeTxValue + 1
	verifyValue(t, chClientOrg1User, expectedValue)

	// Specify user that will be used by dynamic selection service (to retrieve chanincode policy information)
	// This user has to have privileges to query lscc for chaincode data
	mychannelUser := selection.ChannelUser{ChannelID: "orgchannel", UserName: "User1", OrgName: "Org1"}

	// Create SDK setup for channel client with dynamic selection
	sdk, err = fabsdk.New(config.FromFile("../"+integration.ConfigTestFile),
		fabsdk.WithServicePkg(&DynamicSelectionProviderFactory{ChannelUsers: []selection.ChannelUser{mychannelUser}}))
	if err != nil {
		t.Fatalf("Failed to create new SDK: %s", err)
	}

	// Create new client that will use dynamic selection
	chClientOrg2User, err = sdk.NewClient(fabsdk.WithUser("User1"), fabsdk.WithOrg(org2)).Channel("orgchannel")
	if err != nil {
		t.Fatalf("Failed to create new channel client for Org2 user: %s", err)
	}

	// Org2 user moves funds (dynamic selection will inspect chaincode policy to determine endorsers)
	response, err = chClientOrg2User.Execute(chclient.Request{ChaincodeID: "exampleCC", Fcn: "invoke", Args: integration.ExampleCCTxArgs()})
	if err != nil {
		t.Fatalf("Failed to move funds: %s", err)
	}

	expectedValue++
	verifyValue(t, chClientOrg1User, expectedValue)

}

func verifyValue(t *testing.T, chClient chclient.ChannelClient, expected int) {

	// Assert that funds have changed value on org1 peer
	var valueInt int
	for i := 0; i < pollRetries; i++ {
		// Query final value on org1 peer
		response, err := chClient.Query(chclient.Request{ChaincodeID: "exampleCC", Fcn: "invoke", Args: integration.ExampleCCQueryArgs()}, chclient.WithProposalProcessor(orgTestPeer0))
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

func loadOrgUser(t *testing.T, sdk *fabsdk.FabricSDK, orgName string, userName string) fab.IdentityContext {

	session, err := sdk.NewClient(fabsdk.WithUser(userName), fabsdk.WithOrg(orgName)).Session()
	if err != nil {
		t.Fatal(errors.Wrapf(err, "Session failed, %s, %s", orgName, userName))
	}
	return session
}

func loadOrgPeers(t *testing.T, sdk *fabsdk.FabricSDK) {

	org1Peers, err := sdk.Config().PeersConfig(org1)
	if err != nil {
		t.Fatal(err)
	}

	org2Peers, err := sdk.Config().PeersConfig(org2)
	if err != nil {
		t.Fatal(err)
	}

	orgTestPeer0, err = peer.New(sdk.Config(), peer.FromPeerConfig(&apiconfig.NetworkPeer{PeerConfig: org1Peers[0]}))
	if err != nil {
		t.Fatal(err)
	}

	orgTestPeer1, err = peer.New(sdk.Config(), peer.FromPeerConfig(&apiconfig.NetworkPeer{PeerConfig: org2Peers[0]}))
	if err != nil {
		t.Fatal(err)
	}
}

// DynamicSelectionProviderFactory is configured with dynamic (endorser) selection provider
type DynamicSelectionProviderFactory struct {
	defsvc.ProviderFactory
	ChannelUsers []selection.ChannelUser
}

// NewSelectionProvider returns a new implementation of dynamic selection provider
func (f *DynamicSelectionProviderFactory) NewSelectionProvider(config apiconfig.Config) (fab.SelectionProvider, error) {
	return selection.NewSelectionProvider(config, f.ChannelUsers, nil)
}
