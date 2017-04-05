/*
Copyright SecureKey Technologies Inc. All Rights Reserved.


Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at


      http://www.apache.org/licenses/LICENSE-2.0


Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package mocks

import (
	"hash"

	"github.com/hyperledger/fabric/bccsp"
)

// MockCryptoSuite implementation
type MockCryptoSuite struct {
}

// KeyGen mock key gen
func (m *MockCryptoSuite) KeyGen(opts bccsp.KeyGenOpts) (k bccsp.Key, err error) {
	return nil, nil
}

// KeyDeriv mock key derivation
func (m *MockCryptoSuite) KeyDeriv(k bccsp.Key,
	opts bccsp.KeyDerivOpts) (dk bccsp.Key, err error) {
	return nil, nil
}

// KeyImport mock key import
func (m *MockCryptoSuite) KeyImport(raw interface{},
	opts bccsp.KeyImportOpts) (k bccsp.Key, err error) {
	return nil, nil
}

// GetKey mock get key
func (m *MockCryptoSuite) GetKey(ski []byte) (k bccsp.Key, err error) {
	return nil, nil
}

// Hash mock hash
func (m *MockCryptoSuite) Hash(msg []byte, opts bccsp.HashOpts) (hash []byte, err error) {
	return nil, nil
}

// GetHash mock get hash
func (m *MockCryptoSuite) GetHash(opts bccsp.HashOpts) (h hash.Hash, err error) {
	return nil, nil
}

// Sign mock signing
func (m *MockCryptoSuite) Sign(k bccsp.Key, digest []byte,
	opts bccsp.SignerOpts) (signature []byte, err error) {
	return []byte("testSignature"), nil
}

// Verify mock verify
func (m *MockCryptoSuite) Verify(k bccsp.Key, signature, digest []byte,
	opts bccsp.SignerOpts) (valid bool, err error) {
	return false, nil
}

// Encrypt mock encrypt
func (m *MockCryptoSuite) Encrypt(k bccsp.Key, plaintext []byte,
	opts bccsp.EncrypterOpts) (ciphertext []byte, err error) {
	return nil, nil
}

// Decrypt mock decrypt
func (m *MockCryptoSuite) Decrypt(k bccsp.Key, ciphertext []byte,
	opts bccsp.DecrypterOpts) (plaintext []byte, err error) {
	return nil, nil
}
