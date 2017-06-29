/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package fabrictxn

import (
	"fmt"
	"time"

	api "github.com/hyperledger/fabric-sdk-go/api"
	internal "github.com/hyperledger/fabric-sdk-go/pkg/fabric-txn/internal"

	"github.com/op/go-logging"
)

var logger = logging.MustGetLogger("fabric_sdk_go")

//QueryChaincode ...
func QueryChaincode(client api.FabricClient, channel api.Channel, chaincodeID string, args []string) (string, error) {
	err := checkCommonArgs(client, channel, chaincodeID)
	if err != nil {
		return "", err
	}

	transactionProposalResponses, _, err := internal.CreateAndSendTransactionProposal(channel,
		chaincodeID, channel.Name(), args, []api.Peer{channel.PrimaryPeer()}, nil)

	if err != nil {
		return "", fmt.Errorf("CreateAndSendTransactionProposal returned error: %v", err)
	}

	return string(transactionProposalResponses[0].ProposalResponse.GetResponse().Payload), nil
}

// InvokeChaincode ...
func InvokeChaincode(client api.FabricClient, channel api.Channel, targets []api.Peer,
	eventHub api.EventHub, chaincodeID string, args []string, transientData map[string][]byte) error {

	err := checkCommonArgs(client, channel, chaincodeID)
	if err != nil {
		return err
	}

	if eventHub == nil {
		return fmt.Errorf("Eventhub is nil")
	}

	if targets == nil || len(targets) == 0 {
		return fmt.Errorf("No target peers")
	}

	if eventHub.IsConnected() == false {
		err = eventHub.Connect()
		if err != nil {
			return fmt.Errorf("Error connecting to eventhub: %v", err)
		}
		defer eventHub.Disconnect()
	}

	transactionProposalResponses, txID, err := internal.CreateAndSendTransactionProposal(channel,
		chaincodeID, channel.Name(), args, targets, transientData)

	if err != nil {
		return fmt.Errorf("CreateAndSendTransactionProposal returned error: %v", err)
	}

	done, fail := internal.RegisterTxEvent(txID, eventHub)

	_, err = internal.CreateAndSendTransaction(channel, transactionProposalResponses)
	if err != nil {
		return fmt.Errorf("CreateAndSendTransaction returned error: %v", err)
	}

	select {
	case <-done:
	case err := <-fail:
		return fmt.Errorf("invoke Error received from eventhub for txid(%s), error(%v)", txID, err)
	case <-time.After(time.Second * 30):
		return fmt.Errorf("invoke Didn't receive block event for txid(%s)", txID)
	}

	return nil
}

// checkCommonArgs ...
func checkCommonArgs(client api.FabricClient, channel api.Channel, chaincodeID string) error {
	if client == nil {
		return fmt.Errorf("Client is nil")
	}

	if channel == nil {
		return fmt.Errorf("Channel is nil")
	}

	if chaincodeID == "" {
		return fmt.Errorf("ChaincodeID is empty")
	}

	return nil
}

// RegisterCCEvent registers chain code event on the given eventhub
// @returns {chan bool} channel which receives true when the event is complete
// @returns {object} ChainCodeCBE object handle that should be used to unregister
func RegisterCCEvent(chainCodeID string, eventID string, eventHub api.EventHub) (chan bool, *api.ChainCodeCBE) {
	done := make(chan bool)

	// Register callback for CE
	rce := eventHub.RegisterChaincodeEvent(chainCodeID, eventID, func(ce *api.ChaincodeEvent) {
		logger.Debugf("Received CC event: %v\n", ce)
		done <- true
	})

	return done, rce
}
