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
	pb "github.com/hyperledger/fabric-protos-go/peer"
	lb "github.com/hyperledger/fabric-protos-go/peer/lifecycle"
	"github.com/stretchr/testify/require"

	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/fab"
	lifecyclepkg "github.com/hyperledger/fabric-sdk-go/pkg/fab/ccpackager/lifecycle"
	fcmocks "github.com/hyperledger/fabric-sdk-go/pkg/fab/mocks"
	"github.com/hyperledger/fabric-sdk-go/pkg/fab/resource"
	"github.com/hyperledger/fabric-sdk-go/third_party/github.com/hyperledger/fabric/common/policydsl"
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

func TestClient_LifecycleApproveCC(t *testing.T) {
	const channelID = "channel1"

	respBytes, err := proto.Marshal(&lb.ApproveChaincodeDefinitionForMyOrgResult{})
	require.NoError(t, err)

	peer1 := &fcmocks.MockPeer{Payload: respBytes}
	rc := setupDefaultResMgmtClient(t)

	req := LifecycleApproveCCRequest{
		Name:      "cc1",
		Version:   "v1",
		PackageID: "pkg1",
		Sequence:  1,
	}

	t.Run("Success", func(t *testing.T) {
		txnID, err := rc.LifecycleApproveCC(channelID, req, WithTargets(peer1))
		require.NoError(t, err)
		require.NotEmpty(t, txnID)
	})

	t.Run("No channel ID -> error", func(t *testing.T) {
		txnID, err := rc.LifecycleApproveCC("", req, WithTargets(peer1))
		require.EqualError(t, err, "channel ID is required")
		require.Empty(t, txnID)
	})

	t.Run("No name -> error", func(t *testing.T) {
		req := LifecycleApproveCCRequest{
			Version:   "v1",
			PackageID: "pkg1",
			Sequence:  1,
		}

		txnID, err := rc.LifecycleApproveCC(channelID, req, WithTargets(peer1))
		require.EqualError(t, err, "name is required")
		require.Empty(t, txnID)
	})

	t.Run("No version -> error", func(t *testing.T) {
		req := LifecycleApproveCCRequest{
			Name:      "cc1",
			PackageID: "pkg1",
			Sequence:  1,
		}

		txnID, err := rc.LifecycleApproveCC(channelID, req, WithTargets(peer1))
		require.EqualError(t, err, "version is required")
		require.Empty(t, txnID)
	})

	t.Run("Get targets -> error", func(t *testing.T) {
		errExpected := fmt.Errorf("injected targets error")

		rc := setupDefaultResMgmtClient(t)
		rc.lifecycleProcessor.getCCProposalTargets = func(string, requestOptions) ([]fab.Peer, error) { return nil, errExpected }

		txnID, err := rc.LifecycleApproveCC(channelID, req, WithTargets(peer1))
		require.EqualError(t, err, errExpected.Error())
		require.Empty(t, txnID)
	})

	t.Run("CreateProposal -> error", func(t *testing.T) {
		errExpected := fmt.Errorf("injected proposal error")

		rc := setupDefaultResMgmtClient(t)

		lifecycleResource := &MockLifecycleResource{}
		lifecycleResource.CreateApproveProposalReturns(nil, errExpected)
		rc.lifecycleProcessor.lifecycleResource = lifecycleResource

		txnID, err := rc.LifecycleApproveCC(channelID, req, WithTargets(peer1))
		require.Error(t, err, errExpected.Error())
		require.Contains(t, err.Error(), errExpected.Error())
		require.Empty(t, txnID)
	})

	t.Run("VerifySignature -> error", func(t *testing.T) {
		errExpected := fmt.Errorf("injected signature error")

		rc := setupDefaultResMgmtClient(t)

		rc.lifecycleProcessor.verifyTPSignature = func(fab.ChannelService, []*fab.TransactionProposalResponse) error { return errExpected }

		txnID, err := rc.LifecycleApproveCC(channelID, req, WithTargets(peer1))
		require.Error(t, err, errExpected.Error())
		require.Contains(t, err.Error(), errExpected.Error())
		require.NotEmpty(t, txnID)
	})

	t.Run("Channel provider -> error", func(t *testing.T) {
		ctx := setupTestContext("test", "Org1MSP")
		ctx.SetEndpointConfig(getNetworkConfig(t))
		cp := &MockChannelProvider{}

		ctx.SetCustomChannelProvider(cp)

		rc := setupResMgmtClient(t, ctx, getDefaultTargetFilterOption())
		rc.lifecycleProcessor.getCCProposalTargets = func(string, requestOptions) ([]fab.Peer, error) { return []fab.Peer{peer1}, nil }

		t.Run("ChannelService error", func(t *testing.T) {
			errExpected := fmt.Errorf("injected provider error")
			cp.ChannelServiceReturns(nil, errExpected)

			txnID, err := rc.LifecycleApproveCC(channelID, req, WithTargets(peer1))
			require.Error(t, err)
			require.Contains(t, err.Error(), errExpected.Error())
			require.Empty(t, txnID)
		})

		t.Run("Transactor error", func(t *testing.T) {
			errExpected := fmt.Errorf("injected transactor error")
			cs := &MockChannelService{}
			cs.TransactorReturns(nil, errExpected)
			cp.ChannelServiceReturns(cs, nil)

			txnID, err := rc.LifecycleApproveCC(channelID, req, WithTargets(peer1))
			require.Error(t, err)
			require.Contains(t, err.Error(), errExpected.Error())
			require.Empty(t, txnID)
		})

		t.Run("EventService error", func(t *testing.T) {
			errExpected := fmt.Errorf("injected event service error")
			cs := &MockChannelService{}
			cs.EventServiceReturns(nil, errExpected)
			cp.ChannelServiceReturns(cs, nil)

			txnID, err := rc.LifecycleApproveCC(channelID, req, WithTargets(peer1))
			require.Error(t, err)
			require.Contains(t, err.Error(), errExpected.Error())
			require.Empty(t, txnID)
		})
	})
}

