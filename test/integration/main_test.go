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

	testSetup := BaseSetupImpl{
		ConfigFile: "../fixtures/config/config_test.yaml",
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
