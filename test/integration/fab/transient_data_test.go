/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package fab

import (
	"path"
	"testing"

	"github.com/hyperledger/fabric-sdk-go/api/apitxn"
	"github.com/hyperledger/fabric-sdk-go/test/integration"
	"github.com/hyperledger/fabric-sdk-go/test/metadata"
)

// TestTransient ...
func TestTransient(t *testing.T) {

	testSetup := integration.BaseSetupImpl{
		ConfigFile:      "../" + integration.ConfigTestFile,
		ChannelID:       "mychannel",
		OrgID:           org1Name,
		ChannelConfig:   path.Join("../../../", metadata.ChannelConfigPath, "mychannel.tx"),
		ConnectEventHub: true,
	}

	if err := testSetup.Initialize(t); err != nil {
		t.Fatalf(err.Error())
	}

	if err := testSetup.InstallAndInstantiateExampleCC(); err != nil {
		t.Fatalf("InstallAndInstantiateExampleCC return error: %v", err)
	}

	fcn := "invoke"
	transientData := "Transient data test..."

	transientDataMap := make(map[string][]byte)
	transientDataMap["result"] = []byte(transientData)

	transactionProposalResponse, _, err := testSetup.CreateAndSendTransactionProposal(testSetup.Channel, testSetup.ChainCodeID, fcn, integration.ExampleCCTxArgs(), []apitxn.ProposalProcessor{testSetup.Channel.PrimaryPeer()}, transientDataMap)
	if err != nil {
		t.Fatalf("CreateAndSendTransactionProposal return error: %v", err)
	}
	strResponse := string(transactionProposalResponse[0].ProposalResponse.GetResponse().Payload)
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
	transactionProposalResponse, _, err = testSetup.CreateAndSendTransactionProposal(testSetup.Channel, testSetup.ChainCodeID, fcn, integration.ExampleCCTxArgs(), []apitxn.ProposalProcessor{testSetup.Channel.PrimaryPeer()}, transientDataMap)
	if err != nil {
		t.Fatalf("CreateAndSendTransactionProposal with empty transient data return an error: %v", err)
	}
	//validate that transient data does not exist in proposal
	strResponse = string(transactionProposalResponse[0].ProposalResponse.GetResponse().Payload)
	if len(strResponse) != 0 {
		t.Fatalf("Transient data validation has failed. An empty transient data was expected but %s was returned", strResponse)
	}

}
