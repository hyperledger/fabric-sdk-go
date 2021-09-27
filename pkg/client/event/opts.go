/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package event

import (
	"time"

	"github.com/hyperledger/fabric-sdk-go/pkg/fab/events/deliverclient/seek"
)

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

// WithBlockNum indicates the block number from which events are to be received.
// Only deliverclient supports this
func WithBlockNum(from uint64) ClientOption {
	return func(c *Client) error {
		c.fromBlock = from
		return nil
	}
}

// WithSeekType indicates the  type of seek desired - newest, oldest or from given block
// Only deliverclient supports this
func WithSeekType(seek seek.Type) ClientOption {
	return func(c *Client) error {
		c.seekType = seek
		return nil
	}
}

// WithEventConsumerTimeout is the timeout when sending events to a registered consumer.
// If < 0, if buffer full, unblocks immediately and does not send.
// If 0, if buffer full, will block and guarantee the event will be sent out.
// If > 0, if buffer full, blocks util timeout.
func WithEventConsumerTimeout(value time.Duration) ClientOption {
	return func(c *Client) error {
		c.eventConsumerTimeout = &value
		return nil
	}
}
