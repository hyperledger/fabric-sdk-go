/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package integration

import (
	"fmt"
	"strconv"
	"testing"
	"time"

	"github.com/hyperledger/fabric-sdk-go/fabric-client/util"
)

func TestChainCodeInvoke(t *testing.T) {

	testSetup := BaseSetupImpl{
		ConfigFile:      "../fixtures/config/config_test.yaml",
		ChainID:         "mychannel",
		ChannelConfig:   "../fixtures/channel/mychannel.tx",
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
	done, rce := util.RegisterCCEvent(testSetup.ChainCodeID, eventID, testSetup.EventHub)

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
