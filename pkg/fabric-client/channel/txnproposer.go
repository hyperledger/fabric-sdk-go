/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package channel

import (
	"fmt"
	"net/http"

	"github.com/golang/protobuf/proto"
	"github.com/hyperledger/fabric-sdk-go/pkg/errors"

	protos_utils "github.com/hyperledger/fabric-sdk-go/internal/github.com/hyperledger/fabric/protos/utils"
	"github.com/hyperledger/fabric-sdk-go/third_party/github.com/hyperledger/fabric/protos/common"
	pb "github.com/hyperledger/fabric-sdk-go/third_party/github.com/hyperledger/fabric/protos/peer"

	fab "github.com/hyperledger/fabric-sdk-go/api/apifabclient"
	"github.com/hyperledger/fabric-sdk-go/api/apitxn"
	"github.com/hyperledger/fabric-sdk-go/pkg/fabric-client/internal/txnproc"
)

// SendTransactionProposal sends the created proposal to peer for endorsement.
// TODO: return the entire request or just the txn ID?
func (c *Channel) SendTransactionProposal(request apitxn.ChaincodeInvokeRequest) ([]*apitxn.TransactionProposalResponse, apitxn.TransactionID, error) {
	request, err := c.chaincodeInvokeRequestAddDefaultPeers(request)
	if err != nil {
		return nil, apitxn.TransactionID{}, err
	}

	return SendTransactionProposalWithChannelID(c.name, request, c.clientContext)
}

// SendTransactionProposalWithChannelID sends the created proposal to peer for endorsement.
// TODO: return the entire request or just the txn ID?
func SendTransactionProposalWithChannelID(channelID string, request apitxn.ChaincodeInvokeRequest, clientContext ClientContext) ([]*apitxn.TransactionProposalResponse, apitxn.TransactionID, error) {
	if err := validateChaincodeInvokeRequest(request); err != nil {
		return nil, apitxn.TransactionID{}, errors.WithMessage(err, "validateChaincodeInvokeRequest failed")
	}

	request, err := chaincodeInvokeRequestAddTxnID(request, clientContext)
	if err != nil {
		return nil, request.TxnID, err
	}

	proposal, err := newTransactionProposal(channelID, request, clientContext)
	if err != nil {
		return nil, request.TxnID, err
	}

	responses, err := txnproc.SendTransactionProposalToProcessors(proposal, request.Targets)
	return responses, request.TxnID, err
}

func validateChaincodeInvokeRequest(request apitxn.ChaincodeInvokeRequest) error {
	if request.ChaincodeID == "" {
		return errors.New("ChaincodeID is required")
	}

	if request.Fcn == "" {
		return errors.New("Fcn is required")
	}

	if request.Targets == nil || len(request.Targets) < 1 {
		return errors.New("Targets is required")
	}
	return nil
}

func chaincodeInvokeRequestAddTxnID(request apitxn.ChaincodeInvokeRequest, clientContext ClientContext) (apitxn.ChaincodeInvokeRequest, error) {
	// create txn id (if needed)
	if request.TxnID.ID == "" {
		txid, err := clientContext.NewTxnID()
		if err != nil {
			return request, errors.WithMessage(err, "NewTxnID failed")
		}
		request.TxnID = txid
	}

	return request, nil
}

func (c *Channel) chaincodeInvokeRequestAddDefaultPeers(request apitxn.ChaincodeInvokeRequest) (apitxn.ChaincodeInvokeRequest, error) {
	// Use default peers if targets are not specified.
	if request.Targets == nil || len(request.Targets) == 0 {
		if c.peers == nil || len(c.peers) == 0 {
			return request, errors.New("targets were not specified and no peers have been configured")
		}

		request.Targets = c.txnProcessors()
	}
	return request, nil
}

// newTransactionProposal creates a proposal for transaction. This involves assembling the proposal
// with the data (chaincodeName, function to call, arguments, transient data, etc.) and signing it using the private key corresponding to the
// ECert to sign.
func newTransactionProposal(channelID string, request apitxn.ChaincodeInvokeRequest, clientContext ClientContext) (*apitxn.TransactionProposal, error) {

	// Add function name to arguments
	argsArray := make([][]byte, len(request.Args)+1)
	argsArray[0] = []byte(request.Fcn)
	for i, arg := range request.Args {
		argsArray[i+1] = arg
	}

	// create invocation spec to target a chaincode with arguments
	ccis := &pb.ChaincodeInvocationSpec{ChaincodeSpec: &pb.ChaincodeSpec{
		Type: pb.ChaincodeSpec_GOLANG, ChaincodeId: &pb.ChaincodeID{Name: request.ChaincodeID},
		Input: &pb.ChaincodeInput{Args: argsArray}}}

	// create a proposal from a ChaincodeInvocationSpec
	if clientContext.UserContext() == nil {
		return nil, errors.New("User context is nil")
	}
	creator, err := clientContext.UserContext().Identity()
	if err != nil {
		return nil, errors.WithMessage(err, "Failed to get user context identity")
	}

	proposal, _, err := protos_utils.CreateChaincodeProposalWithTxIDNonceAndTransient(request.TxnID.ID, common.HeaderType_ENDORSER_TRANSACTION, channelID, ccis, request.TxnID.Nonce, creator, request.TransientMap)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create chaincode proposal")
	}

	// sign proposal bytes
	proposalBytes, err := proto.Marshal(proposal)
	if err != nil {
		return nil, errors.Wrap(err, "marshal proposal failed")
	}

	user := clientContext.UserContext()
	if user == nil {
		return nil, errors.WithMessage(err, "failed to get user context")
	}

	signingMgr := clientContext.SigningManager()
	if signingMgr == nil {
		return nil, errors.New("signing manager is nil")
	}

	signature, err := signingMgr.Sign(proposalBytes, user.PrivateKey())
	if err != nil {
		return nil, err
	}

	// construct the transaction proposal
	signedProposal := pb.SignedProposal{ProposalBytes: proposalBytes, Signature: signature}
	tp := apitxn.TransactionProposal{
		TxnID:          request.TxnID,
		SignedProposal: &signedProposal,
		Proposal:       proposal,
	}

	return &tp, nil
}

