/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package resmgmtclient

import (
	"fmt"
	"net"
	"strings"
	"testing"
	"time"

	"google.golang.org/grpc"

	fab "github.com/hyperledger/fabric-sdk-go/api/apifabclient"
	resmgmt "github.com/hyperledger/fabric-sdk-go/api/apitxn/resmgmtclient"
	"github.com/hyperledger/fabric-sdk-go/pkg/errors"
	"github.com/hyperledger/fabric-sdk-go/pkg/fabric-client/peer"
	"github.com/hyperledger/fabric-sdk-go/third_party/github.com/hyperledger/fabric/common/cauthdsl"

	"github.com/hyperledger/fabric-sdk-go/pkg/config"
	txnmocks "github.com/hyperledger/fabric-sdk-go/pkg/fabric-txn/mocks"

	"github.com/hyperledger/fabric-sdk-go/pkg/fabric-client/channel"
	fcmocks "github.com/hyperledger/fabric-sdk-go/pkg/fabric-client/mocks"
	pb "github.com/hyperledger/fabric-sdk-go/third_party/github.com/hyperledger/fabric/protos/peer"
)

func TestJoinChannel(t *testing.T) {

	grpcServer := grpc.NewServer()
	defer grpcServer.Stop()

	// Setup mock targets
	endorserServer, addr := startEndorserServer(t, grpcServer)
	time.Sleep(2 * time.Second)

	client := setupTestClient("test", "Org1MSP")

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

	user := fcmocks.NewMockUserWithMSPID("test", "")
	client.SetUserContext(user)

	_, err = NewResourceMgmtClient(client, discovery, nil, config)
	if err == nil {
		t.Fatal("Should have failed due to missing msp")
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

	// Setup test client with different msp (default targets cannot be calculated)
	client := setupTestClient("test", "otherMSP")
	config := getNetworkConfig(t)

	// Create new resource management client ("otherMSP")
	rc = setupResMgmtClient(client, nil, config, t)

	// Test missing default targets
	err = rc.JoinChannel("mychannel")
	if err == nil {
		t.Fatalf("InstallCC should have failed with no default targets error")
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

	// Test missing default targets
	err = rc.JoinChannelWithOpts("mychannel", resmgmt.JoinChannelOpts{TargetFilter: &MSPFilter{mspID: "MspID"}})
	if err == nil {
		t.Fatalf("InstallCC should have failed with no targets error")
	}

}

func TestJoinChannelDiscoveryError(t *testing.T) {

	// Setup test client and config
	client := setupTestClient("test", "Org1MSP")
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

	client := setupTestClient("test", "Org1MSP")

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

func TestIsChaincodeInstalled(t *testing.T) {

	rc := setupDefaultResMgmtClient(t)

	peer := &fcmocks.MockPeer{MockName: "Peer1", MockURL: "http://peer1.com", MockRoles: []string{}, MockCert: nil, MockMSP: "Org1MSP"}

	// Chaincode found request
	req := resmgmt.InstallCCRequest{Name: "name", Version: "version", Path: "path"}

	// Test chaincode installed (valid peer)
	installed, err := rc.isChaincodeInstalled(req, peer)
	if err != nil {
		t.Fatal(err)
	}
	if !installed {
		t.Fatalf("CC should have been installed: %s", req)
	}

	// Chaincode not found request
	req = resmgmt.InstallCCRequest{Name: "ID", Version: "v0", Path: "path"}

	// Test chaincode installed
	installed, err = rc.isChaincodeInstalled(req, peer)
	if err != nil {
		t.Fatal(err)
	}
	if installed {
		t.Fatalf("CC should NOT have been installed: %s", req)
	}

	// Test error retrieving installed cc info (peer is nil)
	_, err = rc.isChaincodeInstalled(req, nil)
	if err == nil {
		t.Fatalf("Should have failed with error in get installed chaincodes")
	}

}

func TestInstallCCWithOpts(t *testing.T) {

	rc := setupDefaultResMgmtClient(t)

	// Setup targets
	var peers []fab.Peer
	peer := fcmocks.MockPeer{MockName: "Peer1", MockURL: "http://peer1.com", MockRoles: []string{}, MockCert: nil, MockMSP: "Org1MSP"}
	peers = append(peers, &peer)

	// Already installed chaincode request
	req := resmgmt.InstallCCRequest{Name: "name", Version: "version", Path: "path", Package: &fab.CCPackage{Type: 1, Code: []byte("code")}}
	responses, err := rc.InstallCCWithOpts(req, resmgmt.InstallCCOpts{Targets: peers})
	if err != nil {
		t.Fatal(err)
	}

	if responses == nil || len(responses) != 1 {
		t.Fatal("Should have one 'already installed' response")
	}

	if !strings.Contains(responses[0].Info, "already installed") {
		t.Fatal("Should have 'already installed' info set")
	}

	if responses[0].Target != peer.MockURL {
		t.Fatalf("Expecting %s target URL, got %s", peer.MockURL, responses[0].Target)
	}

	// Chaincode not found request (it will be installed)
	req = resmgmt.InstallCCRequest{Name: "ID", Version: "v0", Path: "path", Package: &fab.CCPackage{Type: 1, Code: []byte("code")}}
	responses, err = rc.InstallCCWithOpts(req, resmgmt.InstallCCOpts{Targets: peers})
	if err != nil {
		t.Fatal(err)
	}

	if responses[0].Target != peer.MockURL {
		t.Fatal("Wrong target URL set")
	}

	if strings.Contains(responses[0].Info, "already installed") {
		t.Fatal("Should not have 'already installed' info set since it was not previously installed")
	}

	// Chaincode that causes generic (system) error in installed chaincodes
	req = resmgmt.InstallCCRequest{Name: "error", Version: "v0", Path: "path", Package: &fab.CCPackage{Type: 1, Code: []byte("code")}}
	_, err = rc.InstallCCWithOpts(req, resmgmt.InstallCCOpts{Targets: peers})
	if err == nil {
		t.Fatalf("Should have failed since install cc returns an error in the client")
	}

	// Chaincode that causes response error in installed chaincodes
	req = resmgmt.InstallCCRequest{Name: "errorInResponse", Version: "v0", Path: "path", Package: &fab.CCPackage{Type: 1, Code: []byte("code")}}
	responses, err = rc.InstallCCWithOpts(req, resmgmt.InstallCCOpts{Targets: peers})
	if err != nil {
		t.Fatal(err)
	}
}

func TestInstallCC(t *testing.T) {

	rc := setupDefaultResMgmtClient(t)

	// Chaincode that is not installed already (it will be installed)
	req := resmgmt.InstallCCRequest{Name: "ID", Version: "v0", Path: "path", Package: &fab.CCPackage{Type: 1, Code: []byte("code")}}
	responses, err := rc.InstallCC(req)
	if err != nil {
		t.Fatal(err)
	}
	if responses == nil || len(responses) != 1 {
		t.Fatal("Should have one successful response")
	}

	expected := "http://peer1.com"
	if responses[0].Target != expected {
		t.Fatalf("Expecting %s target URL, got %s", expected, responses[0].Target)
	}

	if responses[0].Status != 0 {
		t.Fatalf("Expecting %d status, got %d", 0, responses[0].Status)
	}

	// Chaincode that causes generic (system) error in installed chaincodes
	req = resmgmt.InstallCCRequest{Name: "error", Version: "v0", Path: "path", Package: &fab.CCPackage{Type: 1, Code: []byte("code")}}
	_, err = rc.InstallCC(req)
	if err == nil {
		t.Fatalf("Should have failed since install cc returns an error in the client")
	}
}

func TestInstallCCRequiredParameters(t *testing.T) {

	rc := setupDefaultResMgmtClient(t)

	// Test missing required parameters
	req := resmgmt.InstallCCRequest{}
	_, err := rc.InstallCC(req)
	if err == nil {
		t.Fatalf("Should have failed for empty install cc request")
	}

	// Test missing chaincode ID
	req = resmgmt.InstallCCRequest{Name: "", Version: "v0", Path: "path"}
	_, err = rc.InstallCC(req)
	if err == nil {
		t.Fatalf("Should have failed for empty cc ID")
	}

	// Test missing chaincode version
	req = resmgmt.InstallCCRequest{Name: "ID", Version: "", Path: "path"}
	_, err = rc.InstallCC(req)
	if err == nil {
		t.Fatalf("Should have failed for empty cc version")
	}

	// Test missing chaincode path
	req = resmgmt.InstallCCRequest{Name: "ID", Version: "v0", Path: ""}
	_, err = rc.InstallCC(req)
	if err == nil {
		t.Fatalf("InstallCC should have failed for empty cc path")
	}

	// Test missing chaincode package
	req = resmgmt.InstallCCRequest{Name: "ID", Version: "v0", Path: "path"}
	_, err = rc.InstallCC(req)
	if err == nil {
		t.Fatalf("InstallCC should have failed for nil chaincode package")
	}

	// Setup test client with different msp (default targets cannot be calculated)
	client := setupTestClient("test", "otherMSP")
	config := getNetworkConfig(t)

	// Create new resource management client ("otherMSP")
	rc = setupResMgmtClient(client, nil, config, t)
	req = resmgmt.InstallCCRequest{Name: "ID", Version: "v0", Path: "path", Package: &fab.CCPackage{Type: 1, Code: []byte("code")}}

	// Test missing default targets
	_, err = rc.InstallCC(req)
	if err == nil {
		t.Fatalf("InstallCC should have failed with no default targets error")
	}

}

func TestInstallCCWithOptsRequiredParameters(t *testing.T) {

	rc := setupDefaultResMgmtClient(t)

	// Test missing required parameters
	req := resmgmt.InstallCCRequest{}
	opts := resmgmt.InstallCCOpts{}
	_, err := rc.InstallCCWithOpts(req, opts)
	if err == nil {
		t.Fatalf("Should have failed for empty install cc request")
	}

	// Test missing chaincode ID
	req = resmgmt.InstallCCRequest{Name: "", Version: "v0", Path: "path"}
	_, err = rc.InstallCCWithOpts(req, opts)
	if err == nil {
		t.Fatalf("Should have failed for empty cc ID")
	}

	// Test missing chaincode version
	req = resmgmt.InstallCCRequest{Name: "ID", Version: "", Path: "path"}
	_, err = rc.InstallCCWithOpts(req, opts)
	if err == nil {
		t.Fatalf("Should have failed for empty cc version")
	}

	// Test missing chaincode path
	req = resmgmt.InstallCCRequest{Name: "ID", Version: "v0", Path: ""}
	_, err = rc.InstallCCWithOpts(req, opts)
	if err == nil {
		t.Fatalf("InstallCC should have failed for empty cc path")
	}

	// Test missing chaincode package
	req = resmgmt.InstallCCRequest{Name: "ID", Version: "v0", Path: "path"}
	_, err = rc.InstallCCWithOpts(req, opts)
	if err == nil {
		t.Fatalf("InstallCC should have failed for nil chaincode package")
	}

	// Valid request
	req = resmgmt.InstallCCRequest{Name: "name", Version: "version", Path: "path", Package: &fab.CCPackage{Type: 1, Code: []byte("code")}}

	// Setup targets
	var peers []fab.Peer
	peer := fcmocks.MockPeer{MockName: "Peer1", MockURL: "http://peer1.com", MockRoles: []string{}, MockCert: nil, MockMSP: "Org1MSP"}
	peers = append(peers, &peer)

	// Test both targets and filter provided (error condition)
	_, err = rc.InstallCCWithOpts(req, resmgmt.InstallCCOpts{Targets: peers, TargetFilter: &MSPFilter{mspID: "Org1MSP"}})
	if err == nil {
		t.Fatalf("Should have failed if both target and filter provided")
	}

	// Setup test client with different msp
	client := setupTestClient("test", "otherMSP")
	config := getNetworkConfig(t)

	// Create new resource management client ("otherMSP")
	rc = setupResMgmtClient(client, nil, config, t)

	// No targets and no filter -- default filter msp doesn't match discovery service peer msp
	_, err = rc.InstallCCWithOpts(req, resmgmt.InstallCCOpts{})
	if err == nil {
		t.Fatalf("Should have failed with no targets error")
	}

	// Test filter only provided (filter rejects discovery service peer msp)
	_, err = rc.InstallCCWithOpts(req, resmgmt.InstallCCOpts{TargetFilter: &MSPFilter{mspID: "Org2MSP"}})
	if err == nil {
		t.Fatalf("Should have failed with no targets since filter rejected all discovery targets")
	}
}

func TestInstallCCDiscoveryError(t *testing.T) {

	// Setup test client and config
	client := setupTestClient("test", "Org1MSP")
	config := getNetworkConfig(t)

	// Create resource management client with discovery service that will generate an error
	rc := setupResMgmtClient(client, errors.New("Test Error"), config, t)

	// Test InstallCC discovery service error
	req := resmgmt.InstallCCRequest{Name: "ID", Version: "v0", Path: "path", Package: &fab.CCPackage{Type: 1, Code: []byte("code")}}
	_, err := rc.InstallCC(req)
	if err == nil {
		t.Fatalf("Should have failed to install cc with discovery error")
	}

	// Test InstallCC discovery service error
	// if targets are not provided discovery service is used
	_, err = rc.InstallCCWithOpts(req, resmgmt.InstallCCOpts{})
	if err == nil {
		t.Fatalf("Should have failed to install cc with opts with discovery error")
	}

}

func TestInstantiateCCRequiredParameters(t *testing.T) {

	rc := setupDefaultResMgmtClient(t)

	// Test missing required parameters
	req := resmgmt.InstantiateCCRequest{}

	// Test empty channel name
	err := rc.InstantiateCC("", req)
	if err == nil {
		t.Fatalf("Should have failed for empty request")
	}

	// Test empty request
	err = rc.InstantiateCC("mychannel", req)
	if err == nil {
		t.Fatalf("Should have failed for empty request")
	}

	// Test missing chaincode ID
	req = resmgmt.InstantiateCCRequest{Name: "", Version: "v0", Path: "path"}
	err = rc.InstantiateCC("mychannel", req)
	if err == nil {
		t.Fatalf("Should have failed for empty cc name")
	}

	// Test missing chaincode version
	req = resmgmt.InstantiateCCRequest{Name: "ID", Version: "", Path: "path"}
	err = rc.InstantiateCC("mychannel", req)
	if err == nil {
		t.Fatalf("Should have failed for empty cc version")
	}

	// Test missing chaincode path
	req = resmgmt.InstantiateCCRequest{Name: "ID", Version: "v0", Path: ""}
	err = rc.InstantiateCC("mychannel", req)
	if err == nil {
		t.Fatalf("Should have failed for empty cc path")
	}

	// Test missing chaincode policy
	req = resmgmt.InstantiateCCRequest{Name: "ID", Version: "v0", Path: "path"}
	err = rc.InstantiateCC("mychannel", req)
	if err == nil {
		t.Fatalf("Should have failed for nil chaincode policy")
	}

	// Setup test client with different msp (default targets cannot be calculated)
	client := setupTestClient("test", "otherMSP")
	config := getNetworkConfig(t)

	// Create new resource management client ("otherMSP")
	rc = setupResMgmtClient(client, nil, config, t)

	// Valid request
	ccPolicy := cauthdsl.SignedByMspMember("otherMSP")
	req = resmgmt.InstantiateCCRequest{Name: "name", Version: "version", Path: "path", Policy: ccPolicy}

	// Test missing default targets
	err = rc.InstantiateCC("mychannel", req)
	if err == nil {
		t.Fatalf("InstallCC should have failed with no default targets error")
	}

}

func TestInstantiateCCWithOptsRequiredParameters(t *testing.T) {

	rc := setupDefaultResMgmtClient(t)

	// Test missing required parameters
	req := resmgmt.InstantiateCCRequest{}
	opts := resmgmt.InstantiateCCOpts{}

	// Test empty channel name
	err := rc.InstantiateCCWithOpts("", req, opts)
	if err == nil {
		t.Fatalf("Should have failed for empty channel name")
	}

	// Test empty request
	err = rc.InstantiateCCWithOpts("mychannel", req, opts)
	if err == nil {
		t.Fatalf("Should have failed for empty instantiate cc request")
	}

	// Test missing chaincode ID
	req = resmgmt.InstantiateCCRequest{Name: "", Version: "v0", Path: "path"}
	err = rc.InstantiateCCWithOpts("mychannel", req, opts)
	if err == nil {
		t.Fatalf("Should have failed for empty cc name")
	}

	// Test missing chaincode version
	req = resmgmt.InstantiateCCRequest{Name: "ID", Version: "", Path: "path"}
	err = rc.InstantiateCCWithOpts("mychannel", req, opts)
	if err == nil {
		t.Fatalf("Should have failed for empty cc version")
	}

	// Test missing chaincode path
	req = resmgmt.InstantiateCCRequest{Name: "ID", Version: "v0", Path: ""}
	err = rc.InstantiateCCWithOpts("mychannel", req, opts)
	if err == nil {
		t.Fatalf("Should have failed for empty cc path")
	}

	// Test missing chaincode policy
	req = resmgmt.InstantiateCCRequest{Name: "ID", Version: "v0", Path: "path"}
	err = rc.InstantiateCCWithOpts("mychannel", req, opts)
	if err == nil {
		t.Fatalf("Should have failed for missing chaincode policy")
	}

	// Valid request
	ccPolicy := cauthdsl.SignedByMspMember("Org1MSP")
	req = resmgmt.InstantiateCCRequest{Name: "name", Version: "version", Path: "path", Policy: ccPolicy}

	// Setup targets
	var peers []fab.Peer
	peer := fcmocks.MockPeer{MockName: "Peer1", MockURL: "http://peer1.com", MockRoles: []string{}, MockCert: nil, MockMSP: "Org1MSP"}
	peers = append(peers, &peer)

	// Test both targets and filter provided (error condition)
	err = rc.InstantiateCCWithOpts("mychannel", req, resmgmt.InstantiateCCOpts{Targets: peers, TargetFilter: &MSPFilter{mspID: "Org1MSP"}})
	if err == nil {
		t.Fatalf("Should have failed if both target and filter provided")
	}

	// Setup test client with different msp
	client := setupTestClient("test", "otherMSP")
	config := getNetworkConfig(t)

	// Create new resource management client ("otherMSP")
	rc = setupResMgmtClient(client, nil, config, t)

	// No targets and no filter -- default filter msp doesn't match discovery service peer msp
	err = rc.InstantiateCCWithOpts("mychannel", req, resmgmt.InstantiateCCOpts{})
	if err == nil {
		t.Fatalf("Should have failed with no targets error")
	}

	// Test filter only provided (filter rejects discovery service peer msp)
	err = rc.InstantiateCCWithOpts("mychannel", req, resmgmt.InstantiateCCOpts{TargetFilter: &MSPFilter{mspID: "Org2MSP"}})
	if err == nil {
		t.Fatalf("Should have failed with no targets since filter rejected all discovery targets")
	}
}

func TestInstantiateCCDiscoveryError(t *testing.T) {

	// Setup test client and config
	client := setupTestClient("test", "Org1MSP")
	config := getNetworkConfig(t)

	// Create resource management client with discovery service that will generate an error
	rc := setupResMgmtClient(client, errors.New("Test Error"), config, t)

	ccPolicy := cauthdsl.SignedByMspMember("Org1MSP")
	req := resmgmt.InstantiateCCRequest{Name: "name", Version: "version", Path: "path", Policy: ccPolicy}

	// Test InstantiateCC create new discovery service per channel error
	err := rc.InstantiateCC("error", req)
	if err == nil {
		t.Fatalf("Should have failed to instantiate cc with create discovery service error")
	}

	// Test InstantiateCC discovery service get peers error
	err = rc.InstantiateCC("mychannel", req)
	if err == nil {
		t.Fatalf("Should have failed to instantiate cc with get peers discovery error")
	}

	// Test InstantiateCCWithOpts create new discovery service per channel error
	err = rc.InstantiateCCWithOpts("error", req, resmgmt.InstantiateCCOpts{})
	if err == nil {
		t.Fatalf("Should have failed to instantiate cc with opts with create discovery service error")
	}

	// Test InstantiateCCWithOpts discovery service get peers error
	// if targets are not provided discovery service is used
	err = rc.InstantiateCCWithOpts("mychannel", req, resmgmt.InstantiateCCOpts{})
	if err == nil {
		t.Fatalf("Should have failed to instantiate cc with opts with get peers discovery error")
	}

}

func TestUpgradeCCRequiredParameters(t *testing.T) {

	rc := setupDefaultResMgmtClient(t)

	// Test missing required parameters
	req := resmgmt.UpgradeCCRequest{}

	// Test empty channel name
	err := rc.UpgradeCC("", req)
	if err == nil {
		t.Fatalf("Should have failed for empty channel name")
	}

	// Test empty request
	err = rc.UpgradeCC("mychannel", req)
	if err == nil {
		t.Fatalf("Should have failed for empty upgrade cc request")
	}

	// Test missing chaincode ID
	req = resmgmt.UpgradeCCRequest{Name: "", Version: "v0", Path: "path"}
	err = rc.UpgradeCC("mychannel", req)
	if err == nil {
		t.Fatalf("Should have failed for empty cc name")
	}

	// Test missing chaincode version
	req = resmgmt.UpgradeCCRequest{Name: "ID", Version: "", Path: "path"}
	err = rc.UpgradeCC("mychannel", req)
	if err == nil {
		t.Fatalf("Should have failed for empty cc version")
	}

	// Test missing chaincode path
	req = resmgmt.UpgradeCCRequest{Name: "ID", Version: "v0", Path: ""}
	err = rc.UpgradeCC("mychannel", req)
	if err == nil {
		t.Fatalf("Should have failed for empty cc path")
	}

	// Test missing chaincode policy
	req = resmgmt.UpgradeCCRequest{Name: "ID", Version: "v0", Path: "path"}
	err = rc.UpgradeCC("mychannel", req)
	if err == nil {
		t.Fatalf("Should have failed for nil chaincode policy")
	}

	// Setup test client with different msp (default targets cannot be calculated)
	client := setupTestClient("test", "otherMSP")
	config := getNetworkConfig(t)

	// Create new resource management client ("otherMSP")
	rc = setupResMgmtClient(client, nil, config, t)

	// Valid request
	ccPolicy := cauthdsl.SignedByMspMember("otherMSP")
	req = resmgmt.UpgradeCCRequest{Name: "name", Version: "version", Path: "path", Policy: ccPolicy}

	// Test missing default targets
	err = rc.UpgradeCC("mychannel", req)
	if err == nil {
		t.Fatalf("Should have failed with no default targets error")
	}

}

func TestUpgradeCCWithOptsRequiredParameters(t *testing.T) {

	rc := setupDefaultResMgmtClient(t)

	// Test missing required parameters
	req := resmgmt.UpgradeCCRequest{}
	opts := resmgmt.UpgradeCCOpts{}

	// Test empty channel name
	err := rc.UpgradeCCWithOpts("", req, opts)
	if err == nil {
		t.Fatalf("Should have failed for empty channel name")
	}

	// Test empty request
	err = rc.UpgradeCCWithOpts("mychannel", req, opts)
	if err == nil {
		t.Fatalf("Should have failed for empty upgrade cc request")
	}

	// Test missing chaincode ID
	req = resmgmt.UpgradeCCRequest{Name: "", Version: "v0", Path: "path"}
	err = rc.UpgradeCCWithOpts("mychannel", req, opts)
	if err == nil {
		t.Fatalf("Should have failed for empty cc name")
	}

	// Test missing chaincode version
	req = resmgmt.UpgradeCCRequest{Name: "ID", Version: "", Path: "path"}
	err = rc.UpgradeCCWithOpts("mychannel", req, opts)
	if err == nil {
		t.Fatalf("Should have failed for empty cc version")
	}

	// Test missing chaincode path
	req = resmgmt.UpgradeCCRequest{Name: "ID", Version: "v0", Path: ""}
	err = rc.UpgradeCCWithOpts("mychannel", req, opts)
	if err == nil {
		t.Fatalf("UpgradeCC should have failed for empty cc path")
	}

	// Test missing chaincode policy
	req = resmgmt.UpgradeCCRequest{Name: "ID", Version: "v0", Path: "path"}
	err = rc.UpgradeCCWithOpts("mychannel", req, opts)
	if err == nil {
		t.Fatalf("UpgradeCC should have failed for missing chaincode policy")
	}

	// Valid request
	ccPolicy := cauthdsl.SignedByMspMember("Org1MSP")
	req = resmgmt.UpgradeCCRequest{Name: "name", Version: "version", Path: "path", Policy: ccPolicy}

	// Setup targets
	var peers []fab.Peer
	peer := fcmocks.MockPeer{MockName: "Peer1", MockURL: "http://peer1.com", MockRoles: []string{}, MockCert: nil, MockMSP: "Org1MSP"}
	peers = append(peers, &peer)

	// Test both targets and filter provided (error condition)
	err = rc.UpgradeCCWithOpts("mychannel", req, resmgmt.UpgradeCCOpts{Targets: peers, TargetFilter: &MSPFilter{mspID: "Org1MSP"}})
	if err == nil {
		t.Fatalf("Should have failed if both target and filter provided")
	}

	// Setup test client with different msp
	client := setupTestClient("test", "otherMSP")
	config := getNetworkConfig(t)

	// Create new resource management client ("otherMSP")
	rc = setupResMgmtClient(client, nil, config, t)

	// No targets and no filter -- default filter msp doesn't match discovery service peer msp
	err = rc.UpgradeCCWithOpts("mychannel", req, resmgmt.UpgradeCCOpts{})
	if err == nil {
		t.Fatalf("Should have failed with no targets error")
	}

	// Test filter only provided (filter rejects discovery service peer msp)
	err = rc.UpgradeCCWithOpts("mychannel", req, resmgmt.UpgradeCCOpts{TargetFilter: &MSPFilter{mspID: "Org2MSP"}})
	if err == nil {
		t.Fatalf("Should have failed with no targets since filter rejected all discovery targets")
	}
}

func TestUpgradeCCDiscoveryError(t *testing.T) {

	// Setup test client and config
	client := setupTestClient("test", "Org1MSP")
	config := getNetworkConfig(t)

	// Create resource management client with discovery service that will generate an error
	rc := setupResMgmtClient(client, errors.New("Test Error"), config, t)

	// Test UpgradeCC discovery service error
	ccPolicy := cauthdsl.SignedByMspMember("Org1MSP")
	req := resmgmt.UpgradeCCRequest{Name: "name", Version: "version", Path: "path", Policy: ccPolicy}

	// Test error while creating discovery service for channel "error"
	err := rc.UpgradeCC("error", req)
	if err == nil {
		t.Fatalf("Should have failed to upgrade cc with discovery error")
	}

	// Test error in discovery service while getting peers
	err = rc.UpgradeCC("mychannel", req)
	if err == nil {
		t.Fatalf("Should have failed to upgrade cc with discovery error")
	}

	// Test UpgradeCCWithOpts discovery service error when creating discovery service for channel 'error'
	err = rc.UpgradeCCWithOpts("error", req, resmgmt.UpgradeCCOpts{})
	if err == nil {
		t.Fatalf("Should have failed to upgrade cc with opts with discovery error")
	}

	// Test UpgradeCCWithOpts discovery service error
	// if targets are not provided discovery service is used to get targets
	err = rc.UpgradeCCWithOpts("mychannel", req, resmgmt.UpgradeCCOpts{})
	if err == nil {
		t.Fatalf("Should have failed to upgrade cc with opts with discovery error")
	}

}

func TestCCProposal(t *testing.T) {

	grpcServer := grpc.NewServer()
	defer grpcServer.Stop()

	// Setup mock targets
	endorserServer, addr := startEndorserServer(t, grpcServer)
	time.Sleep(2 * time.Second)

	client := setupTestClient("Admin", "Org1MSP")

	// Create test channel and add it to the client (no added orderer yet)
	channel, _ := channel.NewChannel("mychannel", client)
	client.SetChannel("mychannel", channel)

	// Setup resource management client
	cfg, err := config.InitConfig("./testdata/ccproposal_test.yaml")
	if err != nil {
		t.Fatal(err)
	}
	rc := setupResMgmtClient(client, nil, cfg, t)

	// Setup target peers
	var peers []fab.Peer
	peer, _ := peer.NewPeer(addr, fcmocks.NewMockConfig())
	peers = append(peers, peer)

	// Create mock orderer
	orderer := fcmocks.NewMockOrderer("", nil)

	// Add orderer to the channel
	err = channel.AddOrderer(orderer)
	if err != nil {
		t.Fatalf("Error adding orderer: %v", err)
	}

	rc = setupResMgmtClient(client, nil, cfg, t)

	ccPolicy := cauthdsl.SignedByMspMember("Org1MSP")
	instantiateReq := resmgmt.InstantiateCCRequest{Name: "name", Version: "version", Path: "path", Policy: ccPolicy}

	// Test failed proposal error handling (endorser returns an error)
	endorserServer.ProposalError = errors.New("Test Error")

	err = rc.InstantiateCCWithOpts("mychannel", instantiateReq, resmgmt.InstantiateCCOpts{Targets: peers})
	if err == nil {
		t.Fatalf("Should have failed to instantiate cc due to endorser error")
	}

	upgradeRequest := resmgmt.UpgradeCCRequest{Name: "name", Version: "version", Path: "path", Policy: ccPolicy}
	err = rc.UpgradeCCWithOpts("mychannel", upgradeRequest, resmgmt.UpgradeCCOpts{Targets: peers})
	if err == nil {
		t.Fatalf("Should have failed to upgrade cc due to endorser error")
	}

	// Remove endorser error
	endorserServer.ProposalError = nil

	// Test error connecting to event hub
	err = rc.InstantiateCC("mychannel", instantiateReq)
	if err == nil {
		t.Fatalf("Should have failed to get event hub since not setup")
	}

	// Start mock event hub
	eventServer, err := fcmocks.StartMockEventServer(fmt.Sprintf("%s:%d", "127.0.0.1", 7053))
	if err != nil {
		t.Fatalf("Failed to start mock event hub: %v", err)
	}
	defer eventServer.Stop()

	// Test error in commit
	err = rc.InstantiateCC("mychannel", instantiateReq)
	if err == nil {
		t.Fatalf("Should have failed due to error in commit")
	}

	// Test invalid function (only 'instatiate' and 'upgrade' are supported)
	err = rc.sendCCProposalWithOpts(3, "mychannel", instantiateReq, resmgmt.InstantiateCCOpts{Targets: peers})
	if err == nil {
		t.Fatalf("Should have failed for invalid function name")
	}

	// Test no event source in config
	cfg, err = config.InitConfig("./testdata/event_source_missing_test.yaml")
	if err != nil {
		t.Fatal(err)
	}

	rc = setupResMgmtClient(client, nil, cfg, t)
	err = rc.InstantiateCC("mychannel", instantiateReq)
	if err == nil {
		t.Fatalf("Should have failed since no event source has been configured")
	}

	expected := "unable to find event source for channel"
	if !strings.Contains(err.Error(), expected) {
		t.Fatalf("Expecting '%s', got '%s'", expected, err.Error())
	}

}

func setupTestDiscovery(discErr error, peers []fab.Peer) (fab.DiscoveryProvider, error) {

	mockDiscovery, err := txnmocks.NewMockDiscoveryProvider(discErr, peers)
	if err != nil {
		return nil, errors.WithMessage(err, "NewMockDiscoveryProvider failed")
	}

	return mockDiscovery, nil
}

func getNetworkConfig(t *testing.T) *config.Config {
	config, err := config.InitConfig("../../../test/fixtures/config/config_test.yaml")
	if err != nil {
		t.Fatal(err)
	}

	return config
}

func setupDefaultResMgmtClient(t *testing.T) *ResourceMgmtClient {
	client := setupTestClient("test", "Org1MSP")
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

func setupTestClient(userName string, mspID string) *fcmocks.MockClient {
	client := fcmocks.NewMockClient()
	user := fcmocks.NewMockUserWithMSPID(userName, mspID)
	cryptoSuite := &fcmocks.MockCryptoSuite{}
	client.SaveUserToStateStore(user, true)
	client.SetUserContext(user)
	client.SetCryptoSuite(cryptoSuite)

	return client
}

func startEndorserServer(t *testing.T, grpcServer *grpc.Server) (*fcmocks.MockEndorserServer, string) {
	lis, err := net.Listen("tcp", "127.0.0.1:7051")
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
