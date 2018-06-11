/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package endpoint

import (
	"crypto/x509"

	"github.com/hyperledger/fabric-sdk-go/pkg/common/options"
	"github.com/hyperledger/fabric-sdk-go/pkg/fab/comm"

	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/fab"
)

// EventEndpoint extends a Peer endpoint and provides the
// event URL, which, in the case of Eventhub, is different
// from the peer endpoint
type EventEndpoint struct {
	Certificate *x509.Certificate
	fab.Peer
	EvtURL string
	opts   []options.Opt
}

// EventURL returns the event URL
func (e *EventEndpoint) EventURL() string {
	return e.EvtURL
}

// Opts returns additional options for the event connection
func (e *EventEndpoint) Opts() []options.Opt {
	return e.opts
}

// FromPeerConfig creates a new EventEndpoint from the given config
func FromPeerConfig(config fab.EndpointConfig, peer fab.Peer, peerCfg *fab.PeerConfig) *EventEndpoint {
	opts := comm.OptsFromPeerConfig(peerCfg)
	opts = append(opts, comm.WithConnectTimeout(config.Timeout(fab.EventHubConnection)))

	return &EventEndpoint{
		Peer:   peer,
		EvtURL: peerCfg.EventURL,
		opts:   opts,
	}
}
