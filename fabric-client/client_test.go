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

package fabricclient

import (
	"fmt"
	"io/ioutil"
	"testing"

	kvs "github.com/hyperledger/fabric-sdk-go/fabric-client/keyvaluestore"

	bccspFactory "github.com/hyperledger/fabric/bccsp/factory"
)

func TestClientMethods(t *testing.T) {
	client := NewClient()
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
	user = NewUser("someUser")
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

	//Client tests: Should throw "stateStore is nil"
	err = client.SaveUserToStateStore(user, false)
	if err == nil {
		t.Fatalf("client.SaveUserToStateStore didn't return error")
	}
	if err.Error() != "stateStore is nil" {
		t.Fatalf("client.SaveUserToStateStore didn't return right error")
	}

	//Client tests: Create new chain
	chain, err := client.NewChain("someChain")
	if err != nil {
		t.Fatalf("client.NewChain return error[%s]", err)
	}
	if chain.GetName() != "someChain" {
		t.Fatalf("client.NewChain create wrong chain")
	}
	chain1 := client.GetChain("someChain")
	if chain1.GetName() != "someChain" {
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
	client := NewClient()

	configTx, err := ioutil.ReadFile("../test/fixtures/channel/testchannel.tx")
	if err != nil {
		t.Fatalf(err.Error())
	}
	// Setup mock orderer
	orderer := &mockOrderer{MockURL: fmt.Sprintf("0.0.0.0:1234")}

	// Create channel without envelope
	chain, err := client.CreateChannel(&CreateChannelRequest{
		Orderer: orderer,
		Name:    "testchannel",
	})
	if err == nil {
		t.Fatalf("Expected error creating channel without envelope")
	}

	// Create channel without orderer
	chain, err = client.CreateChannel(&CreateChannelRequest{
		Envelope: configTx,
		Name:     "testchannel",
	})
	if err == nil {
		t.Fatalf("Expected error creating channel without orderer")
	}

	// Create channel without name
	chain, err = client.CreateChannel(&CreateChannelRequest{
		Envelope: configTx,
		Orderer:  orderer,
	})
	if err == nil {
		t.Fatalf("Expected error creating channel without name")
	}

	// Test with valid cofiguration
	chain, err = client.CreateChannel(&CreateChannelRequest{
		Envelope: configTx,
		Orderer:  orderer,
		Name:     "testchannel",
	})
	if err != nil {
		t.Fatalf("Did not expect error from create channel. Got error: %s", err.Error())
	}
	if chain == nil {
		t.Fatalf("Nil chain returned from CreateChannel")
	}
	if chain.GetName() != "testchannel" {
		t.Fatalf("Invalid name %s of chain. Expecting testchannel", chain.GetName())
	}
	mspManager := chain.GetMSPManager()
	if mspManager == nil {
		t.Fatalf("nil MSPManager on new chain")
	}
	msps, err := mspManager.GetMSPs()
	if err != nil || len(msps) == 0 {
		t.Fatalf("At least one MSP expected in MSPManager")
	}
}

func TestQueryMethodsOnClient(t *testing.T) {
	client := NewClient()

	_, err := client.QueryChannels(nil)
	if err == nil {
		t.Fatalf("QueryChanels: peer cannot be nil")
	}

	_, err = client.QueryInstalledChaincodes(nil)
	if err == nil {
		t.Fatalf("QueryInstalledChaincodes: peer cannot be nil")
	}

}
