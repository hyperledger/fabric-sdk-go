/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package comm

import (
	"context"
	"testing"
	"time"

	"google.golang.org/grpc/keepalive"

	fabcontext "github.com/hyperledger/fabric-sdk-go/pkg/common/context"
	eventmocks "github.com/hyperledger/fabric-sdk-go/pkg/fab/events/mocks"
	fabmocks "github.com/hyperledger/fabric-sdk-go/pkg/fab/mocks"

	pb "github.com/hyperledger/fabric-sdk-go/third_party/github.com/hyperledger/fabric/protos/peer"
	"github.com/pkg/errors"
	"google.golang.org/grpc"
)

var testStream = func(grpcconn *grpc.ClientConn) (grpc.ClientStream, error) {
	return pb.NewDeliverClient(grpcconn).Deliver(context.Background())
}

var invalidStream = func(grpcconn *grpc.ClientConn) (grpc.ClientStream, error) {
	return nil, errors.New("simulated error creating stream")
}

func TestConnection(t *testing.T) {
	channelID := "testchannel"

	context := newMockContext()
	chConfig := fabmocks.NewMockChannelCfg(channelID)

	_, err := NewConnection(context, chConfig, testStream, "")
	if err == nil {
		t.Fatalf("expected error creating new connection with empty URL")
	}
	_, err = NewConnection(context, chConfig, testStream, "invalidhost:0000",
		WithFailFast(true),
		WithCertificate(nil),
		WithHostOverride(""),
		WithKeepAliveParams(keepalive.ClientParameters{}),
		WithConnectTimeout(3*time.Second),
	)
	if err == nil {
		t.Fatalf("expected error creating new connection with invalid URL")
	}
	_, err = NewConnection(context, chConfig, invalidStream, peerURL)
	if err == nil {
		t.Fatalf("expected error creating new connection with invalid stream but got none")
	}

	conn, err := NewConnection(context, chConfig, testStream, peerURL)
	if err != nil {
		t.Fatalf("error creating new connection: %s", err)
	}
	if conn.Closed() {
		t.Fatalf("expected connection to be open")
	}
	if conn.Stream() == nil {
		t.Fatalf("got invalid stream")
	}
	if _, err := context.SerializedIdentity(); err != nil {
		t.Fatalf("error getting identity")
	}

	time.Sleep(1 * time.Second)

	conn.Close()
	if !conn.Closed() {
		t.Fatalf("expected connection to be closed")
	}

	// Calling close again should be ignored
	conn.Close()
}

// Use the Deliver server for testing
var testServer *eventmocks.MockEventhubServer
var endorserAddr []string

func newMockContext() fabcontext.Client {
	return fabmocks.NewMockContext(fabmocks.NewMockUser("test"))
}
