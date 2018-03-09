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

// User is a representation of a Fabric user
type User struct {
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
func (u *User) Name() string {
	return u.name
}

// EnrollmentCertificate Returns the underlying ECert representing this user’s identity.
func (u *User) EnrollmentCertificate() []byte {
	return u.enrollmentCertificate
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

// SerializedIdentity returns client's serialized identity
func (u *User) SerializedIdentity() ([]byte, error) {
	serializedIdentity := &pb_msp.SerializedIdentity{Mspid: u.MspID(),
		IdBytes: u.EnrollmentCertificate()}
	identity, err := proto.Marshal(serializedIdentity)
	if err != nil {
		return nil, errors.Wrap(err, "marshal serializedIdentity failed")
	}
	return identity, nil
}

// UserStore is responsible for UserData persistence
type UserStore interface {
	Store(UserData) error
	Load(UserIdentifier) (UserData, error)
}
