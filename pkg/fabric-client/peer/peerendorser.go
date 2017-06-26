/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package peer

import (
	"context"
	"fmt"
	"time"

	"github.com/hyperledger/fabric-sdk-go/api"
	pb "github.com/hyperledger/fabric/protos/peer"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

// peerEndorser enables access to a GRPC-based endorser for running transaction proposal simulations
type peerEndorser struct {
	grpcDialOption []grpc.DialOption
	target         string
}

func newPeerEndorser(target string, certificate string, serverHostOverride string, dialTimeout time.Duration, dialBlocking bool, config api.Config) (peerEndorser, error) {
	if len(target) == 0 {
		return peerEndorser{}, fmt.Errorf("Target is required")
	}

	// Construct dialer options for the connection
	var opts []grpc.DialOption
	opts = append(opts, grpc.WithTimeout(dialTimeout)) // TODO: should be configurable
	if dialBlocking {                                  // TODO: configurable?
		opts = append(opts, grpc.WithBlock())
	}

	if config.IsTLSEnabled() {
		if len(certificate) == 0 {
			return peerEndorser{}, fmt.Errorf("Certificate is required")
		}

		tlsCaCertPool, err := config.GetTLSCACertPool(certificate)
		if err != nil {
			return peerEndorser{}, err
		}
		creds := credentials.NewClientTLSFromCert(tlsCaCertPool, serverHostOverride)
		opts = append(opts, grpc.WithTransportCredentials(creds))
	} else {
		opts = append(opts, grpc.WithInsecure())
	}

	pc := peerEndorser{grpcDialOption: opts, target: target}

	return pc, nil
}

// ProcessProposal sends the transaction proposal to a peer and returns the response.
func (p *peerEndorser) ProcessProposal(proposal *api.TransactionProposal) (*api.TransactionProposalResponse, error) {
	proposalResponse, err := p.sendProposal(proposal)
	if err != nil {
		return nil, err
	}

	return &api.TransactionProposalResponse{
		Proposal:         proposal,
		ProposalResponse: proposalResponse,
		Endorser:         p.target, // TODO: what format is expected for Endorser? Just target? URL?
		Status:           proposalResponse.GetResponse().Status,
	}, nil
}

func (p *peerEndorser) conn() (*grpc.ClientConn, error) {
	return grpc.Dial(p.target, p.grpcDialOption...)
}

func (p *peerEndorser) releaseConn(conn *grpc.ClientConn) {
	conn.Close()
}

func (p *peerEndorser) sendProposal(proposal *api.TransactionProposal) (*pb.ProposalResponse, error) {
	conn, err := p.conn()
	if err != nil {
		return nil, err
	}
	defer p.releaseConn(conn)

	endorserClient := pb.NewEndorserClient(conn)
	return endorserClient.ProcessProposal(context.Background(), proposal.SignedProposal)
}
