/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package channel

import (
	"net/http"

	"github.com/golang/protobuf/proto"
	"github.com/pkg/errors"

	"github.com/hyperledger/fabric-sdk-go/pkg/common/context"
	"github.com/hyperledger/fabric-sdk-go/pkg/context/api/fab"
	"github.com/hyperledger/fabric-sdk-go/pkg/errors/multi"
	"github.com/hyperledger/fabric-sdk-go/pkg/fab/txn"
	"github.com/hyperledger/fabric-sdk-go/pkg/logging"
	"github.com/hyperledger/fabric-sdk-go/third_party/github.com/hyperledger/fabric/protos/common"
	pb "github.com/hyperledger/fabric-sdk-go/third_party/github.com/hyperledger/fabric/protos/peer"
)

var logger = logging.NewLogger("fabsdk/fab")

const (
	lscc           = "lscc"
	lsccChaincodes = "getchaincodes"
)

// Ledger is a client that provides access to the underlying ledger of a channel.
type Ledger struct {
	ctx    context.Client
	chName string
}

// NewLedger constructs a Ledger client for the current context and named channel.
func NewLedger(ctx context.Client, chName string) (*Ledger, error) {
	l := Ledger{
		ctx:    ctx,
		chName: chName,
	}
	return &l, nil
}

// QueryInfo queries for various useful information on the state of the channel
// (height, known peers).
func (c *Ledger) QueryInfo(targets []fab.ProposalProcessor) ([]*fab.BlockchainInfoResponse, error) {
	logger.Debug("queryInfo - start")

	cir := createChannelInfoInvokeRequest(c.chName)
	tprs, errs := queryChaincode(c.ctx, c.chName, cir, targets)

	responses := []*fab.BlockchainInfoResponse{}
	for _, tpr := range tprs {
		r, err := createBlockchainInfo(tpr)
		if err != nil {
			errs = multi.Append(errs, errors.WithMessage(err, "From target: "+tpr.Endorser))
		} else {
			responses = append(responses, &fab.BlockchainInfoResponse{Endorser: tpr.Endorser, Status: tpr.Status, BCI: r})
		}
	}
	return responses, errs
}

func createBlockchainInfo(tpr *fab.TransactionProposalResponse) (*common.BlockchainInfo, error) {
	response := common.BlockchainInfo{}
	err := proto.Unmarshal(tpr.ProposalResponse.GetResponse().Payload, &response)
	if err != nil {
		return nil, errors.Wrap(err, "unmarshal of transaction proposal response failed")
	}
	return &response, nil
}

// QueryBlockByHash queries the ledger for Block by block hash.
// This query will be made to specified targets.
// Returns the block.
func (c *Ledger) QueryBlockByHash(blockHash []byte, targets []fab.ProposalProcessor) ([]*common.Block, error) {

	if blockHash == nil {
		return nil, errors.New("blockHash is required")
	}

	cir := createBlockByHashInvokeRequest(c.chName, blockHash)
	tprs, errs := queryChaincode(c.ctx, c.chName, cir, targets)

	responses := []*common.Block{}
	for _, tpr := range tprs {
		r, err := createCommonBlock(tpr)
		if err != nil {
			errs = multi.Append(errs, errors.WithMessage(err, "From target: "+tpr.Endorser))
		} else {
			responses = append(responses, r)
		}
	}
	return responses, errs
}

// QueryBlock queries the ledger for Block by block number.
// This query will be made to specified targets.
// blockNumber: The number which is the ID of the Block.
// It returns the block.
func (c *Ledger) QueryBlock(blockNumber int, targets []fab.ProposalProcessor) ([]*common.Block, error) {

	if blockNumber < 0 {
		return nil, errors.New("blockNumber must be a positive integer")
	}

	cir := createBlockByNumberInvokeRequest(c.chName, blockNumber)
	tprs, errs := queryChaincode(c.ctx, c.chName, cir, targets)

	responses := []*common.Block{}
	for _, tpr := range tprs {
		r, err := createCommonBlock(tpr)
		if err != nil {
			errs = multi.Append(errs, errors.WithMessage(err, "From target: "+tpr.Endorser))
		} else {
			responses = append(responses, r)
		}
	}
	return responses, errs
}

func createCommonBlock(tpr *fab.TransactionProposalResponse) (*common.Block, error) {
	response := common.Block{}
	err := proto.Unmarshal(tpr.ProposalResponse.GetResponse().Payload, &response)
	if err != nil {
		return nil, errors.Wrap(err, "unmarshal of transaction proposal response failed")
	}
	return &response, err
}

// QueryTransaction queries the ledger for Transaction by number.
// This query will be made to specified targets.
// Returns the ProcessedTransaction information containing the transaction.
func (c *Ledger) QueryTransaction(transactionID fab.TransactionID, targets []fab.ProposalProcessor) ([]*pb.ProcessedTransaction, error) {

	cir := createTransactionByIDInvokeRequest(c.chName, transactionID)
	tprs, errs := queryChaincode(c.ctx, c.chName, cir, targets)

	responses := []*pb.ProcessedTransaction{}
	for _, tpr := range tprs {
		r, err := createProcessedTransaction(tpr)
		if err != nil {
			errs = multi.Append(errs, errors.WithMessage(err, "From target: "+tpr.Endorser))
		} else {
			responses = append(responses, r)
		}
	}

	return responses, errs
}

func createProcessedTransaction(tpr *fab.TransactionProposalResponse) (*pb.ProcessedTransaction, error) {
	response := pb.ProcessedTransaction{}
	err := proto.Unmarshal(tpr.ProposalResponse.GetResponse().Payload, &response)
	if err != nil {
		return nil, errors.Wrap(err, "unmarshal of transaction proposal response failed")
	}
	return &response, err
}

