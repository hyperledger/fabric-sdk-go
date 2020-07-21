/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package msp

import (
	"bytes"
	"path/filepath"
	"testing"

	"github.com/hyperledger/fabric-sdk-go/internal/github.com/hyperledger/fabric-ca/sdkinternal/pkg/util"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/core"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/msp"
	"github.com/hyperledger/fabric-sdk-go/pkg/core/config"
	"github.com/hyperledger/fabric-sdk-go/pkg/core/cryptosuite"
	cryptosuiteimpl "github.com/hyperledger/fabric-sdk-go/pkg/core/cryptosuite/bccsp/sw"
	"github.com/hyperledger/fabric-sdk-go/test/metadata"
)

func TestUserMethods(t *testing.T) {

	testUserMSPID := "testUserMSPID"
	testUsername := "testUsername"

	configPath := filepath.Join(metadata.GetProjectPath(), metadata.SDKConfigPath, configTestFile)
	configBackend, err := config.FromFile(configPath)()
	if err != nil {
		t.Fatalf("Failed to read config: %s", err)
	}
	cryptoConfig := cryptosuite.ConfigFromBackend(configBackend...)
	if err != nil {
		t.Fatalf("Failed to read config: %s", err)
	}
	identityConfig, err := ConfigFromBackend(configBackend...)
	if err != nil {
		t.Fatalf("Failed to read config: %s", err)
	}
	// Delete all private keys from the crypto suite store
	// and users from the user store
	cleanupTestPath(t, cryptoConfig.KeyStorePath())
	defer cleanupTestPath(t, cryptoConfig.KeyStorePath())
	cleanupTestPath(t, identityConfig.CredentialStorePath())
	defer cleanupTestPath(t, identityConfig.CredentialStorePath())

	cryptoSuite, err := cryptosuiteimpl.GetSuiteByConfig(cryptoConfig)
	if cryptoSuite == nil {
		t.Fatalf("Failed initialize cryptoSuite: %s", err)
	}

	// Missing enrollment cert
	userData := &msp.UserData{
		MSPID: testUserMSPID,
		ID:    testUsername,
	}
	_, err = newUser(userData, cryptoSuite)
	if err == nil {
		t.Fatal("Expected newUser to fail when missing enrollment cert")
	}

	// User not enrolled (have cert, but private key is not in crypto store)
	userData.EnrollmentCertificate = generatedCertBytes
	_, err = newUser(userData, cryptoSuite)
	if err == nil {
		t.Fatal("Expected newUser to fail when user is not enrolled")
	}

	// Import the key into the crypto suite's private key storage.
	// This is normally done when a new user in enrolled
	verifyUserIdentity(cryptoSuite, t, userData, testUserMSPID, testUsername)

}

func verifyUserIdentity(cryptoSuite core.CryptoSuite, t *testing.T, userData *msp.UserData, testUserMSPID string, testUsername string) {
	generatedKey, err := util.ImportBCCSPKeyFromPEMBytes(generatedKeyBytes, cryptoSuite, false)
	if err != nil {
		t.Fatalf("ImportBCCSPKeyFromPEMBytes failed %s", err)
	}
	user, err := newUser(userData, cryptoSuite)
	if err != nil {
		t.Fatalf("newUser failed: %s", err)
	}
	// Check MSPID
	if user.Identifier().MSPID != testUserMSPID {
		t.Fatal("user.SetMSPID Failed to MSP.")
	}
	// Check Name
	if user.Identifier().ID != testUsername {
		t.Fatal("NewUser create wrong user")
	}
	// Check EnrolmentCert
	verifyBytes(t, user.EnrollmentCertificate(), generatedCertBytes)
	// Check PrivateKey
	verifyBytes(t, user.PrivateKey().SKI(), generatedKey.SKI())
}

func verifyBytes(t *testing.T, v interface{}, expected []byte) error {
	var vbytes []byte
	var ok bool
	if v == nil {
		vbytes = nil
	} else {
		vbytes, ok = v.([]byte)
		if !ok {
			t.Fatal("value is not []byte")
		}
	}
	if !bytes.Equal(vbytes, expected) {
		t.Fatal("value from store comparison failed")
	}
	return nil
}
