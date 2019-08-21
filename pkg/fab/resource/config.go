/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package resource

import (
	"github.com/golang/protobuf/proto"
	"github.com/pkg/errors"

	"github.com/hyperledger/fabric-protos-go/common"
	"github.com/hyperledger/fabric-sdk-go/internal/github.com/hyperledger/fabric/common/crypto"
	fcutils "github.com/hyperledger/fabric-sdk-go/internal/github.com/hyperledger/fabric/common/util"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/context"
)

// CreateConfigSignature creates a ConfigSignature for the current context
func CreateConfigSignature(ctx context.Client, config []byte) (*common.ConfigSignature, error) {
	cfd, e := GetConfigSignatureData(ctx, config)
	if e != nil {
		return nil, e
	}

	signingMgr := ctx.SigningManager()
	signature, err := signingMgr.Sign(cfd.SigningBytes, ctx.PrivateKey())
	if err != nil {
		return nil, errors.WithMessage(err, "signing of channel config failed")
	}

	// build the return object
	configSignature := common.ConfigSignature{
		SignatureHeader: cfd.SignatureHeaderBytes,
		Signature:       signature,
	}
	return &configSignature, nil
}

// ConfigSignatureData holds data ready to be signed (SigningBytes) + Signature Header
//    When building the common.ConfigSignature instance with the signed SigningBytes from the external tool,
//    assign the returned ConfigSignatureData.SignatureHeader as part of the new ConfigSignature instance.
type ConfigSignatureData struct {
	SignatureHeader      common.SignatureHeader
	SignatureHeaderBytes []byte
	SigningBytes         []byte
}

type identitySerializer interface {
	// Serialize takes an identity object and converts it to the byte representation.
	Serialize() ([]byte, error)
}

// GetConfigSignatureData will prepare a ConfigSignatureData comprising:
// SignatureHeader, its marshaled []byte and the full signing []byte to be used for signing (by an external tool) a Channel Config
func GetConfigSignatureData(creator identitySerializer, config []byte) (signatureHeaderData ConfigSignatureData, e error) {
	creatorBytes, err := creator.Serialize()
	if err != nil {
		e = errors.WithMessage(err, "failed to get user context's identity")
		return
	}

	// generate a random nonce
	nonce, err := crypto.GetRandomNonce()
	if err != nil {
		e = errors.WithMessage(err, "nonce creation failed")
		return
	}

	signatureHeaderData = ConfigSignatureData{}
	// signature is across a signature header and the config update
	signatureHeaderData.SignatureHeader = common.SignatureHeader{
		Creator: creatorBytes,
		Nonce:   nonce,
	}

	signatureHeaderData.SignatureHeaderBytes, err = proto.Marshal(&signatureHeaderData.SignatureHeader)
	if err != nil {
		e = errors.Wrap(err, "marshal signatureHeader failed")
		return
	}

	// get all the bytes to be signed together, then sign
	signatureHeaderData.SigningBytes = fcutils.ConcatenateBytes(signatureHeaderData.SignatureHeaderBytes, config)

	return
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

//CreateConfigUpdateEnvelope creates configuration update envelope proto
func CreateConfigUpdateEnvelope(data []byte) (*common.ConfigUpdateEnvelope, error) {

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
	configEnvelope := &common.ConfigUpdateEnvelope{}
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
