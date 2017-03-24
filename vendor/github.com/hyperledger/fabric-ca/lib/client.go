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
	"github.com/hyperledger/fabric/bccsp/factory"
	"github.com/mitchellh/mapstructure"
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
	Config                        *ClientConfig
	initialized                   bool
	keyFile, certFile, caCertsDir string
}

// Init initializes the client
func (c *Client) Init() error {
	if !c.initialized {
		cfg := c.Config
		if cfg.MSPDir == "" {
			cfg.MSPDir = "msp"
		}
		mspDir, err := util.MakeFileAbs(cfg.MSPDir, c.HomeDir)
		if err != nil {
			return err
		}
		cfg.MSPDir = mspDir
		// Key directory and file
		keyDir := path.Join(mspDir, "keystore")
		err = os.MkdirAll(keyDir, 0700)
		if err != nil {
			return fmt.Errorf("Failed to create keystore directory: %s", err)
		}
		c.keyFile = path.Join(keyDir, "key.pem")
		// Cert directory and file
		certDir := path.Join(mspDir, "signcerts")
		err = os.MkdirAll(certDir, 0755)
		if err != nil {
			return fmt.Errorf("Failed to create signcerts directory: %s", err)
		}
		c.certFile = path.Join(certDir, "cert.pem")
		// CA certs directory
		c.caCertsDir = path.Join(mspDir, "cacerts")
		err = os.MkdirAll(c.caCertsDir, 0755)
		if err != nil {
			return fmt.Errorf("Failed to create cacerts directory: %s", err)
		}
		err = factory.InitFactories(nil)
		if err != nil {
			return fmt.Errorf("Failed to initialize BCCSP: %s", err)
		}
		c.initialized = true
	}
	return nil
}

// GetServerInfoResponse is the response from the GetServerInfo call
type GetServerInfoResponse struct {
	// CAName is the name of the CA
	CAName string
	// CAChain is the PEM-encoded bytes of the fabric-ca-server's CA chain.
	// The 1st element of the chain is the root CA cert
	CAChain []byte
}

// GetServerInfo returns generic server information
func (c *Client) GetServerInfo() (*GetServerInfoResponse, error) {
	err := c.Init()
	if err != nil {
		return nil, err
	}
	req, err := c.newGet("info")
	if err != nil {
		return nil, err
	}
	netSI := &serverInfoResponseNet{}
	err = c.SendReq(req, netSI)
	if err != nil {
		return nil, err
	}
	localSI := &GetServerInfoResponse{}
	err = c.net2LocalServerInfo(netSI, localSI)
	if err != nil {
		return nil, err
	}
	return localSI, nil
}

// Convert from network to local server information
func (c *Client) net2LocalServerInfo(net *serverInfoResponseNet, local *GetServerInfoResponse) error {
	caChain, err := util.B64Decode(net.CAChain)
	if err != nil {
		return err
	}
	local.CAName = net.CAName
	local.CAChain = caChain
	return nil
}

// EnrollmentResponse is the response from Client.Enroll and Identity.Reenroll
type EnrollmentResponse struct {
	Identity   *Identity
	ServerInfo GetServerInfoResponse
}

