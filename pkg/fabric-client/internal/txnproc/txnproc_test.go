/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package txnproc

import (
	"fmt"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/hyperledger/fabric-sdk-go/api/apifabclient"
	"github.com/hyperledger/fabric-sdk-go/api/apifabclient/mocks"
	pb "github.com/hyperledger/fabric-sdk-go/third_party/github.com/hyperledger/fabric/protos/peer"
	"github.com/stretchr/testify/assert"
)

func TestSendTransactionProposalToProcessors(t *testing.T) {

	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()
	proc := mock_apifabclient.NewMockProposalProcessor(mockCtrl)

	tp := apifabclient.TransactionProposal{
		SignedProposal: &pb.SignedProposal{},
	}
	tpr := apifabclient.TransactionProposalResult{Endorser: "example.com", Status: 99, Proposal: tp, ProposalResponse: nil}
	proc.EXPECT().ProcessTransactionProposal(tp).Return(tpr, nil)
	targets := []apifabclient.ProposalProcessor{proc}

	result, err := SendTransactionProposalToProcessors(&apifabclient.TransactionProposal{
		SignedProposal: &pb.SignedProposal{},
	}, nil)

	if result != nil || err == nil || err.Error() != "targets is required" {
		t.Fatalf("Test SendTransactionProposal failed, validation on peer is nil is not working as expected: %v", err)
	}

	result, err = SendTransactionProposalToProcessors(&apifabclient.TransactionProposal{
		SignedProposal: &pb.SignedProposal{},
	}, []apifabclient.ProposalProcessor{})

	if result != nil || err == nil || err.Error() != "targets is required" {
		t.Fatalf("Test SendTransactionProposal failed, validation on missing peer objects is not working: %v", err)
	}

	result, err = SendTransactionProposalToProcessors(&apifabclient.TransactionProposal{
		SignedProposal: nil,
	}, nil)

	if result != nil || err == nil || err.Error() != "signedProposal is required" {
		t.Fatal("Test SendTransactionProposal failed, validation on signedProposal is nil is not working as expected")
	}

	result, err = SendTransactionProposalToProcessors(&apifabclient.TransactionProposal{
		SignedProposal: &pb.SignedProposal{},
	}, targets)

	if result == nil || err != nil {
		t.Fatalf("Test SendTransactionProposal failed, with error '%s'", err.Error())
	}
}

func TestProposalResponseError(t *testing.T) {
	testError := fmt.Errorf("Test Error")

	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()
	proc := mock_apifabclient.NewMockProposalProcessor(mockCtrl)

	tp := apifabclient.TransactionProposal{
		SignedProposal: &pb.SignedProposal{},
	}

	// Test with error from lower layer
	tpr := apifabclient.TransactionProposalResult{Endorser: "example.com", Status: 200,
		Proposal: tp, ProposalResponse: nil}
	proc.EXPECT().ProcessTransactionProposal(tp).Return(tpr, testError)
	targets := []apifabclient.ProposalProcessor{proc}
	resp, _ := SendTransactionProposalToProcessors(&apifabclient.TransactionProposal{
		SignedProposal: &pb.SignedProposal{},
	}, targets)
	assert.Equal(t, testError, resp[0].Err)
}
