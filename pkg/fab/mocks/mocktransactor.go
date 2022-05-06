/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package mocks

import (
	reqContex "context"
	"net/http"

	"github.com/golang/protobuf/proto"
	pb "github.com/hyperledger/fabric-protos-go/peer"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/fab"
)

// MockTransactor provides an implementation of Transactor that exposes all its context.
type MockTransactor struct {
	Ctx       reqContex.Context
	ChannelID string
	Orderers  []fab.Orderer
}

// CreateTransactionHeader creates a Transaction Header based on the current context.
func (t *MockTransactor) CreateTransactionHeader(opts ...fab.TxnHeaderOpt) (fab.TransactionHeader, error) {
	return &MockTransactionHeader{}, nil
}

// SendTransactionProposal sends a TransactionProposal to the target peers.
func (t *MockTransactor) SendTransactionProposal(proposal *fab.TransactionProposal, targets []fab.ProposalProcessor) ([]*fab.TransactionProposalResponse, error) {
	response := make([]*fab.TransactionProposalResponse, 1)
	txResponse := &pb.Response{Message: "success", Payload: []byte("abc"), Status: http.StatusOK}
	response[0] = &fab.TransactionProposalResponse{Endorser: "example.com", Status: 200,
		ProposalResponse: &pb.ProposalResponse{
			Response:    txResponse,
			Payload:     getProposalResponsePayloadBytes(txResponse),
			Endorsement: &pb.Endorsement{Endorser: []byte("example.com"), Signature: []byte("signature")},
		},
	}
	return response, nil
}

// CreateTransaction create a transaction with proposal response.
func (t *MockTransactor) CreateTransaction(request fab.TransactionRequest) (*fab.Transaction, error) {
	response := &fab.Transaction{
		Proposal: &fab.TransactionProposal{
			Proposal: &pb.Proposal{},
		},
		Transaction: &pb.Transaction{},
	}
	return response, nil
}

// SendTransaction send a transaction to the chain’s orderer service (one or more orderer endpoints) for consensus and committing to the ledger.
func (t *MockTransactor) SendTransaction(tx *fab.Transaction) (*fab.TransactionResponse, error) {
	response := &fab.TransactionResponse{
		Orderer: "example.com",
	}
	return response, nil
}

func getProposalResponsePayloadBytes(response *pb.Response) []byte {
	chaincodeAction := &pb.ChaincodeAction{
		Response: response,
	}
	chaincodeActionBytes, err := proto.Marshal(chaincodeAction)
	if err != nil {
		panic(err)
	}

	prp := &pb.ProposalResponsePayload{
		Extension: chaincodeActionBytes,
	}
	payloadBytes, err := proto.Marshal(prp)
	if err != nil {
		panic(err)
	}

	return payloadBytes
}
