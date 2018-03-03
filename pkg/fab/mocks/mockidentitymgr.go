/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package mocks

import (
	"github.com/hyperledger/fabric-sdk-go/pkg/context/api/core"
	"github.com/pkg/errors"
)

// MockIdentityManager is a mock IdentityManager
type MockIdentityManager struct {
}

// NewMockIdentityManager Constructor for a identity manager.
func NewMockIdentityManager(orgName string, cryptoProvider core.CryptoSuite, config core.Config) (core.IdentityManager, error) {
	mcm := MockIdentityManager{}
	return &mcm, nil
}

// GetSigningIdentity will return an identity that can be used to cryptographically sign an object
func (mgr *MockIdentityManager) GetSigningIdentity(userName string) (*core.SigningIdentity, error) {

	si := core.SigningIdentity{
		MspID: "Org1MSP",
	}
	return &si, nil
}

// GetUser will return a user for a given user name
func (mgr *MockIdentityManager) GetUser(userName string) (core.User, error) {
	return nil, nil
}

// Enroll enrolls a user with a Fabric network
func (mgr *MockIdentityManager) Enroll(enrollmentID string, enrollmentSecret string) error {
	return errors.New("not implemented")
}

// Reenroll re-enrolls a user
func (mgr *MockIdentityManager) Reenroll(user core.User) error {
	return errors.New("not implemented")
}

// Register registers a user with a Fabric network
func (mgr *MockIdentityManager) Register(request *core.RegistrationRequest) (string, error) {
	return "", errors.New("not implemented")
}

// Revoke revokes a user
func (mgr *MockIdentityManager) Revoke(request *core.RevocationRequest) (*core.RevocationResponse, error) {
	return nil, errors.New("not implemented")
}

// CAName return the name of a CA associated with this identity manager
func (mgr *MockIdentityManager) CAName() string {
	return ""
}
