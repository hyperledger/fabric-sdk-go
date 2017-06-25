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
	"crypto/dsa"
	"crypto/ecdsa"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"strconv"
	"time"

	"github.com/cloudflare/cfssl/config"
	cfcsr "github.com/cloudflare/cfssl/csr"
	"github.com/cloudflare/cfssl/initca"
	"github.com/cloudflare/cfssl/log"
	"github.com/cloudflare/cfssl/signer"
	"github.com/hyperledger/fabric-ca/api"
	"github.com/hyperledger/fabric-ca/lib/dbutil"
	"github.com/hyperledger/fabric-ca/lib/ldap"
	"github.com/hyperledger/fabric-ca/lib/spi"
	"github.com/hyperledger/fabric-ca/lib/tcert"
	"github.com/hyperledger/fabric-ca/lib/tls"
	"github.com/hyperledger/fabric-ca/util"
	"github.com/hyperledger/fabric/bccsp"
	"github.com/jmoiron/sqlx"

	_ "github.com/go-sql-driver/mysql" // import to support MySQL
	_ "github.com/lib/pq"              // import to support Postgres
	_ "github.com/mattn/go-sqlite3"    // import to support SQLite3
)

const (
	defaultDatabaseType = "sqlite3"
)

var (
	// Default root CA certificate expiration is 15 years (in hours).
	defaultRootCACertificateExpiration = "131400h"
	// Default intermediate CA certificate expiration is 5 years (in hours).
	defaultIntermediateCACertificateExpiration = parseDuration("43800h")
	// Default issued certificate expiration is 1 year (in hours).
	defaultIssuedCertificateExpiration = parseDuration("8760h")
)

// CA represents a certificate authority which signs, issues and revokes certificates
type CA struct {
	// The home directory for the CA
	HomeDir string
	// The CA's configuration
	Config *CAConfig
	// The file path of the config file
	ConfigFilePath string
	// The database handle used to store certificates and optionally
	// the user registry information, unless LDAP it enabled for the
	// user registry function.
	db *sqlx.DB
	// The crypto service provider (BCCSP)
	csp bccsp.BCCSP
	// The certificate DB accessor
	certDBAccessor *CertDBAccessor
	// The user registry
	registry spi.UserRegistry
	// The signer used for enrollment
	enrollSigner signer.Signer
	// The options to use in verifying a signature in token-based authentication
	verifyOptions *x509.VerifyOptions
	// The tcert manager for this CA
	tcertMgr *tcert.Mgr
	// The key tree
	keyTree *tcert.KeyTree
	// The server hosting this CA
	server *Server
}

const (
	certificateError = "Invalid certificate in file"
)

// NewCA creates a new CA with the specified
// home directory, parent server URL, and config
func NewCA(caFile string, config *CAConfig, server *Server, renew bool) (*CA, error) {
	ca := new(CA)
	ca.ConfigFilePath = caFile
	err := initCA(ca, filepath.Dir(caFile), config, server, renew)
	if err != nil {
		return nil, err
	}
	return ca, nil
}

// initCA will initialize the passed in pointer to a CA struct
func initCA(ca *CA, homeDir string, config *CAConfig, server *Server, renew bool) error {
	ca.HomeDir = homeDir
	ca.Config = config
	ca.server = server

	err := ca.init(renew)
	if err != nil {
		return err
	}

	return nil
}

// Init initializes an instance of a CA
func (ca *CA) init(renew bool) (err error) {
	log.Debugf("Init CA with home %s and config %+v", ca.HomeDir, *ca.Config)
	// Initialize the config, setting defaults, etc
	err = ca.initConfig()
	if err != nil {
		return err
	}
	// Initialize the crypto layer (BCCSP) for this CA
	ca.csp, err = util.InitBCCSP(&ca.Config.CSP, "", ca.HomeDir)
	if err != nil {
		return err
	}
	// Initialize key materials
	err = ca.initKeyMaterial(renew)
	if err != nil {
		return err
	}
	// Initialize the database
	err = ca.initDB()
	if err != nil {
		return err
	}
	// Initialize the enrollment signer
	err = ca.initEnrollmentSigner()
	if err != nil {
		return err
	}
	// Initialize TCert handling
	keyfile := ca.Config.CA.Keyfile
	certfile := ca.Config.CA.Certfile
	ca.tcertMgr, err = tcert.LoadMgr(keyfile, certfile, ca.csp)
	if err != nil {
		return err
	}
	// FIXME: The root prekey must be stored persistently in DB and retrieved here if not found
	rootKey, err := genRootKey(ca.csp)
	if err != nil {
		return err
	}
	ca.keyTree = tcert.NewKeyTree(ca.csp, rootKey)
	log.Debug("CA initialization successful")
	// Successful initialization
	return nil
}

