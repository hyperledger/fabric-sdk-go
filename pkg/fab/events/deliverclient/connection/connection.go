/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package connection

import (
	"context"
	"fmt"
	"io"

	"github.com/golang/protobuf/proto"
	cb "github.com/hyperledger/fabric-protos-go/common"
	ab "github.com/hyperledger/fabric-protos-go/orderer"
	pb "github.com/hyperledger/fabric-protos-go/peer"
	"github.com/hyperledger/fabric-sdk-go/internal/github.com/hyperledger/fabric/common/crypto"
	"github.com/hyperledger/fabric-sdk-go/internal/github.com/hyperledger/fabric/protoutil"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/logging"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/options"
	fabcontext "github.com/hyperledger/fabric-sdk-go/pkg/common/providers/context"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/fab"
	"github.com/hyperledger/fabric-sdk-go/pkg/fab/comm"
	clientdisp "github.com/hyperledger/fabric-sdk-go/pkg/fab/events/client/dispatcher"
	"github.com/pkg/errors"
	"google.golang.org/grpc"
)

var logger = logging.NewLogger("fabsdk/fab")

type deliverStream interface {
	grpc.ClientStream
	Send(*cb.Envelope) error
	Recv() (*pb.DeliverResponse, error)
}

// DeliverConnection manages the connection to the deliver server
type DeliverConnection struct {
	*comm.StreamConnection
	url string
}

// StreamProvider creates a deliver stream
type StreamProvider func(pb.DeliverClient) (stream deliverStream, cancel func(), err error)

var (
	// Deliver creates a Deliver stream
	Deliver = func(client pb.DeliverClient) (deliverStream, func(), error) {
		ctx, cancel := context.WithCancel(context.Background())
		stream, err := client.Deliver(ctx)
		return stream, cancel, err
	}

	// DeliverFiltered creates a DeliverFiltered stream
	DeliverFiltered = func(client pb.DeliverClient) (deliverStream, func(), error) {
		ctx, cancel := context.WithCancel(context.Background())
		stream, err := client.DeliverFiltered(ctx)
		return stream, cancel, err
	}
)

// New returns a new Deliver Server connection
func New(ctx fabcontext.Client, chConfig fab.ChannelCfg, streamProvider StreamProvider, url string, opts ...options.Opt) (*DeliverConnection, error) {
	logger.Debugf("Connecting to %s...", url)
	connect, err := comm.NewStreamConnection(
		ctx, chConfig,
		func(grpcconn *grpc.ClientConn) (grpc.ClientStream, func(), error) {
			return streamProvider(pb.NewDeliverClient(grpcconn))
		},
		url, opts...,
	)
	if err != nil {
		return nil, err
	}

	return &DeliverConnection{
		StreamConnection: connect,
		url:              url,
	}, nil
}

func (c *DeliverConnection) deliverStream() deliverStream {
	if c.Stream() == nil {
		return nil
	}
	stream, ok := c.Stream().(deliverStream)
	if !ok {
		panic(fmt.Sprintf("invalid DeliverStream type %T", c.Stream()))
	}
	return stream
}

// Send sends a seek request to the deliver server
func (c *DeliverConnection) Send(seekInfo *ab.SeekInfo) error {
	if c.Closed() {
		return errors.New("connection is closed")
	}

	logger.Debugf("Sending %#v", seekInfo)

	env, err := c.createSignedEnvelope(seekInfo)
	if err != nil {
		return err
	}

	return c.deliverStream().Send(env)
}

// Receive receives events from the deliver server
func (c *DeliverConnection) Receive(eventch chan<- interface{}) {
	for {
		stream := c.deliverStream()
		if stream == nil {
			logger.Warn("The stream has closed. Terminating loop.")
			break
		}

		in, err := stream.Recv()

		logger.Debugf("Got deliver response: %#v", in)

		if c.Closed() {
			logger.Debugf("The connection has closed with error [%s]. Terminating loop.", err)
			break
		}

		if err == io.EOF {
			// This signifies that the stream has been terminated at the client-side. No need to send an event.
			logger.Debug("Received EOF from stream.")
			break
		}

		if err != nil {
			logger.Warnf("Received error from stream: [%s]. Sending disconnected event.", err)
			eventch <- clientdisp.NewDisconnectedEvent(err)
			break
		}

		eventch <- NewEvent(in, c.url)
	}
	logger.Debug("Exiting stream listener")
}

func (c *DeliverConnection) createSignedEnvelope(msg proto.Message) (*cb.Envelope, error) {
	// TODO: Do we need to make these configurable?
	var msgVersion int32
	var epoch uint64

	payloadChannelHeader := protoutil.MakeChannelHeader(cb.HeaderType_DELIVER_SEEK_INFO, msgVersion, c.ChannelConfig().ID(), epoch)
	payloadChannelHeader.TlsCertHash = c.TLSCertHash()

	data, err := proto.Marshal(msg)
	if err != nil {
		return nil, err
	}

	identity, err := c.Context().Serialize()
	if err != nil {
		return nil, err
	}

	nonce, err := crypto.GetRandomNonce()
	if err != nil {
		return nil, err
	}

	payloadSignatureHeader := &cb.SignatureHeader{
		Creator: identity,
		Nonce:   nonce,
	}

	paylBytes := protoutil.MarshalOrPanic(&cb.Payload{
		Header: protoutil.MakePayloadHeader(payloadChannelHeader, payloadSignatureHeader),
		Data:   data,
	})

	signature, err := c.Context().SigningManager().Sign(paylBytes, c.Context().PrivateKey())
	if err != nil {
		return nil, err
	}

	return &cb.Envelope{Payload: paylBytes, Signature: signature}, nil
}

// Event contains the deliver event as well as the event source
type Event struct {
	SourceURL string
	Event     interface{}
}

// NewEvent returns a deliver event
func NewEvent(event interface{}, sourceURL string) *Event {
	return &Event{
		SourceURL: sourceURL,
		Event:     event,
	}
}
