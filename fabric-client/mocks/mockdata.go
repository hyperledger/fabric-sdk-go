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
	"io/ioutil"
	"log"

	"github.com/hyperledger/fabric-sdk-go/fabric-client/util"
	fabric_config "github.com/hyperledger/fabric/common/config"
	ledger_util "github.com/hyperledger/fabric/core/ledger/util"
	"github.com/hyperledger/fabric/protos/common"
	mb "github.com/hyperledger/fabric/protos/msp"
	ab "github.com/hyperledger/fabric/protos/orderer"
	"github.com/hyperledger/fabric/protos/peer"
)

// NewSimpleMockBlock returns a simple mock block
func NewSimpleMockBlock() *common.Block {
	return &common.Block{
		Data: &common.BlockData{
			Data: [][]byte{[]byte("test")},
		},
	}
}

// MockConfigGroupBuilder is used to build a mock ConfigGroup
type MockConfigGroupBuilder struct {
	Version        uint64
	ModPolicy      string
	OrdererAddress string
	MSPNames       []string
}

// MockConfigBlockBuilder is used to build a mock Chain configuration block
type MockConfigBlockBuilder struct {
	MockConfigGroupBuilder
	Index           uint64
	LastConfigIndex uint64
}

// MockConfigUpdateEnvelopeBuilder builds a mock ConfigUpdateEnvelope
type MockConfigUpdateEnvelopeBuilder struct {
	MockConfigGroupBuilder
	ChannelID string
}

// Build creates a mock Chain configuration Block
func (b *MockConfigBlockBuilder) Build() *common.Block {
	return &common.Block{
		Header: &common.BlockHeader{
			Number: b.Index,
		},
		Metadata: b.buildBlockMetadata(),
		Data: &common.BlockData{
			Data: b.buildBlockEnvelopeBytes(),
		},
	}
}

// buildBlockMetadata builds BlockMetadata that contains an array of bytes in the following order:
// 	0: SIGNATURES
// 	1: LAST_CONFIG
// 	2: TRANSACTIONS_FILTER
// 	3: ORDERER
func (b *MockConfigBlockBuilder) buildBlockMetadata() *common.BlockMetadata {
	return &common.BlockMetadata{
		Metadata: [][]byte{
			b.buildSignaturesMetaDataBytes(),
			util.MarshalOrPanic(b.buildLastConfigMetaData()),
			b.buildTransactionsFilterMetaDataBytes(),
			b.buildOrdererMetaDataBytes(),
		},
	}
}

func (b *MockConfigBlockBuilder) buildSignaturesMetaDataBytes() []byte {
	return []byte("test signatures")
}

func (b *MockConfigBlockBuilder) buildLastConfigMetaData() *common.Metadata {
	return &common.Metadata{
		Value: util.MarshalOrPanic(b.buildLastConfig()),
	}
}

func (b *MockConfigBlockBuilder) buildTransactionsFilterMetaDataBytes() []byte {
	return []byte(ledger_util.TxValidationFlags{uint8(peer.TxValidationCode_VALID)})
}

func (b *MockConfigBlockBuilder) buildOrdererMetaDataBytes() []byte {
	// TODO: What's the structure of this?
	return []byte("orderer meta data")
}

func (b *MockConfigBlockBuilder) buildLastConfig() *common.LastConfig {
	return &common.LastConfig{Index: b.LastConfigIndex}
}

func (b *MockConfigBlockBuilder) buildBlockEnvelopeBytes() [][]byte {
	return [][]byte{util.MarshalOrPanic(b.buildEnvelope())}
}

func (b *MockConfigBlockBuilder) buildEnvelope() *common.Envelope {
	return &common.Envelope{
		Payload: util.MarshalOrPanic(b.buildPayload()),
	}
}

func (b *MockConfigBlockBuilder) buildPayload() *common.Payload {
	return &common.Payload{
		Header: b.buildHeader(),
		Data:   util.MarshalOrPanic(b.buildConfigEnvelope()),
	}
}

func (b *MockConfigBlockBuilder) buildHeader() *common.Header {
	return &common.Header{
		ChannelHeader: util.MarshalOrPanic(b.buildChannelHeader()),
	}
}

func (b *MockConfigBlockBuilder) buildChannelHeader() *common.ChannelHeader {
	return &common.ChannelHeader{
		Type: int32(common.HeaderType_CONFIG),
	}
}

