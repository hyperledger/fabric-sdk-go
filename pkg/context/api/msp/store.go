/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package msp

// UserData is the representation of User in UserStore
// PrivateKey is stored separately, in the crypto store
type UserData struct {
	ID                    string
	MSPID                 string
	EnrollmentCertificate []byte
}

// UserStore is responsible for UserData persistence
type UserStore interface {
	Store(*UserData) error
	Load(IdentityIdentifier) (*UserData, error)
}

// PrivKeyKey is a composite key for accessing a private key in the key store
type PrivKeyKey struct {
	ID    string
	MSPID string
	SKI   []byte
}
