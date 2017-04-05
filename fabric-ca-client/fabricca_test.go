/*
Copyright SecureKey Technologies Inc. All Rights Reserved.


Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at


      http://www.apache.org/licenses/LICENSE-2.0


Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package fabricca

import (
	"fmt"
	"io/ioutil"
	"os"
	"testing"

	"github.com/hyperledger/fabric-sdk-go/config"
	"github.com/hyperledger/fabric-sdk-go/fabric-ca-client/mocks"
	"github.com/hyperledger/fabric-sdk-go/fabric-client"
)

// Load testing config
func TestMain(m *testing.M) {
	err := config.InitConfig("../test/fixtures/config/config_test.yaml")
	if err != nil {
		fmt.Println(err.Error())
	}
	os.Exit(m.Run())

}

func TestEnrollWithMissingParameters(t *testing.T) {

	fabricCAClient, err := NewFabricCAClient()
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

	fabricCAClient, err := NewFabricCAClient()
	if err != nil {
		t.Fatalf("NewFabricCAClient returned error: %v", err)
	}
	mockKey := &mocks.MockKey{}
	user := fabricclient.NewUser("test")
	// Register with nil request
	_, err = fabricCAClient.Register(user, nil)
	if err == nil {
		t.Fatalf("Expected error with nil request")
	}
	//Register with nil user
	_, err = fabricCAClient.Register(nil, &RegistrationRequest{})
	if err == nil {
		t.Fatalf("Expected error with nil user")
	}
	// Register with nil user cert and key
	_, err = fabricCAClient.Register(user, &RegistrationRequest{})
	if err == nil {
		t.Fatalf("Expected error without user enrolment information")
	}
	user.SetEnrollmentCertificate(readCert(t))
	user.SetPrivateKey(mockKey)
	// Register without registration name parameter
	_, err = fabricCAClient.Register(user, &RegistrationRequest{})
	if err.Error() != "Error Registering User: Register was called without a Name set" {
		t.Fatalf("Expected error without registration information. Got: %s", err.Error())
	}
	// Register without registration affiliation parameter
	_, err = fabricCAClient.Register(user, &RegistrationRequest{Name: "test"})
	if err.Error() != "Error Registering User: Registration request does not have an affiliation" {
		t.Fatalf("Expected error without registration information. Got: %s", err.Error())
	}
	// Register with valid request
	var attributes []Attribute
	attributes = append(attributes, Attribute{Key: "test1", Value: "test2"})
	attributes = append(attributes, Attribute{Key: "test2", Value: "test3"})
	_, err = fabricCAClient.Register(user, &RegistrationRequest{Name: "test",
		Affiliation: "test", Attributes: attributes})
	if err == nil {
		t.Fatalf("Expected PEM decoding error with test cert")
	}
}

func TestRevoke(t *testing.T) {

	fabricCAClient, err := NewFabricCAClient()
	if err != nil {
		t.Fatalf("NewFabricCAClient returned error: %v", err)
	}
	mockKey := &mocks.MockKey{}
	user := fabricclient.NewUser("test")
	// Revoke with nil request
	err = fabricCAClient.Revoke(user, nil)
	if err == nil {
		t.Fatalf("Expected error with nil request")
	}
	//Revoke with nil user
	err = fabricCAClient.Revoke(nil, &RevocationRequest{})
	if err == nil {
		t.Fatalf("Expected error with nil user")
	}
	user.SetEnrollmentCertificate(readCert(t))
	user.SetPrivateKey(mockKey)
	err = fabricCAClient.Revoke(user, &RevocationRequest{})
	if err == nil {
		t.Fatalf("Expected decoding error with test cert")
	}
}

// Reads a random cert for testing
func readCert(t *testing.T) []byte {
	cert, err := ioutil.ReadFile("../test/fixtures/root.pem")
	if err != nil {
		t.Fatalf("Error reading cert: %s", err.Error())
	}
	return cert
}
