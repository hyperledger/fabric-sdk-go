/*
Copyright IBM Corp, SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package consumer

import (
	"context"
	"io"
	"sync"
	"time"

	"github.com/golang/protobuf/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"

	apiconfig "github.com/hyperledger/fabric-sdk-go/api/apiconfig"
	fab "github.com/hyperledger/fabric-sdk-go/api/apifabclient"
	"github.com/hyperledger/fabric-sdk-go/pkg/config/comm"
	"github.com/hyperledger/fabric-sdk-go/pkg/config/urlutil"
	"github.com/hyperledger/fabric-sdk-go/third_party/github.com/hyperledger/fabric/protos/peer"
	ehpb "github.com/hyperledger/fabric-sdk-go/third_party/github.com/hyperledger/fabric/protos/peer"

	consumer "github.com/hyperledger/fabric-sdk-go/internal/github.com/hyperledger/fabric/events/consumer"
	"github.com/hyperledger/fabric-sdk-go/pkg/errors"
	"github.com/hyperledger/fabric-sdk-go/pkg/logging"
)

var logger = logging.NewLogger("fabric_sdk_go")

const defaultTimeout = time.Second * 3

type eventsClient struct {
	sync.RWMutex
	peerAddress            string
	regTimeout             time.Duration
	stream                 ehpb.Events_ChatClient
	adapter                consumer.EventAdapter
	TLSCertificate         string
	TLSServerHostOverride  string
	clientConn             *grpc.ClientConn
	client                 fab.FabricClient
	processEventsCompleted chan struct{}
}

//NewEventsClient Returns a new grpc.ClientConn to the configured local PEER.
func NewEventsClient(client fab.FabricClient, peerAddress string, certificate string, serverhostoverride string, regTimeout time.Duration, adapter consumer.EventAdapter) (fab.EventsClient, error) {
	var err error
	if regTimeout < 100*time.Millisecond {
		regTimeout = 100 * time.Millisecond
		err = errors.New("regTimeout >= 0, setting to 100 msec")
	} else if regTimeout > 60*time.Second {
		regTimeout = 60 * time.Second
		err = errors.New("regTimeout > 60, setting to 60 sec")
	}
	return &eventsClient{sync.RWMutex{}, peerAddress, regTimeout, nil, adapter,
		certificate, serverhostoverride, nil, client, nil}, err
}

//newEventsClientConnectionWithAddress Returns a new grpc.ClientConn to the configured local PEER.
func newEventsClientConnectionWithAddress(peerAddress string, certificate string, serverHostOverride string, config apiconfig.Config) (*grpc.ClientConn, error) {
	var opts []grpc.DialOption
	opts = append(opts, grpc.WithTimeout(config.TimeoutOrDefault(apiconfig.EventHub)))
	if urlutil.IsTLSEnabled(peerAddress) {
		tlsConfig, err := comm.TLSConfig(certificate, serverHostOverride, config)
		if err != nil {
			return nil, err
		}

		opts = append(opts, grpc.WithTransportCredentials(credentials.NewTLS(tlsConfig)))
	} else {
		opts = append(opts, grpc.WithInsecure())
	}
	conn, err := grpc.Dial(urlutil.ToAddress(peerAddress), opts...)
	if err != nil {
		return nil, err
	}

	return conn, err
}

func (ec *eventsClient) send(emsg *ehpb.Event) error {
	ec.Lock()
	defer ec.Unlock()

	user, err := ec.client.LoadUserFromStateStore("")
	if err != nil {
		return errors.WithMessage(err, "LoadUserFromStateStore failed")
	}
	payload, err := proto.Marshal(emsg)
	if err != nil {
		return errors.Wrap(err, "marshal event failed")
	}

	signingMgr := ec.client.SigningManager()
	if signingMgr == nil {
		return errors.New("signing manager is nil")
	}

	signature, err := signingMgr.Sign(payload, user.PrivateKey())
	if err != nil {
		return errors.WithMessage(err, "sign failed")
	}
	signedEvt := &peer.SignedEvent{EventBytes: payload, Signature: signature}

	return ec.stream.Send(signedEvt)
}

// RegisterAsync - registers interest in a event and doesn't wait for a response
func (ec *eventsClient) RegisterAsync(ies []*ehpb.Interest) error {
	if ec.client.UserContext() == nil {
		return errors.New("user context is nil")
	}
	creator, err := ec.client.UserContext().Identity()
	if err != nil {
		return errors.WithMessage(err, "user context identity retrieval failed")
	}
	emsg := &ehpb.Event{
		Event:   &ehpb.Event_Register{Register: &ehpb.Register{Events: ies}},
		Creator: creator,
	}
	if err = ec.send(emsg); err != nil {
		logger.Errorf("error on Register send %s\n", err)
	}
	return err
}

// register - registers interest in a event
func (ec *eventsClient) register(ies []*ehpb.Interest) error {
	var err error
	if err = ec.RegisterAsync(ies); err != nil {
		return err
	}

	regChan := make(chan struct{})
	go func() {
		defer close(regChan)
		in, inerr := ec.stream.Recv()
		if inerr != nil {
			err = inerr
			return
		}
		switch in.Event.(type) {
		case *ehpb.Event_Register:
		case nil:
			err = errors.New("nil object for register")
		default:
			err = errors.New("invalid object for register")
		}
	}()
	select {
	case <-regChan:
	case <-time.After(ec.regTimeout):
		err = errors.New("register timeout")
	}
	return err
}

// UnregisterAsync - Unregisters interest in a event and doesn't wait for a response
func (ec *eventsClient) UnregisterAsync(ies []*ehpb.Interest) error {
	if ec.client.UserContext() == nil {
		return errors.New("user context is required")
	}
	creator, err := ec.client.UserContext().Identity()
	if err != nil {
		return errors.WithMessage(err, "user context identity retrieval failed")
	}

	emsg := &ehpb.Event{
		Event:   &ehpb.Event_Unregister{Unregister: &ehpb.Unregister{Events: ies}},
		Creator: creator,
	}

	if err = ec.send(emsg); err != nil {
		err = errors.Wrap(err, "unregister send failed")
	}

	return err
}

// unregister - unregisters interest in a event
func (ec *eventsClient) Unregister(ies []*ehpb.Interest) error {
	var err error
	if err = ec.UnregisterAsync(ies); err != nil {
		return err
	}

	regChan := make(chan struct{})
	go func() {
		defer close(regChan)
		in, inerr := ec.stream.Recv()
		if inerr != nil {
			err = inerr
			return
		}
		switch in.Event.(type) {
		case *ehpb.Event_Unregister:
		case nil:
			err = errors.New("nil object for unregister")
		default:
			err = errors.New("invalid object for unregister")
		}
	}()
	select {
	case <-regChan:
	case <-time.After(ec.regTimeout):
		err = errors.New("unregister timeout")
	}
	return err
}

// Recv receives next event - use when client has not called Start
func (ec *eventsClient) Recv() (*ehpb.Event, error) {
	in, err := ec.stream.Recv()
	if err == io.EOF {
		// read done
		if ec.adapter != nil {
			ec.adapter.Disconnected(nil)
		}
		return nil, err
	}
	if err != nil {
		if ec.adapter != nil {
			ec.adapter.Disconnected(err)
		}
		return nil, err
	}
	return in, nil
}
func (ec *eventsClient) processEvents() error {
	defer ec.stream.CloseSend()
	defer close(ec.processEventsCompleted)

	for {
		in, err := ec.stream.Recv()
		if err == io.EOF {
			// read done.
			if ec.adapter != nil {
				ec.adapter.Disconnected(nil)
			}
			return nil
		}
		if err != nil {
			if ec.adapter != nil {
				ec.adapter.Disconnected(err)
			}
			return err
		}
		if ec.adapter != nil {
			cont, err := ec.adapter.Recv(in)
			if !cont {
				return err
			}
		}
	}
}

//Start establishes connection with Event hub and registers interested events with it
func (ec *eventsClient) Start() error {
	conn, err := newEventsClientConnectionWithAddress(ec.peerAddress, ec.TLSCertificate, ec.TLSServerHostOverride, ec.client.Config())
	if err != nil {
		return errors.WithMessage(err, "events connection failed")
	}
	ec.clientConn = conn

	ies, err := ec.adapter.GetInterestedEvents()
	if err != nil {
		return errors.Wrap(err, "interested events retrieval failed")
	}

	if len(ies) == 0 {
		return errors.New("interested events is required")
	}

	serverClient := ehpb.NewEventsClient(conn)
	ec.stream, err = serverClient.Chat(context.Background())
	if err != nil {
		return errors.Wrap(err, "events connection failed")
	}

	if err = ec.register(ies); err != nil {
		return err
	}

	ec.processEventsCompleted = make(chan struct{})
	go ec.processEvents()

	return nil
}

//Stop terminates connection with event hub
func (ec *eventsClient) Stop() error {
	var timeoutErr error

	if ec.stream == nil {
		// in case the stream/chat server has not been established earlier, we assume that it's closed, successfully
		return nil
	}
	//this closes only sending direction of the stream; event is still there
	//read will not return an error
	err := ec.stream.CloseSend()
	if err != nil {
		return err
	}

	select {
	// Server ended its send stream in response to CloseSend()
	case <-ec.processEventsCompleted:
		// Timeout waiting for server to end stream
	case <-time.After(ec.client.Config().TimeoutOrDefault(apiconfig.EventHub)):
		timeoutErr = errors.New("close event stream timeout")
	}

	//close  client connection
	if ec.clientConn != nil {
		err := ec.clientConn.Close()
		if err != nil {
			return err
		}
	}

	if timeoutErr != nil {
		return timeoutErr
	}

	return nil
}
