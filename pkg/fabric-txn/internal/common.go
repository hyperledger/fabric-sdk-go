/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package internal

import (
	"fmt"

	fab "github.com/hyperledger/fabric-sdk-go/api/apifabclient"
	"github.com/hyperledger/fabric-sdk-go/api/apitxn"
	"github.com/hyperledger/fabric/common/crypto"
	pb "github.com/hyperledger/fabric/protos/peer"
	protos_utils "github.com/hyperledger/fabric/protos/utils"
	"github.com/op/go-logging"
)

var logger = logging.MustGetLogger("fabric_sdk_go")

// CreateAndSendTransactionProposal ...
func CreateAndSendTransactionProposal(sender apitxn.ProposalSender, chainCodeID string, channelID string,
	args []string, targets []apitxn.ProposalProcessor, transientData map[string][]byte) ([]*apitxn.TransactionProposalResponse, string, error) {

	signedProposal, err := sender.CreateTransactionProposal(chainCodeID, channelID, args, true, transientData)
	if err != nil {
		return nil, "", fmt.Errorf("SendTransactionProposal returned error: %v", err)
	}

	transactionProposalResponses, err := sender.SendTransactionProposal(signedProposal, 0, targets)
	if err != nil {
		return nil, "", fmt.Errorf("SendTransactionProposal returned error: %v", err)
	}

	for _, v := range transactionProposalResponses {
		if v.Err != nil {
			return nil, signedProposal.TransactionID, fmt.Errorf("invoke Endorser %s returned error: %v", v.Endorser, v.Err)
		}
		logger.Debugf("invoke Endorser '%s' returned ProposalResponse status:%v\n", v.Endorser, v.Status)
	}

	return transactionProposalResponses, signedProposal.TransactionID, nil
}

// CreateAndSendTransaction ...
func CreateAndSendTransaction(sender apitxn.Sender, resps []*apitxn.TransactionProposalResponse) ([]*apitxn.TransactionResponse, error) {

	tx, err := sender.CreateTransaction(resps)
	if err != nil {
		return nil, fmt.Errorf("CreateTransaction returned error: %v", err)
	}

	transactionResponses, err := sender.SendTransaction(tx)
	if err != nil {
		return nil, fmt.Errorf("SendTransaction returned error: %v", err)

	}
	for _, v := range transactionResponses {
		if v.Err != nil {
			return nil, fmt.Errorf("Orderer %s returned error: %v", v.Orderer, v.Err)
		}
	}

	return transactionResponses, nil
}

// RegisterTxEvent registers on the given eventhub for the given transaction id
// returns a boolean channel which receives true when the event is complete and
// an error channel for errors
func RegisterTxEvent(txID string, eventHub fab.EventHub) (chan bool, chan error) {
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

// GenerateRandomNonce generates a random nonce
func GenerateRandomNonce() ([]byte, error) {
	return crypto.GetRandomNonce()
}

// ComputeTxID computes a transaction ID from a given nonce and creator ID
func ComputeTxID(nonce []byte, creator []byte) (string, error) {
	return protos_utils.ComputeProposalTxID(nonce, creator)
}
