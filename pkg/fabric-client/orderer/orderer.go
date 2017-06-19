/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package orderer

import (
	"crypto/x509"
	"fmt"
	"time"

	api "github.com/hyperledger/fabric-sdk-go/api"
	"google.golang.org/grpc/credentials"

	"github.com/hyperledger/fabric/protos/common"
	ab "github.com/hyperledger/fabric/protos/orderer"
	"github.com/op/go-logging"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
)

var logger = logging.MustGetLogger("fabric_sdk_go")

type orderer struct {
	url            string
	grpcDialOption []grpc.DialOption
}

// CreateNewOrdererWithRootCAs Returns a new Orderer instance using the passed in orderer root CAs
func CreateNewOrdererWithRootCAs(url string, ordererRootCAs [][]byte, serverHostOverride string, config api.Config) (api.Orderer, error) {
	if config.IsTLSEnabled() {
		tlsCaCertPool, err := config.GetTLSCACertPoolFromRoots(ordererRootCAs)
		if err != nil {
			return nil, err
		}
		return createNewOrdererWithCertPool(url, tlsCaCertPool, serverHostOverride), nil
	}
	return createNewOrdererWithoutTLS(url), nil
}

func createNewOrdererWithoutTLS(url string) api.Orderer {
	var opts []grpc.DialOption
	opts = append(opts, grpc.WithTimeout(time.Second*3))
	opts = append(opts, grpc.WithInsecure())
	return &orderer{url: url, grpcDialOption: opts}
}

func createNewOrdererWithCertPool(url string, tlsCaCertPool *x509.CertPool, serverHostOverride string) api.Orderer {
	var opts []grpc.DialOption
	opts = append(opts, grpc.WithTimeout(time.Second*3))
	creds := credentials.NewClientTLSFromCert(tlsCaCertPool, serverHostOverride)
	opts = append(opts, grpc.WithTransportCredentials(creds))
	return &orderer{url: url, grpcDialOption: opts}
}

// GetURL Get the Orderer url. Required property for the instance objects.
// @returns {string} The address of the Orderer
func (o *orderer) GetURL() string {
	return o.url
}

// SendBroadcast Send the created transaction to Orderer.
func (o *orderer) SendBroadcast(envelope *api.SignedEnvelope) (*common.Status, error) {
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
// @param {*SignedEnvelope} envelope that contains the seek request for blocks
// @return {chan *common.Block} channel with the requested blocks
// @return {chan error} a buffered channel that can contain a single error
func (o *orderer) SendDeliver(envelope *api.SignedEnvelope) (chan *common.Block,
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

// NewOrderer Returns a Orderer instance
func NewOrderer(url string, certificate string, serverHostOverride string, config api.Config) (api.Orderer, error) {
	var opts []grpc.DialOption
	opts = append(opts, grpc.WithTimeout(time.Second*3))
	if config.IsTLSEnabled() {
		tlsCaCertPool, err := config.GetTLSCACertPool(certificate)
		if err != nil {
			return nil, err
		}
		creds := credentials.NewClientTLSFromCert(tlsCaCertPool, serverHostOverride)
		opts = append(opts, grpc.WithTransportCredentials(creds))
	} else {
		opts = append(opts, grpc.WithInsecure())
	}
	return &orderer{url: url, grpcDialOption: opts}, nil
}
