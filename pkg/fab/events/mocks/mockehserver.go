/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package mocks

import (
	"fmt"
	"io"
	"sync"

	"github.com/golang/protobuf/proto"
	pb "github.com/hyperledger/fabric-sdk-go/third_party/github.com/hyperledger/fabric/protos/peer"
)

// MockEventhubServer is a mock event hub server
type MockEventhubServer struct {
	sync.RWMutex
	disconnErr error
}

// NewMockEventhubServer returns a new MockEventhubServer
func NewMockEventhubServer() *MockEventhubServer {
	return new(MockEventhubServer)
}

// Disconnect terminates the stream and returns the given error to the client
func (s *MockEventhubServer) Disconnect(err error) {
	s.Lock()
	defer s.Unlock()
	s.disconnErr = err
}

func (s *MockEventhubServer) disconnectErr() error {
	s.RLock()
	defer s.RUnlock()
	return s.disconnErr
}

// Chat starts a listener on the given chat stream
func (s *MockEventhubServer) Chat(srv pb.Events_ChatServer) error {
	for {
		signedEvt, err := srv.Recv()
		if err == io.EOF || signedEvt == nil {
			break
		}

		err = s.disconnectErr()
		if err != nil {
			return err
		}

		var emsg pb.Event
		if err := proto.Unmarshal(signedEvt.EventBytes, &emsg); err != nil {
			panic(fmt.Sprintf("Error unmarshalling event bytes: %s", err))
		}

		switch emsg.Event.(type) {
		case *pb.Event_Register:
			// Send back the same event (which is what the event hub server currently does)
			send(srv, emsg)
		case *pb.Event_Unregister:
			// Send back the same event (which is what the event hub server currently does)
			send(srv, emsg)
		default:
			panic(fmt.Sprintf("Unsupported message type: %T", emsg))
		}
	}
	return nil
}

func send(srv pb.Events_ChatServer, emsg pb.Event) {
	if err := srv.Send(&emsg); err != nil {
		panic(fmt.Sprintf("Error Send event: %s", err))
	}
}
