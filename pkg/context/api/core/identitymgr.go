/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package core

import (
	"errors"
)

var (
	// ErrCARegistrarNotFound indicates the CA registrar was not found
	ErrCARegistrarNotFound = errors.New("CA registrar not found")
)

// SigningIdentity is the identity object that encapsulates the user's private key for signing
// and the user's enrollment certificate (identity)
type SigningIdentity struct {
	MspID          string
	EnrollmentCert []byte
	PrivateKey     Key
}

// IdentityManager provides management of identities in a Fabric network
type IdentityManager interface {
	GetSigningIdentity(name string) (*SigningIdentity, error)
	GetUser(name string) (User, error)
}
