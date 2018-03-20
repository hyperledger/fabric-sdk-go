/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package config

import (
	"bytes"
	"crypto/tls"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"io"
	"io/ioutil"
	"math/rand"
	"os"
	"path"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/spf13/viper"

	"github.com/hyperledger/fabric-sdk-go/pkg/common/errors/status"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/logging"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/core"
	"github.com/hyperledger/fabric-sdk-go/pkg/core/config/cryptoutil"
	"github.com/hyperledger/fabric-sdk-go/pkg/core/config/endpoint"
	"github.com/pkg/errors"

	"regexp"

	"sync"

	cs "github.com/hyperledger/fabric-sdk-go/pkg/core/cryptosuite"
)

var logger = logging.NewLogger("fabsdk/core")

const (
	cmdRoot                        = "FABRIC_SDK"
	defaultTimeout                 = time.Second * 5
	defaultConnIdleTimeout         = time.Second * 30
	defaultCacheSweepInterval      = time.Second * 15
	defaultEventServiceIdleTimeout = time.Minute * 2
	defaultResMgmtTimeout          = time.Second * 180
	defaultExecuteTimeout          = time.Second * 180
)

var logModules = [...]string{"fabsdk", "fabsdk/client", "fabsdk/core", "fabsdk/fab", "fabsdk/common",
	"fabsdk/msp", "fabsdk/util", "fabsdk/context"}

// Config represents the configuration for the client
type Config struct {
	tlsCerts            []*x509.Certificate
	networkConfig       *core.NetworkConfig
	networkConfigCached bool
	configViper         *viper.Viper
	peerMatchers        map[int]*regexp.Regexp
	ordererMatchers     map[int]*regexp.Regexp
	caMatchers          map[int]*regexp.Regexp
	opts                options
	certPoolLock        sync.Mutex
}

type options struct {
	envPrefix    string
	templatePath string
	template     *Config
}

// Option configures the package.
type Option func(opts *options) error

// FromReader loads configuration from in.
// configType can be "json" or "yaml".
func FromReader(in io.Reader, configType string, opts ...Option) core.ConfigProvider {
	return func() (core.Config, error) {
		c, err := newConfig(opts...)
		if err != nil {
			return nil, err
		}

		if configType == "" {
			return nil, errors.New("empty config type")
		}

		// read config from bytes array, but must set ConfigType
		// for viper to properly unmarshal the bytes array
		c.configViper.SetConfigType(configType)
		c.configViper.MergeConfig(in)

		return initConfig(c)
	}
}

// FromFile reads from named config file
func FromFile(name string, opts ...Option) core.ConfigProvider {
	return func() (core.Config, error) {
		c, err := newConfig(opts...)
		if err != nil {
			return nil, err
		}

		if name == "" {
			return nil, errors.New("filename is required")
		}

		// create new viper
		c.configViper.SetConfigFile(name)

		// If a config file is found, read it in.
		err = c.configViper.MergeInConfig()
		if err == nil {
			logger.Debugf("Using config file: %s", c.configViper.ConfigFileUsed())
		} else {
			return nil, errors.Wrap(err, "loading config file failed")
		}

		return initConfig(c)
	}
}

// FromRaw will initialize the configs from a byte array
func FromRaw(configBytes []byte, configType string, opts ...Option) core.ConfigProvider {
	buf := bytes.NewBuffer(configBytes)
	logger.Debugf("config.FromRaw buf Len is %d, Cap is %d: %s", buf.Len(), buf.Cap(), buf)

	return FromReader(buf, configType, opts...)
}

/*
// FromDefaultPath loads configuration from the default path
func FromDefaultPath(opts ...Option) (*Config, error) {
	optsWithDef := append(opts, withTemplatePathFromEnv("CONFIG_PATH"))

	c, err := newConfig(optsWithDef...)
	if err != nil {
		return nil, err
	}
	if c.opts.templatePath == "" {
		return nil, errors.New("Configuration path is not set")
	}

	return initConfig(c)
}
*/

// WithEnvPrefix defines the prefix for environment variable overrides.
// See viper SetEnvPrefix for more information.
func WithEnvPrefix(prefix string) Option {
	return func(opts *options) error {
		opts.envPrefix = prefix
		return nil
	}
}

/*
// WithTemplatePath loads the named file to populate a configuration template prior to loading the instance configuration.
func WithTemplatePath(path string) Option {
	return func(opts *options) error {
		opts.templatePath = path
		return nil
	}
}
*/

/*
func withTemplatePathFromEnv(e string) Option {
	return func(opts *options) error {
		if opts.templatePath == "" {
			opts.templatePath = os.Getenv(opts.envPrefix + "_" + e)
		}

		return nil
	}
}
*/

func newConfig(opts ...Option) (*Config, error) {
	o := options{
		envPrefix: cmdRoot,
	}

	for _, option := range opts {
		err := option(&o)
		if err != nil {
			return nil, errors.WithMessage(err, "Error in option passed to New")
		}
	}

	v := newViper(o.envPrefix)
	c := Config{
		configViper: v,
		opts:        o,
	}

	err := c.loadTemplateConfig()
	if err != nil {
		return nil, err
	}

	return &c, nil
}

func newViper(cmdRootPrefix string) *viper.Viper {
	myViper := viper.New()
	myViper.SetEnvPrefix(cmdRootPrefix)
	myViper.AutomaticEnv()
	replacer := strings.NewReplacer(".", "_")
	myViper.SetEnvKeyReplacer(replacer)
	return myViper
}

func initConfig(c *Config) (*Config, error) {
	setLogLevel(c.configViper)

	if err := c.cacheNetworkConfiguration(); err != nil {
		return nil, errors.WithMessage(err, "network configuration load failed")
	}

	for _, logModule := range logModules {
		logger.Debugf("config %s logging level is set to: %s", logModule, logging.ParseString(logging.GetLevel(logModule)))
	}

	//Compile the entityMatchers
	c.peerMatchers = make(map[int]*regexp.Regexp)
	c.ordererMatchers = make(map[int]*regexp.Regexp)
	c.caMatchers = make(map[int]*regexp.Regexp)

	matchError := c.compileMatchers()
	if matchError != nil {
		return nil, matchError
	}

	return c, nil
}

// setLogLevel will set the log level of the client
func setLogLevel(myViper *viper.Viper) {
	loggingLevelString := myViper.GetString("client.logging.level")
	logLevel := logging.INFO
	if loggingLevelString != "" {
		const logModule = "fabsdk" // TODO: allow more flexability in setting levels for different modules
		logger.Debugf("%s logging level from the config: %v", logModule, loggingLevelString)
		var err error
		logLevel, err = logging.LogLevel(loggingLevelString)
		if err != nil {
			panic(err)
		}
	}

	// TODO: allow separate settings for each
	for _, logModule := range logModules {
		logging.SetLevel(logModule, logLevel)
	}
}

