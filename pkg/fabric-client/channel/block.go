/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package channel

import (
	"github.com/golang/protobuf/proto"

	ab "github.com/hyperledger/fabric-sdk-go/internal/github.com/hyperledger/fabric/protos/orderer"
	"github.com/hyperledger/fabric-sdk-go/third_party/github.com/hyperledger/fabric/protos/common"

	ccomm "github.com/hyperledger/fabric-sdk-go/pkg/config/comm"
	"github.com/hyperledger/fabric-sdk-go/pkg/fabric-client/txn"
	"github.com/pkg/errors"
)

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
		ChannelID:   c.name,
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
