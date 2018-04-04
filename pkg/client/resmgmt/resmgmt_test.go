/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package resmgmt

import (
	"fmt"
	"net"
	"net/http"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/hyperledger/fabric-sdk-go/pkg/common/errors/retry"

	"github.com/golang/protobuf/proto"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc"

	txnmocks "github.com/hyperledger/fabric-sdk-go/pkg/client/common/mocks"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/errors/status"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/context"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/fab"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/msp"
	contextImpl "github.com/hyperledger/fabric-sdk-go/pkg/context"
	configImpl "github.com/hyperledger/fabric-sdk-go/pkg/core/config"
	fabImpl "github.com/hyperledger/fabric-sdk-go/pkg/fab"
	fcmocks "github.com/hyperledger/fabric-sdk-go/pkg/fab/mocks"
	"github.com/hyperledger/fabric-sdk-go/pkg/fab/peer"
	"github.com/hyperledger/fabric-sdk-go/pkg/fab/resource/api"
	"github.com/hyperledger/fabric-sdk-go/pkg/fabsdk/provider/fabpvdr"
	mspmocks "github.com/hyperledger/fabric-sdk-go/pkg/msp/test/mockmsp"
	"github.com/hyperledger/fabric-sdk-go/third_party/github.com/hyperledger/fabric/common/cauthdsl"
	"github.com/hyperledger/fabric-sdk-go/third_party/github.com/hyperledger/fabric/protos/common"
	pb "github.com/hyperledger/fabric-sdk-go/third_party/github.com/hyperledger/fabric/protos/peer"
)

const channelConfig = "../../../test/fixtures/fabric/v1.0/channel/mychannel.tx"
const networkCfg = "../../../test/fixtures/config/config_test.yaml"

func TestJoinChannelFail(t *testing.T) {

	grpcServer := grpc.NewServer()
	defer grpcServer.Stop()

	endorserServer, addr := startEndorserServer(t, grpcServer)
	ctx := setupTestContext("test", "Org1MSP")

	// Create mock orderer with simple mock block
	orderer := fcmocks.NewMockOrderer("", nil)
	defer orderer.Close()
	orderer.EnqueueForSendDeliver(fcmocks.NewSimpleMockBlock())
	orderer.EnqueueForSendDeliver(common.Status_SUCCESS)

	setupCustomOrderer(ctx, orderer)

	rc := setupResMgmtClient(ctx, nil, t)

	// Setup target peers
	peer1, _ := peer.New(fcmocks.NewMockEndpointConfig(), peer.WithURL("grpc://"+addr))

	// Test fail with send proposal error
	endorserServer.ProposalError = errors.New("Test Error")
	err := rc.JoinChannel("mychannel", WithTargets(peer1))

	if err == nil {
		t.Fatal("Should have failed to get genesis block")
	}

}

func TestJoinChannelSuccess(t *testing.T) {
	grpcServer := grpc.NewServer()
	defer grpcServer.Stop()

	_, addr := startEndorserServer(t, grpcServer)
	ctx := setupTestContext("test", "Org1MSP")

	// Create mock orderer with simple mock block
	orderer := fcmocks.NewMockOrderer("", nil)
	defer orderer.Close()
	orderer.EnqueueForSendDeliver(fcmocks.NewSimpleMockBlock())
	orderer.EnqueueForSendDeliver(common.Status_SUCCESS)

	setupCustomOrderer(ctx, orderer)

	rc := setupResMgmtClient(ctx, nil, t)

	// Setup target peers
	peer1, _ := peer.New(fcmocks.NewMockEndpointConfig(), peer.WithURL("grpc://"+addr))

	// Test valid join channel request (success)
	err := rc.JoinChannel("mychannel", WithTargets(peer1))
	if err != nil {
		t.Fatal(err)
	}

}

func TestWithFilterOption(t *testing.T) {
	ctx := setupTestContext("test", "Org1MSP")
	rc := setupResMgmtClient(ctx, nil, t, getDefaultTargetFilterOption())
	if rc == nil {
		t.Fatal("Expected Resource Management Client to be set")
	}
}

func TestJoinChannelWithFilter(t *testing.T) {
	grpcServer := grpc.NewServer()
	defer grpcServer.Stop()

	_, addr := startEndorserServer(t, grpcServer)
	ctx := setupTestContext("test", "Org1MSP")

	// Create mock orderer with simple mock block
	orderer := fcmocks.NewMockOrderer("", nil)
	defer orderer.Close()
	orderer.EnqueueForSendDeliver(fcmocks.NewSimpleMockBlock())
	orderer.EnqueueForSendDeliver(common.Status_SUCCESS)
	setupCustomOrderer(ctx, orderer)

	//the target filter ( client option) will be set
	rc := setupResMgmtClient(ctx, nil, t)

	// Setup target peers
	peer1, _ := peer.New(fcmocks.NewMockEndpointConfig(), peer.WithURL("grpc://"+addr))

	// Test valid join channel request (success)
	err := rc.JoinChannel("mychannel", WithTargets(peer1))
	if err != nil {
		t.Fatal(err)
	}
}

func TestNoSigningUserFailure(t *testing.T) {
	user := mspmocks.NewMockSigningIdentity("test", "")

	// Setup client without user context
	fabCtx := fcmocks.NewMockContext(user)
	config := getNetworkConfig(t)
	fabCtx.SetEndpointConfig(config)

	clientCtx := createClientContext(contextImpl.Client{
		Providers:       fabCtx,
		SigningIdentity: fabCtx,
	})

	_, err := New(clientCtx)
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

	// Setup test client with different msp (default targets cannot be calculated)
	ctx := setupTestContext("test", "otherMSP")
	config := getNetworkConfig(t)
	ctx.SetEndpointConfig(config)

	// Create new resource management client ("otherMSP")
	rc = setupResMgmtClient(ctx, nil, t)

	// Test missing default targets
	err = rc.JoinChannel("mychannel")

	assert.NotNil(t, err, "error should have been returned")
	s, ok := status.FromError(err)
	assert.True(t, ok, "status code should be available")
	assert.Equal(t, status.NoPeersFound.ToInt32(), s.Code, "code should be no peers found")
}

