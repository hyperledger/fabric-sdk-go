/*
Copyright IBM Corp. 2016 All Rights Reserved.

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

package ldap

import (
	"errors"
	"fmt"
	"net"
	"net/url"
	"strconv"
	"strings"

	"github.com/cloudflare/cfssl/log"
	"github.com/hyperledger/fabric-ca/lib/spi"
	ctls "github.com/hyperledger/fabric-ca/lib/tls"
	ldap "gopkg.in/ldap.v2"
)

var (
	dnAttr          = []string{"dn"}
	errNotSupported = errors.New("Not supported")
)

// Config is the configuration object for this LDAP client
type Config struct {
	Enabled     bool   `def:"false" help:"Enable the LDAP client for authentication and attributes"`
	URL         string `help:"LDAP client URL of form ldap://adminDN:adminPassword@host[:port]/base"`
	UserFilter  string `def:"(uid=%s)" help:"The LDAP user filter to use when searching for users"`
	GroupFilter string `def:"(memberUid=%s)" help:"The LDAP group filter for a single affiliation group"`
	TLS         ctls.ClientTLSConfig
}

// NewClient creates an LDAP client
func NewClient(cfg *Config) (*Client, error) {
	log.Debugf("Creating new LDAP client for %+v", cfg)
	if cfg == nil {
		return nil, errors.New("LDAP configuration is nil")
	}
	if cfg.URL == "" {
		return nil, errors.New("LDAP configuration requires a 'URL'")
	}
	u, err := url.Parse(cfg.URL)
	if err != nil {
		return nil, err
	}
	var defaultPort string
	switch u.Scheme {
	case "ldap":
		defaultPort = "389"
	case "ldaps":
		defaultPort = "636"
	default:
		return nil, fmt.Errorf("invalid LDAP scheme: %s", u.Scheme)
	}
	var host, port string
	if strings.Index(u.Host, ":") < 0 {
		host = u.Host
		port = defaultPort
	} else {
		host, port, err = net.SplitHostPort(u.Host)
		if err != nil {
			return nil, fmt.Errorf("invalid LDAP host:port (%s): %s", u.Host, err)
		}
	}
	portVal, err := strconv.Atoi(port)
	if err != nil {
		return nil, fmt.Errorf("invalid LDAP port (%s): %s", port, err)
	}
	c := new(Client)
	c.Host = host
	c.Port = portVal
	c.UseSSL = u.Scheme == "ldaps"
	if u.User != nil {
		c.AdminDN = u.User.Username()
		c.AdminPassword, _ = u.User.Password()
	}
	c.Base = u.Path
	if c.Base != "" && strings.HasPrefix(c.Base, "/") {
		c.Base = c.Base[1:]
	}
	c.UserFilter = cfgVal(cfg.UserFilter, "(uid=%s)")
	c.GroupFilter = cfgVal(cfg.GroupFilter, "(memberUid=%s)")
	c.TLS = &cfg.TLS
	log.Debug("LDAP client was successfully created")
	return c, nil
}

func cfgVal(val1, val2 string) string {
	if val1 != "" {
		return val1
	}
	return val2
}

// Client is an LDAP client
type Client struct {
	Host          string
	Port          int
	UseSSL        bool
	AdminDN       string
	AdminPassword string
	Base          string
	UserFilter    string // e.g. "(uid=%s)"
	GroupFilter   string // e.g. "(memberUid=%s)"
	AdminConn     *ldap.Conn
	TLS           *ctls.ClientTLSConfig
}

// GetUser returns a user object for username and attribute values
// for the requested attribute names
func (lc *Client) GetUser(username string, attrNames []string) (spi.User, error) {

	log.Debugf("Getting user '%s'", username)

	// Connect to the LDAP server as admin if not already connected
	err := lc.adminConnect()
	if err != nil {
		return nil, err
	}

	// Search for the given username
	sreq := ldap.NewSearchRequest(
		lc.Base, ldap.ScopeWholeSubtree,
		ldap.NeverDerefAliases, 0, 0, false,
		fmt.Sprintf(lc.UserFilter, username),
		attrNames,
		nil,
	)
	sresp, err := lc.AdminConn.Search(sreq)
	if err != nil {
		return nil, fmt.Errorf("LDAP search failure: %s; search request: %+v", err, sreq)
	}
	// Make sure there was exactly one match found
	if len(sresp.Entries) < 1 {
		return nil, fmt.Errorf("User '%s' does not exist in LDAP directory", username)
	}
	if len(sresp.Entries) > 1 {
		return nil, fmt.Errorf("Multiple users with name '%s' exist in LDAP directory", username)
	}

	DN := sresp.Entries[0].DN

	// Create the map of attributes
	attrs := make(map[string]string)
	for _, attrName := range attrNames {
		if attrName == "dn" {
			attrs["dn"] = DN
		} else {
			attrs[attrName] = sresp.Entries[0].GetAttributeValue(attrName)
		}
	}

	// Construct the user object
	user := &User{
		name:   username,
		dn:     DN,
		attrs:  attrs,
		client: lc,
	}

	log.Debug("Successfully retrieved user '%s', DN: %s", username, DN)

	return user, nil
}

// GetUserInfo gets user information from database
func (lc *Client) GetUserInfo(id string) (spi.UserInfo, error) {
	var userInfo spi.UserInfo
	return userInfo, errNotSupported
}

// InsertUser inserts a user
func (lc *Client) InsertUser(user spi.UserInfo) error {
	return errNotSupported
}

// UpdateUser updates a user
func (lc *Client) UpdateUser(user spi.UserInfo) error {
	return errNotSupported
}

// DeleteUser deletes a user
func (lc *Client) DeleteUser(id string) error {
	return errNotSupported
}

// GetAffiliation returns an affiliation group
func (lc *Client) GetAffiliation(name string) (spi.Affiliation, error) {
	return nil, errNotSupported
}

// GetRootAffiliation returns the root affiliation group
func (lc *Client) GetRootAffiliation() (spi.Affiliation, error) {
	return nil, errNotSupported
}

// InsertAffiliation adds an affiliation group
func (lc *Client) InsertAffiliation(name string, prekey string) error {
	return errNotSupported
}

// DeleteAffiliation deletes an affiliation group
func (lc *Client) DeleteAffiliation(name string) error {
	return errNotSupported
}

// Create an admin connection to the LDAP server and cache it in the client
func (lc *Client) adminConnect() error {
	if lc.AdminConn == nil {
		conn, err := lc.newConnection()
		if err != nil {
			return err
		}
		lc.AdminConn = conn
	}
	return nil
}

// Connect to the LDAP server and bind as user as admin user as specified in LDAP URL
func (lc *Client) newConnection() (conn *ldap.Conn, err error) {
	address := fmt.Sprintf("%s:%d", lc.Host, lc.Port)
	if !lc.UseSSL {
		log.Debug("Connecting to LDAP server over TCP")
		conn, err = ldap.Dial("tcp", address)
		if err != nil {
			return conn, fmt.Errorf("Failed to connect to LDAP server over TCP at %s: %s", address, err)
		}
	} else {
		log.Debug("Connecting to LDAP server over TLS")
		tlsConfig, err2 := ctls.GetClientTLSConfig(lc.TLS)
		if err2 != nil {
			return nil, fmt.Errorf("Failed to get client TLS config: %s", err2)
		}

		tlsConfig.ServerName = lc.Host

		conn, err = ldap.DialTLS("tcp", address, tlsConfig)
		if err != nil {
			return conn, fmt.Errorf("Failed to connect to LDAP server over TLS at %s: %s", address, err)
		}
	}
	// Bind with a read only user
	if lc.AdminDN != "" && lc.AdminPassword != "" {
		log.Debug("Binding to the LDAP server as admin user %s", lc.AdminDN)
		err := conn.Bind(lc.AdminDN, lc.AdminPassword)
		if err != nil {
			return nil, fmt.Errorf("LDAP bind failure as %s: %s", lc.AdminDN, err)
		}
	}
	return conn, nil
}

// User represents a single user
type User struct {
	name   string
	dn     string
	attrs  map[string]string
	client *Client
}

// GetName returns the user's enrollment ID, which is the DN (Distinquished Name)
func (u *User) GetName() string {
	return u.dn
}

// Login logs a user in using password
func (u *User) Login(password string) error {

	// Get a connection to use to bind over as the user to check the password
	conn, err := u.client.newConnection()
	if err != nil {
		return err
	}
	defer conn.Close()

	// Bind calls the LDAP server to check the user's password
	err = conn.Bind(u.dn, password)
	if err != nil {
		return fmt.Errorf("LDAP authentication failure for user '%s' (DN=%s): %s", u.name, u.dn, err)
	}

	return nil

}

// GetAffiliationPath returns the affiliation path for this user
func (u *User) GetAffiliationPath() []string {
	return reverse(strings.Split(u.dn, ","))
}

// GetAttribute returns the value of an attribute, or "" if not found
func (u *User) GetAttribute(name string) string {
	return u.attrs[name]
}

// Returns a slice with the elements reversed
func reverse(in []string) []string {
	size := len(in)
	out := make([]string, size)
	for i := 0; i < size; i++ {
		out[i] = in[size-i-1]
	}
	return out
}
