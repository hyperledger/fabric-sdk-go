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
)

//
// Orderer via chain setOrderer/getOrderer
//
// Set the orderer URL through the chain setOrderer method. Verify that the
// orderer URL was set correctly through the getOrderer method. Repeat the
// process by updating the orderer URL to a different address.
//
func TestOrdererViaChain(t *testing.T) {
	client := NewClient()
	chain, err := client.NewChain("testChain-orderer-member")
	if err != nil {
		t.Fatalf("error from NewChain %v", err)
	}
	orderer, _ := CreateNewOrderer("localhost:7050", "", "")
	chain.AddOrderer(orderer)

	orderers := chain.GetOrderers()
	if orderers == nil || len(orderers) != 1 || orderers[0].GetURL() != "localhost:7050" {
		t.Fatalf("Failed to retieve the new orderer URL from the chain")
	}
	chain.RemoveOrderer(orderer)
	orderer2, err := CreateNewOrderer("localhost:7054", "", "")
	if err != nil {
		t.Fatalf("Failed to CreateNewOrderer error(%v)", err)
	}
	chain.AddOrderer(orderer2)
	orderers = chain.GetOrderers()

	if orderers == nil || len(orderers) != 1 || orderers[0].GetURL() != "localhost:7054" {
		t.Fatalf("Failed to retieve the new orderer URL from the chain")
	}

}

//
// Orderer via chain missing orderer
//
// Attempt to send a request to the orderer with the sendTransaction method
// before the orderer URL was set. Verify that an error is reported when tying
// to send the request.
//
func TestPeerViaChainMissingOrderer(t *testing.T) {
	client := NewClient()
	chain, err := client.NewChain("testChain-orderer-member2")
	if err != nil {
		t.Fatalf("error from NewChain %v", err)
	}
	_, err = chain.SendTransaction(nil)
	if err == nil {
		t.Fatalf("SendTransaction didn't return error")
	}
	if err.Error() != "orderers is nil" {
		t.Fatalf("SendTransaction didn't return right error")
	}

}

//
// Orderer via chain nil data
//
// Attempt to send a request to the orderer with the sendTransaction method
// with the data set to null. Verify that an error is reported when tying
// to send null data.
//
func TestOrdererViaChainNilData(t *testing.T) {
	client := NewClient()
	chain, err := client.NewChain("testChain-orderer-member2")
	if err != nil {
		t.Fatalf("error from NewChain %v", err)
	}
	orderer, err := CreateNewOrderer("localhost:7050", "", "")
	if err != nil {
		t.Fatalf("Failed to CreateNewOrderer error(%v)", err)
	}
	chain.AddOrderer(orderer)
	_, err = chain.SendTransaction(nil)
	if err == nil {
		t.Fatalf("SendTransaction didn't return error")
	}
	if err.Error() != "proposal is nil" {
		t.Fatalf("SendTransaction didn't return right error")
	}
}
