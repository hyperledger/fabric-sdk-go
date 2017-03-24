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

	//Client tests: getUserContext successful nill user
	user, err := client.GetUserContext("")
	if err != nil {
		t.Fatalf("client.GetUserContext return error[%s]", err)
	}
	if user != nil {
		t.Fatalf("client.GetUserContext should return nil user")
	}

	//Client tests: Should return error "user is nil"
	err = client.SetUserContext(nil, false)
	if err == nil {
		t.Fatalf("client.SetUserContext didn't return error")
	}
	if err.Error() != "user is nil" {
		t.Fatalf("client.SetUserContext didn't return right error")
	}

	//Client tests: getUserContext with no context in memory or persisted returns nil
	user, err = client.GetUserContext("someUser")
	if err != nil {
		t.Fatalf("client.GetUserContext return error[%s]", err)
	}
	if user != nil {
		t.Fatalf("client.GetUserContext should return nil user")
	}

	//Client tests: successfully setUserContext with skipPersistence true
	user = NewUser("someUser")
	err = client.SetUserContext(user, true)
	if err != nil {
		t.Fatalf("client.SetUserContext return error[%s]", err)
	}
	user, err = client.GetUserContext("someUser")
	if err != nil {
		t.Fatalf("client.GetUserContext return error[%s]", err)
	}
	if user == nil {
		t.Fatalf("client.GetUserContext return nil user")
	}
	if user.GetName() != "someUser" {
		t.Fatalf("client.GetUserContext didn't return the right user")
	}

	//Client tests: Should throw "stateStore is nil"
	err = client.SetUserContext(user, false)
	if err == nil {
		t.Fatalf("client.SetUserContext didn't return error")
	}
	if err.Error() != "stateStore is nil" {
		t.Fatalf("client.SetUserContext didn't return right error")
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
