/*
Copyright IBM Corp. 2017 All Rights Reserved.

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

package lib

import (
	"github.com/cloudflare/cfssl/config"
	"github.com/cloudflare/cfssl/csr"
	"github.com/hyperledger/fabric-ca/lib/ldap"
	"github.com/hyperledger/fabric-ca/lib/tls"
	"github.com/hyperledger/fabric-ca/util"
	"github.com/hyperledger/fabric/bccsp/factory"
)

const (
	// DefaultServerPort is the default listening port for the fabric-ca server
	DefaultServerPort = 7054

	// DefaultServerAddr is the default listening address for the fabric-ca server
	DefaultServerAddr = "0.0.0.0"
)

// ServerConfig is the fabric-ca server's config
// The tags are recognized by the RegisterFlags function in fabric-ca/lib/util.go
// and are as follows:
// "def" - the default value of the field;
// "opt" - the optional one character short name to use on the command line;
// "help" - the help message to display on the command line;
// "skip" - to skip the field.
type ServerConfig struct {
	Port         int    `def:"7054" opt:"p" help:"Listening port of fabric-ca-server"`
	Address      string `def:"0.0.0.0" help:"Listening address of fabric-ca-server"`
	Debug        bool   `def:"false" opt:"d" help:"Enable debug level logging"`
	TLS          tls.ServerTLSConfig
	CSP          *factory.FactoryOpts
	CA           ServerConfigCA
	Signing      *config.Signing
	CSR          csr.CertificateRequest
	Registry     ServerConfigRegistry
	Affiliations map[string]interface{}
	LDAP         ldap.Config
	DB           ServerConfigDB
	Remote       string `skip:"true"`
	Client       *ClientConfig
}

// ServerConfigCA is the CA config for the fabric-ca server
type ServerConfigCA struct {
	Name      string `opt:"n" help:"Certificate Authority name"`
	Keyfile   string `def:"ca-key.pem" help:"PEM-encoded CA key file"`
	Certfile  string `def:"ca-cert.pem" help:"PEM-encoded CA certificate file"`
	Chainfile string `def:"ca-chain.pem" help:"PEM-encoded CA chain file"`
}

// ServerConfigDB is the database part of the server's config
type ServerConfigDB struct {
	Type       string `def:"sqlite3" help:"Type of database; one of: sqlite3, postgres, mysql"`
	Datasource string `def:"fabric-ca-server.db" help:"Data source which is database specific"`
	TLS        tls.ClientTLSConfig
}

// ServerConfigRegistry is the registry part of the server's config
type ServerConfigRegistry struct {
	MaxEnrollments int `def:"0" help:"Maximum number of enrollments; valid if LDAP not enabled"`
	Identities     []ServerConfigIdentity
}

// ServerConfigIdentity is identity information in the server's config
type ServerConfigIdentity struct {
	Name           string
	Pass           string `secret:"password"`
	Type           string
	Affiliation    string
	MaxEnrollments int
	Attrs          map[string]string
}

func (sc *ServerConfigIdentity) String() string {
	return util.StructToString(sc)
}
