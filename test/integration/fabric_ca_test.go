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

package integration

import (
	"crypto/x509"
	"encoding/pem"
	"testing"

	config "github.com/hyperledger/fabric-sdk-go/config"
	fabric_client "github.com/hyperledger/fabric-sdk-go/fabric-client"
	kvs "github.com/hyperledger/fabric-sdk-go/fabric-client/keyvaluestore"
	"github.com/hyperledger/fabric/bccsp"
	bccspFactory "github.com/hyperledger/fabric/bccsp/factory"

	fabric_ca_client "github.com/hyperledger/fabric-sdk-go/fabric-ca-client"
)

// this test uses the FabricCAServices to enroll a user, and
// saves the enrollment materials into a key value store.
// then uses the Client class to load the member from the
// key value store
func TestEnroll(t *testing.T) {
	InitConfigForFabricCA()
	client := fabric_client.NewClient()

	err := bccspFactory.InitFactories(&bccspFactory.FactoryOpts{
		ProviderName: "SW",
		SwOpts: &bccspFactory.SwOpts{
			HashFamily: config.GetSecurityAlgorithm(),
			SecLevel:   config.GetSecurityLevel(),
			FileKeystore: &bccspFactory.FileKeystoreOpts{
				KeyStorePath: config.GetKeyStorePath(),
			},
			Ephemeral: false,
		},
	})
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

	fabricCAClient, err := fabric_ca_client.NewFabricCAClient()
	if err != nil {
		t.Fatalf("NewFabricCAClient return error: %v", err)
	}
	key, cert, err := fabricCAClient.Enroll("testUser2", "user2")
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
	if cert509.Subject.CommonName != "testUser2" {
		t.Fatalf("CommonName in x509 cert is not the enrollmentID")
	}

	keyPem, _ := pem.Decode(key)
	if err != nil {
		t.Fatalf("pem Decode return error: %v", err)
	}
	user := fabric_client.NewUser("testUser2")
	k, err := client.GetCryptoSuite().KeyImport(keyPem.Bytes, &bccsp.ECDSAPrivateKeyImportOpts{Temporary: false})
	if err != nil {
		t.Fatalf("KeyImport return error: %v", err)
	}
	user.SetPrivateKey(k)
	user.SetEnrollmentCertificate(cert)
	err = client.SetUserContext(user, false)
	if err != nil {
		t.Fatalf("client.SetUserContext return error: %v", err)
	}
	user, err = client.GetUserContext("testUser2")
	if err != nil {
		t.Fatalf("client.GetUserContext return error: %v", err)
	}
	if user == nil {
		t.Fatalf("client.GetUserContext return nil")
	}

}

func InitConfigForFabricCA() {
	config.InitConfig("./test_resources/config/config_test.yaml")
}
