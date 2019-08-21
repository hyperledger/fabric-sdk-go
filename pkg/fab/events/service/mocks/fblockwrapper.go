/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package mocks

import (
	pb "github.com/hyperledger/fabric-protos-go/peer"
)

// FilteredBlockWrapper wraps the FilteredBlock and conforms to the Block interface
type FilteredBlockWrapper struct {
	block *pb.FilteredBlock
}

// NewFilteredBlockWrapper returns a new Filtered Block wrapper
func NewFilteredBlockWrapper(block *pb.FilteredBlock) *FilteredBlockWrapper {
	return &FilteredBlockWrapper{block: block}
}

// Block returns the filtered block
func (w *FilteredBlockWrapper) Block() *pb.FilteredBlock {
	return w.block
}

// Number returns the block number
func (w *FilteredBlockWrapper) Number() uint64 {
	return w.block.Number
}

// SetNumber sets the block number
func (w *FilteredBlockWrapper) SetNumber(number uint64) {
	w.block.Number = number
}
