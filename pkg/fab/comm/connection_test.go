/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package comm

import (
	"testing"
	"time"

	eventmocks "github.com/hyperledger/fabric-sdk-go/pkg/fab/events/mocks"
	fabmocks "github.com/hyperledger/fabric-sdk-go/pkg/fab/mocks"
	mspmocks "github.com/hyperledger/fabric-sdk-go/pkg/msp/test/mockmsp"
)

func TestConnection(t *testing.T) {
	context := newMockContext()

	_, err := NewConnection(context, "")
	if err == nil {
		t.Fatal("expected error creating new connection with empty URL")
	}
	conn, err := NewConnection(context, peerURL)
	if err != nil {
		t.Fatalf("error creating new connection: %s", err)
	}
	if conn.Closed() {
		t.Fatal("expected connection to be open")
	}
	if _, err := context.Serialize(); err != nil {
		t.Fatal("error getting identity")
	}

	time.Sleep(1 * time.Second)

	conn.Close()
	if !conn.Closed() {
		t.Fatal("expected connection to be closed")
	}

	// Calling close again should be ignored
	conn.Close()
}

// Use the mock deliver server for testing
var testServer *eventmocks.MockDeliverServer
var endorserAddr []string

func newMockContext() *fabmocks.MockContext {
	context := fabmocks.NewMockContext(mspmocks.NewMockSigningIdentity("test", "test"))
	context.SetCustomInfraProvider(NewMockInfraProvider())
	return context
}
