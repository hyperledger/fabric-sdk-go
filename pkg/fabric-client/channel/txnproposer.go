/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package channel

import (
	"fmt"
	"net/http"
	"sync"

	"github.com/golang/protobuf/proto"

	"github.com/hyperledger/fabric/bccsp"
	"github.com/hyperledger/fabric/protos/common"
	pb "github.com/hyperledger/fabric/protos/peer"
	protos_utils "github.com/hyperledger/fabric/protos/utils"

	fab "github.com/hyperledger/fabric-sdk-go/api/apifabclient"
	"github.com/hyperledger/fabric-sdk-go/api/apitxn"
	fc "github.com/hyperledger/fabric-sdk-go/pkg/fabric-client/internal"
)

// CreateTransactionProposal creates a proposal for transaction. This involves assembling the proposal
// with the data (chaincodeName, function to call, arguments, transient data, etc.) and signing it using the private key corresponding to the
// ECert to sign.
func (c *Channel) CreateTransactionProposal(chaincodeName string,
	args []string, sign bool, transientData map[string][]byte) (*apitxn.TransactionProposal, error) {
	return createTransactionProposal(chaincodeName, c.Name(), args, sign, transientData, c.clientContext)
}

func createTransactionProposal(chaincodeName string, channelID string,
	args []string, sign bool, transientData map[string][]byte, clientContext fab.FabricClient) (*apitxn.TransactionProposal, error) {

	argsArray := make([][]byte, len(args))
	for i, arg := range args {
		argsArray[i] = []byte(arg)
	}
	ccis := &pb.ChaincodeInvocationSpec{ChaincodeSpec: &pb.ChaincodeSpec{
		Type: pb.ChaincodeSpec_GOLANG, ChaincodeId: &pb.ChaincodeID{Name: chaincodeName},
		Input: &pb.ChaincodeInput{Args: argsArray}}}

	creator, err := clientContext.GetIdentity()
	if err != nil {
		return nil, fmt.Errorf("Error getting creator: %v", err)
	}

	// create a proposal from a ChaincodeInvocationSpec
	proposal, txID, err := protos_utils.CreateChaincodeProposalWithTransient(common.HeaderType_ENDORSER_TRANSACTION, channelID, ccis, creator, transientData)
	if err != nil {
		return nil, fmt.Errorf("Could not create chaincode proposal, err %s", err)
	}

	proposalBytes, err := proto.Marshal(proposal)
	if err != nil {
		return nil, fmt.Errorf("Error marshalling proposal: %v", err)
	}

	user, err := clientContext.LoadUserFromStateStore("")
	if err != nil {
		return nil, fmt.Errorf("Error loading user from store: %s", err)
	}

	signature, err := fc.SignObjectWithKey(proposalBytes, user.PrivateKey(),
		&bccsp.SHAOpts{}, nil, clientContext.GetCryptoSuite())
	if err != nil {
		return nil, err
	}
	signedProposal := &pb.SignedProposal{ProposalBytes: proposalBytes, Signature: signature}
	return &apitxn.TransactionProposal{
		TransactionID:  txID,
		SignedProposal: signedProposal,
		Proposal:       proposal,
	}, nil

}