func TestClient_LifecycleQueryApprovedCC(t *testing.T) {
	const packageID = "pkg1"
	const cc1 = "cc1"
	const v1 = "v1"
	const channel1 = "channel1"

	applicationPolicy := &pb.ApplicationPolicy{
		Type: &pb.ApplicationPolicy_SignaturePolicy{
			SignaturePolicy: policydsl.AcceptAllPolicy,
		},
	}

	policyBytes, err := proto.Marshal(applicationPolicy)
	require.NoError(t, err)

	response := &lb.QueryApprovedChaincodeDefinitionResult{
		Sequence:            1,
		Version:             v1,
		ValidationParameter: policyBytes,
		Source: &lb.ChaincodeSource{
			Type: &lb.ChaincodeSource_LocalPackage{
				LocalPackage: &lb.ChaincodeSource_Local{
					PackageId: packageID,
				},
			},
		},
		Collections: &pb.CollectionConfigPackage{},
	}

	responseBytes, err := proto.Marshal(response)
	require.NoError(t, err)

	peer1 := &fcmocks.MockPeer{Payload: responseBytes}

	req := LifecycleQueryApprovedCCRequest{
		Name:     cc1,
		Sequence: 1,
	}

	t.Run("Success", func(t *testing.T) {
		rc := setupDefaultResMgmtClient(t)

		resp, err := rc.LifecycleQueryApprovedCC(channel1, req, WithTargets(peer1))
		require.NoError(t, err)
		require.NotNil(t, resp)
	})

	t.Run("No targets", func(t *testing.T) {
		rc := setupDefaultResMgmtClient(t)

		resp, err := rc.LifecycleQueryApprovedCC(channel1, req)
		require.EqualError(t, err, "only one target is supported")
		require.Empty(t, resp)
	})

	t.Run("Marshal error", func(t *testing.T) {
		rc := setupDefaultResMgmtClient(t)

		resp, err := rc.LifecycleQueryApprovedCC(channel1, req, WithTargets(&fcmocks.MockPeer{Payload: []byte("invalid payload")}))
		require.Error(t, err)
		require.Contains(t, err.Error(), "failed to unmarshal proposal response's response payload")
		require.Empty(t, resp)
	})

	t.Run("No channel ID -> error", func(t *testing.T) {
		rc := setupDefaultResMgmtClient(t)

		resp, err := rc.LifecycleQueryApprovedCC("", req, WithTargets(peer1))
		require.EqualError(t, err, "channel ID is required")
		require.Empty(t, resp)
	})

	t.Run("No name -> error", func(t *testing.T) {
		rc := setupDefaultResMgmtClient(t)

		resp, err := rc.LifecycleQueryApprovedCC(channel1, LifecycleQueryApprovedCCRequest{}, WithTargets(peer1))
		require.EqualError(t, err, "name is required")
		require.Empty(t, resp)
	})
}

