/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/
package resmgmt

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"time"

	"github.com/golang/protobuf/proto"

	"github.com/hyperledger/fabric-protos-go/common"
	"github.com/hyperledger/fabric-sdk-go/internal/github.com/hyperledger/fabric/common/cauthdsl"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/context"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/fab"
	sdkCtx "github.com/hyperledger/fabric-sdk-go/pkg/context"
	"github.com/hyperledger/fabric-sdk-go/pkg/fab/mocks"
	"github.com/hyperledger/fabric-sdk-go/pkg/fab/resource"
	mspmocks "github.com/hyperledger/fabric-sdk-go/pkg/msp/test/mockmsp"
	"github.com/hyperledger/fabric-sdk-go/test/metadata"
)

func Example() {

	// Create new resource management client
	c, err := New(mockClientProvider())
	if err != nil {
		fmt.Println("failed to create client")
	}

	// Read channel configuration tx
	channelConfigTxPath := filepath.Join(metadata.GetProjectPath(), metadata.ChannelConfigPath, channelConfigTxFile)
	r, err := os.Open(channelConfigTxPath)
	if err != nil {
		fmt.Printf("failed to open channel config: %s\n", err)
	}
	defer r.Close()

	// Create new channel 'mychannel'
	_, err = c.SaveChannel(SaveChannelRequest{ChannelID: "mychannel", ChannelConfig: r})
	if err != nil {
		fmt.Printf("failed to save channel: %s\n", err)
	}

	peer := mockPeer()

	// Peer joins channel 'mychannel'
	err = c.JoinChannel("mychannel", WithTargets(peer))
	if err != nil {
		fmt.Printf("failed to join channel: %s\n", err)
	}

	// Install example chaincode to peer
	installReq := InstallCCRequest{Name: "ExampleCC", Version: "v0", Path: "path", Package: &resource.CCPackage{Type: 1, Code: []byte("bytes")}}
	_, err = c.InstallCC(installReq, WithTargets(peer))
	if err != nil {
		fmt.Printf("failed to install chaincode: %s\n", err)
	}

	// Instantiate example chaincode on channel 'mychannel'
	ccPolicy := cauthdsl.SignedByMspMember("Org1MSP")
	instantiateReq := InstantiateCCRequest{Name: "ExampleCC", Version: "v0", Path: "path", Policy: ccPolicy}
	_, err = c.InstantiateCC("mychannel", instantiateReq, WithTargets(peer))
	if err != nil {
		fmt.Printf("failed to install chaincode: %s\n", err)
	}

	fmt.Println("Network setup completed")

	// Output: Network setup completed
}

func ExampleNew() {

	ctx := mockClientProvider()

	c, err := New(ctx)
	if err != nil {
		fmt.Println("failed to create client")
	}

	if c != nil {
		fmt.Println("resource management client created")
	}

	// Output: resource management client created
}

func ExampleWithDefaultTargetFilter() {

	ctx := mockClientProvider()

	c, err := New(ctx, WithDefaultTargetFilter(&urlTargetFilter{url: "example.com"}))
	if err != nil {
		fmt.Println("failed to create client")
	}

	if c != nil {
		fmt.Println("resource management client created with url target filter")
	}

	// Output: resource management client created with url target filter
}

// urlTargetFilter filters targets based on url
type urlTargetFilter struct {
	url string
}

// Accept returns true if this peer is to be included in the target list
func (f *urlTargetFilter) Accept(peer fab.Peer) bool {
	return peer.URL() == f.url
}

func ExampleWithParentContext() {

	c, err := New(mockClientProvider())
	if err != nil {
		fmt.Println("failed to create client")
	}

	clientContext, err := mockClientProvider()()
	if err != nil {
		fmt.Println("failed to return client context")
		return
	}

	// get parent context and cancel
	parentContext, cancel := sdkCtx.NewRequest(clientContext, sdkCtx.WithTimeout(20*time.Second))
	defer cancel()

	channels, err := c.QueryChannels(WithParentContext(parentContext), WithTargets(mockPeer()))
	if err != nil {
		fmt.Printf("failed to query for blockchain info: %s\n", err)
	}

	if channels != nil {
		fmt.Println("Retrieved channels that peer belongs to")
	}

	// Output: Retrieved channels that peer belongs to
}

