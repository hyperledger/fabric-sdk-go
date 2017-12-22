/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package fab

import (
	"bytes"
	"crypto/x509"
	"encoding/pem"
	"math/rand"
	"strconv"
	"testing"
	"time"

	"github.com/hyperledger/fabric-sdk-go/api/apiconfig"
	ca "github.com/hyperledger/fabric-sdk-go/api/apifabca"

	client "github.com/hyperledger/fabric-sdk-go/pkg/fabric-client"
	"github.com/hyperledger/fabric-sdk-go/pkg/fabric-client/identity"
	kvs "github.com/hyperledger/fabric-sdk-go/pkg/fabric-client/keyvaluestore"
	"github.com/hyperledger/fabric-sdk-go/pkg/fabric-client/peer"
	"github.com/hyperledger/fabric-sdk-go/pkg/fabric-client/signingmgr"

	cryptosuite "github.com/hyperledger/fabric-sdk-go/pkg/cryptosuite/bccsp"
	fabricCAClient "github.com/hyperledger/fabric-sdk-go/pkg/fabric-ca-client"
)

const (
	org1Name = "Org1"
	org2Name = "Org2"
)

// This test loads/enrols an admin user
// Using the admin, it registers, enrols, and revokes a test user
func TestRegisterEnrollRevoke(t *testing.T) {
	mspID, err := testFabricConfig.MspID(org1Name)
	if err != nil {
		t.Fatalf("GetMspId() returned error: %v", err)
	}

	caConfig, err := testFabricConfig.CAConfig(org1Name)
	if err != nil {
		t.Fatalf("GetCAConfig returned error: %s", err)
	}

	client := client.NewClient(testFabricConfig)

	cryptoSuiteProvider, err := cryptosuite.GetSuiteByConfig(testFabricConfig)
	if err != nil {
		t.Fatalf("Failed getting cryptosuite from config : %s", err)
	}

	client.SetCryptoSuite(cryptoSuiteProvider)
	stateStore, err := kvs.CreateNewFileKeyValueStore("/tmp/enroll_user")
	if err != nil {
		t.Fatalf("CreateNewFileKeyValueStore return error[%s]", err)
	}
	client.SetStateStore(stateStore)

	caClient, err := fabricCAClient.NewFabricCAClient(org1Name, testFabricConfig, cryptoSuiteProvider)
	if err != nil {
		t.Fatalf("NewFabricCAClient return error: %v", err)
	}

	// Admin user is used to register, enroll and revoke a test user
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
		if certPem == nil {
			t.Fatal("Fail to decode pem block")
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
		CAName:      caConfig.CAName,
	}
	enrolmentSecret, err := caClient.Register(adminUser, &registerRequest)
	if err != nil {
		t.Fatalf("Error from Register: %s", err)
	}
	t.Logf("Registered User: %s, Secret: %s", userName, enrolmentSecret)
	// Enrol the previously registered user
	ekey, ecert, err := caClient.Enroll(userName, enrolmentSecret)
	if err != nil {
		t.Fatalf("Error enroling user: %s", err.Error())
	}
	//re-enroll
	t.Logf("** Attempt to re-enrolled user:  '%s'", userName)
	//create new user object and set certificate and private key of the previously enrolled user
	enrolleduser := identity.NewUser(userName, mspID)
	enrolleduser.SetEnrollmentCertificate(ecert)
	enrolleduser.SetPrivateKey(ekey)
	//reenroll
	_, reenrollCert, err := caClient.Reenroll(enrolleduser)
	if err != nil {
		t.Fatalf("Error Reenroling user: %s", err.Error())
	}
	t.Logf("** User '%s' was re-enrolled", userName)
	if bytes.Equal(ecert, reenrollCert) {
		t.Fatalf("Error Reenroling user. Enrollmet and Reenrollment certificates are the same.")
	}

	revokeRequest := ca.RevocationRequest{Name: userName, CAName: "ca.org1.example.com"}
	err = caClient.Revoke(adminUser, &revokeRequest)
	if err != nil {
		t.Fatalf("Error from Revoke: %s", err)
	}

}

func TestEnrollOrg2(t *testing.T) {

	cryptoSuiteProvider, err := cryptosuite.GetSuiteByConfig(testFabricConfig)
	if err != nil {
		t.Fatalf("Failed getting cryptosuite from config : %s", err)
	}

	caClient, err := fabricCAClient.NewFabricCAClient(org2Name, testFabricConfig, cryptoSuiteProvider)
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

func TestEnrollAndTransact(t *testing.T) {
	mspID, err := testFabricConfig.MspID(org1Name)
	if err != nil {
		t.Fatalf("GetMspId() returned error: %v", err)
	}
	peers, err := testFabricConfig.PeersConfig(org1Name)
	if err != nil {
		t.Fatalf("Failed to get peer config : %s", err)
	}
	networkPeer := &apiconfig.NetworkPeer{PeerConfig: peers[0], MspID: mspID}
	testPeer, err := peer.NewPeerFromConfig(networkPeer, testFabricConfig)
	if err != nil {
		t.Fatalf("Failed to create peer from config : %s", err)
	}

	cryptoSuiteProvider, err := cryptosuite.GetSuiteByConfig(testFabricConfig)
	if err != nil {
		t.Fatalf("Failed getting cryptosuite from config : %s", err)
	}
	signingManager, err := signingmgr.NewSigningManager(cryptoSuiteProvider, testFabricConfig)
	if err != nil {
		t.Fatalf("Could not create signing manager: %s", err)
	}

	caClient, err := fabricCAClient.NewFabricCAClient(org1Name, testFabricConfig, cryptoSuiteProvider)
	if err != nil {
		t.Fatalf("NewFabricCAClient returned error: %v", err)
	}

	key, cert, err := caClient.Enroll("admin", "adminpw")
	if err != nil {
		t.Fatalf("Enroll returned error: %v", err)
	}

	myUser := identity.NewUser("myUser", mspID)
	myUser.SetEnrollmentCertificate(cert)
	myUser.SetPrivateKey(key)

	testClient := client.NewClient(testFabricConfig)
	testClient.SetUserContext(myUser)
	testClient.SetSigningManager(signingManager)

	_, err = testClient.QueryChannels(testPeer)
	if err != nil {
		t.Fatalf("Failed to query with enrolled user : %s", err)
	}
}

func createRandomName() string {
	rand.Seed(time.Now().UnixNano())
	return "user" + strconv.Itoa(rand.Intn(500000))
}
