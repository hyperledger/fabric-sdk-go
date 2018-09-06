/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package options

import (
	"sort"

	"github.com/hyperledger/fabric-sdk-go/pkg/common/errors/retry"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/logging"
	copts "github.com/hyperledger/fabric-sdk-go/pkg/common/options"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/fab"
)

var logger = logging.NewLogger("fabsdk/client")

// PeerFilter filters out unwanted peers
type PeerFilter func(peer fab.Peer) bool

// PeerSorter sorts the peers
type PeerSorter func(peers []fab.Peer) []fab.Peer

// PrioritySelector determines how likely a peer is to be
// selected over another peer.
// A positive return value means peer1 is selected;
// negative return value means the peer2 is selected;
// zero return value means their priorities are the same
type PrioritySelector func(peer1, peer2 fab.Peer) int

// Params defines the parameters of a selection service request
type Params struct {
	PeerFilter PeerFilter
	PeerSorter PeerSorter
	RetryOpts  retry.Opts
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

// WithPeerSorter sets a peer sorter function which provides per-request
// sorting of peers
func WithPeerSorter(value PeerSorter) copts.Opt {
	return func(p copts.Params) {
		if setter, ok := p.(peerSorterSetter); ok {
			setter.SetPeerSorter(value)
		}
	}
}

// WithPrioritySelector sets a priority selector function which provides per-request
// prioritization of peers
func WithPrioritySelector(value PrioritySelector) copts.Opt {
	return WithPeerSorter(func(peers []fab.Peer) []fab.Peer {
		return sortPeers(peers, value)
	})
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

type peerSorterSetter interface {
	SetPeerSorter(value PeerSorter)
}

// SetPeerSorter sets the priority selector
func (p *Params) SetPeerSorter(value PeerSorter) {
	logger.Debugf("PeerSorter: %#+v", value)
	p.PeerSorter = value
}

type retryOptsSetter interface {
	SetRetryOpts(value retry.Opts)
}

// SetRetryOpts sets the priority selector
func (p *Params) SetRetryOpts(value retry.Opts) {
	logger.Debugf("RetryOpts: %#+v", value)
	p.RetryOpts = value
}

type peers []fab.Peer

func sortPeers(peers peers, ps PrioritySelector) []fab.Peer {
	sort.Sort(&sorter{
		peers:            peers,
		PrioritySelector: ps,
	})
	return peers
}

type sorter struct {
	peers
	PrioritySelector PrioritySelector
}

func (es *sorter) Len() int {
	return len(es.peers)
}

func (es *sorter) Less(i, j int) bool {
	return es.PrioritySelector(es.peers[i], es.peers[j]) > 0
}

func (es *sorter) Swap(i, j int) {
	es.peers[i], es.peers[j] = es.peers[j], es.peers[i]
}
