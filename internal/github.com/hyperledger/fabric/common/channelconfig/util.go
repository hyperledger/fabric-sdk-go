/*
Copyright IBM Corp. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/
/*
Notice: This file has been modified for Hyperledger Fabric SDK Go usage.
Please review third_party pinning scripts and patches for more details.
*/

package channelconfig

import (
	"math"

	bccsp "github.com/hyperledger/fabric-sdk-go/internal/github.com/hyperledger/fabric/sdkpatch/cryptosuitebridge"
)

const (
	// ReadersPolicyKey is the key used for the read policy
	ReadersPolicyKey = "Readers"

	// WritersPolicyKey is the key used for the read policy
	WritersPolicyKey = "Writers"

	// AdminsPolicyKey is the key used for the read policy
	AdminsPolicyKey = "Admins"

	defaultHashingAlgorithm = bccsp.SHA256

	defaultBlockDataHashingStructureWidth = math.MaxUint32
)
