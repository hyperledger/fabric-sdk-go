/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package channel

import (
	"time"

	"github.com/golang/protobuf/proto"

	ab "github.com/hyperledger/fabric-sdk-go/internal/github.com/hyperledger/fabric/protos/orderer"
	"github.com/hyperledger/fabric-sdk-go/third_party/github.com/hyperledger/fabric/protos/common"
	protos_utils "github.com/hyperledger/fabric-sdk-go/third_party/github.com/hyperledger/fabric/protos/utils"

	ccomm "github.com/hyperledger/fabric-sdk-go/pkg/config/comm"
	fc "github.com/hyperledger/fabric-sdk-go/pkg/fabric-client/internal"
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
	creator, err := c.clientContext.Identity()
	if err != nil {
		return nil, errors.WithMessage(err, "failed to get creator identity")
	}

	txnID, err := txn.NewID(c.clientContext)
	if err != nil {
		return nil, errors.WithMessage(err, "failed to calculate transaction id")
	}

	// now build the seek info , will be used once the channel is created
	// to get the genesis block back
	seekStart := fc.NewSpecificSeekPosition(0)
	seekStop := fc.NewSpecificSeekPosition(0)
	seekInfo := &ab.SeekInfo{
		Start:    seekStart,
		Stop:     seekStop,
		Behavior: ab.SeekInfo_BLOCK_UNTIL_READY,
	}
	protos_utils.MakeChannelHeader(common.HeaderType_DELIVER_SEEK_INFO, 1, c.Name(), 0)
	tlsCertHash := ccomm.TLSCertHash(c.clientContext.Config())
	seekInfoHeader, err := txn.BuildChannelHeader(common.HeaderType_DELIVER_SEEK_INFO, c.Name(), txnID.ID, 0, "", time.Now(), tlsCertHash)
	if err != nil {
		return nil, errors.Wrap(err, "BuildChannelHeader failed")
	}
	seekHeader, err := fc.BuildHeader(creator, seekInfoHeader, txnID.Nonce)
	if err != nil {
		return nil, errors.Wrap(err, "BuildHeader failed")
	}
	seekPayload := &common.Payload{
		Header: seekHeader,
		Data:   fc.MarshalOrPanic(seekInfo),
	}
	seekPayloadBytes := fc.MarshalOrPanic(seekPayload)

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

// block retrieves the block at the given position
func (c *Channel) block(pos *ab.SeekPosition) (*common.Block, error) {
	nonce, err := fc.GenerateRandomNonce()
	if err != nil {
		return nil, errors.Wrap(err, "GenerateRandomNonce failed")
	}

	creator, err := c.clientContext.Identity()
	if err != nil {
		return nil, errors.WithMessage(err, "serializing identity failed")
	}

	txID, err := protos_utils.ComputeProposalTxID(nonce, creator)
	if err != nil {
		return nil, errors.Wrap(err, "generating TX ID failed")
	}

	tlsCertHash := ccomm.TLSCertHash(c.clientContext.Config())
	seekInfoHeader, err := txn.BuildChannelHeader(common.HeaderType_DELIVER_SEEK_INFO, c.Name(), txID, 0, "", time.Now(), tlsCertHash)
	if err != nil {
		return nil, errors.Wrap(err, "BuildChannelHeader failed")
	}

	seekInfoHeaderBytes, err := proto.Marshal(seekInfoHeader)
	if err != nil {
		return nil, errors.Wrap(err, "marshal seek info failed")
	}

	signatureHeader := &common.SignatureHeader{
		Creator: creator,
		Nonce:   nonce,
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
