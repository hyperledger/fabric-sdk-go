/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package mocks

import (
	"crypto/x509"

	"github.com/hyperledger/fabric-sdk-go/api/apifabclient"
	"github.com/hyperledger/fabric-sdk-go/api/apitxn"
	"github.com/hyperledger/fabric-sdk-go/third_party/github.com/hyperledger/fabric/protos/common"

	pb "github.com/hyperledger/fabric-sdk-go/third_party/github.com/hyperledger/fabric/protos/peer"
)

// MockEventHub Mock EventHub
type MockEventHub struct {
	RegisteredTxCallbacks chan func(string, pb.TxValidationCode, error)
}

// NewMockEventHub creates a new mock EventHub
func NewMockEventHub() *MockEventHub {
	return &MockEventHub{RegisteredTxCallbacks: make(chan func(string, pb.TxValidationCode, error))}
}

// SetPeerAddr not implemented
func (m *MockEventHub) SetPeerAddr(peerURL string, certificate *x509.Certificate, serverHostOverride string) {
	// Not implemented
}

// IsConnected not implemented
func (m *MockEventHub) IsConnected() bool {
	return false
}

// Connect not implemented
func (m *MockEventHub) Connect() error {
	return nil
}

// Disconnect not implemented
func (m *MockEventHub) Disconnect() error {
	return nil
}

// RegisterChaincodeEvent not implemented
func (m *MockEventHub) RegisterChaincodeEvent(ccid string, eventname string, callback func(*apifabclient.ChaincodeEvent)) *apifabclient.ChainCodeCBE {
	return nil
}

// UnregisterChaincodeEvent not implemented
func (m *MockEventHub) UnregisterChaincodeEvent(cbe *apifabclient.ChainCodeCBE) {
	return
}

// RegisterTxEvent not implemented
func (m *MockEventHub) RegisterTxEvent(txnID apitxn.TransactionID, callback func(string, pb.TxValidationCode, error)) {
	go func() { m.RegisteredTxCallbacks <- callback }()
	return
}

// UnregisterTxEvent not implemented
func (m *MockEventHub) UnregisterTxEvent(txnID apitxn.TransactionID) {
	return
}

// RegisterBlockEvent not implemented
func (m *MockEventHub) RegisterBlockEvent(callback func(*common.Block)) {
	return
}

// UnregisterBlockEvent not implemented
func (m *MockEventHub) UnregisterBlockEvent(callback func(*common.Block)) {
	return
}
