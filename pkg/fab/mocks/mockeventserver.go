/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package mocks

import (
	"fmt"
	"net"
	"sync"

	"github.com/golang/protobuf/proto"
	"github.com/hyperledger/fabric-sdk-go/pkg/util/test"
	pb "github.com/hyperledger/fabric-sdk-go/third_party/github.com/hyperledger/fabric/protos/peer"
	"google.golang.org/grpc"
)

// MockEventServer ...
type MockEventServer struct {
	server  pb.Events_ChatServer
	channel chan *pb.Event
	srv     *grpc.Server
	wg      sync.WaitGroup
}

// Start the mock event server
func (m *MockEventServer) Start(address string) string {
	if m.srv != nil {
		panic("MockEventServer already started")
	}
	m.srv = grpc.NewServer()

	lis, err := net.Listen("tcp", address)
	if err != nil {
		panic(fmt.Sprintf("starting test server failed %s", err))
	}

	addr := lis.Addr().String()

	test.Logf("Starting MockEventServer [%s]", addr)
	pb.RegisterEventsServer(m.srv, m)
	m.wg.Add(1)
	go func() {
		defer m.wg.Done()
		if err := m.srv.Serve(lis); err != nil {
			test.Logf("StartMockEventServer failed [%s]", err)
		}
	}()

	return addr
}

// Stop the mock event server and wait for completion.
func (m *MockEventServer) Stop() {
	if m.srv == nil {
		panic("MockEventServer not started")
	}

	m.srv.Stop()
	m.wg.Wait()
	m.srv = nil
}

// Chat for event chatting
func (m *MockEventServer) Chat(srv pb.Events_ChatServer) error {
	m.server = srv
	m.channel = make(chan *pb.Event)
	in, err := srv.Recv()
	if err != nil {
		return err
	}
	evt := &pb.Event{}
	err = proto.Unmarshal(in.EventBytes, evt)
	if err != nil {
		return fmt.Errorf("error unmarshaling the event bytes in the SignedEvent: %s", err)
	}
	switch evt.Event.(type) {
	case *pb.Event_Register:
		if err := srv.Send(&pb.Event{Event: &pb.Event_Register{Register: &pb.Register{}}}); err != nil {
			return err
		}
	}
	for {
		event := <-m.channel
		if err := srv.Send(event); err != nil {
			return err
		}
	}
}

// SendMockEvent used for sending mock events to event server
func (m *MockEventServer) SendMockEvent(event *pb.Event) {
	m.channel <- event
}
