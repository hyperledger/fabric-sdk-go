/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package mocks

import (
	"github.com/hyperledger/fabric/protos/common"
	po "github.com/hyperledger/fabric/protos/orderer"
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
}

// Broadcast mock broadcast
func (m *MockBroadcastServer) Broadcast(server po.AtomicBroadcast_BroadcastServer) error {

	if m.BroadcastError != nil {
		return m.BroadcastError
	}

	if m.BroadcastInternalServerError {
		server.Send(broadcastResponseError)
		return nil
	}
	server.Send(broadcastResponseSuccess)
	return nil
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
