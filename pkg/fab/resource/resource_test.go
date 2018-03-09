/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package resource

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net"
	"path"
	"testing"
	"time"

	"github.com/hyperledger/fabric-sdk-go/pkg/common/context"
	"github.com/hyperledger/fabric-sdk-go/pkg/context/api/fab"
	"github.com/hyperledger/fabric-sdk-go/pkg/fab/mocks"
	"github.com/hyperledger/fabric-sdk-go/pkg/fab/peer"
	"github.com/hyperledger/fabric-sdk-go/pkg/fab/resource/api"
	"github.com/hyperledger/fabric-sdk-go/test/metadata"
	"github.com/hyperledger/fabric-sdk-go/third_party/github.com/hyperledger/fabric/protos/common"
	pb "github.com/hyperledger/fabric-sdk-go/third_party/github.com/hyperledger/fabric/protos/peer"
	"github.com/pkg/errors"
	"google.golang.org/grpc"
)

func TestSignChannelConfig(t *testing.T) {
	ctx := setupContext()

	configTx, err := ioutil.ReadFile(path.Join("../../../", metadata.ChannelConfigPath, "mychannel.tx"))
	if err != nil {
		t.Fatalf(err.Error())
	}

	_, err = SignChannelConfig(ctx, nil, nil)
	if err == nil {
		t.Fatalf("Expected 'channel configuration required")
	}

	_, err = SignChannelConfig(ctx, configTx, nil)
	if err != nil {
		t.Fatalf("Expected 'channel configuration required %v", err)
	}
}

func TestCreateChannel(t *testing.T) {
	ctx := setupContext()

	configTx, err := ioutil.ReadFile(path.Join("../../../", metadata.ChannelConfigPath, "mychannel.tx"))
	if err != nil {
		t.Fatalf(err.Error())
	}

	// Setup mock orderer
	verifyBroadcast := make(chan *fab.SignedEnvelope)
	orderer := mocks.NewMockOrderer(fmt.Sprintf("0.0.0.0:1234"), verifyBroadcast)

	// Create channel without envelope
	_, err = CreateChannel(ctx, api.CreateChannelRequest{
		Orderer: orderer,
		Name:    "mychannel",
	})
	if err == nil {
		t.Fatalf("Expected error creating channel without envelope")
	}

	// Create channel without orderer
	_, err = CreateChannel(ctx, api.CreateChannelRequest{
		Envelope: configTx,
		Name:     "mychannel",
	})
	if err == nil {
		t.Fatalf("Expected error creating channel without orderer")
	}

	// Create channel without name
	_, err = CreateChannel(ctx, api.CreateChannelRequest{
		Envelope: configTx,
		Orderer:  orderer,
	})
	if err == nil {
		t.Fatalf("Expected error creating channel without name")
	}

	// Test with valid cofiguration
	request := api.CreateChannelRequest{
		Envelope: configTx,
		Orderer:  orderer,
		Name:     "mychannel",
	}
	_, err = CreateChannel(ctx, request)
	if err != nil {
		t.Fatalf("Did not expect error from create channel. Got error: %v", err)
	}
	select {
	case b := <-verifyBroadcast:
		logger.Debugf("Verified broadcast: %v", b)
	case <-time.After(time.Second):
		t.Fatalf("Expected broadcast")
	}
}

func TestJoinChannel(t *testing.T) {
	var peers []fab.ProposalProcessor

	grpcServer := grpc.NewServer()
	defer grpcServer.Stop()

	endorserServer, addr := startEndorserServer(t, grpcServer)
	peer, _ := peer.New(mocks.NewMockConfig(), peer.WithURL("grpc://"+addr), peer.WithInsecure())
	peers = append(peers, peer)

	orderer := mocks.NewMockOrderer("", nil)
	defer orderer.Close()
	orderer.EnqueueForSendDeliver(mocks.NewSimpleMockBlock())
	orderer.EnqueueForSendDeliver(common.Status_SUCCESS)

	ctx := setupContext()

	genesisBlock := mocks.NewSimpleMockBlock()

	request := api.JoinChannelRequest{
		Targets: peers,
		//GenesisBlock: genesisBlock,
	}
	err := JoinChannel(ctx, request)
	if err == nil {
		t.Fatalf("Should not have been able to join channel because of missing GenesisBlock parameter")
	}

	request = api.JoinChannelRequest{
		Targets:      peers,
		GenesisBlock: genesisBlock,
	}
	if err == nil {
		t.Fatalf("Should not have been able to join channel because of invalid targets")
	}

	// Test join channel with valid arguments
	err = JoinChannel(ctx, request)
	if err != nil {
		t.Fatalf("Did not expect error from join channel. Got: %s", err)
	}

	// Test failed proposal error handling
	endorserServer.ProposalError = errors.New("Test Error")
	request = api.JoinChannelRequest{
		Targets: peers,
	}
	err = JoinChannel(ctx, request)
	if err == nil {
		t.Fatalf("Expected error")
	}
}

