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
	"crypto/tls"
	"errors"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strconv"

	"github.com/cloudflare/cfssl/config"
	cfcsr "github.com/cloudflare/cfssl/csr"
	"github.com/cloudflare/cfssl/initca"
	"github.com/cloudflare/cfssl/log"
	"github.com/cloudflare/cfssl/signer"
	"github.com/cloudflare/cfssl/signer/universal"
	"github.com/hyperledger/fabric-ca/api"
	"github.com/hyperledger/fabric-ca/lib/dbutil"
	"github.com/hyperledger/fabric-ca/lib/ldap"
	"github.com/hyperledger/fabric-ca/lib/spi"
	"github.com/hyperledger/fabric-ca/util"
	"github.com/hyperledger/fabric/bccsp"
	"github.com/hyperledger/fabric/bccsp/factory"
	"github.com/jmoiron/sqlx"

	_ "github.com/go-sql-driver/mysql" // import to support MySQL
	_ "github.com/lib/pq"              // import to support Postgres
	_ "github.com/mattn/go-sqlite3"    // import to support SQLite3
)

const (
	defaultDatabaseType = "sqlite3"
)

// Server is the fabric-ca server
type Server struct {
	// The home directory for the server
	HomeDir string
	// BlockingStart if true makes the Start function blocking;
	// It is non-blocking by default.
	BlockingStart bool
	// The server's configuration
	Config *ServerConfig
	// The parent server URL, which is non-null if this is an intermediate server
	ParentServerURL string
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
	// The server mux
	mux *http.ServeMux
	// The current listener for this server
	listener net.Listener
	// An error which occurs when serving
	serveError error
}

// Init initializes a fabric-ca server
func (s *Server) Init(renew bool) (err error) {
	// Initialize the config, setting defaults, etc
	err = s.initConfig()
	if err != nil {
		return err
	}
	// Initialize the Crypto Service Provider
	s.csp = factory.GetDefault()
	// Initialize key materials
	err = s.initKeyMaterial(renew)
	if err != nil {
		return err
	}
	// Initialize the database
	err = s.initDB()
	if err != nil {
		return err
	}
	// Initialize the enrollment signer
	err = s.initEnrollmentSigner()
	if err != nil {
		return err
	}
	// Successful initialization
	return nil
}

// Start the fabric-ca server
func (s *Server) Start() (err error) {

	s.serveError = nil

	if s.listener != nil {
		return errors.New("server is already started")
	}

	// Initialize the server
	err = s.Init(false)
	if err != nil {
		return err
	}

	// Register http handlers
	s.registerHandlers()

	// Start listening and serving
	return s.listenAndServe()

}

// Stop the server
// WARNING: This forcefully closes the listening socket and may cause
// requests in transit to fail, and so is only used for testing.
// A graceful shutdown will be supported with golang 1.8.
func (s *Server) Stop() error {
	if s.listener == nil {
		return errors.New("server is not currently started")
	}
	err := s.listener.Close()
	s.listener = nil
	return err
}

// Initialize the fabric-ca server's key material
func (s *Server) initKeyMaterial(renew bool) error {
	log.Debugf("Init with home %s and config %+v", s.HomeDir, *s.Config)

	// Make the path names absolute in the config
	s.makeFileNamesAbsolute()

	keyFile := s.Config.CA.Keyfile
	certFile := s.Config.CA.Certfile

	// If we aren't renewing and the key and cert files exist, do nothing
	if !renew {
		// If they both exist, the server was already initialized
		keyFileExists := util.FileExists(keyFile)
		certFileExists := util.FileExists(certFile)
		if keyFileExists && certFileExists {
			log.Info("The CA key and certificate files already exist")
			log.Infof("Key file location: %s", keyFile)
			log.Infof("Certificate file location: %s", certFile)
			return nil
		}
	}

	// Get the CA cert and key
	cert, key, err := s.getCACertAndKey()
	if err != nil {
		return fmt.Errorf("Failed to initialize CA: %s", err)
	}

	// Store the key and certificate to file
	err = writeFile(keyFile, key, 0600)
	if err != nil {
		return fmt.Errorf("Failed to store key: %s", err)
	}
	err = writeFile(certFile, cert, 0644)
	if err != nil {
		return fmt.Errorf("Failed to store certificate: %s", err)
	}
	log.Info("The CA key and certificate files were generated")
	log.Infof("Key file location: %s", keyFile)
	log.Infof("Certificate file location: %s", certFile)
	return nil
}

