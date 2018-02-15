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
	ProcessTransactionProposal(ProcessProposalRequest) (*TransactionProposalResponse, error)
}

// ProposalSender provides the ability for a transaction proposal to be created and sent.
//
// TODO: CreateChaincodeInvokeProposal should be refactored as it is mostly a factory method.
type ProposalSender interface {
	CreateChaincodeInvokeProposal(ChaincodeInvokeRequest) (*TransactionProposal, error)
	SendTransactionProposal(*TransactionProposal, []ProposalProcessor) ([]*TransactionProposalResponse, error)
}

// TransactionID contains the ID of a Fabric Transaction Proposal
type TransactionID struct {
	ID    string
	Nonce []byte
}

// ChaincodeInvokeRequest contains the parameters for sending a transaction proposal.
//
// Deprecated: this struct has been replaced by ChaincodeInvokeProposal.
type ChaincodeInvokeRequest struct {
	Targets      []ProposalProcessor // Deprecated: this parameter is ignored in the new codes and will be removed shortly.
	ChaincodeID  string
	TransientMap map[string][]byte
	Fcn          string
	Args         [][]byte
}

// TransactionProposal contains a marashalled transaction proposal.
type TransactionProposal struct {
	TxnID TransactionID // TODO: remove?
	*pb.Proposal
}

// ProcessProposalRequest requests simulation of a proposed transaction from transaction processors.
type ProcessProposalRequest struct {
	SignedProposal *pb.SignedProposal
}

// TransactionProposalResponse respresents the result of transaction proposal processing.
type TransactionProposalResponse struct {
	Endorser string
	Status   int32
	*pb.ProposalResponse
}
