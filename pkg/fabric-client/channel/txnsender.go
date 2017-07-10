/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package channel

import (
	"fmt"
	"math/rand"
	"sync"
	"time"

	"github.com/golang/protobuf/proto"
	proto_ts "github.com/golang/protobuf/ptypes/timestamp"
	fab "github.com/hyperledger/fabric-sdk-go/api/apifabclient"
	"github.com/hyperledger/fabric-sdk-go/api/apitxn"
	fc "github.com/hyperledger/fabric-sdk-go/pkg/fabric-client/internal"
	"github.com/hyperledger/fabric-sdk-go/pkg/fabric-client/internal/txnproc"
	"github.com/hyperledger/fabric/bccsp"
	"github.com/hyperledger/fabric/protos/common"
	mspprotos "github.com/hyperledger/fabric/protos/msp"
	pb "github.com/hyperledger/fabric/protos/peer"
	protos_utils "github.com/hyperledger/fabric/protos/utils"
)

func init() {
	rand.Seed(time.Now().UnixNano())
}

// CreateTransaction create a transaction with proposal response, following the endorsement policy.
func (c *Channel) CreateTransaction(resps []*apitxn.TransactionProposalResponse) (*apitxn.Transaction, error) {
	if len(resps) == 0 {
		return nil, fmt.Errorf("At least one proposal response is necessary")
	}

	proposal := &resps[0].Proposal

	// the original header
	hdr, err := protos_utils.GetHeader(proposal.Proposal.Header)
	if err != nil {
		return nil, fmt.Errorf("Could not unmarshal the proposal header")
	}

	// the original payload
	pPayl, err := protos_utils.GetChaincodeProposalPayload(proposal.Proposal.Payload)
	if err != nil {
		return nil, fmt.Errorf("Could not unmarshal the proposal payload")
	}

	// get header extensions so we have the visibility field
	hdrExt, err := protos_utils.GetChaincodeHeaderExtension(hdr)
	if err != nil {
		return nil, err
	}

	// This code is commented out because the ProposalResponsePayload Extension ChaincodeAction Results
	// return from endorsements is different so the compare will fail

	//            var a1 []byte
	//            for n, r := range resps {
	//                            if n == 0 {
	//                                            a1 = r.Payload
	//                                            if r.Response.Status != 200 {
	//                                                            return nil, fmt.Errorf("Proposal response was not successful, error code %d, msg %s", r.Response.Status, r.Response.Message)
	//                                            }
	//                                            continue
	//                            }

	//                            if bytes.Compare(a1, r.Payload) != 0 {
	//                                            return nil, fmt.Errorf("ProposalResponsePayloads do not match")
	//                            }
	//            }

	for _, r := range resps {
		if r.ProposalResponse.Response.Status != 200 {
			return nil, fmt.Errorf("Proposal response was not successful, error code %d, msg %s", r.ProposalResponse.Response.Status, r.ProposalResponse.Response.Message)
		}
	}

	// fill endorsements
	endorsements := make([]*pb.Endorsement, len(resps))
	for n, r := range resps {
		endorsements[n] = r.ProposalResponse.Endorsement
	}
	// create ChaincodeEndorsedAction
	cea := &pb.ChaincodeEndorsedAction{ProposalResponsePayload: resps[0].ProposalResponse.Payload, Endorsements: endorsements}

	// obtain the bytes of the proposal payload that will go to the transaction
	propPayloadBytes, err := protos_utils.GetBytesProposalPayloadForTx(pPayl, hdrExt.PayloadVisibility)
	if err != nil {
		return nil, err
	}

	// serialize the chaincode action payload
	cap := &pb.ChaincodeActionPayload{ChaincodeProposalPayload: propPayloadBytes, Action: cea}
	capBytes, err := protos_utils.GetBytesChaincodeActionPayload(cap)
	if err != nil {
		return nil, err
	}

	// create a transaction
	taa := &pb.TransactionAction{Header: hdr.SignatureHeader, Payload: capBytes}
	taas := make([]*pb.TransactionAction, 1)
	taas[0] = taa

	return &apitxn.Transaction{
		Transaction: &pb.Transaction{Actions: taas},
		Proposal:    proposal,
	}, nil
}

