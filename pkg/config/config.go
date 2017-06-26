/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package config

import (
	"crypto/x509"
	"encoding/pem"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"strconv"
	"strings"

	api "github.com/hyperledger/fabric-sdk-go/api"

	bccspFactory "github.com/hyperledger/fabric/bccsp/factory"
	"github.com/op/go-logging"
	"github.com/spf13/viper"
)

var myViper = viper.New()
var log = logging.MustGetLogger("fabric_sdk_go")
var format = logging.MustStringFormatter(
	`%{color}%{time:15:04:05.000} [%{module}] %{level:.4s} : %{color:reset} %{message}`,
)

const cmdRoot = "fabric_sdk"

type config struct {
}

// InitConfig ...
// initConfig reads in config file
func InitConfig(configFile string) (api.Config, error) {
	return InitConfigWithCmdRoot(configFile, cmdRoot)
}

// InitConfigWithCmdRoot reads in a config file and allows the
// environment variable prefixed to be specified
func InitConfigWithCmdRoot(configFile string, cmdRootPrefix string) (api.Config, error) {
	myViper.SetEnvPrefix(cmdRootPrefix)
	myViper.AutomaticEnv()
	replacer := strings.NewReplacer(".", "_")
	myViper.SetEnvKeyReplacer(replacer)

	if configFile != "" {
		// create new viper
		myViper.SetConfigFile(configFile)
		// If a config file is found, read it in.
		err := myViper.ReadInConfig()

		if err == nil {
			log.Infof("Using config file: %s", myViper.ConfigFileUsed())
		} else {
			return nil, fmt.Errorf("Fatal error config file: %v", err)
		}
	}
	log.Debug(myViper.GetString("client.fabricCA.serverURL"))
	backend := logging.NewLogBackend(os.Stderr, "", 0)
	backendFormatter := logging.NewBackendFormatter(backend, format)

	loggingLevelString := myViper.GetString("client.logging.level")
	logLevel := logging.INFO
	if loggingLevelString != "" {
		log.Infof("fabric_sdk_go Logging level: %v", loggingLevelString)
		var err error
		logLevel, err = logging.LogLevel(loggingLevelString)
		if err != nil {
			panic(err)
		}
	}
	logging.SetBackend(backendFormatter).SetLevel(logging.Level(logLevel), "fabric_sdk_go")

	return &config{}, nil

}

//GetServerURL Read configuration option for the fabric CA server URL
func (c *config) GetServerURL() string {
	return strings.Replace(myViper.GetString("client.fabricCA.serverURL"), "$GOPATH", os.Getenv("GOPATH"), -1)
}

//GetServerCertFiles Read configuration option for the server certificate files
func (c *config) GetServerCertFiles() []string {
	certFiles := myViper.GetStringSlice("client.fabricCA.certfiles")
	certFileModPath := make([]string, len(certFiles))
	for i, v := range certFiles {
		certFileModPath[i] = strings.Replace(v, "$GOPATH", os.Getenv("GOPATH"), -1)
	}
	return certFileModPath
}

//GetFabricCAClientKeyFile Read configuration option for the fabric CA client key file
func (c *config) GetFabricCAClientKeyFile() string {
	return strings.Replace(myViper.GetString("client.fabricCA.client.keyfile"), "$GOPATH", os.Getenv("GOPATH"), -1)
}

//GetFabricCAClientCertFile Read configuration option for the fabric CA client cert file
func (c *config) GetFabricCAClientCertFile() string {
	return strings.Replace(myViper.GetString("client.fabricCA.client.certfile"), "$GOPATH", os.Getenv("GOPATH"), -1)
}

//GetFabricCATLSEnabledFlag Read configuration option for the fabric CA TLS flag
func (c *config) GetFabricCATLSEnabledFlag() bool {
	return myViper.GetBool("client.fabricCA.tlsEnabled")
}

// GetFabricClientViper returns the internal viper instance used by the
// SDK to read configuration options
func (c *config) GetFabricClientViper() *viper.Viper {
	return myViper
}

// GetPeersConfig Retrieves the fabric peers from the config file provided
func (c *config) GetPeersConfig() ([]api.PeerConfig, error) {
	peersConfig := []api.PeerConfig{}
	err := myViper.UnmarshalKey("client.peers", &peersConfig)
	if err != nil {
		return nil, err
	}
	for index, p := range peersConfig {
		if p.Host == "" {
			return nil, fmt.Errorf("host key not exist or empty for peer %d", index)
		}
		if p.Port == 0 {
			return nil, fmt.Errorf("port key not exist or empty for peer %d", index)
		}
		if c.IsTLSEnabled() && p.TLS.Certificate == "" {
			return nil, fmt.Errorf("tls.certificate not exist or empty for peer %d", index)
		}
		peersConfig[index].TLS.Certificate = strings.Replace(p.TLS.Certificate, "$GOPATH",
			os.Getenv("GOPATH"), -1)
	}
	return peersConfig, nil
}

