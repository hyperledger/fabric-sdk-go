/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/
package event

import (
	"math"
	"testing"
	"time"

	txnmocks "github.com/hyperledger/fabric-sdk-go/pkg/client/common/mocks"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/options"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/context"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/fab"
	contextImpl "github.com/hyperledger/fabric-sdk-go/pkg/context"
	fcmocks "github.com/hyperledger/fabric-sdk-go/pkg/fab/mocks"
	mspmocks "github.com/hyperledger/fabric-sdk-go/pkg/msp/test/mockmsp"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"

	pb "github.com/hyperledger/fabric-protos-go/peer"
	"github.com/hyperledger/fabric-sdk-go/pkg/fab/events/deliverclient/seek"
	"github.com/hyperledger/fabric-sdk-go/pkg/fab/events/service"
	"github.com/hyperledger/fabric-sdk-go/pkg/fab/events/service/dispatcher"
	servicemocks "github.com/hyperledger/fabric-sdk-go/pkg/fab/events/service/mocks"
)

var (
	channelID   = "testChannel"
	defaultOpts = []options.Opt{}
	sourceURL   = "localhost:9051"
)

func TestNewEventClient(t *testing.T) {

	fabCtx := setupCustomTestContext(t, nil)
	ctx := createChannelContext(fabCtx, channelID)

	_, err := New(ctx)
	if err != nil {
		t.Fatalf("Failed to create new event client: %s", err)
	}

	_, err = New(ctx, WithBlockEvents(), WithSeekType(seek.Newest), WithBlockNum(math.MaxUint64), WithEventConsumerTimeout(500*time.Millisecond), WithChaincodeID("testChaincode"))
	if err != nil {
		t.Fatalf("Failed to create new event client: %s", err)
	}

	ctxErr := createChannelContextWithError(fabCtx, channelID)
	_, err = New(ctxErr)
	if err == nil {
		t.Fatal("Should have failed with 'Test Error'")
	}
}

func TestNewEventClientWithFromBlock(t *testing.T) {

	fabCtx := setupCustomTestContext(t, nil)
	ctx := createChannelContext(fabCtx, channelID)

	_, err := New(ctx)
	if err != nil {
		t.Fatalf("Failed to create new event client: %s", err)
	}

	_, err = New(ctx, WithBlockEvents(), WithSeekType(seek.FromBlock), WithBlockNum(100), WithChaincodeID("testChaincode"))
	if err != nil {
		t.Fatalf("Failed to create new event client: %s", err)
	}
}

func TestBlockEvents(t *testing.T) {

	eventService, eventProducer, err := newServiceWithMockProducer(defaultOpts, withBlockLedger(sourceURL))
	if err != nil {
		t.Fatalf("error creating channel event client: %s", err)
	}
	defer eventProducer.Close()
	defer eventService.Stop()

	fabCtx := setupCustomTestContext(t, nil)
	ctx := createChannelContext(fabCtx, channelID)

	client, err := New(ctx)
	if err != nil {
		t.Fatalf("Failed to create new event client: %s", err)
	}

	client.eventService = eventService

	registration, eventch, err := client.RegisterBlockEvent()
	if err != nil {
		t.Fatalf("error registering for block events: %s", err)
	}
	defer client.Unregister(registration)

	eventProducer.Ledger().NewBlock(channelID)

	select {
	case _, ok := <-eventch:
		if !ok {
			t.Fatalf("unexpected closed channel")
		}
	case <-time.After(5 * time.Second):
		t.Fatal("timed out waiting for block event")
	}
}

func TestFilteredBlockEvents(t *testing.T) {

	eventService, eventProducer, err := newServiceWithMockProducer(defaultOpts, withFilteredBlockLedger(sourceURL))
	if err != nil {
		t.Fatalf("error creating channel event client: %s", err)
	}
	defer eventProducer.Close()
	defer eventService.Stop()

	fabCtx := setupCustomTestContext(t, nil)
	ctx := createChannelContext(fabCtx, channelID)

	client, err := New(ctx)
	if err != nil {
		t.Fatalf("Failed to create new event client: %s", err)
	}

	client.eventService = eventService

	registration, eventch, err := client.RegisterFilteredBlockEvent()
	if err != nil {
		t.Fatalf("error registering for filtered block events: %s", err)
	}
	defer client.Unregister(registration)

	txID1 := "1234"
	txCode1 := pb.TxValidationCode_VALID
	txID2 := "5678"
	txCode2 := pb.TxValidationCode_ENDORSEMENT_POLICY_FAILURE

	eventProducer.Ledger().NewFilteredBlock(
		channelID,
		servicemocks.NewFilteredTx(txID1, txCode1),
		servicemocks.NewFilteredTx(txID2, txCode2),
	)

	select {
	case fbevent, ok := <-eventch:
		if !ok {
			t.Fatal("unexpected closed channel")
		}
		if fbevent.FilteredBlock == nil {
			t.Fatal("Expecting filtered block but got nil")
		}
		if fbevent.FilteredBlock.ChannelId != channelID {
			t.Fatalf("Expecting channel [%s] but got [%s]", channelID, fbevent.FilteredBlock.ChannelId)
		}
	case <-time.After(5 * time.Second):
		t.Fatal("timed out waiting for filtered block event")
	}
}

