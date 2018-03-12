/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package resmgmt

import (
	"time"

	"github.com/hyperledger/fabric-sdk-go/pkg/common/context"
	"github.com/hyperledger/fabric-sdk-go/pkg/context/api/fab"
	"github.com/pkg/errors"
)

//WithTargets encapsulates fab.Peer targets to resmgmtclient RequestOption
func WithTargets(targets ...fab.Peer) RequestOption {
	return func(ctx context.Client, opts *requestOptions) error {
		opts.Targets = targets
		return nil
	}
}

//WithTarget encapsulates fab.Peer target to RequestOption
func WithTarget(target fab.Peer) RequestOption {
	return func(ctx context.Client, opts *requestOptions) error {
		opts.Targets = []fab.Peer{target}
		return nil
	}
}

//WithTargetFilter encapsulates  resmgmtclient TargetFilter targets to resmgmtclient RequestOption
func WithTargetFilter(targetFilter TargetFilter) RequestOption {
	return func(ctx context.Client, opts *requestOptions) error {
		opts.TargetFilter = targetFilter
		return nil
	}
}

//WithTimeout encapsulates time.Duration to resmgmtclient RequestOption
func WithTimeout(timeout time.Duration) RequestOption {
	return func(ctx context.Client, opts *requestOptions) error {
		opts.Timeout = timeout
		return nil
	}
}

//WithOrdererURL allows an orderer to be specified for the request.
//The orderer will be looked-up based on the url argument.
//A default orderer implementation will be used.
func WithOrdererURL(ordererID string) RequestOption {
	return func(ctx context.Client, opts *requestOptions) error {

		ordererCfg, err := ctx.Config().OrdererConfig(ordererID)
		if err != nil {
			return errors.WithMessage(err, "orderer not found")
		}
		if ordererCfg == nil {
			return errors.New("orderer not found")
		}

		orderer, err := ctx.InfraProvider().CreateOrdererFromConfig(ordererCfg)
		if err != nil {
			return errors.WithMessage(err, "creating orderer from config failed")
		}

		return WithOrderer(orderer)(ctx, opts)
	}
}

//WithOrderer allows an orderer to be specified for the request.
func WithOrderer(orderer fab.Orderer) RequestOption {
	return func(ctx context.Client, opts *requestOptions) error {
		opts.Orderer = orderer
		return nil
	}
}
