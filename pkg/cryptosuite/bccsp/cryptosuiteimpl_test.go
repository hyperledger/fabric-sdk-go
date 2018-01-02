/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package bccsp

import (
	"errors"
	"hash"
	"testing"

	"strings"

	"github.com/golang/mock/gomock"
	"github.com/hyperledger/fabric-sdk-go/api/apiconfig/mocks"
	"github.com/hyperledger/fabric-sdk-go/api/apicryptosuite"
	"github.com/hyperledger/fabric-sdk-go/internal/github.com/hyperledger/fabric/bccsp"
	"github.com/hyperledger/fabric-sdk-go/pkg/logging/utils"
)

const (
	mockIdentifier   = "mock-test"
	signedIdentifier = "-signed"
	signingKey       = "signing-key"
	hashMessage      = "-msg-bytes"
	sampleKey        = "sample-key"
	getKey           = "-getkey"
	keyImport        = "-keyimport"
	keyGen           = "-keygent"
)

func TestCryptoSuite(t *testing.T) {

	//Get BCCSP implementation
	samplebccsp := getMockBCCSP(mockIdentifier)

	//Get cryptosuite
	samplecryptoSuite := GetSuite(samplebccsp)

	//Verify CryptSuite
	verifyCryptoSuite(t, samplecryptoSuite)

}

func TestCryptoSuiteByConfig(t *testing.T) {

	//Prepare Config
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()
	mockConfig := mock_apiconfig.NewMockConfig(mockCtrl)
	mockConfig.EXPECT().SecurityProvider().Return("SW")
	mockConfig.EXPECT().SecurityAlgorithm().Return("SHA2")
	mockConfig.EXPECT().SecurityLevel().Return(256)
	mockConfig.EXPECT().KeyStorePath().Return("/tmp/msp")
	mockConfig.EXPECT().Ephemeral().Return(false)

	//Get cryptosuite using config
	samplecryptoSuite, err := GetSuiteByConfig(mockConfig)
	utils.VerifyEmpty(t, err, "Not supposed to get error on GetSuiteByConfig call : %s", err)
	utils.VerifyNotEmpty(t, samplecryptoSuite, "Supposed to get valid cryptosuite")

	hashbytes, err := samplecryptoSuite.Hash([]byte(hashMessage), &bccsp.SHAOpts{})
	utils.VerifyEmpty(t, err, "Not supposed to get error on GetSuiteByConfig call : %s", err)
	utils.VerifyNotEmpty(t, hashbytes, "Supposed to get valid hash from sample cryptosuite")

}

func TestCryptoSuiteByConfigFailures(t *testing.T) {

	//Prepare Config
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()
	mockConfig := mock_apiconfig.NewMockConfig(mockCtrl)
	mockConfig.EXPECT().SecurityProvider().Return("SW")
	mockConfig.EXPECT().SecurityAlgorithm().Return("SHA2")
	mockConfig.EXPECT().SecurityLevel().Return(100)
	mockConfig.EXPECT().KeyStorePath().Return("/tmp/msp")
	mockConfig.EXPECT().Ephemeral().Return(false)

	//Get cryptosuite using config
	samplecryptoSuite, err := GetSuiteByConfig(mockConfig)
	utils.VerifyNotEmpty(t, err, "Supposed to get error on GetSuiteByConfig call : %s", err)
	utils.VerifyEmpty(t, samplecryptoSuite, "Not supposed to get valid cryptosuite")

	if !strings.HasPrefix(err.Error(), "Could not initialize BCCSP SW") {
		t.Fatalf("Didn't get expected failure, got %s instead", err)
	}

}

// TestCreateInvalidBCCSPSecurityLevel will test cryptsuite creation with invalid BCCSP options
func TestCreateInvalidBCCSPSecurityLevel(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()
	mockConfig := mock_apiconfig.NewMockConfig(mockCtrl)

	mockConfig.EXPECT().SecurityProvider().Return("SW")
	mockConfig.EXPECT().SecurityAlgorithm().Return("SHA2")
	mockConfig.EXPECT().SecurityLevel().Return(100)
	mockConfig.EXPECT().KeyStorePath().Return("/tmp/msp")
	mockConfig.EXPECT().Ephemeral().Return(false)

	_, err := GetSuiteByConfig(mockConfig)
	if !strings.Contains(err.Error(), "Security level not supported [100]") {
		t.Fatalf("Expected invalid security level error, but got %v", err.Error())
	}

}

// TestCreateInvalidBCCSPHashFamily will test cryptsuite creation with bad HashFamily
func TestCreateInvalidBCCSPHashFamily(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()
	mockConfig := mock_apiconfig.NewMockConfig(mockCtrl)

	mockConfig.EXPECT().SecurityProvider().Return("SW")
	mockConfig.EXPECT().SecurityAlgorithm().Return("ABC")
	mockConfig.EXPECT().SecurityLevel().Return(256)
	mockConfig.EXPECT().KeyStorePath().Return("/tmp/msp")
	mockConfig.EXPECT().Ephemeral().Return(false)

	_, err := GetSuiteByConfig(mockConfig)
	if !strings.Contains(err.Error(), "Hash Family not supported [ABC]") {
		t.Fatalf("Expected invalid hash family error, but got %v", err.Error())
	}
}

