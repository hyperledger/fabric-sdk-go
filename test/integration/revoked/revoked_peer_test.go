/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package revoked

import (
	"path"
	"strings"
	"testing"

	"github.com/hyperledger/fabric-sdk-go/pkg/common/errors/retry"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/errors/status"
	contextAPI "github.com/hyperledger/fabric-sdk-go/pkg/common/providers/context"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/fab"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/msp"

	packager "github.com/hyperledger/fabric-sdk-go/pkg/fab/ccpackager/gopackager"
	"github.com/hyperledger/fabric-sdk-go/pkg/fabsdk"

	mspclient "github.com/hyperledger/fabric-sdk-go/pkg/client/msp"
	"github.com/hyperledger/fabric-sdk-go/pkg/client/resmgmt"
	"github.com/hyperledger/fabric-sdk-go/test/integration"
	"github.com/hyperledger/fabric-sdk-go/test/metadata"

	"os"

	"runtime"

	"github.com/hyperledger/fabric-sdk-go/pkg/client/channel"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/core"
	"github.com/hyperledger/fabric-sdk-go/pkg/core/config"
	"github.com/hyperledger/fabric-sdk-go/pkg/core/config/lookup"
	"github.com/hyperledger/fabric-sdk-go/pkg/core/mocks"
	"github.com/hyperledger/fabric-sdk-go/third_party/github.com/hyperledger/fabric/common/cauthdsl"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	org1             = "Org1"
	org2             = "Org2"
	ordererAdminUser = "Admin"
	ordererOrgName   = "ordererorg"
	org1AdminUser    = "Admin"
	org2AdminUser    = "Admin"
	org1User         = "User1"
	channelID        = "orgchannel"
	configPath       = "../../fixtures/config/config_test.yaml"
)

// SDK
var sdk *fabsdk.FabricSDK

// Org MSP clients
var org1MspClient *mspclient.Client
var org2MspClient *mspclient.Client

// Peers used for testing
var orgTestPeer0 fab.Peer
var orgTestPeer1 fab.Peer

func TestMain(m *testing.M) {
	err := setup()
	defer teardown()
	var r int
	if err == nil {
		r = m.Run()
	}
	defer os.Exit(r)
	runtime.Goexit()
}

func setup() error {
	// Create SDK setup for the integration tests
	var err error
	sdk, err = fabsdk.New(getConfigBackend())
	if err != nil {
		return errors.Wrap(err, "Failed to create new SDK")
	}

	org1MspClient, err = mspclient.New(sdk.Context(), mspclient.WithOrg(org1))
	if err != nil {
		return errors.Wrap(err, "failed to create org1MspClient, err")
	}

	org2MspClient, err = mspclient.New(sdk.Context(), mspclient.WithOrg(org2))
	if err != nil {
		return errors.Wrap(err, "failed to create org2MspClient, err")
	}

	return nil
}

func teardown() {
	if sdk != nil {
		sdk.Close()
	}
}

// TestRevokedPeer
func TestRevokedPeer(t *testing.T) {

	// Delete all private keys from the crypto suite store
	// and users from the user store at the end
	integration.CleanupUserData(t, sdk)
	defer integration.CleanupUserData(t, sdk)

	//prepare contexts
	ordererClientContext := sdk.Context(fabsdk.WithUser(ordererAdminUser), fabsdk.WithOrg(ordererOrgName))
	org1AdminClientContext := sdk.Context(fabsdk.WithUser(org1AdminUser), fabsdk.WithOrg(org1))
	org2AdminClientContext := sdk.Context(fabsdk.WithUser(org2AdminUser), fabsdk.WithOrg(org2))
	org1ChannelClientContext := sdk.ChannelContext(channelID, fabsdk.WithUser(org1User), fabsdk.WithOrg(org1))

	// Channel management client is responsible for managing channels (create/update channel)
	chMgmtClient, err := resmgmt.New(ordererClientContext)
	if err != nil {
		t.Fatal(err)
	}

	// Get signing identity that is used to sign create channel request
	org1AdminUser, err := org1MspClient.GetSigningIdentity(org1AdminUser)
	if err != nil {
		t.Fatalf("failed to get org1AdminUser, err : %s", err)
	}

	org2AdminUser, err := org2MspClient.GetSigningIdentity(org2AdminUser)
	if err != nil {
		t.Fatalf("failed to get org2AdminUser, err : %s", err)
	}

	createChannel(org1AdminUser, org2AdminUser, chMgmtClient, t)

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
	_, err = chClientOrg1User.Query(channel.Request{ChaincodeID: "exampleCC", Fcn: "invoke", Args: integration.ExampleCCQueryArgs()})
	if err == nil {
		t.Fatal("Expected error: '....Description: could not find chaincode with name 'exampleCC',,, ")
	}
}

