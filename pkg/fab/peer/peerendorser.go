/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package peer

import (
	reqContext "context"
	"crypto/x509"
	"regexp"
	"time"

	"github.com/pkg/errors"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/keepalive"
	grpcstatus "google.golang.org/grpc/status"

	"github.com/hyperledger/fabric-protos-go/common"
	pb "github.com/hyperledger/fabric-protos-go/peer"
	"github.com/hyperledger/fabric-sdk-go/internal/github.com/hyperledger/fabric/protoutil"
	"github.com/hyperledger/fabric-sdk-go/pkg/client/common/verifier"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/errors/status"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/fab"
	"github.com/hyperledger/fabric-sdk-go/pkg/context"
	"github.com/hyperledger/fabric-sdk-go/pkg/core/config/comm"
	"github.com/hyperledger/fabric-sdk-go/pkg/core/config/endpoint"
)

const (
	// GRPC max message size (same as Fabric)
	maxCallRecvMsgSize = 100 * 1024 * 1024
	maxCallSendMsgSize = 100 * 1024 * 1024
)

var chaincodeNotFoundPattern = regexp.MustCompile(`(chaincode [^ ]+ not found)|(could not find chaincode with name)|(cannot get package for chaincode)`)

// peerEndorser enables access to a GRPC-based endorser for running transaction proposal simulations
type peerEndorser struct {
	grpcDialOption []grpc.DialOption
	target         string
	dialTimeout    time.Duration
	commManager    fab.CommManager
}

type peerEndorserRequest struct {
	target             string
	certificate        *x509.Certificate
	serverHostOverride string
	config             fab.EndpointConfig
	kap                keepalive.ClientParameters
	failFast           bool
	allowInsecure      bool
	commManager        fab.CommManager
}

func newPeerEndorser(endorseReq *peerEndorserRequest) (*peerEndorser, error) {
	if len(endorseReq.target) == 0 {
		return nil, errors.New("target is required")
	}

	// Construct dialer options for the connection
	var grpcOpts []grpc.DialOption
	if endorseReq.kap.Time > 0 {
		grpcOpts = append(grpcOpts, grpc.WithKeepaliveParams(endorseReq.kap))
	}
	grpcOpts = append(grpcOpts, grpc.WithDefaultCallOptions(grpc.WaitForReady(!endorseReq.failFast)))

	if endpoint.AttemptSecured(endorseReq.target, endorseReq.allowInsecure) {
		tlsConfig, err := comm.TLSConfig(endorseReq.certificate, endorseReq.serverHostOverride, endorseReq.config)
		if err != nil {
			return nil, err
		}
		//verify if certificate was expired or not yet valid
		tlsConfig.VerifyPeerCertificate = func(rawCerts [][]byte, verifiedChains [][]*x509.Certificate) error {
			return verifier.VerifyPeerCertificate(rawCerts, verifiedChains)
		}
		grpcOpts = append(grpcOpts, grpc.WithTransportCredentials(credentials.NewTLS(tlsConfig)))
	} else {
		grpcOpts = append(grpcOpts, grpc.WithInsecure())
	}

	grpcOpts = append(grpcOpts, grpc.WithDefaultCallOptions(grpc.MaxCallRecvMsgSize(maxCallRecvMsgSize),
		grpc.MaxCallSendMsgSize(maxCallSendMsgSize)))

	timeout := endorseReq.config.Timeout(fab.PeerConnection)

	pc := &peerEndorser{
		grpcDialOption: grpcOpts,
		target:         endpoint.ToAddress(endorseReq.target),
		dialTimeout:    timeout,
		commManager:    endorseReq.commManager,
	}

	return pc, nil
}

