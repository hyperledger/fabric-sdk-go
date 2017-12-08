/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package fabricca

import (
	"io/ioutil"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/hyperledger/fabric-sdk-go/api/apiconfig/mocks"
	"github.com/hyperledger/fabric-sdk-go/pkg/errors"

	config "github.com/hyperledger/fabric-sdk-go/api/apiconfig"
	ca "github.com/hyperledger/fabric-sdk-go/api/apifabca"

	"github.com/hyperledger/fabric-sdk-go/api/apicryptosuite"
	"github.com/hyperledger/fabric-sdk-go/pkg/cryptosuite"
	cryptosuiteimpl "github.com/hyperledger/fabric-sdk-go/pkg/cryptosuite/bccsp"
	"github.com/hyperledger/fabric-sdk-go/pkg/fabric-ca-client/mocks"
)

var configImp config.Config
var cryptoSuiteProvider apicryptosuite.CryptoSuite
var org1 = "peerorg1"
var caServerURL = "http://localhost:8090"
var wrongCAServerURL = "http://localhost:8091"

// TestMain Load testing config
func TestMain(m *testing.M) {
	configImp = mocks.NewMockConfig(caServerURL)
	cryptoSuiteProvider, _ = cryptosuiteimpl.GetSuiteByConfig(configImp)
	if cryptoSuiteProvider == nil {
		panic("Failed initialize cryptoSuiteProvider")
	}
	// Start Http Server
	go mocks.StartFabricCAMockServer(strings.TrimPrefix(caServerURL, "http://"))
	// Allow HTTP server to start
	time.Sleep(1 * time.Second)
	os.Exit(m.Run())
}

// TestEnroll will test multiple enrol scenarios
func TestEnroll(t *testing.T) {

	fabricCAClient, err := NewFabricCAClient(org1, configImp, cryptoSuiteProvider)
	if err != nil {
		t.Fatalf("NewFabricCAClient return error: %v", err)
	}
	_, _, err = fabricCAClient.Enroll("", "user1")
	if err == nil {
		t.Fatalf("Enroll didn't return error")
	}
	if err.Error() != "enrollmentID required" {
		t.Fatalf("Enroll didn't return right error")
	}
	_, _, err = fabricCAClient.Enroll("test", "")
	if err == nil {
		t.Fatalf("Enroll didn't return error")
	}
	if err.Error() != "enrollmentSecret required" {
		t.Fatalf("Enroll didn't return right error")
	}
	_, _, err = fabricCAClient.Enroll("enrollmentID", "enrollmentSecret")
	if err != nil {
		t.Fatalf("fabricCAClient Enroll return error %v", err)
	}

	wrongConfigImp := mocks.NewMockConfig(wrongCAServerURL)
	fabricCAClient, err = NewFabricCAClient(org1, wrongConfigImp, cryptoSuiteProvider)
	if err != nil {
		t.Fatalf("NewFabricCAClient return error: %v", err)
	}
	_, _, err = fabricCAClient.Enroll("enrollmentID", "enrollmentSecret")
	if err == nil {
		t.Fatalf("Enroll didn't return error")
	}
	if !strings.Contains(err.Error(), "enroll failed") {
		t.Fatalf("Expected error enroll failed. Got: %s", err)
	}

}

