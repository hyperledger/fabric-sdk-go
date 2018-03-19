/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package mockmsp

import "github.com/hyperledger/fabric-sdk-go/internal/github.com/hyperledger/fabric/bccsp"

// MockKey mockcore BCCSP key
type MockKey struct {
}

// Bytes ...
func (m *MockKey) Bytes() ([]byte, error) {
	return []byte("Not implemented"), nil
}

// SKI ...
func (m *MockKey) SKI() []byte {
	return []byte("Not implemented")
}

// Symmetric ...
func (m *MockKey) Symmetric() bool {
	return false
}

// Private ...
func (m *MockKey) Private() bool {
	return true
}

// PublicKey ...
func (m *MockKey) PublicKey() (bccsp.Key, error) {
	return m, nil
}