func TestJoinChannelWithOptsRequiredParameters(t *testing.T) {

	grpcServer := grpc.NewServer()
	defer grpcServer.Stop()

	_, addr := startEndorserServer(t, grpcServer)

	ctx := setupTestContextWithDiscoveryError("test", "Org1MSP", nil)
	network := getNetworkConfig(t)
	ctx.SetEndpointConfig(network)

	// Create mock orderer with simple mock block
	orderer := fcmocks.NewMockOrderer("", nil)
	defer orderer.Close()
	orderer.EnqueueForSendDeliver(fcmocks.NewSimpleMockBlock())
	orderer.EnqueueForSendDeliver(common.Status_SUCCESS)
	setupCustomOrderer(ctx, orderer)

	rc := setupResMgmtClient(ctx, nil, t, getDefaultTargetFilterOption())

	// Test empty channel name for request with no opts
	err := rc.JoinChannel("")
	if err == nil {
		t.Fatalf("Should have failed for empty channel name")
	}

	var peers []fab.Peer
	peer1, _ := peer.New(fcmocks.NewMockEndpointConfig(), peer.WithURL("grpc://"+addr), peer.WithMSPID("Org1MSP"))
	peers = append(peers, peer1)

	// Test both targets and filter provided (error condition)
	err = rc.JoinChannel("mychannel", WithTargets(peers...), WithTargetFilter(&mspFilter{mspID: "MSPID"}))
	if err == nil || !strings.Contains(err.Error(), "If targets are provided, filter cannot be provided") {
		t.Fatalf("Should have failed if both target and filter provided")
	}

	// Test targets only
	err = rc.JoinChannel("mychannel", WithTargets(peers...))
	if err != nil {
		t.Fatalf(err.Error())
	}

	// Test filter only (filter has no match)
	err = rc.JoinChannel("mychannel", WithTargetFilter(&mspFilter{mspID: "MSPID"}))
	assert.NotNil(t, err, "error should have been returned")
	s, ok := status.FromError(err)
	assert.True(t, ok, "status code should be available")
	assert.Equal(t, status.NoPeersFound.ToInt32(), s.Code, "code should be no peers found")

	//Some cleanup before further test
	orderer = fcmocks.NewMockOrderer("", nil)
	defer orderer.Close()
	orderer.EnqueueForSendDeliver(fcmocks.NewSimpleMockBlock())
	orderer.EnqueueForSendDeliver(common.Status_SUCCESS)
	setupCustomOrderer(ctx, orderer)

	rc = setupResMgmtClient(ctx, nil, t, getDefaultTargetFilterOption())
	disProvider, _ := fcmocks.NewMockDiscoveryProvider(nil, peers)
	rc.discovery, _ = disProvider.CreateDiscoveryService("mychannel")

	// Test filter only (filter has a match)
	err = rc.JoinChannel("mychannel", WithTargetFilter(&mspFilter{mspID: "Org1MSP"}))
	if err != nil {
		t.Fatalf(err.Error())
	}

}

