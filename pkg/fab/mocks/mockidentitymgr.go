/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package mocks

import (
	"fmt"

	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/msp"
	mspmocks "github.com/hyperledger/fabric-sdk-go/pkg/msp/test/mockmsp"
	"github.com/pkg/errors"
)

// MockIdentityManager is a mock IdentityManager
type MockIdentityManager struct {
	users map[string]msp.SigningIdentity
}

// UsersOptions holds optional users
type UsersOptions struct {
	users map[string]msp.SigningIdentity
}

// UsersOption describes a functional parameter for the New constructor
type UsersOption func(*UsersOptions) error

// WithUsers option
func WithUsers(users map[string]msp.SigningIdentity) UsersOption {
	return func(mgr *UsersOptions) error {
		if mgr.users != nil {
			return errors.New("already initialized")
		}
		mgr.users = users
		return nil
	}
}

// WithUser option
func WithUser(username string, org string) UsersOption {
	return func(mgr *UsersOptions) error {
		if mgr.users != nil {
			return errors.New("already initialized")
		}
		mgr.users = make(map[string]msp.SigningIdentity)
		mgr.users[username] = mspmocks.NewMockSigningIdentity(username, org)
		return nil
	}
}

// NewMockIdentityManager Constructor for a identity manager.
func NewMockIdentityManager(opts ...UsersOption) msp.IdentityManager {

	manager := MockIdentityManager{}

	usersOptions := UsersOptions{}

	for _, param := range opts {
		err := param(&usersOptions)
		if err != nil {
			panic(fmt.Errorf("failed to create IdentityManager: %s", err))
		}
	}
	if usersOptions.users != nil {
		manager.users = usersOptions.users
	} else {
		manager.users = make(map[string]msp.SigningIdentity)
	}

	return &manager
}

// GetSigningIdentity will return an identity that can be used to cryptographically sign an object
func (mgr *MockIdentityManager) GetSigningIdentity(id string) (msp.SigningIdentity, error) {
	si, ok := mgr.users[id]
	if !ok {
		return nil, msp.ErrUserNotFound
	}
	return si, nil
}

// CreateSigningIdentity creates a signing identity with the given options
func (mgr *MockIdentityManager) CreateSigningIdentity(opts ...msp.SigningIdentityOption) (msp.SigningIdentity, error) {
	return nil, errors.New("not implemented")
}
