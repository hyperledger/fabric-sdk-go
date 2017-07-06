/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package fabricca

import (
	"fmt"
	"io/ioutil"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/hyperledger/fabric-sdk-go/api/apiconfig/mocks"

	config "github.com/hyperledger/fabric-sdk-go/api/apiconfig"
	ca "github.com/hyperledger/fabric-sdk-go/api/apifabca"
	"github.com/hyperledger/fabric/bccsp"
	bccspFactory "github.com/hyperledger/fabric/bccsp/factory"

	"github.com/hyperledger/fabric-sdk-go/pkg/fabric-ca-client/mocks"
)

var configImp config.Config
var org1 = "peerorg1"

// Load testing config
func TestMain(m *testing.M) {
	configImp = mocks.NewMockConfig("http://localhost:8090")
	// Start Http Server
	go mocks.StartFabricCAMockServer("localhost:8090")
	// Allow HTTP server to start
	time.Sleep(1 * time.Second)
	os.Exit(m.Run())
}

func TestEnroll(t *testing.T) {

	fabricCAClient, err := NewFabricCAClient(configImp, org1)
	if err != nil {
		t.Fatalf("NewFabricCAClient return error: %v", err)
	}
	_, _, err = fabricCAClient.Enroll("", "user1")
	if err == nil {
		t.Fatalf("Enroll didn't return error")
	}
	if err.Error() != "enrollmentID is empty" {
		t.Fatalf("Enroll didn't return right error")
	}
	_, _, err = fabricCAClient.Enroll("test", "")
	if err == nil {
		t.Fatalf("Enroll didn't return error")
	}
	if err.Error() != "enrollmentSecret is empty" {
		t.Fatalf("Enroll didn't return right error")
	}
	_, _, err = fabricCAClient.Enroll("enrollmentID", "enrollmentSecret")
	if err != nil {
		t.Fatalf("fabricCAClient Enroll return error %v", err)
	}

	wrongConfigImp := mocks.NewMockConfig("http://localhost:8091")
	fabricCAClient, err = NewFabricCAClient(wrongConfigImp, org1)
	if err != nil {
		t.Fatalf("NewFabricCAClient return error: %v", err)
	}
	_, _, err = fabricCAClient.Enroll("enrollmentID", "enrollmentSecret")
	if err == nil {
		t.Fatalf("Enroll didn't return error")
	}
	if !strings.Contains(err.Error(), "Enroll failed") {
		t.Fatalf("Expected error when fabric-ca is down. Got: %s", err)
	}

}