// load Default config
func (c *Config) loadTemplateConfig() error {
	// get Environment Default Config Path
	templatePath := c.opts.templatePath
	if templatePath == "" {
		return nil
	}

	// if set, use it to load default config
	c.configViper.AddConfigPath(SubstPathVars(templatePath))
	err := c.configViper.ReadInConfig() // Find and read the config file
	if err != nil {                     // Handle errors reading the config file
		return errors.Wrap(err, "loading config file failed")
	}
	return nil
}

// Client returns the Client config
func (c *Config) Client() (*core.ClientConfig, error) {
	config, err := c.NetworkConfig()
	if err != nil {
		return nil, err
	}
	client := config.Client

	client.TLSCerts.Path = SubstPathVars(client.TLSCerts.Path)
	client.TLSCerts.Client.Key.Path = SubstPathVars(client.TLSCerts.Client.Key.Path)
	client.TLSCerts.Client.Cert.Path = SubstPathVars(client.TLSCerts.Client.Cert.Path)

	return &client, nil
}

// CAConfig returns the CA configuration.
func (c *Config) CAConfig(org string) (*core.CAConfig, error) {
	config, err := c.NetworkConfig()
	if err != nil {
		return nil, err
	}

	caName, err := c.getCAName(org)
	if err != nil {
		return nil, err
	}
	caConfig := config.CertificateAuthorities[strings.ToLower(caName)]

	return &caConfig, nil
}

// CAServerCertPems Read configuration option for the server certificates
// will send a list of cert pem contents directly from the config bytes array
func (c *Config) CAServerCertPems(org string) ([]string, error) {
	config, err := c.NetworkConfig()
	if err != nil {
		return nil, err
	}
	caName, err := c.getCAName(org)
	if err != nil {
		return nil, err
	}
	if _, ok := config.CertificateAuthorities[strings.ToLower(caName)]; !ok {
		return nil, errors.Errorf("CA Server Name '%s' not found", caName)
	}
	certFilesPem := config.CertificateAuthorities[caName].TLSCACerts.Pem
	certPems := make([]string, len(certFilesPem))
	for i, v := range certFilesPem {
		certPems[i] = string(v)
	}

	return certPems, nil
}

// CAServerCertPaths Read configuration option for the server certificates
// will send a list of cert file paths
func (c *Config) CAServerCertPaths(org string) ([]string, error) {
	config, err := c.NetworkConfig()
	if err != nil {
		return nil, err
	}
	caName, err := c.getCAName(org)
	if err != nil {
		return nil, err
	}
	if _, ok := config.CertificateAuthorities[strings.ToLower(caName)]; !ok {
		return nil, errors.Errorf("CA Server Name '%s' not found", caName)
	}

	certFiles := strings.Split(config.CertificateAuthorities[caName].TLSCACerts.Path, ",")

	certFileModPath := make([]string, len(certFiles))
	for i, v := range certFiles {
		certFileModPath[i] = SubstPathVars(v)
	}

	return certFileModPath, nil
}

func (c *Config) getCAName(org string) (string, error) {
	config, err := c.NetworkConfig()
	if err != nil {
		return "", err
	}

	logger.Debug("Getting cert authority for org: %s.", org)

	if len(config.Organizations[strings.ToLower(org)].CertificateAuthorities) == 0 {
		return "", errors.Errorf("organization %s has no Certificate Authorities setup. Make sure each org has at least 1 configured", org)
	}
	//for now, we're only loading the first Cert Authority by default. TODO add logic to support passing the Cert Authority ID needed by the client.
	certAuthorityName := config.Organizations[strings.ToLower(org)].CertificateAuthorities[0]
	logger.Debugf("Cert authority for org: %s is %s", org, certAuthorityName)

	if certAuthorityName == "" {
		return "", errors.Errorf("certificate authority empty for %s. Make sure each org has at least 1 non empty certificate authority name", org)
	}

	if _, ok := config.CertificateAuthorities[strings.ToLower(certAuthorityName)]; !ok {
		logger.Debugf("Could not find Certificate Authority for [%s], trying with Entity Matchers", certAuthorityName)
		_, mappedHost, err := c.tryMatchingCAConfig(strings.ToLower(certAuthorityName))
		if err != nil {
			return "", errors.WithMessage(err, fmt.Sprintf("CA Server Name %s not found", certAuthorityName))
		}
		logger.Debugf("Mapped Certificate Authority for [%s] to [%s]", certAuthorityName, mappedHost)
		return mappedHost, nil
	}

	return certAuthorityName, nil
}

// CAClientKeyPath Read configuration option for the fabric CA client key file
func (c *Config) CAClientKeyPath(org string) (string, error) {
	config, err := c.NetworkConfig()
	if err != nil {
		return "", err
	}

	caName, err := c.getCAName(org)
	if err != nil {
		return "", err
	}
	if _, ok := config.CertificateAuthorities[strings.ToLower(caName)]; !ok {
		return "", errors.Errorf("CA Server Name '%s' not found", caName)
	}
	return SubstPathVars(config.CertificateAuthorities[strings.ToLower(caName)].TLSCACerts.Client.Key.Path), nil
}

// CAClientKeyPem Read configuration option for the fabric CA client key pem embedded in the client config
func (c *Config) CAClientKeyPem(org string) (string, error) {
	config, err := c.NetworkConfig()
	if err != nil {
		return "", err
	}

	caName, err := c.getCAName(org)
	if err != nil {
		return "", err
	}
	if _, ok := config.CertificateAuthorities[strings.ToLower(caName)]; !ok {
		return "", errors.Errorf("CA Server Name '%s' not found", caName)
	}

	ca := config.CertificateAuthorities[strings.ToLower(caName)]
	if len(ca.TLSCACerts.Client.Key.Pem) == 0 {
		return "", errors.New("Empty Client Key Pem")
	}

	return ca.TLSCACerts.Client.Key.Pem, nil
}

// CAClientCertPath Read configuration option for the fabric CA client cert file
func (c *Config) CAClientCertPath(org string) (string, error) {
	config, err := c.NetworkConfig()
	if err != nil {
		return "", err
	}

	caName, err := c.getCAName(org)
	if err != nil {
		return "", err
	}
	if _, ok := config.CertificateAuthorities[strings.ToLower(caName)]; !ok {
		return "", errors.Errorf("CA Server Name %s not found", caName)
	}
	return SubstPathVars(config.CertificateAuthorities[strings.ToLower(caName)].TLSCACerts.Client.Cert.Path), nil
}

