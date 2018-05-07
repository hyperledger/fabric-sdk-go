/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package ledger

import (
	reqContext "context"
	"time"

	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/context"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/fab"
	"github.com/hyperledger/fabric-sdk-go/pkg/fab/comm"
	"github.com/pkg/errors"
)

const (
	minTargets = 1
	maxTargets = 1
)

// ClientOption describes a functional parameter for the New constructor
type ClientOption func(*Client) error

// WithDefaultTargetFilter option to configure new
func WithDefaultTargetFilter(filter fab.TargetFilter) ClientOption {
	return func(rmc *Client) error {
		rmc.filter = filter
		return nil
	}
}

//RequestOption func for each requestOptions argument
type RequestOption func(ctx context.Client, opts *requestOptions) error

//requestOptions contains options for operations performed by LedgerClient
type requestOptions struct {
	Targets       []fab.Peer                        // target peers
	TargetFilter  fab.TargetFilter                  // target filter
	MaxTargets    int                               // maximum number of targets to select
	MinTargets    int                               // min number of targets that have to respond with no error (or agree on result)
	Timeouts      map[fab.TimeoutType]time.Duration //timeout options for ledger query operations
	ParentContext reqContext.Context                //parent grpc context for ledger operations
}

//WithTargets allows for overriding of the target peers per request.
func WithTargets(targets ...fab.Peer) RequestOption {
	return func(ctx context.Client, opts *requestOptions) error {

		// Validate targets
		for _, t := range targets {
			if t == nil {
				return errors.New("target is nil")
			}
		}

		opts.Targets = targets
		return nil
	}
}

// WithTargetEndpoints allows overriding of the target peers per request.
// Targets are specified by name or URL, and the SDK will create the underlying peer objects.
func WithTargetEndpoints(keys ...string) RequestOption {
	return func(ctx context.Client, opts *requestOptions) error {

		var targets []fab.Peer

		for _, url := range keys {

			peerCfg, err := comm.NetworkPeerConfig(ctx.EndpointConfig(), url)
			if err != nil {
				return err
			}

			peer, err := ctx.InfraProvider().CreatePeerFromConfig(peerCfg)
			if err != nil {
				return errors.WithMessage(err, "creating peer from config failed")
			}

			targets = append(targets, peer)
		}

		return WithTargets(targets...)(ctx, opts)
	}
}

// WithTargetFilter specifies a per-request target peer-filter.
func WithTargetFilter(targetFilter fab.TargetFilter) RequestOption {
	return func(ctx context.Client, opts *requestOptions) error {
		opts.TargetFilter = targetFilter
		return nil
	}
}

//WithMaxTargets specifies maximum number of targets to select per request.
// Default value for maximum number of targets is 1.
func WithMaxTargets(maxTargets int) RequestOption {
	return func(ctx context.Client, opts *requestOptions) error {
		opts.MaxTargets = maxTargets
		return nil
	}
}

//WithMinTargets specifies minimum number of targets that have to respond with no error (or agree on result).
// Default value for minimum number of targets is 1.
func WithMinTargets(minTargets int) RequestOption {
	return func(ctx context.Client, opts *requestOptions) error {
		opts.MinTargets = minTargets
		return nil
	}
}

//WithTimeout encapsulates key value pairs of timeout type, timeout duration to Options
//for QueryInfo, QueryBlock, QueryBlockByHash,  QueryBlockByTxID, QueryTransaction, QueryConfig functions
func WithTimeout(timeoutType fab.TimeoutType, timeout time.Duration) RequestOption {
	return func(ctx context.Client, o *requestOptions) error {
		if o.Timeouts == nil {
			o.Timeouts = make(map[fab.TimeoutType]time.Duration)
		}
		o.Timeouts[timeoutType] = timeout
		return nil
	}
}

//WithParentContext encapsulates grpc parent context
func WithParentContext(parentContext reqContext.Context) RequestOption {
	return func(ctx context.Client, o *requestOptions) error {
		o.ParentContext = parentContext
		return nil
	}
}
