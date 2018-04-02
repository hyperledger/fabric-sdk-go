/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package event

// ClientOption describes a functional parameter for the New constructor
type ClientOption func(*Client) error

// WithBlockEvents indicates that block events are to be received.
// Note that the caller must have sufficient privileges for this option.
func WithBlockEvents() ClientOption {
	return func(c *Client) error {
		c.permitBlockEvents = true
		return nil
	}
}
