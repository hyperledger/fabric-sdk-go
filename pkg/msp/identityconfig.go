/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package msp

import (
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"strconv"
	"strings"

	commtls "github.com/hyperledger/fabric-sdk-go/pkg/core/config/comm/tls"

	"github.com/pkg/errors"

	"regexp"

	"io/ioutil"

	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/core"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/msp"
	"github.com/hyperledger/fabric-sdk-go/pkg/core/config/endpoint"
	"github.com/hyperledger/fabric-sdk-go/pkg/core/config/lookup"
	logApi "github.com/hyperledger/fabric-sdk-go/pkg/core/logging/api"
	fabImpl "github.com/hyperledger/fabric-sdk-go/pkg/fab"
	"github.com/hyperledger/fabric-sdk-go/pkg/util/pathvar"
)

var defaultCAServerSchema = "https"
var defaultCAServerListenPort = 7054

//ConfigFromBackend returns identity config implementation of given backend
func ConfigFromBackend(coreBackend ...core.ConfigBackend) (msp.IdentityConfig, error) {

	//create identity config
	config := &IdentityConfig{backend: lookup.New(coreBackend...)}

	//preload config identities
	err := config.loadIdentityConfigEntities()
	if err != nil {
		return nil, errors.WithMessage(err, "failed to create identity config from backends")
	}

	return config, nil
}

// IdentityConfig represents the identity configuration for the client
type IdentityConfig struct {
	client              *msp.ClientConfig
	caConfigs           map[string]*msp.CAConfig
	backend             *lookup.ConfigLookup
	caKeyStorePath      string
	credentialStorePath string
	caMatchers          []matcherEntry
	tlsCertPool         commtls.CertPool
}

//entityMatchers for identity configuration
type entityMatchers struct {
	matchers map[string][]MatchConfig
}

//matcher entry mapping regex to match config
type matcherEntry struct {
	regex       *regexp.Regexp
	matchConfig MatchConfig
}

//identityConfigEntity contains all config definitions needed
type identityConfigEntity struct {
	Client                 ClientConfig
	Organizations          map[string]fabImpl.OrganizationConfig
	CertificateAuthorities map[string]CAConfig
}

// ClientConfig defines client configuration in identity config
type ClientConfig struct {
	Organization    string
	Logging         logApi.LoggingType
	CryptoConfig    msp.CCType
	TLSCerts        ClientTLSConfig
	CredentialStore msp.CredentialStoreType
}

//ClientTLSConfig defines client TLS configuration in identity config
type ClientTLSConfig struct {
	//Client TLS information
	Client         endpoint.TLSKeyPair
	SystemCertPool bool
}

// CAConfig defines a CA configuration in identity config
type CAConfig struct {
	ID          string
	URL         string
	GRPCOptions map[string]interface{}
	TLSCACerts  endpoint.MutualTLSConfig
	Registrar   msp.EnrollCredentials
	CAName      string
}

// MatchConfig contains match pattern and substitution pattern
// for pattern matching of network configured hostnames or channel names with static config
type MatchConfig struct {
	Pattern string

	// these are used for hostname mapping
	URLSubstitutionExp                  string
	SSLTargetOverrideURLSubstitutionExp string
	MappedHost                          string

	// this is used for Name mapping instead of hostname mappings
	MappedName string

	//IgnoreEndpoint option to exclude given entity from any kind of search or from entity list
	IgnoreEndpoint bool
}

// Client returns the Client config
func (c *IdentityConfig) Client() *msp.ClientConfig {
	return c.client
}

// CAConfig returns the CA configuration.
func (c *IdentityConfig) CAConfig(caID string) (*msp.CAConfig, bool) {
	cfg, ok := c.caConfigs[strings.ToLower(caID)]
	return cfg, ok
}

//CAClientCert read configuration for the fabric CA client cert bytes for given org
func (c *IdentityConfig) CAClientCert(caID string) ([]byte, bool) {
	cfg, ok := c.caConfigs[strings.ToLower(caID)]
	if ok {
		//for now, we're only loading the first Cert Authority by default.
		return cfg.TLSCAClientCert, true
	}
	return nil, false
}

//CAClientKey read configuration for the fabric CA client key bytes for given org
func (c *IdentityConfig) CAClientKey(caID string) ([]byte, bool) {
	cfg, ok := c.caConfigs[strings.ToLower(caID)]
	if ok {
		//for now, we're only loading the first Cert Authority by default.
		return cfg.TLSCAClientKey, true
	}
	return nil, false
}