// Initialize the CA's key material
func (ca *CA) initKeyMaterial(renew bool) error {
	log.Debug("Initialize key material")

	// Make the path names absolute in the config
	err := ca.makeFileNamesAbsolute()
	if err != nil {
		return err
	}

	keyFile := ca.Config.CA.Keyfile
	certFile := ca.Config.CA.Certfile

	// If we aren't renewing and the key and cert files exist, do nothing
	if !renew {
		// If they both exist, the CA was already initialized
		keyFileExists := util.FileExists(keyFile)
		certFileExists := util.FileExists(certFile)
		if keyFileExists && certFileExists {
			log.Info("The CA key and certificate files already exist")
			log.Infof("Key file location: %s", keyFile)
			log.Infof("Certificate file location: %s", certFile)
			err = ca.validateCert(certFile, keyFile)
			if err != nil {
				return fmt.Errorf("Validation of certificate and key failed: %s", err)
			}
			// Load CN from existing enrollment information and set CSR accordingly
			// CN needs to be set, having a multi CA setup requires a unique CN and can't
			// be left blank
			ca.Config.CSR.CN, err = ca.loadCNFromEnrollmentInfo(certFile)
			if err != nil {
				return err
			}
			return nil
		}

		// If key file does not exist but certFile does, key file is probably
		// stored by BCCSP, so check for that now.
		if certFileExists {
			_, _, _, err = util.GetSignerFromCertFile(certFile, ca.csp)
			if err == nil {
				// Yes, it is stored by BCCSP
				log.Info("The CA key and certificate already exist")
				log.Infof("The key is stored by BCCSP provider '%s'", ca.Config.CSP.ProviderName)
				log.Infof("The certificate is at: %s", certFile)
				// Load CN from existing enrollment information and set CSR accordingly
				// CN needs to be set, having a multi CA setup requires a unique CN and can't
				// be left blank
				ca.Config.CSR.CN, err = ca.loadCNFromEnrollmentInfo(certFile)
				if err != nil {
					return err
				}
				return nil
			}
		}
	}

	// Get the CA cert
	cert, err := ca.getCACert()
	if err != nil {
		return err
	}
	// Store the certificate to file
	err = writeFile(certFile, cert, 0644)
	if err != nil {
		return fmt.Errorf("Failed to store certificate: %s", err)
	}
	log.Infof("The CA key and certificate were generated for CA %s", ca.Config.CA.Name)
	log.Infof("The key was stored by BCCSP provider '%s'", ca.Config.CSP.ProviderName)
	log.Infof("The certificate is at: %s", certFile)
	return nil
}

