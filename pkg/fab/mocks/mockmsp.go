/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package mocks

import (
	msp_protos "github.com/hyperledger/fabric-protos-go/msp"
	"github.com/hyperledger/fabric-sdk-go/internal/github.com/hyperledger/fabric/msp"
)

// MockMSP implements mock msp
type MockMSP struct {
	Err error
}

// NewMockMSP creates mock msp
func NewMockMSP(err error) *MockMSP {
	return &MockMSP{Err: err}
}

// DeserializeIdentity mockcore deserialize identity
func (m *MockMSP) DeserializeIdentity(serializedIdentity []byte) (msp.Identity, error) {
	if m.Err != nil && m.Err.Error() == "DeserializeIdentity" {
		return nil, m.Err
	}
	return &MockIdentity{Err: m.Err}, nil
}

// IsWellFormed  checks if the given identity can be deserialized into its provider-specific form
func (m *MockMSP) IsWellFormed(identity *msp_protos.SerializedIdentity) error {
	return nil
}

// Setup the MSP instance according to configuration information
func (m *MockMSP) Setup(config *msp_protos.MSPConfig) error {
	return nil
}

// GetMSPs Provides a list of Membership Service providers
func (m *MockMSP) GetMSPs() (map[string]msp.MSP, error) {
	return nil, nil
}

// GetVersion returns the version of this MSP
func (m *MockMSP) GetVersion() msp.MSPVersion {
	return 0
}

// GetType returns the provider type
func (m *MockMSP) GetType() msp.ProviderType {
	return 0
}

// GetIdentifier returns the provider identifier
func (m *MockMSP) GetIdentifier() (string, error) {
	return "", nil
}

// GetSigningIdentity returns a signing identity corresponding to the provided identifier
func (m *MockMSP) GetSigningIdentity(identifier *msp.IdentityIdentifier) (msp.SigningIdentity, error) {
	return nil, nil
}

// GetDefaultSigningIdentity returns the default signing identity
func (m *MockMSP) GetDefaultSigningIdentity() (msp.SigningIdentity, error) {
	return nil, nil
}

// GetTLSRootCerts returns the TLS root certificates for this MSP
func (m *MockMSP) GetTLSRootCerts() [][]byte {
	return nil
}

// GetTLSIntermediateCerts returns the TLS intermediate root certificates for this MSP
func (m *MockMSP) GetTLSIntermediateCerts() [][]byte {
	return nil
}

// Validate checks whether the supplied identity is valid
func (m *MockMSP) Validate(id msp.Identity) error {
	return nil
}

// SatisfiesPrincipal checks whether the identity matches
// the description supplied in MSPPrincipal.
func (m *MockMSP) SatisfiesPrincipal(id msp.Identity, principal *msp_protos.MSPPrincipal) error {
	return nil
}
