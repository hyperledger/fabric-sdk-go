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
)

// MockDeliverServer is a mock deliver server
type MockDeliverServer struct {
	sync.RWMutex
	disconnErr error
}

// NewMockDeliverServer returns a new MockDeliverServer
func NewMockDeliverServer() *MockDeliverServer {
	return new(MockDeliverServer)
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
	srv.Send(&pb.DeliverResponse{
		Type: &pb.DeliverResponse_Status{
			Status: cb.Status_SUCCESS,
		},
	})

	for {
		envelope, err := srv.Recv()
		if err == io.EOF || envelope == nil {
			break
		}

		err = s.disconnectErr()
		if err != nil {
			return err
		}

		srv.Send(&pb.DeliverResponse{
			Type: &pb.DeliverResponse_Block{
				Block: &cb.Block{},
			},
		})
	}
	return nil
}

// DeliverFiltered delivers a stream of filtered blocks
func (s *MockDeliverServer) DeliverFiltered(srv pb.Deliver_DeliverFilteredServer) error {
	srv.Send(&pb.DeliverResponse{
		Type: &pb.DeliverResponse_Status{
			Status: cb.Status_SUCCESS,
		},
	})
	for {
		envelope, err := srv.Recv()
		if err == io.EOF || envelope == nil {
			break
		}

		err = s.disconnectErr()
		if err != nil {
			return err
		}

		srv.Send(&pb.DeliverResponse{
			Type: &pb.DeliverResponse_FilteredBlock{
				FilteredBlock: &pb.FilteredBlock{},
			},
		})
	}
	return nil
}
