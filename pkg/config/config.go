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
	"io/ioutil"
	"math/rand"
	"os"
	"path"
	"strings"
	"time"

	"github.com/spf13/viper"

	"github.com/hyperledger/fabric-sdk-go/api/apiconfig"

	"github.com/hyperledger/fabric-sdk-go/api/apilogging"
	"github.com/hyperledger/fabric-sdk-go/pkg/config/urlutil"
	"github.com/hyperledger/fabric-sdk-go/pkg/errors"
	"github.com/hyperledger/fabric-sdk-go/pkg/logging"
	lu "github.com/hyperledger/fabric-sdk-go/pkg/logging/utils"
)

var logger = logging.NewLogger(logModule)

const (
	cmdRoot        = "fabric_sdk"
	logModule      = "fabric_sdk_go"
	defaultTimeout = time.Second * 5
)

// Config represents the configuration for the client
type Config struct {
	tlsCertPool         *x509.CertPool
	networkConfig       *apiconfig.NetworkConfig
	networkConfigCached bool
	configViper         *viper.Viper
}

// InitConfigFromBytes will initialize the configs from a bytes array
func InitConfigFromBytes(configBytes []byte, configType string) (*Config, error) {
	myViper := getNewViper(cmdRoot)

	if configType == "" {
		return nil, errors.New("empty config type")
	}

	if len(configBytes) > 0 {
		buf := bytes.NewBuffer(configBytes)
		logger.Debugf("buf Len is %d, Cap is %d: %s", buf.Len(), buf.Cap(), buf)
		// read config from bytes array, but must set ConfigType
		// for viper to properly unmarshal the bytes array
		myViper.SetConfigType(configType)
		myViper.ReadConfig(buf)
	} else {
		return nil, errors.New("empty config bytes array")
	}

	setLogLevel(myViper)

	tlsCertPool, err := getCertPool(myViper)
	if err != nil {
		return nil, err
	}

	return &Config{tlsCertPool: tlsCertPool, configViper: myViper}, nil
}

// getNewViper returns a new instance of viper
func getNewViper(cmdRootPrefix string) *viper.Viper {
	myViper := viper.New()
	myViper.SetEnvPrefix(cmdRootPrefix)
	myViper.AutomaticEnv()
	replacer := strings.NewReplacer(".", "_")
	myViper.SetEnvKeyReplacer(replacer)
	return myViper
}

// InitConfig reads in config file
func InitConfig(configFile string) (*Config, error) {
	return initConfigWithCmdRoot(configFile, cmdRoot)
}

// initConfigWithCmdRoot reads in a config file and allows the
// environment variable prefixed to be specified
func initConfigWithCmdRoot(configFile string, cmdRootPrefix string) (*Config, error) {
	myViper := getNewViper(cmdRootPrefix)

	// load default config
	// for now, loading DefaultConfig only works with File, not []Byte due to viper's ConfigType
	err := loadDefaultConfig(myViper)
	if err != nil {
		return nil, err
	}

	if configFile != "" {
		// create new viper
		myViper.SetConfigFile(configFile)
		// If a config file is found, read it in.
		err := myViper.ReadInConfig()

		if err == nil {
			logger.Debugf("Using config file: %s", myViper.ConfigFileUsed())
		} else {
			return nil, errors.Wrap(err, "loading config file failed")
		}
	}

	setLogLevel(myViper)
	tlsCertPool, err := getCertPool(myViper)
	if err != nil {
		return nil, err
	}

	logger.Infof("%s logging level is set to: %s", logModule, lu.LogLevelString(logging.GetLevel(logModule)))
	return &Config{tlsCertPool: tlsCertPool, configViper: myViper}, nil
}

func getCertPool(myViper *viper.Viper) (*x509.CertPool, error) {
	tlsCertPool := x509.NewCertPool()
	if myViper.GetBool("client.tlsCerts.systemCertPool") == true {
		var err error
		if tlsCertPool, err = x509.SystemCertPool(); err != nil {
			return nil, err
		}
		logger.Debugf("Loaded system cert pool of size: %d", len(tlsCertPool.Subjects()))
	}
	return tlsCertPool, nil
}

