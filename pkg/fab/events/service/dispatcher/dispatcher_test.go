/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package dispatcher

import (
	"bytes"
	"testing"
	"time"

	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/fab"
	"github.com/hyperledger/fabric-sdk-go/pkg/fab/events/service/blockfilter"
	"github.com/hyperledger/fabric-sdk-go/pkg/fab/events/service/blockfilter/headertypefilter"
	servicemocks "github.com/hyperledger/fabric-sdk-go/pkg/fab/events/service/mocks"
	cb "github.com/hyperledger/fabric-sdk-go/third_party/github.com/hyperledger/fabric/protos/common"
	pb "github.com/hyperledger/fabric-sdk-go/third_party/github.com/hyperledger/fabric/protos/peer"
)

var sourceURL = "localhost:9051"

func TestInvalidUnregister(t *testing.T) {
	dispatcher := New()
	if err := dispatcher.Start(); err != nil {
		t.Fatalf("Error starting dispatcher: %s", err)
	}

	dispatcherEventch, err := dispatcher.EventCh()
	if err != nil {
		t.Fatalf("Error getting event channel from dispatcher: %s", err)
	}

	// Make sure the client doesn't panic with invalid registration
	dispatcherEventch <- NewUnregisterEvent("invalid registration")
}

func TestBlockEvents(t *testing.T) {
	channelID := "testchannel"
	dispatcher := New(
		WithEventConsumerBufferSize(100),
		WithEventConsumerTimeout(2*time.Second),
	)
	if err := dispatcher.Start(); err != nil {
		t.Fatalf("Error starting dispatcher: %s", err)
	}

	dispatcherEventch, err := dispatcher.EventCh()
	if err != nil {
		t.Fatalf("Error getting event channel from dispatcher: %s", err)
	}

	eventch := make(chan *fab.BlockEvent, 10)
	regch := make(chan fab.Registration)
	errch := make(chan error)

	dispatcherEventch <- NewRegisterBlockEvent(blockfilter.AcceptAny, eventch, regch, errch)

	var reg fab.Registration
	select {
	case reg = <-regch:
	case err := <-errch:
		t.Fatalf("Error registering for block events: %s", err)
	}

	dispatcherEventch <- NewBlockEvent(servicemocks.NewBlockProducer().NewBlock(channelID), sourceURL)

	select {
	case event, ok := <-eventch:
		if !ok {
			t.Fatalf("unexpected closed channel")
		}
		if event.SourceURL != sourceURL {
			t.Fatalf("expecting source URL [%s] but got [%s]", sourceURL, event.SourceURL)
		}
	case <-time.After(5 * time.Second):
		t.Fatal("timed out waiting for block event")
	}

	dispatcherEventch <- NewUnregisterEvent(reg)

	stopResp := make(chan error)
	dispatcherEventch <- NewStopEvent(stopResp)
	if err := <-stopResp; err != nil {
		t.Fatalf("Error stopping dispatcher: %s", err)
	}
}

