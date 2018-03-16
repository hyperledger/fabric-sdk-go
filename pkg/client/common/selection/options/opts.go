/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package options

import (
	"github.com/hyperledger/fabric-sdk-go/pkg/common/logging"
	copts "github.com/hyperledger/fabric-sdk-go/pkg/common/options"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/fab"
)

var logger = logging.NewLogger("fabsdk/client")

// PeerFilter filters out unwanted peers
type PeerFilter func(peer fab.Peer) bool

// Params defines the parameters of a selection service request
type Params struct {
	PeerFilter PeerFilter
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
		if setter, ok := p.(peerFilterSetter); ok {
			setter.SetPeerFilter(value)
		}
	}
}

type peerFilterSetter interface {
	SetPeerFilter(value PeerFilter)
}

// SetPeerFilter sets the peer filter
func (p *Params) SetPeerFilter(value PeerFilter) {
	logger.Debugf("PeerFilter: %#v", value)
	p.PeerFilter = value
}