func (b *MockConfigBlockBuilder) buildConfigEnvelope() *common.ConfigEnvelope {
	return &common.ConfigEnvelope{Config: b.buildConfig()}
}

func (b *MockConfigBlockBuilder) buildConfig() *common.Config {
	return &common.Config{
		Sequence:     0,
		ChannelGroup: b.buildConfigGroup(),
	}
}

func (b *MockConfigGroupBuilder) buildConfigGroup() *common.ConfigGroup {
	return &common.ConfigGroup{
		Groups: map[string]*common.ConfigGroup{
			"Orderer":     b.buildOrdererGroup(),
			"Application": b.buildApplicationGroup(),
		},
		Policies: map[string]*common.ConfigPolicy{
			"BlockValidation": b.buildBasicConfigPolicy(),
			"Writers":         b.buildBasicConfigPolicy(),
			"Readers":         b.buildBasicConfigPolicy(),
			"Admins":          b.buildBasicConfigPolicy(),
		},
		Values: map[string]*common.ConfigValue{
			fabric_config.OrdererAddressesKey: b.buildOrdererAddressesConfigValue(),
		},
		Version:   b.Version,
		ModPolicy: b.ModPolicy,
	}
}

func (b *MockConfigGroupBuilder) buildOrdererAddressesConfigValue() *common.ConfigValue {
	return &common.ConfigValue{
		Version:   b.Version,
		ModPolicy: b.ModPolicy,
		Value:     util.MarshalOrPanic(b.buildOrdererAddresses())}
}

func (b *MockConfigGroupBuilder) buildOrdererAddresses() *common.OrdererAddresses {
	return &common.OrdererAddresses{
		Addresses: []string{b.OrdererAddress},
	}
}

func (b *MockConfigGroupBuilder) buildOrdererGroup() *common.ConfigGroup {
	return &common.ConfigGroup{
		Groups: map[string]*common.ConfigGroup{
			"OrdererMSP": b.buildMSPGroup("OrdererMSP"),
		},
		Policies: map[string]*common.ConfigPolicy{
			"BlockValidation": b.buildBasicConfigPolicy(),
			"Writers":         b.buildBasicConfigPolicy(),
			"Readers":         b.buildBasicConfigPolicy(),
			"Admins":          b.buildBasicConfigPolicy(),
		},
		Values: map[string]*common.ConfigValue{
			fabric_config.BatchSizeKey: b.buildBatchSizeConfigValue(),
			// TODO: More
		},
		Version:   b.Version,
		ModPolicy: b.ModPolicy,
	}
}

func (b *MockConfigGroupBuilder) buildMSPGroup(mspName string) *common.ConfigGroup {
	return &common.ConfigGroup{
		Groups: nil,
		Policies: map[string]*common.ConfigPolicy{
			"Admins":  b.buildSignatureConfigPolicy(),
			"Writers": b.buildSignatureConfigPolicy(),
			"Readers": b.buildSignatureConfigPolicy(),
		},
		Values: map[string]*common.ConfigValue{
			fabric_config.MSPKey: b.buildMSPConfigValue(mspName),
			// TODO: More
		},
		Version:   b.Version,
		ModPolicy: b.ModPolicy,
	}
}

func (b *MockConfigGroupBuilder) buildMSPConfigValue(name string) *common.ConfigValue {
	return &common.ConfigValue{
		Version:   b.Version,
		ModPolicy: b.ModPolicy,
		Value:     util.MarshalOrPanic(b.buildMSPConfig(name))}
}

func (b *MockConfigGroupBuilder) buildBatchSizeConfigValue() *common.ConfigValue {
	return &common.ConfigValue{
		Version:   b.Version,
		ModPolicy: b.ModPolicy,
		Value:     util.MarshalOrPanic(b.buildBatchSize())}
}

func (b *MockConfigGroupBuilder) buildBatchSize() *ab.BatchSize {
	return &ab.BatchSize{
		MaxMessageCount:   10,
		AbsoluteMaxBytes:  103809024,
		PreferredMaxBytes: 524288,
	}
}

func (b *MockConfigGroupBuilder) buildMSPConfig(name string) *mb.MSPConfig {
	return &mb.MSPConfig{
		Type:   0,
		Config: util.MarshalOrPanic(b.buildfabricMSPConfig(name)),
	}
}

