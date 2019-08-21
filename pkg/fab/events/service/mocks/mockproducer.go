/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package mocks

import (
	"sync"
	"sync/atomic"
	"time"

	pb "github.com/hyperledger/fabric-protos-go/peer"
	"github.com/hyperledger/fabric-sdk-go/pkg/util/test"
)

// Ledger is a MockLedger
type Ledger interface {
	// Register registers an event consumer
	Register(consumer Consumer)

	// Unregister unregisters the given consumer
	Unregister(consumer Consumer)

	// NewBlock creates a new block
	NewBlock(channelID string, transactions ...*TxInfo) Block

	// NewFilteredBlock returns a new filtered block
	NewFilteredBlock(channelID string, filteredTx ...*pb.FilteredTransaction)

	// SendFrom sends block events to all registered consumers from the
	// given block number
	SendFrom(blockNum uint64)
}

// MockProducer produces events for unit testing
type MockProducer struct {
	sync.RWMutex
	rcvch         chan interface{}
	eventChannels []chan interface{}
	ledger        Ledger
	closed        int32
}

// NewMockProducer returns a new MockProducer
func NewMockProducer(ledger Ledger) *MockProducer {
	c := &MockProducer{
		rcvch:  make(chan interface{}, 100),
		ledger: ledger,
	}
	go c.listen()
	ledger.Register(c.rcvch)
	return c
}

// Close closes the event producer
func (c *MockProducer) Close() {
	if !atomic.CompareAndSwapInt32(&c.closed, 0, 1) {
		// Already closed
		return
	}

	c.ledger.Unregister(c.rcvch)
	close(c.rcvch)
}

// Register registers an event channel
func (c *MockProducer) Register() <-chan interface{} {
	c.Lock()
	defer c.Unlock()

	eventch := make(chan interface{})
	c.eventChannels = append(c.eventChannels, eventch)
	return eventch
}

// Unregister unregisters an event channel
func (c *MockProducer) Unregister(eventch chan<- interface{}) {
	c.Lock()
	defer c.Unlock()

	for i, e := range c.eventChannels {
		if e == eventch {
			if i != 0 {
				c.eventChannels = c.eventChannels[1:]
			}
			c.eventChannels = c.eventChannels[1:]
			close(eventch)
			return
		}
	}
}

// Ledger returns the mock ledger
func (c *MockProducer) Ledger() Ledger {
	return c.ledger
}

func (c *MockProducer) listen() {
	for {
		event, ok := <-c.rcvch
		if !ok {
			// Channel is closed
			c.unregisterAll()
			return
		}
		c.notifyAll(event)
	}
}

func (c *MockProducer) notifyAll(event interface{}) {
	c.RLock()
	defer c.RUnlock()

	for _, eventch := range c.eventChannels {
		send(eventch, event)
	}
}

func (c *MockProducer) unregisterAll() {
	c.Lock()
	defer c.Unlock()

	for _, eventch := range c.eventChannels {
		close(eventch)
	}
	c.eventChannels = nil
}

func send(eventch chan<- interface{}, event interface{}) {
	defer func() {
		// During shutdown, events may still be produced and we may
		// get a 'send on closed channel' panic. Just log and ignore the error.
		if p := recover(); p != nil {
			test.Logf("panic while submitting event %#v: %s", event, p)
		}
	}()

	select {
	case eventch <- event:
	case <-time.After(5 * time.Second):
		test.Logf("***** Timed out sending event.")
	}
}
