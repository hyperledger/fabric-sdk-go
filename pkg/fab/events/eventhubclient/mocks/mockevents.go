/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package mocks

import (
	"github.com/golang/protobuf/ptypes/timestamp"
	cb "github.com/hyperledger/fabric-sdk-go/third_party/github.com/hyperledger/fabric/protos/common"
	pb "github.com/hyperledger/fabric-sdk-go/third_party/github.com/hyperledger/fabric/protos/peer"
)

// NewBlockEvent returns a new mock block event initialized with the given block
func NewBlockEvent(block *cb.Block) *pb.Event {
	return &pb.Event{
		Creator:   []byte("some-id"),
		Timestamp: &timestamp.Timestamp{Seconds: 1000},
		Event: &pb.Event_Block{
			Block: block,
		},
	}
}

// NewFilteredBlockEvent returns a new mock filtered block event initialized with the given filtered block
func NewFilteredBlockEvent(fblock *pb.FilteredBlock) *pb.Event {
	return &pb.Event{
		Creator:   []byte("some-id"),
		Timestamp: &timestamp.Timestamp{Seconds: 1000},
		Event: &pb.Event_FilteredBlock{
			FilteredBlock: fblock,
		},
	}
}
