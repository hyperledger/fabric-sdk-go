/*
Copyright IBM Corp, SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package consumer

import (
	"fmt"
	"io"
	"sync"
	"time"

	"github.com/golang/protobuf/proto"
	apiconfig "github.com/hyperledger/fabric-sdk-go/api/apiconfig"
	fab "github.com/hyperledger/fabric-sdk-go/api/apifabclient"
	fc "github.com/hyperledger/fabric-sdk-go/pkg/fabric-client/internal"
	"github.com/hyperledger/fabric/bccsp"
	consumer "github.com/hyperledger/fabric/events/consumer"
	"github.com/hyperledger/fabric/protos/peer"
	ehpb "github.com/hyperledger/fabric/protos/peer"
	logging "github.com/op/go-logging"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

var logger = logging.MustGetLogger("fabric_sdk_go")

const defaultTimeout = time.Second * 3

type eventsClient struct {
	sync.RWMutex
	peerAddress           string
	regTimeout            time.Duration
	stream                ehpb.Events_ChatClient
	adapter               consumer.EventAdapter
	TLSCertificate        string
	TLSServerHostOverride string
	clientConn            *grpc.ClientConn
	client                fab.FabricClient
}

//NewEventsClient Returns a new grpc.ClientConn to the configured local PEER.
func NewEventsClient(client fab.FabricClient, peerAddress string, certificate string, serverhostoverride string, regTimeout time.Duration, adapter consumer.EventAdapter) (fab.EventsClient, error) {
	var err error
	if regTimeout < 100*time.Millisecond {
		regTimeout = 100 * time.Millisecond
		err = fmt.Errorf("regTimeout >= 0, setting to 100 msec")
	} else if regTimeout > 60*time.Second {
		regTimeout = 60 * time.Second
		err = fmt.Errorf("regTimeout > 60, setting to 60 sec")
	}
	return &eventsClient{sync.RWMutex{}, peerAddress, regTimeout, nil, adapter, certificate, serverhostoverride, nil, client}, err
}

//newEventsClientConnectionWithAddress Returns a new grpc.ClientConn to the configured local PEER.
func newEventsClientConnectionWithAddress(peerAddress string, certificate string, serverhostoverride string, config apiconfig.Config) (*grpc.ClientConn, error) {
	var opts []grpc.DialOption
	opts = append(opts, grpc.WithTimeout(config.TimeoutOrDefault(apiconfig.EventHub)))
	if config.IsTLSEnabled() {
		tlsCaCertPool, err := config.TLSCACertPool(certificate)
		if err != nil {
			return nil, err
		}
		creds := credentials.NewClientTLSFromCert(tlsCaCertPool, serverhostoverride)
		opts = append(opts, grpc.WithTransportCredentials(creds))
	} else {
		opts = append(opts, grpc.WithInsecure())
	}
	conn, err := grpc.Dial(peerAddress, opts...)
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
		return fmt.Errorf("LoadUserFromStateStore returned error: %s", err)
	}
	payload, err := proto.Marshal(emsg)
	if err != nil {
		return fmt.Errorf("Error marshaling message: %s", err)
	}
	signature, err := fc.SignObjectWithKey(payload, user.PrivateKey(),
		&bccsp.SHAOpts{}, nil, ec.client.CryptoSuite())
	if err != nil {
		return fmt.Errorf("Error signing message: %s", err)
	}
	signedEvt := &peer.SignedEvent{EventBytes: payload, Signature: signature}

	return ec.stream.Send(signedEvt)
}

// RegisterAsync - registers interest in a event and doesn't wait for a response
func (ec *eventsClient) RegisterAsync(ies []*ehpb.Interest) error {
	if ec.client.UserContext() == nil {
		return fmt.Errorf("User context needs to be set")
	}
	creator, err := ec.client.UserContext().Identity()
	if err != nil {
		return fmt.Errorf("Error getting creator: %v", err)
	}
	emsg := &ehpb.Event{
		Event:   &ehpb.Event_Register{Register: &ehpb.Register{Events: ies}},
		Creator: creator,
	}
	if err = ec.send(emsg); err != nil {
		fmt.Printf("error on Register send %s\n", err)
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
			err = fmt.Errorf("invalid nil object for register")
		default:
			err = fmt.Errorf("invalid registration object")
		}
	}()
	select {
	case <-regChan:
	case <-time.After(ec.regTimeout):
		err = fmt.Errorf("timeout waiting for registration")
	}
	return err
}

// UnregisterAsync - Unregisters interest in a event and doesn't wait for a response
func (ec *eventsClient) UnregisterAsync(ies []*ehpb.Interest) error {
	emsg := &ehpb.Event{Event: &ehpb.Event_Unregister{Unregister: &ehpb.Unregister{Events: ies}}}
	var err error
	if err = ec.send(emsg); err != nil {
		err = fmt.Errorf("error on unregister send %s", err)
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
			err = fmt.Errorf("invalid nil object for unregister")
		default:
			err = fmt.Errorf("invalid unregistration object")
		}
	}()
	select {
	case <-regChan:
	case <-time.After(ec.regTimeout):
		err = fmt.Errorf("timeout waiting for unregistration")
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
		return fmt.Errorf("Could not create client conn to %s (%v)", ec.peerAddress, err)
	}
	ec.clientConn = conn

	ies, err := ec.adapter.GetInterestedEvents()
	if err != nil {
		return fmt.Errorf("error getting interested events:%s", err)
	}

	if len(ies) == 0 {
		return fmt.Errorf("must supply interested events")
	}

	serverClient := ehpb.NewEventsClient(conn)
	ec.stream, err = serverClient.Chat(context.Background())
	if err != nil {
		return fmt.Errorf("Could not create client conn to %s (%v)", ec.peerAddress, err)
	}

	if err = ec.register(ies); err != nil {
		return err
	}

	go ec.processEvents()

	return nil
}

//Stop terminates connection with event hub
func (ec *eventsClient) Stop() error {
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
	//close  client connection
	if ec.clientConn != nil {
		err := ec.clientConn.Close()
		if err != nil {
			return err
		}
	}
	return nil
}
