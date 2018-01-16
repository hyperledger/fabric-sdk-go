/*
Copyright SecureKey Technologies Inc., Unchain B.V. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package peer

import (
	"crypto/x509"

	"github.com/hyperledger/fabric-sdk-go/api/apiconfig"
	"github.com/hyperledger/fabric-sdk-go/api/apitxn"
	"github.com/hyperledger/fabric-sdk-go/pkg/config/urlutil"
)

// NewPeerTLSFromCert constructs a Peer given its endpoint configuration settings.
// url is the URL with format of "host:port".
// certificate is ...
// serverNameOverride is passed to NewClientTLSFromCert in grpc/credentials.
// Deprecated: use peer.New() instead
func NewPeerTLSFromCert(url string, certPath string, serverHostOverride string, config apiconfig.Config) (*Peer, error) {
	var certificate *x509.Certificate
	var err error

	if urlutil.IsTLSEnabled(url) {
		certConfig := apiconfig.TLSConfig{Path: certPath}
		certificate, err = certConfig.TLSCert()

		if err != nil {
			return nil, err
		}
	}

	// TODO: config is declaring TLS but cert & serverHostOverride is being passed-in...
	conn, err := newPeerEndorser(url, certificate, serverHostOverride, connBlocking, config)
	if err != nil {
		return nil, err
	}

	return NewPeerFromProcessor(url, conn, config)
}

// NewPeerFromConfig constructs a Peer from given peer configuration and global configuration setting.
// Deprecated: use peer.New() instead
func NewPeerFromConfig(peerCfg *apiconfig.NetworkPeer, config apiconfig.Config) (*Peer, error) {

	serverHostOverride := ""
	if str, ok := peerCfg.GRPCOptions["ssl-target-name-override"].(string); ok {
		serverHostOverride = str
	}
	var certificate *x509.Certificate
	var err error

	if urlutil.IsTLSEnabled(peerCfg.URL) {
		certificate, err = peerCfg.TLSCACerts.TLSCert()

		if err != nil {
			return nil, err
		}
	}
	conn, err := newPeerEndorser(peerCfg.URL, certificate, serverHostOverride, connBlocking, config)
	if err != nil {
		return nil, err
	}

	newPeer, err := NewPeerFromProcessor(peerCfg.URL, conn, config)
	if err != nil {
		return nil, err
	}

	// TODO: Remove upon making peer interface immutable
	newPeer.SetMSPID(peerCfg.MspID)

	return newPeer, nil
}

// NewPeer constructs a Peer given its endpoint configuration settings.
// url is the URL with format of "host:port".
// Deprecated: use peer.New() instead
func NewPeer(url string, config apiconfig.Config) (*Peer, error) {
	conn, err := newPeerEndorser(url, nil, "", connBlocking, config)
	if err != nil {
		return nil, err
	}

	return NewPeerFromProcessor(url, conn, config)
}

// NewPeerFromProcessor constructs a Peer with a ProposalProcessor to simulate transactions.
// Deprecated: use peer.New() instead
func NewPeerFromProcessor(url string, processor apitxn.ProposalProcessor, config apiconfig.Config) (*Peer, error) {
	return &Peer{url: url, processor: processor}, nil
}