func ExampleWithTargets() {

	c, err := New(mockClientProvider())
	if err != nil {
		fmt.Println("failed to create client")
	}

	response, err := c.QueryChannels(WithTargets(mockPeer()))
	if err != nil {
		fmt.Printf("failed to query channels: %s\n", err)
	}

	if response != nil {
		fmt.Println("Retrieved channels")
	}

	// Output: Retrieved channels
}

func ExampleWithTargetFilter() {

	c, err := New(mockClientProvider())
	if err != nil {
		fmt.Println("failed to create client")
	}

	ccPolicy := cauthdsl.SignedByMspMember("Org1MSP")
	req := InstantiateCCRequest{Name: "ExampleCC", Version: "v0", Path: "path", Policy: ccPolicy}

	resp, err := c.InstantiateCC("mychannel", req, WithTargetFilter(&urlTargetFilter{url: "http://peer1.com"}))
	if err != nil {
		fmt.Printf("failed to install chaincode: %s\n", err)
	}

	if resp.TransactionID == "" {
		fmt.Println("Failed to instantiate chaincode")
	}

	fmt.Println("Chaincode instantiated")

	// Output: Chaincode instantiated

}

func ExampleClient_SaveChannel() {

	c, err := New(mockClientProvider())
	if err != nil {
		fmt.Printf("failed to create client: %s\n", err)
	}

	channelConfigTxPath := filepath.Join(metadata.GetProjectPath(), metadata.ChannelConfigPath, channelConfigTxFile)
	r, err := os.Open(channelConfigTxPath)
	if err != nil {
		fmt.Printf("failed to open channel config: %s\n", err)
	}
	defer r.Close()

	resp, err := c.SaveChannel(SaveChannelRequest{ChannelID: "mychannel", ChannelConfig: r})
	if err != nil {
		fmt.Printf("failed to save channel: %s\n", err)
	}

	if resp.TransactionID == "" {
		fmt.Println("Failed to save channel")
	}

	fmt.Println("Saved channel")

	// Output: Saved channel
}

func ExampleClient_SaveChannel_withOrdererEndpoint() {

	c, err := New(mockClientProvider())
	if err != nil {
		fmt.Printf("failed to create client: %s\n", err)
	}

	channelConfigTxPath := filepath.Join(metadata.GetProjectPath(), metadata.ChannelConfigPath, channelConfigTxFile)
	r, err := os.Open(channelConfigTxPath)
	if err != nil {
		fmt.Printf("failed to open channel config: %s\n", err)
	}
	defer r.Close()

	resp, err := c.SaveChannel(SaveChannelRequest{ChannelID: "mychannel", ChannelConfig: r}, WithOrdererEndpoint("example.com"))
	if err != nil {
		fmt.Printf("failed to save channel: %s\n", err)
	}

	if resp.TransactionID == "" {
		fmt.Println("Failed to save channel")
	}

	fmt.Println("Saved channel")

	// Output: Saved channel

}

func ExampleClient_JoinChannel() {

	c, err := New(mockClientProvider())
	if err != nil {
		fmt.Println("failed to create client")
	}

	err = c.JoinChannel("mychannel", WithTargets(mockPeer()))
	if err != nil {
		fmt.Printf("failed to join channel: %s\n", err)
	}

	fmt.Println("Joined channel")

	// Output: Joined channel
}

func ExampleClient_InstallCC() {

	c, err := New(mockClientProvider())
	if err != nil {
		fmt.Println("failed to create client")
	}

	req := InstallCCRequest{Name: "ExampleCC", Version: "v0", Path: "path", Package: &resource.CCPackage{Type: 1, Code: []byte("bytes")}}
	responses, err := c.InstallCC(req, WithTargets(mockPeer()))
	if err != nil {
		fmt.Printf("failed to install chaincode: %s\n", err)
	}

	if len(responses) > 0 {
		fmt.Println("Chaincode installed")
	}

	// Output: Chaincode installed
}