// Get the CA certificate for this CA
func (ca *CA) getCACert() (cert []byte, err error) {
	log.Debugf("Getting CA cert; parent server URL is '%s'", ca.Config.Intermediate.ParentServer.URL)
	if ca.Config.Intermediate.ParentServer.URL != "" {
		// This is an intermediate CA, so call the parent fabric-ca-server
		// to get the cert
		clientCfg := ca.Config.Client
		if clientCfg == nil {
			clientCfg = &ClientConfig{}
		}
		// Copy over the intermediate configuration into client configuration
		clientCfg.TLS = ca.Config.Intermediate.TLS
		clientCfg.Enrollment = ca.Config.Intermediate.Enrollment
		clientCfg.CAName = ca.Config.Intermediate.ParentServer.CAName
		clientCfg.CSR = ca.Config.CSR
		if ca.Config.CSR.CN != "" {
			return nil, fmt.Errorf("CN '%s' cannot be specified for an intermediate CA. Remove CN from CSR section for enrollment of intermediate CA to be successful", ca.Config.CSR.CN)
		}
		if clientCfg.Enrollment.Profile == "" {
			clientCfg.Enrollment.Profile = "ca"
		}
		if clientCfg.Enrollment.CSR == nil {
			clientCfg.Enrollment.CSR = &api.CSRInfo{}
		}
		log.Debugf("Intermediate enrollment request: %v", clientCfg.Enrollment)
		var resp *EnrollmentResponse
		resp, err = clientCfg.Enroll(ca.Config.Intermediate.ParentServer.URL, ca.HomeDir)
		if err != nil {
			return nil, err
		}
		// Set the CN for an intermediate server to be the ID used to enroll with root CA
		ca.Config.CSR.CN = resp.Identity.GetName()
		ecert := resp.Identity.GetECert()
		if ecert == nil {
			return nil, errors.New("No enrollment certificate returned by parent server")
		}
		cert = ecert.Cert()
		// Store the chain file as the concatenation of the parent's chain plus the cert.
		chainPath := ca.Config.CA.Chainfile
		if chainPath == "" {
			chainPath, err = util.MakeFileAbs("ca-chain.pem", ca.HomeDir)
			if err != nil {
				return nil, fmt.Errorf("Failed to create intermediate chain file path: %s", err)
			}
			ca.Config.CA.Chainfile = chainPath
		}
		chain := ca.concatChain(resp.ServerInfo.CAChain, cert)
		err = os.MkdirAll(path.Dir(chainPath), 0755)
		if err != nil {
			return nil, fmt.Errorf("Failed to create intermediate chain file directory: %s", err)
		}
		err = util.WriteFile(chainPath, chain, 0644)
		if err != nil {
			return nil, fmt.Errorf("Failed to create intermediate chain file: %s", err)
		}
		log.Debugf("Stored intermediate certificate chain at %s", chainPath)
	} else {
		// This is a root CA, so create a CSR (Certificate Signing Request)
		if ca.Config.CSR.CN == "" {
			ca.Config.CSR.CN = "fabric-ca-server"
		}
		csr := &ca.Config.CSR
		if csr.CA == nil {
			csr.CA = &cfcsr.CAConfig{}
		}
		if csr.CA.Expiry == "" {
			csr.CA.Expiry = defaultRootCACertificateExpiration
		}
		req := cfcsr.CertificateRequest{
			CN:    csr.CN,
			Names: csr.Names,
			Hosts: csr.Hosts,
			// FIXME: NewBasicKeyRequest only does ecdsa 256; use config
			KeyRequest:   cfcsr.NewBasicKeyRequest(),
			CA:           csr.CA,
			SerialNumber: csr.SerialNumber,
		}
		log.Debugf("Root CA certificate request: %+v", req)
		// Generate the key/signer
		_, cspSigner, err := util.BCCSPKeyRequestGenerate(&req, ca.csp)
		if err != nil {
			return nil, err
		}
		// Call CFSSL to initialize the CA
		cert, _, err = initca.NewFromSigner(&req, cspSigner)
		if err != nil {
			return nil, fmt.Errorf("Failed to create new CA certificate: %s", err)
		}
	}
	return cert, nil
}

// Return a certificate chain which is the concatenation of chain and cert
func (ca *CA) concatChain(chain []byte, cert []byte) []byte {
	result := make([]byte, len(chain)+len(cert))
	copy(result[:len(chain)], chain)
	copy(result[len(chain):], cert)
	return result
}

// Get the certificate chain for the CA
func (ca *CA) getCAChain() (chain []byte, err error) {
	if ca.Config == nil {
		return nil, errors.New("The server has no configuration")
	}
	certAuth := &ca.Config.CA
	// If the chain file exists, we always return the chain from here
	if util.FileExists(certAuth.Chainfile) {
		return util.ReadFile(certAuth.Chainfile)
	}
	// Otherwise, if this is a root CA, we always return the contents of the CACertfile
	if ca.Config.Intermediate.ParentServer.URL == "" {
		return util.ReadFile(certAuth.Certfile)
	}
	// If this is an intermediate CA but the ca.Chainfile doesn't exist,
	// it is an error.  It should have been created during intermediate CA enrollment.
	return nil, fmt.Errorf("Chain file does not exist at %s", certAuth.Chainfile)
}

