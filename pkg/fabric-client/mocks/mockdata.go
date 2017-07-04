/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package mocks

import (
	"fmt"

	"github.com/golang/protobuf/proto"
	fabric_config "github.com/hyperledger/fabric/common/config"
	ledger_util "github.com/hyperledger/fabric/core/ledger/util"
	"github.com/hyperledger/fabric/protos/common"
	mb "github.com/hyperledger/fabric/protos/msp"
	ab "github.com/hyperledger/fabric/protos/orderer"
	pp "github.com/hyperledger/fabric/protos/peer"
)

// NewSimpleMockBlock returns a simple mock block
func NewSimpleMockBlock() *common.Block {
	return &common.Block{
		Data: &common.BlockData{
			Data: [][]byte{[]byte("test")},
		},
		Header: &common.BlockHeader{
			DataHash:     []byte(""),
			PreviousHash: []byte(""),
			Number:       1,
		},
		Metadata: &common.BlockMetadata{
			Metadata: [][]byte{[]byte("test")},
		},
	}
}

// NewSimpleMockError returns a error
func NewSimpleMockError() error {
	return fmt.Errorf("Test Error")
}

// MockConfigGroupBuilder is used to build a mock ConfigGroup
type MockConfigGroupBuilder struct {
	Version        uint64
	ModPolicy      string
	OrdererAddress string
	MSPNames       []string
	RootCA         string
	Groups         map[string]*common.ConfigGroup
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
			marshalOrPanic(b.buildLastConfigMetaData()),
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
		Value: marshalOrPanic(b.buildLastConfig()),
	}
}

func (b *MockConfigBlockBuilder) buildTransactionsFilterMetaDataBytes() []byte {
	return []byte(ledger_util.TxValidationFlags{uint8(pp.TxValidationCode_VALID)})
}

func (b *MockConfigBlockBuilder) buildOrdererMetaDataBytes() []byte {
	// TODO: What's the structure of this?
	return []byte("orderer meta data")
}

func (b *MockConfigBlockBuilder) buildLastConfig() *common.LastConfig {
	return &common.LastConfig{Index: b.LastConfigIndex}
}

func (b *MockConfigBlockBuilder) buildBlockEnvelopeBytes() [][]byte {
	return [][]byte{marshalOrPanic(b.buildEnvelope())}
}

func (b *MockConfigBlockBuilder) buildEnvelope() *common.Envelope {
	return &common.Envelope{
		Payload: marshalOrPanic(b.buildPayload()),
	}
}

func (b *MockConfigBlockBuilder) buildPayload() *common.Payload {
	return &common.Payload{
		Header: b.buildHeader(),
		Data:   marshalOrPanic(b.buildConfigEnvelope()),
	}
}

func (b *MockConfigBlockBuilder) buildHeader() *common.Header {
	return &common.Header{
		ChannelHeader: marshalOrPanic(b.buildChannelHeader()),
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
		Value:     marshalOrPanic(b.buildOrdererAddresses())}
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
			fabric_config.BatchSizeKey:                 b.buildBatchSizeConfigValue(),
			fabric_config.AnchorPeersKey:               b.buildAnchorPeerConfigValue(),
			fabric_config.ConsensusTypeKey:             b.buildConsensusTypeConfigValue(),
			fabric_config.BatchTimeoutKey:              b.buildBatchTimeoutConfigValue(),
			fabric_config.ChannelRestrictionsKey:       b.buildChannelRestrictionsConfigValue(),
			fabric_config.HashingAlgorithmKey:          b.buildHashingAlgorithmConfigValue(),
			fabric_config.BlockDataHashingStructureKey: b.buildBlockDataHashingStructureConfigValue(),
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
		Value:     marshalOrPanic(b.buildMSPConfig(name))}
}

func (b *MockConfigGroupBuilder) buildBatchSizeConfigValue() *common.ConfigValue {
	return &common.ConfigValue{
		Version:   b.Version,
		ModPolicy: b.ModPolicy,
		Value:     marshalOrPanic(b.buildBatchSize())}
}

func (b *MockConfigGroupBuilder) buildAnchorPeerConfigValue() *common.ConfigValue {
	return &common.ConfigValue{
		Version:   b.Version,
		ModPolicy: b.ModPolicy,
		Value:     marshalOrPanic(b.buildAnchorPeer())}
}

func (b *MockConfigGroupBuilder) buildConsensusTypeConfigValue() *common.ConfigValue {
	return &common.ConfigValue{
		Version:   b.Version,
		ModPolicy: b.ModPolicy,
		Value:     marshalOrPanic(b.buildConsensusType())}
}

func (b *MockConfigGroupBuilder) buildBatchTimeoutConfigValue() *common.ConfigValue {
	return &common.ConfigValue{
		Version:   b.Version,
		ModPolicy: b.ModPolicy,
		Value:     marshalOrPanic(b.buildBatchTimeout())}
}

