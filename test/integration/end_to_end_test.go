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
	"math/rand"
	"strconv"
	"testing"
	"time"

	fabric_client "github.com/hyperledger/fabric-sdk-go/fabric-client"
	events "github.com/hyperledger/fabric-sdk-go/fabric-client/events"

	config "github.com/hyperledger/fabric-sdk-go/config"
)

var chainCodeID = ""
var chainID = "mychannel"
var chainCodePath = "github.com/example_cc"
var chainCodeVersion = "v0"

func TestChainCodeInvoke(t *testing.T) {
	InitConfigForEndToEnd()
	testSetup := BaseSetupImpl{}

	eventHub := testSetup.GetEventHub(t, nil)
	queryChain, invokeChain, deployChain := testSetup.GetChains(t)
	testSetup.SetupChaincodeDeploy()
	err := installCC(deployChain)
	if err != nil {
		t.Fatalf("installCC return error: %v", err)
	}
	err = instantiateCC(deployChain, eventHub)
	if err != nil {
		t.Fatalf("instantiateCC return error: %v", err)
	}
	// Get Query value before invoke
	value, err := getQueryValue(t, queryChain)
	if err != nil {
		t.Fatalf("getQueryValue return error: %v", err)
	}
	fmt.Printf("*** QueryValue before invoke %s\n", value)

	_, err = invoke(t, invokeChain, eventHub)
	if err != nil {
		t.Fatalf("invoke return error: %v", err)
	}

	valueAfterInvoke, err := getQueryValue(t, queryChain)
	if err != nil {
		t.Errorf("getQueryValue return error: %v", err)
		return
	}
	fmt.Printf("*** QueryValue after invoke %s\n", valueAfterInvoke)

	valueInt, _ := strconv.Atoi(value)
	valueInt = valueInt + 1
	valueAfterInvokeInt, _ := strconv.Atoi(valueAfterInvoke)
	if valueInt != valueAfterInvokeInt {
		t.Fatalf("SendTransaction didn't change the QueryValue")

	}

}

func getQueryValue(t *testing.T, chain fabric_client.Chain) (string, error) {

	var args []string
	args = append(args, "invoke")
	args = append(args, "query")
	args = append(args, "b")

	signedProposal, err := chain.CreateTransactionProposal(chainCodeID, chainID, args, true, nil)
	if err != nil {
		return "", fmt.Errorf("SendTransactionProposal return error: %v", err)
	}
	transactionProposalResponses, err := chain.SendTransactionProposal(signedProposal, 0, nil)
	if err != nil {
		return "", fmt.Errorf("SendTransactionProposal return error: %v", err)
	}

	for _, v := range transactionProposalResponses {
		if v.Err != nil {
			return "", fmt.Errorf("query Endorser %s return error: %v", v.Endorser, v.Err)
		}
		return string(v.GetResponsePayload()), nil
	}
	return "", nil
}

func invoke(t *testing.T, chain fabric_client.Chain, eventHub events.EventHub) (string, error) {

	var args []string
	args = append(args, "invoke")
	args = append(args, "move")
	args = append(args, "a")
	args = append(args, "b")
	args = append(args, "1")

	signedProposal, err := chain.CreateTransactionProposal(chainCodeID, chainID, args, true, nil)
	if err != nil {
		return "", fmt.Errorf("SendTransactionProposal return error: %v", err)
	}
	transactionProposalResponse, err := chain.SendTransactionProposal(signedProposal, 0, nil)
	if err != nil {
		return "", fmt.Errorf("SendTransactionProposal return error: %v", err)
	}

	for _, v := range transactionProposalResponse {
		if v.Err != nil {
			return "", fmt.Errorf("invoke Endorser %s return error: %v", v.Endorser, v.Err)
		}
		fmt.Printf("invoke Endorser '%s' return ProposalResponse status:%v\n", v.Endorser, v.Status)
	}

	tx, err := chain.CreateTransaction(transactionProposalResponse)
	if err != nil {
		return "", fmt.Errorf("CreateTransaction return error: %v", err)

	}
	transactionResponse, err := chain.SendTransaction(tx)
	if err != nil {
		return "", fmt.Errorf("SendTransaction return error: %v", err)

	}
	for _, v := range transactionResponse {
		if v.Err != nil {
			return "", fmt.Errorf("Orderer %s return error: %v", v.Orderer, v.Err)
		}
	}
	done := make(chan bool)
	fail := make(chan error)
	eventHub.RegisterTxEvent(signedProposal.TransactionID, func(txId string, err error) {
		if err != nil {
			fail <- err
		} else {
			fmt.Printf("invoke receive success event for txid(%s)\n", txId)
			done <- true
		}
	})

	select {
	case <-done:
	case <-fail:
		return "", fmt.Errorf("invoke Error received from eventhub for txid(%s) error(%v)", signedProposal.TransactionID, fail)
	case <-time.After(time.Second * 30):
		return "", fmt.Errorf("invoke Didn't receive block event for txid(%s)", signedProposal.TransactionID)
	}
	return signedProposal.TransactionID, nil

}

