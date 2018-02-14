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

	"github.com/hyperledger/fabric-sdk-go/api/apiconfig"
	fab "github.com/hyperledger/fabric-sdk-go/api/apifabclient"
	resmgmt "github.com/hyperledger/fabric-sdk-go/api/apitxn/resmgmtclient"
	"github.com/hyperledger/fabric-sdk-go/pkg/fabric-client/peer"
	"github.com/hyperledger/fabric-sdk-go/pkg/fabsdk/provider/fabpvdr"
	"github.com/hyperledger/fabric-sdk-go/third_party/github.com/hyperledger/fabric/common/cauthdsl"
	"github.com/pkg/errors"

	"github.com/hyperledger/fabric-sdk-go/pkg/config"
	txnmocks "github.com/hyperledger/fabric-sdk-go/pkg/fabric-txn/mocks"

	"github.com/hyperledger/fabric-sdk-go/pkg/fabric-client/channel"
	fcmocks "github.com/hyperledger/fabric-sdk-go/pkg/fabric-client/mocks"
	pb "github.com/hyperledger/fabric-sdk-go/third_party/github.com/hyperledger/fabric/protos/peer"
)

func TestJoinChannelFail(t *testing.T) {

	ctx := setupTestContext("test", "Org1MSP")

	// Setup resource management client
	config := getNetworkConfig(t)
	ctx.SetConfig(config)

	rc := setupResMgmtClient(ctx, nil, t)
	rc.resource = fcmocks.NewMockInvalidResource()

	// Setup target peers
	peer1, _ := peer.New(fcmocks.NewMockConfig())

	// Test fail genesis block retrieval (no orderer)
	err := rc.JoinChannel("mychannel", resmgmt.WithTargets(peer1))
	if err == nil {
		t.Fatal("Should have failed to get genesis block")
	}
}

func TestJoinChannel(t *testing.T) {
	grpcServer := grpc.NewServer()
	defer grpcServer.Stop()

	ctx := setupTestContext("test", "Org1MSP")

	// Create mock orderer with simple mock block
	orderer := fcmocks.NewMockOrderer("", nil)
	orderer.(fcmocks.MockOrderer).EnqueueForSendDeliver(fcmocks.NewSimpleMockBlock())

	rc := setupResMgmtClient(ctx, nil, t)

	channel, err := channel.New(ctx, fcmocks.NewMockChannelCfg("mychannel"))
	if err != nil {
		t.Fatalf("Error setting up channel: %v", err)
	}
	err = channel.AddOrderer(orderer)
	if err != nil {
		t.Fatalf("Error adding orderer: %v", err)
	}
	rc.channelProvider.(*fcmocks.MockChannelProvider).SetChannel("mychannel", channel)

	// Setup target peers
	var peers []fab.Peer
	peer1, _ := peer.New(fcmocks.NewMockConfig(), peer.WithURL("example.com"))
	peers = append(peers, peer1)

	// Test valid join channel request (success)
	err = rc.JoinChannel("mychannel", resmgmt.WithTargets(peer1))
	if err != nil {
		t.Fatal(err)
	}
}

func TestNoSigningUserFailure(t *testing.T) {
	user := fcmocks.NewMockUserWithMSPID("test", "")

	// Setup client without user context
	fabCtx := fcmocks.NewMockContext(user)
	config := getNetworkConfig(t)
	fabCtx.SetConfig(config)
	resource := fcmocks.NewMockResource()

	discovery, err := setupTestDiscovery(nil, nil)
	if err != nil {
		t.Fatalf("Failed to setup discovery service: %s", err)
	}

	chProvider, err := fcmocks.NewMockChannelProvider(fabCtx)
	if err != nil {
		t.Fatalf("Failed to setup channel provider: %s", err)
	}

	ctx := Context{
		ProviderContext:   fabCtx,
		IdentityContext:   fabCtx,
		Resource:          resource,
		ChannelProvider:   chProvider,
		DiscoveryProvider: discovery,
	}
	_, err = New(ctx, nil)
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
	ctx := setupTestContext("test", "otherMSP")
	config := getNetworkConfig(t)
	ctx.SetConfig(config)

	// Create new resource management client ("otherMSP")
	rc = setupResMgmtClient(ctx, nil, t)

	// Test missing default targets
	err = rc.JoinChannel("mychannel")
	if err == nil || !strings.Contains(err.Error(), "No targets available") {
		t.Fatalf("InstallCC should have failed with no default targets error")
	}

}