// CAServerCerts Read configuration option for the server certificates
// will send a list of cert bytes for given org
func (c *IdentityConfig) CAServerCerts(caID string) ([][]byte, bool) {
	cfg, ok := c.caConfigs[strings.ToLower(caID)]
	if ok {
		//for now, we're only loading the first Cert Authority by default.
		return cfg.TLSCAServerCerts, true
	}
	return nil, false
}

// TLSCACertPool returns the configured cert pool.
func (c *IdentityConfig) TLSCACertPool() commtls.CertPool {
	return c.tlsCertPool
}

// CAKeyStorePath returns the same path as KeyStorePath() without the
// 'keystore' directory added. This is done because the fabric-ca-client
// adds this to the path
func (c *IdentityConfig) CAKeyStorePath() string {
	return c.caKeyStorePath
}

// CredentialStorePath returns the user store path
func (c *IdentityConfig) CredentialStorePath() string {
	return c.credentialStorePath
}

//loadIdentityConfigEntities loads config entities and dictionaries for searches
func (c *IdentityConfig) loadIdentityConfigEntities() error {
	configEntity := identityConfigEntity{}

	err := c.backend.UnmarshalKey("client", &configEntity.Client)
	logger.Debugf("Client is: %+v", configEntity.Client)
	if err != nil {
		return errors.WithMessage(err, "failed to parse 'client' config item to identityConfigEntity.Client type")
	}

	err = c.backend.UnmarshalKey("organizations", &configEntity.Organizations)
	logger.Debugf("organizations are: %+v", configEntity.Organizations)
	if err != nil {
		return errors.WithMessage(err, "failed to parse 'organizations' config item to identityConfigEntity.Organizations type")
	}

	err = c.backend.UnmarshalKey("certificateAuthorities", &configEntity.CertificateAuthorities)
	logger.Debugf("certificateAuthorities are: %+v", configEntity.CertificateAuthorities)
	if err != nil {
		return errors.WithMessage(err, "failed to parse 'certificateAuthorities' config item to identityConfigEntity.CertificateAuthorities type")
	}
	// Populate ID from the lookup keys
	for caID := range configEntity.CertificateAuthorities {
		ca := configEntity.CertificateAuthorities[caID]
		ca.ID = caID
		configEntity.CertificateAuthorities[caID] = ca
	}

	//compile CA matchers
	err = c.compileMatchers()
	if err != nil {
		return errors.WithMessage(err, "failed to compile certificate authority matchers")
	}

	err = c.loadClientTLSConfig(&configEntity)
	if err != nil {
		return errors.WithMessage(err, "failed to load client TLSConfig ")
	}

	err = c.loadCATLSConfig(&configEntity)
	if err != nil {
		return errors.WithMessage(err, "failed to load CA TLSConfig ")
	}

	err = c.loadAllCAConfigs(&configEntity)
	if err != nil {
		return errors.WithMessage(err, "failed to load all CA configs ")
	}

	err = c.loadTLSCertPool(&configEntity)
	if err != nil {
		return errors.WithMessage(err, "failed to load TLS Cert Pool")
	}

	c.caKeyStorePath = pathvar.Subst(c.backend.GetString("client.credentialStore.cryptoStore.path"))
	c.credentialStorePath = pathvar.Subst(c.backend.GetString("client.credentialStore.path"))

	return nil
}

func (c *IdentityConfig) loadTLSCertPool(ce *identityConfigEntity) error {

	useSystemCertPool := ce.Client.TLSCerts.SystemCertPool

	var err error
	c.tlsCertPool, err = commtls.NewCertPool(useSystemCertPool)
	if err != nil {
		return errors.WithMessage(err, "failed to create cert pool")
	}

	// preemptively add all TLS certs to cert pool as adding them at request time
	// is expensive
	for _, ca := range c.caConfigs {
		if len(ca.TLSCAServerCerts) == 0 && !useSystemCertPool {
			return errors.New(fmt.Sprintf("Org '%s' doesn't have defined tlsCACerts", ca.ID))
		}
		for _, cacert := range ca.TLSCAServerCerts {
			ok := appendCertsFromPEM(c.tlsCertPool, cacert)
			if !ok {
				return errors.New("Failed to process certificate")
			}
		}
	}

	//update cert pool
	if _, err := c.tlsCertPool.Get(); err != nil {
		return errors.WithMessage(err, "cert pool load failed")
	}
	return nil
}

