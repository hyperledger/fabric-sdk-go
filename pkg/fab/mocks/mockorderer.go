/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package mocks

import (
	reqContext "context"
	"fmt"

	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/fab"
	"github.com/hyperledger/fabric-sdk-go/third_party/github.com/hyperledger/fabric/protos/common"
)

// MockOrderer is a mock fabricclient.Orderer
// Nothe that calling broadcast doesn't deliver anythng. This implies
// that the broadcast side and the deliver side are totally
// independent from the mocking point of view.
type MockOrderer struct {
	OrdererURL        string
	BroadcastListener chan *fab.SignedEnvelope
	BroadcastErrors   chan error
	Deliveries        chan *common.Block
	DeliveryErrors    chan error
	// These queues are used to detach the client, to avoid deadlocks
	BroadcastQueue chan *fab.SignedEnvelope
	DeliveryQueue  chan interface{}
}

// NewMockOrderer ...
func NewMockOrderer(url string, broadcastListener chan *fab.SignedEnvelope) *MockOrderer {
	o := &MockOrderer{
		OrdererURL:        url,
		BroadcastListener: broadcastListener,
		BroadcastErrors:   make(chan error, 100),
		Deliveries:        make(chan *common.Block, 1),
		DeliveryErrors:    make(chan error, 1),
		BroadcastQueue:    make(chan *fab.SignedEnvelope, 100),
		DeliveryQueue:     make(chan interface{}, 100),
	}

	if broadcastListener != nil {
		go broadcast(o)
	}
	go delivery(o)
	return o
}

func broadcast(o *MockOrderer) {
	for {
		value, ok := <-o.BroadcastQueue
		if !ok {
			close(o.BroadcastListener)
			return
		}
		o.BroadcastListener <- value
	}
}

func delivery(o *MockOrderer) {
	for {
		value, ok := <-o.DeliveryQueue
		if !ok {
			close(o.Deliveries)
			return
		}
		switch value.(type) {
		case common.Status:
		case *common.Block:
			o.Deliveries <- value.(*common.Block)
		case error:
			o.DeliveryErrors <- value.(error)
		default:
			panic(fmt.Sprintf("Value not *common.Block nor error: %+v", value))
		}
	}
}

// URL returns the URL of the mock Orderer
func (o *MockOrderer) URL() string {
	return o.OrdererURL
}

// SendBroadcast accepts client broadcast calls and reports them to the listener channel
// Returns the first enqueued error, or nil if there are no enqueued errors
func (o *MockOrderer) SendBroadcast(ctx reqContext.Context, envelope *fab.SignedEnvelope) (*common.Status, error) {
	// Report this call to the listener
	if o.BroadcastListener != nil {
		o.BroadcastQueue <- envelope
	}
	select {
	case err := <-o.BroadcastErrors:
		return nil, err
	default:
		return nil, nil
	}
}

// SendDeliver returns the channels for delivery of prepared mock values and errors (if any)
func (o *MockOrderer) SendDeliver(ctx reqContext.Context, envelope *fab.SignedEnvelope) (chan *common.Block, chan error) {
	return o.Deliveries, o.DeliveryErrors
}

// CloseQueue ends the mock broadcast and delivery queues
func (o *MockOrderer) CloseQueue() {
	close(o.BroadcastQueue)
	close(o.DeliveryQueue)
}

// EnqueueSendBroadcastError enqueues error
func (o *MockOrderer) EnqueueSendBroadcastError(err error) {
	o.BroadcastErrors <- err
}

// EnqueueForSendDeliver enqueues a mock value (block or error) for delivery
func (o *MockOrderer) EnqueueForSendDeliver(value interface{}) {
	switch value.(type) {
	case common.Status:
		o.DeliveryQueue <- value
	case *common.Block:
		o.DeliveryQueue <- value
	case error:
		o.DeliveryQueue <- value
	default:
		panic(fmt.Sprintf("Value not *common.Block nor error: %+v", value))
	}
}
