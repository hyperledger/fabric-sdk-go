/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package integration

import (
	"bytes"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"math/rand"
	"os"
	"strconv"
	"testing"
	"time"

	ca "github.com/hyperledger/fabric-sdk-go/api/apifabca"

	config "github.com/hyperledger/fabric-sdk-go/api/apiconfig"

	client "github.com/hyperledger/fabric-sdk-go/pkg/fabric-client"
	"github.com/hyperledger/fabric-sdk-go/pkg/fabric-client/identity"
	kvs "github.com/hyperledger/fabric-sdk-go/pkg/fabric-client/keyvaluestore"
	bccspFactory "github.com/hyperledger/fabric/bccsp/factory"

	fabricCAClient "github.com/hyperledger/fabric-sdk-go/pkg/fabric-ca-client"
)

var org1Name = "peerorg1"
var org2Name = "peerorg2"
var testFabricCAConfig config.Config

func TestMain(m *testing.M) {
	var err error

	testSetup := BaseSetupImpl{
		ConfigFile: "../fixtures/config/config_test.yaml",
	}

	testFabricCAConfig, err = testSetup.InitConfig()
	if err != nil {
		fmt.Printf("Failed InitConfig [%s]\n", err)
		os.Exit(1)
	}

	os.Exit(m.Run())
}

// This test loads/enrols an admin user
// Using the admin, it registers, enrols, and revokes a test user
func TestRegisterEnrollRevoke(t *testing.T) {
	mspID, err := testFabricCAConfig.MspID(org1Name)
	if err != nil {
		t.Fatalf("GetMspId() returned error: %v", err)
	}

	caConfig, err := testFabricCAConfig.CAConfig(org1Name)
	if err != nil {
		t.Fatalf("GetCAConfig returned error: %s", err)
	}

	client := client.NewClient(testFabricCAConfig)

	err = bccspFactory.InitFactories(testFabricCAConfig.CSPConfig())
	if err != nil {
		t.Fatalf("Failed getting ephemeral software-based BCCSP [%s]", err)
	}

	cryptoSuite := bccspFactory.GetDefault()

	client.SetCryptoSuite(cryptoSuite)
	stateStore, err := kvs.CreateNewFileKeyValueStore("/tmp/enroll_user")
	if err != nil {
		t.Fatalf("CreateNewFileKeyValueStore return error[%s]", err)
	}
	client.SetStateStore(stateStore)

	caClient, err := fabricCAClient.NewFabricCAClient(testFabricCAConfig, org1Name)
	if err != nil {
		t.Fatalf("NewFabricCAClient return error: %v", err)
	}

	// Admin user is used to register, enrol and revoke a test user
	adminUser, err := client.LoadUserFromStateStore("admin")

	if err != nil {
		t.Fatalf("client.LoadUserFromStateStore return error: %v", err)
	}
	if adminUser == nil {
		key, cert, err := caClient.Enroll("admin", "adminpw")
		if err != nil {
			t.Fatalf("Enroll return error: %v", err)
		}
		if key == nil {
			t.Fatalf("private key return from Enroll is nil")
		}
		if cert == nil {
			t.Fatalf("cert return from Enroll is nil")
		}

		certPem, _ := pem.Decode(cert)
		if err != nil {
			t.Fatalf("pem Decode return error: %v", err)
		}

		cert509, err := x509.ParseCertificate(certPem.Bytes)
		if err != nil {
			t.Fatalf("x509 ParseCertificate return error: %v", err)
		}
		if cert509.Subject.CommonName != "admin" {
			t.Fatalf("CommonName in x509 cert is not the enrollmentID")
		}
		adminUser2 := identity.NewUser("admin", mspID)
		adminUser2.SetPrivateKey(key)
		adminUser2.SetEnrollmentCertificate(cert)
		err = client.SaveUserToStateStore(adminUser2, false)
		if err != nil {
			t.Fatalf("client.SaveUserToStateStore return error: %v", err)
		}
		adminUser, err = client.LoadUserFromStateStore("admin")
		if err != nil {
			t.Fatalf("client.LoadUserFromStateStore return error: %v", err)
		}
		if adminUser == nil {
			t.Fatalf("client.LoadUserFromStateStore return nil")
		}
	}

	// Register a random user
	userName := createRandomName()
	registerRequest := ca.RegistrationRequest{
		Name:        userName,
		Type:        "user",
		Affiliation: "org1.department1",
		CAName:      caConfig.Name,
	}
	enrolmentSecret, err := caClient.Register(adminUser, &registerRequest)
	if err != nil {
		t.Fatalf("Error from Register: %s", err)
	}
	fmt.Printf("Registered User: %s, Secret: %s\n", userName, enrolmentSecret)
	// Enrol the previously registered user
	ekey, ecert, err := caClient.Enroll(userName, enrolmentSecret)
	if err != nil {
		t.Fatalf("Error enroling user: %s", err.Error())
	}
	//re-enroll
	fmt.Printf("** Attempt to re-enrolled user:  '%s'\n", userName)
	//create new user object and set certificate and private key of the previously enrolled user
	enrolleduser := identity.NewUser(userName, mspID)
	enrolleduser.SetEnrollmentCertificate(ecert)
	enrolleduser.SetPrivateKey(ekey)
	//reenroll
	_, reenrollCert, err := caClient.Reenroll(enrolleduser)
	if err != nil {
		t.Fatalf("Error Reenroling user: %s", err.Error())
	}
	fmt.Printf("** User '%s' was re-enrolled \n", userName)
	if bytes.Equal(ecert, reenrollCert) {
		t.Fatalf("Error Reenroling user. Enrollmet and Reenrollment certificates are the same.")
	}

	revokeRequest := ca.RevocationRequest{Name: userName, CAName: "ca-org1"}
	err = caClient.Revoke(adminUser, &revokeRequest)
	if err != nil {
		t.Fatalf("Error from Revoke: %s", err)
	}

}

func TestEnrolOrg2(t *testing.T) {
	caClient, err := fabricCAClient.NewFabricCAClient(testFabricCAConfig, org2Name)
	if err != nil {
		t.Fatalf("NewFabricCAClient return error: %v", err)
	}

	key, cert, err := caClient.Enroll("admin", "adminpw")
	if err != nil {
		t.Fatalf("Enroll returned error: %v", err)
	}
	if key == nil {
		t.Fatalf("Expected enrol to return a private key")
	}
	if cert == nil {
		t.Fatalf("Expected enrol to return an enrolment cert")
	}
}

func createRandomName() string {
	rand.Seed(time.Now().UnixNano())
	return "user" + strconv.Itoa(rand.Intn(500000))
}
