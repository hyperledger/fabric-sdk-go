/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package txn

import (
	"encoding/hex"
	"hash"
	"time"

	"github.com/golang/protobuf/proto"
	"github.com/golang/protobuf/ptypes"
	"github.com/pkg/errors"

	"github.com/hyperledger/fabric-protos-go/common"
	pb "github.com/hyperledger/fabric-protos-go/peer"
	"github.com/hyperledger/fabric-sdk-go/internal/github.com/hyperledger/fabric/common/crypto"
	contextApi "github.com/hyperledger/fabric-sdk-go/pkg/common/providers/context"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/fab"
	"github.com/hyperledger/fabric-sdk-go/pkg/core/cryptosuite"
)

// TransactionHeader contains metadata for a transaction created by the SDK.
type TransactionHeader struct {
	id        fab.TransactionID
	creator   []byte
	nonce     []byte
	channelID string
}

// TransactionID returns the transaction's computed identifier.
func (th *TransactionHeader) TransactionID() fab.TransactionID {
	return th.id
}

// Creator returns the transaction creator's identity bytes.
func (th *TransactionHeader) Creator() []byte {
	return th.creator
}

// Nonce returns the transaction's generated nonce.
func (th *TransactionHeader) Nonce() []byte {
	return th.nonce
}

// ChannelID returns the transaction's target channel identifier.
func (th *TransactionHeader) ChannelID() string {
	return th.channelID
}

// NewHeader computes a TransactionID from the current user context and holds
// metadata to create transaction proposals.
func NewHeader(ctx contextApi.Client, channelID string, opts ...fab.TxnHeaderOpt) (*TransactionHeader, error) {
	var options fab.TxnHeaderOptions
	for _, opt := range opts {
		opt(&options)
	}

	nonce := options.Nonce
	if nonce == nil {
		// generate a random nonce
		var err error
		nonce, err = crypto.GetRandomNonce()
		if err != nil {
			return nil, errors.WithMessage(err, "nonce creation failed")
		}
	}

	creator := options.Creator
	if creator == nil {
		var err error
		creator, err = ctx.Serialize()
		if err != nil {
			return nil, errors.WithMessage(err, "identity from context failed")
		}
	}

	ho := cryptosuite.GetSHA256Opts() // TODO: make configurable
	h, err := ctx.CryptoSuite().GetHash(ho)
	if err != nil {
		return nil, errors.WithMessage(err, "hash function creation failed")
	}

	id, err := computeTxnID(nonce, creator, h)
	if err != nil {
		return nil, errors.WithMessage(err, "txn ID computation failed")
	}

	txnID := TransactionHeader{
		id:        fab.TransactionID(id),
		creator:   creator,
		nonce:     nonce,
		channelID: channelID,
	}

	return &txnID, nil
}

func computeTxnID(nonce, creator []byte, h hash.Hash) (string, error) {
	b := append(nonce, creator...)

	_, err := h.Write(b)
	if err != nil {
		return "", err
	}
	digest := h.Sum(nil)
	id := hex.EncodeToString(digest)

	return id, nil
}

// signPayload signs payload
func signPayload(ctx contextApi.Client, payload *common.Payload) (*fab.SignedEnvelope, error) {
	payloadBytes, err := proto.Marshal(payload)
	if err != nil {
		return nil, errors.WithMessage(err, "marshaling of payload failed")
	}

	signingMgr := ctx.SigningManager()
	signature, err := signingMgr.Sign(payloadBytes, ctx.PrivateKey())
	if err != nil {
		return nil, errors.WithMessage(err, "signing of payload failed")
	}
	return &fab.SignedEnvelope{Payload: payloadBytes, Signature: signature}, nil
}

// ChannelHeaderOpts holds the parameters to create a ChannelHeader.
type ChannelHeaderOpts struct {
	TxnHeader   *TransactionHeader
	Epoch       uint64
	ChaincodeID string
	Timestamp   time.Time
	TLSCertHash []byte
}

// CreateChannelHeader is a utility method to build a common chain header (TODO refactor)
//
// TODO: Determine if this function should be exported after refactoring is completed.
func CreateChannelHeader(headerType common.HeaderType, opts ChannelHeaderOpts) (*common.ChannelHeader, error) {
	logger.Debugf("buildChannelHeader - headerType: %s channelID: %s txID: %d epoch: %d chaincodeID: %s timestamp: %v", headerType, opts.TxnHeader.channelID, opts.TxnHeader.id, opts.Epoch, opts.ChaincodeID, opts.Timestamp)
	channelHeader := &common.ChannelHeader{
		Type:        int32(headerType),
		ChannelId:   opts.TxnHeader.channelID,
		TxId:        string(opts.TxnHeader.id),
		Epoch:       opts.Epoch,
		TlsCertHash: opts.TLSCertHash,
	}

	if opts.Timestamp.IsZero() {
		opts.Timestamp = time.Now()
	}

	ts, err := ptypes.TimestampProto(opts.Timestamp)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create timestamp in channel header")
	}
	channelHeader.Timestamp = ts

	if opts.ChaincodeID != "" {
		ccID := &pb.ChaincodeID{
			Name: opts.ChaincodeID,
		}
		headerExt := &pb.ChaincodeHeaderExtension{
			ChaincodeId: ccID,
		}
		headerExtBytes, err := proto.Marshal(headerExt)
		if err != nil {
			return nil, errors.Wrap(err, "marshal header extension failed")
		}
		channelHeader.Extension = headerExtBytes
	}
	return channelHeader, nil
}

// createHeader creates a Header from a ChannelHeader.
func createHeader(th *TransactionHeader, channelHeader *common.ChannelHeader) (*common.Header, error) { //nolint

	signatureHeader := &common.SignatureHeader{
		Creator: th.creator,
		Nonce:   th.nonce,
	}
	sh, err := proto.Marshal(signatureHeader)
	if err != nil {
		return nil, errors.Wrap(err, "marshal signatureHeader failed")
	}
	ch, err := proto.Marshal(channelHeader)
	if err != nil {
		return nil, errors.Wrap(err, "marshal channelHeader failed")
	}
	header := common.Header{
		SignatureHeader: sh,
		ChannelHeader:   ch,
	}
	return &header, nil
}

// CreatePayload creates a slice of payload bytes from a ChannelHeader and a data slice.
func CreatePayload(txh *TransactionHeader, channelHeader *common.ChannelHeader, data []byte) (*common.Payload, error) {
	header, err := createHeader(txh, channelHeader)
	if err != nil {
		return nil, errors.Wrap(err, "header creation failed")
	}

	payload := common.Payload{
		Header: header,
		Data:   data,
	}

	return &payload, nil
}

// CreateSignatureHeader creates a SignatureHeader based on the nonce and creator of the transaction header.
func CreateSignatureHeader(txh *TransactionHeader) (*common.SignatureHeader, error) {

	signatureHeader := common.SignatureHeader{
		Creator: txh.creator,
		Nonce:   txh.nonce,
	}

	return &signatureHeader, nil
}
