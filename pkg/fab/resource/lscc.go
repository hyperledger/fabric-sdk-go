/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package resource

import (
	pb "github.com/hyperledger/fabric-protos-go/peer"
	"github.com/hyperledger/fabric-sdk-go/internal/github.com/hyperledger/fabric/protoutil"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/fab"
	"github.com/hyperledger/fabric-sdk-go/pkg/fab/txn"
	"github.com/pkg/errors"
)

const (
	lscc                    = "lscc"
	lsccInstall             = "install"
	lsccInstalledChaincodes = "getinstalledchaincodes"
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
func CreateChaincodeInstallProposal(txh fab.TransactionHeader, request ChaincodeInstallRequest) (*fab.TransactionProposal, error) {
	cir, err := createInstallInvokeRequest(request)
	if err != nil {
		return nil, errors.WithMessage(err, "creating lscc install invocation request failed")
	}

	return txn.CreateChaincodeInvokeProposal(txh, cir)
}

func createInstallInvokeRequest(request ChaincodeInstallRequest) (fab.ChaincodeInvokeRequest, error) {
	// Generate arguments for install
	args := [][]byte{}

	ccds := &pb.ChaincodeDeploymentSpec{ChaincodeSpec: &pb.ChaincodeSpec{
		Type: request.Package.Type, ChaincodeId: &pb.ChaincodeID{Name: request.Name, Path: request.Path, Version: request.Version}},
		CodePackage: request.Package.Code}

	ccdsBytes, err := protoutil.Marshal(ccds)
	if err != nil {
		return fab.ChaincodeInvokeRequest{}, errors.WithMessage(err, "marshal of chaincode deployment spec failed")
	}
	args = append(args, ccdsBytes)

	cir := fab.ChaincodeInvokeRequest{
		ChaincodeID: lscc,
		Fcn:         lsccInstall,
		Args:        args,
	}
	return cir, nil
}

func createInstalledChaincodesInvokeRequest() fab.ChaincodeInvokeRequest {
	cir := fab.ChaincodeInvokeRequest{
		ChaincodeID: lscc,
		Fcn:         lsccInstalledChaincodes,
	}
	return cir
}
