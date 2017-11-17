/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package resmgmtclient

import (
	"net"
	"testing"
	"time"

	"google.golang.org/grpc"

	fab "github.com/hyperledger/fabric-sdk-go/api/apifabclient"
	resmgmt "github.com/hyperledger/fabric-sdk-go/api/apitxn/resmgmtclient"
	"github.com/pkg/errors"

	"github.com/hyperledger/fabric-sdk-go/pkg/config"
	txnmocks "github.com/hyperledger/fabric-sdk-go/pkg/fabric-txn/mocks"

	"github.com/hyperledger/fabric-sdk-go/pkg/fabric-client/channel"
	fcmocks "github.com/hyperledger/fabric-sdk-go/pkg/fabric-client/mocks"
	"github.com/hyperledger/fabric-sdk-go/pkg/fabric-client/peer"
	pb "github.com/hyperledger/fabric-sdk-go/third_party/github.com/hyperledger/fabric/protos/peer"
)

func TestJoinChannel(t *testing.T) {

	grpcServer := grpc.NewServer()
	defer grpcServer.Stop()

	// Setup mock targets
	endorserServer, addr := startEndorserServer(t, grpcServer)
	time.Sleep(2 * time.Second)

	client := setupTestClient()

	// Create test channel and add it to the client (no added orderer yet)
	channel, _ := channel.NewChannel("mychannel", client)
	client.SetChannel("mychannel", channel)

	// Setup resource management client
	config := getNetworkConfig(t)
	rc := setupResMgmtClient(client, nil, config, t)

	// Setup target peers
	var peers []fab.Peer
	peer, _ := peer.NewPeer(addr, fcmocks.NewMockConfig())
	peers = append(peers, peer)

	// Test fail genesis block retrieval (no orderer)
	err := rc.JoinChannelWithOpts("mychannel", resmgmt.JoinChannelOpts{Targets: peers})
	if err == nil {
		t.Fatal("Should have failed to get genesis block")
	}

	// Create mock orderer with simple mock block
	orderer := fcmocks.NewMockOrderer("", nil)
	orderer.(fcmocks.MockOrderer).EnqueueForSendDeliver(fcmocks.NewSimpleMockBlock())

	// Add orderer to the channel
	err = channel.AddOrderer(orderer)
	if err != nil {
		t.Fatalf("Error adding orderer: %v", err)
	}

	rc = setupResMgmtClient(client, nil, config, t)

	// Test valid join channel request (success)
	err = rc.JoinChannelWithOpts("mychannel", resmgmt.JoinChannelOpts{Targets: peers})
	if err != nil {
		t.Fatal(err)
	}

	orderer.(fcmocks.MockOrderer).EnqueueForSendDeliver(fcmocks.NewSimpleMockBlock())

	// Test fails because configured peer is not running
	err = rc.JoinChannelWithOpts("mychannel", resmgmt.JoinChannelOpts{})
	if err == nil {
		t.Fatal("Should have failed due to configured peer is not running")
	}

	orderer.(fcmocks.MockOrderer).EnqueueForSendDeliver(fcmocks.NewSimpleMockBlock())

	// Test failed proposal error handling
	endorserServer.ProposalError = errors.New("Test Error")

	// Test proposal error
	err = rc.JoinChannelWithOpts("mychannel", resmgmt.JoinChannelOpts{Targets: peers})
	if err == nil {
		t.Fatal("Should have failed with proposal error")
	}

}

func TestNoSigningUserFailure(t *testing.T) {

	// Setup client without user context
	client := fcmocks.NewMockClient()
	config := getNetworkConfig(t)

	discovery, err := setupTestDiscovery(nil, nil)
	if err != nil {
		t.Fatalf("Failed to setup discovery service: %s", err)
	}

	_, err = NewResourceMgmtClient(client, discovery, nil, config)
	if err == nil {
		t.Fatal("Should have failed due to missing signing user")
	}

}

func TestJoinChannelRequiredParameters(t *testing.T) {

	rc := setupDefaultResMgmtClient(t)

	// Test empty channel name
	err := rc.JoinChannel("")
	if err == nil {
		t.Fatalf("Should have failed for empty channel name")
	}

	// Test error when creating channel from configuration
	err = rc.JoinChannel("error")
	if err == nil {
		t.Fatalf("Should have failed with generated error in NewChannel")
	}

}

