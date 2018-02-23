/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package credentialmgr

import (
	"math/rand"
	"os"
	"strconv"
	"strings"
	"testing"

	fabricCaUtil "github.com/hyperledger/fabric-sdk-go/internal/github.com/hyperledger/fabric-ca/util"
	"github.com/hyperledger/fabric-sdk-go/pkg/context/api"
	"github.com/hyperledger/fabric-sdk-go/pkg/core/config"
	"github.com/hyperledger/fabric-sdk-go/pkg/core/cryptosuite"
	"github.com/hyperledger/fabric-sdk-go/pkg/core/cryptosuite/bccsp/sw"
	"github.com/hyperledger/fabric-sdk-go/pkg/fab/identity"
	fcmocks "github.com/hyperledger/fabric-sdk-go/pkg/fab/mocks"
	"github.com/pkg/errors"
)

var testPrivKey = `-----BEGIN PRIVATE KEY-----
MIGHAgEAMBMGByqGSM49AgEGCCqGSM49AwEHBG0wawIBAQQgp4qKKB0WCEfx7XiB
5Ul+GpjM1P5rqc6RhjD5OkTgl5OhRANCAATyFT0voXX7cA4PPtNstWleaTpwjvbS
J3+tMGTG67f+TdCfDxWYMpQYxLlE8VkbEzKWDwCYvDZRMKCQfv2ErNvb
-----END PRIVATE KEY-----`

var testCert = `-----BEGIN CERTIFICATE-----
MIICGTCCAcCgAwIBAgIRALR/1GXtEud5GQL2CZykkOkwCgYIKoZIzj0EAwIwczEL
MAkGA1UEBhMCVVMxEzARBgNVBAgTCkNhbGlmb3JuaWExFjAUBgNVBAcTDVNhbiBG
cmFuY2lzY28xGTAXBgNVBAoTEG9yZzEuZXhhbXBsZS5jb20xHDAaBgNVBAMTE2Nh
Lm9yZzEuZXhhbXBsZS5jb20wHhcNMTcwNzI4MTQyNzIwWhcNMjcwNzI2MTQyNzIw
WjBbMQswCQYDVQQGEwJVUzETMBEGA1UECBMKQ2FsaWZvcm5pYTEWMBQGA1UEBxMN
U2FuIEZyYW5jaXNjbzEfMB0GA1UEAwwWVXNlcjFAb3JnMS5leGFtcGxlLmNvbTBZ
MBMGByqGSM49AgEGCCqGSM49AwEHA0IABPIVPS+hdftwDg8+02y1aV5pOnCO9tIn
f60wZMbrt/5N0J8PFZgylBjEuUTxWRsTMpYPAJi8NlEwoJB+/YSs29ujTTBLMA4G
A1UdDwEB/wQEAwIHgDAMBgNVHRMBAf8EAjAAMCsGA1UdIwQkMCKAIIeR0TY+iVFf
mvoEKwaToscEu43ZXSj5fTVJornjxDUtMAoGCCqGSM49BAMCA0cAMEQCID+dZ7H5
AiaiI2BjxnL3/TetJ8iFJYZyWvK//an13WV/AiARBJd/pI5A7KZgQxJhXmmR8bie
XdsmTcdRvJ3TS/6HCA==
-----END CERTIFICATE-----`

var msp = "Org1"

func TestCredentialManager(t *testing.T) {

	config, err := config.FromFile("../../../test/fixtures/config/config_test.yaml")()
	if err != nil {
		t.Fatalf(err.Error())
	}
	netConfig, err := config.NetworkConfig()
	if err != nil {
		t.Fatalf("Failed to setup netConfig: %s", err)
	}
	orgConfig, ok := netConfig.Organizations[strings.ToLower(msp)]
	if !ok {
		t.Fatalf("Failed to setup orgConfig: %s", err)
	}
	mspID := orgConfig.MspID

	clientCofig, err := config.Client()
	if err != nil {
		t.Fatalf("Unable to retrieve client config: %v", err)
	}

	// Cleanup key store and user store
	cleanup(t, config.KeyStorePath())
	defer cleanup(t, config.KeyStorePath())
	cleanup(t, clientCofig.CredentialStore.Path)
	defer cleanup(t, clientCofig.CredentialStore.Path)

	cryptoSuite, err := sw.GetSuiteByConfig(config)
	if err != nil {
		t.Fatalf("Failed to setup cryptoSuite: %s", err)
	}

	// the same location used by credential manager.
	// in the future all will use common user store instance from the SDK context
	userStore, err := identity.NewCertFileUserStore(clientCofig.CredentialStore.Path, cryptoSuite)
	if err != nil {
		t.Fatalf("Failed to setup userStore: %s", err)
	}

	credentialMgr, err := NewCredentialManager(msp, config, cryptoSuite)
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

	testUserName := createRandomName()

	// Should not find the user
	if err := checkSigningIdentity(credentialMgr, testUserName); err != api.ErrUserNotFound {
		t.Fatalf("expected ErrUserNotFound, got: %s", err)
	}

	// "Manually" enroll User1
	_, err = fabricCaUtil.ImportBCCSPKeyFromPEMBytes([]byte(testPrivKey), cryptoSuite, false)
	if err != nil {
		t.Fatalf("ImportBCCSPKeyFromPEMBytes failed [%s]", err)
	}
	user1 := identity.NewUser(mspID, testUserName)
	user1.SetEnrollmentCertificate([]byte(testCert))
	err = userStore.Store(user1)
	if err != nil {
		t.Fatalf("userStore.Store: %s", err)
	}

	// Should succeed after enrollment
	if err := checkSigningIdentity(credentialMgr, testUserName); err != nil {
		t.Fatalf("checkSigningIdentity failed: %s", err)
	}
}

func checkSigningIdentity(credentialMgr api.CredentialManager, user string) error {
	id, err := credentialMgr.GetSigningIdentity(user)
	if err == api.ErrUserNotFound {
		return err
	}
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

	credentialMgr, err := NewCredentialManager(msp, config, cryptosuite.GetDefault())
	if err != nil {
		t.Fatalf("Failed to setup credential manager: %s", err)
	}

	_, err = credentialMgr.GetSigningIdentity("")
	if err == nil {
		t.Fatalf("Should get error for empty user name")
	}

	_, err = credentialMgr.GetSigningIdentity("Non-Existent")
	if err != api.ErrUserNotFound {
		t.Fatalf("Should get ErrUserNotFound for non-existent user, got %v", err)
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

func createRandomName() string {
	return "user" + strconv.Itoa(rand.Intn(500000))
}

func cleanup(t *testing.T, storePath string) {
	err := os.RemoveAll(storePath)
	if err != nil {
		t.Fatalf("Cleaning up directory '%s' failed: %v", storePath, err)
	}
}