// TestCreateInvalidSecurityProviderPanic will test cryptsuite creation with bad HashFamily
func TestCreateInvalidSecurityProviderPanic(t *testing.T) {

	defer func() {
		if r := recover(); r == nil {
			t.Errorf("was supposed to panic")
		}
	}()

	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()
	mockConfig := mock_apiconfig.NewMockConfig(mockCtrl)

	mockConfig.EXPECT().SecurityProvider().Return("XYZ")
	mockConfig.EXPECT().SecurityProvider().Return("XYZ")

	GetSuiteByConfig(mockConfig)
	t.Fatalf("Getting cryptosuite with invalid security provider supposed to panic")
}

func verifyCryptoSuite(t *testing.T, samplecryptoSuite apicryptosuite.CryptoSuite) {
	//Test cryptosuite.Sign
	signedBytes, err := samplecryptoSuite.Sign(GetKey(getMockKey(signingKey)), nil, nil)
	utils.VerifyEmpty(t, err, "Not supposed to get any error for samplecryptoSuite.GetKey : %s", err)
	utils.VerifyTrue(t, string(signedBytes) == mockIdentifier+signedIdentifier, "Got unexpected result from samplecryptoSuite.Sign")

	//Test cryptosuite.Hash
	hashBytes, err := samplecryptoSuite.Hash([]byte(hashMessage), &bccsp.SHAOpts{})
	utils.VerifyEmpty(t, err, "Not supposed to get any error for samplecryptoSuite.GetKey")
	utils.VerifyTrue(t, string(hashBytes) == mockIdentifier+hashMessage, "Got unexpected result from samplecryptoSuite.Hash")

	//Test cryptosuite.GetKey
	key, err := samplecryptoSuite.GetKey([]byte(sampleKey))
	utils.VerifyEmpty(t, err, "Not supposed to get any error for samplecryptoSuite.GetKey")
	utils.VerifyNotEmpty(t, key, "Not supposed to get empty key for samplecryptoSuite.GetKey")

	keyBytes, err := key.Bytes()
	utils.VerifyEmpty(t, err, "Not supposed to get any error for samplecryptoSuite.GetKey().GetBytes()")
	utils.VerifyTrue(t, string(keyBytes) == sampleKey+getKey, "Not supposed to get empty bytes for samplecryptoSuite.GetKey().GetBytes()")

	skiBytes := key.SKI()
	utils.VerifyTrue(t, string(skiBytes) == sampleKey+getKey, "Not supposed to get empty bytes for samplecryptoSuite.GetKey().GetSKI()")

	utils.VerifyTrue(t, key.Private(), "Not supposed to get false for samplecryptoSuite.GetKey().Private()")
	utils.VerifyTrue(t, key.Symmetric(), "Not supposed to get false for samplecryptoSuite.GetKey().Symmetric()")

	publikey, err := key.PublicKey()
	utils.VerifyEmpty(t, err, "Not supposed to get any error for samplecryptoSuite.GetKey().PublicKey()")
	utils.VerifyNotEmpty(t, publikey, "Not supposed to get empty key for samplecryptoSuite.GetKey().PublicKey()")

	//Test cryptosuite.KeyImport
	key, err = samplecryptoSuite.KeyImport(nil, &bccsp.X509PublicKeyImportOpts{Temporary: true})
	utils.VerifyEmpty(t, err, "Not supposed to get any error for samplecryptoSuite.KeyImport")
	utils.VerifyNotEmpty(t, key, "Not supposed to get empty key for samplecryptoSuite.KeyImport")

	keyBytes, err = key.Bytes()
	utils.VerifyEmpty(t, err, "Not supposed to get any error for samplecryptoSuite.KeyImport().GetBytes()")
	utils.VerifyTrue(t, string(keyBytes) == mockIdentifier+keyImport, "Unexpected bytes for samplecryptoSuite.KeyImport().GetBytes()")

	skiBytes = key.SKI()
	utils.VerifyTrue(t, string(skiBytes) == mockIdentifier+keyImport, "Unexpected bytes for samplecryptoSuite.KeyImport().GetSKI()")

	utils.VerifyTrue(t, key.Private(), "Not supposed to get false for samplecryptoSuite.KeyImport().Private()")
	utils.VerifyTrue(t, key.Symmetric(), "Not supposed to get false for samplecryptoSuite.KeyImport().Symmetric()")

	publikey, err = key.PublicKey()
	utils.VerifyEmpty(t, err, "Not supposed to get any error for samplecryptoSuite.KeyImport().PublicKey()")
	utils.VerifyNotEmpty(t, publikey, "Not supposed to get empty key for samplecryptoSuite.KeyImport().PublicKey()")

	//Test cryptosuite.KeyGen
	key, err = samplecryptoSuite.KeyGen(nil)
	utils.VerifyEmpty(t, err, "Not supposed to get any error for samplecryptoSuite.KeyGen")
	utils.VerifyNotEmpty(t, key, "Not supposed to get empty key for samplecryptoSuite.KeyGen")

	keyBytes, err = key.Bytes()
	utils.VerifyEmpty(t, err, "Not supposed to get any error for samplecryptoSuite.KeyGen().GetBytes()")
	utils.VerifyTrue(t, string(keyBytes) == mockIdentifier+keyGen, "Unexpected bytes for samplecryptoSuite.KeyGen().GetBytes()")

	skiBytes = key.SKI()
	utils.VerifyTrue(t, string(skiBytes) == mockIdentifier+keyGen, "Unexpected bytes for samplecryptoSuite.KeyGen().GetSKI()")

	utils.VerifyTrue(t, key.Private(), "Not supposed to get false for samplecryptoSuite.KeyGen().Private()")
	utils.VerifyTrue(t, key.Symmetric(), "Not supposed to get false for samplecryptoSuite.KeyGen().Symmetric()")

	publikey, err = key.PublicKey()
	utils.VerifyEmpty(t, err, "Not supposed to get any error for samplecryptoSuite.KeyGen().PublicKey()")
	utils.VerifyNotEmpty(t, publikey, "Not supposed to get empty key for samplecryptoSuite.KeyGen().PublicKey()")

	//Test cryptosuite.GetHash
	hash, err := samplecryptoSuite.GetHash(&bccsp.SHA256Opts{})
	utils.VerifyNotEmpty(t, err, "Supposed to get error for samplecryptoSuite.GetHash")
	utils.VerifyEmpty(t, hash, "Supposed to get empty hash for samplecryptoSuite.GetHash")

	//Test cryptosuite.GetHash
	valid, err := samplecryptoSuite.Verify(GetKey(getMockKey(signingKey)), nil, nil, nil)
	utils.VerifyEmpty(t, err, "Not supposed to get error for samplecryptoSuite.Verify")
	utils.VerifyTrue(t, valid, "Supposed to get true for samplecryptoSuite.Verify")
}

