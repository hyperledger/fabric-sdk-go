/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package mocks

import (
	reqContext "context"
	"fmt"
	"net"
	"sync"

	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/fab"
	"github.com/hyperledger/fabric-sdk-go/pkg/util/test"
	"github.com/hyperledger/fabric-sdk-go/third_party/github.com/hyperledger/fabric/protos/common"
	"github.com/pkg/errors"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

// MockOrderer is a mock fabricclient.Orderer
// Note that calling broadcast doesn't deliver anythng. This implies
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
		switch value := value.(type) {
		case common.Status:
		case *common.Block:
			o.Deliveries <- value
		case error:
			o.DeliveryErrors <- value
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

// MockGrpcOrderer is a mock fabricclient.Orderer to test
// connectivity to the orderer. It only wraps a GRPC server.
// Note that calling broadcast doesn't deliver anythng.
// This implies that the broadcast side and the deliver side are totally
// independent from the mocking point of view.
type MockGrpcOrderer struct {
	Creds      credentials.TransportCredentials
	srv        *grpc.Server
	wg         sync.WaitGroup
	OrdererURL string
}

// NewMockGrpcOrderer will create a new instance for the given url and TLS credentials (optional)
func NewMockGrpcOrderer(url string, tls credentials.TransportCredentials) *MockGrpcOrderer {
	o := &MockGrpcOrderer{
		OrdererURL: url,
		Creds:      tls,
	}

	return o
}

// Start with start the underlying GRPC server for this MockGrpcOrderer
// it updates the OrdererUrl with the address returned by the GRPC server
func (o *MockGrpcOrderer) Start() string {
	// pass in TLS creds if present
	if o.Creds != nil {
		o.srv = grpc.NewServer(grpc.Creds(o.Creds))
	} else {
		o.srv = grpc.NewServer()
	}
	lis, err := net.Listen("tcp", o.OrdererURL)
	if err != nil {
		panic(fmt.Sprintf("Error starting GRPC Orderer %s", err))
	}
	addr := lis.Addr().String()

	test.Logf("Starting MockGrpcOrderer [%s]", addr)
	o.wg.Add(1)
	go func() {
		defer o.wg.Done()
		if err := o.srv.Serve(lis); err != nil {
			test.Logf("Start MockGrpcOrderer failed [%s]", err)
		}
	}()

	o.OrdererURL = addr
	return addr
}

// Stop the mock broadcast server and wait for completion.
func (o *MockGrpcOrderer) Stop() {
	if o.srv == nil {
		panic("MockGrpcOrderer not started")
	}
	test.Logf("Stopping MockGrpcOrderer [%s]", o.OrdererURL)
	o.srv.Stop()
	o.wg.Wait()
	o.srv = nil
	test.Logf("Stopped MockGrpcOrderer [%s]", o.OrdererURL)
}

// URL returns the URL of the mock GRPC Orderer
func (o *MockGrpcOrderer) URL() string {
	return o.OrdererURL
}

// SendBroadcast accepts client broadcast calls and attempts connection to the grpc server
// it does not attempt to broadcast the envelope, it only tries to connect to the server
func (o *MockGrpcOrderer) SendBroadcast(ctx reqContext.Context, envelope *fab.SignedEnvelope) (*common.Status, error) {
	test.Logf("creating connection [%s]", o.OrdererURL)
	var err error
	if o.Creds != nil {
		_, err = grpc.DialContext(ctx, o.OrdererURL, grpc.WithTransportCredentials(o.Creds))
	} else {
		_, err = grpc.DialContext(ctx, o.OrdererURL, grpc.WithInsecure())
	}
	if err != nil {
		return nil, errors.WithMessage(err, "dialing orderer failed")
	}

	return nil, nil
}

// SendDeliver is not used and can be implemented for special GRPC connectivity in the future
func (o *MockGrpcOrderer) SendDeliver(ctx reqContext.Context, envelope *fab.SignedEnvelope) (chan *common.Block, chan error) {
	return nil, nil
}