func TestBlockEventsWithFilter(t *testing.T) {
	channelID := "testchannel"
	dispatcher := New()
	if err := dispatcher.Start(); err != nil {
		t.Fatalf("Error starting dispatcher: %s", err)
	}

	dispatcherEventch, err := dispatcher.EventCh()
	if err != nil {
		t.Fatalf("Error getting event channel from dispatcher: %s", err)
	}

	regch := make(chan fab.Registration)
	errch := make(chan error)

	beventch := make(chan *fab.BlockEvent, 10)
	dispatcherEventch <- NewRegisterBlockEvent(headertypefilter.New(cb.HeaderType_CONFIG, cb.HeaderType_CONFIG_UPDATE), beventch, regch, errch)

	var breg fab.Registration
	select {
	case breg = <-regch:
	case err := <-errch:
		t.Fatalf("Error registering for block events: %s", err)
	}

	fbeventch := make(chan *fab.FilteredBlockEvent, 10)
	dispatcherEventch <- NewRegisterFilteredBlockEvent(fbeventch, regch, errch)

	var fbreg fab.Registration
	select {
	case breg = <-regch:
	case err := <-errch:
		t.Fatalf("Error registering for filtered block events: %s", err)
	}

	txID1 := "1234"
	txCode1 := pb.TxValidationCode_VALID
	txID2 := "5678"
	txCode2 := pb.TxValidationCode_ENDORSEMENT_POLICY_FAILURE

	eventProducer := servicemocks.NewBlockProducer()

	dispatcherEventch <- NewBlockEvent(eventProducer.NewBlock(channelID,
		servicemocks.NewTransaction(txID1, txCode1, cb.HeaderType_CONFIG)), sourceURL,
	)
	dispatcherEventch <- NewBlockEvent(eventProducer.NewBlock(channelID,
		servicemocks.NewTransaction(txID2, txCode2, cb.HeaderType_CONFIG_UPDATE)), sourceURL,
	)
	dispatcherEventch <- NewBlockEvent(eventProducer.NewBlock(channelID,
		servicemocks.NewTransaction(txID2, txCode2, cb.HeaderType_ENDORSER_TRANSACTION)), sourceURL,
	)

	numBlockEventsReceived := 0
	numBlockEventsExpected := 2
	numFilteredBlockEventsReceived := 0
	numFilteredBlockEventsExpected := 3

	checkBlockEventsWithFilter(beventch, t, numBlockEventsReceived, fbeventch, numFilteredBlockEventsReceived, numBlockEventsExpected, numFilteredBlockEventsExpected)

	dispatcherEventch <- NewUnregisterEvent(breg)
	dispatcherEventch <- NewUnregisterEvent(fbreg)

	stopResp := make(chan error)
	dispatcherEventch <- NewStopEvent(stopResp)
	if err := <-stopResp; err != nil {
		t.Fatalf("Error stopping dispatcher: %s", err)
	}
}

func checkBlockEventsWithFilter(beventch chan *fab.BlockEvent, t *testing.T, numBlockEventsReceived int, fbeventch chan *fab.FilteredBlockEvent, numFilteredBlockEventsReceived int, numBlockEventsExpected int, numFilteredBlockEventsExpected int) {
	done := false
	for !done {
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
			done = true
		}
	}
}

func TestFilteredBlockEvents(t *testing.T) {
	channelID := "testchannel"
	dispatcher := New()
	if err := dispatcher.Start(); err != nil {
		t.Fatalf("Error starting dispatcher: %s", err)
	}

	dispatcherEventch, err := dispatcher.EventCh()
	if err != nil {
		t.Fatalf("Error getting event channel from dispatcher: %s", err)
	}

	regch := make(chan fab.Registration)
	errch := make(chan error)
	fbeventch := make(chan *fab.FilteredBlockEvent, 10)
	dispatcherEventch <- NewRegisterFilteredBlockEvent(fbeventch, regch, errch)

	var reg fab.Registration
	select {
	case reg = <-regch:
	case err := <-errch:
		t.Fatalf("Error registering for filtered block events: %s", err)
	}

	txID1 := "1234"
	txCode1 := pb.TxValidationCode_VALID
	txID2 := "5678"
	txCode2 := pb.TxValidationCode_ENDORSEMENT_POLICY_FAILURE

	dispatcherEventch <- NewFilteredBlockEvent(servicemocks.NewBlockProducer().NewFilteredBlock(
		channelID,
		servicemocks.NewFilteredTx(txID1, txCode1),
		servicemocks.NewFilteredTx(txID2, txCode2),
	), sourceURL)

	checkFbEvent(fbeventch, t, channelID)

	dispatcherEventch <- NewUnregisterEvent(reg)

	stopResp := make(chan error)
	dispatcherEventch <- NewStopEvent(stopResp)
	if err := <-stopResp; err != nil {
		t.Fatalf("Error stopping dispatcher: %s", err)
	}
}

func checkFbEvent(fbeventch chan *fab.FilteredBlockEvent, t *testing.T, channelID string) {
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
		if fbevent.SourceURL != sourceURL {
			t.Fatalf("expecting source URL [%s] but got [%s]", sourceURL, fbevent.SourceURL)
		}
	case <-time.After(5 * time.Second):
		t.Fatal("timed out waiting for filtered block event")
	}
}

