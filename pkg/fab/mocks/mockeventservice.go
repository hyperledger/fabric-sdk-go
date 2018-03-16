/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package mocks

import (
	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/fab"
	"github.com/hyperledger/fabric-sdk-go/pkg/fab/events/service/dispatcher"
)

// MockEventService implements a mock event service
type MockEventService struct {
	TxStatusRegCh chan *dispatcher.TxStatusReg
}

// NewMockEventService returns a new mock event service
func NewMockEventService() *MockEventService {
	return &MockEventService{
		TxStatusRegCh: make(chan *dispatcher.TxStatusReg, 1),
	}
}

// RegisterBlockEvent registers for block events.
func (m *MockEventService) RegisterBlockEvent(filter ...fab.BlockFilter) (fab.Registration, <-chan *fab.BlockEvent, error) {
	panic("not implemented")
}

// RegisterFilteredBlockEvent registers for filtered block events.
func (m *MockEventService) RegisterFilteredBlockEvent() (fab.Registration, <-chan *fab.FilteredBlockEvent, error) {
	panic("not implemented")
}

// RegisterChaincodeEvent registers for chaincode events.
func (m *MockEventService) RegisterChaincodeEvent(ccID, eventFilter string) (fab.Registration, <-chan *fab.CCEvent, error) {
	panic("not implemented")
}

// RegisterTxStatusEvent registers for transaction status events.
func (m *MockEventService) RegisterTxStatusEvent(txID string) (fab.Registration, <-chan *fab.TxStatusEvent, error) {
	eventCh := make(chan *fab.TxStatusEvent)
	reg := &dispatcher.TxStatusReg{
		Eventch: eventCh,
		TxID:    txID,
	}
	m.TxStatusRegCh <- reg
	return reg, eventCh, nil
}

// Unregister removes the given registration.
func (m *MockEventService) Unregister(reg fab.Registration) {
	// Nothing to do
}