func installCC(chain fabric_client.Chain) error {

	transactionProposalResponse, _, err := chain.SendInstallProposal(chainCodeID, chainCodePath, chainCodeVersion, nil, nil)
	if err != nil {
		return fmt.Errorf("SendInstallProposal return error: %v", err)
	}

	for _, v := range transactionProposalResponse {
		if v.Err != nil {
			return fmt.Errorf("SendInstallProposal Endorser %s return error: %v", v.Endorser, v.Err)
		}
		fmt.Printf("SendInstallProposal Endorser '%s' return ProposalResponse status:%v\n", v.Endorser, v.Status)
	}

	return nil

}

func instantiateCC(chain fabric_client.Chain, eventHub events.EventHub) error {

	var args []string
	args = append(args, "init")
	args = append(args, "a")
	args = append(args, "100")
	args = append(args, "b")
	args = append(args, "200")

	transactionProposalResponse, txID, err := chain.SendInstantiateProposal(chainCodeID, chainID, args, chainCodePath, chainCodeVersion, nil)
	if err != nil {
		return fmt.Errorf("SendInstantiateProposal return error: %v", err)
	}

	for _, v := range transactionProposalResponse {
		if v.Err != nil {
			return fmt.Errorf("SendInstantiateProposal Endorser %s return error: %v", v.Endorser, v.Err)
		}
		fmt.Printf("SendInstantiateProposal Endorser '%s' return ProposalResponse status:%v\n", v.Endorser, v.Status)
	}

	tx, err := chain.CreateTransaction(transactionProposalResponse)
	if err != nil {
		return fmt.Errorf("CreateTransaction return error: %v", err)

	}
	transactionResponse, err := chain.SendTransaction(tx)
	if err != nil {
		return fmt.Errorf("SendTransaction return error: %v", err)

	}
	for _, v := range transactionResponse {
		if v.Err != nil {
			return fmt.Errorf("Orderer %s return error: %v", v.Orderer, v.Err)
		}
	}
	done := make(chan bool)
	fail := make(chan error)

	eventHub.RegisterTxEvent(txID, func(txId string, err error) {
		if err != nil {
			fail <- err
		} else {
			fmt.Printf("instantiateCC receive success event for txid(%s)\n", txId)
			done <- true
		}

	})

	select {
	case <-done:
	case <-fail:
		return fmt.Errorf("instantiateCC Error received from eventhub for txid(%s) error(%v)", txID, fail)
	case <-time.After(time.Second * 30):
		return fmt.Errorf("instantiateCC Didn't receive block event for txid(%s)", txID)
	}
	return nil

}

func InitConfigForEndToEnd() {
	err := config.InitConfig("../fixtures/config/config_test.yaml")
	if err != nil {
		fmt.Println(err.Error())
	}
}

func randomString(strlen int) string {
	const chars = "abcdefghijklmnopqrstuvwxyz0123456789"
	result := make([]byte, strlen)
	for i := 0; i < strlen; i++ {
		result[i] = chars[rand.Intn(len(chars))]
	}
	return string(result)
}

func init() {
	rand.Seed(time.Now().UnixNano())
	chainCodeID = randomString(10)
}
