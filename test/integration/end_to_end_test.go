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

var chainCodeId = ""
var chainId = "testchainid"
var chainCodePath = "github.com/example_cc"
var chainCodeVersion = "v0"

func TestChainCodeInvoke(t *testing.T) {
	InitConfigForEndToEnd()
	testSetup := BaseSetupImpl{}

	eventHub := testSetup.GetEventHub(t, nil)
	querychain, invokechain := testSetup.GetChains(t)
	testSetup.SetupChaincodeDeploy()
	err := installCC(invokechain)
	if err != nil {
		t.Fatalf("installCC return error: %v", err)
	}
	err = instantiateCC(invokechain, eventHub)
	if err != nil {
		t.Fatalf("installCC return error: %v", err)
	}
	// Get Query value before invoke
	value, err := getQueryValue(t, querychain)
	if err != nil {
		t.Fatalf("getQueryValue return error: %v", err)
	}
	fmt.Printf("*** QueryValue before invoke %s\n", value)

	err = invoke(t, invokechain, eventHub)
	if err != nil {
		t.Fatalf("invoke return error: %v", err)
	}

	valueAfterInvoke, err := getQueryValue(t, querychain)
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

	signedProposal, err := chain.CreateTransactionProposal(chainCodeId, chainId, args, true, nil)
	if err != nil {
		return "", fmt.Errorf("SendTransactionProposal return error: %v", err)
	}
	transactionProposalResponses, err := chain.SendTransactionProposal(signedProposal, 0)
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

func invoke(t *testing.T, chain fabric_client.Chain, eventHub events.EventHub) error {

	var args []string
	args = append(args, "invoke")
	args = append(args, "move")
	args = append(args, "a")
	args = append(args, "b")
	args = append(args, "1")

	signedProposal, err := chain.CreateTransactionProposal(chainCodeId, chainId, args, true, nil)
	if err != nil {
		return fmt.Errorf("SendTransactionProposal return error: %v", err)
	}
	transactionProposalResponse, err := chain.SendTransactionProposal(signedProposal, 0)
	if err != nil {
		return fmt.Errorf("SendTransactionProposal return error: %v", err)
	}

	for _, v := range transactionProposalResponse {
		if v.Err != nil {
			return fmt.Errorf("invoke Endorser %s return error: %v", v.Endorser, v.Err)
		}
		fmt.Printf("invoke Endorser '%s' return ProposalResponse:%v\n", v.Endorser, v.GetResponsePayload())
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
	eventHub.RegisterTxEvent(signedProposal.TransactionID, func(txId string, err error) {
		fmt.Printf("receive success event for txid(%s)\n", txId)
		done <- true
	})

	select {
	case <-done:
	case <-time.After(time.Second * 20):
		return fmt.Errorf("Didn't receive block event for txid(%s)\n", signedProposal.TransactionID)
	}
	return nil

}

func installCC(chain fabric_client.Chain) error {

	transactionProposalResponse, _, err := chain.SendInstallProposal(chainCodeId, chainCodePath, chainCodeVersion, nil)
	if err != nil {
		return fmt.Errorf("SendInstallProposal return error: %v", err)
	}

	for _, v := range transactionProposalResponse {
		if v.Err != nil {
			return fmt.Errorf("SendInstallProposal Endorser %s return error: %v", v.Endorser, v.Err)
		}
		fmt.Printf("SendInstallProposal Endorser '%s' return ProposalResponse:%v\n", v.Endorser, v.GetResponsePayload())
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

	transactionProposalResponse, txID, err := chain.SendInstantiateProposal(chainCodeId, chainId, args, chainCodePath, chainCodeVersion)
	if err != nil {
		return fmt.Errorf("SendInstantiateProposal return error: %v", err)
	}

	for _, v := range transactionProposalResponse {
		if v.Err != nil {
			return fmt.Errorf("SendInstantiateProposal Endorser %s return error: %v", v.Endorser, v.Err)
		}
		fmt.Printf("SendInstantiateProposal Endorser '%s' return ProposalResponse:%v\n", v.Endorser, v.GetResponsePayload())
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
	eventHub.RegisterTxEvent(txID, func(txId string, err error) {
		fmt.Printf("receive success event for txid(%s)\n", txId)
		done <- true
	})

	select {
	case <-done:
	case <-time.After(time.Second * 20):
		return fmt.Errorf("Didn't receive block event for txid(%s)\n", txID)
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
	chainCodeId = randomString(10)
}
