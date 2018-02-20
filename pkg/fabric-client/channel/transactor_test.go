/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package channel

import (
	"testing"

	"github.com/hyperledger/fabric-sdk-go/api/apifabclient"

	"github.com/hyperledger/fabric-sdk-go/pkg/fabric-client/mocks"
	"github.com/hyperledger/fabric-sdk-go/pkg/fabric-client/txn"
	"github.com/stretchr/testify/assert"
)

func TestCreateTxnID(t *testing.T) {
	transactor := createTransactor(t)
	createTxnID(t, transactor)
}

func TestTransactionProposal(t *testing.T) {
	transactor := createTransactor(t)
	tp := createTransactionProposal(t, transactor)
	createTransactionProposalResponse(t, transactor, tp)
}

func TestTransaction(t *testing.T) {
	transactor := createTransactor(t)
	tp := createTransactionProposal(t, transactor)
	tpr := createTransactionProposalResponse(t, transactor, tp)

	request := apifabclient.TransactionRequest{
		Proposal:          tp,
		ProposalResponses: tpr,
	}
	tx, err := txn.New(request)
	assert.Nil(t, err)

	_, err = transactor.SendTransaction(tx)
	assert.Nil(t, err)
}

func TestTransactionBadStatus(t *testing.T) {
	transactor := createTransactor(t)
	tp := createTransactionProposal(t, transactor)
	tpr := createTransactionProposalResponseBadStatus(t, transactor, tp)

	request := apifabclient.TransactionRequest{
		Proposal:          tp,
		ProposalResponses: tpr,
	}
	_, err := txn.New(request)
	assert.NotNil(t, err)
}

func createTransactor(t *testing.T) apifabclient.Transactor {
	user := mocks.NewMockUser("test")
	ctx := mocks.NewMockContext(user)
	orderer := mocks.NewMockOrderer("", nil)
	chConfig := mocks.NewMockChannelCfg("testChannel")

	transactor, err := NewTransactor(ctx, chConfig)
	transactor.orderers = []apifabclient.Orderer{orderer}
	assert.Nil(t, err)

	return transactor
}

func createTxnID(t *testing.T, transactor apifabclient.Transactor) apifabclient.TransactionID {
	txid, err := transactor.CreateTransactionID()
	assert.Nil(t, err, "creation of transaction ID failed")

	return txid
}

func createTransactionProposal(t *testing.T, transactor apifabclient.Transactor) *apifabclient.TransactionProposal {
	request := apifabclient.ChaincodeInvokeRequest{
		ChaincodeID: "example",
		Fcn:         "fcn",
	}

	txid := createTxnID(t, transactor)
	tp, err := txn.CreateChaincodeInvokeProposal(txid, "testChannel", request)
	assert.Nil(t, err)

	assert.NotEmpty(t, tp.TxnID.ID)
	assert.NotEmpty(t, tp.TxnID.Nonce)

	return tp
}

func createTransactionProposalResponse(t *testing.T, transactor apifabclient.Transactor, tp *apifabclient.TransactionProposal) []*apifabclient.TransactionProposalResponse {

	peer := mocks.MockPeer{MockName: "Peer1", MockURL: "http://peer1.com", MockRoles: []string{}, Status: 200}
	tpr, err := transactor.SendTransactionProposal(tp, []apifabclient.ProposalProcessor{&peer})
	assert.Nil(t, err)

	return tpr
}

func createTransactionProposalResponseBadStatus(t *testing.T, transactor apifabclient.Transactor, tp *apifabclient.TransactionProposal) []*apifabclient.TransactionProposalResponse {

	peer := mocks.MockPeer{MockName: "Peer1", MockURL: "http://peer1.com", MockRoles: []string{}, Status: 500}
	tpr, err := transactor.SendTransactionProposal(tp, []apifabclient.ProposalProcessor{&peer})
	assert.Nil(t, err)

	return tpr
}

// TestOrderersFromChannelCfg uses an orderer that exists in the configuration.
func TestOrderersFromChannelCfg(t *testing.T) {
	user := mocks.NewMockUser("test")
	ctx := mocks.NewMockContext(user)
	chConfig := mocks.NewMockChannelCfg("testChannel")
	chConfig.(*mocks.MockChannelCfg).MockOrderers = []string{"example.com"}

	o, err := orderersFromChannelCfg(ctx, chConfig)
	assert.Nil(t, err)
	assert.NotEmpty(t, o)
}

// TestOrderersFromChannelCfg uses an orderer that does not exist in the configuration.
func TestOrderersFromChannelCfgBadTLS(t *testing.T) {
	user := mocks.NewMockUser("test")
	ctx := mocks.NewMockContext(user)
	chConfig := mocks.NewMockChannelCfg("testChannel")
	chConfig.(*mocks.MockChannelCfg).MockOrderers = []string{"doesnotexist.com"}

	o, err := orderersFromChannelCfg(ctx, chConfig)
	assert.Nil(t, err)
	assert.NotEmpty(t, o)
}
