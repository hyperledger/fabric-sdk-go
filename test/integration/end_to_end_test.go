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
	"time"

	fcUtil "github.com/hyperledger/fabric-sdk-go/fabric-client/helpers"
)

func TestChainCodeInvoke(t *testing.T) {

	testSetup := BaseSetupImpl{
		ConfigFile:      "../fixtures/config/config_test.yaml",
		ChainID:         "testchannel",
		ChannelConfig:   "../fixtures/channel/testchannel.tx",
		ConnectEventHub: true,
	}

	if err := testSetup.Initialize(); err != nil {
		t.Fatalf(err.Error())
	}

	if err := testSetup.InstallAndInstantiateExampleCC(); err != nil {
		t.Fatalf("InstallAndInstantiateExampleCC return error: %v", err)
	}

	// Get Query value before invoke
	value, err := testSetup.QueryAsset()
	if err != nil {
		t.Fatalf("getQueryValue return error: %v", err)
	}
	fmt.Printf("*** QueryValue before invoke %s\n", value)

	eventID := "test([a-zA-Z]+)"

	// Register callback for chaincode event
	done, rce := fcUtil.RegisterCCEvent(testSetup.ChainCodeID, eventID, testSetup.EventHub)

	_, err = testSetup.MoveFunds()
	if err != nil {
		t.Fatalf("Move funds return error: %v", err)
	}

	select {
	case <-done:
	case <-time.After(time.Second * 20):
		t.Fatalf("Did NOT receive CC for eventId(%s)\n", eventID)
	}

	testSetup.EventHub.UnregisterChaincodeEvent(rce)

	valueAfterInvoke, err := testSetup.QueryAsset()
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
