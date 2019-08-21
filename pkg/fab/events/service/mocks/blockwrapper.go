/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package mocks

import (
	cb "github.com/hyperledger/fabric-protos-go/common"
)

// BlockWrapper wraps the Block and conforms to the Block interface
type BlockWrapper struct {
	block *cb.Block
}

// NewBlockWrapper returns a new Block wrapper
func NewBlockWrapper(block *cb.Block) *BlockWrapper {
	return &BlockWrapper{block: block}
}

// Block returns the block
func (w *BlockWrapper) Block() *cb.Block {
	return w.block
}

// Number returns the block number
func (w *BlockWrapper) Number() uint64 {
	return w.block.Header.Number
}

// SetNumber sets the block number
func (w *BlockWrapper) SetNumber(number uint64) {
	w.block.Header.Number = number
}