// TODO: There should be a strategy for choosing processors.
func (c *Channel) txnProcessors() []apitxn.ProposalProcessor {
	return peersToTxnProcessors(c.Peers())
}

// peersToTxnProcessors converts a slice of Peers to a slice of ProposalProcessors
func peersToTxnProcessors(peers []fab.Peer) []apitxn.ProposalProcessor {
	tpp := make([]apitxn.ProposalProcessor, len(peers))

	for i := range peers {
		tpp[i] = peers[i]
	}
	return tpp
}

// ProposalBytes returns the serialized transaction.
func (c *Channel) ProposalBytes(tp *apitxn.TransactionProposal) ([]byte, error) {
	return proto.Marshal(tp.SignedProposal)
}

func (c *Channel) signProposal(proposal *pb.Proposal) (*pb.SignedProposal, error) {
	user := c.clientContext.UserContext()
	if user == nil {
		return nil, errors.New("user context is nil")
	}

	proposalBytes, err := proto.Marshal(proposal)
	if err != nil {
		return nil, errors.Wrap(err, "mashal proposal failed")
	}

	signingMgr := c.clientContext.SigningManager()
	if signingMgr == nil {
		return nil, errors.New("signing manager is nil")
	}

	signature, err := signingMgr.Sign(proposalBytes, user.PrivateKey())
	if err != nil {
		return nil, errors.WithMessage(err, "signing proposal failed")
	}

	return &pb.SignedProposal{ProposalBytes: proposalBytes, Signature: signature}, nil
}

// JoinChannel sends a join channel proposal to one or more endorsing peers
// Will get the genesis block from the defined orderer to be used
// in the proposal.
// request: An object containing the following fields:
// `targets` : required - An array of `Peer` objects that will join
//             this channel
// `block`   : the genesis block of the channel
//             see GenesisBlock() method
// `txId`    : required - String of the transaction id
// `nonce`   : required - Integer of the once time number
// See /protos/peer/proposal_response.proto
func (c *Channel) JoinChannel(request *fab.JoinChannelRequest) error {
	logger.Debug("joinChannel - start")

	// verify that we have targets (Peers) to join this channel
	// defined by the caller
	if request == nil {
		return errors.New("join channel request is required")
	}

	// verify that a Peer(s) has been selected to join this channel
	if request.Targets == nil {
		return errors.New("missing targets input parameter with the peer objects for the join channel proposal")
	}

	// verify that we have transaction id
	if request.TxnID.ID == "" {
		return errors.New("missing txId input parameter with the required transaction identifier")
	}

	// verify that we have the nonce
	if request.TxnID.Nonce == nil {
		return errors.New("missing nonce input parameter with the required single use number")
	}

	if request.GenesisBlock == nil {
		return errors.New("missing block input parameter with the required genesis block")
	}

	if c.clientContext.UserContext() == nil {
		return errors.New("user context is nil")
	}
	creator, err := c.clientContext.UserContext().Identity()
	if err != nil {
		return errors.WithMessage(err, "getting creator identity failed")
	}

	genesisBlockBytes, err := proto.Marshal(request.GenesisBlock)
	if err != nil {
		return errors.Wrap(err, "marshal genesis block failed")
	}

	// Create join channel transaction proposal for target peers
	joinCommand := "JoinChain"
	var args [][]byte
	args = append(args, []byte(joinCommand))
	args = append(args, genesisBlockBytes)
	ccSpec := &pb.ChaincodeSpec{
		Type:        pb.ChaincodeSpec_GOLANG,
		ChaincodeId: &pb.ChaincodeID{Name: "cscc"},
		Input:       &pb.ChaincodeInput{Args: args},
	}
	cciSpec := &pb.ChaincodeInvocationSpec{
		ChaincodeSpec: ccSpec,
	}

	proposal, _, err := protos_utils.CreateChaincodeProposalWithTxIDNonceAndTransient(request.TxnID.ID, common.HeaderType_ENDORSER_TRANSACTION, "", cciSpec, request.TxnID.Nonce, creator, nil)
	if err != nil {
		return errors.Wrap(err, "failed to build chaincode proposal")
	}
	signedProposal, err := c.signProposal(proposal)
	if err != nil {
		return errors.WithMessage(err, "signing proposal failed")
	}
	transactionProposal := &apitxn.TransactionProposal{
		TxnID:          request.TxnID,
		SignedProposal: signedProposal,
		Proposal:       proposal,
	}

	targets := peersToTxnProcessors(request.Targets)

	// Send join proposal
	proposalResponses, err := txnproc.SendTransactionProposalToProcessors(transactionProposal, targets)
	if err != nil {
		return errors.WithMessage(err, "sending join transaction proposal failed")
	}
	// Check responses from target peers for success/failure and join all errors
	var joinError string
	for _, response := range proposalResponses {
		if response.Err != nil {
			joinError = joinError +
				fmt.Sprintf("join channel proposal response error: %s \n",
					response.Err.Error())
		} else if response.Status != http.StatusOK {
			joinError = joinError +
				fmt.Sprintf("join channel proposal HTTP response status: %d \n", response.Status)
		}
	}

	if joinError != "" {
		return errors.New(joinError)
	}

	return nil
}
