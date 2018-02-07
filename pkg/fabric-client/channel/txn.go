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
	"github.com/hyperledger/fabric-sdk-go/pkg/errors/status"
	"github.com/hyperledger/fabric-sdk-go/pkg/fabric-client/txn"
)

// CCProposalType reflects transitions in the chaincode lifecycle
type CCProposalType int

// Define chaincode proposal types
const (
	Instantiate CCProposalType = iota
	Upgrade
)

// CreateTransaction create a transaction with proposal response, following the endorsement policy.
func (c *Channel) CreateTransaction(resps []*fab.TransactionProposalResponse) (*fab.Transaction, error) {
	return txn.New(resps)
}

// SendTransaction send a transaction to the chain’s orderer service (one or more orderer endpoints) for consensus and committing to the ledger.
func (c *Channel) SendTransaction(tx *fab.Transaction) (*fab.TransactionResponse, error) {
	return txn.Send(c.clientContext, tx, c.Orderers())
}

// SendTransactionProposal sends the created proposal to peer for endorsement.
// TODO: return the entire request or just the txn ID?
func (c *Channel) SendTransactionProposal(request fab.ChaincodeInvokeRequest, targets []fab.ProposalProcessor) ([]*fab.TransactionProposalResponse, fab.TransactionID, error) {
	tp, err := txn.NewProposal(c.clientContext, c.name, request)
	if err != nil {
		return nil, fab.TransactionID{}, errors.WithMessage(err, "new transaction proposal failed")
	}

	targets, err = c.chaincodeInvokeRequestAddDefaultPeers(targets)
	if err != nil {
		return nil, fab.TransactionID{}, err
	}

	tpr, err := txn.SendProposal(tp, targets)
	if err != nil {
		return nil, fab.TransactionID{}, errors.WithMessage(err, "send transaction proposal failed")
	}

	return tpr, tp.TxnID, nil
}

// SendInstantiateProposal sends an instantiate proposal to one or more endorsing peers.
// chaincodeName: required - The name of the chain.
// args: optional - string Array arguments specific to the chaincode being instantiated
// chaincodePath: required - string of the path to the location of the source code of the chaincode
// chaincodeVersion: required - string of the version of the chaincode
// chaincodePolicy: required - chaincode signature policy
// collConfig: optional - private data collection configuration
func (c *Channel) SendInstantiateProposal(chaincodeName string,
	args [][]byte, chaincodePath string, chaincodeVersion string,
	chaincodePolicy *common.SignaturePolicyEnvelope,
	collConfig []*common.CollectionConfig, targets []fab.ProposalProcessor) ([]*fab.TransactionProposalResponse, fab.TransactionID, error) {

	return c.sendCCProposal(Instantiate, chaincodeName, args, chaincodePath, chaincodeVersion, chaincodePolicy, collConfig, targets)

}

// SendUpgradeProposal sends an upgrade proposal to one or more endorsing peers.
// chaincodeName: required - The name of the chain.
// args: optional - string Array arguments specific to the chaincode being upgraded
// chaincodePath: required - string of the path to the location of the source code of the chaincode
// chaincodeVersion: required - string of the version of the chaincode
func (c *Channel) SendUpgradeProposal(chaincodeName string,
	args [][]byte, chaincodePath string, chaincodeVersion string,
	chaincodePolicy *common.SignaturePolicyEnvelope, targets []fab.ProposalProcessor) ([]*fab.TransactionProposalResponse, fab.TransactionID, error) {

	return c.sendCCProposal(Upgrade, chaincodeName, args, chaincodePath, chaincodeVersion, chaincodePolicy, nil, targets)

}

func validateChaincodeInvokeRequest(request fab.ChaincodeInvokeRequest) error {
	if request.ChaincodeID == "" {
		return errors.New("ChaincodeID is required")
	}

	if request.Fcn == "" {
		return errors.New("Fcn is required")
	}
	return nil
}

func (c *Channel) chaincodeInvokeRequestAddDefaultPeers(targets []fab.ProposalProcessor) ([]fab.ProposalProcessor, error) {
	// Use default peers if targets are not specified.
	if targets == nil || len(targets) == 0 {
		if c.peers == nil || len(c.peers) == 0 {
			return nil, status.New(status.ClientStatus, status.NoPeersFound.ToInt32(),
				"targets were not specified and no peers have been configured", nil)
		}

		return c.txnProcessors(), nil
	}
	return targets, nil
}

