/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package orderer

import (
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"

	"github.com/hyperledger/fabric-sdk-go/api/apiconfig"
	fab "github.com/hyperledger/fabric-sdk-go/api/apifabclient"
	ab "github.com/hyperledger/fabric-sdk-go/internal/github.com/hyperledger/fabric/protos/orderer"
	"github.com/hyperledger/fabric-sdk-go/pkg/config/comm"
	"github.com/hyperledger/fabric-sdk-go/pkg/config/urlutil"
	"github.com/hyperledger/fabric-sdk-go/third_party/github.com/hyperledger/fabric/protos/common"

	"github.com/hyperledger/fabric-sdk-go/pkg/errors"
	"github.com/hyperledger/fabric-sdk-go/pkg/logging"
)

var logger = logging.NewLogger("fabric_sdk_go")

// Orderer allows a client to broadcast a transaction.
type Orderer struct {
	url            string
	grpcDialOption []grpc.DialOption
}

// NewOrderer Returns a Orderer instance
func NewOrderer(url string, certificate string, serverHostOverride string, config apiconfig.Config) (*Orderer, error) {
	var opts []grpc.DialOption
	opts = append(opts, grpc.WithTimeout(config.TimeoutOrDefault(apiconfig.OrdererConnection)))
	if urlutil.IsTLSEnabled(url) {
		tlsConfig, err := comm.TLSConfig(certificate, serverHostOverride, config)
		if err != nil {
			return nil, err
		}

		opts = append(opts, grpc.WithTransportCredentials(credentials.NewTLS(tlsConfig)))
	} else {
		opts = append(opts, grpc.WithInsecure())
	}
	return &Orderer{url: urlutil.ToAddress(url), grpcDialOption: opts}, nil
}

// NewOrdererFromConfig returns an Orderer instance constructed from orderer config
func NewOrdererFromConfig(ordererCfg *apiconfig.OrdererConfig, config apiconfig.Config) (*Orderer, error) {

	serverHostOverride := ""
	if str, ok := ordererCfg.GRPCOptions["ssl-target-name-override"].(string); ok {
		serverHostOverride = str
	}

	return NewOrderer(ordererCfg.URL, ordererCfg.TLSCACerts.Path, serverHostOverride, config)
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
		return nil, errors.Wrap(err, "NewAtomicBroadcastClient failed")
	}
	done := make(chan bool)
	var broadcastErr error
	var broadcastStatus *common.Status

	go func() {
		for {
			broadcastResponse, err := broadcastStream.Recv()
			logger.Debugf("Orderer.broadcastStream - response:%v, error:%v\n", broadcastResponse, err)
			if err != nil {
				broadcastErr = errors.Wrap(err, "broadcast recv failed")
				done <- true
				return
			}
			broadcastStatus = &broadcastResponse.Status
			if broadcastResponse.Status == common.Status_SUCCESS {
				done <- true
				return
			}
			if broadcastResponse.Status != common.Status_SUCCESS {
				broadcastErr = errors.Errorf("broadcast response is not success %v", broadcastResponse.Status)
				done <- true
				return
			}
		}
	}()
	if err := broadcastStream.Send(&common.Envelope{
		Payload:   envelope.Payload,
		Signature: envelope.Signature,
	}); err != nil {
		return nil, errors.Wrap(err, "failed to send envelope to orderer")
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
	errs := make(chan error, 1)
	// Validate envelope
	if envelope == nil {
		errs <- errors.New("envelope is nil")
		return responses, errs
	}
	// Establish connection to Ordering Service
	conn, err := grpc.Dial(o.url, o.grpcDialOption...)
	if err != nil {
		errs <- err
		return responses, errs
	}
	// Create atomic broadcast client
	broadcastStream, err := ab.NewAtomicBroadcastClient(conn).
		Deliver(context.Background())
	if err != nil {
		errs <- errors.Wrap(err, "NewAtomicBroadcastClient failed")
		return responses, errs
	}
	// Send block request envolope
	logger.Debugf("Requesting blocks from ordering service")
	if err := broadcastStream.Send(&common.Envelope{
		Payload:   envelope.Payload,
		Signature: envelope.Signature,
	}); err != nil {
		errs <- errors.Wrap(err, "failed to send block request to orderer")
		return responses, errs
	}
	// Receive blocks from the GRPC stream and put them on the channel
	go func() {
		defer conn.Close()
		for {
			response, err := broadcastStream.Recv()
			if err != nil {
				errs <- errors.Wrap(err, "recv from ordering service failed")
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
					errs <- errors.Errorf("error status from ordering service %s",
						t.Status)
					return
				}

			// Response is a requested block
			case *ab.DeliverResponse_Block:
				logger.Debug("Received block from ordering service")
				responses <- response.GetBlock()
			// Unknown response
			default:
				errs <- errors.Errorf("unknown response from ordering service %s", t)
				return
			}
		}
	}()

	return responses, errs
}