func TestClient_LifecycleCheckCCCommitReadiness(t *testing.T) {
	const cc1 = "cc1"
	const v1 = "v1"
	const channel1 = "channel1"

	response := &lb.CheckCommitReadinessResult{
		Approvals: map[string]bool{"org1": true, "org2": false},
	}

	responseBytes, err := proto.Marshal(response)
	require.NoError(t, err)

	peer1 := &fcmocks.MockPeer{Payload: responseBytes}

	req := LifecycleCheckCCCommitReadinessRequest{
		Name:     cc1,
		Version:  v1,
		Sequence: 1,
	}

	ctx := setupTestContext("test", "Org1MSP")
	ctx.SetEndpointConfig(getNetworkConfig(t))

	cs := &MockChannelService{}
	transactor := &MockTransactor{}

	result := []*fab.TransactionProposalResponse{
		{
			ProposalResponse: &pb.ProposalResponse{
				Response: &pb.Response{
					Payload: responseBytes,
				},
			},
		},
	}

	transactor.SendTransactionProposalReturns(result, nil)
	cs.TransactorReturns(transactor, nil)

	cp := &MockChannelProvider{}
	cp.ChannelServiceReturns(cs, nil)

	rc := setupResMgmtClient(t, ctx, getDefaultTargetFilterOption())
	rc.lifecycleProcessor.verifyTPSignature = func(fab.ChannelService, []*fab.TransactionProposalResponse) error { return nil }
	rc.lifecycleProcessor.getCCProposalTargets = func(string, requestOptions) ([]fab.Peer, error) { return []fab.Peer{peer1}, nil }

	t.Run("Success", func(t *testing.T) {
		ctx.SetCustomChannelProvider(cp)

		resp, err := rc.LifecycleCheckCCCommitReadiness(channel1, req, WithTargets(peer1))
		require.NoError(t, err)
		require.NotNil(t, resp)
		require.Len(t, resp.Approvals, 2)
	})

	t.Run("No channel ID -> error", func(t *testing.T) {
		ctx.SetCustomChannelProvider(cp)

		resp, err := rc.LifecycleCheckCCCommitReadiness("", req, WithTargets(peer1))
		require.EqualError(t, err, "channel ID is required")
		require.Empty(t, resp)
	})

	t.Run("No name -> error", func(t *testing.T) {
		ctx.SetCustomChannelProvider(cp)

		req := LifecycleCheckCCCommitReadinessRequest{
			Version:  v1,
			Sequence: 1,
		}

		resp, err := rc.LifecycleCheckCCCommitReadiness(channel1, req, WithTargets(peer1))
		require.EqualError(t, err, "name is required")
		require.Empty(t, resp)
	})

	t.Run("No version -> error", func(t *testing.T) {
		ctx.SetCustomChannelProvider(cp)

		req := LifecycleCheckCCCommitReadinessRequest{
			Name:     cc1,
			Sequence: 1,
		}

		resp, err := rc.LifecycleCheckCCCommitReadiness(channel1, req, WithTargets(peer1))
		require.EqualError(t, err, "version is required")
		require.Empty(t, resp)
	})

	t.Run("Get targets -> error", func(t *testing.T) {
		ctx.SetCustomChannelProvider(cp)

		errExpected := fmt.Errorf("injected targets error")

		rc := setupResMgmtClient(t, ctx, getDefaultTargetFilterOption())
		rc.lifecycleProcessor.getCCProposalTargets = func(channelID string, opts requestOptions) ([]fab.Peer, error) { return nil, errExpected }

		resp, err := rc.LifecycleCheckCCCommitReadiness(channel1, req, WithTargets(peer1))
		require.EqualError(t, err, errExpected.Error())
		require.Empty(t, resp)
	})

	t.Run("ChannelService -> error", func(t *testing.T) {
		errExpected := fmt.Errorf("injected provider error")

		cp := &MockChannelProvider{}
		cp.ChannelServiceReturns(nil, errExpected)
		ctx.SetCustomChannelProvider(cp)

		resp, err := rc.LifecycleCheckCCCommitReadiness(channel1, req, WithTargets(peer1))
		require.Error(t, err)
		require.Contains(t, err.Error(), errExpected.Error())
		require.Empty(t, resp)
	})

	t.Run("Transactor -> error", func(t *testing.T) {
		errExpected := fmt.Errorf("injected transactor error")
		cs := &MockChannelService{}
		cs.TransactorReturns(nil, errExpected)

		cp := &MockChannelProvider{}
		cp.ChannelServiceReturns(cs, nil)

		ctx.SetCustomChannelProvider(cp)

		resp, err := rc.LifecycleCheckCCCommitReadiness(channel1, req, WithTargets(peer1))
		require.Error(t, err)
		require.Contains(t, err.Error(), errExpected.Error())
		require.Empty(t, resp)
	})

	t.Run("Signature -> error", func(t *testing.T) {
		errExpected := fmt.Errorf("injected signature error")

		ctx.SetCustomChannelProvider(cp)

		rc := setupResMgmtClient(t, ctx, getDefaultTargetFilterOption())
		rc.lifecycleProcessor.getCCProposalTargets = func(string, requestOptions) ([]fab.Peer, error) { return []fab.Peer{peer1}, nil }
		rc.lifecycleProcessor.verifyTPSignature = func(fab.ChannelService, []*fab.TransactionProposalResponse) error { return errExpected }

		resp, err := rc.LifecycleCheckCCCommitReadiness(channel1, req, WithTargets(peer1))
		require.Error(t, err)
		require.Contains(t, err.Error(), errExpected.Error())
		require.Empty(t, resp)
	})

	t.Run("CreateProposal -> error", func(t *testing.T) {
		errExpected := fmt.Errorf("injected create proposal error")

		ctx.SetCustomChannelProvider(cp)

		rc := setupResMgmtClient(t, ctx, getDefaultTargetFilterOption())
		rc.lifecycleProcessor.getCCProposalTargets = func(string, requestOptions) ([]fab.Peer, error) { return []fab.Peer{peer1}, nil }

		lr := &MockLifecycleResource{}
		lr.CreateCheckCommitReadinessProposalReturns(nil, errExpected)

		rc.lifecycleProcessor.lifecycleResource = lr

		resp, err := rc.LifecycleCheckCCCommitReadiness(channel1, req, WithTargets(peer1))
		require.Error(t, err)
		require.Contains(t, err.Error(), errExpected.Error())
		require.Empty(t, resp)
	})

	t.Run("No responses -> error", func(t *testing.T) {
		cs := &MockChannelService{}
		transactor := &MockTransactor{}

		transactor.SendTransactionProposalReturns(nil, nil)
		cs.TransactorReturns(transactor, nil)

		cp := &MockChannelProvider{}
		cp.ChannelServiceReturns(cs, nil)

		ctx.SetCustomChannelProvider(cp)

		resp, err := rc.LifecycleCheckCCCommitReadiness(channel1, req, WithTargets(peer1))
		require.EqualError(t, err, "no responses")
		require.Empty(t, resp)
	})

	t.Run("Endorsements not matching -> error", func(t *testing.T) {
		responseBytes1, err := proto.Marshal(&lb.CheckCommitReadinessResult{
			Approvals: map[string]bool{"org1": true, "org2": false},
		})
		require.NoError(t, err)

		responseBytes2, err := proto.Marshal(&lb.CheckCommitReadinessResult{
			Approvals: map[string]bool{"org3": true},
		})
		require.NoError(t, err)

		cs := &MockChannelService{}
		transactor := &MockTransactor{}

		result := []*fab.TransactionProposalResponse{
			{
				ProposalResponse: &pb.ProposalResponse{
					Response: &pb.Response{
						Payload: responseBytes1,
					},
				},
			},
			{
				ProposalResponse: &pb.ProposalResponse{
					Response: &pb.Response{
						Payload: responseBytes2,
					},
				},
			},
		}

		transactor.SendTransactionProposalReturns(result, nil)
		cs.TransactorReturns(transactor, nil)

		cp := &MockChannelProvider{}
		cp.ChannelServiceReturns(cs, nil)

		ctx.SetCustomChannelProvider(cp)

		resp, err := rc.LifecycleCheckCCCommitReadiness(channel1, req, WithTargets(peer1))
		require.Error(t, err)
		require.Contains(t, err.Error(), "responses from endorsers do not match")
		require.Empty(t, resp)
	})
}

