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
	"github.com/hyperledger/fabric-ca/api"
	"github.com/hyperledger/fabric-ca/lib/ldap"
	"github.com/hyperledger/fabric-ca/lib/tls"
	"github.com/hyperledger/fabric-ca/util"
	"github.com/hyperledger/fabric/bccsp/factory"
)

const (
	// defaultCACfgTemplate is the a CA's default configuration file template
	defaultCACfgTemplate = `
#############################################################################
# This file contains information specific to a single Certificate Authority (CA).
# A single fabric-ca-server can service multiple CAs.  The server's configuration
# file contains configuration information for the default CA, and each of these
# CA-specific files define configuration settings for a non-default CA.
#
# The only required configuration item in each CA-specific file is a unique
# CA name (see "ca.name" below).  Each CA name in the same fabric-ca-server
# must be unique. All other configuration settings needed for this CA are
# taken from the default CA settings, or you may override those settings by
# adding the setting to this file.
#
# For example, you should provide a different username and password for the
# bootstrap identity as found in the "identities" subsection of the "registry"
# section.
#
# See the server's configuration file for comments on all settings.
# All settings pertaining to the server's listening endpoint are by definition
# server-specific and so will be ignored in a CA configuration file.
#############################################################################
ca:
  # Name of this CA
  name: <<<CANAME>>>

###########################################################################
#  Certificate Signing Request section for generating the CA certificate
###########################################################################
csr:
  cn: <<<COMMONNAME>>>
`
)

// CAConfig is the CA instance's config
// The tags are recognized by the RegisterFlags function in fabric-ca/lib/util.go
// and are as follows:
// "def" - the default value of the field;
// "opt" - the optional one character short name to use on the command line;
// "help" - the help message to display on the command line;
// "skip" - to skip the field.
type CAConfig struct {
	CA           CAInfo
	Signing      *config.Signing
	CSR          api.CSRInfo
	Registry     CAConfigRegistry
	Affiliations map[string]interface{}
	LDAP         ldap.Config
	DB           CAConfigDB
	CSP          *factory.FactoryOpts `mapstructure:"bccsp"`
	// Optional client config for an intermediate server which acts as a client
	// of the root (or parent) server
	Client       *ClientConfig
	Intermediate IntermediateCA
}

// CAInfo is the CA information on a fabric-ca-server
type CAInfo struct {
	Name      string `opt:"n" help:"Certificate Authority name"`
	Keyfile   string `def:"ca-key.pem" help:"PEM-encoded CA key file"`
	Certfile  string `def:"ca-cert.pem" help:"PEM-encoded CA certificate file"`
	Chainfile string `def:"ca-chain.pem" help:"PEM-encoded CA chain file"`
}

// CAConfigDB is the database part of the server's config
type CAConfigDB struct {
	Type       string `def:"sqlite3" help:"Type of database; one of: sqlite3, postgres, mysql"`
	Datasource string `def:"fabric-ca-server.db" help:"Data source which is database specific"`
	TLS        tls.ClientTLSConfig
}

// CAConfigRegistry is the registry part of the server's config
type CAConfigRegistry struct {
	MaxEnrollments int `def:"-1" help:"Maximum number of enrollments; valid if LDAP not enabled"`
	Identities     []CAConfigIdentity
}

// CAConfigIdentity is identity information in the server's config
type CAConfigIdentity struct {
	Name           string
	Pass           string `secret:"password"`
	Type           string
	Affiliation    string
	MaxEnrollments int
	Attrs          map[string]string
}

// ParentServer contains URL for the parent server and the name of CA inside
// the server to connect to
type ParentServer struct {
	URL    string `opt:"u" help:"URL of the parent fabric-ca-server (e.g. http://<username>:<password>@<address>:<port)"`
	CAName string `help:"Name of the CA to connect to on fabric-ca-serve"`
}

// IntermediateCA contains parent server information, TLS configuration, and
// enrollment request for an intermetiate CA
type IntermediateCA struct {
	ParentServer ParentServer
	TLS          tls.ClientTLSConfig
	Enrollment   api.EnrollmentRequest
}

func (cc *CAConfigIdentity) String() string {
	return util.StructToString(cc)
}
