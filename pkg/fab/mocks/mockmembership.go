/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package mocks

// MockMembership mock member id
type MockMembership struct {
	ValidateErr error
	VerifyErr   error
}

// NewMockMembership new mock member id
func NewMockMembership() *MockMembership {
	return &MockMembership{}
}

// Validate if the given ID was issued by the channel's members
func (m *MockMembership) Validate(serializedID []byte) error {
	return m.ValidateErr
}

// Verify the given signature
func (m *MockMembership) Verify(serializedID []byte, msg []byte, sig []byte) error {
	return m.VerifyErr
}
