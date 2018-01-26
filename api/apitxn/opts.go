/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package apitxn

import (
	"time"
)

//WithTimeout encapsulates time.Duration to Option
func WithTimeout(timeout time.Duration) Option {
	return func(opts *Opts) error {
		opts.Timeout = timeout
		return nil
	}
}

//WithNotifier encapsulates Response to Option
func WithNotifier(notifier chan Response) Option {
	return func(opts *Opts) error {
		opts.Notifier = notifier
		return nil
	}
}

//WithProposalProcessor encapsulates ProposalProcessors to Option
func WithProposalProcessor(proposalProcessors ...ProposalProcessor) Option {
	return func(opts *Opts) error {
		opts.ProposalProcessors = proposalProcessors
		return nil
	}
}
