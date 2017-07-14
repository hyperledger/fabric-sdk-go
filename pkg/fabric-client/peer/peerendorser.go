/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package peer

import (
	"context"
	"fmt"

	"github.com/hyperledger/fabric-sdk-go/api/apiconfig"
	"github.com/hyperledger/fabric-sdk-go/api/apitxn"
	pb "github.com/hyperledger/fabric/protos/peer"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

// peerEndorser enables access to a GRPC-based endorser for running transaction proposal simulations
type peerEndorser struct {
	grpcDialOption []grpc.DialOption
	target         string
}

// TransactionProposalError represents an error condition that prevented proposal processing.
type TransactionProposalError struct {
	Endorser string
	Proposal apitxn.TransactionProposal
	Err      error
}

func newPeerEndorser(target string, certificate string, serverHostOverride string,
	dialBlocking bool, config apiconfig.Config) (
	peerEndorser, error) {
	if len(target) == 0 {
		return peerEndorser{}, fmt.Errorf("Target is required")
	}

	// Construct dialer options for the connection
	var opts []grpc.DialOption
	opts = append(opts, grpc.WithTimeout(config.TimeoutOrDefault(apiconfig.Endorser)))
	if dialBlocking { // TODO: configurable?
		opts = append(opts, grpc.WithBlock())
	}

	if config.IsTLSEnabled() {
		certPool, _ := config.TLSCACertPool("")
		if len(certificate) == 0 && len(certPool.Subjects()) == 0 {
			return peerEndorser{}, fmt.Errorf("Certificate is required")
		}

		tlsCaCertPool, err := config.TLSCACertPool(certificate)
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

// ProcessTransactionProposal sends the transaction proposal to a peer and returns the response.
func (p *peerEndorser) ProcessTransactionProposal(proposal apitxn.TransactionProposal) (apitxn.TransactionProposalResult, error) {
	logger.Debugf("Processing proposal using endorser :%s", p.target)

	proposalResponse, err := p.sendProposal(proposal)
	if err != nil {
		tpe := TransactionProposalError{
			Endorser: p.target,
			Proposal: proposal,
			Err:      err,
		}
		return apitxn.TransactionProposalResult{}, &tpe
	}

	return apitxn.TransactionProposalResult{
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

func (p *peerEndorser) sendProposal(proposal apitxn.TransactionProposal) (*pb.ProposalResponse, error) {
	conn, err := p.conn()
	if err != nil {
		return nil, err
	}
	defer p.releaseConn(conn)

	endorserClient := pb.NewEndorserClient(conn)
	return endorserClient.ProcessProposal(context.Background(), proposal.SignedProposal)
}

func (tpe TransactionProposalError) Error() string {
	return fmt.Sprintf("Transaction processor (%s) returned error '%s' for proposal: %v", tpe.Endorser, tpe.Err.Error(), tpe.Proposal)
}
