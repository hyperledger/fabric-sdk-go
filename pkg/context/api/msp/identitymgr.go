/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package msp

import (
	"github.com/hyperledger/fabric-sdk-go/pkg/context/api/core"
)

// Identity supplies the serialized identity and key reference.
type Identity interface {
	MSPID() string
	SerializedIdentity() ([]byte, error)
	PrivateKey() core.Key
}

// SigningIdentity is the identity object that encapsulates the user's private key for signing
// and the user's enrollment certificate (identity)
type SigningIdentity struct {
	MSPID          string
	EnrollmentCert []byte
	PrivateKey     core.Key
}

// IdentityManager provides management of identities in a Fabric network
type IdentityManager interface {
	GetSigningIdentity(name string) (*SigningIdentity, error)
	GetUser(name string) (User, error)
}
