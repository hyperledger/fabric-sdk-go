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
	"github.com/hyperledger/fabric/core/crypto/primitives"
	ab "github.com/hyperledger/fabric/protos/orderer"

	"github.com/golang/protobuf/proto"
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