func TestJoinChannelWithOptsRequiredParameters(t *testing.T) {

	rc := setupDefaultResMgmtClient(t)

	// Test empty channel name for request with no opts
	err := rc.JoinChannel("")
	if err == nil {
		t.Fatalf("Should have failed for empty channel name")
	}

	var peers []fab.Peer
	peer := fcmocks.MockPeer{MockName: "Peer1", MockURL: "http://peer1.com", MockRoles: []string{}, MockCert: nil}
	peers = append(peers, &peer)

	// Test both targets and filter provided (error condition)
	err = rc.JoinChannel("mychannel", resmgmt.WithTargets(peers...), resmgmt.WithTargetFilter(&MSPFilter{mspID: "MspID"}))
	if err == nil || !strings.Contains(err.Error(), "If targets are provided, filter cannot be provided") {
		t.Fatalf("Should have failed if both target and filter provided")
	}

	// Test targets only
	err = rc.JoinChannel("mychannel", resmgmt.WithTargets(peers...))
	if err != nil {
		t.Fatalf(err.Error())
	}

	// Test filter only (filter has no match)
	err = rc.JoinChannel("mychannel", resmgmt.WithTargetFilter(&MSPFilter{mspID: "MspID"}))
	if err == nil || !strings.Contains(err.Error(), "No targets available") {
		t.Fatalf("InstallCC should have failed with no targets error")
	}

	// Test filter only (filter has a match)
	err = rc.JoinChannel("mychannel", resmgmt.WithTargetFilter(&MSPFilter{mspID: "Org1MSP"}))
	if err != nil {
		t.Fatalf(err.Error())
	}

}

func TestJoinChannelDiscoveryError(t *testing.T) {

	// Setup test client and config
	ctx := setupTestContext("test", "Org1MSP")
	config := getNetworkConfig(t)
	ctx.SetConfig(config)

	// Create resource management client with discovery service that will generate an error
	rc := setupResMgmtClient(ctx, errors.New("Test Error"), t)

	err := rc.JoinChannel("mychannel")
	if err == nil {
		t.Fatalf("Should have failed to join channel with discovery error")
	}

	// If targets are not provided discovery service is used
	err = rc.JoinChannel("mychannel")
	if err == nil {
		t.Fatalf("Should have failed to join channel with discovery error")
	}

}

