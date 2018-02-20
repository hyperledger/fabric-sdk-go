/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package resource

import (
	"time"

	"github.com/golang/protobuf/ptypes"
	"github.com/hyperledger/fabric-sdk-go/pkg/fabric-client/txn"
	"github.com/pkg/errors"

	pb "github.com/hyperledger/fabric-sdk-go/third_party/github.com/hyperledger/fabric/protos/peer"
	protos_utils "github.com/hyperledger/fabric-sdk-go/third_party/github.com/hyperledger/fabric/protos/utils"

	fab "github.com/hyperledger/fabric-sdk-go/api/apifabclient"
)

// ChaincodeInstallRequest requests chaincode installation on the network
type ChaincodeInstallRequest struct {
	Name    string
	Path    string
	Version string
	Package *ChaincodePackage
}

// ChaincodePackage contains package type and bytes required to create CDS
type ChaincodePackage struct {
	Type pb.ChaincodeSpec_Type
	Code []byte
}

// CreateChaincodeInstallProposal creates an install chaincode proposal.
func CreateChaincodeInstallProposal(ctx fab.IdentityContext, request ChaincodeInstallRequest) (*fab.TransactionProposal, error) {

	// Generate arguments for install
	args := [][]byte{}
	timestamp := time.Now()
	ts, err := ptypes.TimestampProto(timestamp)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create timestamp in install proposal")
	}

	ccds := &pb.ChaincodeDeploymentSpec{ChaincodeSpec: &pb.ChaincodeSpec{
		Type: request.Package.Type, ChaincodeId: &pb.ChaincodeID{Name: request.Name, Path: request.Path, Version: request.Version}},
		CodePackage: request.Package.Code, EffectiveDate: ts}
	ccdsBytes, err := protos_utils.Marshal(ccds)
	if err != nil {
		return nil, errors.WithMessage(err, "marshal of chaincode deployment spec failed")
	}
	args = append(args, ccdsBytes)

	args = append(args, []byte("escc"))
	args = append(args, []byte("vscc"))

	cir := fab.ChaincodeInvokeRequest{
		ChaincodeID: "lscc",
		Fcn:         "install",
		Args:        args,
	}
	return txn.CreateChaincodeInvokeProposal(ctx, "", cir)
}
