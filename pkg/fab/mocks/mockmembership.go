/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package mocks

// MockMembership mock member id
type MockMembership struct {
	ValidateErr error
	VerifyErr   error
	excludeMSPs []string
}

// NewMockMembership new mock member id
func NewMockMembership() *MockMembership {
	return &MockMembership{}
}

// NewMockMembershipWithMSPFilter return new mock membership where given MSPs will be excluded for ContainsMSP test
func NewMockMembershipWithMSPFilter(mspsToBeExlcluded []string) *MockMembership {
	return &MockMembership{excludeMSPs: mspsToBeExlcluded}
}

// Validate if the given ID was issued by the channel's members
func (m *MockMembership) Validate(serializedID []byte) error {
	return m.ValidateErr
}

// Verify the given signature
func (m *MockMembership) Verify(serializedID []byte, msg []byte, sig []byte) error {
	return m.VerifyErr
}

// ContainsMSP mocks membership.ContainsMSP
func (m *MockMembership) ContainsMSP(msp string) bool {
	for _, v := range m.excludeMSPs {
		if v == msp {
			return false
		}
	}
	return true
}
