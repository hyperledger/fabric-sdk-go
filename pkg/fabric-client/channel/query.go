/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package channel

import (
	"fmt"
	"strconv"

	"github.com/golang/protobuf/proto"

	"github.com/hyperledger/fabric/protos/common"
	pb "github.com/hyperledger/fabric/protos/peer"

	txn "github.com/hyperledger/fabric-sdk-go/api/apitxn"
)

const (
	systemChannel = ""
)

// QueryInfo queries for various useful information on the state of the channel
// (height, known peers).
// This query will be made to the primary peer.
func (c *Channel) QueryInfo() (*common.BlockchainInfo, error) {
	logger.Debug("queryInfo - start")

	// prepare arguments to call qscc GetChainInfo function
	var args []string
	args = append(args, c.Name())

	payload, err := c.queryBySystemChaincodeByTarget("qscc", "GetChainInfo", args, c.PrimaryPeer())
	if err != nil {
		return nil, fmt.Errorf("Invoke qscc GetChainInfo return error: %v", err)
	}

	bci := &common.BlockchainInfo{}
	err = proto.Unmarshal(payload, bci)
	if err != nil {
		return nil, fmt.Errorf("Unmarshal BlockchainInfo return error: %v", err)
	}

	return bci, nil
}

// QueryBlockByHash queries the ledger for Block by block hash.
// This query will be made to the primary peer.
// Returns the block.
func (c *Channel) QueryBlockByHash(blockHash []byte) (*common.Block, error) {

	if blockHash == nil {
		return nil, fmt.Errorf("Blockhash bytes are required")
	}

	// prepare arguments to call qscc GetBlockByNumber function
	var args []string
	args = append(args, c.Name())
	args = append(args, string(blockHash[:len(blockHash)]))

	payload, err := c.queryBySystemChaincodeByTarget("qscc", "GetBlockByHash", args, c.PrimaryPeer())
	if err != nil {
		return nil, fmt.Errorf("Invoke qscc GetBlockByHash return error: %v", err)
	}

	block := &common.Block{}
	err = proto.Unmarshal(payload, block)
	if err != nil {
		return nil, fmt.Errorf("Unmarshal Block return error: %v", err)
	}

	return block, nil
}

// QueryBlock queries the ledger for Block by block number.
// This query will be made to the primary peer.
// blockNumber: The number which is the ID of the Block.
// It returns the block.
func (c *Channel) QueryBlock(blockNumber int) (*common.Block, error) {

	if blockNumber < 0 {
		return nil, fmt.Errorf("Block number must be positive integer")
	}

	// prepare arguments to call qscc GetBlockByNumber function
	var args []string
	args = append(args, c.Name())
	args = append(args, strconv.Itoa(blockNumber))

	payload, err := c.queryBySystemChaincodeByTarget("qscc", "GetBlockByNumber", args, c.PrimaryPeer())
	if err != nil {
		return nil, fmt.Errorf("Invoke qscc GetBlockByNumber return error: %v", err)
	}

	block := &common.Block{}
	err = proto.Unmarshal(payload, block)
	if err != nil {
		return nil, fmt.Errorf("Unmarshal Block return error: %v", err)
	}

	return block, nil
}

// QueryTransaction queries the ledger for Transaction by number.
// This query will be made to the primary peer.
// Returns the ProcessedTransaction information containing the transaction.
// TODO: add optional target
func (c *Channel) QueryTransaction(transactionID string) (*pb.ProcessedTransaction, error) {

	// prepare arguments to call qscc GetTransactionByID function
	var args []string
	args = append(args, c.Name())
	args = append(args, transactionID)

	payload, err := c.queryBySystemChaincodeByTarget("qscc", "GetTransactionByID", args, c.PrimaryPeer())
	if err != nil {
		return nil, fmt.Errorf("Invoke qscc GetBlockByNumber return error: %v", err)
	}

	transaction := new(pb.ProcessedTransaction)
	err = proto.Unmarshal(payload, transaction)
	if err != nil {
		return nil, fmt.Errorf("Unmarshal ProcessedTransaction return error: %v", err)
	}

	return transaction, nil
}

