/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package channel

import (
	"bytes"
	"net/http"
	"strconv"
	"strings"

	"github.com/golang/protobuf/proto"
	"github.com/pkg/errors"

	fab "github.com/hyperledger/fabric-sdk-go/api/apifabclient"
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
func NewLedger(ctx fab.Context, chName string) *Ledger {
	l := Ledger{
		ctx:    ctx,
		chName: chName,
	}
	return &l
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
	tprs, err := queryChaincode(c.ctx, systemChannel, request, targets)
	processed, err := processTxnProposalResponse(tprs, err, createBlockchainInfo)

	responses := []*common.BlockchainInfo{}
	for _, p := range processed {
		responses = append(responses, p.(*common.BlockchainInfo))
	}
	return responses, err
}

func createBlockchainInfo(tpr *fab.TransactionProposalResponse, err error) (interface{}, error) {
	response := common.BlockchainInfo{}
	if err != nil {
		// response had an error - do not process.
		return &response, err
	}

	err = proto.Unmarshal(tpr.ProposalResponse.GetResponse().Payload, &response)
	if err != nil {
		return nil, errors.Wrap(err, "unmarshal of transaction proposal response failed")
	}
	return &response, err
}

// QueryBlockByHash queries the ledger for Block by block hash.
// This query will be made to the primary peer.
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
	tprs, err := queryChaincode(c.ctx, systemChannel, request, targets)
	processed, err := processTxnProposalResponse(tprs, err, createCommonBlock)

	responses := []*common.Block{}
	for _, p := range processed {
		responses = append(responses, p.(*common.Block))
	}
	return responses, err
}

// QueryBlock queries the ledger for Block by block number.
// This query will be made to the primary peer.
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

	tprs, err := queryChaincode(c.ctx, systemChannel, request, targets)
	processed, err := processTxnProposalResponse(tprs, err, createCommonBlock)

	responses := []*common.Block{}
	for _, p := range processed {
		responses = append(responses, p.(*common.Block))
	}
	return responses, err
}

func createCommonBlock(tpr *fab.TransactionProposalResponse, err error) (interface{}, error) {
	response := common.Block{}
	if err != nil {
		// response had an error - do not process.
		return &response, err
	}

	err = proto.Unmarshal(tpr.ProposalResponse.GetResponse().Payload, &response)
	if err != nil {
		return nil, errors.Wrap(err, "unmarshal of transaction proposal response failed")
	}
	return &response, err
}

// QueryTransaction queries the ledger for Transaction by number.
// This query will be made to the primary peer.
// Returns the ProcessedTransaction information containing the transaction.
// TODO: add optional target
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

	tprs, err := queryChaincode(c.ctx, systemChannel, request, targets)
	processed, err := processTxnProposalResponse(tprs, err, createProcessedTransaction)

	responses := []*pb.ProcessedTransaction{}
	for _, p := range processed {
		responses = append(responses, p.(*pb.ProcessedTransaction))
	}
	return responses, err
}

func createProcessedTransaction(tpr *fab.TransactionProposalResponse, err error) (interface{}, error) {
	response := pb.ProcessedTransaction{}
	if err != nil {
		// response had an error - do not process.
		return &response, err
	}

	err = proto.Unmarshal(tpr.ProposalResponse.GetResponse().Payload, &response)
	if err != nil {
		return nil, errors.Wrap(err, "unmarshal of transaction proposal response failed")
	}
	return &response, err
}

// QueryInstantiatedChaincodes queries the instantiated chaincodes on this channel.
// This query will be made to the primary peer.
func (c *Ledger) QueryInstantiatedChaincodes(targets []fab.ProposalProcessor) ([]*pb.ChaincodeQueryResponse, error) {
	request := fab.ChaincodeInvokeRequest{
		ChaincodeID: "lscc",
		Fcn:         "getchaincodes",
	}

	tprs, err := queryChaincode(c.ctx, c.chName, request, targets)
	processed, err := processTxnProposalResponse(tprs, err, createChaincodeQueryResponse)

	responses := []*pb.ChaincodeQueryResponse{}
	for _, p := range processed {
		responses = append(responses, p.(*pb.ChaincodeQueryResponse))
	}
	return responses, err
}

