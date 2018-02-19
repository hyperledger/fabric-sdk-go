/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package txn

import (
	"encoding/hex"
	"hash"
	"time"

	"github.com/hyperledger/fabric-sdk-go/pkg/cryptosuite"

	"github.com/golang/protobuf/proto"
	"github.com/golang/protobuf/ptypes"
	"github.com/pkg/errors"

	fab "github.com/hyperledger/fabric-sdk-go/api/apifabclient"

	"github.com/hyperledger/fabric-sdk-go/internal/github.com/hyperledger/fabric/common/crypto"
	"github.com/hyperledger/fabric-sdk-go/third_party/github.com/hyperledger/fabric/protos/common"
	pb "github.com/hyperledger/fabric-sdk-go/third_party/github.com/hyperledger/fabric/protos/peer"
)

// NewID computes a TransactionID for the current user context
//
// TODO: Determine if this function should be exported after refactoring is completed.
func NewID(ctx fab.Context) (fab.TransactionID, error) {
	// generate a random nonce
	nonce, err := crypto.GetRandomNonce()
	if err != nil {
		return fab.TransactionID{}, errors.WithMessage(err, "nonce creation failed")
	}

	creator, err := ctx.Identity()
	if err != nil {
		return fab.TransactionID{}, errors.WithMessage(err, "identity from context failed")
	}

	ho := cryptosuite.GetSHA256Opts() // TODO: make configurable
	h, err := ctx.CryptoSuite().GetHash(ho)
	if err != nil {
		return fab.TransactionID{}, errors.WithMessage(err, "hash function creation failed")
	}

	id, err := computeTxnID(nonce, creator, h)
	if err != nil {
		return fab.TransactionID{}, errors.WithMessage(err, "txn ID computation failed")
	}

	txnID := fab.TransactionID{
		ID:    id,
		Nonce: nonce,
	}

	return txnID, nil
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

// SignPayload signs payload
//
// TODO: Determine if this function should be exported after refactoring is completed.
func SignPayload(ctx context, payload []byte) (*fab.SignedEnvelope, error) {
	signingMgr := ctx.SigningManager()
	signature, err := signingMgr.Sign(payload, ctx.PrivateKey())
	if err != nil {
		return nil, err
	}
	return &fab.SignedEnvelope{Payload: payload, Signature: signature}, nil
}

// ChannelHeaderOpts holds the parameters to create a ChannelHeader.
type ChannelHeaderOpts struct {
	ChannelID   string
	TxnID       fab.TransactionID
	Epoch       uint64
	ChaincodeID string
	Timestamp   time.Time
	TLSCertHash []byte
}

// CreateChannelHeader is a utility method to build a common chain header (TODO refactor)
//
// TODO: Determine if this function should be exported after refactoring is completed.
func CreateChannelHeader(headerType common.HeaderType, opts ChannelHeaderOpts) (*common.ChannelHeader, error) {
	logger.Debugf("buildChannelHeader - headerType: %s channelID: %s txID: %d epoch: % chaincodeID: %s timestamp: %v", headerType, opts.ChannelID, opts.TxnID.ID, opts.Epoch, opts.ChaincodeID, opts.Timestamp)
	channelHeader := &common.ChannelHeader{
		Type:        int32(headerType),
		ChannelId:   opts.ChannelID,
		TxId:        opts.TxnID.ID,
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

// CreateHeader creates a Header from a ChannelHeader.
func CreateHeader(ctx fab.IdentityContext, channelHeader *common.ChannelHeader, txnID fab.TransactionID) (*common.Header, error) {
	creator, err := ctx.Identity()
	if err != nil {
		return nil, errors.WithMessage(err, "extracting creator from identity context failed")
	}

	signatureHeader := &common.SignatureHeader{
		Creator: creator,
		Nonce:   txnID.Nonce,
	}
	sh, err := proto.Marshal(signatureHeader)
	if err != nil {
		return nil, errors.Wrap(err, "marshal signatureHeader failed")
	}
	ch, err := proto.Marshal(channelHeader)
	if err != nil {
		return nil, errors.Wrap(err, "marshal channelHeader failed")
	}
	header := &common.Header{
		SignatureHeader: sh,
		ChannelHeader:   ch,
	}
	return header, nil
}
