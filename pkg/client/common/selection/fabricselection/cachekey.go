/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package fabricselection

import (
	"encoding/json"

	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/fab"
)

type cacheKey struct {
	chaincodes []*fab.ChaincodeCall
}

func newCacheKey(chaincodes []*fab.ChaincodeCall) *cacheKey {
	return &cacheKey{chaincodes: chaincodes}
}

func (k *cacheKey) String() string {
	bytes, err := json.Marshal(k.chaincodes)
	if err != nil {
		logger.Errorf("unexpected error marshalling chaincodes: %s", err)
		return ""
	}
	return string(bytes)
}
