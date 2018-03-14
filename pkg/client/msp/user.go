/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package msp

import (
	"github.com/hyperledger/fabric-sdk-go/pkg/context/api/core"
	"github.com/pkg/errors"
)

var (
	// ErrUserNotFound indicates the user was not found
	ErrUserNotFound = errors.New("user not found")
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

// User represents users that have been enrolled and represented by
// an enrollment certificate (ECert) and a signing key. The ECert must have
// been signed by one of the CAs the blockchain network has been configured to trust.
// An enrolled user (having a signing key and ECert) can conduct chaincode deployments,
// transactions and queries with the Chain.
//
// User ECerts can be obtained from a CA beforehand as part of deploying the application,
// or it can be obtained from the optional Fabric COP service via its enrollment process.
//
// Sometimes User identities are confused with Peer identities. User identities represent
// signing capability because it has access to the private key, while Peer identities in
// the context of the application/SDK only has the certificate for verifying signatures.
// An application cannot use the Peer identity to sign things because the application doesn’t
// have access to the Peer identity’s private key.
type User interface {
	MSPID() string
	Name() string
	SerializedIdentity() ([]byte, error)
	PrivateKey() core.Key
	EnrollmentCertificate() []byte
}

// UserData is the representation of User in UserStore
// PrivateKey is stored separately, in the crypto store
type UserData struct {
	Name                  string
	MSPID                 string
	EnrollmentCertificate []byte
}

// UserStore is responsible for UserData persistence
type UserStore interface {
	Store(*UserData) error
	Load(UserIdentifier) (*UserData, error)
}

// UserIdentifier is the User's unique identifier
type UserIdentifier struct {
	MSPID string
	Name  string
}

// PrivKeyKey is a composite key for accessing a private key in the key store
type PrivKeyKey struct {
	MSPID    string
	UserName string
	SKI      []byte
}

// CertKey is a composite key for accessing a cert in the cert store
type CertKey struct {
	MSPID    string
	UserName string
}
