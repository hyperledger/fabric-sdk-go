/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package chconfig

import (
	reqContext "context"
	"math/rand"
	"regexp"

	"github.com/golang/protobuf/proto"
	"github.com/hyperledger/fabric-protos-go/common"
	mb "github.com/hyperledger/fabric-protos-go/msp"
	pb "github.com/hyperledger/fabric-protos-go/peer"
	channelConfig "github.com/hyperledger/fabric-sdk-go/internal/github.com/hyperledger/fabric/common/channelconfig"
	imsp "github.com/hyperledger/fabric-sdk-go/internal/github.com/hyperledger/fabric/msp"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/errors/retry"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/logging"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/context"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/fab"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/msp"
	contextImpl "github.com/hyperledger/fabric-sdk-go/pkg/context"
	"github.com/hyperledger/fabric-sdk-go/pkg/fab/channel"
	"github.com/hyperledger/fabric-sdk-go/pkg/fab/resource"
	"github.com/pkg/errors"
)

var logger = logging.NewLogger("fabsdk/fab")

//overrideRetryHandler is private and used for unit-tests to test query retry behaviors
var overrideRetryHandler retry.Handler
var versionCapabilityPattern = regexp.MustCompile(`^V(\d+(_\d+?)*)$`)

// Opts contains options for retrieving channel configuration
type Opts struct {
	Orderer      fab.Orderer // if configured, channel config will be retrieved from this orderer
	Targets      []fab.Peer  // if configured, channel config will be retrieved from peers (targets)
	MinResponses int         // used with targets option; min number of success responses (from targets/peers)
	MaxTargets   int         //if configured, channel config will be retrieved for these number of random targets
	RetryOpts    retry.Opts  //opts for channel query retry handler
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
	id           string
	blockNumber  uint64
	msps         []*mb.MSPConfig
	anchorPeers  []*fab.OrgAnchorPeer
	orderers     []string
	versions     *fab.Versions
	capabilities map[fab.ConfigGroupKey]map[string]bool
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

// BlockNumber returns the channel config block number
func (cfg *ChannelCfg) BlockNumber() uint64 {
	return cfg.blockNumber
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

// HasCapability indicates whether or not the given group has the given capability
func (cfg *ChannelCfg) HasCapability(group fab.ConfigGroupKey, capability string) bool {
	groupCapabilities, ok := cfg.capabilities[group]
	if !ok {
		return false
	}
	if groupCapabilities[capability] {
		return true
	}

	// Special handling for version capabilities: V1_1 is supported if V1_2 or V1_3
	// are supported; V1_2 is supported if V1_3 is supported, etc.
	if isVersionCapability(capability) {
		for c := range groupCapabilities {
			if isVersionCapability(c) && c > capability {
				logger.Debugf("[%s] is greater than [%s] and therefore capability is supported", c, capability)
				return true
			}
		}
	}
	return false
}

// New channel config implementation
func New(channelID string, options ...Option) (*ChannelConfig, error) {
	opts, err := prepareOpts(options...)
	if err != nil {
		return nil, err
	}

	return &ChannelConfig{channelID: channelID, opts: opts}, nil
}

// QueryBlock returns channel configuration
func (c *ChannelConfig) QueryBlock(reqCtx reqContext.Context) (*common.Block, error) {

	if c.opts.Orderer != nil {
		return c.queryBlockFromOrderer(reqCtx)
	}

	return c.queryBlockFromPeers(reqCtx)
}

// Query returns channel configuration
func (c *ChannelConfig) Query(reqCtx reqContext.Context) (fab.ChannelCfg, error) {

	if c.opts.Orderer != nil {
		return c.queryOrderer(reqCtx)
	}

	return c.queryPeers(reqCtx)
}

func (c *ChannelConfig) queryPeers(reqCtx reqContext.Context) (*ChannelCfg, error) {
	block, err := c.queryBlockFromPeers(reqCtx)

	if err != nil {
		return nil, errors.WithMessage(err, "QueryBlockConfig failed")
	}
	return extractConfig(c.channelID, block)

}

func (c *ChannelConfig) queryBlockFromPeers(reqCtx reqContext.Context) (*common.Block, error) {
	ctx, ok := contextImpl.RequestClientContext(reqCtx)
	if !ok {
		return nil, errors.New("failed get client context from reqContext for signPayload")
	}

	l, err := channel.NewLedger(c.channelID)
	if err != nil {
		return nil, errors.WithMessage(err, "ledger client creation failed")
	}

	c.resolveOptsFromConfig(ctx)

	targets := []fab.ProposalProcessor{}
	if c.opts.Targets == nil {
		// Calculate targets from config
		targets, err = c.calculateTargetsFromConfig(ctx)
		if err != nil {
			return nil, err
		}
	} else {
		targets = peersToTxnProcessors(c.opts.Targets)
	}

	retryHandler := retry.New(c.opts.RetryOpts)

	//Unit test purpose only
	if overrideRetryHandler != nil {
		retryHandler = overrideRetryHandler
	}

	block, err := retry.NewInvoker(retryHandler).Invoke(
		func() (interface{}, error) {
			return l.QueryConfigBlock(reqCtx, targets, &channel.TransactionProposalResponseVerifier{MinResponses: c.opts.MinResponses})
		},
	)

	if err != nil {
		return nil, errors.WithMessage(err, "QueryBlockConfig failed")
	}
	return block.(*common.Block), nil

}

func (c *ChannelConfig) calculateTargetsFromConfig(ctx context.Client) ([]fab.ProposalProcessor, error) {
	targets := []fab.ProposalProcessor{}
	chPeers := ctx.EndpointConfig().ChannelPeers(c.channelID)
	if len(chPeers) == 0 {
		return nil, errors.Errorf("no channel peers configured for channel [%s]", c.channelID)
	}

	for _, p := range chPeers {
		newPeer, err := ctx.InfraProvider().CreatePeerFromConfig((&p.NetworkPeer))
		if err != nil || newPeer == nil {
			return nil, errors.WithMessage(err, "NewPeer failed")
		}

		// Pick peers in the same MSP as the context since only they can query system chaincode
		if newPeer.MSPID() == ctx.Identifier().MSPID {
			targets = append(targets, newPeer)
		}
	}

	targets = randomMaxTargets(targets, c.opts.MaxTargets)
	return targets, nil
}

func (c *ChannelConfig) queryOrderer(reqCtx reqContext.Context) (*ChannelCfg, error) {

	block, err := c.queryBlockFromOrderer(reqCtx)
	if err != nil {
		return nil, errors.WithMessage(err, "LastConfigFromOrderer failed")
	}

	return extractConfig(c.channelID, block)
}

func (c *ChannelConfig) queryBlockFromOrderer(reqCtx reqContext.Context) (*common.Block, error) {

	return resource.LastConfigFromOrderer(reqCtx, c.channelID, c.opts.Orderer, resource.WithRetry(c.opts.RetryOpts))
}

//resolveOptsFromConfig loads opts from config if not loaded/initialized
func (c *ChannelConfig) resolveOptsFromConfig(ctx context.Client) {

	if c.opts.MaxTargets != 0 && c.opts.MinResponses != 0 && c.opts.RetryOpts.RetryableCodes != nil {
		//already loaded
		return
	}

	chSdkCfg := ctx.EndpointConfig().ChannelConfig(c.channelID)

	//resolve opts
	c.resolveMaxResponsesOptsFromConfig(chSdkCfg)
	c.resolveMinResponsesOptsFromConfig(chSdkCfg)
	c.resolveRetryOptsFromConfig(chSdkCfg)

	//apply default to missing opts
	c.applyDefaultOpts()

}

func (c *ChannelConfig) resolveMaxResponsesOptsFromConfig(chSdkCfg *fab.ChannelEndpointConfig) {
	if c.opts.MaxTargets == 0 {
		c.opts.MaxTargets = chSdkCfg.Policies.QueryChannelConfig.MaxTargets
	}
}

func (c *ChannelConfig) resolveMinResponsesOptsFromConfig(chSdkCfg *fab.ChannelEndpointConfig) {
	if c.opts.MinResponses == 0 {
		c.opts.MinResponses = chSdkCfg.Policies.QueryChannelConfig.MinResponses
	}
}

func (c *ChannelConfig) resolveRetryOptsFromConfig(chSdkCfg *fab.ChannelEndpointConfig) {
	if c.opts.RetryOpts.RetryableCodes == nil {
		c.opts.RetryOpts = chSdkCfg.Policies.QueryChannelConfig.RetryOpts
		c.opts.RetryOpts.RetryableCodes = retry.ChannelConfigRetryableCodes
	}
}

func (c *ChannelConfig) applyDefaultOpts() {
	if c.opts.RetryOpts.Attempts == 0 {
		c.opts.RetryOpts.Attempts = retry.DefaultAttempts
	}

	if c.opts.RetryOpts.InitialBackoff == 0 {
		c.opts.RetryOpts.InitialBackoff = retry.DefaultInitialBackoff
	}

	if c.opts.RetryOpts.BackoffFactor == 0 {
		c.opts.RetryOpts.BackoffFactor = retry.DefaultBackoffFactor
	}

	if c.opts.RetryOpts.MaxBackoff == 0 {
		c.opts.RetryOpts.MaxBackoff = retry.DefaultMaxBackoff
	}

	if c.opts.RetryOpts.RetryableCodes == nil {
		c.opts.RetryOpts.RetryableCodes = retry.ChannelConfigRetryableCodes
	}
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

// WithRetryOpts encapsulates retry opts to Option
func WithRetryOpts(retryOpts retry.Opts) Option {
	return func(opts *Opts) error {
		opts.RetryOpts = retryOpts
		return nil
	}
}

// prepareOpts Reads channel config options from Option array
func prepareOpts(options ...Option) (Opts, error) {
	opts := Opts{}
	for _, option := range options {
		err := option(&opts)
		if err != nil {
			return opts, errors.WithMessage(err, "Failed to read query config opts")
		}
	}

	return opts, nil
}

func extractConfig(channelID string, block *common.Block) (*ChannelCfg, error) {
	if block.Header == nil {
		return nil, errors.New("expected header in block")
	}

	configEnvelope, err := resource.CreateConfigEnvelope(block.Data.Data[0])
	if err != nil {
		return nil, err
	}

	group := configEnvelope.Config.ChannelGroup

	versions := &fab.Versions{
		Channel: &common.ConfigGroup{},
	}

	config := &ChannelCfg{
		id:           channelID,
		blockNumber:  block.Header.Number,
		msps:         []*mb.MSPConfig{},
		anchorPeers:  []*fab.OrgAnchorPeer{},
		orderers:     []string{},
		versions:     versions,
		capabilities: make(map[fab.ConfigGroupKey]map[string]bool),
	}

	err = loadConfig(config, config.versions.Channel, group, "", "")
	if err != nil {
		return nil, errors.WithMessage(err, "load config items from config group failed")
	}

	logger.Debugf("loaded channel config: %+v", config)

	return config, err

}

func loadConfig(configItems *ChannelCfg, versionsGroup *common.ConfigGroup, group *common.ConfigGroup, name string, org string) error {
	if group == nil {
		return nil
	}

	versionsGroup.Version = group.Version
	versionsGroup.ModPolicy = group.ModPolicy

	groups := group.GetGroups()
	if groups != nil {
		versionsGroup.Groups = make(map[string]*common.ConfigGroup)
		for key, configGroup := range groups {
			logger.Debugf("loadConfigGroup - %s - found config group ==> %s", name, key)
			// The Application group is where config settings are that we want to find
			versionsGroup.Groups[key] = &common.ConfigGroup{}
			err := loadConfig(configItems, versionsGroup.Groups[key], configGroup, key, key)
			if err != nil {
				return err
			}
		}
	}

	values := group.GetValues()
	if values != nil {
		versionsGroup.Values = make(map[string]*common.ConfigValue)
		for key, configValue := range values {
			versionsGroup.Values[key] = &common.ConfigValue{}
			err := loadConfigValue(configItems, key, versionsGroup.Values[key], configValue, name, org)
			if err != nil {
				return err
			}

		}
	}

	loadConfigGroupPolicies(versionsGroup, group)

	return nil
}

func loadConfigGroupPolicies(versionsGroup *common.ConfigGroup, group *common.ConfigGroup) {
	policies := group.GetPolicies()
	if policies != nil {
		versionsGroup.Policies = make(map[string]*common.ConfigPolicy)
		for key, configPolicy := range policies {
			versionsGroup.Policies[key] = &common.ConfigPolicy{}

			versionsGroup.Policies[key].Version = configPolicy.Version
			versionsGroup.Policies[key].Policy = configPolicy.Policy
			versionsGroup.Policies[key].ModPolicy = configPolicy.ModPolicy
		}
	}
}

func loadAnchorPeers(configValue *common.ConfigValue, configItems *ChannelCfg, org string) error {
	anchorPeers := &pb.AnchorPeers{}
	err := proto.Unmarshal(configValue.Value, anchorPeers)
	if err != nil {
		return errors.Wrap(err, "unmarshal anchor peers from config failed")
	}

	if len(anchorPeers.AnchorPeers) > 0 {
		for _, anchorPeer := range anchorPeers.AnchorPeers {
			oap := &fab.OrgAnchorPeer{Org: org, Host: anchorPeer.Host, Port: anchorPeer.Port}
			configItems.anchorPeers = append(configItems.anchorPeers, oap)
		}
	}
	return nil
}

func loadMSPKey(configValue *common.ConfigValue, configItems *ChannelCfg) error {
	mspConfig := &mb.MSPConfig{}
	err := proto.Unmarshal(configValue.Value, mspConfig)
	if err != nil {
		return errors.Wrap(err, "unmarshal MSPConfig from config failed")
	}

	mspType := imsp.ProviderType(mspConfig.Type)
	if mspType != imsp.FABRIC {
		return errors.Errorf("unsupported MSP type (%v)", mspType)
	}

	configItems.msps = append(configItems.msps, mspConfig)
	return nil

}

func loadOrdererAddressesKey(configValue *common.ConfigValue, configItems *ChannelCfg) error {
	ordererAddresses := &common.OrdererAddresses{}
	err := proto.Unmarshal(configValue.Value, ordererAddresses)
	if err != nil {
		return errors.Wrap(err, "unmarshal orderer addresses from config failed")
	}
	if len(ordererAddresses.Addresses) > 0 {
		configItems.orderers = append(configItems.orderers, ordererAddresses.Addresses...)
	}
	return nil

}

func loadCapabilities(configValue *common.ConfigValue, configItems *ChannelCfg, groupName string) error {
	capabilities := &common.Capabilities{}
	err := proto.Unmarshal(configValue.Value, capabilities)
	if err != nil {
		return errors.Wrap(err, "unmarshal capabilities from config failed")
	}
	capabilityMap := make(map[string]bool)
	for capability := range capabilities.Capabilities {
		capabilityMap[capability] = true
	}
	configItems.capabilities[fab.ConfigGroupKey(groupName)] = capabilityMap
	return nil
}

func loadConfigValue(configItems *ChannelCfg, key string, versionsValue *common.ConfigValue, configValue *common.ConfigValue, groupName string, org string) error {
	versionsValue.Version = configValue.Version
	versionsValue.Value = configValue.Value

	switch key {
	case channelConfig.AnchorPeersKey:
		if err := loadAnchorPeers(configValue, configItems, org); err != nil {
			return err
		}
	case channelConfig.MSPKey:
		if err := loadMSPKey(configValue, configItems); err != nil {
			return err
		}
	case channelConfig.CapabilitiesKey:
		if err := loadCapabilities(configValue, configItems, groupName); err != nil {
			return err
		}
	//case channelConfig.ConsensusTypeKey:
	//	consensusType := &ab.ConsensusType{}
	//	err := proto.Unmarshal(configValue.Value, consensusType)
	//	if err != nil {
	//		return errors.Wrap(err, "unmarshal ConsensusType from config failed")
	//	}
	//
	//	logger.Debugf("loadConfigValue - %s   - Consensus type value :: %s", groupName, consensusType.Type)
	//	// TODO: Do something with this value
	//case channelConfig.BatchSizeKey:
	//	batchSize := &ab.BatchSize{}
	//	err := proto.Unmarshal(configValue.Value, batchSize)
	//	if err != nil {
	//		return errors.Wrap(err, "unmarshal batch size from config failed")
	//	}
	//
	//	// TODO: Do something with this value

	//case channelConfig.BatchTimeoutKey:
	//	batchTimeout := &ab.BatchTimeout{}
	//	err := proto.Unmarshal(configValue.Value, batchTimeout)
	//	if err != nil {
	//		return errors.Wrap(err, "unmarshal batch timeout from config failed")
	//	}
	//	// TODO: Do something with this value

	//case channelConfig.ChannelRestrictionsKey:
	//	channelRestrictions := &ab.ChannelRestrictions{}
	//	err := proto.Unmarshal(configValue.Value, channelRestrictions)
	//	if err != nil {
	//		return errors.Wrap(err, "unmarshal channel restrictions from config failed")
	//	}
	//	// TODO: Do something with this value

	//case channelConfig.HashingAlgorithmKey:
	//	hashingAlgorithm := &common.HashingAlgorithm{}
	//	err := proto.Unmarshal(configValue.Value, hashingAlgorithm)
	//	if err != nil {
	//		return errors.Wrap(err, "unmarshal hashing algorithm from config failed")
	//	}
	//	// TODO: Do something with this value

	//case channelConfig.ConsortiumKey:
	//	consortium := &common.Consortium{}
	//	err := proto.Unmarshal(configValue.Value, consortium)
	//	if err != nil {
	//		return errors.Wrap(err, "unmarshal consortium from config failed")
	//	}
	//	// TODO: Do something with this value

	//case channelConfig.BlockDataHashingStructureKey:
	//	bdhstruct := &common.BlockDataHashingStructure{}
	//	err := proto.Unmarshal(configValue.Value, bdhstruct)
	//	if err != nil {
	//		return errors.Wrap(err, "unmarshal block data hashing structure from config failed")
	//	}
	//	// TODO: Do something with this value

	case channelConfig.OrdererAddressesKey:
		if err := loadOrdererAddressesKey(configValue, configItems); err != nil {
			return err
		}

	default:
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

func isVersionCapability(capability string) bool {
	return versionCapabilityPattern.MatchString(capability)
}