func TestBlockAndFilteredBlockEvents(t *testing.T) {
	channelID := "testchannel"
	dispatcher := New()
	if err := dispatcher.Start(); err != nil {
		t.Fatalf("Error starting dispatcher: %s", err)
	}

	dispatcherEventch, err := dispatcher.EventCh()
	if err != nil {
		t.Fatalf("Error getting event channel from dispatcher: %s", err)
	}

	errch := make(chan error)
	regch := make(chan fab.Registration)

	beventch := make(chan *fab.BlockEvent, 10)
	dispatcherEventch <- NewRegisterBlockEvent(blockfilter.AcceptAny, beventch, regch, errch)

	var breg fab.Registration
	select {
	case breg = <-regch:
	case err := <-errch:
		t.Fatalf("Error registering for block events: %s", err)
	}

	fbeventch := make(chan *fab.FilteredBlockEvent, 10)
	dispatcherEventch <- NewRegisterFilteredBlockEvent(fbeventch, regch, errch)

	var fbreg fab.Registration
	select {
	case fbreg = <-regch:
	case err := <-errch:
		t.Fatalf("Error registering for filtered block events: %s", err)
	}

	txID1 := "1234"
	txCode1 := pb.TxValidationCode_VALID
	txID2 := "5678"
	txCode2 := pb.TxValidationCode_ENDORSEMENT_POLICY_FAILURE

	dispatcherEventch <- NewBlockEvent(servicemocks.NewBlockProducer().NewBlock(channelID,
		servicemocks.NewTransaction(txID1, txCode1, cb.HeaderType_CONFIG),
		servicemocks.NewTransaction(txID2, txCode2, cb.HeaderType_ENDORSER_TRANSACTION),
	), sourceURL)

	numReceived := 0
	numExpected := 2

	checkBlockAndFilteredBlockEvents(fbeventch, t, channelID, numReceived, beventch, numExpected)

	dispatcherEventch <- NewUnregisterEvent(breg)
	dispatcherEventch <- NewUnregisterEvent(fbreg)

	stopResp := make(chan error)
	dispatcherEventch <- NewStopEvent(stopResp)
	if err := <-stopResp; err != nil {
		t.Fatalf("Error stopping dispatcher: %s", err)
	}
}

func checkBlockAndFilteredBlockEvents(fbeventch chan *fab.FilteredBlockEvent, t *testing.T, channelID string, numReceived int, beventch chan *fab.BlockEvent, numExpected int) {
	done := false
	for !done {
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
			done = true
		}
	}
}

func TestTxStatusEvents(t *testing.T) {
	channelID := "testchannel"
	dispatcher := New()
	if err := dispatcher.Start(); err != nil {
		t.Fatalf("Error starting dispatcher: %s", err)
	}

	dispatcherEventch, err := dispatcher.EventCh()
	if err != nil {
		t.Fatalf("Error getting event channel from dispatcher: %s", err)
	}

	txID1 := "1234"
	txCode1 := pb.TxValidationCode_VALID
	txID2 := "5678"
	txCode2 := pb.TxValidationCode_ENDORSEMENT_POLICY_FAILURE

	regch := make(chan fab.Registration)
	errch := make(chan error)

	eventch := make(chan *fab.TxStatusEvent, 10)
	dispatcherEventch <- NewRegisterTxStatusEvent(txID1, eventch, regch, errch)

	var reg1 fab.Registration
	select {
	case reg1 = <-regch:
	case err1 := <-errch:
		t.Fatalf("error registering for TxStatus events: %s", err1)
	}

	eventch = make(chan *fab.TxStatusEvent, 10)
	dispatcherEventch <- NewRegisterTxStatusEvent(txID1, eventch, regch, errch)

	select {
	case <-regch:
		t.Fatal("expecting error registering multiple times for TxStatus events but got registration")
	case err = <-errch:
	}

	if err == nil {
		t.Fatal("expecting error registering multiple times for TxStatus events")
	}

	dispatcherEventch <- NewUnregisterEvent(reg1)
	time.Sleep(100 * time.Millisecond)

	eventch1, dispatcherEventch, reg1 := registerEvent(dispatcherEventch, txID1, regch, errch, t)

	eventch2, dispatcherEventch, reg2 := registerEvent(dispatcherEventch, txID2, regch, errch, t)

	fblockEvent := NewFilteredBlockEvent(servicemocks.NewBlockProducer().NewFilteredBlock(
		channelID,
		servicemocks.NewFilteredTx(txID1, txCode1),
		servicemocks.NewFilteredTx(txID2, txCode2),
	), sourceURL)
	dispatcherEventch <- fblockEvent

	checkTxStatusEvents(fblockEvent, eventch1, t, txID1, txCode1, eventch2, txID2, txCode2)

	dispatcherEventch <- NewUnregisterEvent(reg1)
	dispatcherEventch <- NewUnregisterEvent(reg2)

	stopResp := make(chan error)
	dispatcherEventch <- NewStopEvent(stopResp)
	if err := <-stopResp; err != nil {
		t.Fatalf("Error stopping dispatcher: %s", err)
	}
}

