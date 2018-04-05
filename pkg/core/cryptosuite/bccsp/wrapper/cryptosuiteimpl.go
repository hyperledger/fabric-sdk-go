/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package wrapper

import (
	"hash"

	"github.com/hyperledger/fabric-sdk-go/internal/github.com/hyperledger/fabric/bccsp"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/core"
)

//NewCryptoSuite returns cryptosuite adaptor for given bccsp.BCCSP implementation
func NewCryptoSuite(bccsp bccsp.BCCSP) core.CryptoSuite {
	return &CryptoSuite{
		BCCSP: bccsp,
	}
}

//GetKey returns implementation of of cryptosuite.Key
func GetKey(newkey bccsp.Key) core.Key {
	return &key{newkey}
}

// CryptoSuite provides a wrapper of BCCSP
type CryptoSuite struct {
	BCCSP bccsp.BCCSP
}

// KeyGen is a wrapper of BCCSP.KeyGen
func (c *CryptoSuite) KeyGen(opts core.KeyGenOpts) (k core.Key, err error) {
	key, err := c.BCCSP.KeyGen(opts)
	return GetKey(key), err
}

// KeyImport is a wrapper of BCCSP.KeyImport
func (c *CryptoSuite) KeyImport(raw interface{}, opts core.KeyImportOpts) (k core.Key, err error) {
	key, err := c.BCCSP.KeyImport(raw, opts)
	return GetKey(key), err
}

// GetKey is a wrapper of BCCSP.GetKey
func (c *CryptoSuite) GetKey(ski []byte) (k core.Key, err error) {
	key, err := c.BCCSP.GetKey(ski)
	return GetKey(key), err
}

// Hash is a wrapper of BCCSP.Hash
func (c *CryptoSuite) Hash(msg []byte, opts core.HashOpts) (hash []byte, err error) {
	return c.BCCSP.Hash(msg, opts)
}

// GetHash is a wrapper of BCCSP.GetHash
func (c *CryptoSuite) GetHash(opts core.HashOpts) (h hash.Hash, err error) {
	return c.BCCSP.GetHash(opts)
}

// Sign is a wrapper of BCCSP.Sign
func (c *CryptoSuite) Sign(k core.Key, digest []byte, opts core.SignerOpts) (signature []byte, err error) {
	return c.BCCSP.Sign(k.(*key).key, digest, opts)
}

// Verify is a wrapper of BCCSP.Verify
func (c *CryptoSuite) Verify(k core.Key, signature, digest []byte, opts core.SignerOpts) (valid bool, err error) {
	return c.BCCSP.Verify(k.(*key).key, signature, digest, opts)
}

type key struct {
	key bccsp.Key
}

func (k *key) Bytes() ([]byte, error) {
	return k.key.Bytes()
}

func (k *key) SKI() []byte {
	return k.key.SKI()
}

func (k *key) Symmetric() bool {
	return k.key.Symmetric()
}

func (k *key) Private() bool {
	return k.key.Private()
}

func (k *key) PublicKey() (core.Key, error) {
	key, err := k.key.PublicKey()
	return GetKey(key), err
}
