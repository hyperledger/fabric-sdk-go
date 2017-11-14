/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/
package integration

import (
	"fmt"
	"os"
	"testing"

	config "github.com/hyperledger/fabric-sdk-go/api/apiconfig"
	"github.com/hyperledger/fabric-sdk-go/pkg/logging"
	"github.com/hyperledger/fabric-sdk-go/pkg/logging/deflogger"
)

var testFabricConfig config.Config

func TestMain(m *testing.M) {
	setup()
	r := m.Run()
	teardown()
	os.Exit(r)
}

func setup() {

	//Setup logging provider in advance for tests to make sure none of the test logs are being skipped
	if !logging.IsLoggerInitialized() {
		logging.InitLogger(deflogger.GetLoggingProvider())
	}

	// do any test setup for all tests here...
	var err error

	testSetup := BaseSetupImpl{
		ConfigFile: ConfigTestFile,
	}

	testFabricConfig, err = testSetup.InitConfig()
	if err != nil {
		fmt.Printf("Failed InitConfig [%s]\n", err)
		os.Exit(1)
	}
}

func teardown() {
	// do any teadown activities here ..
	testFabricConfig = nil
}
