/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package resource

import (
	reqContext "context"
	"fmt"
	"testing"

	"github.com/golang/protobuf/proto"
	pb "github.com/hyperledger/fabric-protos-go/peer"
	lb "github.com/hyperledger/fabric-protos-go/peer/lifecycle"
	"github.com/stretchr/testify/require"

	"github.com/hyperledger/fabric-sdk-go/internal/github.com/hyperledger/fabric/common/policydsl"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/context"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/fab"
	contextImpl "github.com/hyperledger/fabric-sdk-go/pkg/context"
	"github.com/hyperledger/fabric-sdk-go/pkg/fab/mocks"
	"github.com/hyperledger/fabric-sdk-go/pkg/fab/txn"
)

func TestLifecycle_Install(t *testing.T) {
	lc := NewLifecycle()
	require.NotNil(t, lc)

	ctx := setupContext()

	reqCtx, cancel := contextImpl.NewRequest(ctx)
	defer cancel()

	t.Run("Success", func(t *testing.T) {
		resp, err := lc.Install(reqCtx, []byte("install package"), []fab.ProposalProcessor{&mocks.MockPeer{}})
		require.NoError(t, err)
		require.NotEmpty(t, resp)
	})

	t.Run("No package", func(t *testing.T) {
		resp, err := lc.Install(reqCtx, nil, []fab.ProposalProcessor{&mocks.MockPeer{}})
		require.EqualError(t, err, "chaincode package is required")
		require.Empty(t, resp)
	})

	t.Run("No targets", func(t *testing.T) {
		resp, err := lc.Install(reqCtx, []byte("install package"), nil)
		require.EqualError(t, err, "targets is required")
		require.Empty(t, resp)
	})

	t.Run("Marshal error", func(t *testing.T) {
		errExpected := fmt.Errorf("injected marshal error")

		lc := NewLifecycle()
		require.NotNil(t, lc)

		lc.protoMarshal = func(pb proto.Message) ([]byte, error) { return nil, errExpected }

		resp, err := lc.Install(reqCtx, []byte("install package"), []fab.ProposalProcessor{&mocks.MockPeer{}})
		require.Error(t, err)
		require.Contains(t, err.Error(), errExpected.Error())
		require.Empty(t, resp)
	})

	t.Run("Unmarshal error", func(t *testing.T) {
		errExpected := fmt.Errorf("injected unmarshal error")

		lc := NewLifecycle()
		require.NotNil(t, lc)

		lc.protoUnmarshal = func(buf []byte, pb proto.Message) error { return errExpected }

		resp, err := lc.Install(reqCtx, []byte("install package"), []fab.ProposalProcessor{&mocks.MockPeer{}})
		require.Error(t, err)
		require.Contains(t, err.Error(), errExpected.Error())
		require.Empty(t, resp)
	})

	t.Run("Context error", func(t *testing.T) {
		lc := NewLifecycle()
		require.NotNil(t, lc)

		lc.newContext = func(ctx reqContext.Context) (context.Client, bool) { return nil, false }

		resp, err := lc.Install(reqCtx, []byte("install package"), []fab.ProposalProcessor{&mocks.MockPeer{}})
		require.EqualError(t, err, "failed get client context from reqContext for txn header")
		require.Empty(t, resp)
	})

	t.Run("Txn Header error", func(t *testing.T) {
		errExpected := fmt.Errorf("injected Txn Header error")

		lc := NewLifecycle()
		require.NotNil(t, lc)

		lc.newTxnHeader = func(ctx context.Client, channelID string, opts ...fab.TxnHeaderOpt) (*txn.TransactionHeader, error) {
			return nil, errExpected
		}

		resp, err := lc.Install(reqCtx, []byte("install package"), []fab.ProposalProcessor{&mocks.MockPeer{}})
		require.Error(t, err)
		require.Contains(t, err.Error(), errExpected.Error())
		require.Empty(t, resp)
	})
}

