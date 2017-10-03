/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package config

import (
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"io/ioutil"
	"math/rand"
	"os"
	"path"
	"path/filepath"
	"strings"
	"time"

	"github.com/spf13/viper"

	"github.com/hyperledger/fabric-sdk-go/api/apiconfig"

	"github.com/hyperledger/fabric-sdk-go/pkg/errors"
	"github.com/hyperledger/fabric-sdk-go/pkg/logging"
	bccspFactory "github.com/hyperledger/fabric-sdk-go/third_party/github.com/hyperledger/fabric/bccsp/factory"
	"github.com/hyperledger/fabric-sdk-go/third_party/github.com/hyperledger/fabric/bccsp/pkcs11"
)

var myViper = viper.New()
var logger = logging.NewLogger("fabric_sdk_go")

const (
	cmdRoot        = "fabric_sdk"
	defaultTimeout = time.Second * 5
)

// Config represents the configuration for the client
type Config struct {
	tlsCertPool         *x509.CertPool
	networkConfig       *apiconfig.NetworkConfig
	networkConfigCached bool
}

// InitConfig ...
// initConfig reads in config file
func InitConfig(configFile string) (apiconfig.Config, error) {
	return InitConfigWithCmdRoot(configFile, cmdRoot)
}

// InitConfigWithCmdRoot reads in a config file and allows the
// environment variable prefixed to be specified
func InitConfigWithCmdRoot(configFile string, cmdRootPrefix string) (*Config, error) {
	myViper.SetEnvPrefix(cmdRootPrefix)
	myViper.AutomaticEnv()
	replacer := strings.NewReplacer(".", "_")
	myViper.SetEnvKeyReplacer(replacer)
	err := loadDefaultConfig()
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

	loggingLevelString := myViper.GetString("client.logging.level")
	logLevel := logging.INFO
	if loggingLevelString != "" {
		logger.Infof("fabric_sdk_go Logging level from the config: %v", loggingLevelString)
		var err error
		logLevel, err = logging.LogLevel(loggingLevelString)
		if err != nil {
			panic(err)
		}
	}
	logging.SetLevel(logging.Level(logLevel), "fabric_sdk_go")

	logger.Infof("fabric_sdk_go Logging level is finally set to: %s", logging.GetLevel("fabric_sdk_go"))
	return &Config{tlsCertPool: x509.NewCertPool()}, nil
}