func createChaincodeQueryResponse(tpr *fab.TransactionProposalResponse, err error) (interface{}, error) {
	response := pb.ChaincodeQueryResponse{}
	if err != nil {
		// response had an error - do not process.
		return &response, err
	}

	err = proto.Unmarshal(tpr.ProposalResponse.GetResponse().Payload, &response)
	if err != nil {
		return nil, errors.Wrap(err, "unmarshal of transaction proposal response failed")
	}
	return &response, err
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
	if err != nil {
		return nil, errors.WithMessage(err, "queryChaincode failed")
	}

	responses, err := filterProposalResponses(tpr, err)
	if err != nil {
		return nil, err
	}

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

type txnProposalResponseOp func(*fab.TransactionProposalResponse, error) (interface{}, error)

func processTxnProposalResponse(tprs []*fab.TransactionProposalResponse, tperr error, op txnProposalResponseOp) ([]interface{}, error) {

	// examine errors from peers and prepare error slice that can be checked during each response' processing.
	var errs MultiError
	if tperr != nil {
		var ok bool
		errs, ok = tperr.(MultiError)
		if !ok {
			return nil, errors.WithMessage(tperr, "chaincode query failed")
		}
	} else {
		errs = make([]error, len(tprs))
	}

	// process each response and set processing error, if needed.
	responses := []interface{}{}
	var resperrs MultiError
	isErr := false
	for i, tpr := range tprs {
		var resp interface{}
		var err error

		resp, err = op(tpr, errs[i])
		if err != nil {
			isErr = true
		}

		responses = append(responses, resp)
		resperrs = append(resperrs, err)
	}

	// when any error has occurred return responses and errors as a MultiError.
	if isErr {
		return responses, resperrs
	}
	return responses, nil
}

func filterProposalResponses(tprs []*fab.TransactionProposalResponse, tperr error) ([][]byte, error) {
	// examine errors from peers and prepare error slice that can be checked during each response' processing.
	var errs MultiError
	if tperr != nil {
		var ok bool
		errs, ok = tperr.(MultiError)
		if !ok {
			return nil, errors.WithMessage(tperr, "chaincode query failed")
		}
	} else {
		errs = make(MultiError, len(tprs))
	}

	responses := [][]byte{}
	errMsg := ""
	for i, tpr := range tprs {
		if errs[i] != nil {
			errMsg = errMsg + errs[i].Error() + "\n"
		} else {
			responses = append(responses, tpr.ProposalResponse.GetResponse().Payload)
		}
	}

	if len(errMsg) > 0 {
		return responses, errors.New(errMsg)
	}
	return responses, nil
}

// MultiError represents a slice of errors originating from each target peer.
type MultiError []error

func (me MultiError) Error() string {
	msg := []string{}
	for _, e := range me {
		msg = append(msg, e.Error())
	}
	return strings.Join(msg, ",")
}

func queryChaincode(ctx fab.Context, channel string, request fab.ChaincodeInvokeRequest, targets []fab.ProposalProcessor) ([]*fab.TransactionProposalResponse, error) {
	errors := MultiError{}
	responses := []*fab.TransactionProposalResponse{}
	isErr := false

	// TODO: this can be done concurrently.
	for _, target := range targets {
		resp, err := queryChaincodeWithTarget(ctx, channel, request, target)

		responses = append(responses, resp)
		errors = append(errors, err)

		if err != nil {
			isErr = true
		}
	}
	if isErr {
		return responses, errors
	}
	return responses, nil
}

func queryChaincodeWithTarget(ctx fab.Context, channel string, request fab.ChaincodeInvokeRequest, target fab.ProposalProcessor) (*fab.TransactionProposalResponse, error) {

	targets := []fab.ProposalProcessor{target}

	tp, err := txn.NewProposal(ctx, channel, request)
	if err != nil {
		return nil, errors.WithMessage(err, "NewProposal failed")
	}

	tpr, err := txn.SendProposal(tp, targets)
	if err != nil {
		return nil, errors.WithMessage(err, "SendProposal failed")
	}

	err = validateResponse(tpr[0])
	if err != nil {
		return nil, errors.WithMessage(err, "transaction proposal failed")
	}

	return tpr[0], nil
}

func validateResponse(response *fab.TransactionProposalResponse) error {
	if response.Err != nil {
		return errors.Errorf("error from %s (%s)", response.Endorser, response.Err.Error())
	}

	if response.Status != http.StatusOK {
		return errors.Errorf("bad status from %s (%d)", response.Endorser, response.Status)
	}

	return nil
}
