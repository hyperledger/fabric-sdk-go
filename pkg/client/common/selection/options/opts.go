/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package options

import (
	"github.com/hyperledger/fabric-sdk-go/pkg/common/errors/retry"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/logging"
	copts "github.com/hyperledger/fabric-sdk-go/pkg/common/options"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/fab"
)

var logger = logging.NewLogger("fabsdk/client")

// PeerFilter filters out unwanted peers
type PeerFilter func(peer fab.Peer) bool

// PrioritySelector determines how likely a peer is to be
// selected over another peer.
// A positive return value means peer1 is selected;
// negative return value means the peer2 is selected;
// zero return value means their priorities are the same
type PrioritySelector func(peer1, peer2 fab.Peer) int

// Params defines the parameters of a selection service request
type Params struct {
	PeerFilter       PeerFilter
	PrioritySelector PrioritySelector
	RetryOpts        retry.Opts
}

// NewParams creates new parameters based on the provided options
func NewParams(opts []copts.Opt) *Params {
	params := &Params{}
	copts.Apply(params, opts)
	return params
}

// WithPeerFilter sets a peer filter which provides per-request filtering of peers
func WithPeerFilter(value PeerFilter) copts.Opt {
	return func(p copts.Params) {
		logger.Debugf("Checking PeerFilter: %#+v", value)
		if setter, ok := p.(peerFilterSetter); ok {
			setter.SetPeerFilter(value)
		}
	}
}

// WithPrioritySelector sets a priority selector function which provides per-request
// prioritization of peers
func WithPrioritySelector(value PrioritySelector) copts.Opt {
	return func(p copts.Params) {
		if setter, ok := p.(prioritySelectorSetter); ok {
			setter.SetPrioritySelector(value)
		}
	}
}

// WithRetryOpts sets the retry options
func WithRetryOpts(value retry.Opts) copts.Opt {
	return func(p copts.Params) {
		if setter, ok := p.(retryOptsSetter); ok {
			setter.SetRetryOpts(value)
		}
	}
}

type peerFilterSetter interface {
	SetPeerFilter(value PeerFilter)
}

// SetPeerFilter sets the peer filter
func (p *Params) SetPeerFilter(value PeerFilter) {
	logger.Debugf("PeerFilter: %#+v", value)
	p.PeerFilter = value
}

type prioritySelectorSetter interface {
	SetPrioritySelector(value PrioritySelector)
}

// SetPrioritySelector sets the priority selector
func (p *Params) SetPrioritySelector(value PrioritySelector) {
	logger.Debugf("PrioritySelector: %#+v", value)
	p.PrioritySelector = value
}

type retryOptsSetter interface {
	SetRetryOpts(value retry.Opts)
}

// SetRetryOpts sets the priority selector
func (p *Params) SetRetryOpts(value retry.Opts) {
	logger.Debugf("RetryOpts: %#+v", value)
	p.RetryOpts = value
}
