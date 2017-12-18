/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package events

import (
	"encoding/hex"
	"testing"
	"time"

	fab "github.com/hyperledger/fabric-sdk-go/api/apifabclient"
	"github.com/hyperledger/fabric-sdk-go/api/apitxn"
	"github.com/hyperledger/fabric-sdk-go/third_party/github.com/hyperledger/fabric/protos/common"
	pb "github.com/hyperledger/fabric-sdk-go/third_party/github.com/hyperledger/fabric/protos/peer"

	ledger_util "github.com/hyperledger/fabric-sdk-go/internal/github.com/hyperledger/fabric/core/ledger/util"
	fcConsumer "github.com/hyperledger/fabric-sdk-go/internal/github.com/hyperledger/fabric/events/consumer"
	"github.com/hyperledger/fabric-sdk-go/pkg/cryptosuite"
	"github.com/hyperledger/fabric-sdk-go/pkg/errors"
	client "github.com/hyperledger/fabric-sdk-go/pkg/fabric-client"
	internal "github.com/hyperledger/fabric-sdk-go/pkg/fabric-client/internal"
	"github.com/hyperledger/fabric-sdk-go/pkg/fabric-client/mocks"
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

func (mecf *mockEventClientFactory) newEventsClient(client fab.FabricClient, peerAddress string, certificate string, serverHostOverride string, regTimeout time.Duration, adapter fcConsumer.EventAdapter) (fab.EventsClient, error) {
	mec := &mockEventClient{
		PeerAddress: peerAddress,
		RegTimeout:  regTimeout,
		Adapter:     adapter,
		events:      make(chan pb.Event),
	}
	mecf.clients = append(mecf.clients, mec)
	return mec, nil
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

// Unregister does not unregister anything anywhere but acts like all is well
func (mec *mockEventClient) Unregister(ies []*pb.Interest) error {
	return mec.UnregisterAsync(ies)
}

// Recv will return mock events sent to the event channel. Warning! This might block indefinitely
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

func createMockedEventHub(t *testing.T) (*EventHub, *mockEventClientFactory) {
	eventHub, err := NewEventHub(client.NewClient(mocks.NewMockConfig()))
	if err != nil {
		t.Fatalf("Error creating event hub: %v", err)
	}

	var clientFactory mockEventClientFactory
	eventHub.eventsClientFactory = &clientFactory

	eventHub.SetPeerAddr("mock://mock", "", "")

	err = eventHub.Connect()
	if err != nil {
		t.Fatalf("Failed to connect: %v", err)
		return nil, nil
	}

	return eventHub, &clientFactory
}

// MockTxEventBuilder builds a mock TX event block
type MockTxEventBuilder struct {
	ChannelID string
	TxID      string
}

// MockCCEventBuilder builds a mock chaincode event
type MockCCEventBuilder struct {
	CCID      string
	EventName string
	Payload   []byte
}

// MockCCBlockEventBuilder builds a mock CC event block
type MockCCBlockEventBuilder struct {
	CCID      string
	EventName string
	ChannelID string
	TxID      string
	Payload   []byte
}

// Build builds a mock TX event block
func (b *MockTxEventBuilder) Build() *pb.Event_Block {
	return &pb.Event_Block{
		Block: &common.Block{
			Header:   &common.BlockHeader{},
			Metadata: b.buildBlockMetadata(),
			Data: &common.BlockData{
				Data: [][]byte{internal.MarshalOrPanic(b.buildEnvelope())},
			},
		},
	}
}

func (b *MockTxEventBuilder) buildBlockMetadata() *common.BlockMetadata {
	return &common.BlockMetadata{
		Metadata: [][]byte{
			[]byte{},
			[]byte{},
			b.buildTransactionsFilterMetaDataBytes(),
			[]byte{},
		},
	}
}

func (b *MockTxEventBuilder) buildTransactionsFilterMetaDataBytes() []byte {
	return []byte(ledger_util.TxValidationFlags{uint8(pb.TxValidationCode_VALID)})
}

// Build builds a mock chaincode event
func (b *MockCCEventBuilder) Build() *pb.Event_ChaincodeEvent {
	return &pb.Event_ChaincodeEvent{
		ChaincodeEvent: &pb.ChaincodeEvent{
			ChaincodeId: b.CCID,
			EventName:   b.EventName,
			Payload:     b.Payload,
		},
	}
}

func (b *MockTxEventBuilder) buildEnvelope() *common.Envelope {
	return &common.Envelope{
		Payload: internal.MarshalOrPanic(b.buildPayload()),
	}
}

func (b *MockTxEventBuilder) buildPayload() *common.Payload {
	return &common.Payload{
		Header: &common.Header{
			ChannelHeader: internal.MarshalOrPanic(b.buildChannelHeader()),
		},
	}
}

func (b *MockTxEventBuilder) buildChannelHeader() *common.ChannelHeader {
	return &common.ChannelHeader{
		TxId:      b.TxID,
		ChannelId: b.ChannelID,
	}
}

// Build builds a mock chaincode event block
func (b *MockCCBlockEventBuilder) Build() *pb.Event_Block {
	return b.BuildWithTxValidationFlag(true)
}

// BuildWithTxValidationFlag builds a mock chaincode event block with valid/invalid TxValidation Flag (set in the argument)
func (b *MockCCBlockEventBuilder) BuildWithTxValidationFlag(isValid bool) *pb.Event_Block {
	return &pb.Event_Block{
		Block: &common.Block{
			Header:   &common.BlockHeader{},
			Metadata: b.buildBlockMetadataWithValidFlag(isValid),
			Data: &common.BlockData{
				Data: [][]byte{internal.MarshalOrPanic(b.buildEnvelope())},
			},
		},
	}
}

func (b *MockCCBlockEventBuilder) buildBlockMetadata() *common.BlockMetadata {
	return b.buildBlockMetadataWithValidFlag(true)
}

func (b *MockCCBlockEventBuilder) buildBlockMetadataWithValidFlag(isValid bool) *common.BlockMetadata {
	return &common.BlockMetadata{
		Metadata: [][]byte{
			[]byte{},
			[]byte{},
			b.buildTransactionsFilterMetaDataBytesWithValidFlag(isValid),
			[]byte{},
		},
	}
}

func (b *MockCCBlockEventBuilder) buildEnvelope() *common.Envelope {
	return &common.Envelope{
		Payload: internal.MarshalOrPanic(b.buildPayload()),
	}
}

func (b *MockCCBlockEventBuilder) buildTransactionsFilterMetaDataBytes() []byte {
	return b.buildTransactionsFilterMetaDataBytesWithValidFlag(true)
}

func (b *MockCCBlockEventBuilder) buildTransactionsFilterMetaDataBytesWithValidFlag(isValidTx bool) []byte {
	if isValidTx {
		return []byte(ledger_util.TxValidationFlags{uint8(pb.TxValidationCode_VALID)})
	}
	// return transaction with any non valid flag
	return []byte(ledger_util.TxValidationFlags{uint8(pb.TxValidationCode_BAD_COMMON_HEADER)})
}

func (b *MockCCBlockEventBuilder) buildPayload() *common.Payload {
	logger.Debug("MockCCBlockEventBuilder.buildPayload")
	return &common.Payload{
		Header: &common.Header{
			ChannelHeader: internal.MarshalOrPanic(b.buildChannelHeader()),
		},
		Data: internal.MarshalOrPanic(b.buildTransaction()),
	}
}

func (b *MockCCBlockEventBuilder) buildChannelHeader() *common.ChannelHeader {
	return &common.ChannelHeader{
		Type:      int32(common.HeaderType_ENDORSER_TRANSACTION),
		TxId:      b.TxID,
		ChannelId: b.ChannelID,
	}
}

func (b *MockCCBlockEventBuilder) buildTransaction() *pb.Transaction {
	return &pb.Transaction{
		Actions: []*pb.TransactionAction{b.buildTransactionAction()},
	}
}

func (b *MockCCBlockEventBuilder) buildTransactionAction() *pb.TransactionAction {
	return &pb.TransactionAction{
		Header:  []byte{},
		Payload: internal.MarshalOrPanic(b.buildChaincodeActionPayload()),
	}
}

func (b *MockCCBlockEventBuilder) buildChaincodeActionPayload() *pb.ChaincodeActionPayload {
	return &pb.ChaincodeActionPayload{
		Action: b.buildChaincodeEndorsedAction(),
		ChaincodeProposalPayload: []byte{},
	}
}

func (b *MockCCBlockEventBuilder) buildChaincodeEndorsedAction() *pb.ChaincodeEndorsedAction {
	return &pb.ChaincodeEndorsedAction{
		ProposalResponsePayload: internal.MarshalOrPanic(b.buildProposalResponsePayload()),
		Endorsements:            []*pb.Endorsement{},
	}
}

func (b *MockCCBlockEventBuilder) buildProposalResponsePayload() *pb.ProposalResponsePayload {
	return &pb.ProposalResponsePayload{
		ProposalHash: []byte("somehash"),
		Extension:    internal.MarshalOrPanic(b.buildChaincodeAction()),
	}
}

func (b *MockCCBlockEventBuilder) buildChaincodeAction() *pb.ChaincodeAction {
	return &pb.ChaincodeAction{
		Events: internal.MarshalOrPanic(b.buildChaincodeEvent()),
	}
}

func (b *MockCCBlockEventBuilder) buildChaincodeEvent() *pb.ChaincodeEvent {
	return &pb.ChaincodeEvent{
		ChaincodeId: b.CCID,
		EventName:   b.EventName,
		TxId:        b.TxID,
		Payload:     b.Payload,
	}
}

func generateTxID() apitxn.TransactionID {
	nonce, err := internal.GenerateRandomNonce()
	if err != nil {
		panic(errors.WithMessage(err, "GenerateRandomNonce failed"))
	}
	digest, err := cryptosuite.GetDefault().Hash(
		nonce,
		cryptosuite.GetSHA256Opts())
	if err != nil {
		panic(errors.Wrap(err, "hashing nonce failed"))
	}

	txnid := apitxn.TransactionID{
		ID:    hex.EncodeToString(digest),
		Nonce: nonce,
	}

	return txnid
}