func setupContext() context.Client {
	user := mocks.NewMockUser("test")
	ctx := mocks.NewMockContext(user)
	return ctx
}

func TestQueryByChaincode(t *testing.T) {
	ctx := setupContext()

	peer := mocks.MockPeer{MockName: "Peer1", MockURL: "peer1.example.com", MockRoles: []string{}, MockCert: nil, Payload: []byte("A"), Status: 200}

	request := fab.ChaincodeInvokeRequest{
		ChaincodeID: "cc",
		Fcn:         "Hello",
	}
	resp, err := queryChaincodeWithTarget(ctx, request, &peer)
	if err != nil {
		t.Fatalf("Failed to query: %s", err)
	}
	expectedResp := []byte("A")

	if !bytes.Equal(resp, expectedResp) {
		t.Fatalf("Unexpected transaction proposal response: %v (expected %v)", resp, expectedResp)
	}
}

func TestQueryByChaincodeBadStatus(t *testing.T) {
	ctx := setupContext()

	peer := mocks.MockPeer{MockName: "Peer1", MockURL: "http://peer1.com", MockRoles: []string{}, MockCert: nil, Payload: []byte("A"), Status: 99}

	request := fab.ChaincodeInvokeRequest{
		ChaincodeID: "cc",
		Fcn:         "Hello",
	}
	_, err := queryChaincodeWithTarget(ctx, request, &peer)
	if err == nil {
		t.Fatalf("expected failure due to bad status")
	}
}

func TestQueryByChaincodeError(t *testing.T) {
	ctx := setupContext()

	peer := mocks.MockPeer{MockName: "Peer1", MockURL: "http://peer1.com", MockRoles: []string{}, MockCert: nil, Payload: []byte("A"), Error: errors.New("error")}

	request := fab.ChaincodeInvokeRequest{
		ChaincodeID: "cc",
		Fcn:         "Hello",
	}
	_, err := queryChaincodeWithTarget(ctx, request, &peer)
	if err == nil {
		t.Fatalf("expected failure due to error")
	}
}

func TestGenesisBlockOrdererErr(t *testing.T) {
	const channelName = "testchannel"
	ctx := setupContext()

	orderer := mocks.NewMockOrderer("", nil)
	defer orderer.Close()
	orderer.EnqueueForSendDeliver(mocks.NewSimpleMockError())

	_, err := GenesisBlockFromOrderer(ctx, channelName, orderer)

	if err == nil {
		t.Fatal("GenesisBlock Test supposed to fail with error")
	}
}

func TestGenesisBlock(t *testing.T) {
	const channelName = "testchannel"
	ctx := setupContext()

	orderer := mocks.NewMockOrderer("", nil)
	defer orderer.Close()
	orderer.EnqueueForSendDeliver(mocks.NewSimpleMockBlock())
	orderer.EnqueueForSendDeliver(common.Status_SUCCESS)

	_, err := GenesisBlockFromOrderer(ctx, channelName, orderer)

	if err != nil {
		t.Fatalf("GenesisBlock failed: %s", err)
	}
}

/*
// TODO - make a much shorter timeout for this test.
func TestGenesisBlockOrdererTimeout(t *testing.T) {
	const channelName = "testchannel"

	ctx := setupContext()
	orderer := mocks.NewMockOrderer("", nil)

	_, err := GenesisBlockFromOrderer(ctx, channelName, orderer)

	//It should fail with timeout
	if err == nil || !strings.HasSuffix(err.Error(), "timeout waiting for response from orderer") {
		t.Fatal("GenesisBlock Test supposed to fail with timeout error")
	}
}*/

func TestGenesisBlockOrderer(t *testing.T) {
	const channelName = "testchannel"
	ctx := setupContext()

	orderer := mocks.NewMockOrderer("", nil)
	defer orderer.Close()
	orderer.EnqueueForSendDeliver(mocks.NewSimpleMockError())

	//Call get Genesis block
	_, err := GenesisBlockFromOrderer(ctx, channelName, orderer)

	//Expecting error
	if err == nil {
		t.Fatal("GenesisBlock Test supposed to fail with error")
	}
}

const testAddress = "127.0.0.1:0"

func startEndorserServer(t *testing.T, grpcServer *grpc.Server) (*mocks.MockEndorserServer, string) {
	lis, err := net.Listen("tcp", testAddress)
	addr := lis.Addr().String()

	endorserServer := &mocks.MockEndorserServer{}
	pb.RegisterEndorserServer(grpcServer, endorserServer)
	if err != nil {
		t.Logf("Error starting test server %s", err)
		t.FailNow()
	}
	t.Logf("Starting test server on %s\n", addr)
	go grpcServer.Serve(lis)
	return endorserServer, addr
}
