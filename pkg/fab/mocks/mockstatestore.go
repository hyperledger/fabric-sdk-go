/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package mocks

import (
	"github.com/hyperledger/fabric-sdk-go/pkg/context/api/core"
)

// MockStateStore is mock signing manager
type MockStateStore struct {
	cryptoProvider core.CryptoSuite
	hashOpts       core.HashOpts
	signerOpts     core.SignerOpts
}

// NewMockStateStore Constructor for a mock signing manager.
func NewMockStateStore() core.KVStore {
	return &MockStateStore{}
}

// Store sets the value for the key.
func (s *MockStateStore) Store(key interface{}, value interface{}) error {
	return nil
}

//Load returns the value stored in the store for a key.
func (s *MockStateStore) Load(key interface{}) (interface{}, error) {
	return nil, nil
}

//Delete deletes the value for a key.
func (s *MockStateStore) Delete(key interface{}) error {
	return nil
}
