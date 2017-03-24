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

package lib

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"strconv"
	"strings"

	cfsslapi "github.com/cloudflare/cfssl/api"
	"github.com/cloudflare/cfssl/csr"
	"github.com/cloudflare/cfssl/log"
	"github.com/cloudflare/cfssl/signer"
	"github.com/hyperledger/fabric-ca/api"
	"github.com/hyperledger/fabric-ca/lib/tls"
	"github.com/hyperledger/fabric-ca/util"
)

const (
	clientConfigFile = "client-config.json"
)

// NewClient is the constructor for the fabric-ca client API
func NewClient(configFile string) (*Client, error) {
	c := new(Client)

	if configFile != "" {
		if _, err := os.Stat(configFile); err != nil {
			log.Info("Fabric-ca client configuration file not found. Using Defaults...")
		} else {
			var config []byte
			var err error
			config, err = ioutil.ReadFile(configFile)
			if err != nil {
				return nil, err
			}
			// Override any defaults
			err = util.Unmarshal([]byte(config), c, "NewClient")
			if err != nil {
				return nil, err
			}
		}
	}

	var cfg = new(ClientConfig)
	c.Config = cfg

	// Set defaults
	if c.Config.URL == "" {
		c.Config.URL = util.GetServerURL()
	}

	if c.HomeDir == "" {
		c.HomeDir = filepath.Dir(util.GetDefaultConfigFile("fabric-ca-client"))
	}

	if _, err := os.Stat(c.HomeDir); err != nil {
		if os.IsNotExist(err) {
			_, err := util.CreateClientHome()
			if err != nil {
				return nil, err
			}
		}
	}

	return c, nil
}

// Client is the fabric-ca client object
type Client struct {
	// HomeDir is the home directory
	HomeDir string `json:"homeDir,omitempty"`

	// The client's configuration
	Config *ClientConfig
}

// Enroll enrolls a new identity
// @param req The enrollment request
func (c *Client) Enroll(req *api.EnrollmentRequest) (*Identity, error) {
	log.Debugf("Enrolling %+v", req)

	// Generate the CSR
	csrPEM, key, err := c.GenCSR(req.CSR, req.Name)
	if err != nil {
		log.Debugf("enroll failure generating CSR: %s", err)
		return nil, err
	}

	// Get the body of the request
	sreq := signer.SignRequest{
		Hosts:   signer.SplitHosts(req.Hosts),
		Request: string(csrPEM),
		Profile: req.Profile,
		Label:   req.Label,
	}
	body, err := util.Marshal(sreq, "SignRequest")
	if err != nil {
		return nil, err
	}

	// Send the CSR to the fabric-ca server with basic auth header
	post, err := c.NewPost("enroll", body)
	if err != nil {
		return nil, err
	}
	post.SetBasicAuth(req.Name, req.Secret)
	result, err := c.SendPost(post)
	if err != nil {
		return nil, err
	}

	// Create an identity from the key and certificate in the response
	return c.newIdentityFromResponse(result, req.Name, key)
}

// newIdentityFromResponse returns an Identity for enroll and reenroll responses
// @param result The result from server
// @param id Name of identity being enrolled or reenrolled
// @param key The private key which was used to sign the request
func (c *Client) newIdentityFromResponse(result interface{}, id string, key []byte) (*Identity, error) {
	log.Debugf("newIdentityFromResponse %s", id)
	certByte, err := base64.StdEncoding.DecodeString(result.(string))
	if err != nil {
		return nil, fmt.Errorf("Invalid response format from server: %s", err)
	}
	return newIdentity(c, id, key, certByte), nil
}

// GenCSR generates a CSR (Certificate Signing Request)
func (c *Client) GenCSR(req *api.CSRInfo, id string) ([]byte, []byte, error) {
	log.Debugf("GenCSR %+v", req)

	cr := c.newCertificateRequest(req)
	cr.CN = id

	csrPEM, key, err := csr.ParseRequest(cr)
	if err != nil {
		log.Debugf("failed generating CSR: %s", err)
		return nil, nil, err
	}

	return csrPEM, key, nil
}