// Get the CA certificate and key for this server
func (s *Server) getCACertAndKey() (cert, key []byte, err error) {
	log.Debugf("Getting CA cert and key; parent server URL is '%s'", s.ParentServerURL)
	if s.ParentServerURL != "" {
		// This is an intermediate CA, so call the parent fabric-ca-server
		// to get the key and cert
		clientCfg := s.Config.Client
		if clientCfg == nil {
			clientCfg = &ClientConfig{}
		}
		if clientCfg.Enrollment.Profile == "" {
			clientCfg.Enrollment.Profile = "ca"
		}
		if clientCfg.Enrollment.CSR == nil {
			clientCfg.Enrollment.CSR = &api.CSRInfo{}
		}
		if clientCfg.Enrollment.CSR.CA == nil {
			clientCfg.Enrollment.CSR.CA = &cfcsr.CAConfig{PathLength: 0, PathLenZero: true}
		}
		log.Debugf("Intermediate enrollment request: %v", clientCfg.Enrollment)
		var resp *EnrollmentResponse
		resp, err = clientCfg.Enroll(s.ParentServerURL, s.HomeDir)
		if err != nil {
			return nil, nil, err
		}
		ecert := resp.Identity.GetECert()
		if ecert == nil {
			return nil, nil, errors.New("No ECert from parent server")
		}
		cert = ecert.Cert()
		key = ecert.Key()
		// Store the chain file as the concatenation of the parent's chain plus the cert.
		chainPath := s.Config.CA.Chainfile
		if chainPath == "" {
			chainPath, err = util.MakeFileAbs("ca-chain.pem", s.HomeDir)
			if err != nil {
				return nil, nil, fmt.Errorf("Failed to create intermediate chain file path: %s", err)
			}
		}
		chain := s.concatChain(resp.ServerInfo.CAChain, cert)
		err = os.MkdirAll(path.Dir(chainPath), 0755)
		if err != nil {
			return nil, nil, fmt.Errorf("Failed to create intermediate chain file directory: %s", err)
		}
		err = util.WriteFile(chainPath, chain, 0644)
		if err != nil {
			return nil, nil, fmt.Errorf("Failed to create intermediate chain file: %s", err)
		}
		log.Debugf("Stored intermediate certificate chain at %s", chainPath)
	} else {
		// This is a root CA, so call cfssl to get the key and cert.
		csr := &s.Config.CSR
		req := cfcsr.CertificateRequest{
			CN:    csr.CN,
			Names: csr.Names,
			Hosts: csr.Hosts,
			// FIXME: NewBasicKeyRequest only does ecdsa 256; use config
			KeyRequest:   cfcsr.NewBasicKeyRequest(),
			CA:           csr.CA,
			SerialNumber: csr.SerialNumber,
		}
		// Call CFSSL to initialize the CA
		cert, _, key, err = initca.New(&req)
	}
	if err != nil {
		return nil, nil, err
	}
	return cert, key, nil
}

// Return a chain which is the concatenation of chain and cert
func (s *Server) concatChain(chain []byte, cert []byte) []byte {
	result := make([]byte, len(chain)+len(cert))
	copy(result[:len(chain)], chain)
	copy(result[len(chain):], cert)
	return result
}

// Get the CA chain
func (s *Server) getCAChain() (chain []byte, err error) {
	if s.Config == nil {
		return nil, errors.New("The server has no configuration")
	}
	ca := &s.Config.CA
	// If the chain file exists, we always return the chain from here
	if util.FileExists(ca.Chainfile) {
		return util.ReadFile(ca.Chainfile)
	}
	// Otherwise, if this is a root CA, we always return the contents of the CACertfile
	if s.ParentServerURL == "" {
		return util.ReadFile(ca.Certfile)
	}
	// If this is an intermediate CA but the ca.Chainfile doesn't exist,
	// it is an error.  It should have been created during intermediate CA enrollment.
	return nil, fmt.Errorf("Chain file does not exist at %s", ca.Chainfile)
}

