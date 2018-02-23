/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package orderer

import (
	grpcContext "context"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/keepalive"
	grpcstatus "google.golang.org/grpc/status"

	ab "github.com/hyperledger/fabric-sdk-go/internal/github.com/hyperledger/fabric/protos/orderer"
	"github.com/hyperledger/fabric-sdk-go/pkg/core/config/comm"
	"github.com/hyperledger/fabric-sdk-go/pkg/errors/status"
	"github.com/hyperledger/fabric-sdk-go/third_party/github.com/hyperledger/fabric/protos/common"
	"github.com/spf13/cast"

	"crypto/x509"

	"github.com/hyperledger/fabric-sdk-go/pkg/context/api/core"
	"github.com/hyperledger/fabric-sdk-go/pkg/context/api/fab"
	"github.com/hyperledger/fabric-sdk-go/pkg/core/config/urlutil"
	"github.com/hyperledger/fabric-sdk-go/pkg/logging"
	"github.com/pkg/errors"
)

var logger = logging.NewLogger("fabric_sdk_go")

// Orderer allows a client to broadcast a transaction.
type Orderer struct {
	config               core.Config
	url                  string
	tlsCACert            *x509.Certificate
	serverName           string
	grpcDialOption       []grpc.DialOption
	kap                  keepalive.ClientParameters
	dialTimeout          time.Duration
	failFast             bool
	transportCredentials credentials.TransportCredentials
	secured              bool
	allowInsecure        bool
}

// Option describes a functional parameter for the New constructor
type Option func(*Orderer) error

// New Returns a Orderer instance
func New(config core.Config, opts ...Option) (*Orderer, error) {
	orderer := &Orderer{config: config}

	for _, opt := range opts {
		err := opt(orderer)

		if err != nil {
			return nil, err
		}
	}
	var grpcOpts []grpc.DialOption
	if orderer.kap.Time > 0 {
		grpcOpts = append(grpcOpts, grpc.WithKeepaliveParams(orderer.kap))
	}
	grpcOpts = append(grpcOpts, grpc.WithDefaultCallOptions(grpc.FailFast(orderer.failFast)))
	orderer.dialTimeout = config.TimeoutOrDefault(core.OrdererConnection)

	//tls config
	tlsConfig, err := comm.TLSConfig(orderer.tlsCACert, orderer.serverName, config)
	if err != nil {
		return nil, err
	}

	orderer.grpcDialOption = grpcOpts
	orderer.transportCredentials = credentials.NewTLS(tlsConfig)
	orderer.secured = urlutil.AttemptSecured(orderer.url)
	orderer.url = urlutil.ToAddress(orderer.url)

	return orderer, nil
}

// WithURL is a functional option for the orderer.New constructor that configures the orderer's URL.
func WithURL(url string) Option {
	return func(o *Orderer) error {
		o.url = url

		return nil
	}
}

// WithTLSCert is a functional option for the orderer.New constructor that configures the orderer's TLS certificate
func WithTLSCert(tlsCACert *x509.Certificate) Option {
	return func(o *Orderer) error {
		o.tlsCACert = tlsCACert

		return nil
	}
}

// WithServerName is a functional option for the orderer.New constructor that configures the orderer's server name
func WithServerName(serverName string) Option {
	return func(o *Orderer) error {
		o.serverName = serverName

		return nil
	}
}

// WithInsecure is a functional option for the orderer.New constructor that configures the orderer's grpc insecure option
func WithInsecure() Option {
	return func(o *Orderer) error {
		o.allowInsecure = true

		return nil
	}
}

// FromOrdererConfig is a functional option for the orderer.New constructor that configures a new orderer
// from a apiconfig.OrdererConfig struct
func FromOrdererConfig(ordererCfg *core.OrdererConfig) Option {
	return func(o *Orderer) error {
		o.url = ordererCfg.URL

		var err error

		o.tlsCACert, err = ordererCfg.TLSCACerts.TLSCert()

		if err != nil {
			//Ignore empty cert errors,
			errStatus, ok := err.(*status.Status)
			if !ok || errStatus.Code != status.EmptyCert.ToInt32() {
				return err
			}
		}

		o.serverName = getServerNameOverride(ordererCfg)
		o.kap = getKeepAliveOptions(ordererCfg)
		o.failFast = getFailFast(ordererCfg)
		o.allowInsecure = isInsecureConnectionAllowed(ordererCfg)

		return nil
	}
}

// FromOrdererName is a functional option for the orderer.New constructor that obtains an apiconfig.OrdererConfig
// by name from the apiconfig.Config supplied to the constructor, and then constructs a new orderer from it
func FromOrdererName(name string) Option {
	return func(o *Orderer) error {
		ordererCfg, err := o.config.OrdererConfig(name)

		if err != nil {
			return err
		}

		return FromOrdererConfig(ordererCfg)(o)
	}
}

func getServerNameOverride(ordererCfg *core.OrdererConfig) string {
	serverNameOverride := ""
	if str, ok := ordererCfg.GRPCOptions["ssl-target-name-override"].(string); ok {
		serverNameOverride = str
	}
	return serverNameOverride
}

func getFailFast(ordererCfg *core.OrdererConfig) bool {

	var failFast = true
	if ff, ok := ordererCfg.GRPCOptions["fail-fast"].(bool); ok {
		failFast = cast.ToBool(ff)
	}
	return failFast
}

