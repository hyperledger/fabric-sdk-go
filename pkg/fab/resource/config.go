/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package resource

import (
	"github.com/golang/protobuf/proto"
	"github.com/pkg/errors"

	"github.com/hyperledger/fabric-sdk-go/internal/github.com/hyperledger/fabric/common/crypto"
	fcutils "github.com/hyperledger/fabric-sdk-go/internal/github.com/hyperledger/fabric/common/util"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/context"
	"github.com/hyperledger/fabric-sdk-go/third_party/github.com/hyperledger/fabric/protos/common"
)

// CreateConfigSignature creates a ConfigSignature for the current context.
func CreateConfigSignature(ctx context.Client, config []byte) (*common.ConfigSignature, error) {

	creator, err := ctx.Serialize()
	if err != nil {
		return nil, errors.WithMessage(err, "failed to get user context's identity")
	}

	// generate a random nonce
	nonce, err := crypto.GetRandomNonce()
	if err != nil {
		return nil, errors.WithMessage(err, "nonce creation failed")
	}

	// signature is across a signature header and the config update
	signatureHeader := &common.SignatureHeader{
		Creator: creator,
		Nonce:   nonce,
	}
	signatureHeaderBytes, err := proto.Marshal(signatureHeader)
	if err != nil {
		return nil, errors.Wrap(err, "marshal signatureHeader failed")
	}

	// get all the bytes to be signed together, then sign
	signingBytes := fcutils.ConcatenateBytes(signatureHeaderBytes, config)
	signingMgr := ctx.SigningManager()
	signature, err := signingMgr.Sign(signingBytes, ctx.PrivateKey())
	if err != nil {
		return nil, errors.WithMessage(err, "signing of channel config failed")
	}

	// build the return object
	configSignature := common.ConfigSignature{
		SignatureHeader: signatureHeaderBytes,
		Signature:       signature,
	}
	return &configSignature, nil
}

// ExtractChannelConfig extracts the protobuf 'ConfigUpdate' object out of the 'ConfigEnvelope'.
func ExtractChannelConfig(configEnvelope []byte) ([]byte, error) {

	envelope := &common.Envelope{}
	err := proto.Unmarshal(configEnvelope, envelope)
	if err != nil {
		return nil, errors.Wrap(err, "unmarshal config envelope failed")
	}

	payload := &common.Payload{}
	err = proto.Unmarshal(envelope.Payload, payload)
	if err != nil {
		return nil, errors.Wrap(err, "unmarshal envelope payload failed")
	}

	configUpdateEnvelope := &common.ConfigUpdateEnvelope{}
	err = proto.Unmarshal(payload.Data, configUpdateEnvelope)
	if err != nil {
		return nil, errors.Wrap(err, "unmarshal config update envelope")
	}

	return configUpdateEnvelope.ConfigUpdate, nil
}

// CreateConfigEnvelope creates configuration envelope proto
func CreateConfigEnvelope(data []byte) (*common.ConfigEnvelope, error) {

	envelope := &common.Envelope{}
	if err := proto.Unmarshal(data, envelope); err != nil {
		return nil, errors.Wrap(err, "unmarshal envelope from config block failed")
	}
	payload := &common.Payload{}
	if err := proto.Unmarshal(envelope.Payload, payload); err != nil {
		return nil, errors.Wrap(err, "unmarshal payload from envelope failed")
	}
	channelHeader := &common.ChannelHeader{}
	if err := proto.Unmarshal(payload.Header.ChannelHeader, channelHeader); err != nil {
		return nil, errors.Wrap(err, "unmarshal payload from envelope failed")
	}
	if common.HeaderType(channelHeader.Type) != common.HeaderType_CONFIG {
		return nil, errors.New("block must be of type 'CONFIG'")
	}
	configEnvelope := &common.ConfigEnvelope{}
	if err := proto.Unmarshal(payload.Data, configEnvelope); err != nil {
		return nil, errors.Wrap(err, "unmarshal config envelope failed")
	}

	return configEnvelope, nil
}

// GetLastConfigFromBlock returns the LastConfig data from the given block
func GetLastConfigFromBlock(block *common.Block) (*common.LastConfig, error) {
	if block.Metadata == nil {
		return nil, errors.New("block metadata is nil")
	}
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
