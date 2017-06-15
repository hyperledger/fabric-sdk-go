/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package fabricclient

// TODO: Move protos to this library
import (
	"encoding/pem"
	"errors"

	pb "github.com/hyperledger/fabric/protos/peer"
)

// mockPeer is a mock fabricsdk.Peer.
type mockPeer struct {
	MockName  string
	MockURL   string
	MockRoles []string
	MockCert  *pem.Block
}

// ConnectEventSource does not connect anywhere
func (p *mockPeer) ConnectEventSource() {
	// done.
}

// IsEventListened always returns true
func (p *mockPeer) IsEventListened(event string, chain Chain) (bool, error) {
	return true, nil
}

// AddListener is not implemented
func (p *mockPeer) AddListener(eventType string, eventTypeData interface{}, eventCallback interface{}) (string, error) {
	return "", errors.New("Not implemented")
}

// RemoveListener is not implemented
func (p *mockPeer) RemoveListener(eventListenerRef string) (bool, error) {
	return false, errors.New("Not implemented")
}

// GetName returns the mock peer's mock name
func (p mockPeer) GetName() string {
	return p.MockName
}

// SetName sets the mock peer's mock name
func (p *mockPeer) SetName(name string) {
	p.MockName = name
}

// GetRoles returns the mock peer's mock roles
func (p *mockPeer) GetRoles() []string {
	return p.MockRoles
}

// SetRoles sets the mock peer's mock roles
func (p *mockPeer) SetRoles(roles []string) {
	p.MockRoles = roles
}

// GetEnrollmentCertificate returns the mock peer's mock enrollment certificate
func (p *mockPeer) GetEnrollmentCertificate() *pem.Block {
	return p.MockCert
}

// SetEnrollmentCertificate sets the mock peer's mock enrollment certificate
func (p *mockPeer) SetEnrollmentCertificate(pem *pem.Block) {
	p.MockCert = pem
}

// GetURL returns the mock peer's mock URL
func (p *mockPeer) GetURL() string {
	return p.MockURL
}

// SendProposal does not send anything anywhere but returns an empty mock ProposalResponse
func (p *mockPeer) SendProposal(tp *TransactionProposal) (*TransactionProposalResponse, error) {
	return &TransactionProposalResponse{
		Endorser:         p.MockURL,
		Proposal:         tp,
		ProposalResponse: &pb.ProposalResponse{},
	}, nil
}
