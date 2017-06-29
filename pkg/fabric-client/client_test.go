/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package fabricclient

import (
	"fmt"
	"io/ioutil"
	"testing"
	"time"

	api "github.com/hyperledger/fabric-sdk-go/api"
	mocks "github.com/hyperledger/fabric-sdk-go/pkg/fabric-client/mocks"
	fcUser "github.com/hyperledger/fabric-sdk-go/pkg/fabric-client/user"

	kvs "github.com/hyperledger/fabric-sdk-go/pkg/fabric-client/keyvaluestore"
	bccspFactory "github.com/hyperledger/fabric/bccsp/factory"
)

var testMsp = "testMsp"

func TestClientMethods(t *testing.T) {
	client := NewClient(mocks.NewMockConfig())
	if client.GetCryptoSuite() != nil {
		t.Fatalf("Client getCryptoSuite should initially be nil")
	}
	err := bccspFactory.InitFactories(nil)
	if err != nil {
		t.Fatalf("Failed getting ephemeral software-based BCCSP [%s]", err)
	}
	cryptoSuite := bccspFactory.GetDefault()

	client.SetCryptoSuite(cryptoSuite)
	if client.GetCryptoSuite() == nil {
		t.Fatalf("Client getCryptoSuite should not be nil after setCryptoSuite")
	}

	//Client tests: LoadUserFromStateStore successful nill user
	user, err := client.LoadUserFromStateStore("")
	if err != nil {
		t.Fatalf("client.LoadUserFromStateStore return error[%s]", err)
	}
	if user != nil {
		t.Fatalf("client.LoadUserFromStateStore should return nil user")
	}

	//Client tests: Should return error "user is nil"
	err = client.SaveUserToStateStore(nil, false)
	if err == nil {
		t.Fatalf("client.SaveUserToStateStore didn't return error")
	}
	if err.Error() != "user is nil" {
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
	user = fcUser.NewUser("someUser", testMsp)
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
	if user.GetName() != "someUser" {
		t.Fatalf("client.LoadUserFromStateStore didn't return the right user")
	}

	if user.GetMspID() != testMsp {
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
	chain1 := client.GetChannel("someChain")
	if chain1.Name() != "someChain" {
		t.Fatalf("client.NewChain create wrong chain")
	}

	stateStore, err := kvs.CreateNewFileKeyValueStore("/tmp/keyvaluestore")
	if err != nil {
		t.Fatalf("CreateNewFileKeyValueStore return error[%s]", err)
	}
	client.SetStateStore(stateStore)
	client.GetStateStore().SetValue("testvalue", []byte("data"))
	value, err := client.GetStateStore().GetValue("testvalue")
	if err != nil {
		t.Fatalf("client.GetStateStore().GetValue() return error[%s]", err)
	}
	if string(value) != "data" {
		t.Fatalf("client.GetStateStore().GetValue() didn't return the right value")
	}

}

func TestCreateChannel(t *testing.T) {
	client := NewClient(mocks.NewMockConfig())

	configTx, err := ioutil.ReadFile("../../test/fixtures/channel/mychannel.tx")
	if err != nil {
		t.Fatalf(err.Error())
	}

	// Setup mock orderer
	verifyBroadcast := make(chan *api.SignedEnvelope)
	orderer := mocks.NewMockOrderer(fmt.Sprintf("0.0.0.0:1234"), verifyBroadcast)

	// Create channel without envelope
	err = client.CreateChannel(&api.CreateChannelRequest{
		Orderer: orderer,
		Name:    "mychannel",
	})
	if err == nil {
		t.Fatalf("Expected error creating channel without envelope")
	}

	// Create channel without orderer
	err = client.CreateChannel(&api.CreateChannelRequest{
		Envelope: configTx,
		Name:     "mychannel",
	})
	if err == nil {
		t.Fatalf("Expected error creating channel without orderer")
	}

	// Create channel without name
	err = client.CreateChannel(&api.CreateChannelRequest{
		Envelope: configTx,
		Orderer:  orderer,
	})
	if err == nil {
		t.Fatalf("Expected error creating channel without name")
	}

	// Test with valid cofiguration
	request := &api.CreateChannelRequest{
		Envelope: configTx,
		Orderer:  orderer,
		Name:     "mychannel",
	}
	err = client.CreateChannel(request)
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
