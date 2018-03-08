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

	"github.com/hyperledger/fabric-sdk-go/pkg/context/api/fab"
	"github.com/hyperledger/fabric-sdk-go/pkg/core/config"
	packager "github.com/hyperledger/fabric-sdk-go/pkg/fab/ccpackager/gopackager"
	"github.com/hyperledger/fabric-sdk-go/pkg/fab/resource/api"
	"github.com/hyperledger/fabric-sdk-go/pkg/fabsdk"
	"github.com/hyperledger/fabric-sdk-go/test/integration"
	"github.com/hyperledger/fabric-sdk-go/test/metadata"
	"github.com/pkg/errors"
)

const (
	chainCodeName = "install"
	chainCodePath = "github.com/example_cc"
)

func TestChaincodeInstal(t *testing.T) {

	testSetup := &integration.BaseSetupImpl{
		ConfigFile:    "../" + integration.ConfigTestFile,
		ChannelID:     "mychannel",
		OrgID:         org1Name,
		ChannelConfig: path.Join("../../../", metadata.ChannelConfigPath, "mychannel.tx"),
	}

	sdk, err := fabsdk.New(config.FromFile(testSetup.ConfigFile))
	if err != nil {
		t.Fatalf("Failed to create new SDK: %s", err)
	}
	defer sdk.Close()

	if err := testSetup.Initialize(sdk); err != nil {
		t.Fatalf(err.Error())
	}

	testChaincodeInstallUsingChaincodePath(t, sdk, testSetup)

	testChaincodeInstallUsingChaincodePackage(t, sdk, testSetup)
}

// Test chaincode install using chaincodePath to create chaincodePackage
func testChaincodeInstallUsingChaincodePath(t *testing.T, sdk *fabsdk.FabricSDK, testSetup *integration.BaseSetupImpl) {
	chainCodeVersion := getRandomCCVersion()

	ccPkg, err := packager.NewCCPackage(chainCodePath, integration.GetDeployPath())
	if err != nil {
		t.Fatalf("Failed to package chaincode")
	}

	// Low level resource
	client, err := getResource(sdk, "Admin", orgName)
	if err != nil {
		t.Fatalf("Failed to get resource: %s", err)
	}

	if err := installCC(client, chainCodeName, chainCodePath, chainCodeVersion, ccPkg, testSetup.Targets); err != nil {
		t.Fatalf("installCC return error: %v", err)
	}

	chaincodeQueryResponse, err := client.QueryInstalledChaincodes(testSetup.Targets[0])
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
	err = installCC(client, chainCodeName, chainCodePath, chainCodeVersion, ccPkg, testSetup.Targets)
	if err == nil {
		t.Fatalf("install same chaincode didn't return error")
	}
	if strings.Contains(err.Error(), "chaincodes/install.v"+chainCodeVersion+" exists") {
		t.Fatalf("install same chaincode didn't return the correct error")
	}
}

// Test chaincode install using chaincodePackage[byte]
func testChaincodeInstallUsingChaincodePackage(t *testing.T, sdk *fabsdk.FabricSDK, testSetup *integration.BaseSetupImpl) {

	chainCodeVersion := getRandomCCVersion()

	ccPkg, err := packager.NewCCPackage(chainCodePath, integration.GetDeployPath())
	if err != nil {
		t.Fatalf("PackageCC return error: %s", err)
	}

	// Low level resource
	client, err := getResource(sdk, "Admin", orgName)
	if err != nil {
		t.Fatalf("Failed to get resource: %s", err)
	}

	err = installCC(client, "install", "github.com/example_cc_pkg", chainCodeVersion, ccPkg, testSetup.Targets)
	if err != nil {
		t.Fatalf("installCC return error: %v", err)
	}

	//Install same chaincode again, should fail
	err = installCC(client, "install", chainCodePath, chainCodeVersion, ccPkg, testSetup.Targets)
	if err == nil {
		t.Fatalf("install same chaincode didn't return error")
	}
	if strings.Contains(err.Error(), "chaincodes/install.v"+chainCodeVersion+" exists") {
		t.Fatalf("install same chaincode didn't return the correct error")
	}
}

// installCC use low level client to install chaincode
func installCC(client api.Resource, name string, path string, version string, ccPackage *api.CCPackage, targets []fab.ProposalProcessor) error {

	icr := api.InstallChaincodeRequest{Name: name, Path: path, Version: version, Package: ccPackage, Targets: targets}

	_, _, err := client.InstallChaincode(icr)
	if err != nil {
		return errors.WithMessage(err, "InstallChaincode failed")
	}

	return nil
}

func getRandomCCVersion() string {
	return "v0" + strconv.Itoa(rand.Intn(10000000))
}
