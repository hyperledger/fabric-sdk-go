/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package chclient

import (
	"time"

	"github.com/hyperledger/fabric-sdk-go/api/apifabclient"
	"github.com/hyperledger/fabric-sdk-go/pkg/errors/retry"
)

//WithTimeout encapsulates time.Duration to Option
func WithTimeout(timeout time.Duration) Option {
	return func(opts *Opts) error {
		opts.Timeout = timeout
		return nil
	}
}

//WithProposalProcessor encapsulates ProposalProcessors to Option
func WithProposalProcessor(proposalProcessors ...apifabclient.ProposalProcessor) Option {
	return func(opts *Opts) error {
		opts.ProposalProcessors = proposalProcessors
		return nil
	}
}

// WithRetry option to configure retries
func WithRetry(opt retry.Opts) Option {
	return func(opts *Opts) error {
		opts.Retry = opt
		return nil
	}
}
