/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package resource

import (
	"bytes"
	"encoding/json"
	"fmt"

	"github.com/hyperledger/fabric-sdk-go/internal/github.com/hyperledger/fabric/common/tools/protolator"
	"github.com/hyperledger/fabric-sdk-go/internal/github.com/hyperledger/fabric/protoutil"
	"github.com/hyperledger/fabric-sdk-go/internal/github.com/hyperledger/fabric/sdkinternal/configtxgen/encoder"
	"github.com/pkg/errors"

	"github.com/hyperledger/fabric-sdk-go/internal/github.com/hyperledger/fabric/sdkinternal/configtxgen/localconfig"

	"github.com/hyperledger/fabric-sdk-go/pkg/fab/resource/genesisconfig"
	cb "github.com/hyperledger/fabric-sdk-go/third_party/github.com/hyperledger/fabric/protos/common"
)

// See https://github.com/hyperledger/fabric/blob/be235fd3a236f792a525353d9f9586c8b0d4a61a/cmd/configtxgen/main.go

// CreateGenesisBlock creates a genesis block for a channel
func CreateGenesisBlock(config *genesisconfig.Profile, channelID string) ([]byte, error) {
	localConfig, err := genesisToLocalConfig(config)
	if err != nil {
		return nil, err
	}
	pgen, err := encoder.NewBootstrapper(localConfig)
	if err != nil {
		return nil, errors.WithMessage(err, "could not create bootstrapper")
	}
	logger.Debug("Generating genesis block")
	if config.Orderer == nil {
		return nil, errors.Errorf("refusing to generate block which is missing orderer section")
	}
	if config.Consortiums == nil {
		logger.Warn("Genesis block does not contain a consortiums group definition.  This block cannot be used for orderer bootstrap.")
	}
	genesisBlock := pgen.GenesisBlockForChannel(channelID)
	logger.Debug("Writing genesis block")
	return protoutil.Marshal(genesisBlock)
}

func genesisToLocalConfig(config *genesisconfig.Profile) (*localconfig.Profile, error) {
	b, err := json.Marshal(config)
	if err != nil {
		return nil, err
	}
	c := &localconfig.Profile{}
	err = json.Unmarshal(b, c)
	if err != nil {
		return nil, err
	}
	return c, nil
}

// InspectGenesisBlock inspects a block
func InspectGenesisBlock(data []byte) (string, error) {
	logger.Debug("Parsing genesis block")
	block, err := protoutil.UnmarshalBlock(data)
	if err != nil {
		return "", fmt.Errorf("error unmarshaling to block: %s", err)
	}
	var buf bytes.Buffer
	err = protolator.DeepMarshalJSON(&buf, block)
	if err != nil {
		return "", fmt.Errorf("malformed block contents: %s", err)
	}
	return buf.String(), nil
}

// CreateChannelCreateTx creates a Fabric transaction for creating a channel
func CreateChannelCreateTx(conf, baseProfile *genesisconfig.Profile, channelID string) ([]byte, error) {
	logger.Debug("Generating new channel configtx")

	localConf, err := genesisToLocalConfig(conf)
	if err != nil {
		return nil, err
	}
	localBaseProfile, err := genesisToLocalConfig(baseProfile)
	if err != nil {
		return nil, err
	}

	var configtx *cb.Envelope
	if baseProfile == nil {
		configtx, err = encoder.MakeChannelCreationTransaction(channelID, nil, localConf)
	} else {
		configtx, err = encoder.MakeChannelCreationTransactionWithSystemChannelContext(channelID, nil, localConf, localBaseProfile)
	}
	if err != nil {
		return nil, err
	}

	logger.Debug("Writing new channel tx")
	return protoutil.Marshal(configtx)
}

// InspectChannelCreateTx inspects a Fabric transaction for creating a channel
func InspectChannelCreateTx(data []byte) (string, error) {
	logger.Debug("Parsing transaction")
	env, err := protoutil.UnmarshalEnvelope(data)
	if err != nil {
		return "", fmt.Errorf("Error unmarshaling envelope: %s", err)
	}
	var buf bytes.Buffer
	err = protolator.DeepMarshalJSON(&buf, env)
	if err != nil {
		return "", fmt.Errorf("malformed transaction contents: %s", err)
	}
	return buf.String(), nil
}