// setLogLevel will set the log level of the client
func setLogLevel(myViper *viper.Viper) {
	loggingLevelString := myViper.GetString("client.logging.level")
	logLevel := apilogging.INFO
	if loggingLevelString != "" {
		logger.Debugf("%s logging level from the config: %v", logModule, loggingLevelString)
		var err error
		logLevel, err = logging.LogLevel(loggingLevelString)
		if err != nil {
			panic(err)
		}
	}
	logging.SetLevel(logModule, logLevel)
}

// load Default config
func loadDefaultConfig(myViper *viper.Viper) error {
	// get Environment Default Config Path
	defaultPath := os.Getenv("FABRIC_SDK_CONFIG_PATH")
	if defaultPath == "" {
		return nil
	}
	// if set, use it to load default config
	myViper.AddConfigPath(substPathVars(defaultPath))
	err := myViper.ReadInConfig() // Find and read the config file
	if err != nil {               // Handle errors reading the config file
		return errors.Wrap(err, "loading config file failed")
	}
	return nil
}

// Client returns the Client config
func (c *Config) Client() (*apiconfig.ClientConfig, error) {
	config, err := c.NetworkConfig()
	if err != nil {
		return nil, err
	}
	client := config.Client

	client.TLSCerts.Path = substPathVars(client.TLSCerts.Path)
	client.TLSCerts.Client.Keyfile = substPathVars(client.TLSCerts.Client.Keyfile)
	client.TLSCerts.Client.Certfile = substPathVars(client.TLSCerts.Client.Certfile)

	return &client, nil
}

// CAConfig returns the CA configuration.
func (c *Config) CAConfig(org string) (*apiconfig.CAConfig, error) {
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
		certFileModPath[i] = substPathVars(v)
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
	return substPathVars(config.CertificateAuthorities[strings.ToLower(caName)].TLSCACerts.Client.Keyfile), nil
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
	if len(ca.TLSCACerts.Client.CertPem) == 0 {
		return "", errors.New("Empty Client Key Pem")
	}

	return ca.TLSCACerts.Client.KeyPem, nil
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
		return "", errors.Errorf("CA Server Name '%s' not found", caName)
	}
	return substPathVars(config.CertificateAuthorities[strings.ToLower(caName)].TLSCACerts.Client.Certfile), nil
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
	if len(ca.TLSCACerts.Client.CertPem) == 0 {
		return "", errors.New("Empty Client Cert Pem")
	}

	return ca.TLSCACerts.Client.CertPem, nil
}

// TimeoutOrDefault reads connection timeouts for the given connection type
func (c *Config) TimeoutOrDefault(conn apiconfig.TimeoutType) time.Duration {
	var timeout time.Duration
	switch conn {
	case apiconfig.Endorser:
		timeout = c.configViper.GetDuration("client.peer.timeout.connection")
	case apiconfig.Query:
		timeout = c.configViper.GetDuration("client.peer.timeout.queryResponse")
	case apiconfig.ExecuteTx:
		timeout = c.configViper.GetDuration("client.peer.timeout.executeTxResponse")
	case apiconfig.EventHub:
		timeout = c.configViper.GetDuration("client.eventService.timeout.connection")
	case apiconfig.EventReg:
		timeout = c.configViper.GetDuration("client.eventService.timeout.registrationResponse")
	case apiconfig.OrdererConnection:
		timeout = c.configViper.GetDuration("client.orderer.timeout.connection")
	case apiconfig.OrdererResponse:
		timeout = c.configViper.GetDuration("client.orderer.timeout.response")

	}
	if timeout == 0 {
		timeout = defaultTimeout
	}

	return timeout
}

// MspID returns the MSP ID for the requested organization
func (c *Config) MspID(org string) (string, error) {
	config, err := c.NetworkConfig()
	if err != nil {
		return "", err
	}
	// viper lowercases all key maps, org is lower case
	mspID := config.Organizations[strings.ToLower(org)].MspID
	if mspID == "" {
		return "", errors.Errorf("MSP ID is empty for org: %s", org)
	}

	return mspID, nil
}

