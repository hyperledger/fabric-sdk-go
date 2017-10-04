/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package internal

import (
	fab "github.com/hyperledger/fabric-sdk-go/api/apifabclient"
	"github.com/hyperledger/fabric-sdk-go/api/apitxn"
	pb "github.com/hyperledger/fabric-sdk-go/third_party/github.com/hyperledger/fabric/protos/peer"

	"github.com/hyperledger/fabric-sdk-go/pkg/errors"
	"github.com/hyperledger/fabric-sdk-go/pkg/logging"
)

var logger = logging.NewLogger("fabric_sdk_go")

// CreateAndSendTransactionProposal ...
func CreateAndSendTransactionProposal(sender apitxn.ProposalSender, chainCodeID string,
	fcn string, args [][]byte, targets []apitxn.ProposalProcessor, transientData map[string][]byte) ([]*apitxn.TransactionProposalResponse, apitxn.TransactionID, error) {

	request := apitxn.ChaincodeInvokeRequest{
		Targets:      targets,
		Fcn:          fcn,
		Args:         args,
		TransientMap: transientData,
		ChaincodeID:  chainCodeID,
	}
	transactionProposalResponses, txnID, err := sender.SendTransactionProposal(request)
	if err != nil {
		return nil, txnID, err
	}

	for _, v := range transactionProposalResponses {
		if v.Err != nil {
			logger.Debugf("SendTransactionProposal failed (%v, %s)", v.Endorser, v.Err.Error())
			return nil, request.TxnID, errors.Wrap(v.Err, "SendTransactionProposal failed")
		}
		logger.Debugf("invoke Endorser '%s' returned ProposalResponse status:%v", v.Endorser, v.Status)
	}

	return transactionProposalResponses, txnID, nil
}

// CreateAndSendTransaction ...
func CreateAndSendTransaction(sender apitxn.Sender, resps []*apitxn.TransactionProposalResponse) (*apitxn.TransactionResponse, error) {

	tx, err := sender.CreateTransaction(resps)
	if err != nil {
		return nil, errors.WithMessage(err, "CreateTransaction failed")
	}

	transactionResponse, err := sender.SendTransaction(tx)
	if err != nil {
		return nil, errors.WithMessage(err, "SendTransaction failed")

	}
	if transactionResponse.Err != nil {
		logger.Debugf("orderer %s failed (%s)", transactionResponse.Orderer, transactionResponse.Err.Error())
		return nil, errors.Wrap(transactionResponse.Err, "orderer failed")
	}

	return transactionResponse, nil
}

// RegisterTxEvent registers on the given eventhub for the given transaction id
// returns a TxValidationCode channel which receives the validation code when the
// transaction completes. If the code is TxValidationCode_VALID then
// the transaction committed successfully, otherwise the code indicates the error
// that occurred.
func RegisterTxEvent(txID apitxn.TransactionID, eventHub fab.EventHub) chan pb.TxValidationCode {
	chcode := make(chan pb.TxValidationCode)

	eventHub.RegisterTxEvent(txID, func(txId string, code pb.TxValidationCode, err error) {
		logger.Debugf("Received code(%s) for txid(%s)\n", code, txId)
		chcode <- code
	})

	return chcode
}
