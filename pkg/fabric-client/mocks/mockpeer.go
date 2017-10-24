/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package mocks

// TODO: Move protos to this library
import (
	"encoding/pem"

	"github.com/hyperledger/fabric-sdk-go/api/apitxn"
	pb "github.com/hyperledger/fabric-sdk-go/third_party/github.com/hyperledger/fabric/protos/peer"
)

// MockPeer is a mock fabricsdk.Peer.
type MockPeer struct {
	MockName  string
	MockURL   string
	MockRoles []string
	MockCert  *pem.Block
	Payload   []byte
	MockMSP   string
}

// NewMockPeer creates basic mock peer
func NewMockPeer(name string, url string) *MockPeer {
	mp := &MockPeer{MockName: name, MockURL: url}
	return mp
}

// Name returns the mock peer's mock name
func (p MockPeer) Name() string {
	return p.MockName
}

// SetName sets the mock peer's mock name
func (p *MockPeer) SetName(name string) {
	p.MockName = name
}

// MSPID gets the Peer mspID.
func (p *MockPeer) MSPID() string {
	return p.MockMSP
}

// SetMSPID sets the Peer mspID.
func (p *MockPeer) SetMSPID(mspID string) {
	p.MockMSP = mspID
}

// Roles returns the mock peer's mock roles
func (p *MockPeer) Roles() []string {
	return p.MockRoles
}

// SetRoles sets the mock peer's mock roles
func (p *MockPeer) SetRoles(roles []string) {
	p.MockRoles = roles
}

// EnrollmentCertificate returns the mock peer's mock enrollment certificate
func (p *MockPeer) EnrollmentCertificate() *pem.Block {
	return p.MockCert
}

// SetEnrollmentCertificate sets the mock peer's mock enrollment certificate
func (p *MockPeer) SetEnrollmentCertificate(pem *pem.Block) {
	p.MockCert = pem
}

// URL returns the mock peer's mock URL
func (p *MockPeer) URL() string {
	return p.MockURL
}

// ProcessTransactionProposal does not send anything anywhere but returns an empty mock ProposalResponse
func (p *MockPeer) ProcessTransactionProposal(tp apitxn.TransactionProposal) (apitxn.TransactionProposalResult, error) {
	return apitxn.TransactionProposalResult{
		Endorser:         p.MockURL,
		Proposal:         tp,
		ProposalResponse: &pb.ProposalResponse{Response: &pb.Response{Message: "success", Status: 99, Payload: p.Payload}},
	}, nil

}