// IsTLSEnabled ...
func (c *config) IsTLSEnabled() bool {
	return myViper.GetBool("client.tls.enabled")
}

// GetTLSCACertPool ...
// TODO: Should be related to configuration.
func (c *config) GetTLSCACertPool(tlsCertificate string) (*x509.CertPool, error) {
	certPool := x509.NewCertPool()
	if tlsCertificate != "" {
		rawData, err := ioutil.ReadFile(tlsCertificate)
		if err != nil {
			return nil, err
		}

		cert, err := loadCAKey(rawData)
		if err != nil {
			return nil, err
		}

		certPool.AddCert(cert)
	}

	return certPool, nil
}

// GetTLSCACertPoolFromRoots ...
func (c *config) GetTLSCACertPoolFromRoots(ordererRootCAs [][]byte) (*x509.CertPool, error) {
	certPool := x509.NewCertPool()

	for _, root := range ordererRootCAs {
		cert, err := loadCAKey(root)
		if err != nil {
			return nil, err
		}

		certPool.AddCert(cert)
	}

	return certPool, nil
}

// IsSecurityEnabled ...
func (c *config) IsSecurityEnabled() bool {
	return myViper.GetBool("client.security.enabled")
}

// TcertBatchSize ...
func (c *config) TcertBatchSize() int {
	return myViper.GetInt("client.tcert.batch.size")
}

// GetSecurityAlgorithm ...
func (c *config) GetSecurityAlgorithm() string {
	return myViper.GetString("client.security.hashAlgorithm")
}

// GetSecurityLevel ...
func (c *config) GetSecurityLevel() int {
	return myViper.GetInt("client.security.level")

}

// GetOrdererHost ...
func (c *config) GetOrdererHost() string {
	return myViper.GetString("client.orderer.host")
}

// GetOrdererPort ...
func (c *config) GetOrdererPort() string {
	return strconv.Itoa(myViper.GetInt("client.orderer.port"))
}

// GetOrdererTLSServerHostOverride ...
func (c *config) GetOrdererTLSServerHostOverride() string {
	return myViper.GetString("client.orderer.tls.serverhostoverride")
}

// GetOrdererTLSCertificate ...
func (c *config) GetOrdererTLSCertificate() string {
	return strings.Replace(myViper.GetString("client.orderer.tls.certificate"), "$GOPATH", os.Getenv("GOPATH"), -1)
}

// GetFabricCAID ...
func (c *config) GetFabricCAID() string {
	return myViper.GetString("client.fabricCA.id")
}

//GetFabricCAName Read the fabric CA name
func (c *config) GetFabricCAName() string {
	return myViper.GetString("client.fabricCA.name")
}

// GetKeyStorePath ...
func (c *config) GetKeyStorePath() string {
	return path.Join(c.GetFabricCAHomeDir(), c.GetFabricCAMspDir(), "keystore")
}

// GetFabricCAHomeDir ...
func (c *config) GetFabricCAHomeDir() string {
	return myViper.GetString("client.fabricCA.homeDir")
}

// GetFabricCAMspDir ...
func (c *config) GetFabricCAMspDir() string {
	return myViper.GetString("client.fabricCA.mspDir")
}

// GetCryptoConfigPath ...
func (c *config) GetCryptoConfigPath() string {
	return strings.Replace(myViper.GetString("client.cryptoconfig.path"), "$GOPATH", os.Getenv("GOPATH"), -1)
}

// loadCAKey
func loadCAKey(rawData []byte) (*x509.Certificate, error) {
	block, _ := pem.Decode(rawData)

	if block != nil {
		pub, err := x509.ParseCertificate(block.Bytes)
		if err != nil {
			return nil, errors.New("Failed to parse certificate: " + err.Error())
		}

		return pub, nil
	}
	return nil, errors.New("No pem data found")
}

// GetCSPConfig ...
func (c *config) GetCSPConfig() *bccspFactory.FactoryOpts {
	return &bccspFactory.FactoryOpts{
		ProviderName: "SW",
		SwOpts: &bccspFactory.SwOpts{
			HashFamily: c.GetSecurityAlgorithm(),
			SecLevel:   c.GetSecurityLevel(),
			FileKeystore: &bccspFactory.FileKeystoreOpts{
				KeyStorePath: c.GetKeyStorePath(),
			},
			Ephemeral: false,
		},
	}
}
