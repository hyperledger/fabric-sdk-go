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

package util

import (
	"fmt"

	fabricClient "github.com/hyperledger/fabric-sdk-go/fabric-client"

	fabricCAClient "github.com/hyperledger/fabric-sdk-go/fabric-ca-client"

	kvs "github.com/hyperledger/fabric-sdk-go/fabric-client/keyvaluestore"
	bccspFactory "github.com/hyperledger/fabric/bccsp/factory"
)

// CreateAndSendTransactionProposal combines create and send transaction proposal methods into one method.
// See CreateTransactionProposal and SendTransactionProposal
func CreateAndSendTransactionProposal(chain fabricClient.Chain, chainCodeID string, chainID string, args []string, targets []fabricClient.Peer, transientData map[string][]byte) ([]*fabricClient.TransactionProposalResponse, string, error) {

	signedProposal, err := chain.CreateTransactionProposal(chainCodeID, chainID, args, true, transientData)
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
		logger.Debugf("invoke Endorser '%s' return ProposalResponse status:%v\n", v.Endorser, v.Status)
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

// GetClient initializes and returns a client based on config and user
func GetClient(name string, pwd string, stateStorePath string) (fabricClient.Client, error) {
	client := fabricClient.NewClient()

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
		fabricCAClient, err := fabricCAClient.NewFabricCAClient()
		if err != nil {
			return nil, fmt.Errorf("NewFabricCAClient return error: %v", err)
		}
		key, cert, err := fabricCAClient.Enroll(name, pwd)
		if err != nil {
			return nil, fmt.Errorf("Enroll return error: %v", err)
		}
		user = fabricClient.NewUser(name)
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