// ProcessTransactionProposal sends the transaction proposal to a peer and returns the response.
func (p *peerEndorser) ProcessTransactionProposal(ctx reqContext.Context, request fab.ProcessProposalRequest) (*fab.TransactionProposalResponse, error) {
	logger.Debugf("Processing proposal using endorser: %s", p.target)

	proposalResponse, err := p.sendProposal(ctx, request)
	if err != nil {
		tpr := fab.TransactionProposalResponse{Endorser: p.target}
		return &tpr, errors.Wrapf(err, "Transaction processing for endorser [%s]", p.target)
	}

	chaincodeStatus, err := getChaincodeResponseStatus(proposalResponse)
	if err != nil {
		return nil, errors.WithMessage(err, "chaincode response status parsing failed")
	}

	tpr := fab.TransactionProposalResponse{
		ProposalResponse: proposalResponse,
		Endorser:         p.target,
		ChaincodeStatus:  chaincodeStatus,
		Status:           proposalResponse.GetResponse().Status,
	}
	return &tpr, nil
}

func (p *peerEndorser) conn(ctx reqContext.Context) (*grpc.ClientConn, error) {
	commManager, ok := context.RequestCommManager(ctx)
	if !ok {
		commManager = p.commManager
	}

	ctx, cancel := reqContext.WithTimeout(ctx, p.dialTimeout)
	defer cancel()

	return commManager.DialContext(ctx, p.target, p.grpcDialOption...)
}

func (p *peerEndorser) releaseConn(ctx reqContext.Context, conn *grpc.ClientConn) {
	commManager, ok := context.RequestCommManager(ctx)
	if !ok {
		commManager = p.commManager
	}

	commManager.ReleaseConn(conn)
}

func (p *peerEndorser) sendProposal(ctx reqContext.Context, proposal fab.ProcessProposalRequest) (*pb.ProposalResponse, error) {
	conn, err := p.conn(ctx)
	if err != nil {
		rpcStatus, ok := grpcstatus.FromError(err)
		if ok {
			return nil, errors.WithMessage(status.NewFromGRPCStatus(rpcStatus), "connection failed")
		}
		return nil, status.New(status.EndorserClientStatus, status.ConnectionFailed.ToInt32(), err.Error(), []interface{}{p.target})
	}
	defer p.releaseConn(ctx, conn)

	endorserClient := pb.NewEndorserClient(conn)
	resp, err := endorserClient.ProcessProposal(ctx, proposal.SignedProposal)

	if err != nil {
		logger.Errorf("process proposal failed [%s]", err)
		rpcStatus, ok := grpcstatus.FromError(err)

		if ok {
			err = status.NewFromGRPCStatus(rpcStatus)
		}
	} else {
		//check error from response (for :fabric v1.2 and later)
		err = extractChaincodeErrorFromResponse(resp)
	}

	return resp, err
}

//extractChaincodeErrorFromResponse extracts chaincode error from proposal response
func extractChaincodeErrorFromResponse(resp *pb.ProposalResponse) error {
	if resp.Response.Status < int32(common.Status_SUCCESS) || resp.Response.Status >= int32(common.Status_BAD_REQUEST) {
		details := []interface{}{resp.Endorsement, resp.Response.Payload}
		if chaincodeNotFoundPattern.MatchString(resp.Response.Message) {
			return status.New(status.EndorserClientStatus, int32(status.ChaincodeNameNotFound), resp.Response.Message, details)
		}
		return status.New(status.ChaincodeStatus, resp.Response.Status, resp.Response.Message, details)
	}
	return nil
}

// getChaincodeResponseStatus gets the actual response status from response.Payload.extension.Response.status, as fabric always returns actual 200
func getChaincodeResponseStatus(response *pb.ProposalResponse) (int32, error) {
	if response.Payload != nil {
		payload, err := protoutil.UnmarshalProposalResponsePayload(response.Payload)
		if err != nil {
			return 0, errors.Wrap(err, "unmarshal of proposal response payload failed")
		}

		extension, err := protoutil.UnmarshalChaincodeAction(payload.Extension)
		if err != nil {
			return 0, errors.Wrap(err, "unmarshal of chaincode action failed")
		}

		if extension != nil && extension.Response != nil {
			return extension.Response.Status, nil
		}
	}
	return response.Response.Status, nil
}