// SendTransaction send a transaction to the chain’s orderer service (one or more orderer endpoints) for consensus and committing to the ledger.
func (c *Channel) SendTransaction(tx *apitxn.Transaction) (*apitxn.TransactionResponse, error) {
	if c.orderers == nil || len(c.orderers) == 0 {
		return nil, fmt.Errorf("orderers is nil")
	}
	if tx == nil {
		return nil, fmt.Errorf("Transaction is nil")
	}
	if tx.Proposal == nil || tx.Proposal.Proposal == nil {
		return nil, fmt.Errorf("proposal is nil")
	}

	// the original header
	hdr, err := protos_utils.GetHeader(tx.Proposal.Proposal.Header)
	if err != nil {
		return nil, fmt.Errorf("Could not unmarshal the proposal header")
	}
	// serialize the tx
	txBytes, err := protos_utils.GetBytesTransaction(tx.Transaction)
	if err != nil {
		return nil, err
	}

	// create the payload
	payl := &common.Payload{Header: hdr, Data: txBytes}
	paylBytes, err := protos_utils.GetBytesPayload(payl)
	if err != nil {
		return nil, err
	}

	// here's the envelope
	envelope, err := c.SignPayload(paylBytes)
	if err != nil {
		return nil, err
	}

	transactionResponse, err := c.BroadcastEnvelope(envelope)
	if err != nil {
		return nil, err
	}

	return transactionResponse, nil
}

// SendInstantiateProposal sends an instantiate proposal to one or more endorsing peers.
// chaincodeName: required - The name of the chain.
// args: optional - string Array arguments specific to the chaincode being instantiated
// chaincodePath: required - string of the path to the location of the source code of the chaincode
// chaincodeVersion: required - string of the version of the chaincode
func (c *Channel) SendInstantiateProposal(chaincodeName string,
	args []string, chaincodePath string, chaincodeVersion string, targets []apitxn.ProposalProcessor) ([]*apitxn.TransactionProposalResponse, apitxn.TransactionID, error) {

	if chaincodeName == "" {
		return nil, apitxn.TransactionID{}, fmt.Errorf("Missing 'chaincodeName' parameter")
	}
	if chaincodePath == "" {
		return nil, apitxn.TransactionID{}, fmt.Errorf("Missing 'chaincodePath' parameter")
	}
	if chaincodeVersion == "" {

		return nil, apitxn.TransactionID{}, fmt.Errorf("Missing 'chaincodeVersion' parameter")
	}

	// TODO: We should validate that targets are added to the channel.
	if targets == nil || len(targets) < 1 {
		return nil, apitxn.TransactionID{}, fmt.Errorf("Missing peer objects for instantiate CC proposal")
	}

	argsArray := make([][]byte, len(args))
	for i, arg := range args {
		argsArray[i] = []byte(arg)
	}

	ccds := &pb.ChaincodeDeploymentSpec{ChaincodeSpec: &pb.ChaincodeSpec{
		Type: pb.ChaincodeSpec_GOLANG, ChaincodeId: &pb.ChaincodeID{Name: chaincodeName, Path: chaincodePath, Version: chaincodeVersion},
		Input: &pb.ChaincodeInput{Args: argsArray}}}

	if c.clientContext.UserContext() == nil {
		return nil, apitxn.TransactionID{}, fmt.Errorf("User context needs to be set")
	}
	creator, err := c.clientContext.UserContext().Identity()
	if err != nil {
		return nil, apitxn.TransactionID{}, fmt.Errorf("Error getting creator: %v", err)
	}
	chaincodePolicy, err := buildChaincodePolicy(c.clientContext.UserContext().MspID())
	if err != nil {
		return nil, apitxn.TransactionID{}, err
	}
	chaincodePolicyBytes, err := protos_utils.Marshal(chaincodePolicy)
	if err != nil {
		return nil, apitxn.TransactionID{}, err
	}
	// create a proposal from a chaincodeDeploymentSpec
	proposal, txID, err := protos_utils.CreateDeployProposalFromCDS(c.Name(), ccds, creator, chaincodePolicyBytes, []byte("escc"), []byte("vscc"))
	if err != nil {
		return nil, apitxn.TransactionID{}, fmt.Errorf("Could not create chaincode Deploy proposal, err %s", err)
	}

	signedProposal, err := c.signProposal(proposal)
	if err != nil {
		return nil, apitxn.TransactionID{}, err
	}

	txnID := apitxn.TransactionID{ID: txID} // Nonce is missing

	transactionProposalResponse, err := txnproc.SendTransactionProposalToProcessors(&apitxn.TransactionProposal{
		SignedProposal: signedProposal,
		Proposal:       proposal,
		TxnID:          txnID,
	}, targets)

	return transactionProposalResponse, txnID, err
}

