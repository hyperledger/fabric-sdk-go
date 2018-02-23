/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package credentialmgr

import (
	"strings"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/hyperledger/fabric-sdk-go/internal/github.com/hyperledger/fabric-ca/util"
	"github.com/hyperledger/fabric-sdk-go/pkg/context/api/core"
	camocks "github.com/hyperledger/fabric-sdk-go/pkg/context/api/fab/mocks"
	"github.com/hyperledger/fabric-sdk-go/pkg/fabric-client/identity"

	"github.com/hyperledger/fabric-sdk-go/pkg/config"
	"github.com/hyperledger/fabric-sdk-go/pkg/cryptosuite/bccsp/sw"
)

func TestCredentialManagerWithEnrollment(t *testing.T) {
	config, err := config.FromFile("../../../test/fixtures/config/config_test.yaml")()
	if err != nil {
		t.Fatalf(err.Error())
	}
	clientCofig, err := config.Client()
	if err != nil {
		t.Fatalf("Unable to retrieve client config: %v", err)
	}
	netConfig, err := config.NetworkConfig()
	if err != nil {
		t.Fatalf("NetworkConfig failed: %s", err)
	}
	orgName := "Org1"
	orgConfig, ok := netConfig.Organizations[strings.ToLower(orgName)]
	if !ok {
		t.Fatalf("org config not found: %s", orgName)
	}

	// Delete all private keys from the crypto suite store
	// and users from the user store
	keyStorePath := config.KeyStorePath()
	cleanup(t, keyStorePath)
	defer cleanup(t, keyStorePath)
	cleanup(t, clientCofig.CredentialStore.Path)
	defer cleanup(t, clientCofig.CredentialStore.Path)

	cs, err := sw.GetSuiteByConfig(config)

	credentialMgr, err := NewCredentialManager(orgName, config, cs)
	if err != nil {
		t.Fatalf("Failed to setup credential manager: %s", err)
	}

	if err := checkSigningIdentity(credentialMgr, "User1"); err != nil {
		t.Fatalf("checkSigningIdentity failed: %s", err)
	}

	// Refers to the same location used by the CredentialManager
	userStore, err := identity.NewCertFileUserStore(clientCofig.CredentialStore.Path, cs)
	if err != nil {
		t.Fatalf("Failed to setup userStore: %s", err)
	}

	userToEnroll := "enrollmentID"

	if err := checkSigningIdentity(credentialMgr, userToEnroll); err == nil {
		t.Fatalf("checkSigningIdentity should fail for user who hasn't been enrolled")
	}

	// Enroll the user

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	caClient := camocks.NewMockFabricCAClient(ctrl)
	prepareForEnroll(t, caClient, cs)

	_, certBytes, err := caClient.Enroll(userToEnroll, "enrollmentSecret")
	if err != nil {
		t.Fatalf("fabricCAClient Enroll failed: %v", err)
	}

	// Private key is saved to key store by Enroll()
	// For now, the app has to save the cert (user)
	user := identity.NewUser(orgConfig.MspID, userToEnroll)
	user.SetEnrollmentCertificate([]byte(certBytes))
	err = userStore.Store(user)
	if err != nil {
		t.Fatalf("userStore.Store: %s", err)
	}

	if err := checkSigningIdentity(credentialMgr, userToEnroll); err != nil {
		t.Fatalf("checkSigningIdentity shouldn't fail for user who hasn been just enrolled: %s", err)
	}
}

// Simulate caClient.Enroll()
func prepareForEnroll(t *testing.T, mc *camocks.MockFabricCAClient, cs core.CryptoSuite) {
	// A real caClient.Enroll() generates a CSR. In the process, a crypto suite generates
	// a new key pair, and the private key is stored into crypto suite private key storage.

	// "Generated" private key
	keyBytes := []byte(`-----BEGIN PRIVATE KEY-----
MIGHAgEAMBMGByqGSM49AgEGCCqGSM49AwEHBG0wawIBAQQg5Ahcehypz6IpAYy6
DtIf5zZsRjP4PtsmDhLbBJsXmD6hRANCAAR+YRAn8dFpDQDyvDA7JKPl5PoZenj3
m1KOnMry/mOZcnXnTIh2ASV4ss8VluzBcyHGAv7BCmxXxDkjcV9eybv8
-----END PRIVATE KEY-----`)

	// "Generated" cert, the "result" of a CA CSR processing
	certBytes := []byte(`-----BEGIN CERTIFICATE-----
MIICGjCCAcCgAwIBAgIRAIQkbh9nsGnLmDalAVlj8sUwCgYIKoZIzj0EAwIwczEL
MAkGA1UEBhMCVVMxEzARBgNVBAgTCkNhbGlmb3JuaWExFjAUBgNVBAcTDVNhbiBG
cmFuY2lzY28xGTAXBgNVBAoTEG9yZzEuZXhhbXBsZS5jb20xHDAaBgNVBAMTE2Nh
Lm9yZzEuZXhhbXBsZS5jb20wHhcNMTcwNzI4MTQyNzIwWhcNMjcwNzI2MTQyNzIw
WjBbMQswCQYDVQQGEwJVUzETMBEGA1UECBMKQ2FsaWZvcm5pYTEWMBQGA1UEBxMN
U2FuIEZyYW5jaXNjbzEfMB0GA1UEAwwWQWRtaW5Ab3JnMS5leGFtcGxlLmNvbTBZ
MBMGByqGSM49AgEGCCqGSM49AwEHA0IABH5hECfx0WkNAPK8MDsko+Xk+hl6ePeb
Uo6cyvL+Y5lydedMiHYBJXiyzxWW7MFzIcYC/sEKbFfEOSNxX17Ju/yjTTBLMA4G
A1UdDwEB/wQEAwIHgDAMBgNVHRMBAf8EAjAAMCsGA1UdIwQkMCKAIIeR0TY+iVFf
mvoEKwaToscEu43ZXSj5fTVJornjxDUtMAoGCCqGSM49BAMCA0gAMEUCIQDVf8cL
NrfToiPzJpEFPGF+/8CpzOkl91oz+XJsvdgf5wIgI/e8mpvpplUQbU52+LejA36D
CsbWERvZPjR/GFEDEvc=
-----END CERTIFICATE-----`)

	var privateKey core.Key
	var err error

	mc.EXPECT().Enroll(gomock.Any(), gomock.Any()).Do(func(enrollmentID string, enrollmentSecret string) {
		// Import the key into the crypto suite's private key storage.
		// This is normally done by a crypto suite when a new key is generated
		privateKey, err = util.ImportBCCSPKeyFromPEMBytes(keyBytes, cs, false)
	}).Return(privateKey, certBytes, err)
}