// see x509.AppendCertsFromPEM
func appendCertsFromPEM(c commtls.CertPool, pemCerts []byte) (ok bool) {
	for len(pemCerts) > 0 {
		var block *pem.Block
		block, pemCerts = pem.Decode(pemCerts)
		if block == nil {
			break
		}
		if block.Type != "CERTIFICATE" || len(block.Headers) != 0 {
			continue
		}

		cert, err := x509.ParseCertificate(block.Bytes)
		if err != nil {
			continue
		}

		c.Add(cert)
		ok = true
	}

	return
}

//loadClientTLSConfig pre-loads all TLSConfig bytes in client config
func (c *IdentityConfig) loadClientTLSConfig(configEntity *identityConfigEntity) error {
	//Clients Config
	//resolve paths and org name
	configEntity.Client.Organization = strings.ToLower(configEntity.Client.Organization)
	configEntity.Client.TLSCerts.Client.Key.Path = pathvar.Subst(configEntity.Client.TLSCerts.Client.Key.Path)
	configEntity.Client.TLSCerts.Client.Cert.Path = pathvar.Subst(configEntity.Client.TLSCerts.Client.Cert.Path)

	//pre load client key and cert bytes
	err := configEntity.Client.TLSCerts.Client.Key.LoadBytes()
	if err != nil {
		return errors.WithMessage(err, "failed to load client key")
	}

	err = configEntity.Client.TLSCerts.Client.Cert.LoadBytes()
	if err != nil {
		return errors.WithMessage(err, "failed to load client cert")
	}

	c.client = &msp.ClientConfig{
		Organization:    configEntity.Client.Organization,
		Logging:         configEntity.Client.Logging,
		CryptoConfig:    configEntity.Client.CryptoConfig,
		CredentialStore: configEntity.Client.CredentialStore,
		TLSKey:          configEntity.Client.TLSCerts.Client.Key.Bytes(),
		TLSCert:         configEntity.Client.TLSCerts.Client.Cert.Bytes(),
	}

	return nil
}

//loadCATLSConfig pre-loads all TLSConfig bytes in certificate authorities
func (c *IdentityConfig) loadCATLSConfig(configEntity *identityConfigEntity) error {
	//CA Config
	for ca, caConfig := range configEntity.CertificateAuthorities {
		//resolve paths
		caConfig.TLSCACerts.Path = pathvar.Subst(caConfig.TLSCACerts.Path)
		caConfig.TLSCACerts.Client.Key.Path = pathvar.Subst(caConfig.TLSCACerts.Client.Key.Path)
		caConfig.TLSCACerts.Client.Cert.Path = pathvar.Subst(caConfig.TLSCACerts.Client.Cert.Path)
		//pre load key and cert bytes
		err := caConfig.TLSCACerts.Client.Key.LoadBytes()
		if err != nil {
			return errors.WithMessage(err, "failed to load ca key")
		}

		err = caConfig.TLSCACerts.Client.Cert.LoadBytes()
		if err != nil {
			return errors.WithMessage(err, "failed to load ca cert")
		}
		configEntity.CertificateAuthorities[ca] = caConfig
	}

	return nil
}

func (c *IdentityConfig) loadAllCAConfigs(configEntity *identityConfigEntity) error {

	configs := make(map[string]*msp.CAConfig)

	for caID := range configEntity.CertificateAuthorities {

		matchedCaConfig, ok := c.tryMatchingCAConfig(configEntity, strings.ToLower(caID))
		if !ok {
			continue
		}

		logger.Debugf("Mapped Certificate Authority [%s]", caID)
		mspCAConfig, err := c.getMSPCAConfig(matchedCaConfig)
		if err != nil {
			return err
		}
		configs[strings.ToLower(caID)] = mspCAConfig
	}

	c.caConfigs = configs
	return nil
}

func (c *IdentityConfig) getMSPCAConfig(caConfig *CAConfig) (*msp.CAConfig, error) {

	serverCerts, err := c.getServerCerts(caConfig)
	if err != nil {
		return nil, err
	}

	var URL string
	if caConfig.URL == "" {
		URL = defaultCAServerSchema + "://" + caConfig.ID + ":" + strconv.Itoa(defaultCAServerListenPort)
	} else {
		URL = caConfig.URL
	}

	return &msp.CAConfig{
		ID:               caConfig.ID,
		URL:              URL,
		GRPCOptions:      caConfig.GRPCOptions,
		Registrar:        caConfig.Registrar,
		CAName:           caConfig.CAName,
		TLSCAClientCert:  caConfig.TLSCACerts.Client.Cert.Bytes(),
		TLSCAClientKey:   caConfig.TLSCACerts.Client.Key.Bytes(),
		TLSCAServerCerts: serverCerts,
	}, nil
}