func (c *Config) tryMatchingCAConfig(caName string) (*core.CAConfig, string, error) {
	networkConfig, err := c.NetworkConfig()
	if err != nil {
		return nil, "", err
	}
	//Return if no caMatchers are configured
	if len(c.caMatchers) == 0 {
		return nil, "", errors.New("no CertAuthority entityMatchers are found")
	}

	//sort the keys
	var keys []int
	for k := range c.caMatchers {
		keys = append(keys, k)
	}
	sort.Ints(keys)

	//loop over certAuthorityEntityMatchers to find the matching Cert
	for _, k := range keys {
		v := c.caMatchers[k]
		if v.MatchString(caName) {
			// get the matching Config from the index number
			certAuthorityMatchConfig := networkConfig.EntityMatchers["certificateauthorities"][k]
			//Get the certAuthorityMatchConfig from mapped host
			caConfig, ok := networkConfig.CertificateAuthorities[strings.ToLower(certAuthorityMatchConfig.MappedHost)]
			if !ok {
				return nil, certAuthorityMatchConfig.MappedHost, errors.New("failed to load config from matched CertAuthority")
			}
			_, isPortPresentInCAName := c.getPortIfPresent(caName)
			//if substitution url is empty, use the same network certAuthority url
			if certAuthorityMatchConfig.URLSubstitutionExp == "" {
				port, isPortPresent := c.getPortIfPresent(caConfig.URL)

				caConfig.URL = caName
				//append port of matched config
				if isPortPresent && !isPortPresentInCAName {
					caConfig.URL += ":" + strconv.Itoa(port)
				}
			} else {
				//else, replace url with urlSubstitutionExp if it doesnt have any variable declarations like $
				if strings.Index(certAuthorityMatchConfig.URLSubstitutionExp, "$") < 0 {
					caConfig.URL = certAuthorityMatchConfig.URLSubstitutionExp
				} else {
					//if the urlSubstitutionExp has $ variable declarations, use regex replaceallstring to replace networkhostname with substituionexp pattern
					caConfig.URL = v.ReplaceAllString(caName, certAuthorityMatchConfig.URLSubstitutionExp)
				}
			}

			return &caConfig, certAuthorityMatchConfig.MappedHost, nil
		}
	}

	return nil, "", errors.WithStack(status.New(status.ClientStatus, status.NoMatchingCertificateAuthorityEntity.ToInt32(), "no matching certAuthority config found", nil))
}

// CAClientCertPem Read configuration option for the fabric CA client cert pem embedded in the client config
func (c *Config) CAClientCertPem(org string) (string, error) {
	config, err := c.NetworkConfig()
	if err != nil {
		return "", err
	}

	caName, err := c.getCAName(org)
	if err != nil {
		return "", err
	}

	if _, ok := config.CertificateAuthorities[strings.ToLower(caName)]; !ok {
		return "", errors.Errorf("CA Server Name '%s' not found", caName)
	}

	ca := config.CertificateAuthorities[strings.ToLower(caName)]
	if len(ca.TLSCACerts.Client.Cert.Pem) == 0 {
		return "", errors.New("Empty Client Cert Pem")
	}

	return ca.TLSCACerts.Client.Cert.Pem, nil
}

// TimeoutOrDefault reads timeouts for the given timeout type, if not found, defaultTimeout is returned
func (c *Config) TimeoutOrDefault(tType core.TimeoutType) time.Duration {
	timeout := c.getTimeout(tType)
	if timeout == 0 {
		timeout = defaultTimeout
	}

	return timeout
}

// Timeout reads timeouts for the given timeout type, the default is 0 if type is not found in config
func (c *Config) Timeout(tType core.TimeoutType) time.Duration {
	return c.getTimeout(tType)
}

// EventServiceType returns the type of event service client to use
func (c *Config) EventServiceType() core.EventServiceType {
	etype := c.configViper.GetString("client.eventService.type")
	switch etype {
	case "deliver":
		return core.DeliverEventServiceType
	default:
		return core.EventHubEventServiceType
	}
}

func (c *Config) getTimeout(tType core.TimeoutType) time.Duration {
	var timeout time.Duration
	switch tType {
	case core.EndorserConnection:
		timeout = c.configViper.GetDuration("client.peer.timeout.connection")
	case core.Query:
		timeout = c.configViper.GetDuration("client.global.timeout.query")
	case core.Execute:
		timeout = c.configViper.GetDuration("client.global.timeout.execute")
		if timeout == 0 {
			timeout = defaultExecuteTimeout
		}
	case core.DiscoveryGreylistExpiry:
		timeout = c.configViper.GetDuration("client.peer.timeout.discovery.greylistExpiry")
	case core.PeerResponse:
		timeout = c.configViper.GetDuration("client.peer.timeout.response")
	case core.EventHubConnection:
		timeout = c.configViper.GetDuration("client.eventService.timeout.connection")
	case core.EventReg:
		timeout = c.configViper.GetDuration("client.eventService.timeout.registrationResponse")
	case core.OrdererConnection:
		timeout = c.configViper.GetDuration("client.orderer.timeout.connection")
	case core.OrdererResponse:
		timeout = c.configViper.GetDuration("client.orderer.timeout.response")
	case core.ChannelConfigRefresh:
		timeout = c.configViper.GetDuration("client.global.cache.channelConfig")
	case core.ChannelMembershipRefresh:
		timeout = c.configViper.GetDuration("client.global.cache.channelMembership")
	case core.CacheSweepInterval: // EXPERIMENTAL - do we need this to be configurable?
		timeout = c.configViper.GetDuration("client.cache.interval.sweep")
		if timeout == 0 {
			timeout = defaultCacheSweepInterval
		}
	case core.ConnectionIdle:
		timeout = c.configViper.GetDuration("client.global.cache.connectionIdle")
		if timeout == 0 {
			timeout = defaultConnIdleTimeout
		}
	case core.EventServiceIdle:
		timeout = c.configViper.GetDuration("client.global.cache.eventServiceIdle")
		if timeout == 0 {
			timeout = defaultEventServiceIdleTimeout
		}
	case core.ResMgmt:
		timeout = c.configViper.GetDuration("client.global.timeout.resmgmt")
		if timeout == 0 {
			timeout = defaultResMgmtTimeout
		}
	}

	return timeout
}

// MSPID returns the MSP ID for the requested organization
func (c *Config) MSPID(org string) (string, error) {
	config, err := c.NetworkConfig()
	if err != nil {
		return "", err
	}
	// viper lowercases all key maps, org is lower case
	mspID := config.Organizations[strings.ToLower(org)].MSPID
	if mspID == "" {
		return "", errors.Errorf("MSP ID is empty for org: %s", org)
	}

	return mspID, nil
}

