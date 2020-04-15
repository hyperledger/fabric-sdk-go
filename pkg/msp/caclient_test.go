/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package msp

import (
	"errors"
	"fmt"
	"path/filepath"
	"strings"
	"testing"

	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/test/mockfab"

	"github.com/golang/mock/gomock"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/core"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/msp"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/test/mockcontext"
	mockmspApi "github.com/hyperledger/fabric-sdk-go/pkg/common/providers/test/mockmsp"
	"github.com/hyperledger/fabric-sdk-go/pkg/core/config"
	"github.com/hyperledger/fabric-sdk-go/pkg/core/config/endpoint"
	"github.com/hyperledger/fabric-sdk-go/pkg/core/config/lookup"
	bccspwrapper "github.com/hyperledger/fabric-sdk-go/pkg/core/cryptosuite/bccsp/wrapper"
	"github.com/hyperledger/fabric-sdk-go/pkg/core/mocks"
	"github.com/hyperledger/fabric-sdk-go/pkg/fab"
	"github.com/hyperledger/fabric-sdk-go/pkg/msp/api"
	"github.com/hyperledger/fabric-sdk-go/pkg/msp/test/mockmsp"
)

// TestEnrollAndReenroll tests enrol/reenroll scenarios
func TestEnrollAndReenroll(t *testing.T) {

	f := textFixture{}
	f.setup()
	defer f.close()

	orgMSPID := mspIDByOrgName(t, f.endpointConfig, org1)

	// Empty enrollment ID
	err := f.caClient.Enroll(&api.EnrollmentRequest{Name: "", Secret: "user1"})
	if err == nil {
		t.Fatal("Enroll didn't return error")
	}

	// Empty enrollment secret
	err = f.caClient.Enroll(&api.EnrollmentRequest{Name: "enrolledUsername", Secret: ""})
	if err == nil {
		t.Fatal("Enroll didn't return error")
	}

	// Successful enrollment
	enrollUsername := createRandomName()
	_, err = f.userStore.Load(msp.IdentityIdentifier{MSPID: orgMSPID, ID: enrollUsername})
	if err != msp.ErrUserNotFound {
		t.Fatal("Expected to not find user in user store")
	}
	err = f.caClient.Enroll(&api.EnrollmentRequest{Name: enrollUsername, Secret: "enrollmentSecret"})
	if err != nil {
		t.Fatalf("identityManager Enroll return error %s", err)
	}
	enrolledUserData, err := f.userStore.Load(msp.IdentityIdentifier{MSPID: orgMSPID, ID: enrollUsername})
	if err != nil {
		t.Fatal("Expected to load user from user store")
	}

	// Reenroll with empty user
	err = f.caClient.Reenroll(&api.ReenrollmentRequest{Name: ""})
	if err == nil {
		t.Fatal("Expected error with enpty user")
	}
	if err.Error() != "user name missing" {
		t.Fatalf("Expected error user required. Got: %s", err)
	}

	// Reenroll with appropriate user
	reenrollWithAppropriateUser(f, t, enrolledUserData)
}

func reenrollWithAppropriateUser(f textFixture, t *testing.T, enrolledUserData *msp.UserData) {
	iManager, ok := f.identityManagerProvider.IdentityManager("org1")
	if !ok {
		t.Fatal("failed to get identity manager")
	}
	enrolledUser, err := iManager.(*IdentityManager).NewUser(enrolledUserData)
	if err != nil {
		t.Fatalf("newUser return error %s", err)
	}
	err = f.caClient.Reenroll(&api.ReenrollmentRequest{Name: enrolledUser.Identifier().ID})
	if err != nil {
		t.Fatalf("Reenroll return error %s", err)
	}
}

