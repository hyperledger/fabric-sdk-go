/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package orgs

import (
	"fmt"
	"testing"
	"time"

	ca "github.com/hyperledger/fabric-sdk-go/api/apifabca"
	fab "github.com/hyperledger/fabric-sdk-go/api/apifabclient"
	"github.com/hyperledger/fabric-sdk-go/api/apitxn"
	"github.com/hyperledger/fabric-sdk-go/pkg/config"
	client "github.com/hyperledger/fabric-sdk-go/pkg/fabric-client"
	"github.com/hyperledger/fabric-sdk-go/pkg/fabric-client/channel"
	"github.com/hyperledger/fabric-sdk-go/pkg/fabric-client/events"
	"github.com/hyperledger/fabric-sdk-go/pkg/fabric-client/orderer"
	"github.com/hyperledger/fabric-sdk-go/pkg/fabric-client/peer"
	"github.com/hyperledger/fabric-sdk-go/pkg/fabric-txn/admin"
	"github.com/hyperledger/fabric-sdk-go/test/integration"
	"github.com/hyperledger/fabric/bccsp/factory"
)

var org1 = "peerorg1"
var org2 = "peerorg2"

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
	failTestIfError(err, t)

	// Instantiate client
	orgTestClient = client.NewClient(configImpl)

	// Initialize crypto suite
	err = factory.InitFactories(configImpl.CSPConfig())
	failTestIfError(err, t)
	cryptoSuite := factory.GetDefault()
	orgTestClient.SetCryptoSuite(cryptoSuite)
}

func createTestChannel(t *testing.T) {
	var err error

	orgTestChannel, err = channel.NewChannel("orgchannel", orgTestClient)
	failTestIfError(err, t)

	orgTestChannel.AddPeer(orgTestPeer0)
	orgTestChannel.AddPeer(orgTestPeer1)
	orgTestChannel.SetPrimaryPeer(orgTestPeer0)

	orgTestChannel.AddOrderer(orgTestOrderer)

	foundChannel, err = integration.HasPrimaryPeerJoinedChannel(orgTestClient, org1User, orgTestChannel)
	failTestIfError(err, t)

	if foundChannel {
		return
	}

	err = admin.CreateOrUpdateChannel(orgTestClient, ordererAdminUser, org1AdminUser,
		orgTestChannel, "../../fixtures/channel/orgchannel.tx")
	failTestIfError(err, t)
	// Allow orderer to process channel creation
	time.Sleep(time.Millisecond * 500)
}

func joinTestChannel(t *testing.T) {
	if foundChannel {
		return
	}

	// Get peer0 to join channel
	orgTestChannel.RemovePeer(orgTestPeer1)
	err := admin.JoinChannel(orgTestClient, org1AdminUser, orgTestChannel)
	failTestIfError(err, t)

	// Get peer1 to join channel
	orgTestChannel.RemovePeer(orgTestPeer0)
	orgTestChannel.AddPeer(orgTestPeer1)
	orgTestChannel.SetPrimaryPeer(orgTestPeer1)
	err = admin.JoinChannel(orgTestClient, org2AdminUser, orgTestChannel)
	failTestIfError(err, t)
}

func installAndInstantiate(t *testing.T) {
	if foundChannel {
		return
	}

	orgTestClient.SetUserContext(org1AdminUser)
	admin.SendInstallCC(orgTestClient, "exampleCC",
		"github.com/example_cc", "0", nil, []fab.Peer{orgTestPeer0}, "../../fixtures")

	orgTestClient.SetUserContext(org2AdminUser)
	err := admin.SendInstallCC(orgTestClient, "exampleCC",
		"github.com/example_cc", "0", nil, []fab.Peer{orgTestPeer1}, "../../fixtures")
	failTestIfError(err, t)

	err = admin.SendInstantiateCC(orgTestChannel, "exampleCC",
		generateInitArgs(), "github.com/example_cc", "0", []apitxn.ProposalProcessor{orgTestPeer1}, peer1EventHub)
	failTestIfError(err, t)
}

func loadOrderer(t *testing.T) {
	ordererConfig, err := orgTestClient.Config().RandomOrdererConfig()
	failTestIfError(err, t)

	orgTestOrderer, err = orderer.NewOrderer(fmt.Sprintf("%s:%d", ordererConfig.Host,
		ordererConfig.Port), ordererConfig.TLS.Certificate,
		ordererConfig.TLS.ServerHostOverride, orgTestClient.Config())
	failTestIfError(err, t)
}

func loadOrgPeers(t *testing.T) {
	org1Peers, err := orgTestClient.Config().PeersConfig(org1)
	failTestIfError(err, t)

	org2Peers, err := orgTestClient.Config().PeersConfig(org2)
	failTestIfError(err, t)

	orgTestPeer0, err = peer.NewPeerTLSFromCert(fmt.Sprintf("%s:%d", org1Peers[0].Host,
		org1Peers[0].Port), org1Peers[0].TLS.Certificate,
		org1Peers[0].TLS.ServerHostOverride, orgTestClient.Config())
	failTestIfError(err, t)

	orgTestPeer1, err = peer.NewPeerTLSFromCert(fmt.Sprintf("%s:%d", org2Peers[0].Host,
		org2Peers[0].Port), org2Peers[0].TLS.Certificate,
		org2Peers[0].TLS.ServerHostOverride, orgTestClient.Config())
	failTestIfError(err, t)

	peer0EventHub, err = events.NewEventHub(orgTestClient)
	failTestIfError(err, t)

	peer0EventHub.SetPeerAddr(fmt.Sprintf("%s:%d", org1Peers[0].EventHost,
		org1Peers[0].EventPort), org1Peers[0].TLS.Certificate,
		org1Peers[0].TLS.ServerHostOverride)

	orgTestClient.SetUserContext(org1User)
	err = peer0EventHub.Connect()
	failTestIfError(err, t)

	peer1EventHub, err = events.NewEventHub(orgTestClient)
	failTestIfError(err, t)

	peer1EventHub.SetPeerAddr(fmt.Sprintf("%s:%d", org2Peers[0].EventHost,
		org2Peers[0].EventPort), org2Peers[0].TLS.Certificate,
		org2Peers[0].TLS.ServerHostOverride)

	orgTestClient.SetUserContext(org2User)
	err = peer1EventHub.Connect()
	failTestIfError(err, t)
}

// loadOrgUsers Loads all the users required to perform this test
func loadOrgUsers(t *testing.T) {
	var err error

	ordererAdminUser, err = integration.GetOrdererAdmin(orgTestClient, org1)
	failTestIfError(err, t)
	org1AdminUser, err = integration.GetAdmin(orgTestClient, "org1", org1)
	failTestIfError(err, t)
	org2AdminUser, err = integration.GetAdmin(orgTestClient, "org2", org2)
	failTestIfError(err, t)
	org1User, err = integration.GetUser(orgTestClient, "org1", org1)
	failTestIfError(err, t)
	org2User, err = integration.GetUser(orgTestClient, "org2", org2)
	failTestIfError(err, t)
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

func failTestIfError(err error, t *testing.T) {
	if err != nil {
		t.Fatal(err)
	}
}
