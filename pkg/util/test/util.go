/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package test

import (
	"fmt"

	"github.com/hyperledger/fabric-protos-go/peer"

	"github.com/pkg/errors"

	"github.com/hyperledger/fabric-sdk-go/internal/github.com/hyperledger/fabric/protoutil"

	"github.com/golang/protobuf/proto"
	"github.com/hyperledger/fabric-protos-go/common"
	"github.com/hyperledger/fabric-protos-go/orderer"
)

// AddACL adds an ACL config value to channel config
func AddACL(config *common.Config, policyName, policy string) error {

	aclsConfigValue, ok := config.ChannelGroup.Groups["Application"].Values["ACLs"]
	if !ok {
		return errors.New("ACL missing from Application config")
	}
	acls := &peer.ACLs{}
	err := proto.Unmarshal(aclsConfigValue.Value, acls)
	if err != nil {
		return err
	}
	acls.Acls[policyName] = &peer.APIResource{PolicyRef: policy}
	aclsConfigValue.Value = protoutil.MarshalOrPanic(acls)

	return nil
}

// VerifyACL verifies an ACL config value
func VerifyACL(config *common.Config, expectedPolicyName, expectedPolicy string) error {

	aclsConfigValue, ok := config.ChannelGroup.Groups["Application"].Values["ACLs"]
	if !ok {
		return errors.New("ACL missing from Application config")
	}
	acls := &peer.ACLs{}
	err := proto.Unmarshal(aclsConfigValue.Value, acls)
	if err != nil {
		return err
	}
	resource, ok := acls.Acls[expectedPolicyName]
	if !ok {
		return errors.Errorf("missing expected policy name: %s", expectedPolicyName)
	}
	if resource.PolicyRef != expectedPolicy {
		return errors.Errorf("unexpected policy ref: %s, expected: %s", resource.PolicyRef, expectedPolicy)
	}

	return nil
}

// ModifyMaxMessageCount increments the orderer's BatchSize.MaxMessageCount in a channel config
func ModifyMaxMessageCount(config *common.Config) (uint32, error) {

	// Modify Config
	batchSizeBytes := config.ChannelGroup.Groups["Orderer"].Values["BatchSize"].Value
	batchSize := &orderer.BatchSize{}
	if err := proto.Unmarshal(batchSizeBytes, batchSize); err != nil {
		return 0, err
	}
	batchSize.MaxMessageCount = batchSize.MaxMessageCount + 1
	newMatchSizeBytes, err := proto.Marshal(batchSize)
	if err != nil {
		return 0, err
	}
	config.ChannelGroup.Groups["Orderer"].Values["BatchSize"].Value = newMatchSizeBytes

	return batchSize.MaxMessageCount, nil
}

// VerifyMaxMessageCount verifies the orderer's BatchSize.MaxMessageCount in a channel config
func VerifyMaxMessageCount(config *common.Config, expected uint32) error {

	batchSizeBytes := config.ChannelGroup.Groups["Orderer"].Values["BatchSize"].Value
	batchSize := &orderer.BatchSize{}
	if err := proto.Unmarshal(batchSizeBytes, batchSize); err != nil {
		return err
	}

	if batchSize.MaxMessageCount != expected {
		return fmt.Errorf("Unexpected MaxMessageCount. actual: %d, expected: %d", batchSize.MaxMessageCount, expected)
	}
	return nil
}
