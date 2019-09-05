/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package test

import (
	"fmt"

	"github.com/golang/protobuf/proto"
	"github.com/hyperledger/fabric-protos-go/common"
	"github.com/hyperledger/fabric-protos-go/orderer"
)

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
