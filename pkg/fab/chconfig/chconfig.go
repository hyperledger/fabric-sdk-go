/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package chconfig

import (
	reqContext "context"
	"math/rand"

	"github.com/golang/protobuf/proto"

	"github.com/hyperledger/fabric-sdk-go/pkg/common/logging"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/fab"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/msp"
	"github.com/hyperledger/fabric-sdk-go/pkg/fab/channel"
	"github.com/hyperledger/fabric-sdk-go/pkg/fab/resource"
	"github.com/hyperledger/fabric-sdk-go/third_party/github.com/hyperledger/fabric/protos/common"
	mb "github.com/hyperledger/fabric-sdk-go/third_party/github.com/hyperledger/fabric/protos/msp"
	ab "github.com/hyperledger/fabric-sdk-go/third_party/github.com/hyperledger/fabric/protos/orderer"
	pb "github.com/hyperledger/fabric-sdk-go/third_party/github.com/hyperledger/fabric/protos/peer"
	"github.com/pkg/errors"

	channelConfig "github.com/hyperledger/fabric-sdk-go/internal/github.com/hyperledger/fabric/common/channelconfig"

	imsp "github.com/hyperledger/fabric-sdk-go/internal/github.com/hyperledger/fabric/msp"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/context"
	contextImpl "github.com/hyperledger/fabric-sdk-go/pkg/context"
)

var logger = logging.NewLogger("fabsdk/fab")

const (
	defaultMinResponses = 1
	defaultMaxTargets   = 1
)

// Opts contains options for retrieving channel configuration
type Opts struct {
	Orderer      fab.Orderer // if configured, channel config will be retrieved from this orderer
	Targets      []fab.Peer  // if configured, channel config will be retrieved from peers (targets)
	MinResponses int         // used with targets option; min number of success responses (from targets/peers)
	MaxTargets   int         //if configured, channel config will be retrieved for these number of random targets
}

// Option func for each Opts argument
type Option func(opts *Opts) error

// Context holds the providers and identity
type Context struct {
	context.Providers
	msp.Identity
}

// ChannelConfig implements query channel configuration
type ChannelConfig struct {
	channelID string
	opts      Opts
}

// ChannelCfg contains channel configuration
type ChannelCfg struct {
	id          string
	msps        []*mb.MSPConfig
	anchorPeers []*fab.OrgAnchorPeer
	orderers    []string
	versions    *fab.Versions
}

// NewChannelCfg creates channel cfg
// TODO: This is temporary, Remove once we have config injected in sdk
func NewChannelCfg(channelID string) *ChannelCfg {
	return &ChannelCfg{id: channelID}
}

// ID returns the channel ID
func (cfg *ChannelCfg) ID() string {
	return cfg.id
}

// MSPs returns msps
func (cfg *ChannelCfg) MSPs() []*mb.MSPConfig {
	return cfg.msps
}

// AnchorPeers returns anchor peers
func (cfg *ChannelCfg) AnchorPeers() []*fab.OrgAnchorPeer {
	return cfg.anchorPeers
}

// Orderers returns orderers
func (cfg *ChannelCfg) Orderers() []string {
	return cfg.orderers
}

// Versions returns versions
func (cfg *ChannelCfg) Versions() *fab.Versions {
	return cfg.versions
}

// New channel config implementation
func New(channelID string, options ...Option) (*ChannelConfig, error) {
	opts, err := prepareOpts(options...)
	if err != nil {
		return nil, err
	}

	return &ChannelConfig{channelID: channelID, opts: opts}, nil
}

// Query returns channel configuration
func (c *ChannelConfig) Query(reqCtx reqContext.Context) (fab.ChannelCfg, error) {

	if c.opts.Orderer != nil {
		return c.queryOrderer(reqCtx)
	}

	return c.queryPeers(reqCtx)
}

func (c *ChannelConfig) queryPeers(reqCtx reqContext.Context) (*ChannelCfg, error) {

	ctx, ok := contextImpl.RequestClientContext(reqCtx)
	if !ok {
		return nil, errors.New("failed get client context from reqContext for signPayload")
	}

	l, err := channel.NewLedger(c.channelID)
	if err != nil {
		return nil, errors.WithMessage(err, "ledger client creation failed")
	}

	targets := []fab.ProposalProcessor{}
	if c.opts.Targets == nil {

		// Calculate targets from config
		chPeers, err := ctx.Config().ChannelPeers(c.channelID)
		if err != nil {
			return nil, errors.WithMessage(err, "read configuration for channel peers failed")
		}

		for _, p := range chPeers {
			newPeer, err := ctx.InfraProvider().CreatePeerFromConfig((&p.NetworkPeer))
			if err != nil || newPeer == nil {
				return nil, errors.WithMessage(err, "NewPeer failed")
			}

			targets = append(targets, newPeer)
		}

		targets = randomMaxTargets(targets, c.opts.MaxTargets)

	} else {
		targets = peersToTxnProcessors(c.opts.Targets)
	}

	configEnvelope, err := l.QueryConfigBlock(reqCtx, targets, &channel.TransactionProposalResponseVerifier{MinResponses: c.opts.MinResponses})
	if err != nil {
		return nil, errors.WithMessage(err, "QueryBlockConfig failed")
	}

	return extractConfig(c.channelID, configEnvelope)
}