func (c *Config) cacheNetworkConfiguration() error {
	networkConfig := core.NetworkConfig{}
	networkConfig.Name = c.configViper.GetString("name")
	networkConfig.Xtype = c.configViper.GetString("x-type")
	networkConfig.Description = c.configViper.GetString("description")
	networkConfig.Version = c.configViper.GetString("version")

	err := c.configViper.UnmarshalKey("client", &networkConfig.Client)
	logger.Debugf("Client is: %+v", networkConfig.Client)
	if err != nil {
		return err
	}
	err = c.configViper.UnmarshalKey("channels", &networkConfig.Channels)
	logger.Debugf("channels are: %+v", networkConfig.Channels)
	if err != nil {
		return err
	}
	err = c.configViper.UnmarshalKey("organizations", &networkConfig.Organizations)
	logger.Debugf("organizations are: %+v", networkConfig.Organizations)
	if err != nil {
		return err
	}
	err = c.configViper.UnmarshalKey("orderers", &networkConfig.Orderers)
	logger.Debugf("orderers are: %+v", networkConfig.Orderers)
	if err != nil {
		return err
	}
	err = c.configViper.UnmarshalKey("peers", &networkConfig.Peers)
	logger.Debugf("peers are: %+v", networkConfig.Peers)
	if err != nil {
		return err
	}
	err = c.configViper.UnmarshalKey("certificateAuthorities", &networkConfig.CertificateAuthorities)
	logger.Debugf("certificateAuthorities are: %+v", networkConfig.CertificateAuthorities)
	if err != nil {
		return err
	}

	err = c.configViper.UnmarshalKey("entityMatchers", &networkConfig.EntityMatchers)
	logger.Debugf("Matchers are: %+v", networkConfig.EntityMatchers)
	if err != nil {
		return err
	}

	c.networkConfig = &networkConfig
	c.networkConfigCached = true
	return nil
}

// OrderersConfig returns a list of defined orderers
func (c *Config) OrderersConfig() ([]core.OrdererConfig, error) {
	orderers := []core.OrdererConfig{}
	config, err := c.NetworkConfig()
	if err != nil {
		return nil, err
	}

	for _, orderer := range config.Orderers {

		if orderer.TLSCACerts.Path != "" {
			orderer.TLSCACerts.Path = SubstPathVars(orderer.TLSCACerts.Path)
		} else if len(orderer.TLSCACerts.Pem) == 0 && c.configViper.GetBool("client.tlsCerts.systemCertPool") == false {
			errors.Errorf("Orderer has no certs configured. Make sure TLSCACerts.Pem or TLSCACerts.Path is set for %s", orderer.URL)
		}

		orderers = append(orderers, orderer)
	}

	return orderers, nil
}

// RandomOrdererConfig returns a pseudo-random orderer from the network config
func (c *Config) RandomOrdererConfig() (*core.OrdererConfig, error) {
	orderers, err := c.OrderersConfig()
	if err != nil {
		return nil, err
	}

	return randomOrdererConfig(orderers)
}

// randomOrdererConfig returns a pseudo-random orderer from the list of orderers
func randomOrdererConfig(orderers []core.OrdererConfig) (*core.OrdererConfig, error) {

	rs := rand.NewSource(time.Now().Unix())
	r := rand.New(rs)
	randomNumber := r.Intn(len(orderers))

	return &orderers[randomNumber], nil
}

// OrdererConfig returns the requested orderer
func (c *Config) OrdererConfig(name string) (*core.OrdererConfig, error) {
	config, err := c.NetworkConfig()
	if err != nil {
		return nil, err
	}
	orderer, ok := config.Orderers[strings.ToLower(name)]
	if !ok {
		logger.Debugf("Could not find Orderer for [%s], trying with Entity Matchers", name)
		matchingOrdererConfig, matchErr := c.tryMatchingOrdererConfig(strings.ToLower(name))
		if matchErr != nil {
			return nil, errors.WithMessage(matchErr, "unable to find Orderer Config")
		}
		logger.Debugf("Found matching Orderer Config for [%s]", name)
		orderer = *matchingOrdererConfig
	}

	if orderer.TLSCACerts.Path != "" {
		orderer.TLSCACerts.Path = SubstPathVars(orderer.TLSCACerts.Path)
	}

	return &orderer, nil
}

// PeersConfig Retrieves the fabric peers for the specified org from the
// config file provided
func (c *Config) PeersConfig(org string) ([]core.PeerConfig, error) {
	config, err := c.NetworkConfig()
	if err != nil {
		return nil, err
	}

	peersConfig := config.Organizations[strings.ToLower(org)].Peers
	peers := []core.PeerConfig{}

	for _, peerName := range peersConfig {
		p := config.Peers[strings.ToLower(peerName)]
		if err = c.verifyPeerConfig(p, peerName, endpoint.IsTLSEnabled(p.URL)); err != nil {
			logger.Debugf("Could not verify Peer for [%s], trying with Entity Matchers", peerName)
			matchingPeerConfig, matchErr := c.tryMatchingPeerConfig(peerName)
			if matchErr != nil {
				return nil, errors.WithMessage(err, "unable to find Peer Config")
			}
			logger.Debugf("Found a matchingPeerConfig for [%s]", peerName)
			p = *matchingPeerConfig
		}
		if p.TLSCACerts.Path != "" {
			p.TLSCACerts.Path = SubstPathVars(p.TLSCACerts.Path)
		}

		peers = append(peers, p)
	}
	return peers, nil
}

func (c *Config) getPortIfPresent(url string) (int, bool) {
	s := strings.Split(url, ":")
	if len(s) > 1 {
		if port, err := strconv.Atoi(s[len(s)-1]); err == nil {
			return port, true
		}
	}
	return 0, false
}

