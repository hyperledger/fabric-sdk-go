/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package msp

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/pkg/errors"

	"sort"

	"regexp"

	"io/ioutil"

	"github.com/hyperledger/fabric-sdk-go/pkg/common/errors/status"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/core"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/fab"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/msp"
	fabImpl "github.com/hyperledger/fabric-sdk-go/pkg/fab"
	"github.com/hyperledger/fabric-sdk-go/pkg/util/pathvar"
)

//ConfigFromBackend returns identity config implementation of give backend
func ConfigFromBackend(coreBackend core.ConfigBackend) (msp.IdentityConfig, error) {
	endpointConfig, err := fabImpl.ConfigFromBackend(coreBackend)
	if err != nil {
		return nil, errors.New("failed load identity configuration")
	}
	return &IdentityConfig{endpointConfig.(*fabImpl.EndpointConfig)}, nil
}

// IdentityConfig represents the identity configuration for the client
type IdentityConfig struct {
	endpointConfig *fabImpl.EndpointConfig
}

// Client returns the Client config
func (c *IdentityConfig) Client() (*msp.ClientConfig, error) {
	config, err := c.networkConfig()
	if err != nil {
		return nil, err
	}
	client := config.Client

	client.Organization = strings.ToLower(client.Organization)
	client.TLSCerts.Path = pathvar.Subst(client.TLSCerts.Path)
	client.TLSCerts.Client.Key.Path = pathvar.Subst(client.TLSCerts.Client.Key.Path)
	client.TLSCerts.Client.Cert.Path = pathvar.Subst(client.TLSCerts.Client.Cert.Path)

	return &client, nil
}

// CAConfig returns the CA configuration.
func (c *IdentityConfig) CAConfig(org string) (*msp.CAConfig, error) {
	config, err := c.networkConfig()
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

func (c *IdentityConfig) getCAName(org string) (string, error) {
	config, err := c.networkConfig()
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

//CAClientCert read configuration for the fabric CA client cert bytes for given org
func (c *IdentityConfig) CAClientCert(org string) ([]byte, error) {
	config, err := c.networkConfig()
	if err != nil {
		return nil, err
	}

	caName, err := c.getCAName(org)
	if err != nil {
		return nil, err
	}

	caConfig, ok := config.CertificateAuthorities[strings.ToLower(caName)]
	if !ok {
		return nil, errors.Errorf("CA Server Name %s not found", caName)
	}

	//subst path
	caConfig.TLSCACerts.Client.Cert.Path = pathvar.Subst(caConfig.TLSCACerts.Client.Cert.Path)

	return caConfig.TLSCACerts.Client.Cert.Bytes()
}

//CAClientKey read configuration for the fabric CA client key bytes for given org
func (c *IdentityConfig) CAClientKey(org string) ([]byte, error) {
	config, err := c.networkConfig()
	if err != nil {
		return nil, err
	}

	caName, err := c.getCAName(org)
	if err != nil {
		return nil, err
	}

	caConfig, ok := config.CertificateAuthorities[strings.ToLower(caName)]
	if !ok {
		return nil, errors.Errorf("CA Server Name %s not found", caName)
	}

	//subst path
	caConfig.TLSCACerts.Client.Key.Path = pathvar.Subst(caConfig.TLSCACerts.Client.Key.Path)

	return caConfig.TLSCACerts.Client.Key.Bytes()
}

// CAServerCerts Read configuration option for the server certificates
// will send a list of cert bytes for given org
func (c *IdentityConfig) CAServerCerts(org string) ([][]byte, error) {
	config, err := c.networkConfig()
	if err != nil {
		return nil, err
	}
	caName, err := c.getCAName(org)
	if err != nil {
		return nil, err
	}
	caConfig, ok := config.CertificateAuthorities[strings.ToLower(caName)]
	if !ok {
		return nil, errors.Errorf("CA Server Name '%s' not found", caName)
	}

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
	certFiles := strings.Split(config.CertificateAuthorities[caName].TLSCACerts.Path, ",")
	serverCerts = make([][]byte, len(certFiles))
	for i, certPath := range certFiles {
		bytes, err := ioutil.ReadFile(pathvar.Subst(certPath))
		if err != nil {
			return nil, errors.Wrapf(err, "failed to load pem bytes from path %s", certPath)
		}
		serverCerts[i] = bytes
	}
	return serverCerts, nil
}

// CAKeyStorePath returns the same path as KeyStorePath() without the
// 'keystore' directory added. This is done because the fabric-ca-client
// adds this to the path
func (c *IdentityConfig) CAKeyStorePath() string {
	return pathvar.Subst(c.endpointConfig.Backend().GetString("client.credentialStore.cryptoStore.path"))
}

// CredentialStorePath returns the user store path
func (c *IdentityConfig) CredentialStorePath() string {
	return pathvar.Subst(c.endpointConfig.Backend().GetString("client.credentialStore.path"))
}

// NetworkConfig returns the network configuration defined in the config file
func (c *IdentityConfig) networkConfig() (*fab.NetworkConfig, error) {
	if c.endpointConfig == nil {
		return nil, errors.New("network config not initialized for identity config")
	}
	return c.endpointConfig.NetworkConfig()
}

func (c *IdentityConfig) tryMatchingCAConfig(caName string) (*msp.CAConfig, string, error) {
	//Return if no caMatchers are configured
	caMatchers := c.endpointConfig.CAMatchers()
	if len(caMatchers) == 0 {
		return nil, "", errors.New("no CertAuthority entityMatchers are found")
	}

	//sort the keys
	var keys []int
	for k := range caMatchers {
		keys = append(keys, k)
	}
	sort.Ints(keys)

	//loop over certAuthorityEntityMatchers to find the matching Cert
	for _, k := range keys {
		v := caMatchers[k]
		if v.MatchString(caName) {
			return c.findMatchingCert(caName, v, k)
		}
	}

	return nil, "", errors.WithStack(status.New(status.ClientStatus, status.NoMatchingCertificateAuthorityEntity.ToInt32(), "no matching certAuthority config found", nil))
}

func (c *IdentityConfig) findMatchingCert(caName string, v *regexp.Regexp, k int) (*msp.CAConfig, string, error) {
	networkConfig, err := c.networkConfig()
	if err != nil {
		return nil, "", err
	}
	// get the matching Config from the index number
	certAuthorityMatchConfig := networkConfig.EntityMatchers["certificateauthority"][k]
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
		if !strings.Contains(certAuthorityMatchConfig.URLSubstitutionExp, "$") {
			caConfig.URL = certAuthorityMatchConfig.URLSubstitutionExp
		} else {
			//if the urlSubstitutionExp has $ variable declarations, use regex replaceallstring to replace networkhostname with substituionexp pattern
			caConfig.URL = v.ReplaceAllString(caName, certAuthorityMatchConfig.URLSubstitutionExp)
		}
	}

	return &caConfig, certAuthorityMatchConfig.MappedHost, nil
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