// TestWrongURL tests creation of CAClient with wrong URL
func TestWrongURL(t *testing.T) {

	f := textFixture{}
	f.setup()
	defer f.close()

	configBackend, err := getInvalidURLBackend()
	if err != nil {
		panic(fmt.Sprintf("Failed to get config backend: %s", err))
	}

	wrongURLIdentityConfig, err := ConfigFromBackend(configBackend...)
	if err != nil {
		panic(fmt.Sprintf("Failed to read config: %s", err))
	}

	wrongURLEndpointConfig, err := fab.ConfigFromBackend(configBackend...)
	if err != nil {
		panic(fmt.Sprintf("Failed to read config: %s", err))
	}

	iManager, ok := f.identityManagerProvider.IdentityManager("Org1")
	if !ok {
		t.Fatal("failed to get identity manager")
	}

	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()
	mockContext := mockcontext.NewMockClient(mockCtrl)
	mockContext.EXPECT().EndpointConfig().Return(wrongURLEndpointConfig).AnyTimes()
	mockContext.EXPECT().IdentityConfig().Return(wrongURLIdentityConfig).AnyTimes()
	mockContext.EXPECT().CryptoSuite().Return(f.cryptoSuite).AnyTimes()
	mockContext.EXPECT().UserStore().Return(f.userStore).AnyTimes()
	mockContext.EXPECT().IdentityManager("Org1").Return(iManager, true).AnyTimes()

	//f.caClient, err = NewCAClient(org1, f.identityManager, f.userStore, f.cryptoSuite, wrongURLConfigConfig)
	f.caClient, err = NewCAClient(org1, mockContext)
	if err != nil {
		t.Fatalf("NewidentityManagerClient return error: %s", err)
	}
	err = f.caClient.Enroll(&api.EnrollmentRequest{Name: "enrollmentID", Secret: "enrollmentSecret"})
	if err == nil {
		t.Fatal("Enroll didn't return error")
	}

}

// TestNoConfiguredCAs tests creation of CAClient when there are no configured CAs
func TestNoConfiguredCAs(t *testing.T) {

	f := textFixture{}
	f.setup()
	defer f.close()

	configBackend, err := getNoCAConfigBackend()
	if err != nil {
		panic(fmt.Sprintf("Failed to get config backend: %s", err))
	}

	wrongURLEndpointConfig, err := fab.ConfigFromBackend(configBackend...)
	if err != nil {
		panic(fmt.Sprintf("Failed to read config: %s", err))
	}

	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()
	mockContext := mockcontext.NewMockClient(mockCtrl)
	mockContext.EXPECT().EndpointConfig().Return(wrongURLEndpointConfig).AnyTimes()
	mockContext.EXPECT().IdentityConfig().Return(f.identityConfig).AnyTimes()
	mockContext.EXPECT().CryptoSuite().Return(f.cryptoSuite).AnyTimes()
	mockContext.EXPECT().UserStore().Return(f.userStore).AnyTimes()

	_, err = NewCAClient(org1, mockContext)
	if err == nil || !strings.Contains(err.Error(), "no CAs configured") {
		t.Fatal("Expected error when there are no configured CAs")
	}

}

// TestRegister tests multiple scenarios of registering a test (mocked or nil user) and their certs
func TestRegister(t *testing.T) {

	f := textFixture{}
	f.setup()
	defer f.close()

	// Register with nil request
	_, err := f.caClient.Register(nil)
	if err == nil {
		t.Fatal("Expected error with nil request")
	}

	// Register without registration name parameter
	_, err = f.caClient.Register(&api.RegistrationRequest{})
	if err == nil {
		t.Fatal("Expected error without registration name parameter")
	}

	// Register with valid request
	var attributes []api.Attribute
	attributes = append(attributes, api.Attribute{Name: "test1", Value: "test2"})
	attributes = append(attributes, api.Attribute{Name: "test2", Value: "test3"})
	secret, err := f.caClient.Register(&api.RegistrationRequest{Name: "test", Affiliation: "test", Attributes: attributes})
	if err != nil {
		t.Fatalf("identityManager Register return error %s", err)
	}
	if secret != "mockSecretValue" {
		t.Fatalf("identityManager Register return wrong value %s", secret)
	}
}

// TestCreateIdentity tests creating identity
func TestCreateIdentity(t *testing.T) {

	f := textFixture{}
	f.setup()
	defer f.close()

	// Create with nil request
	_, err := f.caClient.CreateIdentity(nil)
	if err == nil {
		t.Fatal("Expected error with nil request")
	}

	// Create without required parameters
	_, err = f.caClient.CreateIdentity(&api.IdentityRequest{Affiliation: "Org1"})
	if err == nil || !strings.Contains(err.Error(), "ID is required") {
		t.Fatal("Expected error due to missing required parameters")
	}

	// Create identity with valid request
	var attributes []api.Attribute
	attributes = append(attributes, api.Attribute{Name: "test1", Value: "test2"})
	attributes = append(attributes, api.Attribute{Name: "test2", Value: "test3"})
	identity, err := f.caClient.CreateIdentity(&api.IdentityRequest{ID: "test", Affiliation: "test", Attributes: attributes})
	if err != nil {
		t.Fatalf("create identity return error %s", err)
	}
	if identity.Secret != "top-secret" {
		t.Fatalf("create identity returned wrong value %s", identity.Secret)
	}

	// Create identity with ID only
	identity, err = f.caClient.CreateIdentity(&api.IdentityRequest{ID: "test1"})
	if err != nil {
		t.Fatalf("create identity return error %s", err)
	}
	if identity.Secret != "top-secret" {
		t.Fatalf("create identity returned wrong value %s", identity.Secret)
	}
}

