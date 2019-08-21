/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package mocks

import (
	msp_protos "github.com/hyperledger/fabric-protos-go/msp"
	"github.com/hyperledger/fabric-sdk-go/internal/github.com/hyperledger/fabric/msp"
)

// MockMSPManager implements mock msp manager
type MockMSPManager struct {
	MSPs map[string]msp.MSP
	Err  error
}

// NewMockMSPManager mockcore msp manager
func NewMockMSPManager(msps map[string]msp.MSP) *MockMSPManager {
	return &MockMSPManager{MSPs: msps}
}

// NewMockMSPManagerWithError mockcore msp manager
func NewMockMSPManagerWithError(msps map[string]msp.MSP, err error) *MockMSPManager {
	return &MockMSPManager{MSPs: msps, Err: err}
}

// DeserializeIdentity mockcore deserialize identity
func (mgr *MockMSPManager) DeserializeIdentity(serializedIdentity []byte) (msp.Identity, error) {
	return nil, nil
}

// IsWellFormed  checks if the given identity can be deserialized into its provider-specific form
func (mgr *MockMSPManager) IsWellFormed(identity *msp_protos.SerializedIdentity) error {
	return nil
}

// Setup the MSP manager instance according to configuration information
func (mgr *MockMSPManager) Setup(msps []msp.MSP) error {
	return nil
}

// GetMSPs Provides a list of Membership Service providers
func (mgr *MockMSPManager) GetMSPs() (map[string]msp.MSP, error) {
	if mgr.Err != nil && mgr.Err.Error() == "GetMSPs" {
		return nil, mgr.Err
	}

	return mgr.MSPs, nil
}
