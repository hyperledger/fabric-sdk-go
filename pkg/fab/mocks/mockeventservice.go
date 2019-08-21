/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package mocks

import (
	"time"

	pb "github.com/hyperledger/fabric-protos-go/peer"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/fab"
	"github.com/hyperledger/fabric-sdk-go/pkg/fab/events/service/dispatcher"
)

// MockEventService implements a mock event service
type MockEventService struct {
	TxStatusRegCh    chan *dispatcher.TxStatusReg
	TxValidationCode pb.TxValidationCode
	Timeout          bool
}

// NewMockEventService returns a new mock event service
func NewMockEventService() *MockEventService {
	return &MockEventService{
		TxStatusRegCh: make(chan *dispatcher.TxStatusReg, 1),
	}
}

// RegisterBlockEvent registers for block events.
func (m *MockEventService) RegisterBlockEvent(filter ...fab.BlockFilter) (fab.Registration, <-chan *fab.BlockEvent, error) {
	eventCh := make(chan *fab.BlockEvent)
	reg := &dispatcher.BlockReg{
		Eventch: eventCh,
	}
	return reg, eventCh, nil
}

// RegisterFilteredBlockEvent registers for filtered block events.
func (m *MockEventService) RegisterFilteredBlockEvent() (fab.Registration, <-chan *fab.FilteredBlockEvent, error) {
	eventCh := make(chan *fab.FilteredBlockEvent)
	reg := &dispatcher.FilteredBlockReg{
		Eventch: eventCh,
	}
	return reg, eventCh, nil
}

// RegisterChaincodeEvent registers for chaincode events.
func (m *MockEventService) RegisterChaincodeEvent(ccID, eventFilter string) (fab.Registration, <-chan *fab.CCEvent, error) {
	eventCh := make(chan *fab.CCEvent)
	reg := &dispatcher.ChaincodeReg{
		Eventch:     eventCh,
		ChaincodeID: ccID,
		EventFilter: eventFilter,
	}
	return reg, eventCh, nil
}

// RegisterTxStatusEvent registers for transaction status events.
func (m *MockEventService) RegisterTxStatusEvent(txID string) (fab.Registration, <-chan *fab.TxStatusEvent, error) {
	eventCh := make(chan *fab.TxStatusEvent)
	reg := &dispatcher.TxStatusReg{
		Eventch: eventCh,
		TxID:    txID,
	}
	m.TxStatusRegCh <- reg

	if !m.Timeout {

		go func() {
			select {
			case txStatusReg := <-m.TxStatusRegCh:
				txStatusReg.Eventch <- &fab.TxStatusEvent{TxID: txStatusReg.TxID, TxValidationCode: m.TxValidationCode}
			case <-time.After(5 * time.Second):
				panic("time out not expected")
			}
		}()

	}

	return reg, eventCh, nil
}

// Unregister removes the given registration.
func (m *MockEventService) Unregister(reg fab.Registration) {
	// Nothing to do
}