func getKeepAliveOptions(ordererCfg *core.OrdererConfig) keepalive.ClientParameters {

	var kap keepalive.ClientParameters
	if kaTime, ok := ordererCfg.GRPCOptions["keep-alive-time"].(time.Duration); ok {
		kap.Time = cast.ToDuration(kaTime)
	}
	if kaTimeout, ok := ordererCfg.GRPCOptions["keep-alive-timeout"].(time.Duration); ok {
		kap.Timeout = cast.ToDuration(kaTimeout)
	}
	if kaPermit, ok := ordererCfg.GRPCOptions["keep-alive-permit"].(time.Duration); ok {
		kap.PermitWithoutStream = cast.ToBool(kaPermit)
	}
	return kap
}

func isInsecureConnectionAllowed(ordererCfg *core.OrdererConfig) bool {
	//allowInsecure used only when protocol is missing from URL
	allowInsecure := !urlutil.HasProtocol(ordererCfg.URL)
	boolVal, ok := ordererCfg.GRPCOptions["allow-insecure"].(bool)
	if ok {
		return allowInsecure && boolVal
	}
	return false
}

// URL Get the Orderer url. Required property for the instance objects.
// Returns the address of the Orderer.
func (o *Orderer) URL() string {
	return o.url
}

// SendBroadcast Send the created transaction to Orderer.
func (o *Orderer) SendBroadcast(envelope *fab.SignedEnvelope) (*common.Status, error) {
	return o.sendBroadcast(envelope, o.secured)
}

// SendBroadcast Send the created transaction to Orderer.
func (o *Orderer) sendBroadcast(envelope *fab.SignedEnvelope, secured bool) (*common.Status, error) {
	var grpcOpts []grpc.DialOption
	grpcOpts = append(grpcOpts, o.grpcDialOption...)
	if secured {
		grpcOpts = append(grpcOpts, grpc.WithTransportCredentials(o.transportCredentials))
	} else {
		grpcOpts = append(grpcOpts, grpc.WithInsecure())
	}

	ctx := grpcContext.Background()
	ctx, cancel := grpcContext.WithTimeout(ctx, o.dialTimeout)
	defer cancel()

	conn, err := grpc.DialContext(ctx, o.url, grpcOpts...)

	if err != nil {
		return nil, status.New(status.OrdererClientStatus, status.ConnectionFailed.ToInt32(), err.Error(), nil)
	}
	defer conn.Close()
	broadcastStream, err := ab.NewAtomicBroadcastClient(conn).Broadcast(ctx)
	if err != nil {
		rpcStatus, ok := grpcstatus.FromError(err)
		if ok {
			err = status.NewFromGRPCStatus(rpcStatus)
		}
		logger.Error("NewAtomicBroadcastClient failed, cause : ", err)
		if secured && o.allowInsecure {
			//If secured mode failed and allow insecure is enabled then retry in insecure mode
			logger.Debug("Secured sendBroadcast failed, attempting insecured")
			return o.sendBroadcast(envelope, false)
		}
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
				rpcStatus, ok := grpcstatus.FromError(err)
				if ok {
					err = status.NewFromGRPCStatus(rpcStatus)
				}
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
				broadcastErr = status.New(status.OrdererServerStatus, int32(broadcastResponse.Status), broadcastResponse.Info, nil)
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
func (o *Orderer) SendDeliver(envelope *fab.SignedEnvelope) (chan *common.Block, chan error, grpcContext.CancelFunc) {
	return o.sendDeliver(envelope, o.secured)
}

// SendDeliver sends a deliver request to the ordering service and returns the
// blocks requested
// envelope: contains the seek request for blocks
func (o *Orderer) sendDeliver(envelope *fab.SignedEnvelope, secured bool) (chan *common.Block, chan error, grpcContext.CancelFunc) {
	responses := make(chan *common.Block)
	errs := make(chan error, 1)

	// Establish connection to Ordering Service
	var grpcOpts []grpc.DialOption
	grpcOpts = append(grpcOpts, o.grpcDialOption...)
	if secured {
		grpcOpts = append(o.grpcDialOption, grpc.WithTransportCredentials(o.transportCredentials))
	} else {
		grpcOpts = append(o.grpcDialOption, grpc.WithInsecure())
	}

	ctx := grpcContext.Background()
	ctx, cancel := grpcContext.WithTimeout(ctx, o.dialTimeout)

	conn, err := grpc.DialContext(ctx, o.url, grpcOpts...)
	if err != nil {
		errs <- err
		return responses, errs, cancel
	}

	// Create atomic broadcast client
	broadcastStream, err := ab.NewAtomicBroadcastClient(conn).Deliver(ctx)
	if err != nil {
		logger.Error("NewAtomicBroadcastClient failed, cause : ", err)
		if secured && o.allowInsecure {
			//If secured mode failed and allow insecure is enabled then retry in insecure mode
			logger.Debug("Secured sendBroadcast failed, attempting insecured")

			cancel()
			return o.sendDeliver(envelope, false)
		}
		errs <- errors.Wrap(err, "NewAtomicBroadcastClient failed")
		return responses, errs, cancel
	}
	// Send block request envelope
	logger.Debugf("Requesting blocks from ordering service")
	if err := broadcastStream.Send(&common.Envelope{
		Payload:   envelope.Payload,
		Signature: envelope.Signature,
	}); err != nil {
		errs <- errors.Wrap(err, "failed to send block request to orderer")
		return responses, errs, cancel
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
	return responses, errs, cancel
}
