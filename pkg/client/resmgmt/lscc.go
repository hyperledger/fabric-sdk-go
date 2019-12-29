/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package resmgmt

import (
	"github.com/golang/protobuf/proto"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/fab"
	"github.com/hyperledger/fabric-sdk-go/pkg/fab/txn"
	"github.com/pkg/errors"

	"github.com/hyperledger/fabric-protos-go/common"
	pb "github.com/hyperledger/fabric-protos-go/peer"
	"github.com/hyperledger/fabric-sdk-go/internal/github.com/hyperledger/fabric/protoutil"
)

const (
	lscc        = "lscc"
	lsccDeploy  = "deploy"
	lsccUpgrade = "upgrade"
	escc        = "escc"
	vscc        = "vscc"
)

// chaincodeProposalType reflects transitions in the chaincode lifecycle
type chaincodeProposalType int

// Define chaincode proposal types
const (
	InstantiateChaincode chaincodeProposalType = iota
	UpgradeChaincode
)

// chaincodeDeployRequest holds parameters for creating an instantiate or upgrade chaincode proposal.
type chaincodeDeployRequest struct {
	Name       string
	Path       string
	Version    string
	Lang       pb.ChaincodeSpec_Type
	Args       [][]byte
	Policy     *common.SignaturePolicyEnvelope
	CollConfig []*pb.CollectionConfig
}

// createChaincodeDeployProposal creates an instantiate or upgrade chaincode proposal.
func createChaincodeDeployProposal(txh fab.TransactionHeader, deploy chaincodeProposalType, channelID string, chaincode chaincodeDeployRequest) (*fab.TransactionProposal, error) {

	// Generate arguments for deploy (channel, marshaled CCDS, marshaled chaincode policy, marshaled collection policy)
	args := [][]byte{}
	args = append(args, []byte(channelID))

	ccds := &pb.ChaincodeDeploymentSpec{ChaincodeSpec: &pb.ChaincodeSpec{
		Type: chaincode.Lang, ChaincodeId: &pb.ChaincodeID{Name: chaincode.Name, Path: chaincode.Path, Version: chaincode.Version},
		Input: &pb.ChaincodeInput{Args: chaincode.Args}}}
	ccdsBytes, err := protoutil.Marshal(ccds)
	if err != nil {
		return nil, errors.WithMessage(err, "marshal of chaincode deployment spec failed")
	}
	args = append(args, ccdsBytes)

	chaincodePolicyBytes, err := protoutil.Marshal(chaincode.Policy)
	if err != nil {
		return nil, errors.WithMessage(err, "marshal of chaincode policy failed")
	}
	args = append(args, chaincodePolicyBytes)

	args = append(args, []byte(escc))
	args = append(args, []byte(vscc))

	if chaincode.CollConfig != nil {
		collConfigBytes, err := proto.Marshal(&pb.CollectionConfigPackage{Config: chaincode.CollConfig})
		if err != nil {
			return nil, errors.WithMessage(err, "marshal of collection policy failed")
		}
		args = append(args, collConfigBytes)
	}

	// Fcn is deploy or upgrade
	fcn := ""
	switch deploy {
	case InstantiateChaincode:
		fcn = lsccDeploy
	case UpgradeChaincode:
		fcn = lsccUpgrade
	default:
		return nil, errors.New("chaincode deployment type unknown")
	}

	cir := fab.ChaincodeInvokeRequest{
		ChaincodeID: lscc,
		Lang:        chaincode.Lang,
		Fcn:         fcn,
		Args:        args,
	}
	return txn.CreateChaincodeInvokeProposal(txh, cir)
}
