/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package mocks

import (
	"io"
	"sync"

	cb "github.com/hyperledger/fabric-sdk-go/third_party/github.com/hyperledger/fabric/protos/common"
	pb "github.com/hyperledger/fabric-sdk-go/third_party/github.com/hyperledger/fabric/protos/peer"
	"github.com/pkg/errors"
)

// MockDeliverServer is a mock deliver server
type MockDeliverServer struct {
	sync.RWMutex
	status     cb.Status
	disconnErr error
}

// NewMockDeliverServer returns a new MockDeliverServer
func NewMockDeliverServer() *MockDeliverServer {
	return &MockDeliverServer{
		status: cb.Status_UNKNOWN,
	}
}

// SetStatus sets the status to return when calling Deliver or DeliverFiltered
func (s *MockDeliverServer) SetStatus(status cb.Status) {
	s.Lock()
	defer s.Unlock()
	s.status = status
}

// Status returns the status that's returned when calling Deliver or DeliverFiltered
func (s *MockDeliverServer) Status() cb.Status {
	s.RLock()
	defer s.RUnlock()
	return s.status
}

// Disconnect terminates the stream and returns the given error to the client
func (s *MockDeliverServer) Disconnect(err error) {
	s.Lock()
	defer s.Unlock()
	s.disconnErr = err
}

func (s *MockDeliverServer) disconnectErr() error {
	s.RLock()
	defer s.RUnlock()
	return s.disconnErr
}

// Deliver delivers a stream of blocks
func (s *MockDeliverServer) Deliver(srv pb.Deliver_DeliverServer) error {
	status := s.Status()
	if status != cb.Status_UNKNOWN {
		err := srv.Send(&pb.DeliverResponse{
			Type: &pb.DeliverResponse_Status{
				Status: status,
			},
		})
		return errors.Errorf("returning error status: %s %s", status, err)
	}

	for {
		envelope, err := srv.Recv()
		if err == io.EOF || envelope == nil {
			break
		}

		err = s.disconnectErr()
		if err != nil {
			return err
		}

		err1 := srv.Send(&pb.DeliverResponse{
			Type: &pb.DeliverResponse_Block{
				Block: &cb.Block{},
			},
		})
		if err1 != nil {
			return err1
		}
	}
	return nil
}

// DeliverFiltered delivers a stream of filtered blocks
func (s *MockDeliverServer) DeliverFiltered(srv pb.Deliver_DeliverFilteredServer) error {
	if s.status != cb.Status_UNKNOWN {
		err1 := srv.Send(&pb.DeliverResponse{
			Type: &pb.DeliverResponse_Status{
				Status: s.status,
			},
		})
		return errors.Errorf("returning error status: %s %s", s.status, err1)
	}

	for {
		envelope, err := srv.Recv()
		if err == io.EOF || envelope == nil {
			break
		}

		err = s.disconnectErr()
		if err != nil {
			return err
		}

		err1 := srv.Send(&pb.DeliverResponse{
			Type: &pb.DeliverResponse_FilteredBlock{
				FilteredBlock: &pb.FilteredBlock{},
			},
		})
		if err1 != nil {
			return err1
		}
	}
	return nil
}
