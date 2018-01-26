/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package sdk

import (
	"path"
	"testing"
	"time"

	"github.com/hyperledger/fabric-sdk-go/api/apitxn"
	pb "github.com/hyperledger/fabric-sdk-go/third_party/github.com/hyperledger/fabric/protos/peer"

	"github.com/hyperledger/fabric-sdk-go/pkg/config"
	"github.com/hyperledger/fabric-sdk-go/pkg/fabsdk"
	"github.com/hyperledger/fabric-sdk-go/test/integration"
	"github.com/hyperledger/fabric-sdk-go/test/metadata"
)

const (
	org1Name = "Org1"
)

func TestChannelClient(t *testing.T) {

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

	// Create SDK setup for the integration tests
	sdk, err := fabsdk.New(config.FromFile(testSetup.ConfigFile))
	if err != nil {
		t.Fatalf("Failed to create new SDK: %s", err)
	}

	chClient, err := sdk.NewClient(fabsdk.WithUser("User1")).Channel(testSetup.ChannelID)
	if err != nil {
		t.Fatalf("Failed to create new channel client: %s", err)
	}

	// Synchronous query
	testQuery("200", testSetup.ChainCodeID, chClient, t)

	transientData := "Some data"
	transientDataMap := make(map[string][]byte)
	transientDataMap["result"] = []byte(transientData)

	// Synchronous transaction
	response, _, err := chClient.Execute(
		apitxn.Request{
			ChaincodeID:  testSetup.ChainCodeID,
			Fcn:          "invoke",
			Args:         integration.ExampleCCTxArgs(),
			TransientMap: transientDataMap,
		})
	if err != nil {
		t.Fatalf("Failed to move funds: %s", err)
	}
	// The example CC should return the transient data as a response
	if string(response) != transientData {
		t.Fatalf("Expecting response [%s] but got [%s]", transientData, response)
	}

	// Verify transaction using asynchronous query
	testQueryWithOpts("201", testSetup.ChainCodeID, chClient, t)

	// Asynchronous transaction
	testAsyncTransaction(testSetup.ChainCodeID, chClient, t)

	// Verify asynchronous transaction
	testQuery("202", testSetup.ChainCodeID, chClient, t)

	// Verify that filter error and commit error did not modify value
	testQuery("202", testSetup.ChainCodeID, chClient, t)

	// Test register and receive chaincode event
	testChaincodeEvent(testSetup.ChainCodeID, chClient, t)

	// Verify transaction with chain code event completed
	testQuery("203", testSetup.ChainCodeID, chClient, t)

	// Release channel client resources
	err = chClient.Close()
	if err != nil {
		t.Fatalf("Failed to close channel client: %v", err)
	}

}

func testQuery(expected string, ccID string, chClient apitxn.ChannelClient, t *testing.T) {

	result, err := chClient.Query(apitxn.Request{ChaincodeID: ccID, Fcn: "invoke", Args: integration.ExampleCCQueryArgs()})
	if err != nil {
		t.Fatalf("Failed to invoke example cc: %s", err)
	}

	if string(result) != expected {
		t.Fatalf("Expecting %s, got %s", expected, result)
	}
}

func testQueryWithOpts(expected string, ccID string, chClient apitxn.ChannelClient, t *testing.T) {

	notifier := make(chan apitxn.Response)
	result, err := chClient.Query(apitxn.Request{ChaincodeID: ccID, Fcn: "invoke", Args: integration.ExampleCCQueryArgs()}, apitxn.WithNotifier(notifier))
	if err != nil {
		t.Fatalf("Failed to invoke example cc asynchronously: %s", err)
	}
	if result != nil {
		t.Fatalf("Expecting empty, got %s", result)
	}

	select {
	case response := <-notifier:
		if response.Error != nil {
			t.Fatalf("Query returned error: %s", response.Error)
		}
		if string(response.Payload) != expected {
			t.Fatalf("Expecting %s, got %s", expected, response.Payload)
		}
	case <-time.After(time.Second * 20):
		t.Fatalf("Query Request timed out")
	}

}

func testAsyncTransaction(ccID string, chClient apitxn.ChannelClient, t *testing.T) {

	txNotifier := make(chan apitxn.Response)
	_, _, err := chClient.Execute(apitxn.Request{ChaincodeID: ccID, Fcn: "invoke", Args: integration.ExampleCCTxArgs()}, apitxn.WithNotifier(txNotifier))
	if err != nil {
		t.Fatalf("Failed to move funds: %s", err)
	}

	select {
	case response := <-txNotifier:
		if response.Error != nil {
			t.Fatalf("Execute returned error: %s", response.Error)
		}
		if response.TxValidationCode != pb.TxValidationCode_VALID {
			t.Fatalf("Expecting TxValidationCode to be TxValidationCode_VALID but received: %s", response.TxValidationCode)
		}
	case <-time.After(time.Second * 20):
		t.Fatalf("Execute timed out")
	}
}

type TestTxFilter struct {
	err          error
	errResponses error
}

func (tf *TestTxFilter) ProcessTxProposalResponse(txProposalResponse []*apitxn.TransactionProposalResponse) ([]*apitxn.TransactionProposalResponse, error) {
	if tf.err != nil {
		return nil, tf.err
	}

	var newResponses []*apitxn.TransactionProposalResponse

	if tf.errResponses != nil {
		// 404 will cause transaction commit error
		txProposalResponse[0].ProposalResponse.Response.Status = 404
	}

	newResponses = append(newResponses, txProposalResponse[0])
	return newResponses, nil
}

func testChaincodeEvent(ccID string, chClient apitxn.ChannelClient, t *testing.T) {

	eventID := "test([a-zA-Z]+)"

	// Register chaincode event (pass in channel which receives event details when the event is complete)
	notifier := make(chan *apitxn.CCEvent)
	rce := chClient.RegisterChaincodeEvent(notifier, ccID, eventID)

	// Synchronous transaction
	_, txID, err := chClient.Execute(apitxn.Request{ChaincodeID: ccID, Fcn: "invoke", Args: integration.ExampleCCTxArgs()})
	if err != nil {
		t.Fatalf("Failed to move funds: %s", err)
	}

	select {
	case ccEvent := <-notifier:
		t.Logf("Received cc event: %s", ccEvent)
		if ccEvent.TxID != txID.ID {
			t.Fatalf("CCEvent(%s) and Execute(%s) transaction IDs don't match", ccEvent.TxID, txID.ID)
		}
	case <-time.After(time.Second * 20):
		t.Fatalf("Did NOT receive CC for eventId(%s)\n", eventID)
	}

	// Unregister chain code event using registration handle
	err = chClient.UnregisterChaincodeEvent(rce)
	if err != nil {
		t.Fatalf("Unregister cc event failed: %s", err)
	}

}