// newCertificateRequest creates a certificate request which is used to generate
// a CSR (Certificate Signing Request)
func (c *Client) newCertificateRequest(req *api.CSRInfo) *csr.CertificateRequest {
	cr := csr.CertificateRequest{}
	if req != nil && req.Names != nil {
		cr.Names = req.Names
	}
	if req != nil && req.Hosts != nil {
		cr.Hosts = req.Hosts
	} else {
		// Default requested hosts are local hostname
		hostname, _ := os.Hostname()
		if hostname != "" {
			cr.Hosts = make([]string, 1)
			cr.Hosts[0] = hostname
		}
	}
	if req != nil && req.KeyRequest != nil {
		cr.KeyRequest = req.KeyRequest
	}
	if req != nil {
		cr.CA = req.CA
		cr.SerialNumber = req.SerialNumber
	}
	return &cr
}

// LoadMyIdentity loads the client's identity from disk
func (c *Client) LoadMyIdentity() (*Identity, error) {
	return c.LoadIdentity(c.GetMyKeyFile(), c.GetMyCertFile())
}

// StoreMyIdentity stores my identity to disk
func (c *Client) StoreMyIdentity(key, cert []byte) error {
	err := util.WriteFile(c.GetMyKeyFile(), key, 0600)
	if err != nil {
		return err
	}
	return util.WriteFile(c.GetMyCertFile(), cert, 0644)
}

// GetMyKeyFile returns the path to this identity's key file
func (c *Client) GetMyKeyFile() string {
	file := os.Getenv("FABRIC_CA_KEY_FILE")
	if file == "" {
		file = path.Join(c.GetMyEnrollmentDir(), "key.pem")
	}
	return file
}

// GetMyCertFile returns the path to this identity's certificate file
func (c *Client) GetMyCertFile() string {
	file := os.Getenv("FABRIC_CA_CERT_FILE")
	if file == "" {
		file = path.Join(c.GetMyEnrollmentDir(), "cert.pem")
	}
	return file
}

// GetMyEnrollmentDir returns the path to this identity's enrollment directory
func (c *Client) GetMyEnrollmentDir() string {
	dir := os.Getenv("FABRIC_CA_ENROLLMENT_DIR")
	if dir == "" {
		dir = c.HomeDir
	}
	return dir
}

// LoadIdentity loads an identity from disk
func (c *Client) LoadIdentity(keyFile, certFile string) (*Identity, error) {
	key, err := util.ReadFile(keyFile)
	if err != nil {
		return nil, err
	}
	cert, err := util.ReadFile(certFile)
	if err != nil {
		return nil, err
	}
	return c.NewIdentity(key, cert)
}

// NewIdentity creates a new identity
func (c *Client) NewIdentity(key, cert []byte) (*Identity, error) {
	name, err := util.GetEnrollmentIDFromPEM(cert)
	if err != nil {
		return nil, err
	}
	return newIdentity(c, name, key, cert), nil
}

// LoadCSRInfo reads CSR (Certificate Signing Request) from a file
// @parameter path The path to the file contains CSR info in JSON format
func (c *Client) LoadCSRInfo(path string) (*api.CSRInfo, error) {
	csrJSON, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var csrInfo api.CSRInfo
	err = util.Unmarshal(csrJSON, &csrInfo, "LoadCSRInfo")
	if err != nil {
		return nil, err
	}
	return &csrInfo, nil
}

// NewPost create a new post request
func (c *Client) NewPost(endpoint string, reqBody []byte) (*http.Request, error) {
	curl, cerr := c.getURL(endpoint)
	if cerr != nil {
		return nil, cerr
	}
	req, err := http.NewRequest("POST", curl, bytes.NewReader(reqBody))
	if err != nil {
		return nil, fmt.Errorf("Failed posting to %s: %s", curl, err)
	}
	return req, nil
}