// Initialize the configuration for the CA setting any defaults and making filenames absolute
func (ca *CA) initConfig() (err error) {
	// Init home directory if not set
	if ca.HomeDir == "" {
		ca.HomeDir, err = os.Getwd()
		if err != nil {
			return fmt.Errorf("Failed to initialize CA's home directory: %s", err)
		}
	}
	log.Debugf("CA Home Directory: %s", ca.HomeDir)
	// Init config if not set
	if ca.Config == nil {
		ca.Config = new(CAConfig)
		ca.Config.Registry.MaxEnrollments = -1
	}
	// Set config defaults
	cfg := ca.Config
	if cfg.CA.Certfile == "" {
		cfg.CA.Certfile = "ca-cert.pem"
	}
	if cfg.CA.Keyfile == "" {
		cfg.CA.Keyfile = "ca-key.pem"
	}
	if cfg.CSR.CA == nil {
		cfg.CSR.CA = &cfcsr.CAConfig{}
	}
	if cfg.CSR.CA.Expiry == "" {
		cfg.CSR.CA.Expiry = defaultRootCACertificateExpiration
	}
	if cfg.Signing == nil {
		cfg.Signing = &config.Signing{}
	}
	cs := cfg.Signing
	if cs.Profiles == nil {
		cs.Profiles = make(map[string]*config.SigningProfile)
	}
	caProfile := cs.Profiles["ca"]
	initSigningProfile(&caProfile,
		defaultIntermediateCACertificateExpiration,
		true)
	cs.Profiles["ca"] = caProfile
	initSigningProfile(
		&cs.Default,
		defaultIssuedCertificateExpiration,
		false)
	// Set log level if debug is true
	if ca.server.Config.Debug {
		log.Level = log.LevelDebug
	}
	ca.normalizeStringSlices()
	return nil
}

// VerifyCertificate verifies that 'cert' was issued by this CA
// Return nil if successful; otherwise, return an error.
func (ca *CA) VerifyCertificate(cert *x509.Certificate) error {
	opts, err := ca.getVerifyOptions()
	if err != nil {
		return fmt.Errorf("Failed to get verify options: %s", err)
	}
	_, err = cert.Verify(*opts)
	if err != nil {
		return fmt.Errorf("Failed to verify certificate: %s", err)
	}
	return nil
}

// Get the options to verify
func (ca *CA) getVerifyOptions() (*x509.VerifyOptions, error) {
	if ca.verifyOptions != nil {
		return ca.verifyOptions, nil
	}
	chain, err := ca.getCAChain()
	if err != nil {
		return nil, err
	}
	block, rest := pem.Decode(chain)
	if block == nil {
		return nil, errors.New("No root certificate was found")
	}
	rootCert, err := x509.ParseCertificate(block.Bytes)
	if err != nil {
		return nil, fmt.Errorf("Failed to parse root certificate: %s", err)
	}
	rootPool := x509.NewCertPool()
	rootPool.AddCert(rootCert)
	var intPool *x509.CertPool
	if len(rest) > 0 {
		intPool = x509.NewCertPool()
		if !intPool.AppendCertsFromPEM(rest) {
			return nil, errors.New("Failed to add intermediate PEM certificates")
		}
	}
	ca.verifyOptions = &x509.VerifyOptions{
		Roots:         rootPool,
		Intermediates: intPool,
		KeyUsages:     []x509.ExtKeyUsage{x509.ExtKeyUsageAny},
	}
	return ca.verifyOptions, nil
}