func (c *ChannelConfig) queryOrderer(reqCtx reqContext.Context) (*ChannelCfg, error) {

	configEnvelope, err := resource.LastConfigFromOrderer(reqCtx, c.channelID, c.opts.Orderer)
	if err != nil {
		return nil, errors.WithMessage(err, "LastConfigFromOrderer failed")
	}

	return extractConfig(c.channelID, configEnvelope)
}

// WithPeers encapsulates peers to Option
func WithPeers(peers []fab.Peer) Option {
	return func(opts *Opts) error {
		opts.Targets = peers
		return nil
	}
}

// WithMinResponses encapsulates minimum responses to Option
func WithMinResponses(min int) Option {
	return func(opts *Opts) error {
		opts.MinResponses = min
		return nil
	}
}

// WithOrderer encapsulates orderer to Option
func WithOrderer(orderer fab.Orderer) Option {
	return func(opts *Opts) error {
		opts.Orderer = orderer
		return nil
	}
}

// WithMaxTargets encapsulates minTargets to Option
func WithMaxTargets(maxTargets int) Option {
	return func(opts *Opts) error {
		opts.MaxTargets = maxTargets
		return nil
	}
}

// prepareQueryConfigOpts Reads channel config options from Option array
func prepareOpts(options ...Option) (Opts, error) {
	opts := Opts{}
	for _, option := range options {
		err := option(&opts)
		if err != nil {
			return opts, errors.WithMessage(err, "Failed to read query config opts")
		}
	}

	//resolve defaults
	if opts.MinResponses == 0 {
		opts.MinResponses = defaultMinResponses
	}
	if opts.MaxTargets == 0 {
		opts.MaxTargets = defaultMaxTargets
	}

	return opts, nil
}

func extractConfig(channelID string, configEnvelope *common.ConfigEnvelope) (*ChannelCfg, error) {

	group := configEnvelope.Config.ChannelGroup

	versions := &fab.Versions{
		Channel: &common.ConfigGroup{},
	}

	config := &ChannelCfg{
		id:          channelID,
		msps:        []*mb.MSPConfig{},
		anchorPeers: []*fab.OrgAnchorPeer{},
		orderers:    []string{},
		versions:    versions,
	}

	err := loadConfig(config, config.versions.Channel, group, "base", "", true)
	if err != nil {
		return nil, errors.WithMessage(err, "load config items from config group failed")
	}

	logger.Debugf("channel config: %v", config)

	return config, err

}

func loadConfig(configItems *ChannelCfg, versionsGroup *common.ConfigGroup, group *common.ConfigGroup, name string, org string, top bool) error {
	logger.Debugf("loadConfigGroup - %s - START groups Org: %s", name, org)
	if group == nil {
		return nil
	}

	logger.Debugf("loadConfigGroup - %s   - version %v", name, group.Version)
	logger.Debugf("loadConfigGroup - %s   - mod policy %s", name, group.ModPolicy)
	logger.Debugf("loadConfigGroup - %s - >> groups", name)

	groups := group.GetGroups()
	if groups != nil {
		versionsGroup.Groups = make(map[string]*common.ConfigGroup)
		for key, configGroup := range groups {
			logger.Debugf("loadConfigGroup - %s - found config group ==> %s", name, key)
			// The Application group is where config settings are that we want to find
			versionsGroup.Groups[key] = &common.ConfigGroup{}
			loadConfig(configItems, versionsGroup.Groups[key], configGroup, name+"."+key, key, false)
		}
	} else {
		logger.Debugf("loadConfigGroup - %s - no groups", name)
	}
	logger.Debugf("loadConfigGroup - %s - << groups", name)

	logger.Debugf("loadConfigGroup - %s - >> values", name)
	values := group.GetValues()
	if values != nil {
		versionsGroup.Values = make(map[string]*common.ConfigValue)
		for key, configValue := range values {
			versionsGroup.Values[key] = &common.ConfigValue{}
			loadConfigValue(configItems, key, versionsGroup.Values[key], configValue, name, org)
		}
	} else {
		logger.Debugf("loadConfigGroup - %s - no values", name)
	}
	logger.Debugf("loadConfigGroup - %s - << values", name)

	logger.Debugf("loadConfigGroup - %s - >> policies", name)
	policies := group.GetPolicies()
	if policies != nil {
		versionsGroup.Policies = make(map[string]*common.ConfigPolicy)
		for key, configPolicy := range policies {
			versionsGroup.Policies[key] = &common.ConfigPolicy{}
			loadConfigPolicy(configItems, key, versionsGroup.Policies[key], configPolicy, name, org)
		}
	} else {
		logger.Debugf("loadConfigGroup - %s - no policies", name)
	}
	logger.Debugf("loadConfigGroup - %s - << policies", name)
	logger.Debugf("loadConfigGroup - %s - < group", name)
	return nil
}

