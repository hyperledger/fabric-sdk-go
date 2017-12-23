/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package credentialmgr

import (
	"testing"

	"github.com/hyperledger/fabric-sdk-go/pkg/config"
	fcmocks "github.com/hyperledger/fabric-sdk-go/pkg/fabric-client/mocks"
)

func TestCredentialManager(t *testing.T) {

	config, err := config.InitConfig("../../../test/fixtures/config/config_test.yaml")
	if err != nil {
		t.Fatalf(err.Error())
	}

	credentialMgr, err := NewCredentialManager("Org1", config, &fcmocks.MockCryptoSuite{})
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

	_, err = credentialMgr.GetSigningIdentity("User1")
	if err != nil {
		t.Fatalf("Failed to retrieve signing identity: %s", err)
	}

}

func TestInvalidOrgCredentialManager(t *testing.T) {

	config, err := config.InitConfig("../../../test/fixtures/config/config_test.yaml")
	if err != nil {
		t.Fatalf(err.Error())
	}

	// Invalid Org
	credentialMgr, err := NewCredentialManager("invalidOrg", config, &fcmocks.MockCryptoSuite{})
	if err == nil {
		t.Fatalf("Should have failed to setup manager for invalid org")
	}

	// Valid Org, Invalid User
	credentialMgr, err = NewCredentialManager("Org1", config, &fcmocks.MockCryptoSuite{})
	if err != nil {
		t.Fatalf("Failed to setup credential manager: %s", err)
	}
	_, err = credentialMgr.GetSigningIdentity("testUser")
	if err == nil {
		t.Fatalf("Should have failed to retrieve signing identity for invalid user name")
	}

}

func TestCredentialManagerFromEmbeddedCryptoConfig(t *testing.T) {
	config, err := config.InitConfig("../../../test/fixtures/config/config_test_embedded_pems.yaml")

	if err != nil {
		t.Fatalf(err.Error())
	}

	credentialMgr, err := NewCredentialManager("Org1", config, &fcmocks.MockCryptoSuite{})
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

	_, err = credentialMgr.GetSigningIdentity("EmbeddedUser")
	if err != nil {
		t.Fatalf("Failed to retrieve signing identity: %+v", err)
	}

	_, err = credentialMgr.GetSigningIdentity("EmbeddedUserWithPaths")
	if err != nil {
		t.Fatalf("Failed to retrieve signing identity: %+v", err)
	}

	_, err = credentialMgr.GetSigningIdentity("EmbeddedUserMixed")
	if err != nil {
		t.Fatalf("Failed to retrieve signing identity: %+v", err)
	}

	_, err = credentialMgr.GetSigningIdentity("EmbeddedUserMixed2")
	if err != nil {
		t.Fatalf("Failed to retrieve signing identity: %+v", err)
	}
}
