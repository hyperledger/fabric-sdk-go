/*
Copyright State Street Corp. 2018 All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/
/*
Notice: This file has been modified for Hyperledger Fabric SDK Go usage.
Please review third_party pinning scripts and patches for more details.
*/

package channelconfig

import (
	pb "github.com/hyperledger/fabric-protos-go/peer"
)

// aclsProvider provides mappings for resource to policy names
type aclsProvider struct {
	aclPolicyRefs map[string]string
}

func (ag *aclsProvider) PolicyRefForAPI(aclName string) string {
	return ag.aclPolicyRefs[aclName]
}

// this translates policies to absolute paths if needed
func newAPIsProvider(acls map[string]*pb.APIResource) *aclsProvider {
	aclPolicyRefs := make(map[string]string)

	for key, acl := range acls {
		// If the policy is fully qualified, ie to /Channel/Application/Readers leave it alone
		// otherwise, make it fully qualified referring to /Channel/Application/policyName
		if '/' != acl.PolicyRef[0] {
			aclPolicyRefs[key] = "/" + ChannelGroupKey + "/" + ApplicationGroupKey + "/" + acl.PolicyRef
		} else {
			aclPolicyRefs[key] = acl.PolicyRef
		}
	}

	return &aclsProvider{
		aclPolicyRefs: aclPolicyRefs,
	}
}