// RegisterBootstrapUser registers the bootstrap user with appropriate privileges
func (s *Server) RegisterBootstrapUser(user, pass, affiliation string) error {
	// Initialize the config, setting defaults, etc
	log.Debugf("RegisterBootstrapUser - User: %s, Pass: %s, affiliation: %s", user, pass, affiliation)

	if user == "" || pass == "" {
		return errors.New("empty user and/or pass not allowed")
	}
	err := s.initConfig()
	if err != nil {
		return fmt.Errorf("Failed to register bootstrap user '%s': %s", user, err)
	}

	id := ServerConfigIdentity{
		Name:           user,
		Pass:           pass,
		Type:           "user",
		Affiliation:    affiliation,
		MaxEnrollments: s.Config.Registry.MaxEnrollments,
		Attrs: map[string]string{
			"hf.Registrar.Roles":         "client,user,peer,validator,auditor",
			"hf.Registrar.DelegateRoles": "client,user,validator,auditor",
			"hf.Revoker":                 "true",
			"hf.IntermediateCA":          "true",
		},
	}
	registry := &s.Config.Registry
	registry.Identities = append(registry.Identities, id)
	log.Debugf("Registered bootstrap identity: %+v", &id)
	return nil
}

// Do any ize the config, setting any defaults and making filenames absolute
func (s *Server) initConfig() (err error) {
	// Init home directory if not set
	if s.HomeDir == "" {
		s.HomeDir, err = os.Getwd()
		if err != nil {
			return fmt.Errorf("Failed to initialize server's home directory: %s", err)
		}
	}
	// Init config if not set
	if s.Config == nil {
		s.Config = new(ServerConfig)
	}
	// Set config defaults
	cfg := s.Config
	if cfg.Address == "" {
		cfg.Address = DefaultServerAddr
	}
	if cfg.Port == 0 {
		cfg.Port = DefaultServerPort
	}
	if cfg.CA.Certfile == "" {
		cfg.CA.Certfile = "ca-cert.pem"
	}
	if cfg.CA.Keyfile == "" {
		cfg.CA.Keyfile = "ca-key.pem"
	}
	if cfg.CSR.CN == "" {
		cfg.CSR.CN = "fabric-ca-server"
	}
	// Set log level if debug is true
	if cfg.Debug {
		log.Level = log.LevelDebug
	}
	// Init the BCCSP
	err = factory.InitFactories(s.Config.CSP)
	if err != nil {
		panic(fmt.Errorf("Could not initialize BCCSP Factories [%s]", err))
	}

	return nil
}

// Initialize the database for the server
func (s *Server) initDB() error {
	db := &s.Config.DB

	var err error
	var exists bool

	if db.Type == "" || db.Type == defaultDatabaseType {

		db.Type = defaultDatabaseType

		if db.Datasource == "" {
			db.Datasource = "fabric-ca-server.db"
		}

		db.Datasource, err = util.MakeFileAbs(db.Datasource, s.HomeDir)
		if err != nil {
			return err
		}
	}

	log.Debugf("Initializing '%s' data base at '%s'", db.Type, db.Datasource)

	switch db.Type {
	case defaultDatabaseType:
		s.db, exists, err = dbutil.NewUserRegistrySQLLite3(db.Datasource)
		if err != nil {
			return err
		}
	case "postgres":
		s.db, exists, err = dbutil.NewUserRegistryPostgres(db.Datasource, &db.TLS)
		if err != nil {
			return err
		}
	case "mysql":
		s.db, exists, err = dbutil.NewUserRegistryMySQL(db.Datasource, &db.TLS)
		if err != nil {
			return err
		}
	default:
		return fmt.Errorf("Invalid db.type in config file: '%s'; must be 'sqlite3', 'postgres', or 'mysql'", db.Type)
	}

	// Set the certificate DB accessor
	s.certDBAccessor = NewCertDBAccessor(s.db)

	// Initialize the user registry.
	// If LDAP is not configured, the fabric-ca server functions as a user
	// registry based on the database.
	err = s.initUserRegistry()
	if err != nil {
		return err
	}

	// If the DB doesn't exist, bootstrap it
	if !exists {
		err = s.loadUsersTable()
		if err != nil {
			return err
		}
		err = s.loadAffiliationsTable()
		if err != nil {
			return err
		}
	}
	log.Infof("Initialized %s data base at %s", db.Type, db.Datasource)
	return nil
}

// Initialize the user registry interface
func (s *Server) initUserRegistry() error {
	log.Debug("Initializing user registry")
	var err error
	ldapCfg := &s.Config.LDAP

	if ldapCfg.Enabled {
		// Use LDAP for the user registry
		s.registry, err = ldap.NewClient(ldapCfg)
		log.Debugf("Initialized LDAP user registry; err=%s", err)
		return err
	}

	// Use the DB for the user registry
	dbAccessor := new(Accessor)
	dbAccessor.SetDB(s.db)
	s.registry = dbAccessor
	log.Debug("Initialized DB user registry")
	return nil
}

