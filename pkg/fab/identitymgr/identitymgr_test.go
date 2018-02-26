/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package identitymgr

import (
	"fmt"
	"io/ioutil"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/pkg/errors"

	contextApi "github.com/hyperledger/fabric-sdk-go/pkg/context/api"
	"github.com/hyperledger/fabric-sdk-go/pkg/context/api/core"
	"github.com/hyperledger/fabric-sdk-go/pkg/context/api/core/mocks"
	"github.com/hyperledger/fabric-sdk-go/pkg/context/api/fab"
	"github.com/hyperledger/fabric-sdk-go/pkg/core/config"
	"github.com/hyperledger/fabric-sdk-go/pkg/core/cryptosuite"
	cryptosuiteimpl "github.com/hyperledger/fabric-sdk-go/pkg/core/cryptosuite/bccsp/sw"
	bccspwrapper "github.com/hyperledger/fabric-sdk-go/pkg/core/cryptosuite/bccsp/wrapper"
	"github.com/hyperledger/fabric-sdk-go/pkg/fab/identity"
	"github.com/hyperledger/fabric-sdk-go/pkg/fab/identitymgr/mocks"
)

const (
	org1               = "Org1"
	caServerURL        = "http://localhost:8090"
	wrongCAServerURL   = "http://localhost:8091"
	dummyUserStorePath = "/tmp/userstore"
)

var (
	fullConfig     core.Config
	cryptoSuite    core.CryptoSuite
	wrongURLConfig core.Config
	userStore      contextApi.UserStore
)

// TestMain Load testing config
func TestMain(m *testing.M) {

	var err error
	fullConfig, err = config.FromFile("testdata/config_test.yaml")()
	if err != nil {
		panic(fmt.Sprintf("Failed to read full config: %v", err))
	}

	// Delete all private keys from the crypto suite store
	// and users from the user store
	cleanup(fullConfig.KeyStorePath())
	defer cleanup(fullConfig.KeyStorePath())
	cleanup(fullConfig.CredentialStorePath())
	defer cleanup(fullConfig.CredentialStorePath())

	cryptoSuite, err = cryptosuiteimpl.GetSuiteByConfig(fullConfig)
	if cryptoSuite == nil {
		panic(fmt.Sprintf("Failed initialize cryptoSuite: %v", err))
	}
	if fullConfig.CredentialStorePath() != "" {
		userStore, err = identity.NewCertFileUserStore(fullConfig.CredentialStorePath(), cryptoSuite)
		if err != nil {
			panic(fmt.Sprintf("creating a user store failed: %v", err))
		}
	}

	wrongURLConfig, err = config.FromFile("testdata/config_test_wrong_url.yaml")()
	if err != nil {
		panic(fmt.Sprintf("Failed to read full config: %v", err))
	}

	// Start Http Server
	go mocks.StartFabricCAMockServer(strings.TrimPrefix(caServerURL, "http://"), cryptoSuite)
	// Allow HTTP server to start
	time.Sleep(1 * time.Second)

	os.Exit(m.Run())
}

