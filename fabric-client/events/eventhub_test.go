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

package events

import (
	"sync/atomic"
	"testing"

	"fmt"

	"time"

	"reflect"

	"github.com/hyperledger/fabric/protos/common"
	pb "github.com/hyperledger/fabric/protos/peer"
)

func TestDeadlock(t *testing.T) {
	channelID := "testchannel"
	ccID := "testccid"

	eventHub, clientFactory := createMockedEventHub(t)
	if t.Failed() {
		return
	}

	fmt.Printf("EventHub Concurrency test\n")

	client := clientFactory.clients[0]
	if client == nil {
		t.Fatalf("No client")
	}

	threads := 20
	eventsPerThread := 100
	eventsSent := eventsPerThread * threads

	// The test should be done in milliseconds but if there's
	// a deadlock then we don't want it to hang
	timeout := 5 * time.Second

	// create a flood of TX events
	txCompletion := newMultiCompletionHandler(eventsSent, timeout)
	go flood(eventsPerThread, threads, func() {
		transactionID := generateTxID()
		received := newCompletionHandler(timeout)
		eventHub.RegisterTxEvent(transactionID, func(txID string, err error) {
			txCompletion.done()
			received.done()
		})

		go client.MockEvent(&pb.Event{
			Event: (&MockTxEventBuilder{
				TxID:      transactionID,
				ChannelID: channelID,
			}).Build(),
		})

		// Wait for the TX event and then unregister
		received.wait()
		eventHub.UnregisterTxEvent(transactionID)
	})

	// create a flood of CC events
	ccCompletion := newMultiCompletionHandler(eventsSent, timeout)
	go flood(eventsPerThread, threads, func() {
		eventName := generateTxID()
		received := newCompletionHandler(timeout)
		registration := eventHub.RegisterChaincodeEvent(ccID, eventName, func(event *ChaincodeEvent) {
			ccCompletion.done()
			received.done()
		})

		go client.MockEvent(&pb.Event{
			Event: (&MockCCEventBuilder{
				CCID:      ccID,
				EventName: eventName,
			}).Build(),
		})

		// Wait for the CC event and then unregister
		received.wait()
		eventHub.UnregisterChaincodeEvent(registration)
	})

	// Wait for all events to be received
	txCompletion.wait()
	ccCompletion.wait()

	if txCompletion.numDone() != eventsSent {
		t.Errorf("Sent %d Tx events but received %d - could indicate a deadlock", eventsSent, txCompletion.numDone())
	} else {
		fmt.Printf("Received all %d TX events.\n", txCompletion.numDone())
	}

	if ccCompletion.numDone() != eventsSent {
		t.Errorf("Sent %d CC events but received %d - could indicate a deadlock", eventsSent, ccCompletion.numDone())
	} else {
		fmt.Printf("Received all %d CC events.\n", ccCompletion.numDone())
	}
}

func TestChaincodeEvent(t *testing.T) {
	ccID := "someccid"
	eventName := "someevent"

	eventHub, clientFactory := createMockedEventHub(t)
	if t.Failed() {
		return
	}

	fmt.Printf("EventHub Chaincode event test\n")

	client := clientFactory.clients[0]
	if client == nil {
		t.Fatalf("No client")
	}

	eventReceived := make(chan *ChaincodeEvent)

	// Register for CC event
	registration := eventHub.RegisterChaincodeEvent(ccID, eventName, func(event *ChaincodeEvent) {
		eventReceived <- event
	})

	// Publish CC event
	go client.MockEvent(&pb.Event{
		Event: (&MockCCEventBuilder{
			CCID:      ccID,
			EventName: eventName,
		}).Build(),
	})

	// Wait for the CC event
	var event *ChaincodeEvent
	select {
	case event = <-eventReceived:
		eventHub.UnregisterChaincodeEvent(registration)
	case <-time.After(time.Second * 5):
		t.Fatalf("Timed out waiting for CC event")
	}

	// Check CC event
	if event.ChaincodeId != ccID {
		t.Fatalf("Expecting chaincode ID [%s] but got [%s]", ccID, event.ChaincodeId)
	}
	if event.EventName != eventName {
		t.Fatalf("Expecting event name [%s] but got [%s]", eventName, event.EventName)
	}
}

