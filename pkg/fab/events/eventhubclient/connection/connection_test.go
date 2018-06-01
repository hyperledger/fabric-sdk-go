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

	"github.com/hyperledger/fabric-sdk-go/pkg/fab/comm"
	clientdisp "github.com/hyperledger/fabric-sdk-go/pkg/fab/events/client/dispatcher"
	eventmocks "github.com/hyperledger/fabric-sdk-go/pkg/fab/events/mocks"
	fabmocks "github.com/hyperledger/fabric-sdk-go/pkg/fab/mocks"
	mspmocks "github.com/hyperledger/fabric-sdk-go/pkg/msp/test/mockmsp"
	pb "github.com/hyperledger/fabric-sdk-go/third_party/github.com/hyperledger/fabric/protos/peer"
	"github.com/pkg/errors"
	"google.golang.org/grpc"
)

const (
	eventAddressListen = "localhost:0"
)

var eventAddress string
var eventURL string

func TestInvalidConnectionOpts(t *testing.T) {
	if _, err := New(newMockContext(), fabmocks.NewMockChannelCfg("channelid"), "grpcs://invalidhost:7053"); err == nil {
		t.Fatal("expecting error creating new connection with invaid address but got none")
	}
}

func TestConnection(t *testing.T) {
	channelID := "mychannel"
	conn, err := New(newMockContext(), fabmocks.NewMockChannelCfg(channelID), eventURL)
	if err != nil {
		t.Fatalf("error creating new connection: %s", err)
	}

	conn.Close()

	// Calling close again should be ignored
	conn.Close()
}

func TestSend(t *testing.T) {
	channelID := "mychannel"
	conn, err := New(newMockContext(), fabmocks.NewMockChannelCfg(channelID), eventURL)
	if err != nil {
		t.Fatalf("error creating new connection: %s", err)
	}

	eventch := make(chan interface{})

	go conn.Receive(eventch)

	emsg := &pb.Event{
		Event: &pb.Event_Register{
			Register: &pb.Register{
				Events: []*pb.Interest{
					{EventType: pb.EventType_FILTEREDBLOCK},
				},
			},
		},
	}

	t.Log("Sending register event...")
	if err := conn.Send(emsg); err != nil {
		t.Fatalf("Error sending register interest event: %s", err)
	}

	select {
	case e, ok := <-eventch:
		if !ok {
			t.Fatal("unexpected closed connection")
		}
		t.Logf("Got response: %#v", e)
		eventHubEvent, ok := e.(*Event)
		if !ok {
			t.Fatalf("expected EventHubEvent but got %T", e)
		}
		evt, ok := eventHubEvent.Event.(*pb.Event)
		if !ok {
			t.Fatalf("expected Event but got %T", eventHubEvent.Event)
		}
		if !ok {
			t.Fatalf("expected register response but got %T", evt.Event)
		}
	case <-time.After(5 * time.Second):
		t.Fatal("timed out waiting for event")
	}

	emsg = &pb.Event{
		Event: &pb.Event_Unregister{
			Unregister: &pb.Unregister{
				Events: []*pb.Interest{
					{EventType: pb.EventType_FILTEREDBLOCK},
				},
			},
		},
	}

	t.Log("Sending unregister event...")
	if err := conn.Send(emsg); err != nil {
		t.Fatalf("Error sending unregister interest event: %s", err)
	}

	checkEvent(eventch, t)

	conn.Close()
}

func checkEvent(eventch chan interface{}, t *testing.T) {
	select {
	case e, ok := <-eventch:
		if !ok {
			t.Fatal("unexpected closed connection")
		}
		t.Logf("Got response: %#v", e)
		eventHubEvent, ok := e.(*Event)
		if !ok {
			t.Fatalf("expected EventHubEvent but got %T", e)
		}
		evt, ok := eventHubEvent.Event.(*pb.Event)
		if !ok {
			t.Fatalf("expected Event but got %T", eventHubEvent.Event)
		}
		_, ok = evt.Event.(*pb.Event_Unregister)
		if !ok {
			t.Fatalf("expected unregister response but got %T", evt.Event)
		}
	case <-time.After(5 * time.Second):
		t.Fatal("timed out waiting for event")
	}
}

func TestDisconnected(t *testing.T) {
	channelID := "mychannel"
	conn, err := New(newMockContext(), fabmocks.NewMockChannelCfg(channelID), eventURL)
	if err != nil {
		t.Fatalf("error creating new connection: %s", err)
	}

	eventch := make(chan interface{})

	go conn.Receive(eventch)

	emsg := &pb.Event{
		Event: &pb.Event_Register{
			Register: &pb.Register{
				Events: []*pb.Interest{
					{EventType: pb.EventType_FILTEREDBLOCK},
				},
			},
		},
	}

	if err := conn.Send(emsg); err != nil {
		t.Fatalf("Error sending register interest event: %s", err)
	}

	ehServer.Disconnect(errors.New("simulating disconnect"))

	select {
	case e, ok := <-eventch:
		if !ok {
			t.Fatal("unexpected closed connection")
		}
		_, ok = e.(*clientdisp.DisconnectedEvent)
		if !ok {
			t.Fatalf("expected DisconnectedEvent but got %T", e)
		}
	case <-time.After(5 * time.Second):
		t.Fatal("timed out waiting for event")
	}

	conn.Close()
}

var ehServer *eventmocks.MockEventhubServer

func TestMain(m *testing.M) {
	var opts []grpc.ServerOption
	grpcServer := grpc.NewServer(opts...)

	lis, err := net.Listen("tcp", eventAddressListen)
	if err != nil {
		panic(fmt.Sprintf("Error starting events listener %s", err))
	}

	eventAddress = lis.Addr().String()
	eventURL = "grpc://" + eventAddress

	ehServer = eventmocks.NewMockEventhubServer()

	pb.RegisterEventsServer(grpcServer, ehServer)

	go grpcServer.Serve(lis)

	time.Sleep(2 * time.Second)
	os.Exit(m.Run())
}

func newMockContext() *fabmocks.MockContext {
	context := fabmocks.NewMockContext(mspmocks.NewMockSigningIdentity("user1", "Org1MSP"))
	context.SetCustomInfraProvider(comm.NewMockInfraProvider())
	return context
}