// QueryInstantiatedChaincodes queries the instantiated chaincodes on this channel.
// This query will be made to specified targets.
func (c *Ledger) QueryInstantiatedChaincodes(targets []fab.ProposalProcessor) ([]*pb.ChaincodeQueryResponse, error) {
	cir := createChaincodeInvokeRequest()
	tprs, errs := queryChaincode(c.ctx, c.chName, cir, targets)

	responses := []*pb.ChaincodeQueryResponse{}
	for _, tpr := range tprs {
		r, err := createChaincodeQueryResponse(tpr)
		if err != nil {
			errs = multi.Append(errs, errors.WithMessage(err, "From target: "+tpr.Endorser))
		} else {
			responses = append(responses, r)
		}
	}
	return responses, errs
}

func createChaincodeQueryResponse(tpr *fab.TransactionProposalResponse) (*pb.ChaincodeQueryResponse, error) {
	response := pb.ChaincodeQueryResponse{}
	err := proto.Unmarshal(tpr.ProposalResponse.GetResponse().Payload, &response)
	if err != nil {
		return nil, errors.Wrap(err, "unmarshal of transaction proposal response failed")
	}
	return &response, nil
}

// QueryConfigBlock returns the current configuration block for the specified channel. If the
// peer doesn't belong to the channel, return error
func (c *Ledger) QueryConfigBlock(targets []fab.ProposalProcessor, minResponses int) (*common.ConfigEnvelope, error) {

	if len(targets) == 0 {
		return nil, errors.New("target(s) required")
	}

	if minResponses <= 0 {
		return nil, errors.New("Minimum endorser has to be greater than zero")
	}

	cir := createConfigBlockInvokeRequest(c.chName)
	tprs, err := queryChaincode(c.ctx, c.chName, cir, targets)
	if err != nil && len(tprs) == 0 {
		return nil, errors.WithMessage(err, "queryChaincode failed")
	}

	if len(tprs) < minResponses {
		return nil, errors.Errorf("Required minimum %d endorsments got %d", minResponses, len(tprs))
	}

	block, err := createCommonBlock(tprs[0])
	if err != nil {
		return nil, err
	}

	// Compare block data from  remaining responses
	for _, tpr := range tprs[1:] {
		b, err := createCommonBlock(tpr)
		if err != nil {
			return nil, err
		}

		if !proto.Equal(block.Data, b.Data) {
			return nil, errors.New("Payloads for config block do not match")
		}
	}

	if block.Data == nil || block.Data.Data == nil {
		return nil, errors.New("config block data is nil")
	}

	if len(block.Data.Data) != 1 {
		return nil, errors.New("config block must contain one transaction")
	}

	return createConfigEnvelope(block.Data.Data[0])

}

func collectProposalResponses(tprs []*fab.TransactionProposalResponse) [][]byte {
	responses := [][]byte{}
	for _, tpr := range tprs {
		responses = append(responses, tpr.ProposalResponse.GetResponse().Payload)
	}

	return responses
}

func queryChaincode(ctx context.Client, channelID string, request fab.ChaincodeInvokeRequest, targets []fab.ProposalProcessor) ([]*fab.TransactionProposalResponse, error) {
	txh, err := txn.NewHeader(ctx, channelID)
	if err != nil {
		return nil, errors.WithMessage(err, "creation of transaction ID failed")
	}

	tp, err := txn.CreateChaincodeInvokeProposal(txh, request)
	if err != nil {
		return nil, errors.WithMessage(err, "NewProposal failed")
	}
	tprs, errs := txn.SendProposal(ctx, tp, targets)

	return filterResponses(tprs, errs)
}

func filterResponses(responses []*fab.TransactionProposalResponse, errs error) ([]*fab.TransactionProposalResponse, error) {
	filteredResponses := responses[:0]
	for _, response := range responses {
		if response.Status == http.StatusOK {
			filteredResponses = append(filteredResponses, response)
		} else {
			errs = multi.Append(errs, errors.Errorf("bad status from %s (%d)", response.Endorser, response.Status))
		}
	}

	return filteredResponses, errs
}

func createChaincodeInvokeRequest() fab.ChaincodeInvokeRequest {
	cir := fab.ChaincodeInvokeRequest{
		ChaincodeID: lscc,
		Fcn:         lsccChaincodes,
	}
	return cir
}

func createConfigEnvelope(data []byte) (*common.ConfigEnvelope, error) {

	envelope := &common.Envelope{}
	if err := proto.Unmarshal(data, envelope); err != nil {
		return nil, errors.Wrap(err, "unmarshal envelope from config block failed")
	}
	payload := &common.Payload{}
	if err := proto.Unmarshal(envelope.Payload, payload); err != nil {
		return nil, errors.Wrap(err, "unmarshal payload from envelope failed")
	}
	channelHeader := &common.ChannelHeader{}
	if err := proto.Unmarshal(payload.Header.ChannelHeader, channelHeader); err != nil {
		return nil, errors.Wrap(err, "unmarshal payload from envelope failed")
	}
	if common.HeaderType(channelHeader.Type) != common.HeaderType_CONFIG {
		return nil, errors.New("block must be of type 'CONFIG'")
	}
	configEnvelope := &common.ConfigEnvelope{}
	if err := proto.Unmarshal(payload.Data, configEnvelope); err != nil {
		return nil, errors.Wrap(err, "unmarshal config envelope failed")
	}

	return configEnvelope, nil
}
