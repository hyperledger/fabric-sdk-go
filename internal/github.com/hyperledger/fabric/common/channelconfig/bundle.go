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
	flogging "github.com/hyperledger/fabric-sdk-go/internal/github.com/hyperledger/fabric/sdkpatch/logbridge"
)

var logger = flogging.MustGetLogger("common.channelconfig")

// RootGroupKey is the key for namespacing the channel config, especially for
// policy evaluation.
const RootGroupKey = "Channel"
