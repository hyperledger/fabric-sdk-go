/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package mocks

import (
	"io"
	"sync"

	cb "github.com/hyperledger/fabric-protos-go/common"
	pb "github.com/hyperledger/fabric-protos-go/peer"
	"github.com/hyperledger/fabric-sdk-go/pkg/fab/mocks"
	"github.com/hyperledger/fabric-sdk-go/pkg/util/test"
	"github.com/pkg/errors"
)

// MockDeliverServer is a mock deliver server
type MockDeliverServer struct {
	sync.RWMutex
	status     cb.Status
	disconnErr error

	// Note: the mock broadcast server should setup either deliveries or fileteredDeliveries, not both at once.
	//       the same for mock endorser server, it should either call NewMockDeliverServerWithDeliveries or NewMockDeliverServerWithFilteredDeliveries
	//       to get a new instance of MockDeliverServer

	//for mocking communication with a mockBroadCastServer, this channel will receive common blocks sent by that mockBroadcastServer
	deliveries <-chan *cb.Block

	// for mocking communcation with mockBroadCastServer, this channel will received filtered blocks sent by that mockBradcastServer
	filteredDeliveries <-chan *pb.FilteredBlock
}

// NewMockDeliverServer returns a new MockDeliverServer
func NewMockDeliverServer() *MockDeliverServer {
	return &MockDeliverServer{
		status: cb.Status_UNKNOWN,
	}
}

// NewMockDeliverServerWithDeliveries returns a new MockDeliverServer using Deliveries channel with common.Block
func NewMockDeliverServerWithDeliveries(d <-chan *cb.Block) *MockDeliverServer {
	return &MockDeliverServer{
		status:     cb.Status_UNKNOWN,
		deliveries: d,
	}
}

// NewMockDeliverServerWithFilteredDeliveries returns a new MockDeliverServer using filteredDeliveries channel with FilteredBlock
func NewMockDeliverServerWithFilteredDeliveries(d <-chan *pb.FilteredBlock) *MockDeliverServer {
	return &MockDeliverServer{
		status:             cb.Status_UNKNOWN,
		filteredDeliveries: d,
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
	disconnect := make(chan bool)

	go s.handleEvents(srv, disconnect)

	for {
		envelope, err := srv.Recv()
		if err == io.EOF || envelope == nil {
			break
		}

		err = s.disconnectErr()
		if err != nil {
			return err
		}

		newBlock := mocks.NewSimpleMockBlock()
		err1 := srv.Send(&pb.DeliverResponse{
			Type: &pb.DeliverResponse_Block{
				Block: newBlock,
			},
		})
		if err1 != nil {
			return err1
		}
		if err == io.EOF {
			test.Logf("*** mockdelserver err is io.EOF or envelope == nil, disconnecting from Deliver..")
			disconnect <- true
			break
		}
	}
	return nil
}

// DeliverWithPrivateData is not implemented
func (s *MockDeliverServer) DeliverWithPrivateData(pb.Deliver_DeliverWithPrivateDataServer) error {
	return errors.New("not implemented")
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
	disconnect := make(chan bool)

	go s.handleFilteredEvents(srv, disconnect)
	for {
		envelope, err := srv.Recv()
		if err == io.EOF || envelope == nil {
			disconnect <- true
			break
		}

		err = s.disconnectErr()
		if err != nil {
			disconnect <- true
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
		if err == io.EOF {
			test.Logf("*** mockdelserver err is io.EOF or envelope == nil, disconnecting from DeliverFiltered..")
			disconnect <- true
			break
		}
	}
	return nil
}

func (s *MockDeliverServer) handleEvents(srv pb.Deliver_DeliverServer, disconnect chan bool) {
	for {
		select {
		case block, ok := <-s.deliveries:
			if ok {
				//test.Logf("handling block event:[%+v]", block)
				err1 := srv.Send(&pb.DeliverResponse{
					Type: &pb.DeliverResponse_Block{
						Block: block,
					},
				})
				if err1 != nil {
					test.Logf("got error during handle block event: %s", err1)
				}
			} else {
				test.Logf("channel is closed")
				return
			}
		case <-disconnect:
			return
		}
	}
}

func (s *MockDeliverServer) handleFilteredEvents(srv pb.Deliver_DeliverServer, disconnect chan bool) {
	for {
		select {
		case filteredBlock, ok := <-s.filteredDeliveries:
			if ok {
				//test.Logf("handling filteredBlock event: [%+v], blockNumber: %i", filteredBlock, filteredBlock.Number)
				err1 := srv.Send(&pb.DeliverResponse{
					Type: &pb.DeliverResponse_FilteredBlock{
						FilteredBlock: filteredBlock,
					},
				})
				if err1 != nil {
					test.Logf("got error during handle filteredBlock event: %s", err1)
				}
			} else {
				test.Logf("channel is closed")
				return
			}
		case <-disconnect:
			return
		}
	}
}
