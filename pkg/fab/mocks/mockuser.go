/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package mocks

import (
	"github.com/hyperledger/fabric-sdk-go/pkg/context/api/core"
	"github.com/hyperledger/fabric-sdk-go/pkg/context/api/msp"
)

// MockUser ...
type MockUser struct {
	name                  string
	mspID                 string
	roles                 []string
	privateKey            core.Key
	enrollmentCertificate []byte
}

// NewMockUser ...
/**
 * Constructor for a user.
 *
 * @param {string} name - The user name
 */
func NewMockUser(name string) msp.User {
	return &MockUser{name: name}
}

//NewMockUserWithMSPID to return mock user with MSP ids
func NewMockUserWithMSPID(name string, mspid string) msp.User {
	return &MockUser{name: name, mspID: mspid}
}

// Name ...
/**
 * Get the user name.
 * @returns {string} The user name.
 */
func (u *MockUser) Name() string {
	return u.name
}

// Roles ...
/**
 * Get the roles.
 * @returns {[]string} The roles.
 */
func (u *MockUser) Roles() []string {
	return u.roles
}

// SetRoles ...
/**
 * Set the roles.
 * @param roles {[]string} The roles.
 */
func (u *MockUser) SetRoles(roles []string) {
	u.roles = roles
}

// EnrollmentCertificate ...
/**
 * Returns the underlying ECert representing this user’s identity.
 */
func (u *MockUser) EnrollmentCertificate() []byte {
	return u.enrollmentCertificate
}

// SetEnrollmentCertificate ...
/**
 * Set the user’s Enrollment Certificate.
 */
func (u *MockUser) SetEnrollmentCertificate(cert []byte) {
	u.enrollmentCertificate = cert
}

// SetPrivateKey ...
func (u *MockUser) SetPrivateKey(privateKey core.Key) {
	u.privateKey = privateKey
}

// PrivateKey ...
func (u *MockUser) PrivateKey() core.Key {
	return u.privateKey
}

// SetMSPID sets the MSP for this user
func (u *MockUser) SetMSPID(mspID string) {
	u.mspID = mspID
}

// MSPID returns the MSP for this user
func (u *MockUser) MSPID() string {
	return u.mspID
}

// SerializedIdentity returns MockUser's serialized identity
func (u *MockUser) SerializedIdentity() ([]byte, error) {
	return []byte("test"), nil
}

// GenerateTcerts ...
/**
 * Gets a batch of TCerts to use for transaction. there is a 1-to-1 relationship between
 * TCert and Transaction. The TCert can be generated locally by the SDK using the user’s crypto materials.
 * @param {int} count how many in the batch to obtain
 * @param {[]string} attributes  list of attributes to include in the TCert
 * @return {[]tcert} An array of TCerts
 */
func (u *MockUser) GenerateTcerts(count int, attributes []string) {

}
