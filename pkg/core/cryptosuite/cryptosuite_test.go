/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package cryptosuite

import (
	"testing"

	"sync/atomic"

	"github.com/hyperledger/fabric-sdk-go/pkg/core/cryptosuite/bccsp/sw"
	"github.com/hyperledger/fabric-sdk-go/pkg/logging/testutils"
)

const (
	shaHashOptsAlgorithm       = "SHA"
	sha256HashOptsAlgorithm    = "SHA256"
	ecdsap256KeyGenOpts        = "ECDSAP256"
	setDefAlreadySetErrorMsg   = "default crypto suite is already set"
	InvalidDefSuiteSetErrorMsg = "attempting to set invalid default suite"
)

func TestGetDefault(t *testing.T) {

	//At the beginning default suite is nil if no attempts have been made to set or get one
	testutils.VerifyEmpty(t, defaultCryptoSuite, "default suite should be nil if no attempts have been made to set or get one")

	//Now try to get default, it will create one and return
	defSuite := GetDefault()
	testutils.VerifyNotEmpty(t, defSuite, "Not supposed to be nil defaultCryptSuite")
	testutils.VerifyNotEmpty(t, defaultCryptoSuite, "default suite should have been initialized")
	testutils.VerifyTrue(t, atomic.LoadInt32(&initialized) > 0, "'initialized' flag supposed to be set to 1")

	hashbytes, err := defSuite.Hash([]byte("Sample message"), GetSHAOpts())
	testutils.VerifyEmpty(t, err, "Not supposed to get error on defaultCryptSuite.Hash() call : %s", err)
	testutils.VerifyNotEmpty(t, hashbytes, "Supposed to get valid hash from defaultCryptSuite.Hash()")

	//Now try to get default, which is already created
	defSuite = GetDefault()
	testutils.VerifyNotEmpty(t, defSuite, "Not supposed to be nil defaultCryptSuite")
	testutils.VerifyNotEmpty(t, defaultCryptoSuite, "default suite should have been initialized")
	testutils.VerifyTrue(t, atomic.LoadInt32(&initialized) > 0, "'initialized' flag supposed to be set to 1")

	hashbytes, err = defSuite.Hash([]byte("Sample message"), GetSHAOpts())
	testutils.VerifyEmpty(t, err, "Not supposed to get error on defaultCryptSuite.Hash() call : %s", err)
	testutils.VerifyNotEmpty(t, hashbytes, "Supposed to get valid hash from defaultCryptSuite.Hash()")

	//Now attempt to set default suite
	err = SetDefault(nil)
	testutils.VerifyNotEmpty(t, err, "supposed to get error when SetDefault() gets called after GetDefault()")
	testutils.VerifyTrue(t, err.Error() == setDefAlreadySetErrorMsg, "unexpected error : expected [%s], got [%s]", setDefAlreadySetErrorMsg, err.Error())

	//Reset
	defaultCryptoSuite = nil
	atomic.StoreInt32(&initialized, 0)

	//Now attempt to set invalid default suite
	err = SetDefault(nil)
	testutils.VerifyNotEmpty(t, err, "supposed to get error when invalid default suite is set")
	testutils.VerifyTrue(t, err.Error() == InvalidDefSuiteSetErrorMsg, "unexpected error : expected [%s], got [%s]", InvalidDefSuiteSetErrorMsg, err.Error())

	s, err := sw.GetSuiteWithDefaultEphemeral()
	if err != nil {
		t.Fatalf("Unable to get default cryptosuite")
	}

	err = SetDefault(s)
	testutils.VerifyEmpty(t, err, "Not supposed to get error when valid default suite is set")

}

func TestHashOpts(t *testing.T) {

	//Get CryptoSuite SHA Opts
	hashOpts := GetSHAOpts()
	testutils.VerifyNotEmpty(t, hashOpts, "Not supposed to be empty shaHashOpts")
	testutils.VerifyTrue(t, hashOpts.Algorithm() == shaHashOptsAlgorithm, "Unexpected SHA hash opts, expected [%s], got [%s]", shaHashOptsAlgorithm, hashOpts.Algorithm())

	//Get CryptoSuite SHA256 Opts
	hashOpts = GetSHA256Opts()
	testutils.VerifyNotEmpty(t, hashOpts, "Not supposed to be empty sha256HashOpts")
	testutils.VerifyTrue(t, hashOpts.Algorithm() == sha256HashOptsAlgorithm, "Unexpected SHA hash opts, expected [%v], got [%v]", sha256HashOptsAlgorithm, hashOpts.Algorithm())

}

func TestKeyGenOpts(t *testing.T) {

	keygenOpts := GetECDSAP256KeyGenOpts(true)
	testutils.VerifyNotEmpty(t, keygenOpts, "Not supposed to be empty ECDSAP256KeyGenOpts")
	testutils.VerifyTrue(t, keygenOpts.Ephemeral(), "Expected keygenOpts.Ephemeral() ==> true")
	testutils.VerifyTrue(t, keygenOpts.Algorithm() == ecdsap256KeyGenOpts, "Unexpected SHA hash opts, expected [%v], got [%v]", ecdsap256KeyGenOpts, keygenOpts.Algorithm())

	keygenOpts = GetECDSAP256KeyGenOpts(false)
	testutils.VerifyNotEmpty(t, keygenOpts, "Not supposed to be empty ECDSAP256KeyGenOpts")
	testutils.VerifyFalse(t, keygenOpts.Ephemeral(), "Expected keygenOpts.Ephemeral() ==> false")
	testutils.VerifyTrue(t, keygenOpts.Algorithm() == ecdsap256KeyGenOpts, "Unexpected SHA hash opts, expected [%v], got [%v]", ecdsap256KeyGenOpts, keygenOpts.Algorithm())

}
