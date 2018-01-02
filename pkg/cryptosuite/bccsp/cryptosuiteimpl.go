/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package bccsp

import (
	"hash"

	"github.com/hyperledger/fabric-sdk-go/api/apiconfig"
	"github.com/hyperledger/fabric-sdk-go/api/apicryptosuite"
	"github.com/hyperledger/fabric-sdk-go/internal/github.com/hyperledger/fabric/bccsp"
	bccspFactory "github.com/hyperledger/fabric-sdk-go/internal/github.com/hyperledger/fabric/bccsp/factory"
	"github.com/hyperledger/fabric-sdk-go/pkg/logging"
)

var logger = logging.NewLogger("fabric_sdk_go")

//GetSuite returns cryptosuite adaptor for given bccsp.BCCSP implementation
func GetSuite(bccsp bccsp.BCCSP) apicryptosuite.CryptoSuite {
	return &cryptoSuite{bccsp}
}

//GetSuiteByConfig returns cryptosuite adaptor for bccsp loaded according to given config
func GetSuiteByConfig(config apiconfig.Config) (apicryptosuite.CryptoSuite, error) {
	opts := GetOptsByConfig(config)
	bccsp, err := bccspFactory.GetBCCSPFromOpts(opts)

	if err != nil {
		return nil, err
	}
	return &cryptoSuite{bccsp}, nil
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

func (c *cryptoSuite) GetHash(opts apicryptosuite.HashOpts) (h hash.Hash, err error) {
	return c.bccsp.GetHash(opts)
}

func (c *cryptoSuite) Sign(k apicryptosuite.Key, digest []byte, opts apicryptosuite.SignerOpts) (signature []byte, err error) {
	return c.bccsp.Sign(k.(*key).key, digest, opts)
}

func (c *cryptoSuite) Verify(k apicryptosuite.Key, signature, digest []byte, opts apicryptosuite.SignerOpts) (valid bool, err error) {
	return c.bccsp.Verify(k.(*key).key, signature, digest, opts)
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
