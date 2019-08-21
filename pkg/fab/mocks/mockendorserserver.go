/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package mocks

import (
	"fmt"
	"net"
	"sync"

	"github.com/golang/protobuf/proto"
	"github.com/hyperledger/fabric-protos-go/ledger/rwset/kvrwset"
	pb "github.com/hyperledger/fabric-protos-go/peer"
	"github.com/hyperledger/fabric-sdk-go/internal/github.com/hyperledger/fabric/core/ledger/kvledger/txmgmt/rwsetutil"
	"github.com/hyperledger/fabric-sdk-go/pkg/util/test"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

// MockEndorserServer mock endoreser server to process endorsement proposals
type MockEndorserServer struct {
	Creds         credentials.TransportCredentials
	ProposalError error
	wg            sync.WaitGroup
	AddkvWrite    bool
	srv           *grpc.Server
}

// ProcessProposal mock implementation that returns success if error is not set
// error if it is
func (m *MockEndorserServer) ProcessProposal(context context.Context,
	proposal *pb.SignedProposal) (*pb.ProposalResponse, error) {
	if m.ProposalError == nil {
		return &pb.ProposalResponse{Response: &pb.Response{
			Status: 200,
		}, Endorsement: &pb.Endorsement{Endorser: []byte("endorser"), Signature: []byte("signature")},
			Payload: m.createProposalResponsePayload()}, nil
	}
	return &pb.ProposalResponse{Response: &pb.Response{
		Status:  500,
		Message: m.ProposalError.Error(),
	}}, m.ProposalError
}

func (m *MockEndorserServer) createProposalResponsePayload() []byte {

	prp := &pb.ProposalResponsePayload{}
	ccAction := &pb.ChaincodeAction{}
	txRwSet := &rwsetutil.TxRwSet{}

	if m.AddkvWrite {
		txRwSet.NsRwSets = []*rwsetutil.NsRwSet{
			{NameSpace: "ns1", KvRwSet: &kvrwset.KVRWSet{
				Reads:  []*kvrwset.KVRead{{Key: "key1", Version: &kvrwset.Version{BlockNum: 1, TxNum: 1}}},
				Writes: []*kvrwset.KVWrite{{Key: "key2", IsDelete: false, Value: []byte("value2")}},
			}}}
	}

	txRWSetBytes, err := txRwSet.ToProtoBytes()
	if err != nil {
		return nil
	}
	ccAction.Results = txRWSetBytes
	ccActionBytes, err := proto.Marshal(ccAction)
	if err != nil {
		return nil
	}
	prp.Extension = ccActionBytes
	prpBytes, err := proto.Marshal(prp)
	if err != nil {
		return nil
	}
	return prpBytes
}

// Start the mock broadcast server
func (m *MockEndorserServer) Start(address string) string {
	if m.srv != nil {
		panic("MockBroadcastServer already started")
	}

	// pass in TLS creds if present
	if m.Creds != nil {
		m.srv = grpc.NewServer(grpc.Creds(m.Creds))
	} else {
		m.srv = grpc.NewServer()
	}

	lis, err := net.Listen("tcp", address)
	if err != nil {
		panic(fmt.Sprintf("Error starting BroadcastServer %s", err))
	}
	addr := lis.Addr().String()

	test.Logf("Starting MockEventServer [%s]", addr)
	pb.RegisterEndorserServer(m.srv, m)
	m.wg.Add(1)
	go func() {
		defer m.wg.Done()
		if err := m.srv.Serve(lis); err != nil {
			test.Logf("StartMockBroadcastServer failed [%s]", err)
		}
	}()

	return addr
}

// Stop the mock broadcast server and wait for completion.
func (m *MockEndorserServer) Stop() {
	if m.srv == nil {
		panic("MockBroadcastServer not started")
	}

	m.srv.Stop()
	m.wg.Wait()
	m.srv = nil
}
