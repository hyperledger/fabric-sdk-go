/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package peer

import (
	"encoding/pem"

	"github.com/hyperledger/fabric-sdk-go/api/apiconfig"
	fab "github.com/hyperledger/fabric-sdk-go/api/apifabclient"
	"github.com/hyperledger/fabric-sdk-go/api/apitxn"
	"github.com/hyperledger/fabric-sdk-go/pkg/logging"
)

var logger = logging.NewLogger("fabric_sdk_go")

const (
	connBlocking = true
)

// Peer represents a node in the target blockchain network to which
// HFC sends endorsement proposals, transaction ordering or query requests.
type Peer struct {
	processor             apitxn.ProposalProcessor
	name                  string
	mspID                 string
	roles                 []string
	enrollmentCertificate *pem.Block
	url                   string
}

// NewPeerTLSFromCert constructs a Peer given its endpoint configuration settings.
// url is the URL with format of "host:port".
// certificate is ...
// serverNameOverride is passed to NewClientTLSFromCert in grpc/credentials.
func NewPeerTLSFromCert(url string, certificate string, serverHostOverride string, config apiconfig.Config) (*Peer, error) {
	// TODO: config is declaring TLS but cert & serverHostOverride is being passed-in...
	conn, err := newPeerEndorser(url, certificate, serverHostOverride, connBlocking, config)
	if err != nil {
		return nil, err
	}

	return NewPeerFromProcessor(url, &conn, config)
}

// NewPeerFromConfig constructs a Peer from given peer configuration and global configuration setting.
func NewPeerFromConfig(peerCfg *apiconfig.NetworkPeer, config apiconfig.Config) (*Peer, error) {

	serverHostOverride := ""
	if str, ok := peerCfg.GRPCOptions["ssl-target-name-override"].(string); ok {
		serverHostOverride = str
	}

	conn, err := newPeerEndorser(peerCfg.URL, peerCfg.TLSCACerts.Path, serverHostOverride, connBlocking, config)
	if err != nil {
		return nil, err
	}

	newPeer, err := NewPeerFromProcessor(peerCfg.URL, &conn, config)
	if err != nil {
		return nil, err
	}

	// TODO: Remove upon making peer interface immutable
	newPeer.SetMSPID(peerCfg.MspID)

	return newPeer, nil
}

// NewPeer constructs a Peer given its endpoint configuration settings.
// url is the URL with format of "host:port".
func NewPeer(url string, config apiconfig.Config) (*Peer, error) {
	conn, err := newPeerEndorser(url, "", "", connBlocking, config)
	if err != nil {
		return nil, err
	}

	return NewPeerFromProcessor(url, &conn, config)
}

// NewPeerFromProcessor constructs a Peer with a ProposalProcessor to simulate transactions.
func NewPeerFromProcessor(url string, processor apitxn.ProposalProcessor, config apiconfig.Config) (*Peer, error) {
	return &Peer{url: url, processor: processor}, nil
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
