/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package mocks

import (
	"crypto/sha256"

	"github.com/golang/protobuf/proto"
	"github.com/golang/protobuf/ptypes/timestamp"

	"github.com/hyperledger/fabric-protos-go/common"
	mb "github.com/hyperledger/fabric-protos-go/msp"
	ab "github.com/hyperledger/fabric-protos-go/orderer"
	pp "github.com/hyperledger/fabric-protos-go/peer"
	cutil "github.com/hyperledger/fabric-sdk-go/internal/github.com/hyperledger/fabric/common/util"
	"github.com/hyperledger/fabric-sdk-go/internal/github.com/hyperledger/fabric/protoutil"

	"time"

	channelConfig "github.com/hyperledger/fabric-sdk-go/internal/github.com/hyperledger/fabric/common/channelconfig"
	ledger_util "github.com/hyperledger/fabric-sdk-go/internal/github.com/hyperledger/fabric/core/ledger/util"
	"github.com/pkg/errors"
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
	return errors.New("Test Error")
}

/*
func NewMockGenesisBlock() *common.Block {
	file := file.Open()
}*/

// MockConfigGroupBuilder is used to build a mock ConfigGroup
type MockConfigGroupBuilder struct {
	Version                 uint64
	ModPolicy               string
	OrdererAddress          string
	MSPNames                []string
	RootCA                  string
	Groups                  map[string]*common.ConfigGroup
	ChannelCapabilities     []string
	ApplicationCapabilities []string
	OrdererCapabilities     []string
	PolicyRefs              []string
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
			channelConfig.OrdererAddressesKey: b.buildOrdererAddressesConfigValue(),
			channelConfig.CapabilitiesKey:     b.buildCapabilitiesConfigValue(b.ChannelCapabilities),
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
			channelConfig.ConsensusTypeKey:       b.buildConsensusTypeConfigValue(),
			channelConfig.BatchSizeKey:           b.buildBatchSizeConfigValue(),
			channelConfig.BatchTimeoutKey:        b.buildBatchTimeoutConfigValue(),
			channelConfig.ChannelRestrictionsKey: b.buildChannelRestrictionsConfigValue(),
			channelConfig.CapabilitiesKey:        b.buildCapabilitiesConfigValue(b.OrdererCapabilities),
			channelConfig.KafkaBrokersKey:        b.buildKafkaBrokersConfigValue(),
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
			channelConfig.MSPKey: b.buildMSPConfigValue(mspName),
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

func (b *MockConfigGroupBuilder) buildCapabilitiesConfigValue(capabilityNames []string) *common.ConfigValue {
	return &common.ConfigValue{
		Version:   b.Version,
		ModPolicy: b.ModPolicy,
		Value:     marshalOrPanic(b.buildCapabilities(capabilityNames))}
}

func (b *MockConfigGroupBuilder) buildKafkaBrokersConfigValue() *common.ConfigValue {
	return &common.ConfigValue{
		Version:   b.Version,
		ModPolicy: b.ModPolicy,
		Value:     marshalOrPanic(b.buildKafkaBrokers())}
}

func (b *MockConfigGroupBuilder) buildACLsConfigValue(policyRefs []string) *common.ConfigValue {
	return &common.ConfigValue{
		Version:   b.Version,
		ModPolicy: b.ModPolicy,
		Value:     marshalOrPanic(b.buildACLs(policyRefs))}
}

func (b *MockConfigGroupBuilder) buildBatchSize() *ab.BatchSize {
	return &ab.BatchSize{
		MaxMessageCount:   10,
		AbsoluteMaxBytes:  103809024,
		PreferredMaxBytes: 524288,
	}
}

func (b *MockConfigGroupBuilder) buildConsensusType() *ab.ConsensusType {
	return &ab.ConsensusType{
		Type:  "kafka",
		State: ab.ConsensusType_STATE_NORMAL,
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

func (b *MockConfigGroupBuilder) buildCapabilities(capabilityNames []string) *common.Capabilities {
	capabilities := make(map[string]*common.Capability)
	for _, capability := range capabilityNames {
		capabilities[capability] = &common.Capability{}
	}
	return &common.Capabilities{
		Capabilities: capabilities,
	}
}

func (b *MockConfigGroupBuilder) buildKafkaBrokers() *ab.KafkaBrokers {
	brokers := []string{"kafkabroker"}
	return &ab.KafkaBrokers{
		Brokers: brokers,
	}
}

func (b *MockConfigGroupBuilder) buildACLs(policyRefs []string) *pp.ACLs {
	acls := make(map[string]*pp.APIResource)
	for _, policyRef := range policyRefs {
		acls[policyRef] = &pp.APIResource{PolicyRef: policyRef}
	}
	return &pp.ACLs{
		Acls: acls,
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
			channelConfig.CapabilitiesKey: b.buildCapabilitiesConfigValue(b.ApplicationCapabilities),
			channelConfig.ACLsKey:         b.buildACLsConfigValue(b.PolicyRefs),
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

// CreateBlockWithCCEvent creates a mock block
func CreateBlockWithCCEvent(events *pp.ChaincodeEvent, txID string,
	channelID string) (*common.Block, error) {
	return CreateBlockWithCCEventAndTxStatus(events, txID, channelID, pp.TxValidationCode_VALID)
}

// CreateBlockWithCCEventAndTxStatus creates a mock block with the given CC event and TX validation code
func CreateBlockWithCCEventAndTxStatus(events *pp.ChaincodeEvent, txID string,
	channelID string, txValidationCode pp.TxValidationCode) (*common.Block, error) {
	chdr := &common.ChannelHeader{
		Type:    int32(common.HeaderType_ENDORSER_TRANSACTION),
		Version: 1,
		Timestamp: &timestamp.Timestamp{
			Seconds: time.Now().Unix(),
			Nanos:   0,
		},
		ChannelId: channelID,
		TxId:      txID}
	hdr := &common.Header{ChannelHeader: protoutil.MarshalOrPanic(chdr)}
	payload := &common.Payload{Header: hdr}
	cea := &pp.ChaincodeEndorsedAction{}
	ccaPayload := &pp.ChaincodeActionPayload{Action: cea}
	env := &common.Envelope{}
	taa := &pp.TransactionAction{}
	taas := make([]*pp.TransactionAction, 1)
	taas[0] = taa
	tx := &pp.Transaction{Actions: taas}

	pHashBytes := []byte("proposal_hash")
	pResponse := &pp.Response{Status: 200}
	results := []byte("results")
	eventBytes, err := protoutil.GetBytesChaincodeEvent(events)
	if err != nil {
		return nil, err
	}
	ccaPayload.Action.ProposalResponsePayload, err = protoutil.GetBytesProposalResponsePayload(pHashBytes, pResponse, results, eventBytes, nil)
	if err != nil {
		return nil, err
	}
	tx.Actions[0].Payload, err = protoutil.GetBytesChaincodeActionPayload(ccaPayload)
	if err != nil {
		return nil, err
	}
	payload.Data, err = protoutil.GetBytesTransaction(tx)
	if err != nil {
		return nil, err
	}
	env.Payload, err = protoutil.GetBytesPayload(payload)
	if err != nil {
		return nil, err
	}
	ebytes, err := protoutil.GetBytesEnvelope(env)
	if err != nil {
		return nil, err
	}

	block := newBlock(1, []byte{})
	block.Data.Data = append(block.Data.Data, ebytes)

	blockbytes := cutil.ConcatenateBytes(block.Data.Data...)
	block.Header.DataHash = computeSHA256(blockbytes)

	txsfltr := ledger_util.NewTxValidationFlags(len(block.Data.Data))
	for i := 0; i < len(block.Data.Data); i++ {
		txsfltr[i] = uint8(txValidationCode)
	}

	block.Metadata.Metadata[common.BlockMetadataIndex_TRANSACTIONS_FILTER] = txsfltr

	return block, nil
}

// NewBlock construct a block with no data and no metadata.
func newBlock(seqNum uint64, previousHash []byte) *common.Block {
	block := &common.Block{}
	block.Header = &common.BlockHeader{}
	block.Header.Number = seqNum
	block.Header.PreviousHash = previousHash
	block.Data = &common.BlockData{}

	var metadataContents [][]byte
	for i := 0; i < len(common.BlockMetadataIndex_name); i++ {
		metadataContents = append(metadataContents, []byte{})
	}
	block.Metadata = &common.BlockMetadata{Metadata: metadataContents}

	return block
}

func computeSHA256(data []byte) (hash []byte) {
	h := sha256.New()
	_, err := h.Write(data)
	if err != nil {
		panic("unable to create digest")
	}
	return h.Sum(nil)
}
