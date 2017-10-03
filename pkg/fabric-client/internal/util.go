/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package internal

import (
	"github.com/golang/protobuf/proto"

	ab "github.com/hyperledger/fabric-sdk-go/internal/github.com/hyperledger/fabric/protos/orderer"
	"github.com/hyperledger/fabric-sdk-go/third_party/github.com/hyperledger/fabric/protos/common"

	"github.com/hyperledger/fabric-sdk-go/internal/github.com/hyperledger/fabric/common/crypto"
	protos_utils "github.com/hyperledger/fabric-sdk-go/internal/github.com/hyperledger/fabric/protos/utils"
	"github.com/hyperledger/fabric-sdk-go/pkg/errors"
)

// CreateSeekGenesisBlockRequest creates a seek request for block 0 on the specified
// channel. This request is sent to the ordering service to request blocks
func CreateSeekGenesisBlockRequest(channelName string, creator []byte) []byte {
	return MarshalOrPanic(&common.Payload{
		Header: &common.Header{
			ChannelHeader: MarshalOrPanic(&common.ChannelHeader{
				ChannelId: channelName,
			}),
			SignatureHeader: MarshalOrPanic(&common.SignatureHeader{Creator: creator}),
		},
		Data: MarshalOrPanic(&ab.SeekInfo{
			Start:    &ab.SeekPosition{Type: &ab.SeekPosition_Specified{Specified: &ab.SeekSpecified{Number: 0}}},
			Stop:     &ab.SeekPosition{Type: &ab.SeekPosition_Specified{Specified: &ab.SeekSpecified{Number: 0}}},
			Behavior: ab.SeekInfo_BLOCK_UNTIL_READY,
		}),
	})
}

// GenerateRandomNonce generates a random nonce
func GenerateRandomNonce() ([]byte, error) {
	return crypto.GetRandomNonce()
}

// ComputeTxID computes a transaction ID from a given nonce and creator ID
func ComputeTxID(nonce []byte, creator []byte) (string, error) {
	return protos_utils.ComputeProposalTxID(nonce, creator)
}

// NewNewestSeekPosition returns a SeekPosition that requests the newest block
func NewNewestSeekPosition() *ab.SeekPosition {
	return &ab.SeekPosition{Type: &ab.SeekPosition_Newest{Newest: &ab.SeekNewest{}}}
}

// NewSpecificSeekPosition returns a SeekPosition that requests the block at the given index
func NewSpecificSeekPosition(index uint64) *ab.SeekPosition {
	return &ab.SeekPosition{Type: &ab.SeekPosition_Specified{Specified: &ab.SeekSpecified{Number: index}}}
}

// GetLastConfigFromBlock returns the LastConfig data from the given block
func GetLastConfigFromBlock(block *common.Block) (*common.LastConfig, error) {
	metadata := &common.Metadata{}
	err := proto.Unmarshal(block.Metadata.Metadata[common.BlockMetadataIndex_LAST_CONFIG], metadata)
	if err != nil {
		return nil, errors.Wrap(err, "unmarshal block metadata failed")
	}

	lastConfig := &common.LastConfig{}
	err = proto.Unmarshal(metadata.Value, lastConfig)
	if err != nil {
		return nil, errors.Wrap(err, "unmarshal last config from metadata failed")
	}

	return lastConfig, err
}

// BuildHeader ...
func BuildHeader(creator []byte, channelHeader *common.ChannelHeader, nonce []byte) (*common.Header, error) {
	signatureHeader := &common.SignatureHeader{
		Creator: creator,
		Nonce:   nonce,
	}
	sh, err := proto.Marshal(signatureHeader)
	if err != nil {
		return nil, errors.Wrap(err, "marshal signatureHeader failed")
	}
	ch, err := proto.Marshal(channelHeader)
	if err != nil {
		return nil, errors.Wrap(err, "marshal channelHeader failed")
	}
	header := &common.Header{
		SignatureHeader: sh,
		ChannelHeader:   ch,
	}
	return header, nil
}

// MarshalOrPanic serializes a protobuf message and panics if this operation fails.
func MarshalOrPanic(pb proto.Message) []byte {
	data, err := proto.Marshal(pb)
	if err != nil {
		panic(err)
	}
	return data
}
