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
	"testing"

	fabricClient "github.com/hyperledger/fabric-sdk-go/fabric-client"
	"github.com/hyperledger/fabric-sdk-go/fabric-client/util"
)

// TestTransient ...
func TestTransient(t *testing.T) {

	testSetup := BaseSetupImpl{
		ConfigFile:      "../fixtures/config/config_test.yaml",
		ChainID:         "mychannel",
		ChannelConfig:   "../fixtures/channel/mychannel.tx",
		ConnectEventHub: true,
	}

	if err := testSetup.Initialize(); err != nil {
		t.Fatalf(err.Error())
	}

	if err := testSetup.InstallAndInstantiateExampleCC(); err != nil {
		t.Fatalf("InstallAndInstantiateExampleCC return error: %v", err)
	}

	var args []string
	args = append(args, "invoke")
	args = append(args, "move")
	args = append(args, "a")
	args = append(args, "b")
	args = append(args, "0")
	transientData := "Transient data test..."

	transientDataMap := make(map[string][]byte)
	transientDataMap["result"] = []byte(transientData)

	transactionProposalResponse, _, err := util.CreateAndSendTransactionProposal(testSetup.Chain, testSetup.ChainCodeID, testSetup.ChainID, args, []fabricClient.Peer{testSetup.Chain.GetPrimaryPeer()}, transientDataMap)
	if err != nil {
		t.Fatalf("CreateAndSendTransactionProposal return error: %v", err)
	}
	strResponse := string(transactionProposalResponse[0].GetResponsePayload())
	//validate transient data exists in proposal
	if len(strResponse) == 0 {
		t.Fatalf("Transient data does not exist: expected %s", transientData)
	}
	//verify transient data content
	if strResponse != transientData {
		t.Fatalf("Expected '%s' in transient data field. Received '%s' ", transientData, strResponse)
	}
	//transient data null
	transientDataMap["result"] = []byte{}
	transactionProposalResponse, _, err = util.CreateAndSendTransactionProposal(testSetup.Chain, testSetup.ChainCodeID, testSetup.ChainID, args, []fabricClient.Peer{testSetup.Chain.GetPrimaryPeer()}, transientDataMap)
	if err != nil {
		t.Fatalf("CreateAndSendTransactionProposal with empty transient data return an error: %v", err)
	}
	//validate that transient data does not exist in proposal
	strResponse = string(transactionProposalResponse[0].GetResponsePayload())
	if len(strResponse) != 0 {
		t.Fatalf("Transient data validation has failed. An empty transient data was expected but %s was returned", strResponse)
	}

}
