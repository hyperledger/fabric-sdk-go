/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package service

import (
	"fmt"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/hyperledger/fabric-sdk-go/pkg/common/options"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/fab"

	"github.com/hyperledger/fabric-sdk-go/pkg/fab/events/service/dispatcher"
	servicemocks "github.com/hyperledger/fabric-sdk-go/pkg/fab/events/service/mocks"

	"github.com/hyperledger/fabric-sdk-go/pkg/fab/events/service/blockfilter/headertypefilter"
	cb "github.com/hyperledger/fabric-protos-go/common"
	pb "github.com/hyperledger/fabric-protos-go/peer"
	"github.com/pkg/errors"
)

var (
	defaultOpts = []options.Opt{}
	sourceURL   = "localhost:9051"
)

func TestInvalidUnregister(t *testing.T) {
	eventService, eventProducer, err := newServiceWithMockProducer(defaultOpts)
	if err != nil {
		t.Fatalf("error creating channel event client: %s", err)
	}
	defer eventProducer.Close()
	defer eventService.Stop()

	// Make sure the client doesn't panic with invalid registration
	eventService.Unregister("invalid registration")
}

func TestBlockEvents(t *testing.T) {
	channelID := "mychannel"
	eventService, eventProducer, err := newServiceWithMockProducer(defaultOpts, withBlockLedger(sourceURL))
	if err != nil {
		t.Fatalf("error creating channel event client: %s", err)
	}
	defer eventProducer.Close()
	defer eventService.Stop()

	registration, eventch, err := eventService.RegisterBlockEvent()
	if err != nil {
		t.Fatalf("error registering for block events: %s", err)
	}
	defer eventService.Unregister(registration)

	eventProducer.Ledger().NewBlock(channelID)

	select {
	case _, ok := <-eventch:
		if !ok {
			t.Fatal("unexpected closed channel")
		}
	case <-time.After(5 * time.Second):
		t.Fatal("timed out waiting for block event")
	}
}

func TestBlockEventsWithFilter(t *testing.T) {
	channelID := "mychannel"
	eventService, eventProducer, err := newServiceWithMockProducer(defaultOpts, withBlockLedger(sourceURL))
	if err != nil {
		t.Fatalf("error creating channel event client: %s", err)
	}
	defer eventProducer.Close()
	defer eventService.Stop()

	// Only want to see Config and Config Update blocks
	breg, beventch, err := eventService.RegisterBlockEvent(headertypefilter.New(cb.HeaderType_CONFIG, cb.HeaderType_CONFIG_UPDATE))
	if err != nil {
		t.Fatalf("error registering for block events: %s", err)
	}
	defer eventService.Unregister(breg)

	fbreg, fbeventch, err := eventService.RegisterFilteredBlockEvent()
	if err != nil {
		t.Fatalf("error registering for filtered block events: %s", err)
	}
	defer eventService.Unregister(fbreg)

	txID1 := "1234"
	txCode1 := pb.TxValidationCode_VALID
	txID2 := "5678"
	txCode2 := pb.TxValidationCode_ENDORSEMENT_POLICY_FAILURE

	eventProducer.Ledger().NewBlock(channelID,
		servicemocks.NewTransaction(txID1, txCode1, cb.HeaderType_CONFIG),
	)
	eventProducer.Ledger().NewBlock(channelID,
		servicemocks.NewTransaction(txID2, txCode2, cb.HeaderType_CONFIG_UPDATE),
	)
	eventProducer.Ledger().NewBlock(channelID,
		servicemocks.NewTransaction(txID2, txCode2, cb.HeaderType_ENDORSER_TRANSACTION),
	)
	checkBlockEventsWithFilter(t, beventch, fbeventch)

}

func checkBlockEventsWithFilter(t *testing.T, beventch <-chan *fab.BlockEvent, fbeventch <-chan *fab.FilteredBlockEvent) {
	numBlockEventsReceived := 0
	numBlockEventsExpected := 2
	numFilteredBlockEventsReceived := 0
	numFilteredBlockEventsExpected := 3

	for {
		select {
		case _, ok := <-beventch:
			if !ok {
				t.Fatal("unexpected closed channel")
			}
			numBlockEventsReceived++
		case _, ok := <-fbeventch:
			if !ok {
				t.Fatal("unexpected closed channel")
			}
			numFilteredBlockEventsReceived++
		case <-time.After(2 * time.Second):
			if numBlockEventsReceived != numBlockEventsExpected {
				t.Fatalf("Expecting %d block events but got %d", numBlockEventsExpected, numBlockEventsReceived)
			}
			if numFilteredBlockEventsReceived != numFilteredBlockEventsExpected {
				t.Fatalf("Expecting %d filtered block events but got %d", numFilteredBlockEventsExpected, numFilteredBlockEventsReceived)
			}
			return
		}
	}
}

