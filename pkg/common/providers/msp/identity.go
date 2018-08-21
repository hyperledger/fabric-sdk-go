/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package msp

import (
	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/core"
	"github.com/pkg/errors"
)

var (
	// ErrUserNotFound indicates the user was not found
	ErrUserNotFound = errors.New("user not found")
)

// IdentityOption captures options used for creating a new SigningIdentity instance
type IdentityOption struct {
	Cert       []byte
	PrivateKey []byte
}

// SigningIdentityOption describes a functional parameter for creating a new SigningIdentity instance
type SigningIdentityOption func(*IdentityOption) error

// WithPrivateKey can be passed as an option when creating a new SigningIdentity.
// It cannot be used without the WithCert option
func WithPrivateKey(key []byte) SigningIdentityOption {
	return func(o *IdentityOption) error {
		o.PrivateKey = key
		return nil
	}
}

// WithCert can be passed as an option when creating a new SigningIdentity.
// When used alone, SDK will lookup the corresponding private key.
func WithCert(cert []byte) SigningIdentityOption {
	return func(o *IdentityOption) error {
		o.Cert = cert
		return nil
	}
}

// IdentityManager provides management of identities in Fabric network
type IdentityManager interface {
	GetSigningIdentity(name string) (SigningIdentity, error)
	CreateSigningIdentity(ops ...SigningIdentityOption) (SigningIdentity, error)
}

// Identity represents a Fabric client identity
type Identity interface {

	// Identifier returns the identifier of that identity
	Identifier() *IdentityIdentifier

	// Verify a signature over some message using this identity as reference
	Verify(msg []byte, sig []byte) error

	// Serialize converts an identity to bytes
	Serialize() ([]byte, error)

	// EnrollmentCertificate Returns the underlying ECert representing this user’s identity.
	EnrollmentCertificate() []byte
}

// SigningIdentity is an extension of Identity to cover signing capabilities.
type SigningIdentity interface {

	// Extends Identity
	Identity

	// Sign the message
	Sign(msg []byte) ([]byte, error)

	// GetPublicVersion returns the public parts of this identity
	PublicVersion() Identity

	// PrivateKey returns the crypto suite representation of the private key
	PrivateKey() core.Key
}

// IdentityIdentifier is a holder for the identifier of a specific
// identity, naturally namespaced, by its provider identifier.
type IdentityIdentifier struct {

	// The identifier of the associated membership service provider
	MSPID string

	// The identifier for an identity within a provider
	ID string
}
