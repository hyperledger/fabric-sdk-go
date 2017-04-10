/*
Copyright SecureKey Technologies Inc. All Rights Reserved.


Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at


      http://www.apache.org/licenses/LICENSE-2.0


Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package integration

import (
	"fmt"
	"time"

	fabricClient "github.com/hyperledger/fabric-sdk-go/fabric-client"
	"github.com/hyperledger/fabric-sdk-go/fabric-client/events"
)

// CreateAndSendTransactionProposal combines create and send transaction proposal methods into one method.
// See CreateTransactionProposal and SendTransactionProposal
func CreateAndSendTransactionProposal(chain fabricClient.Chain, chainCodeID string, chainID string, args []string, targets []fabricClient.Peer) ([]*fabricClient.TransactionProposalResponse, string, error) {

	signedProposal, err := chain.CreateTransactionProposal(chainCodeID, chainID, args, true, nil)
	if err != nil {
		return nil, "", fmt.Errorf("SendTransactionProposal return error: %v", err)
	}

	transactionProposalResponses, err := chain.SendTransactionProposal(signedProposal, 0, targets)
	if err != nil {
		return nil, "", fmt.Errorf("SendTransactionProposal return error: %v", err)
	}

	for _, v := range transactionProposalResponses {
		if v.Err != nil {
			return nil, signedProposal.TransactionID, fmt.Errorf("invoke Endorser %s return error: %v", v.Endorser, v.Err)
		}
		fmt.Printf("invoke Endorser '%s' return ProposalResponse status:%v\n", v.Endorser, v.Status)
	}

	return transactionProposalResponses, signedProposal.TransactionID, nil
}

// CreateAndSendTransaction combines create and send transaction methods into one method.
// See CreateTransaction and SendTransaction
func CreateAndSendTransaction(chain fabricClient.Chain, resps []*fabricClient.TransactionProposalResponse) ([]*fabricClient.TransactionResponse, error) {

	tx, err := chain.CreateTransaction(resps)
	if err != nil {
		return nil, fmt.Errorf("CreateTransaction return error: %v", err)
	}

	transactionResponse, err := chain.SendTransaction(tx)
	if err != nil {
		return nil, fmt.Errorf("SendTransaction return error: %v", err)

	}
	for _, v := range transactionResponse {
		if v.Err != nil {
			return nil, fmt.Errorf("Orderer %s return error: %v", v.Orderer, v.Err)
		}
	}

	return transactionResponse, nil
}

// RegisterEvent registers on the given eventhub for the give transaction
// returns a boolean channel which receives true when the event is complete
// and an error channel for errors
func RegisterEvent(txID string, eventHub events.EventHub) (chan bool, chan error) {
	done := make(chan bool)
	fail := make(chan error)

	eventHub.RegisterTxEvent(txID, func(txId string, err error) {
		if err != nil {
			fmt.Printf("Received error event for txid(%s)\n", txId)
			fail <- err
		} else {
			fmt.Printf("Received success event for txid(%s)\n", txId)
			done <- true
		}
	})

	return done, fail
}

// RegisterCCEvent registers chain code event on the given eventhub
// @returns {chan bool} channel which receives true when the event is complete
// @returns {object} ChainCodeCBE object handle that should be used to unregister
func RegisterCCEvent(chainCodeID, eventID string, eventHub events.EventHub) (chan bool, *events.ChainCodeCBE) {
	done := make(chan bool)

	// Register callback for CE
	rce := eventHub.RegisterChaincodeEvent(chainCodeID, eventID, func(ce *events.ChaincodeEvent) {
		fmt.Printf("Received CC event ( %s ): \n%v\n", time.Now().Format(time.RFC850), ce)
		done <- true
	})

	return done, rce
}
