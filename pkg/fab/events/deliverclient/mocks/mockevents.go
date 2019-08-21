/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package mocks

import (
	"fmt"

	cb "github.com/hyperledger/fabric-protos-go/common"
	pb "github.com/hyperledger/fabric-protos-go/peer"
	"github.com/hyperledger/fabric-sdk-go/pkg/fab/events/deliverclient/connection"
	servicemocks "github.com/hyperledger/fabric-sdk-go/pkg/fab/events/service/mocks"
)

// NewBlockEvent returns a new mock block event initialized with the given block
func NewBlockEvent(block *cb.Block, sourceURL string) *connection.Event {
	return connection.NewEvent(
		&pb.DeliverResponse{
			Type: &pb.DeliverResponse_Block{
				Block: block,
			},
		}, sourceURL,
	)
}

// NewFilteredBlockEvent returns a new mock filtered block event initialized with the given filtered block
func NewFilteredBlockEvent(fblock *pb.FilteredBlock, sourceURL string) *connection.Event {
	return connection.NewEvent(
		&pb.DeliverResponse{
			Type: &pb.DeliverResponse_FilteredBlock{
				FilteredBlock: fblock,
			},
		}, sourceURL,
	)
}

// BlockEventFactory creates block events
var BlockEventFactory = func(block servicemocks.Block, sourceURL string) servicemocks.BlockEvent {
	b, ok := block.(*servicemocks.BlockWrapper)
	if !ok {
		panic(fmt.Sprintf("Invalid block type: %T", block))
	}
	return NewBlockEvent(b.Block(), sourceURL)
}

// FilteredBlockEventFactory creates filtered block events
var FilteredBlockEventFactory = func(block servicemocks.Block, sourceURL string) servicemocks.BlockEvent {
	b, ok := block.(*servicemocks.FilteredBlockWrapper)
	if !ok {
		panic(fmt.Sprintf("Invalid block type: %T", block))
	}
	return NewFilteredBlockEvent(b.Block(), sourceURL)
}
