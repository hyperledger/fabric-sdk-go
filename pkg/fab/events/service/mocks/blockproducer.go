/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package mocks

import (
	"sync/atomic"

	cb "github.com/hyperledger/fabric-protos-go/common"
	pb "github.com/hyperledger/fabric-protos-go/peer"
)

// BlockProducer is a block BlockProducer that ensures the block
// number is set sequencially
type BlockProducer struct {
	blockNum uint64
}

// NewBlockProducer returns a new block producer
func NewBlockProducer() *BlockProducer {
	return &BlockProducer{}
}

// NewBlock returns a new block
func (p *BlockProducer) NewBlock(channelID string, transactions ...*TxInfo) *cb.Block {
	block := NewBlock(channelID, transactions...)
	block.Header.Number = p.blockNum
	atomic.AddUint64(&p.blockNum, 1)
	return block
}

// NewFilteredBlock returns a new filtered block
func (p *BlockProducer) NewFilteredBlock(channelID string, filteredTx ...*pb.FilteredTransaction) *pb.FilteredBlock {
	block := NewFilteredBlock(channelID, filteredTx...)
	block.Number = p.blockNum
	atomic.AddUint64(&p.blockNum, 1)
	return block
}
