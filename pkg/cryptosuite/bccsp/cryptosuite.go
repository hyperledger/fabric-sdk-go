/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package bccsp

import (
	"fmt"

	"encoding/json"

	"github.com/hyperledger/fabric-sdk-go/api/apiconfig"
	"github.com/hyperledger/fabric-sdk-go/api/apicryptosuite"
	"github.com/hyperledger/fabric-sdk-go/internal/github.com/hyperledger/fabric/bccsp"
	bccspFactory "github.com/hyperledger/fabric-sdk-go/internal/github.com/hyperledger/fabric/bccsp/factory"
	"github.com/hyperledger/fabric-sdk-go/internal/github.com/hyperledger/fabric/bccsp/pkcs11"
	"github.com/hyperledger/fabric-sdk-go/pkg/logging"
)

var logger = logging.NewLogger("fabric_sdk_go")

//GetSuite returns cryptosuite adaptor for given bccsp.BCCSP implementation
func GetSuite(bccsp bccsp.BCCSP) apicryptosuite.CryptoSuite {
	return &cryptoSuite{bccsp}
}

//GetSuiteByConfig returns cryptosuite adaptor for bccsp loaded according to given config
func GetSuiteByConfig(config apiconfig.Config) (apicryptosuite.CryptoSuite, error) {
	opts := getOptsByConfig(config)
	bccsp, err := bccspFactory.GetBCCSPFromOpts(opts)

	if err != nil {
		return nil, err
	}
	return &cryptoSuite{bccsp}, nil
}

//GetCryptoOptsJSON returns factory opts in json format
func GetCryptoOptsJSON(config apiconfig.Config) ([]byte, error) {
	opts := getOptsByConfig(config)
	jsonBytes, err := json.Marshal(opts)
	if err != nil {
		return nil, err
	}
	return jsonBytes, nil
}

//GetSHAOpts returns bccsp SHA hashing opts
func GetSHAOpts() apicryptosuite.HashOpts {
	return &bccsp.SHAOpts{}
}

func getOptsByConfig(c apiconfig.Config) *bccspFactory.FactoryOpts {
	var opts *bccspFactory.FactoryOpts

	switch c.SecurityProvider() {
	case "SW":
		opts = &bccspFactory.FactoryOpts{
			ProviderName: "SW",
			SwOpts: &bccspFactory.SwOpts{
				HashFamily: c.SecurityAlgorithm(),
				SecLevel:   c.SecurityLevel(),
				FileKeystore: &bccspFactory.FileKeystoreOpts{
					KeyStorePath: c.KeyStorePath(),
				},
				Ephemeral: c.Ephemeral(),
			},
		}
		logger.Debug("Initialized SW ")
		bccspFactory.InitFactories(opts)
		return opts

	case "PKCS11":
		pkks := pkcs11.FileKeystoreOpts{KeyStorePath: c.KeyStorePath()}
		opts = &bccspFactory.FactoryOpts{
			ProviderName: "PKCS11",
			Pkcs11Opts: &pkcs11.PKCS11Opts{
				SecLevel:     c.SecurityLevel(),
				HashFamily:   c.SecurityAlgorithm(),
				Ephemeral:    c.Ephemeral(),
				FileKeystore: &pkks,
				Library:      c.SecurityProviderLibPath(),
				Pin:          c.SecurityProviderPin(),
				Label:        c.SecurityProviderLabel(),
				SoftVerify:   c.SoftVerify(),
			},
		}
		logger.Debug("Initialized PKCS11 ")
		bccspFactory.InitFactories(opts)
		return opts
	default:
		panic(fmt.Sprintf("Unsupported BCCSP Provider: %s", c.SecurityProvider()))

	}
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
