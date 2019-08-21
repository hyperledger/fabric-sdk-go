/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package mocks

import (
	"crypto"

	"time"

	msp_protos "github.com/hyperledger/fabric-protos-go/msp"
	"github.com/hyperledger/fabric-sdk-go/internal/github.com/hyperledger/fabric/msp"
)

// MockIdentity implements identity
type MockIdentity struct {
	Err error
}

// NewMockIdentity creates new mock identity
func NewMockIdentity(err error) (msp.Identity, error) {
	return &MockIdentity{Err: err}, nil
}

// ExpiresAt returns the time at which the Identity expires.
func (id *MockIdentity) ExpiresAt() time.Time {
	return time.Time{}
}

// SatisfiesPrincipal returns null if this instance matches the supplied principal or an error otherwise
func (id *MockIdentity) SatisfiesPrincipal(principal *msp_protos.MSPPrincipal) error {
	return nil
}

// GetIdentifier returns the identifier (MSPID/IDID) for this instance
func (id *MockIdentity) GetIdentifier() *msp.IdentityIdentifier {
	return nil
}

// GetMSPIdentifier returns the MSP identifier for this instance
func (id *MockIdentity) GetMSPIdentifier() string {
	return ""
}

// Validate returns nil if this instance is a valid identity or an error otherwise
func (id *MockIdentity) Validate() error {
	if id.Err != nil && id.Err.Error() == "Validate" {
		return id.Err
	}
	return nil
}

// GetOrganizationalUnits returns the OU for this instance
func (id *MockIdentity) GetOrganizationalUnits() []*msp.OUIdentifier {
	return nil
}

// Verify checks against a signature and a message
// to determine whether this identity produced the
// signature; it returns nil if so or an error otherwise
func (id *MockIdentity) Verify(msg []byte, sig []byte) error {
	if id.Err != nil && id.Err.Error() == "Verify" {
		return id.Err
	}
	return nil
}

// Serialize returns a byte array representation of this identity
func (id *MockIdentity) Serialize() ([]byte, error) {
	return nil, nil
}

// Anonymous ...
func (id *MockIdentity) Anonymous() bool {
	return false
}

// MockSigningIdentity ...
type MockSigningIdentity struct {
	// we embed everything from a base identity
	MockIdentity

	// signer corresponds to the object that can produce signatures from this identity
	Signer crypto.Signer
}

// NewMockSigningIdentity ...
func NewMockSigningIdentity() (msp.SigningIdentity, error) {
	return &MockSigningIdentity{}, nil
}

// Sign produces a signature over msg, signed by this instance
func (id *MockSigningIdentity) Sign(msg []byte) ([]byte, error) {
	return []byte(""), nil
}

// GetPublicVersion ...
func (id *MockSigningIdentity) GetPublicVersion() msp.Identity {
	return nil
}
