/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package resmgmt

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/golang/protobuf/proto"
	lb "github.com/hyperledger/fabric-protos-go/peer/lifecycle"
	"github.com/stretchr/testify/require"

	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/fab"
	lifecyclepkg "github.com/hyperledger/fabric-sdk-go/pkg/fab/ccpackager/lifecycle"
	fcmocks "github.com/hyperledger/fabric-sdk-go/pkg/fab/mocks"
	"github.com/hyperledger/fabric-sdk-go/pkg/fab/resource"
)

func TestClient_LifecycleInstallCC(t *testing.T) {
	req := LifecycleInstallCCRequest{
		Label:   "cc1",
		Package: []byte("cc package"),
	}

	packageID := lifecyclepkg.ComputePackageID(req.Label, req.Package)

	response := &lb.InstallChaincodeResult{
		PackageId: packageID,
		Label:     req.Label,
	}

	peer1 := &fcmocks.MockPeer{MockName: "Peer1", MockURL: "grpc://peer1.com", MockRoles: []string{}, MockCert: nil, MockMSP: "Org1MSP", Status: http.StatusOK}

	t.Run("Success", func(t *testing.T) {
		rc := setupDefaultResMgmtClient(t)

		lifecycleResource := &MockLifecycleResource{}
		lifecycleResource.GetInstalledPackageReturns(nil, fmt.Errorf("chaincode install package '%s' not found", packageID))
		lifecycleResource.InstallReturns([]*resource.LifecycleInstallProposalResponse{{
			TransactionProposalResponse: &fab.TransactionProposalResponse{
				Endorser: peer1.Name(),
				Status:   200,
			},
			InstallChaincodeResult: response,
		}}, nil)
		rc.lifecycleProcessor.lifecycleResource = lifecycleResource

		resp, err := rc.LifecycleInstallCC(req, WithTargets(peer1))
		require.NoError(t, err)
		require.Len(t, resp, 1)
		require.Equal(t, packageID, resp[0].PackageID)
	})

	t.Run("Already installed", func(t *testing.T) {
		rc := setupDefaultResMgmtClient(t)

		resp, err := rc.LifecycleInstallCC(req, WithTargets(peer1))
		require.NoError(t, err)
		require.Empty(t, resp)
	})

	t.Run("No label error", func(t *testing.T) {
		rc := setupDefaultResMgmtClient(t)

		resp, err := rc.LifecycleInstallCC(LifecycleInstallCCRequest{Package: []byte("cc package")}, WithTargets(peer1))
		require.EqualError(t, err, "label is required")
		require.Empty(t, resp)
	})

	t.Run("No package error", func(t *testing.T) {
		rc := setupDefaultResMgmtClient(t)

		resp, err := rc.LifecycleInstallCC(LifecycleInstallCCRequest{Label: "cc1"}, WithTargets(peer1))
		require.EqualError(t, err, "package is required")
		require.Empty(t, resp)
	})

	t.Run("No targets error", func(t *testing.T) {
		rc := setupDefaultResMgmtClient(t)

		resp, err := rc.LifecycleInstallCC(req)
		require.Error(t, err)
		require.Contains(t, err.Error(), "no targets available")
		require.Empty(t, resp)
	})

	t.Run("Install error", func(t *testing.T) {
		errExpected := fmt.Errorf("injected error")

		rc := setupDefaultResMgmtClient(t)
		lifecycleResource := &MockLifecycleResource{}
		lifecycleResource.InstallReturns(nil, errExpected)
		lifecycleResource.GetInstalledPackageReturns(nil, fmt.Errorf("chaincode install package '%s' not found", packageID))

		rc.lifecycleProcessor.lifecycleResource = lifecycleResource

		resp, err := rc.LifecycleInstallCC(req, WithTargets(peer1))
		require.Error(t, err)
		require.Contains(t, err.Error(), errExpected.Error())
		require.Empty(t, resp)
	})
}

