/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package msp

import (
	"strconv"
	"strings"

	"github.com/pkg/errors"

	"sort"

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

//ConfigFromBackend returns identity config implementation of given backend
func ConfigFromBackend(coreBackend ...core.ConfigBackend) (msp.IdentityConfig, error) {

	//create identity config
	config := &IdentityConfig{backend: lookup.New(coreBackend...),
		caMatchers: make(map[int]*regexp.Regexp)}

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
	caConfigsByOrg      map[string][]*msp.CAConfig
	backend             *lookup.ConfigLookup
	caKeyStorePath      string
	credentialStorePath string
	entityMatchers      *entityMatchers
	caMatchers          map[int]*regexp.Regexp
}

//entityMatchers for identity configuration
type entityMatchers struct {
	matchers map[string][]MatchConfig
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
	Client endpoint.TLSKeyPair
}

// CAConfig defines a CA configuration in identity config
type CAConfig struct {
	URL        string
	TLSCACerts endpoint.MutualTLSConfig
	Registrar  msp.EnrollCredentials
	CAName     string
}

// MatchConfig contains match pattern and substitution pattern
// for pattern matching of network configured hostnames or channel names with static config
type MatchConfig struct {
	Pattern string

	// these are used for hostname mapping
	URLSubstitutionExp string
	MappedHost         string

	// this is used for Name mapping instead of hostname mappings
	MappedName string
}

// Client returns the Client config
func (c *IdentityConfig) Client() *msp.ClientConfig {
	return c.client
}

// CAConfig returns the CA configuration.
func (c *IdentityConfig) CAConfig(org string) (*msp.CAConfig, bool) {
	caConfigs, ok := c.caConfigsByOrg[strings.ToLower(org)]
	if ok {
		//for now, we're only loading the first Cert Authority by default.
		return caConfigs[0], true
	}
	return nil, false
}

//CAClientCert read configuration for the fabric CA client cert bytes for given org
func (c *IdentityConfig) CAClientCert(org string) ([]byte, bool) {
	caConfigs, ok := c.caConfigsByOrg[strings.ToLower(org)]
	if ok {
		//for now, we're only loading the first Cert Authority by default.
		return caConfigs[0].TLSCAClientCert, true
	}

	return nil, false
}

//CAClientKey read configuration for the fabric CA client key bytes for given org
func (c *IdentityConfig) CAClientKey(org string) ([]byte, bool) {
	caConfigs, ok := c.caConfigsByOrg[strings.ToLower(org)]
	if ok {
		//for now, we're only loading the first Cert Authority by default.
		return caConfigs[0].TLSCAClientKey, true
	}

	return nil, false
}

// CAServerCerts Read configuration option for the server certificates
// will send a list of cert bytes for given org
func (c *IdentityConfig) CAServerCerts(org string) ([][]byte, bool) {
	caConfigs, ok := c.caConfigsByOrg[strings.ToLower(org)]
	if ok {
		//for now, we're only loading the first Cert Authority by default.
		return caConfigs[0].TLSCAServerCerts, true
	}
	return nil, false
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

	c.caKeyStorePath = pathvar.Subst(c.backend.GetString("client.credentialStore.cryptoStore.path"))
	c.credentialStorePath = pathvar.Subst(c.backend.GetString("client.credentialStore.path"))

	return nil
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

	caConfigsByOrg := make(map[string][]*msp.CAConfig)

	for orgName, orgConfig := range configEntity.Organizations {
		var caConfigs []*msp.CAConfig
		for _, caName := range orgConfig.CertificateAuthorities {
			if caName == "" {
				continue
			}
			caConfig, ok := configEntity.CertificateAuthorities[strings.ToLower(caName)]
			if !ok {
				logger.Debugf("Could not find Certificate Authority for [%s], trying with Entity Matchers", caName)
				matchedCaConfig, mappedHost := c.tryMatchingCAConfig(configEntity, strings.ToLower(caName))
				if mappedHost == "" {
					return errors.Errorf("CA Server Name [%s] not found", caName)
				}
				caConfig = *matchedCaConfig
				logger.Debugf("Mapped Certificate Authority for [%s] to [%s]", caName, mappedHost)
			}
			mspCAConfig, err := c.getMSPCAConfig(&caConfig)
			if err != nil {
				return err
			}
			caConfigs = append(caConfigs, mspCAConfig)
		}
		caConfigsByOrg[strings.ToLower(orgName)] = caConfigs
	}

	c.caConfigsByOrg = caConfigsByOrg
	return nil
}

func (c *IdentityConfig) getMSPCAConfig(caConfig *CAConfig) (*msp.CAConfig, error) {

	serverCerts, err := c.getServerCerts(caConfig)
	if err != nil {
		return nil, err
	}

	return &msp.CAConfig{
		URL:              caConfig.URL,
		Registrar:        caConfig.Registrar,
		CAName:           caConfig.CAName,
		TLSCAClientCert:  caConfig.TLSCACerts.Client.Cert.Bytes(),
		TLSCAClientKey:   caConfig.TLSCACerts.Client.Key.Bytes(),
		TLSCAServerCerts: serverCerts,
	}, nil
}

func (c *IdentityConfig) getServerCerts(caConfig *CAConfig) ([][]byte, error) {

	var serverCerts [][]byte

	//check for pems first
	pems := caConfig.TLSCACerts.Pem
	if len(pems) > 0 {
		serverCerts = make([][]byte, len(pems))
		for i, pem := range pems {
			serverCerts[i] = []byte(pem)
		}
		return serverCerts, nil
	}

	//check for files if pems not found
	certFiles := strings.Split(caConfig.TLSCACerts.Path, ",")
	serverCerts = make([][]byte, len(certFiles))
	for i, certPath := range certFiles {
		bytes, err := ioutil.ReadFile(pathvar.Subst(certPath))
		if err != nil {
			return nil, errors.WithMessage(err, "failed to load server certs")
		}
		serverCerts[i] = bytes
	}

	return serverCerts, nil
}

func (c *IdentityConfig) compileMatchers() error {
	entityMatchers := entityMatchers{}

	err := c.backend.UnmarshalKey("entityMatchers", &entityMatchers.matchers)
	logger.Debugf("Matchers are: %+v", entityMatchers)
	if err != nil {
		return errors.WithMessage(err, "failed to parse 'entityMatchers' config item")
	}

	if entityMatchers.matchers["certificateauthority"] != nil {
		certMatchersConfig := entityMatchers.matchers["certificateauthority"]
		var err error
		for i := 0; i < len(certMatchersConfig); i++ {
			if certMatchersConfig[i].Pattern != "" {
				c.caMatchers[i], err = regexp.Compile(certMatchersConfig[i].Pattern)
				if err != nil {
					return err
				}
			}
		}
	}
	c.entityMatchers = &entityMatchers
	return nil
}

func (c *IdentityConfig) tryMatchingCAConfig(configEntity *identityConfigEntity, caName string) (*CAConfig, string) {
	//Return if no caMatchers are configured
	if len(c.caMatchers) == 0 {
		return nil, ""
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
			return c.findMatchingCert(configEntity, caName, v, k)
		}
	}

	return nil, ""
}

