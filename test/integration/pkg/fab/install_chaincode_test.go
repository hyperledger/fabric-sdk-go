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
	"github.com/hyperledger/fabric-sdk-go/pkg/fabsdk"
	"github.com/hyperledger/fabric-sdk-go/test/integration"
	"github.com/hyperledger/fabric-protos-go/common"
	"github.com/hyperledger/fabric-protos-go/peer"
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

	testChaincodeInstallUsingChaincodePath(t, sdk, testSetup)

	testChaincodeInstallUsingChaincodePackage(t, sdk, testSetup)
}

// Test chaincode install using chaincodePath to create chaincodePackage
func testChaincodeInstallUsingChaincodePath(t *testing.T, sdk *fabsdk.FabricSDK, testSetup *integration.BaseSetupImpl) {
	chainCodeVersion := getRandomCCVersion()

	ccPkg, err := packager.NewCCPackage(chainCodePath, integration.GetDeployPath())
	if err != nil {
		t.Fatal("Failed to package chaincode")
	}

	// Low level resource
	reqCtx, cancel, err := getContext(sdk, "Admin", org1Name)
	if err != nil {
		t.Fatalf("Failed to get resource: %s", err)
	}
	defer cancel()

	peers, err := getProposalProcessors(sdk, "Admin", testSetup.OrgID, testSetup.Targets)
	require.Nil(t, err, "creating peers failed")

	if err := installCC(t, reqCtx, chainCodeName, chainCodePath, chainCodeVersion, ccPkg, peers); err != nil {
		t.Fatalf("installCC return error: %s", err)
	}

	chaincodeQueryResponse, err := resource.QueryInstalledChaincodes(reqCtx, peers[0], resource.WithRetry(retry.DefaultResMgmtOpts))

	if err != nil {
		t.Fatalf("QueryInstalledChaincodes return error: %s", err)
	}
	retrieveInstalledCC(chaincodeQueryResponse, chainCodeVersion, t)
	//Install same chaincode again, should fail
	err = installCC(t, reqCtx, chainCodeName, chainCodePath, chainCodeVersion, ccPkg, peers)

	if err == nil {
		t.Fatal("install same chaincode didn't return error")
	}
	if strings.Contains(err.Error(), "chaincodes/install.v"+chainCodeVersion+" exists") {
		t.Fatalf("install same chaincode didn't return the correct error. It returned: %s", err)
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
		t.Fatal("Failed to retrieve installed chaincode.")
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
	reqCtx, cancel, err := getContext(sdk, "Admin", org1Name)
	if err != nil {
		t.Fatalf("Failed to get resource: %s", err)
	}
	defer cancel()

	peers, err := getProposalProcessors(sdk, "Admin", testSetup.OrgID, testSetup.Targets)
	require.Nil(t, err, "creating peers failed")

	err = installCC(t, reqCtx, "install", "github.com/example_cc_pkg", chainCodeVersion, ccPkg, peers)

	if err != nil {
		t.Fatalf("installCC return error: %s", err)
	}

	//Install same chaincode again, should fail
	err = installCC(t, reqCtx, "install", chainCodePath, chainCodeVersion, ccPkg, peers)

	if err == nil {
		t.Fatal("install same chaincode didn't return error")
	}
	if strings.Contains(err.Error(), "chaincodes/install.v"+chainCodeVersion+" exists") {
		t.Fatal("install same chaincode didn't return the correct error")
	}
}

// installCC use low level client to install chaincode
func installCC(t *testing.T, reqCtx reqContext.Context, name string, path string, version string, ccPackage *resource.CCPackage, targets []fab.ProposalProcessor) error {

	icr := resource.InstallChaincodeRequest{Name: name, Path: path, Version: version, Package: ccPackage}

	r, _, err := resource.InstallChaincode(reqCtx, icr, targets, resource.WithRetry(retry.DefaultResMgmtOpts))
	if err != nil {
		return errors.WithMessage(err, "InstallChaincode failed")
	}
	t.Logf("resource.InstallChaincode, responses: [%s]", r)

	// check if response status is not success
	for _, response := range r {
		// return on first not success status
		if response.Status != int32(common.Status_SUCCESS) {
			return errors.Errorf("InstallChaincode returned response status: [%d], cc status: [%d], message: [%s]", response.Status, response.ChaincodeStatus, response.GetResponse().Message)
		}
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
