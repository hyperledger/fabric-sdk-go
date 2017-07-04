/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package fabricca

import (
	"fmt"
	"io/ioutil"
	"os"
	"testing"

	config "github.com/hyperledger/fabric-sdk-go/api/apiconfig"
	ca "github.com/hyperledger/fabric-sdk-go/api/apifabca"

	"github.com/hyperledger/fabric-sdk-go/pkg/fabric-ca-client/mocks"
)

var configImp config.Config
var org1 = "peerorg1"

// Load testing config
func TestMain(m *testing.M) {
	var err error
	configImp = mocks.NewMockConfig()
	if err != nil {
		fmt.Println(err.Error())
	}
	os.Exit(m.Run())
}

func TestEnrollWithMissingParameters(t *testing.T) {

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
}

func TestRegister(t *testing.T) {

	fabricCAClient, err := NewFabricCAClient(configImp, org1)
	if err != nil {
		t.Fatalf("NewFabricCAClient returned error: %v", err)
	}
	mockKey := &mocks.MockKey{}
	user := mocks.NewMockUser("test")
	// Register with nil request
	_, err = fabricCAClient.Register(user, nil)
	if err == nil {
		t.Fatalf("Expected error with nil request")
	}
	//Register with nil user
	_, err = fabricCAClient.Register(nil, &ca.RegistrationRequest{})
	if err == nil {
		t.Fatalf("Expected error with nil user")
	}
	// Register with nil user cert and key
	_, err = fabricCAClient.Register(user, &ca.RegistrationRequest{})
	if err == nil {
		t.Fatalf("Expected error without user enrolment information")
	}
	user.SetEnrollmentCertificate(readCert(t))
	user.SetPrivateKey(mockKey)
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
	_, err = fabricCAClient.Register(user, &ca.RegistrationRequest{Name: "test",
		Affiliation: "test", Attributes: attributes})
	if err == nil {
		t.Fatalf("Expected PEM decoding error with test cert")
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
	//Revoke with nil user
	err = fabricCAClient.Revoke(nil, &ca.RevocationRequest{})
	if err == nil {
		t.Fatalf("Expected error with nil user")
	}
	user.SetEnrollmentCertificate(readCert(t))
	user.SetPrivateKey(mockKey)
	err = fabricCAClient.Revoke(user, &ca.RevocationRequest{})
	if err == nil {
		t.Fatalf("Expected decoding error with test cert")
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