func TestJoinChannelDiscoveryError(t *testing.T) {

	// Setup test client and config
	ctx := setupTestContextWithDiscoveryError("test", "Org1MSP", errors.New("Test Error"))
	config := getNetworkConfig(t)
	ctx.SetEndpointConfig(config)

	// Create resource management client with discovery service that will generate an error
	rc := setupResMgmtClient(ctx, nil, t)

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

func TestOrdererConfigFail(t *testing.T) {

	ctx := setupTestContext("test", "Org1MSP")

	// No channel orderer, no global orderer
	configBackend, err := configImpl.FromFile("./testdata/noorderer_test.yaml")()
	assert.Nil(t, err)

	noOrdererConfig, err := fabImpl.ConfigFromBackend(configBackend)
	assert.Nil(t, err)

	ctx.SetEndpointConfig(noOrdererConfig)
	rc := setupResMgmtClient(ctx, nil, t)

	orderer, err := rc.ordererConfig("mychannel")
	assert.Nil(t, orderer)
	assert.NotNil(t, err, "should fail since no orderer has been configured")
}

/*
func TestOrdererConfigFromOpts(t *testing.T) {
	ctx := setupTestContext("test", "Org1MSP")

	// No channel orderer, no global orderer
	noOrdererConfig, err := config.FromFile("./testdata/ccproposal_test.yaml")()
	assert.Nil(t, err)

	ctx.SetEndpointConfig(noOrdererConfig)
	rc := setupResMgmtClient(ctx, nil, t)

	opts := Opts{}
	orderer, err := rc.ordererConfig(&opts, "mychannel")
	assert.Nil(t, orderer)
	assert.NotNil(t, err, "should fail since no orderer has been configured")
}*/

func TestJoinChannelNoOrdererConfig(t *testing.T) {

	ctx := setupTestContext("test", "Org1MSP")

	// No channel orderer, no global orderer
	configBackend, err := configImpl.FromFile("./testdata/noorderer_test.yaml")()
	if err != nil {
		t.Fatal(err)
	}
	noOrdererConfig, err := fabImpl.ConfigFromBackend(configBackend)
	if err != nil {
		t.Fatal(err)
	}
	ctx.SetEndpointConfig(noOrdererConfig)
	rc := setupResMgmtClient(ctx, nil, t)

	err = rc.JoinChannel("mychannel")
	assert.NotNil(t, err, "Should have failed to join channel since no orderer has been configured")

	// Misconfigured channel orderer
	configBackend, err = configImpl.FromFile("./testdata/invalidchorderer_test.yaml")()
	if err != nil {
		t.Fatal(err)
	}
	invalidChOrdererConfig, err := fabImpl.ConfigFromBackend(configBackend)
	if err != nil {
		t.Fatal(err)
	}
	ctx.SetEndpointConfig(invalidChOrdererConfig)

	rc = setupResMgmtClient(ctx, nil, t)

	err = rc.JoinChannel("mychannel")
	if err == nil {
		t.Fatalf("Should have failed to join channel since channel orderer has been misconfigured")
	}

	// Misconfigured global orderer (cert cannot be loaded)
	configBackend, err = configImpl.FromFile("./testdata/invalidorderer_test.yaml")()
	if err != nil {
		t.Fatal(err)
	}
	invalidOrdererConfig, err := fabImpl.ConfigFromBackend(configBackend)
	if err != nil {
		t.Fatal(err)
	}
	ctx.SetEndpointConfig(invalidOrdererConfig)
	customFabProvider := fabpvdr.New(ctx.EndpointConfig())
	customFabProvider.Initialize(ctx)
	ctx.SetCustomInfraProvider(customFabProvider)

	rc = setupResMgmtClient(ctx, nil, t)

	err = rc.JoinChannel("mychannel")
	if err == nil {
		t.Fatalf("Should have failed to join channel since global orderer certs are not configured properly")
	}
}

func TestIsChaincodeInstalled(t *testing.T) {

	rc := setupDefaultResMgmtClient(t)

	//prepare sample response
	response := new(pb.ChaincodeQueryResponse)
	chaincodes := make([]*pb.ChaincodeInfo, 1)
	chaincodes[0] = &pb.ChaincodeInfo{Name: "test-name", Path: "test-path", Version: "test-version"}
	response.Chaincodes = chaincodes
	responseBytes, err := proto.Marshal(response)
	if err != nil {
		t.Fatal("failed to marshal sample response")
	}

	peer1 := &fcmocks.MockPeer{MockName: "Peer1", MockURL: "grpc://peer1.com", MockRoles: []string{}, MockCert: nil, MockMSP: "Org1MSP", Status: http.StatusOK, Payload: responseBytes}

	// Chaincode found request
	req := InstallCCRequest{Name: "test-name", Path: "test-path", Version: "test-version"}

	reqCtx, cancel := contextImpl.NewRequest(rc.ctx, contextImpl.WithTimeout(10*time.Second))
	defer cancel()
	// Test chaincode installed (valid peer)
	installed, err := rc.isChaincodeInstalled(reqCtx, req, peer1, retry.Opts{})
	if err != nil {
		t.Fatal(err)
	}
	if !installed {
		t.Fatalf("CC should have been installed: %v", req)
	}

	// Chaincode not found request
	req = InstallCCRequest{Name: "ID", Version: "v0", Path: "path"}

	// Test chaincode installed
	installed, err = rc.isChaincodeInstalled(reqCtx, req, peer1, retry.Opts{})
	if err != nil {
		t.Fatal(err)
	}
	if installed {
		t.Fatalf("CC should NOT have been installed: %s", req)
	}

	// Test error retrieving installed cc info (peer is nil)
	_, err = rc.isChaincodeInstalled(reqCtx, req, nil, retry.Opts{})
	if err == nil {
		t.Fatalf("Should have failed with error in get installed chaincodes")
	}

}

func TestQueryInstalledChaincodes(t *testing.T) {

	rc := setupDefaultResMgmtClient(t)

	// Test error
	_, err := rc.QueryInstalledChaincodes()
	if err == nil {
		t.Fatalf("QueryInstalledChaincodes: peer cannot be nil")
	}

	peer := &fcmocks.MockPeer{MockName: "Peer1", MockURL: "http://peer1.com", MockRoles: []string{}, MockCert: nil, MockMSP: "Org1MSP", Status: http.StatusOK}

	// Test success (valid peer)
	_, err = rc.QueryInstalledChaincodes(WithTargets(peer))
	if err != nil {
		t.Fatal(err)
	}

}

func TestQueryChannels(t *testing.T) {

	rc := setupDefaultResMgmtClient(t)

	//prepare sample response
	response := new(pb.ChannelQueryResponse)
	channels := make([]*pb.ChannelInfo, 1)
	channels[0] = &pb.ChannelInfo{ChannelId: "test"}
	response.Channels = channels

	responseBytes, err := proto.Marshal(response)
	if err != nil {
		t.Fatal("failed to marshal sample response")
	}

	// Test error
	_, err = rc.QueryChannels()
	if err == nil {
		t.Fatalf("QueryChannels: peer cannot be nil")
	}

	peer := &fcmocks.MockPeer{MockName: "Peer1", MockURL: "http://peer1.com", MockRoles: []string{}, MockCert: nil, MockMSP: "Org1MSP", Status: http.StatusOK, Payload: responseBytes}

	// Test success (valid peer)
	found := false
	response, err = rc.QueryChannels(WithTargets(peer))
	if err != nil {
		t.Fatalf("failed to query channel for peer: %s", err)
	}
	for _, responseChannel := range response.Channels {
		if responseChannel.ChannelId == "test" {
			found = true
		}
	}

	if !found {
		t.Fatal("Peer has not joined 'test' channel")
	}

}

func TestInstallCCWithOpts(t *testing.T) {

	rc := setupDefaultResMgmtClient(t)

	//prepare sample response
	response := new(pb.ChaincodeQueryResponse)
	chaincodes := make([]*pb.ChaincodeInfo, 1)
	chaincodes[0] = &pb.ChaincodeInfo{Name: "name", Path: "path", Version: "version"}
	response.Chaincodes = chaincodes
	responseBytes, err := proto.Marshal(response)
	assert.Nil(t, err, "marshal should not have failed")

	// Setup targets
	peer1 := fcmocks.MockPeer{MockName: "Peer1", MockURL: "http://peer1.com",
		Status: http.StatusOK, MockRoles: []string{}, MockCert: nil, MockMSP: "Org1MSP", Payload: responseBytes}

	// Already installed chaincode request
	req := InstallCCRequest{Name: "name", Version: "version", Path: "path", Package: &api.CCPackage{Type: 1, Code: []byte("code")}}
	responses, err := rc.InstallCC(req, WithTargets(&peer1))
	if err != nil {
		t.Fatal(err)
	}

	if responses == nil || len(responses) != 1 {
		t.Fatal("Should have one 'already installed' response")
	}

	if !strings.Contains(responses[0].Info, "already installed") {
		t.Fatal("Should have 'already installed' info set")
	}

	if responses[0].Target != peer1.MockURL {
		t.Fatalf("Expecting %s target URL, got %s", peer1.MockURL, responses[0].Target)
	}

	// Chaincode not found request (it will be installed)
	req = InstallCCRequest{Name: "ID", Version: "v0", Path: "path", Package: &api.CCPackage{Type: 1, Code: []byte("code")}}
	responses, err = rc.InstallCC(req, WithTargets(&peer1))
	if err != nil {
		t.Fatal(err)
	}

	if responses[0].Target != peer1.MockURL {
		t.Fatal("Wrong target URL set")
	}

	if strings.Contains(responses[0].Info, "already installed") {
		t.Fatal("Should not have 'already installed' info set since it was not previously installed")
	}

	// Chaincode that causes generic (system) error in installed chaincodes

	//prepare sample response

	//prepare sample response
	response = new(pb.ChaincodeQueryResponse)
	chaincodes = make([]*pb.ChaincodeInfo, 1)
	chaincodes[0] = &pb.ChaincodeInfo{Name: "name1", Path: "path1", Version: "version1"}
	response.Chaincodes = chaincodes
	responseBytes, _ = proto.Marshal(response)

	peer1 = fcmocks.MockPeer{MockName: "Peer1", MockURL: "http://peer1.com",
		Status: http.StatusOK, MockRoles: []string{}, MockCert: nil, MockMSP: "Org1MSP", Payload: responseBytes}

	req = InstallCCRequest{Name: "error", Version: "v0", Path: "", Package: &api.CCPackage{Type: 1, Code: []byte("code")}}
	_, err = rc.InstallCC(req, WithTargets(&peer1))
	if err == nil {
		t.Fatalf("Should have failed since install cc returns an error in the client")
	}
}

func TestInstallError(t *testing.T) {
	rc := setupDefaultResMgmtClient(t)

	testErr := fmt.Errorf("Test error message")
	peer1 := fcmocks.MockPeer{MockName: "Peer1", MockURL: "http://peer1.com",
		Status: http.StatusInternalServerError, MockRoles: []string{}, MockCert: nil, MockMSP: "Org1MSP", Error: testErr}

	peer2 := fcmocks.MockPeer{MockName: "Peer1", MockURL: "http://peer1.com",
		Status: http.StatusOK, MockRoles: []string{}, MockCert: nil, MockMSP: "Org1MSP"}

	req := InstallCCRequest{Name: "ID", Version: "v0", Path: "path", Package: &api.CCPackage{Type: 1, Code: []byte("code")}}
	_, err := rc.InstallCC(req, WithTargets(&peer1, &peer2))
	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), testErr.Error())
}

func TestInstallCC(t *testing.T) {
	rc := setupDefaultResMgmtClient(t)

	peer2 := fcmocks.MockPeer{MockName: "Peer1", MockURL: "http://peer1.com",
		Status: http.StatusOK, MockRoles: []string{}, MockCert: nil, MockMSP: "Org1MSP"}

	// Chaincode that is not installed already (it will be installed)
	req := InstallCCRequest{Name: "ID", Version: "v0", Path: "path", Package: &api.CCPackage{Type: 1, Code: []byte("code")}}
	responses, err := rc.InstallCC(req, WithTargets(&peer2))
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

	if responses[0].Status != http.StatusOK {
		t.Fatalf("Expecting %d status, got %d", 0, responses[0].Status)
	}

	// Chaincode that causes generic (system) error in installed chaincodes
	req = InstallCCRequest{Name: "error", Version: "v0", Path: "", Package: &api.CCPackage{Type: 1, Code: []byte("code")}}
	_, err = rc.InstallCC(req)
	if err == nil {
		t.Fatalf("Should have failed since install cc returns an error in the client")
	}
}