func TestRegister(t *testing.T) {

	fabricCAClient, err := NewFabricCAClient(configImp, org1)
	if err != nil {
		t.Fatalf("NewFabricCAClient returned error: %v", err)
	}
	user := mocks.NewMockUser("test")
	// Register with nil request
	_, err = fabricCAClient.Register(user, nil)
	if err == nil {
		t.Fatalf("Expected error with nil request")
	}
	if err.Error() != "Registration request cannot be nil" {
		t.Fatalf("Expected error with nil request. Got: %s", err.Error())
	}

	//Register with nil user
	_, err = fabricCAClient.Register(nil, &ca.RegistrationRequest{})
	if err == nil {
		t.Fatalf("Expected error with nil user")
	}
	if err.Error() != "Error creating signing identity: Valid user required to create signing identity" {
		t.Fatalf("Expected error with nil user. Got: %s", err.Error())
	}
	// Register with nil user cert and key
	_, err = fabricCAClient.Register(user, &ca.RegistrationRequest{})
	if err == nil {
		t.Fatalf("Expected error without user enrolment information")
	}
	if err.Error() != "Error creating signing identity: Unable to read user enrolment information to create signing identity" {
		t.Fatalf("Expected error without user enrolment information. Got: %s", err.Error())
	}

	user.SetEnrollmentCertificate(readCert(t))
	key, err := bccspFactory.GetDefault().KeyGen(&bccsp.ECDSAP256KeyGenOpts{})
	if err != nil {
		t.Fatalf("KeyGen return error %v", err)
	}
	user.SetPrivateKey(key)
	// Register without registration name parameter
	_, err = fabricCAClient.Register(user, &ca.RegistrationRequest{})
	if err.Error() != "Error Registering User: Register was called without a Name set" {
		t.Fatalf("Expected error without registration information. Got: %s", err.Error())
	}
	// Register without registration affiliation parameter
	_, err = fabricCAClient.Register(user, &ca.RegistrationRequest{Name: "test"})
	if err.Error() != "Error Registering User: Registration request does not have an affiliation" {
		t.Fatalf("Expected error without registration information. Got: %s", err.Error())
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

func TestRevoke(t *testing.T) {

	fabricCAClient, err := NewFabricCAClient(configImp, org1)
	if err != nil {
		t.Fatalf("NewFabricCAClient returned error: %v", err)
	}
	mockKey := &mocks.MockKey{}
	user := mocks.NewMockUser("test")
	// Revoke with nil request
	err = fabricCAClient.Revoke(user, nil)
	if err == nil {
		t.Fatalf("Expected error with nil request")
	}
	if err.Error() != "Revocation request cannot be nil" {
		t.Fatalf("Expected error with nil request. Got: %s", err.Error())
	}
	//Revoke with nil user
	err = fabricCAClient.Revoke(nil, &ca.RevocationRequest{})
	if err == nil {
		t.Fatalf("Expected error with nil user")
	}
	if err.Error() != "Error creating signing identity: Valid user required to create signing identity" {
		t.Fatalf("Expected error with nil user. Got: %s", err.Error())
	}
	user.SetEnrollmentCertificate(readCert(t))
	user.SetPrivateKey(mockKey)
	err = fabricCAClient.Revoke(user, &ca.RevocationRequest{})
	if err == nil {
		t.Fatalf("Expected decoding error with test cert")
	}
}

func TestReenroll(t *testing.T) {

	fabricCAClient, err := NewFabricCAClient(configImp, org1)
	if err != nil {
		t.Fatalf("NewFabricCAClient returned error: %v", err)
	}
	user := mocks.NewMockUser("")
	// Reenroll with nil user
	_, _, err = fabricCAClient.Reenroll(nil)
	if err == nil {
		t.Fatalf("Expected error with nil user")
	}
	if err.Error() != "User does not exist" {
		t.Fatalf("Expected error with nil user. Got: %s", err.Error())
	}
	// Reenroll with user.Name is empty
	_, _, err = fabricCAClient.Reenroll(user)
	if err == nil {
		t.Fatalf("Expected error with user.Name is empty")
	}
	if err.Error() != "User is empty" {
		t.Fatalf("Expected error with user.Name is empty. Got: %s", err.Error())
	}
	// Reenroll with user.EnrollmentCertificate is empty
	user = mocks.NewMockUser("testUser")
	_, _, err = fabricCAClient.Reenroll(user)
	if err == nil {
		t.Fatalf("Expected error with user.EnrollmentCertificate is empty")
	}
	if err.Error() != "Reenroll has failed; Cannot create user identity: Unable to read user enrolment information to create signing identity" {
		t.Fatalf("Expected error with user.EnrollmentCertificate is empty. Got: %s", err.Error())
	}
	// Reenroll with appropriate user
	user.SetEnrollmentCertificate(readCert(t))
	key, err := bccspFactory.GetDefault().KeyGen(&bccsp.ECDSAP256KeyGenOpts{})
	if err != nil {
		t.Fatalf("KeyGen return error %v", err)
	}
	user.SetPrivateKey(key)
	_, _, err = fabricCAClient.Reenroll(user)
	if err != nil {
		t.Fatalf("Reenroll return error %v", err)
	}

	// Reenroll with wrong fabric-ca server url
	wrongConfigImp := mocks.NewMockConfig("http://localhost:8091")
	fabricCAClient, err = NewFabricCAClient(wrongConfigImp, org1)
	if err != nil {
		t.Fatalf("NewFabricCAClient return error: %v", err)
	}
	_, _, err = fabricCAClient.Reenroll(user)
	if err == nil {
		t.Fatalf("Expected error with wrong fabric-ca server url")
	}
	if !strings.Contains(err.Error(), "ReEnroll failed: POST failure") {
		t.Fatalf("Expected error with wrong fabric-ca server url. Got: %s", err.Error())
	}
}

func TestGetCAName(t *testing.T) {

	fabricCAClient, err := NewFabricCAClient(configImp, org1)
	if err != nil {
		t.Fatalf("NewFabricCAClient returned error: %v", err)
	}
	if fabricCAClient.CAName() != "test" {
		t.Fatalf("CAName returned wrong value: %s", fabricCAClient.CAName())
	}
}

func TestCreateNewFabricCAClient(t *testing.T) {

	_, err := NewFabricCAClient(configImp, "")
	if err.Error() != "Organization and config are required to load CA config" {
		t.Fatalf("Expected error without oganization information. Got: %s", err.Error())
	}

	_, err = NewFabricCAClient(nil, org1)
	if err.Error() != "Organization and config are required to load CA config" {
		t.Fatalf("Expected error without config information. Got: %s", err.Error())
	}

	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()
	mockConfig := mock_apiconfig.NewMockConfig(mockCtrl)

	mockConfig.EXPECT().CAConfig(org1).Return(nil, fmt.Errorf("CAConfig error"))

	_, err = NewFabricCAClient(mockConfig, org1)
	if err.Error() != "CAConfig error" {
		t.Fatalf("Expected error from CAConfig. Got: %s", err.Error())
	}

	mockConfig.EXPECT().CAConfig(org1).Return(&config.CAConfig{}, nil)
	mockConfig.EXPECT().CAServerCertFiles(org1).Return(nil, fmt.Errorf("CAServerCertFiles error"))
	_, err = NewFabricCAClient(mockConfig, org1)
	if err.Error() != "CAServerCertFiles error" {
		t.Fatalf("Expected error from CAServerCertFiles. Got: %s", err.Error())
	}

	mockConfig.EXPECT().CAConfig(org1).Return(&config.CAConfig{}, nil)
	mockConfig.EXPECT().CAServerCertFiles(org1).Return([]string{"test"}, nil)
	mockConfig.EXPECT().CAClientCertFile(org1).Return("", fmt.Errorf("CAClientCertFile error"))
	_, err = NewFabricCAClient(mockConfig, org1)
	if err.Error() != "CAClientCertFile error" {
		t.Fatalf("Expected error from CAClientCertFile. Got: %s", err.Error())
	}

	mockConfig.EXPECT().CAConfig(org1).Return(&config.CAConfig{}, nil)
	mockConfig.EXPECT().CAServerCertFiles(org1).Return([]string{"test"}, nil)
	mockConfig.EXPECT().CAClientCertFile(org1).Return("", nil)
	mockConfig.EXPECT().CAClientKeyFile(org1).Return("", fmt.Errorf("CAClientKeyFile error"))
	_, err = NewFabricCAClient(mockConfig, org1)
	if err.Error() != "CAClientKeyFile error" {
		t.Fatalf("Expected error from CAClientKeyFile. Got: %s", err.Error())
	}

	mockConfig.EXPECT().CAConfig(org1).Return(&config.CAConfig{}, nil)
	mockConfig.EXPECT().CAServerCertFiles(org1).Return([]string{"test"}, nil)
	mockConfig.EXPECT().CAClientCertFile(org1).Return("", nil)
	mockConfig.EXPECT().CAClientKeyFile(org1).Return("", nil)
	mockConfig.EXPECT().CAKeyStorePath().Return("/\\wq")
	mockConfig.EXPECT().CSPConfig().Return(nil)
	_, err = NewFabricCAClient(mockConfig, org1)
	if !strings.Contains(err.Error(), "New fabricCAClient failed") {
		t.Fatalf("Expected error from client init. Got: %s", err.Error())
	}

}

// Reads a random cert for testing
func readCert(t *testing.T) []byte {
	cert, err := ioutil.ReadFile("../../test/fixtures/root.pem")
	if err != nil {
		t.Fatalf("Error reading cert: %s", err.Error())
	}
	return cert
}

func TestInterfaces(t *testing.T) {
	var apiCA ca.FabricCAClient
	var ca FabricCA

	apiCA = &ca
	if apiCA == nil {
		t.Fatalf("this shouldn't happen.")
	}
}