func TestClient_LifecycleCommitCC(t *testing.T) {
	const channelID = "channel1"

	respBytes, err := proto.Marshal(&lb.CommitChaincodeDefinitionResult{})
	require.NoError(t, err)

	peer1 := &fcmocks.MockPeer{Payload: respBytes}
	rc := setupDefaultResMgmtClient(t)

	req := LifecycleCommitCCRequest{
		Name:     "cc1",
		Version:  "v1",
		Sequence: 1,
	}

	t.Run("Success", func(t *testing.T) {
		txnID, err := rc.LifecycleCommitCC(channelID, req, WithTargets(peer1))
		require.NoError(t, err)
		require.NotEmpty(t, txnID)
	})

	t.Run("No channel ID -> error", func(t *testing.T) {
		txnID, err := rc.LifecycleCommitCC("", req, WithTargets(peer1))
		require.EqualError(t, err, "channel ID is required")
		require.Empty(t, txnID)
	})

	t.Run("No name -> error", func(t *testing.T) {
		req := LifecycleCommitCCRequest{
			Version:  "v1",
			Sequence: 1,
		}

		txnID, err := rc.LifecycleCommitCC(channelID, req, WithTargets(peer1))
		require.EqualError(t, err, "name is required")
		require.Empty(t, txnID)
	})

	t.Run("No version -> error", func(t *testing.T) {
		req := LifecycleCommitCCRequest{
			Name:     "cc1",
			Sequence: 1,
		}

		txnID, err := rc.LifecycleCommitCC(channelID, req, WithTargets(peer1))
		require.EqualError(t, err, "version is required")
		require.Empty(t, txnID)
	})

	t.Run("Get targets -> error", func(t *testing.T) {
		errExpected := fmt.Errorf("injected targets error")

		rc := setupDefaultResMgmtClient(t)
		rc.lifecycleProcessor.getCCProposalTargets = func(string, requestOptions) ([]fab.Peer, error) { return nil, errExpected }

		txnID, err := rc.LifecycleCommitCC(channelID, req, WithTargets(peer1))
		require.EqualError(t, err, errExpected.Error())
		require.Empty(t, txnID)
	})

	t.Run("CreateProposal -> error", func(t *testing.T) {
		errExpected := fmt.Errorf("injected proposal error")

		rc := setupDefaultResMgmtClient(t)

		lifecycleResource := &MockLifecycleResource{}
		lifecycleResource.CreateCommitProposalReturns(nil, errExpected)
		rc.lifecycleProcessor.lifecycleResource = lifecycleResource

		txnID, err := rc.LifecycleCommitCC(channelID, req, WithTargets(peer1))
		require.Error(t, err, errExpected.Error())
		require.Contains(t, err.Error(), errExpected.Error())
		require.Empty(t, txnID)
	})

	t.Run("VerifySignature -> error", func(t *testing.T) {
		errExpected := fmt.Errorf("injected signature error")

		rc := setupDefaultResMgmtClient(t)

		rc.lifecycleProcessor.verifyTPSignature = func(fab.ChannelService, []*fab.TransactionProposalResponse) error { return errExpected }

		txnID, err := rc.LifecycleCommitCC(channelID, req, WithTargets(peer1))
		require.Error(t, err, errExpected.Error())
		require.Contains(t, err.Error(), errExpected.Error())
		require.NotEmpty(t, txnID)
	})

	t.Run("Channel provider -> error", func(t *testing.T) {
		ctx := setupTestContext("test", "Org1MSP")
		ctx.SetEndpointConfig(getNetworkConfig(t))
		cp := &MockChannelProvider{}

		ctx.SetCustomChannelProvider(cp)

		rc := setupResMgmtClient(t, ctx, getDefaultTargetFilterOption())
		rc.lifecycleProcessor.getCCProposalTargets = func(string, requestOptions) ([]fab.Peer, error) { return []fab.Peer{peer1}, nil }

		t.Run("ChannelService error", func(t *testing.T) {
			errExpected := fmt.Errorf("injected provider error")
			cp.ChannelServiceReturns(nil, errExpected)

			txnID, err := rc.LifecycleCommitCC(channelID, req, WithTargets(peer1))
			require.Error(t, err)
			require.Contains(t, err.Error(), errExpected.Error())
			require.Empty(t, txnID)
		})

		t.Run("Transactor error", func(t *testing.T) {
			errExpected := fmt.Errorf("injected transactor error")
			cs := &MockChannelService{}
			cs.TransactorReturns(nil, errExpected)
			cp.ChannelServiceReturns(cs, nil)

			txnID, err := rc.LifecycleCommitCC(channelID, req, WithTargets(peer1))
			require.Error(t, err)
			require.Contains(t, err.Error(), errExpected.Error())
			require.Empty(t, txnID)
		})

		t.Run("EventService error", func(t *testing.T) {
			errExpected := fmt.Errorf("injected event service error")
			cs := &MockChannelService{}
			cs.EventServiceReturns(nil, errExpected)
			cp.ChannelServiceReturns(cs, nil)

			txnID, err := rc.LifecycleCommitCC(channelID, req, WithTargets(peer1))
			require.Error(t, err)
			require.Contains(t, err.Error(), errExpected.Error())
			require.Empty(t, txnID)
		})
	})
}

