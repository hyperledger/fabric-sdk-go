/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package channel

import (
	"crypto/ecdsa"
	"crypto/rand"
	"crypto/sha256"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"net"
	"sync"

	"github.com/golang/protobuf/proto"
	cb "github.com/hyperledger/fabric-protos-go/common"
	"github.com/hyperledger/fabric-protos-go/discovery"
	"github.com/hyperledger/fabric-protos-go/ledger/rwset/kvrwset"
	"github.com/hyperledger/fabric-protos-go/msp"
	pb "github.com/hyperledger/fabric-protos-go/peer"
	"github.com/hyperledger/fabric-sdk-go/internal/github.com/hyperledger/fabric/bccsp/utils"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/fab"
	discmocks "github.com/hyperledger/fabric-sdk-go/pkg/fab/discovery/mocks"
	eventmocks "github.com/hyperledger/fabric-sdk-go/pkg/fab/events/mocks"
	"github.com/hyperledger/fabric-sdk-go/pkg/util/test"
	"github.com/hyperledger/fabric-sdk-go/third_party/github.com/hyperledger/fabric/core/ledger/kvledger/txmgmt/rwsetutil"
	"github.com/pkg/errors"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

// MockEndorserServer mock endorser server to process endorsement proposals
type MockEndorserServer struct {
	ProposalError       error
	Creds               credentials.TransportCredentials
	mockPeer            *MockPeer
	srv                 *grpc.Server
	wg                  sync.WaitGroup
	AddkvWrite          bool
	DeliveriesListener  chan *cb.Block
	FilteredDelListener chan *pb.FilteredBlock
}

// ProcessProposal mock implementation that returns success (through mockPeer) if error is not set
// error if it is
func (m *MockEndorserServer) ProcessProposal(context context.Context,
	proposal *pb.SignedProposal) (*pb.ProposalResponse, error) {
	fcn, err := m.getFuncNameFromProposal(proposal)
	if err != nil {
		return nil, errors.Wrap(err, "Error getting function name from mockPeer")
	}
	if m.ProposalError == nil {

		tp, err := m.GetMockPeer().ProcessTransactionProposal(fab.TransactionProposal{}, fcn)
		if err != nil {
			return &pb.ProposalResponse{Response: &pb.Response{
				Status:  500,
				Message: err.Error(),
			}}, err
		}
		return tp.ProposalResponse, nil

	}

	return &pb.ProposalResponse{Response: &pb.Response{
		Status:  500,
		Message: m.ProposalError.Error(),
	}}, m.ProposalError
}

func (m *MockEndorserServer) getFuncNameFromProposal(proposal *pb.SignedProposal) ([]byte, error) {
	pr := &pb.Proposal{}
	err := proto.Unmarshal(proposal.GetProposalBytes(), pr)
	if err != nil {
		return nil, err
	}
	cpp := &pb.ChaincodeProposalPayload{}
	err = proto.Unmarshal(pr.Payload, cpp)
	if err != nil {
		return nil, err
	}

	cic := &pb.ChaincodeInvocationSpec{}
	err = proto.Unmarshal(cpp.Input, cic)
	if err != nil {
		return nil, err
	}
	return cic.ChaincodeSpec.Input.Args[0], nil
}

// Start the mock endorser server
func (m *MockEndorserServer) Start(address string, filteredChannel chan *pb.FilteredBlock) string {
	if m.srv != nil {
		panic("MockEndorserServer already started")
	}

	// pass in TLS creds if present
	if m.Creds != nil {
		m.srv = grpc.NewServer(grpc.Creds(m.Creds))
	} else {
		m.srv = grpc.NewServer()
	}

	// set the filtered block channel used by the mock delivery server
	m.FilteredDelListener = filteredChannel
	m.registerDiscoveryAndDeliveryServers(address)

	lis, err := net.Listen("tcp", address)
	if err != nil {
		panic(fmt.Sprintf("Error starting EndorserServer %s", err))
	}
	addr := lis.Addr().String()

	test.Logf("Starting MockEndorserServer [%s]", addr)
	pb.RegisterEndorserServer(m.srv, m)

	m.wg.Add(1)
	go func() {
		defer m.wg.Done()
		if err := m.srv.Serve(lis); err != nil {
			test.Logf("StartEndorserServer failed [%s]", err)
		}
	}()

	return addr
}

// Stop the mock broadcast server and wait for completion.
func (m *MockEndorserServer) Stop() {
	if m.srv == nil {
		panic("MockEndorserServer not started")
	}

	m.srv.Stop()
	m.wg.Wait()
	m.srv = nil
}

// GetMockPeer will return the mock endorser's mock peer in a thread safe way
func (m *MockEndorserServer) GetMockPeer() *MockPeer {
	var v = func() *MockPeer {
		m.wg.Add(1)
		defer m.wg.Done()
		return m.mockPeer
	}

	return v()
}

// SetMockPeer will write the mock endorser's mock peer in a thread safe way
func (m *MockEndorserServer) SetMockPeer(mPeer *MockPeer) {
	func(p *MockPeer) {
		m.wg.Add(1)
		defer m.wg.Done()
		m.mockPeer = p
	}(mPeer)
}

func (m *MockEndorserServer) registerDiscoveryAndDeliveryServers(peerAddress string) {
	//register DiscoverService and DeliveryService
	discoverServer := discmocks.NewServer(
		discmocks.WithLocalPeers(
			&discmocks.MockDiscoveryPeerEndpoint{
				MSPID:        "Org1MSP",
				Endpoint:     peerAddress,
				LedgerHeight: 26,
			},
		),
		discmocks.WithPeers(
			&discmocks.MockDiscoveryPeerEndpoint{
				MSPID:        "Org1MSP",
				Endpoint:     peerAddress,
				LedgerHeight: 26,
			},
		),
	)

	discovery.RegisterDiscoveryServer(m.srv, discoverServer)

	deliverServer := eventmocks.NewMockDeliverServerWithFilteredDeliveries(m.FilteredDelListener)
	pb.RegisterDeliverServer(m.srv, deliverServer)
}

// MockPeer is a mock fabricsdk.Peer
type MockPeer struct {
	RWLock               *sync.RWMutex
	Error                error
	MockName             string
	MockURL              string
	MockRoles            []string
	MockCert             *pem.Block
	Payload              map[string][]byte
	ResponseMessage      string
	MockMSP              string
	Status               int32
	KVWrite              bool
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
func (p *MockPeer) ProcessTransactionProposal(tp fab.TransactionProposal, funcName []byte) (*fab.TransactionProposalResponse, error) {
	if p.RWLock != nil {
		p.RWLock.Lock()
		defer p.RWLock.Unlock()
	}
	p.ProcessProposalCalls++

	if p.Endorser == nil {
		// We serialize identities by prepending the MSPID and appending the ASN.1 DER content of the cert
		sID := &msp.SerializedIdentity{Mspid: "Org1MSP", IdBytes: []byte(CertPem)}
		endorser, err := proto.Marshal(sID)
		if err != nil {
			return nil, err
		}
		p.Endorser = endorser
	}

	block, _ := pem.Decode(KeyPem)
	lowLevelKey, err := x509.ParseECPrivateKey(block.Bytes)
	if err != nil {
		return nil, errors.Wrap(err, "Error received while parsing EC Key")
	}
	proposalResponsePayload, err := p.createProposalResponsePayload()
	if err != nil {
		return nil, errors.Wrap(err, "Error received while creating proposal response")
	}
	sigma, err := SignECDSA(lowLevelKey, append(proposalResponsePayload, p.Endorser...))
	if err != nil {
		return nil, errors.Wrap(err, "Error received while signing proposal for endorser with EC key")
	}

	payload, ok := p.Payload[string(funcName)]
	if !ok {
		payload, ok = p.Payload[string("default")]
		if !ok {
			fmt.Printf("payload for func(%s) not found\n", funcName)
		}
	}
	return &fab.TransactionProposalResponse{
		Endorser: p.MockURL,
		Status:   p.Status,
		ProposalResponse: &pb.ProposalResponse{Response: &pb.Response{
			Message: p.ResponseMessage, Status: p.Status, Payload: payload}, Payload: proposalResponsePayload,
			Endorsement: &pb.Endorsement{Endorser: p.Endorser, Signature: sigma}},
	}, p.Error

}

func (p *MockPeer) createProposalResponsePayload() ([]byte, error) {

	prp := &pb.ProposalResponsePayload{}
	ccAction := &pb.ChaincodeAction{}
	txRwSet := &rwsetutil.TxRwSet{}
	var kvWrite []*kvrwset.KVWrite
	if p.KVWrite {
		kvWrite = []*kvrwset.KVWrite{{Key: "key2", IsDelete: false, Value: []byte("value2")}}
	}
	txRwSet.NsRwSets = []*rwsetutil.NsRwSet{
		{NameSpace: "ns1", KvRwSet: &kvrwset.KVRWSet{
			Reads:  []*kvrwset.KVRead{{Key: "key1", Version: &kvrwset.Version{BlockNum: 1, TxNum: 1}}},
			Writes: kvWrite,
		}}}

	txRWSetBytes, err := txRwSet.ToProtoBytes()
	if err != nil {
		return nil, err
	}

	ccAction.Results = txRWSetBytes
	ccActionBytes, err := proto.Marshal(ccAction)
	if err != nil {
		return nil, err
	}
	prp.Extension = ccActionBytes
	prpBytes, err := proto.Marshal(prp)
	if err != nil {
		return nil, err
	}
	return prpBytes, nil
}

// SignECDSA sign with ec key
func SignECDSA(k *ecdsa.PrivateKey, digest []byte) (signature []byte, err error) {
	hash := sha256.New()
	_, err = hash.Write(digest)
	if err != nil {
		return nil, err
	}

	r, s, err := ecdsa.Sign(rand.Reader, k, hash.Sum(nil))
	if err != nil {
		return nil, err
	}

	s, err = utils.ToLowS(&k.PublicKey, s)
	if err != nil {
		return nil, err
	}

	return utils.MarshalECDSASignature(r, s)
}

// CertPem certificate
var CertPem = `-----BEGIN CERTIFICATE-----
MIICCjCCAbGgAwIBAgIQOcq9Om9VwUe9hGN0TTGw1DAKBggqhkjOPQQDAjBYMQsw
CQYDVQQGEwJVUzETMBEGA1UECBMKQ2FsaWZvcm5pYTEWMBQGA1UEBxMNU2FuIEZy
YW5jaXNjbzENMAsGA1UEChMET3JnMTENMAsGA1UEAxMET3JnMTAeFw0xNzA1MDgw
OTMwMzRaFw0yNzA1MDYwOTMwMzRaMGUxCzAJBgNVBAYTAlVTMRMwEQYDVQQIEwpD
YWxpZm9ybmlhMRYwFAYDVQQHEw1TYW4gRnJhbmNpc2NvMRUwEwYDVQQKEwxPcmcx
LXNlcnZlcjExEjAQBgNVBAMTCWxvY2FsaG9zdDBZMBMGByqGSM49AgEGCCqGSM49
AwEHA0IABAm+2CZhbmsnA+HKQynXKz7fVZvvwlv/DdNg3Mdg7lIcP2z0b07/eAZ5
0chdJNcjNAd/QAj/mmGG4dObeo4oTKGjUDBOMA4GA1UdDwEB/wQEAwIFoDAdBgNV
HSUEFjAUBggrBgEFBQcDAQYIKwYBBQUHAwIwDAYDVR0TAQH/BAIwADAPBgNVHSME
CDAGgAQBAgMEMAoGCCqGSM49BAMCA0cAMEQCIG55RvN4Boa0WS9UcIb/tI2YrAT8
EZd/oNnZYlbxxyvdAiB6sU9xAn4oYIW9xtrrOISv3YRg8rkCEATsagQfH8SiLg==
-----END CERTIFICATE-----`

// KeyPem ec private key
var KeyPem = []byte(`-----BEGIN EC PRIVATE KEY-----
MHcCAQEEICfXQtVmdQAlp/l9umWJqCXNTDurmciDNmGHPpxHwUK/oAoGCCqGSM49
AwEHoUQDQgAECb7YJmFuaycD4cpDKdcrPt9Vm+/CW/8N02Dcx2DuUhw/bPRvTv94
BnnRyF0k1yM0B39ACP+aYYbh05t6jihMoQ==
-----END EC PRIVATE KEY-----`)
