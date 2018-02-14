/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package apifabclient

import (
	pb "github.com/hyperledger/fabric-sdk-go/third_party/github.com/hyperledger/fabric/protos/peer"
)

// ProposalProcessor simulates transaction proposal, so that a client can submit the result for ordering.
type ProposalProcessor interface {
	ProcessTransactionProposal(proposal TransactionProposal) (TransactionProposalResponse, error)
}

// ProposalSender provides the ability for a transaction proposal to be created and sent.
type ProposalSender interface {
	SendTransactionProposal(ChaincodeInvokeRequest, []ProposalProcessor) ([]*TransactionProposalResponse, TransactionID, error)
}

// TransactionID contains the ID of a Fabric Transaction Proposal
type TransactionID struct {
	ID    string
	Nonce []byte
}

// ChaincodeInvokeRequest contains the parameters for sending a transaction proposal.
type ChaincodeInvokeRequest struct {
	Targets      []ProposalProcessor // TODO: remove
	ChaincodeID  string
	TransientMap map[string][]byte
	Fcn          string
	Args         [][]byte
}

// TransactionProposal requests simulation of a proposed transaction from transaction processors.
type TransactionProposal struct {
	TxnID TransactionID

	SignedProposal *pb.SignedProposal
	Proposal       *pb.Proposal
}

// TransactionProposalResponse respresents the result of transaction proposal processing.
type TransactionProposalResponse struct {
	Endorser string
	Status   int32

	Proposal         TransactionProposal
	ProposalResponse *pb.ProposalResponse
}