func TestClient_LifecycleGetInstalledCCPackage(t *testing.T) {
	installedPackage := []byte("cc package")
	packageID := lifecyclepkg.ComputePackageID("cc1", installedPackage)

	response := &lb.GetInstalledChaincodePackageResult{
		ChaincodeInstallPackage: installedPackage,
	}

	responseBytes, err := proto.Marshal(response)
	require.NoError(t, err)

	peer1 := &fcmocks.MockPeer{MockName: "Peer1", MockURL: "grpc://peer1.com", MockRoles: []string{}, MockCert: nil, MockMSP: "Org1MSP", Status: http.StatusOK, Payload: responseBytes}

	t.Run("Success", func(t *testing.T) {
		rc := setupDefaultResMgmtClient(t)

		resp, err := rc.LifecycleGetInstalledCCPackage(packageID, WithTargets(peer1))
		require.NoError(t, err)
		require.Equal(t, installedPackage, resp)
	})

	t.Run("No targets", func(t *testing.T) {
		rc := setupDefaultResMgmtClient(t)

		resp, err := rc.LifecycleGetInstalledCCPackage(packageID)
		require.EqualError(t, err, "only one target is supported")
		require.Empty(t, resp)
	})

	t.Run("Get install package error", func(t *testing.T) {
		errExpected := fmt.Errorf("injected error")

		rc := setupDefaultResMgmtClient(t)

		lifecycleResource := &MockLifecycleResource{}
		lifecycleResource.GetInstalledPackageReturns(nil, errExpected)
		rc.lifecycleProcessor.lifecycleResource = lifecycleResource

		resp, err := rc.LifecycleGetInstalledCCPackage(packageID, WithTargets(peer1))
		require.EqualError(t, err, errExpected.Error())
		require.Empty(t, resp)
	})
}

func TestClient_LifecycleQueryInstalled(t *testing.T) {
	const packageID = "pkg1"
	const label = "label1"
	const cc1 = "cc1"
	const v1 = "v1"
	const channel1 = "channel1"

	response := &lb.QueryInstalledChaincodesResult{
		InstalledChaincodes: []*lb.QueryInstalledChaincodesResult_InstalledChaincode{
			{
				PackageId: packageID,
				Label:     label,
				References: map[string]*lb.QueryInstalledChaincodesResult_References{
					channel1: {
						Chaincodes: []*lb.QueryInstalledChaincodesResult_Chaincode{
							{
								Name:    cc1,
								Version: v1,
							},
						},
					},
				},
			},
		},
	}

	responseBytes, err := proto.Marshal(response)
	require.NoError(t, err)

	peer1 := &fcmocks.MockPeer{Payload: responseBytes}

	t.Run("Success", func(t *testing.T) {
		rc := setupDefaultResMgmtClient(t)

		resp, err := rc.LifecycleQueryInstalledCC(WithTargets(peer1))
		require.NoError(t, err)
		require.Len(t, resp, 1)
		require.Equal(t, packageID, resp[0].PackageID)
		require.Equal(t, label, resp[0].Label)
		require.Len(t, resp[0].References, 1)

		references, ok := resp[0].References[channel1]
		require.True(t, ok)
		require.Len(t, references, 1)
		require.Equal(t, cc1, references[0].Name)
		require.Equal(t, v1, references[0].Version)
	})

	t.Run("No targets", func(t *testing.T) {
		rc := setupDefaultResMgmtClient(t)

		resp, err := rc.LifecycleQueryInstalledCC()
		require.EqualError(t, err, "only one target is supported")
		require.Empty(t, resp)
	})

	t.Run("Marshal error", func(t *testing.T) {
		rc := setupDefaultResMgmtClient(t)

		resp, err := rc.LifecycleQueryInstalledCC(WithTargets(&fcmocks.MockPeer{Payload: []byte("invalid payload")}))
		require.Error(t, err)
		require.Contains(t, err.Error(), "failed to unmarshal proposal response's response payload")
		require.Empty(t, resp)
	})
}
