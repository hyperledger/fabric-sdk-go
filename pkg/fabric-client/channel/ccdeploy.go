/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package channel

import (
	"github.com/golang/protobuf/proto"
	"github.com/pkg/errors"

	"github.com/hyperledger/fabric-sdk-go/third_party/github.com/hyperledger/fabric/protos/common"
	pb "github.com/hyperledger/fabric-sdk-go/third_party/github.com/hyperledger/fabric/protos/peer"
	protos_utils "github.com/hyperledger/fabric-sdk-go/third_party/github.com/hyperledger/fabric/protos/utils"

	fab "github.com/hyperledger/fabric-sdk-go/api/apifabclient"
)

// ChaincodeProposalType reflects transitions in the chaincode lifecycle
type ChaincodeProposalType int

// Define chaincode proposal types
const (
	InstantiateChaincode ChaincodeProposalType = iota
	UpgradeChaincode
)

// ChaincodeDeployRequest holds parameters for creating an instantiate or upgrade chaincode proposal.
type ChaincodeDeployRequest struct {
	Name       string
	Path       string
	Version    string
	Args       [][]byte
	Policy     *common.SignaturePolicyEnvelope
	CollConfig []*common.CollectionConfig
}

// CreateChaincodeDeployProposal creates an instantiate or upgrade chaincode proposal.
func CreateChaincodeDeployProposal(ctx fab.IdentityContext, deploy ChaincodeProposalType, channelID string, chaincode ChaincodeDeployRequest) (*fab.TransactionProposal, error) {

	ccds := &pb.ChaincodeDeploymentSpec{ChaincodeSpec: &pb.ChaincodeSpec{
		Type: pb.ChaincodeSpec_GOLANG, ChaincodeId: &pb.ChaincodeID{Name: chaincode.Name, Path: chaincode.Path, Version: chaincode.Version},
		Input: &pb.ChaincodeInput{Args: chaincode.Args}}}

	creator, err := ctx.Identity()
	if err != nil {
		return nil, errors.Wrap(err, "getting user context's identity failed")
	}
	chaincodePolicyBytes, err := protos_utils.Marshal(chaincode.Policy)
	if err != nil {
		return nil, err
	}
	var collConfigBytes []byte
	if chaincode.CollConfig != nil {
		var err error
		collConfigBytes, err = proto.Marshal(&common.CollectionConfigPackage{Config: chaincode.CollConfig})
		if err != nil {
			return nil, err
		}
	}

	var proposal *pb.Proposal
	var txID string

	switch deploy {

	case InstantiateChaincode:
		proposal, txID, err = protos_utils.CreateDeployProposalFromCDS(channelID, ccds, creator, chaincodePolicyBytes, []byte("escc"), []byte("vscc"), collConfigBytes)
		if err != nil {
			return nil, errors.Wrap(err, "create instantiate chaincode proposal failed")
		}
	case UpgradeChaincode:
		proposal, txID, err = protos_utils.CreateUpgradeProposalFromCDS(channelID, ccds, creator, chaincodePolicyBytes, []byte("escc"), []byte("vscc"))
		if err != nil {
			return nil, errors.Wrap(err, "create  upgrade chaincode proposal failed")
		}
	default:
		return nil, errors.Errorf("chaincode proposal type %d not supported", deploy)
	}

	txnID := fab.TransactionID{ID: txID} // Nonce is missing
	tp := fab.TransactionProposal{
		Proposal: proposal,
		TxnID:    txnID,
	}

	return &tp, err
}
