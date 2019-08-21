/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package mocks

import (
	"fmt"
	"sync"

	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/fab"

	pb "github.com/hyperledger/fabric-protos-go/peer"
)

// Block is n abstract block
type Block interface {
	Number() uint64
	SetNumber(blockNum uint64)
}

// BlockEvent is an abstract event
type BlockEvent interface{}

// Consumer is a consumer of a BlockEvent
type Consumer chan interface{}

// EventFactory creates block events
type EventFactory func(block Block, sourceURL string) BlockEvent

// BlockEventFactory creates block events
var BlockEventFactory = func(block Block, sourceURL string) BlockEvent {
	b, ok := block.(*BlockWrapper)
	if !ok {
		panic(fmt.Sprintf("Invalid block type: %T", block))
	}
	return &fab.BlockEvent{Block: b.Block(), SourceURL: sourceURL}
}

// FilteredBlockEventFactory creates filtered block events
var FilteredBlockEventFactory = func(block Block, sourceURL string) BlockEvent {
	b, ok := block.(*FilteredBlockWrapper)
	if !ok {
		panic(fmt.Sprintf("Invalid block type: %T", block))
	}
	return &fab.FilteredBlockEvent{FilteredBlock: b.Block(), SourceURL: sourceURL}
}

// MockLedger is a mock ledger that stores blocks sequentially
type MockLedger struct {
	sync.RWMutex
	blockProducer *BlockProducer
	consumers     []Consumer
	blocks        []Block
	eventFactory  EventFactory
	sourceURL     string
}

// NewMockLedger creates a new MockLedger
func NewMockLedger(eventFactory EventFactory, sourceURL string) *MockLedger {
	return &MockLedger{
		eventFactory:  eventFactory,
		blockProducer: NewBlockProducer(),
		sourceURL:     sourceURL,
	}
}

// BlockProducer returns the block producer
func (l *MockLedger) BlockProducer() *BlockProducer {
	return l.blockProducer
}

// Register registers an event consumer
func (l *MockLedger) Register(consumer Consumer) {
	l.Lock()
	defer l.Unlock()
	l.consumers = append(l.consumers, consumer)
}

// Unregister unregisters the given consumer
func (l *MockLedger) Unregister(Consumer Consumer) {
	l.Lock()
	defer l.Unlock()

	for i, p := range l.consumers {
		if p == Consumer {
			if i != 0 {
				l.consumers = l.consumers[1:]
			}
			l.consumers = l.consumers[1:]
			break
		}
	}
}

// NewBlock stores a new block on the ledger
func (l *MockLedger) NewBlock(channelID string, transactions ...*TxInfo) Block {
	l.Lock()
	defer l.Unlock()
	block := NewBlockWrapper(l.blockProducer.NewBlock(channelID, transactions...))
	l.Store(block)
	return block
}

// NewFilteredBlock stores a new filtered block on the ledger
func (l *MockLedger) NewFilteredBlock(channelID string, filteredTx ...*pb.FilteredTransaction) {
	l.Lock()
	defer l.Unlock()
	l.Store(NewFilteredBlockWrapper(l.blockProducer.NewFilteredBlock(channelID, filteredTx...)))
}

// Store stores the given block to the ledger
func (l *MockLedger) Store(block Block) {
	l.blocks = append(l.blocks, block)

	for _, p := range l.consumers {
		blockEvent := l.eventFactory(block, l.sourceURL)
		p <- blockEvent
	}
}

// SendFrom sends block events to all registered consumers from the
// given block number
func (l *MockLedger) SendFrom(blockNum uint64) {
	l.RLock()
	defer l.RUnlock()

	if blockNum >= uint64(len(l.blocks)) {
		return
	}

	for _, block := range l.blocks[blockNum:] {
		for _, p := range l.consumers {
			p <- l.eventFactory(block, l.sourceURL)
		}
	}
}
