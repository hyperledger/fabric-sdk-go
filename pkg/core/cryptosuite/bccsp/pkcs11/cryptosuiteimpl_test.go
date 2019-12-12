/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package pkcs11

import (
	"bytes"
	"crypto/sha256"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"

	"github.com/hyperledger/fabric-sdk-go/internal/github.com/hyperledger/fabric/bccsp"
	pkcsFactory "github.com/hyperledger/fabric-sdk-go/internal/github.com/hyperledger/fabric/bccsp/factory/pkcs11"
	"github.com/hyperledger/fabric-sdk-go/internal/github.com/hyperledger/fabric/bccsp/pkcs11"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/core"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/test/mockcore"
)

var securityLevel = 256

const (
	providerTypePKCS11 = "PKCS11"
)

func TestBadConfig(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	mockConfig := mockcore.NewMockCryptoSuiteConfig(mockCtrl)
	mockConfig.EXPECT().SecurityProvider().Return("UNKNOWN")
	mockConfig.EXPECT().SecurityProvider().Return("UNKNOWN")

	//Get cryptosuite using config
	_, err := GetSuiteByConfig(mockConfig)
	if err == nil {
		t.Fatal("Unknown security provider should return error")
	}
}
func TestCryptoSuiteByConfigPKCS11(t *testing.T) {

	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	//Prepare Config
	providerLib, softHSMPin, softHSMTokenLabel := pkcs11.FindPKCS11Lib()

	mockConfig := mockcore.NewMockCryptoSuiteConfig(mockCtrl)
	mockConfig.EXPECT().SecurityProvider().Return("pkcs11")
	mockConfig.EXPECT().SecurityAlgorithm().Return("SHA2")
	mockConfig.EXPECT().SecurityLevel().Return(256)
	mockConfig.EXPECT().SecurityProviderLibPath().Return(providerLib)
	mockConfig.EXPECT().SecurityProviderLabel().Return(softHSMTokenLabel)
	mockConfig.EXPECT().SecurityProviderPin().Return(softHSMPin)
	mockConfig.EXPECT().SoftVerify().Return(true)

	//Get cryptosuite using config
	c, err := GetSuiteByConfig(mockConfig)
	if err != nil {
		t.Fatalf("Not supposed to get error, but got: %s", err)
	}

	verifyHashFn(t, c)
}

func TestCryptoSuiteByConfigPKCS11Failure(t *testing.T) {

	//Prepare Config
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()
	//Prepare Config
	mockConfig := mockcore.NewMockCryptoSuiteConfig(mockCtrl)
	mockConfig.EXPECT().SecurityProvider().Return("pkcs11")
	mockConfig.EXPECT().SecurityAlgorithm().Return("SHA2")
	mockConfig.EXPECT().SecurityLevel().Return(256)
	mockConfig.EXPECT().SecurityProviderLibPath().Return("")
	mockConfig.EXPECT().SecurityProviderLabel().Return("")
	mockConfig.EXPECT().SecurityProviderPin().Return("")
	mockConfig.EXPECT().SoftVerify().Return(true)

	//Get cryptosuite using config
	samplecryptoSuite, err := GetSuiteByConfig(mockConfig)
	assert.NotEmpty(t, err, "Supposed to get error on GetSuiteByConfig call : %s", err)
	assert.Empty(t, samplecryptoSuite, "Not supposed to get valid cryptosuite")
}

func TestPKCS11CSPConfigWithValidOptions(t *testing.T) {
	opts := configurePKCS11Options("SHA2", securityLevel)
	f := &pkcsFactory.PKCS11Factory{}

	csp, err := f.Get(opts)
	if err != nil {
		t.Fatalf(err.Error())
	}
	if csp == nil {
		t.Fatal("BCCSP PKCS11 was not configured")
	}
	t.Logf("TestPKCS11CSPConfigWithValidOptions passed. BCCSP PKCS11 provider was configured (%+v)", csp)

}

func TestPKCS11CSPConfigWithEmptyHashFamily(t *testing.T) {

	opts := configurePKCS11Options("", securityLevel)

	f := &pkcsFactory.PKCS11Factory{}
	t.Logf("PKCS11 factory name: %s", f.Name())
	_, err := f.Get(opts)
	if err == nil {
		t.Fatal("Expected error 'Hash Family not supported'")
	}
	t.Log("TestPKCS11CSPConfigWithEmptyHashFamily passed.")

}

func TestPKCS11CSPConfigWithIncorrectLevel(t *testing.T) {

	opts := configurePKCS11Options("SHA2", 100)

	f := &pkcsFactory.PKCS11Factory{}
	t.Logf("PKCS11 factory name: %s", f.Name())
	_, err := f.Get(opts)
	if err == nil {
		t.Fatal("Expected error 'Failed initializing configuration'")
	}

}

func TestPKCS11CSPConfigWithEmptyProviderName(t *testing.T) {
	f := &pkcsFactory.PKCS11Factory{}
	if f.Name() != providerTypePKCS11 {
		t.Fatalf("Expected default name for PKCS11. Got %s", f.Name())
	}
}

func configurePKCS11Options(hashFamily string, securityLevel int) *pkcs11.PKCS11Opts {
	providerLib, softHSMPin, softHSMTokenLabel := pkcs11.FindPKCS11Lib()

	//PKCS11 options
	pkcsOpt := pkcs11.PKCS11Opts{
		SecLevel:   securityLevel,
		HashFamily: hashFamily,
		Library:    providerLib,
		Pin:        softHSMPin,
		Label:      softHSMTokenLabel,
		Ephemeral:  false,
	}

	return &pkcsOpt

}

func verifyHashFn(t *testing.T, c core.CryptoSuite) {
	msg := []byte("Hello")
	e := sha256.Sum256(msg)
	a, err := c.Hash(msg, &bccsp.SHA256Opts{})
	if err != nil {
		t.Fatalf("Not supposed to get error, but got: %s", err)
	}

	if !bytes.Equal(a, e[:]) {
		t.Fatal("Expected SHA 256 hash function")
	}
}
