/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package mocks

import (
	"github.com/hyperledger/fabric-sdk-go/pkg/context/api"
	"github.com/hyperledger/fabric-sdk-go/pkg/context/api/core"
	"github.com/hyperledger/fabric-sdk-go/pkg/context/api/fab"
	"github.com/pkg/errors"
)

// MockIdentityManager is a mock IdentityManager
type MockIdentityManager struct {
}

// NewMockIdentityManager Constructor for a identity manager.
func NewMockIdentityManager(orgName string, config core.Config, cryptoProvider core.CryptoSuite) (fab.IdentityManager, error) {
	mcm := MockIdentityManager{}
	return &mcm, nil
}

// GetSigningIdentity will return an identity that can be used to cryptographically sign an object
func (mgr *MockIdentityManager) GetSigningIdentity(userName string) (*api.SigningIdentity, error) {

	si := api.SigningIdentity{}
	return &si, nil
}

// Enroll enrolls a user with a Fabric network
func (mgr *MockIdentityManager) Enroll(enrollmentID string, enrollmentSecret string) (core.Key, []byte, error) {
	return nil, nil, errors.New("not implemented")
}

// Reenroll re-enrolls a user
func (mgr *MockIdentityManager) Reenroll(user api.User) (core.Key, []byte, error) {
	return nil, nil, errors.New("not implemented")
}

// Register registers a user with a Fabric network
func (mgr *MockIdentityManager) Register(request *fab.RegistrationRequest) (string, error) {
	return "", errors.New("not implemented")
}

// Revoke revokes a user
func (mgr *MockIdentityManager) Revoke(request *fab.RevocationRequest) (*fab.RevocationResponse, error) {
	return nil, errors.New("not implemented")
}

// CAName return the name of a CA associated with this identity manager
func (mgr *MockIdentityManager) CAName() string {
	return ""
}
