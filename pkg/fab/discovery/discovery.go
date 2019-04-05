/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package discovery

import (
	"context"
	"sync"

	discclient "github.com/hyperledger/fabric-sdk-go/internal/github.com/hyperledger/fabric/discovery/client"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/errors/multi"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/logging"
	fabcontext "github.com/hyperledger/fabric-sdk-go/pkg/common/providers/context"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/fab"
	corecomm "github.com/hyperledger/fabric-sdk-go/pkg/core/config/comm"
	"github.com/hyperledger/fabric-sdk-go/pkg/fab/comm"
	"github.com/hyperledger/fabric-sdk-go/third_party/github.com/hyperledger/fabric/protos/discovery"
	"github.com/pkg/errors"
	"google.golang.org/grpc"
)

var logger = logging.NewLogger("fabsdk/fab")

const (
	signerCacheSize = 10 // TODO: set an appropriate value (and perhaps make configurable)
)

// Client implements a Discovery client
type Client struct {
	ctx      fabcontext.Client
	authInfo *discovery.AuthInfo
}

// New returns a new Discover client
func New(ctx fabcontext.Client) (*Client, error) {
	authInfo, err := newAuthInfo(ctx)
	if err != nil {
		return nil, err
	}

	return &Client{
		ctx:      ctx,
		authInfo: authInfo,
	}, nil
}

// Response extends the response from the Discovery invocation on the peer
// by adding the endpoint URL of the peer that was invoked.
type Response interface {
	discclient.Response
	Target() string
}

// Send retrieves information about channel peers, endorsers, and MSP config from the
// given set of peers. A set of successful responses is returned and/or an error
// is returned from each of the peers that was unsuccessful (note that if more than one peer returned
// an error then the returned error may be cast to multi.Errors).
func (c *Client) Send(ctx context.Context, req *discclient.Request, targets ...fab.PeerConfig) ([]Response, error) {
	if len(targets) == 0 {
		return nil, errors.New("no targets specified")
	}

	var lock sync.Mutex
	var wg sync.WaitGroup
	wg.Add(len(targets))

	var responses []Response
	var errs error

	for _, t := range targets {
		go func(target fab.PeerConfig) {
			defer wg.Done()

			resp, err := c.send(ctx, req, target)
			lock.Lock()
			if err != nil {
				errs = multi.Append(errs, errors.WithMessage(err, "From target: "+target.URL))
				logger.Debugf("... got discovery error response from [%s]: %s", target.URL, err)
			} else {
				responses = append(responses, &response{Response: resp, target: target.URL})
				logger.Debugf("... got discovery response from [%s]", target.URL)
			}
			lock.Unlock()
		}(t)
	}
	wg.Wait()

	return responses, errs
}

func (c *Client) send(reqCtx context.Context, req *discclient.Request, target fab.PeerConfig) (discclient.Response, error) {
	opts := comm.OptsFromPeerConfig(&target)
	opts = append(opts, comm.WithConnectTimeout(c.ctx.EndpointConfig().Timeout(fab.DiscoveryConnection)))

	conn, err := comm.NewConnection(c.ctx, target.URL, opts...)
	if err != nil {
		return nil, err
	}
	defer conn.Close()

	discClient := discclient.NewClient(
		func() (*grpc.ClientConn, error) {
			return conn.ClientConn(), nil
		},
		func(msg []byte) ([]byte, error) {
			return c.ctx.SigningManager().Sign(msg, c.ctx.PrivateKey())
		},
		signerCacheSize,
	)
	return discClient.Send(reqCtx, req, c.authInfo)
}

type response struct {
	discclient.Response
	target string
}

// Target returns the target peer URL
func (r *response) Target() string {
	return r.target
}

func newAuthInfo(ctx fabcontext.Client) (*discovery.AuthInfo, error) {
	identity, err := ctx.Serialize()
	if err != nil {
		return nil, err
	}

	hash, err := corecomm.TLSCertHash(ctx.EndpointConfig())
	if err != nil {
		return nil, errors.Wrapf(err, "failed to get tls cert hash")
	}

	return &discovery.AuthInfo{
		ClientIdentity:    identity,
		ClientTlsCertHash: hash,
	}, nil
}
