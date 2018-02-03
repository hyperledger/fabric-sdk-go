/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package channel

import (
	"github.com/golang/protobuf/proto"

	ab "github.com/hyperledger/fabric-sdk-go/internal/github.com/hyperledger/fabric/protos/orderer"
	"github.com/hyperledger/fabric-sdk-go/third_party/github.com/hyperledger/fabric/protos/common"
	protos_utils "github.com/hyperledger/fabric-sdk-go/third_party/github.com/hyperledger/fabric/protos/utils"

	ccomm "github.com/hyperledger/fabric-sdk-go/pkg/config/comm"
	"github.com/hyperledger/fabric-sdk-go/pkg/fabric-client/txn"
	"github.com/pkg/errors"
)

// GenesisBlock returns the genesis block from the defined orderer that may be
// used in a join request
// request: An object containing the following fields:
//          `txId` : required - String of the transaction id
//          `nonce` : required - Integer of the once time number
//
// See /protos/peer/proposal_response.proto
func (c *Channel) GenesisBlock() (*common.Block, error) {
	logger.Debug("GenesisBlock - start")

	// verify that we have an orderer configured
	if len(c.Orderers()) == 0 {
		return nil, errors.New("GenesisBlock missing orderer assigned to this channel for the GenesisBlock request")
	}

	txnID, err := txn.NewID(c.clientContext)
	if err != nil {
		return nil, errors.WithMessage(err, "failed to calculate transaction id")
	}

	// now build the seek info , will be used once the channel is created
	// to get the genesis block back
	seekStart := newSpecificSeekPosition(0)
	seekStop := newSpecificSeekPosition(0)
	seekInfo := &ab.SeekInfo{
		Start:    seekStart,
		Stop:     seekStop,
		Behavior: ab.SeekInfo_BLOCK_UNTIL_READY,
	}

	tlsCertHash := ccomm.TLSCertHash(c.clientContext.Config())
	channelHeaderOpts := txn.ChannelHeaderOpts{
		ChannelID:   c.Name(),
		TxnID:       txnID,
		TLSCertHash: tlsCertHash,
	}
	seekInfoHeader, err := txn.CreateChannelHeader(common.HeaderType_DELIVER_SEEK_INFO, channelHeaderOpts)
	if err != nil {
		return nil, errors.Wrap(err, "BuildChannelHeader failed")
	}
	seekHeader, err := txn.CreateHeader(c.clientContext, seekInfoHeader, txnID)
	if err != nil {
		return nil, errors.Wrap(err, "BuildHeader failed")
	}
	seekPayload := &common.Payload{
		Header: seekHeader,
		Data:   protos_utils.MarshalOrPanic(seekInfo),
	}
	seekPayloadBytes := protos_utils.MarshalOrPanic(seekPayload)

	signedEnvelope, err := txn.SignPayload(c.clientContext, seekPayloadBytes)
	if err != nil {
		return nil, errors.WithMessage(err, "SignPayload failed")
	}

	block, err := txn.SendEnvelope(c.clientContext, signedEnvelope, c.Orderers())
	if err != nil {
		return nil, errors.WithMessage(err, "SendEnvelope failed")
	}
	return block, nil
}

// createSeekGenesisBlockRequest creates a seek request for block 0 on the specified
// channel. This request is sent to the ordering service to request blocks
func createSeekGenesisBlockRequest(channelName string, creator []byte) []byte {
	return protos_utils.MarshalOrPanic(&common.Payload{
		Header: &common.Header{
			ChannelHeader: protos_utils.MarshalOrPanic(&common.ChannelHeader{
				ChannelId: channelName,
			}),
			SignatureHeader: protos_utils.MarshalOrPanic(&common.SignatureHeader{Creator: creator}),
		},
		Data: protos_utils.MarshalOrPanic(&ab.SeekInfo{
			Start:    &ab.SeekPosition{Type: &ab.SeekPosition_Specified{Specified: &ab.SeekSpecified{Number: 0}}},
			Stop:     &ab.SeekPosition{Type: &ab.SeekPosition_Specified{Specified: &ab.SeekSpecified{Number: 0}}},
			Behavior: ab.SeekInfo_BLOCK_UNTIL_READY,
		}),
	})
}

// block retrieves the block at the given position
func (c *Channel) block(pos *ab.SeekPosition) (*common.Block, error) {

	creator, err := c.clientContext.Identity()
	if err != nil {
		return nil, errors.WithMessage(err, "serializing identity failed")
	}

	txnID, err := txn.NewID(c.clientContext)
	if err != nil {
		return nil, errors.Wrap(err, "generating TX ID failed")
	}

	channelHeaderOpts := txn.ChannelHeaderOpts{
		ChannelID:   c.Name(),
		TxnID:       txnID,
		TLSCertHash: ccomm.TLSCertHash(c.clientContext.Config()),
	}
	seekInfoHeader, err := txn.CreateChannelHeader(common.HeaderType_DELIVER_SEEK_INFO, channelHeaderOpts)
	if err != nil {
		return nil, errors.Wrap(err, "NewChannelHeader failed")
	}

	seekInfoHeaderBytes, err := proto.Marshal(seekInfoHeader)
	if err != nil {
		return nil, errors.Wrap(err, "marshal seek info failed")
	}

	signatureHeader := &common.SignatureHeader{
		Creator: creator,
		Nonce:   txnID.Nonce,
	}

	signatureHeaderBytes, err := proto.Marshal(signatureHeader)
	if err != nil {
		return nil, errors.Wrap(err, "marshal signature header failed")
	}

	seekHeader := &common.Header{
		ChannelHeader:   seekInfoHeaderBytes,
		SignatureHeader: signatureHeaderBytes,
	}

	seekInfo := &ab.SeekInfo{
		Start:    pos,
		Stop:     pos,
		Behavior: ab.SeekInfo_BLOCK_UNTIL_READY,
	}

	seekInfoBytes, err := proto.Marshal(seekInfo)
	if err != nil {
		return nil, errors.Wrap(err, "marshal seek info failed")
	}

	seekPayload := &common.Payload{
		Header: seekHeader,
		Data:   seekInfoBytes,
	}

	seekPayloadBytes, err := proto.Marshal(seekPayload)
	if err != nil {
		return nil, err
	}

	signedEnvelope, err := txn.SignPayload(c.clientContext, seekPayloadBytes)
	if err != nil {
		return nil, errors.WithMessage(err, "SignPayload failed")
	}

	return txn.SendEnvelope(c.clientContext, signedEnvelope, c.Orderers())
}

// newNewestSeekPosition returns a SeekPosition that requests the newest block
func newNewestSeekPosition() *ab.SeekPosition {
	return &ab.SeekPosition{Type: &ab.SeekPosition_Newest{Newest: &ab.SeekNewest{}}}
}

// newSpecificSeekPosition returns a SeekPosition that requests the block at the given index
func newSpecificSeekPosition(index uint64) *ab.SeekPosition {
	return &ab.SeekPosition{Type: &ab.SeekPosition_Specified{Specified: &ab.SeekSpecified{Number: index}}}
}
