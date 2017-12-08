/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package fabricclient

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"path"
	"testing"
	"time"

	fab "github.com/hyperledger/fabric-sdk-go/api/apifabclient"
	"github.com/hyperledger/fabric-sdk-go/internal/github.com/hyperledger/fabric/bccsp/factory"
	"github.com/hyperledger/fabric-sdk-go/test/metadata"

	"github.com/hyperledger/fabric-sdk-go/pkg/cryptosuite"
	"github.com/hyperledger/fabric-sdk-go/pkg/fabric-client/identity"
	kvs "github.com/hyperledger/fabric-sdk-go/pkg/fabric-client/keyvaluestore"
	mocks "github.com/hyperledger/fabric-sdk-go/pkg/fabric-client/mocks"
)

var testMsp = "testMsp"

func TestClientMethods(t *testing.T) {
	client := NewClient(mocks.NewMockConfig())
	if client.CryptoSuite() != nil {
		t.Fatalf("Client CryptoSuite should initially be nil")
	}
	err := factory.InitFactories(nil)
	if err != nil {
		t.Fatalf("Failed getting ephemeral software-based BCCSP [%s]", err)
	}

	client.SetCryptoSuite(cryptosuite.GetDefault())
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
	err = client.SaveUserToStateStore(nil, false)
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

	//Client tests: successfully SaveUserToStateStore with skipPersistence true
	user = identity.NewUser("someUser", testMsp)
	err = client.SaveUserToStateStore(user, true)
	if err != nil {
		t.Fatalf("client.SaveUserToStateStore return error[%s]", err)
	}
	user, err = client.LoadUserFromStateStore("someUser")
	if err != nil {
		t.Fatalf("client.LoadUserFromStateStore return error[%s]", err)
	}
	if user == nil {
		t.Fatalf("client.LoadUserFromStateStore return nil user")
	}
	if user.Name() != "someUser" {
		t.Fatalf("client.LoadUserFromStateStore didn't return the right user")
	}

	if user.MspID() != testMsp {
		t.Fatalf("client.LoadUserFromStateStore didn't return the right msp")
	}

	//Client tests: Should throw "stateStore is nil"
	err = client.SaveUserToStateStore(user, false)
	if err == nil {
		t.Fatalf("client.SaveUserToStateStore didn't return error")
	}
	if err.Error() != "stateStore is nil" {
		t.Fatalf("client.SaveUserToStateStore didn't return right error")
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

	stateStore, err := kvs.CreateNewFileKeyValueStore("/tmp/keyvaluestore")
	if err != nil {
		t.Fatalf("CreateNewFileKeyValueStore return error[%s]", err)
	}
	client.SetStateStore(stateStore)
	client.StateStore().SetValue("testvalue", []byte("data"))
	value, err := client.StateStore().Value("testvalue")
	if err != nil {
		t.Fatalf("client.StateStore().GetValue() return error[%s]", err)
	}
	if string(value) != "data" {
		t.Fatalf("client.StateStore().GetValue() didn't return the right value")
	}

	// Set and use siging manager
	client.SetSigningManager(mocks.NewMockSigningManager())

	greeting := []byte("Hello")
	signedObj, err := client.SigningManager().Sign(greeting, user.PrivateKey())
	if err != nil {
		t.Fatalf("Failed to sign object.")
	}

	if !bytes.Equal(signedObj, greeting) {
		t.Fatalf("Expecting Hello, got %s", signedObj)
	}

}

func TestCreateChannel(t *testing.T) {
	client := NewClient(mocks.NewMockConfig())

	configTx, err := ioutil.ReadFile(path.Join("../../", metadata.ChannelConfigPath, "mychannel.tx"))
	if err != nil {
		t.Fatalf(err.Error())
	}

	// Setup mock orderer
	verifyBroadcast := make(chan *fab.SignedEnvelope)
	orderer := mocks.NewMockOrderer(fmt.Sprintf("0.0.0.0:1234"), verifyBroadcast)

	// Create channel without envelope
	_, err = client.CreateChannel(fab.CreateChannelRequest{
		Orderer: orderer,
		Name:    "mychannel",
	})
	if err == nil {
		t.Fatalf("Expected error creating channel without envelope")
	}

	// Create channel without orderer
	_, err = client.CreateChannel(fab.CreateChannelRequest{
		Envelope: configTx,
		Name:     "mychannel",
	})
	if err == nil {
		t.Fatalf("Expected error creating channel without orderer")
	}

	// Create channel without name
	_, err = client.CreateChannel(fab.CreateChannelRequest{
		Envelope: configTx,
		Orderer:  orderer,
	})
	if err == nil {
		t.Fatalf("Expected error creating channel without name")
	}

	// Test with valid cofiguration
	request := fab.CreateChannelRequest{
		Envelope: configTx,
		Orderer:  orderer,
		Name:     "mychannel",
	}
	_, err = client.CreateChannel(request)
	if err != nil {
		t.Fatalf("Did not expect error from create channel. Got error: %v", err)
	}
	select {
	case b := <-verifyBroadcast:
		logger.Debugf("Verified broadcast: %v", b)
	case <-time.After(time.Second):
		t.Fatalf("Expected broadcast")
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

func TestInterfaces(t *testing.T) {
	var apiClient fab.FabricClient
	var client Client

	apiClient = &client
	if apiClient == nil {
		t.Fatalf("this shouldn't happen.")
	}
}