func (b *MockConfigGroupBuilder) buildfabricMSPConfig(name string) *mb.FabricMSPConfig {
	return &mb.FabricMSPConfig{
		Name:                          name,
		Admins:                        [][]byte{},
		IntermediateCerts:             [][]byte{},
		OrganizationalUnitIdentifiers: []*mb.FabricOUIdentifier{},
		RevocationList:                [][]byte{},
		RootCerts:                     [][]byte{b.buildRootCertBytes()},
		SigningIdentity:               nil,
	}
}

func (b *MockConfigGroupBuilder) buildRootCertBytes() []byte {
	pem, err := ioutil.ReadFile("../test/fixtures/root.pem")
	if err != nil {
		log.Fatal(err)
	}
	return pem
}

func (b *MockConfigGroupBuilder) buildBasicConfigPolicy() *common.ConfigPolicy {
	return &common.ConfigPolicy{
		Version:   b.Version,
		ModPolicy: b.ModPolicy,
		Policy:    &common.Policy{},
	}
}

func (b *MockConfigGroupBuilder) buildSignatureConfigPolicy() *common.ConfigPolicy {
	return &common.ConfigPolicy{
		Version:   b.Version,
		ModPolicy: b.ModPolicy,
		Policy:    b.buildSignaturePolicy(),
	}
}

func (b *MockConfigGroupBuilder) buildSignaturePolicy() *common.Policy {
	return &common.Policy{
		Type:   int32(common.Policy_SIGNATURE),
		Policy: util.MarshalOrPanic(b.buildSignedBySignaturePolicy()),
	}
}

func (b *MockConfigGroupBuilder) buildSignedBySignaturePolicy() *common.SignaturePolicy {
	return &common.SignaturePolicy{
		Type: &common.SignaturePolicy_SignedBy{
			SignedBy: 0,
		},
	}
}

func (b *MockConfigGroupBuilder) buildApplicationGroup() *common.ConfigGroup {
	groups := make(map[string]*common.ConfigGroup)
	for _, name := range b.MSPNames {
		groups[name] = b.buildMSPGroup(name)
	}

	return &common.ConfigGroup{
		Groups: groups,
		Policies: map[string]*common.ConfigPolicy{
			"Admins":  b.buildSignatureConfigPolicy(),
			"Writers": b.buildSignatureConfigPolicy(),
			"Readers": b.buildSignatureConfigPolicy(),
		},
		Values: map[string]*common.ConfigValue{
			fabric_config.BatchSizeKey: b.buildBatchSizeConfigValue(),
			// TODO: More
		},
		Version:   b.Version,
		ModPolicy: b.ModPolicy,
	}
}

// Build builds an Envelope that contains a mock ConfigUpdateEnvelope
func (b *MockConfigUpdateEnvelopeBuilder) Build() *common.Envelope {
	return &common.Envelope{
		Payload: util.MarshalOrPanic(b.buildPayload()),
	}
}

// BuildBytes builds an Envelope that contains a mock ConfigUpdateEnvelope and returns the marshaled bytes
func (b *MockConfigUpdateEnvelopeBuilder) BuildBytes() []byte {
	return util.MarshalOrPanic(b.Build())
}

func (b *MockConfigUpdateEnvelopeBuilder) buildPayload() *common.Payload {
	return &common.Payload{
		Header: b.buildHeader(),
		Data:   util.MarshalOrPanic(b.buildConfigUpdateEnvelope()),
	}
}

func (b *MockConfigUpdateEnvelopeBuilder) buildHeader() *common.Header {
	return &common.Header{
		ChannelHeader: util.MarshalOrPanic(&common.ChannelHeader{
			Type: int32(common.HeaderType_CONFIG_UPDATE)},
		),
	}
}

func (b *MockConfigUpdateEnvelopeBuilder) buildConfigUpdateEnvelope() *common.ConfigUpdateEnvelope {
	return &common.ConfigUpdateEnvelope{
		ConfigUpdate: util.MarshalOrPanic(b.buildConfigUpdate()),
		Signatures:   nil,
	}
}

func (b *MockConfigUpdateEnvelopeBuilder) buildConfigUpdate() *common.ConfigUpdate {
	return &common.ConfigUpdate{
		ChannelId: b.ChannelID,
		ReadSet:   b.buildConfigGroup(),
		WriteSet:  b.buildConfigGroup(),
	}
}