func (c *Config) cacheNetworkConfiguration() error {
	c.networkConfig = new(apiconfig.NetworkConfig)
	c.networkConfig.Name = c.configViper.GetString("name")
	c.networkConfig.Xtype = c.configViper.GetString("x-type")
	c.networkConfig.Description = c.configViper.GetString("description")
	c.networkConfig.Version = c.configViper.GetString("version")

	err := c.configViper.UnmarshalKey("client", &c.networkConfig.Client)
	logger.Debugf("Client is: %+v", c.networkConfig.Client)
	if err != nil {
		return err
	}
	err = c.configViper.UnmarshalKey("channels", &c.networkConfig.Channels)
	logger.Debugf("channels are: %+v", c.networkConfig.Channels)
	if err != nil {
		return err
	}
	err = c.configViper.UnmarshalKey("organizations", &c.networkConfig.Organizations)
	logger.Debugf("organizations are: %+v", c.networkConfig.Organizations)
	if err != nil {
		return err
	}
	err = c.configViper.UnmarshalKey("orderers", &c.networkConfig.Orderers)
	logger.Debugf("orderers are: %+v", c.networkConfig.Orderers)
	if err != nil {
		return err
	}
	err = c.configViper.UnmarshalKey("peers", &c.networkConfig.Peers)
	logger.Debugf("peers are: %+v", c.networkConfig.Peers)
	if err != nil {
		return err
	}
	err = c.configViper.UnmarshalKey("certificateAuthorities", &c.networkConfig.CertificateAuthorities)
	logger.Debugf("certificateAuthorities are: %+v", c.networkConfig.CertificateAuthorities)
	if err != nil {
		return err
	}

	c.networkConfigCached = true
	return err
}

// OrderersConfig returns a list of defined orderers
func (c *Config) OrderersConfig() ([]apiconfig.OrdererConfig, error) {
	orderers := []apiconfig.OrdererConfig{}
	config, err := c.NetworkConfig()
	if err != nil {
		return nil, err
	}

	for _, orderer := range config.Orderers {

		if orderer.TLSCACerts.Path != "" {
			orderer.TLSCACerts.Path = substPathVars(orderer.TLSCACerts.Path)
		} else if len(orderer.TLSCACerts.Pem) == 0 && c.configViper.GetBool("client.tlsCerts.systemCertPool") == false {
			errors.Errorf("Orderer has no certs configured. Make sure TLSCACerts.Pem or TLSCACerts.Path is set for %s", orderer.URL)
		}

		orderers = append(orderers, orderer)
	}

	return orderers, nil
}

// RandomOrdererConfig returns a pseudo-random orderer from the network config
func (c *Config) RandomOrdererConfig() (*apiconfig.OrdererConfig, error) {
	orderers, err := c.OrderersConfig()
	if err != nil {
		return nil, err
	}

	return randomOrdererConfig(orderers)
}

// randomOrdererConfig returns a pseudo-random orderer from the list of orderers
func randomOrdererConfig(orderers []apiconfig.OrdererConfig) (*apiconfig.OrdererConfig, error) {

	rs := rand.NewSource(time.Now().Unix())
	r := rand.New(rs)
	randomNumber := r.Intn(len(orderers))

	return &orderers[randomNumber], nil
}

// OrdererConfig returns the requested orderer
func (c *Config) OrdererConfig(name string) (*apiconfig.OrdererConfig, error) {
	config, err := c.NetworkConfig()
	if err != nil {
		return nil, err
	}
	orderer, ok := config.Orderers[strings.ToLower(name)]
	if !ok {
		return nil, nil
	}

	if orderer.TLSCACerts.Path != "" {
		orderer.TLSCACerts.Path = substPathVars(orderer.TLSCACerts.Path)
	}

	return &orderer, nil
}

// PeersConfig Retrieves the fabric peers for the specified org from the
// config file provided
func (c *Config) PeersConfig(org string) ([]apiconfig.PeerConfig, error) {
	config, err := c.NetworkConfig()
	if err != nil {
		return nil, err
	}

	peersConfig := config.Organizations[strings.ToLower(org)].Peers
	peers := []apiconfig.PeerConfig{}

	for _, peerName := range peersConfig {
		p := config.Peers[strings.ToLower(peerName)]
		if err = c.verifyPeerConfig(p, peerName, urlutil.IsTLSEnabled(p.URL)); err != nil {
			return nil, err
		}
		if p.TLSCACerts.Path != "" {
			p.TLSCACerts.Path = substPathVars(p.TLSCACerts.Path)
		}

		peers = append(peers, p)
	}
	return peers, nil
}

