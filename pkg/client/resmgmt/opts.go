/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package resmgmt

import (
	"time"

	"github.com/hyperledger/fabric-sdk-go/pkg/context/api/fab"
)

//WithTargets encapsulates fab.Peer targets to resmgmtclient RequestOption
func WithTargets(targets ...fab.Peer) RequestOption {
	return func(opts *Opts) error {
		opts.Targets = targets
		return nil
	}
}

//WithTargetFilter encapsulates  resmgmtclient TargetFilter targets to resmgmtclient RequestOption
func WithTargetFilter(targetFilter TargetFilter) RequestOption {
	return func(opts *Opts) error {
		opts.TargetFilter = targetFilter
		return nil
	}
}

//WithTimeout encapsulates time.Duration to resmgmtclient RequestOption
func WithTimeout(timeout time.Duration) RequestOption {
	return func(opts *Opts) error {
		opts.Timeout = timeout
		return nil
	}
}

//WithOrdererID encapsulates OrdererID to RequestOption
func WithOrdererID(ordererID string) RequestOption {
	return func(opts *Opts) error {
		opts.OrdererID = ordererID
		return nil
	}
}
