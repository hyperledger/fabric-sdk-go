/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package resmgmt

import (
	reqContext "context"
	"time"

	"github.com/hyperledger/fabric-sdk-go/pkg/common/errors/retry"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/context"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/fab"
	"github.com/hyperledger/fabric-sdk-go/pkg/fab/comm"
	"github.com/pkg/errors"
)

// WithTargets allows overriding of the target peers for the request.
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

// WithTargetEndpoints allows overriding of the target peers for the request.
// Targets are specified by name or URL, and the SDK will create the underlying peer
// objects.
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

// WithTargetFilter enables a target filter for the request.
func WithTargetFilter(targetFilter fab.TargetFilter) RequestOption {
	return func(ctx context.Client, opts *requestOptions) error {
		opts.TargetFilter = targetFilter
		return nil
	}
}

//WithTimeout encapsulates key value pairs of timeout type, timeout duration to Options
//if not provided, default timeout configuration from config will be used
func WithTimeout(timeoutType fab.TimeoutType, timeout time.Duration) RequestOption {
	return func(ctx context.Client, o *requestOptions) error {
		if o.Timeouts == nil {
			o.Timeouts = make(map[fab.TimeoutType]time.Duration)
		}
		o.Timeouts[timeoutType] = timeout
		return nil
	}
}

// WithOrdererEndpoint allows an orderer to be specified for the request.
// The orderer will be looked-up based on the key argument.
// key argument can be a name or url
func WithOrdererEndpoint(key string) RequestOption {

	return func(ctx context.Client, opts *requestOptions) error {

		ordererCfg, found := ctx.EndpointConfig().OrdererConfig(key)
		if !found {
			return errors.Errorf("orderer not found for url : %s", key)
		}

		orderer, err := ctx.InfraProvider().CreateOrdererFromConfig(ordererCfg)
		if err != nil {
			return errors.WithMessage(err, "creating orderer from config failed")
		}

		return WithOrderer(orderer)(ctx, opts)
	}
}

// WithOrderer allows an orderer to be specified for the request.
func WithOrderer(orderer fab.Orderer) RequestOption {
	return func(ctx context.Client, opts *requestOptions) error {
		opts.Orderer = orderer
		return nil
	}
}

//WithParentContext encapsulates grpc parent context.
func WithParentContext(parentContext reqContext.Context) RequestOption {
	return func(ctx context.Client, o *requestOptions) error {
		o.ParentContext = parentContext
		return nil
	}
}

// WithRetry sets retry options.
func WithRetry(retryOpt retry.Opts) RequestOption {
	return func(ctx context.Client, o *requestOptions) error {
		o.Retry = retryOpt
		return nil
	}
}