// PeerConfig Retrieves a specific peer from the configuration by org and name
func (c *Config) PeerConfig(org string, name string) (*apiconfig.PeerConfig, error) {
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
		return nil, nil
	}

	if peerConfig.TLSCACerts.Path != "" {
		peerConfig.TLSCACerts.Path = substPathVars(peerConfig.TLSCACerts.Path)
	}
	return &peerConfig, nil
}

// NetworkConfig returns the network configuration defined in the config file
func (c *Config) NetworkConfig() (*apiconfig.NetworkConfig, error) {
	if c.networkConfigCached {
		return c.networkConfig, nil
	}

	if err := c.cacheNetworkConfiguration(); err != nil {
		return nil, errors.WithMessage(err, "network configuration load failed")
	}
	return c.networkConfig, nil
}

// ChannelConfig returns the channel configuration
func (c *Config) ChannelConfig(name string) (*apiconfig.ChannelConfig, error) {
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
func (c *Config) ChannelOrderers(name string) ([]apiconfig.OrdererConfig, error) {
	orderers := []apiconfig.OrdererConfig{}
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
func (c *Config) ChannelPeers(name string) ([]apiconfig.ChannelPeer, error) {
	netConfig, err := c.NetworkConfig()
	if err != nil {
		return nil, err
	}

	// viper lowercases all key maps
	chConfig, ok := netConfig.Channels[strings.ToLower(name)]
	if !ok {
		return nil, errors.Errorf("channel config not found for %s", name)
	}

	peers := []apiconfig.ChannelPeer{}

	for peerName, chPeerConfig := range chConfig.Peers {

		// Get generic peer configuration
		p, ok := netConfig.Peers[strings.ToLower(peerName)]
		if !ok {
			return nil, errors.Errorf("peer config not found for %s", peerName)
		}

		if err = c.verifyPeerConfig(p, peerName, urlutil.IsTLSEnabled(p.URL)); err != nil {
			return nil, err
		}

		if p.TLSCACerts.Path != "" {
			p.TLSCACerts.Path = substPathVars(p.TLSCACerts.Path)
		}

		mspID, err := c.PeerMspID(peerName)
		if err != nil {
			return nil, errors.Errorf("failed to retrieve msp id for peer %s", peerName)
		}

		networkPeer := apiconfig.NetworkPeer{PeerConfig: p, MspID: mspID}

		peer := apiconfig.ChannelPeer{PeerChannelConfig: chPeerConfig, NetworkPeer: networkPeer}

		peers = append(peers, peer)
	}

	return peers, nil

}

// NetworkPeers returns the network peers configuration
func (c *Config) NetworkPeers() ([]apiconfig.NetworkPeer, error) {
	netConfig, err := c.NetworkConfig()
	if err != nil {
		return nil, err
	}

	netPeers := []apiconfig.NetworkPeer{}

	for name, p := range netConfig.Peers {

		if err = c.verifyPeerConfig(p, name, urlutil.IsTLSEnabled(p.URL)); err != nil {
			return nil, err
		}

		if p.TLSCACerts.Path != "" {
			p.TLSCACerts.Path = substPathVars(p.TLSCACerts.Path)
		}

		mspID, err := c.PeerMspID(name)
		if err != nil {
			return nil, errors.Errorf("failed to retrieve msp id for peer %s", name)
		}

		netPeer := apiconfig.NetworkPeer{PeerConfig: p, MspID: mspID}
		netPeers = append(netPeers, netPeer)
	}

	return netPeers, nil
}

// PeerMspID returns msp that peer belongs to
func (c *Config) PeerMspID(name string) (string, error) {
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
				mspID = org.MspID
				break
			}
		}
	}

	return mspID, nil

}