func registerEvent(dispatcherEventch chan<- interface{}, txID string, regch chan fab.Registration, errch chan error, t *testing.T) (chan *fab.TxStatusEvent, chan<- interface{}, fab.Registration) {
	eventch := make(chan *fab.TxStatusEvent, 10)
	dispatcherEventch <- NewRegisterTxStatusEvent(txID, eventch, regch, errch)
	var reg fab.Registration
	select {
	case reg = <-regch:
	case err := <-errch:
		t.Fatalf("error registering for TxStatus events: %s", err)
	}
	return eventch, dispatcherEventch, reg
}

func checkTxStatusEvents(fblockEvent *fab.FilteredBlockEvent, eventch1 chan *fab.TxStatusEvent, t *testing.T, txID1 string, txCode1 pb.TxValidationCode, eventch2 chan *fab.TxStatusEvent, txID2 string, txCode2 pb.TxValidationCode) {
	expectedBlockNumber := fblockEvent.FilteredBlock.Number
	numExpected := 2
	numReceived := 0
	for {
		select {
		case event, ok := <-eventch1:
			numReceived = checkEventCh1(ok, t, event, txID1, txCode1, numReceived, expectedBlockNumber)
		case event, ok := <-eventch2:
			if !ok {
				t.Fatal("unexpected closed channel")
			} else {
				checkTxStatusEvent(t, event, txID2, txCode2)
				numReceived++
			}
			if event.SourceURL != sourceURL {
				t.Fatalf("expecting source URL [%s] but got [%s]", sourceURL, event.SourceURL)
			}
		case <-time.After(5 * time.Second):
			t.Fatalf("timed out waiting for [%d] TxStatus events. Only received [%d]", numExpected, numReceived)
		}

		if numReceived >= numExpected {
			break
		}
	}
	if numReceived != numExpected {
		t.Fatalf("expecting [%d] TxStatus events but got [%d]", numExpected, numReceived)
	}
}

func checkEventCh1(ok bool, t *testing.T, event *fab.TxStatusEvent, txID1 string, txCode1 pb.TxValidationCode, numReceived int, expectedBlockNumber uint64) int {
	if !ok {
		t.Fatal("unexpected closed channel")
	} else {
		checkTxStatusEvent(t, event, txID1, txCode1)
		numReceived++
	}
	if event.SourceURL != sourceURL {
		t.Fatalf("expecting source URL [%s] but got [%s]", sourceURL, event.SourceURL)
	}
	if event.BlockNumber != expectedBlockNumber {
		t.Fatalf("expecting block number [%d] but got [%d]", expectedBlockNumber, event.BlockNumber)
	}
	return numReceived
}

