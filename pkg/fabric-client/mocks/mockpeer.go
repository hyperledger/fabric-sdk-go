/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package mocks

// TODO: Move protos to this library
import (
	"encoding/pem"
	"sync"

	"github.com/golang/protobuf/proto"

	"github.com/hyperledger/fabric-sdk-go/api/apifabclient"
	msp "github.com/hyperledger/fabric-sdk-go/third_party/github.com/hyperledger/fabric/protos/msp"
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
	mp := &MockPeer{MockName: name, MockURL: url, Status: 200, RWLock: &sync.RWMutex{}}
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
func (p *MockPeer) ProcessTransactionProposal(tp apifabclient.TransactionProposal) (apifabclient.TransactionProposalResponse, error) {
	if p.RWLock != nil {
		p.RWLock.Lock()
		defer p.RWLock.Unlock()
	}
	p.ProcessProposalCalls++

	if p.Endorser == nil {
		// We serialize identities by prepending the MSPID and appending the ASN.1 DER content of the cert
		sID := &msp.SerializedIdentity{Mspid: "Org1MSP", IdBytes: []byte(certPem)}
		endorser, err := proto.Marshal(sID)
		if err != nil {
			return apifabclient.TransactionProposalResponse{}, err
		}
		p.Endorser = endorser
	}

	return apifabclient.TransactionProposalResponse{
		Endorser: p.MockURL,
		Proposal: tp,
		Status:   p.Status,
		ProposalResponse: &pb.ProposalResponse{Response: &pb.Response{
			Message: p.ResponseMessage, Status: p.Status, Payload: p.Payload},
			Endorsement: &pb.Endorsement{Endorser: p.Endorser, Signature: []byte("signature")}},
	}, p.Error

}

var certPem = `-----BEGIN CERTIFICATE-----
MIIC5TCCAkagAwIBAgIUMYhiY5MS3jEmQ7Fz4X/e1Dx33J0wCgYIKoZIzj0EAwQw
gYwxCzAJBgNVBAYTAkNBMRAwDgYDVQQIEwdPbnRhcmlvMRAwDgYDVQQHEwdUb3Jv
bnRvMREwDwYDVQQKEwhsaW51eGN0bDEMMAoGA1UECxMDTGFiMTgwNgYDVQQDEy9s
aW51eGN0bCBFQ0MgUm9vdCBDZXJ0aWZpY2F0aW9uIEF1dGhvcml0eSAoTGFiKTAe
Fw0xNzEyMDEyMTEzMDBaFw0xODEyMDEyMTEzMDBaMGMxCzAJBgNVBAYTAkNBMRAw
DgYDVQQIEwdPbnRhcmlvMRAwDgYDVQQHEwdUb3JvbnRvMREwDwYDVQQKEwhsaW51
eGN0bDEMMAoGA1UECxMDTGFiMQ8wDQYDVQQDDAZzZGtfZ28wdjAQBgcqhkjOPQIB
BgUrgQQAIgNiAAT6I1CGNrkchIAEmeJGo53XhDsoJwRiohBv2PotEEGuO6rMyaOu
pulj2VOj+YtgWw4ZtU49g4Nv6rq1QlKwRYyMwwRJSAZHIUMhYZjcDi7YEOZ3Fs1h
xKmIxR+TTR2vf9KjgZAwgY0wDgYDVR0PAQH/BAQDAgWgMBMGA1UdJQQMMAoGCCsG
AQUFBwMCMAwGA1UdEwEB/wQCMAAwHQYDVR0OBBYEFDwS3xhpAWs81OVWvZt+iUNL
z26DMB8GA1UdIwQYMBaAFLRasbknomawJKuQGiyKs/RzTCujMBgGA1UdEQQRMA+C
DWZhYnJpY19zZGtfZ28wCgYIKoZIzj0EAwQDgYwAMIGIAkIAk1MxMogtMtNO0rM8
gw2rrxqbW67ulwmMQzp6EJbm/28T2pIoYWWyIwpzrquypI7BOuf8is5b7Jcgn9oz
7sdMTggCQgF7/8ZFl+wikAAPbciIL1I+LyCXKwXosdFL6KMT6/myYjsGNeeDeMbg
3YkZ9DhdH1tN4U/h+YulG/CkKOtUATtQxg==
-----END CERTIFICATE-----`

var keyPem = `-----BEGIN EC PRIVATE KEY-----
MIGkAgEBBDByldj7VTpqTQESGgJpR9PFW9b6YTTde2WN6/IiBo2nW+CIDmwQgmAl
c/EOc9wmgu+gBwYFK4EEACKhZANiAAT6I1CGNrkchIAEmeJGo53XhDsoJwRiohBv
2PotEEGuO6rMyaOupulj2VOj+YtgWw4ZtU49g4Nv6rq1QlKwRYyMwwRJSAZHIUMh
YZjcDi7YEOZ3Fs1hxKmIxR+TTR2vf9I=
-----END EC PRIVATE KEY-----`
