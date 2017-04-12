/*
Copyright SecureKey Technologies Inc. All Rights Reserved.


Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at


      http://www.apache.org/licenses/LICENSE-2.0


Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package config

import (
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"io/ioutil"
	"os"
	"strconv"
	"strings"

	"github.com/op/go-logging"
	"github.com/spf13/viper"
)

// PeerConfig ...
type PeerConfig struct {
	Host                  string
	Port                  string
	EventHost             string
	EventPort             string
	TLSCertificate        string
	TLSServerHostOverride string
	Primary               bool
}

var myViper = viper.New()
var log = logging.MustGetLogger("fabric_sdk_go")
var format = logging.MustStringFormatter(
	`%{color}%{time:15:04:05.000} [%{module}] %{level:.4s} : %{color:reset} %{message}`,
)

// InitConfig ...
// initConfig reads in config file
func InitConfig(configFile string) error {

	if configFile != "" {
		// create new viper
		myViper.SetConfigFile(configFile)
		// If a config file is found, read it in.
		err := myViper.ReadInConfig()

		if err == nil {
			log.Infof("Using config file: %s", myViper.ConfigFileUsed())
		} else {
			return fmt.Errorf("Fatal error config file: %v", err)
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

	return nil
}

//GetServerURL Read configuration option for the fabric CA server URL
func GetServerURL() string {
	return strings.Replace(myViper.GetString("client.fabricCA.serverURL"), "$GOPATH", os.Getenv("GOPATH"), -1)
}

//GetServerCertFiles Read configuration option for the server certificate files
func GetServerCertFiles() []string {
	certFiles := myViper.GetStringSlice("client.fabricCA.certfiles")
	certFileModPath := make([]string, len(certFiles))
	for i, v := range certFiles {
		certFileModPath[i] = strings.Replace(v, "$GOPATH", os.Getenv("GOPATH"), -1)
	}
	return certFileModPath
}

//GetFabricCAClientKeyFile Read configuration option for the fabric CA client key file
func GetFabricCAClientKeyFile() string {
	return strings.Replace(myViper.GetString("client.fabricCA.client.keyfile"), "$GOPATH", os.Getenv("GOPATH"), -1)
}

//GetFabricCAClientCertFile Read configuration option for the fabric CA client cert file
func GetFabricCAClientCertFile() string {
	return strings.Replace(myViper.GetString("client.fabricCA.client.certfile"), "$GOPATH", os.Getenv("GOPATH"), -1)
}

//GetFabricCATLSEnabledFlag Read configuration option for the fabric CA TLS flag
func GetFabricCATLSEnabledFlag() bool {
	return myViper.GetBool("client.fabricCA.tlsEnabled")
}

// GetFabricClientViper returns the internal viper instance used by the
// SDK to read configuration options
func GetFabricClientViper() *viper.Viper {
	return myViper
}

// GetPeersConfig ...
func GetPeersConfig() []PeerConfig {
	peersConfig := []PeerConfig{}
	peers := myViper.GetStringMap("client.peers")
	for key, value := range peers {
		mm, ok := value.(map[string]interface{})
		var host string
		var port int
		var primary bool
		var eventHost string
		var eventPort int
		var tlsCertificate string
		var tlsServerHostOverride string

		if ok {
			host, _ = mm["host"].(string)
			port, _ = mm["port"].(int)
			primary, _ = mm["primary"].(bool)
			eventHost, _ = mm["event_host"].(string)
			eventPort, _ = mm["event_port"].(int)
			tlsCertificate, _ = mm["tls"].(map[string]interface{})["certificate"].(string)
			tlsServerHostOverride, _ = mm["tls"].(map[string]interface{})["serverhostoverride"].(string)

		} else {
			mm1 := value.(map[interface{}]interface{})
			host, _ = mm1["host"].(string)
			port, _ = mm1["port"].(int)
			primary, _ = mm1["primary"].(bool)
			eventHost, _ = mm1["event_host"].(string)
			eventPort, _ = mm1["event_port"].(int)
			tlsCertificate, _ = mm1["tls"].(map[string]interface{})["certificate"].(string)
			tlsServerHostOverride, _ = mm1["tls"].(map[string]interface{})["serverhostoverride"].(string)

		}

		p := PeerConfig{Host: host, Port: strconv.Itoa(port), EventHost: eventHost, EventPort: strconv.Itoa(eventPort),
			TLSCertificate: tlsCertificate, TLSServerHostOverride: tlsServerHostOverride, Primary: primary}
		if p.Host == "" {
			panic(fmt.Sprintf("host key not exist or empty for %s", key))
		}
		if p.Port == "" {
			panic(fmt.Sprintf("port key not exist or empty for %s", key))
		}

		if IsTLSEnabled() && p.TLSCertificate == "" {
			panic(fmt.Sprintf("tls.certificate not exist or empty for %s", key))
		}

		p.TLSCertificate = strings.Replace(p.TLSCertificate, "$GOPATH", os.Getenv("GOPATH"), -1)
		peersConfig = append(peersConfig, p)
	}
	return peersConfig

}

// IsTLSEnabled ...
func IsTLSEnabled() bool {
	return myViper.GetBool("client.tls.enabled")
}

// GetTLSCACertPool ...
func GetTLSCACertPool(tlsCertificate string) (*x509.CertPool, error) {
	certPool := x509.NewCertPool()
	if tlsCertificate != "" {
		rawData, err := ioutil.ReadFile(tlsCertificate)
		if err != nil {
			return nil, err
		}

		certPool.AddCert(loadCAKey(rawData))
	}

	return certPool, nil
}

// IsSecurityEnabled ...
func IsSecurityEnabled() bool {
	return myViper.GetBool("client.security.enabled")
}

// TcertBatchSize ...
func TcertBatchSize() int {
	return myViper.GetInt("client.tcert.batch.size")
}

// GetSecurityAlgorithm ...
func GetSecurityAlgorithm() string {
	return myViper.GetString("client.security.hashAlgorithm")
}

// GetSecurityLevel ...
func GetSecurityLevel() int {
	return myViper.GetInt("client.security.level")

}

// GetOrdererHost ...
func GetOrdererHost() string {
	return myViper.GetString("client.orderer.host")
}

// GetOrdererPort ...
func GetOrdererPort() string {
	return strconv.Itoa(myViper.GetInt("client.orderer.port"))
}

// GetOrdererTLSServerHostOverride ...
func GetOrdererTLSServerHostOverride() string {
	return myViper.GetString("client.orderer.tls.serverhostoverride")
}

// GetOrdererTLSCertificate ...
func GetOrdererTLSCertificate() string {
	return strings.Replace(myViper.GetString("client.orderer.tls.certificate"), "$GOPATH", os.Getenv("GOPATH"), -1)
}

// GetFabricCAID ...
func GetFabricCAID() string {
	return myViper.GetString("client.fabricCA.id")
}

// GetKeyStorePath ...
func GetKeyStorePath() string {
	return myViper.GetString("client.keystore.path")
}

// loadCAKey
func loadCAKey(rawData []byte) *x509.Certificate {
	block, _ := pem.Decode(rawData)

	pub, err := x509.ParseCertificate(block.Bytes)
	if err != nil {
		panic(err)
	}
	return pub
}