// TestModifyIdentity tests updating identity
func TestModifyIdentity(t *testing.T) {

	f := textFixture{}
	f.setup()
	defer f.close()

	// Update with nil request
	_, err := f.caClient.ModifyIdentity(nil)
	if err == nil {
		t.Fatal("Expected error with nil request")
	}

	// Update without required parameters
	_, err = f.caClient.ModifyIdentity(&api.IdentityRequest{Affiliation: "Org1"})
	if err == nil || !strings.Contains(err.Error(), "ID is required") {
		t.Fatal("Expected error due to missing required parameters")
	}

	// Update identity with valid request
	identity, err := f.caClient.ModifyIdentity(&api.IdentityRequest{ID: "123", Affiliation: "org2", Secret: "new-top-secret"})
	if err != nil {
		t.Fatalf("update identity return error %s", err)
	}
	if identity.Secret != "new-top-secret" {
		t.Fatalf("update identity returned wrong value: %s", identity.Secret)
	}

	// Update identity without affiliation
	identity, err = f.caClient.ModifyIdentity(&api.IdentityRequest{ID: "123", Secret: "new-top-secret"})
	if err != nil {
		t.Fatalf("update identity return error %s", err)
	}
	if identity.Secret != "new-top-secret" {
		t.Fatalf("update identity returned wrong value: %s", identity.Secret)
	}
}

// TestRemoveIdentity tests removing an identity
func TestRemoveIdentity(t *testing.T) {

	f := textFixture{}
	f.setup()
	defer f.close()

	// Remove with nil request
	_, err := f.caClient.RemoveIdentity(nil)
	if err == nil {
		t.Fatal("Expected error with nil request")
	}

	// Remove without required parameters
	_, err = f.caClient.RemoveIdentity(&api.RemoveIdentityRequest{Force: false})
	if err == nil || !strings.Contains(err.Error(), "ID is required") {
		t.Fatal("Expected error due to missing required parameters")
	}

	// Remove identity with valid request
	identity, err := f.caClient.RemoveIdentity(&api.RemoveIdentityRequest{ID: "123"})
	if err != nil {
		t.Fatalf("remove identity return error %s", err)
	}
	if identity.Secret != "" {
		t.Fatalf("update identity returned wrong value: %s", identity.Secret)
	}
}

// TestCreateIdentity tests retrieving an identity by id
func TestGetIdentity(t *testing.T) {

	f := textFixture{}
	f.setup()
	defer f.close()

	// Get without required identity id parameter
	_, err := f.caClient.GetIdentity("", "")
	if err == nil || !strings.Contains(err.Error(), "id is required") {
		t.Fatal("Expected error due to missing required parameter")
	}

	// Get identity with valid request
	response, err := f.caClient.GetIdentity("123", "")
	if err != nil {
		t.Fatalf("get identity return error %s", err)
	}

	if response == nil {
		t.Fatal("get identity response is nil")
	}

}

// TestGetAllIdentities tests retrieving identities
func TestGetAllIdentities(t *testing.T) {

	f := textFixture{}
	f.setup()
	defer f.close()

	responses, err := f.caClient.GetAllIdentities("")
	if err != nil {
		t.Fatalf("get identities return error %s", err)
	}

	if len(responses) != 2 {
		t.Fatalf("expecting %d, got %d responses", 2, len(responses))
	}

}

// TestEmbeddedRegistar tests registration with embedded registrar identity
func TestEmbeddedRegistar(t *testing.T) {

	embeddedRegistrarBackend, err := getEmbeddedRegistrarConfigBackend()
	if err != nil {
		t.Fatalf("Failed to get config backend, cause: %s", err)
	}

	f := textFixture{}
	f.setup(embeddedRegistrarBackend...)
	defer f.close()

	// Register with valid request
	var attributes []api.Attribute
	attributes = append(attributes, api.Attribute{Name: "test1", Value: "test2"})
	attributes = append(attributes, api.Attribute{Name: "test2", Value: "test3"})
	secret, err := f.caClient.Register(&api.RegistrationRequest{Name: "withEmbeddedRegistrar", Affiliation: "test", Attributes: attributes})
	if err != nil {
		t.Fatalf("identityManager Register return error %s", err)
	}
	if secret != "mockSecretValue" {
		t.Fatalf("identityManager Register return wrong value %s", secret)
	}
}

