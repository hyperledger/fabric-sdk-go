/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package util

import (
	"fmt"

	client "github.com/hyperledger/fabric-sdk-go/pkg/fabric-client"
	sdkUser "github.com/hyperledger/fabric-sdk-go/pkg/fabric-client/user"

	api "github.com/hyperledger/fabric-sdk-go/api"
	fabricCAClient "github.com/hyperledger/fabric-sdk-go/pkg/fabric-ca-client"

	kvs "github.com/hyperledger/fabric-sdk-go/pkg/fabric-client/keyvaluestore"
	bccspFactory "github.com/hyperledger/fabric/bccsp/factory"
)

// CreateAndSendTransactionProposal combines create and send transaction proposal methods into one method.
// See CreateTransactionProposal and SendTransactionProposal
func CreateAndSendTransactionProposal(channel api.Channel, chainCodeID string, channelID string, args []string, targets []api.Peer, transientData map[string][]byte) ([]*api.TransactionProposalResponse, string, error) {

	signedProposal, err := channel.CreateTransactionProposal(chainCodeID, channelID, args, true, transientData)
	if err != nil {
		return nil, "", fmt.Errorf("SendTransactionProposal return error: %v", err)
	}

	transactionProposalResponses, err := channel.SendTransactionProposal(signedProposal, 0, targets)
	if err != nil {
		return nil, "", fmt.Errorf("SendTransactionProposal return error: %v", err)
	}

	for _, v := range transactionProposalResponses {
		if v.Err != nil {
			return nil, signedProposal.TransactionID, fmt.Errorf("invoke Endorser %s return error: %v", v.Endorser, v.Err)
		}
		logger.Debugf("invoke Endorser '%s' return ProposalResponse status:%v\n", v.Endorser, v.Status)
	}

	return transactionProposalResponses, signedProposal.TransactionID, nil
}

// CreateAndSendTransaction combines create and send transaction methods into one method.
// See CreateTransaction and SendTransaction
func CreateAndSendTransaction(channel api.Channel, resps []*api.TransactionProposalResponse) ([]*api.TransactionResponse, error) {

	tx, err := channel.CreateTransaction(resps)
	if err != nil {
		return nil, fmt.Errorf("CreateTransaction return error: %v", err)
	}

	transactionResponse, err := channel.SendTransaction(tx)
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

// GetClient initializes and returns a client based on config and user
func GetClient(name string, pwd string, stateStorePath string, config api.Config) (api.FabricClient, error) {
	client := client.NewClient(config)

	cryptoSuite := bccspFactory.GetDefault()

	client.SetCryptoSuite(cryptoSuite)
	stateStore, err := kvs.CreateNewFileKeyValueStore(stateStorePath)
	if err != nil {
		return nil, fmt.Errorf("CreateNewFileKeyValueStore return error[%s]", err)
	}
	client.SetStateStore(stateStore)
	user, err := client.LoadUserFromStateStore(name)
	if err != nil {
		return nil, fmt.Errorf("client.LoadUserFromStateStore return error: %v", err)
	}
	if user == nil {
		fabricCAClient, err := fabricCAClient.NewFabricCAClient(config)
		if err != nil {
			return nil, fmt.Errorf("NewFabricCAClient return error: %v", err)
		}
		key, cert, err := fabricCAClient.Enroll(name, pwd)
		if err != nil {
			return nil, fmt.Errorf("Enroll return error: %v", err)
		}
		user = sdkUser.NewUser(name)
		user.SetPrivateKey(key)
		user.SetEnrollmentCertificate(cert)
		err = client.SaveUserToStateStore(user, false)
		if err != nil {
			return nil, fmt.Errorf("client.SaveUserToStateStore return error: %v", err)
		}
	}

	client.SetUserContext(user)

	return client, nil
}
