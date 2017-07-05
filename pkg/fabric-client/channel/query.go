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

	fab "github.com/hyperledger/fabric-sdk-go/api/apifabclient"
	txn "github.com/hyperledger/fabric-sdk-go/api/apitxn"
)

// QueryInfo queries for various useful information on the state of the channel
// (height, known peers).
// This query will be made to the primary peer.
func (c *Channel) QueryInfo() (*common.BlockchainInfo, error) {
	logger.Debug("queryInfo - start")

	// prepare arguments to call qscc GetChainInfo function
	var args []string
	args = append(args, "GetChainInfo")
	args = append(args, c.Name())

	payload, err := c.queryByChaincodeByTarget("qscc", args, c.PrimaryPeer())
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
	args = append(args, "GetBlockByHash")
	args = append(args, c.Name())
	args = append(args, string(blockHash[:len(blockHash)]))

	payload, err := c.queryByChaincodeByTarget("qscc", args, c.PrimaryPeer())
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
	args = append(args, "GetBlockByNumber")
	args = append(args, c.Name())
	args = append(args, strconv.Itoa(blockNumber))

	payload, err := c.queryByChaincodeByTarget("qscc", args, c.PrimaryPeer())
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
func (c *Channel) QueryTransaction(transactionID string) (*pb.ProcessedTransaction, error) {

	// prepare arguments to call qscc GetTransactionByID function
	var args []string
	args = append(args, "GetTransactionByID")
	args = append(args, c.Name())
	args = append(args, transactionID)

	payload, err := c.queryByChaincodeByTarget("qscc", args, c.PrimaryPeer())
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

	payload, err := c.queryByChaincodeByTarget("lscc", []string{"getchaincodes"}, c.PrimaryPeer())
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

/**
* Generic helper for query functionality for chain
* This query will be made to one target peer and will return one result only.
* @parame {string} chaincode name
* @param {[]string} invoke arguments
* @param {Peer} target peer
* @returns {[]byte} payload
 */
func (c *Channel) queryByChaincodeByTarget(chaincodeName string, args []string, target fab.Peer) ([]byte, error) {
	return QueryChaincodeByTarget(chaincodeName, args, target, c.clientContext)
}

// QueryByChaincode sends a proposal to one or more endorsing peers that will be handled by the chaincode.
// This request will be presented to the chaincode 'invoke' and must understand
// from the arguments that this is a query request. The chaincode must also return
// results in the byte array format and the caller will have to be able to decode.
// these results.
// chaincodeName: chaincode name.
// args: invoke arguments
// targets: target peers
// Returns an array of payloads.
func (c *Channel) QueryByChaincode(chaincodeName string, args []string, targets []txn.ProposalProcessor) ([][]byte, error) {
	return queryChaincode(chaincodeName, args, targets, c.clientContext)
}

// QueryChaincodeByTarget ...
func QueryChaincodeByTarget(chaincodeName string, args []string, target txn.ProposalProcessor, clientContext fab.FabricClient) ([]byte, error) {
	queryResponses, err := queryChaincode(chaincodeName, args, []txn.ProposalProcessor{target}, clientContext)
	if err != nil {
		return nil, fmt.Errorf("QueryChaincodeByTarget return error: %v", err)
	}

	// we are only querying one peer hence one result
	if len(queryResponses) != 1 {
		return nil, fmt.Errorf("queryByChaincodeByTarget should have one result only - result number: %d", len(queryResponses))
	}

	return queryResponses[0], nil
}

func queryChaincode(chaincodeName string, args []string, targets []txn.ProposalProcessor, clientContext fab.FabricClient) ([][]byte, error) {
	if chaincodeName == "" {
		return nil, fmt.Errorf("Missing chaincode name")
	}

	if args == nil || len(args) < 1 {
		return nil, fmt.Errorf("Missing invoke arguments")
	}

	if targets == nil || len(targets) < 1 {
		return nil, fmt.Errorf("Missing target peers")
	}

	logger.Debugf("Calling %s function %v on targets: %s\n", chaincodeName, args[0], targets)

	signedProposal, err := createTransactionProposal(chaincodeName, "", args, true, nil, clientContext)
	if err != nil {
		return nil, fmt.Errorf("CreateTransactionProposal return error: %v", err)
	}

	transactionProposalResponses, err := SendTransactionProposal(signedProposal, 0, targets)
	if err != nil {
		return nil, fmt.Errorf("SendTransactionProposal return error: %v", err)
	}

	var responses [][]byte
	errMsg := ""
	for _, response := range transactionProposalResponses {
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
