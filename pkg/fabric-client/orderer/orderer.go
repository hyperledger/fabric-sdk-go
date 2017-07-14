/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package orderer

import (
	"fmt"

	"github.com/hyperledger/fabric-sdk-go/api/apiconfig"
	fab "github.com/hyperledger/fabric-sdk-go/api/apifabclient"
	"google.golang.org/grpc/credentials"

	"github.com/hyperledger/fabric/protos/common"
	ab "github.com/hyperledger/fabric/protos/orderer"
	"github.com/op/go-logging"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
)

var logger = logging.MustGetLogger("fabric_sdk_go")

// Orderer allows a client to broadcast a transaction.
type Orderer struct {
	url            string
	grpcDialOption []grpc.DialOption
}

// NewOrderer Returns a Orderer instance
func NewOrderer(url string, certificate string, serverHostOverride string, config apiconfig.Config) (*Orderer, error) {
	var opts []grpc.DialOption
	opts = append(opts, grpc.WithTimeout(config.TimeoutOrDefault(apiconfig.Orderer)))
	if config.IsTLSEnabled() {
		tlsCaCertPool, err := config.TLSCACertPool(certificate)
		if err != nil {
			return nil, err
		}
		creds := credentials.NewClientTLSFromCert(tlsCaCertPool, serverHostOverride)
		opts = append(opts, grpc.WithTransportCredentials(creds))
	} else {
		opts = append(opts, grpc.WithInsecure())
	}
	return &Orderer{url: url, grpcDialOption: opts}, nil
}

// URL Get the Orderer url. Required property for the instance objects.
// Returns the address of the Orderer.
func (o *Orderer) URL() string {
	return o.url
}

// SendBroadcast Send the created transaction to Orderer.
func (o *Orderer) SendBroadcast(envelope *fab.SignedEnvelope) (*common.Status, error) {
	conn, err := grpc.Dial(o.url, o.grpcDialOption...)
	if err != nil {
		return nil, err
	}
	defer conn.Close()

	broadcastStream, err := ab.NewAtomicBroadcastClient(conn).Broadcast(context.Background())
	if err != nil {
		return nil, fmt.Errorf("Error Create NewAtomicBroadcastClient %v", err)
	}
	done := make(chan bool)
	var broadcastErr error
	var broadcastStatus *common.Status

	go func() {
		for {
			broadcastResponse, err := broadcastStream.Recv()
			logger.Debugf("Orderer.broadcastStream - response:%v, error:%v\n", broadcastResponse, err)
			if err != nil {
				broadcastErr = fmt.Errorf("error broadcast response : %v", err)
				done <- true
				return
			}
			broadcastStatus = &broadcastResponse.Status
			if broadcastResponse.Status == common.Status_SUCCESS {
				done <- true
				return
			}
			if broadcastResponse.Status != common.Status_SUCCESS {
				broadcastErr = fmt.Errorf("broadcast response is not success : %v", broadcastResponse.Status)
				done <- true
				return
			}
		}
	}()
	if err := broadcastStream.Send(&common.Envelope{
		Payload:   envelope.Payload,
		Signature: envelope.Signature,
	}); err != nil {
		return nil, fmt.Errorf("Failed to send a envelope to orderer: %v", err)
	}
	broadcastStream.CloseSend()
	<-done
	return broadcastStatus, broadcastErr
}

// SendDeliver sends a deliver request to the ordering service and returns the
// blocks requested
// envelope: contains the seek request for blocks
func (o *Orderer) SendDeliver(envelope *fab.SignedEnvelope) (chan *common.Block,
	chan error) {
	responses := make(chan *common.Block)
	errors := make(chan error, 1)
	// Validate envelope
	if envelope == nil {
		errors <- fmt.Errorf("Envelope cannot be nil")
		return responses, errors
	}
	// Establish connection to Ordering Service
	conn, err := grpc.Dial(o.url, o.grpcDialOption...)
	if err != nil {
		errors <- err
		return responses, errors
	}
	// Create atomic broadcast client
	broadcastStream, err := ab.NewAtomicBroadcastClient(conn).
		Deliver(context.Background())
	if err != nil {
		errors <- fmt.Errorf("Error creating NewAtomicBroadcastClient %s", err)
		return responses, errors
	}
	// Send block request envolope
	logger.Debugf("Requesting blocks from ordering service")
	if err := broadcastStream.Send(&common.Envelope{
		Payload:   envelope.Payload,
		Signature: envelope.Signature,
	}); err != nil {
		errors <- fmt.Errorf("Failed to send block request to orderer: %s", err)
		return responses, errors
	}
	// Receive blocks from the GRPC stream and put them on the channel
	go func() {
		defer conn.Close()
		for {
			response, err := broadcastStream.Recv()
			if err != nil {
				errors <- fmt.Errorf("Got error from ordering service: %s", err)
				return
			}
			// Assert response type
			switch t := response.Type.(type) {
			// Seek operation success, no more resposes
			case *ab.DeliverResponse_Status:
				if t.Status == common.Status_SUCCESS {
					close(responses)
					return
				}
				if t.Status != common.Status_SUCCESS {
					errors <- fmt.Errorf("Got error status from ordering service: %s",
						t.Status)
					return
				}

			// Response is a requested block
			case *ab.DeliverResponse_Block:
				logger.Debug("Received block from ordering service")
				responses <- response.GetBlock()
			// Unknown response
			default:
				errors <- fmt.Errorf("Received unknown response from ordering service: %s", t)
				return
			}
		}
	}()

	return responses, errors
}
