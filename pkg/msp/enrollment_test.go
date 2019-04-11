/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package msp

import (
	"fmt"
	"path/filepath"
	"strings"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/hyperledger/fabric-sdk-go/internal/github.com/hyperledger/fabric-ca/util"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/core"
	providersFab "github.com/hyperledger/fabric-sdk-go/pkg/common/providers/fab"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/msp"
	"github.com/hyperledger/fabric-sdk-go/pkg/core/config"
	"github.com/hyperledger/fabric-sdk-go/pkg/core/cryptosuite"
	"github.com/hyperledger/fabric-sdk-go/pkg/core/cryptosuite/bccsp/sw"
	"github.com/hyperledger/fabric-sdk-go/pkg/fab"
	"github.com/hyperledger/fabric-sdk-go/pkg/msp/api"
	apimocks "github.com/hyperledger/fabric-sdk-go/pkg/msp/test/mockmspapi"
	"github.com/hyperledger/fabric-sdk-go/test/metadata"
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
	configPath := filepath.Join(metadata.GetProjectPath(), metadata.SDKConfigPath, configTestFile)
	configBackend, err := config.FromFile(configPath)()
	if err != nil {
		t.Fatalf(err.Error())
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

	clientConfig := identityConfig.Client()
	netConfig := endpointConfig.NetworkConfig()
	if netConfig == nil {
		t.Fatal("Failed to get network config")
	}
	orgName := "Org1"
	orgConfig, ok := netConfig.Organizations[strings.ToLower(orgName)]
	if !ok {
		t.Fatalf("org config not found: %s", orgName)
	}
	userToEnrollMSPID = orgConfig.MSPID

	// Delete all private keys from the crypto suite store
	// and users from the user store
	keyStorePath := cryptoConfig.KeyStorePath()
	credentialStorePath := identityConfig.CredentialStorePath()
	cleanupTestPath(t, keyStorePath)
	defer cleanupTestPath(t, keyStorePath)
	cleanupTestPath(t, credentialStorePath)
	defer cleanupTestPath(t, credentialStorePath)

	checkSigningIdentityWithEnrollment(cryptoConfig, t, identityConfig, orgName, endpointConfig, clientConfig)
}

func checkSigningIdentityWithEnrollment(cryptoConfig core.CryptoSuiteConfig, t *testing.T, identityConfig msp.IdentityConfig, orgName string, endpointConfig providersFab.EndpointConfig, clientConfig *msp.ClientConfig) {
	cs, err := sw.GetSuiteByConfig(cryptoConfig)
	if err != nil {
		t.Fatalf("Failed to get suite by config: %s", err)
	}
	userStore := userStoreFromConfig(t, identityConfig)
	identityMgr, err := NewIdentityManager(orgName, userStore, cs, endpointConfig)
	if err != nil {
		t.Fatalf("Failed to setup credential manager: %s", err)
	}
	if err = checkSigningIdentity(identityMgr, "User1"); err != nil {
		t.Fatalf("checkSigningIdentity failed: %s", err)
	}
	// Refers to the same location used by the IdentityManager
	enrollmentTestUserStore, err = NewCertFileUserStore(clientConfig.CredentialStore.Path)
	if err != nil {
		t.Fatalf("Failed to setup userStore: %s", err)
	}
	if err = checkSigningIdentity(identityMgr, userToEnroll); err == nil {
		t.Fatalf("checkSigningIdentity should fail for user who hasn't been enrolled")
	}
	// Enroll the user
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	caClient := apimocks.NewMockCAClient(ctrl)
	prepareForEnroll(t, caClient, cs)
	err = caClient.Enroll(&api.EnrollmentRequest{Name: userToEnroll, Secret: "enrollmentSecret"})
	if err != nil {
		t.Fatalf("fabricCAClient Enroll failed: %s", err)
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

	mc.EXPECT().Enroll(gomock.Any()).Do(func(enrollmentRequest *api.EnrollmentRequest) {

		// Simulate key and cert management normally done by the SDK

		// Import the key into the crypto suite's private key storage.
		// This is normally done by a crypto suite when a new key is generated
		_, err = util.ImportBCCSPKeyFromPEMBytes(generatedKeyBytes, cs, false)

		// Save the "new" cert to user store
		// This is done by IdentityManagement.Enroll()
		user := &msp.UserData{
			MSPID:                 userToEnrollMSPID,
			ID:                    userToEnroll,
			EnrollmentCertificate: []byte(generatedCertBytes),
		}
		err = enrollmentTestUserStore.Store(user)
		if err != nil {
			t.Fatalf("userStore.Store: %s", err)
		}

	}).Return(err)
}
