/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package integration

import (
	"fmt"
	"os"
	"testing"

	api "github.com/hyperledger/fabric-sdk-go/api/apiconfig"
	pkcsFactory "github.com/hyperledger/fabric-sdk-go/third_party/github.com/hyperledger/fabric/bccsp/factory"
	pkcs11 "github.com/hyperledger/fabric-sdk-go/third_party/github.com/hyperledger/fabric/bccsp/pkcs11"
)

var configImpl api.Config
var securityLevel = 256

const (
	providerTypePKCS11 = "PKCS11"
)

func TestPKCS11CSPConfigWithValidOptions(t *testing.T) {
	opts := configurePKCS11Options("SHA2", securityLevel)
	f := &pkcsFactory.PKCS11Factory{}
	//
	csp, err := f.Get(opts)
	if err != nil {
		t.Fatalf(err.Error())
	}
	if csp == nil {
		t.Fatalf("BCCSP PKCS11 was not configured")
	}
	fmt.Println("TestPKCS11CSPConfigWithValidOptions passed. BCCSP PKCS11 provider was configured\n", csp)

}

func TestPKCS11CSPConfigWithEmptyHashFamily(t *testing.T) {

	opts := configurePKCS11Options("", securityLevel)

	f := &pkcsFactory.PKCS11Factory{}
	fmt.Println(f.Name())
	_, err := f.Get(opts)
	if err == nil {
		t.Fatalf("Expected error 'Hash Family not supported'")
	}
	fmt.Println("TestPKCS11CSPConfigWithEmptyHashFamily passed. ")

}

func TestPKCS11CSPConfigWithIncorrectLevel(t *testing.T) {

	opts := configurePKCS11Options("SHA2", 100)

	f := &pkcsFactory.PKCS11Factory{}
	fmt.Println(f.Name())
	_, err := f.Get(opts)
	if err == nil {
		t.Fatalf("Expected error 'Failed initializing configuration'")
	}

}

func TestPKCS11CSPConfigWithEmptyProviderName(t *testing.T) {
	f := &pkcsFactory.PKCS11Factory{}
	if f.Name() != providerTypePKCS11 {
		t.Fatalf("Expected default name for PKCS11. Got %s", f.Name())
	}
}

func configurePKCS11Options(hashFamily string, securityLevel int) *pkcsFactory.FactoryOpts {
	providerLib, softHSMPin, softHSMTokenLabel := pkcs11.FindPKCS11Lib()

	pkks := pkcs11.FileKeystoreOpts{KeyStorePath: os.TempDir()}
	//PKCS11 options
	pkcsOpt := pkcs11.PKCS11Opts{
		SecLevel:     securityLevel,
		HashFamily:   hashFamily,
		FileKeystore: &pkks,
		Library:      providerLib,
		Pin:          softHSMPin,
		Label:        softHSMTokenLabel,
		Ephemeral:    false,
	}

	opts := &pkcsFactory.FactoryOpts{
		ProviderName: providerTypePKCS11,
		Pkcs11Opts:   &pkcsOpt,
	}
	pkcsFactory.InitFactories(opts)
	return opts

}
