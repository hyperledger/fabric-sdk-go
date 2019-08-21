/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package resource

import (
	"github.com/golang/protobuf/proto"
	"github.com/hyperledger/fabric-protos-go/common"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/fab"
	"github.com/pkg/errors"
)

const (
	cscc            = "cscc"
	csccJoinChannel = "JoinChain"
	csccChannels    = "GetChannels"
)

func createJoinChannelInvokeRequest(genesisBlock *common.Block) (fab.ChaincodeInvokeRequest, error) { //nolint

	genesisBlockBytes, err := proto.Marshal(genesisBlock)
	if err != nil {
		return fab.ChaincodeInvokeRequest{}, errors.Wrap(err, "marshal genesis block failed")
	}

	// Create join channel transaction proposal for target peers
	var args [][]byte
	args = append(args, genesisBlockBytes)

	cir := fab.ChaincodeInvokeRequest{
		ChaincodeID: cscc,
		Fcn:         csccJoinChannel,
		Args:        args,
	}

	return cir, nil
}

func createChannelsInvokeRequest() fab.ChaincodeInvokeRequest {
	cir := fab.ChaincodeInvokeRequest{
		ChaincodeID: cscc,
		Fcn:         csccChannels,
	}
	return cir
}
