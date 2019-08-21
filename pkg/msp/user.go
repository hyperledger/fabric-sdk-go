/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package msp

import (
	"github.com/golang/protobuf/proto"

	pb_msp "github.com/hyperledger/fabric-protos-go/msp"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/core"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/msp"
	"github.com/pkg/errors"
)

// User is a representation of a Fabric user
type User struct {
	id                    string
	mspID                 string
	enrollmentCertificate []byte
	privateKey            core.Key
}

// Identifier returns user identifier
func (u *User) Identifier() *msp.IdentityIdentifier {
	return &msp.IdentityIdentifier{MSPID: u.mspID, ID: u.id}
}

// Verify a signature over some message using this identity as reference
func (u *User) Verify(msg []byte, sig []byte) error {
	return errors.New("not implemented")
}

// Serialize converts an identity to bytes
func (u *User) Serialize() ([]byte, error) {
	serializedIdentity := &pb_msp.SerializedIdentity{
		Mspid:   u.mspID,
		IdBytes: u.enrollmentCertificate,
	}
	identity, err := proto.Marshal(serializedIdentity)
	if err != nil {
		return nil, errors.Wrap(err, "marshal serializedIdentity failed")
	}
	return identity, nil
}

// EnrollmentCertificate Returns the underlying ECert representing this user’s identity.
func (u *User) EnrollmentCertificate() []byte {
	return u.enrollmentCertificate
}

// PrivateKey returns the crypto suite representation of the private key
func (u *User) PrivateKey() core.Key {
	return u.privateKey
}

// PublicVersion returns the public parts of this identity
func (u *User) PublicVersion() msp.Identity {
	return u
}

// Sign the message
func (u *User) Sign(msg []byte) ([]byte, error) {
	return nil, errors.New("Sign() function not implemented")
}
