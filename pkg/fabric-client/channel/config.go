/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package channel

import (
	"github.com/golang/protobuf/proto"

	"github.com/hyperledger/fabric-sdk-go/api/apicryptosuite"
	"github.com/hyperledger/fabric-sdk-go/third_party/github.com/hyperledger/fabric/protos/common"
	mb "github.com/hyperledger/fabric-sdk-go/third_party/github.com/hyperledger/fabric/protos/msp"

	"github.com/hyperledger/fabric-sdk-go/internal/github.com/hyperledger/fabric/msp"
	"github.com/pkg/errors"
)

// ChannelConfig queries for the current config block for this channel.
// This transaction will be made to the orderer.
// @returns {ConfigEnvelope} Object containing the configuration items.
// @see /protos/orderer/ab.proto
// @see /protos/common/configtx.proto
func (c *Channel) ChannelConfig() (*common.ConfigEnvelope, error) {
	logger.Debugf("channelConfig - start for channel %s", c.name)

	// Get the newest block
	block, err := c.block(newNewestSeekPosition())
	if err != nil {
		return nil, err
	}
	logger.Debugf("channelConfig - Retrieved newest block number: %d\n", block.Header.Number)

	// Get the index of the last config block
	lastConfig, err := getLastConfigFromBlock(block)
	if err != nil {
		return nil, errors.Wrap(err, "GetLastConfigFromBlock failed")
	}
	logger.Debugf("channelConfig - Last config index: %d\n", lastConfig.Index)

	// Get the last config block
	block, err = c.block(newSpecificSeekPosition(lastConfig.Index))

	if err != nil {
		return nil, errors.WithMessage(err, "retrieve block failed")
	}
	logger.Debugf("channelConfig - Last config block number %d, Number of tx: %d", block.Header.Number, len(block.Data.Data))

	if len(block.Data.Data) != 1 {
		return nil, errors.New("config block must contain one transaction")
	}

	return createConfigEnvelope(block.Data.Data[0])

}

func createConfigEnvelope(data []byte) (*common.ConfigEnvelope, error) {

	envelope := &common.Envelope{}
	if err := proto.Unmarshal(data, envelope); err != nil {
		return nil, errors.Wrap(err, "unmarshal envelope from config block failed")
	}
	payload := &common.Payload{}
	if err := proto.Unmarshal(envelope.Payload, payload); err != nil {
		return nil, errors.Wrap(err, "unmarshal payload from envelope failed")
	}
	channelHeader := &common.ChannelHeader{}
	if err := proto.Unmarshal(payload.Header.ChannelHeader, channelHeader); err != nil {
		return nil, errors.Wrap(err, "unmarshal payload from envelope failed")
	}
	if common.HeaderType(channelHeader.Type) != common.HeaderType_CONFIG {
		return nil, errors.New("block must be of type 'CONFIG'")
	}
	configEnvelope := &common.ConfigEnvelope{}
	if err := proto.Unmarshal(payload.Data, configEnvelope); err != nil {
		return nil, errors.Wrap(err, "unmarshal config envelope failed")
	}

	return configEnvelope, nil
}

func loadMSPs(mspConfigs []*mb.MSPConfig, cs apicryptosuite.CryptoSuite) ([]msp.MSP, error) {
	logger.Debugf("loadMSPs - start number of msps=%d", len(mspConfigs))

	msps := []msp.MSP{}
	for _, config := range mspConfigs {
		mspType := msp.ProviderType(config.Type)
		if mspType != msp.FABRIC {
			return nil, errors.Errorf("MSP type not supported: %v", mspType)
		}
		if len(config.Config) == 0 {
			return nil, errors.Errorf("MSP configuration missing the payload in the 'Config' property")
		}

		fabricConfig := &mb.FabricMSPConfig{}
		err := proto.Unmarshal(config.Config, fabricConfig)
		if err != nil {
			return nil, errors.Wrap(err, "unmarshal FabricMSPConfig from config failed")
		}

		if fabricConfig.Name == "" {
			return nil, errors.New("MSP Configuration missing name")
		}

		// with this method we are only dealing with verifying MSPs, not local MSPs. Local MSPs are instantiated
		// from user enrollment materials (see User class). For verifying MSPs the root certificates are always
		// required
		if len(fabricConfig.RootCerts) == 0 {
			return nil, errors.New("MSP Configuration missing root certificates required for validating signing certificates")
		}

		// get the application org names
		var orgs []string
		orgUnits := fabricConfig.OrganizationalUnitIdentifiers
		for _, orgUnit := range orgUnits {
			logger.Debugf("loadMSPs - found org of :: %s", orgUnit.OrganizationalUnitIdentifier)
			orgs = append(orgs, orgUnit.OrganizationalUnitIdentifier)
		}

		// TODO: Do something with orgs
		// TODO: Configure MSP version (rather than MSP 1.0)
		newMSP, err := msp.NewBccspMsp(msp.MSPv1_0, cs)
		if err != nil {
			return nil, errors.Wrap(err, "instantiate MSP failed")
		}

		if err := newMSP.Setup(config); err != nil {
			return nil, errors.Wrap(err, "configure MSP failed")
		}

		mspID, _ := newMSP.GetIdentifier()
		logger.Debugf("loadMSPs - adding msp=%s", mspID)

		msps = append(msps, newMSP)
	}

	logger.Debugf("loadMSPs - loaded %d MSPs", len(msps))
	return msps, nil
}

// getLastConfigFromBlock returns the LastConfig data from the given block
func getLastConfigFromBlock(block *common.Block) (*common.LastConfig, error) {
	if block.Metadata == nil {
		return nil, errors.New("block metadata is nil")
	}
	metadata := &common.Metadata{}
	err := proto.Unmarshal(block.Metadata.Metadata[common.BlockMetadataIndex_LAST_CONFIG], metadata)
	if err != nil {
		return nil, errors.Wrap(err, "unmarshal block metadata failed")
	}

	lastConfig := &common.LastConfig{}
	err = proto.Unmarshal(metadata.Value, lastConfig)
	if err != nil {
		return nil, errors.Wrap(err, "unmarshal last config from metadata failed")
	}

	return lastConfig, err
}