func TestTxStatusEvents(t *testing.T) {
	chanID := "mychannel"
	eventService, eventProducer, err := newServiceWithMockProducer(defaultOpts, withFilteredBlockLedger(sourceURL))
	if err != nil {
		t.Fatalf("error creating channel event client: %s", err)
	}
	defer eventProducer.Close()
	defer eventService.Stop()

	fabCtx := setupCustomTestContext(t, nil)
	ctx := createChannelContext(fabCtx, chanID)

	client, err := New(ctx)
	if err != nil {
		t.Fatalf("Failed to create new event client: %s", err)
	}

	client.eventService = eventService

	txID1 := "1234"
	txID2 := "5678"

	if _, _, err1 := client.RegisterTxStatusEvent(""); err1 == nil {
		t.Fatal("expecting error registering for TxStatus event without a TX ID but got none")
	}

	reg1, eventch1, err := client.RegisterTxStatusEvent(txID1)
	if err != nil {
		t.Fatalf("error registering for TxStatus events: %s", err)
	}
	defer client.Unregister(reg1)

	reg2, eventch2, err := client.RegisterTxStatusEvent(txID2)
	if err != nil {
		t.Fatalf("error registering for TxStatus events: %s", err)
	}
	defer client.Unregister(reg2)
	validateTxStatusEvents(t, eventProducer, eventch1, eventch2, chanID, txID1, txID2)

}

func validateTxStatusEvents(t *testing.T, eventProducer *servicemocks.MockProducer, eventch1 <-chan *fab.TxStatusEvent, eventch2 <-chan *fab.TxStatusEvent, chanID string, txID1 string, txID2 string) {
	txCode1 := pb.TxValidationCode_VALID
	txCode2 := pb.TxValidationCode_ENDORSEMENT_POLICY_FAILURE
	eventProducer.Ledger().NewFilteredBlock(
		chanID,
		servicemocks.NewFilteredTx(txID1, txCode1),
		servicemocks.NewFilteredTx(txID2, txCode2),
	)

	numExpected := 2
	numReceived := 0
	done := false
	for !done {
		select {
		case event, ok := <-eventch1:
			if !ok {
				t.Fatal("unexpected closed channel")
			} else {
				checkTxStatusEvent(t, event, txID1, txCode1)
				numReceived++
			}
		case event, ok := <-eventch2:
			if !ok {
				t.Fatal("unexpected closed channel")
			} else {
				checkTxStatusEvent(t, event, txID2, txCode2)
				numReceived++
			}
		case <-time.After(5 * time.Second):
			t.Fatalf("timed out waiting for [%d] TxStatus events. Only received [%d]", numExpected, numReceived)
		}

		if numReceived == numExpected {
			break
		}
	}
}

func TestCCEvents(t *testing.T) {
	chanID := "mychannel"
	eventService, eventProducer, err := newServiceWithMockProducer(defaultOpts, withFilteredBlockLedger(sourceURL))
	if err != nil {
		t.Fatalf("error creating channel event client: %s", err)
	}
	defer eventProducer.Close()
	defer eventService.Stop()

	fabCtx := setupCustomTestContext(t, nil)
	ctx := createChannelContext(fabCtx, chanID)

	client, err := New(ctx)
	if err != nil {
		t.Fatalf("Failed to create new event client: %s", err)
	}

	client.eventService = eventService

	ccID1 := "mycc1"
	ccID2 := "mycc2"
	ccFilter1 := "event1"
	ccFilter2 := "event.*"

	if _, _, err1 := client.RegisterChaincodeEvent("", ccFilter1); err1 == nil {
		t.Fatal("expecting error registering for chaincode events without CC ID but got none")
	}

	reg1, eventch1, err := client.RegisterChaincodeEvent(ccID1, ccFilter1)
	if err != nil {
		t.Fatalf("error registering for block events: %s", err)
	}
	defer client.Unregister(reg1)

	reg2, eventch2, err := client.RegisterChaincodeEvent(ccID2, ccFilter2)
	if err != nil {
		t.Fatalf("error registering for chaincode events: %s", err)
	}
	defer client.Unregister(reg2)
	validateCCEvents(t, eventProducer, eventch1, eventch2, chanID, ccID1, ccID2)

}

func validateCCEvents(t *testing.T, eventProducer *servicemocks.MockProducer, eventch1 <-chan *fab.CCEvent, eventch2 <-chan *fab.CCEvent, chanID string, ccID1 string, ccID2 string) {
	event1 := "event1"
	event2 := "event2"
	event3 := "event3"
	eventProducer.Ledger().NewFilteredBlock(
		chanID,
		servicemocks.NewFilteredTxWithCCEvent("txid1", ccID1, event1),
		servicemocks.NewFilteredTxWithCCEvent("txid2", ccID2, event2),
		servicemocks.NewFilteredTxWithCCEvent("txid3", ccID2, event3),
	)

	numExpected := 3
	numReceived := 0
	done := false
	for !done {
		select {
		case event, ok := <-eventch1:
			if !ok {
				t.Fatal("unexpected closed channel")
			} else {
				checkCCEvent(t, event, ccID1, event1)
				numReceived++
			}
		case event, ok := <-eventch2:
			if !ok {
				t.Fatal("unexpected closed channel")
			} else {
				checkCCEvent(t, event, ccID2, event2, event3)
				numReceived++
			}
		case <-time.After(5 * time.Second):
			t.Fatalf("timed out waiting for [%d] CC events. Only received [%d]", numExpected, numReceived)
		}

		if numReceived == numExpected {
			break
		}
	}
}

