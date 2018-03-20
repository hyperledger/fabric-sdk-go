/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package orderer

import (
	reqContext "context"
	"crypto/x509"
	"time"

	"github.com/pkg/errors"
	"github.com/spf13/cast"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/keepalive"
	grpcstatus "google.golang.org/grpc/status"

	ab "github.com/hyperledger/fabric-sdk-go/internal/github.com/hyperledger/fabric/protos/orderer"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/errors/status"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/logging"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/core"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/fab"
	"github.com/hyperledger/fabric-sdk-go/pkg/context"
	"github.com/hyperledger/fabric-sdk-go/pkg/core/config/comm"
	"github.com/hyperledger/fabric-sdk-go/pkg/core/config/endpoint"
	"github.com/hyperledger/fabric-sdk-go/third_party/github.com/hyperledger/fabric/protos/common"
)

var logger = logging.NewLogger("fabsdk/fab")

const (
	// GRPC max message size (same as Fabric)
	maxCallRecvMsgSize = 100 * 1024 * 1024
	maxCallSendMsgSize = 100 * 1024 * 1024
)

// Orderer allows a client to broadcast a transaction.
type Orderer struct {
	config         core.Config
	url            string
	serverName     string
	tlsCACert      *x509.Certificate
	grpcDialOption []grpc.DialOption
	kap            keepalive.ClientParameters
	dialTimeout    time.Duration
	failFast       bool
	allowInsecure  bool
	commManager    fab.CommManager
}

// Option describes a functional parameter for the New constructor
type Option func(*Orderer) error

// New Returns a Orderer instance
func New(config core.Config, opts ...Option) (*Orderer, error) {
	orderer := &Orderer{
		config:      config,
		commManager: &defCommManager{},
	}

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
	if endpoint.AttemptSecured(orderer.url, orderer.allowInsecure) {
		//tls config
		tlsConfig, err := comm.TLSConfig(orderer.tlsCACert, orderer.serverName, config)
		if err != nil {
			return nil, err
		}
		grpcOpts = append(grpcOpts, grpc.WithTransportCredentials(credentials.NewTLS(tlsConfig)))
	} else {
		grpcOpts = append(grpcOpts, grpc.WithInsecure())
	}

	grpcOpts = append(grpcOpts, grpc.WithDefaultCallOptions(grpc.MaxCallRecvMsgSize(maxCallRecvMsgSize),
		grpc.MaxCallSendMsgSize(maxCallSendMsgSize)))

	orderer.dialTimeout = config.TimeoutOrDefault(core.OrdererConnection)
	orderer.url = endpoint.ToAddress(orderer.url)
	orderer.grpcDialOption = grpcOpts

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
	allowInsecure, ok := ordererCfg.GRPCOptions["allow-insecure"].(bool)
	if ok {
		return allowInsecure
	}
	return false
}

func (o *Orderer) conn(ctx reqContext.Context) (*grpc.ClientConn, error) {
	// Establish connection to Ordering Service
	ctx, cancel := reqContext.WithTimeout(ctx, o.dialTimeout)
	defer cancel()

	commManager, ok := context.RequestCommManager(ctx)
	if !ok {
		commManager = o.commManager
	}

	return commManager.DialContext(ctx, o.url, o.grpcDialOption...)
}

func (o *Orderer) releaseConn(ctx reqContext.Context, conn *grpc.ClientConn) {
	commManager, ok := context.RequestCommManager(ctx)
	if !ok {
		commManager = o.commManager
	}

	commManager.ReleaseConn(conn)
}

// URL Get the Orderer url. Required property for the instance objects.
// Returns the address of the Orderer.
func (o *Orderer) URL() string {
	return o.url
}

// SendBroadcast Send the created transaction to Orderer.
func (o *Orderer) SendBroadcast(ctx reqContext.Context, envelope *fab.SignedEnvelope) (*common.Status, error) {
	conn, err := o.conn(ctx)
	if err != nil {
		rpcStatus, ok := grpcstatus.FromError(err)
		if ok {
			return nil, errors.WithMessage(status.NewFromGRPCStatus(rpcStatus), "connection failed")
		}

		return nil, status.New(status.OrdererClientStatus, status.ConnectionFailed.ToInt32(), err.Error(), nil)
	}
	defer o.releaseConn(ctx, conn)

	broadcastClient, err := ab.NewAtomicBroadcastClient(conn).Broadcast(ctx)
	if err != nil {
		rpcStatus, ok := grpcstatus.FromError(err)
		if ok {
			err = status.NewFromGRPCStatus(rpcStatus)
		}
		return nil, errors.Wrap(err, "NewAtomicBroadcastClient failed")
	}

	responses := make(chan common.Status)
	errs := make(chan error, 1)

	go broadcastStream(broadcastClient, responses, errs)

	err = broadcastClient.Send(&common.Envelope{
		Payload:   envelope.Payload,
		Signature: envelope.Signature,
	})
	if err != nil {
		return nil, errors.Wrap(err, "failed to send envelope to orderer")
	}
	if err = broadcastClient.CloseSend(); err != nil {
		logger.Debugf("unable to close broadcast client [%s]", err)
	}

	select {
	case broadcastStatus := <-responses:
		return &broadcastStatus, nil
	case broadcastErr := <-errs:
		return nil, broadcastErr
	}
}

