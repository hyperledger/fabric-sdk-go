/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package peer

import (
	"context"
	"encoding/pem"
	"fmt"

	"crypto/x509"

	"github.com/spf13/cast"
	"google.golang.org/grpc"
	"google.golang.org/grpc/keepalive"

	"github.com/hyperledger/fabric-sdk-go/pkg/context/api/core"
	"github.com/hyperledger/fabric-sdk-go/pkg/context/api/fab"
	"github.com/hyperledger/fabric-sdk-go/pkg/core/config/urlutil"
	"github.com/hyperledger/fabric-sdk-go/pkg/errors/status"
	"github.com/hyperledger/fabric-sdk-go/pkg/logging"
)

var logger = logging.NewLogger("fabric_sdk_go")

const (
	connBlocking = false
)

type connProvider interface {
	DialContext(ctx context.Context, target string, opts ...grpc.DialOption) (conn *grpc.ClientConn, err error)
	ReleaseConn(conn *grpc.ClientConn)
}

// Peer represents a node in the target blockchain network to which
// HFC sends endorsement proposals, transaction ordering or query requests.
type Peer struct {
	config                core.Config
	certificate           *x509.Certificate
	serverName            string
	processor             fab.ProposalProcessor
	name                  string
	mspID                 string
	roles                 []string
	enrollmentCertificate *pem.Block
	url                   string
	kap                   keepalive.ClientParameters
	failFast              bool
	inSecure              bool
	connector             connProvider
}

// Option describes a functional parameter for the New constructor
type Option func(*Peer) error

// New Returns a new Peer instance
func New(config core.Config, opts ...Option) (*Peer, error) {
	peer := &Peer{
		config:    config,
		connector: &defConnector{},
	}
	var err error

	for _, opt := range opts {
		err := opt(peer)

		if err != nil {
			return nil, err
		}
	}

	if peer.processor == nil {
		// TODO: config is declaring TLS but cert & serverHostOverride is being passed-in...
		endorseRequest := peerEndorserRequest{
			target:             peer.url,
			certificate:        peer.certificate,
			serverHostOverride: peer.serverName,
			dialBlocking:       connBlocking,
			config:             peer.config,
			kap:                peer.kap,
			failFast:           peer.failFast,
			allowInsecure:      peer.inSecure,
			connector:          peer.connector,
		}
		peer.processor, err = newPeerEndorser(&endorseRequest)

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

// WithInsecure is a functional option for the peer.New constructor that configures the peer's grpc insecure option
func WithInsecure() Option {
	return func(p *Peer) error {
		p.inSecure = true

		return nil
	}
}

// FromPeerConfig is a functional option for the peer.New constructor that configures a new peer
// from a apiconfig.NetworkPeer struct
func FromPeerConfig(peerCfg *core.NetworkPeer) Option {
	return func(p *Peer) error {

		p.url = peerCfg.URL
		p.serverName = getServerNameOverride(peerCfg)
		p.inSecure = isInsecureConnectionAllowed(peerCfg)

		var err error
		p.certificate, err = peerCfg.TLSCACerts.TLSCert()

		if err != nil {
			//Ignore empty cert errors,
			errStatus, ok := err.(*status.Status)
			if !ok || errStatus.Code != status.EmptyCert.ToInt32() {
				return err
			}
		}

		// TODO: Remove upon making peer interface immutable
		p.mspID = peerCfg.MspID
		p.kap = getKeepAliveOptions(peerCfg)
		p.failFast = getFailFast(peerCfg)
		return nil
	}
}

func getServerNameOverride(peerCfg *core.NetworkPeer) string {
	serverHostOverride := ""
	if str, ok := peerCfg.GRPCOptions["ssl-target-name-override"].(string); ok {
		serverHostOverride = str
	}

	return serverHostOverride
}

func getFailFast(peerCfg *core.NetworkPeer) bool {
	var failFast = true
	if ff, ok := peerCfg.GRPCOptions["fail-fast"].(bool); ok {
		failFast = cast.ToBool(ff)
	}

	return failFast
}

func getKeepAliveOptions(peerCfg *core.NetworkPeer) keepalive.ClientParameters {

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
func isInsecureConnectionAllowed(peerCfg *core.NetworkPeer) bool {
	//allowInsecure used only when protocol is missing from URL
	allowInsecure := !urlutil.HasProtocol(peerCfg.URL)
	boolVal, ok := peerCfg.GRPCOptions["allow-insecure"].(bool)
	if ok {
		return allowInsecure && boolVal
	}
	return false
}

// WithPeerProcessor is a functional option for the peer.New constructor that configures the peer's proposal processor
func WithPeerProcessor(processor fab.ProposalProcessor) Option {
	return func(p *Peer) error {
		p.processor = processor

		return nil
	}
}

// WithConnProvider allows a custom GRPC connection provider to be used.
func WithConnProvider(provider connProvider) Option {
	return func(p *Peer) error {
		p.connector = provider

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
func (p *Peer) ProcessTransactionProposal(proposal fab.ProcessProposalRequest) (*fab.TransactionProposalResponse, error) {
	return p.processor.ProcessTransactionProposal(proposal)
}

func (p *Peer) String() string {
	return fmt.Sprintf("%s (%s)", p.name, p.url)
}

// PeersToTxnProcessors converts a slice of Peers to a slice of TxnProposalProcessors
func PeersToTxnProcessors(peers []fab.Peer) []fab.ProposalProcessor {
	tpp := make([]fab.ProposalProcessor, len(peers))

	for i := range peers {
		tpp[i] = peers[i]
	}
	return tpp
}

type defConnector struct{}

func (*defConnector) DialContext(ctx context.Context, target string, opts ...grpc.DialOption) (conn *grpc.ClientConn, err error) {
	opts = append(opts, grpc.WithBlock())
	return grpc.DialContext(ctx, target, opts...)
}

func (*defConnector) ReleaseConn(conn *grpc.ClientConn) {
	conn.Close()
}
