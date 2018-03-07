/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package peer

import (
	reqContext "context"
	"crypto/x509"
	"time"

	"github.com/pkg/errors"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/keepalive"
	grpcstatus "google.golang.org/grpc/status"

	"github.com/hyperledger/fabric-sdk-go/pkg/context/api/core"
	"github.com/hyperledger/fabric-sdk-go/pkg/context/api/fab"
	"github.com/hyperledger/fabric-sdk-go/pkg/core/config/comm"
	"github.com/hyperledger/fabric-sdk-go/pkg/core/config/urlutil"
	"github.com/hyperledger/fabric-sdk-go/pkg/errors/status"
	pb "github.com/hyperledger/fabric-sdk-go/third_party/github.com/hyperledger/fabric/protos/peer"
)

// peerEndorser enables access to a GRPC-based endorser for running transaction proposal simulations
type peerEndorser struct {
	grpcDialOption       []grpc.DialOption
	target               string
	dialTimeout          time.Duration
	failFast             bool
	transportCredentials credentials.TransportCredentials
	secured              bool
	allowInsecure        bool
	connector            connProvider
}

type peerEndorserRequest struct {
	target             string
	certificate        *x509.Certificate
	serverHostOverride string
	config             core.Config
	kap                keepalive.ClientParameters
	failFast           bool
	allowInsecure      bool
	connector          connProvider
}

func newPeerEndorser(endorseReq *peerEndorserRequest) (*peerEndorser, error) {
	if len(endorseReq.target) == 0 {
		return nil, errors.New("target is required")
	}

	// Construct dialer options for the connection
	var opts []grpc.DialOption
	if endorseReq.kap.Time > 0 || endorseReq.kap.Timeout > 0 {
		opts = append(opts, grpc.WithKeepaliveParams(endorseReq.kap))
	}
	opts = append(opts, grpc.WithDefaultCallOptions(grpc.FailFast(endorseReq.failFast)))

	timeout := endorseReq.config.TimeoutOrDefault(core.EndorserConnection)

	tlsConfig, err := comm.TLSConfig(endorseReq.certificate, endorseReq.serverHostOverride, endorseReq.config)
	if err != nil {
		return nil, err
	}

	pc := &peerEndorser{grpcDialOption: opts, target: urlutil.ToAddress(endorseReq.target), dialTimeout: timeout,
		transportCredentials: credentials.NewTLS(tlsConfig), secured: urlutil.AttemptSecured(endorseReq.target),
		allowInsecure: endorseReq.allowInsecure, connector: endorseReq.connector}

	return pc, nil
}

// ProcessTransactionProposal sends the transaction proposal to a peer and returns the response.
func (p *peerEndorser) ProcessTransactionProposal(ctx reqContext.Context, request fab.ProcessProposalRequest) (*fab.TransactionProposalResponse, error) {
	logger.Debugf("Processing proposal using endorser: %s", p.target)

	proposalResponse, err := p.sendProposal(ctx, request, p.secured)
	if err != nil {
		tpr := fab.TransactionProposalResponse{Endorser: p.target}
		return &tpr, errors.Wrapf(err, "Transaction processing for endorser [%s]", p.target)
	}

	tpr := fab.TransactionProposalResponse{
		ProposalResponse: proposalResponse,
		Endorser:         p.target,
		Status:           proposalResponse.GetResponse().Status,
	}
	return &tpr, nil
}

func (p *peerEndorser) conn(ctx reqContext.Context, secured bool) (*grpc.ClientConn, error) {
	// Establish connection to Ordering Service
	var grpcOpts []grpc.DialOption
	if secured {
		grpcOpts = append(p.grpcDialOption, grpc.WithTransportCredentials(p.transportCredentials))
	} else {
		grpcOpts = append(p.grpcDialOption, grpc.WithInsecure())
	}

	ctx, cancel := reqContext.WithTimeout(ctx, p.dialTimeout)
	defer cancel()

	return p.connector.DialContext(ctx, p.target, grpcOpts...)
}

func (p *peerEndorser) sendProposal(ctx reqContext.Context, proposal fab.ProcessProposalRequest, secured bool) (*pb.ProposalResponse, error) {
	conn, err := p.conn(ctx, secured)
	if err != nil {
		if secured && p.allowInsecure {
			//If secured mode failed and allow insecure is enabled then retry in insecure mode
			logger.Debug("Secured NewEndorserClient failed, attempting insecured")
			return p.sendProposal(ctx, proposal, false)
		}
		rpcStatus, ok := grpcstatus.FromError(err)
		if ok {
			return nil, errors.WithMessage(status.NewFromGRPCStatus(rpcStatus), "connection failed")
		}
		return nil, status.New(status.EndorserClientStatus, status.ConnectionFailed.ToInt32(), err.Error(), []interface{}{p.target})
	}
	defer p.connector.ReleaseConn(conn)

	endorserClient := pb.NewEndorserClient(conn)
	resp, err := endorserClient.ProcessProposal(ctx, proposal.SignedProposal)
	if err != nil {
		logger.Errorf("process proposal failed [%s]", err)
		rpcStatus, ok := grpcstatus.FromError(err)
		if ok {
			err = status.NewFromGRPCStatus(rpcStatus)
		}
	}
	return resp, err
}
