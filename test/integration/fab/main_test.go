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

	config "github.com/hyperledger/fabric-sdk-go/api/apiconfig"
	"github.com/hyperledger/fabric-sdk-go/pkg/fabsdk"
	"github.com/hyperledger/fabric-sdk-go/test/integration"
	"github.com/hyperledger/fabric-sdk-go/test/metadata"
)

var testFabricConfig config.Config

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

func initializeTests(t *testing.T, chainCodeID string) integration.BaseSetupImpl {
	testSetup := integration.BaseSetupImpl{
		ConfigFile: "../" + integration.ConfigTestFile,

		ChannelID:       "mychannel",
		OrgID:           org1Name,
		ChannelConfig:   path.Join("../../", metadata.ChannelConfigPath, "mychannel.tx"),
		ConnectEventHub: true,
	}

	if err := testSetup.Initialize(); err != nil {
		t.Fatalf(err.Error())
	}

	if err := integration.InstallAndInstantiateCC(testSetup.SDK, fabsdk.WithUser("Admin"), testSetup.OrgID, chainCodeID, "github.com/events_cc", "v0", integration.GetDeployPath(), nil); err != nil {
		t.Fatalf("InstallAndInstantiateCC return error: %v", err)
	}

	return testSetup
}