// load Default confid
func loadDefaultConfig() error {
	// get Environment Default Config Path
	defaultPath := os.Getenv("FABRIC_SDK_CONFIG_PATH")
	if defaultPath != "" { // if set, use it to load default config
		myViper.AddConfigPath(strings.Replace(defaultPath, "$GOPATH", os.Getenv("GOPATH"), -1))
	} else { // else fallback to default DEV path
		devPath := filepath.Join(os.Getenv("GOPATH"), "src", "github.com", "hyperledger", "fabric-sdk-go", "pkg", "config")
		myViper.AddConfigPath(devPath)
	}
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

// CAServerCertFiles Read configuration option for the server certificate files
func (c *Config) CAServerCertFiles(org string) ([]string, error) {
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
		certFileModPath[i] = strings.Replace(v, "$GOPATH", os.Getenv("GOPATH"), -1)
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

// CAClientKeyFile Read configuration option for the fabric CA client key file
func (c *Config) CAClientKeyFile(org string) (string, error) {
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
	return strings.Replace(config.CertificateAuthorities[strings.ToLower(caName)].TLSCACerts.Client.Keyfile,
		"$GOPATH", os.Getenv("GOPATH"), -1), nil
}

// CAClientCertFile Read configuration option for the fabric CA client cert file
func (c *Config) CAClientCertFile(org string) (string, error) {
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
	return strings.Replace(config.CertificateAuthorities[strings.ToLower(caName)].TLSCACerts.Client.Certfile,
		"$GOPATH", os.Getenv("GOPATH"), -1), nil
}

// TimeoutOrDefault reads connection timeouts for the given connection type
func (c *Config) TimeoutOrDefault(conn apiconfig.TimeoutType) time.Duration {
	var timeout time.Duration
	switch conn {
	case apiconfig.Endorser:
		timeout = myViper.GetDuration("client.peer.timeout.connection")
	case apiconfig.Query:
		timeout = myViper.GetDuration("client.peer.timeout.queryResponse")
	case apiconfig.ExecuteTx:
		timeout = myViper.GetDuration("client.peer.timeout.executeTxResponse")
	case apiconfig.EventHub:
		timeout = myViper.GetDuration("client.eventService.timeout.connection")
	case apiconfig.EventReg:
		timeout = myViper.GetDuration("client.eventService.timeout.registrationResponse")
	case apiconfig.OrdererConnection:
		timeout = myViper.GetDuration("client.orderer.timeout.connection")
	case apiconfig.OrdererResponse:
		timeout = myViper.GetDuration("client.orderer.timeout.response")

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

// FabricClientViper returns the internal viper instance used by the
// SDK to read configuration options
func FabricClientViper() *viper.Viper {
	return myViper
}

func (c *Config) cacheNetworkConfiguration() error {
	c.networkConfig = new(apiconfig.NetworkConfig)
	c.networkConfig.Name = myViper.GetString("name")
	c.networkConfig.Xtype = myViper.GetString("x-type")
	c.networkConfig.Description = myViper.GetString("description")
	c.networkConfig.Version = myViper.GetString("version")

	err := myViper.UnmarshalKey("client", &c.networkConfig.Client)
	logger.Debugf("Client is: %+v", c.networkConfig.Client)
	if err != nil {
		return err
	}
	err = myViper.UnmarshalKey("channels", &c.networkConfig.Channels)
	logger.Debugf("channels are: %+v", c.networkConfig.Channels)
	if err != nil {
		return err
	}
	err = myViper.UnmarshalKey("organizations", &c.networkConfig.Organizations)
	logger.Debugf("organizations are: %+v", c.networkConfig.Organizations)
	if err != nil {
		return err
	}
	err = myViper.UnmarshalKey("orderers", &c.networkConfig.Orderers)
	logger.Debugf("orderers are: %+v", c.networkConfig.Orderers)
	if err != nil {
		return err
	}
	err = myViper.UnmarshalKey("peers", &c.networkConfig.Peers)
	logger.Debugf("peers are: %+v", c.networkConfig.Peers)
	if err != nil {
		return err
	}
	err = myViper.UnmarshalKey("certificateAuthorities", &c.networkConfig.CertificateAuthorities)
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
			orderer.TLSCACerts.Path = strings.Replace(orderer.TLSCACerts.Path, "$GOPATH",
				os.Getenv("GOPATH"), -1)
		}

		orderers = append(orderers, orderer)
	}

	return orderers, nil
}

// RandomOrdererConfig returns a pseudo-random orderer from the network config
func (c *Config) RandomOrdererConfig() (*apiconfig.OrdererConfig, error) {
	config, err := c.NetworkConfig()
	if err != nil {
		return nil, err
	}

	rs := rand.NewSource(time.Now().Unix())
	r := rand.New(rs)
	randomNumber := r.Intn(len(config.Orderers))

	var i int
	for _, value := range config.Orderers {
		if value.TLSCACerts.Path != "" {
			value.TLSCACerts.Path = strings.Replace(value.TLSCACerts.Path, "$GOPATH",
				os.Getenv("GOPATH"), -1)
		}
		if i == randomNumber {
			return &value, nil
		}
		i++
	}

	return nil, nil
}

// OrdererConfig returns the requested orderer
func (c *Config) OrdererConfig(name string) (*apiconfig.OrdererConfig, error) {
	config, err := c.NetworkConfig()
	if err != nil {
		return nil, err
	}
	orderer := config.Orderers[strings.ToLower(name)]
	if orderer.TLSCACerts.Path != "" {
		orderer.TLSCACerts.Path = strings.Replace(orderer.TLSCACerts.Path, "$GOPATH",
			os.Getenv("GOPATH"), -1)
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
		if err = verifyPeerConfig(p, peerName, c.IsTLSEnabled()); err != nil {
			return nil, err
		}
		if p.TLSCACerts.Path != "" {
			p.TLSCACerts.Path = strings.Replace(p.TLSCACerts.Path, "$GOPATH",
				os.Getenv("GOPATH"), -1)
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
		return nil, errors.Errorf("peer %s is not part of orgianzation %s", name, org)
	}
	peerConfig := config.Peers[strings.ToLower(name)]
	if peerConfig.TLSCACerts.Path != "" {
		peerConfig.TLSCACerts.Path = strings.Replace(peerConfig.TLSCACerts.Path, "$GOPATH",
			os.Getenv("GOPATH"), -1)
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
		return nil, errors.Errorf("channel config not found for %s", name)
	}

	return &ch, nil
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

		if err = verifyPeerConfig(p, peerName, c.IsTLSEnabled()); err != nil {
			return nil, err
		}

		if p.TLSCACerts.Path != "" {
			p.TLSCACerts.Path = strings.Replace(p.TLSCACerts.Path, "$GOPATH", os.Getenv("GOPATH"), -1)
		}

		var mspID string

		// Find organisation/msp that peer belongs to
		for _, org := range netConfig.Organizations {
			for i := 0; i < len(org.Peers); i++ {
				if strings.EqualFold(org.Peers[i], peerName) {
					// peer belongs to this org add org msp
					mspID = org.MspID
					break
				}
			}
		}

		peer := apiconfig.ChannelPeer{PeerChannelConfig: chPeerConfig, PeerConfig: p, MspID: mspID}

		peers = append(peers, peer)
	}

	return peers, nil

}

func verifyPeerConfig(p apiconfig.PeerConfig, peerName string, tlsEnabled bool) error {
	if p.URL == "" {
		return errors.Errorf("URL does not exist or empty for peer %s", peerName)
	}
	if p.EventURL == "" {
		return errors.Errorf("event URL does not exist or empty for peer %s", peerName)
	}
	if tlsEnabled && p.TLSCACerts.Pem == "" && p.TLSCACerts.Path == "" {
		return errors.Errorf("tls.certificate does not exist or empty for peer %s", peerName)
	}
	return nil
}

// IsTLSEnabled is TLS enabled?
func (c *Config) IsTLSEnabled() bool {
	return myViper.GetBool("client.tls.enabled")
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
	return myViper.GetBool("client.BCCSP.security.enabled")
}

// SecurityAlgorithm ...
func (c *Config) SecurityAlgorithm() string {
	return myViper.GetString("client.BCCSP.security.hashAlgorithm")
}

// SecurityLevel ...
func (c *Config) SecurityLevel() int {
	return myViper.GetInt("client.BCCSP.security.level")
}

//SecurityProvider provider SW or PKCS11
func (c *Config) SecurityProvider() string {
	return myViper.GetString("client.BCCSP.security.default.provider")
}

//Ephemeral flag
func (c *Config) Ephemeral() bool {
	return myViper.GetBool("client.BCCSP.security.ephemeral")
}

//SoftVerify flag
func (c *Config) SoftVerify() bool {
	return myViper.GetBool("client.BCCSP.security.softVerify")
}

//SecurityProviderLibPath will be set only if provider is PKCS11
func (c *Config) SecurityProviderLibPath() string {
	configuredLibs := myViper.GetString("client.BCCSP.security.library")
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
	return myViper.GetString("client.BCCSP.security.pin")
}

//SecurityProviderLabel will be set only if provider is PKCS11
func (c *Config) SecurityProviderLabel() string {
	return myViper.GetString("client.BCCSP.security.label")
}

// KeyStorePath returns the keystore path used by BCCSP
func (c *Config) KeyStorePath() string {
	keystorePath := strings.Replace(myViper.GetString("client.credentialStore.cryptoStore.path"),
		"$GOPATH", os.Getenv("GOPATH"), -1)
	return path.Join(keystorePath, "keystore")
}

// CAKeyStorePath returns the same path as KeyStorePath() without the
// 'keystore' directory added. This is done because the fabric-ca-client
// adds this to the path
func (c *Config) CAKeyStorePath() string {
	return strings.Replace(myViper.GetString("client.credentialStore.cryptoStore.path"),
		"$GOPATH", os.Getenv("GOPATH"), -1)
}

// CryptoConfigPath ...
func (c *Config) CryptoConfigPath() string {
	return strings.Replace(myViper.GetString("client.cryptoconfig.path"),
		"$GOPATH", os.Getenv("GOPATH"), -1)
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

// CSPConfig ...
func (c *Config) CSPConfig() *bccspFactory.FactoryOpts {
	switch c.SecurityProvider() {
	case "SW":
		opts := &bccspFactory.FactoryOpts{
			ProviderName: "SW",
			SwOpts: &bccspFactory.SwOpts{
				HashFamily: c.SecurityAlgorithm(),
				SecLevel:   c.SecurityLevel(),
				FileKeystore: &bccspFactory.FileKeystoreOpts{
					KeyStorePath: c.KeyStorePath(),
				},
				Ephemeral: c.Ephemeral(),
			},
		}
		logger.Debug("Initialized SW ")
		bccspFactory.InitFactories(opts)
		return opts

	case "PKCS11":
		pkks := pkcs11.FileKeystoreOpts{KeyStorePath: c.KeyStorePath()}
		opts := &bccspFactory.FactoryOpts{
			ProviderName: "PKCS11",
			Pkcs11Opts: &pkcs11.PKCS11Opts{
				SecLevel:     c.SecurityLevel(),
				HashFamily:   c.SecurityAlgorithm(),
				Ephemeral:    c.Ephemeral(),
				FileKeystore: &pkks,
				Library:      c.SecurityProviderLibPath(),
				Pin:          c.SecurityProviderPin(),
				Label:        c.SecurityProviderLabel(),
				SoftVerify:   c.SoftVerify(),
			},
		}
		logger.Debug("Initialized PKCS11 ")
		bccspFactory.InitFactories(opts)
		return opts

	default:
		panic(fmt.Sprintf("Unsupported BCCSP Provider: %s", c.SecurityProvider()))

	}
}
