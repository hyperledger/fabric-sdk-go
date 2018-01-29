/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package resmgmtclient

import (
	"time"

	fab "github.com/hyperledger/fabric-sdk-go/api/apifabclient"
)

//WithTargets encapsulates fab.Peer targets to resmgmtclient Option
func WithTargets(targets ...fab.Peer) Option {
	return func(opts *Opts) error {
		opts.Targets = targets
		return nil
	}
}

//WithTargetFilter encapsulates  resmgmtclient TargetFilter targets to resmgmtclient Option
func WithTargetFilter(targetFilter TargetFilter) Option {
	return func(opts *Opts) error {
		opts.TargetFilter = targetFilter
		return nil
	}
}

//WithTimeout encapsulates time.Duration to resmgmtclient Option
func WithTimeout(timeout time.Duration) Option {
	return func(opts *Opts) error {
		opts.Timeout = timeout
		return nil
	}
}
