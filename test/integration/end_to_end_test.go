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
	"fmt"
	"strconv"
	"testing"
)

func TestChainCodeInvoke(t *testing.T) {
	testSetup := BaseSetupImpl{}

	testSetup.InitConfig()

	eventHub, err := testSetup.GetEventHub(nil)
	if err != nil {
		t.Fatalf("GetEventHub return error: %v", err)
	}
	chain, err := testSetup.GetChain()
	if err != nil {
		t.Fatalf("GetChain return error: %v", err)
	}
	// Create and join channel represented by 'chain'
	testSetup.CreateAndJoinChannel(t, chain)

	err = testSetup.InstallCC(chain, chainCodeID, chainCodePath, chainCodeVersion, nil, nil)
	if err != nil {
		t.Fatalf("installCC return error: %v", err)
	}
	err = testSetup.InstantiateCC(chain, eventHub)
	if err != nil {
		t.Fatalf("instantiateCC return error: %v", err)
	}
	// Get Query value before invoke
	value, err := testSetup.GetQueryValue(chain)
	if err != nil {
		t.Fatalf("getQueryValue return error: %v", err)
	}
	fmt.Printf("*** QueryValue before invoke %s\n", value)

	_, err = testSetup.Invoke(chain, eventHub)
	if err != nil {
		t.Fatalf("invoke return error: %v", err)
	}

	valueAfterInvoke, err := testSetup.GetQueryValue(chain)
	if err != nil {
		t.Errorf("getQueryValue return error: %v", err)
		return
	}
	fmt.Printf("*** QueryValue after invoke %s\n", valueAfterInvoke)

	valueInt, _ := strconv.Atoi(value)
	valueInt = valueInt + 1
	valueAfterInvokeInt, _ := strconv.Atoi(valueAfterInvoke)
	if valueInt != valueAfterInvokeInt {
		t.Fatalf("SendTransaction didn't change the QueryValue")

	}

}