func ExampleClient_InstantiateCC() {

	c, err := New(mockClientProvider())
	if err != nil {
		fmt.Println("failed to create client")
	}

	ccPolicy := cauthdsl.SignedByMspMember("Org1MSP")
	req := InstantiateCCRequest{Name: "ExampleCC", Version: "v0", Path: "path", Policy: ccPolicy}

	resp, err := c.InstantiateCC("mychannel", req)
	if err != nil {
		fmt.Printf("failed to install chaincode: %s\n", err)
	}

	if resp.TransactionID == "" {
		fmt.Println("Failed to instantiate chaincode")
	}

	fmt.Println("Chaincode instantiated")

	// Output: Chaincode instantiated
}

func ExampleClient_UpgradeCC() {

	c, err := New(mockClientProvider())
	if err != nil {
		fmt.Println("failed to create client")
	}

	ccPolicy := cauthdsl.SignedByMspMember("Org1MSP")
	req := UpgradeCCRequest{Name: "ExampleCC", Version: "v1", Path: "path", Policy: ccPolicy}

	resp, err := c.UpgradeCC("mychannel", req, WithTargets(mockPeer()))
	if err != nil {
		fmt.Printf("failed to upgrade chaincode: %s\n", err)
	}

	if resp.TransactionID == "" {
		fmt.Println("Failed to upgrade chaincode")
	}

	fmt.Println("Chaincode upgraded")

	// Output: Chaincode upgraded
}

func ExampleClient_QueryChannels() {

	c, err := New(mockClientProvider())
	if err != nil {
		fmt.Println("failed to create client")
	}

	response, err := c.QueryChannels(WithTargets(mockPeer()))
	if err != nil {
		fmt.Printf("failed to query channels: %s\n", err)
	}

	if response != nil {
		fmt.Println("Retrieved channels")
	}

	// Output: Retrieved channels
}

func ExampleClient_QueryInstalledChaincodes() {

	c, err := New(mockClientProvider())
	if err != nil {
		fmt.Println("failed to create client")
	}

	response, err := c.QueryInstalledChaincodes(WithTargets(mockPeer()))
	if err != nil {
		fmt.Printf("failed to query installed chaincodes: %s\n", err)
	}

	if response != nil {
		fmt.Println("Retrieved installed chaincodes")
	}

	// Output: Retrieved installed chaincodes
}

func ExampleClient_QueryInstantiatedChaincodes() {

	c, err := New(mockClientProvider())
	if err != nil {
		fmt.Println("failed to create client")
	}

	response, err := c.QueryInstantiatedChaincodes("mychannel", WithTargets(mockPeer()))
	if err != nil {
		fmt.Printf("failed to query instantiated chaincodes: %s\n", err)
	}

	if response != nil {
		fmt.Println("Retrieved instantiated chaincodes")
	}

	// Output: Retrieved instantiated chaincodes
}

func mockClientProvider() context.ClientProvider {

	ctx := mocks.NewMockContext(mspmocks.NewMockSigningIdentity("test", "Org1MSP"))

	configlBlockBytes, err := ioutil.ReadFile(filepath.Join("testdata", "config.block"))
	if err != nil {
		fmt.Printf("opening config.block file failed: %s\n", err)
	}
	configBlock := &common.Block{}
	err = proto.Unmarshal(configlBlockBytes, configBlock)
	if err != nil {
		fmt.Printf("unmarshalling configBlock failed: %s\n", err)
	}

	// Create mock orderer with simple mock block
	orderer := mocks.NewMockOrderer("", nil)
	orderer.EnqueueForSendDeliver(
		configBlock,
		common.Status_SUCCESS,
	)
	orderer.EnqueueForSendDeliver(
		configBlock,
		common.Status_SUCCESS,
	)
	orderer.EnqueueForSendDeliver(
		configBlock,
		common.Status_SUCCESS,
	)
	orderer.EnqueueForSendDeliver(
		configBlock,
		common.Status_SUCCESS,
	)
	orderer.CloseQueue()

	setupCustomOrderer(ctx, orderer)

	clientProvider := func() (context.Client, error) {
		return ctx, nil
	}

	return clientProvider
}

func mockPeer() fab.Peer {
	return &mocks.MockPeer{MockName: "Peer1", MockURL: "http://peer1.com", Status: 200}
}
