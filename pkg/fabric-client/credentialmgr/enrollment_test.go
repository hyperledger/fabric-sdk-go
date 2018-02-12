/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package credentialmgr

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/hyperledger/fabric-sdk-go/api/apiconfig"
	"github.com/hyperledger/fabric-sdk-go/api/apicryptosuite"
	camocks "github.com/hyperledger/fabric-sdk-go/api/apifabca/mocks"
	"github.com/hyperledger/fabric-sdk-go/api/kvstore"
	"github.com/hyperledger/fabric-sdk-go/internal/github.com/hyperledger/fabric-ca/util"
	"github.com/hyperledger/fabric-sdk-go/pkg/fabric-client/credentialmgr/persistence"
	"github.com/pkg/errors"

	"github.com/hyperledger/fabric-sdk-go/pkg/config"
	"github.com/hyperledger/fabric-sdk-go/pkg/cryptosuite/bccsp/sw"
)

func TestCredentialManagerWithEnrollment(t *testing.T) {
	config, err := config.FromFile("../../../test/fixtures/config/config_test.yaml")()
	if err != nil {
		t.Fatalf(err.Error())
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

	cs, err := sw.GetSuiteByConfig(config)

	credentialMgr, err := NewCredentialManager(orgName, config, cs)
	if err != nil {
		t.Fatalf("Failed to setup credential manager: %s", err)
	}

	if err := checkSigningIdentity(credentialMgr, "User1"); err != nil {
		t.Fatalf("checkSigningIdentity failed: %s", err)
	}

	userToEnroll := "enrollmentID"
	certLookupKey := &persistence.CertKey{
		MspID:    orgConfig.MspID,
		UserName: userToEnroll,
	}

	// Refers to the same location used by the CredentialManager for looking up certs
	certStore, err := getCertStore(config, orgName)
	if err != nil {
		t.Fatalf("NewFileCertStore failed: %v", err)
	}

	// // Delete all private keys from the crypto suite store
	keyStorePath := config.KeyStorePath()
	err = cleanup(keyStorePath)
	if err != nil {
		t.Fatalf("cleanup keyStorePath failed: %v", err)
	}

	// Delete userToEnroll from cert store in case it's there
	err = certStore.Delete(certLookupKey)
	if err != nil {
		t.Fatalf("certStore.Delete failed: %v", err)
	}

	if err := checkSigningIdentity(credentialMgr, userToEnroll); err == nil {
		t.Fatalf("checkSigningIdentity should fail for user who hasn't been enrolled")
	}

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	caClient := camocks.NewMockFabricCAClient(ctrl)
	prepareForEnroll(t, caClient, cs)

	_, certBytes, err := caClient.Enroll(userToEnroll, "enrollmentSecret")
	if err != nil {
		t.Fatalf("fabricCAClient Enroll failed: %v", err)
	}
	if certBytes == nil || len(certBytes) == 0 {
		t.Fatalf("Got an empty cert from Enrill()")
	}

	err = certStore.Store(certLookupKey, certBytes)
	if err != nil {
		t.Fatalf("certStore.Store: %v", err)
	}

	if err := checkSigningIdentity(credentialMgr, userToEnroll); err != nil {
		t.Fatalf("checkSigningIdentity shouldn't fail for user who hasn been just enrolled: %s", err)
	}
}

// Simulate caClient.Enroll()
func prepareForEnroll(t *testing.T, mc *camocks.MockFabricCAClient, cs apicryptosuite.CryptoSuite) {
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

	var privateKey apicryptosuite.Key
	var err error

	mc.EXPECT().Enroll(gomock.Any(), gomock.Any()).Do(func(enrollmentID string, enrollmentSecret string) {
		// Import the key into the crypto suite's private key storage.
		// This is normally done by a crypto suite when a new key is generated
		privateKey, err = util.ImportBCCSPKeyFromPEMBytes(keyBytes, cs, false)
	}).Return(privateKey, certBytes, err)
}

func getCertStore(config apiconfig.Config, orgName string) (kvstore.KVStore, error) {
	netConfig, err := config.NetworkConfig()
	if err != nil {
		return nil, err
	}
	orgConfig, ok := netConfig.Organizations[strings.ToLower(orgName)]
	if !ok {
		return nil, errors.New("org config retrieval failed")
	}
	orgCryptoPathTemplate := orgConfig.CryptoPath
	if !filepath.IsAbs(orgCryptoPathTemplate) {
		orgCryptoPathTemplate = filepath.Join(config.CryptoConfigPath(), orgCryptoPathTemplate)
	}
	fmt.Printf("orgCryptoPathTemplate: %s\n", orgCryptoPathTemplate)
	certStore, err := persistence.NewFileCertStore(orgCryptoPathTemplate)
	if err != nil {
		return nil, errors.Wrapf(err, "creating a cert store failed")
	}
	return certStore, nil
}

func cleanup(storePath string) error {
	d, err := os.Open(storePath)
	if err != nil {
		return err
	}
	defer d.Close()
	names, err := d.Readdirnames(-1)
	if err != nil {
		return err
	}
	for _, name := range names {
		err = os.RemoveAll(filepath.Join(storePath, name))
		if err != nil {
			return err
		}
	}
	return nil
}
