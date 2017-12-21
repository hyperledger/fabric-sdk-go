/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package channel

import (
	"math/rand"
	"sync"
	"time"

	"github.com/golang/protobuf/proto"
	"github.com/golang/protobuf/ptypes"
	fab "github.com/hyperledger/fabric-sdk-go/api/apifabclient"
	"github.com/hyperledger/fabric-sdk-go/api/apitxn"
	"github.com/hyperledger/fabric-sdk-go/pkg/errors"

	"github.com/hyperledger/fabric-sdk-go/api/apiconfig"
	protos_utils "github.com/hyperledger/fabric-sdk-go/internal/github.com/hyperledger/fabric/protos/utils"
	"github.com/hyperledger/fabric-sdk-go/pkg/fabric-client/internal/txnproc"
	"github.com/hyperledger/fabric-sdk-go/third_party/github.com/hyperledger/fabric/protos/common"
	pb "github.com/hyperledger/fabric-sdk-go/third_party/github.com/hyperledger/fabric/protos/peer"
)

func init() {
	rand.Seed(time.Now().UnixNano())
}

// CCProposalType reflects transitions in the chaincode lifecycle
type CCProposalType int

// Define chaincode proposal types
const (
	Instantiate CCProposalType = iota
	Upgrade
)

// CreateTransaction create a transaction with proposal response, following the endorsement policy.
func (c *Channel) CreateTransaction(resps []*apitxn.TransactionProposalResponse) (*apitxn.Transaction, error) {
	if len(resps) == 0 {
		return nil, errors.New("at least one proposal response is necessary")
	}

	proposal := &resps[0].Proposal

	// the original header
	hdr, err := protos_utils.GetHeader(proposal.Proposal.Header)
	if err != nil {
		return nil, errors.Wrap(err, "unmarshal proposal header failed")
	}

	// the original payload
	pPayl, err := protos_utils.GetChaincodeProposalPayload(proposal.Proposal.Payload)
	if err != nil {
		return nil, errors.Wrap(err, "unmarshal proposal payload failed")
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
	//                                                            return nil, errors.Errorf("proposal response was not successful, error code %d, msg %s", r.Response.Status, r.Response.Message)
	//                                            }
	//                                            continue
	//                            }

	//                            if bytes.Compare(a1, r.Payload) != 0 {
	//                                            return nil, errors.New("ProposalResponsePayloads do not match")
	//                            }
	//            }

	for _, r := range resps {
		if r.ProposalResponse.Response.Status != 200 {
			return nil, errors.Errorf("proposal response was not successful, error code %d, msg %s", r.ProposalResponse.Response.Status, r.ProposalResponse.Response.Message)
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
		return nil, errors.New("orderers is nil")
	}
	if tx == nil {
		return nil, errors.New("transaction is nil")
	}
	if tx.Proposal == nil || tx.Proposal.Proposal == nil {
		return nil, errors.New("proposal is nil")
	}

	// the original header
	hdr, err := protos_utils.GetHeader(tx.Proposal.Proposal.Header)
	if err != nil {
		return nil, errors.Wrap(err, "unmarshal proposal header failed")
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
	args [][]byte, chaincodePath string, chaincodeVersion string,
	chaincodePolicy *common.SignaturePolicyEnvelope, targets []apitxn.ProposalProcessor) ([]*apitxn.TransactionProposalResponse, apitxn.TransactionID, error) {

	return c.sendCCProposal(Instantiate, chaincodeName, args, chaincodePath, chaincodeVersion, chaincodePolicy, targets)

}

// SendUpgradeProposal sends an upgrade proposal to one or more endorsing peers.
// chaincodeName: required - The name of the chain.
// args: optional - string Array arguments specific to the chaincode being upgraded
// chaincodePath: required - string of the path to the location of the source code of the chaincode
// chaincodeVersion: required - string of the version of the chaincode
func (c *Channel) SendUpgradeProposal(chaincodeName string,
	args [][]byte, chaincodePath string, chaincodeVersion string,
	chaincodePolicy *common.SignaturePolicyEnvelope, targets []apitxn.ProposalProcessor) ([]*apitxn.TransactionProposalResponse, apitxn.TransactionID, error) {

	return c.sendCCProposal(Upgrade, chaincodeName, args, chaincodePath, chaincodeVersion, chaincodePolicy, targets)

}

// helper function that sends an instantiate or upgrade chaincode proposal to one or more endorsing peers
func (c *Channel) sendCCProposal(ccProposalType CCProposalType, chaincodeName string,
	args [][]byte, chaincodePath string, chaincodeVersion string,
	chaincodePolicy *common.SignaturePolicyEnvelope, targets []apitxn.ProposalProcessor) ([]*apitxn.TransactionProposalResponse, apitxn.TransactionID, error) {

	if chaincodeName == "" {
		return nil, apitxn.TransactionID{}, errors.New("chaincodeName is required")
	}
	if chaincodePath == "" {
		return nil, apitxn.TransactionID{}, errors.New("chaincodePath is required")
	}
	if chaincodeVersion == "" {
		return nil, apitxn.TransactionID{}, errors.New("chaincodeVersion is required")
	}
	if chaincodePolicy == nil {
		return nil, apitxn.TransactionID{}, errors.New("chaincodePolicy is required")
	}

	if targets == nil || len(targets) < 1 {
		return nil, apitxn.TransactionID{}, errors.New("missing peer objects for chaincode proposal")
	}

	ccds := &pb.ChaincodeDeploymentSpec{ChaincodeSpec: &pb.ChaincodeSpec{
		Type: pb.ChaincodeSpec_GOLANG, ChaincodeId: &pb.ChaincodeID{Name: chaincodeName, Path: chaincodePath, Version: chaincodeVersion},
		Input: &pb.ChaincodeInput{Args: args}}}

	if c.clientContext.UserContext() == nil {
		return nil, apitxn.TransactionID{}, errors.New("user context is nil")
	}
	creator, err := c.clientContext.UserContext().Identity()
	if err != nil {
		return nil, apitxn.TransactionID{}, errors.Wrap(err, "getting user context's identity failed")
	}
	chaincodePolicyBytes, err := protos_utils.Marshal(chaincodePolicy)
	if err != nil {
		return nil, apitxn.TransactionID{}, err
	}

	var proposal *pb.Proposal
	var txID string

	switch ccProposalType {

	case Instantiate:
		proposal, txID, err = protos_utils.CreateDeployProposalFromCDS(c.Name(), ccds, creator, chaincodePolicyBytes, []byte("escc"), []byte("vscc"))
		if err != nil {
			return nil, apitxn.TransactionID{}, errors.Wrap(err, "create instantiate chaincode proposal failed")
		}
	case Upgrade:
		proposal, txID, err = protos_utils.CreateUpgradeProposalFromCDS(c.Name(), ccds, creator, chaincodePolicyBytes, []byte("escc"), []byte("vscc"))
		if err != nil {
			return nil, apitxn.TransactionID{}, errors.Wrap(err, "create  upgrade chaincode proposal failed")
		}
	default:
		return nil, apitxn.TransactionID{}, errors.Errorf("chaincode proposal type %d not supported", ccProposalType)
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

// SignPayload signs payload
func (c *Channel) SignPayload(payload []byte) (*fab.SignedEnvelope, error) {
	//Get user info
	user := c.clientContext.UserContext()
	if user == nil {
		return nil, errors.New("user is nil")
	}

	signingMgr := c.clientContext.SigningManager()
	if signingMgr == nil {
		return nil, errors.New("signing manager is nil")
	}

	signature, err := signingMgr.Sign(payload, user.PrivateKey())
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
		return nil, errors.New("orderers not set")
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
			Err: errors.Wrapf(err, "calling orderer '%s' failed", orderer.URL())}
	}

	logger.Debugf("Receive Success Response from orderer\n")
	return &apitxn.TransactionResponse{Orderer: orderer.URL(), Err: nil}
}

// SendEnvelope sends the given envelope to each orderer and returns a block response
func (c *Channel) SendEnvelope(envelope *fab.SignedEnvelope) (*common.Block, error) {
	if c.orderers == nil || len(c.orderers) == 0 {
		return nil, errors.New("orderers not set")
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

			blocks, errs := orderer.SendDeliver(envelope)
			select {
			case block := <-blocks:
				mutex.Lock()
				if blockResponse == nil {
					blockResponse = block
					done <- true
				}
				mutex.Unlock()

			case err := <-errs:
				mutex.Lock()
				if errorResponse == nil {
					errorResponse = err
				}
				outstandingRequests--
				if outstandingRequests == 0 {
					done <- true
				}
				mutex.Unlock()

			case <-time.After(c.ClientContext().Config().TimeoutOrDefault(apiconfig.OrdererResponse)):
				mutex.Lock()
				if errorResponse == nil {
					errorResponse = errors.New("timeout waiting for response from orderer")
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
		return nil, errors.Wrap(errorResponse, "error returned from orderer service")
	}

	return nil, errors.New("unexpected: didn't receive a block from any of the orderer servces and didn't receive any error")
}

// BuildChannelHeader is a utility method to build a common chain header (TODO refactor)
func BuildChannelHeader(headerType common.HeaderType, channelID string, txID string, epoch uint64, chaincodeID string, timestamp time.Time, tlsCertHash []byte) (*common.ChannelHeader, error) {
	logger.Debugf("buildChannelHeader - headerType: %s channelID: %s txID: %d epoch: % chaincodeID: %s timestamp: %v", headerType, channelID, txID, epoch, chaincodeID, timestamp)
	channelHeader := &common.ChannelHeader{
		Type:        int32(headerType),
		Version:     1,
		ChannelId:   channelID,
		TxId:        txID,
		Epoch:       epoch,
		TlsCertHash: tlsCertHash,
	}

	ts, err := ptypes.TimestampProto(timestamp)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create timestamp in channel header")
	}
	channelHeader.Timestamp = ts

	if chaincodeID != "" {
		ccID := &pb.ChaincodeID{
			Name: chaincodeID,
		}
		headerExt := &pb.ChaincodeHeaderExtension{
			ChaincodeId: ccID,
		}
		headerExtBytes, err := proto.Marshal(headerExt)
		if err != nil {
			return nil, errors.Wrap(err, "marshal header extension failed")
		}
		channelHeader.Extension = headerExtBytes
	}
	return channelHeader, nil
}