func (c *Config) tryMatchingPeerConfig(peerName string) (*core.PeerConfig, error) {
	networkConfig, err := c.NetworkConfig()
	if err != nil {
		return nil, err
	}
	//Return if no peerMatchers are configured
	if len(c.peerMatchers) == 0 {
		return nil, errors.New("no Peer entityMatchers are found")
	}

	//sort the keys
	var keys []int
	for k := range c.peerMatchers {
		keys = append(keys, k)
	}
	sort.Ints(keys)

	//loop over peerentityMatchers to find the matching peer
	for _, k := range keys {
		v := c.peerMatchers[k]
		if v.MatchString(peerName) {
			// get the matching matchConfig from the index number
			peerMatchConfig := networkConfig.EntityMatchers["peer"][k]
			//Get the peerConfig from mapped host
			peerConfig, ok := networkConfig.Peers[strings.ToLower(peerMatchConfig.MappedHost)]
			if !ok {
				return nil, errors.New("failed to load config from matched Peer")
			}

			// Make a copy of GRPC options (as it is manipulated below)
			peerConfig.GRPCOptions = copyPropertiesMap(peerConfig.GRPCOptions)

			_, isPortPresentInPeerName := c.getPortIfPresent(peerName)
			//if substitution url is empty, use the same network peer url
			if peerMatchConfig.URLSubstitutionExp == "" {
				port, isPortPresent := c.getPortIfPresent(peerConfig.URL)
				peerConfig.URL = peerName
				//append port of matched config
				if isPortPresent && !isPortPresentInPeerName {
					peerConfig.URL += ":" + strconv.Itoa(port)
				}
			} else {
				//else, replace url with urlSubstitutionExp if it doesnt have any variable declarations like $
				if strings.Index(peerMatchConfig.URLSubstitutionExp, "$") < 0 {
					peerConfig.URL = peerMatchConfig.URLSubstitutionExp
				} else {
					//if the urlSubstitutionExp has $ variable declarations, use regex replaceallstring to replace networkhostname with substituionexp pattern
					peerConfig.URL = v.ReplaceAllString(peerName, peerMatchConfig.URLSubstitutionExp)
				}

			}

			//if eventSubstitution url is empty, use the same network peer url
			if peerMatchConfig.EventURLSubstitutionExp == "" {
				port, isPortPresent := c.getPortIfPresent(peerConfig.EventURL)
				peerConfig.EventURL = peerName
				//append port of matched config
				if isPortPresent && !isPortPresentInPeerName {
					peerConfig.EventURL += ":" + strconv.Itoa(port)
				}
			} else {
				//else, replace url with eventUrlSubstitutionExp if it doesnt have any variable declarations like $
				if strings.Index(peerMatchConfig.EventURLSubstitutionExp, "$") < 0 {
					peerConfig.EventURL = peerMatchConfig.EventURLSubstitutionExp
				} else {
					//if the eventUrlSubstitutionExp has $ variable declarations, use regex replaceallstring to replace networkhostname with eventsubstituionexp pattern
					peerConfig.EventURL = v.ReplaceAllString(peerName, peerMatchConfig.EventURLSubstitutionExp)
				}

			}

			//if sslTargetOverrideUrlSubstitutionExp is empty, use the same network peer host
			if peerMatchConfig.SSLTargetOverrideURLSubstitutionExp == "" {
				if strings.Index(peerName, ":") < 0 {
					peerConfig.GRPCOptions["ssl-target-name-override"] = peerName
				} else {
					//Remove port and protocol of the peerName
					s := strings.Split(peerName, ":")
					if isPortPresentInPeerName {
						peerConfig.GRPCOptions["ssl-target-name-override"] = s[len(s)-2]
					} else {
						peerConfig.GRPCOptions["ssl-target-name-override"] = s[len(s)-1]
					}
				}

			} else {
				//else, replace url with sslTargetOverrideUrlSubstitutionExp if it doesnt have any variable declarations like $
				if strings.Index(peerMatchConfig.SSLTargetOverrideURLSubstitutionExp, "$") < 0 {
					peerConfig.GRPCOptions["ssl-target-name-override"] = peerMatchConfig.SSLTargetOverrideURLSubstitutionExp
				} else {
					//if the sslTargetOverrideUrlSubstitutionExp has $ variable declarations, use regex replaceallstring to replace networkhostname with eventsubstituionexp pattern
					peerConfig.GRPCOptions["ssl-target-name-override"] = v.ReplaceAllString(peerName, peerMatchConfig.SSLTargetOverrideURLSubstitutionExp)
				}

			}
			return &peerConfig, nil
		}
	}

	return nil, errors.WithStack(status.New(status.ClientStatus, status.NoMatchingPeerEntity.ToInt32(), "no matching peer config found", nil))
}

func (c *Config) tryMatchingOrdererConfig(ordererName string) (*core.OrdererConfig, error) {
	networkConfig, err := c.NetworkConfig()
	if err != nil {
		return nil, err
	}
	//Return if no ordererMatchers are configured
	if len(c.ordererMatchers) == 0 {
		return nil, errors.New("no Orderer entityMatchers are found")
	}

	//sort the keys
	var keys []int
	for k := range c.ordererMatchers {
		keys = append(keys, k)
	}
	sort.Ints(keys)

	//loop over ordererentityMatchers to find the matching orderer
	for _, k := range keys {
		v := c.ordererMatchers[k]
		if v.MatchString(ordererName) {
			// get the matching matchConfig from the index number
			ordererMatchConfig := networkConfig.EntityMatchers["orderer"][k]
			//Get the ordererConfig from mapped host
			ordererConfig, ok := networkConfig.Orderers[strings.ToLower(ordererMatchConfig.MappedHost)]
			if !ok {
				return nil, errors.New("failed to load config from matched Orderer")
			}

			// Make a copy of GRPC options (as it is manipulated below)
			ordererConfig.GRPCOptions = copyPropertiesMap(ordererConfig.GRPCOptions)

			_, isPortPresentInOrdererName := c.getPortIfPresent(ordererName)
			//if substitution url is empty, use the same network orderer url
			if ordererMatchConfig.URLSubstitutionExp == "" {
				port, isPortPresent := c.getPortIfPresent(ordererConfig.URL)
				ordererConfig.URL = ordererName

				//append port of matched config
				if isPortPresent && !isPortPresentInOrdererName {
					ordererConfig.URL += ":" + strconv.Itoa(port)
				}
			} else {
				//else, replace url with urlSubstitutionExp if it doesnt have any variable declarations like $
				if strings.Index(ordererMatchConfig.URLSubstitutionExp, "$") < 0 {
					ordererConfig.URL = ordererMatchConfig.URLSubstitutionExp
				} else {
					//if the urlSubstitutionExp has $ variable declarations, use regex replaceallstring to replace networkhostname with substituionexp pattern
					ordererConfig.URL = v.ReplaceAllString(ordererName, ordererMatchConfig.URLSubstitutionExp)
				}
			}

			//if sslTargetOverrideUrlSubstitutionExp is empty, use the same network peer host
			if ordererMatchConfig.SSLTargetOverrideURLSubstitutionExp == "" {
				if strings.Index(ordererName, ":") < 0 {
					ordererConfig.GRPCOptions["ssl-target-name-override"] = ordererName
				} else {
					//Remove port and protocol of the ordererName
					s := strings.Split(ordererName, ":")
					if isPortPresentInOrdererName {
						ordererConfig.GRPCOptions["ssl-target-name-override"] = s[len(s)-2]
					} else {
						ordererConfig.GRPCOptions["ssl-target-name-override"] = s[len(s)-1]
					}
				}

			} else {
				//else, replace url with sslTargetOverrideUrlSubstitutionExp if it doesnt have any variable declarations like $
				if strings.Index(ordererMatchConfig.SSLTargetOverrideURLSubstitutionExp, "$") < 0 {
					ordererConfig.GRPCOptions["ssl-target-name-override"] = ordererMatchConfig.SSLTargetOverrideURLSubstitutionExp
				} else {
					//if the sslTargetOverrideUrlSubstitutionExp has $ variable declarations, use regex replaceallstring to replace networkhostname with eventsubstituionexp pattern
					ordererConfig.GRPCOptions["ssl-target-name-override"] = v.ReplaceAllString(ordererName, ordererMatchConfig.SSLTargetOverrideURLSubstitutionExp)
				}

			}
			return &ordererConfig, nil
		}
	}

	return nil, errors.WithStack(status.New(status.ClientStatus, status.NoMatchingOrdererEntity.ToInt32(), "no matching orderer config found", nil))
}