// SignPayload ... TODO.
func (c *Channel) SignPayload(payload []byte) (*fab.SignedEnvelope, error) {
	//Get user info
	user := c.clientContext.UserContext()
	if user == nil {
		return nil, fmt.Errorf("User is nil")
	}

	signature, err := fc.SignObjectWithKey(payload, user.PrivateKey(),
		&bccsp.SHAOpts{}, nil, c.clientContext.CryptoSuite())
	if err != nil {
		return nil, err
	}
	// here's the envelope
	return &fab.SignedEnvelope{Payload: payload, Signature: signature}, nil
}

// BroadcastEnvelope will send the given envelope to some orderer, picking random endpoints
// until all are exhausted
func (c *Channel) BroadcastEnvelope(envelope *fab.SignedEnvelope) (*apitxn.TransactionResponse, error) {
	// Check if orderers are defined
	if len(c.orderers) == 0 {
		return nil, fmt.Errorf("orderers not set")
	}

	// Copy aside the ordering service endpoints
	orderers := []fab.Orderer{}
	for _, o := range c.orderers {
		orderers = append(orderers, o)
	}

	// Iterate them in a random order and try broadcasting 1 by 1
	var errResp *apitxn.TransactionResponse
	for _, i := range rand.Perm(len(orderers)) {
		resp := c.sendBroadcast(envelope, orderers[i])
		if resp.Err != nil {
			errResp = resp
		} else {
			return resp, nil
		}
	}
	return errResp, nil
}

func (c *Channel) sendBroadcast(envelope *fab.SignedEnvelope, orderer fab.Orderer) *apitxn.TransactionResponse {
	logger.Debugf("Broadcasting envelope to orderer :%s\n", orderer.URL())
	if _, err := orderer.SendBroadcast(envelope); err != nil {
		logger.Debugf("Receive Error Response from orderer :%v\n", err)
		return &apitxn.TransactionResponse{Orderer: orderer.URL(),
			Err: fmt.Errorf("Error calling orderer '%s':  %s", orderer.URL(), err)}
	}

	logger.Debugf("Receive Success Response from orderer\n")
	return &apitxn.TransactionResponse{Orderer: orderer.URL(), Err: nil}
}