// TestRegisterNoRegistrar tests registration with no configured registrar identity
func TestRegisterNoRegistrar(t *testing.T) {

	noRegistrarBackend, err := getNoRegistrarBackend()
	if err != nil {
		t.Fatalf("Failed to get config backend, cause: %s", err)
	}

	f := textFixture{}
	f.setup(noRegistrarBackend...)
	defer f.close()

	// Register with nil request
	_, err = f.caClient.Register(nil)
	if err != api.ErrCARegistrarNotFound {
		t.Fatalf("Expected ErrCARegistrarNotFound, got: %s", err)
	}

	// Register without registration name parameter
	_, err = f.caClient.Register(&api.RegistrationRequest{})
	if err != api.ErrCARegistrarNotFound {
		t.Fatalf("Expected ErrCARegistrarNotFound, got: %s", err)
	}

	// Register with valid request
	var attributes []api.Attribute
	attributes = append(attributes, api.Attribute{Name: "test1", Value: "test2"})
	attributes = append(attributes, api.Attribute{Name: "test2", Value: "test3"})
	_, err = f.caClient.Register(&api.RegistrationRequest{Name: "test", Affiliation: "test", Attributes: attributes})
	if err != api.ErrCARegistrarNotFound {
		t.Fatalf("Expected ErrCARegistrarNotFound, got: %s", err)
	}
}

// TestRevoke will test multiple revoking a user with a nil request or a nil user
// TODO - improve Revoke test coverage
func TestRevoke(t *testing.T) {

	f := textFixture{}
	f.setup()
	defer f.close()

	// Revoke with nil request
	_, err := f.caClient.Revoke(nil)
	if err == nil {
		t.Fatal("Expected error with nil request")
	}

	mockKey := bccspwrapper.GetKey(&mockmsp.MockKey{})
	user := mockmsp.NewMockSigningIdentity("test", "test")
	user.SetEnrollmentCertificate(readCert(t))
	user.SetPrivateKey(mockKey)

	_, err = f.caClient.Revoke(&api.RevocationRequest{})
	if err != nil {
		t.Fatalf("Revoke return error %s", err)
	}
}

// TestCAConfigError will test CAClient creation with bad CAConfig
func TestCAConfigError(t *testing.T) {

	f := textFixture{}
	f.setup()
	defer f.close()

	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()
	mockContext := mockcontext.NewMockClient(mockCtrl)

	mockIdentityConfig := mockmspApi.NewMockIdentityConfig(mockCtrl)
	mockIdentityConfig.EXPECT().CAConfig(org1CA).Return(nil, false)
	mockIdentityConfig.EXPECT().CredentialStorePath().Return(dummyUserStorePath).AnyTimes()

	mockContext.EXPECT().IdentityConfig().Return(mockIdentityConfig)
	mockContext.EXPECT().EndpointConfig().Return(f.endpointConfig).AnyTimes()

	_, err := NewCAClient(org1, mockContext)
	if err == nil || !strings.Contains(err.Error(), "error initializing CA [ca.org1.example.com]") {
		t.Fatalf("Expected error from CAConfig. Got: %s", err)
	}
}

// TestCAServerCertPathsError will test CAClient creation with missing CAServerCertPaths
func TestCAServerCertPathsError(t *testing.T) {

	f := textFixture{}
	f.setup()
	defer f.close()

	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()
	mockIdentityConfig := mockmspApi.NewMockIdentityConfig(mockCtrl)
	mockIdentityConfig.EXPECT().CAConfig(org1CA).Return(&msp.CAConfig{}, true).AnyTimes()
	mockIdentityConfig.EXPECT().CredentialStorePath().Return(dummyUserStorePath).AnyTimes()
	mockIdentityConfig.EXPECT().CAServerCerts(org1CA).Return(nil, false)

	mockContext := mockcontext.NewMockClient(mockCtrl)
	mockContext.EXPECT().EndpointConfig().Return(f.endpointConfig).AnyTimes()
	mockContext.EXPECT().IdentityConfig().Return(mockIdentityConfig).AnyTimes()
	mockContext.EXPECT().UserStore().Return(&mockmsp.MockUserStore{}).AnyTimes()
	mockContext.EXPECT().CryptoSuite().Return(f.cryptoSuite).AnyTimes()

	_, err := NewCAClient(org1, mockContext)
	if err == nil || !strings.Contains(err.Error(), "has no corresponding server certs in the configs") {
		t.Fatalf("Expected error from CAServerCertPaths. Got: %s", err)
	}
}