func copyPropertiesMap(origMap map[string]interface{}) map[string]interface{} {
	newMap := make(map[string]interface{}, len(origMap))
	for k, v := range origMap {
		newMap[k] = v
	}
	return newMap
}

func (c *Config) findMatchingPeer(peerName string) (string, error) {
	networkConfig, err := c.NetworkConfig()
	if err != nil {
		return "", err
	}
	//Return if no peerMatchers are configured
	if len(c.peerMatchers) == 0 {
		return "", errors.New("no Peer entityMatchers are found")
	}

	//sort the keys
	var keys []int
	for k := range c.peerMatchers {
		keys = append(keys, k)
	}
	sort.Ints(keys)

	//loop over peerentityMatchers to find the matching peer
	for _, k := range keys {
		v := c.peerMatchers[k]
		if v.MatchString(peerName) {
			// get the matching matchConfig from the index number
			peerMatchConfig := networkConfig.EntityMatchers["peer"][k]
			return peerMatchConfig.MappedHost, nil
		}
	}

	return "", errors.WithStack(status.New(status.ClientStatus, status.NoMatchingPeerEntity.ToInt32(), "no matching peer config found", nil))
}

func (c *Config) compileMatchers() error {
	networkConfig, err := c.NetworkConfig()
	if err != nil {
		return err
	}
	//return no error if entityMatchers is not configured
	if networkConfig.EntityMatchers == nil {
		return nil
	}

	if networkConfig.EntityMatchers["peer"] != nil {
		peerMatchersConfig := networkConfig.EntityMatchers["peer"]
		for i := 0; i < len(peerMatchersConfig); i++ {
			if peerMatchersConfig[i].Pattern != "" {
				c.peerMatchers[i], err = regexp.Compile(peerMatchersConfig[i].Pattern)
				if err != nil {
					return err
				}
			}
		}
	}
	if networkConfig.EntityMatchers["orderer"] != nil {
		ordererMatchersConfig := networkConfig.EntityMatchers["orderer"]
		for i := 0; i < len(ordererMatchersConfig); i++ {
			if ordererMatchersConfig[i].Pattern != "" {
				c.ordererMatchers[i], err = regexp.Compile(ordererMatchersConfig[i].Pattern)
				if err != nil {
					return err
				}
			}
		}
	}
	if networkConfig.EntityMatchers["certificateauthorities"] != nil {
		certMatchersConfig := networkConfig.EntityMatchers["certificateauthorities"]
		for i := 0; i < len(certMatchersConfig); i++ {
			if certMatchersConfig[i].Pattern != "" {
				c.caMatchers[i], err = regexp.Compile(certMatchersConfig[i].Pattern)
				if err != nil {
					return err
				}
			}
		}
	}
	return nil
}

// PeerConfigByURL retrieves PeerConfig by URL
func (c *Config) PeerConfigByURL(url string) (*core.PeerConfig, error) {
	config, err := c.NetworkConfig()
	if err != nil {
		return nil, err
	}

	var matchPeerConfig *core.PeerConfig
	staticPeers := config.Peers
	for _, staticPeerConfig := range staticPeers {
		if strings.EqualFold(staticPeerConfig.URL, url) {
			matchPeerConfig = &staticPeerConfig
			break
		}
	}

	if matchPeerConfig == nil {
		// try to match from entity matchers
		logger.Debugf("Could not find Peer for url [%s], trying with Entity Matchers", url)
		matchPeerConfig, err = c.tryMatchingPeerConfig(url)
		if err != nil {
			return nil, errors.WithMessage(err, "No Peer found with the url from config")
		}
		logger.Debugf("Found MatchingPeerConfig for url [%s]", url)
	}

	if matchPeerConfig != nil && matchPeerConfig.TLSCACerts.Path != "" {
		matchPeerConfig.TLSCACerts.Path = SubstPathVars(matchPeerConfig.TLSCACerts.Path)
	}

	return matchPeerConfig, nil
}

// PeerConfig Retrieves a specific peer from the configuration by org and name
func (c *Config) PeerConfig(org string, name string) (*core.PeerConfig, error) {
	config, err := c.NetworkConfig()
	if err != nil {
		return nil, err
	}

	peersConfig := config.Organizations[strings.ToLower(org)].Peers
	peerInOrg := false
	for _, p := range peersConfig {
		if p == name {
			peerInOrg = true
		}
	}
	if !peerInOrg {
		return nil, errors.Errorf("peer %s is not part of organization %s", name, org)
	}

	peerConfig, ok := config.Peers[strings.ToLower(name)]
	if !ok {
		logger.Debugf("Could not find Peer for [%s], trying with Entity Matchers", name)
		matchingPeerConfig, matchErr := c.tryMatchingPeerConfig(strings.ToLower(name))
		if matchErr != nil {
			return nil, errors.WithMessage(matchErr, "unable to find peer config")
		}
		logger.Debugf("Found MatchingPeerConfig for [%s]", name)
		peerConfig = *matchingPeerConfig
	}

	if peerConfig.TLSCACerts.Path != "" {
		peerConfig.TLSCACerts.Path = SubstPathVars(peerConfig.TLSCACerts.Path)
	}
	return &peerConfig, nil
}

// PeerConfig Retrieves a specific peer by name
func (c *Config) peerConfig(name string) (*core.PeerConfig, error) {
	config, err := c.NetworkConfig()
	if err != nil {
		return nil, err
	}
	peerConfig, ok := config.Peers[strings.ToLower(name)]
	if !ok {
		logger.Debugf("Could not find PeerConfig for [%s], trying with Entity Matchers", name)
		matchingPeerConfig, matchErr := c.tryMatchingPeerConfig(strings.ToLower(name))
		if matchErr != nil {
			return nil, errors.WithMessage(matchErr, "unable to find peer config")
		}
		logger.Debugf("Found MatchingPeerConfig for [%s]", name)
		peerConfig = *matchingPeerConfig
	}

	if peerConfig.TLSCACerts.Path != "" {
		peerConfig.TLSCACerts.Path = SubstPathVars(peerConfig.TLSCACerts.Path)
	}
	return &peerConfig, nil
}

// NetworkConfig returns the network configuration defined in the config file
func (c *Config) NetworkConfig() (*core.NetworkConfig, error) {
	if c.networkConfigCached {
		return c.networkConfig, nil
	}

	if err := c.cacheNetworkConfiguration(); err != nil {
		return nil, errors.WithMessage(err, "network configuration load failed")
	}
	return c.networkConfig, nil
}