// SendPost sends a request to the LDAP server and returns a response
func (c *Client) SendPost(req *http.Request) (interface{}, error) {
	reqStr := util.HTTPRequestToString(req)
	log.Debugf("Sending request\n%s", reqStr)

	var tr = new(http.Transport)

	if c.Config.TLS.Enabled {
		log.Info("TLS Enabled")

		err := tls.AbsTLSClient(&c.Config.TLS, c.HomeDir)
		if err != nil {
			return nil, err
		}

		tlsConfig, err := tls.GetClientTLSConfig(&c.Config.TLS)
		if err != nil {
			return nil, fmt.Errorf("Failed to get client TLS config: %s", err)
		}

		tr.TLSClientConfig = tlsConfig
	}

	httpClient := &http.Client{Transport: tr}
	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("POST failure [%s]; not sending\n%s", err, reqStr)
	}
	var respBody []byte
	if resp.Body != nil {
		respBody, err = ioutil.ReadAll(resp.Body)
		defer resp.Body.Close()
		if err != nil {
			return nil, fmt.Errorf("Failed to read response [%s] of request:\n%s", err, reqStr)
		}
		log.Debugf("Received response\n%s", util.HTTPResponseToString(resp))
	}
	var body *cfsslapi.Response
	if respBody != nil && len(respBody) > 0 {
		body = new(cfsslapi.Response)
		err = json.Unmarshal(respBody, body)
		if err != nil {
			return nil, fmt.Errorf("Failed to parse response [%s] for request:\n%s", err, reqStr)
		}
		if len(body.Errors) > 0 {
			msg := body.Errors[0].Message
			return nil, fmt.Errorf("Error response from server was '%s' for request:\n%s", msg, reqStr)
		}
	}
	scode := resp.StatusCode
	if scode >= 400 {
		return nil, fmt.Errorf("Failed with server status code %d for request:\n%s", scode, reqStr)
	}
	if body == nil {
		return nil, nil
	}
	if !body.Success {
		return nil, fmt.Errorf("Server returned failure for request:\n%s", reqStr)
	}
	log.Debugf("Response body result: %+v", body.Result)
	return body.Result, nil
}

func (c *Client) getURL(endpoint string) (string, error) {
	nurl, err := NormalizeURL(c.Config.URL)
	if err != nil {
		return "", err
	}
	rtn := fmt.Sprintf("%s/api/v1/cfssl/%s", nurl, endpoint)
	return rtn, nil
}

// Enrollment checks to see if client is enrolled (i.e. enrollment information exists)
func (c *Client) Enrollment() error {
	if !util.FileExists(c.GetMyCertFile()) || !util.FileExists(c.GetMyKeyFile()) {
		return errors.New("Enrollment information does not exist. Please execute enroll command first. Example: fabric-ca-client enroll -u http://user:userpw@serverAddr:serverPort")
	}
	return nil
}

func (c *Client) getClientConfig(path string) ([]byte, error) {
	log.Debug("Retrieving client config")
	// fcaClient := filepath.Join(path, clientConfigFile)
	fileBytes, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}
	return fileBytes, nil
}

// NormalizeURL normalizes a URL (from cfssl)
func NormalizeURL(addr string) (*url.URL, error) {
	addr = strings.TrimSpace(addr)
	u, err := url.Parse(addr)
	if err != nil {
		return nil, err
	}
	if u.Opaque != "" {
		u.Host = net.JoinHostPort(u.Scheme, u.Opaque)
		u.Opaque = ""
	} else if u.Path != "" && !strings.Contains(u.Path, ":") {
		u.Host = net.JoinHostPort(u.Path, util.GetServerPort())
		u.Path = ""
	} else if u.Scheme == "" {
		u.Host = u.Path
		u.Path = ""
	}
	if u.Scheme != "https" {
		u.Scheme = "http"
	}
	_, port, err := net.SplitHostPort(u.Host)
	if err != nil {
		_, port, err = net.SplitHostPort(u.Host + ":" + util.GetServerPort())
		if err != nil {
			return nil, err
		}
	}
	if port != "" {
		_, err = strconv.Atoi(port)
		if err != nil {
			return nil, err
		}
	}
	return u, nil
}
