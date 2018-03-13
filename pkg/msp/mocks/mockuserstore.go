/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package mocks

import (
	"github.com/hyperledger/fabric-sdk-go/pkg/context/api/msp"
)

// MockUserStore ...
type MockUserStore struct {
}

// Store ...
func (m *MockUserStore) Store(*msp.UserData) error {
	return nil
}

// Load ...
func (m *MockUserStore) Load(msp.UserIdentifier) (*msp.UserData, error) {
	return &msp.UserData{}, nil
}
