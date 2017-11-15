/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package cryptosuite

import (
	"github.com/hyperledger/fabric-sdk-go/api/apicryptosuite"
	"github.com/hyperledger/fabric-sdk-go/third_party/github.com/hyperledger/fabric/bccsp"
)

//GetSuite returns cryptosuite adaptor
func GetSuite(bccsp bccsp.BCCSP) apicryptosuite.CryptoSuite {
	return &cryptoSuite{bccsp}
}

//GetKey returns implementation of of cryptosuite.Key
func GetKey(newkey bccsp.Key) apicryptosuite.Key {
	return &key{newkey}
}

type cryptoSuite struct {
	bccsp bccsp.BCCSP
}

func (c *cryptoSuite) KeyGen(opts apicryptosuite.KeyGenOpts) (k apicryptosuite.Key, err error) {
	key, err := c.bccsp.KeyGen(opts)
	return GetKey(key), err
}

func (c *cryptoSuite) KeyImport(raw interface{}, opts apicryptosuite.KeyImportOpts) (k apicryptosuite.Key, err error) {
	key, err := c.bccsp.KeyImport(raw, opts)
	return GetKey(key), err
}

func (c *cryptoSuite) GetKey(ski []byte) (k apicryptosuite.Key, err error) {
	key, err := c.bccsp.GetKey(ski)
	return GetKey(key), err
}

func (c *cryptoSuite) Hash(msg []byte, opts apicryptosuite.HashOpts) (hash []byte, err error) {
	return c.bccsp.Hash(msg, opts)
}

func (c *cryptoSuite) Sign(k apicryptosuite.Key, digest []byte, opts apicryptosuite.SignerOpts) (signature []byte, err error) {
	return c.bccsp.Sign(k.(*key).key, digest, opts)
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

func (k *key) PublicKey() (apicryptosuite.Key, error) {
	key, err := k.key.PublicKey()
	return GetKey(key), err
}