// TestEnrollAndReenroll tests enrol/reenroll scenarios
func TestEnrollAndReenroll(t *testing.T) {

	identityManager, err := New(org1, fullConfig, cryptoSuite)
	if err != nil {
		t.Fatalf("NewidentityManagerClient return error: %v", err)
	}
	_, _, err = identityManager.Enroll("", "user1")
	if err == nil {
		t.Fatalf("Enroll didn't return error")
	}
	_, _, err = identityManager.Enroll("test", "")
	if err == nil {
		t.Fatalf("Enroll didn't return error")
	}
	_, _, err = identityManager.Enroll("enrollmentID", "enrollmentSecret")
	if err != nil {
		t.Fatalf("identityManager Enroll return error %v", err)
	}

	user := mocks.NewMockUser("")
	// Reenroll with nil user
	_, _, err = identityManager.Reenroll(nil)
	if err == nil {
		t.Fatalf("Expected error with nil user")
	}
	if err.Error() != "user required" {
		t.Fatalf("Expected error user required. Got: %s", err.Error())
	}

	// Reenroll with user.Name is empty
	_, _, err = identityManager.Reenroll(user)
	if err == nil {
		t.Fatalf("Expected error with user.Name is empty")
	}
	if err.Error() != "user name missing" {
		t.Fatalf("Expected error user name missing. Got: %s", err.Error())
	}

	// Reenroll with user.EnrollmentCertificate is empty
	user = mocks.NewMockUser("testUser")
	_, _, err = identityManager.Reenroll(user)
	if err == nil {
		t.Fatalf("Expected error with user.EnrollmentCertificate is empty")
	}
	if !strings.Contains(err.Error(), "createSigningIdentity failed") {
		t.Fatalf("Expected error createSigningIdentity failed. Got: %s", err.Error())
	}

	// Reenroll with appropriate user
	user.SetEnrollmentCertificate(readCert(t))
	key, err := cryptosuite.GetDefault().KeyGen(cryptosuite.GetECDSAP256KeyGenOpts(true))
	if err != nil {
		t.Fatalf("KeyGen return error %v", err)
	}
	user.SetPrivateKey(key)
	_, _, err = identityManager.Reenroll(user)
	if err != nil {
		t.Fatalf("Reenroll return error %v", err)
	}

	// Try going against wrong CA URL
	identityManager, err = New(org1, wrongURLConfig, cryptoSuite)
	if err != nil {
		t.Fatalf("NewidentityManagerClient return error: %v", err)
	}
	_, _, err = identityManager.Enroll("enrollmentID", "enrollmentSecret")
	if err == nil {
		t.Fatalf("Enroll didn't return error")
	}

}

// TestRegister tests multiple scenarios of registering a test (mocked or nil user) and their certs
func TestRegister(t *testing.T) {

	identityManager, err := New(org1, fullConfig, cryptoSuite)
	if err != nil {
		t.Fatalf("NewidentityManagerClient returned error: %v", err)
	}

	// Register with nil request
	_, err = identityManager.Register(nil)
	if err == nil {
		t.Fatalf("Expected error with nil request")
	}

	// Register without registration name parameter
	_, err = identityManager.Register(&fab.RegistrationRequest{})
	if err == nil {
		t.Fatalf("Expected error without registration name parameter")
	}

	// Register with valid request
	var attributes []fab.Attribute
	attributes = append(attributes, fab.Attribute{Key: "test1", Value: "test2"})
	attributes = append(attributes, fab.Attribute{Key: "test2", Value: "test3"})
	secret, err := identityManager.Register(&fab.RegistrationRequest{Name: "test", Affiliation: "test", Attributes: attributes})
	if err != nil {
		t.Fatalf("identityManager Register return error %v", err)
	}
	if secret != "mockSecretValue" {
		t.Fatalf("identityManager Register return wrong value %s", secret)
	}
}

// TestRevoke will test multiple revoking a user with a nil request or a nil user
// TODO - improve Revoke test coverage
func TestRevoke(t *testing.T) {

	cryptoSuite, err := cryptosuiteimpl.GetSuiteByConfig(fullConfig)
	if err != nil {
		t.Fatalf("cryptosuite.GetSuiteByConfig returned error: %v", err)
	}

	identityManager, err := New(org1, fullConfig, cryptoSuite)
	if err != nil {
		t.Fatalf("NewidentityManagerClient returned error: %v", err)
	}
	mockKey := bccspwrapper.GetKey(&mocks.MockKey{})

	// Revoke with nil request
	_, err = identityManager.Revoke(nil)
	if err == nil {
		t.Fatalf("Expected error with nil request")
	}

	user := mocks.NewMockUser("test")
	user.SetEnrollmentCertificate(readCert(t))
	user.SetPrivateKey(mockKey)

	_, err = identityManager.Revoke(&fab.RevocationRequest{})
	if err == nil {
		t.Fatalf("Expected decoding error with test cert")
	}
}