// TestCAClientCertPathError will test CAClient creation with missing CAClientCertPath
func TestCAClientCertPathError(t *testing.T) {

	f := textFixture{}
	f.setup()
	defer f.close()

	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()
	mockIdentityConfig := mockmspApi.NewMockIdentityConfig(mockCtrl)
	mockIdentityConfig.EXPECT().CAConfig(org1CA).Return(&msp.CAConfig{}, true).AnyTimes()
	mockIdentityConfig.EXPECT().CredentialStorePath().Return(dummyUserStorePath).AnyTimes()
	mockIdentityConfig.EXPECT().CAServerCerts(org1CA).Return([][]byte{[]byte("test")}, true)
	mockIdentityConfig.EXPECT().CAClientCert(org1CA).Return(nil, false)

	mockContext := mockcontext.NewMockClient(mockCtrl)
	mockContext.EXPECT().EndpointConfig().Return(f.endpointConfig).AnyTimes()
	mockContext.EXPECT().IdentityConfig().Return(mockIdentityConfig).AnyTimes()
	mockContext.EXPECT().UserStore().Return(&mockmsp.MockUserStore{}).AnyTimes()
	mockContext.EXPECT().CryptoSuite().Return(f.cryptoSuite).AnyTimes()

	_, err := NewCAClient(org1, mockContext)
	if err == nil || !strings.Contains(err.Error(), "has no corresponding client certs in the configs") {
		t.Fatalf("Expected error from CAClientCertPath. Got: %s", err)
	}
}

// TestCAClientKeyPathError will test CAClient creation with missing CAClientKeyPath
func TestCAClientKeyPathError(t *testing.T) {

	f := textFixture{}
	f.setup()
	defer f.close()

	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	mockIdentityConfig := mockmspApi.NewMockIdentityConfig(mockCtrl)
	mockIdentityConfig.EXPECT().CAConfig(org1CA).Return(&msp.CAConfig{}, true).AnyTimes()
	mockIdentityConfig.EXPECT().CredentialStorePath().Return(dummyUserStorePath).AnyTimes()
	mockIdentityConfig.EXPECT().CAServerCerts(org1CA).Return([][]byte{[]byte("test")}, true)
	mockIdentityConfig.EXPECT().CAClientCert(org1CA).Return([]byte(""), true)
	mockIdentityConfig.EXPECT().CAClientKey(org1CA).Return(nil, false)

	mockContext := mockcontext.NewMockClient(mockCtrl)
	mockContext.EXPECT().EndpointConfig().Return(f.endpointConfig).AnyTimes()
	mockContext.EXPECT().IdentityConfig().Return(mockIdentityConfig).AnyTimes()
	mockContext.EXPECT().UserStore().Return(&mockmsp.MockUserStore{}).AnyTimes()
	mockContext.EXPECT().CryptoSuite().Return(f.cryptoSuite).AnyTimes()

	_, err := NewCAClient(org1, mockContext)
	if err == nil || !strings.Contains(err.Error(), "has no corresponding client keys in the configs") {
		t.Fatalf("Expected error from CAClientKeyPath. Got: %s", err)
	}
}

