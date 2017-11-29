/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package mocks

import (
	"fmt"
	"net"

	"github.com/golang/protobuf/proto"
	pb "github.com/hyperledger/fabric-sdk-go/third_party/github.com/hyperledger/fabric/protos/peer"
	"google.golang.org/grpc"
)

// MockEventServer ...
type MockEventServer struct {
	server     pb.Events_ChatServer
	grpcServer *grpc.Server
	channel    chan *pb.Event
}

// StartMockEventServer will start mock event server for unit testing purpose
func StartMockEventServer(testAddress string) (*MockEventServer, error) {
	grpcServer := grpc.NewServer()
	grpcServer.GetServiceInfo()
	lis, err := net.Listen("tcp", testAddress)
	if err != nil {
		return nil, fmt.Errorf("Error starting test server %s", err)
	}
	eventServer := &MockEventServer{grpcServer: grpcServer}
	pb.RegisterEventsServer(grpcServer, eventServer)
	fmt.Printf("Starting mock event server\n")
	go grpcServer.Serve(lis)

	return eventServer, nil
}

// Chat for event chatting
func (m *MockEventServer) Chat(srv pb.Events_ChatServer) error {
	m.server = srv
	m.channel = make(chan *pb.Event)
	in, _ := srv.Recv()
	evt := &pb.Event{}
	err := proto.Unmarshal(in.EventBytes, evt)
	if err != nil {
		return fmt.Errorf("error unmarshaling the event bytes in the SignedEvent: %s", err)
	}
	switch evt.Event.(type) {
	case *pb.Event_Register:
		srv.Send(&pb.Event{Event: &pb.Event_Register{Register: &pb.Register{}}})
	}
	for {
		event := <-m.channel
		srv.Send(event)
	}
}

// SendMockEvent used for sending mock events to event server
func (m *MockEventServer) SendMockEvent(event *pb.Event) {
	m.channel <- event
}

// Stop mock event
func (m *MockEventServer) Stop() {
	m.grpcServer.Stop()
}
