/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package fabricclient

import (
	"bytes"
	"path"
	"testing"

	"github.com/pkg/errors"

	"github.com/hyperledger/fabric-sdk-go/pkg/cryptosuite/bccsp/sw"
	kvs "github.com/hyperledger/fabric-sdk-go/pkg/fabric-client/keyvaluestore"
	mocks "github.com/hyperledger/fabric-sdk-go/pkg/fabric-client/mocks"
)

var testMsp = "testMsp"

func TestClientMethods(t *testing.T) {
	client := NewClient(mocks.NewMockConfig())
	if client.CryptoSuite() != nil {
		t.Fatalf("Client CryptoSuite should initially be nil")
	}

	s, err := sw.GetSuiteWithDefaultEphemeral()
	if err != nil {
		t.Fatalf("Failed getting ephemeral software-based BCCSP [%s]", err)
	}

	client.SetCryptoSuite(s)
	if client.CryptoSuite() == nil {
		t.Fatalf("Client CryptoSuite should not be nil after setCryptoSuite")
	}

	//Client tests: LoadUserFromStateStore successful nill user
	user, err := client.LoadUserFromStateStore("")
	if err != nil {
		t.Fatalf("client.LoadUserFromStateStore return error[%s]", err)
	}
	if user != nil {
		t.Fatalf("client.LoadUserFromStateStore should return nil user")
	}

	//Client tests: Should return error "user required"
	err = client.SaveUserToStateStore(nil)
	if err == nil {
		t.Fatalf("client.SaveUserToStateStore didn't return error")
	}
	if err.Error() != "user required" {
		t.Fatalf("client.SaveUserToStateStore didn't return right error")
	}

	//Client tests: LoadUserFromStateStore with no context in memory or persisted returns nil
	user, err = client.LoadUserFromStateStore("someUser")
	if err != nil {
		t.Fatalf("client.LoadUserFromStateStore return error[%s]", err)
	}
	if user != nil {
		t.Fatalf("client.LoadUserFromStateStore should return nil user")
	}

	//Client tests: Should throw "stateStore is nil"
	client.SetStateStore(nil)
	err = client.SaveUserToStateStore(mocks.NewMockUser("hello"))
	if err == nil {
		t.Fatalf("client.SaveUserToStateStore didn't return error")
	}
	if err.Error() != "stateStore is nil" {
		t.Fatalf("client.SaveUserToStateStore didn't return right error: %v", err)
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

	stateStorePath := "/tmp/keyvaluestore"
	stateStore, err := kvs.NewFileKeyValueStore(&kvs.FileKeyValueStoreOptions{
		Path: stateStorePath,
		KeySerializer: func(key interface{}) (string, error) {
			keyString, ok := key.(string)
			if !ok {
				return "", errors.New("converting key to string failed")
			}
			return path.Join(stateStorePath, keyString+".json"), nil
		},
	})
	if err != nil {
		t.Fatalf("CreateNewFileKeyValueStore return error[%s]", err)
	}
	client.SetStateStore(stateStore)
	client.StateStore().Store("testvalue", []byte("data"))
	value, err := client.StateStore().Load("testvalue")
	if err != nil {
		t.Fatalf("client.StateStore().Load() return error[%s]", err)
	}
	valueBytes, ok := value.([]byte)
	if !ok {
		t.Fatalf("client.StateStore().Load() returned wrong data type")
	}
	if string(valueBytes) != "data" {
		t.Fatalf("client.StateStore().GetValue() didn't return the right value")
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
