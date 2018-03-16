/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package comm

import (
	"sync/atomic"

	"github.com/pkg/errors"

	"github.com/hyperledger/fabric-sdk-go/pkg/common/logging"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/options"
	fabcontext "github.com/hyperledger/fabric-sdk-go/pkg/common/providers/context"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/core"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/fab"
	"github.com/hyperledger/fabric-sdk-go/pkg/context"
	"github.com/hyperledger/fabric-sdk-go/pkg/core/config/comm"
	"github.com/hyperledger/fabric-sdk-go/pkg/core/config/endpoint"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

var logger = logging.NewLogger("fabsdk/fab")

const (
	// GRPC max message size (same as Fabric)
	maxCallRecvMsgSize = 100 * 1024 * 1024
	maxCallSendMsgSize = 100 * 1024 * 1024
)

// StreamProvider creates a GRPC stream
type StreamProvider func(conn *grpc.ClientConn) (grpc.ClientStream, error)

// GRPCConnection manages the GRPC connection and client stream
type GRPCConnection struct {
	context     fabcontext.Client
	chConfig    fab.ChannelCfg
	conn        *grpc.ClientConn
	stream      grpc.ClientStream
	commManager fab.CommManager
	tlsCertHash []byte
	done        int32
}

// NewConnection creates a new connection
func NewConnection(ctx fabcontext.Client, chConfig fab.ChannelCfg, streamProvider StreamProvider, url string, opts ...options.Opt) (*GRPCConnection, error) {
	if url == "" {
		return nil, errors.New("server URL not specified")
	}

	params := defaultParams()
	options.Apply(params, opts)

	dialOpts, err := newDialOpts(ctx.Config(), url, params)
	if err != nil {
		return nil, err
	}

	reqCtx, cancel := context.NewRequest(ctx, context.WithTimeout(params.connectTimeout))
	defer cancel()

	commManager, ok := context.RequestCommManager(reqCtx)
	if !ok {
		return nil, errors.New("unable to get comm manager")
	}

	grpcconn, err := commManager.DialContext(reqCtx, endpoint.ToAddress(url), dialOpts...)
	if err != nil {
		return nil, errors.Wrapf(err, "could not connect to %s", url)
	}

	stream, err := streamProvider(grpcconn)
	if err != nil {
		commManager.ReleaseConn(grpcconn)
		return nil, errors.Wrapf(err, "could not create stream to %s", url)
	}

	if stream == nil {
		return nil, errors.New("unexpected nil stream received from provider")
	}

	return &GRPCConnection{
		context:     ctx,
		chConfig:    chConfig,
		commManager: commManager,
		conn:        grpcconn,
		stream:      stream,
		tlsCertHash: comm.TLSCertHash(ctx.Config()),
	}, nil
}

// ChannelConfig returns the channel configuration
func (c *GRPCConnection) ChannelConfig() fab.ChannelCfg {
	return c.chConfig
}

// Close closes the connection
func (c *GRPCConnection) Close() {
	if !c.setClosed() {
		logger.Debugf("Already closed")
		return
	}

	logger.Debug("Closing stream....")
	if err := c.stream.CloseSend(); err != nil {
		logger.Warnf("error closing GRPC stream: %s", err)
	}

	logger.Debug("Releasing connection....")
	c.commManager.ReleaseConn(c.conn)

	logger.Debug("... connection successfully closed.")
}

// Closed returns true if the connection has been closed
func (c *GRPCConnection) Closed() bool {
	return atomic.LoadInt32(&c.done) == 1
}

func (c *GRPCConnection) setClosed() bool {
	return atomic.CompareAndSwapInt32(&c.done, 0, 1)
}

// Stream returns the GRPC stream
func (c *GRPCConnection) Stream() grpc.Stream {
	return c.stream
}

// TLSCertHash returns the hash of the TLS cert
func (c *GRPCConnection) TLSCertHash() []byte {
	return c.tlsCertHash
}

// Context returns the context of the client establishing the connection
func (c *GRPCConnection) Context() fabcontext.Client {
	return c.context
}

func newDialOpts(config core.Config, url string, params *params) ([]grpc.DialOption, error) {
	var dialOpts []grpc.DialOption

	if params.keepAliveParams.Time > 0 || params.keepAliveParams.Timeout > 0 {
		dialOpts = append(dialOpts, grpc.WithKeepaliveParams(params.keepAliveParams))
	}

	dialOpts = append(dialOpts, grpc.WithDefaultCallOptions(grpc.FailFast(params.failFast)))

	if endpoint.AttemptSecured(url, params.insecure) {
		tlsConfig, err := comm.TLSConfig(params.certificate, params.hostOverride, config)
		if err != nil {
			return nil, err
		}
		dialOpts = append(dialOpts, grpc.WithTransportCredentials(credentials.NewTLS(tlsConfig)))
		logger.Debugf("Creating a secure connection to [%s] with TLS HostOverride [%s]", url, params.hostOverride)
	} else {
		logger.Debugf("Creating an insecure connection [%s]", url)
		dialOpts = append(dialOpts, grpc.WithInsecure())
	}

	dialOpts = append(dialOpts, grpc.WithDefaultCallOptions(grpc.MaxCallRecvMsgSize(maxCallRecvMsgSize),
		grpc.MaxCallSendMsgSize(maxCallSendMsgSize)))

	return dialOpts, nil
}
