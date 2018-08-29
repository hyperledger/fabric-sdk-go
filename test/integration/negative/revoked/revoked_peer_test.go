// +build disabled

/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package revoked

import (
	"strings"
	"testing"

	"github.com/hyperledger/fabric-sdk-go/pkg/common/errors/retry"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/errors/status"
	contextAPI "github.com/hyperledger/fabric-sdk-go/pkg/common/providers/context"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/fab"

	packager "github.com/hyperledger/fabric-sdk-go/pkg/fab/ccpackager/gopackager"
	"github.com/hyperledger/fabric-sdk-go/pkg/fabsdk"

	"github.com/hyperledger/fabric-sdk-go/pkg/client/resmgmt"
	"github.com/hyperledger/fabric-sdk-go/test/integration"

	"github.com/hyperledger/fabric-sdk-go/pkg/client/channel"
	"github.com/hyperledger/fabric-sdk-go/third_party/github.com/hyperledger/fabric/common/cauthdsl"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	org1AdminUser = "Admin"
	org2AdminUser = "Admin"
	org1User      = "User1"
)

// Peers used for testing
var orgTestPeer0 fab.Peer
var orgTestPeer1 fab.Peer

// TestRevokedPeer
func TestRevokedPeer(t *testing.T) {

	// Delete all private keys from the crypto suite store
	// and users from the user store at the end
	integration.CleanupUserData(t, sdk)
	defer integration.CleanupUserData(t, sdk)

	//prepare contexts
	org1AdminClientContext := sdk.Context(fabsdk.WithUser(org1AdminUser), fabsdk.WithOrg(org1))
	org2AdminClientContext := sdk.Context(fabsdk.WithUser(org2AdminUser), fabsdk.WithOrg(org2))
	org1ChannelClientContext := sdk.ChannelContext(channelID, fabsdk.WithUser(org1User), fabsdk.WithOrg(org1))

	// Org1 resource management client (Org1 is default org)
	org1ResMgmt, err := resmgmt.New(org1AdminClientContext)
	if err != nil {
		t.Fatalf("Failed to create new resource management client: %s", err)
	}

	// Org1 peers join channel
	if err = org1ResMgmt.JoinChannel("orgchannel", resmgmt.WithRetry(retry.DefaultResMgmtOpts), resmgmt.WithOrdererEndpoint("orderer.example.com")); err != nil {
		t.Fatalf("Org1 peers failed to JoinChannel: %s", err)
	}

	// Org2 resource management client
	org2ResMgmt, err := resmgmt.New(org2AdminClientContext)
	if err != nil {
		t.Fatal(err)
	}

	// Org2 peers join channel
	if err = org2ResMgmt.JoinChannel("orgchannel", resmgmt.WithRetry(retry.DefaultResMgmtOpts), resmgmt.WithOrdererEndpoint("orderer.example.com")); err != nil {
		t.Fatalf("Org2 peers failed to JoinChannel: %s", err)
	}

	// Create chaincode package for example cc
	createCC(t, org1ResMgmt, org2ResMgmt)

	// Load specific targets for move funds test - one of the
	//targets has its certificate revoked
	loadOrgPeers(t, org1AdminClientContext)

	queryCC(org1ChannelClientContext, t)

}

func queryCC(org1ChannelClientContext contextAPI.ChannelProvider, t *testing.T) {
	// Org1 user connects to 'orgchannel'
	chClientOrg1User, err := channel.New(org1ChannelClientContext)
	if err != nil {
		t.Fatalf("Failed to create new channel client for Org1 user: %s", err)
	}
	// Org1 user queries initial value on both peers
	// Since one of the peers on channel has certificate revoked, eror is expected here
	// Error in container is :
	// .... identity 0 does not satisfy principal:
	// Could not validate identity against certification chain, err The certificate has been revoked
	_, err = chClientOrg1User.Query(channel.Request{ChaincodeID: "exampleCC", Fcn: "invoke", Args: integration.ExampleCCDefaultQueryArgs()},
		channel.WithRetry(retry.DefaultChannelOpts))
	if err == nil {
		t.Fatal("Expected error: '....Description: could not find chaincode with name 'exampleCC',,, ")
	}
}

func createCC(t *testing.T, org1ResMgmt *resmgmt.Client, org2ResMgmt *resmgmt.Client) {
	ccPkg, err := packager.NewCCPackage("github.com/example_cc", integration.GetDeployPath())
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
	_, err = org2ResMgmt.InstallCC(installCCReq, resmgmt.WithRetry(retry.DefaultResMgmtOpts))
	if err != nil {
		t.Fatal(err)
	}
	// Set up chaincode policy to 'two-of-two msps'
	ccPolicy, err := cauthdsl.FromString("AND ('Org1MSP.member','Org2MSP.member')")
	require.NoErrorf(t, err, "Error creating cc policy with both orgs to approve")
	// Org1 resource manager will instantiate 'example_cc' on 'orgchannel'
	_, err = org1ResMgmt.InstantiateCC(
		"orgchannel",
		resmgmt.InstantiateCCRequest{
			Name:    "exampleCC",
			Path:    "github.com/example_cc",
			Version: "0",
			Args:    integration.ExampleCCInitArgs(),
			Policy:  ccPolicy,
		},
		resmgmt.WithTargetEndpoints("peer1.org2.example.com"),
	)
	require.Errorf(t, err, "Expecting error instantiating CC on peer with revoked certificate")
	stat, ok := status.FromError(err)
	require.Truef(t, ok, "Expecting error to be a status error")
	require.Equalf(t, stat.Code, int32(status.SignatureVerificationFailed), "Expecting signature verification error due to revoked cert")
	require.Truef(t, strings.Contains(err.Error(), "the creator certificate is not valid"), "Expecting error message to contain 'the creator certificate is not valid'")
}

func loadOrgPeers(t *testing.T, ctxProvider contextAPI.ClientProvider) {

	ctx, err := ctxProvider()
	if err != nil {
		t.Fatalf("context creation failed: %s", err)
	}

	org1Peers, ok := ctx.EndpointConfig().PeersConfig(org1)
	assert.True(t, ok)

	org2Peers, ok := ctx.EndpointConfig().PeersConfig(org2)
	assert.True(t, ok)

	orgTestPeer0, err = ctx.InfraProvider().CreatePeerFromConfig(&fab.NetworkPeer{PeerConfig: org1Peers[0]})
	if err != nil {
		t.Fatal(err)
	}

	orgTestPeer1, err = ctx.InfraProvider().CreatePeerFromConfig(&fab.NetworkPeer{PeerConfig: org2Peers[0]})
	if err != nil {
		t.Fatal(err)
	}

}