func TestInstallCCRequiredParameters(t *testing.T) {

	rc := setupDefaultResMgmtClient(t)

	// Test missing required parameters
	req := InstallCCRequest{}
	_, err := rc.InstallCC(req)
	if err == nil {
		t.Fatalf("Should have failed for empty install cc request")
	}

	// Test missing chaincode ID
	req = InstallCCRequest{Name: "", Version: "v0", Path: "path"}
	_, err = rc.InstallCC(req)
	if err == nil {
		t.Fatalf("Should have failed for empty cc ID")
	}

	// Test missing chaincode version
	req = InstallCCRequest{Name: "ID", Version: "", Path: "path"}
	_, err = rc.InstallCC(req)
	if err == nil {
		t.Fatalf("Should have failed for empty cc version")
	}

	// Test missing chaincode path
	req = InstallCCRequest{Name: "ID", Version: "v0", Path: ""}
	_, err = rc.InstallCC(req)
	if err == nil {
		t.Fatalf("InstallCC should have failed for empty cc path")
	}

	// Test missing chaincode package
	req = InstallCCRequest{Name: "ID", Version: "v0", Path: "path"}
	_, err = rc.InstallCC(req)
	if err == nil {
		t.Fatalf("InstallCC should have failed for nil chaincode package")
	}

	// Setup test client with different msp (default targets cannot be calculated)
	ctx := setupTestContext("test", "otherMSP")
	config := getNetworkConfig(t)
	ctx.SetEndpointConfig(config)

	// Create new resource management client ("otherMSP")
	rc = setupResMgmtClient(ctx, nil, t)
	req = InstallCCRequest{Name: "ID", Version: "v0", Path: "path", Package: &api.CCPackage{Type: 1, Code: []byte("code")}}

	// Test missing default targets
	_, err = rc.InstallCC(req)
	if err == nil {
		t.Fatalf("InstallCC should have failed with no default targets error")
	}

}

func TestInstallCCWithOptsRequiredParameters(t *testing.T) {

	rc := setupDefaultResMgmtClient(t)

	// Test missing required parameters
	req := InstallCCRequest{}
	_, err := rc.InstallCC(req)
	if err == nil {
		t.Fatalf("Should have failed for empty install cc request")
	}

	// Test missing chaincode ID
	req = InstallCCRequest{Name: "", Version: "v0", Path: "path"}
	_, err = rc.InstallCC(req)
	if err == nil {
		t.Fatalf("Should have failed for empty cc ID")
	}

	// Test missing chaincode version
	req = InstallCCRequest{Name: "ID", Version: "", Path: "path"}
	_, err = rc.InstallCC(req)
	if err == nil {
		t.Fatalf("Should have failed for empty cc version")
	}

	// Test missing chaincode path
	req = InstallCCRequest{Name: "ID", Version: "v0", Path: ""}
	_, err = rc.InstallCC(req)
	if err == nil {
		t.Fatalf("InstallCC should have failed for empty cc path")
	}

	// Test missing chaincode package
	req = InstallCCRequest{Name: "ID", Version: "v0", Path: "path"}
	_, err = rc.InstallCC(req)
	if err == nil {
		t.Fatalf("InstallCC should have failed for nil chaincode package")
	}

	// Valid request
	req = InstallCCRequest{Name: "name", Version: "version", Path: "path", Package: &api.CCPackage{Type: 1, Code: []byte("code")}}

	// Setup targets
	var peers []fab.Peer
	peers = append(peers, &fcmocks.MockPeer{MockName: "Peer1", MockURL: "http://peer1.com", MockRoles: []string{}, MockCert: nil, MockMSP: "Org1MSP"})

	// Test both targets and filter provided (error condition)
	_, err = rc.InstallCC(req, WithTargets(peers...), WithTargetFilter(&mspFilter{mspID: "Org1MSP"}))
	if err == nil {
		t.Fatalf("Should have failed if both target and filter provided")
	}

	// Setup test client with different msp
	ctx := setupTestContext("test", "otherMSP")
	config := getNetworkConfig(t)
	ctx.SetEndpointConfig(config)

	// Create new resource management client ("otherMSP")
	rc = setupResMgmtClient(ctx, nil, t)

	// No targets and no filter -- default filter msp doesn't match discovery service peer msp
	_, err = rc.InstallCC(req)
	if err == nil {
		t.Fatalf("Should have failed with no targets error")
	}

	// Test filter only provided (filter rejects discovery service peer msp)
	_, err = rc.InstallCC(req, WithTargetFilter(&mspFilter{mspID: "Org2MSP"}))
	if err == nil {
		t.Fatalf("Should have failed with no targets since filter rejected all discovery targets")
	}
}

func TestInstallCCDiscoveryError(t *testing.T) {

	// Setup test client and config
	ctx := setupTestContextWithDiscoveryError("test", "Org1MSP", errors.New("Test Error"))
	config := getNetworkConfig(t)
	ctx.SetEndpointConfig(config)

	// Create resource management client with discovery service that will generate an error
	rc := setupResMgmtClient(ctx, nil, t)

	// Test InstallCC discovery service error
	req := InstallCCRequest{Name: "ID", Version: "v0", Path: "path", Package: &api.CCPackage{Type: 1, Code: []byte("code")}}
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
	req := InstantiateCCRequest{}

	// Test empty channel name
	_, err := rc.InstantiateCC("", req)
	if err == nil {
		t.Fatalf("Should have failed for empty request")
	}

	// Test empty request
	_, err = rc.InstantiateCC("mychannel", req)
	if err == nil {
		t.Fatalf("Should have failed for empty request")
	}

	// Test missing chaincode ID
	req = InstantiateCCRequest{Name: "", Version: "v0", Path: "path"}
	_, err = rc.InstantiateCC("mychannel", req)
	if err == nil {
		t.Fatalf("Should have failed for empty cc name")
	}

	// Test missing chaincode version
	req = InstantiateCCRequest{Name: "ID", Version: "", Path: "path"}
	_, err = rc.InstantiateCC("mychannel", req)
	if err == nil {
		t.Fatalf("Should have failed for empty cc version")
	}

	// Test missing chaincode path
	req = InstantiateCCRequest{Name: "ID", Version: "v0", Path: ""}
	_, err = rc.InstantiateCC("mychannel", req)
	if err == nil {
		t.Fatalf("Should have failed for empty cc path")
	}

	// Test missing chaincode policy
	req = InstantiateCCRequest{Name: "ID", Version: "v0", Path: "path"}
	_, err = rc.InstantiateCC("mychannel", req)
	if err == nil {
		t.Fatalf("Should have failed for nil chaincode policy")
	}

	// Setup test client with different msp (default targets cannot be calculated)
	ctx := setupTestContext("test", "otherMSP")
	config := getNetworkConfig(t)
	ctx.SetEndpointConfig(config)

	// Create new resource management client ("otherMSP")
	rc = setupResMgmtClient(ctx, nil, t)

	// Valid request
	ccPolicy := cauthdsl.SignedByMspMember("otherMSP")
	req = InstantiateCCRequest{Name: "name", Version: "version", Path: "path", Policy: ccPolicy}

	// Test missing default targets
	_, err = rc.InstantiateCC("mychannel", req)
	if err == nil {
		t.Fatalf("InstallCC should have failed with no default targets error")
	}

}

