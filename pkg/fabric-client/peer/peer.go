/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package peer

import (
	"encoding/pem"
	"time"

	api "github.com/hyperledger/fabric-sdk-go/api"

	pb "github.com/hyperledger/fabric/protos/peer"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

type peer struct {
	url                   string
	grpcDialOption        []grpc.DialOption
	name                  string
	roles                 []string
	enrollmentCertificate *pem.Block
}

// ConnectEventSource ...
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
func (p *peer) ConnectEventSource() {
	//to do
}

// IsEventListened ...
/**
 * A network call that discovers if at least one listener has been connected to the target
 * Peer for a given event. This helps application instance to decide whether it needs to
 * connect to the event source in a crash recovery or multiple instance deployment.
 * @param {string} eventName required
 * @param {Channel} channel optional
 * @result {bool} Whether the said event has been listened on by some application instance on that chain.
 */
func (p *peer) IsEventListened(event string, channel api.Channel) (bool, error) {
	//to do
	return false, nil
}

// AddListener ...
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
func (p *peer) AddListener(eventType string, eventTypeData interface{}, eventCallback interface{}) (string, error) {
	//to do
	return "", nil
}

// RemoveListener ...
/**
 * Unregisters a listener.
 * @param {string} eventListenerRef Reference returned by SDK for event listener.
 * @return {bool} Success / Failure status
 */
func (p *peer) RemoveListener(eventListenerRef string) (bool, error) {
	return false, nil
	//to do
}

// GetName ...
/**
 * Get the Peer name. Required property for the instance objects.
 * @returns {string} The name of the Peer
 */
func (p *peer) GetName() string {
	return p.name
}

// SetName ...
/**
 * Set the Peer name / id.
 * @param {string} name
 */
func (p *peer) SetName(name string) {
	p.name = name
}

// GetRoles ...
/**
 * Get the user’s roles the Peer participates in. It’s an array of possible values
 * in “client”, and “auditor”. The member service defines two more roles reserved
 * for peer membership: “peer” and “validator”, which are not exposed to the applications.
 * @returns {[]string} The roles for this user.
 */
func (p *peer) GetRoles() []string {
	return p.roles
}

// SetRoles ...
/**
 * Set the user’s roles the Peer participates in. See getRoles() for legitimate values.
 * @param {[]string} roles The list of roles for the user.
 */
func (p *peer) SetRoles(roles []string) {
	p.roles = roles
}

// GetEnrollmentCertificate ...
/**
 * Returns the Peer's enrollment certificate.
 * @returns {pem.Block} Certificate in PEM format signed by the trusted CA
 */
func (p *peer) GetEnrollmentCertificate() *pem.Block {
	return p.enrollmentCertificate
}

// SetEnrollmentCertificate ...
/**
 * Set the Peer’s enrollment certificate.
 * @param {pem.Block} enrollment Certificate in PEM format signed by the trusted CA
 */
func (p *peer) SetEnrollmentCertificate(pem *pem.Block) {
	p.enrollmentCertificate = pem
}

// GetURL ...
/**
 * Get the Peer url. Required property for the instance objects.
 * @returns {string} The address of the Peer
 */
func (p *peer) GetURL() string {
	return p.url
}

// SendProposal ...
/**
 * Send  the created proposal to peer for endorsement.
 */
func (p *peer) SendProposal(proposal *api.TransactionProposal) (*api.TransactionProposalResponse, error) {
	conn, err := grpc.Dial(p.url, p.grpcDialOption...)
	if err != nil {
		return nil, err
	}
	defer conn.Close()
	endorserClient := pb.NewEndorserClient(conn)
	proposalResponse, err := endorserClient.ProcessProposal(context.Background(), proposal.SignedProposal)
	if err != nil {
		return nil, err
	}
	return &api.TransactionProposalResponse{
		Proposal:         proposal,
		ProposalResponse: proposalResponse,
		Endorser:         p.url,
		Status:           proposalResponse.GetResponse().Status,
	}, nil
}

// NewPeer ...
/**
 * Constructs a Peer given its endpoint configuration settings.
 *
 * @param {string} url The URL with format of "host:port".
 */
func NewPeer(url string, certificate string, serverHostOverride string, config api.Config) (api.Peer, error) {
	var opts []grpc.DialOption
	opts = append(opts, grpc.WithTimeout(time.Second*3))
	if config.IsTLSEnabled() {
		tlsCaCertPool, err := config.GetTLSCACertPool(certificate)
		if err != nil {
			return nil, err
		}
		creds := credentials.NewClientTLSFromCert(tlsCaCertPool, serverHostOverride)
		opts = append(opts, grpc.WithTransportCredentials(creds))
	} else {
		opts = append(opts, grpc.WithInsecure())
	}
	return &peer{url: url, grpcDialOption: opts, name: "", roles: nil}, nil

}
