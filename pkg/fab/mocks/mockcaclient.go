/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package mocks

import (
	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/core"
	"github.com/hyperledger/fabric-sdk-go/pkg/msp/api"
	"github.com/pkg/errors"
)

// MockCAClient is a mock CAClient
type MockCAClient struct {
}

// NewMockCAClient Constructor for a CA client.
func NewMockCAClient(orgName string, cryptoProvider core.CryptoSuite, config core.Config) (api.CAClient, error) {
	mcm := MockCAClient{}
	return &mcm, nil
}

// Enroll enrolls a user with a Fabric network
func (mgr *MockCAClient) Enroll(enrollmentID string, enrollmentSecret string) error {
	return errors.New("not implemented")
}

// Reenroll re-enrolls a user
func (mgr *MockCAClient) Reenroll(enrollmentID string) error {
	return errors.New("not implemented")
}

// Register registers a user with a Fabric network
func (mgr *MockCAClient) Register(request *api.RegistrationRequest) (string, error) {
	return "", errors.New("not implemented")
}

// Revoke revokes a user
func (mgr *MockCAClient) Revoke(request *api.RevocationRequest) (*api.RevocationResponse, error) {
	return nil, errors.New("not implemented")
}
