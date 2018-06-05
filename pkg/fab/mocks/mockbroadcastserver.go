/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package mocks

import (
	"fmt"
	"io"
	"net"
	"sync"

	po "github.com/hyperledger/fabric-sdk-go/internal/github.com/hyperledger/fabric/protos/orderer"
	"github.com/hyperledger/fabric-sdk-go/pkg/util/test"
	"github.com/hyperledger/fabric-sdk-go/third_party/github.com/hyperledger/fabric/protos/common"
	"google.golang.org/grpc"
)

// TestBlock is a test block
var TestBlock = &po.DeliverResponse{
	Type: &po.DeliverResponse_Block{
		Block: &common.Block{
			Data: &common.BlockData{
				Data: [][]byte{[]byte("test")},
			},
		},
	},
}

var broadcastResponseSuccess = &po.BroadcastResponse{Status: common.Status_SUCCESS}
var broadcastResponseError = &po.BroadcastResponse{Status: common.Status_INTERNAL_SERVER_ERROR}

// MockBroadcastServer mock broadcast server
type MockBroadcastServer struct {
	DeliverError                 error
	BroadcastInternalServerError bool
	DeliverResponse              *po.DeliverResponse
	BroadcastError               error
	BroadcastCustomResponse      *po.BroadcastResponse
	srv                          *grpc.Server
	wg                           sync.WaitGroup
}

// Broadcast mock broadcast
func (m *MockBroadcastServer) Broadcast(server po.AtomicBroadcast_BroadcastServer) error {
	_, err := server.Recv()
	if err == io.EOF {
		return nil
	}
	if err != nil {
		return err
	}
	if m.BroadcastError != nil {
		return m.BroadcastError
	}

	if m.BroadcastInternalServerError {
		return server.Send(broadcastResponseError)
	}

	if m.BroadcastCustomResponse != nil {
		return server.Send(m.BroadcastCustomResponse)
	}

	return server.Send(broadcastResponseSuccess)
}

// Deliver mock deliver
func (m *MockBroadcastServer) Deliver(server po.AtomicBroadcast_DeliverServer) error {
	if m.DeliverError != nil {
		return m.DeliverError
	}

	if m.DeliverResponse != nil {
		if _, err := server.Recv(); err != nil {
			return err
		}
		if err := server.SendMsg(m.DeliverResponse); err != nil {
			return err
		}
		return nil
	}

	if _, err := server.Recv(); err != nil {
		return err
	}
	if err := server.Send(TestBlock); err != nil {
		return err
	}

	return nil
}

// Start the mock broadcast server
func (m *MockBroadcastServer) Start(address string) string {
	if m.srv != nil {
		panic("MockBroadcastServer already started")
	}
	m.srv = grpc.NewServer()

	lis, err := net.Listen("tcp", address)
	if err != nil {
		panic(fmt.Sprintf("Error starting BroadcastServer %s", err))
	}
	addr := lis.Addr().String()

	test.Logf("Starting MockEventServer [%s]", addr)
	po.RegisterAtomicBroadcastServer(m.srv, m)
	m.wg.Add(1)
	go func() {
		defer m.wg.Done()
		if err := m.srv.Serve(lis); err != nil {
			test.Logf("StartMockBroadcastServer failed [%s]", err)
		}
	}()

	return addr
}

// Stop the mock broadcast server and wait for completion.
func (m *MockBroadcastServer) Stop() {
	if m.srv == nil {
		panic("MockBroadcastServer not started")
	}

	m.srv.Stop()
	m.wg.Wait()
	m.srv = nil
}
