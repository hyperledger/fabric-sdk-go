/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package identity

import (
	"github.com/golang/protobuf/proto"

	"github.com/hyperledger/fabric-sdk-go/pkg/context/api/core"
	pb_msp "github.com/hyperledger/fabric-sdk-go/third_party/github.com/hyperledger/fabric/protos/msp"
	"github.com/pkg/errors"
)

// User represents a Fabric user registered at an MSP
type User struct {
	name                  string
	mspID                 string
	roles                 []string
	privateKey            core.Key
	enrollmentCertificate []byte
}

// NewUser Constructor for a user.
// @param {string} mspID - The mspID for this user
// @param {string} name - The user name
// @returns {User} new user
func NewUser(mspID string, name string) *User {
	return &User{mspID: mspID, name: name}
}

// Name Get the user name.
// @returns {string} The user name.
func (u *User) Name() string {
	return u.name
}

// Roles Get the roles.
// @returns {[]string} The roles.
func (u *User) Roles() []string {
	return u.roles
}

// SetRoles Set the roles.
// @param roles {[]string} The roles.
func (u *User) SetRoles(roles []string) {
	u.roles = roles
}

// EnrollmentCertificate Returns the underlying ECert representing this user’s identity.
func (u *User) EnrollmentCertificate() []byte {
	return u.enrollmentCertificate
}

// SetEnrollmentCertificate Set the user’s Enrollment Certificate.
func (u *User) SetEnrollmentCertificate(cert []byte) {
	u.enrollmentCertificate = cert
}

// SetPrivateKey sets the crypto suite representation of the private key
// for this user
func (u *User) SetPrivateKey(privateKey core.Key) {
	u.privateKey = privateKey
}

// PrivateKey returns the crypto suite representation of the private key
func (u *User) PrivateKey() core.Key {
	return u.privateKey
}

// SetMspID sets the MSP for this user
func (u *User) SetMspID(mspID string) {
	u.mspID = mspID
}

// MspID returns the MSP for this user
func (u *User) MspID() string {
	return u.mspID
}

// Identity returns client's serialized identity
func (u *User) Identity() ([]byte, error) {
	serializedIdentity := &pb_msp.SerializedIdentity{Mspid: u.MspID(),
		IdBytes: u.EnrollmentCertificate()}
	identity, err := proto.Marshal(serializedIdentity)
	if err != nil {
		return nil, errors.Wrap(err, "marshal serializedIdentity failed")
	}
	return identity, nil
}

// GenerateTcerts Gets a batch of TCerts to use for transaction. there is a 1-to-1 relationship between
// TCert and Transaction. The TCert can be generated locally by the SDK using the user’s crypto materials.
// @param {int} count how many in the batch to obtain
// @param {[]string} attributes  list of attributes to include in the TCert
// @return {[]tcert} An array of TCerts
func (u *User) GenerateTcerts(count int, attributes []string) {
	// not yet implemented
}
