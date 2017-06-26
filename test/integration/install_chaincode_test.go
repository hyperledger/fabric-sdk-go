/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package integration

import (
	"fmt"
	"math/rand"
	"strconv"
	"strings"
	"testing"
	"time"

	packager "github.com/hyperledger/fabric-sdk-go/pkg/fabric-client/packager"
	"github.com/hyperledger/fabric-sdk-go/pkg/util"
)

const (
	chainCodeName = "install"
	chainCodePath = "github.com/example_cc"
)

func TestChaincodeInstal(t *testing.T) {

	testSetup := &BaseSetupImpl{
		ConfigFile:      "../fixtures/config/config_test.yaml",
		ChannelID:       "mychannel",
		ChannelConfig:   "../fixtures/channel/mychannel.tx",
		ConnectEventHub: true,
	}

	if err := testSetup.Initialize(); err != nil {
		t.Fatalf(err.Error())
	}

	testChaincodeInstallUsingChaincodePath(t, testSetup)

	testChaincodeInstallUsingChaincodePackage(t, testSetup)
}

// Test chaincode install using chaincodePath to create chaincodePackage
func testChaincodeInstallUsingChaincodePath(t *testing.T, testSetup *BaseSetupImpl) {
	chainCodeVersion := getRandomCCVersion()

	// Install and Instantiate Events CC
	// Retrieve installed chaincodes
	client := testSetup.Client

	if err := testSetup.InstallCC(chainCodeName, chainCodePath, chainCodeVersion, nil); err != nil {
		t.Fatalf("installCC return error: %v", err)
	}

	// set Client User Context to Admin
	testSetup.Client.SetUserContext(testSetup.AdminUser)
	defer testSetup.Client.SetUserContext(testSetup.NormalUser)
	chaincodeQueryResponse, err := client.QueryInstalledChaincodes(testSetup.Channel.GetPrimaryPeer())
	if err != nil {
		t.Fatalf("QueryInstalledChaincodes return error: %v", err)
	}
	ccFound := false
	for _, chaincode := range chaincodeQueryResponse.Chaincodes {
		if chaincode.Name == chainCodeName && chaincode.Path == chainCodePath && chaincode.Version == chainCodeVersion {
			fmt.Printf("Found chaincode: %s\n", chaincode)
			ccFound = true
		}
	}

	if !ccFound {
		t.Fatalf("Failed to retrieve installed chaincode.")
	}
	//Install same chaincode again, should fail
	err = testSetup.InstallCC(chainCodeName, chainCodePath, chainCodeVersion, nil)
	if err == nil {
		t.Fatalf("install same chaincode didn't return error")
	}
	if strings.Contains(err.Error(), "chaincodes/install.v"+chainCodeVersion+" exists") {
		t.Fatalf("install same chaincode didn't return the correct error")
	}
}

// Test chaincode install using chaincodePackage[byte]
func testChaincodeInstallUsingChaincodePackage(t *testing.T, testSetup *BaseSetupImpl) {

	chainCodeVersion := getRandomCCVersion()
	util.ChangeGOPATHToDeploy(testSetup.GetDeployPath())
	chaincodePackage, err := packager.PackageCC(chainCodePath, "")
	util.ResetGOPATH()
	if err != nil {
		t.Fatalf("PackageCC return error: %s", err)
	}

	err = testSetup.InstallCC("install", "github.com/example_cc_pkg", chainCodeVersion, chaincodePackage)
	if err != nil {
		t.Fatalf("installCC return error: %v", err)
	}
	//Install same chaincode again, should fail
	err = testSetup.InstallCC("install", chainCodePath, chainCodeVersion, chaincodePackage)
	if err == nil {
		t.Fatalf("install same chaincode didn't return error")
	}
	if strings.Contains(err.Error(), "chaincodes/install.v"+chainCodeVersion+" exists") {
		t.Fatalf("install same chaincode didn't return the correct error")
	}
}

func getRandomCCVersion() string {
	rand.Seed(time.Now().UnixNano())
	return "v0" + strconv.Itoa(rand.Intn(10000000))
}
