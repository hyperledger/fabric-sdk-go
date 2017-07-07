/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package txnproc

import (
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/hyperledger/fabric-sdk-go/api/apitxn"
	"github.com/hyperledger/fabric-sdk-go/api/apitxn/mocks"
	pb "github.com/hyperledger/fabric/protos/peer"
)

func TestSendTransactionProposalToProcessors(t *testing.T) {

	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()
	proc := mock_apitxn.NewMockProposalProcessor(mockCtrl)

	tp := apitxn.TransactionProposal{
		SignedProposal: &pb.SignedProposal{},
	}
	tpr := apitxn.TransactionProposalResult{Endorser: "example.com", Status: 99, Proposal: tp, ProposalResponse: nil}
	proc.EXPECT().ProcessTransactionProposal(tp).Return(tpr, nil)
	targets := []apitxn.ProposalProcessor{proc}

	result, err := SendTransactionProposalToProcessors(&apitxn.TransactionProposal{
		SignedProposal: &pb.SignedProposal{},
	}, nil)

	if result != nil || err == nil || err.Error() != "Missing peer objects for sending transaction proposal" {
		t.Fatalf("Test SendTransactionProposal failed, validation on peer is nil is not working as expected: %v", err)
	}

	result, err = SendTransactionProposalToProcessors(&apitxn.TransactionProposal{
		SignedProposal: &pb.SignedProposal{},
	}, []apitxn.ProposalProcessor{})

	if result != nil || err == nil || err.Error() != "Missing peer objects for sending transaction proposal" {
		t.Fatalf("Test SendTransactionProposal failed, validation on missing peer objects is not working: %v", err)
	}

	result, err = SendTransactionProposalToProcessors(&apitxn.TransactionProposal{
		SignedProposal: nil,
	}, nil)

	if result != nil || err == nil || err.Error() != "signedProposal is nil" {
		t.Fatal("Test SendTransactionProposal failed, validation on signedProposal is nil is not working as expected")
	}

	result, err = SendTransactionProposalToProcessors(&apitxn.TransactionProposal{
		SignedProposal: &pb.SignedProposal{},
	}, targets)

	if result == nil || err != nil {
		t.Fatalf("Test SendTransactionProposal failed, with error '%s'", err.Error())
	}
}
