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
	"github.com/op/go-logging"
)

var logger = logging.MustGetLogger("fabric_sdk_go")

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

//
// TODO: The following placeholders need to be examined - implement or delete.
//

// ConnectEventSource (placeholder)
/**
 * Since practically all Peers are event producers, when constructing a Peer instance,
 * an application can designate it as the event source for the application. Typically
 * only one of the Peers on a Chain needs to be the event source, because all Peers on
 * the Chain produce the same events. This method tells the SDK which Peer(s) to use as
 * the event source for the client application. It is the responsibility of the SDK to
 * manage the connection lifecycle to the Peer’s EventHub. It is the responsibility of
 * the Client Application to understand and inform the selected Peer as to which event
 * types it wants to receive and the call back functions to use.
 * @returns {Future} This gives the app a handle to attach “success” and “error” listeners
 */
func (p *Peer) ConnectEventSource() {
	//to do
}

// IsEventListened (placeholder)
/**
 * A network call that discovers if at least one listener has been connected to the target
 * Peer for a given event. This helps application instance to decide whether it needs to
 * connect to the event source in a crash recovery or multiple instance deployment.
 * @param {string} eventName required
 * @param {Channel} channel optional
 * @result {bool} Whether the said event has been listened on by some application instance on that chain.
 */
func (p *Peer) IsEventListened(event string, channel fab.Channel) (bool, error) {
	//to do
	return false, nil
}

// AddListener (placeholder)
/**
 * For a Peer that is connected to eventSource, the addListener registers an EventCallBack for a
 * set of event types. addListener can be invoked multiple times to support differing EventCallBack
 * functions receiving different types of events.
 *
 * Note that the parameters below are optional in certain languages, like Java, that constructs an
 * instance of a listener interface, and pass in that instance as the parameter.
 * @param {string} eventType : ie. Block, Chaincode, Transaction
 * @param  {object} eventTypeData : Object Specific for event type as necessary, currently needed
 * for “Chaincode” event type, specifying a matching pattern to the event name set in the chaincode(s)
 * being executed on the target Peer, and for “Transaction” event type, specifying the transaction ID
 * @param {struct} eventCallback Client Application class registering for the callback.
 * @returns {string} An ID reference to the event listener.
 */
func (p *Peer) AddListener(eventType string, eventTypeData interface{}, eventCallback interface{}) (string, error) {
	//to do
	return "", nil
}

// RemoveListener (placeholder)
/**
 * Unregisters a listener.
 * @param {string} eventListenerRef Reference returned by SDK for event listener.
 * @return {bool} Success / Failure status
 */
func (p *Peer) RemoveListener(eventListenerRef string) (bool, error) {
	return false, nil
	//to do
}

// PeersToTxnProcessors converts a slice of Peers to a slice of TxnProposalProcessors
func PeersToTxnProcessors(peers []fab.Peer) []apitxn.ProposalProcessor {
	tpp := make([]apitxn.ProposalProcessor, len(peers))

	for i := range peers {
		tpp[i] = peers[i]
	}
	return tpp
}