func TestFilteredBlockEvents(t *testing.T) {
	channelID := "mychannel"
	eventService, eventProducer, err := newServiceWithMockProducer(defaultOpts, withFilteredBlockLedger(sourceURL))
	if err != nil {
		t.Fatalf("error creating channel event client: %s", err)
	}
	defer eventProducer.Close()
	defer eventService.Stop()

	registration, eventch, err := eventService.RegisterFilteredBlockEvent()
	if err != nil {
		t.Fatalf("error registering for filtered block events: %s", err)
	}
	defer eventService.Unregister(registration)

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

func TestBlockAndFilteredBlockEvents(t *testing.T) {
	channelID := "mychannel"
	eventService, eventProducer, err := newServiceWithMockProducer(defaultOpts, withBlockLedger(sourceURL))
	if err != nil {
		t.Fatalf("error creating channel event client: %s", err)
	}
	defer eventProducer.Close()
	defer eventService.Stop()

	breg, beventch, err := eventService.RegisterBlockEvent()
	if err != nil {
		t.Fatalf("error registering for block events: %s", err)
	}
	defer eventService.Unregister(breg)

	fbreg, fbeventch, err := eventService.RegisterFilteredBlockEvent()
	if err != nil {
		t.Fatalf("error registering for filtered block events: %s", err)
	}
	defer eventService.Unregister(fbreg)

	txID1 := "1234"
	txCode1 := pb.TxValidationCode_VALID
	txID2 := "5678"
	txCode2 := pb.TxValidationCode_ENDORSEMENT_POLICY_FAILURE

	eventProducer.Ledger().NewBlock(channelID,
		servicemocks.NewTransaction(txID1, txCode1, cb.HeaderType_CONFIG),
		servicemocks.NewTransaction(txID2, txCode2, cb.HeaderType_CONFIG_UPDATE),
	)
	checkBlockAndFilteredBlockEvents(t, beventch, fbeventch, channelID)

}

func checkBlockAndFilteredBlockEvents(t *testing.T, beventch <-chan *fab.BlockEvent, fbeventch <-chan *fab.FilteredBlockEvent, channelID string) {
	numReceived := 0
	numExpected := 2

	for {
		select {
		case fbevent, ok := <-fbeventch:
			if !ok {
				t.Fatal("unexpected closed channel")
			}
			if fbevent.FilteredBlock == nil {
				t.Fatal("Expecting filtered block but got nil")
			}
			if fbevent.FilteredBlock.ChannelId != channelID {
				t.Fatalf("Expecting channel [%s] but got [%s]", channelID, fbevent.FilteredBlock.ChannelId)
			}
			numReceived++
		case _, ok := <-beventch:
			if !ok {
				t.Fatal("unexpected closed channel")
			}
			numReceived++
		case <-time.After(2 * time.Second):
			if numReceived != numExpected {
				t.Fatalf("Expecting %d events but got %d", numExpected, numReceived)
			}
			return
		}
	}
}

func TestTxStatusEvents(t *testing.T) {
	channelID := "mychannel"
	eventService, eventProducer, err := newServiceWithMockProducer(defaultOpts, withFilteredBlockLedger(sourceURL))
	if err != nil {
		t.Fatalf("error creating channel event client: %s", err)
	}
	defer eventProducer.Close()
	defer eventService.Stop()

	txID1 := "1234"
	txCode1 := pb.TxValidationCode_VALID
	txID2 := "5678"
	txCode2 := pb.TxValidationCode_ENDORSEMENT_POLICY_FAILURE

	if _, _, err1 := eventService.RegisterTxStatusEvent(""); err1 == nil {
		t.Fatal("expecting error registering for TxStatus event without a TX ID but got none")
	}
	reg1, _, err := eventService.RegisterTxStatusEvent(txID1)
	if err != nil {
		t.Fatalf("error registering for TxStatus events: %s", err)
	}
	_, _, err = eventService.RegisterTxStatusEvent(txID1)
	if err == nil {
		t.Fatalf("expecting error registering multiple times for TxStatus events: %s", err)
	}
	eventService.Unregister(reg1)

	reg1, eventch1, err := eventService.RegisterTxStatusEvent(txID1)
	if err != nil {
		t.Fatalf("error registering for TxStatus events: %s", err)
	}
	defer eventService.Unregister(reg1)

	reg2, eventch2, err := eventService.RegisterTxStatusEvent(txID2)
	if err != nil {
		t.Fatalf("error registering for TxStatus events: %s", err)
	}
	defer eventService.Unregister(reg2)

	eventProducer.Ledger().NewFilteredBlock(
		channelID,
		servicemocks.NewFilteredTx(txID1, txCode1),
		servicemocks.NewFilteredTx(txID2, txCode2),
	)

	checkTxStatusEvents(eventch1, t, txID1, txCode1, eventch2, txID2, txCode2)
}

func checkTxStatusEvents(eventch1 <-chan *fab.TxStatusEvent, t *testing.T, txID1 string, txCode1 pb.TxValidationCode, eventch2 <-chan *fab.TxStatusEvent, txID2 string, txCode2 pb.TxValidationCode) {
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
	channelID := "mychannel"
	eventService, eventProducer, err := newServiceWithMockProducer(defaultOpts, withFilteredBlockLedger(sourceURL))
	if err != nil {
		t.Fatalf("error creating channel event client: %s", err)
	}
	defer eventProducer.Close()
	defer eventService.Stop()

	ccID1 := "mycc1"
	ccID2 := "mycc2"
	ccFilter1 := "event1"
	ccFilter2 := "event.*"
	event1 := "event1"
	event2 := "event2"
	event3 := "event3"

	if _, _, err1 := eventService.RegisterChaincodeEvent("", ccFilter1); err1 == nil {
		t.Fatal("expecting error registering for chaincode events without CC ID but got none")
	}
	if _, _, err2 := eventService.RegisterChaincodeEvent(ccID1, ""); err2 == nil {
		t.Fatal("expecting error registering for chaincode events without event filter but got none")
	}
	if _, _, err3 := eventService.RegisterChaincodeEvent(ccID1, ".(xxx"); err3 == nil {
		t.Fatal("expecting error registering for chaincode events with invalid (regular expression) event filter but got none")
	}
	reg1, _, err := eventService.RegisterChaincodeEvent(ccID1, ccFilter1)
	if err != nil {
		t.Fatalf("error registering for chaincode events: %s", err)
	}
	_, _, err = eventService.RegisterChaincodeEvent(ccID1, ccFilter1)
	if err == nil {
		t.Fatalf("expecting error registering multiple times for chaincode events: %s", err)
	}
	eventService.Unregister(reg1)

	reg1, eventch1, err := eventService.RegisterChaincodeEvent(ccID1, ccFilter1)
	if err != nil {
		t.Fatalf("error registering for block events: %s", err)
	}
	defer eventService.Unregister(reg1)

	reg2, eventch2, err := eventService.RegisterChaincodeEvent(ccID2, ccFilter2)
	if err != nil {
		t.Fatalf("error registering for chaincode events: %s", err)
	}
	defer eventService.Unregister(reg2)

	eventProducer.Ledger().NewFilteredBlock(
		channelID,
		servicemocks.NewFilteredTxWithCCEvent("txid1", ccID1, event1),
		servicemocks.NewFilteredTxWithCCEvent("txid2", ccID2, event2),
		servicemocks.NewFilteredTxWithCCEvent("txid3", ccID2, event3),
	)

	checkCCEvents(eventch1, t, ccID1, event1, eventch2, ccID2, event2, event3)
}

func checkCCEvents(eventch1 <-chan *fab.CCEvent, t *testing.T, ccID1 string, event1 string, eventch2 <-chan *fab.CCEvent, ccID2 string, event2 string, event3 string) {
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

// TestConcurrentEvents ensures that the channel event client is thread-safe
func TestConcurrentEvents(t *testing.T) {
	var numEvents uint = 1000
	channelID := "mychannel"

	eventService, eventProducer, err := newServiceWithMockProducer(
		[]options.Opt{
			dispatcher.WithEventConsumerBufferSize(numEvents),
			dispatcher.WithEventConsumerTimeout(time.Second),
		},
		withBlockLedger(sourceURL),
	)
	if err != nil {
		t.Fatalf("error creating channel event client: %s", err)
	}

	t.Run("Block Events", func(t *testing.T) {
		t.Parallel()
		if err := testConcurrentBlockEvents(channelID, numEvents, eventService, eventProducer); err != nil {
			t.Fatalf("error in testConcurrentBlockEvents: %s", err)
		}
	})
	t.Run("Filtered Block Events", func(t *testing.T) {
		t.Parallel()
		if err := testConcurrentFilteredBlockEvents(channelID, numEvents, eventService, eventProducer); err != nil {
			t.Fatalf("error in testConcurrentBlockEvents: %s", err)
		}
	})
	t.Run("Chaincode Events", func(t *testing.T) {
		t.Parallel()
		if err := testConcurrentCCEvents(channelID, numEvents, eventService, eventProducer); err != nil {
			t.Fatalf("error in testConcurrentBlockEvents: %s", err)
		}
	})
	t.Run("Tx Status Events", func(t *testing.T) {
		t.Parallel()
		if err := testConcurrentTxStatusEvents(channelID, numEvents, eventService, eventProducer); err != nil {
			t.Fatalf("error in testConcurrentBlockEvents: %s", err)
		}
	})
}

func testConcurrentBlockEvents(channelID string, numEvents uint, eventService fab.EventService, eventProducer *servicemocks.MockProducer) error {
	registration, eventch, err := eventService.RegisterBlockEvent()
	if err != nil {
		return errors.Errorf("error registering for block events: %s", err)
	}

	go func() {
		var i uint
		for i = 0; i < numEvents+10; i++ {
			eventProducer.Ledger().NewBlock(channelID,
				servicemocks.NewTransaction(fmt.Sprintf("txid_fb_%d", i), pb.TxValidationCode_VALID, cb.HeaderType_CONFIG_UPDATE),
			)
		}
	}()

	var numReceived uint
	done := false

	for !done {
		select {
		case _, ok := <-eventch:
			if !ok {
				done = true
			} else {
				numReceived++
				if numReceived == numEvents {
					// Unregister will close the event channel
					// and done will be set to true
					eventService.Unregister(registration)
				}
			}
		case <-time.After(5 * time.Second):
			if numReceived < numEvents {
				return errors.Errorf("Expected [%d] events but received [%d]", numEvents, numReceived)
			}
		}
	}

	return nil
}

func testConcurrentFilteredBlockEvents(channelID string, numEvents uint, eventService fab.EventService, conn *servicemocks.MockProducer) error {
	registration, eventch, err := eventService.RegisterFilteredBlockEvent()
	if err != nil {
		return errors.Errorf("error registering for filtered block events: %s", err)
	}
	defer eventService.Unregister(registration)

	sendNewBlock(numEvents, conn, channelID)

	var numReceived uint
	done := false

	for !done {
		select {
		case fbevent, ok := <-eventch:
			if !ok {
				done = true
			} else {
				if fbevent.FilteredBlock == nil {
					return errors.New("Expecting filtered block but got nil")
				}
				if fbevent.FilteredBlock.ChannelId != channelID {
					return errors.Errorf("Expecting channel [%s] but got [%s]", channelID, fbevent.FilteredBlock.ChannelId)
				}
				numReceived++
				if numReceived == numEvents {
					// Unregister will close the event channel and done will be set to true
					return nil
					// eventService.Unregister(registration)
				}
			}
		case <-time.After(5 * time.Second):
			if numReceived < numEvents {
				return errors.Errorf("Expected [%d] events but received [%d]", numEvents, numReceived)
			}
		}
	}

	return nil
}

func sendNewBlock(numEvents uint, conn *servicemocks.MockProducer, channelID string) {
	go func() {
		var i uint
		for _ = 0; i < numEvents; i++ {
			conn.Ledger().NewBlock(channelID,
				servicemocks.NewTransaction(
					fmt.Sprintf("txid_fb_%d", i), pb.TxValidationCode_VALID, cb.HeaderType_CONFIG_UPDATE),
			)
		}
	}()
}

func testConcurrentCCEvents(channelID string, numEvents uint, eventService fab.EventService, conn *servicemocks.MockProducer) error {
	ccID := "mycc1"
	ccFilter := "event.*"
	event1 := "event1"

	reg, eventch, err := eventService.RegisterChaincodeEvent(ccID, ccFilter)
	if err != nil {
		return errors.New("error registering for chaincode events")
	}

	go func() {
		var i uint
		for i = 0; i < numEvents+10; i++ {
			conn.Ledger().NewBlock(channelID,
				servicemocks.NewTransactionWithCCEvent(fmt.Sprintf("txid_cc_%d", i), pb.TxValidationCode_VALID, ccID, event1, nil),
			)
		}
	}()

	var numReceived uint
	done := false
	for !done {
		select {
		case _, ok := <-eventch:
			if !ok {
				done = true
			} else {
				numReceived++
			}
		case <-time.After(5 * time.Second):
			if numReceived < numEvents {
				return errors.Errorf("timed out waiting for [%d] CC events but received [%d]", numEvents, numReceived)
			}
		}

		if numReceived == numEvents {
			// Unregister will close the event channel and done will be set to true
			eventService.Unregister(reg)
		}
	}

	return nil
}

func testConcurrentTxStatusEvents(channelID string, numEvents uint, eventService fab.EventService, conn *servicemocks.MockProducer) error {
	var wg sync.WaitGroup

	wg.Add(int(numEvents))

	var errs []error
	var mutex sync.Mutex

	var receivedEvents uint32
	for i := 0; i < int(numEvents); i++ {
		txID := fmt.Sprintf("txid_tx_%d", i)
		go func() {
			defer wg.Done()

			reg, eventch, err := eventService.RegisterTxStatusEvent(txID)
			if err != nil {
				mutex.Lock()
				errs = append(errs, errors.New("Error registering for TxStatus event"))
				mutex.Unlock()
				return
			}
			defer eventService.Unregister(reg)

			conn.Ledger().NewBlock(channelID,
				servicemocks.NewTransaction(txID, pb.TxValidationCode_VALID, cb.HeaderType_ENDORSER_TRANSACTION),
			)

			select {
			case _, ok := <-eventch:
				if !ok {
					mutex.Lock()
					errs = append(errs, errors.New("unexpected closed channel"))
					mutex.Unlock()
				} else {
					atomic.AddUint32(&receivedEvents, 1)
				}
			case <-time.After(5 * time.Second):
				mutex.Lock()
				errs = append(errs, errors.New("timed out waiting for TxStatus event"))
				mutex.Unlock()
			}
		}()
	}

	wg.Wait()

	if len(errs) > 0 {
		return errors.Errorf("Received %d events and %d errors. First error %s", receivedEvents, len(errs), errs[0])
	}
	return nil
}

func checkTxStatusEvent(t *testing.T, event *fab.TxStatusEvent, expectedTxID string, expectedCode pb.TxValidationCode) {
	if event.TxID != expectedTxID {
		t.Fatalf("expecting event for TxID [%s] but received event for TxID [%s]", expectedTxID, event.TxID)
	}
	if event.TxValidationCode != expectedCode {
		t.Fatalf("expecting TxValidationCode [%s] but received [%s]", expectedCode, event.TxValidationCode)
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

func newServiceWithMockProducer(opts []options.Opt, pOpts ...producerOpt) (*Service, *servicemocks.MockProducer, error) {
	service := New(dispatcher.New(opts...), opts...)
	if err := service.Start(); err != nil {
		return nil, nil, err
	}

	eventch, err := service.Dispatcher().EventCh()
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

	return service, eventProducer, nil
}

func TestTransfer(t *testing.T) {
	t.Run("Transfer", func(t *testing.T) {
		testTransfer(t, func(service *Service) (fab.EventSnapshot, error) {
			return service.Transfer()
		})
	})
	t.Run("StopAndTransfer", func(t *testing.T) {
		testTransfer(t, func(service *Service) (fab.EventSnapshot, error) {
			return service.Transfer()
		})
	})
}

type transferFunc func(*Service) (fab.EventSnapshot, error)

func testTransfer(t *testing.T, transferFunc transferFunc) {
	channelID := "mychannel"
	eventService1, eventProducer1, err := newServiceWithMockProducer(defaultOpts, withBlockLedger(sourceURL))
	if err != nil {
		t.Fatalf("error creating channel event client: %s", err)
	}

	breg, beventch, err := eventService1.RegisterBlockEvent()
	require.NoErrorf(t, err, "error registering for block events")

	eventProducer1.Ledger().NewBlock(channelID)

	select {
	case _, ok := <-beventch:
		require.Truef(t, ok, "unexpected closed channel")
	case <-time.After(5 * time.Second):
		t.Fatal("timed out waiting for block event")
	}

	// Transfer all event registrations into a snapshot
	snapshot, err := transferFunc(eventService1)
	require.NoErrorf(t, err, "error in StopAndTransfer")
	require.NotNil(t, snapshot)
	eventProducer1.Close()

	// Use the snapshot with a new event service
	eventService2, eventProducer2, err := newServiceWithMockProducer(
		[]options.Opt{dispatcher.WithSnapshot(snapshot)},
		withBlockLedger(sourceURL))
	if err != nil {
		t.Fatalf("error creating channel event client: %s", err)
	}
	defer eventProducer2.Close()
	defer eventService2.Stop()

	eventProducer2.Ledger().NewBlock(channelID)

	select {
	case _, ok := <-beventch:
		require.Truef(t, ok, "unexpected closed channel")
	case <-time.After(5 * time.Second):
		t.Fatal("timed out waiting for block event")
	}

	eventService2.Unregister(breg)
}