// TestRegister tests multiple scenarios of registering a test (mocked or nil user) and their certs
func TestRegister(t *testing.T) {

	fabricCAClient, err := NewFabricCAClient(org1, configImp, cryptoSuiteProvider)
	if err != nil {
		t.Fatalf("NewFabricCAClient returned error: %v", err)
	}
	user := mocks.NewMockUser("test")
	// Register with nil request
	_, err = fabricCAClient.Register(user, nil)
	if err == nil {
		t.Fatalf("Expected error with nil request")
	}
	if err.Error() != "registration request required" {
		t.Fatalf("Expected error registration request required. Got: %s", err.Error())
	}

	//Register with nil user
	_, err = fabricCAClient.Register(nil, &ca.RegistrationRequest{})
	if err == nil {
		t.Fatalf("Expected error with nil user")
	}
	if !strings.Contains(err.Error(), "failed to create request for signing identity") {
		t.Fatalf("Expected error failed to create request for signing identity. Got: %s", err.Error())
	}
	// Register with nil user cert and key
	_, err = fabricCAClient.Register(user, &ca.RegistrationRequest{})
	if err == nil {
		t.Fatalf("Expected error without user enrolment information")
	}
	if !strings.Contains(err.Error(), "failed to create request for signing identity") {
		t.Fatalf("Expected error failed to create request for signing identity. Got: %s", err.Error())
	}

	user.SetEnrollmentCertificate(readCert(t))
	key, err := cryptosuite.GetDefault().KeyGen(cryptosuite.GetECDSAP256KeyGenOpts(true))
	if err != nil {
		t.Fatalf("KeyGen return error %v", err)
	}
	user.SetPrivateKey(key)
	// Register without registration name parameter
	_, err = fabricCAClient.Register(user, &ca.RegistrationRequest{})
	if !strings.Contains(err.Error(), "failed to register user") {
		t.Fatalf("Expected error failed to register user. Got: %s", err.Error())
	}

	// Register with valid request
	var attributes []ca.Attribute
	attributes = append(attributes, ca.Attribute{Key: "test1", Value: "test2"})
	attributes = append(attributes, ca.Attribute{Key: "test2", Value: "test3"})
	secret, err := fabricCAClient.Register(user, &ca.RegistrationRequest{Name: "test",
		Affiliation: "test", Attributes: attributes})
	if err != nil {
		t.Fatalf("fabricCAClient Register return error %v", err)
	}
	if secret != "mockSecretValue" {
		t.Fatalf("fabricCAClient Register return wrong value %s", secret)
	}
}

// TestRevoke will test multiple revoking a user with a nil request or a nil user
func TestRevoke(t *testing.T) {

	cryptoSuiteProvider, err := cryptosuiteimpl.GetSuiteByConfig(configImp)
	if err != nil {
		t.Fatalf("cryptosuite.GetSuiteByConfig returned error: %v", err)
	}

	fabricCAClient, err := NewFabricCAClient(org1, configImp, cryptoSuiteProvider)
	if err != nil {
		t.Fatalf("NewFabricCAClient returned error: %v", err)
	}
	mockKey := cryptosuiteimpl.GetKey(&mocks.MockKey{})
	user := mocks.NewMockUser("test")
	// Revoke with nil request
	err = fabricCAClient.Revoke(user, nil)
	if err == nil {
		t.Fatalf("Expected error with nil request")
	}
	if err.Error() != "revocation request required" {
		t.Fatalf("Expected error revocation request required. Got: %s", err.Error())
	}
	//Revoke with nil user
	err = fabricCAClient.Revoke(nil, &ca.RevocationRequest{})
	if err == nil {
		t.Fatalf("Expected error with nil user")
	}
	if !strings.Contains(err.Error(), "failed to create request for signing identity") {
		t.Fatalf("Expected error failed to create request for signing identity. Got: %s", err.Error())
	}
	user.SetEnrollmentCertificate(readCert(t))
	user.SetPrivateKey(mockKey)
	err = fabricCAClient.Revoke(user, &ca.RevocationRequest{})
	if err == nil {
		t.Fatalf("Expected decoding error with test cert")
	}
}

