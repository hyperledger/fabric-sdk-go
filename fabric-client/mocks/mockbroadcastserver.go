/*
Copyright SecureKey Technologies Inc. All Rights Reserved.


Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at


      http://www.apache.org/licenses/LICENSE-2.0


Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package mocks

import (
	"github.com/hyperledger/fabric/protos/common"
	"github.com/hyperledger/fabric/protos/orderer"
)

// TestBlock is a test block
var TestBlock = &orderer.DeliverResponse{
	Type: &orderer.DeliverResponse_Block{
		Block: &common.Block{
			Data: &common.BlockData{
				Data: [][]byte{[]byte("test")},
			},
		},
	},
}

var broadcastResponseSuccess = &orderer.BroadcastResponse{Status: common.Status_SUCCESS}
var broadcastResponseError = &orderer.BroadcastResponse{Status: common.Status_INTERNAL_SERVER_ERROR}

// MockBroadcastServer mock broadcast server
type MockBroadcastServer struct {
	DeliverError                 error
	BroadcastInternalServerError bool
}

// Broadcast mock broadcast
func (m *MockBroadcastServer) Broadcast(server orderer.AtomicBroadcast_BroadcastServer) error {
	if m.BroadcastInternalServerError {
		server.Send(broadcastResponseError)
		return nil
	}
	server.Send(broadcastResponseSuccess)
	return nil
}

// Deliver mock deliver
func (m *MockBroadcastServer) Deliver(server orderer.AtomicBroadcast_DeliverServer) error {
	if m.DeliverError != nil {
		return m.DeliverError
	}

	server.Recv()
	server.Send(TestBlock)

	return nil
}
