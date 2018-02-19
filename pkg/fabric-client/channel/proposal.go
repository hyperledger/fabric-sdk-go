/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package channel

import (
	"github.com/golang/protobuf/proto"
	"github.com/hyperledger/fabric-sdk-go/pkg/fabric-client/txn"
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

	// Generate arguments for deploy (channel, marshaled CCDS, marshaled chaincode policy, marshaled collection policy)
	args := [][]byte{}
	args = append(args, []byte(channelID))

	ccds := &pb.ChaincodeDeploymentSpec{ChaincodeSpec: &pb.ChaincodeSpec{
		Type: pb.ChaincodeSpec_GOLANG, ChaincodeId: &pb.ChaincodeID{Name: chaincode.Name, Path: chaincode.Path, Version: chaincode.Version},
		Input: &pb.ChaincodeInput{Args: chaincode.Args}}}
	ccdsBytes, err := protos_utils.Marshal(ccds)
	if err != nil {
		return nil, errors.WithMessage(err, "marshal of chaincode deployment spec failed")
	}
	args = append(args, ccdsBytes)

	chaincodePolicyBytes, err := protos_utils.Marshal(chaincode.Policy)
	if err != nil {
		return nil, errors.WithMessage(err, "marshal of chaincode policy failed")
	}
	args = append(args, chaincodePolicyBytes)

	args = append(args, []byte("escc"))
	args = append(args, []byte("vscc"))

	if chaincode.CollConfig != nil {
		var err error
		collConfigBytes, err := proto.Marshal(&common.CollectionConfigPackage{Config: chaincode.CollConfig})
		if err != nil {
			return nil, errors.WithMessage(err, "marshal of collection policy failed")
		}
		args = append(args, collConfigBytes)
	}

	// Fcn is deploy or upgrade
	fcn := ""
	switch deploy {
	case InstantiateChaincode:
		fcn = "deploy"
	case UpgradeChaincode:
		fcn = "upgrade"
	default:
		return nil, errors.WithMessage(err, "chaincode deployment type unknown")
	}

	cir := fab.ChaincodeInvokeRequest{
		ChaincodeID: "lscc",
		Fcn:         fcn,
		Args:        args,
	}
	return txn.CreateChaincodeInvokeProposal(ctx, channelID, cir)
}