func TestJoinChannelNoOrdererConfig(t *testing.T) {

	ctx := setupTestContext("test", "Org1MSP")

	// No channel orderer, no global orderer
	noOrdererConfig, err := config.FromFile("./testdata/noorderer_test.yaml")()
	if err != nil {
		t.Fatal(err)
	}
	ctx.SetConfig(noOrdererConfig)
	rc := setupResMgmtClient(ctx, nil, t)

	err = rc.JoinChannel("mychannel")
	if err == nil {
		t.Fatalf("Should have failed to join channel since no orderer has been configured")
	}

	// Misconfigured channel orderer
	invalidChOrdererConfig, err := config.FromFile("./testdata/invalidchorderer_test.yaml")()
	if err != nil {
		t.Fatal(err)
	}
	ctx.SetConfig(invalidChOrdererConfig)
	rc = setupResMgmtClient(ctx, nil, t)

	err = rc.JoinChannel("mychannel")
	if err == nil {
		t.Fatalf("Should have failed to join channel since channel orderer has been misconfigured")
	}

	// Misconfigured global orderer (cert cannot be loaded)
	invalidOrdererConfig, err := config.FromFile("./testdata/invalidorderer_test.yaml")()
	if err != nil {
		t.Fatal(err)
	}
	ctx.SetConfig(invalidOrdererConfig)
	rc = setupResMgmtClient(ctx, nil, t)

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
		t.Fatalf("CC should have been installed: %v", req)
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
	peer := fcmocks.MockPeer{MockName: "Peer1", MockURL: "http://peer1.com",
		Status: 200, MockRoles: []string{}, MockCert: nil, MockMSP: "Org1MSP"}
	peers = append(peers, &peer)

	// Already installed chaincode request
	req := resmgmt.InstallCCRequest{Name: "name", Version: "version", Path: "path", Package: &fab.CCPackage{Type: 1, Code: []byte("code")}}
	responses, err := rc.InstallCC(req, resmgmt.WithTargets(peers...))
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
	responses, err = rc.InstallCC(req, resmgmt.WithTargets(peers...))
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
	_, err = rc.InstallCC(req, resmgmt.WithTargets(peers...))
	if err == nil {
		t.Fatalf("Should have failed since install cc returns an error in the client")
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
	ctx := setupTestContext("test", "otherMSP")
	config := getNetworkConfig(t)
	ctx.SetConfig(config)

	// Create new resource management client ("otherMSP")
	rc = setupResMgmtClient(ctx, nil, t)
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

	// Valid request
	req = resmgmt.InstallCCRequest{Name: "name", Version: "version", Path: "path", Package: &fab.CCPackage{Type: 1, Code: []byte("code")}}

	// Setup targets
	var peers []fab.Peer
	peer := fcmocks.MockPeer{MockName: "Peer1", MockURL: "http://peer1.com", MockRoles: []string{}, MockCert: nil, MockMSP: "Org1MSP"}
	peers = append(peers, &peer)

	// Test both targets and filter provided (error condition)
	_, err = rc.InstallCC(req, resmgmt.WithTargets(peers...), resmgmt.WithTargetFilter(&MSPFilter{mspID: "Org1MSP"}))
	if err == nil {
		t.Fatalf("Should have failed if both target and filter provided")
	}

	// Setup test client with different msp
	ctx := setupTestContext("test", "otherMSP")
	config := getNetworkConfig(t)
	ctx.SetConfig(config)

	// Create new resource management client ("otherMSP")
	rc = setupResMgmtClient(ctx, nil, t)

	// No targets and no filter -- default filter msp doesn't match discovery service peer msp
	_, err = rc.InstallCC(req)
	if err == nil {
		t.Fatalf("Should have failed with no targets error")
	}

	// Test filter only provided (filter rejects discovery service peer msp)
	_, err = rc.InstallCC(req, resmgmt.WithTargetFilter(&MSPFilter{mspID: "Org2MSP"}))
	if err == nil {
		t.Fatalf("Should have failed with no targets since filter rejected all discovery targets")
	}
}

func TestInstallCCDiscoveryError(t *testing.T) {

	// Setup test client and config
	ctx := setupTestContext("test", "Org1MSP")
	config := getNetworkConfig(t)
	ctx.SetConfig(config)

	// Create resource management client with discovery service that will generate an error
	rc := setupResMgmtClient(ctx, errors.New("Test Error"), t)

	// Test InstallCC discovery service error
	req := resmgmt.InstallCCRequest{Name: "ID", Version: "v0", Path: "path", Package: &fab.CCPackage{Type: 1, Code: []byte("code")}}
	_, err := rc.InstallCC(req)
	if err == nil {
		t.Fatalf("Should have failed to install cc with discovery error")
	}

	// Test InstallCC discovery service error
	// if targets are not provided discovery service is used
	_, err = rc.InstallCC(req)
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
	ctx := setupTestContext("test", "otherMSP")
	config := getNetworkConfig(t)
	ctx.SetConfig(config)

	// Create new resource management client ("otherMSP")
	rc = setupResMgmtClient(ctx, nil, t)

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

	// Test empty channel name
	err := rc.InstantiateCC("", req)
	if err == nil {
		t.Fatalf("Should have failed for empty channel name")
	}

	// Test empty request
	err = rc.InstantiateCC("mychannel", req)
	if err == nil {
		t.Fatalf("Should have failed for empty instantiate cc request")
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
	err = rc.InstantiateCC("mychannel", req, resmgmt.WithTargets(peers...), resmgmt.WithTargetFilter(&MSPFilter{mspID: "Org1MSP"}))
	if err == nil {
		t.Fatalf("Should have failed if both target and filter provided")
	}

	// Setup test client with different msp
	ctx := setupTestContext("test", "otherMSP")
	config := getNetworkConfig(t)
	ctx.SetConfig(config)

	// Create new resource management client ("otherMSP")
	rc = setupResMgmtClient(ctx, nil, t)

	// No targets and no filter -- default filter msp doesn't match discovery service peer msp
	err = rc.InstantiateCC("mychannel", req)
	if err == nil {
		t.Fatalf("Should have failed with no targets error")
	}

	// Test filter only provided (filter rejects discovery service peer msp)
	err = rc.InstantiateCC("mychannel", req, resmgmt.WithTargetFilter(&MSPFilter{mspID: "Org2MSP"}))
	if err == nil {
		t.Fatalf("Should have failed with no targets since filter rejected all discovery targets")
	}
}

func TestInstantiateCCDiscoveryError(t *testing.T) {

	// Setup test client and config
	ctx := setupTestContext("test", "Org1MSP")
	config := getNetworkConfig(t)
	ctx.SetConfig(config)

	// Create resource management client with discovery service that will generate an error
	rc := setupResMgmtClient(ctx, errors.New("Test Error"), t)

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
	err = rc.InstantiateCC("error", req)
	if err == nil {
		t.Fatalf("Should have failed to instantiate cc with opts with create discovery service error")
	}

	// Test InstantiateCCWithOpts discovery service get peers error
	// if targets are not provided discovery service is used
	err = rc.InstantiateCC("mychannel", req)
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
	ctx := setupTestContext("test", "otherMSP")
	config := getNetworkConfig(t)
	ctx.SetConfig(config)

	// Create new resource management client ("otherMSP")
	rc = setupResMgmtClient(ctx, nil, t)

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
		t.Fatalf("UpgradeCC should have failed for empty cc path")
	}

	// Test missing chaincode policy
	req = resmgmt.UpgradeCCRequest{Name: "ID", Version: "v0", Path: "path"}
	err = rc.UpgradeCC("mychannel", req)
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
	err = rc.UpgradeCC("mychannel", req, resmgmt.WithTargets(peers...), resmgmt.WithTargetFilter(&MSPFilter{mspID: "Org1MSP"}))
	if err == nil {
		t.Fatalf("Should have failed if both target and filter provided")
	}

	// Setup test client with different msp
	ctx := setupTestContext("test", "otherMSP")
	config := getNetworkConfig(t)
	ctx.SetConfig(config)

	// Create new resource management client ("otherMSP")
	rc = setupResMgmtClient(ctx, nil, t)

	// No targets and no filter -- default filter msp doesn't match discovery service peer msp
	err = rc.UpgradeCC("mychannel", req)
	if err == nil {
		t.Fatalf("Should have failed with no targets error")
	}

	// Test filter only provided (filter rejects discovery service peer msp)
	err = rc.UpgradeCC("mychannel", req, resmgmt.WithTargetFilter(&MSPFilter{mspID: "Org2MSP"}))
	if err == nil {
		t.Fatalf("Should have failed with no targets since filter rejected all discovery targets")
	}
}

func TestUpgradeCCDiscoveryError(t *testing.T) {

	// Setup test client and config
	ctx := setupTestContext("test", "Org1MSP")
	config := getNetworkConfig(t)
	ctx.SetConfig(config)

	// Create resource management client with discovery service that will generate an error
	rc := setupResMgmtClient(ctx, errors.New("Test Error"), t)

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
	err = rc.UpgradeCC("error", req)
	if err == nil {
		t.Fatalf("Should have failed to upgrade cc with opts with discovery error")
	}

	// Test UpgradeCCWithOpts discovery service error
	// if targets are not provided discovery service is used to get targets
	err = rc.UpgradeCC("mychannel", req)
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

	ctx := setupTestContext("Admin", "Org1MSP")

	// Setup resource management client
	cfg, err := config.FromFile("./testdata/ccproposal_test.yaml")()
	if err != nil {
		t.Fatal(err)
	}
	ctx.SetConfig(cfg)

	rc := setupResMgmtClient(ctx, nil, t)

	// Setup target peers
	var peers []fab.Peer
	peer1, _ := peer.New(fcmocks.NewMockConfig(), peer.WithURL(addr))
	peers = append(peers, peer1)

	// Create mock orderer
	orderer := fcmocks.NewMockOrderer("", nil)

	rc = setupResMgmtClient(ctx, nil, t)

	channel, err := channel.New(ctx, fcmocks.NewMockChannelCfg("mychannel"))
	if err != nil {
		t.Fatalf("Error setting up channel: %v", err)
	}
	err = channel.AddOrderer(orderer)
	if err != nil {
		t.Fatalf("Error adding orderer: %v", err)
	}
	rc.channelProvider.(*fcmocks.MockChannelProvider).SetChannel("mychannel", channel)

	ccPolicy := cauthdsl.SignedByMspMember("Org1MSP")
	instantiateReq := resmgmt.InstantiateCCRequest{Name: "name", Version: "version", Path: "path", Policy: ccPolicy}

	// Test failed proposal error handling (endorser returns an error)
	endorserServer.ProposalError = errors.New("Test Error")

	err = rc.InstantiateCC("mychannel", instantiateReq, resmgmt.WithTargets(peers...))
	if err == nil {
		t.Fatalf("Should have failed to instantiate cc due to endorser error")
	}

	upgradeRequest := resmgmt.UpgradeCCRequest{Name: "name", Version: "version", Path: "path", Policy: ccPolicy}
	err = rc.UpgradeCC("mychannel", upgradeRequest, resmgmt.WithTargets(peers...))
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
	err = rc.sendCCProposal(3, "mychannel", instantiateReq, resmgmt.WithTargets(peers...))
	if err == nil {
		t.Fatalf("Should have failed for invalid function name")
	}

	// Test no event source in config
	cfg, err = config.FromFile("./testdata/event_source_missing_test.yaml")()
	if err != nil {
		t.Fatal(err)
	}
	ctx.SetConfig(cfg)

	rc = setupResMgmtClient(ctx, nil, t)
	err = rc.InstantiateCC("mychannel", instantiateReq)
	if err == nil {
		t.Fatalf("Should have failed since no event source has been configured")
	}
}

func setupTestDiscovery(discErr error, peers []fab.Peer) (fab.DiscoveryProvider, error) {

	mockDiscovery, err := txnmocks.NewMockDiscoveryProvider(discErr, peers)
	if err != nil {
		return nil, errors.WithMessage(err, "NewMockDiscoveryProvider failed")
	}

	return mockDiscovery, nil
}

func getNetworkConfig(t *testing.T) apiconfig.Config {
	config, err := config.FromFile("../../../test/fixtures/config/config_test.yaml")()
	if err != nil {
		t.Fatal(err)
	}

	return config
}

func setupDefaultResMgmtClient(t *testing.T) *ResourceMgmtClient {
	ctx := setupTestContext("test", "Org1MSP")
	network := getNetworkConfig(t)
	ctx.SetConfig(network)

	return setupResMgmtClient(ctx, nil, t)
}

func setupResMgmtClient(fabCtx fab.Context, discErr error, t *testing.T) *ResourceMgmtClient {

	fabProvider := fabpvdr.New(fabCtx)

	discovery, err := setupTestDiscovery(discErr, nil)
	if err != nil {
		t.Fatalf("Failed to setup discovery service: %s", err)
	}

	chProvider, err := fcmocks.NewMockChannelProvider(fabCtx)
	if err != nil {
		t.Fatalf("Failed to setup channel provider: %s", err)
	}

	resource := fcmocks.NewMockResource()

	ctx := Context{
		ProviderContext:   fabCtx,
		IdentityContext:   fabCtx,
		Resource:          resource,
		ChannelProvider:   chProvider,
		DiscoveryProvider: discovery,
		FabricProvider:    fabProvider,
	}
	resClient, err := New(ctx, nil)
	if err != nil {
		t.Fatalf("Failed to create new channel management client: %s", err)
	}

	return resClient
}

func setupTestContext(userName string, mspID string) *fcmocks.MockContext {
	user := fcmocks.NewMockUserWithMSPID(userName, mspID)
	ctx := fcmocks.NewMockContext(user)
	return ctx
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
