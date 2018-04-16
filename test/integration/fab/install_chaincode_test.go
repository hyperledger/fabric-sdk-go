/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package fab

import (
	reqContext "context"
	"math/rand"
	"strconv"
	"strings"
	"testing"

	"github.com/hyperledger/fabric-sdk-go/pkg/common/errors/retry"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/fab"
	"github.com/hyperledger/fabric-sdk-go/pkg/context"
	packager "github.com/hyperledger/fabric-sdk-go/pkg/fab/ccpackager/gopackager"
	"github.com/hyperledger/fabric-sdk-go/pkg/fab/resource"
	"github.com/hyperledger/fabric-sdk-go/pkg/fab/resource/api"
	"github.com/hyperledger/fabric-sdk-go/pkg/fabsdk"
	"github.com/hyperledger/fabric-sdk-go/test/integration"
	"github.com/hyperledger/fabric-sdk-go/third_party/github.com/hyperledger/fabric/protos/peer"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/require"
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
	reqCtx, cancel, err := getContext(sdk, "Admin", orgName)
	if err != nil {
		t.Fatalf("Failed to get resource: %s", err)
	}
	defer cancel()

	peers, err := getProposalProcessors(sdk, "Admin", testSetup.OrgID, testSetup.Targets)
	require.Nil(t, err, "creating peers failed")

	if err := installCC(reqCtx, chainCodeName, chainCodePath, chainCodeVersion, ccPkg, peers); err != nil {
		t.Fatalf("installCC return error: %v", err)
	}

	chaincodeQueryResponse, err := resource.QueryInstalledChaincodes(reqCtx, peers[0], resource.WithRetry(retry.DefaultResMgmtOpts))

	if err != nil {
		t.Fatalf("QueryInstalledChaincodes return error: %v", err)
	}
	retrieveInstalledCC(chaincodeQueryResponse, chainCodeVersion, t)
	//Install same chaincode again, should fail
	err = installCC(reqCtx, chainCodeName, chainCodePath, chainCodeVersion, ccPkg, peers)

	if err == nil {
		t.Fatalf("install same chaincode didn't return error")
	}
	if strings.Contains(err.Error(), "chaincodes/install.v"+chainCodeVersion+" exists") {
		t.Fatalf("install same chaincode didn't return the correct error")
	}
}

func retrieveInstalledCC(chaincodeQueryResponse *peer.ChaincodeQueryResponse, chainCodeVersion string, t *testing.T) {
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
}

// Test chaincode install using chaincodePackage[byte]
func testChaincodeInstallUsingChaincodePackage(t *testing.T, sdk *fabsdk.FabricSDK, testSetup *integration.BaseSetupImpl) {

	chainCodeVersion := getRandomCCVersion()

	ccPkg, err := packager.NewCCPackage(chainCodePath, integration.GetDeployPath())
	if err != nil {
		t.Fatalf("PackageCC return error: %s", err)
	}

	// Low level resource
	reqCtx, cancel, err := getContext(sdk, "Admin", orgName)
	if err != nil {
		t.Fatalf("Failed to get resource: %s", err)
	}
	defer cancel()

	peers, err := getProposalProcessors(sdk, "Admin", testSetup.OrgID, testSetup.Targets)
	require.Nil(t, err, "creating peers failed")

	err = installCC(reqCtx, "install", "github.com/example_cc_pkg", chainCodeVersion, ccPkg, peers)

	if err != nil {
		t.Fatalf("installCC return error: %v", err)
	}

	//Install same chaincode again, should fail
	err = installCC(reqCtx, "install", chainCodePath, chainCodeVersion, ccPkg, peers)

	if err == nil {
		t.Fatalf("install same chaincode didn't return error")
	}
	if strings.Contains(err.Error(), "chaincodes/install.v"+chainCodeVersion+" exists") {
		t.Fatalf("install same chaincode didn't return the correct error")
	}
}

// installCC use low level client to install chaincode
func installCC(reqCtx reqContext.Context, name string, path string, version string, ccPackage *api.CCPackage, targets []fab.ProposalProcessor) error {

	icr := api.InstallChaincodeRequest{Name: name, Path: path, Version: version, Package: ccPackage}

	_, _, err := resource.InstallChaincode(reqCtx, icr, targets, resource.WithRetry(retry.DefaultResMgmtOpts))
	if err != nil {
		return errors.WithMessage(err, "InstallChaincode failed")
	}

	return nil
}

func getRandomCCVersion() string {
	return "v0" + strconv.Itoa(rand.Intn(10000000))
}

func getContext(sdk *fabsdk.FabricSDK, user string, orgName string) (reqContext.Context, reqContext.CancelFunc, error) {

	ctx := sdk.Context(fabsdk.WithUser(user), fabsdk.WithOrg(orgName))

	clientContext, err := ctx()
	if err != nil {
		return nil, nil, errors.WithMessage(err, "create context failed")
	}

	reqCtx, cancel := context.NewRequest(&context.Client{Providers: clientContext, SigningIdentity: clientContext}, context.WithTimeoutType(fab.PeerResponse))
	return reqCtx, cancel, nil
}