func (c *Config) verifyPeerConfig(p apiconfig.PeerConfig, peerName string, tlsEnabled bool) error {
	if p.URL == "" {
		return errors.Errorf("URL does not exist or empty for peer %s", peerName)
	}
	if p.EventURL == "" {
		return errors.Errorf("event URL does not exist or empty for peer %s", peerName)
	}
	if tlsEnabled && len(p.TLSCACerts.Pem) == 0 && p.TLSCACerts.Path == "" && c.configViper.GetBool("client.tlsCerts.systemCertPool") == false {
		return errors.Errorf("tls.certificate does not exist or empty for peer %s", peerName)
	}
	return nil
}

// SetTLSCACertPool allows a user to set a global cert pool with a set of
// root TLS CAs that will be used for all outgoing connections
func (c *Config) SetTLSCACertPool(certPool *x509.CertPool) {
	if certPool == nil {
		certPool = x509.NewCertPool()
	}
	c.tlsCertPool = certPool
}

// TLSCACertPool returns the configured cert pool. If a tlsCertificate path
// is provided, the certficate is added to the pool
func (c *Config) TLSCACertPool(tlsCertificate string) (*x509.CertPool, error) {
	if tlsCertificate != "" {
		rawData, err := ioutil.ReadFile(tlsCertificate)
		if err != nil {
			return nil, err
		}

		cert, err := loadCAKey(rawData)
		if err != nil {
			return nil, err
		}

		c.tlsCertPool.AddCert(cert)
	}

	return c.tlsCertPool, nil
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

// KeyStorePath returns the keystore path used by BCCSP
func (c *Config) KeyStorePath() string {
	keystorePath := substPathVars(c.configViper.GetString("client.credentialStore.cryptoStore.path"))
	return path.Join(keystorePath, "keystore")
}

// CAKeyStorePath returns the same path as KeyStorePath() without the
// 'keystore' directory added. This is done because the fabric-ca-client
// adds this to the path
func (c *Config) CAKeyStorePath() string {
	return substPathVars(c.configViper.GetString("client.credentialStore.cryptoStore.path"))
}

// CryptoConfigPath ...
func (c *Config) CryptoConfigPath() string {
	return substPathVars(c.configViper.GetString("client.cryptoconfig.path"))
}

// TLSClientCerts loads the client's certs for mutual TLS
// It checks the config for embedded pem files before looking for cert files
func (c *Config) TLSClientCerts() ([]tls.Certificate, error) {
	config, err := c.NetworkConfig()
	if err != nil {
		return nil, err
	}

	clientConfig := config.Client
	var clientCerts tls.Certificate
	var cb, kb []byte
	if clientConfig.TLSCerts.Client.CertPem != "" {
		cb = []byte(clientConfig.TLSCerts.Client.CertPem)
	} else if clientConfig.TLSCerts.Client.Certfile != "" {
		cb, err = loadByteKeyOrCertFromFile(&clientConfig, false)
		if err != nil {
			return nil, errors.Wrapf(err, "Failed to load cert from file path '%s'", clientConfig.TLSCerts.Client.Certfile)
		}
	}

	if clientConfig.TLSCerts.Client.KeyPem != "" {
		kb = []byte(clientConfig.TLSCerts.Client.KeyPem)
	} else if clientConfig.TLSCerts.Client.Keyfile != "" {
		kb, err = loadByteKeyOrCertFromFile(&clientConfig, true)
		if err != nil {
			return nil, errors.Wrapf(err, "Failed to load key from file path '%s'", clientConfig.TLSCerts.Client.Keyfile)
		}
	}

	if len(cb) == 0 && len(kb) == 0 {
		// if no cert found in the config, return empty cert chain
		return []tls.Certificate{clientCerts}, nil
	}

	// load the key/cert pair from []byte
	clientCerts, err = tls.X509KeyPair(cb, kb)
	if err != nil {
		return nil, errors.Errorf("Error loading cert/key pair as TLS client credentials: %v", err)
	}

	return []tls.Certificate{clientCerts}, nil
}

func loadByteKeyOrCertFromFile(c *apiconfig.ClientConfig, isKey bool) ([]byte, error) {
	var path string
	a := "key"
	if isKey {
		path = substPathVars(c.TLSCerts.Client.Keyfile)
		c.TLSCerts.Client.Keyfile = path
	} else {
		a = "cert"
		path = substPathVars(c.TLSCerts.Client.Certfile)
		c.TLSCerts.Client.Certfile = path
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
