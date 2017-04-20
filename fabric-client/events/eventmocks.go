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
	"encoding/hex"
	"fmt"
	"testing"
	"time"

	consumer "github.com/hyperledger/fabric-sdk-go/fabric-client/events/consumer"
	"github.com/hyperledger/fabric-sdk-go/fabric-client/util"
	"github.com/hyperledger/fabric/bccsp"
	"github.com/hyperledger/fabric/bccsp/factory"
	ledger_util "github.com/hyperledger/fabric/core/ledger/util"
	fcConsumer "github.com/hyperledger/fabric/events/consumer"
	"github.com/hyperledger/fabric/protos/common"
	pb "github.com/hyperledger/fabric/protos/peer"
)

type mockEventClientMockEventRegistration struct {
	Action string
	ies    []*pb.Interest
}

type mockEventClient struct {
	PeerAddress string
	RegTimeout  time.Duration
	Adapter     fcConsumer.EventAdapter

	Started       int
	Stopped       int
	Registrations []mockEventClientMockEventRegistration

	events chan pb.Event
}

type mockEventClientFactory struct {
	clients []*mockEventClient
}

func (mecf *mockEventClientFactory) newEventsClient(peerAddress string, certificate string, serverHostOverride string, regTimeout time.Duration, adapter fcConsumer.EventAdapter) (consumer.EventsClient, error) {
	client := &mockEventClient{
		PeerAddress: peerAddress,
		RegTimeout:  regTimeout,
		Adapter:     adapter,
		events:      make(chan pb.Event),
	}
	mecf.clients = append(mecf.clients, client)
	return client, nil
}

// MockEvent mocks an event
func (mec *mockEventClient) MockEvent(msg *pb.Event) (bool, error) {
	if mec.Started > mec.Stopped {
		return mec.Adapter.Recv(msg)
	}

	mec.events <- *msg
	return true, nil
}

// RegisterAsync does not register anything anywhere but acts like all is well
func (mec *mockEventClient) RegisterAsync(ies []*pb.Interest) error {
	mec.Registrations = append(mec.Registrations, mockEventClientMockEventRegistration{
		Action: "register",
		ies:    ies,
	})
	return nil
}

// UnregisterAsync does not unregister anything anywhere but acts like all is well
func (mec *mockEventClient) UnregisterAsync(ies []*pb.Interest) error {
	mec.Registrations = append(mec.Registrations, mockEventClientMockEventRegistration{
		Action: "register",
		ies:    ies,
	})
	return nil
}

// Recv will return mock events sent to the event channel. Warning! This might block indefinately
func (mec *mockEventClient) Recv() (*pb.Event, error) {
	event := <-mec.events
	return &event, nil
}

// Start does not start anything
func (mec *mockEventClient) Start() error {
	mec.Started++
	return nil
}

// Stop does not stop anything
func (mec *mockEventClient) Stop() error {
	mec.Stopped++
	return nil
}

func createMockedEventHub(t *testing.T) (*eventHub, *mockEventClientFactory) {
	eventHub, ok := NewEventHub().(*eventHub)
	if !ok {
		t.Fatalf("Could not create eventHub")
		return nil, nil
	}

	var clientFactory mockEventClientFactory
	eventHub.eventsClientFactory = &clientFactory

	eventHub.SetPeerAddr("mock://mock", "", "")

	err := eventHub.Connect()
	if err != nil {
		t.Fatalf("Failed to connect: %v", err)
		return nil, nil
	}

	return eventHub, &clientFactory
}

func buildMockTxEvent(txID string) *pb.Event_Block {
	return &pb.Event_Block{
		Block: &common.Block{
			Header:   &common.BlockHeader{},
			Metadata: buildBlockMetadata(),
			Data: &common.BlockData{
				Data: [][]byte{util.MarshalOrPanic(buildEnvelope(txID))},
			},
		},
	}
}

func buildMockCCEvent(ccID string, eventName string) *pb.Event_ChaincodeEvent {
	return &pb.Event_ChaincodeEvent{
		ChaincodeEvent: &pb.ChaincodeEvent{
			ChaincodeId: ccID,
			EventName:   eventName,
		},
	}
}

func buildBlockMetadata() *common.BlockMetadata {
	return &common.BlockMetadata{
		Metadata: [][]byte{
			[]byte{},
			[]byte{},
			buildTransactionsFilterMetaDataBytes(),
			[]byte{},
		},
	}
}

func buildTransactionsFilterMetaDataBytes() []byte {
	return []byte(ledger_util.TxValidationFlags{uint8(pb.TxValidationCode_VALID)})
}

func buildEnvelope(txID string) *common.Envelope {
	return &common.Envelope{
		Payload: util.MarshalOrPanic(buildPayload(txID)),
	}
}

func buildPayload(txID string) *common.Payload {
	return &common.Payload{
		Header: &common.Header{
			ChannelHeader: util.MarshalOrPanic(buildChannelHeader(txID)),
		},
	}
}

func buildChannelHeader(txID string) *common.ChannelHeader {
	return &common.ChannelHeader{
		TxId:      txID,
		ChannelId: "testchannel",
	}
}

func generateTxID() string {
	nonce, err := util.GenerateRandomNonce()
	if err != nil {
		panic(fmt.Errorf("error generating nonce: %v", err))
	}
	digest, err := factory.GetDefault().Hash(
		nonce,
		&bccsp.SHA256Opts{})
	if err != nil {
		panic(fmt.Errorf("error hashing nonce: %v", err))
	}
	return hex.EncodeToString(digest)
}
