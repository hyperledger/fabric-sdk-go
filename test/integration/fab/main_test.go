/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/
package fab

import (
	"fmt"
	"os"
	"path"
	"testing"

	"github.com/hyperledger/fabric-sdk-go/pkg/context/api/core"
	"github.com/hyperledger/fabric-sdk-go/pkg/core/config"
	"github.com/hyperledger/fabric-sdk-go/pkg/fabsdk"
	"github.com/hyperledger/fabric-sdk-go/test/integration"
	"github.com/hyperledger/fabric-sdk-go/test/metadata"
)

var testFabricConfig core.Config

func TestMain(m *testing.M) {
	setup()
	r := m.Run()
	teardown()
	os.Exit(r)
}

func setup() {
	// do any test setup for all tests here...
	var err error

	testSetup := integration.BaseSetupImpl{
		ConfigFile: "../" + integration.ConfigTestFile,
	}

	testFabricConfig, err = testSetup.InitConfig()()
	if err != nil {
		fmt.Printf("Failed InitConfig [%s]\n", err)
		os.Exit(1)
	}
}

func teardown() {
	// do any teadown activities here ..
	testFabricConfig = nil
}

func initializeTests(t *testing.T, chainCodeID string) (integration.BaseSetupImpl, *fabsdk.FabricSDK) {
	testSetup := integration.BaseSetupImpl{
		ConfigFile: "../" + integration.ConfigTestFile,

		ChannelID:     "mychannel",
		OrgID:         org1Name,
		ChannelConfig: path.Join("../../", metadata.ChannelConfigPath, "mychannel.tx"),
	}

	sdk, err := fabsdk.New(config.FromFile(testSetup.ConfigFile))
	if err != nil {
		t.Fatalf("Failed to create new SDK: %s", err)
	}

	if err := testSetup.Initialize(sdk); err != nil {
		t.Fatalf(err.Error())
	}

	if err := integration.InstallAndInstantiateCC(sdk, fabsdk.WithUser("Admin"), testSetup.OrgID, chainCodeID, "github.com/events_cc", "v0", integration.GetDeployPath(), nil); err != nil {
		t.Fatalf("InstallAndInstantiateCC return error: %v", err)
	}

	return testSetup, sdk
}
