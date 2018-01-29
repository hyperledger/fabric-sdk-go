/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package chmgmtclient

//WithOrdererID encapsulates OrdererID to Option
func WithOrdererID(ordererID string) Option {
	return func(opts *Opts) error {
		opts.OrdererID = ordererID
		return nil
	}
}
