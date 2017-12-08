/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package identity

import (
	"testing"

	"io/ioutil"

	cryptosuite "github.com/hyperledger/fabric-sdk-go/pkg/cryptosuite/bccsp"
	"github.com/hyperledger/fabric-sdk-go/pkg/fabric-ca-client/mocks"
)

func TestUserMethods(t *testing.T) {
	user := NewUser("testUser", "testMSP")

	//test Name
	if user.Name() != "testUser" {
		t.Fatalf("NewUser create wrong user")
	}

	// test Roles
	var roles []string
	roles = append(roles, "admin")
	roles = append(roles, "user")
	user.SetRoles(roles)

	if user.Roles()[0] != "admin" {
		t.Fatalf("user.GetRoles() return wrong user")
	}
	if user.Roles()[1] != "user" {
		t.Fatalf("user.GetRoles() return wrong user")
	}

	// test PrivateKey
	privateKey := cryptosuite.GetKey(&mocks.MockKey{})

	user.SetPrivateKey(privateKey)

	returnKey := user.PrivateKey()
	if returnKey == nil {
		t.Fatalf("GetKey() after SetKey() returned nil.")
	}

	if returnKey != privateKey {
		t.Fatalf("user.SetKey() and GetKey() don't return matching keys.")
	}

	// test TCerts
	var attributes []string
	user.GenerateTcerts(1, attributes) // TODO implement test when function is implemented

	// test EnrolmentCert
	cert := readCert(t)
	user.SetEnrollmentCertificate(cert)
	setCert := user.EnrollmentCertificate()
	if len(cert) != len(setCert) {
		t.Fatal("user.SetEnrollmentCertificate did not set the same cert.")
	}

	// test MSP
	user.SetMspID("test")
	mspID := user.MspID()
	if mspID != "test" {
		t.Fatal("user.SetMspID Failed to MSP.")
	}

}

// Reads a random cert for testing
func readCert(t *testing.T) []byte {
	cert, err := ioutil.ReadFile("testdata/root.pem")
	if err != nil {
		t.Fatalf("Error reading cert: %s", err.Error())
	}
	return cert
}
