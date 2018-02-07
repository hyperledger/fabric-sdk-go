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

	fab "github.com/hyperledger/fabric-sdk-go/api/apifabclient"
	"github.com/hyperledger/fabric-sdk-go/pkg/fabric-client/mocks"
	"github.com/hyperledger/fabric-sdk-go/pkg/fabric-client/peer"
	"github.com/hyperledger/fabric-sdk-go/test/metadata"
	pb "github.com/hyperledger/fabric-sdk-go/third_party/github.com/hyperledger/fabric/protos/peer"
	"github.com/pkg/errors"
	"google.golang.org/grpc"
)

func TestCreateChannel(t *testing.T) {
	client := setupTestClient()

	configTx, err := ioutil.ReadFile(path.Join("../../../", metadata.ChannelConfigPath, "mychannel.tx"))
	if err != nil {
		t.Fatalf(err.Error())
	}

	// Setup mock orderer
	verifyBroadcast := make(chan *fab.SignedEnvelope)
	orderer := mocks.NewMockOrderer(fmt.Sprintf("0.0.0.0:1234"), verifyBroadcast)

	// Create channel without envelope
	_, err = client.CreateChannel(fab.CreateChannelRequest{
		Orderer: orderer,
		Name:    "mychannel",
	})
	if err == nil {
		t.Fatalf("Expected error creating channel without envelope")
	}

	// Create channel without orderer
	_, err = client.CreateChannel(fab.CreateChannelRequest{
		Envelope: configTx,
		Name:     "mychannel",
	})
	if err == nil {
		t.Fatalf("Expected error creating channel without orderer")
	}

	// Create channel without name
	_, err = client.CreateChannel(fab.CreateChannelRequest{
		Envelope: configTx,
		Orderer:  orderer,
	})
	if err == nil {
		t.Fatalf("Expected error creating channel without name")
	}

	// Test with valid cofiguration
	request := fab.CreateChannelRequest{
		Envelope: configTx,
		Orderer:  orderer,
		Name:     "mychannel",
	}
	_, err = client.CreateChannel(request)
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
	peer, _ := peer.New(mocks.NewMockConfig(), peer.WithURL(addr))
	peers = append(peers, peer)
	orderer := mocks.NewMockOrderer("", nil)
	orderer.(mocks.MockOrderer).EnqueueForSendDeliver(mocks.NewSimpleMockBlock())

	client := setupTestClient()

	genesisBlock := mocks.NewSimpleMockBlock()

	request := fab.JoinChannelRequest{
		Targets: peers,
		//GenesisBlock: genesisBlock,
	}
	err := client.JoinChannel(request)
	if err == nil {
		t.Fatalf("Should not have been able to join channel because of missing GenesisBlock parameter")
	}

	request = fab.JoinChannelRequest{
		Targets:      peers,
		GenesisBlock: genesisBlock,
	}
	if err == nil {
		t.Fatalf("Should not have been able to join channel because of invalid targets")
	}

	// Test join channel with valid arguments
	err = client.JoinChannel(request)
	if err != nil {
		t.Fatalf("Did not expect error from join channel. Got: %s", err)
	}

	// Test failed proposal error handling
	endorserServer.ProposalError = errors.New("Test Error")
	request = fab.JoinChannelRequest{
		Targets: peers,
	}
	err = client.JoinChannel(request)
	if err == nil {
		t.Fatalf("Expected error")
	}
}

func setupTestClient() *Resource {
	user := mocks.NewMockUser("test")
	ctx := mocks.NewMockContext(user)
	return New(ctx)
}

func TestQueryByChaincode(t *testing.T) {
	c := setupTestClient()

	peer := mocks.MockPeer{MockName: "Peer1", MockURL: "peer1.example.com", MockRoles: []string{}, MockCert: nil, Payload: []byte("A"), Status: 200}

	request := fab.ChaincodeInvokeRequest{
		ChaincodeID: "cc",
		Fcn:         "Hello",
	}
	resp, err := c.queryChaincodeWithTarget(request, &peer)
	if err != nil {
		t.Fatalf("Failed to query: %s", err)
	}
	expectedResp := []byte("A")

	if !bytes.Equal(resp, expectedResp) {
		t.Fatalf("Unexpected transaction proposal response: %v (expected %v)", resp, expectedResp)
	}
}

func TestQueryByChaincodeBadStatus(t *testing.T) {
	c := setupTestClient()

	peer := mocks.MockPeer{MockName: "Peer1", MockURL: "http://peer1.com", MockRoles: []string{}, MockCert: nil, Payload: []byte("A"), Status: 99}

	request := fab.ChaincodeInvokeRequest{
		ChaincodeID: "cc",
		Fcn:         "Hello",
	}
	_, err := c.queryChaincodeWithTarget(request, &peer)
	if err == nil {
		t.Fatalf("expected failure due to bad status")
	}
}

func TestQueryByChaincodeError(t *testing.T) {
	c := setupTestClient()

	peer := mocks.MockPeer{MockName: "Peer1", MockURL: "http://peer1.com", MockRoles: []string{}, MockCert: nil, Payload: []byte("A"), Error: errors.New("error")}

	request := fab.ChaincodeInvokeRequest{
		ChaincodeID: "cc",
		Fcn:         "Hello",
	}
	_, err := c.queryChaincodeWithTarget(request, &peer)
	if err == nil {
		t.Fatalf("expected failure due to error")
	}
}

func TestGenesisBlockOrdererErr(t *testing.T) {
	const channelName = "testchannel"
	client := setupTestClient()

	orderer := mocks.NewMockOrderer("", nil)
	orderer.(mocks.MockOrderer).EnqueueForSendDeliver(mocks.NewSimpleMockError())

	_, err := client.GenesisBlockFromOrderer(channelName, orderer)

	if err == nil {
		t.Fatal("GenesisBlock Test supposed to fail with error")
	}
}

func TestGenesisBlock(t *testing.T) {
	const channelName = "testchannel"
	client := setupTestClient()

	orderer := mocks.NewMockOrderer("", nil)
	orderer.(mocks.MockOrderer).EnqueueForSendDeliver(mocks.NewSimpleMockBlock())

	_, err := client.GenesisBlockFromOrderer(channelName, orderer)

	if err != nil {
		t.Fatalf("GenesisBlock failed: %s", err)
	}
}

/*
// TODO - make a much shorter timeout for this test.
func TestGenesisBlockOrdererTimeout(t *testing.T) {
	const channelName = "testchannel"

	client := setupTestClient()
	orderer := mocks.NewMockOrderer("", nil)

	_, err := client.GenesisBlockFromOrderer(channelName, orderer)

	//It should fail with timeout
	if err == nil || !strings.HasSuffix(err.Error(), "timeout waiting for response from orderer") {
		t.Fatal("GenesisBlock Test supposed to fail with timeout error")
	}
}*/

func TestGenesisBlockOrderer(t *testing.T) {
	const channelName = "testchannel"
	client := setupTestClient()

	orderer := mocks.NewMockOrderer("", nil)
	orderer.(mocks.MockOrderer).EnqueueForSendDeliver(mocks.NewSimpleMockError())

	//Call get Genesis block
	_, err := client.GenesisBlockFromOrderer(channelName, orderer)

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
