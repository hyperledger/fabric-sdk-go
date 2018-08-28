/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package event

import (
	"testing"
	"time"

	"github.com/hyperledger/fabric-sdk-go/pkg/client/channel"
	"github.com/hyperledger/fabric-sdk-go/pkg/client/event"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/errors/retry"
	"github.com/hyperledger/fabric-sdk-go/pkg/fab/events/deliverclient/seek"
	"github.com/hyperledger/fabric-sdk-go/pkg/fabsdk"
	"github.com/hyperledger/fabric-sdk-go/test/integration"
)

func TestDefaultEventClient(t *testing.T) {

	// using shared SDK instance to increase test speed
	sdk := mainSDK
	testSetup := mainTestSetup
	chaincodeID := mainChaincodeID

	// prepare channel client context
	org1ChannelClientContext := sdk.ChannelContext(testSetup.ChannelID, fabsdk.WithUser(org1User), fabsdk.WithOrg(org1Name))

	// get channel client (used to generate transactions)
	chClient, err := channel.New(org1ChannelClientContext)
	if err != nil {
		t.Fatalf("Failed to create new channel client: %s", err)
	}

	// get default event client (with filtered block events)
	eventClient, err := event.New(org1ChannelClientContext, event.WithSeekType(seek.Newest))
	if err != nil {
		t.Fatalf("Failed to create new events client: %s", err)
	}

	// test register and receive chaincode event (payload is not expected)
	testCCEvent(chaincodeID, chClient, eventClient, false, t)

	// test register filter block event
	testRegisterFilteredBlockEvent(chaincodeID, chClient, eventClient, t)

	// default event client (with filtered blocks) is not allowed to register for block events
	_, _, err = eventClient.RegisterBlockEvent()
	if err == nil {
		t.Fatal("Default events client should have failed to register for block events")
	}
}

func TestEventsClientWithBlockEvents(t *testing.T) {

	// using shared SDK instance to increase test speed
	sdk := mainSDK
	testSetup := mainTestSetup
	chaincodeID := mainChaincodeID

	// prepare channel client context
	org1ChannelClientContext := sdk.ChannelContext(testSetup.ChannelID, fabsdk.WithUser(org1User), fabsdk.WithOrg(org1Name))

	// get channel client (used to generate transactions)
	chClient, err := channel.New(org1ChannelClientContext)
	if err != nil {
		t.Fatalf("Failed to create new channel client: %s", err)
	}

	// create event client with block events
	eventClient, err := event.New(org1ChannelClientContext, event.WithBlockEvents(), event.WithSeekType(seek.Newest))
	if err != nil {
		t.Fatalf("Failed to create new events client with block events: %s", err)
	}

	// test register and receive chaincode event (payload is expected since we are set for receiving block events)
	testCCEvent(chaincodeID, chClient, eventClient, true, t)

	// test register block and filter block event
	testRegisterBlockEvent(chaincodeID, chClient, eventClient, t)
	testRegisterFilteredBlockEvent(chaincodeID, chClient, eventClient, t)
}

func testCCEvent(ccID string, chClient *channel.Client, eventClient *event.Client, expectPayload bool, t *testing.T) {

	eventID := integration.GenerateRandomID()
	payload := "Test Payload"

	// Register chaincode event (pass in channel which receives event details when the event is complete)
	reg, notifier, err := eventClient.RegisterChaincodeEvent(ccID, eventID)
	if err != nil {
		t.Fatalf("Failed to register cc event: %s", err)
	}
	defer eventClient.Unregister(reg)

	response, err := chClient.Execute(channel.Request{ChaincodeID: ccID, Fcn: "invoke", Args: append(integration.ExampleCCTxRandomSetArgs(), []byte(eventID))},
		channel.WithRetry(retry.DefaultChannelOpts))
	if err != nil {
		t.Fatalf("Failed to move funds: %s", err)
	}

	select {
	case ccEvent := <-notifier:
		t.Logf("Received cc event: %#v", ccEvent)
		if expectPayload && string(ccEvent.Payload[:]) != payload {
			t.Fatal("Did not receive 'Test Payload'")
		}

		if !expectPayload && string(ccEvent.Payload[:]) != "" {
			t.Fatalf("Expected empty payload, got %s", ccEvent.Payload[:])
		}
		if ccEvent.TxID != string(response.TransactionID) {
			t.Fatalf("CCEvent(%s) and Execute(%s) transaction IDs don't match", ccEvent.TxID, string(response.TransactionID))
		}
	case <-time.After(time.Second * 20):
		t.Fatalf("Did NOT receive CC for eventId(%s)\n", eventID)
	}
}

func testRegisterBlockEvent(ccID string, chClient *channel.Client, eventClient *event.Client, t *testing.T) {

	breg, beventch, err := eventClient.RegisterBlockEvent()
	if err != nil {
		t.Fatalf("Error registering for block events: %s", err)
	}
	defer eventClient.Unregister(breg)

	response, err := chClient.Execute(channel.Request{ChaincodeID: ccID, Fcn: "invoke", Args: integration.ExampleCCTxRandomSetArgs()},
		channel.WithRetry(retry.DefaultChannelOpts))
	if err != nil {
		t.Fatalf("Failed to move funds: %s", err)
	}

	select {
	case e, ok := <-beventch:
		if !ok {
			t.Fatal("unexpected closed channel while waiting for block event")
		}
		t.Logf("Received block event: %#v", e)
		if e.Block == nil {
			t.Fatal("Expecting block in block event but got nil")
		}
	case <-time.After(time.Second * 20):
		t.Fatalf("Did NOT receive block event for txID(%s)\n", response.TransactionID)
	}
}

func testRegisterFilteredBlockEvent(ccID string, chClient *channel.Client, eventClient *event.Client, t *testing.T) {

	fbreg, fbeventch, err := eventClient.RegisterFilteredBlockEvent()
	if err != nil {
		t.Fatalf("Error registering for block events: %s", err)
	}
	defer eventClient.Unregister(fbreg)

	response, err := chClient.Execute(channel.Request{ChaincodeID: ccID, Fcn: "invoke", Args: integration.ExampleCCTxRandomSetArgs()},
		channel.WithRetry(retry.DefaultChannelOpts))
	if err != nil {
		t.Fatalf("Failed to move funds: %s", err)
	}

	select {
	case event, ok := <-fbeventch:
		if !ok {
			t.Fatal("unexpected closed channel while waiting for filtered block event")
		}
		if event.FilteredBlock == nil {
			t.Fatal("Expecting filtered block in filtered block event but got nil")
		}
		t.Logf("Received filtered block event: %#v", event)

	case <-time.After(time.Second * 20):
		t.Fatalf("Did NOT receive filtered block event for txID(%s)\n", response.TransactionID)
	}
}
