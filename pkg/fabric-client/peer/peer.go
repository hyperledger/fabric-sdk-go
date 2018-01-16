/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package peer

import (
	"encoding/pem"

	"crypto/x509"

	"github.com/hyperledger/fabric-sdk-go/api/apiconfig"
	fab "github.com/hyperledger/fabric-sdk-go/api/apifabclient"
	"github.com/hyperledger/fabric-sdk-go/api/apitxn"
	"github.com/hyperledger/fabric-sdk-go/pkg/config/urlutil"
	"github.com/hyperledger/fabric-sdk-go/pkg/logging"
)

var logger = logging.NewLogger("fabric_sdk_go")

const (
	connBlocking = true
)

// Peer represents a node in the target blockchain network to which
// HFC sends endorsement proposals, transaction ordering or query requests.
type Peer struct {
	config                apiconfig.Config
	certificate           *x509.Certificate
	serverName            string
	processor             apitxn.ProposalProcessor
	name                  string
	mspID                 string
	roles                 []string
	enrollmentCertificate *pem.Block
	url                   string
}

// Option describes a functional parameter for the New constructor
type Option func(*Peer) error

// New Returns a new Peer instance
func New(config apiconfig.Config, opts ...Option) (*Peer, error) {
	peer := &Peer{config: config}
	var err error

	for _, opt := range opts {
		err := opt(peer)

		if err != nil {
			return nil, err
		}
	}

	if peer.processor == nil {
		// TODO: config is declaring TLS but cert & serverHostOverride is being passed-in...
		peer.processor, err = newPeerEndorser(peer.url, peer.certificate, peer.serverName, connBlocking, peer.config)

		if err != nil {
			return nil, err
		}
	}

	return peer, nil
}

// WithURL is a functional option for the peer.New constructor that configures the peer's URL
func WithURL(url string) Option {
	return func(p *Peer) error {
		p.url = url

		return nil
	}
}

// WithTLSCert is a functional option for the peer.New constructor that configures the peer's TLS certificate
func WithTLSCert(certificate *x509.Certificate) Option {
	return func(p *Peer) error {
		p.certificate = certificate

		return nil
	}
}

// WithServerName is a functional option for the peer.New constructor that configures the peer's server name
func WithServerName(serverName string) Option {
	return func(p *Peer) error {
		p.serverName = serverName

		return nil
	}
}

// FromPeerConfig is a functional option for the peer.New constructor that configures a new peer
// from a apiconfig.NetworkPeer struct
func FromPeerConfig(peerCfg *apiconfig.NetworkPeer) Option {
	return func(p *Peer) error {
		p.url = peerCfg.URL
		p.serverName = getServerNameOverride(peerCfg)

		var err error

		if urlutil.IsTLSEnabled(peerCfg.URL) {
			p.certificate, err = peerCfg.TLSCACerts.TLSCert()

			if err != nil {
				return err
			}
		}

		// TODO: Remove upon making peer interface immutable
		p.mspID = peerCfg.MspID

		return nil
	}
}

func getServerNameOverride(peerCfg *apiconfig.NetworkPeer) string {
	serverHostOverride := ""
	if str, ok := peerCfg.GRPCOptions["ssl-target-name-override"].(string); ok {
		serverHostOverride = str
	}

	return serverHostOverride
}

// WithPeerProcessor is a functional option for the peer.New constructor that configures the peer's proposal processor
func WithPeerProcessor(processor apitxn.ProposalProcessor) Option {
	return func(p *Peer) error {
		p.processor = processor

		return nil
	}
}

// Name gets the Peer name.
func (p *Peer) Name() string {
	return p.name
}

// SetName sets the Peer name / id.
func (p *Peer) SetName(name string) {
	p.name = name
}

// MSPID gets the Peer mspID.
func (p *Peer) MSPID() string {
	return p.mspID
}

// SetMSPID sets the Peer mspID.
func (p *Peer) SetMSPID(mspID string) {
	p.mspID = mspID
}

// Roles gets the user’s roles the Peer participates in. It’s an array of possible values
// in “client”, and “auditor”. The member service defines two more roles reserved
// for peer membership: “peer” and “validator”, which are not exposed to the applications.
// It returns the roles for this user.
func (p *Peer) Roles() []string {
	return p.roles
}

// SetRoles sets the user’s roles the Peer participates in. See getRoles() for legitimate values.
// roles is the list of roles for the user.
func (p *Peer) SetRoles(roles []string) {
	p.roles = roles
}

// EnrollmentCertificate returns the Peer's enrollment certificate.
// It returns the certificate in PEM format signed by the trusted CA.
func (p *Peer) EnrollmentCertificate() *pem.Block {
	return p.enrollmentCertificate
}

// SetEnrollmentCertificate set the Peer’s enrollment certificate.
// pem is the enrollment Certificate in PEM format signed by the trusted CA.
func (p *Peer) SetEnrollmentCertificate(pem *pem.Block) {
	p.enrollmentCertificate = pem
}

// URL gets the Peer URL. Required property for the instance objects.
// It returns the address of the Peer.
func (p *Peer) URL() string {
	return p.url
}

// ProcessTransactionProposal sends the created proposal to peer for endorsement.
func (p *Peer) ProcessTransactionProposal(proposal apitxn.TransactionProposal) (apitxn.TransactionProposalResult, error) {
	return p.processor.ProcessTransactionProposal(proposal)
}

// PeersToTxnProcessors converts a slice of Peers to a slice of TxnProposalProcessors
func PeersToTxnProcessors(peers []fab.Peer) []apitxn.ProposalProcessor {
	tpp := make([]apitxn.ProposalProcessor, len(peers))

	for i := range peers {
		tpp[i] = peers[i]
	}
	return tpp
}
