/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package discovery

import (
	"context"
	"strings"
	"sync"

	"github.com/hyperledger/fabric-protos-go/discovery"
	discclient "github.com/hyperledger/fabric-sdk-go/internal/github.com/hyperledger/fabric/discovery/client"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/logging"
	fabcontext "github.com/hyperledger/fabric-sdk-go/pkg/common/providers/context"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/fab"
	corecomm "github.com/hyperledger/fabric-sdk-go/pkg/core/config/comm"
	"github.com/hyperledger/fabric-sdk-go/pkg/fab/comm"
	"github.com/pkg/errors"
	"google.golang.org/grpc"
)

var logger = logging.NewLogger("fabsdk/fab")

const (
	signerCacheSize = 10 // TODO: set an appropriate value (and perhaps make configurable)
)

//Client gives ability to send discovery request to multiple targets.
//There are cases when multiple targets requested and some of them are hanging, recommended to cancel ctx after first successful response.
//Note: "access denied" is a success response, so check for it after response evaluation.
type Client interface {
	Send(ctx context.Context, req *Request, targets ...fab.PeerConfig) (<-chan Response, error)
}

// Client implements a Discovery client
type client struct {
	ctx      fabcontext.Client
	authInfo *discovery.AuthInfo
}

// New returns a new Discover client
func New(ctx fabcontext.Client) (Client, error) {
	authInfo, err := newAuthInfo(ctx)
	if err != nil {
		return nil, err
	}

	return &client{
		ctx:      ctx,
		authInfo: authInfo,
	}, nil
}

// Response extends the response from the Discovery invocation on the peer
// by adding the endpoint URL of the peer that was invoked.
type Response interface {
	discclient.Response
	Target() string
	Error() error
}

// NewIndifferentFilter returns NoPriorities/NoExclusion filter.
// Note: this method was added just to allow users of Client to be able to use Endorsers method in response, which requires Filter as an argument.
// It's impossible to implement interface because Filter is placed under internal dir which is not available to end user.
// A user should filter peers by himself.
func NewIndifferentFilter() discclient.Filter {
	return discclient.NewFilter(discclient.NoPriorities, discclient.NoExclusion)
}

// Send retrieves information about channel peers, endorsers, and MSP config from the
// given set of peers. A channel of successful responses is returned and an error if there is not targets.
// Each Response contains Error method to check if there is an error.
func (c *client) Send(ctx context.Context, req *Request, targets ...fab.PeerConfig) (<-chan Response, error) {
	if len(targets) == 0 {
		return nil, errors.New("no targets specified")
	}

	//buffered channel is used because don't want to handle hanging goroutine on writing to the channel
	respCh := make(chan Response, len(targets))
	var requests sync.WaitGroup

	for _, t := range targets {
		requests.Add(1)

		go func(target fab.PeerConfig) {
			defer requests.Done()

			discoveryResponse, err := c.send(ctx, req.r, target)
			resp := response{target: target.URL, Response: discoveryResponse}

			if err != nil {
				if !isContextCanceled(err) {
					resp.err = errors.WithMessage(err, "From target: "+target.URL)
					logger.Debugf("... got discovery error response from [%s]: %s", target.URL, err)
				} else {
					logger.Debugf("... request to [%s] cancelled", target.URL)
				}
			} else {
				logger.Debugf("... got discovery response from [%s]", target.URL)
			}

			respCh <- resp
		}(t)
	}

	//this method is responsible for respCh channel, so we need to wait until all workers are done and close respCh
	go func() {
		requests.Wait()
		close(respCh)
	}()

	return respCh, nil
}

func (c *client) send(reqCtx context.Context, req *discclient.Request, target fab.PeerConfig) (discclient.Response, error) {
	opts := comm.OptsFromPeerConfig(&target)
	opts = append(opts, comm.WithConnectTimeout(c.ctx.EndpointConfig().Timeout(fab.DiscoveryConnection)))
	opts = append(opts, comm.WithParentContext(reqCtx))

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
	err    error
}

// Target returns the target peer URL
func (r response) Target() string {
	return r.target
}

// Error returns an error if it exists
func (r response) Error() error {
	return r.err
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

func isContextCanceled(err error) bool {
	return strings.Contains(err.Error(), context.Canceled.Error())
}
