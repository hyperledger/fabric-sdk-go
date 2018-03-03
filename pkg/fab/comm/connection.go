/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package comm

import (
	"context"
	"sync/atomic"

	"github.com/pkg/errors"

	fabcontext "github.com/hyperledger/fabric-sdk-go/pkg/context"
	"github.com/hyperledger/fabric-sdk-go/pkg/context/api/core"
	"github.com/hyperledger/fabric-sdk-go/pkg/core/config/comm"
	"github.com/hyperledger/fabric-sdk-go/pkg/core/config/urlutil"
	"github.com/hyperledger/fabric-sdk-go/pkg/logging"
	"github.com/hyperledger/fabric-sdk-go/pkg/options"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

var logger = logging.NewLogger("fabric_sdk_go")

// StreamProvider creates a GRPC stream
type StreamProvider func(conn *grpc.ClientConn) (grpc.ClientStream, error)

// GRPCConnection manages the GRPC connection and client stream
type GRPCConnection struct {
	channelID   string
	conn        *grpc.ClientConn
	stream      grpc.ClientStream
	context     fabcontext.Client
	tlsCertHash []byte
	done        int32
}

// NewConnection creates a new connection
func NewConnection(ctx fabcontext.Client, channelID string, streamProvider StreamProvider, url string, opts ...options.Opt) (*GRPCConnection, error) {
	if url == "" {
		return nil, errors.New("server URL not specified")
	}

	params := defaultParams()
	options.Apply(params, opts)

	dialOpts, err := newDialOpts(ctx.Config(), url, params)
	if err != nil {
		return nil, err
	}

	grpcctx := context.Background()
	grpcctx, cancel := context.WithTimeout(grpcctx, params.connectTimeout)
	defer cancel()

	grpcconn, err := grpc.DialContext(grpcctx, urlutil.ToAddress(url), dialOpts...)
	if err != nil {
		return nil, errors.Wrapf(err, "could not connect to %s", url)
	}

	stream, err := streamProvider(grpcconn)
	if err != nil {
		if err := grpcconn.Close(); err != nil {
			logger.Warnf("error closing GRPC connection: %s", err)
		}
		return nil, errors.Wrapf(err, "could not create stream to %s", url)
	}

	if stream == nil {
		return nil, errors.New("unexpected nil stream received from provider")
	}

	return &GRPCConnection{
		channelID:   channelID,
		conn:        grpcconn,
		stream:      stream,
		context:     ctx,
		tlsCertHash: comm.TLSCertHash(ctx.Config()),
	}, nil
}

// ChannelID returns the ID of the channel
func (c *GRPCConnection) ChannelID() string {
	return c.channelID
}

// Close closes the connection
func (c *GRPCConnection) Close() {
	if !c.setClosed() {
		logger.Debugf("Already closed")
		return
	}

	logger.Debugf("Closing stream....")
	if err := c.stream.CloseSend(); err != nil {
		logger.Warnf("error closing GRPC stream: %s", err)
	}

	logger.Debugf("Closing connection....")
	if err := c.conn.Close(); err != nil {
		logger.Warnf("error closing GRPC connection: %s", err)
	}
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

	if urlutil.IsTLSEnabled(url) {
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

	return dialOpts, nil
}
