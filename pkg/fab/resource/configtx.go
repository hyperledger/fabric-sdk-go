/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package resource

import (
	"bytes"
	"encoding/json"
	"fmt"

	"github.com/golang/protobuf/proto"

	"github.com/hyperledger/fabric-sdk-go/internal/github.com/hyperledger/fabric/common/channelconfig"
	"github.com/hyperledger/fabric-sdk-go/internal/github.com/hyperledger/fabric/protoutil"
	"github.com/hyperledger/fabric-sdk-go/internal/github.com/hyperledger/fabric/sdkinternal/configtxgen/encoder"
	"github.com/hyperledger/fabric-sdk-go/internal/github.com/hyperledger/fabric/sdkinternal/configtxlator/update"
	"github.com/pkg/errors"

	localconfig "github.com/hyperledger/fabric-sdk-go/internal/github.com/hyperledger/fabric/sdkinternal/configtxgen/genesisconfig"

	"github.com/hyperledger/fabric-config/protolator"
	"github.com/hyperledger/fabric-protos-go/common"
	cb "github.com/hyperledger/fabric-protos-go/common"
	"github.com/hyperledger/fabric-sdk-go/pkg/fab/resource/genesisconfig"
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
	genesisBlock := pgen.GenesisBlockForChannel(channelID)
	logger.Debug("Writing genesis block")
	return protoutil.Marshal(genesisBlock)
}

// CreateGenesisBlockForOrderer creates a genesis block for a channel
func CreateGenesisBlockForOrderer(config *genesisconfig.Profile, channelID string) ([]byte, error) {
	if config.Consortiums == nil {
		return nil, errors.Errorf("Genesis block does not contain a consortiums group definition. This block cannot be used for orderer bootstrap.")
	}
	return CreateGenesisBlock(config, channelID)
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

// InspectBlock inspects a block
func InspectBlock(data []byte) (string, error) {
	if len(data) == 0 {
		return "", fmt.Errorf("missing block")
	}
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

// CreateAnchorPeersUpdate creates an anchor peers update transaction
func CreateAnchorPeersUpdate(conf *genesisconfig.Profile, channelID string, asOrg string) (*common.Envelope, error) {
	logger.Debug("Generating anchor peer update")
	if asOrg == "" {
		return nil, fmt.Errorf("Must specify an organization to update the anchor peer for")
	}

	if conf.Application == nil {
		return nil, fmt.Errorf("Cannot update anchor peers without an application section")
	}

	localConf, err := genesisToLocalConfig(conf)
	if err != nil {
		return nil, err
	}

	original, err := encoder.NewChannelGroup(localConf)
	if err != nil {
		return nil, errors.WithMessage(err, "error parsing profile as channel group")
	}
	original.Groups[channelconfig.ApplicationGroupKey].Version = 1

	updated := proto.Clone(original).(*cb.ConfigGroup)

	originalOrg, ok := original.Groups[channelconfig.ApplicationGroupKey].Groups[asOrg]
	if !ok {
		return nil, errors.Errorf("org with name '%s' does not exist in config", asOrg)
	}

	if _, ok = originalOrg.Values[channelconfig.AnchorPeersKey]; !ok {
		return nil, errors.Errorf("org '%s' does not have any anchor peers defined", asOrg)
	}

	delete(originalOrg.Values, channelconfig.AnchorPeersKey)

	updt, err := update.Compute(&cb.Config{ChannelGroup: original}, &cb.Config{ChannelGroup: updated})
	if err != nil {
		return nil, errors.WithMessage(err, "could not compute update")
	}
	updt.ChannelId = channelID

	newConfigUpdateEnv := &cb.ConfigUpdateEnvelope{
		ConfigUpdate: protoutil.MarshalOrPanic(updt),
	}

	return protoutil.CreateSignedEnvelope(cb.HeaderType_CONFIG_UPDATE, channelID, nil, newConfigUpdateEnv, 0, 0)

}