func TestJoinChannelWithOptsRequiredParameters(t *testing.T) {

	rc := setupDefaultResMgmtClient(t)

	// Test empty channel name for request with opts
	err := rc.JoinChannelWithOpts("", resmgmt.JoinChannelOpts{})
	if err == nil {
		t.Fatalf("Should have failed for empty channel name")
	}

	var peers []fab.Peer
	peer := fcmocks.MockPeer{MockName: "Peer1", MockURL: "http://peer1.com", MockRoles: []string{}, MockCert: nil}
	peers = append(peers, &peer)

	// Test both targets and filter provided (error condition)
	err = rc.JoinChannelWithOpts("mychannel", resmgmt.JoinChannelOpts{Targets: peers, TargetFilter: &MSPFilter{mspID: "MspID"}})
	if err == nil {
		t.Fatalf("Should have failed if both target and filter provided")
	}

}

func TestJoinChannelDiscoveryError(t *testing.T) {

	// Setup test client and config
	client := setupTestClient()
	config := getNetworkConfig(t)

	// Create resource management client with discovery service that will generate an error
	rc := setupResMgmtClient(client, errors.New("Test Error"), config, t)

	err := rc.JoinChannel("mychannel")
	if err == nil {
		t.Fatalf("Should have failed to join channel with discovery error")
	}

	// If targets are not provided discovery service is used
	err = rc.JoinChannelWithOpts("mychannel", resmgmt.JoinChannelOpts{})
	if err == nil {
		t.Fatalf("Should have failed to join channel with discovery error")
	}

}

func TestJoinChannelNoOrdererConfig(t *testing.T) {

	client := setupTestClient()

	// No channel orderer, no global orderer
	noOrdererConfig, err := config.InitConfig("./testdata/noorderer_test.yaml")
	if err != nil {
		t.Fatal(err)
	}
	rc := setupResMgmtClient(client, nil, noOrdererConfig, t)

	err = rc.JoinChannel("mychannel")
	if err == nil {
		t.Fatalf("Should have failed to join channel since no orderer has been configured")
	}

	// Misconfigured channel orderer
	invalidChOrdererConfig, err := config.InitConfig("./testdata/invalidchorderer_test.yaml")
	if err != nil {
		t.Fatal(err)
	}
	rc.config = invalidChOrdererConfig

	err = rc.JoinChannel("mychannel")
	if err == nil {
		t.Fatalf("Should have failed to join channel since channel orderer has been misconfigured")
	}

	// Misconfigured global orderer (cert cannot be loaded)
	invalidOrdererConfig, err := config.InitConfig("./testdata/invalidorderer_test.yaml")
	if err != nil {
		t.Fatal(err)
	}
	rc.config = invalidOrdererConfig

	err = rc.JoinChannel("mychannel")
	if err == nil {
		t.Fatalf("Should have failed to join channel since global orderer certs are not configured properly")
	}

}

func setupTestDiscovery(discErr error, peers []fab.Peer) (fab.DiscoveryService, error) {

	mockDiscovery, err := txnmocks.NewMockDiscoveryProvider(discErr, peers)
	if err != nil {
		return nil, errors.WithMessage(err, "NewMockDiscoveryProvider failed")
	}

	return mockDiscovery.NewDiscoveryService("")
}

func getNetworkConfig(t *testing.T) *config.Config {
	config, err := config.InitConfig("../../../test/fixtures/config/config_test.yaml")
	if err != nil {
		t.Fatal(err)
	}

	return config
}

func setupDefaultResMgmtClient(t *testing.T) *ResourceMgmtClient {
	client := setupTestClient()
	network := getNetworkConfig(t)
	return setupResMgmtClient(client, nil, network, t)
}

func setupResMgmtClient(client *fcmocks.MockClient, discErr error, config *config.Config, t *testing.T) *ResourceMgmtClient {

	discovery, err := setupTestDiscovery(discErr, nil)
	if err != nil {
		t.Fatalf("Failed to setup discovery service: %s", err)
	}

	resClient, err := NewResourceMgmtClient(client, discovery, nil, config)
	if err != nil {
		t.Fatalf("Failed to create new channel management client: %s", err)
	}

	return resClient
}

func setupTestClient() *fcmocks.MockClient {
	client := fcmocks.NewMockClient()
	user := fcmocks.NewMockUser("test")
	cryptoSuite := &fcmocks.MockCryptoSuite{}
	client.SaveUserToStateStore(user, true)
	client.SetUserContext(user)
	client.SetCryptoSuite(cryptoSuite)

	return client
}

func startEndorserServer(t *testing.T, grpcServer *grpc.Server) (*fcmocks.MockEndorserServer, string) {
	lis, err := net.Listen("tcp", "127.0.0.1:0")
	addr := lis.Addr().String()

	endorserServer := &fcmocks.MockEndorserServer{}
	pb.RegisterEndorserServer(grpcServer, endorserServer)
	if err != nil {
		t.Logf("Error starting test server %s", err)
		t.FailNow()
	}
	t.Logf("Starting test server on %s\n", addr)
	go grpcServer.Serve(lis)
	return endorserServer, addr
}