// QueryInstantiatedChaincodes queries the instantiated chaincodes on this channel.
// This query will be made to the primary peer.
func (c *Channel) QueryInstantiatedChaincodes() (*pb.ChaincodeQueryResponse, error) {

	var args []string

	payload, err := c.queryBySystemChaincodeByTarget("lscc", "getchaincodes", args, c.PrimaryPeer())
	if err != nil {
		return nil, fmt.Errorf("Invoke lscc getchaincodes return error: %v", err)
	}

	response := new(pb.ChaincodeQueryResponse)
	err = proto.Unmarshal(payload, response)
	if err != nil {
		return nil, fmt.Errorf("Unmarshal ChaincodeQueryResponse return error: %v", err)
	}

	return response, nil
}

// QueryByChaincode sends a proposal to one or more endorsing peers that will be handled by the chaincode.
// This request will be presented to the chaincode 'invoke' and must understand
// from the arguments that this is a query request. The chaincode must also return
// results in the byte array format and the caller will have to be able to decode.
// these results.
func (c *Channel) QueryByChaincode(request txn.ChaincodeInvokeRequest) ([][]byte, error) {
	request, err := c.chaincodeInvokeRequestAddDefaultPeers(request)
	if err != nil {
		return nil, err
	}
	return queryByChaincode(c.name, request, c.clientContext)
}

func filterProposalResponses(tpr []*txn.TransactionProposalResponse) ([][]byte, error) {
	var responses [][]byte
	errMsg := ""
	for _, response := range tpr {
		if response.Err != nil {
			errMsg = errMsg + response.Err.Error() + "\n"
		} else {
			responses = append(responses, response.ProposalResponse.GetResponse().Payload)
		}
	}

	if len(errMsg) > 0 {
		return responses, fmt.Errorf(errMsg)
	}
	return responses, nil
}

func queryByChaincode(channelID string, request txn.ChaincodeInvokeRequest, clientContext ClientContext) ([][]byte, error) {
	if err := validateChaincodeInvokeRequest(request); err != nil {
		return nil, err
	}

	transactionProposalResponses, _, err := sendTransactionProposal(channelID, request, clientContext)
	if err != nil {
		return nil, fmt.Errorf("SendTransactionProposal return error: %v", err)
	}

	return filterProposalResponses(transactionProposalResponses)
}

// queryBySystemChaincodeByTarget is an internal helper function that queries system chaincode.
// This function is not exported to keep the external interface of this package to only expose
// request structs.
func (c *Channel) queryBySystemChaincodeByTarget(chaincodeID string, fcn string, args []string, target txn.ProposalProcessor) ([]byte, error) {
	targets := []txn.ProposalProcessor{target}
	request := txn.ChaincodeInvokeRequest{
		ChaincodeID: chaincodeID,
		Fcn:         fcn,
		Args:        args,
		Targets:     targets,
	}
	responses, err := c.QueryBySystemChaincode(request)

	// we are only querying one peer hence one result
	if err != nil || len(responses) != 1 {
		return nil, fmt.Errorf("QueryBySystemChaincode should have one result only - result number: %d", len(responses))
	}

	return responses[0], nil
}

// QueryBySystemChaincode invokes a system chaincode
func (c *Channel) QueryBySystemChaincode(request txn.ChaincodeInvokeRequest) ([][]byte, error) {
	request, err := c.chaincodeInvokeRequestAddDefaultPeers(request)
	if err != nil {
		return nil, err
	}
	return queryByChaincode(systemChannel, request, c.clientContext)
}

// QueryBySystemChaincode invokes a system chaincode
// TODO - should be moved.
func QueryBySystemChaincode(request txn.ChaincodeInvokeRequest, clientContext ClientContext) ([][]byte, error) {
	return queryByChaincode(systemChannel, request, clientContext)
}