// TestCAClientKeyPathError will test CAClient creation with bad TLSCACertPool
func TestCAClientTLSCACertPoolError(t *testing.T) {

	f := textFixture{}
	f.setup()
	defer f.close()

	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	certPoolErr := errors.New("certPoolErr")
	tlsCACertPool := &mockfab.MockCertPool{nil, certPoolErr}

	mockIdentityConfig := mockmspApi.NewMockIdentityConfig(mockCtrl)
	mockIdentityConfig.EXPECT().CAConfig(org1CA).Return(&msp.CAConfig{}, true).AnyTimes()
	mockIdentityConfig.EXPECT().CredentialStorePath().Return(dummyUserStorePath).AnyTimes()
	mockIdentityConfig.EXPECT().CAServerCerts(org1CA).Return([][]byte{[]byte("test")}, true)
	mockIdentityConfig.EXPECT().CAClientCert(org1CA).Return([]byte(""), true)
	mockIdentityConfig.EXPECT().CAClientKey(org1CA).Return([]byte("testCAclientkey"), true)
	mockIdentityConfig.EXPECT().TLSCACertPool().Return(tlsCACertPool)

	mockContext := mockcontext.NewMockClient(mockCtrl)
	mockContext.EXPECT().EndpointConfig().Return(f.endpointConfig).AnyTimes()
	mockContext.EXPECT().IdentityConfig().Return(mockIdentityConfig).AnyTimes()
	mockContext.EXPECT().UserStore().Return(&mockmsp.MockUserStore{}).AnyTimes()
	mockContext.EXPECT().CryptoSuite().Return(f.cryptoSuite).AnyTimes()

	_, err := NewCAClient(org1, mockContext)
	if err == nil || !strings.Contains(err.Error(), "couldn't load configured cert pool") {
		t.Fatalf("Expected error from TLSCACertPool. Got: %s", err)
	}
}

// TestInterfaces will test if the interface instantiation happens properly, ie no nil returned
func TestInterfaces(t *testing.T) {
	var apiClient api.CAClient
	var cl CAClientImpl

	apiClient = &cl
	if apiClient == nil {
		t.Fatal("this shouldn't happen.")
	}
}

func TestAddAffiliation(t *testing.T) {
	f := textFixture{}
	f.setup()
	defer f.close()

	// Add with nil request
	_, err := f.caClient.AddAffiliation(nil)
	if err == nil {
		t.Fatal("Expected error with nil request")
	}

	// Add without required parameters
	_, err = f.caClient.AddAffiliation(&api.AffiliationRequest{})
	if err == nil || !strings.Contains(err.Error(), "Name is required") {
		t.Fatal("Expected error due to missing required parameter")
	}

	resp, err := f.caClient.AddAffiliation(&api.AffiliationRequest{Name: "test1.com", Force: true})
	if err != nil {
		t.Fatalf("Add affiliation return error %s", err)
	}

	if resp.Name != "test1.com" {
		t.Fatalf("add affiliation returned wrong value %s", resp.Name)
	}
}

func TestModifyAffiliation(t *testing.T) {
	f := textFixture{}
	f.setup()
	defer f.close()

	// Modify with nil request
	_, err := f.caClient.ModifyAffiliation(nil)
	if err == nil {
		t.Fatal("Expected error with nil request")
	}

	// Modify without required parameters
	_, err = f.caClient.ModifyAffiliation(&api.ModifyAffiliationRequest{})
	if err == nil || !strings.Contains(err.Error(), "Name and NewName are required") {
		t.Fatal("Expected error due to missing required parameters")
	}

	resp, err := f.caClient.ModifyAffiliation(&api.ModifyAffiliationRequest{NewName: "test1new.com", AffiliationRequest: api.AffiliationRequest{Name: "123"}})
	if err != nil {
		t.Fatalf("Modify affiliation return error %s", err)
	}

	if resp.Name != "test1new.com" {
		t.Fatalf("Modify affiliation returned wrong value %s", resp.Name)
	}
}

func TestRemoveAffiliation(t *testing.T) {
	f := textFixture{}
	f.setup()
	defer f.close()

	// Remove with nil request
	_, err := f.caClient.RemoveAffiliation(nil)
	if err == nil {
		t.Fatal("Expected error with nil request")
	}

	// Remove without required parameters
	_, err = f.caClient.RemoveAffiliation(&api.AffiliationRequest{})
	if err == nil || !strings.Contains(err.Error(), "Name is required") {
		t.Fatal("Expected error due to missing required parameters")
	}

	resp, err := f.caClient.RemoveAffiliation(&api.AffiliationRequest{Name: "123"})
	if err != nil {
		t.Fatalf("Remove affiliation return error %s", err)
	}

	if resp.Name != "test1.com" {
		t.Fatalf("Remove affiliation returned wrong value %s", resp.Name)
	}
}

func TestGetAffiliation(t *testing.T) {
	f := textFixture{}
	f.setup()
	defer f.close()

	// Get without required parameter
	_, err := f.caClient.GetAffiliation("", "")
	if err == nil || !strings.Contains(err.Error(), "affiliation is required") {
		t.Fatal("Expected error due to missing required parameter")
	}

	// Get affiliation with valid request
	resp, err := f.caClient.GetAffiliation("123", "")
	if err != nil {
		t.Fatalf("Get affiliation return error %s", err)
	}

	if resp == nil {
		t.Fatal("Get affiliation response is nil")
	}
}

