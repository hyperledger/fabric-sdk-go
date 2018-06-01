/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/
package sdk

import (
	"fmt"
	"os"
	"path"
	"testing"

	"github.com/hyperledger/fabric-sdk-go/pkg/fabsdk"
	"github.com/hyperledger/fabric-sdk-go/test/integration"
	"github.com/hyperledger/fabric-sdk-go/test/metadata"
)

var mainSDK *fabsdk.FabricSDK
var mainTestSetup *integration.BaseSetupImpl
var mainChaincodeID string

func TestMain(m *testing.M) {
	setup()
	r := m.Run()
	teardown()
	os.Exit(r)
}

func setup() {
	testSetup := integration.BaseSetupImpl{
		ChannelID:         "mychannel",
		OrgID:             org1Name,
		ChannelConfigFile: path.Join("../../../", metadata.ChannelConfigPath, "mychannel.tx"),
	}

	sdk, err := fabsdk.New(integration.ConfigBackend)
	if err != nil {
		panic(fmt.Sprintf("Failed to create new SDK: %s", err))
	}

	// Delete all private keys from the crypto suite store
	// and users from the user store
	integration.CleanupUserData(nil, sdk)

	if err := testSetup.Initialize(sdk); err != nil {
		panic(err.Error())
	}

	chaincodeID := integration.GenerateRandomID()
	if _, err := integration.InstallAndInstantiateExampleCC(sdk, fabsdk.WithUser("Admin"), testSetup.OrgID, chaincodeID); err != nil {
		panic(fmt.Sprintf("InstallAndInstantiateExampleCC return error: %s", err))
	}

	mainSDK = sdk
	mainTestSetup = &testSetup
	mainChaincodeID = chaincodeID
}

func teardown() {
	integration.CleanupUserData(nil, mainSDK)
	mainSDK.Close()
}