// Initialize the database for the CA
func (ca *CA) initDB() error {
	db := &ca.Config.DB

	var err error
	var exists bool

	if db.Type == "" || db.Type == defaultDatabaseType {

		db.Type = defaultDatabaseType

		if db.Datasource == "" {
			db.Datasource = "fabric-ca-server.db"
		}

		db.Datasource, err = util.MakeFileAbs(db.Datasource, ca.HomeDir)
		if err != nil {
			return err
		}
	}

	log.Debugf("Initializing '%s' database at '%s'", db.Type, db.Datasource)

	switch db.Type {
	case defaultDatabaseType:
		ca.db, exists, err = dbutil.NewUserRegistrySQLLite3(db.Datasource)
		if err != nil {
			return err
		}
	case "postgres":
		ca.db, exists, err = dbutil.NewUserRegistryPostgres(db.Datasource, &db.TLS)
		if err != nil {
			return err
		}
	case "mysql":
		ca.db, exists, err = dbutil.NewUserRegistryMySQL(db.Datasource, &db.TLS, ca.csp)
		if err != nil {
			return err
		}
	default:
		return fmt.Errorf("Invalid db.type in config file: '%s'; must be 'sqlite3', 'postgres', or 'mysql'", db.Type)
	}

	// Set the certificate DB accessor
	ca.certDBAccessor = NewCertDBAccessor(ca.db)

	// Initialize the user registry.
	// If LDAP is not configured, the fabric-ca CA functions as a user
	// registry based on the database.
	err = ca.initUserRegistry()
	if err != nil {
		return err
	}

	// If the DB doesn't exist, bootstrap it
	if !exists {
		// Since users come from LDAP when enabled,
		// load them from the config file only when LDAP is disabled
		if !ca.Config.LDAP.Enabled {
			err = ca.loadUsersTable()
			if err != nil {
				return err
			}
		}
		err = ca.loadAffiliationsTable()
		if err != nil {
			return err
		}
	}
	log.Infof("Initialized %s database at %s", db.Type, db.Datasource)
	return nil
}

// Initialize the user registry interface
func (ca *CA) initUserRegistry() error {
	log.Debug("Initializing identity registry")
	var err error
	ldapCfg := &ca.Config.LDAP

	if ldapCfg.Enabled {
		// Use LDAP for the user registry
		ca.registry, err = ldap.NewClient(ldapCfg, ca.server.csp)
		log.Debugf("Initialized LDAP identity registry; err=%s", err)
		if err == nil {
			log.Info("Successfully initialized LDAP client")
		} else {
			log.Warningf("Failed to initialize LDAP client; err=%s", err)
		}
		return err
	}

	// Use the DB for the user registry
	dbAccessor := new(Accessor)
	dbAccessor.SetDB(ca.db)
	ca.registry = dbAccessor
	log.Debug("Initialized DB identity registry")
	return nil
}

// Initialize the enrollment signer
func (ca *CA) initEnrollmentSigner() (err error) {
	log.Debug("Initializing enrollment signer")
	c := ca.Config

	// If there is a config, use its signing policy. Otherwise create a default policy.
	var policy *config.Signing
	if c.Signing != nil {
		policy = c.Signing
	} else {
		policy = &config.Signing{
			Profiles: map[string]*config.SigningProfile{},
			Default:  config.DefaultConfig(),
		}
		policy.Default.CAConstraint.IsCA = true
	}

	// Make sure the policy reflects the new remote
	parentServerURL := ca.Config.Intermediate.ParentServer.URL
	if parentServerURL != "" {
		err = policy.OverrideRemotes(parentServerURL)
		if err != nil {
			return fmt.Errorf("Failed initializing enrollment signer: %s", err)
		}
	}

	ca.enrollSigner, err = util.BccspBackedSigner(c.CA.Certfile, c.CA.Keyfile, policy, ca.csp)
	if err != nil {
		return err
	}
	ca.enrollSigner.SetDBAccessor(ca.certDBAccessor)

	// Successful enrollment
	return nil
}

// loadUsersTable adds the configured users to the table if not already found
func (ca *CA) loadUsersTable() error {
	log.Debug("Loading identity table")
	registry := &ca.Config.Registry
	for _, id := range registry.Identities {
		log.Debugf("Loading identity '%s'", id.Name)
		err := ca.addIdentity(&id, false)
		if err != nil {
			return err
		}
	}
	log.Debug("Successfully loaded identity table")
	return nil
}

// loadAffiliationsTable adds the configured affiliations to the table
func (ca *CA) loadAffiliationsTable() error {
	log.Debug("Loading affiliations table")
	err := ca.loadAffiliationsTableR(ca.Config.Affiliations, "")
	if err == nil {
		log.Debug("Successfully loaded affiliations table")
	}
	log.Debug("Successfully loaded groups table")
	return nil
}