func (c *IdentityConfig) getServerCerts(caConfig *CAConfig) ([][]byte, error) {

	var serverCerts [][]byte

	// check for pems first
	pems := caConfig.TLSCACerts.Pem
	if len(pems) > 0 {
		serverCerts = make([][]byte, len(pems))
		for i, p := range pems {
			serverCerts[i] = []byte(p)
		}
		return serverCerts, nil
	}

	// check for files if pems not found
	if caConfig.TLSCACerts.Path != "" {
		certFiles := strings.Split(caConfig.TLSCACerts.Path, ",")
		serverCerts = make([][]byte, len(certFiles))
		for i, certPath := range certFiles {
			bytes, err := ioutil.ReadFile(pathvar.Subst(certPath))
			if err != nil {
				return nil, errors.WithMessage(err, "failed to load server certs")
			}
			serverCerts[i] = bytes
		}
	}

	// Can return nil. It's OK if SystemCertPool is true
	return serverCerts, nil
}

func (c *IdentityConfig) compileMatchers() error {
	entMatchers := entityMatchers{}

	err := c.backend.UnmarshalKey("entityMatchers", &entMatchers.matchers)
	logger.Debugf("Matchers are: %+v", entMatchers)
	if err != nil {
		return errors.WithMessage(err, "failed to parse 'entMatchers' config item")
	}

	caMatcherConfigs := entMatchers.matchers["certificateauthority"]
	c.caMatchers = make([]matcherEntry, len(caMatcherConfigs))

	if len(caMatcherConfigs) > 0 {
		for i, v := range caMatcherConfigs {
			regex, err := regexp.Compile(v.Pattern)
			if err != nil {
				return err
			}
			c.caMatchers[i] = matcherEntry{regex: regex, matchConfig: v}
		}
	}

	return nil
}

func (c *IdentityConfig) tryMatchingCAConfig(configEntity *identityConfigEntity, caID string) (*CAConfig, bool) {

	//loop over certAuthorityEntityMatchers to find the matching CA Config
	for _, matcher := range c.caMatchers {
		if matcher.regex.MatchString(caID) {
			return c.findMatchingCAConfig(configEntity, caID, matcher)
		}
	}

	//Direct lookup, if no caMatchers are configured or no matcher matched
	caConfig, ok := configEntity.CertificateAuthorities[strings.ToLower(caID)]
	if !ok {
		return nil, false
	}

	if caConfig.GRPCOptions == nil {
		caConfig.GRPCOptions = make(map[string]interface{})
	}

	return &caConfig, true
}

func (c *IdentityConfig) findMatchingCAConfig(configEntity *identityConfigEntity, caID string, matcher matcherEntry) (*CAConfig, bool) {

	if matcher.matchConfig.IgnoreEndpoint {
		logger.Debugf("Ignoring CA `%s` since entity matcher 'IgnoreEndpoint' flag is on", caID)
		return nil, false
	}

	mappedHost := matcher.matchConfig.MappedHost
	if strings.Contains(mappedHost, "$") {
		mappedHost = matcher.regex.ReplaceAllString(caID, mappedHost)
	}

	//Get the certAuthorityMatchConfig from mapped host
	caConfig, ok := configEntity.CertificateAuthorities[strings.ToLower(mappedHost)]
	if !ok {
		return nil, false
	}

	if matcher.matchConfig.URLSubstitutionExp != "" {
		caConfig.URL = matcher.matchConfig.URLSubstitutionExp
		//check for regex replace '$'
		if strings.Contains(caConfig.URL, "$") {
			caConfig.URL = matcher.regex.ReplaceAllString(caID, caConfig.URL)
		}
	}

	if caConfig.GRPCOptions == nil {
		caConfig.GRPCOptions = make(map[string]interface{})
	}

	//SSLTargetOverrideURLSubstitutionExp if found use from entity matcher otherwise use from mapped host
	if matcher.matchConfig.SSLTargetOverrideURLSubstitutionExp != "" {
		hostOverride := matcher.matchConfig.SSLTargetOverrideURLSubstitutionExp
		//check for regex replace '$'
		if strings.Contains(hostOverride, "$") {
			hostOverride = matcher.regex.ReplaceAllString(caID, hostOverride)
		}
		caConfig.GRPCOptions["ssl-target-name-override"] = hostOverride
	}

	return &caConfig, true
}