func TestInstantiateCCWithOptsRequiredParameters(t *testing.T) {

	rc := setupDefaultResMgmtClient(t)

	// Test missing required parameters
	req := InstantiateCCRequest{}

	// Test empty channel name
	_, err := rc.InstantiateCC("", req)
	if err == nil {
		t.Fatalf("Should have failed for empty channel name")
	}

	// Test empty request
	_, err = rc.InstantiateCC("mychannel", req)
	if err == nil {
		t.Fatalf("Should have failed for empty instantiate cc request")
	}

	// Test missing chaincode ID
	req = InstantiateCCRequest{Name: "", Version: "v0", Path: "path"}
	_, err = rc.InstantiateCC("mychannel", req)
	if err == nil {
		t.Fatalf("Should have failed for empty cc name")
	}

	// Test missing chaincode version
	req = InstantiateCCRequest{Name: "ID", Version: "", Path: "path"}
	_, err = rc.InstantiateCC("mychannel", req)
	if err == nil {
		t.Fatalf("Should have failed for empty cc version")
	}

	// Test missing chaincode path
	req = InstantiateCCRequest{Name: "ID", Version: "v0", Path: ""}
	_, err = rc.InstantiateCC("mychannel", req)
	if err == nil {
		t.Fatalf("Should have failed for empty cc path")
	}

	// Test missing chaincode policy
	req = InstantiateCCRequest{Name: "ID", Version: "v0", Path: "path"}
	_, err = rc.InstantiateCC("mychannel", req)
	if err == nil {
		t.Fatalf("Should have failed for missing chaincode policy")
	}

	// Valid request
	ccPolicy := cauthdsl.SignedByMspMember("Org1MSP")
	req = InstantiateCCRequest{Name: "name", Version: "version", Path: "path", Policy: ccPolicy}

	// Setup targets
	var peers []fab.Peer
	peers = append(peers, &fcmocks.MockPeer{MockName: "Peer1", MockURL: "http://peer1.com", MockRoles: []string{}, MockCert: nil, MockMSP: "Org1MSP"})

	// Test both targets and filter provided (error condition)
	_, err = rc.InstantiateCC("mychannel", req, WithTargets(peers...), WithTargetFilter(&mspFilter{mspID: "Org1MSP"}))
	if err == nil {
		t.Fatalf("Should have failed if both target and filter provided")
	}

	// Setup test client with different msp
	ctx := setupTestContext("test", "otherMSP")
	config := getNetworkConfig(t)
	ctx.SetEndpointConfig(config)

	// Create new resource management client ("otherMSP")
	rc = setupResMgmtClient(ctx, nil, t)

	// No targets and no filter -- default filter msp doesn't match discovery service peer msp
	_, err = rc.InstantiateCC("mychannel", req)
	if err == nil {
		t.Fatalf("Should have failed with no targets error")
	}

	// Test filter only provided (filter rejects discovery service peer msp)
	_, err = rc.InstantiateCC("mychannel", req, WithTargetFilter(&mspFilter{mspID: "Org2MSP"}))
	if err == nil {
		t.Fatalf("Should have failed with no targets since filter rejected all discovery targets")
	}
}

func TestInstantiateCCDiscoveryError(t *testing.T) {

	// Setup test client and config
	ctx := setupTestContext("test", "Org1MSP")
	config := getNetworkConfig(t)
	ctx.SetEndpointConfig(config)

	// Create resource management client with discovery service that will generate an error
	rc := setupResMgmtClient(ctx, errors.New("Test Error"), t)

	orderer := fcmocks.NewMockOrderer("", nil)

	transactor := txnmocks.MockTransactor{
		Ctx:       ctx,
		ChannelID: "mychannel",
		Orderers:  []fab.Orderer{orderer},
	}
	rc.ctx.InfraProvider().(*fcmocks.MockInfraProvider).SetCustomTransactor(&transactor)

	// Start mock event hub
	eventServer, err := fcmocks.StartMockEventServer(fmt.Sprintf("%s:%d", "127.0.0.1", 7053))
	if err != nil {
		t.Fatalf("Failed to start mock event hub: %v", err)
	}
	defer eventServer.Stop()

	ccPolicy := cauthdsl.SignedByMspMember("Org1MSP")
	req := InstantiateCCRequest{Name: "name", Version: "version", Path: "path", Policy: ccPolicy}

	// Test InstantiateCC create new discovery service per channel error
	_, err = rc.InstantiateCC("error", req)
	if err == nil {
		t.Fatalf("Should have failed to instantiate cc with create discovery service error")
	}

	// Test InstantiateCC discovery service get peers error
	_, err = rc.InstantiateCC("mychannel", req)
	if err == nil {
		t.Fatalf("Should have failed to instantiate cc with get peers discovery error")
	}

	// Test InstantiateCCWithOpts create new discovery service per channel error
	_, err = rc.InstantiateCC("error", req)
	if err == nil {
		t.Fatalf("Should have failed to instantiate cc with opts with create discovery service error")
	}

	// Test InstantiateCCWithOpts discovery service get peers error
	// if targets are not provided discovery service is used
	_, err = rc.InstantiateCC("mychannel", req)
	if err == nil {
		t.Fatalf("Should have failed to instantiate cc with opts with get peers discovery error")
	}

}

func TestUpgradeCCRequiredParameters(t *testing.T) {

	rc := setupDefaultResMgmtClient(t)

	// Test missing required parameters
	req := UpgradeCCRequest{}

	// Test empty channel name
	_, err := rc.UpgradeCC("", req)
	if err == nil {
		t.Fatalf("Should have failed for empty channel name")
	}

	// Test empty request
	_, err = rc.UpgradeCC("mychannel", req)
	if err == nil {
		t.Fatalf("Should have failed for empty upgrade cc request")
	}

	// Test missing chaincode ID
	req = UpgradeCCRequest{Name: "", Version: "v0", Path: "path"}
	_, err = rc.UpgradeCC("mychannel", req)
	if err == nil {
		t.Fatalf("Should have failed for empty cc name")
	}

	// Test missing chaincode version
	req = UpgradeCCRequest{Name: "ID", Version: "", Path: "path"}
	_, err = rc.UpgradeCC("mychannel", req)
	if err == nil {
		t.Fatalf("Should have failed for empty cc version")
	}

	// Test missing chaincode path
	req = UpgradeCCRequest{Name: "ID", Version: "v0", Path: ""}
	_, err = rc.UpgradeCC("mychannel", req)
	if err == nil {
		t.Fatalf("Should have failed for empty cc path")
	}

	// Test missing chaincode policy
	req = UpgradeCCRequest{Name: "ID", Version: "v0", Path: "path"}
	_, err = rc.UpgradeCC("mychannel", req)
	if err == nil {
		t.Fatalf("Should have failed for nil chaincode policy")
	}

	// Setup test client with different msp (default targets cannot be calculated)
	ctx := setupTestContext("test", "otherMSP")
	config := getNetworkConfig(t)
	ctx.SetEndpointConfig(config)

	// Create new resource management client ("otherMSP")
	rc = setupResMgmtClient(ctx, nil, t)

	// Valid request
	ccPolicy := cauthdsl.SignedByMspMember("otherMSP")
	req = UpgradeCCRequest{Name: "name", Version: "version", Path: "path", Policy: ccPolicy}

	// Test missing default targets
	_, err = rc.UpgradeCC("mychannel", req)
	if err == nil {
		t.Fatalf("Should have failed with no default targets error")
	}

}

