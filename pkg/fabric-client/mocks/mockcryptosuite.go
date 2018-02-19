/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package mocks

import (
	"crypto/sha256"
	"hash"

	"github.com/hyperledger/fabric-sdk-go/api/apicryptosuite"
)

// MockCryptoSuite implementation
type MockCryptoSuite struct {
}

// KeyGen mock key gen
func (m *MockCryptoSuite) KeyGen(opts apicryptosuite.KeyGenOpts) (k apicryptosuite.Key, err error) {
	return nil, nil
}

// KeyImport mock key import
func (m *MockCryptoSuite) KeyImport(raw interface{},
	opts apicryptosuite.KeyImportOpts) (k apicryptosuite.Key, err error) {
	return nil, nil
}

// GetKey mock get key
func (m *MockCryptoSuite) GetKey(ski []byte) (k apicryptosuite.Key, err error) {
	return nil, nil
}

// Hash mock hash
func (m *MockCryptoSuite) Hash(msg []byte, opts apicryptosuite.HashOpts) (hash []byte, err error) {
	return nil, nil
}

// GetHash mock get hash
func (m *MockCryptoSuite) GetHash(opts apicryptosuite.HashOpts) (hash.Hash, error) {
	return sha256.New(), nil
}

// Sign mock signing
func (m *MockCryptoSuite) Sign(k apicryptosuite.Key, digest []byte,
	opts apicryptosuite.SignerOpts) (signature []byte, err error) {
	return []byte("testSignature"), nil
}

//Verify mock verify implementation
func (m *MockCryptoSuite) Verify(k apicryptosuite.Key, signature, digest []byte, opts apicryptosuite.SignerOpts) (valid bool, err error) {
	return true, nil
}
