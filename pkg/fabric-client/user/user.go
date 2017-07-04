/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package user

import (
	api "github.com/hyperledger/fabric-sdk-go/api"
	"github.com/hyperledger/fabric/bccsp"
)

type user struct {
	name                  string
	mspID                 string
	roles                 []string
	privateKey            bccsp.Key
	enrollmentCertificate []byte
}

// JSON representation of the user struct
type JSON struct {
	MspID                 string
	Roles                 []string
	PrivateKeySKI         []byte
	EnrollmentCertificate []byte
}

// NewUser Constructor for a user.
// @param {string} name - The user name
// @param {string} mspID - The mspID for this user
// @returns {api.User} new user
func NewUser(name string, mspID string) api.User {
	return &user{name: name, mspID: mspID}
}

// Name Get the user name.
// @returns {string} The user name.
func (u *user) Name() string {
	return u.name
}

// Roles Get the roles.
// @returns {[]string} The roles.
func (u *user) Roles() []string {
	return u.roles
}

// SetRoles Set the roles.
// @param roles {[]string} The roles.
func (u *user) SetRoles(roles []string) {
	u.roles = roles
}

// EnrollmentCertificate Returns the underlying ECert representing this user’s identity.
func (u *user) EnrollmentCertificate() []byte {
	return u.enrollmentCertificate
}

// SetEnrollmentCertificate Set the user’s Enrollment Certificate.
func (u *user) SetEnrollmentCertificate(cert []byte) {
	u.enrollmentCertificate = cert
}

// SetPrivateKey sets the crypto suite representation of the private key
// for this user
func (u *user) SetPrivateKey(privateKey bccsp.Key) {
	u.privateKey = privateKey
}

// PrivateKey returns the crypto suite representation of the private key
func (u *user) PrivateKey() bccsp.Key {
	return u.privateKey
}

// SetMspID sets the MSP for this user
func (u *user) SetMspID(mspID string) {
	u.mspID = mspID
}

// MspID returns the MSP for this user
func (u *user) MspID() string {
	return u.mspID
}

// GenerateTcerts Gets a batch of TCerts to use for transaction. there is a 1-to-1 relationship between
// TCert and Transaction. The TCert can be generated locally by the SDK using the user’s crypto materials.
// @param {int} count how many in the batch to obtain
// @param {[]string} attributes  list of attributes to include in the TCert
// @return {[]tcert} An array of TCerts
func (u *user) GenerateTcerts(count int, attributes []string) {
	// not yet implemented
}