// Enroll enrolls a new identity
// @param req The enrollment request
func (c *Client) Enroll(req *api.EnrollmentRequest) (*EnrollmentResponse, error) {
	log.Debugf("Enrolling %+v", &req)

	err := c.Init()
	if err != nil {
		return nil, err
	}

	// Generate the CSR
	csrPEM, key, err := c.GenCSR(req.CSR, req.Name)
	if err != nil {
		log.Debugf("Enroll failure generating CSR: %s", err)
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
	post, err := c.newPost("enroll", body)
	if err != nil {
		return nil, err
	}
	post.SetBasicAuth(req.Name, req.Secret)
	var result enrollmentResponseNet
	err = c.SendReq(post, &result)
	if err != nil {
		return nil, err
	}

	// Create the enrollment response
	return c.newEnrollmentResponse(&result, req.Name, key)
}

// newEnrollmentResponse creates a client enrollment response from a network response
// @param result The result from server
// @param id Name of identity being enrolled or reenrolled
// @param key The private key which was used to sign the request
func (c *Client) newEnrollmentResponse(result *enrollmentResponseNet, id string, key []byte) (*EnrollmentResponse, error) {
	log.Debugf("newEnrollmentResponse %s", id)
	certByte, err := util.B64Decode(result.Cert)
	if err != nil {
		return nil, fmt.Errorf("Invalid response format from server: %s", err)
	}
	resp := &EnrollmentResponse{
		Identity: newIdentity(c, id, key, certByte),
	}
	err = c.net2LocalServerInfo(&result.ServerInfo, &resp.ServerInfo)
	if err != nil {
		return nil, err
	}
	return resp, nil
}

// GenCSR generates a CSR (Certificate Signing Request)
func (c *Client) GenCSR(req *api.CSRInfo, id string) ([]byte, []byte, error) {
	log.Debugf("GenCSR %+v", &req)

	err := c.Init()
	if err != nil {
		return nil, nil, err
	}

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
	err := c.Init()
	if err != nil {
		return nil, err
	}
	return c.LoadIdentity(c.keyFile, c.certFile)
}

// StoreMyIdentity stores my identity to disk
func (c *Client) StoreMyIdentity(key, cert []byte) error {
	err := c.Init()
	if err != nil {
		return err
	}
	err = util.WriteFile(c.keyFile, key, 0600)
	if err != nil {
		return fmt.Errorf("Failed to store my key: %s", err)
	}
	log.Infof("Stored client key at %s", c.keyFile)
	err = util.WriteFile(c.certFile, cert, 0644)
	if err != nil {
		return fmt.Errorf("Failed to store my certificate: %s", err)
	}
	log.Infof("Stored client certificate at %s", c.certFile)
	return nil
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

// NewGet create a new GET request
func (c *Client) newGet(endpoint string) (*http.Request, error) {
	curl, err := c.getURL(endpoint)
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequest("GET", curl, bytes.NewReader([]byte{}))
	if err != nil {
		return nil, fmt.Errorf("Failed creating GET request for %s: %s", curl, err)
	}
	return req, nil
}

// NewPost create a new post request
func (c *Client) newPost(endpoint string, reqBody []byte) (*http.Request, error) {
	curl, err := c.getURL(endpoint)
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequest("POST", curl, bytes.NewReader(reqBody))
	if err != nil {
		return nil, fmt.Errorf("Failed posting to %s: %s", curl, err)
	}
	return req, nil
}

// SendReq sends a request to the fabric-ca-server and fills in the result
func (c *Client) SendReq(req *http.Request, result interface{}) (err error) {

	reqStr := util.HTTPRequestToString(req)
	log.Debugf("Sending request\n%s", reqStr)

	err = c.Init()
	if err != nil {
		return err
	}

	var tr = new(http.Transport)

	if c.Config.TLS.Enabled {
		log.Info("TLS Enabled")

		err = tls.AbsTLSClient(&c.Config.TLS, c.HomeDir)
		if err != nil {
			return err
		}

		tlsConfig, err2 := tls.GetClientTLSConfig(&c.Config.TLS)
		if err2 != nil {
			return fmt.Errorf("Failed to get client TLS config: %s", err2)
		}

		tr.TLSClientConfig = tlsConfig
	}

	httpClient := &http.Client{Transport: tr}
	resp, err := httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("POST failure [%s]; not sending\n%s", err, reqStr)
	}
	var respBody []byte
	if resp.Body != nil {
		respBody, err = ioutil.ReadAll(resp.Body)
		defer resp.Body.Close()
		if err != nil {
			return fmt.Errorf("Failed to read response [%s] of request:\n%s", err, reqStr)
		}
		log.Debugf("Received response\n%s", util.HTTPResponseToString(resp))
	}
	var body *cfsslapi.Response
	if respBody != nil && len(respBody) > 0 {
		body = new(cfsslapi.Response)
		err = json.Unmarshal(respBody, body)
		if err != nil {
			return fmt.Errorf("Failed to parse response [%s] for request:\n%s", err, reqStr)
		}
		if len(body.Errors) > 0 {
			msg := body.Errors[0].Message
			return fmt.Errorf("Error response from server was: %s", msg)
		}
	}
	scode := resp.StatusCode
	if scode >= 400 {
		return fmt.Errorf("Failed with server status code %d for request:\n%s", scode, reqStr)
	}
	if body == nil {
		return fmt.Errorf("Empty response body:\n%s", reqStr)
	}
	if !body.Success {
		return fmt.Errorf("Server returned failure for request:\n%s", reqStr)
	}
	log.Debugf("Response body result: %+v", body.Result)
	if result != nil {
		return mapstructure.Decode(body.Result, result)
	}
	return nil
}

func (c *Client) getURL(endpoint string) (string, error) {
	nurl, err := NormalizeURL(c.Config.URL)
	if err != nil {
		return "", err
	}
	rtn := fmt.Sprintf("%s/%s", nurl, endpoint)
	return rtn, nil
}

// CheckEnrollment returns an error if this client is not enrolled
func (c *Client) CheckEnrollment() error {
	err := c.Init()
	if err != nil {
		return err
	}
	if !util.FileExists(c.certFile) || !util.FileExists(c.keyFile) {
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
