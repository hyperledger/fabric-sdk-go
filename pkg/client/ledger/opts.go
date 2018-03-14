/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package ledger

import (
	"time"

	"github.com/hyperledger/fabric-sdk-go/pkg/common/context"
	"github.com/hyperledger/fabric-sdk-go/pkg/context/api/fab"
	"github.com/hyperledger/fabric-sdk-go/pkg/core/config"
	"github.com/pkg/errors"
)

const (
	minTargets = 1
	maxTargets = 1
)

// ClientOption describes a functional parameter for the New constructor
type ClientOption func(*Client) error

// WithDefaultTargetFilter option to configure new
func WithDefaultTargetFilter(filter TargetFilter) ClientOption {
	return func(rmc *Client) error {
		rmc.filter = filter
		return nil
	}
}

//RequestOption func for each requestOptions argument
type RequestOption func(ctx context.Client, opts *requestOptions) error

// TargetFilter allows for filtering target peers
type TargetFilter interface {
	// Accept returns true if peer should be included in the list of target peers
	Accept(peer fab.Peer) bool
}

//requestOptions contains options for operations performed by LedgerClient
type requestOptions struct {
	Targets      []fab.Peer    // target peers
	TargetFilter TargetFilter  // target filter
	MaxTargets   int           // maximum number of targets to select
	MinTargets   int           // min number of targets that have to respond with no error (or agree on result)
	Timeout      time.Duration //timeout options for QueryInfo,QueryBlockByHash,QueryBlock,QueryTransaction,QueryConfig
}

//WithTargets encapsulates fab.Peer targets to ledger RequestOption
func WithTargets(targets ...fab.Peer) RequestOption {
	return func(ctx context.Client, opts *requestOptions) error {
		opts.Targets = targets
		return nil
	}
}

// WithTargetURLs allows overriding of the target peers for the request.
// Targets are specified by URL, and the SDK will create the underlying peer
// objects.
func WithTargetURLs(urls ...string) RequestOption {
	return func(ctx context.Client, opts *requestOptions) error {

		var targets []fab.Peer

		for _, url := range urls {

			peerCfg, err := config.NetworkPeerConfigFromURL(ctx.Config(), url)
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

//WithTargetFilter encapsulates TargetFilter targets to ledger RequestOption
func WithTargetFilter(targetFilter TargetFilter) RequestOption {
	return func(ctx context.Client, opts *requestOptions) error {
		opts.TargetFilter = targetFilter
		return nil
	}
}

//WithMaxTargets encapsulates max targets to ledger RequestOption
func WithMaxTargets(maxTargets int) RequestOption {
	return func(ctx context.Client, opts *requestOptions) error {
		opts.MaxTargets = maxTargets
		return nil
	}
}

//WithMinTargets encapsulates min targets to ledger RequestOption
func WithMinTargets(minTargets int) RequestOption {
	return func(ctx context.Client, opts *requestOptions) error {
		opts.MinTargets = minTargets
		return nil
	}
}

//WithTimeout encapsulates timeout to ledger RequestOption
//for QueryInfo,QueryBlockByHash,QueryBlock,QueryTransaction,QueryConfig functions
func WithTimeout(timeout time.Duration) RequestOption {
	return func(ctx context.Client, opts *requestOptions) error {
		opts.Timeout = timeout
		return nil
	}
}