// SendEnvelope sends the given envelope to each orderer and returns a block response
func (c *Channel) SendEnvelope(envelope *fab.SignedEnvelope) (*common.Block, error) {
	if c.orderers == nil || len(c.orderers) == 0 {
		return nil, fmt.Errorf("orderers not set")
	}

	var blockResponse *common.Block
	var errorResponse error
	var mutex sync.Mutex
	outstandingRequests := len(c.orderers)
	done := make(chan bool)

	// Send the request to all orderers and return as soon as one responds with a block.
	for _, o := range c.orderers {

		go func(orderer fab.Orderer) {
			logger.Debugf("Broadcasting envelope to orderer :%s\n", orderer.URL())

			blocks, errors := orderer.SendDeliver(envelope)
			select {
			case block := <-blocks:
				mutex.Lock()
				if blockResponse == nil {
					blockResponse = block
					done <- true
				}
				mutex.Unlock()

			case err := <-errors:
				mutex.Lock()
				if errorResponse == nil {
					errorResponse = err
				}
				outstandingRequests--
				if outstandingRequests == 0 {
					done <- true
				}
				mutex.Unlock()

			case <-time.After(time.Second * 5):
				mutex.Lock()
				if errorResponse == nil {
					errorResponse = fmt.Errorf("Timeout waiting for response from orderer")
				}
				outstandingRequests--
				if outstandingRequests == 0 {
					done <- true
				}
				mutex.Unlock()
			}
		}(o)
	}

	<-done

	if blockResponse != nil {
		return blockResponse, nil
	}

	// There must be an error
	if errorResponse != nil {
		return nil, fmt.Errorf("error returned from orderer service: %v", errorResponse)
	}

	return nil, fmt.Errorf("unexpected: didn't receive a block from any of the orderer servces and didn't receive any error")
}

// BuildChannelHeader is a utility method to build a common chain header (TODO refactor)
func BuildChannelHeader(headerType common.HeaderType, channelID string, txID string, epoch uint64, chaincodeID string, timestamp time.Time) (*common.ChannelHeader, error) {
	logger.Debugf("buildChannelHeader - headerType: %s channelID: %s txID: %d epoch: % chaincodeID: %s timestamp: %v", headerType, channelID, txID, epoch, chaincodeID, timestamp)
	channelHeader := &common.ChannelHeader{
		Type:      int32(headerType),
		Version:   1,
		ChannelId: channelID,
		TxId:      txID,
		Epoch:     epoch,
	}
	if !timestamp.IsZero() {
		ts := &proto_ts.Timestamp{
			Seconds: int64(timestamp.Second()),
			Nanos:   int32(timestamp.Nanosecond()),
		}
		channelHeader.Timestamp = ts
	}
	if chaincodeID != "" {
		ccID := &pb.ChaincodeID{
			Name: chaincodeID,
		}
		headerExt := &pb.ChaincodeHeaderExtension{
			ChaincodeId: ccID,
		}
		headerExtBytes, err := proto.Marshal(headerExt)
		if err != nil {
			return nil, fmt.Errorf("Error marshaling header extension: %v", err)
		}
		channelHeader.Extension = headerExtBytes
	}
	return channelHeader, nil
}

// internal utility method to build chaincode policy
// FIXME: for now always construct a 'Signed By any member of an organization by mspid' policy
func buildChaincodePolicy(mspid string) (*common.SignaturePolicyEnvelope, error) {
	// Define MSPRole
	memberRole, err := proto.Marshal(&mspprotos.MSPRole{Role: mspprotos.MSPRole_MEMBER, MspIdentifier: mspid})
	if err != nil {
		return nil, fmt.Errorf("Error marshal MSPRole: %s", err)
	}

	// construct a list of msp principals to select from using the 'n out of' operator
	onePrn := &mspprotos.MSPPrincipal{
		PrincipalClassification: mspprotos.MSPPrincipal_ROLE,
		Principal:               memberRole}

	// construct 'signed by msp principal at index 0'
	signedBy := &common.SignaturePolicy{Type: &common.SignaturePolicy_SignedBy{SignedBy: 0}}

	// construct 'one of one' policy
	oneOfone := &common.SignaturePolicy{Type: &common.SignaturePolicy_NOutOf_{NOutOf: &common.SignaturePolicy_NOutOf{
		N: 1, Rules: []*common.SignaturePolicy{signedBy}}}}

	p := &common.SignaturePolicyEnvelope{
		Version:    0,
		Rule:       oneOfone,
		Identities: []*mspprotos.MSPPrincipal{onePrn},
	}
	return p, nil
}