// TestReenroll will test multiple scenarios of re enrolling a user
func TestReenroll(t *testing.T) {

	fabricCAClient, err := NewFabricCAClient(org1, configImp, cryptoSuiteProvider)
	if err != nil {
		t.Fatalf("NewFabricCAClient returned error: %v", err)
	}
	user := mocks.NewMockUser("")
	// Reenroll with nil user
	_, _, err = fabricCAClient.Reenroll(nil)
	if err == nil {
		t.Fatalf("Expected error with nil user")
	}
	if err.Error() != "user required" {
		t.Fatalf("Expected error user required. Got: %s", err.Error())
	}
	// Reenroll with user.Name is empty
	_, _, err = fabricCAClient.Reenroll(user)
	if err == nil {
		t.Fatalf("Expected error with user.Name is empty")
	}
	if err.Error() != "user name missing" {
		t.Fatalf("Expected error user name missing. Got: %s", err.Error())
	}
	// Reenroll with user.EnrollmentCertificate is empty
	user = mocks.NewMockUser("testUser")
	_, _, err = fabricCAClient.Reenroll(user)
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
	_, _, err = fabricCAClient.Reenroll(user)
	if err != nil {
		t.Fatalf("Reenroll return error %v", err)
	}

	// Reenroll with wrong fabric-ca server url
	wrongConfigImp := mocks.NewMockConfig(wrongCAServerURL)
	fabricCAClient, err = NewFabricCAClient(org1, wrongConfigImp, cryptoSuiteProvider)
	if err != nil {
		t.Fatalf("NewFabricCAClient return error: %v", err)
	}
	_, _, err = fabricCAClient.Reenroll(user)
	if err == nil {
		t.Fatalf("Expected error with wrong fabric-ca server url")
	}
	if !strings.Contains(err.Error(), "reenroll failed") {
		t.Fatalf("Expected error with wrong fabric-ca server url. Got: %s", err.Error())
	}
}

// TestGetCAName will test the CAName is properly created once a new FabricCAClient is created
func TestGetCAName(t *testing.T) {
	fabricCAClient, err := NewFabricCAClient(org1, configImp, cryptoSuiteProvider)
	if err != nil {
		t.Fatalf("NewFabricCAClient returned error: %v", err)
	}
	if fabricCAClient.CAName() != "test" {
		t.Fatalf("CAName returned wrong value: %s", fabricCAClient.CAName())
	}
}

// TestCreateNewFabricCAClientOrgAndConfigMissingFailure tests for newFabricCA Client creation with a missing Config and Org
func TestCreateNewFabricCAClientOrgAndConfigMissingFailure(t *testing.T) {
	_, err := NewFabricCAClient("", configImp, cryptoSuiteProvider)
	if err.Error() != "organization, config and cryptoSuite are required to load CA config" {
		t.Fatalf("Expected error without oganization information. Got: %s", err.Error())
	}
	_, err = NewFabricCAClient(org1, nil, cryptoSuiteProvider)
	if err.Error() != "organization, config and cryptoSuite are required to load CA config" {
		t.Fatalf("Expected error without config information. Got: %s", err.Error())
	}

	_, err = NewFabricCAClient(org1, configImp, nil)
	if err.Error() != "organization, config and cryptoSuite are required to load CA config" {
		t.Fatalf("Expected error without cryptosuite information. Got: %s", err.Error())
	}
}

// TestCreateNewFabricCAClientCAConfigMissingFailure will test newFabricCA Client creation with with CAConfig
func TestCreateNewFabricCAClientCAConfigMissingFailure(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()
	mockConfig := mock_apiconfig.NewMockConfig(mockCtrl)

	mockConfig.EXPECT().CAConfig(org1).Return(nil, errors.New("CAConfig error"))

	_, err := NewFabricCAClient(org1, mockConfig, cryptoSuiteProvider)
	if err.Error() != "CAConfig error" {
		t.Fatalf("Expected error from CAConfig. Got: %s", err.Error())
	}
}

// TestCreateNewFabricCAClientCertFilesMissingFailure will test newFabricCA Client creation with missing CA Cert files
func TestCreateNewFabricCAClientCertFilesMissingFailure(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()
	mockConfig := mock_apiconfig.NewMockConfig(mockCtrl)
	mockConfig.EXPECT().CAConfig(org1).Return(&config.CAConfig{URL: ""}, nil)
	mockConfig.EXPECT().CAServerCertPaths(org1).Return(nil, errors.New("CAServerCertPaths error"))
	_, err := NewFabricCAClient(org1, mockConfig, cryptoSuiteProvider)
	if err.Error() != "CAServerCertPaths error" {
		t.Fatalf("Expected error from CAServerCertPaths. Got: %s", err.Error())
	}
}

