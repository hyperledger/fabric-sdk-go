/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package fab

import (
	"math/rand"
	"path"
	"strconv"
	"strings"
	"testing"
	"time"

	packager "github.com/hyperledger/fabric-sdk-go/pkg/fabric-client/ccpackager/gopackager"
	"github.com/hyperledger/fabric-sdk-go/test/integration"
	"github.com/hyperledger/fabric-sdk-go/test/metadata"
)

const (
	chainCodeName = "install"
	chainCodePath = "github.com/example_cc"
)

func TestChaincodeInstal(t *testing.T) {

	testSetup := &integration.BaseSetupImpl{
		ConfigFile:      "../" + integration.ConfigTestFile,
		ChannelID:       "mychannel",
		OrgID:           org1Name,
		ChannelConfig:   path.Join("../../../", metadata.ChannelConfigPath, "mychannel.tx"),
		ConnectEventHub: true,
	}

	if err := testSetup.Initialize(t); err != nil {
		t.Fatalf(err.Error())
	}

	testChaincodeInstallUsingChaincodePath(t, testSetup)

	testChaincodeInstallUsingChaincodePackage(t, testSetup)
}

// Test chaincode install using chaincodePath to create chaincodePackage
func testChaincodeInstallUsingChaincodePath(t *testing.T, testSetup *integration.BaseSetupImpl) {
	chainCodeVersion := getRandomCCVersion()

	// Install and Instantiate Events CC
	// Retrieve installed chaincodes
	client := testSetup.Client

	ccPkg, err := packager.NewCCPackage(chainCodePath, testSetup.GetDeployPath())
	if err != nil {
		t.Fatalf("Failed to package chaincode")
	}

	if err := testSetup.InstallCC(chainCodeName, chainCodePath, chainCodeVersion, ccPkg); err != nil {
		t.Fatalf("installCC return error: %v", err)
	}

	chaincodeQueryResponse, err := client.QueryInstalledChaincodes(testSetup.Channel.PrimaryPeer())
	if err != nil {
		t.Fatalf("QueryInstalledChaincodes return error: %v", err)
	}
	ccFound := false
	for _, chaincode := range chaincodeQueryResponse.Chaincodes {
		if chaincode.Name == chainCodeName && chaincode.Path == chainCodePath && chaincode.Version == chainCodeVersion {
			t.Logf("Found chaincode: %s", chaincode)
			ccFound = true
		}
	}

	if !ccFound {
		t.Fatalf("Failed to retrieve installed chaincode.")
	}
	//Install same chaincode again, should fail
	err = testSetup.InstallCC(chainCodeName, chainCodePath, chainCodeVersion, ccPkg)
	if err == nil {
		t.Fatalf("install same chaincode didn't return error")
	}
	if strings.Contains(err.Error(), "chaincodes/install.v"+chainCodeVersion+" exists") {
		t.Fatalf("install same chaincode didn't return the correct error")
	}
}

// Test chaincode install using chaincodePackage[byte]
func testChaincodeInstallUsingChaincodePackage(t *testing.T, testSetup *integration.BaseSetupImpl) {

	chainCodeVersion := getRandomCCVersion()

	ccPkg, err := packager.NewCCPackage(chainCodePath, testSetup.GetDeployPath())
	if err != nil {
		t.Fatalf("PackageCC return error: %s", err)
	}

	err = testSetup.InstallCC("install", "github.com/example_cc_pkg", chainCodeVersion, ccPkg)
	if err != nil {
		t.Fatalf("installCC return error: %v", err)
	}

	//Install same chaincode again, should fail
	err = testSetup.InstallCC("install", chainCodePath, chainCodeVersion, ccPkg)
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
