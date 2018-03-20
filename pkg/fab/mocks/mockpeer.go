/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package mocks

// TODO: Move protos to this library
import (
	reqContext "context"
	"encoding/pem"
	"sync"

	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/fab"
	pb "github.com/hyperledger/fabric-sdk-go/third_party/github.com/hyperledger/fabric/protos/peer"
)

// MockPeer is a mock fabricsdk.Peer.
type MockPeer struct {
	RWLock               *sync.RWMutex
	Error                error
	MockName             string
	MockURL              string
	MockRoles            []string
	MockCert             *pem.Block
	Payload              []byte
	ResponseMessage      string
	MockMSP              string
	Status               int32
	ProcessProposalCalls int
	Endorser             []byte
}

// NewMockPeer creates basic mock peer
func NewMockPeer(name string, url string) *MockPeer {
	mp := &MockPeer{MockName: name, MockMSP: "Org1MSP", MockURL: url, Status: 200, RWLock: &sync.RWMutex{}}
	return mp
}

// Name returns the mock peer's mock name
func (p *MockPeer) Name() string {
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
func (p *MockPeer) ProcessTransactionProposal(ctx reqContext.Context, tp fab.ProcessProposalRequest) (*fab.TransactionProposalResponse, error) {
	if p.RWLock != nil {
		p.RWLock.Lock()
		defer p.RWLock.Unlock()
	}
	p.ProcessProposalCalls++

	return &fab.TransactionProposalResponse{
		Endorser: p.MockURL,
		Status:   p.Status,
		ProposalResponse: &pb.ProposalResponse{Response: &pb.Response{
			Message: p.ResponseMessage, Status: p.Status, Payload: p.Payload},
			Endorsement: &pb.Endorsement{Endorser: p.Endorser, Signature: []byte("signature")}},
	}, p.Error

}
