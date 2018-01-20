/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package internal

import (
	"github.com/hyperledger/fabric-sdk-go/api/apifabclient"
	"github.com/hyperledger/fabric-sdk-go/internal/github.com/hyperledger/fabric/common/crypto"

	"github.com/hyperledger/fabric-sdk-go/api/apitxn"
	"github.com/hyperledger/fabric-sdk-go/pkg/logging"
	protos_utils "github.com/hyperledger/fabric-sdk-go/third_party/github.com/hyperledger/fabric/protos/utils"
)

var logger = logging.NewLogger("fabric_sdk_go")

// NewTxnID computes a TransactionID for the current user context
func NewTxnID(signingIdentity apifabclient.IdentityContext) (apitxn.TransactionID, error) {
	// generate a random nonce
	nonce, err := crypto.GetRandomNonce()
	if err != nil {
		return apitxn.TransactionID{}, err
	}

	creator, err := signingIdentity.Identity()
	if err != nil {
		return apitxn.TransactionID{}, err
	}

	id, err := protos_utils.ComputeProposalTxID(nonce, creator)
	if err != nil {
		return apitxn.TransactionID{}, err
	}

	txnID := apitxn.TransactionID{
		ID:    id,
		Nonce: nonce,
	}

	return txnID, nil
}