// SendTransactionProposal sends the created proposal to peer for endorsement.
func (c *Channel) SendTransactionProposal(proposal *apitxn.TransactionProposal, retry int, targets []apitxn.ProposalProcessor) ([]*apitxn.TransactionProposalResponse, error) {
	if proposal == nil || proposal.SignedProposal == nil {
		return nil, fmt.Errorf("signedProposal is nil")
	}

	if targets == nil || len(targets) == 0 {
		if c.peers == nil || len(c.peers) == 0 {
			return nil, fmt.Errorf("peers and target peers is nil or empty")
		}

		return c.SendTransactionProposal(proposal, retry, c.txnProcessors())
	}

	return SendTransactionProposal(proposal, retry, targets)
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

// SendTransactionProposal ... TODO (should be refactored)
func SendTransactionProposal(proposal *apitxn.TransactionProposal, retry int, targets []apitxn.ProposalProcessor) ([]*apitxn.TransactionProposalResponse, error) {

	if proposal == nil || proposal.SignedProposal == nil {
		return nil, fmt.Errorf("signedProposal is nil")
	}

	if len(targets) < 1 {
		return nil, fmt.Errorf("Missing peer objects for sending transaction proposal")
	}

	var responseMtx sync.Mutex
	var transactionProposalResponses []*apitxn.TransactionProposalResponse
	var wg sync.WaitGroup

	for _, p := range targets {
		wg.Add(1)
		go func(processor apitxn.ProposalProcessor) {
			defer wg.Done()

			r, err := processor.ProcessTransactionProposal(*proposal)
			if err != nil {
				logger.Debugf("Received error response from txn proposal processing: %v", err)
				// Error is handled downstream.
			}

			tpr := apitxn.TransactionProposalResponse{
				TransactionProposalResult: r, Err: err}

			responseMtx.Lock()
			transactionProposalResponses = append(transactionProposalResponses, &tpr)
			responseMtx.Unlock()
		}(p)
	}
	wg.Wait()
	return transactionProposalResponses, nil
}

// ProposalBytes returns the serialized transaction.
func (c *Channel) ProposalBytes(tp *apitxn.TransactionProposal) ([]byte, error) {
	return proto.Marshal(tp.SignedProposal)
}

func (c *Channel) signProposal(proposal *pb.Proposal) (*pb.SignedProposal, error) {
	user, err := c.clientContext.LoadUserFromStateStore("")
	if err != nil {
		return nil, fmt.Errorf("LoadUserFromStateStore return error: %s", err)
	}

	proposalBytes, err := proto.Marshal(proposal)
	if err != nil {
		return nil, fmt.Errorf("Error mashalling proposal: %s", err)
	}

	signature, err := fc.SignObjectWithKey(proposalBytes, user.PrivateKey(), &bccsp.SHAOpts{}, nil, c.clientContext.GetCryptoSuite())
	if err != nil {
		return nil, fmt.Errorf("Error signing proposal: %s", err)
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
		return fmt.Errorf("JoinChannel - error: Missing all required input request parameters")
	}

	// verify that a Peer(s) has been selected to join this channel
	if request.Targets == nil {
		return fmt.Errorf("JoinChannel - error: Missing targets input parameter with the peer objects for the join channel proposal")
	}

	// verify that we have transaction id
	if request.TxID == "" {
		return fmt.Errorf("JoinChannel - error: Missing txId input parameter with the required transaction identifier")
	}

	// verify that we have the nonce
	if request.Nonce == nil {
		return fmt.Errorf("JoinChannel - error: Missing nonce input parameter with the required single use number")
	}

	if request.GenesisBlock == nil {
		return fmt.Errorf("JoinChannel - error: Missing block input parameter with the required genesis block")
	}

	creator, err := c.clientContext.GetIdentity()
	if err != nil {
		return fmt.Errorf("Error getting creator ID: %v", err)
	}

	genesisBlockBytes, err := proto.Marshal(request.GenesisBlock)
	if err != nil {
		return fmt.Errorf("Error marshalling genesis block: %v", err)
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

	proposal, txID, err := protos_utils.CreateChaincodeProposalWithTxIDNonceAndTransient(request.TxID, common.HeaderType_ENDORSER_TRANSACTION, "", cciSpec, request.Nonce, creator, nil)
	if err != nil {
		return fmt.Errorf("Error building proposal: %v", err)
	}
	signedProposal, err := c.signProposal(proposal)
	if err != nil {
		return fmt.Errorf("Error signing proposal: %v", err)
	}
	transactionProposal := &apitxn.TransactionProposal{
		TransactionID:  txID,
		SignedProposal: signedProposal,
		Proposal:       proposal,
	}

	targets := peersToTxnProcessors(request.Targets)

	// Send join proposal
	proposalResponses, err := c.SendTransactionProposal(transactionProposal, 0, targets)
	if err != nil {
		return fmt.Errorf("Error sending join transaction proposal: %s", err)
	}
	// Check responses from target peers for success/failure and join all errors
	var joinError string
	for _, response := range proposalResponses {
		if response.Err != nil {
			joinError = joinError +
				fmt.Sprintf("Join channel proposal response error: %s \n",
					response.Err.Error())
		} else if response.Status != http.StatusOK {
			joinError = joinError +
				fmt.Sprintf("Join channel proposal HTTP response error: %s \n",
					response.Err.Error())
		}
	}

	if joinError != "" {
		return fmt.Errorf(joinError)
	}

	return nil
}
