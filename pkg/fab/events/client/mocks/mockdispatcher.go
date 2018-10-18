/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package mocks

// MockDispatcher is a mock Dispatcher
type MockDispatcher struct {
	Error     error
	LastBlock uint64
}

// Start returns the configured error
func (d *MockDispatcher) Start() error {
	return d.Error
}

// EventCh simply returns the configured error
func (d *MockDispatcher) EventCh() (chan<- interface{}, error) {
	return nil, d.Error
}

// LastBlockNum returns the last block number
func (d *MockDispatcher) LastBlockNum() uint64 {
	return d.LastBlock
}
