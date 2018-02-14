/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package channel

import (
	"bytes"
	"net/http"
	"strconv"

	"github.com/golang/protobuf/proto"
	"github.com/pkg/errors"

	fab "github.com/hyperledger/fabric-sdk-go/api/apifabclient"
	"github.com/hyperledger/fabric-sdk-go/pkg/errors/multi"
	"github.com/hyperledger/fabric-sdk-go/pkg/fabric-client/txn"
	"github.com/hyperledger/fabric-sdk-go/third_party/github.com/hyperledger/fabric/protos/common"
	pb "github.com/hyperledger/fabric-sdk-go/third_party/github.com/hyperledger/fabric/protos/peer"
)

const (
	systemChannel = ""
)

// Ledger is a client that provides access to the underlying ledger of a channel.
type Ledger struct {
	ctx    fab.Context
	chName string
}

// NewLedger constructs a Ledger client for the current context and named channel.
func NewLedger(ctx fab.Context, chName string) (*Ledger, error) {
	l := Ledger{
		ctx:    ctx,
		chName: chName,
	}
	return &l, nil
}

// QueryInfo queries for various useful information on the state of the channel
// (height, known peers).
func (c *Ledger) QueryInfo(targets []fab.ProposalProcessor) ([]*common.BlockchainInfo, error) {
	logger.Debug("queryInfo - start")

	// prepare arguments to call qscc GetChainInfo function
	var args [][]byte
	args = append(args, []byte(c.chName))

	request := fab.ChaincodeInvokeRequest{
		ChaincodeID: "qscc",
		Fcn:         "GetChainInfo",
		Args:        args,
	}
	tprs, errs := queryChaincode(c.ctx, systemChannel, request, targets)

	responses := []*common.BlockchainInfo{}
	for _, tpr := range tprs {
		r, err := createBlockchainInfo(tpr)
		if err != nil {
			errs = multi.Append(errs, errors.WithMessage(err, "From target: "+tpr.Endorser))
		} else {
			responses = append(responses, r)
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

	// prepare arguments to call qscc GetBlockByNumber function
	var args [][]byte
	args = append(args, []byte(c.chName))
	args = append(args, blockHash[:len(blockHash)])

	request := fab.ChaincodeInvokeRequest{
		ChaincodeID: "qscc",
		Fcn:         "GetBlockByHash",
		Args:        args,
	}
	tprs, errs := queryChaincode(c.ctx, systemChannel, request, targets)

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

	// prepare arguments to call qscc GetBlockByNumber function
	var args [][]byte
	args = append(args, []byte(c.chName))
	args = append(args, []byte(strconv.Itoa(blockNumber)))

	request := fab.ChaincodeInvokeRequest{
		ChaincodeID: "qscc",
		Fcn:         "GetBlockByNumber",
		Args:        args,
	}

	tprs, errs := queryChaincode(c.ctx, systemChannel, request, targets)

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
func (c *Ledger) QueryTransaction(transactionID string, targets []fab.ProposalProcessor) ([]*pb.ProcessedTransaction, error) {

	// prepare arguments to call qscc GetTransactionByID function
	var args [][]byte
	args = append(args, []byte(c.chName))
	args = append(args, []byte(transactionID))

	request := fab.ChaincodeInvokeRequest{
		ChaincodeID: "qscc",
		Fcn:         "GetTransactionByID",
		Args:        args,
	}

	tprs, errs := queryChaincode(c.ctx, systemChannel, request, targets)

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
	request := fab.ChaincodeInvokeRequest{
		ChaincodeID: "lscc",
		Fcn:         "getchaincodes",
	}

	tprs, errs := queryChaincode(c.ctx, c.chName, request, targets)

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
func (c *Ledger) QueryConfigBlock(peers []fab.Peer, minResponses int) (*common.ConfigEnvelope, error) {

	if len(peers) == 0 {
		return nil, errors.New("peer(s) required")
	}

	if minResponses <= 0 {
		return nil, errors.New("Minimum endorser has to be greater than zero")
	}

	request := fab.ChaincodeInvokeRequest{
		ChaincodeID: "cscc",
		Fcn:         "GetConfigBlock",
		Args:        [][]byte{[]byte(c.chName)},
	}
	tpr, err := queryChaincode(c.ctx, c.chName, request, peersToTxnProcessors(peers))
	if err != nil && len(tpr) == 0 {
		return nil, errors.WithMessage(err, "queryChaincode failed")
	}

	responses := collectProposalResponses(tpr)

	if len(responses) < minResponses {
		return nil, errors.Errorf("Required minimum %d endorsments got %d", minResponses, len(responses))
	}

	r := responses[0]
	for _, p := range responses {
		if bytes.Compare(r, p) != 0 {
			return nil, errors.New("Payloads for config block do not match")
		}
	}

	block := &common.Block{}
	err = proto.Unmarshal(responses[0], block)
	if err != nil {
		return nil, errors.Wrap(err, "unmarshal block failed")
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

func queryChaincode(ctx fab.Context, channel string, request fab.ChaincodeInvokeRequest, targets []fab.ProposalProcessor) ([]*fab.TransactionProposalResponse, error) {
	tp, err := txn.NewProposal(ctx, channel, request)
	if err != nil {
		return nil, errors.WithMessage(err, "NewProposal failed")
	}
	tprs, errs := txn.SendProposal(tp, targets)

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
