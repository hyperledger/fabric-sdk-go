/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package endpoint

import (
	"crypto/x509"
	"time"

	"github.com/hyperledger/fabric-sdk-go/pkg/context/api/core"
	"github.com/hyperledger/fabric-sdk-go/pkg/context/api/fab"
	"github.com/hyperledger/fabric-sdk-go/pkg/errors/status"
	"github.com/hyperledger/fabric-sdk-go/pkg/fab/peer"
	"github.com/spf13/cast"
	"google.golang.org/grpc/keepalive"
)

// EventEndpoint extends a Peer endpoint and provides the
// event URL, which, in the case of Eventhub, is different
// from the peer endpoint
type EventEndpoint struct {
	fab.Peer
	EvtURL          string
	HostOverride    string
	Certificate     *x509.Certificate
	KeepAliveParams keepalive.ClientParameters
	FailFast        bool
	ConnectTimeout  time.Duration
}

// EventURL returns the event URL
func (e *EventEndpoint) EventURL() string {
	return e.EvtURL
}

// FromPeerConfig creates a new EventEndpoint from the given config
func FromPeerConfig(config core.Config, peerCfg core.NetworkPeer) (*EventEndpoint, error) {
	p, err := peer.New(config, peer.FromPeerConfig(&peerCfg))
	if err != nil {
		return nil, err
	}

	certificate, err := peerCfg.TLSCACerts.TLSCert()
	if err != nil {
		//Ignore empty cert errors,
		errStatus, ok := err.(*status.Status)
		if !ok || errStatus.Code != status.EmptyCert.ToInt32() {
			return nil, err
		}
	}

	return &EventEndpoint{
		Peer:            p,
		EvtURL:          peerCfg.EventURL,
		HostOverride:    getServerNameOverride(peerCfg),
		Certificate:     certificate,
		KeepAliveParams: getKeepAliveOptions(peerCfg),
		FailFast:        getFailFast(peerCfg),
		ConnectTimeout:  config.TimeoutOrDefault(core.EventHubConnection),
	}, nil
}

func getServerNameOverride(peerCfg core.NetworkPeer) string {
	if str, ok := peerCfg.GRPCOptions["ssl-target-name-override"].(string); ok {
		return str
	}
	return ""
}

func getFailFast(peerCfg core.NetworkPeer) bool {
	if ff, ok := peerCfg.GRPCOptions["fail-fast"].(bool); ok {
		return cast.ToBool(ff)
	}
	return false
}

func getKeepAliveOptions(peerCfg core.NetworkPeer) keepalive.ClientParameters {
	var kap keepalive.ClientParameters
	if kaTime, ok := peerCfg.GRPCOptions["keep-alive-time"]; ok {
		kap.Time = cast.ToDuration(kaTime)
	}
	if kaTimeout, ok := peerCfg.GRPCOptions["keep-alive-timeout"]; ok {
		kap.Timeout = cast.ToDuration(kaTimeout)
	}
	if kaPermit, ok := peerCfg.GRPCOptions["keep-alive-permit"]; ok {
		kap.PermitWithoutStream = cast.ToBool(kaPermit)
	}
	return kap
}
