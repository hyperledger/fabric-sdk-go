/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package apifabclient

import (
	"github.com/hyperledger/fabric-sdk-go/api/apicryptosuite"
)

// SigningIdentity is the identity object that encapsulates the user's private key for signing
// and the user's enrollment certificate (identity)
type SigningIdentity struct {
	MspID          string
	EnrollmentCert []byte
	PrivateKey     apicryptosuite.Key
}

// CredentialManager retrieves user's signing identity
type CredentialManager interface {
	GetSigningIdentity(name string) (*SigningIdentity, error)
}
