/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package orgs

import (
	"math"
	"strconv"
	"testing"
	"time"

	"github.com/hyperledger/fabric-sdk-go/pkg/errors"
	packager "github.com/hyperledger/fabric-sdk-go/pkg/fabric-client/ccpackager/gopackager"
	"github.com/hyperledger/fabric-sdk-go/pkg/fabric-client/peer"

	"github.com/hyperledger/fabric-sdk-go/api/apiconfig"
	fab "github.com/hyperledger/fabric-sdk-go/api/apifabclient"

	"github.com/hyperledger/fabric-sdk-go/api/apitxn"
	chmgmt "github.com/hyperledger/fabric-sdk-go/api/apitxn/chmgmtclient"
	resmgmt "github.com/hyperledger/fabric-sdk-go/api/apitxn/resmgmtclient"
	"github.com/hyperledger/fabric-sdk-go/def/fabapi"
	"github.com/hyperledger/fabric-sdk-go/test/integration"

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
	sdkOptions := fabapi.Options{
		ConfigFile: "../" + integration.ConfigTestFile,
	}

	sdk, err := fabapi.NewSDK(sdkOptions)
	if err != nil {
		t.Fatalf("Failed to create new SDK: %s", err)
	}

	// Channel management client is responsible for managing channels (create/update channel)
	chMgmtClient, err := sdk.NewChannelMgmtClientWithOpts("Admin", &fabapi.ChannelMgmtClientOpts{OrgName: "ordererorg"})
	if err != nil {
		t.Fatal(err)
	}

	// Create channel (or update if it already exists)
	org1AdminUser := loadOrgUser(t, sdk, org1, "Admin")
	req := chmgmt.SaveChannelRequest{ChannelID: "orgchannel", ChannelConfig: "../../fixtures/channel/orgchannel.tx", SigningUser: org1AdminUser}
	if err = chMgmtClient.SaveChannel(req); err != nil {
		t.Fatal(err)
	}

	// Allow orderer to process channel creation
	time.Sleep(time.Second * 3)

	// Org1 resource management client (Org1 is default org)
	org1ResMgmt, err := sdk.NewResourceMgmtClient("Admin")
	if err != nil {
		t.Fatalf("Failed to create new resource management client: %s", err)
	}

	// Org1 peers join channel
	if err = org1ResMgmt.JoinChannel("orgchannel"); err != nil {
		t.Fatalf("Org1 peers failed to JoinChannel: %s", err)
	}

	// Org2 resource management client
	org2ResMgmt, err := sdk.NewResourceMgmtClientWithOpts("Admin", &fabapi.ResourceMgmtClientOpts{OrgName: org2})
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
	chClientOrg1User, err := sdk.NewChannelClientWithOpts("orgchannel", "User1", &fabapi.ChannelClientOpts{OrgName: org1})
	if err != nil {
		t.Fatalf("Failed to create new channel client for Org1 user: %s", err)
	}

	// Org2 user connects to 'orgchannel'
	chClientOrg2User, err := sdk.NewChannelClientWithOpts("orgchannel", "User1", &fabapi.ChannelClientOpts{OrgName: org2})
	if err != nil {
		t.Fatalf("Failed to create new channel client for Org2 user: %s", err)
	}

	// Org1 user queries initial value on both peers
	initialValue, err := chClientOrg1User.Query(apitxn.QueryRequest{ChaincodeID: "exampleCC", Fcn: "invoke", Args: integration.ExampleCCQueryArgs()})
	if err != nil {
		t.Fatalf("Failed to query funds: %s", err)
	}

	// Org2 user moves funds on org2 peer
	txOpts := apitxn.ExecuteTxOpts{ProposalProcessors: []apitxn.ProposalProcessor{orgTestPeer1}}
	_, err = chClientOrg2User.ExecuteTxWithOpts(apitxn.ExecuteTxRequest{ChaincodeID: "exampleCC", Fcn: "invoke", Args: integration.ExampleCCTxArgs()}, txOpts)
	if err != nil {
		t.Fatalf("Failed to move funds: %s", err)
	}

	// Assert that funds have changed value on org1 peer
	initialInt, _ := strconv.Atoi(string(initialValue))
	var finalInt int
	for i := 0; i < pollRetries; i++ {
		// Query final value on org1 peer
		queryOpts := apitxn.QueryOpts{ProposalProcessors: []apitxn.ProposalProcessor{orgTestPeer0}}
		finalValue, err := chClientOrg1User.QueryWithOpts(apitxn.QueryRequest{ChaincodeID: "exampleCC", Fcn: "invoke", Args: integration.ExampleCCQueryArgs()}, queryOpts)
		if err != nil {
			t.Fatalf("Failed to query funds after transaction: %s", err)
		}
		// If value has not propogated sleep with exponential backoff
		finalInt, _ = strconv.Atoi(string(finalValue))
		if initialInt+1 != finalInt {
			backoffFactor := math.Pow(2, float64(i))
			time.Sleep(time.Millisecond * 50 * time.Duration(backoffFactor))
		} else {
			break
		}
	}
	if initialInt+1 != finalInt {
		t.Fatalf("Org2 'move funds' transaction result was not propagated to Org1. Expected %d, got: %d",
			(initialInt + 1), finalInt)
	}

}

func loadOrgUser(t *testing.T, sdk *fabapi.FabricSDK, orgName string, userName string) fab.User {

	user, err := sdk.NewPreEnrolledUser(orgName, userName)
	if err != nil {
		t.Fatal(errors.Wrapf(err, "NewPreEnrolledUser failed, %s, %s", orgName, userName))
	}

	return user
}

func loadOrgPeers(t *testing.T, sdk *fabapi.FabricSDK) {

	org1Peers, err := sdk.ConfigProvider().PeersConfig(org1)
	if err != nil {
		t.Fatal(err)
	}

	org2Peers, err := sdk.ConfigProvider().PeersConfig(org2)
	if err != nil {
		t.Fatal(err)
	}

	orgTestPeer0, err = peer.NewPeerFromConfig(&apiconfig.NetworkPeer{PeerConfig: org1Peers[0]}, sdk.ConfigProvider())
	if err != nil {
		t.Fatal(err)
	}

	orgTestPeer1, err = peer.NewPeerFromConfig(&apiconfig.NetworkPeer{PeerConfig: org2Peers[0]}, sdk.ConfigProvider())
	if err != nil {
		t.Fatal(err)
	}
}
