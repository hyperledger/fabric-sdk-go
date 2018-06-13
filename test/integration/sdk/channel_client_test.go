/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package sdk

import (
	"strings"
	"testing"
	"time"

	pb "github.com/hyperledger/fabric-sdk-go/third_party/github.com/hyperledger/fabric/protos/peer"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/hyperledger/fabric-sdk-go/pkg/client/channel"
	"github.com/hyperledger/fabric-sdk-go/pkg/client/channel/invoke"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/errors/retry"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/errors/status"

	"github.com/hyperledger/fabric-sdk-go/pkg/core/config"

	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/fab"
	"github.com/hyperledger/fabric-sdk-go/pkg/fabsdk"
	"github.com/hyperledger/fabric-sdk-go/test/integration"
)

const (
	org1Name      = "Org1"
	org1User      = "User1"
	org1AdminUser = "Admin"
)

func TestChannelClient(t *testing.T) {

	// Using shared SDK instance to increase test speed.
	sdk := mainSDK
	testSetup := mainTestSetup
	chaincodeID := mainChaincodeID

	//prepare context
	org1ChannelClientContext := sdk.ChannelContext(testSetup.ChannelID, fabsdk.WithUser(org1User), fabsdk.WithOrg(org1Name))

	//get channel client
	chClient, err := channel.New(org1ChannelClientContext)
	if err != nil {
		t.Fatalf("Failed to create new channel client: %s", err)
	}

	// Synchronous query
	testQuery("200", chaincodeID, chClient, t)

	transientData := "Some data"
	transientDataMap := make(map[string][]byte)
	transientDataMap["result"] = []byte(transientData)

	// Synchronous transaction
	response, err := chClient.Execute(
		channel.Request{
			ChaincodeID:  chaincodeID,
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
	testQueryWithOpts("201", chaincodeID, chClient, t)

	// transaction
	nestedCCID := integration.GenerateRandomID()
	_, err = integration.InstallAndInstantiateExampleCC(sdk, fabsdk.WithUser("Admin"), testSetup.OrgID, nestedCCID)
	require.NoError(t, err)
	testTransaction(chaincodeID, nestedCCID, chClient, t)

	// Verify transaction
	testQuery("202", chaincodeID, chClient, t)

	// Verify that filter error and commit error did not modify value
	testQuery("202", chaincodeID, chClient, t)

	// Test register and receive chaincode event
	testChaincodeEvent(chaincodeID, chClient, t)

	// Verify transaction with chain code event completed
	testQuery("203", chaincodeID, chClient, t)

	// Test invocation of custom handler
	testInvokeHandler(chaincodeID, chClient, t)

	// Test chaincode error
	testChaincodeError(chaincodeID, chClient, t)

	// Test receive event using separate client
	listener, err := channel.New(org1ChannelClientContext)
	if err != nil {
		t.Fatalf("Failed to create new channel client: %s", err)
	}

	testChaincodeEventListener(chaincodeID, chClient, listener, t)
}

func testQuery(expected string, ccID string, chClient *channel.Client, t *testing.T) {

	response, err := chClient.Query(channel.Request{ChaincodeID: ccID, Fcn: "invoke", Args: integration.ExampleCCQueryArgs()},
		channel.WithRetry(retry.DefaultChannelOpts))
	if err != nil {
		t.Fatalf("Failed to invoke example cc: %s", err)
	}

	if string(response.Payload) != expected {
		t.Fatalf("Expecting %s, got %s", expected, response.Payload)
	}
}

func testQueryWithOpts(expected string, ccID string, chClient *channel.Client, t *testing.T) {
	response, err := chClient.Query(channel.Request{ChaincodeID: ccID, Fcn: "invoke", Args: integration.ExampleCCQueryArgs()},
		channel.WithRetry(retry.DefaultChannelOpts))
	if err != nil {
		t.Fatalf("Query returned error: %s", err)
	}
	if string(response.Payload) != expected {
		t.Fatalf("Expecting %s, got %s", expected, response.Payload)
	}
}

func testTransaction(ccID, nestedCCID string, chClient *channel.Client, t *testing.T) {
	response, err := chClient.Execute(
		channel.Request{
			ChaincodeID:     ccID,
			Fcn:             "invoke",
			Args:            integration.ExampleCCTxArgs(),
			InvocationChain: []*fab.ChaincodeCall{{ID: nestedCCID}},
		},
		channel.WithRetry(retry.DefaultChannelOpts),
	)
	require.NoError(t, err, "Failed to move funds")
	assert.Equal(t, pb.TxValidationCode_VALID, response.TxValidationCode, "Expecting TxValidationCode to be TxValidationCode_VALID")
}

type testHandler struct {
	t                *testing.T
	txID             *string
	endorser         *string
	txValidationCode *pb.TxValidationCode
	next             invoke.Handler
}

func (h *testHandler) Handle(requestContext *invoke.RequestContext, clientContext *invoke.ClientContext) {
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
		h.t.Log("Custom handler invoking next handler")
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
		invoke.NewProposalProcessorHandler(
			invoke.NewEndorsementHandler(
				invoke.NewEndorsementValidationHandler(
					&testHandler{
						t:        t,
						txID:     &txID,
						endorser: &endorser,
						next: invoke.NewCommitHandler(
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
		channel.WithTimeout(fab.Execute, 5*time.Second),
	)
	if err != nil {
		t.Fatalf("Failed to invoke example cc asynchronously: %s", err)
	}
	if len(response.Responses) == 0 {
		t.Fatal("Expecting more than one endorsement responses but got none")
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
	reg, notifier, err := chClient.RegisterChaincodeEvent(ccID, eventID)
	if err != nil {
		t.Fatalf("Failed to register cc event: %s", err)
	}
	defer chClient.UnregisterChaincodeEvent(reg)

	response, err := chClient.Execute(channel.Request{ChaincodeID: ccID, Fcn: "invoke", Args: integration.ExampleCCTxArgs()})
	if err != nil {
		t.Fatalf("Failed to move funds: %s", err)
	}

	select {
	case ccEvent := <-notifier:
		t.Logf("Received cc event: %#v", ccEvent)
		if ccEvent.TxID != string(response.TransactionID) {
			t.Fatalf("CCEvent(%s) and Execute(%s) transaction IDs don't match", ccEvent.TxID, string(response.TransactionID))
		}
	case <-time.After(time.Second * 20):
		t.Fatalf("Did NOT receive CC for eventId(%s)\n", eventID)
	}
}

func testChaincodeEventListener(ccID string, chClient *channel.Client, listener *channel.Client, t *testing.T) {

	eventID := integration.GenerateRandomID()

	// Register chaincode event (pass in channel which receives event details when the event is complete)
	reg, notifier, err := listener.RegisterChaincodeEvent(ccID, eventID)
	if err != nil {
		t.Fatalf("Failed to register cc event: %s", err)
	}
	defer chClient.UnregisterChaincodeEvent(reg)

	response, err := chClient.Execute(channel.Request{ChaincodeID: ccID, Fcn: "invoke", Args: append(integration.ExampleCCTxArgs(), []byte(eventID))})
	if err != nil {
		t.Fatalf("Failed to move funds: %s", err)
	}

	select {
	case ccEvent := <-notifier:
		t.Logf("Received cc event: %#v", ccEvent)
		if ccEvent.TxID != string(response.TransactionID) {
			t.Fatalf("CCEvent(%s) and Execute(%s) transaction IDs don't match", ccEvent.TxID, string(response.TransactionID))
		}
	case <-time.After(time.Second * 20):
		t.Fatalf("Did NOT receive CC for eventId(%s)\n", eventID)
	}

}

func testChaincodeError(ccID string, client *channel.Client, t *testing.T) {
	// Try calling unknown function call and expect an error
	r, err := client.Execute(channel.Request{ChaincodeID: ccID, Fcn: "DUMMY_FUNCTION", Args: integration.ExampleCCTxArgs()},
		channel.WithRetry(retry.DefaultChannelOpts))

	t.Logf("testChaincodeError err: %s ***** responses: %v", err, r)
	require.Error(t, err)
	s, ok := status.FromError(err)
	require.True(t, ok, "expected status error")

	checkError := func(s *status.Status) {
		require.EqualValues(t, status.ChaincodeStatus, s.Group, "expected ChaincodeStatus")
		require.Equal(t, int32(500), s.Code)
		require.Equal(t, "Unknown function call", s.Message)
	}

	if s.Code == int32(status.MultipleErrors) {
		t.Logf("Received multiple errors from endorsement:")
		for i, d := range s.Details {
			err, ok := d.(error)
			require.Truef(t, ok, "expecting error from status detail")
			s, ok = status.FromError(err)
			require.True(t, ok, "expected status error")
			t.Logf("(%d) - %#v", i, s)
			checkError(s)
		}
	} else {
		t.Logf("Received single error from endorsement: %#v", s)
		checkError(s)
	}
}

func TestNoEndpoints(t *testing.T) {

	// Using shared SDK instance to increase test speed.
	testSetup := mainTestSetup
	configProvider := config.FromFile("../../fixtures/config/config_test_endpoints.yaml")

	if integration.IsLocal() {
		//If it is a local test then add entity mapping to config backend to parse URLs
		configProvider = integration.AddLocalEntityMapping(configProvider, integration.LocalOrdererPeersConfig)
	}

	sdk, err := fabsdk.New(configProvider)
	if err != nil {
		t.Fatalf("Failed to create new SDK: %s", err)
	}

	// Prepare channel context
	org1AdminChannelContext := sdk.ChannelContext(testSetup.ChannelID, fabsdk.WithUser(org1AdminUser), fabsdk.WithOrg(org1Name))

	// Create new channel client
	chClient, err := channel.New(org1AdminChannelContext)
	if err != nil {
		t.Fatalf("Failed to create new resource management client: %s", err)
	}

	// Test query chaincode: since peer has been disabled for chaincode query this query should fail
	_, err = chClient.Query(channel.Request{ChaincodeID: mainChaincodeID, Fcn: "invoke", Args: integration.ExampleCCQueryArgs()},
		channel.WithRetry(retry.DefaultChannelOpts))
	require.Error(t, err)

	expected1_1Err := "targets were not provided"                   // When running with 1.1 DynamicSelection
	expected1_2Err := "no endorsement combination can be satisfied" // When running with 1.2 FabricSelection
	if !strings.Contains(err.Error(), expected1_1Err) && !strings.Contains(err.Error(), expected1_2Err) {
		t.Fatal("Should have failed due to no chaincode query peers")
	}

	// Test execute transaction: since peer has been disabled for endorsement this transaction should fail
	_, err = chClient.Execute(channel.Request{ChaincodeID: mainChaincodeID, Fcn: "invoke", Args: integration.ExampleCCTxArgs()},
		channel.WithRetry(retry.DefaultChannelOpts))
	if !strings.Contains(err.Error(), expected1_1Err) && !strings.Contains(err.Error(), expected1_2Err) {
		t.Fatal("Should have failed due to no chaincode query peers")
	}
}
