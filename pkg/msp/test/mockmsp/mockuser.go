/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package mockmsp

import (
	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/core"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/msp"
)

// MockSigningIdentity ...
type MockSigningIdentity struct {
	id                    string
	mspid                 string
	enrollmentCertificate []byte
	privateKey            core.Key
}

// NewMockSigningIdentity to return mock user with MSPID
func NewMockSigningIdentity(id string, mspid string) *MockSigningIdentity {
	return &MockSigningIdentity{
		id:    id,
		mspid: mspid,
	}
}

// Identifier returns the identifier of that identity
func (m *MockSigningIdentity) Identifier() *msp.IdentityIdentifier {
	return &msp.IdentityIdentifier{ID: m.id, MSPID: m.mspid}
}

// Verify a signature over some message using this identity as reference
func (m *MockSigningIdentity) Verify(msg []byte, sig []byte) error {
	return nil
}

// Serialize converts an identity to bytes
func (m *MockSigningIdentity) Serialize() ([]byte, error) {
	return []byte(m.id + m.mspid), nil
}

// SetEnrollmentCertificate sets yhe enrollment certificate.
func (m *MockSigningIdentity) SetEnrollmentCertificate(cert []byte) {
	m.enrollmentCertificate = cert
}

// EnrollmentCertificate Returns the underlying ECert representing this user’s identity.
func (m *MockSigningIdentity) EnrollmentCertificate() []byte {
	return m.enrollmentCertificate
}

// Sign the message
func (m *MockSigningIdentity) Sign(msg []byte) ([]byte, error) {
	return nil, nil
}

// PublicVersion returns the public parts of this identity
func (m *MockSigningIdentity) PublicVersion() msp.Identity {
	return nil
}

// SetPrivateKey sets the private key
func (m *MockSigningIdentity) SetPrivateKey(key core.Key) {
	m.privateKey = key
}

// PrivateKey returns the crypto suite representation of the private key
func (m *MockSigningIdentity) PrivateKey() core.Key {
	return m.privateKey
}
