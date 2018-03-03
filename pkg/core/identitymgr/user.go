/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package identitymgr

import (
	"github.com/golang/protobuf/proto"

	"github.com/hyperledger/fabric-sdk-go/pkg/context/api/core"
	pb_msp "github.com/hyperledger/fabric-sdk-go/third_party/github.com/hyperledger/fabric/protos/msp"
	"github.com/pkg/errors"
)

// Internal representation of a Fabric user
type user struct {
	mspID                 string
	name                  string
	enrollmentCertificate []byte
	privateKey            core.Key
}

func userIdentifier(userData UserData) UserIdentifier {
	return UserIdentifier{MspID: userData.MspID, Name: userData.Name}
}

// Name Get the user name.
// @returns {string} The user name.
func (u *user) Name() string {
	return u.name
}

// EnrollmentCertificate Returns the underlying ECert representing this user’s identity.
func (u *user) EnrollmentCertificate() []byte {
	return u.enrollmentCertificate
}

// PrivateKey returns the crypto suite representation of the private key
func (u *user) PrivateKey() core.Key {
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

// SerializedIdentity returns client's serialized identity
func (u *user) SerializedIdentity() ([]byte, error) {
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
func (u *user) GenerateTcerts(count int, attributes []string) {
	// not yet implemented
}

// UserStore is responsible for UserData persistence
type UserStore interface {
	Store(UserData) error
	Load(UserIdentifier) (UserData, error)
}