func TestGetAllAffiliations(t *testing.T) {
	f := textFixture{}
	f.setup()
	defer f.close()

	response, err := f.caClient.GetAllAffiliations("")
	if err != nil {
		t.Fatalf("Get affiliations return error %s", err)
	}

	if len(response.Affiliations) != 1 {
		t.Fatalf("expecting %d, got %d response", 1, len(response.Affiliations))
	}
}

func TestGetCAInfo(t *testing.T) {
	f := textFixture{}
	f.setup()
	defer f.close()

	resp, err := f.caClient.GetCAInfo()
	if err != nil {
		t.Fatalf("Get CA info return error %s", err)
	}

	if resp.CAName != "123" {
		t.Fatalf("expecting 123, got %s", resp.CAName)
	}
}

func getCustomBackend(configPath string) ([]core.ConfigBackend, error) {

	configBackends, err := config.FromFile(configPath)()
	if err != nil {
		return nil, err
	}
	backendMap := make(map[string]interface{})
	backendMap["client"], _ = configBackends[0].Lookup("client")
	backendMap["certificateAuthorities"], _ = configBackends[0].Lookup("certificateAuthorities")
	backendMap["entityMatchers"], _ = configBackends[0].Lookup("entityMatchers")
	backendMap["peers"], _ = configBackends[0].Lookup("peers")
	backendMap["organizations"], _ = configBackends[0].Lookup("organizations")
	backendMap["orderers"], _ = configBackends[0].Lookup("orderers")
	backendMap["channels"], _ = configBackends[0].Lookup("channels")

	backends := append([]core.ConfigBackend{}, &mocks.MockConfigBackend{KeyValueMap: backendMap})
	backends = append(backends, configBackends...)
	return backends, nil
}

func getInvalidURLBackend() ([]core.ConfigBackend, error) {

	configPath := filepath.Join(getConfigPath(), configTestFile)
	mockConfigBackend, err := getCustomBackend(configPath)
	if err != nil {
		return nil, err
	}

	//Create an invalid channel
	networkConfig := identityConfigEntity{}
	//get valid certificate authorities
	err = lookup.New(mockConfigBackend...).UnmarshalKey("certificateAuthorities", &networkConfig.CertificateAuthorities)
	if err != nil {
		return nil, err
	}

	//tamper URLs
	ca1Config := networkConfig.CertificateAuthorities["ca.org1.example.com"]
	ca1Config.URL = "http://localhost:8091"
	ca2Config := networkConfig.CertificateAuthorities["ca.org2.example.com"]
	ca2Config.URL = "http://localhost:8091"

	networkConfig.CertificateAuthorities["ca.org1.example.com"] = ca1Config
	networkConfig.CertificateAuthorities["ca.org2.example.com"] = ca2Config

	//Override backend with this new CertificateAuthorities config
	backendMap := make(map[string]interface{})
	backendMap["certificateAuthorities"] = networkConfig.CertificateAuthorities
	backends := append([]core.ConfigBackend{}, &mocks.MockConfigBackend{KeyValueMap: backendMap})
	backends = append(backends, mockConfigBackend...)

	return backends, nil
}

func getNoRegistrarBackend() ([]core.ConfigBackend, error) {

	configPath := filepath.Join(getConfigPath(), configTestFile)
	mockConfigBackend, err := getCustomBackend(configPath)
	if err != nil {
		return nil, err
	}

	//Create an invalid channel
	networkConfig := identityConfigEntity{}
	//get valid certificate authorities
	err = lookup.New(mockConfigBackend...).UnmarshalKey("certificateAuthorities", &networkConfig.CertificateAuthorities)
	if err != nil {
		return nil, err
	}

	//tamper URLs
	ca1Config := networkConfig.CertificateAuthorities["ca.org1.example.com"]
	ca1Config.Registrar = msp.EnrollCredentials{}
	ca2Config := networkConfig.CertificateAuthorities["ca.org2.example.com"]
	ca1Config.Registrar = msp.EnrollCredentials{}

	networkConfig.CertificateAuthorities["ca.org1.example.com"] = ca1Config
	networkConfig.CertificateAuthorities["ca.org2.example.com"] = ca2Config

	//Override backend with this new CertificateAuthorities config
	backendMap := make(map[string]interface{})
	backendMap["certificateAuthorities"] = networkConfig.CertificateAuthorities
	backends := append([]core.ConfigBackend{}, &mocks.MockConfigBackend{KeyValueMap: backendMap})
	backends = append(backends, mockConfigBackend...)

	return backends, nil
}

