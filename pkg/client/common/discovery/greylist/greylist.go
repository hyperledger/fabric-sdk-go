/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package greylist

import (
	"sync"
	"time"

	"github.com/hyperledger/fabric-sdk-go/pkg/common/errors/status"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/logging"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/fab"
	"github.com/hyperledger/fabric-sdk-go/pkg/core/config/endpoint"
)

var logger = logging.NewLogger("fabsdk/client")

// Filter is a discovery filter that greylists certain peers that are
// known to be down for the configured amount of time
type Filter struct {
	// greylistURLs contains a map of peer URLs as keys and timestamps as values
	// peers are expired from the greylist based on these timestamps
	greylistURLs   sync.Map
	expiryInterval time.Duration
}

// New creates a new greylist filter with the given expiry interval
func New(expire time.Duration) *Filter {
	return &Filter{expiryInterval: expire}
}

// Accept returns whether or not to Accept a peer as a canditate for endorsement
func (b *Filter) Accept(peer fab.Peer) bool {
	peerAddress := endpoint.ToAddress(peer.URL())
	value, ok := b.greylistURLs.Load(peerAddress)
	if ok {
		timeAdded, ok := value.(time.Time)
		if ok && timeAdded.Add(b.expiryInterval).After(time.Now()) {
			logger.Infof("Rejecting peer %s", peer.URL())
			return false
		}
		b.greylistURLs.Delete(peerAddress)
	}

	return true
}

// Greylist the given peer URL
func (b *Filter) Greylist(err error) {
	s, ok := status.FromError(err)
	if !ok {
		return
	}
	if ok, peerURL := required(s); ok && peerURL != "" {
		logger.Infof("Greylisting peer %s", peerURL)
		b.greylistURLs.Store(peerURL, time.Now())
	}
}

// required decides whether the given status error warrants a greylist
// on the peer causing the error
func required(s *status.Status) (bool, string) {
	if s.Group == status.EndorserClientStatus && s.Code == status.ConnectionFailed.ToInt32() {
		return true, peerURLFromConnectionFailedStatus(s.Details)
	}
	return false, ""
}

// peerURLFromConnectionFailedStatus extracts the peer url from the status error
// details
func peerURLFromConnectionFailedStatus(details []interface{}) string {
	if len(details) != 0 {
		url, ok := details[0].(string)
		if ok {
			return endpoint.ToAddress(url)
		}
	}
	return ""
}
