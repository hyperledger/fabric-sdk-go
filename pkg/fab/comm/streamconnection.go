/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package comm

import (
	"sync"

	"github.com/pkg/errors"

	"github.com/hyperledger/fabric-sdk-go/pkg/client/common/verifier"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/options"
	fabcontext "github.com/hyperledger/fabric-sdk-go/pkg/common/providers/context"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/fab"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/peer"
)

// StreamProvider creates a GRPC stream
type StreamProvider func(conn *grpc.ClientConn) (grpc.ClientStream, func(), error)

// StreamConnection manages the GRPC connection and client stream
type StreamConnection struct {
	*GRPCConnection
	chConfig fab.ChannelCfg
	stream   grpc.ClientStream
	cancel   func()
	lock     sync.Mutex
}

// NewStreamConnection creates a new connection with stream
func NewStreamConnection(ctx fabcontext.Client, chConfig fab.ChannelCfg, streamProvider StreamProvider, url string, opts ...options.Opt) (*StreamConnection, error) {
	conn, err := NewConnection(ctx, url, opts...)
	if err != nil {
		return nil, err
	}

	stream, cancel, err := streamProvider(conn.conn)
	if err != nil {
		conn.commManager.ReleaseConn(conn.conn)
		return nil, errors.Wrapf(err, "could not create stream to %s", url)
	}

	if stream == nil {
		return nil, errors.New("unexpected nil stream received from provider")
	}

	peer, ok := peer.FromContext(stream.Context())
	if !ok || peer == nil {
		//return error - certificate is not available
		return nil, errors.Wrap(err, "No peer cert in GRPC stream")

	}

	if peer.AuthInfo != nil {
		tlsInfo := peer.AuthInfo.(credentials.TLSInfo)
		for _, peercert := range tlsInfo.State.PeerCertificates {
			err := verifier.ValidateCertificateDates(peercert)
			if err != nil {
				logger.Error(err)
				return nil, errors.Wrapf(err, "error validating certificate dates for [%v]", peercert.Subject)
			}
		}
	}

	return &StreamConnection{
		GRPCConnection: conn,
		chConfig:       chConfig,
		stream:         stream,
		cancel:         cancel,
	}, nil
}

// ChannelConfig returns the channel configuration
func (c *StreamConnection) ChannelConfig() fab.ChannelCfg {
	return c.chConfig
}

// Close closes the connection
func (c *StreamConnection) Close() {
	c.lock.Lock()
	defer c.lock.Unlock()

	if c.Closed() {
		return
	}

	logger.Debug("Closing stream....")

	c.cancel()

	if err := c.stream.CloseSend(); err != nil {
		logger.Warnf("error closing GRPC stream: %s", err)
	}

	c.GRPCConnection.Close()
}

// Stream returns the GRPC stream
func (c *StreamConnection) Stream() grpc.ClientStream {
	return c.stream
}