// TestGetCAName will test the CAName is properly created once a new identityManagerClient is created
func TestGetCAName(t *testing.T) {

	identityManager, err := New(org1, fullConfig, cryptoSuite)
	if err != nil {
		t.Fatalf("NewidentityManagerClient returned error: %v", err)
	}
	netConfig, err := fullConfig.NetworkConfig()
	if err != nil {
		t.Fatalf("network config retrieval failed: %v", err)
	}
	orgConfig, ok := netConfig.Organizations[strings.ToLower(org1)]
	if !ok {
		t.Fatalf("org config retrieval failed: %v", err)
	}

	if identityManager.CAName() != orgConfig.CertificateAuthorities[0] {
		t.Fatalf("CAName returned wrong value: %s", identityManager.CAName())
	}
}

// TestCreateNewidentityManagerClientCAConfigMissingFailure will test newidentityManager Client creation with with CAConfig
func TestCreateNewidentityManagerClientCAConfigMissingFailure(t *testing.T) {

	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()
	mockConfig := mock_core.NewMockConfig(mockCtrl)

	mockConfig.EXPECT().NetworkConfig().Return(fullConfig.NetworkConfig()).AnyTimes()
	mockConfig.EXPECT().CryptoConfigPath().Return(fullConfig.CryptoConfigPath()).AnyTimes()
	mockConfig.EXPECT().CAConfig(org1).Return(nil, errors.New("CAConfig error"))
	mockConfig.EXPECT().CredentialStorePath().Return(dummyUserStorePath).AnyTimes()
	mgr, err := New(org1, mockConfig, cryptoSuite)
	if err != nil {
		t.Fatalf("failed to create IdentityManager: %v", err)
	}
	_, _, err = mgr.Enroll("a", "b")
	if err == nil || !strings.Contains(err.Error(), "CAConfig error") {
		t.Fatalf("Expected error from CAConfig. Got: %v", err)
	}

}

// TestCreateNewidentityManagerClientCertFilesMissingFailure will test newidentityManager Client creation with missing CA Cert files
func TestCreateNewidentityManagerClientCertFilesMissingFailure(t *testing.T) {

	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()
	mockConfig := mock_core.NewMockConfig(mockCtrl)
	mockConfig.EXPECT().NetworkConfig().Return(fullConfig.NetworkConfig()).AnyTimes()
	mockConfig.EXPECT().CryptoConfigPath().Return(fullConfig.CryptoConfigPath()).AnyTimes()
	mockConfig.EXPECT().CAConfig(org1).Return(&core.CAConfig{}, nil).AnyTimes()
	mockConfig.EXPECT().CredentialStorePath().Return(dummyUserStorePath).AnyTimes()
	mockConfig.EXPECT().CAServerCertPaths(org1).Return(nil, errors.New("CAServerCertPaths error"))
	mgr, err := New(org1, mockConfig, cryptoSuite)
	if err != nil {
		t.Fatalf("failed to create IdentityManager: %v", err)
	}
	_, _, err = mgr.Enroll("a", "b")
	if err == nil || !strings.Contains(err.Error(), "CAServerCertPaths error") {
		t.Fatalf("Expected error from CAServerCertPaths. Got: %v", err)
	}
}

