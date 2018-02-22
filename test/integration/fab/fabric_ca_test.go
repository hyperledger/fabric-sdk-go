/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package fab

import (
	"bytes"
	"math/rand"
	"strconv"
	"testing"

	"github.com/hyperledger/fabric-sdk-go/pkg/context/api/core"

	cryptosuite "github.com/hyperledger/fabric-sdk-go/pkg/core/cryptosuite/bccsp/sw"
	client "github.com/hyperledger/fabric-sdk-go/pkg/fab"
	"github.com/hyperledger/fabric-sdk-go/pkg/fab/identity"
	"github.com/hyperledger/fabric-sdk-go/pkg/fab/peer"
	"github.com/hyperledger/fabric-sdk-go/pkg/fab/signingmgr"

	"github.com/hyperledger/fabric-sdk-go/pkg/context/api/fab"
	fabricCAClient "github.com/hyperledger/fabric-sdk-go/pkg/fab/ca"
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

	stateStorePath := "/tmp/enroll_user"
	client.SetCryptoSuite(cryptoSuiteProvider)
	stateStore, err := identity.NewCertFileUserStore(stateStorePath, cryptoSuiteProvider)
	if err != nil {
		t.Fatalf("CreateNewFileKeyValueStore return error[%s]", err)
	}
	client.SetStateStore(stateStore)

	caClient, err := fabricCAClient.NewFabricCAClient(org1Name, testFabricConfig, cryptoSuiteProvider)
	if err != nil {
		t.Fatalf("NewFabricCAClient return error: %v", err)
	}

	// Register a random user
	userName := createRandomName()
	registerRequest := fab.RegistrationRequest{
		Name:        userName,
		Type:        "user",
		Affiliation: "org1.department1",
		CAName:      caConfig.CAName,
	}
	enrolmentSecret, err := caClient.Register(&registerRequest)
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
	enrolleduser := identity.NewUser(mspID, userName)
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

	revokeRequest := fab.RevocationRequest{Name: userName, CAName: "ca.org1.example.com"}
	_, err = caClient.Revoke(&revokeRequest)
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
	networkPeer := &core.NetworkPeer{PeerConfig: peers[0], MspID: mspID}
	testPeer, err := peer.New(testFabricConfig, peer.FromPeerConfig(networkPeer))
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

	myUser := identity.NewUser(mspID, "myUser")
	myUser.SetEnrollmentCertificate(cert)
	myUser.SetPrivateKey(key)

	testClient := client.NewClient(testFabricConfig)
	testClient.SetCryptoSuite(cryptoSuiteProvider)
	testClient.SetIdentityContext(myUser)
	testClient.SetSigningManager(signingManager)

	_, err = testClient.QueryChannels(testPeer)
	if err != nil {
		t.Fatalf("Failed to query with enrolled user : %s", err)
	}
}

func createRandomName() string {
	return "user" + strconv.Itoa(rand.Intn(500000))
}
