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

	fabmocks "github.com/hyperledger/fabric-sdk-go/pkg/fab/mocks"

	pb "github.com/hyperledger/fabric-protos-go/peer"
	"github.com/pkg/errors"
	"google.golang.org/grpc"
)

var testStream = func(grpcconn *grpc.ClientConn) (grpc.ClientStream, func(), error) {
	ctx, cancel := context.WithCancel(context.Background())
	stream, err := pb.NewDeliverClient(grpcconn).Deliver(ctx)
	return stream, cancel, err
}

var invalidStream = func(grpcconn *grpc.ClientConn) (grpc.ClientStream, func(), error) {
	return nil, nil, errors.New("simulated error creating stream")
}

func TestStreamConnection(t *testing.T) {
	channelID := "testchannel"

	context := newMockContext()
	chConfig := fabmocks.NewMockChannelCfg(channelID)

	_, err := NewStreamConnection(context, chConfig, testStream, "")
	if err == nil {
		t.Fatal("expected error creating new connection with empty URL")
	}
	_, err = NewStreamConnection(context, chConfig, testStream, "invalidhost:0000",
		WithFailFast(true),
		WithCertificate(nil),
		WithInsecure(),
		WithHostOverride(""),
		WithKeepAliveParams(keepalive.ClientParameters{}),
		WithConnectTimeout(3*time.Second),
	)
	if err == nil {
		t.Fatal("expected error creating new connection with invalid URL")
	}
	_, err = NewStreamConnection(context, chConfig, invalidStream, peerURL)
	if err == nil {
		t.Fatal("expected error creating new connection with invalid stream but got none")
	}

	conn, err := NewStreamConnection(context, chConfig, testStream, peerURL)
	if err != nil {
		t.Fatalf("error creating new connection: %s", err)
	}
	if conn.Closed() {
		t.Fatal("expected connection to be open")
	}
	if conn.Stream() == nil {
		t.Fatal("got invalid stream")
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