// helper function that sends an instantiate or upgrade chaincode proposal to one or more endorsing peers
func (c *Channel) sendCCProposal(ccProposalType CCProposalType, chaincodeName string,
	args [][]byte, chaincodePath string, chaincodeVersion string,
	chaincodePolicy *common.SignaturePolicyEnvelope,
	collConfig []*common.CollectionConfig,
	targets []fab.ProposalProcessor) ([]*fab.TransactionProposalResponse, fab.TransactionID, error) {

	if chaincodeName == "" {
		return nil, fab.TransactionID{}, errors.New("chaincodeName is required")
	}
	if chaincodePath == "" {
		return nil, fab.TransactionID{}, errors.New("chaincodePath is required")
	}
	if chaincodeVersion == "" {
		return nil, fab.TransactionID{}, errors.New("chaincodeVersion is required")
	}
	if chaincodePolicy == nil {
		return nil, fab.TransactionID{}, errors.New("chaincodePolicy is required")
	}

	if targets == nil || len(targets) < 1 {
		return nil, fab.TransactionID{}, errors.New("missing peer objects for chaincode proposal")
	}

	ccds := &pb.ChaincodeDeploymentSpec{ChaincodeSpec: &pb.ChaincodeSpec{
		Type: pb.ChaincodeSpec_GOLANG, ChaincodeId: &pb.ChaincodeID{Name: chaincodeName, Path: chaincodePath, Version: chaincodeVersion},
		Input: &pb.ChaincodeInput{Args: args}}}

	creator, err := c.clientContext.Identity()
	if err != nil {
		return nil, fab.TransactionID{}, errors.Wrap(err, "getting user context's identity failed")
	}
	chaincodePolicyBytes, err := protos_utils.Marshal(chaincodePolicy)
	if err != nil {
		return nil, fab.TransactionID{}, err
	}
	var collConfigBytes []byte
	if collConfig != nil {
		var err error
		collConfigBytes, err = proto.Marshal(&common.CollectionConfigPackage{Config: collConfig})
		if err != nil {
			return nil, fab.TransactionID{}, err
		}
	}

	var proposal *pb.Proposal
	var txID string

	switch ccProposalType {

	case Instantiate:
		proposal, txID, err = protos_utils.CreateDeployProposalFromCDS(c.Name(), ccds, creator, chaincodePolicyBytes, []byte("escc"), []byte("vscc"), collConfigBytes)
		if err != nil {
			return nil, fab.TransactionID{}, errors.Wrap(err, "create instantiate chaincode proposal failed")
		}
	case Upgrade:
		proposal, txID, err = protos_utils.CreateUpgradeProposalFromCDS(c.Name(), ccds, creator, chaincodePolicyBytes, []byte("escc"), []byte("vscc"))
		if err != nil {
			return nil, fab.TransactionID{}, errors.Wrap(err, "create  upgrade chaincode proposal failed")
		}
	default:
		return nil, fab.TransactionID{}, errors.Errorf("chaincode proposal type %d not supported", ccProposalType)
	}

	signedProposal, err := txn.SignProposal(c.clientContext, proposal)
	if err != nil {
		return nil, fab.TransactionID{}, err
	}

	txnID := fab.TransactionID{ID: txID} // Nonce is missing

	transactionProposalResponse, err := txn.SendProposal(&fab.TransactionProposal{
		SignedProposal: signedProposal,
		Proposal:       proposal,
		TxnID:          txnID,
	}, targets)

	return transactionProposalResponse, txnID, err
}

// TODO: There should be a strategy for choosing processors.
func (c *Channel) txnProcessors() []fab.ProposalProcessor {
	return peersToTxnProcessors(c.Peers())
}

// peersToTxnProcessors converts a slice of Peers to a slice of ProposalProcessors
func peersToTxnProcessors(peers []fab.Peer) []fab.ProposalProcessor {
	tpp := make([]fab.ProposalProcessor, len(peers))

	for i := range peers {
		tpp[i] = peers[i]
	}
	return tpp
}
