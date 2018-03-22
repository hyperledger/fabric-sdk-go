/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package mocks

import (
	ab "github.com/hyperledger/fabric-sdk-go/internal/github.com/hyperledger/fabric/protos/orderer"
	clientmocks "github.com/hyperledger/fabric-sdk-go/pkg/fab/events/client/mocks"
	"github.com/hyperledger/fabric-sdk-go/pkg/fab/events/deliverclient/connection"
	cb "github.com/hyperledger/fabric-sdk-go/third_party/github.com/hyperledger/fabric/protos/common"
	pb "github.com/hyperledger/fabric-sdk-go/third_party/github.com/hyperledger/fabric/protos/peer"
	"github.com/pkg/errors"
)

const (
	// Connect is the Connect operation (used in the OperationMap)
	Connect clientmocks.Operation = "connect"

	// BadRequestResult indicates that the operation should use invalid seek info
	BadRequestResult clientmocks.Result = "bad-request"

	// ForbiddenResult indicates that the user does not have permission to perform the operation
	ForbiddenResult clientmocks.Result = "forbidden"
)

// MockConnection is a fake connection used for unit testing
type MockConnection struct {
	clientmocks.MockConnection
}

// NewConnection returns a new MockConnection using the given options
func NewConnection(opts ...clientmocks.Opt) *MockConnection {
	return &MockConnection{
		MockConnection: *clientmocks.NewMockConnection(opts...),
	}
}

// Receive implements the MockConnection interface
func (c *MockConnection) Receive(eventch chan<- interface{}) {
	result, ok := c.Result(Connect)
	if ok {
		switch result.Result {
		case BadRequestResult:
			eventch <- c.newDeliverStatusResponse(cb.Status_BAD_REQUEST)
			return
		case ForbiddenResult:
			eventch <- c.newDeliverStatusResponse(cb.Status_FORBIDDEN)
			return
		}
	}
	c.MockConnection.Receive(eventch)
}

// Send mockcore sending seek info to the deliver server
func (c *MockConnection) Send(sinfo *ab.SeekInfo) error {
	if c.Closed() {
		return errors.New("mock connection is closed")
	}

	switch seek := sinfo.Start.Type.(type) {
	case *ab.SeekPosition_Specified:
		// Deliver all blocks from the given block number
		fromBlock := seek.Specified.Number
		c.Ledger().SendFrom(fromBlock)
	case *ab.SeekPosition_Oldest:
		// Deliver all blocks from the beginning
		c.Ledger().SendFrom(0)
	}

	return nil
}

func (c *MockConnection) newDeliverStatusResponse(status cb.Status) *connection.Event {
	return connection.NewEvent(
		&pb.DeliverResponse{
			Type: &pb.DeliverResponse_Status{
				Status: status,
			},
		},
		c.SourceURL(),
	)
}
