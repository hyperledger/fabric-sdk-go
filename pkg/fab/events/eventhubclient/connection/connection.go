/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package connection

import (
	"context"
	"fmt"
	"io"
	"time"

	fabcontext "github.com/hyperledger/fabric-sdk-go/pkg/common/providers/context"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/fab"

	"google.golang.org/grpc"

	"github.com/golang/protobuf/proto"
	"github.com/golang/protobuf/ptypes"
	logging "github.com/hyperledger/fabric-sdk-go/pkg/common/logging"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/options"
	comm "github.com/hyperledger/fabric-sdk-go/pkg/fab/comm"
	clientdisp "github.com/hyperledger/fabric-sdk-go/pkg/fab/events/client/dispatcher"
	pb "github.com/hyperledger/fabric-sdk-go/third_party/github.com/hyperledger/fabric/protos/peer"
	"github.com/pkg/errors"
)

var logger = logging.NewLogger("fabsdk/fab")

// EventHubConnection manages the connection and client stream
// to the event hub server
type EventHubConnection struct {
	*comm.StreamConnection
	url string
}

// New returns a new Connection to the event hub.
func New(ctx fabcontext.Client, chConfig fab.ChannelCfg, url string, opts ...options.Opt) (*EventHubConnection, error) {
	connect, err := comm.NewStreamConnection(
		ctx, chConfig,
		func(grpcconn *grpc.ClientConn) (grpc.ClientStream, error) {
			return pb.NewEventsClient(grpcconn).Chat(context.Background())
		},
		url, opts...,
	)
	if err != nil {
		return nil, err
	}

	return &EventHubConnection{
		StreamConnection: connect,
		url:              url,
	}, nil
}

// EventHubStream returns the event hub chat client
func (c *EventHubConnection) EventHubStream() pb.Events_ChatClient {
	if c.Stream() == nil {
		return nil
	}
	stream, ok := c.Stream().(pb.Events_ChatClient)
	if !ok {
		panic(fmt.Sprintf("invalid events chat client type %T", c.Stream()))
	}
	return stream
}

// Send sends an event to the event hub server
func (c *EventHubConnection) Send(emsg *pb.Event) error {
	creator, err := c.Context().Serialize()
	if err != nil {
		return errors.WithMessage(err, "error getting creator identity")
	}

	timestamp, err := ptypes.TimestampProto(time.Now())
	if err != nil {
		return errors.Wrap(err, "failed to create timestamp")
	}

	event := *emsg
	event.Creator = creator
	event.Timestamp = timestamp
	event.TlsCertHash = c.TLSCertHash()

	evtBytes, err := proto.Marshal(&event)
	if err != nil {
		return err
	}

	signature, err := c.Context().SigningManager().Sign(evtBytes, c.Context().PrivateKey())
	if err != nil {
		return err
	}

	return c.EventHubStream().Send(&pb.SignedEvent{
		EventBytes: evtBytes,
		Signature:  signature,
	})
}

// Receive receives events from the event hub server
func (c *EventHubConnection) Receive(eventch chan<- interface{}) {
	for {
		logger.Debug("Listening for events...")
		if c.EventHubStream() == nil {
			logger.Warn("The stream has closed. Terminating loop.")
			break
		}

		in, err := c.EventHubStream().Recv()

		if c.Closed() {
			logger.Debug("The connection has closed. Terminating loop.")
			break
		}

		if err == io.EOF {
			// This signifies that the stream has been terminated at the client-side. No need to send an event.
			logger.Debug("Received EOF from stream.")
			break
		}

		if err != nil {
			logger.Errorf("Received error from stream: [%s]. Sending disconnected event.", err)
			eventch <- clientdisp.NewDisconnectedEvent(err)
			break
		}
		logger.Debugf("Got event %#v", in)
		eventch <- NewEvent(in, c.url)
	}
	logger.Debug("Exiting stream listener")
}

// Event contains the event hub event as well as the event source
type Event struct {
	SourceURL string
	Event     interface{}
}

// NewEvent returns a new event hub event
func NewEvent(event interface{}, sourceURL string) *Event {
	return &Event{
		SourceURL: sourceURL,
		Event:     event,
	}
}
