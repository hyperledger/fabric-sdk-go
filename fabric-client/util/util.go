/*
Copyright SecureKey Technologies Inc. All Rights Reserved.


Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at


      http://www.apache.org/licenses/LICENSE-2.0


Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package util

import (
	"fmt"
	"time"

	"github.com/hyperledger/fabric/core/crypto/primitives"
	ab "github.com/hyperledger/fabric/protos/orderer"

	"github.com/golang/protobuf/proto"
	google_protobuf "github.com/golang/protobuf/ptypes/timestamp"
	"github.com/hyperledger/fabric/protos/common"
	protos_utils "github.com/hyperledger/fabric/protos/utils"
)

// CreateGenesisBlockRequest creates a seek request for block 0 on the specified
// channel. This request is sent to the ordering service to request blocks
func CreateGenesisBlockRequest(channelName string, creator []byte) []byte {
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

// MarshalOrPanic serializes a protobuf message and panics if this operation fails.
func MarshalOrPanic(pb proto.Message) []byte {
	data, err := proto.Marshal(pb)
	if err != nil {
		panic(err)
	}
	return data
}

// GenerateRandomNonce generates a random nonce
func GenerateRandomNonce() ([]byte, error) {
	return primitives.GetRandomNonce()
}

// ComputeTxID computes a transaction ID from a given nonce and creator ID
func ComputeTxID(nonce []byte, creatorID []byte) (string, error) {
	return protos_utils.ComputeProposalTxID(nonce, creatorID)
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
		return nil, fmt.Errorf("unable to unmarshal meta data at index %d: %v", common.BlockMetadataIndex_LAST_CONFIG, err)
	}

	lastConfig := &common.LastConfig{}
	err = proto.Unmarshal(metadata.Value, lastConfig)
	if err != nil {
		return nil, fmt.Errorf("unable to unmarshal last config from meta data: %v", err)
	}

	return lastConfig, err
}

// BuildChannelHeader builds a ChannelHeader with the given parameters
func BuildChannelHeader(channelName string, headerType common.HeaderType, txID string, epoch uint64) (*common.ChannelHeader, error) {
	now := time.Now()
	channelHeader := &common.ChannelHeader{
		Type:      int32(headerType),
		Version:   1,
		Timestamp: &google_protobuf.Timestamp{Seconds: int64(now.Second()), Nanos: int32(now.Nanosecond())},
		ChannelId: channelName,
		Epoch:     epoch,
		TxId:      txID,
	}
	return channelHeader, nil
}