/*
	Mock implementation of bccsp.BCCSP and bccsp.Key
*/

func getMockBCCSP(identifier string) bccsp.BCCSP {
	return &mockBCCSP{identifier}
}

func getMockKey(identifier string) bccsp.Key {
	return &mockKey{identifier}
}

type mockBCCSP struct {
	identifier string
}

func (mock *mockBCCSP) KeyGen(opts bccsp.KeyGenOpts) (k bccsp.Key, err error) {
	return &mockKey{mock.identifier + keyGen}, nil
}

func (mock *mockBCCSP) KeyDeriv(k bccsp.Key, opts bccsp.KeyDerivOpts) (dk bccsp.Key, err error) {
	return &mockKey{"keyderiv"}, nil
}

func (mock *mockBCCSP) KeyImport(raw interface{}, opts bccsp.KeyImportOpts) (k bccsp.Key, err error) {
	return &mockKey{mock.identifier + keyImport}, nil
}

func (mock *mockBCCSP) GetKey(ski []byte) (k bccsp.Key, err error) {
	return &mockKey{string(ski) + getKey}, nil
}

func (mock *mockBCCSP) Hash(msg []byte, opts bccsp.HashOpts) (hash []byte, err error) {
	return []byte(mock.identifier + string(msg)), nil
}

func (mock *mockBCCSP) GetHash(opts bccsp.HashOpts) (h hash.Hash, err error) {
	return nil, errors.New("Not able to Get Hash")
}

func (mock *mockBCCSP) Sign(k bccsp.Key, digest []byte, opts bccsp.SignerOpts) (signature []byte, err error) {
	return []byte(mock.identifier + signedIdentifier), nil
}

func (mock *mockBCCSP) Verify(k bccsp.Key, signature, digest []byte, opts bccsp.SignerOpts) (valid bool, err error) {
	return true, nil
}

func (mock *mockBCCSP) Encrypt(k bccsp.Key, plaintext []byte, opts bccsp.EncrypterOpts) (ciphertext []byte, err error) {
	return []byte(mock.identifier + "-encrypted"), nil
}

func (mock *mockBCCSP) Decrypt(k bccsp.Key, ciphertext []byte, opts bccsp.DecrypterOpts) (plaintext []byte, err error) {
	return []byte(mock.identifier + "-decrypted"), nil
}

type mockKey struct {
	identifier string
}

func (k *mockKey) Bytes() ([]byte, error) {
	return []byte(k.identifier), nil
}

func (k *mockKey) SKI() []byte {
	return []byte(k.identifier)
}

func (k *mockKey) Symmetric() bool {
	return true
}

func (k *mockKey) Private() bool {
	return true
}

func (k *mockKey) PublicKey() (bccsp.Key, error) {
	return &mockKey{k.identifier + "-public"}, nil
}
