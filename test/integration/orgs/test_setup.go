/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package orgs

import (
	"testing"
	"time"

	ca "github.com/hyperledger/fabric-sdk-go/api/apifabca"
	fab "github.com/hyperledger/fabric-sdk-go/api/apifabclient"
	"github.com/hyperledger/fabric-sdk-go/api/apitxn"
	"github.com/hyperledger/fabric-sdk-go/internal/github.com/hyperledger/fabric/common/cauthdsl"
	"github.com/hyperledger/fabric-sdk-go/pkg/config"
	client "github.com/hyperledger/fabric-sdk-go/pkg/fabric-client"
	"github.com/hyperledger/fabric-sdk-go/pkg/fabric-client/channel"
	"github.com/hyperledger/fabric-sdk-go/pkg/fabric-client/events"
	"github.com/hyperledger/fabric-sdk-go/pkg/fabric-client/orderer"
	"github.com/hyperledger/fabric-sdk-go/pkg/fabric-client/peer"
	"github.com/hyperledger/fabric-sdk-go/pkg/fabric-txn/admin"
	"github.com/hyperledger/fabric-sdk-go/test/integration"
	"github.com/hyperledger/fabric-sdk-go/third_party/github.com/hyperledger/fabric/bccsp/factory"
)

var org1 = "org1"
var org2 = "org2"

// Client
var orgTestClient fab.FabricClient

// Channel
var orgTestChannel fab.Channel

// Orderers
var orgTestOrderer fab.Orderer

// Peers
var orgTestPeer0 fab.Peer
var orgTestPeer1 fab.Peer

// EventHubs
var peer0EventHub fab.EventHub
var peer1EventHub fab.EventHub

// Users
var org1AdminUser ca.User
var org2AdminUser ca.User
var ordererAdminUser ca.User
var org1User ca.User
var org2User ca.User

// Flag to indicate if test has run before (to skip certain steps)
var foundChannel bool

// initializeFabricClient initializes fabric-sdk-go
func initializeFabricClient(t *testing.T) {
	// Initialize configuration
	configImpl, err := config.InitConfig("../../fixtures/config/config_test.yaml")
	if err != nil {
		t.Fatal(err)
	}

	// Instantiate client
	orgTestClient = client.NewClient(configImpl)

	// Initialize crypto suite
	err = factory.InitFactories(configImpl.CSPConfig())
	if err != nil {
		t.Fatal(err)
	}
	cryptoSuite := factory.GetDefault()
	orgTestClient.SetCryptoSuite(cryptoSuite)
}

func createTestChannel(t *testing.T) {
	var err error

	orgTestChannel, err = channel.NewChannel("orgchannel", orgTestClient)
	if err != nil {
		t.Fatal(err)
	}

	orgTestChannel.AddPeer(orgTestPeer0)
	orgTestChannel.AddPeer(orgTestPeer1)
	orgTestChannel.SetPrimaryPeer(orgTestPeer0)

	orgTestChannel.AddOrderer(orgTestOrderer)

	foundChannel, err = integration.HasPrimaryPeerJoinedChannel(orgTestClient, org1User, orgTestChannel)
	if err != nil {
		t.Fatal(err)
	}

	if foundChannel {
		return
	}

	err = admin.CreateOrUpdateChannel(orgTestClient, ordererAdminUser, org1AdminUser,
		orgTestChannel, "../../fixtures/channel/orgchannel.tx")
	if err != nil {
		t.Fatal(err)
	}
	// Allow orderer to process channel creation
	time.Sleep(time.Second * 3)
}

func joinTestChannel(t *testing.T) {
	if foundChannel {
		return
	}

	// Get peer0 to join channel
	orgTestChannel.RemovePeer(orgTestPeer1)
	err := admin.JoinChannel(orgTestClient, org1AdminUser, orgTestChannel)
	if err != nil {
		t.Fatal(err)
	}

	// Get peer1 to join channel
	orgTestChannel.RemovePeer(orgTestPeer0)
	orgTestChannel.AddPeer(orgTestPeer1)
	orgTestChannel.SetPrimaryPeer(orgTestPeer1)
	err = admin.JoinChannel(orgTestClient, org2AdminUser, orgTestChannel)
	if err != nil {
		t.Fatal(err)
	}
}

