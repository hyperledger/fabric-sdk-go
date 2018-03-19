/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package channel

import (
	"strconv"

	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/fab"
)

const (
	qscc                = "qscc"
	qsccTransactionByID = "GetTransactionByID"
	qsccChannelInfo     = "GetChainInfo"
	qsccBlockByHash     = "GetBlockByHash"
	qsccBlockByNumber   = "GetBlockByNumber"
	qsccBlockByTxID     = "GetBlockByTxID"
)

func createTransactionByIDInvokeRequest(channelID string, transactionID fab.TransactionID) fab.ChaincodeInvokeRequest {
	var args [][]byte
	args = append(args, []byte(channelID))
	args = append(args, []byte(transactionID))

	cir := fab.ChaincodeInvokeRequest{
		ChaincodeID: qscc,
		Fcn:         qsccTransactionByID,
		Args:        args,
	}
	return cir
}

func createChannelInfoInvokeRequest(channelID string) fab.ChaincodeInvokeRequest {
	var args [][]byte
	args = append(args, []byte(channelID))

	cir := fab.ChaincodeInvokeRequest{
		ChaincodeID: qscc,
		Fcn:         qsccChannelInfo,
		Args:        args,
	}
	return cir
}

func createBlockByHashInvokeRequest(channelID string, blockHash []byte) fab.ChaincodeInvokeRequest {

	var args [][]byte
	args = append(args, []byte(channelID))
	args = append(args, blockHash)

	cir := fab.ChaincodeInvokeRequest{
		ChaincodeID: qscc,
		Fcn:         qsccBlockByHash,
		Args:        args,
	}
	return cir
}

func createBlockByNumberInvokeRequest(channelID string, blockNumber uint64) fab.ChaincodeInvokeRequest {

	var args [][]byte
	args = append(args, []byte(channelID))
	args = append(args, []byte(strconv.FormatUint(blockNumber, 10)))

	cir := fab.ChaincodeInvokeRequest{
		ChaincodeID: qscc,
		Fcn:         qsccBlockByNumber,
		Args:        args,
	}
	return cir
}

func createBlockByTxIDInvokeRequest(channelID string, transactionID fab.TransactionID) fab.ChaincodeInvokeRequest {
	var args [][]byte
	args = append(args, []byte(channelID))
	args = append(args, []byte(transactionID))

	cir := fab.ChaincodeInvokeRequest{
		ChaincodeID: qscc,
		Fcn:         qsccBlockByTxID,
		Args:        args,
	}
	return cir
}
