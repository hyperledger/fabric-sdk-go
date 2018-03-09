/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package dispatcher

import (
	"testing"
	"time"

	"github.com/hyperledger/fabric-sdk-go/pkg/context/api/fab"
	"github.com/hyperledger/fabric-sdk-go/pkg/fab/events/service/blockfilter"
	"github.com/hyperledger/fabric-sdk-go/pkg/fab/events/service/blockfilter/headertypefilter"
	servicemocks "github.com/hyperledger/fabric-sdk-go/pkg/fab/events/service/mocks"
	cb "github.com/hyperledger/fabric-sdk-go/third_party/github.com/hyperledger/fabric/protos/common"
	pb "github.com/hyperledger/fabric-sdk-go/third_party/github.com/hyperledger/fabric/protos/peer"
)

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

	dispatcherEventch <- servicemocks.NewBlockProducer().NewBlock(channelID)

	select {
	case _, ok := <-eventch:
		if !ok {
			t.Fatalf("unexpected closed channel")
		}
	case <-time.After(5 * time.Second):
		t.Fatalf("timed out waiting for block event")
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

	dispatcherEventch <- eventProducer.NewBlock(channelID,
		servicemocks.NewTransaction(txID1, txCode1, cb.HeaderType_CONFIG),
	)
	dispatcherEventch <- eventProducer.NewBlock(channelID,
		servicemocks.NewTransaction(txID2, txCode2, cb.HeaderType_CONFIG_UPDATE),
	)
	dispatcherEventch <- eventProducer.NewBlock(channelID,
		servicemocks.NewTransaction(txID2, txCode2, cb.HeaderType_ENDORSER_TRANSACTION),
	)

	numBlockEventsReceived := 0
	numBlockEventsExpected := 2
	numFilteredBlockEventsReceived := 0
	numFilteredBlockEventsExpected := 3

	done := false
	for !done {
		select {
		case _, ok := <-beventch:
			if !ok {
				t.Fatalf("unexpected closed channel")
			}
			numBlockEventsReceived++
		case _, ok := <-fbeventch:
			if !ok {
				t.Fatalf("unexpected closed channel")
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

	dispatcherEventch <- NewUnregisterEvent(breg)
	dispatcherEventch <- NewUnregisterEvent(fbreg)

	stopResp := make(chan error)
	dispatcherEventch <- NewStopEvent(stopResp)
	if err := <-stopResp; err != nil {
		t.Fatalf("Error stopping dispatcher: %s", err)
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

	dispatcherEventch <- servicemocks.NewBlockProducer().NewFilteredBlock(
		channelID,
		servicemocks.NewFilteredTx(txID1, txCode1),
		servicemocks.NewFilteredTx(txID2, txCode2),
	)

	select {
	case fbevent, ok := <-fbeventch:
		if !ok {
			t.Fatalf("unexpected closed channel")
		}
		if fbevent.FilteredBlock == nil {
			t.Fatalf("Expecting filtered block but got nil")
		}
		if fbevent.FilteredBlock.ChannelId != channelID {
			t.Fatalf("Expecting channel [%s] but got [%s]", channelID, fbevent.FilteredBlock.ChannelId)
		}
	case <-time.After(5 * time.Second):
		t.Fatalf("timed out waiting for filtered block event")
	}

	dispatcherEventch <- NewUnregisterEvent(reg)

	stopResp := make(chan error)
	dispatcherEventch <- NewStopEvent(stopResp)
	if err := <-stopResp; err != nil {
		t.Fatalf("Error stopping dispatcher: %s", err)
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

	dispatcherEventch <- servicemocks.NewBlockProducer().NewBlock(channelID,
		servicemocks.NewTransaction(txID1, txCode1, cb.HeaderType_CONFIG),
		servicemocks.NewTransaction(txID2, txCode2, cb.HeaderType_ENDORSER_TRANSACTION),
	)

	numReceived := 0
	numExpected := 2

	done := false
	for !done {
		select {
		case fbevent, ok := <-fbeventch:
			if !ok {
				t.Fatalf("unexpected closed channel")
			}
			if fbevent.FilteredBlock == nil {
				t.Fatalf("Expecting filtered block but got nil")
			}
			if fbevent.FilteredBlock.ChannelId != channelID {
				t.Fatalf("Expecting channel [%s] but got [%s]", channelID, fbevent.FilteredBlock.ChannelId)
			}
			numReceived++
		case _, ok := <-beventch:
			if !ok {
				t.Fatalf("unexpected closed channel")
			}
			numReceived++
		case <-time.After(2 * time.Second):
			if numReceived != numExpected {
				t.Fatalf("Expecting %d events but got %d", numExpected, numReceived)
			}
			done = true
		}
	}

	dispatcherEventch <- NewUnregisterEvent(breg)
	dispatcherEventch <- NewUnregisterEvent(fbreg)

	stopResp := make(chan error)
	dispatcherEventch <- NewStopEvent(stopResp)
	if err := <-stopResp; err != nil {
		t.Fatalf("Error stopping dispatcher: %s", err)
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
	case err := <-errch:
		t.Fatalf("error registering for TxStatus events: %s", err)
	}

	eventch = make(chan *fab.TxStatusEvent, 10)
	dispatcherEventch <- NewRegisterTxStatusEvent(txID1, eventch, regch, errch)

	select {
	case <-regch:
		t.Fatalf("expecting error registering multiple times for TxStatus events but got registration")
	case err = <-errch:
	}

	if err == nil {
		t.Fatalf("expecting error registering multiple times for TxStatus events")
	}

	dispatcherEventch <- NewUnregisterEvent(reg1)
	time.Sleep(100 * time.Millisecond)

	eventch1 := make(chan *fab.TxStatusEvent, 10)
	dispatcherEventch <- NewRegisterTxStatusEvent(txID1, eventch1, regch, errch)

	select {
	case reg1 = <-regch:
	case err := <-errch:
		t.Fatalf("error registering for TxStatus events: %s", err)
	}

	eventch2 := make(chan *fab.TxStatusEvent, 10)
	dispatcherEventch <- NewRegisterTxStatusEvent(txID2, eventch2, regch, errch)

	var reg2 fab.Registration
	select {
	case reg2 = <-regch:
	case err := <-errch:
		t.Fatalf("error registering for TxStatus events: %s", err)
	}

	dispatcherEventch <- servicemocks.NewBlockProducer().NewFilteredBlock(
		channelID,
		servicemocks.NewFilteredTx(txID1, txCode1),
		servicemocks.NewFilteredTx(txID2, txCode2),
	)

	numExpected := 2
	numReceived := 0

	for {
		select {
		case event, ok := <-eventch1:
			if !ok {
				t.Fatalf("unexpected closed channel")
			} else {
				checkTxStatusEvent(t, event, txID1, txCode1)
				numReceived++
			}
		case event, ok := <-eventch2:
			if !ok {
				t.Fatalf("unexpected closed channel")
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

	dispatcherEventch <- NewUnregisterEvent(reg1)
	dispatcherEventch <- NewUnregisterEvent(reg2)

	stopResp := make(chan error)
	dispatcherEventch <- NewStopEvent(stopResp)
	if err := <-stopResp; err != nil {
		t.Fatalf("Error stopping dispatcher: %s", err)
	}
}

func TestCCEvents(t *testing.T) {
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
	dispatcherEventch <- NewRegisterChaincodeEvent(ccID1, ccFilter1, eventch, fbrespch, errch)

	var reg1 fab.Registration
	select {
	case reg1 = <-fbrespch:
	case err := <-errch:
		t.Fatalf("error registering for chaincode events: %s", err)
	}

	eventch = make(chan *fab.CCEvent, 10)
	dispatcherEventch <- NewRegisterChaincodeEvent(ccID1, ccFilter1, eventch, fbrespch, errch)

	select {
	case reg1 = <-fbrespch:
		t.Fatalf("expecting error registering multiple times for chaincode events but got registration")
	case err = <-errch:
	}

	if err == nil {
		t.Fatalf("expecting error registering multiple times for chaincode events")
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

	var reg2 fab.Registration
	select {
	case reg2 = <-fbrespch:
	case err := <-errch:
		t.Fatalf("error registering for chaincode events: %s", err)
	}

	dispatcherEventch <- servicemocks.NewBlockProducer().NewFilteredBlock(
		channelID,
		servicemocks.NewFilteredTxWithCCEvent("txid1", ccID1, event1),
		servicemocks.NewFilteredTxWithCCEvent("txid2", ccID2, event2),
		servicemocks.NewFilteredTxWithCCEvent("txid3", ccID2, event3),
	)

	numExpected := 3
	numReceived := 0

	for {
		select {
		case event, ok := <-eventch1:
			if !ok {
				t.Fatalf("unexpected closed channel")
			} else {
				checkCCEvent(t, event, ccID1, event1)
				numReceived++
			}
		case event, ok := <-eventch2:
			if !ok {
				t.Fatalf("unexpected closed channel")
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

	dispatcherEventch <- NewUnregisterEvent(reg1)
	dispatcherEventch <- NewUnregisterEvent(reg2)

	stopResp := make(chan error)
	dispatcherEventch <- NewStopEvent(stopResp)
	if err := <-stopResp; err != nil {
		t.Fatalf("Error stopping dispatcher: %s", err)
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

	select {
	case regInfo, ok := <-eventch:
		if !ok {
			t.Fatalf("unexpected closed channel")
		}
		if regInfo.TotalRegistrations != 2 {
			t.Fatalf("expecting total registrations to be [%d] but received [%d]", 2, regInfo.TotalRegistrations)
		}
		if regInfo.NumBlockRegistrations != 1 {
			t.Fatalf("expecting number of block registrations to be [%d] but received [%d]", 1, regInfo.NumBlockRegistrations)
		}
		if regInfo.NumFilteredBlockRegistrations != 1 {
			t.Fatalf("expecting number of filtered block registrations to be [%d] but received [%d]", 1, regInfo.NumFilteredBlockRegistrations)
		}
	case <-time.After(5 * time.Second):
		t.Fatalf("timed out waiting for registration info")
	}

	dispatcherEventch <- NewUnregisterEvent(fbreg)
	dispatcherEventch <- NewRegistrationInfoEvent(eventch)

	select {
	case regInfo, ok := <-eventch:
		if !ok {
			t.Fatalf("unexpected closed channel")
		}
		if regInfo.NumBlockRegistrations != 1 {
			t.Fatalf("expecting number of block registrations to be [%d] but received [%d]", 1, regInfo.NumBlockRegistrations)
		}
		if regInfo.NumFilteredBlockRegistrations != 0 {
			t.Fatalf("expecting number of filtered block registrations to be [%d] but received [%d]", 0, regInfo.NumFilteredBlockRegistrations)
		}
	case <-time.After(5 * time.Second):
		t.Fatalf("timed out waiting for registration info")
	}

	dispatcherEventch <- NewUnregisterEvent(breg)
	dispatcherEventch <- NewRegistrationInfoEvent(eventch)

	select {
	case regInfo, ok := <-eventch:
		if !ok {
			t.Fatalf("unexpected closed channel")
		}
		if regInfo.TotalRegistrations != 0 {
			t.Fatalf("expecting total registrations to be [%d] but received [%d]", 1, regInfo.TotalRegistrations)
		}
		if regInfo.NumBlockRegistrations != 0 {
			t.Fatalf("expecting number of block registrations to be [%d] but received [%d]", 0, regInfo.NumBlockRegistrations)
		}
		if regInfo.NumFilteredBlockRegistrations != 0 {
			t.Fatalf("expecting number of filtered block registrations to be [%d] but received [%d]", 0, regInfo.NumFilteredBlockRegistrations)
		}
	case <-time.After(5 * time.Second):
		t.Fatalf("timed out waiting for registration info")
	}

	stopResp := make(chan error)
	dispatcherEventch <- NewStopEvent(stopResp)
	if err := <-stopResp; err != nil {
		t.Fatalf("Error stopping dispatcher: %s", err)
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
