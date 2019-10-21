/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package comm

import (
	"context"
	"testing"
	"time"

	eventmocks "github.com/hyperledger/fabric-sdk-go/pkg/fab/events/mocks"
	fabmocks "github.com/hyperledger/fabric-sdk-go/pkg/fab/mocks"
	mspmocks "github.com/hyperledger/fabric-sdk-go/pkg/msp/test/mockmsp"
	"github.com/stretchr/testify/require"
)

func TestConnection(t *testing.T) {
	ctx := newMockContext()

	_, err := NewConnection(ctx, "")
	if err == nil {
		t.Fatal("expected error creating new connection with empty URL")
	}

	conn, err := NewConnection(ctx, peerURL)
	if err != nil {
		t.Fatalf("error creating new connection: %s", err)
	}
	if conn.Closed() {
		t.Fatal("expected connection to be open")
	}
	if _, err := ctx.Serialize(); err != nil {
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

func TestConnection_WithParentContext(t *testing.T) {
	ctx := newMockContext()
	reqCtx, cancel := context.WithCancel(context.Background())
	cancel()
	conn, err := NewConnection(ctx, "localhost:8978", WithParentContext(reqCtx))
	require.Error(t, err)
	require.Contains(t, err.Error(), context.Canceled.Error())
	require.Nil(t, conn)
}

// Use the mock deliver server for testing
var testServer *eventmocks.MockDeliverServer
var endorserAddr []string

func newMockContext() *fabmocks.MockContext {
	context := fabmocks.NewMockContext(mspmocks.NewMockSigningIdentity("test", "test"))
	context.SetCustomInfraProvider(NewMockInfraProvider())
	return context
}
