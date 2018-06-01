/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/
package ledger

import (
	"fmt"
	"time"

	"github.com/golang/protobuf/proto"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/context"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/fab"
	sdkCtx "github.com/hyperledger/fabric-sdk-go/pkg/context"
	"github.com/hyperledger/fabric-sdk-go/pkg/fab/mocks"
)

func Example() {

	ctx := mockChannelProvider("mychannel")

	c, err := New(ctx)
	if err != nil {
		fmt.Println("failed to create client")
	}

	block, err := c.QueryBlock(1)
	if err != nil {
		fmt.Printf("failed to query block: %s\n", err)
	}

	if block != nil {
		fmt.Println("Retrieved block #1")
	}

	// Output: Retrieved block #1
}

func ExampleNew() {

	ctx := mockChannelProvider("mychannel")

	c, err := New(ctx)
	if err != nil {
		fmt.Println(err)
	}

	if c != nil {
		fmt.Println("ledger client created")
	}

	// Output: ledger client created
}

func ExampleWithDefaultTargetFilter() {

	ctx := mockChannelProvider("mychannel")

	c, err := New(ctx, WithDefaultTargetFilter(&urlTargetFilter{url: "example.com"}))
	if err != nil {
		fmt.Println(err)
	}

	if c != nil {
		fmt.Println("ledger client created with url target filter")
	}

	// Output: ledger client created with url target filter
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

	c, err := New(mockChannelProvider("mychannel"))
	if err != nil {
		fmt.Println(err)
	}

	channelContext, err := mockChannelProvider("mychannel")()
	if err != nil {
		fmt.Println("failed to return channel context")
		return
	}

	// get parent context and cancel
	parentContext, cancel := sdkCtx.NewRequest(channelContext, sdkCtx.WithTimeout(20*time.Second))
	defer cancel()

	bci, err := c.QueryInfo(WithParentContext(parentContext))
	if err != nil {
		fmt.Printf("failed to query for blockchain info: %s\n", err)
	}

	if bci != nil {
		fmt.Println("Retrieved blockchain info")
	}

	// Output: Retrieved blockchain info
}

func ExampleWithTargets() {

	c, err := New(mockChannelProvider("mychannel"))
	if err != nil {
		fmt.Println("failed to create client")
	}

	cfg, err := c.QueryConfig(WithTargets(mockPeerWithConfigBlock()))
	if err != nil {
		fmt.Printf("failed to query config with target peer: %s\n", err)
	}

	if cfg != nil {
		fmt.Println("Retrieved config from target peer")
	}

	// Output: Retrieved config from target peer
}

func ExampleWithTargetFilter() {

	c, err := New(mockChannelProvider("mychannel"))
	if err != nil {
		fmt.Println(err)
	}

	block, err := c.QueryBlock(1, WithTargetFilter(&urlTargetFilter{url: "example.com"}))
	if err != nil {
		fmt.Printf("failed to query block: %s\n", err)
	}

	if block != nil {
		fmt.Println("Retrieved block #1 from example.com")
	}

	// Output: Retrieved block #1 from example.com
}

func ExampleClient_QueryInfo() {

	c, err := New(mockChannelProvider("mychannel"))
	if err != nil {
		fmt.Println("failed to create client")
	}

	bci, err := c.QueryInfo()
	if err != nil {
		fmt.Printf("failed to query for blockchain info: %s\n", err)
	}

	if bci != nil {
		fmt.Println("Retrieved ledger info")
	}

	// Output: Retrieved ledger info
}

func ExampleClient_QueryBlock() {

	c, err := New(mockChannelProvider("mychannel"))
	if err != nil {
		fmt.Println("failed to create client")
	}

	block, err := c.QueryBlock(1)
	if err != nil {
		fmt.Printf("failed to query block: %s\n", err)
	}

	if block != nil {
		fmt.Println("Retrieved block #1")
	}

	// Output: Retrieved block #1
}

func ExampleClient_QueryBlockByHash() {

	c, err := New(mockChannelProvider("mychannel"))
	if err != nil {
		fmt.Println("failed to create client")
	}

	block, err := c.QueryBlockByHash([]byte("hash"))
	if err != nil {
		fmt.Printf("failed to query block by hash: %s\n", err)
	}

	if block != nil {
		fmt.Println("Retrieved block by hash")
	}

	// Output: Retrieved block by hash
}

func ExampleClient_QueryBlockByTxID() {

	c, err := New(mockChannelProvider("mychannel"))
	if err != nil {
		fmt.Println("failed to create client")
	}

	block, err := c.QueryBlockByTxID("123")
	if err != nil {
		fmt.Printf("failed to query block by transaction ID: %s\n", err)
	}

	if block != nil {
		fmt.Println("Retrieved block by transaction ID")
	}

	// Output: Retrieved block by transaction ID
}

func ExampleClient_QueryTransaction() {

	c, err := New(mockChannelProvider("mychannel"))
	if err != nil {
		fmt.Println("failed to create client")
	}

	t, err := c.QueryTransaction("123")
	if err != nil {
		fmt.Printf("failed to query transaction: %s\n", err)
	}

	if t != nil {
		fmt.Println("Retrieved transaction")
	}

	// Output: Retrieved transaction
}

func ExampleClient_QueryConfig() {

	c, err := New(mockChannelProvider("mychannel"))
	if err != nil {
		fmt.Println("failed to create client")
	}

	cfg, err := c.QueryConfig(WithTargets(mockPeerWithConfigBlock()))
	if err != nil {
		fmt.Printf("failed to query config: %s\n", err)
	}

	if cfg != nil {
		fmt.Println("Retrieved channel configuration")
	}

	// Output: Retrieved channel configuration
}

func mockChannelProvider(channelID string) context.ChannelProvider {

	channelProvider := func() (context.Channel, error) {
		return mocks.NewMockChannel(channelID)
	}

	return channelProvider
}

func mockPeerWithConfigBlock() fab.Peer {

	// create config block builder in order to create valid payload
	builder := &mocks.MockConfigBlockBuilder{
		MockConfigGroupBuilder: mocks.MockConfigGroupBuilder{
			ModPolicy: "Admins",
			MSPNames: []string{
				"Org1MSP",
				"Org2MSP",
			},
			OrdererAddress: "localhost:7054",
			RootCA:         "",
		},
		Index:           0,
		LastConfigIndex: 0,
	}

	payload, err := proto.Marshal(builder.Build())
	if err != nil {
		fmt.Println("Failed to marshal mock block")
	}

	// peer with valid config block payload
	peer := &mocks.MockPeer{MockName: "Peer1", MockURL: "http://peer1.com", MockRoles: []string{}, MockCert: nil, Payload: payload, Status: 200}

	return peer
}