func broadcastStream(broadcastClient ab.AtomicBroadcast_BroadcastClient, responses chan common.Status, errs chan error) {

	broadcastResponse, err := broadcastClient.Recv()
	logger.Debugf("Orderer.broadcastStream - response:%v, error:%v", broadcastResponse, err)
	if err != nil {
		rpcStatus, ok := grpcstatus.FromError(err)
		if ok {
			err = status.NewFromGRPCStatus(rpcStatus)
		}
		errs <- errors.Wrap(err, "broadcast recv failed")
		return
	}

	if broadcastResponse.Status != common.Status_SUCCESS {
		errs <- status.New(status.OrdererServerStatus, int32(broadcastResponse.Status), broadcastResponse.Info, nil)
		return
	}

	responses <- broadcastResponse.Status
}

// SendDeliver sends a deliver request to the ordering service and returns the
// blocks requested
// envelope: contains the seek request for blocks
func (o *Orderer) SendDeliver(ctx reqContext.Context, envelope *fab.SignedEnvelope) (chan *common.Block, chan error) {

	responses := make(chan *common.Block)
	errs := make(chan error, 1)

	conn, err := o.conn(ctx)
	if err != nil {
		rpcStatus, ok := grpcstatus.FromError(err)
		if ok {
			errs <- errors.WithMessage(status.NewFromGRPCStatus(rpcStatus), "connection failed")
			return responses, errs
		}

		errs <- status.New(status.OrdererClientStatus, status.ConnectionFailed.ToInt32(), err.Error(), nil)
		return responses, errs
	}

	// Create atomic broadcast client
	broadcastClient, err := ab.NewAtomicBroadcastClient(conn).Deliver(ctx)
	if err != nil {
		logger.Errorf("deliver failed [%s]", err)
		o.releaseConn(ctx, conn)

		errs <- errors.Wrap(err, "deliver failed")
		return responses, errs
	}

	// Receive blocks from the GRPC stream and put them on the channel
	go func() {
		blockStream(broadcastClient, responses, errs)
		o.releaseConn(ctx, conn)
	}()

	// Send block request envelope
	logger.Debugf("Requesting blocks from ordering service")
	err = broadcastClient.Send(&common.Envelope{
		Payload:   envelope.Payload,
		Signature: envelope.Signature,
	})
	if err != nil {
		o.releaseConn(ctx, conn)

		errs <- errors.Wrap(err, "failed to send block request to orderer")
		return responses, errs
	}

	if err = broadcastClient.CloseSend(); err != nil {
		logger.Debugf("unable to close deliver client [%s]", err)
	}

	return responses, errs
}

func blockStream(deliverClient ab.AtomicBroadcast_DeliverClient, responses chan *common.Block, errs chan error) {
	for {
		response, err := deliverClient.Recv()
		if err != nil {
			errs <- errors.Wrap(err, "recv from ordering service failed")
			return
		}
		// Assert response type
		switch t := response.Type.(type) {
		// Seek operation success, no more resposes
		case *ab.DeliverResponse_Status:
			logger.Debugf("Received deliver response status from ordering service: %s", t.Status)
			if t.Status != common.Status_SUCCESS {
				errs <- errors.Errorf("error status from ordering service %s", t.Status)
				return
			}
			close(responses)
			return

		// Response is a requested block
		case *ab.DeliverResponse_Block:
			logger.Debug("Received block from ordering service")
			responses <- response.GetBlock()
		// Unknown response
		default:
			errs <- errors.Errorf("unknown response type from ordering service %T", t)
			return
		}
	}
}

type defCommManager struct{}

func (*defCommManager) DialContext(ctx reqContext.Context, target string, opts ...grpc.DialOption) (conn *grpc.ClientConn, err error) {
	logger.Debugf("DialContext [%s]", target)
	opts = append(opts, grpc.WithBlock())
	return grpc.DialContext(ctx, target, opts...)
}

func (*defCommManager) ReleaseConn(conn *grpc.ClientConn) {
	logger.Debugf("ReleaseConn [%p]", conn)
	if err := conn.Close(); err != nil {
		logger.Debugf("unable to close connection [%s]", err)
	}
}
