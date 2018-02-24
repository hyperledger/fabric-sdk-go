/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package mocks

import (
	ab "github.com/hyperledger/fabric-sdk-go/internal/github.com/hyperledger/fabric/protos/orderer"
	clientmocks "github.com/hyperledger/fabric-sdk-go/pkg/fab/events/client/mocks"
	cb "github.com/hyperledger/fabric-sdk-go/third_party/github.com/hyperledger/fabric/protos/common"
	pb "github.com/hyperledger/fabric-sdk-go/third_party/github.com/hyperledger/fabric/protos/peer"
	"github.com/pkg/errors"
)

const (
	// Seek is the Seek operation (used in the OperationMap)
	Seek clientmocks.Operation = "seek"

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

// Send mocks sending seek info to the deliver server
func (c *MockConnection) Send(sinfo *ab.SeekInfo) error {
	if c.Closed() {
		return errors.New("mock connection is closed")
	}

	result, ok := c.Result(Seek)
	if ok && result.Result == clientmocks.NoOpResult {
		// Don't send a response
		return nil
	}

	if ok {
		switch result.Result {
		case BadRequestResult:
			c.ProduceEvent(newDeliverStatusResponse(cb.Status_BAD_REQUEST))
			return nil
		case ForbiddenResult:
			c.ProduceEvent(newDeliverStatusResponse(cb.Status_FORBIDDEN))
			return nil
		}
	}

	c.ProduceEvent(newDeliverStatusResponse(cb.Status_SUCCESS))

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

func newDeliverStatusResponse(status cb.Status) *pb.DeliverResponse_Status {
	return &pb.DeliverResponse_Status{
		Status: status,
	}
}
