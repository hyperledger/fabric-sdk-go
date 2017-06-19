/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package mocks

import (
	api "github.com/hyperledger/fabric-sdk-go/api"
	"github.com/hyperledger/fabric/bccsp"
)

// MockUser ...
type MockUser struct {
	name                  string
	roles                 []string
	PrivateKey            bccsp.Key // ****This key is temporary We use it to sign transaction until we have tcerts
	enrollmentCertificate []byte
}

// NewMockUser ...
/**
 * Constructor for a user.
 *
 * @param {string} name - The user name
 */
func NewMockUser(name string) api.User {
	return &MockUser{name: name}
}

// GetName ...
/**
 * Get the user name.
 * @returns {string} The user name.
 */
func (u *MockUser) GetName() string {
	return u.name
}

// GetRoles ...
/**
 * Get the roles.
 * @returns {[]string} The roles.
 */
func (u *MockUser) GetRoles() []string {
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

// GetEnrollmentCertificate ...
/**
 * Returns the underlying ECert representing this user’s identity.
 */
func (u *MockUser) GetEnrollmentCertificate() []byte {
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
/**
 * deprecated.
 */
func (u *MockUser) SetPrivateKey(privateKey bccsp.Key) {
	u.PrivateKey = privateKey
}

// GetPrivateKey ...
/**
 * deprecated.
 */
func (u *MockUser) GetPrivateKey() bccsp.Key {
	return u.PrivateKey
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