func TestUpgradeCCWithOptsRequiredParameters(t *testing.T) {

	rc := setupDefaultResMgmtClient(t)

	// Test missing required parameters
	req := UpgradeCCRequest{}

	// Test empty channel name
	_, err := rc.UpgradeCC("", req)
	if err == nil {
		t.Fatalf("Should have failed for empty channel name")
	}

	// Test empty request
	_, err = rc.UpgradeCC("mychannel", req)
	if err == nil {
		t.Fatalf("Should have failed for empty upgrade cc request")
	}

	// Test missing chaincode ID
	req = UpgradeCCRequest{Name: "", Version: "v0", Path: "path"}
	_, err = rc.UpgradeCC("mychannel", req)
	if err == nil {
		t.Fatalf("Should have failed for empty cc name")
	}

	// Test missing chaincode version
	req = UpgradeCCRequest{Name: "ID", Version: "", Path: "path"}
	_, err = rc.UpgradeCC("mychannel", req)
	if err == nil {
		t.Fatalf("Should have failed for empty cc version")
	}

	// Test missing chaincode path
	req = UpgradeCCRequest{Name: "ID", Version: "v0", Path: ""}
	_, err = rc.UpgradeCC("mychannel", req)
	if err == nil {
		t.Fatalf("UpgradeCC should have failed for empty cc path")
	}

	// Test missing chaincode policy
	req = UpgradeCCRequest{Name: "ID", Version: "v0", Path: "path"}
	_, err = rc.UpgradeCC("mychannel", req)
	if err == nil {
		t.Fatalf("UpgradeCC should have failed for missing chaincode policy")
	}

	// Valid request
	ccPolicy := cauthdsl.SignedByMspMember("Org1MSP")
	req = UpgradeCCRequest{Name: "name", Version: "version", Path: "path", Policy: ccPolicy}

	// Setup targets
	var peers []fab.Peer
	peers = append(peers, &fcmocks.MockPeer{MockName: "Peer1", MockURL: "http://peer1.com", MockRoles: []string{}, MockCert: nil, MockMSP: "Org1MSP"})

	// Test both targets and filter provided (error condition)
	_, err = rc.UpgradeCC("mychannel", req, WithTargets(peers...), WithTargetFilter(&mspFilter{mspID: "Org1MSP"}))
	if err == nil {
		t.Fatalf("Should have failed if both target and filter provided")
	}

	// Setup test client with different msp
	ctx := setupTestContext("test", "otherMSP")
	config := getNetworkConfig(t)
	ctx.SetEndpointConfig(config)

	// Create new resource management client ("otherMSP")
	rc = setupResMgmtClient(ctx, nil, t)

	// No targets and no filter -- default filter msp doesn't match discovery service peer msp
	_, err = rc.UpgradeCC("mychannel", req)
	if err == nil {
		t.Fatalf("Should have failed with no targets error")
	}

	// Test filter only provided (filter rejects discovery service peer msp)
	_, err = rc.UpgradeCC("mychannel", req, WithTargetFilter(&mspFilter{mspID: "Org2MSP"}))
	if err == nil {
		t.Fatalf("Should have failed with no targets since filter rejected all discovery targets")
	}
}

func TestUpgradeCCDiscoveryError(t *testing.T) {

	// Setup test client and config
	ctx := setupTestContextWithDiscoveryError("test", "Org1MSP", nil)
	config := getNetworkConfig(t)
	ctx.SetEndpointConfig(config)

	// Create resource management client with discovery service that will generate an error
	rc := setupResMgmtClient(ctx, errors.New("Test Error"), t)
	orderer := fcmocks.NewMockOrderer("", nil)

	transactor := txnmocks.MockTransactor{
		Ctx:       ctx,
		ChannelID: "mychannel",
		Orderers:  []fab.Orderer{orderer},
	}
	rc.ctx.InfraProvider().(*fcmocks.MockInfraProvider).SetCustomTransactor(&transactor)

	// Start mock event hub
	eventServer, err := fcmocks.StartMockEventServer(fmt.Sprintf("%s:%d", "127.0.0.1", 7053))
	if err != nil {
		t.Fatalf("Failed to start mock event hub: %v", err)
	}
	defer eventServer.Stop()

	// Test UpgradeCC discovery service error
	ccPolicy := cauthdsl.SignedByMspMember("Org1MSP")
	req := UpgradeCCRequest{Name: "name", Version: "version", Path: "path", Policy: ccPolicy}

	// Test error while creating discovery service for channel "error"
	_, err = rc.UpgradeCC("error", req)
	if err == nil {
		t.Fatalf("Should have failed to upgrade cc with discovery error")
	}

	// Test error in discovery service while getting peers
	_, err = rc.UpgradeCC("mychannel", req)
	if err == nil {
		t.Fatalf("Should have failed to upgrade cc with discovery error")
	}

	// Test UpgradeCCWithOpts discovery service error when creating discovery service for channel 'error'
	_, err = rc.UpgradeCC("error", req)
	if err == nil {
		t.Fatalf("Should have failed to upgrade cc with opts with discovery error")
	}

	// Test UpgradeCCWithOpts discovery service error
	// if targets are not provided discovery service is used to get targets
	_, err = rc.UpgradeCC("mychannel", req)
	if err == nil {
		t.Fatalf("Should have failed to upgrade cc with opts with discovery error")
	}

}