// Initialize the enrollment signer
func (s *Server) initEnrollmentSigner() (err error) {

	c := s.Config

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
	if c.Remote != "" {
		err = policy.OverrideRemotes(c.Remote)
		if err != nil {
			return fmt.Errorf("Failed initializing enrollment signer: %s", err)
		}
	}

	// Get CFSSL's universal root and signer
	root := universal.Root{
		Config: map[string]string{
			"cert-file": c.CA.Certfile,
			"key-file":  c.CA.Keyfile,
		},
		ForceRemote: c.Remote != "",
	}
	s.enrollSigner, err = universal.NewSigner(root, policy)
	if err != nil {
		return err
	}
	s.enrollSigner.SetDBAccessor(s.certDBAccessor)

	// Successful enrollment
	return nil
}

// Register all endpoint handlers
func (s *Server) registerHandlers() {
	s.mux = http.NewServeMux()
	s.registerHandler("info", newInfoHandler, noAuth)
	s.registerHandler("register", newRegisterHandler, token)
	s.registerHandler("enroll", newEnrollHandler, basic)
	s.registerHandler("reenroll", newReenrollHandler, token)
	s.registerHandler("revoke", newRevokeHandler, token)
	s.registerHandler("tcert", newTCertHandler, token)
}

// Register an endpoint handler
func (s *Server) registerHandler(
	path string,
	getHandler func(server *Server) (http.Handler, error),
	at authType) {

	var handler http.Handler

	handler, err := getHandler(s)
	if err != nil {
		log.Warningf("Endpoint '%s' is disabled: %s", path, err)
		return
	}
	handler = &fcaAuthHandler{
		server:   s,
		authType: at,
		next:     handler,
	}
	s.mux.Handle("/"+path, handler)
	// TODO: Remove the following line once all SDKs stop using the prefixed paths
	// See https://jira.hyperledger.org/browse/FAB-2597
	s.mux.Handle("/api/v1/cfssl/"+path, handler)
}

// Starting listening and serving
func (s *Server) listenAndServe() (err error) {

	var listener net.Listener

	c := s.Config

	// Set default listening address and port
	if c.Address == "" {
		c.Address = DefaultServerAddr
	}
	if c.Port == 0 {
		c.Port = DefaultServerPort
	}
	addr := net.JoinHostPort(c.Address, strconv.Itoa(c.Port))

	if c.TLS.Enabled {
		log.Debug("TLS is enabled")
		var cer tls.Certificate
		cer, err = tls.LoadX509KeyPair(c.TLS.CertFile, c.TLS.KeyFile)
		if err != nil {
			return err
		}
		config := &tls.Config{Certificates: []tls.Certificate{cer}}
		listener, err = tls.Listen("tcp", addr, config)
		if err != nil {
			return fmt.Errorf("TLS listen failed: %s", err)
		}
		log.Infof("Listening at https://%s", addr)
	} else {
		listener, err = net.Listen("tcp", addr)
		if err != nil {
			return fmt.Errorf("TCP listen failed: %s", err)
		}
		log.Infof("Listening at http://%s", addr)
	}
	s.listener = listener

	// Start serving requests, either blocking or non-blocking
	if s.BlockingStart {
		return s.serve()
	}
	go s.serve()
	return nil
}

func (s *Server) serve() error {
	s.serveError = http.Serve(s.listener, s.mux)
	log.Errorf("Server has stopped serving: %s", s.serveError)
	if s.listener != nil {
		s.listener.Close()
		s.listener = nil
	}
	return s.serveError
}

// loadUsersTable adds the configured users to the table if not already found
func (s *Server) loadUsersTable() error {
	log.Debug("Loading users table")
	registry := &s.Config.Registry
	for _, id := range registry.Identities {
		log.Debugf("Loading identity '%s'", id.Name)
		err := s.addIdentity(&id, false)
		if err != nil {
			return err
		}
	}
	log.Debug("Successfully loaded users table")
	return nil
}

// loadAffiliationsTable adds the configured affiliations to the table
func (s *Server) loadAffiliationsTable() error {
	log.Debug("Loading affiliations table")
	err := s.loadAffiliationsTableR(s.Config.Affiliations, "")
	if err == nil {
		log.Debug("Successfully loaded affiliations table")
	}
	log.Debug("Successfully loaded groups table")
	return nil
}