func loadConfigPolicy(configItems *ChannelCfg, key string, versionsPolicy *common.ConfigPolicy, configPolicy *common.ConfigPolicy, groupName string, org string) error {
	logger.Debugf("loadConfigPolicy - %s - name: %s", groupName, key)
	logger.Debugf("loadConfigPolicy - %s - version: %d", groupName, configPolicy.Version)
	logger.Debugf("loadConfigPolicy - %s - mod_policy: %s", groupName, configPolicy.ModPolicy)

	versionsPolicy.Version = configPolicy.Version
	return loadPolicy(configItems, versionsPolicy, key, configPolicy.Policy, groupName, org)
}

func loadPolicy(configItems *ChannelCfg, versionsPolicy *common.ConfigPolicy, key string, policy *common.Policy, groupName string, org string) error {

	policyType := common.Policy_PolicyType(policy.Type)

	switch policyType {
	case common.Policy_SIGNATURE:
		sigPolicyEnv := &common.SignaturePolicyEnvelope{}
		err := proto.Unmarshal(policy.Value, sigPolicyEnv)
		if err != nil {
			return errors.Wrap(err, "unmarshal signature policy envelope from config failed")
		}
		logger.Debugf("loadConfigPolicy - %s - policy SIGNATURE :: %v", groupName, sigPolicyEnv.Rule)
		// TODO: Do something with this value
		break

	case common.Policy_MSP:
		// TODO: Not implemented yet
		logger.Debugf("loadConfigPolicy - %s - policy :: MSP POLICY NOT PARSED ", groupName)
		break

	case common.Policy_IMPLICIT_META:
		implicitMetaPolicy := &common.ImplicitMetaPolicy{}
		err := proto.Unmarshal(policy.Value, implicitMetaPolicy)
		if err != nil {
			return errors.Wrap(err, "unmarshal implicit meta policy from config failed")
		}
		logger.Debugf("loadConfigPolicy - %s - policy IMPLICIT_META :: %v", groupName, implicitMetaPolicy)
		// TODO: Do something with this value
		break

	default:
		return errors.Errorf("unknown policy type %v", policyType)
	}
	return nil
}

