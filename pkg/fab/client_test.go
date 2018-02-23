/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package fabricclient

import (
	"bytes"
	"os"
	"path"
	"testing"

	"github.com/golang/mock/gomock"
	fabricCaUtil "github.com/hyperledger/fabric-sdk-go/internal/github.com/hyperledger/fabric-ca/util"
	"github.com/hyperledger/fabric-sdk-go/pkg/context/api"
	"github.com/hyperledger/fabric-sdk-go/pkg/context/api/core"
	"github.com/hyperledger/fabric-sdk-go/pkg/context/api/core/mocks"
	"github.com/hyperledger/fabric-sdk-go/pkg/core/cryptosuite/bccsp/sw"
	"github.com/hyperledger/fabric-sdk-go/pkg/fab/identity"
	mocks "github.com/hyperledger/fabric-sdk-go/pkg/fab/mocks"
	"github.com/pkg/errors"
)

var testMsp = "testMsp"

var storePathRoot = "/tmp/testcertfileuserstore"
var storePath = path.Join(storePathRoot, "-certs")

var testPrivKey1 = `-----BEGIN PRIVATE KEY-----
MIGHAgEAMBMGByqGSM49AgEGCCqGSM49AwEHBG0wawIBAQQgp4qKKB0WCEfx7XiB
5Ul+GpjM1P5rqc6RhjD5OkTgl5OhRANCAATyFT0voXX7cA4PPtNstWleaTpwjvbS
J3+tMGTG67f+TdCfDxWYMpQYxLlE8VkbEzKWDwCYvDZRMKCQfv2ErNvb
-----END PRIVATE KEY-----`

var testCert1 = `-----BEGIN CERTIFICATE-----
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
-----END CERTIFICATE-----`

func crypto(t *testing.T) core.CryptoSuite {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	mockConfig := mock_core.NewMockConfig(mockCtrl)
	mockConfig.EXPECT().SecurityProvider().Return("SW")
	mockConfig.EXPECT().SecurityAlgorithm().Return("SHA2")
	mockConfig.EXPECT().SecurityLevel().Return(256)
	mockConfig.EXPECT().KeyStorePath().Return(path.Join(storePathRoot, "-keys"))
	mockConfig.EXPECT().Ephemeral().Return(false)

	//Get cryptosuite using config
	c, err := sw.GetSuiteByConfig(mockConfig)
	if err != nil {
		t.Fatalf("Not supposed to get error, but got: %v", err)
	}
	return c
}

func cleanup(storePath string) error {
	err := os.RemoveAll(storePath)
	if err != nil {
		return errors.Wrapf(err, "Cleaning up directory '%s' failed", storePath)
	}
	return nil
}

func TestClientMethods(t *testing.T) {

	cleanup(storePathRoot)
	defer cleanup(storePathRoot)

	crypto := crypto(t)
	_, err := fabricCaUtil.ImportBCCSPKeyFromPEMBytes([]byte(testPrivKey1), crypto, false)
	if err != nil {
		t.Fatalf("ImportBCCSPKeyFromPEMBytes failed [%s]", err)
	}

	client := NewClient(mocks.NewMockConfig())
	if client.CryptoSuite() != nil {
		t.Fatalf("Client CryptoSuite should initially be nil")
	}

	client.SetCryptoSuite(crypto)
	if client.CryptoSuite() == nil {
		t.Fatalf("Client CryptoSuite should not be nil after setCryptoSuite")
	}

	// Load nil user with no MspID
	_, err = client.LoadUserFromStateStore("", "nonexistant")
	if err == nil {
		t.Fatalf("should return error for nil MspID")
	}

	// Load nil user with no Name
	_, err = client.LoadUserFromStateStore("nonexistant", "")
	if err == nil {
		t.Fatalf("should return error for nil Name")
	}

	// Save nil user
	err = client.SaveUserToStateStore(nil)
	if err == nil {
		t.Fatalf("should return error for nil user")
	}

	// nil state store
	client.SetStateStore(nil)
	err = client.SaveUserToStateStore(mocks.NewMockUser("hello"))
	if err == nil {
		t.Fatalf("should return error for nil state store")
	}

	//Client tests: Create new chain
	chain, err := client.NewChannel("someChain")
	if err != nil {
		t.Fatalf("client.NewChain return error[%s]", err)
	}
	if chain.Name() != "someChain" {
		t.Fatalf("client.NewChain create wrong chain")
	}
	chain1 := client.Channel("someChain")
	if chain1.Name() != "someChain" {
		t.Fatalf("client.NewChain create wrong chain")
	}

	stateStore, err := identity.NewCertFileUserStore(storePath, crypto)
	if err != nil {
		t.Fatalf("CreateNewFileKeyValueStore return error[%s]", err)
	}
	client.SetStateStore(stateStore)

	// Load unknown user
	_, err = client.LoadUserFromStateStore("nonexistant", "nonexistant")
	if err != api.ErrUserNotFound {
		t.Fatalf("should return ErrUserNotFound, got: %v", err)
	}

	saveUser := identity.NewUser("myname", "mymsp")
	saveUser.SetEnrollmentCertificate([]byte(testCert1))
	client.StateStore().Store(saveUser)
	retrievedUser, err := client.StateStore().Load(api.UserKey{MspID: saveUser.MspID(), Name: saveUser.Name()})
	if err != nil {
		t.Fatalf("client.StateStore().Load() return error[%s]", err)
	}
	if retrievedUser.MspID() != saveUser.MspID() {
		t.Fatalf("MspID doesn't match")
	}
	if retrievedUser.Name() != saveUser.Name() {
		t.Fatalf("Name doesn't match")
	}
	if string(retrievedUser.EnrollmentCertificate()) != string(saveUser.EnrollmentCertificate()) {
		t.Fatalf("EnrollmentCertificate doesn't match")
	}

	// Set and use siging manager
	client.SetSigningManager(mocks.NewMockSigningManager())

	greeting := []byte("Hello")
	signedObj, err := client.SigningManager().Sign(greeting, nil)
	if err != nil {
		t.Fatalf("Failed to sign object.")
	}

	if !bytes.Equal(signedObj, greeting) {
		t.Fatalf("Expecting Hello, got %s", signedObj)
	}
}

func TestQueryMethodsOnClient(t *testing.T) {
	client := NewClient(mocks.NewMockConfig())

	_, err := client.QueryChannels(nil)
	if err == nil {
		t.Fatalf("QueryChanels: peer cannot be nil")
	}

	_, err = client.QueryInstalledChaincodes(nil)
	if err == nil {
		t.Fatalf("QueryInstalledChaincodes: peer cannot be nil")
	}

}