func TestLifecycle_GetInstalledPackage(t *testing.T) {
	lc := NewLifecycle()
	require.NotNil(t, lc)

	ctx := setupContext()

	reqCtx, cancel := contextImpl.NewRequest(ctx)
	defer cancel()

	t.Run("Success", func(t *testing.T) {
		resp, err := lc.GetInstalledPackage(reqCtx, "packageid", &mocks.MockPeer{})
		require.NoError(t, err)
		require.Empty(t, resp)
	})

	t.Run("Marshal error", func(t *testing.T) {
		errExpected := fmt.Errorf("injected marshal error")

		lc := NewLifecycle()
		require.NotNil(t, lc)

		lc.protoMarshal = func(pb proto.Message) ([]byte, error) { return nil, errExpected }

		resp, err := lc.GetInstalledPackage(reqCtx, "packageid", &mocks.MockPeer{})
		require.Error(t, err)
		require.Contains(t, err.Error(), errExpected.Error())
		require.Empty(t, resp)
	})

	t.Run("Unmarshal error", func(t *testing.T) {
		errExpected := fmt.Errorf("injected unmarshal error")

		lc := NewLifecycle()
		require.NotNil(t, lc)

		lc.protoUnmarshal = func(buf []byte, pb proto.Message) error { return errExpected }

		resp, err := lc.GetInstalledPackage(reqCtx, "packageid", &mocks.MockPeer{})
		require.Error(t, err)
		require.Contains(t, err.Error(), errExpected.Error())
		require.Empty(t, resp)
	})

	t.Run("Context error", func(t *testing.T) {
		lc := NewLifecycle()
		require.NotNil(t, lc)

		lc.newContext = func(ctx reqContext.Context) (context.Client, bool) { return nil, false }

		resp, err := lc.GetInstalledPackage(reqCtx, "packageid", &mocks.MockPeer{})
		require.EqualError(t, err, "failed get client context from reqContext for txn header")
		require.Empty(t, resp)
	})

	t.Run("Txn Header error", func(t *testing.T) {
		errExpected := fmt.Errorf("injected Txn Header error")

		lc := NewLifecycle()
		require.NotNil(t, lc)

		lc.newTxnHeader = func(ctx context.Client, channelID string, opts ...fab.TxnHeaderOpt) (*txn.TransactionHeader, error) {
			return nil, errExpected
		}

		resp, err := lc.GetInstalledPackage(reqCtx, "packageid", &mocks.MockPeer{})
		require.Error(t, err)
		require.Contains(t, err.Error(), errExpected.Error())
		require.Empty(t, resp)
	})
}

