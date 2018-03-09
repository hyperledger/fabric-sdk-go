/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package mocks

import (
	"fmt"

	servicemocks "github.com/hyperledger/fabric-sdk-go/pkg/fab/events/service/mocks"
	cb "github.com/hyperledger/fabric-sdk-go/third_party/github.com/hyperledger/fabric/protos/common"
	pb "github.com/hyperledger/fabric-sdk-go/third_party/github.com/hyperledger/fabric/protos/peer"
)

// NewBlockEvent returns a new mock block event initialized with the given block
func NewBlockEvent(block *cb.Block) *pb.DeliverResponse {
	return &pb.DeliverResponse{
		Type: &pb.DeliverResponse_Block{
			Block: block,
		},
	}
}

// NewFilteredBlockEvent returns a new mock filtered block event initialized with the given filtered block
func NewFilteredBlockEvent(fblock *pb.FilteredBlock) *pb.DeliverResponse {
	return &pb.DeliverResponse{
		Type: &pb.DeliverResponse_FilteredBlock{
			FilteredBlock: fblock,
		},
	}
}

// BlockEventFactory creates block events
var BlockEventFactory = func(block servicemocks.Block) servicemocks.BlockEvent {
	b, ok := block.(*servicemocks.BlockWrapper)
	if !ok {
		panic(fmt.Sprintf("Invalid block type: %T", block))
	}
	return NewBlockEvent(b.Block())
}

// FilteredBlockEventFactory creates filtered block events
var FilteredBlockEventFactory = func(block servicemocks.Block) servicemocks.BlockEvent {
	b, ok := block.(*servicemocks.FilteredBlockWrapper)
	if !ok {
		panic(fmt.Sprintf("Invalid block type: %T", block))
	}
	return NewFilteredBlockEvent(b.Block())
}