func TestCCProposal(t *testing.T) {

	ctx := setupTestContext("Admin", "Org1MSP")

	// Setup resource management client
	configBackend, err := configImpl.FromFile("./testdata/ccproposal_test.yaml")()
	if err != nil {
		t.Fatal(err)
	}
	cfg, err := fabImpl.ConfigFromBackend(configBackend)
	if err != nil {
		t.Fatal(err)
	}
	ctx.SetEndpointConfig(cfg)

	// Setup target peers
	var peers []fab.Peer
	peer1, _ := peer.New(fcmocks.NewMockEndpointConfig(), peer.WithURL("127.0.0.1:0"))
	peers = append(peers, peer1)

	// Create mock orderer
	orderer := fcmocks.NewMockOrderer("", nil)
	rc := setupResMgmtClient(ctx, nil, t)

	transactor := txnmocks.MockTransactor{
		Ctx:       ctx,
		ChannelID: "mychannel",
		Orderers:  []fab.Orderer{orderer},
	}
	rc.ctx.InfraProvider().(*fcmocks.MockInfraProvider).SetCustomTransactor(&transactor)

	ccPolicy := cauthdsl.SignedByMspMember("Org1MSP")
	instantiateReq := InstantiateCCRequest{Name: "name", Version: "version", Path: "path", Policy: ccPolicy}

	// Test error connecting to event hub
	_, err = rc.InstantiateCC("mychannel", instantiateReq)
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
	_, err = rc.InstantiateCC("mychannel", instantiateReq)
	if err == nil {
		t.Fatalf("Should have failed due to error in commit")
	}

	// Test invalid function (only 'instatiate' and 'upgrade' are supported)
	reqCtx, cancel := contextImpl.NewRequest(rc.ctx, contextImpl.WithTimeout(10*time.Second))
	defer cancel()
	opts := requestOptions{Targets: peers}
	_, err = rc.sendCCProposal(reqCtx, 3, "mychannel", instantiateReq, opts)
	if err == nil {
		t.Fatalf("Should have failed for invalid function name")
	}

	// Test no event source in config
	configBackend, err = configImpl.FromFile("./testdata/event_source_missing_test.yaml")()
	if err != nil {
		t.Fatal(err)
	}
	cfg, err = fabImpl.ConfigFromBackend(configBackend)
	if err != nil {
		t.Fatal(err)
	}
	ctx.SetEndpointConfig(cfg)
	rc = setupResMgmtClient(ctx, nil, t, getDefaultTargetFilterOption())
	_, err = rc.InstantiateCC("mychannel", instantiateReq)
	if err == nil {
		t.Fatalf("Should have failed since no event source has been configured")
	}
}

func TestCCProposalFailed(t *testing.T) {
	grpcServer := grpc.NewServer()
	defer grpcServer.Stop()

	// Setup mock targets
	endorserServer, address := startEndorserServer(t, grpcServer)
	time.Sleep(2 * time.Second)

	ctx := setupTestContext("Admin", "Org1MSP")

	// Setup resource management client
	configBackend, err := configImpl.FromFile("./testdata/ccproposal_test.yaml")()
	if err != nil {
		t.Fatal(err)
	}

	cfg, err := fabImpl.ConfigFromBackend(configBackend)
	if err != nil {
		t.Fatal(err)
	}
	ctx.SetEndpointConfig(cfg)

	rc := setupResMgmtClient(ctx, nil, t)

	// Setup target peers
	var peers []fab.Peer
	peer1, _ := peer.New(fcmocks.NewMockEndpointConfig(), peer.WithURL(address))
	peers = append(peers, peer1)

	ccPolicy := cauthdsl.SignedByMspMember("Org1MSP")
	instantiateReq := InstantiateCCRequest{Name: "name", Version: "version", Path: "path", Policy: ccPolicy}

	// Test failed proposal error handling (endorser returns an error)
	endorserServer.ProposalError = errors.New("Test Error")

	_, err = rc.InstantiateCC("mychannel", instantiateReq, WithTargets(peers...))
	if err == nil {
		t.Fatalf("Should have failed to instantiate cc due to endorser error")
	}

	upgradeRequest := UpgradeCCRequest{Name: "name", Version: "version", Path: "path", Policy: ccPolicy}
	_, err = rc.UpgradeCC("mychannel", upgradeRequest, WithTargets(peers...))
	if err == nil {
		t.Fatalf("Should have failed to upgrade cc due to endorser error")
	}
	// Remove endorser error
	endorserServer.ProposalError = nil

}

func getDefaultTargetFilterOption() ClientOption {
	targetFilter := &mspFilter{mspID: "Org1MSP"}
	return WithDefaultTargetFilter(targetFilter)
}

func setupTestDiscovery(discErr error, peers []fab.Peer) (fab.DiscoveryProvider, error) {

	mockDiscovery, err := txnmocks.NewMockDiscoveryProvider(discErr, peers)
	if err != nil {
		return nil, errors.WithMessage(err, "NewMockDiscoveryProvider failed")
	}

	return mockDiscovery, nil
}

func setupDefaultResMgmtClient(t *testing.T) *Client {
	ctx := setupTestContext("test", "Org1MSP")
	network := getNetworkConfig(t)
	ctx.SetEndpointConfig(network)
	return setupResMgmtClient(ctx, nil, t, getDefaultTargetFilterOption())
}

func setupResMgmtClient(fabCtx context.Client, discErr error, t *testing.T, opts ...ClientOption) *Client {

	ctx := createClientContext(fabCtx)

	resClient, err := New(ctx, opts...)
	if err != nil {
		t.Fatalf("Failed to create new client with options: %s %v", err, opts)
	}

	return resClient

}

func setupTestContext(username string, mspID string) *fcmocks.MockContext {
	user := mspmocks.NewMockSigningIdentity(username, mspID)
	ctx := fcmocks.NewMockContext(user)
	return ctx
}

func setupCustomOrderer(ctx *fcmocks.MockContext, mockOrderer fab.Orderer) *fcmocks.MockContext {
	mockInfraProvider := &fcmocks.MockInfraProvider{}
	mockInfraProvider.SetCustomOrderer(mockOrderer)
	ctx.SetCustomInfraProvider(mockInfraProvider)
	return ctx
}

