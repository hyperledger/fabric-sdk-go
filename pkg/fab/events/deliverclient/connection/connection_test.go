/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package connection

import (
	"fmt"
	"net"
	"os"
	"testing"
	"time"

	"google.golang.org/grpc/keepalive"

	"github.com/hyperledger/fabric-sdk-go/pkg/fab/comm"
	clientdisp "github.com/hyperledger/fabric-sdk-go/pkg/fab/events/client/dispatcher"
	"github.com/hyperledger/fabric-sdk-go/pkg/fab/events/deliverclient/seek"
	eventmocks "github.com/hyperledger/fabric-sdk-go/pkg/fab/events/mocks"
	fabmocks "github.com/hyperledger/fabric-sdk-go/pkg/fab/mocks"
	cb "github.com/hyperledger/fabric-sdk-go/third_party/github.com/hyperledger/fabric/protos/common"
	pb "github.com/hyperledger/fabric-sdk-go/third_party/github.com/hyperledger/fabric/protos/peer"
	"github.com/pkg/errors"
	"google.golang.org/grpc"
)

type streamType string

const (
	peerAddress = "localhost:9999"
	peerURL     = "grpc://" + peerAddress

	streamTypeDeliver         streamType = "DELIVER"
	streamTypeDeliverFiltered streamType = "DELIVER_FILTERED"
)

var (
	peer        = fabmocks.NewMockPeer("peer1", peerURL)
	invalidPeer = fabmocks.NewMockPeer("peer2", "grpcs://invalidhost:7051")
)

func TestInvalidConnectionOpts(t *testing.T) {
	if _, err := New(newMockContext(), fabmocks.NewMockChannelCfg("mychannel"), Deliver, "grpcs://invalidhost:7051"); err == nil {
		t.Fatalf("expecting error creating new connection with invaid address but got none")
	}
}

func TestConnection(t *testing.T) {
	channelID := "mychannel"
	conn, err := New(newMockContext(), fabmocks.NewMockChannelCfg(channelID), Deliver, peerURL,
		comm.WithConnectTimeout(3*time.Second),
		comm.WithFailFast(true),
		comm.WithKeepAliveParams(
			keepalive.ClientParameters{
				Time:                10 * time.Second,
				Timeout:             10 * time.Second,
				PermitWithoutStream: true,
			},
		),
	)
	if err != nil {
		t.Fatalf("error creating new connection: %s", err)
	}

	conn.Close()

	// Calling close again should be ignored
	conn.Close()
}

func TestForbiddenConnection(t *testing.T) {
	expectedStatus := cb.Status_FORBIDDEN
	deliverServer.SetStatus(expectedStatus)
	defer deliverServer.SetStatus(cb.Status_UNKNOWN)

	channelID := "mychannel"
	conn, err := New(newMockContext(), fabmocks.NewMockChannelCfg(channelID), Deliver, peerURL,
		comm.WithConnectTimeout(3*time.Second),
		comm.WithFailFast(true),
		comm.WithKeepAliveParams(
			keepalive.ClientParameters{
				Time:                10 * time.Second,
				Timeout:             10 * time.Second,
				PermitWithoutStream: true,
			},
		),
	)
	if err != nil {
		t.Fatalf("error creating new connection: %s", err)
	}

	eventch := make(chan interface{})

	go conn.Receive(eventch)

	select {
	case e, ok := <-eventch:
		if !ok {
			t.Fatalf("unexpected closed connection")
		}
		statusResponse := e.(*pb.DeliverResponse).Type.(*pb.DeliverResponse_Status)
		if statusResponse.Status != expectedStatus {
			t.Fatalf("expecting status %s but got %s", expectedStatus, statusResponse.Status)
		}
	case <-time.After(5 * time.Second):
		t.Fatalf("timed out waiting for event")
	}

	conn.Close()
}

func TestSend(t *testing.T) {
	t.Run("SendBlockEvent", func(t *testing.T) {
		testSend(t, streamTypeDeliver)
	})
	t.Run("SendFilteredBlockEvent", func(t *testing.T) {
		testSend(t, streamTypeDeliverFiltered)
	})
}

func TestDisconnected(t *testing.T) {
	channelID := "mychannel"
	conn, err := New(newMockContext(), fabmocks.NewMockChannelCfg(channelID), Deliver, peerURL)
	if err != nil {
		t.Fatalf("error creating new connection: %s", err)
	}

	eventch := make(chan interface{})

	go conn.Receive(eventch)

	deliverServer.Disconnect(errors.New("simulating disconnect"))

	if err := conn.Send(seek.InfoNewest()); err != nil {
		t.Fatalf("error sending seek request for channel [%s]: err", err)
	}

	select {
	case e, ok := <-eventch:
		if !ok {
			t.Fatalf("unexpected closed connection")
		}
		_, ok = e.(*clientdisp.DisconnectedEvent)
		if !ok {
			t.Fatalf("expected DisconnectedEvent but got %T", e)
		}
	case <-time.After(5 * time.Second):
		t.Fatalf("timed out waiting for event")
	}

	conn.Close()
}

func getStreamProvider(streamType streamType) StreamProvider {
	if streamType == streamTypeDeliverFiltered {
		return DeliverFiltered
	}
	return Deliver
}

func testSend(t *testing.T, streamType streamType) {
	channelID := "mychannel"
	conn, err := New(newMockContext(), fabmocks.NewMockChannelCfg(channelID), getStreamProvider(streamType), peerURL)
	if err != nil {
		t.Fatalf("error creating new connection: %s", err)
	}

	eventch := make(chan interface{})

	go conn.Receive(eventch)

	if err := conn.Send(seek.InfoNewest()); err != nil {
		t.Fatalf("error sending seek request for channel [%s]: err", err)
	}

	select {
	case e, ok := <-eventch:
		if !ok {
			t.Fatalf("unexpected closed connection")
		}
		deliverResponse, ok := e.(*pb.DeliverResponse)
		if !ok {
			t.Fatalf("expected deliver response but got %T", e)
		}

		if streamType == streamTypeDeliver && deliverResponse.GetBlock() == nil {
			t.Fatalf("expected deliver response block but got none")
		}
		if streamType == streamTypeDeliverFiltered && deliverResponse.GetFilteredBlock() == nil {
			t.Fatalf("expected deliver response filtered block but got none")
		}
	case <-time.After(5 * time.Second):
		t.Fatalf("timed out waiting for event")
	}

	conn.Close()
}

var deliverServer *eventmocks.MockDeliverServer

func TestMain(m *testing.M) {
	var opts []grpc.ServerOption
	grpcServer := grpc.NewServer(opts...)

	lis, err := net.Listen("tcp", peerAddress)
	if err != nil {
		panic(fmt.Sprintf("Error starting events listener %s", err))
	}

	deliverServer = eventmocks.NewMockDeliverServer()

	pb.RegisterDeliverServer(grpcServer, deliverServer)

	go grpcServer.Serve(lis)

	time.Sleep(2 * time.Second)
	os.Exit(m.Run())
}

func newMockContext() *fabmocks.MockContext {
	return fabmocks.NewMockContext(fabmocks.NewMockUser("user1"))
}