func TestCCEventsUnfiltered(t *testing.T) {
	channelID := "testchannel"
	dispatcher := New()
	if err := dispatcher.Start(); err != nil {
		t.Fatalf("Error starting dispatcher: %s", err)
	}

	dispatcherEventch, err := dispatcher.EventCh()
	if err != nil {
		t.Fatalf("Error getting event channel from dispatcher: %s", err)
	}

	ccID1 := "mycc1"
	ccID2 := "mycc2"
	ccFilter1 := "event1"
	ccFilter2 := "event2"
	event1 := "event1"
	event2 := "event2"
	payload1 := []byte("payload1")
	payload2 := []byte("payload2")

	errch := make(chan error)
	fbrespch := make(chan fab.Registration)
	eventch := make(chan *fab.CCEvent, 10)
	dispatcherEventch <- NewRegisterChaincodeEvent(ccID1, ccFilter1, eventch, fbrespch, errch)

	reg1 := getRegistration(fbrespch, errch, t)

	eventch = make(chan *fab.CCEvent, 10)
	dispatcherEventch <- NewRegisterChaincodeEvent(ccID1, ccFilter1, eventch, fbrespch, errch)

	select {
	case reg1 = <-fbrespch:
		t.Fatal("expecting error registering multiple times for chaincode events but got registration")
	case err = <-errch:
	}

	if err == nil {
		t.Fatal("expecting error registering multiple times for chaincode events")
	}

	dispatcherEventch <- NewUnregisterEvent(reg1)

	eventch1 := make(chan *fab.CCEvent, 10)
	dispatcherEventch <- NewRegisterChaincodeEvent(ccID1, ccFilter1, eventch1, fbrespch, errch)

	select {
	case reg1 = <-fbrespch:
	case err := <-errch:
		t.Fatalf("error registering for chaincode events: %s", err)
	}

	eventch2 := make(chan *fab.CCEvent, 10)
	dispatcherEventch <- NewRegisterChaincodeEvent(ccID2, ccFilter2, eventch2, fbrespch, errch)

	reg2 := getRegistration(fbrespch, errch, t)

	blockEvent := NewBlockEvent(
		servicemocks.NewBlockProducer().NewBlock(
			channelID,
			servicemocks.NewTransactionWithCCEvent("txid1", pb.TxValidationCode_VALID, ccID1, event1, payload1),
			servicemocks.NewTransactionWithCCEvent("txid2", pb.TxValidationCode_VALID, ccID2, event2, payload2),
		), sourceURL)

	dispatcherEventch <- blockEvent

	checkCCEventsUnfiltered(blockEvent, eventch1, t, ccID1, payload1, event1, eventch2, ccID2, payload2, event2)

	dispatcherEventch <- NewUnregisterEvent(reg1)
	dispatcherEventch <- NewUnregisterEvent(reg2)

	stopResp := make(chan error)
	dispatcherEventch <- NewStopEvent(stopResp)
	if err := <-stopResp; err != nil {
		t.Fatalf("Error stopping dispatcher: %s", err)
	}
}

func getRegistration(fbrespch chan fab.Registration, errch chan error, t *testing.T) fab.Registration {
	var reg fab.Registration
	select {
	case reg = <-fbrespch:
	case err := <-errch:
		t.Fatalf("error registering for chaincode events: %s", err)
	}
	return reg
}

func checkCCEventsUnfiltered(blockEvent *fab.BlockEvent, eventch1 chan *fab.CCEvent, t *testing.T, ccID1 string, payload1 []byte, event1 string, eventch2 chan *fab.CCEvent, ccID2 string, payload2 []byte, event2 string) {
	expectedBlockNumber := blockEvent.Block.Header.Number
	numExpected := 2
	numReceived := 0
	for {
		select {
		case event, ok := <-eventch1:
			numReceived = checkEvent1Unfiltered(ok, t, event, ccID1, payload1, event1, numReceived, expectedBlockNumber)
		case event, ok := <-eventch2:
			if !ok {
				t.Fatalf("unexpected closed channel")
			} else {
				checkCCEvent(t, event, ccID2, payload2, event2)
				numReceived++
			}
			if event.SourceURL != sourceURL {
				t.Fatalf("expecting source URL [%s] but got [%s]", sourceURL, event.SourceURL)
			}
			if event.BlockNumber != expectedBlockNumber {
				t.Fatalf("expecting block number [%d] but got [%d]", expectedBlockNumber, event.BlockNumber)
			}
		case <-time.After(5 * time.Second):
			t.Fatalf("timed out waiting for [%d] CC events. Only received [%d]", numExpected, numReceived)
		}

		if numReceived >= numExpected {
			break
		}
	}
	if numReceived != numExpected {
		t.Fatalf("expecting [%d] CC events but got [%d]", numExpected, numReceived)
	}
}

func checkEvent1Unfiltered(ok bool, t *testing.T, event *fab.CCEvent, ccID1 string, payload1 []byte, event1 string, numReceived int, expectedBlockNumber uint64) int {
	if !ok {
		t.Fatalf("unexpected closed channel")
	} else {
		checkCCEvent(t, event, ccID1, payload1, event1)
		numReceived++
	}
	if event.SourceURL != sourceURL {
		t.Fatalf("expecting source URL [%s] but got [%s]", sourceURL, event.SourceURL)
	}
	if event.BlockNumber != expectedBlockNumber {
		t.Fatalf("expecting block number [%d] but got [%d]", expectedBlockNumber, event.BlockNumber)
	}
	return numReceived
}

