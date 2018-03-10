/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package identitymgr

import (
	"bytes"
	"testing"

	"github.com/hyperledger/fabric-sdk-go/internal/github.com/hyperledger/fabric-ca/util"
	"github.com/hyperledger/fabric-sdk-go/pkg/context/api/msp"
	"github.com/hyperledger/fabric-sdk-go/pkg/core/config"
	cryptosuiteimpl "github.com/hyperledger/fabric-sdk-go/pkg/core/cryptosuite/bccsp/sw"
)

func TestUserMethods(t *testing.T) {

	testUserMspID := "testUserMspID"
	testUserName := "testUserName"

	config, err := config.FromFile("../../../test/fixtures/config/config_test.yaml")()
	if err != nil {
		t.Fatalf("Failed to read config: %v", err)
	}
	// Delete all private keys from the crypto suite store
	// and users from the user store
	cleanupTestPath(t, config.KeyStorePath())
	defer cleanupTestPath(t, config.KeyStorePath())
	cleanupTestPath(t, config.CredentialStorePath())
	defer cleanupTestPath(t, config.CredentialStorePath())

	cryptoSuite, err = cryptosuiteimpl.GetSuiteByConfig(config)
	if cryptoSuite == nil {
		t.Fatalf("Failed initialize cryptoSuite: %v", err)
	}

	// Missing enrollment cert
	userData := msp.UserData{
		MspID: testUserMspID,
		Name:  testUserName,
	}
	_, err = newUser(userData, cryptoSuite)
	if err == nil {
		t.Fatalf("Expected newUser to fail when missing enrollment cert")
	}

	// User not enrolled (have cert, but private key is not in crypto store)
	userData.EnrollmentCertificate = generatedCertBytes
	_, err = newUser(userData, cryptoSuite)
	if err == nil {
		t.Fatalf("Expected newUser to fail when user is not enrolled")
	}

	// Import the key into the crypto suite's private key storage.
	// This is normally done when a new user in enrolled
	generatedKey, err := util.ImportBCCSPKeyFromPEMBytes(generatedKeyBytes, cryptoSuite, false)

	user, err := newUser(userData, cryptoSuite)
	if err != nil {
		t.Fatalf("newUser failed: %v", err)
	}

	// Check MspID
	if user.MspID() != testUserMspID {
		t.Fatal("user.SetMspID Failed to MSP.")
	}

	// Check Name
	if user.Name() != testUserName {
		t.Fatalf("NewUser create wrong user")
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
			t.Fatalf("value is not []byte")
		}
	}
	if bytes.Compare(vbytes, expected) != 0 {
		t.Fatalf("value from store comparison failed")
	}
	return nil
}
