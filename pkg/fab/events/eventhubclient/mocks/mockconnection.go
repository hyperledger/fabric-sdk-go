/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package mocks

import (
	"fmt"

	clientmocks "github.com/hyperledger/fabric-sdk-go/pkg/fab/events/client/mocks"
	"github.com/hyperledger/fabric-sdk-go/pkg/fab/events/eventhubclient/connection"
	pb "github.com/hyperledger/fabric-sdk-go/third_party/github.com/hyperledger/fabric/protos/peer"
	"github.com/pkg/errors"
)

const (
	// RegInterests is the register operation (used in the OperationMap)
	RegInterests clientmocks.Operation = "reg-interests"

	// UnregInterests is the unregister operation (used in the OperationMap)
	UnregInterests clientmocks.Operation = "unreg-interests"
)

// MockConnection is a mock event hub connection used for unit testing
type MockConnection struct {
	clientmocks.MockConnection
}

// NewConnection returns a new MockConnection using the given options
func NewConnection(opts ...clientmocks.Opt) *MockConnection {
	return &MockConnection{
		MockConnection: *clientmocks.NewMockConnection(opts...),
	}
}

// Send simulates sending register/unregister events to the event hub
func (c *MockConnection) Send(emsg *pb.Event) error {
	if c.Closed() {
		return errors.New("mock connection is closed")
	}

	if err := c.checkEvent(emsg); err != nil {
		return err
	}

	return nil
}
func (c *MockConnection) checkEvent(emsg *pb.Event) error {
	switch evt := emsg.Event.(type) {
	case *pb.Event_Register:
		result, exists := c.Result(RegInterests)
		if exists {
			switch result.Result {
			case clientmocks.NoOpResult:
				// Don't send a response
				return nil
			case clientmocks.FailResult:
				c.ProduceEvent(c.newRegInterestsResponse(nil))
				return nil
			}
		}
		c.ProduceEvent(c.newRegInterestsResponse(evt.Register.Events))

	case *pb.Event_Unregister:
		result, exists := c.Result(UnregInterests)
		if exists {
			switch result.Result {
			case clientmocks.NoOpResult:
				// Don't send a response
				return nil
			case clientmocks.FailResult:
				c.ProduceEvent(c.newUnregInterestsResponse(nil))
				return nil
			}
		}
		c.ProduceEvent(c.newUnregInterestsResponse(evt.Unregister.Events))

	default:
		panic(fmt.Sprintf("unsupported event type: %T", evt))
	}
	return nil
}

func (c *MockConnection) newRegInterestsResponse(interests []*pb.Interest) *connection.Event {
	return connection.NewEvent(
		&pb.Event{
			Event: &pb.Event_Register{
				Register: &pb.Register{
					Events: interests,
				},
			},
		},
		c.SourceURL(),
	)
}

func (c *MockConnection) newUnregInterestsResponse(interests []*pb.Interest) *connection.Event {
	return connection.NewEvent(
		&pb.Event{
			Event: &pb.Event_Unregister{
				Unregister: &pb.Unregister{
					Events: interests,
				},
			},
		},
		c.SourceURL(),
	)
}
