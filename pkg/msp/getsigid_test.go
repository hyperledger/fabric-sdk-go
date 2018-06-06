/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package msp

import (
	"math/rand"
	"strconv"
	"strings"
	"testing"

	"fmt"

	fabricCaUtil "github.com/hyperledger/fabric-sdk-go/internal/github.com/hyperledger/fabric-ca/util"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/core"
	providersFab "github.com/hyperledger/fabric-sdk-go/pkg/common/providers/fab"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/msp"

	"github.com/hyperledger/fabric-sdk-go/pkg/core/config"
	"github.com/hyperledger/fabric-sdk-go/pkg/core/cryptosuite"
	"github.com/hyperledger/fabric-sdk-go/pkg/core/cryptosuite/bccsp/sw"
	"github.com/hyperledger/fabric-sdk-go/pkg/fab"
	fcmocks "github.com/hyperledger/fabric-sdk-go/pkg/fab/mocks"
	"github.com/pkg/errors"
)

var (
	testPrivKey = `-----BEGIN PRIVATE KEY-----
MIGHAgEAMBMGByqGSM49AgEGCCqGSM49AwEHBG0wawIBAQQgp4qKKB0WCEfx7XiB
5Ul+GpjM1P5rqc6RhjD5OkTgl5OhRANCAATyFT0voXX7cA4PPtNstWleaTpwjvbS
J3+tMGTG67f+TdCfDxWYMpQYxLlE8VkbEzKWDwCYvDZRMKCQfv2ErNvb
-----END PRIVATE KEY-----`

	testCert = `-----BEGIN CERTIFICATE-----
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

	orgName = "Org1"
)

func TestGetSigningIdentity(t *testing.T) {

	cryptoConfig, endpointConfig, identityConfig, orgConfig := getConfigs(t)
	mspID := orgConfig.MSPID
	clientConfig := identityConfig.Client()

	// Cleanup key store and user store
	cleanupTestPath(t, cryptoConfig.KeyStorePath())
	defer cleanupTestPath(t, cryptoConfig.KeyStorePath())
	cleanupTestPath(t, clientConfig.CredentialStore.Path)
	defer cleanupTestPath(t, clientConfig.CredentialStore.Path)

	cryptoSuite, err := sw.GetSuiteByConfig(cryptoConfig)
	if err != nil {
		t.Fatalf("Failed to setup cryptoSuite: %s", err)
	}

	userStore := userStoreFromConfig(t, identityConfig)
	mgr, err := NewIdentityManager(orgName, userStore, cryptoSuite, endpointConfig)
	if err != nil {
		t.Fatalf("Failed to setup credential manager: %s", err)
	}

	_, err = mgr.GetSigningIdentity("")
	if err == nil {
		t.Fatalf("Should have failed to retrieve signing identity for empty user name")
	}

	_, err = mgr.GetSigningIdentity("Non-Existent")
	if err == nil {
		t.Fatalf("Should have failed to retrieve signing identity for non-existent user")
	}

	testUsername := createRandomName()

	// Should not find the user
	if err1 := checkSigningIdentity(mgr, testUsername); err1 != msp.ErrUserNotFound {
		t.Fatalf("expected ErrUserNotFound, got: %s", err1)
	}

	// "Manually" enroll User1
	enrollUser1(cryptoSuite, t, mspID, testUsername, userStore, mgr)
}

func getConfigs(t *testing.T) (core.CryptoSuiteConfig, providersFab.EndpointConfig, msp.IdentityConfig, providersFab.OrganizationConfig) {
	configBackend, err := config.FromFile("../../pkg/core/config/testdata/config_test.yaml")()
	if err != nil {
		t.Fatal(err)
	}
	cryptoConfig := cryptosuite.ConfigFromBackend(configBackend...)
	endpointConfig, err := fab.ConfigFromBackend(configBackend...)
	if err != nil {
		panic(fmt.Sprintf("Failed to read config: %s", err))
	}
	identityConfig, err := ConfigFromBackend(configBackend...)
	if err != nil {
		panic(fmt.Sprintf("Failed to read config: %s", err))
	}
	netConfig := endpointConfig.NetworkConfig()
	if netConfig == nil {
		t.Fatal("Failed to setup netConfig")
	}
	orgConfig, ok := netConfig.Organizations[strings.ToLower(orgName)]
	if !ok {
		t.Fatalf("Failed to setup orgConfig: %s", err)
	}

	return cryptoConfig, endpointConfig, identityConfig, orgConfig
}

func enrollUser1(cryptoSuite core.CryptoSuite, t *testing.T, mspID string, testUsername string, userStore msp.UserStore, mgr *IdentityManager) {
	_, err := fabricCaUtil.ImportBCCSPKeyFromPEMBytes([]byte(testPrivKey), cryptoSuite, false)
	if err != nil {
		t.Fatalf("ImportBCCSPKeyFromPEMBytes failed [%s]", err)
	}
	user1 := &msp.UserData{
		MSPID: mspID,
		ID:    testUsername,
		EnrollmentCertificate: []byte(testCert),
	}
	err = userStore.Store(user1)
	if err != nil {
		t.Fatalf("userStore.Store: %s", err)
	}
	// Should succeed after enrollment
	if err := checkSigningIdentity(mgr, testUsername); err != nil {
		t.Fatalf("checkSigningIdentity failed: %s", err)
	}
}

func checkSigningIdentity(mgr msp.IdentityManager, user string) error {
	id, err := mgr.GetSigningIdentity(user)
	if err == msp.ErrUserNotFound {
		return err
	}
	if err != nil {
		return errors.Wrapf(err, "Failed to retrieve signing identity: %s", err)
	}

	if id == nil {
		return errors.New("SigningIdentity is nil")
	}
	if id.EnrollmentCertificate() == nil {
		return errors.New("Enrollment cert is missing")
	}
	if id.Identifier().MSPID == "" {
		return errors.New("MSPID is missing")
	}
	if id.PrivateKey() == nil {
		return errors.New("private key is missing")
	}
	return nil
}

func TestGetSigningIdentityInvalidOrg(t *testing.T) {

	configBackend, err := config.FromFile("../../pkg/core/config/testdata/config_test.yaml")()
	if err != nil {
		t.Fatal(err)
	}

	endpointConfig, err := fab.ConfigFromBackend(configBackend...)
	if err != nil {
		panic(fmt.Sprintf("Failed to read config: %s", err))
	}

	identityConfig, err := ConfigFromBackend(configBackend...)
	if err != nil {
		panic(fmt.Sprintf("Failed to read config: %s", err))
	}
	userStore := userStoreFromConfig(t, identityConfig)

	// Invalid Org
	_, err = NewIdentityManager("invalidOrg", userStore, &fcmocks.MockCryptoSuite{}, endpointConfig)
	if err == nil {
		t.Fatalf("Should have failed to setup manager for invalid org")
	}

}

func TestGetSigningIdentityFromEmbeddedCryptoConfig(t *testing.T) {

	configBackend, err := config.FromFile("../../pkg/core/config/testdata/config_test_embedded_pems.yaml")()
	if err != nil {
		t.Fatal(err)
	}
	endpointConfig, err := fab.ConfigFromBackend(configBackend...)
	if err != nil {
		panic(fmt.Sprintf("Failed to read config: %s", err))
	}

	identityConfig, err := ConfigFromBackend(configBackend...)
	if err != nil {
		panic(fmt.Sprintf("Failed to read config: %s", err))
	}
	userStore := userStoreFromConfig(t, identityConfig)

	mgr, err := NewIdentityManager(orgName, userStore, cryptosuite.GetDefault(), endpointConfig)
	if err != nil {
		t.Fatalf("Failed to setup credential manager: %s", err)
	}

	checkSigningIdentityFromEmbeddedCryptoConfig(mgr, t)
}

func checkSigningIdentityFromEmbeddedCryptoConfig(mgr *IdentityManager, t *testing.T) {
	_, err := mgr.GetSigningIdentity("")
	if err == nil {
		t.Fatal("Should get error for empty user name")
	}
	_, err = mgr.GetSigningIdentity("Non-Existent")
	if err != msp.ErrUserNotFound {
		t.Fatalf("Should get ErrUserNotFound for non-existent user, got %s", err)
	}
	if err := checkSigningIdentity(mgr, "EmbeddedUser"); err != nil {
		t.Fatalf("checkSigningIdentity failes: %s", err)
	}
	if err := checkSigningIdentity(mgr, "EmbeddedUserWithPaths"); err != nil {
		t.Fatalf("checkSigningIdentity failes: %s", err)
	}
	if err := checkSigningIdentity(mgr, "EmbeddedUserMixed"); err != nil {
		t.Fatalf("checkSigningIdentity failes: %s", err)
	}
	if err := checkSigningIdentity(mgr, "EmbeddedUserMixed2"); err != nil {
		t.Fatalf("checkSigningIdentity failes: %s", err)
	}
}

func createRandomName() string {
	return "user" + strconv.Itoa(rand.Intn(500000))
}