// Recursive function to load the affiliations table hierarchy
func (ca *CA) loadAffiliationsTableR(val interface{}, parentPath string) (err error) {
	var path string
	if val == nil {
		return nil
	}
	switch val.(type) {
	case string:
		path = affiliationPath(val.(string), parentPath)
		err = ca.addAffiliation(path, parentPath)
		if err != nil {
			return err
		}
	case []string:
		for _, ele := range val.([]string) {
			err = ca.loadAffiliationsTableR(ele, parentPath)
			if err != nil {
				return err
			}
		}
	case []interface{}:
		for _, ele := range val.([]interface{}) {
			err = ca.loadAffiliationsTableR(ele, parentPath)
			if err != nil {
				return err
			}
		}
	default:
		for name, ele := range val.(map[string]interface{}) {
			path = affiliationPath(name, parentPath)
			err = ca.addAffiliation(path, parentPath)
			if err != nil {
				return err
			}
			err = ca.loadAffiliationsTableR(ele, path)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

// Add an identity to the registry
func (ca *CA) addIdentity(id *CAConfigIdentity, errIfFound bool) error {
	var err error
	user, _ := ca.registry.GetUser(id.Name, nil)
	if user != nil {
		if errIfFound {
			return fmt.Errorf("Identity '%s' is already registered", id.Name)
		}
		log.Debugf("Loaded identity: %+v", id)
		return nil
	}

	id.MaxEnrollments, err = getMaxEnrollments(id.MaxEnrollments, ca.Config.Registry.MaxEnrollments)
	if err != nil {
		return err
	}

	rec := spi.UserInfo{
		Name:           id.Name,
		Pass:           id.Pass,
		Type:           id.Type,
		Affiliation:    id.Affiliation,
		Attributes:     ca.convertAttrs(id.Attrs),
		MaxEnrollments: id.MaxEnrollments,
	}
	err = ca.registry.InsertUser(rec)
	if err != nil {
		return fmt.Errorf("Failed to insert identity '%s': %s", id.Name, err)
	}
	log.Debugf("Registered identity: %+v", id)
	return nil
}

func (ca *CA) addAffiliation(path, parentPath string) error {
	log.Debugf("Adding affiliation %s", path)
	return ca.registry.InsertAffiliation(path, parentPath)
}

// CertDBAccessor returns the certificate DB accessor for CA
func (ca *CA) CertDBAccessor() *CertDBAccessor {
	return ca.certDBAccessor
}

// DBAccessor returns the registry DB accessor for server
func (ca *CA) DBAccessor() spi.UserRegistry {
	return ca.registry
}

func (ca *CA) convertAttrs(inAttrs map[string]string) []api.Attribute {
	var outAttrs []api.Attribute
	for name, value := range inAttrs {
		outAttrs = append(outAttrs, api.Attribute{
			Name:  name,
			Value: value,
		})
	}
	return outAttrs
}

// Get max enrollments relative to the configured max
func (ca *CA) getMaxEnrollments(requestedMax int) (int, error) {
	configuredMax := ca.Config.Registry.MaxEnrollments
	if requestedMax < 0 {
		return configuredMax, nil
	}
	if configuredMax == 0 {
		// no limit, so grant any request
		return requestedMax, nil
	}
	if requestedMax == 0 && configuredMax != 0 {
		return 0, fmt.Errorf("Infinite enrollments is not permitted; max is %d",
			configuredMax)
	}
	if requestedMax > configuredMax {
		return 0, fmt.Errorf("Max enrollments of %d is not permitted; max is %d",
			requestedMax, configuredMax)
	}
	return requestedMax, nil
}

// Make all file names in the CA config absolute
func (ca *CA) makeFileNamesAbsolute() error {
	log.Debug("Making CA filenames absolute")
	fields := []*string{
		&ca.Config.CA.Certfile,
		&ca.Config.CA.Keyfile,
		&ca.Config.CA.Chainfile,
	}
	err := util.MakeFileNamesAbsolute(fields, ca.HomeDir)
	if err != nil {
		return err
	}
	err = tls.AbsTLSClient(&ca.Config.DB.TLS, ca.HomeDir)
	if err != nil {
		return err
	}
	err = tls.AbsTLSClient(&ca.Config.LDAP.TLS, ca.HomeDir)
	if err != nil {
		return err
	}
	return nil
}

// Convert all comma separated strings to string arrays
func (ca *CA) normalizeStringSlices() {
	fields := []*[]string{
		&ca.Config.CSR.Hosts,
		&ca.Config.DB.TLS.CertFiles,
		&ca.Config.LDAP.TLS.CertFiles,
	}
	for _, namePtr := range fields {
		norm := util.NormalizeStringSlice(*namePtr)
		*namePtr = norm
	}
}

// userHasAttribute returns nil error and the value of the attribute
// if the user has the attribute, or an appropriate error if the user
// does not have this attribute.
func (ca *CA) userHasAttribute(username, attrname string) (string, error) {
	val, err := ca.getUserAttrValue(username, attrname)
	if err != nil {
		return "", err
	}
	if val == "" {
		return "", fmt.Errorf("Identity '%s' does not have attribute '%s'", username, attrname)
	}
	return val, nil
}

// attributeIsTrue returns nil if the attribute has
// one of the following values: "1", "t", "T", "true", "TRUE", "True";
// otherwise it will return an error
func (ca *CA) attributeIsTrue(username, attrname string) error {
	val, err := ca.userHasAttribute(username, attrname)
	if err != nil {
		return err
	}
	val2, err := strconv.ParseBool(val)
	if err != nil {
		return fmt.Errorf("Invalid value for attribute '%s' of identity '%s': %s", attrname, username, err)
	}
	if val2 {
		return nil
	}
	return fmt.Errorf("Attribute '%s' is not set to true for identity '%s'", attrname, username)
}

// getUserAttrValue returns a user's value for an attribute
func (ca *CA) getUserAttrValue(username, attrname string) (string, error) {
	log.Debugf("getUserAttrValue identity=%s, attr=%s", username, attrname)
	user, err := ca.registry.GetUser(username, []string{attrname})
	if err != nil {
		return "", err
	}
	attrval := user.GetAttribute(attrname)
	log.Debugf("getUserAttrValue identity=%s, name=%s, value=%s", username, attrname, attrval)
	return attrval, nil
}

// getUserAffiliation returns a user's affiliation
func (ca *CA) getUserAffiliation(username string) (string, error) {
	log.Debugf("getUserAffilliation identity=%s", username)
	user, err := ca.registry.GetUserInfo(username)
	if err != nil {
		return "", err
	}
	aff := user.Affiliation
	log.Debugf("getUserAffiliation identity=%s, aff=%s", username, aff)
	return aff, nil
}

// Fill the CA info structure appropriately
func (ca *CA) fillCAInfo(info *serverInfoResponseNet) error {
	caChain, err := ca.getCAChain()
	if err != nil {
		return err
	}
	info.CAName = ca.Config.CA.Name
	info.CAChain = util.B64Encode(caChain)
	return nil
}

// Perfroms checks on the provided CA cert to make sure it's valid
func (ca *CA) validateCert(certFile string, keyFile string) error {
	log.Debug("Validating the CA certificate and key")
	var err error
	var certPEM []byte

	certPEM, err = ioutil.ReadFile(certFile)
	if err != nil {
		return fmt.Errorf(certificateError+" '%s': %s", certFile, err)
	}

	cert, err := util.GetX509CertificateFromPEM(certPEM)
	if err != nil {
		return fmt.Errorf(certificateError+" '%s': %s", certFile, err)
	}

	if err = validateDates(cert); err != nil {
		return fmt.Errorf(certificateError+" '%s': %s", certFile, err)
	}
	if err = validateUsage(cert); err != nil {
		return fmt.Errorf(certificateError+" '%s': %s", certFile, err)
	}
	if err = validateIsCA(cert); err != nil {
		return fmt.Errorf(certificateError+" '%s': %s", certFile, err)
	}
	if err = validateKeyType(cert); err != nil {
		return fmt.Errorf(certificateError+" '%s': %s", certFile, err)
	}
	if err = validateKeySize(cert); err != nil {
		return fmt.Errorf(certificateError+" '%s': %s", certFile, err)
	}
	if err = validateMatchingKeys(cert, keyFile); err != nil {
		return fmt.Errorf("Invalid certificate and/or key in files '%s' and '%s': %s", certFile, keyFile, err)
	}
	log.Debug("Validation of CA certificate and key successfull")

	return nil
}

func validateDates(cert *x509.Certificate) error {
	log.Debug("Check CA certificate for valid dates")

	notAfter := cert.NotAfter
	currentTime := time.Now().UTC()

	if currentTime.After(notAfter) {
		return errors.New("Certificate provided has expired")
	}

	notBefore := cert.NotBefore
	if currentTime.Before(notBefore) {
		return errors.New("Certificate provided not valid until later date")
	}

	return nil
}

func validateUsage(cert *x509.Certificate) error {
	log.Debug("Check CA certificate for valid usages")

	if cert.KeyUsage == 0 {
		return errors.New("No usage specified for certificate")
	}

	if cert.KeyUsage&x509.KeyUsageCertSign == 0 {
		return errors.New("'Cert Sign' key usage is required")
	}

	return nil
}

func validateIsCA(cert *x509.Certificate) error {
	log.Debug("Check CA certificate for valid IsCA value")

	if !cert.IsCA {
		return errors.New("Certificate not configured to be used for CA")
	}

	return nil
}

func validateKeyType(cert *x509.Certificate) error {
	log.Debug("Check that key type is supported")

	switch cert.PublicKey.(type) {
	case *dsa.PublicKey:
		return errors.New("Unsupported key type: DSA")
	}

	return nil
}

func validateKeySize(cert *x509.Certificate) error {
	log.Debug("Check that key size is of appropriate length")

	switch cert.PublicKey.(type) {
	case *rsa.PublicKey:
		size := cert.PublicKey.(*rsa.PublicKey).N.BitLen()
		if size < 2048 {
			return errors.New("Key size is less than 2048 bits")
		}
	}

	return nil
}

func validateMatchingKeys(cert *x509.Certificate, keyFile string) error {
	log.Debug("Check that public key and private key match")

	keyPEM, err := ioutil.ReadFile(keyFile)
	if err != nil {
		return err
	}

	pubKey := cert.PublicKey
	switch pubKey.(type) {
	case *rsa.PublicKey:
		privKey, err := util.GetRSAPrivateKey(keyPEM)
		if err != nil {
			return err
		}

		if privKey.PublicKey.N.Cmp(pubKey.(*rsa.PublicKey).N) != 0 {
			return errors.New("Public key and private key do not match")
		}
	case *ecdsa.PublicKey:
		privKey, err := util.GetECPrivateKey(keyPEM)
		if err != nil {
			return err
		}

		if privKey.PublicKey.X.Cmp(pubKey.(*ecdsa.PublicKey).X) != 0 {
			return errors.New("Public key and private key do not match")
		}
	}

	return nil
}

// Load CN from existing enrollment information
func (ca *CA) loadCNFromEnrollmentInfo(certFile string) (string, error) {
	log.Debug("Loading CN from existing enrollment information")
	cert, err := util.ReadFile(certFile)
	if err != nil {
		log.Debugf("No cert found at %s", certFile)
		return "", err
	}
	name, err := util.GetEnrollmentIDFromPEM(cert)
	if err != nil {
		return "", err
	}
	return name, nil
}

func writeFile(file string, buf []byte, perm os.FileMode) error {
	err := os.MkdirAll(filepath.Dir(file), 0755)
	if err != nil {
		return err
	}
	return ioutil.WriteFile(file, buf, perm)
}

func affiliationPath(name, parent string) string {
	if parent == "" {
		return name
	}
	return fmt.Sprintf("%s.%s", parent, name)
}

func parseDuration(str string) time.Duration {
	d, err := time.ParseDuration(str)
	if err != nil {
		panic(err)
	}
	return d
}

func initSigningProfile(spp **config.SigningProfile, expiry time.Duration, isCA bool) {
	sp := *spp
	if sp == nil {
		sp = &config.SigningProfile{CAConstraint: config.CAConstraint{IsCA: isCA}}
		*spp = sp
	}
	if sp.Usage == nil {
		sp.Usage = []string{"cert sign"}
	}
	if sp.Expiry == 0 {
		sp.Expiry = expiry
	}
}