// Recursive function to load the affiliations table hierarchy
func (s *Server) loadAffiliationsTableR(val interface{}, parentPath string) (err error) {
	var path string
	if val == nil {
		return nil
	}
	switch val.(type) {
	case string:
		path = affiliationPath(val.(string), parentPath)
		err = s.addAffiliation(path, parentPath)
		if err != nil {
			return err
		}
	case []string:
		for _, ele := range val.([]string) {
			err = s.loadAffiliationsTableR(ele, parentPath)
			if err != nil {
				return err
			}
		}
	case []interface{}:
		for _, ele := range val.([]interface{}) {
			err = s.loadAffiliationsTableR(ele, parentPath)
			if err != nil {
				return err
			}
		}
	default:
		for name, ele := range val.(map[string]interface{}) {
			path = affiliationPath(name, parentPath)
			err = s.addAffiliation(path, parentPath)
			if err != nil {
				return err
			}
			err = s.loadAffiliationsTableR(ele, path)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

// Add an identity to the registry
func (s *Server) addIdentity(id *ServerConfigIdentity, errIfFound bool) error {
	user, _ := s.registry.GetUser(id.Name, nil)
	if user != nil {
		if errIfFound {
			return fmt.Errorf("Identity '%s' is already registered", id.Name)
		}
		log.Debugf("Loaded identity: %+v", id)
		return nil
	}
	maxEnrollments, err := s.getMaxEnrollments(id.MaxEnrollments)
	if err != nil {
		return err
	}
	rec := spi.UserInfo{
		Name:           id.Name,
		Pass:           id.Pass,
		Type:           id.Type,
		Affiliation:    id.Affiliation,
		Attributes:     s.convertAttrs(id.Attrs),
		MaxEnrollments: maxEnrollments,
	}
	err = s.registry.InsertUser(rec)
	if err != nil {
		return fmt.Errorf("Failed to insert user '%s': %s", id.Name, err)
	}
	log.Debugf("Registered identity: %+v", id)
	return nil
}

func (s *Server) addAffiliation(path, parentPath string) error {
	log.Debugf("Adding affiliation %s", path)
	return s.registry.InsertAffiliation(path, parentPath)
}

// CertDBAccessor returns the certificate DB accessor for server
func (s *Server) CertDBAccessor() *CertDBAccessor {
	return s.certDBAccessor
}

func (s *Server) convertAttrs(inAttrs map[string]string) []api.Attribute {
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
func (s *Server) getMaxEnrollments(requestedMax int) (int, error) {
	configuredMax := s.Config.Registry.MaxEnrollments
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

// Make all file names in the config absolute
func (s *Server) makeFileNamesAbsolute() error {
	fields := []*string{
		&s.Config.CA.Certfile,
		&s.Config.CA.Keyfile,
		&s.Config.CA.Chainfile,
		&s.Config.TLS.CertFile,
		&s.Config.TLS.KeyFile,
	}
	for _, namePtr := range fields {
		abs, err := util.MakeFileAbs(*namePtr, s.HomeDir)
		if err != nil {
			return err
		}
		*namePtr = abs
	}
	return nil
}

// userHasAttribute returns nil if the user has the attribute, or an
// appropriate error if the user does not have this attribute.
func (s *Server) userHasAttribute(username, attrname string) error {
	val, err := s.getUserAttrValue(username, attrname)
	if err != nil {
		return err
	}
	if val == "" {
		return fmt.Errorf("user '%s' does not have attribute '%s'", username, attrname)
	}
	return nil
}

// getUserAttrValue returns a user's value for an attribute
func (s *Server) getUserAttrValue(username, attrname string) (string, error) {
	log.Debugf("getUserAttrValue user=%s, attr=%s", username, attrname)
	user, err := s.registry.GetUser(username, []string{attrname})
	if err != nil {
		return "", err
	}
	attrval := user.GetAttribute(attrname)
	log.Debugf("getUserAttrValue user=%s, name=%s, value=%s", username, attrname, attrval)
	return attrval, nil
}

// getUserAffiliation returns a user's affiliation
func (s *Server) getUserAffiliation(username string) (string, error) {
	log.Debugf("getUserAffilliation user=%s", username)
	user, err := s.registry.GetUserInfo(username)
	if err != nil {
		return "", err
	}
	aff := user.Affiliation
	log.Debugf("getUserAttrValue user=%s, aff=%s, value=%s", username, aff)
	return aff, nil
}

// Fill the server info structure appropriately
func (s *Server) fillServerInfo(info *serverInfoResponseNet) error {
	caChain, err := s.getCAChain()
	if err != nil {
		return err
	}
	info.CAName = s.Config.CA.Name
	info.CAChain = util.B64Encode(caChain)
	return nil
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