func getNoCAConfigBackend() ([]core.ConfigBackend, error) {

	configPath := filepath.Join(getConfigPath(), configTestFile)
	mockConfigBackend, err := getCustomBackend(configPath)
	if err != nil {
		return nil, err
	}

	//Create an empty network config
	networkConfig := identityConfigEntity{}
	//get valid certificate authorities
	err = lookup.New(mockConfigBackend...).UnmarshalKey("organizations", &networkConfig.Organizations)
	if err != nil {
		return nil, err
	}
	org1 := networkConfig.Organizations["org1"]

	//clear certificate authorities
	org1.CertificateAuthorities = []string{}
	networkConfig.Organizations["org1"] = org1

	backendMap := make(map[string]interface{})
	//Override backend with organization config having empty CertificateAuthorities
	backendMap["organizations"] = networkConfig.Organizations
	//Override backend with this nil empty CertificateAuthorities config
	backendMap["certificateAuthorities"] = networkConfig.CertificateAuthorities

	backends := append([]core.ConfigBackend{}, &mocks.MockConfigBackend{KeyValueMap: backendMap})
	backends = append(backends, mockConfigBackend...)

	return backends, nil
}

func getEmbeddedRegistrarConfigBackend() ([]core.ConfigBackend, error) {

	configPath := filepath.Join(getConfigPath(), configTestFile)
	mockConfigBackend, err := getCustomBackend(configPath)
	if err != nil {
		return nil, err
	}

	embeddedRegistrarID := "embeddedregistrar"

	//Create an empty network config
	networkConfig := identityConfigEntity{}
	//get valid certificate authorities
	err = lookup.New(mockConfigBackend...).UnmarshalKey("organizations", &networkConfig.Organizations)
	if err != nil {
		return nil, err
	}
	err = lookup.New(mockConfigBackend...).UnmarshalKey("certificateAuthorities", &networkConfig.CertificateAuthorities)
	if err != nil {
		return nil, err
	}
	//update with embedded registrar
	org1 := networkConfig.Organizations["org1"]
	org1.Users = make(map[string]endpoint.TLSKeyPair)
	org1.Users[embeddedRegistrarID] = endpoint.TLSKeyPair{
		Key: endpoint.TLSConfig{
			Pem: `-----BEGIN PRIVATE KEY-----
MIGHAgEAMBMGByqGSM49AgEGCCqGSM49AwEHBG0wawIBAQQgp4qKKB0WCEfx7XiB
5Ul+GpjM1P5rqc6RhjD5OkTgl5OhRANCAATyFT0voXX7cA4PPtNstWleaTpwjvbS
J3+tMGTG67f+TdCfDxWYMpQYxLlE8VkbEzKWDwCYvDZRMKCQfv2ErNvb
-----END PRIVATE KEY-----`,
		},
		Cert: endpoint.TLSConfig{
			Pem: `-----BEGIN CERTIFICATE-----
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
-----END CERTIFICATE-----`,
		},
	}
	networkConfig.Organizations["org1"] = org1

	//update network certificate authorities
	ca1Config := networkConfig.CertificateAuthorities["ca.org1.example.com"]
	ca1Config.Registrar = msp.EnrollCredentials{EnrollID: embeddedRegistrarID}
	ca2Config := networkConfig.CertificateAuthorities["ca.org2.example.com"]
	ca2Config.Registrar = msp.EnrollCredentials{EnrollID: embeddedRegistrarID}
	networkConfig.CertificateAuthorities["ca.org1.example.com"] = ca1Config
	networkConfig.CertificateAuthorities["ca.org2.example.com"] = ca2Config

	backendMap := make(map[string]interface{})
	//Override backend with updated organization config
	backendMap["organizations"] = networkConfig.Organizations
	//Override backend with updated certificate authorities config
	backendMap["certificateAuthorities"] = networkConfig.CertificateAuthorities

	backends := append([]core.ConfigBackend{}, &mocks.MockConfigBackend{KeyValueMap: backendMap})
	backends = append(backends, mockConfigBackend...)

	return backends, nil
}