func setupTestContextWithDiscoveryError(username string, mspID string, discErr error) *fcmocks.MockContext {
	user := mspmocks.NewMockSigningIdentity(username, mspID)
	dscPvdr, _ := setupTestDiscovery(discErr, nil)
	//ignore err and set whatever you get in dscPvdr
	ctx := fcmocks.NewMockContextWithCustomDiscovery(user, dscPvdr)
	return ctx
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

func getNetworkConfig(t *testing.T) fab.EndpointConfig {
	configBackend, err := configImpl.FromFile(networkCfg)()
	if err != nil {
		t.Fatal(err)
	}

	config, err := fabImpl.ConfigFromBackend(configBackend)
	if err != nil {
		t.Fatal(err)
	}

	return config
}

func TestSaveChannelSuccess(t *testing.T) {

	grpcServer := grpc.NewServer()
	defer grpcServer.Stop()
	_, addr := fcmocks.StartMockBroadcastServer("127.0.0.1:0", grpcServer)

	ctx := setupTestContext("test", "Org1MSP")

	mockConfig := &fcmocks.MockConfig{}
	grpcOpts := make(map[string]interface{})
	grpcOpts["allow-insecure"] = true

	oConfig := &fab.OrdererConfig{
		URL:         addr,
		GRPCOptions: grpcOpts,
	}
	mockConfig.SetCustomOrdererCfg(oConfig)
	ctx.SetEndpointConfig(mockConfig)

	cc := setupResMgmtClient(ctx, nil, t)

	// Test empty channel request
	_, err := cc.SaveChannel(SaveChannelRequest{})
	assert.NotNil(t, err, "Should have failed for empty channel request")

	r, err := os.Open(channelConfig)
	assert.Nil(t, err, "opening channel config file failed")
	defer r.Close()

	// Test empty channel name
	_, err = cc.SaveChannel(SaveChannelRequest{ChannelID: "", ChannelConfig: r})
	assert.NotNil(t, err, "Should have failed for empty channel id")

	// Test empty channel config
	_, err = cc.SaveChannel(SaveChannelRequest{ChannelID: "mychannel"})
	assert.NotNil(t, err, "Should have failed for empty channel config")

	// Test extract configuration error
	r1, err := os.Open("./testdata/extractcherr.tx")
	assert.Nil(t, err, "opening channel config file failed")
	defer r1.Close()

	_, err = cc.SaveChannel(SaveChannelRequest{ChannelID: "mychannel", ChannelConfig: r1})
	assert.NotNil(t, err, "Should have failed to extract configuration")

	// Test sign channel error
	r2, err := os.Open("./testdata/signcherr.tx")
	assert.Nil(t, err, "opening channel config file failed")
	defer r2.Close()

	_, err = cc.SaveChannel(SaveChannelRequest{ChannelID: "mychannel", ChannelConfig: r2})
	assert.NotNil(t, err, "Should have failed to sign configuration")

	// Test valid Save Channel request (success)
	resp, err := cc.SaveChannel(SaveChannelRequest{ChannelID: "mychannel", ChannelConfig: r}, WithOrdererURL("example.com"))
	assert.Nil(t, err, "error should be nil")
	assert.NotEmpty(t, resp.TransactionID, "transaction ID should be populated")

	// Test valid Save Channel request (success / filename)
	resp, err = cc.SaveChannel(SaveChannelRequest{ChannelID: "mychannel", ChannelConfigPath: channelConfig}, WithOrdererURL("example.com"))
	assert.Nil(t, err, "error should be nil")
	assert.NotEmpty(t, resp.TransactionID, "transaction ID should be populated")
}

func TestSaveChannelFailure(t *testing.T) {

	// Set up context with error in create channel
	user := mspmocks.NewMockSigningIdentity("test", "test")
	errCtx := fcmocks.NewMockContext(user)
	network := getNetworkConfig(t)
	errCtx.SetEndpointConfig(network)
	fabCtx := setupTestContext("user", "Org1Msp1")

	clientCtx := createClientContext(contextImpl.Client{
		Providers:       fabCtx,
		SigningIdentity: fabCtx,
	})

	cc, err := New(clientCtx)
	if err != nil {
		t.Fatalf("Failed to create new channel management client: %s", err)
	}

	// Test create channel failure
	r, err := os.Open(channelConfig)
	assert.Nil(t, err, "opening channel config file failed")
	defer r.Close()

	_, err = cc.SaveChannel(SaveChannelRequest{ChannelID: "Invalid", ChannelConfig: r})
	assert.NotNil(t, err, "Should have failed with create channel error")
}

func TestSaveChannelWithOpts(t *testing.T) {

	grpcServer := grpc.NewServer()
	defer grpcServer.Stop()
	_, addr := fcmocks.StartMockBroadcastServer("127.0.0.1:0", grpcServer)

	ctx := setupTestContext("test", "Org1MSP")

	mockConfig := &fcmocks.MockConfig{}
	grpcOpts := make(map[string]interface{})
	grpcOpts["allow-insecure"] = true

	oConfig := &fab.OrdererConfig{
		URL:         addr,
		GRPCOptions: grpcOpts,
	}
	mockConfig.SetCustomOrdererCfg(oConfig)
	mockConfig.SetCustomRandomOrdererCfg(oConfig)
	ctx.SetEndpointConfig(mockConfig)

	cc := setupResMgmtClient(ctx, nil, t)

	// Valid request (same for all options)
	r1, err := os.Open(channelConfig)
	assert.Nil(t, err, "opening channel config file failed")
	defer r1.Close()

	req := SaveChannelRequest{ChannelID: "mychannel", ChannelConfig: r1}

	// Test empty option (default order is random orderer from config)
	opts := WithOrdererURL("")
	resp, err := cc.SaveChannel(req, opts)
	assert.Nil(t, err, "error should be nil")
	assert.NotEmpty(t, resp.TransactionID, "transaction ID should be populated")

	// Test valid orderer ID
	r2, err := os.Open(channelConfig)
	assert.Nil(t, err, "opening channel config file failed")
	defer r2.Close()

	req = SaveChannelRequest{ChannelID: "mychannel", ChannelConfig: r2}

	opts = WithOrdererURL("orderer.example.com")
	resp, err = cc.SaveChannel(req, opts)
	assert.Nil(t, err, "error should be nil")
	assert.NotEmpty(t, resp.TransactionID, "transaction ID should be populated")

	// Test invalid orderer ID
	r3, err := os.Open(channelConfig)
	assert.Nil(t, err, "opening channel config file failed")
	defer r3.Close()

	req = SaveChannelRequest{ChannelID: "mychannel", ChannelConfig: r3}

	mockConfig = &fcmocks.MockConfig{}
	ctx.SetEndpointConfig(mockConfig)

	cc = setupResMgmtClient(ctx, nil, t)

	opts = WithOrdererURL("Invalid")
	_, err = cc.SaveChannel(req, opts)
	assert.NotNil(t, err, "Should have failed for invalid orderer ID")
}

func TestJoinChannelWithInvalidOpts(t *testing.T) {

	cc := setupDefaultResMgmtClient(t)
	opts := WithOrdererURL("Invalid")
	err := cc.JoinChannel("mychannel", opts)
	if err == nil {
		t.Fatal("Should have failed for invalid orderer ID")
	}

}

func TestSaveChannelWithMultipleSigningIdenities(t *testing.T) {
	grpcServer := grpc.NewServer()
	defer grpcServer.Stop()
	_, addr := fcmocks.StartMockBroadcastServer("127.0.0.1:0", grpcServer)
	ctx := setupTestContext("test", "Org1MSP")

	mockConfig := &fcmocks.MockConfig{}
	grpcOpts := make(map[string]interface{})
	grpcOpts["allow-insecure"] = true

	oConfig := &fab.OrdererConfig{
		URL:         addr,
		GRPCOptions: grpcOpts,
	}
	mockConfig.SetCustomRandomOrdererCfg(oConfig)
	mockConfig.SetCustomOrdererCfg(oConfig)
	ctx.SetEndpointConfig(mockConfig)

	cc := setupResMgmtClient(ctx, nil, t)

	// empty list of signing identities (defaults to context user)
	r1, err := os.Open(channelConfig)
	assert.Nil(t, err, "opening channel config file failed")
	defer r1.Close()

	req := SaveChannelRequest{ChannelID: "mychannel", ChannelConfig: r1, SigningIdentities: []msp.SigningIdentity{}}
	resp, err := cc.SaveChannel(req, WithOrdererURL(""))
	assert.Nil(t, err, "Failed to save channel with default signing identity: %s", err)
	assert.NotEmpty(t, resp.TransactionID, "transaction ID should be populated")

	// multiple signing identities
	r2, err := os.Open(channelConfig)
	assert.Nil(t, err, "opening channel config file failed")
	defer r2.Close()

	secondCtx := fcmocks.NewMockContext(mspmocks.NewMockSigningIdentity("second", "second"))
	req = SaveChannelRequest{ChannelID: "mychannel", ChannelConfig: r2, SigningIdentities: []msp.SigningIdentity{cc.ctx, secondCtx}}
	resp, err = cc.SaveChannel(req, WithOrdererURL(""))
	assert.Nil(t, err, "Failed to save channel with multiple signing identities: %s", err)
	assert.NotEmpty(t, resp.TransactionID, "transaction ID should be populated")
}

func createClientContext(fabCtx context.Client) context.ClientProvider {
	return func() (context.Client, error) {
		return fabCtx, nil
	}
}