func checkCCEvent(t *testing.T, event *fab.CCEvent, expectedCCID string, expectedEventNames ...string) {
	if event.ChaincodeID != expectedCCID {
		t.Fatalf("expecting event for CC [%s] but received event for CC [%s]", expectedCCID, event.ChaincodeID)
	}
	found := false
	for _, eventName := range expectedEventNames {
		if event.EventName == eventName {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("expecting one of [%v] but received [%s]", expectedEventNames, event.EventName)
	}
}

func checkTxStatusEvent(t *testing.T, event *fab.TxStatusEvent, expectedTxID string, expectedCode pb.TxValidationCode) {
	if event.TxID != expectedTxID {
		t.Fatalf("expecting event for TxID [%s] but received event for TxID [%s]", expectedTxID, event.TxID)
	}
	if event.TxValidationCode != expectedCode {
		t.Fatalf("expecting TxValidationCode [%s] but received [%s]", expectedCode, event.TxValidationCode)
	}
}

func setupCustomTestContext(t *testing.T, orderers []fab.Orderer) context.ClientProvider {
	user := mspmocks.NewMockSigningIdentity("test", "test")
	ctx := fcmocks.NewMockContext(user)

	if orderers == nil {
		orderer := fcmocks.NewMockOrderer("", nil)
		orderers = []fab.Orderer{orderer}
	}

	transactor := txnmocks.MockTransactor{
		Ctx:       ctx,
		ChannelID: channelID,
		Orderers:  orderers,
	}

	testChannelSvc, err := setupTestChannelService(ctx, orderers)
	testChannelSvc.(*fcmocks.MockChannelService).SetTransactor(&transactor)
	assert.Nil(t, err, "Got error %s", err)

	channelProvider := ctx.MockProviderContext.ChannelProvider()
	channelProvider.(*fcmocks.MockChannelProvider).SetCustomChannelService(testChannelSvc)

	return createClientContext(ctx)
}

func setupTestChannelService(ctx context.Client, orderers []fab.Orderer) (fab.ChannelService, error) {
	chProvider, err := fcmocks.NewMockChannelProvider(ctx)
	if err != nil {
		return nil, errors.WithMessage(err, "mock channel provider creation failed")
	}

	chService, err := chProvider.ChannelService(ctx, channelID)
	if err != nil {
		return nil, errors.WithMessage(err, "mock channel service creation failed")
	}

	return chService, nil
}

func createChannelContext(clientContext context.ClientProvider, channelID string) context.ChannelProvider {

	channelProvider := func() (context.Channel, error) {
		return contextImpl.NewChannel(clientContext, channelID)
	}

	return channelProvider
}

func createChannelContextWithError(clientContext context.ClientProvider, channelID string) context.ChannelProvider {

	channelProvider := func() (context.Channel, error) {
		return nil, errors.New("Test Error")
	}

	return channelProvider
}

func createClientContext(client context.Client) context.ClientProvider {
	return func() (context.Client, error) {
		return client, nil
	}
}

type producerOpts struct {
	ledger *servicemocks.MockLedger
}

type producerOpt func(opts *producerOpts)

func withBlockLedger(source string) producerOpt {
	return func(opts *producerOpts) {
		opts.ledger = servicemocks.NewMockLedger(servicemocks.BlockEventFactory, source)
	}
}

func withFilteredBlockLedger(source string) producerOpt {
	return func(opts *producerOpts) {
		opts.ledger = servicemocks.NewMockLedger(servicemocks.FilteredBlockEventFactory, source)
	}
}

func newServiceWithMockProducer(opts []options.Opt, pOpts ...producerOpt) (*service.Service, *servicemocks.MockProducer, error) {
	serv := service.New(dispatcher.New(opts...), opts...)
	if err := serv.Start(); err != nil {
		return nil, nil, err
	}

	eventch, err := serv.Dispatcher().EventCh()
	if err != nil {
		return nil, nil, err
	}

	popts := producerOpts{}
	for _, opt := range pOpts {
		opt(&popts)
	}

	ledger := popts.ledger
	if popts.ledger == nil {
		ledger = servicemocks.NewMockLedger(servicemocks.BlockEventFactory, sourceURL)
	}

	eventProducer := servicemocks.NewMockProducer(ledger)
	producerch := eventProducer.Register()

	go func() {
		for {
			event, ok := <-producerch
			if !ok {
				return
			}
			eventch <- event
		}
	}()

	return serv, eventProducer, nil
}
