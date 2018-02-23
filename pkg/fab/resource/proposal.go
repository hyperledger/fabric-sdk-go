/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package resource

import (
	"time"

	"github.com/golang/protobuf/proto"
	"github.com/golang/protobuf/ptypes"
	"github.com/pkg/errors"

	"github.com/hyperledger/fabric-sdk-go/internal/github.com/hyperledger/fabric/common/crypto"
	fcutils "github.com/hyperledger/fabric-sdk-go/internal/github.com/hyperledger/fabric/common/util"
	"github.com/hyperledger/fabric-sdk-go/pkg/context"
	"github.com/hyperledger/fabric-sdk-go/pkg/context/api/fab"
	"github.com/hyperledger/fabric-sdk-go/pkg/fab/txn"
	"github.com/hyperledger/fabric-sdk-go/third_party/github.com/hyperledger/fabric/protos/common"
	pb "github.com/hyperledger/fabric-sdk-go/third_party/github.com/hyperledger/fabric/protos/peer"
	protos_utils "github.com/hyperledger/fabric-sdk-go/third_party/github.com/hyperledger/fabric/protos/utils"
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
func CreateChaincodeInstallProposal(txid fab.TransactionID, request ChaincodeInstallRequest) (*fab.TransactionProposal, error) {

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

	cir := fab.ChaincodeInvokeRequest{
		ChaincodeID: "lscc",
		Fcn:         "install",
		Args:        args,
	}

	return txn.CreateChaincodeInvokeProposal(txid, "", cir)
}

// CreateConfigSignature creates a ConfigSignature for the current context.
func CreateConfigSignature(ctx context.Context, config []byte) (*common.ConfigSignature, error) {

	creator, err := ctx.Identity()
	if err != nil {
		return nil, errors.WithMessage(err, "failed to get user context's identity")
	}

	// generate a random nonce
	nonce, err := crypto.GetRandomNonce()
	if err != nil {
		return nil, errors.WithMessage(err, "nonce creation failed")
	}

	// signature is across a signature header and the config update
	signatureHeader := &common.SignatureHeader{
		Creator: creator,
		Nonce:   nonce,
	}
	signatureHeaderBytes, err := proto.Marshal(signatureHeader)
	if err != nil {
		return nil, errors.Wrap(err, "marshal signatureHeader failed")
	}

	// get all the bytes to be signed together, then sign
	signingBytes := fcutils.ConcatenateBytes(signatureHeaderBytes, config)
	signingMgr := ctx.SigningManager()
	signature, err := signingMgr.Sign(signingBytes, ctx.PrivateKey())
	if err != nil {
		return nil, errors.WithMessage(err, "signing of channel config failed")
	}

	// build the return object
	configSignature := common.ConfigSignature{
		SignatureHeader: signatureHeaderBytes,
		Signature:       signature,
	}
	return &configSignature, nil
}

// ExtractChannelConfig extracts the protobuf 'ConfigUpdate' object out of the 'ConfigEnvelope'.
func ExtractChannelConfig(configEnvelope []byte) ([]byte, error) {
	logger.Debug("extractConfigUpdate - start")

	envelope := &common.Envelope{}
	err := proto.Unmarshal(configEnvelope, envelope)
	if err != nil {
		return nil, errors.Wrap(err, "unmarshal config envelope failed")
	}

	payload := &common.Payload{}
	err = proto.Unmarshal(envelope.Payload, payload)
	if err != nil {
		return nil, errors.Wrap(err, "unmarshal envelope payload failed")
	}

	configUpdateEnvelope := &common.ConfigUpdateEnvelope{}
	err = proto.Unmarshal(payload.Data, configUpdateEnvelope)
	if err != nil {
		return nil, errors.Wrap(err, "unmarshal config update envelope")
	}

	return configUpdateEnvelope.ConfigUpdate, nil
}
