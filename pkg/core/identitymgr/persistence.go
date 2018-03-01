/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package identitymgr

// PrivKeyKey is a composite key for accessing a private key in the key store
type PrivKeyKey struct {
	MspID    string
	UserName string
	SKI      []byte
}

// CertKey is a composite key for accessing a cert in the cert store
type CertKey struct {
	MspID    string
	UserName string
}

// UserData is the representation of User in UserStore
// PrivateKey is stored separately, in the crypto store
type UserData struct {
	Name                  string
	MspID                 string
	EnrollmentCertificate []byte
}

// UserIdentifier is the User's unique identifier
type UserIdentifier struct {
	MspID string
	Name  string
}
