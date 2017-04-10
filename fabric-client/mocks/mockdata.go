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

package mocks

import (
	"github.com/hyperledger/fabric-sdk-go/fabric-client/util"
	"github.com/hyperledger/fabric/protos/common"
)

// NewSimpleMockBlock returns a simple mock block
func NewSimpleMockBlock() *common.Block {
	return &common.Block{
		Data: &common.BlockData{
			Data: [][]byte{[]byte("test")},
		},
	}
}

// MockConfigBlockBuilder is used to build a mock Chain configuration block builder
type MockConfigBlockBuilder struct {
	Index           uint64
	LastConfigIndex uint64
}

// Build will create a mock Chain configuration block
func (b *MockConfigBlockBuilder) Build() *common.Block {
	return &common.Block{
		Header: &common.BlockHeader{
			Number: b.Index,
		},
		Metadata: &common.BlockMetadata{
			Metadata: b.buildMetaDataBytes(),
		},
		Data: &common.BlockData{
			Data: b.buildBlockEnvelopeBytes(),
		},
	}
}

func (b *MockConfigBlockBuilder) buildMetaDataBytes() [][]byte {
	return [][]byte{b.buildSignaturesMetaDataBytes(), b.buildLastConfigMetaDataBytes()}
}

func (b *MockConfigBlockBuilder) buildSignaturesMetaDataBytes() []byte {
	return []byte("test signatures")
}

func (b *MockConfigBlockBuilder) buildLastConfigMetaDataBytes() []byte {
	return util.MarshalOrPanic(&common.Metadata{Value: b.buildLastConfigBytes()})
}

func (b *MockConfigBlockBuilder) buildLastConfigBytes() []byte {
	return util.MarshalOrPanic(&common.LastConfig{Index: b.LastConfigIndex})
}

func (b *MockConfigBlockBuilder) buildBlockEnvelopeBytes() [][]byte {
	return [][]byte{b.buildEnvelopeBytes()}
}

func (b *MockConfigBlockBuilder) buildEnvelopeBytes() []byte {
	return util.MarshalOrPanic(&common.Envelope{Payload: b.buildPayloadBytes()})
}

func (b *MockConfigBlockBuilder) buildPayloadBytes() []byte {
	return util.MarshalOrPanic(&common.Payload{Header: b.buildHeader(), Data: b.buildConfigEnvelopeBytes()})
}

func (b *MockConfigBlockBuilder) buildHeader() *common.Header {
	return &common.Header{ChannelHeader: b.buildChannelHeaderBytes()}
}

func (b *MockConfigBlockBuilder) buildChannelHeaderBytes() []byte {
	return util.MarshalOrPanic(&common.ChannelHeader{Type: int32(common.HeaderType_CONFIG)})
}

func (b *MockConfigBlockBuilder) buildConfigEnvelopeBytes() []byte {
	return util.MarshalOrPanic(&common.ConfigEnvelope{Config: b.buildConfig()})
}

func (b *MockConfigBlockBuilder) buildConfig() *common.Config {
	return &common.Config{Sequence: 0, ChannelGroup: b.buildConfigGroup()}
}

func (b *MockConfigBlockBuilder) buildConfigGroup() *common.ConfigGroup {
	return &common.ConfigGroup{}
}

type MockConfigUpdateEnvelopeBuilder struct {
}

func (b *MockConfigUpdateEnvelopeBuilder) BuildBytes() []byte {
	return util.MarshalOrPanic(&common.Envelope{Payload: b.buildPayloadBytes()})
}

func (b *MockConfigUpdateEnvelopeBuilder) buildPayloadBytes() []byte {
	return util.MarshalOrPanic(&common.Payload{Header: b.buildHeader(), Data: b.buildConfigEnvelopeBytes()})
}

func (b *MockConfigUpdateEnvelopeBuilder) buildHeader() *common.Header {
	return &common.Header{ChannelHeader: b.buildChannelHeaderBytes()}
}

func (b *MockConfigUpdateEnvelopeBuilder) buildChannelHeaderBytes() []byte {
	return util.MarshalOrPanic(&common.ChannelHeader{Type: int32(common.HeaderType_CONFIG_UPDATE)})
}

func (b *MockConfigUpdateEnvelopeBuilder) buildConfigEnvelopeBytes() []byte {
	return util.MarshalOrPanic(&common.ConfigEnvelope{Config: b.buildConfig()})
}

func (b *MockConfigUpdateEnvelopeBuilder) buildConfig() *common.Config {
	return &common.Config{Sequence: 0, ChannelGroup: b.buildConfigGroup()}
}

func (b *MockConfigUpdateEnvelopeBuilder) buildConfigGroup() *common.ConfigGroup {
	return &common.ConfigGroup{}
}