func TestClient_LifecycleQueryCommittedCC(t *testing.T) {
	const cc1 = "cc1"
	const cc2 = "cc2"
	const v1 = "v1"
	const channel1 = "channel1"

	lc := resource.NewLifecycle()
	policyBytes, err := lc.MarshalApplicationPolicy(nil, "channel config policy")
	require.NoError(t, err)

	collections := &pb.CollectionConfigPackage{
		Config: []*pb.CollectionConfig{
			{
				Payload: &pb.CollectionConfig_StaticCollectionConfig{
					StaticCollectionConfig: &pb.StaticCollectionConfig{
						Name: "coll1",
					},
				},
			},
		},
	}

	lcDef := &lb.QueryChaincodeDefinitionResult{
		Sequence:            1,
		Version:             v1,
		ValidationParameter: policyBytes,
		Collections:         collections,
		Approvals:           map[string]bool{"org1": true, "org2": false},
	}

	lcDefBytes, err := proto.Marshal(lcDef)
	require.NoError(t, err)

	lcDefs := &lb.QueryChaincodeDefinitionsResult{
		ChaincodeDefinitions: []*lb.QueryChaincodeDefinitionsResult_ChaincodeDefinition{
			{Name: cc1, Sequence: 1, Version: v1, ValidationParameter: policyBytes, Collections: collections},
			{Name: cc2, Sequence: 2, Version: v1, ValidationParameter: policyBytes},
		},
	}

	lcDefsBytes, err := proto.Marshal(lcDefs)
	require.NoError(t, err)

	//this result is used to test a case when peers return list of definitions in unexpected order
	lcDefsInDifferentOrder := &lb.QueryChaincodeDefinitionsResult{
		ChaincodeDefinitions: []*lb.QueryChaincodeDefinitionsResult_ChaincodeDefinition{
			{Name: cc2, Sequence: 2, Version: v1, ValidationParameter: policyBytes},
			{Name: cc1, Sequence: 1, Version: v1, ValidationParameter: policyBytes, Collections: collections},
		},
	}

	lcDefsInDifferentOrderBytes, err := proto.Marshal(lcDefsInDifferentOrder)
	require.NoError(t, err)

	ctx := setupTestContext("test", "Org1MSP")
	ctx.SetEndpointConfig(getNetworkConfig(t))

	cs := &MockChannelService{}
	transactor := &MockTransactor{}

	singleCCResponse := []*fab.TransactionProposalResponse{
		{
			ProposalResponse: &pb.ProposalResponse{
				Response: &pb.Response{
					Payload: lcDefBytes,
				},
			},
		},
	}

	allCCsResponse := []*fab.TransactionProposalResponse{
		{
			ProposalResponse: &pb.ProposalResponse{
				Response: &pb.Response{
					Payload: lcDefsBytes,
				},
			},
		},
		{
			ProposalResponse: &pb.ProposalResponse{
				Response: &pb.Response{
					Payload: lcDefsInDifferentOrderBytes,
				},
			},
		},
	}

	cs.TransactorReturns(transactor, nil)

	cp := &MockChannelProvider{}
	cp.ChannelServiceReturns(cs, nil)

	rc := setupResMgmtClient(t, ctx, getDefaultTargetFilterOption())
	rc.lifecycleProcessor.verifyTPSignature = func(fab.ChannelService, []*fab.TransactionProposalResponse) error { return nil }
	rc.lifecycleProcessor.getCCProposalTargets = func(string, requestOptions) ([]fab.Peer, error) { return []fab.Peer{&fcmocks.MockPeer{}}, nil }

	t.Run("With name -> success", func(t *testing.T) {
		ctx.SetCustomChannelProvider(cp)
		transactor.SendTransactionProposalReturns(singleCCResponse, nil)

		resp, err := rc.LifecycleQueryCommittedCC(channel1, LifecycleQueryCommittedCCRequest{Name: cc1}, WithTargets(&fcmocks.MockPeer{}))
		require.NoError(t, err)
		require.Len(t, resp, 1)
		require.Equal(t, cc1, resp[0].Name)
		require.Len(t, resp[0].Approvals, 2)
	})

	t.Run("No name -> success", func(t *testing.T) {
		ctx.SetCustomChannelProvider(cp)
		transactor.SendTransactionProposalReturns(allCCsResponse, nil)

		resp, err := rc.LifecycleQueryCommittedCC(channel1, LifecycleQueryCommittedCCRequest{}, WithTargets(&fcmocks.MockPeer{}))
		require.NoError(t, err)
		require.Len(t, resp, 2)
		require.Equal(t, cc1, resp[0].Name)
		require.Equal(t, cc2, resp[1].Name)
		require.Empty(t, resp[0].Approvals)
	})

	t.Run("With name unmarshal response -> error", func(t *testing.T) {
		ctx.SetCustomChannelProvider(cp)
		transactor.SendTransactionProposalReturns(singleCCResponse, nil)

		rc := setupResMgmtClient(t, ctx, getDefaultTargetFilterOption())
		rc.lifecycleProcessor.verifyTPSignature = func(fab.ChannelService, []*fab.TransactionProposalResponse) error { return nil }
		rc.lifecycleProcessor.getCCProposalTargets = func(string, requestOptions) ([]fab.Peer, error) { return []fab.Peer{&fcmocks.MockPeer{}}, nil }
		rc.lifecycleProcessor.protoUnmarshal = func(buf []byte, pb proto.Message) error { return fmt.Errorf("injected unmarshal error") }

		resp, err := rc.LifecycleQueryCommittedCC(channel1, LifecycleQueryCommittedCCRequest{Name: cc1}, WithTargets(&fcmocks.MockPeer{}))
		require.EqualError(t, err, "failed to unmarshal proposal response's response payload: injected unmarshal error")
		require.Empty(t, resp)
	})

	t.Run("With name unmarshal policy error -> error", func(t *testing.T) {
		errExpected := fmt.Errorf("injected unmarshal error")

		ctx.SetCustomChannelProvider(cp)
		transactor.SendTransactionProposalReturns(singleCCResponse, nil)

		rc := setupResMgmtClient(t, ctx, getDefaultTargetFilterOption())
		rc.lifecycleProcessor.verifyTPSignature = func(fab.ChannelService, []*fab.TransactionProposalResponse) error { return nil }
		rc.lifecycleProcessor.getCCProposalTargets = func(string, requestOptions) ([]fab.Peer, error) { return []fab.Peer{&fcmocks.MockPeer{}}, nil }

		lr := &MockLifecycleResource{}
		lr.UnmarshalApplicationPolicyReturns(nil, "", errExpected)
		rc.lifecycleProcessor.lifecycleResource = lr

		resp, err := rc.LifecycleQueryCommittedCC(channel1, LifecycleQueryCommittedCCRequest{Name: cc1}, WithTargets(&fcmocks.MockPeer{}))
		require.EqualError(t, err, errExpected.Error())
		require.Empty(t, resp)
	})

	t.Run("No name unmarshal response -> error", func(t *testing.T) {
		ctx.SetCustomChannelProvider(cp)
		transactor.SendTransactionProposalReturns(allCCsResponse, nil)

		rc := setupResMgmtClient(t, ctx, getDefaultTargetFilterOption())
		rc.lifecycleProcessor.verifyTPSignature = func(fab.ChannelService, []*fab.TransactionProposalResponse) error { return nil }
		rc.lifecycleProcessor.getCCProposalTargets = func(string, requestOptions) ([]fab.Peer, error) { return []fab.Peer{&fcmocks.MockPeer{}}, nil }
		rc.lifecycleProcessor.protoUnmarshal = func(buf []byte, pb proto.Message) error { return fmt.Errorf("injected unmarshal error") }

		resp, err := rc.LifecycleQueryCommittedCC(channel1, LifecycleQueryCommittedCCRequest{}, WithTargets(&fcmocks.MockPeer{}))
		require.EqualError(t, err, "failed to unmarshal proposal response's response payload: injected unmarshal error")
		require.Empty(t, resp)
	})

	t.Run("No name unmarshal policy error -> error", func(t *testing.T) {
		errExpected := fmt.Errorf("injected unmarshal error")

		ctx.SetCustomChannelProvider(cp)
		transactor.SendTransactionProposalReturns(allCCsResponse, nil)

		rc := setupResMgmtClient(t, ctx, getDefaultTargetFilterOption())
		rc.lifecycleProcessor.verifyTPSignature = func(fab.ChannelService, []*fab.TransactionProposalResponse) error { return nil }
		rc.lifecycleProcessor.getCCProposalTargets = func(string, requestOptions) ([]fab.Peer, error) { return []fab.Peer{&fcmocks.MockPeer{}}, nil }

		lr := &MockLifecycleResource{}
		lr.UnmarshalApplicationPolicyReturns(nil, "", errExpected)
		rc.lifecycleProcessor.lifecycleResource = lr

		resp, err := rc.LifecycleQueryCommittedCC(channel1, LifecycleQueryCommittedCCRequest{}, WithTargets(&fcmocks.MockPeer{}))
		require.EqualError(t, err, errExpected.Error())
		require.Empty(t, resp)
	})

	t.Run("No channel ID -> error", func(t *testing.T) {
		ctx.SetCustomChannelProvider(cp)

		resp, err := rc.LifecycleQueryCommittedCC("", LifecycleQueryCommittedCCRequest{}, WithTargets(&fcmocks.MockPeer{}))
		require.EqualError(t, err, "channel ID is required")
		require.Empty(t, resp)
	})

	t.Run("Get targets -> error", func(t *testing.T) {
		ctx.SetCustomChannelProvider(cp)

		errExpected := fmt.Errorf("injected targets error")

		rc := setupResMgmtClient(t, ctx, getDefaultTargetFilterOption())
		rc.lifecycleProcessor.getCCProposalTargets = func(channelID string, opts requestOptions) ([]fab.Peer, error) { return nil, errExpected }

		resp, err := rc.LifecycleQueryCommittedCC(channel1, LifecycleQueryCommittedCCRequest{}, WithTargets(&fcmocks.MockPeer{}))
		require.EqualError(t, err, errExpected.Error())
		require.Empty(t, resp)
	})

	t.Run("ChannelService -> error", func(t *testing.T) {
		errExpected := fmt.Errorf("injected provider error")

		cp := &MockChannelProvider{}
		cp.ChannelServiceReturns(nil, errExpected)
		ctx.SetCustomChannelProvider(cp)

		resp, err := rc.LifecycleQueryCommittedCC(channel1, LifecycleQueryCommittedCCRequest{}, WithTargets(&fcmocks.MockPeer{}))
		require.Error(t, err)
		require.Contains(t, err.Error(), errExpected.Error())
		require.Empty(t, resp)
	})

	t.Run("Transactor -> error", func(t *testing.T) {
		errExpected := fmt.Errorf("injected transactor error")
		cs := &MockChannelService{}
		cs.TransactorReturns(nil, errExpected)

		cp := &MockChannelProvider{}
		cp.ChannelServiceReturns(cs, nil)

		ctx.SetCustomChannelProvider(cp)

		resp, err := rc.LifecycleQueryCommittedCC(channel1, LifecycleQueryCommittedCCRequest{}, WithTargets(&fcmocks.MockPeer{}))
		require.Error(t, err)
		require.Contains(t, err.Error(), errExpected.Error())
		require.Empty(t, resp)
	})

	t.Run("Signature -> error", func(t *testing.T) {
		errExpected := fmt.Errorf("injected signature error")

		ctx.SetCustomChannelProvider(cp)

		rc := setupResMgmtClient(t, ctx, getDefaultTargetFilterOption())
		rc.lifecycleProcessor.getCCProposalTargets = func(string, requestOptions) ([]fab.Peer, error) { return []fab.Peer{&fcmocks.MockPeer{}}, nil }
		rc.lifecycleProcessor.verifyTPSignature = func(fab.ChannelService, []*fab.TransactionProposalResponse) error { return errExpected }

		resp, err := rc.LifecycleQueryCommittedCC(channel1, LifecycleQueryCommittedCCRequest{}, WithTargets(&fcmocks.MockPeer{}))
		require.Error(t, err)
		require.Contains(t, err.Error(), errExpected.Error())
		require.Empty(t, resp)
	})

	t.Run("CreateProposal -> error", func(t *testing.T) {
		errExpected := fmt.Errorf("injected create proposal error")

		ctx.SetCustomChannelProvider(cp)

		rc := setupResMgmtClient(t, ctx, getDefaultTargetFilterOption())
		rc.lifecycleProcessor.getCCProposalTargets = func(string, requestOptions) ([]fab.Peer, error) { return []fab.Peer{&fcmocks.MockPeer{}}, nil }

		lr := &MockLifecycleResource{}
		lr.CreateQueryCommittedProposalReturns(nil, errExpected)

		rc.lifecycleProcessor.lifecycleResource = lr

		resp, err := rc.LifecycleQueryCommittedCC(channel1, LifecycleQueryCommittedCCRequest{}, WithTargets(&fcmocks.MockPeer{}))
		require.Error(t, err)
		require.Contains(t, err.Error(), errExpected.Error())
		require.Empty(t, resp)
	})

	t.Run("No responses -> error", func(t *testing.T) {
		cs := &MockChannelService{}
		transactor := &MockTransactor{}

		transactor.SendTransactionProposalReturns(nil, nil)
		cs.TransactorReturns(transactor, nil)

		cp := &MockChannelProvider{}
		cp.ChannelServiceReturns(cs, nil)

		ctx.SetCustomChannelProvider(cp)

		resp, err := rc.LifecycleQueryCommittedCC(channel1, LifecycleQueryCommittedCCRequest{}, WithTargets(&fcmocks.MockPeer{}))
		require.EqualError(t, err, "no responses")
		require.Empty(t, resp)
	})

	t.Run("Endorsements not matching -> error", func(t *testing.T) {
		responseBytes1, err := proto.Marshal(&lb.CheckCommitReadinessResult{
			Approvals: map[string]bool{"org1": true, "org2": false},
		})
		require.NoError(t, err)

		responseBytes2, err := proto.Marshal(&lb.CheckCommitReadinessResult{
			Approvals: map[string]bool{"org3": true},
		})
		require.NoError(t, err)

		cs := &MockChannelService{}
		transactor := &MockTransactor{}

		result := []*fab.TransactionProposalResponse{
			{
				ProposalResponse: &pb.ProposalResponse{
					Response: &pb.Response{
						Payload: responseBytes1,
					},
				},
			},
			{
				ProposalResponse: &pb.ProposalResponse{
					Response: &pb.Response{
						Payload: responseBytes2,
					},
				},
			},
		}

		transactor.SendTransactionProposalReturns(result, nil)
		cs.TransactorReturns(transactor, nil)

		cp := &MockChannelProvider{}
		cp.ChannelServiceReturns(cs, nil)

		ctx.SetCustomChannelProvider(cp)

		resp, err := rc.LifecycleQueryCommittedCC(channel1, LifecycleQueryCommittedCCRequest{}, WithTargets(&fcmocks.MockPeer{}))
		require.Error(t, err)
		require.Contains(t, err.Error(), "responses from endorsers do not match")
		require.Empty(t, resp)
	})
}
