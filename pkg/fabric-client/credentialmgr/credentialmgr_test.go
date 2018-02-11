/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package credentialmgr

import (
	"testing"

	"github.com/hyperledger/fabric-sdk-go/api/apifabclient"
	"github.com/hyperledger/fabric-sdk-go/pkg/config"
	"github.com/hyperledger/fabric-sdk-go/pkg/cryptosuite"
	fcmocks "github.com/hyperledger/fabric-sdk-go/pkg/fabric-client/mocks"
	"github.com/pkg/errors"
)

func TestCredentialManager(t *testing.T) {

	config, err := config.FromFile("../../../test/fixtures/config/config_test.yaml")()
	if err != nil {
		t.Fatalf(err.Error())
	}

	credentialMgr, err := NewCredentialManager("Org1", config, cryptosuite.GetDefault())
	if err != nil {
		t.Fatalf("Failed to setup credential manager: %s", err)
	}

	_, err = credentialMgr.GetSigningIdentity("")
	if err == nil {
		t.Fatalf("Should have failed to retrieve signing identity for empty user name")
	}

	_, err = credentialMgr.GetSigningIdentity("Non-Existent")
	if err == nil {
		t.Fatalf("Should have failed to retrieve signing identity for non-existent user")
	}

	if err := checkSigningIdentity(credentialMgr, "User1"); err != nil {
		t.Fatalf("checkSigningIdentity failes: %s", err)
	}
}

func checkSigningIdentity(credentialMgr apifabclient.CredentialManager, user string) error {
	id, err := credentialMgr.GetSigningIdentity(user)
	if err != nil {
		return errors.Wrapf(err, "Failed to retrieve signing identity: %s", err)
	}

	if id == nil {
		return errors.New("SigningIdentity is nil")
	}
	if id.EnrollmentCert == nil {
		return errors.New("Enrollment cert is missing")
	}
	if id.MspID == "" {
		return errors.New("MspID is missing")
	}
	if id.PrivateKey == nil {
		return errors.New("private key is missing")
	}
	return nil
}

func TestInvalidOrgCredentialManager(t *testing.T) {

	config, err := config.FromFile("../../../test/fixtures/config/config_test.yaml")()
	if err != nil {
		t.Fatalf(err.Error())
	}

	// Invalid Org
	_, err = NewCredentialManager("invalidOrg", config, &fcmocks.MockCryptoSuite{})
	if err == nil {
		t.Fatalf("Should have failed to setup manager for invalid org")
	}

}

func TestCredentialManagerFromEmbeddedCryptoConfig(t *testing.T) {
	config, err := config.FromFile("../../../test/fixtures/config/config_test_embedded_pems.yaml")()

	if err != nil {
		t.Fatalf(err.Error())
	}

	credentialMgr, err := NewCredentialManager("Org1", config, cryptosuite.GetDefault())
	if err != nil {
		t.Fatalf("Failed to setup credential manager: %s", err)
	}

	_, err = credentialMgr.GetSigningIdentity("")
	if err == nil {
		t.Fatalf("Should have failed to retrieve signing identity for empty user name")
	}

	_, err = credentialMgr.GetSigningIdentity("Non-Existent")
	if err == nil {
		t.Fatalf("Should have failed to retrieve signing identity for non-existent user")
	}

	if err := checkSigningIdentity(credentialMgr, "EmbeddedUser"); err != nil {
		t.Fatalf("checkSigningIdentity failes: %s", err)
	}

	if err := checkSigningIdentity(credentialMgr, "EmbeddedUserWithPaths"); err != nil {
		t.Fatalf("checkSigningIdentity failes: %s", err)
	}

	if err := checkSigningIdentity(credentialMgr, "EmbeddedUserMixed"); err != nil {
		t.Fatalf("checkSigningIdentity failes: %s", err)
	}

	if err := checkSigningIdentity(credentialMgr, "EmbeddedUserMixed2"); err != nil {
		t.Fatalf("checkSigningIdentity failes: %s", err)
	}
}
