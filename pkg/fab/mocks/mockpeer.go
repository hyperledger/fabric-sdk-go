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

	"github.com/golang/protobuf/proto"
	"github.com/hyperledger/fabric-protos-go/ledger/rwset/kvrwset"
	pb "github.com/hyperledger/fabric-protos-go/peer"
	"github.com/hyperledger/fabric-sdk-go/internal/github.com/hyperledger/fabric/core/ledger/kvledger/txmgmt/rwsetutil"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/fab"
)

// MockPeer is a mock fabricsdk.Peer.
type MockPeer struct {
	RWLock                  *sync.RWMutex
	Error                   error
	MockName                string
	MockURL                 string
	MockRoles               []string
	MockCert                *pem.Block
	Payload                 []byte
	ResponseMessage         string
	ProposalResponsePayload []byte // Overrides proposal response payload generated from other values
	MockMSP                 string
	Status                  int32
	ProcessProposalCalls    int
	Endorser                []byte
	ChaincodeID             string
	RwSets                  []*rwsetutil.NsRwSet
	properties              fab.Properties
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

// Properties returns the peer's properties
func (p *MockPeer) Properties() fab.Properties {
	return p.properties
}

// SetProperties sets the peer's properties
func (p *MockPeer) SetProperties(properties fab.Properties) {
	p.properties = properties
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
		ProposalResponse: &pb.ProposalResponse{
			Response: &pb.Response{
				Message: p.ResponseMessage,
				Status:  p.Status,
				Payload: p.Payload,
			},
			Endorsement: &pb.Endorsement{
				Endorser:  p.Endorser,
				Signature: []byte("signature"),
			},
			Payload: p.getProposalResponsePayload(),
		},
	}, p.Error
}

// SetChaincodeID sets the ID of the chaincode that was invoked. This ID will be
// set in the ChaincodeAction of the proposal response payload.
func (p *MockPeer) SetChaincodeID(ccID string) {
	p.ChaincodeID = ccID
}

// SetRwSets sets the read-write sets that will be set in the proposal response payload
func (p *MockPeer) SetRwSets(rwSets ...*rwsetutil.NsRwSet) {
	p.RwSets = rwSets
}

// NewRwSet returns a new read-write set for the given chaincode
func NewRwSet(ccID string) *rwsetutil.NsRwSet {
	return &rwsetutil.NsRwSet{
		NameSpace:        ccID,
		KvRwSet:          &kvrwset.KVRWSet{},
		CollHashedRwSets: nil,
	}
}

func (p *MockPeer) getProposalResponsePayload() []byte {
	if len(p.RwSets) == 0 && p.ChaincodeID != "" {
		// Create one RWSet from the specified chaincode ID
		p.SetRwSets(NewRwSet(p.ChaincodeID))
	}

	if p.ChaincodeID == "" && len(p.RwSets) > 0 {
		// Set the chaincode ID to be that of the namespace of the first RWSet
		p.ChaincodeID = p.RwSets[0].NameSpace
	}

	if len(p.ProposalResponsePayload) > 0 {
		return p.ProposalResponsePayload
	}

	payload := p.newProposalResponsePayload()
	payloadBytes, err := proto.Marshal(payload)
	if err != nil {
		panic(err)
	}

	return payloadBytes
}

func (p *MockPeer) newProposalResponsePayload() *pb.ProposalResponsePayload {
	chaincodeAction := p.newChaincodeAction()
	chaincodeActionBytes, err := proto.Marshal(chaincodeAction)
	if err != nil {
		panic(err)
	}

	return &pb.ProposalResponsePayload{
		Extension: chaincodeActionBytes,
	}
}

func (p *MockPeer) newChaincodeAction() *pb.ChaincodeAction {
	chaincodeAction := &pb.ChaincodeAction{
		Events: nil,
		Response: &pb.Response{
			Message: p.ResponseMessage,
			Status:  p.Status,
			Payload: p.Payload,
		},
		Results: p.getRWSet(),
	}

	if p.ChaincodeID != "" {
		chaincodeAction.ChaincodeId = &pb.ChaincodeID{Name: p.ChaincodeID}
	}

	return chaincodeAction
}

func (p *MockPeer) getRWSet() []byte {
	if len(p.RwSets) == 0 {
		return nil
	}

	txRWSet := &rwsetutil.TxRwSet{
		NsRwSets: p.RwSets,
	}
	txRWSetBytes, err := txRWSet.ToProtoBytes()
	if err != nil {
		panic(err)
	}

	return txRWSetBytes
}