// ChannelConfig returns the channel configuration
func (c *Config) ChannelConfig(name string) (*core.ChannelConfig, error) {
	config, err := c.NetworkConfig()
	if err != nil {
		return nil, err
	}

	// viper lowercases all key maps
	ch, ok := config.Channels[strings.ToLower(name)]
	if !ok {
		return nil, nil
	}

	return &ch, nil
}

// ChannelOrderers returns a list of channel orderers
func (c *Config) ChannelOrderers(name string) ([]core.OrdererConfig, error) {
	orderers := []core.OrdererConfig{}
	channel, err := c.ChannelConfig(name)
	if err != nil || channel == nil {
		return nil, errors.Errorf("Unable to retrieve channel config: %s", err)
	}

	for _, chOrderer := range channel.Orderers {
		orderer, err := c.OrdererConfig(chOrderer)
		if err != nil || orderer == nil {
			return nil, errors.Errorf("unable to retrieve orderer config: %s", err)
		}

		orderers = append(orderers, *orderer)
	}

	return orderers, nil
}

// ChannelPeers returns the channel peers configuration
func (c *Config) ChannelPeers(name string) ([]core.ChannelPeer, error) {
	netConfig, err := c.NetworkConfig()
	if err != nil {
		return nil, err
	}

	// viper lowercases all key maps
	chConfig, ok := netConfig.Channels[strings.ToLower(name)]
	if !ok {
		return nil, errors.Errorf("channel config not found for %s", name)
	}

	peers := []core.ChannelPeer{}

	for peerName, chPeerConfig := range chConfig.Peers {

		// Get generic peer configuration
		p, ok := netConfig.Peers[strings.ToLower(peerName)]
		if !ok {
			logger.Debugf("Could not find Peer for [%s], trying with Entity Matchers", peerName)
			matchingPeerConfig, matchErr := c.tryMatchingPeerConfig(strings.ToLower(peerName))
			if matchErr != nil {
				return nil, errors.Errorf("peer config not found for %s", peerName)
			}
			logger.Debugf("Found matchingPeerConfig for [%s]", peerName)
			p = *matchingPeerConfig
		}

		if err = c.verifyPeerConfig(p, peerName, endpoint.IsTLSEnabled(p.URL)); err != nil {
			return nil, err
		}

		if p.TLSCACerts.Path != "" {
			p.TLSCACerts.Path = SubstPathVars(p.TLSCACerts.Path)
		}

		mspID, err := c.PeerMSPID(peerName)
		if err != nil {
			return nil, errors.Errorf("failed to retrieve msp id for peer %s", peerName)
		}

		networkPeer := core.NetworkPeer{PeerConfig: p, MSPID: mspID}

		peer := core.ChannelPeer{PeerChannelConfig: chPeerConfig, NetworkPeer: networkPeer}

		peers = append(peers, peer)
	}

	return peers, nil

}

// NetworkPeers returns the network peers configuration
func (c *Config) NetworkPeers() ([]core.NetworkPeer, error) {
	netConfig, err := c.NetworkConfig()
	if err != nil {
		return nil, err
	}

	netPeers := []core.NetworkPeer{}

	for name, p := range netConfig.Peers {

		if err = c.verifyPeerConfig(p, name, endpoint.IsTLSEnabled(p.URL)); err != nil {
			return nil, err
		}

		if p.TLSCACerts.Path != "" {
			p.TLSCACerts.Path = SubstPathVars(p.TLSCACerts.Path)
		}

		mspID, err := c.PeerMSPID(name)
		if err != nil {
			return nil, errors.Errorf("failed to retrieve msp id for peer %s", name)
		}

		netPeer := core.NetworkPeer{PeerConfig: p, MSPID: mspID}
		netPeers = append(netPeers, netPeer)
	}

	return netPeers, nil
}

// PeerMSPID returns msp that peer belongs to
func (c *Config) PeerMSPID(name string) (string, error) {
	netConfig, err := c.NetworkConfig()
	if err != nil {
		return "", err
	}

	var mspID string

	// Find organisation/msp that peer belongs to
	for _, org := range netConfig.Organizations {
		for i := 0; i < len(org.Peers); i++ {
			if strings.EqualFold(org.Peers[i], name) {
				// peer belongs to this org add org msp
				mspID = org.MSPID
				break
			} else {
				peer, err := c.findMatchingPeer(org.Peers[i])
				if err == nil && strings.EqualFold(peer, name) {
					mspID = org.MSPID
					break
				}
			}
		}
	}

	return mspID, nil

}

func (c *Config) verifyPeerConfig(p core.PeerConfig, peerName string, tlsEnabled bool) error {
	if p.URL == "" {
		return errors.Errorf("URL does not exist or empty for peer %s", peerName)
	}
	if tlsEnabled && len(p.TLSCACerts.Pem) == 0 && p.TLSCACerts.Path == "" && c.configViper.GetBool("client.tlsCerts.systemCertPool") == false {
		return errors.Errorf("tls.certificate does not exist or empty for peer %s", peerName)
	}
	return nil
}

// TLSCACertPool returns the configured cert pool. If a certConfig
// is provided, the certficate is added to the pool
func (c *Config) TLSCACertPool(certs ...*x509.Certificate) (*x509.CertPool, error) {

	c.certPoolLock.Lock()
	defer c.certPoolLock.Unlock()

	//add cert if it is not nil and doesn't exists already
	for _, newCert := range certs {
		if newCert != nil && !c.containsCert(newCert) {
			c.tlsCerts = append(c.tlsCerts, newCert)
		}
	}

	//get new cert pool
	tlsCertPool, err := c.getCertPool()
	if err != nil {
		return nil, errors.WithMessage(err, "failed to create cert pool")
	}

	//add all tls ca certs to cert pool
	for _, cert := range c.tlsCerts {
		tlsCertPool.AddCert(cert)
	}

	return tlsCertPool, nil
}

func (c *Config) containsCert(newCert *x509.Certificate) bool {
	//TODO may need to maintain separate map of {cert.RawSubject, cert} to improve performance on search
	for _, cert := range c.tlsCerts {
		if cert.Equal(newCert) {
			return true
		}
	}
	return false
}

func (c *Config) getCertPool() (*x509.CertPool, error) {
	tlsCertPool := x509.NewCertPool()
	if c.configViper.GetBool("client.tlsCerts.systemCertPool") == true {
		var err error
		if tlsCertPool, err = x509.SystemCertPool(); err != nil {
			return nil, err
		}
		logger.Debugf("Loaded system cert pool of size: %d", len(tlsCertPool.Subjects()))
	}
	return tlsCertPool, nil
}