func loadConfigValue(configItems *ChannelCfg, key string, versionsValue *common.ConfigValue, configValue *common.ConfigValue, groupName string, org string) error {
	logger.Debugf("loadConfigValue - %s - START value name: %s", groupName, key)
	logger.Debugf("loadConfigValue - %s   - version: %d", groupName, configValue.Version)
	logger.Debugf("loadConfigValue - %s   - modPolicy: %s", groupName, configValue.ModPolicy)

	versionsValue.Version = configValue.Version

	switch key {
	case channelConfig.AnchorPeersKey:
		anchorPeers := &pb.AnchorPeers{}
		err := proto.Unmarshal(configValue.Value, anchorPeers)
		if err != nil {
			return errors.Wrap(err, "unmarshal anchor peers from config failed")
		}

		logger.Debugf("loadConfigValue - %s   - AnchorPeers :: %s", groupName, anchorPeers)

		if len(anchorPeers.AnchorPeers) > 0 {
			for _, anchorPeer := range anchorPeers.AnchorPeers {
				oap := &fab.OrgAnchorPeer{Org: org, Host: anchorPeer.Host, Port: anchorPeer.Port}
				configItems.anchorPeers = append(configItems.anchorPeers, oap)
				logger.Debugf("loadConfigValue - %s   - AnchorPeer :: %s:%d:%s", groupName, oap.Host, oap.Port, oap.Org)
			}
		}
		break

	case channelConfig.MSPKey:
		mspConfig := &mb.MSPConfig{}
		err := proto.Unmarshal(configValue.Value, mspConfig)
		if err != nil {
			return errors.Wrap(err, "unmarshal MSPConfig from config failed")
		}

		logger.Debugf("loadConfigValue - %s   - MSP found", groupName)

		mspType := imsp.ProviderType(mspConfig.Type)
		if mspType != imsp.FABRIC {
			return errors.Errorf("unsupported MSP type (%v)", mspType)
		}

		configItems.msps = append(configItems.msps, mspConfig)
		break

	case channelConfig.ConsensusTypeKey:
		consensusType := &ab.ConsensusType{}
		err := proto.Unmarshal(configValue.Value, consensusType)
		if err != nil {
			return errors.Wrap(err, "unmarshal ConsensusType from config failed")
		}

		logger.Debugf("loadConfigValue - %s   - Consensus type value :: %s", groupName, consensusType.Type)
		// TODO: Do something with this value
		break

	case channelConfig.BatchSizeKey:
		batchSize := &ab.BatchSize{}
		err := proto.Unmarshal(configValue.Value, batchSize)
		if err != nil {
			return errors.Wrap(err, "unmarshal batch size from config failed")
		}

		logger.Debugf("loadConfigValue - %s   - BatchSize  maxMessageCount :: %d", groupName, batchSize.MaxMessageCount)
		logger.Debugf("loadConfigValue - %s   - BatchSize  absoluteMaxBytes :: %d", groupName, batchSize.AbsoluteMaxBytes)
		logger.Debugf("loadConfigValue - %s   - BatchSize  preferredMaxBytes :: %d", groupName, batchSize.PreferredMaxBytes)
		// TODO: Do something with this value
		break

	case channelConfig.BatchTimeoutKey:
		batchTimeout := &ab.BatchTimeout{}
		err := proto.Unmarshal(configValue.Value, batchTimeout)
		if err != nil {
			return errors.Wrap(err, "unmarshal batch timeout from config failed")
		}
		logger.Debugf("loadConfigValue - %s   - BatchTimeout timeout value :: %s", groupName, batchTimeout.Timeout)
		// TODO: Do something with this value
		break

	case channelConfig.ChannelRestrictionsKey:
		channelRestrictions := &ab.ChannelRestrictions{}
		err := proto.Unmarshal(configValue.Value, channelRestrictions)
		if err != nil {
			return errors.Wrap(err, "unmarshal channel restrictions from config failed")
		}
		logger.Debugf("loadConfigValue - %s   - ChannelRestrictions max_count value :: %d", groupName, channelRestrictions.MaxCount)
		// TODO: Do something with this value
		break

	case channelConfig.HashingAlgorithmKey:
		hashingAlgorithm := &common.HashingAlgorithm{}
		err := proto.Unmarshal(configValue.Value, hashingAlgorithm)
		if err != nil {
			return errors.Wrap(err, "unmarshal hashing algorithm from config failed")
		}
		logger.Debugf("loadConfigValue - %s   - HashingAlgorithm names value :: %s", groupName, hashingAlgorithm.Name)
		// TODO: Do something with this value
		break

	case channelConfig.ConsortiumKey:
		consortium := &common.Consortium{}
		err := proto.Unmarshal(configValue.Value, consortium)
		if err != nil {
			return errors.Wrap(err, "unmarshal consortium from config failed")
		}
		logger.Debugf("loadConfigValue - %s   - Consortium names value :: %s", groupName, consortium.Name)
		// TODO: Do something with this value
		break

	case channelConfig.BlockDataHashingStructureKey:
		bdhstruct := &common.BlockDataHashingStructure{}
		err := proto.Unmarshal(configValue.Value, bdhstruct)
		if err != nil {
			return errors.Wrap(err, "unmarshal block data hashing structure from config failed")
		}
		logger.Debugf("loadConfigValue - %s   - BlockDataHashingStructure width value :: %s", groupName, bdhstruct.Width)
		// TODO: Do something with this value
		break

	case channelConfig.OrdererAddressesKey:
		ordererAddresses := &common.OrdererAddresses{}
		err := proto.Unmarshal(configValue.Value, ordererAddresses)
		if err != nil {
			return errors.Wrap(err, "unmarshal orderer addresses from config failed")
		}
		logger.Debugf("loadConfigValue - %s   - OrdererAddresses addresses value :: %s", groupName, ordererAddresses.Addresses)
		if len(ordererAddresses.Addresses) > 0 {
			for _, ordererAddress := range ordererAddresses.Addresses {
				configItems.orderers = append(configItems.orderers, ordererAddress)
			}
		}
		break

	default:
		logger.Debugf("loadConfigValue - %s   - value: %s", groupName, configValue.Value)
	}
	return nil
}

// peersToTxnProcessors converts a slice of Peers to a slice of ProposalProcessors
func peersToTxnProcessors(peers []fab.Peer) []fab.ProposalProcessor {
	tpp := make([]fab.ProposalProcessor, len(peers))

	for i := range peers {
		tpp[i] = peers[i]
	}
	return tpp
}

//randomMaxTargets returns random sub set of max length targets
func randomMaxTargets(targets []fab.ProposalProcessor, max int) []fab.ProposalProcessor {
	if len(targets) <= max {
		return targets
	}
	for i := range targets {
		j := rand.Intn(i + 1)
		targets[i], targets[j] = targets[j], targets[i]
	}
	return targets[:max]
}