func (b *MockConfigGroupBuilder) buildChannelRestrictionsConfigValue() *common.ConfigValue {
	return &common.ConfigValue{
		Version:   b.Version,
		ModPolicy: b.ModPolicy,
		Value:     marshalOrPanic(b.buildChannelRestrictions())}
}

func (b *MockConfigGroupBuilder) buildHashingAlgorithmConfigValue() *common.ConfigValue {
	return &common.ConfigValue{
		Version:   b.Version,
		ModPolicy: b.ModPolicy,
		Value:     marshalOrPanic(b.buildHashingAlgorithm())}
}

func (b *MockConfigGroupBuilder) buildBlockDataHashingStructureConfigValue() *common.ConfigValue {
	return &common.ConfigValue{
		Version:   b.Version,
		ModPolicy: b.ModPolicy,
		Value:     marshalOrPanic(b.buildBlockDataHashingStructure())}
}

func (b *MockConfigGroupBuilder) buildBatchSize() *ab.BatchSize {
	return &ab.BatchSize{
		MaxMessageCount:   10,
		AbsoluteMaxBytes:  103809024,
		PreferredMaxBytes: 524288,
	}
}

func (b *MockConfigGroupBuilder) buildAnchorPeer() *pp.AnchorPeers {
	ap := pp.AnchorPeer{Host: "sample-host", Port: 22}
	return &pp.AnchorPeers{
		AnchorPeers: []*pp.AnchorPeer{&ap},
	}
}

func (b *MockConfigGroupBuilder) buildConsensusType() *ab.ConsensusType {
	return &ab.ConsensusType{
		Type: "sample-Consensus-Type",
	}
}

func (b *MockConfigGroupBuilder) buildBatchTimeout() *ab.BatchTimeout {
	return &ab.BatchTimeout{
		Timeout: "123",
	}
}

func (b *MockConfigGroupBuilder) buildChannelRestrictions() *ab.ChannelRestrictions {
	return &ab.ChannelRestrictions{
		MaxCount: 200,
	}
}

func (b *MockConfigGroupBuilder) buildHashingAlgorithm() *common.HashingAlgorithm {
	return &common.HashingAlgorithm{
		Name: "SHA2",
	}
}

func (b *MockConfigGroupBuilder) buildBlockDataHashingStructure() *common.BlockDataHashingStructure {
	return &common.BlockDataHashingStructure{
		Width: 64,
	}
}

func (b *MockConfigGroupBuilder) buildMSPConfig(name string) *mb.MSPConfig {
	return &mb.MSPConfig{
		Type:   0,
		Config: marshalOrPanic(b.buildfabricMSPConfig(name)),
	}
}

func (b *MockConfigGroupBuilder) buildfabricMSPConfig(name string) *mb.FabricMSPConfig {
	return &mb.FabricMSPConfig{
		Name:                          name,
		Admins:                        [][]byte{},
		IntermediateCerts:             [][]byte{},
		OrganizationalUnitIdentifiers: []*mb.FabricOUIdentifier{},
		RevocationList:                [][]byte{},
		RootCerts:                     [][]byte{[]byte(b.RootCA)},
		SigningIdentity:               nil,
	}
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
		Type:  int32(common.Policy_SIGNATURE),
		Value: marshalOrPanic(b.buildSignedBySignaturePolicy()),
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
		Payload: marshalOrPanic(b.buildPayload()),
	}
}

// BuildBytes builds an Envelope that contains a mock ConfigUpdateEnvelope and returns the marshaled bytes
func (b *MockConfigUpdateEnvelopeBuilder) BuildBytes() []byte {
	return marshalOrPanic(b.Build())
}

func (b *MockConfigUpdateEnvelopeBuilder) buildPayload() *common.Payload {
	return &common.Payload{
		Header: b.buildHeader(),
		Data:   marshalOrPanic(b.buildConfigUpdateEnvelope()),
	}
}

func (b *MockConfigUpdateEnvelopeBuilder) buildHeader() *common.Header {
	return &common.Header{
		ChannelHeader: marshalOrPanic(&common.ChannelHeader{
			Type: int32(common.HeaderType_CONFIG_UPDATE)},
		),
	}
}

func (b *MockConfigUpdateEnvelopeBuilder) buildConfigUpdateEnvelope() *common.ConfigUpdateEnvelope {
	return &common.ConfigUpdateEnvelope{
		ConfigUpdate: marshalOrPanic(b.buildConfigUpdate()),
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

// BuildConfigUpdateBytes builds an mock ConfigUpdate returns the marshaled bytes
func (b *MockConfigUpdateEnvelopeBuilder) BuildConfigUpdateBytes() []byte {
	return marshalOrPanic(b.buildConfigUpdate())
}

// marshalOrPanic serializes a protobuf message and panics if this operation fails.
func marshalOrPanic(pb proto.Message) []byte {
	data, err := proto.Marshal(pb)
	if err != nil {
		panic(err)
	}
	return data
}