// TestCreateNewFabricCAClientCertFileErrorFailure will test newFabricCA Client creation with missing CA Cert files, additional scenario
func TestCreateNewFabricCAClientCertFileErrorFailure(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()
	mockConfig := mock_apiconfig.NewMockConfig(mockCtrl)
	mockConfig.EXPECT().CAConfig(org1).Return(&config.CAConfig{URL: ""}, nil)
	mockConfig.EXPECT().CAServerCertPaths(org1).Return([]string{"test"}, nil)
	mockConfig.EXPECT().CAClientCertPath(org1).Return("", errors.New("CAClientCertPath error"))
	_, err := NewFabricCAClient(org1, mockConfig, cryptoSuiteProvider)
	if err.Error() != "CAClientCertPath error" {
		t.Fatalf("Expected error from CAClientCertPath. Got: %s", err.Error())
	}
}

// TestCreateNewFabricCAClientKeyFileErrorFailure will test newFabricCA Client creation with missing CA Cert files and missing key
func TestCreateNewFabricCAClientKeyFileErrorFailure(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()
	mockConfig := mock_apiconfig.NewMockConfig(mockCtrl)
	mockConfig.EXPECT().CAConfig(org1).Return(&config.CAConfig{URL: ""}, nil)
	mockConfig.EXPECT().CAServerCertPaths(org1).Return([]string{"test"}, nil)
	mockConfig.EXPECT().CAClientCertPath(org1).Return("", nil)
	mockConfig.EXPECT().CAClientKeyPath(org1).Return("", errors.New("CAClientKeyPath error"))
	_, err := NewFabricCAClient(org1, mockConfig, cryptoSuiteProvider)
	if err.Error() != "CAClientKeyPath error" {
		t.Fatalf("Expected error from CAClientKeyPath. Got: %s", err.Error())
	}
}

// TestCreateValidBCCSPOptsForNewFabricClient test newFabricCA Client creation with valid inputs, successful scenario
func TestCreateValidBCCSPOptsForNewFabricClient(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()
	mockConfig := mock_apiconfig.NewMockConfig(mockCtrl)
	clientMockObject := &config.ClientConfig{Organization: "org1", Logging: config.LoggingType{Level: "info"}, CryptoConfig: config.CCType{Path: "test/path"}}

	mockConfig.EXPECT().CAConfig(org1).Return(&config.CAConfig{}, nil)
	mockConfig.EXPECT().CAServerCertPaths(org1).Return([]string{"test"}, nil)
	mockConfig.EXPECT().CAClientCertPath(org1).Return("", nil)
	mockConfig.EXPECT().CAClientKeyPath(org1).Return("", nil)
	mockConfig.EXPECT().CAKeyStorePath().Return(os.TempDir())
	mockConfig.EXPECT().Client().Return(clientMockObject, nil)
	mockConfig.EXPECT().SecurityProvider().Return("SW")
	mockConfig.EXPECT().SecurityAlgorithm().Return("SHA2")
	mockConfig.EXPECT().SecurityLevel().Return(256)
	mockConfig.EXPECT().KeyStorePath().Return("/tmp/msp")
	mockConfig.EXPECT().Ephemeral().Return(false)

	newCryptosuiteProvider, err := cryptosuiteimpl.GetSuiteByConfig(mockConfig)

	if err != nil {
		t.Fatalf("Expected fabric client ryptosuite to be created with SW BCCS provider, but got %v", err.Error())
	}

	_, err = NewFabricCAClient(org1, mockConfig, newCryptosuiteProvider)
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
	var apiCA ca.FabricCAClient
	var ca FabricCA

	apiCA = &ca
	if apiCA == nil {
		t.Fatalf("this shouldn't happen.")
	}
}