func TestCCEventsFiltered(t *testing.T) {
	channelID := "testchannel"
	dispatcher := New()
	if err := dispatcher.Start(); err != nil {
		t.Fatalf("Error starting dispatcher: %s", err)
	}

	dispatcherEventch, err := dispatcher.EventCh()
	if err != nil {
		t.Fatalf("Error getting event channel from dispatcher: %s", err)
	}

	ccID1 := "mycc1"
	ccID2 := "mycc2"
	ccFilter1 := "event1"
	ccFilter2 := "event.*"
	event1 := "event1"
	event2 := "event2"
	event3 := "event3"

	errch := make(chan error)
	fbrespch := make(chan fab.Registration)
	eventch := make(chan *fab.CCEvent, 10)
	dispatcherEventch, reg1 := regEvent(dispatcherEventch, ccID1, ccFilter1, eventch, fbrespch, errch, t)

	eventch = make(chan *fab.CCEvent, 10)
	dispatcherEventch <- NewRegisterChaincodeEvent(ccID1, ccFilter1, eventch, fbrespch, errch)

	select {
	case reg1 = <-fbrespch:
		t.Fatal("expecting error registering multiple times for chaincode events but got registration")
	case err = <-errch:
	}

	if err == nil {
		t.Fatal("expecting error registering multiple times for chaincode events")
	}

	dispatcherEventch <- NewUnregisterEvent(reg1)

	eventch1 := make(chan *fab.CCEvent, 10)
	dispatcherEventch <- NewRegisterChaincodeEvent(ccID1, ccFilter1, eventch1, fbrespch, errch)

	select {
	case reg1 = <-fbrespch:
	case err := <-errch:
		t.Fatalf("error registering for chaincode events: %s", err)
	}

	eventch2 := make(chan *fab.CCEvent, 10)
	dispatcherEventch, reg2 := regEvent(dispatcherEventch, ccID2, ccFilter2, eventch2, fbrespch, errch, t)

	dispatcherEventch <- NewFilteredBlockEvent(servicemocks.NewBlockProducer().NewFilteredBlock(
		channelID,
		servicemocks.NewFilteredTxWithCCEvent("txid1", ccID1, event1),
		servicemocks.NewFilteredTxWithCCEvent("txid2", ccID2, event2),
		servicemocks.NewFilteredTxWithCCEvent("txid3", ccID2, event3),
	), sourceURL)

	checkCCEventsFiltered(eventch1, t, ccID1, event1, eventch2, ccID2, event2, event3)

	dispatcherEventch <- NewUnregisterEvent(reg1)
	dispatcherEventch <- NewUnregisterEvent(reg2)

	stopResp := make(chan error)
	dispatcherEventch <- NewStopEvent(stopResp)
	if err := <-stopResp; err != nil {
		t.Fatalf("Error stopping dispatcher: %s", err)
	}
}

func regEvent(dispatcherEventch chan<- interface{}, ccID string, ccFilter string, eventch chan *fab.CCEvent, fbrespch chan fab.Registration, errch chan error, t *testing.T) (chan<- interface{}, fab.Registration) {
	dispatcherEventch <- NewRegisterChaincodeEvent(ccID, ccFilter, eventch, fbrespch, errch)
	var reg fab.Registration
	select {
	case reg = <-fbrespch:
	case err := <-errch:
		t.Fatalf("error registering for chaincode events: %s", err)
	}
	return dispatcherEventch, reg
}

