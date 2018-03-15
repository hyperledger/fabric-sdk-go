/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package mocks

import (
	"io"

	"fmt"
	"net"

	po "github.com/hyperledger/fabric-sdk-go/internal/github.com/hyperledger/fabric/protos/orderer"
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
		server.Recv()
		server.SendMsg(m.DeliverResponse)
		return nil
	}

	server.Recv()
	server.Send(TestBlock)

	return nil
}

//StartMockBroadcastServer starts mock server for unit testing purpose
func StartMockBroadcastServer(broadcastTestURL string, grpcServer *grpc.Server) (*MockBroadcastServer, string) {
	lis, err := net.Listen("tcp", broadcastTestURL)
	if err != nil {
		panic(fmt.Sprintf("Error starting BroadcastServer %s", err))
	}
	addr := lis.Addr().String()

	broadcastServer := new(MockBroadcastServer)
	po.RegisterAtomicBroadcastServer(grpcServer, broadcastServer)
	go grpcServer.Serve(lis)

	return broadcastServer, addr
}
