/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package internal

import (
	"fmt"

	fab "github.com/hyperledger/fabric-sdk-go/api/apifabclient"
	"github.com/hyperledger/fabric-sdk-go/api/apitxn"
	pb "github.com/hyperledger/fabric/protos/peer"
	"github.com/op/go-logging"
)

var logger = logging.MustGetLogger("fabric_sdk_go")

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
// returns a boolean channel which receives true when the event is complete and
// an error channel for errors
func RegisterTxEvent(txID apitxn.TransactionID, eventHub fab.EventHub) (chan bool, chan error) {
	done := make(chan bool)
	fail := make(chan error)

	eventHub.RegisterTxEvent(txID, func(txId string, errorCode pb.TxValidationCode, err error) {
		if err != nil {
			logger.Debugf("Received error event for txid(%s)\n", txId)
			fail <- err
		} else {
			logger.Debugf("Received success event for txid(%s)\n", txId)
			done <- true
		}
	})

	return done, fail
}