func (c *IdentityConfig) findMatchingCert(configEntity *identityConfigEntity, caName string, v *regexp.Regexp, k int) (*CAConfig, string) {
	// get the matching Config from the index number
	certAuthorityMatchConfig := c.entityMatchers.matchers["certificateauthority"][k]
	//Get the certAuthorityMatchConfig from mapped host
	caConfig, ok := configEntity.CertificateAuthorities[strings.ToLower(certAuthorityMatchConfig.MappedHost)]
	if !ok {
		return nil, ""
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
		if !strings.Contains(certAuthorityMatchConfig.URLSubstitutionExp, "$") {
			caConfig.URL = certAuthorityMatchConfig.URLSubstitutionExp
		} else {
			//if the urlSubstitutionExp has $ variable declarations, use regex replaceallstring to replace networkhostname with substituionexp pattern
			caConfig.URL = v.ReplaceAllString(caName, certAuthorityMatchConfig.URLSubstitutionExp)
		}
	}

	return &caConfig, certAuthorityMatchConfig.MappedHost
}

func (c *IdentityConfig) getPortIfPresent(url string) (int, bool) {
	s := strings.Split(url, ":")
	if len(s) > 1 {
		if port, err := strconv.Atoi(s[len(s)-1]); err == nil {
			return port, true
		}
	}
	return 0, false
}
