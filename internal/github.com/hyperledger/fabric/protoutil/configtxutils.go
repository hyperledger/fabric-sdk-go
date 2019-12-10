/*
Copyright IBM Corp. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/
/*
Notice: This file has been modified for Hyperledger Fabric SDK Go usage.
Please review third_party pinning scripts and patches for more details.
*/
/*
Notice: This file has been modified for Hyperledger Fabric SDK Go usage.
Please review third_party pinning scripts and patches for more details.
*/

package protoutil

import "github.com/hyperledger/fabric-protos-go/common"

func NewConfigGroup() *common.ConfigGroup {
	return &common.ConfigGroup{
		Groups:   make(map[string]*common.ConfigGroup),
		Values:   make(map[string]*common.ConfigValue),
		Policies: make(map[string]*common.ConfigPolicy),
	}
}
