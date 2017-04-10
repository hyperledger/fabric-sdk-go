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

package fabricclient

import (
	"github.com/hyperledger/fabric/protos/common"
	ab "github.com/hyperledger/fabric/protos/orderer"
)

// mockOrderer is a mock fabricclient.Orderer
type mockOrderer struct {
	MockURL         string
	MockError       error
	DeliverResponse *ab.DeliverResponse
}

// GetURL returns the mock URL of the mock Orderer
func (o *mockOrderer) GetURL() string {
	return o.MockURL
}

// SendBroadcast mocks sending a broadcast by sending nothing nowhere
func (o *mockOrderer) SendBroadcast(envelope *SignedEnvelope) error {
	return o.MockError
}

// SendBroadcast mocks sending a deliver request to the ordering service
func (o *mockOrderer) SendDeliver(envelope *SignedEnvelope) (chan *common.Block,
	chan error) {
	responses := make(chan *common.Block, 1)
	errors := make(chan error, 1)
	responses <- o.DeliverResponse.GetBlock()
	return responses, errors
}

// NewMockDeliverResponse returns a mock DeliverResponse with the given block
func NewMockDeliverResponse(block *common.Block) *ab.DeliverResponse {
	return &ab.DeliverResponse{
		Type: &ab.DeliverResponse_Block{Block: block},
	}
}