func checkCCEventsFiltered(eventch1 chan *fab.CCEvent, t *testing.T, ccID1 string, event1 string, eventch2 chan *fab.CCEvent, ccID2 string, event2 string, event3 string) {
	numExpected := 3
	numReceived := 0
	for {
		select {
		case event, ok := <-eventch1:
			if !ok {
				t.Fatal("unexpected closed channel")
			} else {
				checkCCEvent(t, event, ccID1, nil, event1)
				numReceived++
			}
		case event, ok := <-eventch2:
			if !ok {
				t.Fatal("unexpected closed channel")
			} else {
				checkCCEvent(t, event, ccID2, nil, event2, event3)
				numReceived++
			}
		case <-time.After(5 * time.Second):
			t.Fatalf("timed out waiting for [%d] CC events. Only received [%d]", numExpected, numReceived)
		}

		if numReceived >= numExpected {
			break
		}
	}
	if numReceived != numExpected {
		t.Fatalf("expecting [%d] CC events but got [%d]", numExpected, numReceived)
	}
}

func TestRegistrationInfo(t *testing.T) {
	dispatcher := New()
	if err := dispatcher.Start(); err != nil {
		t.Fatalf("Error starting dispatcher: %s", err)
	}

	dispatcherEventch, err := dispatcher.EventCh()
	if err != nil {
		t.Fatalf("Error getting event channel from dispatcher: %s", err)
	}

	errch := make(chan error)

	regch := make(chan fab.Registration)
	fbeventch := make(chan *fab.FilteredBlockEvent, 10)
	dispatcherEventch <- NewRegisterFilteredBlockEvent(fbeventch, regch, errch)

	var fbreg fab.Registration
	select {
	case fbreg = <-regch:
	case err := <-errch:
		t.Fatalf("Error registering for filtered block events: %s", err)
	}

	beventch := make(chan *fab.BlockEvent, 10)
	dispatcherEventch <- NewRegisterBlockEvent(headertypefilter.New(cb.HeaderType_CONFIG, cb.HeaderType_CONFIG_UPDATE), beventch, regch, errch)

	var breg fab.Registration
	select {
	case breg = <-regch:
	case err := <-errch:
		t.Fatalf("Error registering for block events: %s", err)
	}

	eventch := make(chan *RegistrationInfo, 1)
	dispatcherEventch <- NewRegistrationInfoEvent(eventch)

	checkEvent(eventch, t, 2, 1, 1, true)

	dispatcherEventch <- NewUnregisterEvent(fbreg)
	dispatcherEventch <- NewRegistrationInfoEvent(eventch)

	checkEvent(eventch, t, 0, 1, 0, false)

	dispatcherEventch <- NewUnregisterEvent(breg)
	dispatcherEventch <- NewRegistrationInfoEvent(eventch)

	checkEvent(eventch, t, 0, 0, 0, true)

	stopResp := make(chan error)
	dispatcherEventch <- NewStopEvent(stopResp)
	if err := <-stopResp; err != nil {
		t.Fatalf("Error stopping dispatcher: %s", err)
	}
}

func checkEvent(eventch chan *RegistrationInfo, t *testing.T, totalRegistrations, numBlockRegistrations, numFilteredBlockRegistrations int, checkTotalRegistrations bool) {
	select {
	case regInfo, ok := <-eventch:
		if !ok {
			t.Fatal("unexpected closed channel")
		}
		if checkTotalRegistrations && regInfo.TotalRegistrations != totalRegistrations {
			t.Fatalf("expecting total registrations to be [%d] but received [%d]", totalRegistrations, regInfo.TotalRegistrations)
		}
		if regInfo.NumBlockRegistrations != numBlockRegistrations {
			t.Fatalf("expecting number of block registrations to be [%d] but received [%d]", numBlockRegistrations, regInfo.NumBlockRegistrations)
		}
		if regInfo.NumFilteredBlockRegistrations != numFilteredBlockRegistrations {
			t.Fatalf("expecting number of filtered block registrations to be [%d] but received [%d]", numFilteredBlockRegistrations, regInfo.NumFilteredBlockRegistrations)
		}
	case <-time.After(5 * time.Second):
		t.Fatal("timed out waiting for registration info")
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

func checkCCEvent(t *testing.T, event *fab.CCEvent, expectedCCID string, expectedPayload []byte, expectedEventNames ...string) {
	if event.ChaincodeID != expectedCCID {
		t.Fatalf("expecting event for CC [%s] but received event for CC [%s]", expectedCCID, event.ChaincodeID)
	}
	if !bytes.Equal(event.Payload, expectedPayload) {
		t.Fatalf("expecting payload [%s] but received payload [%s]", expectedPayload, event.Payload)
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