func installAndInstantiate(t *testing.T) {
	if foundChannel {
		return
	}

	orgTestClient.SetUserContext(org1AdminUser)
	admin.SendInstallCC(orgTestClient, "exampleCC",
		"github.com/example_cc", "0", nil, []fab.Peer{orgTestPeer0}, "../../fixtures/testdata")

	orgTestClient.SetUserContext(org2AdminUser)
	err := admin.SendInstallCC(orgTestClient, "exampleCC",
		"github.com/example_cc", "0", nil, []fab.Peer{orgTestPeer1}, "../../fixtures/testdata")
	if err != nil {
		t.Fatal(err)
	}

	chaincodePolicy := cauthdsl.SignedByAnyMember([]string{
		org1AdminUser.MspID(), org2AdminUser.MspID()})

	err = admin.SendInstantiateCC(orgTestChannel, "exampleCC",
		generateInitArgs(), "github.com/example_cc", "0", chaincodePolicy, []apitxn.ProposalProcessor{orgTestPeer1}, peer1EventHub)
	if err != nil {
		t.Fatal(err)
	}
}

func loadOrderer(t *testing.T) {
	ordererConfig, err := orgTestClient.Config().RandomOrdererConfig()
	if err != nil {
		t.Fatal(err)
	}
	serverHostOverride := ""
	if str, ok := ordererConfig.GrpcOptions["ssl-target-name-override"].(string); ok {
		serverHostOverride = str
	}
	orgTestOrderer, err = orderer.NewOrderer(ordererConfig.URL, ordererConfig.TlsCACerts.Path, serverHostOverride, orgTestClient.Config())
	if err != nil {
		t.Fatal(err)
	}
}

func loadOrgPeers(t *testing.T) {
	org1Peers, err := orgTestClient.Config().PeersConfig(org1)
	if err != nil {
		t.Fatal(err)
	}

	org2Peers, err := orgTestClient.Config().PeersConfig(org2)
	if err != nil {
		t.Fatal(err)
	}
	serverHostOverrideOrg1 := ""
	if str, ok := org1Peers[0].GrpcOptions["ssl-target-name-override"].(string); ok {
		serverHostOverrideOrg1 = str
	}
	orgTestPeer0, err = peer.NewPeerTLSFromCert(org1Peers[0].Url, org1Peers[0].TlsCACerts.Path, serverHostOverrideOrg1, orgTestClient.Config())
	if err != nil {
		t.Fatal(err)
	}
	serverHostOverrideOrg2 := ""
	if str, ok := org2Peers[0].GrpcOptions["ssl-target-name-override"].(string); ok {
		serverHostOverrideOrg2 = str
	}
	orgTestPeer1, err = peer.NewPeerTLSFromCert(org2Peers[0].Url, org2Peers[0].TlsCACerts.Path, serverHostOverrideOrg2, orgTestClient.Config())
	if err != nil {
		t.Fatal(err)
	}

	peer0EventHub, err = events.NewEventHub(orgTestClient)
	if err != nil {
		t.Fatal(err)
	}

	peer0EventHub.SetPeerAddr(org1Peers[0].EventUrl, org1Peers[0].TlsCACerts.Path, serverHostOverrideOrg1)

	orgTestClient.SetUserContext(org1User)
	err = peer0EventHub.Connect()
	if err != nil {
		t.Fatal(err)
	}

	peer1EventHub, err = events.NewEventHub(orgTestClient)
	if err != nil {
		t.Fatal(err)
	}

	peer1EventHub.SetPeerAddr(org2Peers[0].EventUrl, org2Peers[0].TlsCACerts.Path, serverHostOverrideOrg2)

	orgTestClient.SetUserContext(org2User)
	err = peer1EventHub.Connect()
	if err != nil {
		t.Fatal(err)
	}
}

// loadOrgUsers Loads all the users required to perform this test
func loadOrgUsers(t *testing.T) {
	var err error

	ordererAdminUser, err = integration.GetOrdererAdmin(orgTestClient, org1)
	if err != nil {
		t.Fatal(err)
	}
	org1AdminUser, err = integration.GetAdmin(orgTestClient, "org1", org1)
	if err != nil {
		t.Fatal(err)
	}
	org2AdminUser, err = integration.GetAdmin(orgTestClient, "org2", org2)
	if err != nil {
		t.Fatal(err)
	}
	org1User, err = integration.GetUser(orgTestClient, "org1", org1)
	if err != nil {
		t.Fatal(err)
	}
	org2User, err = integration.GetUser(orgTestClient, "org2", org2)
	if err != nil {
		t.Fatal(err)
	}
}

func generateInitArgs() []string {
	var args []string
	args = append(args, "init")
	args = append(args, "a")
	args = append(args, "100")
	args = append(args, "b")
	args = append(args, "200")
	return args
}
