/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package internal

import (
	"fmt"

	fab "github.com/hyperledger/fabric-sdk-go/api/apifabclient"
	"github.com/hyperledger/fabric-sdk-go/api/apitxn"
	"github.com/hyperledger/fabric-sdk-go/pkg/logging"
	pb "github.com/hyperledger/fabric-sdk-go/third_party/github.com/hyperledger/fabric/protos/peer"
)

var logger = logging.NewLogger("fabric_sdk_go")

// CreateAndSendTransactionProposal ...
func CreateAndSendTransactionProposal(sender apitxn.ProposalSender, chainCodeID string,
	fcn string, args []string, targets []apitxn.ProposalProcessor, transientData map[string][]byte) ([]*apitxn.TransactionProposalResponse, apitxn.TransactionID, error) {

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
			return nil, request.TxnID, fmt.Errorf("invoke Endorser %s returned error: %v", v.Endorser, v.Err)
		}
		logger.Debugf("invoke Endorser '%s' returned ProposalResponse status:%v\n", v.Endorser, v.Status)
	}

	return transactionProposalResponses, txnID, nil
}

// CreateAndSendTransaction ...
func CreateAndSendTransaction(sender apitxn.Sender, resps []*apitxn.TransactionProposalResponse) (*apitxn.TransactionResponse, error) {

	tx, err := sender.CreateTransaction(resps)
	if err != nil {
		return nil, fmt.Errorf("CreateTransaction returned error: %v", err)
	}

	transactionResponse, err := sender.SendTransaction(tx)
	if err != nil {
		return nil, fmt.Errorf("SendTransaction returned error: %v", err)

	}
	if transactionResponse.Err != nil {
		return nil, fmt.Errorf("Orderer %s returned error: %v", transactionResponse.Orderer, transactionResponse.Err)
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