// IsSecurityEnabled ...
func (c *Config) IsSecurityEnabled() bool {
	return c.configViper.GetBool("client.BCCSP.security.enabled")
}

// SecurityAlgorithm ...
func (c *Config) SecurityAlgorithm() string {
	return c.configViper.GetString("client.BCCSP.security.hashAlgorithm")
}

// SecurityLevel ...
func (c *Config) SecurityLevel() int {
	return c.configViper.GetInt("client.BCCSP.security.level")
}

//SecurityProvider provider SW or PKCS11
func (c *Config) SecurityProvider() string {
	return c.configViper.GetString("client.BCCSP.security.default.provider")
}

//Ephemeral flag
func (c *Config) Ephemeral() bool {
	return c.configViper.GetBool("client.BCCSP.security.ephemeral")
}

//SoftVerify flag
func (c *Config) SoftVerify() bool {
	return c.configViper.GetBool("client.BCCSP.security.softVerify")
}

//SecurityProviderLibPath will be set only if provider is PKCS11
func (c *Config) SecurityProviderLibPath() string {
	configuredLibs := c.configViper.GetString("client.BCCSP.security.library")
	libPaths := strings.Split(configuredLibs, ",")
	logger.Debug("Configured BCCSP Lib Paths %v", libPaths)
	var lib string
	for _, path := range libPaths {
		if _, err := os.Stat(strings.TrimSpace(path)); !os.IsNotExist(err) {
			lib = strings.TrimSpace(path)
			break
		}
	}
	if lib != "" {
		logger.Debug("Found softhsm library: %s", lib)
	} else {
		logger.Debug("Softhsm library was not found")
	}
	return lib
}

//SecurityProviderPin will be set only if provider is PKCS11
func (c *Config) SecurityProviderPin() string {
	return c.configViper.GetString("client.BCCSP.security.pin")
}

//SecurityProviderLabel will be set only if provider is PKCS11
func (c *Config) SecurityProviderLabel() string {
	return c.configViper.GetString("client.BCCSP.security.label")
}

// CredentialStorePath returns the user store path
func (c *Config) CredentialStorePath() string {
	return SubstPathVars(c.configViper.GetString("client.credentialStore.path"))
}

// KeyStorePath returns the keystore path used by BCCSP
func (c *Config) KeyStorePath() string {
	keystorePath := SubstPathVars(c.configViper.GetString("client.credentialStore.cryptoStore.path"))
	return path.Join(keystorePath, "keystore")
}

// CAKeyStorePath returns the same path as KeyStorePath() without the
// 'keystore' directory added. This is done because the fabric-ca-client
// adds this to the path
func (c *Config) CAKeyStorePath() string {
	return SubstPathVars(c.configViper.GetString("client.credentialStore.cryptoStore.path"))
}

// CryptoConfigPath ...
func (c *Config) CryptoConfigPath() string {
	return SubstPathVars(c.configViper.GetString("client.cryptoconfig.path"))
}

// TLSClientCerts loads the client's certs for mutual TLS
// It checks the config for embedded pem files before looking for cert files
func (c *Config) TLSClientCerts() ([]tls.Certificate, error) {
	clientConfig, err := c.Client()
	if err != nil {
		return nil, err
	}
	var clientCerts tls.Certificate
	var cb, kb []byte
	cb, err = clientConfig.TLSCerts.Client.Cert.Bytes()
	if err != nil {
		return nil, errors.Wrapf(err, "failed to load tls client cert")
	}

	if len(cb) == 0 {
		// if no cert found in the config, return empty cert chain
		return []tls.Certificate{clientCerts}, nil
	}

	// Load private key from cert using default crypto suite
	cs := cs.GetDefault()
	pk, err := cryptoutil.GetPrivateKeyFromCert(cb, cs)

	// If CryptoSuite fails to load private key from cert then load private key from config
	if err != nil || pk == nil {
		logger.Debugf("Reading pk from config, unable to retrieve from cert: %s", err)
		if clientConfig.TLSCerts.Client.Key.Pem != "" {
			kb = []byte(clientConfig.TLSCerts.Client.Key.Pem)
		} else if clientConfig.TLSCerts.Client.Key.Path != "" {
			kb, err = loadByteKeyOrCertFromFile(clientConfig, true)
			if err != nil {
				return nil, errors.Wrapf(err, "Failed to load key from file path '%s'", clientConfig.TLSCerts.Client.Key.Path)
			}
		}

		// load the key/cert pair from []byte
		clientCerts, err = tls.X509KeyPair(cb, kb)
		if err != nil {
			return nil, errors.Errorf("Error loading cert/key pair as TLS client credentials: %v", err)
		}

		return []tls.Certificate{clientCerts}, nil

	}

	// private key was retrieved from cert
	clientCerts, err = cryptoutil.X509KeyPair(cb, pk, cs)
	if err != nil {
		return nil, err
	}

	return []tls.Certificate{clientCerts}, nil
}

// NetworkPeerConfigFromURL fetches the peer configuration based on a URL.
func NetworkPeerConfigFromURL(cfg core.Config, url string) (*core.NetworkPeer, error) {
	peerCfg, err := cfg.PeerConfigByURL(url)
	if err != nil {
		return nil, errors.WithMessage(err, "peer not found")
	}
	if peerCfg == nil {
		return nil, errors.New("peer not found")
	}

	// find MSP ID
	networkPeers, err := cfg.NetworkPeers()
	if err != nil {
		return nil, errors.WithMessage(err, "unable to load network peer config")
	}

	var mspID string
	for _, peer := range networkPeers {
		if peer.URL == peerCfg.URL { // need to use the looked-up URL due to matching
			mspID = peer.MSPID
			break
		}
	}

	np := core.NetworkPeer{
		PeerConfig: *peerCfg,
		MSPID:      mspID,
	}

	return &np, nil
}

func loadByteKeyOrCertFromFile(c *core.ClientConfig, isKey bool) ([]byte, error) {
	var path string
	a := "key"
	if isKey {
		path = SubstPathVars(c.TLSCerts.Client.Key.Path)
		c.TLSCerts.Client.Key.Path = path
	} else {
		a = "cert"
		path = SubstPathVars(c.TLSCerts.Client.Cert.Path)
		c.TLSCerts.Client.Cert.Path = path
	}
	bts, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, errors.Errorf("Error loading %s file from '%s' err: %v", a, path, err)
	}
	return bts, nil
}

// loadCAKey
func loadCAKey(rawData []byte) (*x509.Certificate, error) {
	block, _ := pem.Decode(rawData)

	if block != nil {
		pub, err := x509.ParseCertificate(block.Bytes)
		if err != nil {
			return nil, errors.Wrap(err, "certificate parsing failed")
		}

		return pub, nil
	}
	return nil, errors.New("pem data missing")
}