func createCC(t *testing.T, org1ResMgmt *resmgmt.Client, org2ResMgmt *resmgmt.Client) {
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
	)
	require.Errorf(t, err, "Expecting error instantiating CC on peer with revoked certificate")
	stat, ok := status.FromError(err)
	require.Truef(t, ok, "Expecting error to be a status error")
	require.Equalf(t, stat.Code, int32(status.SignatureVerificationFailed), "Expecting signature verification error due to revoked cert")
	require.Truef(t, strings.Contains(err.Error(), "the creator certificate is not valid"), "Expecting error message to contain 'the creator certificate is not valid'")
}

func createChannel(org1AdminUser msp.SigningIdentity, org2AdminUser msp.SigningIdentity, chMgmtClient *resmgmt.Client, t *testing.T) {
	req := resmgmt.SaveChannelRequest{ChannelID: "orgchannel",
		ChannelConfigPath: path.Join("../../../", metadata.ChannelConfigPath, "orgchannel.tx"),
		SigningIdentities: []msp.SigningIdentity{org1AdminUser, org2AdminUser}}
	txID, err := chMgmtClient.SaveChannel(req, resmgmt.WithRetry(retry.DefaultResMgmtOpts), resmgmt.WithOrdererEndpoint("orderer.example.com"))
	require.Nil(t, err, "error should be nil")
	require.NotEmpty(t, txID, "transaction ID should be populated")
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

func getConfigBackend() core.ConfigProvider {

	return func() ([]core.ConfigBackend, error) {
		configBackends, err := config.FromFile(configPath)()
		if err != nil {
			return nil, errors.Wrap(err, "failed to read config backend from file, %v")
		}
		backendMap := make(map[string]interface{})

		networkConfig := fab.NetworkConfig{}
		//get valid peer config
		err = lookup.New(configBackends...).UnmarshalKey("peers", &networkConfig.Peers)
		if err != nil {
			return nil, errors.Wrap(err, "failed to unmarshal peer network config, %v")
		}

		//customize peer0.org2 to peer1.org2
		peer2 := networkConfig.Peers["peer0.org2.example.com"]
		peer2.URL = "peer1.org2.example.com:9051"
		peer2.EventURL = ""
		peer2.GRPCOptions["ssl-target-name-override"] = "peer1.org2.example.com"

		//remove peer0.org2
		delete(networkConfig.Peers, "peer0.org2.example.com")

		//add peer1.org2
		networkConfig.Peers["peer1.org2.example.com"] = peer2

		//get valid org2
		err = lookup.New(configBackends...).UnmarshalKey("organizations", &networkConfig.Organizations)
		if err != nil {
			return nil, errors.Wrap(err, "failed to unmarshal organizations network config, %v")
		}

		//Customize org2
		org2 := networkConfig.Organizations["org2"]
		org2.Peers = []string{"peer1.org2.example.com"}
		org2.MSPID = "Org2MSP"
		networkConfig.Organizations["org2"] = org2

		//custom channel
		err = lookup.New(configBackends...).UnmarshalKey("channels", &networkConfig.Channels)
		if err != nil {
			return nil, errors.Wrap(err, "failed to unmarshal entityMatchers network config, %v")
		}

		orgChannel := networkConfig.Channels[channelID]
		delete(orgChannel.Peers, "peer0.org2.example.com")
		orgChannel.Peers["peer1.org2.example.com"] = fab.PeerChannelConfig{
			EndorsingPeer:  true,
			ChaincodeQuery: true,
			LedgerQuery:    true,
			EventSource:    false,
		}
		networkConfig.Channels[channelID] = orgChannel

		//Customize backend with update peers, organizations, channels and entity matchers config
		backendMap["peers"] = networkConfig.Peers
		backendMap["organizations"] = networkConfig.Organizations
		backendMap["channels"] = networkConfig.Channels

		backends := append([]core.ConfigBackend{}, &mocks.MockConfigBackend{KeyValueMap: backendMap})
		return append(backends, configBackends...), nil
	}
}
