/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

// Package resource provides access to fabric network resource management, typically using system channel queries.
package resource

import (
	reqContext "context"

	"github.com/golang/protobuf/proto"
	"github.com/hyperledger/fabric-protos-go/common"
	ab "github.com/hyperledger/fabric-protos-go/orderer"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/errors/retry"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/fab"
	contextImpl "github.com/hyperledger/fabric-sdk-go/pkg/context"
	ccomm "github.com/hyperledger/fabric-sdk-go/pkg/core/config/comm"
	"github.com/hyperledger/fabric-sdk-go/pkg/fab/txn"
	"github.com/pkg/errors"
)

// block retrieves the block at the given position
func retrieveBlock(reqCtx reqContext.Context, orderers []fab.Orderer, channel string, pos *ab.SeekPosition, opts options) (*common.Block, error) {
	ctx, ok := contextImpl.RequestClientContext(reqCtx)
	if !ok {
		return nil, errors.New("failed get client context from reqContext for signPayload")
	}
	th, err := txn.NewHeader(ctx, channel)
	if err != nil {
		return nil, errors.Wrap(err, "generating TX ID failed")
	}

	hash, err := ccomm.TLSCertHash(ctx.EndpointConfig())
	if err != nil {
		return nil, errors.Wrapf(err, "failed to get tls cert hash")
	}

	channelHeaderOpts := txn.ChannelHeaderOpts{
		TxnHeader:   th,
		TLSCertHash: hash,
	}

	seekInfoHeader, err := txn.CreateChannelHeader(common.HeaderType_DELIVER_SEEK_INFO, channelHeaderOpts)
	if err != nil {
		return nil, errors.Wrap(err, "CreateChannelHeader failed")
	}

	seekInfoHeaderBytes, err := proto.Marshal(seekInfoHeader)
	if err != nil {
		return nil, errors.Wrap(err, "marshal seek info failed")
	}

	signatureHeader, err := txn.CreateSignatureHeader(th)
	if err != nil {
		return nil, errors.Wrap(err, "CreateSignatureHeader failed")
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

	payload := common.Payload{
		Header: seekHeader,
		Data:   seekInfoBytes,
	}

	resp, err := retry.NewInvoker(retry.New(opts.retry)).Invoke(
		func() (interface{}, error) {
			return txn.SendPayload(reqCtx, &payload, orderers)
		},
	)
	if err != nil {
		return nil, err
	}
	return resp.(*common.Block), err
}

// newNewestSeekPosition returns a SeekPosition that requests the newest block
func newNewestSeekPosition() *ab.SeekPosition {
	return &ab.SeekPosition{Type: &ab.SeekPosition_Newest{Newest: &ab.SeekNewest{}}}
}

// newSpecificSeekPosition returns a SeekPosition that requests the block at the given index
func newSpecificSeekPosition(index uint64) *ab.SeekPosition {
	return &ab.SeekPosition{Type: &ab.SeekPosition_Specified{Specified: &ab.SeekSpecified{Number: index}}}
}
