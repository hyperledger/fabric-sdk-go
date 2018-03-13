/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package fab

import (
	"math/rand"
	"strconv"
	"strings"
	"testing"

	"github.com/hyperledger/fabric-sdk-go/pkg/context"
	"github.com/hyperledger/fabric-sdk-go/pkg/context/api/fab"
	packager "github.com/hyperledger/fabric-sdk-go/pkg/fab/ccpackager/gopackager"
	"github.com/hyperledger/fabric-sdk-go/pkg/fab/resource"
	"github.com/hyperledger/fabric-sdk-go/pkg/fab/resource/api"
	"github.com/hyperledger/fabric-sdk-go/pkg/fabsdk"
	"github.com/hyperledger/fabric-sdk-go/test/integration"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
)

const (
	chainCodeName = "install"
	chainCodePath = "github.com/example_cc"
)

func TestChaincodeInstal(t *testing.T) {

	// Using shared SDK instance to increase test speed.
	sdk := mainSDK
	testSetup := mainTestSetup

	//testSetup := &integration.BaseSetupImpl{
	//	ConfigFile:    "../" + integration.ConfigTestFile,
	//	ChannelID:     "mychannel",
	//	OrgID:         org1Name,
	//	ChannelConfig: path.Join("../../../", metadata.ChannelConfigPath, "mychannel.tx"),
	//}

	//sdk, err := fabsdk.New(config.FromFile(testSetup.ConfigFile))
	//if err != nil {
	//	t.Fatalf("Failed to create new SDK: %s", err)
	//}
	//defer sdk.Close()

	//if err := testSetup.Initialize(sdk); err != nil {
	//	t.Fatalf(err.Error())
	//}

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
	client, err := getContext(sdk, "Admin", orgName)

	if err != nil {
		t.Fatalf("Failed to get resource: %s", err)
	}

	peers, err := getProposalProcessors(sdk, "Admin", testSetup.OrgID, testSetup.Targets)
	assert.Nil(t, err, "creating peers failed")

	if err := installCC(client, chainCodeName, chainCodePath, chainCodeVersion, ccPkg, peers); err != nil {
		t.Fatalf("installCC return error: %v", err)
	}

	chaincodeQueryResponse, err := resource.QueryInstalledChaincodes(client, peers[0])
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
	err = installCC(client, chainCodeName, chainCodePath, chainCodeVersion, ccPkg, peers)
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
	client, err := getContext(sdk, "Admin", orgName)
	if err != nil {
		t.Fatalf("Failed to get resource: %s", err)
	}

	peers, err := getProposalProcessors(sdk, "Admin", testSetup.OrgID, testSetup.Targets)
	assert.Nil(t, err, "creating peers failed")

	err = installCC(client, "install", "github.com/example_cc_pkg", chainCodeVersion, ccPkg, peers)
	if err != nil {
		t.Fatalf("installCC return error: %v", err)
	}

	//Install same chaincode again, should fail
	err = installCC(client, "install", chainCodePath, chainCodeVersion, ccPkg, peers)
	if err == nil {
		t.Fatalf("install same chaincode didn't return error")
	}
	if strings.Contains(err.Error(), "chaincodes/install.v"+chainCodeVersion+" exists") {
		t.Fatalf("install same chaincode didn't return the correct error")
	}
}

// installCC use low level client to install chaincode
func installCC(client *context.Client, name string, path string, version string, ccPackage *api.CCPackage, targets []fab.ProposalProcessor) error {

	icr := api.InstallChaincodeRequest{Name: name, Path: path, Version: version, Package: ccPackage}

	_, _, err := resource.InstallChaincode(client, icr, targets)
	if err != nil {
		return errors.WithMessage(err, "InstallChaincode failed")
	}

	return nil
}

func getRandomCCVersion() string {
	return "v0" + strconv.Itoa(rand.Intn(10000000))
}

func getContext(sdk *fabsdk.FabricSDK, user string, orgName string) (*context.Client, error) {

	ctx := sdk.Context(fabsdk.WithUser(user), fabsdk.WithOrg(orgName))

	clientContext, err := ctx()
	if err != nil {
		return nil, errors.WithMessage(err, "create context failed")
	}

	return &context.Client{Providers: clientContext, Identity: clientContext}, nil

}
