/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package mocks

import (
	cb "github.com/hyperledger/fabric-sdk-go/third_party/github.com/hyperledger/fabric/protos/common"
	pb "github.com/hyperledger/fabric-sdk-go/third_party/github.com/hyperledger/fabric/protos/peer"
)

// NewBlockEvent returns a new mock block event initialized with the given block
func NewBlockEvent(block *cb.Block) *pb.DeliverResponse_Block {
	return &pb.DeliverResponse_Block{
		Block: block,
	}
}

// NewFilteredBlockEvent returns a new mock filtered block event initialized with the given filtered block
func NewFilteredBlockEvent(fblock *pb.FilteredBlock) *pb.DeliverResponse_FilteredBlock {
	return &pb.DeliverResponse_FilteredBlock{
		FilteredBlock: fblock,
	}
}
