/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package sdk

import (
	"path"
	"testing"
	"time"

	pb "github.com/hyperledger/fabric-sdk-go/third_party/github.com/hyperledger/fabric/protos/peer"

	"github.com/hyperledger/fabric-sdk-go/pkg/client/channel"
	"github.com/hyperledger/fabric-sdk-go/pkg/context/api/fab"
	"github.com/hyperledger/fabric-sdk-go/pkg/core/config"
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

	if err := testSetup.Initialize(); err != nil {
		t.Fatalf(err.Error())
	}

	chainCodeID := integration.GenerateRandomID()
	if err := integration.InstallAndInstantiateExampleCC(testSetup.SDK, fabsdk.WithUser("Admin"), testSetup.OrgID, chainCodeID); err != nil {
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
	testQuery("200", chainCodeID, chClient, t)

	transientData := "Some data"
	transientDataMap := make(map[string][]byte)
	transientDataMap["result"] = []byte(transientData)

	// Synchronous transaction
	response, err := chClient.Execute(
		channel.Request{
			ChaincodeID:  chainCodeID,
			Fcn:          "invoke",
			Args:         integration.ExampleCCTxArgs(),
			TransientMap: transientDataMap,
		})
	if err != nil {
		t.Fatalf("Failed to move funds: %s", err)
	}
	// The example CC should return the transient data as a response
	if string(response.Payload) != transientData {
		t.Fatalf("Expecting response [%s] but got [%v]", transientData, response)
	}

	// Verify transaction using query
	testQueryWithOpts("201", chainCodeID, chClient, t)

	// transaction
	testTransaction(chainCodeID, chClient, t)

	// Verify transaction
	testQuery("202", chainCodeID, chClient, t)

	// Verify that filter error and commit error did not modify value
	testQuery("202", chainCodeID, chClient, t)

	// Test register and receive chaincode event
	testChaincodeEvent(chainCodeID, chClient, t)

	// Verify transaction with chain code event completed
	testQuery("203", chainCodeID, chClient, t)

	// Test invocation of custom handler
	testInvokeHandler(chainCodeID, chClient, t)

	// Test receive event using separate client
	listener, err := sdk.NewClient(fabsdk.WithUser("User1")).Channel(testSetup.ChannelID)
	if err != nil {
		t.Fatalf("Failed to create new channel client: %s", err)
	}
	defer listener.Close()

	testChaincodeEventListener(chainCodeID, chClient, listener, t)

	// Release channel client resources
	err = chClient.Close()
	if err != nil {
		t.Fatalf("Failed to close channel client: %v", err)
	}

}

func testQuery(expected string, ccID string, chClient *channel.Client, t *testing.T) {

	response, err := chClient.Query(channel.Request{ChaincodeID: ccID, Fcn: "invoke", Args: integration.ExampleCCQueryArgs()})
	if err != nil {
		t.Fatalf("Failed to invoke example cc: %s", err)
	}

	if string(response.Payload) != expected {
		t.Fatalf("Expecting %s, got %s", expected, response.Payload)
	}
}

func testQueryWithOpts(expected string, ccID string, chClient *channel.Client, t *testing.T) {
	response, err := chClient.Query(channel.Request{ChaincodeID: ccID, Fcn: "invoke", Args: integration.ExampleCCQueryArgs()})
	if err != nil {
		t.Fatalf("Query returned error: %s", err)
	}
	if string(response.Payload) != expected {
		t.Fatalf("Expecting %s, got %s", expected, response.Payload)
	}
}

func testTransaction(ccID string, chClient *channel.Client, t *testing.T) {
	response, err := chClient.Execute(channel.Request{ChaincodeID: ccID, Fcn: "invoke", Args: integration.ExampleCCTxArgs()})
	if err != nil {
		t.Fatalf("Failed to move funds: %s", err)
	}
	if response.TxValidationCode != pb.TxValidationCode_VALID {
		t.Fatalf("Expecting TxValidationCode to be TxValidationCode_VALID but received: %s", response.TxValidationCode)
	}
}

type testHandler struct {
	t                *testing.T
	txID             *string
	endorser         *string
	txValidationCode *pb.TxValidationCode
	next             channel.Handler
}

func (h *testHandler) Handle(requestContext *channel.RequestContext, clientContext *channel.ClientContext) {
	if h.txID != nil {
		*h.txID = string(requestContext.Response.TransactionID)
		h.t.Logf("Custom handler writing TxID [%s]", *h.txID)
	}
	if h.endorser != nil && len(requestContext.Response.Responses) > 0 {
		*h.endorser = requestContext.Response.Responses[0].Endorser
		h.t.Logf("Custom handler writing Endorser [%s]", *h.endorser)
	}
	if h.txValidationCode != nil {
		*h.txValidationCode = requestContext.Response.TxValidationCode
		h.t.Logf("Custom handler writing TxValidationCode [%s]", *h.txValidationCode)
	}
	if h.next != nil {
		h.t.Logf("Custom handler invoking next handler")
		h.next.Handle(requestContext, clientContext)
	}
}

func testInvokeHandler(ccID string, chClient *channel.Client, t *testing.T) {
	// Insert a custom handler before and after the commit.
	// Ensure that the handlers are being called by writing out some data
	// and comparing with response.

	var txID string
	var endorser string
	txValidationCode := pb.TxValidationCode(-1)

	response, err := chClient.InvokeHandler(
		channel.NewProposalProcessorHandler(
			channel.NewEndorsementHandler(
				channel.NewEndorsementValidationHandler(
					&testHandler{
						t:        t,
						txID:     &txID,
						endorser: &endorser,
						next: channel.NewCommitHandler(
							&testHandler{
								t:                t,
								txValidationCode: &txValidationCode,
							},
						),
					},
				),
			),
		),
		channel.Request{
			ChaincodeID: ccID,
			Fcn:         "invoke",
			Args:        integration.ExampleCCTxArgs(),
		},
		channel.WithTimeout(5*time.Second),
	)
	if err != nil {
		t.Fatalf("Failed to invoke example cc asynchronously: %s", err)
	}
	if len(response.Responses) == 0 {
		t.Fatalf("Expecting more than one endorsement responses but got none")
	}
	if txID != string(response.TransactionID) {
		t.Fatalf("Expecting TxID [%s] but got [%s]", string(response.TransactionID), txID)
	}
	if endorser != response.Responses[0].Endorser {
		t.Fatalf("Expecting endorser [%s] but got [%s]", response.Responses[0].Endorser, endorser)
	}
	if txValidationCode != response.TxValidationCode {
		t.Fatalf("Expecting TxValidationCode [%s] but got [%s]", response.TxValidationCode, txValidationCode)
	}
}

type TestTxFilter struct {
	err          error
	errResponses error
}

func (tf *TestTxFilter) ProcessTxProposalResponse(txProposalResponse []*fab.TransactionProposalResponse) ([]*fab.TransactionProposalResponse, error) {
	if tf.err != nil {
		return nil, tf.err
	}

	var newResponses []*fab.TransactionProposalResponse

	if tf.errResponses != nil {
		// 404 will cause transaction commit error
		txProposalResponse[0].ProposalResponse.Response.Status = 404
	}

	newResponses = append(newResponses, txProposalResponse[0])
	return newResponses, nil
}

func testChaincodeEvent(ccID string, chClient *channel.Client, t *testing.T) {

	eventID := "test([a-zA-Z]+)"

	// Register chaincode event (pass in channel which receives event details when the event is complete)
	notifier := make(chan *channel.CCEvent)
	rce, err := chClient.RegisterChaincodeEvent(notifier, ccID, eventID)
	if err != nil {
		t.Fatalf("Failed to register cc event: %s", err)
	}

	response, err := chClient.Execute(channel.Request{ChaincodeID: ccID, Fcn: "invoke", Args: integration.ExampleCCTxArgs()})
	if err != nil {
		t.Fatalf("Failed to move funds: %s", err)
	}

	select {
	case ccEvent := <-notifier:
		t.Logf("Received cc event: %s", ccEvent)
		if ccEvent.TxID != string(response.TransactionID) {
			t.Fatalf("CCEvent(%s) and Execute(%s) transaction IDs don't match", ccEvent.TxID, string(response.TransactionID))
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

func testChaincodeEventListener(ccID string, chClient *channel.Client, listener *channel.Client, t *testing.T) {

	eventID := "test([a-zA-Z]+)"

	// Register chaincode event (pass in channel which receives event details when the event is complete)
	notifier := make(chan *channel.CCEvent)
	rce, err := listener.RegisterChaincodeEvent(notifier, ccID, eventID)
	if err != nil {
		t.Fatalf("Failed to register cc event: %s", err)
	}

	response, err := chClient.Execute(channel.Request{ChaincodeID: ccID, Fcn: "invoke", Args: integration.ExampleCCTxArgs()})
	if err != nil {
		t.Fatalf("Failed to move funds: %s", err)
	}

	select {
	case ccEvent := <-notifier:
		t.Logf("Received cc event: %s", ccEvent)
		if ccEvent.TxID != string(response.TransactionID) {
			t.Fatalf("CCEvent(%s) and Execute(%s) transaction IDs don't match", ccEvent.TxID, string(response.TransactionID))
		}
	case <-time.After(time.Second * 20):
		t.Fatalf("Did NOT receive CC for eventId(%s)\n", eventID)
	}

	// Unregister chain code event using registration handle
	err = listener.UnregisterChaincodeEvent(rce)
	if err != nil {
		t.Fatalf("Unregister cc event failed: %s", err)
	}

}