func TestLifecycle_QueryInstalled(t *testing.T) {
	lc := NewLifecycle()
	require.NotNil(t, lc)

	ctx := setupContext()

	reqCtx, cancel := contextImpl.NewRequest(ctx)
	defer cancel()

	const packageID = "pkg1"
	const label = "label1"
	const cc1 = "cc1"
	const v1 = "v1"
	const channel1 = "channel1"

	result := &lb.QueryInstalledChaincodesResult{
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

	payload, err := proto.Marshal(result)
	require.NoError(t, err)

	t.Run("Success", func(t *testing.T) {
		resp, err := lc.QueryInstalled(reqCtx, &mocks.MockPeer{Payload: payload})
		require.NoError(t, err)
		require.NotNil(t, resp)
		require.Len(t, resp.InstalledChaincodes, 1)
		require.Equal(t, packageID, resp.InstalledChaincodes[0].PackageID)
		require.Equal(t, label, resp.InstalledChaincodes[0].Label)
		require.Len(t, resp.InstalledChaincodes[0].References, 1)

		references, ok := resp.InstalledChaincodes[0].References[channel1]
		require.True(t, ok)
		require.Len(t, references, 1)
		require.Equal(t, cc1, references[0].Name)
		require.Equal(t, v1, references[0].Version)
	})

	t.Run("Marshal error", func(t *testing.T) {
		errExpected := fmt.Errorf("injected marshal error")

		lc := NewLifecycle()
		require.NotNil(t, lc)

		lc.protoMarshal = func(pb proto.Message) ([]byte, error) { return nil, errExpected }

		resp, err := lc.QueryInstalled(reqCtx, &mocks.MockPeer{Payload: payload})
		require.Error(t, err)
		require.Contains(t, err.Error(), errExpected.Error())
		require.Empty(t, resp)
	})

	t.Run("Unmarshal error", func(t *testing.T) {
		errExpected := fmt.Errorf("injected unmarshal error")

		lc := NewLifecycle()
		require.NotNil(t, lc)

		lc.protoUnmarshal = func(buf []byte, pb proto.Message) error { return errExpected }

		resp, err := lc.QueryInstalled(reqCtx, &mocks.MockPeer{Payload: payload})
		require.Error(t, err)
		require.Contains(t, err.Error(), errExpected.Error())
		require.Empty(t, resp)
	})

	t.Run("Context error", func(t *testing.T) {
		lc := NewLifecycle()
		require.NotNil(t, lc)

		lc.newContext = func(ctx reqContext.Context) (context.Client, bool) { return nil, false }

		resp, err := lc.QueryInstalled(reqCtx, &mocks.MockPeer{Payload: payload})
		require.EqualError(t, err, "failed get client context from reqContext for txn header")
		require.Empty(t, resp)
	})

	t.Run("Txn Header error", func(t *testing.T) {
		errExpected := fmt.Errorf("injected Txn Header error")

		lc := NewLifecycle()
		require.NotNil(t, lc)

		lc.newTxnHeader = func(ctx context.Client, channelID string, opts ...fab.TxnHeaderOpt) (*txn.TransactionHeader, error) {
			return nil, errExpected
		}

		resp, err := lc.QueryInstalled(reqCtx, &mocks.MockPeer{Payload: payload})
		require.Error(t, err)
		require.Contains(t, err.Error(), errExpected.Error())
		require.Empty(t, resp)
	})
}

func TestLifecycle_CreateApproveProposal(t *testing.T) {
	lc := NewLifecycle()
	require.NotNil(t, lc)

	t.Run("With package ID -> success", func(t *testing.T) {
		req := &ApproveChaincodeRequest{
			Name:              "cc1",
			Version:           "v1",
			PackageID:         "pkg1",
			Sequence:          1,
			EndorsementPlugin: "eplugin",
			ValidationPlugin:  "vplugin",
			SignaturePolicy:   policydsl.AcceptAllPolicy,
			CollectionConfig:  []*pb.CollectionConfig{},
			InitRequired:      true,
		}

		p, err := lc.CreateApproveProposal(&mocks.MockTransactionHeader{}, req)
		require.NoError(t, err)
		require.NotNil(t, p)
		require.NotNil(t, p.Proposal)
	})

	t.Run("No package ID -> success", func(t *testing.T) {
		req := &ApproveChaincodeRequest{
			Name:              "cc1",
			Version:           "v1",
			PackageID:         "",
			Sequence:          1,
			EndorsementPlugin: "eplugin",
			ValidationPlugin:  "vplugin",
			SignaturePolicy:   policydsl.AcceptAllPolicy,
			CollectionConfig:  []*pb.CollectionConfig{},
			InitRequired:      true,
		}

		p, err := lc.CreateApproveProposal(&mocks.MockTransactionHeader{}, req)
		require.NoError(t, err)
		require.NotNil(t, p)
		require.NotNil(t, p.Proposal)
	})

	t.Run("No policy -> success", func(t *testing.T) {
		req := &ApproveChaincodeRequest{
			Name:              "cc1",
			Version:           "v1",
			Sequence:          1,
			EndorsementPlugin: "eplugin",
			ValidationPlugin:  "vplugin",
			CollectionConfig:  []*pb.CollectionConfig{},
			InitRequired:      true,
		}

		p, err := lc.CreateApproveProposal(&mocks.MockTransactionHeader{}, req)
		require.NoError(t, err)
		require.NotNil(t, p)
		require.NotNil(t, p.Proposal)
	})

	t.Run("Channel config policy -> success", func(t *testing.T) {
		req := &ApproveChaincodeRequest{
			Name:                "cc1",
			Version:             "v1",
			Sequence:            1,
			EndorsementPlugin:   "eplugin",
			ValidationPlugin:    "vplugin",
			ChannelConfigPolicy: "policy",
			CollectionConfig:    []*pb.CollectionConfig{},
			InitRequired:        true,
		}

		p, err := lc.CreateApproveProposal(&mocks.MockTransactionHeader{}, req)
		require.NoError(t, err)
		require.NotNil(t, p)
		require.NotNil(t, p.Proposal)
	})

	t.Run("Both signature policy and channel config policy specified -> error", func(t *testing.T) {
		req := &ApproveChaincodeRequest{
			Name:                "cc1",
			Version:             "v1",
			PackageID:           "",
			Sequence:            1,
			EndorsementPlugin:   "eplugin",
			ValidationPlugin:    "vplugin",
			SignaturePolicy:     policydsl.AcceptAllPolicy,
			ChannelConfigPolicy: "policy",
			CollectionConfig:    []*pb.CollectionConfig{},
			InitRequired:        true,
		}

		p, err := lc.CreateApproveProposal(&mocks.MockTransactionHeader{}, req)
		require.Error(t, err)
		require.Contains(t, err.Error(), "cannot specify both signature policy and channel config policy")
		require.Nil(t, p)
	})

	t.Run("Marshal -> error", func(t *testing.T) {
		errExpected := fmt.Errorf("injected marshal error")

		lc := NewLifecycle()
		require.NotNil(t, lc)

		lc.protoMarshal = func(pb proto.Message) ([]byte, error) { return nil, errExpected }

		req := &ApproveChaincodeRequest{
			Name:              "cc1",
			Version:           "v1",
			PackageID:         "",
			Sequence:          1,
			EndorsementPlugin: "eplugin",
			ValidationPlugin:  "vplugin",
			SignaturePolicy:   policydsl.AcceptAllPolicy,
			CollectionConfig:  []*pb.CollectionConfig{},
			InitRequired:      true,
		}

		p, err := lc.CreateApproveProposal(&mocks.MockTransactionHeader{}, req)
		require.Error(t, err)
		require.Contains(t, err.Error(), errExpected.Error())
		require.Nil(t, p)
	})
}

func TestLifecycle_QueryApproved(t *testing.T) {
	lc := NewLifecycle()
	require.NotNil(t, lc)

	ctx := setupContext()

	reqCtx, cancel := contextImpl.NewRequest(ctx)
	defer cancel()

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

	result := &lb.QueryApprovedChaincodeDefinitionResult{
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

	payload, err := proto.Marshal(result)
	require.NoError(t, err)

	req := &QueryApprovedChaincodeRequest{
		Name:     cc1,
		Sequence: 1,
	}

	t.Run("With signature policy -> Success", func(t *testing.T) {
		resp, err := lc.QueryApproved(reqCtx, channel1, req, &mocks.MockPeer{Payload: payload})
		require.NoError(t, err)
		require.NotNil(t, resp)
		require.NotNil(t, resp.ApprovedChaincode)
		require.Equal(t, cc1, resp.ApprovedChaincode.Name)
		require.Equal(t, v1, resp.ApprovedChaincode.Version)
	})

	t.Run("With channel config policy -> Success", func(t *testing.T) {
		applicationPolicy := &pb.ApplicationPolicy{
			Type: &pb.ApplicationPolicy_ChannelConfigPolicyReference{
				ChannelConfigPolicyReference: "channel policy",
			},
		}

		policyBytes, err := proto.Marshal(applicationPolicy)
		require.NoError(t, err)

		result := &lb.QueryApprovedChaincodeDefinitionResult{
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

		payload, err := proto.Marshal(result)
		require.NoError(t, err)

		resp, err := lc.QueryApproved(reqCtx, channel1, req, &mocks.MockPeer{Payload: payload})
		require.NoError(t, err)
		require.NotNil(t, resp)
		require.NotNil(t, resp.ApprovedChaincode)
		require.Equal(t, cc1, resp.ApprovedChaincode.Name)
		require.Equal(t, v1, resp.ApprovedChaincode.Version)
	})

	t.Run("With unsupported policy -> Error", func(t *testing.T) {
		policyBytes, err := proto.Marshal(&pb.ApplicationPolicy{})
		require.NoError(t, err)

		result := &lb.QueryApprovedChaincodeDefinitionResult{
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

		payload, err := proto.Marshal(result)
		require.NoError(t, err)

		resp, err := lc.QueryApproved(reqCtx, channel1, req, &mocks.MockPeer{Payload: payload})
		require.Error(t, err)
		require.Contains(t, err.Error(), "unsupported policy")
		require.Nil(t, resp)
	})

	t.Run("Marshal error", func(t *testing.T) {
		errExpected := fmt.Errorf("injected marshal error")

		lc := NewLifecycle()
		require.NotNil(t, lc)

		lc.protoMarshal = func(pb proto.Message) ([]byte, error) { return nil, errExpected }

		resp, err := lc.QueryApproved(reqCtx, channel1, req, &mocks.MockPeer{Payload: payload})
		require.Error(t, err)
		require.Contains(t, err.Error(), errExpected.Error())
		require.Empty(t, resp)
	})

	t.Run("Unmarshal error", func(t *testing.T) {
		errExpected := fmt.Errorf("injected unmarshal error")

		lc := NewLifecycle()
		require.NotNil(t, lc)

		lc.protoUnmarshal = func(buf []byte, pb proto.Message) error { return errExpected }

		resp, err := lc.QueryApproved(reqCtx, channel1, req, &mocks.MockPeer{Payload: payload})
		require.Error(t, err)
		require.Contains(t, err.Error(), errExpected.Error())
		require.Empty(t, resp)
	})

	t.Run("Context error", func(t *testing.T) {
		lc := NewLifecycle()
		require.NotNil(t, lc)

		lc.newContext = func(ctx reqContext.Context) (context.Client, bool) { return nil, false }

		resp, err := lc.QueryApproved(reqCtx, channel1, req, &mocks.MockPeer{Payload: payload})
		require.EqualError(t, err, "failed get client context from reqContext for txn header")
		require.Empty(t, resp)
	})

	t.Run("Txn Header error", func(t *testing.T) {
		errExpected := fmt.Errorf("injected Txn Header error")

		lc := NewLifecycle()
		require.NotNil(t, lc)

		lc.newTxnHeader = func(ctx context.Client, channelID string, opts ...fab.TxnHeaderOpt) (*txn.TransactionHeader, error) {
			return nil, errExpected
		}

		resp, err := lc.QueryApproved(reqCtx, channel1, req, &mocks.MockPeer{Payload: payload})
		require.Error(t, err)
		require.Contains(t, err.Error(), errExpected.Error())
		require.Empty(t, resp)
	})
}

func TestLifecycle_CreateCheckCommitReadinessProposal(t *testing.T) {
	lc := NewLifecycle()
	require.NotNil(t, lc)

	t.Run("With package ID -> success", func(t *testing.T) {
		req := &CheckChaincodeCommitReadinessRequest{
			Name:              "cc1",
			Version:           "v1",
			Sequence:          1,
			EndorsementPlugin: "eplugin",
			ValidationPlugin:  "vplugin",
			SignaturePolicy:   policydsl.AcceptAllPolicy,
			CollectionConfig:  []*pb.CollectionConfig{},
			InitRequired:      true,
		}

		p, err := lc.CreateCheckCommitReadinessProposal(&mocks.MockTransactionHeader{}, req)
		require.NoError(t, err)
		require.NotNil(t, p)
		require.NotNil(t, p.Proposal)
	})

	t.Run("No policy -> success", func(t *testing.T) {
		req := &CheckChaincodeCommitReadinessRequest{
			Name:              "cc1",
			Version:           "v1",
			Sequence:          1,
			EndorsementPlugin: "eplugin",
			ValidationPlugin:  "vplugin",
			CollectionConfig:  []*pb.CollectionConfig{},
			InitRequired:      true,
		}

		p, err := lc.CreateCheckCommitReadinessProposal(&mocks.MockTransactionHeader{}, req)
		require.NoError(t, err)
		require.NotNil(t, p)
		require.NotNil(t, p.Proposal)
	})

	t.Run("Channel config policy -> success", func(t *testing.T) {
		req := &CheckChaincodeCommitReadinessRequest{
			Name:                "cc1",
			Version:             "v1",
			Sequence:            1,
			EndorsementPlugin:   "eplugin",
			ValidationPlugin:    "vplugin",
			ChannelConfigPolicy: "policy",
			CollectionConfig:    []*pb.CollectionConfig{},
			InitRequired:        true,
		}

		p, err := lc.CreateCheckCommitReadinessProposal(&mocks.MockTransactionHeader{}, req)
		require.NoError(t, err)
		require.NotNil(t, p)
		require.NotNil(t, p.Proposal)
	})

	t.Run("Both signature policy and channel config policy specified -> error", func(t *testing.T) {
		req := &CheckChaincodeCommitReadinessRequest{
			Name:                "cc1",
			Version:             "v1",
			Sequence:            1,
			EndorsementPlugin:   "eplugin",
			ValidationPlugin:    "vplugin",
			SignaturePolicy:     policydsl.AcceptAllPolicy,
			ChannelConfigPolicy: "policy",
			CollectionConfig:    []*pb.CollectionConfig{},
			InitRequired:        true,
		}

		p, err := lc.CreateCheckCommitReadinessProposal(&mocks.MockTransactionHeader{}, req)
		require.Error(t, err)
		require.Contains(t, err.Error(), "cannot specify both signature policy and channel config policy")
		require.Nil(t, p)
	})

	t.Run("Marshal -> error", func(t *testing.T) {
		errExpected := fmt.Errorf("injected marshal error")

		lc := NewLifecycle()
		require.NotNil(t, lc)

		lc.protoMarshal = func(pb proto.Message) ([]byte, error) { return nil, errExpected }

		req := &CheckChaincodeCommitReadinessRequest{
			Name:              "cc1",
			Version:           "v1",
			Sequence:          1,
			EndorsementPlugin: "eplugin",
			ValidationPlugin:  "vplugin",
			SignaturePolicy:   policydsl.AcceptAllPolicy,
			CollectionConfig:  []*pb.CollectionConfig{},
			InitRequired:      true,
		}

		p, err := lc.CreateCheckCommitReadinessProposal(&mocks.MockTransactionHeader{}, req)
		require.Error(t, err)
		require.Contains(t, err.Error(), errExpected.Error())
		require.Nil(t, p)
	})
}

func TestLifecycle_CreateCommitProposal(t *testing.T) {
	lc := NewLifecycle()
	require.NotNil(t, lc)

	t.Run("With package ID -> success", func(t *testing.T) {
		req := &CommitChaincodeRequest{
			Name:              "cc1",
			Version:           "v1",
			Sequence:          1,
			EndorsementPlugin: "eplugin",
			ValidationPlugin:  "vplugin",
			SignaturePolicy:   policydsl.AcceptAllPolicy,
			CollectionConfig:  []*pb.CollectionConfig{},
			InitRequired:      true,
		}

		p, err := lc.CreateCommitProposal(&mocks.MockTransactionHeader{}, req)
		require.NoError(t, err)
		require.NotNil(t, p)
		require.NotNil(t, p.Proposal)
	})

	t.Run("No package ID -> success", func(t *testing.T) {
		req := &CommitChaincodeRequest{
			Name:              "cc1",
			Version:           "v1",
			Sequence:          1,
			EndorsementPlugin: "eplugin",
			ValidationPlugin:  "vplugin",
			SignaturePolicy:   policydsl.AcceptAllPolicy,
			CollectionConfig:  []*pb.CollectionConfig{},
			InitRequired:      true,
		}

		p, err := lc.CreateCommitProposal(&mocks.MockTransactionHeader{}, req)
		require.NoError(t, err)
		require.NotNil(t, p)
		require.NotNil(t, p.Proposal)
	})

	t.Run("No policy -> success", func(t *testing.T) {
		req := &CommitChaincodeRequest{
			Name:              "cc1",
			Version:           "v1",
			Sequence:          1,
			EndorsementPlugin: "eplugin",
			ValidationPlugin:  "vplugin",
			CollectionConfig:  []*pb.CollectionConfig{},
			InitRequired:      true,
		}

		p, err := lc.CreateCommitProposal(&mocks.MockTransactionHeader{}, req)
		require.NoError(t, err)
		require.NotNil(t, p)
		require.NotNil(t, p.Proposal)
	})

	t.Run("Channel config policy -> success", func(t *testing.T) {
		req := &CommitChaincodeRequest{
			Name:                "cc1",
			Version:             "v1",
			Sequence:            1,
			EndorsementPlugin:   "eplugin",
			ValidationPlugin:    "vplugin",
			ChannelConfigPolicy: "policy",
			CollectionConfig:    []*pb.CollectionConfig{},
			InitRequired:        true,
		}

		p, err := lc.CreateCommitProposal(&mocks.MockTransactionHeader{}, req)
		require.NoError(t, err)
		require.NotNil(t, p)
		require.NotNil(t, p.Proposal)
	})

	t.Run("Both signature policy and channel config policy specified -> error", func(t *testing.T) {
		req := &CommitChaincodeRequest{
			Name:                "cc1",
			Version:             "v1",
			Sequence:            1,
			EndorsementPlugin:   "eplugin",
			ValidationPlugin:    "vplugin",
			SignaturePolicy:     policydsl.AcceptAllPolicy,
			ChannelConfigPolicy: "policy",
			CollectionConfig:    []*pb.CollectionConfig{},
			InitRequired:        true,
		}

		p, err := lc.CreateCommitProposal(&mocks.MockTransactionHeader{}, req)
		require.Error(t, err)
		require.Contains(t, err.Error(), "cannot specify both signature policy and channel config policy")
		require.Nil(t, p)
	})

	t.Run("Marshal -> error", func(t *testing.T) {
		errExpected := fmt.Errorf("injected marshal error")

		lc := NewLifecycle()
		require.NotNil(t, lc)

		lc.protoMarshal = func(pb proto.Message) ([]byte, error) { return nil, errExpected }

		req := &CommitChaincodeRequest{
			Name:              "cc1",
			Version:           "v1",
			Sequence:          1,
			EndorsementPlugin: "eplugin",
			ValidationPlugin:  "vplugin",
			SignaturePolicy:   policydsl.AcceptAllPolicy,
			CollectionConfig:  []*pb.CollectionConfig{},
			InitRequired:      true,
		}

		p, err := lc.CreateCommitProposal(&mocks.MockTransactionHeader{}, req)
		require.Error(t, err)
		require.Contains(t, err.Error(), errExpected.Error())
		require.Nil(t, p)
	})
}

func TestLifecycle_CreateQueryCommittedProposal(t *testing.T) {
	lc := NewLifecycle()
	require.NotNil(t, lc)

	t.Run("With name -> success", func(t *testing.T) {
		req := &QueryCommittedChaincodesRequest{
			Name: "cc1",
		}

		p, err := lc.CreateQueryCommittedProposal(&mocks.MockTransactionHeader{}, req)
		require.NoError(t, err)
		require.NotNil(t, p)
		require.NotNil(t, p.Proposal)
	})

	t.Run("No name -> success", func(t *testing.T) {
		req := &QueryCommittedChaincodesRequest{}

		p, err := lc.CreateQueryCommittedProposal(&mocks.MockTransactionHeader{}, req)
		require.NoError(t, err)
		require.NotNil(t, p)
		require.NotNil(t, p.Proposal)
	})

	t.Run("Marshal -> error", func(t *testing.T) {
		errExpected := fmt.Errorf("injected marshal error")

		lc := NewLifecycle()
		require.NotNil(t, lc)

		lc.protoMarshal = func(pb proto.Message) ([]byte, error) { return nil, errExpected }

		req := &QueryCommittedChaincodesRequest{}

		p, err := lc.CreateQueryCommittedProposal(&mocks.MockTransactionHeader{}, req)
		require.Error(t, err)
		require.Contains(t, err.Error(), errExpected.Error())
		require.Nil(t, p)
	})
}