// TestCreateNewidentityManagerClientCertFileErrorFailure will test newidentityManager Client creation with missing CA Cert files, additional scenario
func TestCreateNewidentityManagerClientCertFileErrorFailure(t *testing.T) {

	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()
	mockConfig := mock_core.NewMockConfig(mockCtrl)
	mockConfig.EXPECT().NetworkConfig().Return(fullConfig.NetworkConfig()).AnyTimes()
	mockConfig.EXPECT().CryptoConfigPath().Return(fullConfig.CryptoConfigPath()).AnyTimes()
	mockConfig.EXPECT().CAConfig(org1).Return(&core.CAConfig{}, nil).AnyTimes()
	mockConfig.EXPECT().CredentialStorePath().Return(dummyUserStorePath).AnyTimes()
	mockConfig.EXPECT().CAServerCertPaths(org1).Return([]string{"test"}, nil)
	mockConfig.EXPECT().CAClientCertPath(org1).Return("", errors.New("CAClientCertPath error"))
	mgr, err := New(org1, mockConfig, cryptoSuite)
	if err != nil {
		t.Fatalf("failed to create IdentityManager: %v", err)
	}
	_, _, err = mgr.Enroll("a", "b")
	if err == nil || !strings.Contains(err.Error(), "CAClientCertPath error") {
		t.Fatalf("Expected error from CAClientCertPath. Got: %v", err)
	}
}

// TestCreateNewidentityManagerClientKeyFileErrorFailure will test newidentityManager Client creation with missing CA Cert files and missing key
func TestCreateNewidentityManagerClientKeyFileErrorFailure(t *testing.T) {

	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()
	mockConfig := mock_core.NewMockConfig(mockCtrl)
	mockConfig.EXPECT().NetworkConfig().Return(fullConfig.NetworkConfig()).AnyTimes()
	mockConfig.EXPECT().CryptoConfigPath().Return(fullConfig.CryptoConfigPath()).AnyTimes()
	mockConfig.EXPECT().CAConfig(org1).Return(&core.CAConfig{}, nil).AnyTimes()
	mockConfig.EXPECT().CredentialStorePath().Return(dummyUserStorePath).AnyTimes()
	mockConfig.EXPECT().CAServerCertPaths(org1).Return([]string{"test"}, nil)
	mockConfig.EXPECT().CAClientCertPath(org1).Return("", nil)
	mockConfig.EXPECT().CAClientKeyPath(org1).Return("", errors.New("CAClientKeyPath error"))
	mgr, err := New(org1, mockConfig, cryptoSuite)
	if err != nil {
		t.Fatalf("failed to create IdentityManager: %v", err)
	}
	_, _, err = mgr.Enroll("a", "b")
	if err == nil || !strings.Contains(err.Error(), "CAClientKeyPath error") {
		t.Fatalf("Expected error from CAClientKeyPath. Got: %v", err)
	}
}

// TestCreateValidBCCSPOptsForNewFabricClient test newidentityManager Client creation with valid inputs, successful scenario
func TestCreateValidBCCSPOptsForNewFabricClient(t *testing.T) {

	newCryptosuiteProvider, err := cryptosuiteimpl.GetSuiteByConfig(fullConfig)

	if err != nil {
		t.Fatalf("Expected fabric client ryptosuite to be created with SW BCCS provider, but got %v", err.Error())
	}
	_, err = New(org1, fullConfig, newCryptosuiteProvider)
	if err != nil {
		t.Fatalf("Expected fabric client to be created with SW BCCS provider, but got %v", err.Error())
	}
}

// readCert Reads a random cert for testing
func readCert(t *testing.T) []byte {
	cert, err := ioutil.ReadFile("testdata/root.pem")
	if err != nil {
		t.Fatalf("Error reading cert: %s", err.Error())
	}
	return cert
}

// TestInterfaces will test if the interface instantiation happens properly, ie no nil returned
func TestInterfaces(t *testing.T) {
	var apiCA fab.IdentityManager
	var ca IdentityManager

	apiCA = &ca
	if apiCA == nil {
		t.Fatalf("this shouldn't happen.")
	}
}

func cleanup(storePath string) {
	err := os.RemoveAll(storePath)
	if err != nil {
		panic(fmt.Sprintf("Failed to remove dir %s: %v\n", storePath, err))
	}
}

func cleanupTestPath(t *testing.T, storePath string) {
	err := os.RemoveAll(storePath)
	if err != nil {
		t.Fatalf("Cleaning up directory '%s' failed: %v", storePath, err)
	}
}