func TestChaincodeBlockEvent(t *testing.T) {
	channelID := "somechannelid"
	ccID := "someccid"
	eventName := "someevent"
	txID := generateTxID()

	eventHub, clientFactory := createMockedEventHub(t)
	if t.Failed() {
		return
	}

	client := clientFactory.clients[0]
	if client == nil {
		t.Fatalf("No client")
	}

	eventReceived := make(chan *ChaincodeEvent)

	// Register for CC event
	registration := eventHub.RegisterChaincodeEvent(ccID, eventName, func(event *ChaincodeEvent) {
		eventReceived <- event
	})

	// Publish CC event
	go client.MockEvent(&pb.Event{
		Event: (&MockCCBlockEventBuilder{
			CCID:      ccID,
			EventName: eventName,
			ChannelID: channelID,
			TxID:      txID,
		}).Build(),
	})

	// Wait for CC event
	var event *ChaincodeEvent
	select {
	case event = <-eventReceived:
		eventHub.UnregisterChaincodeEvent(registration)
	case <-time.After(time.Second * 5):
		t.Fatalf("Timed out waiting for CC event")
	}

	// Check CC event
	if event.ChannelID != channelID {
		t.Fatalf("Expecting channel ID [%s] but got [%s]", channelID, event.ChannelID)
	}
	if event.ChaincodeId != ccID {
		t.Fatalf("Expecting chaincode ID [%s] but got [%s]", ccID, event.ChaincodeId)
	}
	if event.EventName != eventName {
		t.Fatalf("Expecting event name [%s] but got [%s]", eventName, event.EventName)
	}
	if event.TxId == "" {
		t.Fatalf("Expecting TxID [%s] but got [%s]", txID, event.TxId)
	}
}

// completionHandler waits for a single event with a timeout
type completionHandler struct {
	completed chan bool
	timeout   time.Duration
}

// newCompletionHandler creates a new completionHandler
func newCompletionHandler(timeout time.Duration) *completionHandler {
	return &completionHandler{
		timeout:   timeout,
		completed: make(chan bool),
	}
}

// wait will wait until the task(s) has completed or until the timeout
func (c *completionHandler) wait() {
	select {
	case <-c.completed:
	case <-time.After(c.timeout):
	}
}

// done marks the task as completed
func (c *completionHandler) done() {
	c.completed <- true
}

// multiCompletionHandler waits for multiple tasks to complete
type multiCompletionHandler struct {
	completionHandler
	expected     int32
	numCompleted int32
}

// newMultiCompletionHandler creates a new multiCompletionHandler
func newMultiCompletionHandler(expected int, timeout time.Duration) *multiCompletionHandler {
	return &multiCompletionHandler{
		expected: int32(expected),
		completionHandler: completionHandler{
			timeout:   timeout,
			completed: make(chan bool),
		},
	}
}

// done marks a task as completed
func (c *multiCompletionHandler) done() {
	doneSoFar := atomic.AddInt32(&c.numCompleted, 1)
	if doneSoFar >= c.expected {
		c.completed <- true
	}
}

// numDone returns the nmber of tasks that have completed
func (c *multiCompletionHandler) numDone() int {
	return int(c.numCompleted)
}

// flood invokes the given function in the given number of threads,
// the given number of times per thread
func flood(invocationsPerThread int, threads int, f func()) {
	for t := 0; t < threads; t++ {
		go func() {
			for i := 0; i < invocationsPerThread; i++ {
				f()
			}
		}()
	}
}

func TestRegisterBlockEvent(t *testing.T) {
	eventHub, _ := createMockedEventHub(t)
	if t.Failed() {
		return
	}

	// Transaction callback is registered by default
	if len(eventHub.interestedEvents) != 1 || len(eventHub.blockRegistrants) != 1 {
		t.Fatalf("Transaction callback should be registered by default")
	}

	f1 := reflect.ValueOf(eventHub.txCallback)
	f2 := reflect.ValueOf(eventHub.blockRegistrants[0])

	if f1.Pointer() != f2.Pointer() {
		t.Fatalf("Registered callback is not txCallback")
	}

	eventHub.RegisterBlockEvent(testCallback)

	if len(eventHub.blockRegistrants) != 2 {
		t.Fatalf("Failed to add test callback for block event")
	}

	f1 = reflect.ValueOf(testCallback)
	f2 = reflect.ValueOf(eventHub.blockRegistrants[1])

	if f1.Pointer() != f2.Pointer() {
		t.Fatalf("Registered callback is not testCallback")
	}

	eventHub.UnregisterBlockEvent(testCallback)

	if len(eventHub.interestedEvents) != 1 || len(eventHub.blockRegistrants) != 1 {
		t.Fatalf("Failed to unregister testCallback")
	}

	eventHub.UnregisterBlockEvent(eventHub.txCallback)

	if len(eventHub.interestedEvents) != 0 || len(eventHub.blockRegistrants) != 0 {
		t.Fatalf("Failed to unregister txCallback")
	}

}

// private test callback to be executed on block event
func testCallback(block *common.Block) {
	fmt.Println("testCallback called on block")
}
