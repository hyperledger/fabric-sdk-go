/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package msp

import (
	"strings"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/hyperledger/fabric-sdk-go/internal/github.com/hyperledger/fabric-ca/util"
	"github.com/hyperledger/fabric-sdk-go/pkg/context/api/core"
	"github.com/hyperledger/fabric-sdk-go/pkg/context/api/msp"
	"github.com/hyperledger/fabric-sdk-go/pkg/core/config"
	"github.com/hyperledger/fabric-sdk-go/pkg/core/cryptosuite/bccsp/sw"
	apimocks "github.com/hyperledger/fabric-sdk-go/pkg/msp/api/mocks"
)

var (
	// "Generated" cert, the "result" of a CA CSR processing
	generatedCertBytes = []byte(`-----BEGIN CERTIFICATE-----
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

	// "Generated" private key
	generatedKeyBytes = []byte(`-----BEGIN PRIVATE KEY-----
MIGHAgEAMBMGByqGSM49AgEGCCqGSM49AwEHBG0wawIBAQQg5Ahcehypz6IpAYy6
DtIf5zZsRjP4PtsmDhLbBJsXmD6hRANCAAR+YRAn8dFpDQDyvDA7JKPl5PoZenj3
m1KOnMry/mOZcnXnTIh2ASV4ss8VluzBcyHGAv7BCmxXxDkjcV9eybv8
-----END PRIVATE KEY-----`)

	userToEnroll = "enrollmentID"

	userToEnrollMSPID       string
	enrollmentTestUserStore msp.UserStore
)

func TestGetSigningIdentityWithEnrollment(t *testing.T) {
	config, err := config.FromFile("../../test/fixtures/config/config_test.yaml")()
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
	userToEnrollMSPID = orgConfig.MSPID

	// Delete all private keys from the crypto suite store
	// and users from the user store
	keyStorePath := config.KeyStorePath()
	credentialStorePath := config.CredentialStorePath()
	cleanupTestPath(t, keyStorePath)
	defer cleanupTestPath(t, keyStorePath)
	cleanupTestPath(t, credentialStorePath)
	defer cleanupTestPath(t, credentialStorePath)

	cs, err := sw.GetSuiteByConfig(config)
	userStore := userStoreFromConfig(t, config)

	identityMgr, err := NewIdentityManager(orgName, userStore, cs, config)
	if err != nil {
		t.Fatalf("Failed to setup credential manager: %s", err)
	}

	if err := checkSigningIdentity(identityMgr, "User1"); err != nil {
		t.Fatalf("checkSigningIdentity failed: %s", err)
	}

	// Refers to the same location used by the IdentityManager
	enrollmentTestUserStore, err = NewCertFileUserStore(clientCofig.CredentialStore.Path)
	if err != nil {
		t.Fatalf("Failed to setup userStore: %s", err)
	}

	if err := checkSigningIdentity(identityMgr, userToEnroll); err == nil {
		t.Fatalf("checkSigningIdentity should fail for user who hasn't been enrolled")
	}

	// Enroll the user

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	caClient := apimocks.NewMockCAClient(ctrl)
	prepareForEnroll(t, caClient, cs)

	err = caClient.Enroll(userToEnroll, "enrollmentSecret")
	if err != nil {
		t.Fatalf("fabricCAClient Enroll failed: %v", err)
	}

	if err := checkSigningIdentity(identityMgr, userToEnroll); err != nil {
		t.Fatalf("checkSigningIdentity shouldn't fail for user who hasn been just enrolled: %s", err)
	}
}

// Simulate caClient.Enroll()
func prepareForEnroll(t *testing.T, mc *apimocks.MockCAClient, cs core.CryptoSuite) {
	// A real caClient.Enroll() generates a CSR. In the process, a crypto suite generates
	// a new key pair, and the private key is stored into crypto suite private key storage.

	var err error

	mc.EXPECT().Enroll(gomock.Any(), gomock.Any()).Do(func(enrollmentID string, enrollmentSecret string) {

		// Simulate key and cert management normally done by the SDK

		// Import the key into the crypto suite's private key storage.
		// This is normally done by a crypto suite when a new key is generated
		_, err = util.ImportBCCSPKeyFromPEMBytes(generatedKeyBytes, cs, false)

		// Save the "new" cert to user store
		// This is done by IdentityManagement.Enroll()
		user := &msp.UserData{
			MSPID: userToEnrollMSPID,
			Name:  userToEnroll,
			EnrollmentCertificate: []byte(generatedCertBytes),
		}
		err = enrollmentTestUserStore.Store(user)
		if err != nil {
			t.Fatalf("userStore.Store: %s", err)
		}

	}).Return(err)
}
