/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package mocks

import (
	"fmt"

	fab "github.com/hyperledger/fabric-sdk-go/api/apifabclient"

	"github.com/hyperledger/fabric/protos/common"
)

// MockOrderer is a mock fabricclient.Orderer
// Nothe that calling broadcast doesn't deliver anythng. This implies
// that the broadcast side and the deliver side are totally
// independent from the mocking point of view.
type MockOrderer interface {
	fab.Orderer
	// Enqueues a mock error to be returned to the client calling SendBroadcast
	EnqueueSendBroadcastError(err error)
	// Enqueues a mock value (block or error) for delivery
	EnqueueForSendDeliver(value interface{})
}
type mockOrderer struct {
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
func NewMockOrderer(url string, broadcastListener chan *fab.SignedEnvelope) fab.Orderer {
	o := &mockOrderer{
		OrdererURL:        url,
		BroadcastListener: broadcastListener,
		BroadcastErrors:   make(chan error, 100),
		Deliveries:        make(chan *common.Block, 1),
		DeliveryErrors:    make(chan error, 1),
		BroadcastQueue:    make(chan *fab.SignedEnvelope, 100),
		DeliveryQueue:     make(chan interface{}, 100),
	}

	go broadcast(o)
	go delivery(o)
	return o
}

func broadcast(o *mockOrderer) {
	for {
		value := <-o.BroadcastQueue
		o.BroadcastListener <- value
	}
}

func delivery(o *mockOrderer) {
	for {
		value := <-o.DeliveryQueue
		switch value.(type) {
		case *common.Block:
			o.Deliveries <- value.(*common.Block)
		case error:
			o.DeliveryErrors <- value.(error)
		default:
			panic(fmt.Sprintf("Value not *common.Block nor error: %v", value))
		}
	}
}

// URL returns the URL of the mock Orderer
func (o *mockOrderer) URL() string {
	return o.OrdererURL
}

// SendBroadcast accepts client broadcast calls and reports them to the listener channel
// Returns the first enqueued error, or nil if there are no enqueued errors
func (o *mockOrderer) SendBroadcast(envelope *fab.SignedEnvelope) (*common.Status, error) {
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
func (o *mockOrderer) SendDeliver(envelope *fab.SignedEnvelope) (chan *common.Block,
	chan error) {
	return o.Deliveries, o.DeliveryErrors
}

func (o *mockOrderer) EnqueueSendBroadcastError(err error) {
	o.BroadcastErrors <- err
}

// EnqueueForSendDeliver enqueues a mock value (block or error) for delivery
func (o *mockOrderer) EnqueueForSendDeliver(value interface{}) {
	switch value.(type) {
	case *common.Block:
		o.DeliveryQueue <- value.(*common.Block)
	case error:
		o.DeliveryQueue <- value.(error)
	default:
		panic(fmt.Sprintf("Value not *common.Block nor error: %v", value))
	}
}
