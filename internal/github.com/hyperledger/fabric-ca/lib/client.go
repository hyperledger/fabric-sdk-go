/*
Copyright IBM Corp. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/
/*
Notice: This file has been modified for Hyperledger Fabric SDK Go usage.
Please review third_party pinning scripts and patches for more details.
*/

package lib

import (
	"bytes"
	"encoding/json"
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
	proto "github.com/golang/protobuf/proto"
	fp256bn "github.com/hyperledger/fabric-amcl/amcl/FP256BN"
	"github.com/hyperledger/fabric-ca/api"
	"github.com/hyperledger/fabric-ca/lib/client/credential"
	idemixcred "github.com/hyperledger/fabric-ca/lib/client/credential/idemix"
	x509cred "github.com/hyperledger/fabric-ca/lib/client/credential/x509"
	"github.com/hyperledger/fabric-ca/lib/common"
	"github.com/hyperledger/fabric-ca/lib/streamer"
	"github.com/hyperledger/fabric-ca/lib/tls"
	"github.com/hyperledger/fabric-ca/util"
	"github.com/hyperledger/fabric/bccsp"
	"github.com/hyperledger/fabric/idemix"
	"github.com/mitchellh/mapstructure"
	"github.com/pkg/errors"
)

// Client is the fabric-ca client object
type Client struct {
	// The client's home directory
	HomeDir string `json:"homeDir,omitempty"`
	// The client's configuration
	Config *ClientConfig
	// Denotes if the client object is already initialized
	initialized bool
	// File and directory paths
	keyFile, certFile, idemixCredFile, idemixCredsDir, ipkFile, caCertsDir string
	// The crypto service provider (BCCSP)
	csp bccsp.BCCSP
	// HTTP client associated with this Fabric CA client
	httpClient *http.Client
	// Public key of Idemix issuer
	issuerPublicKey *idemix.IssuerPublicKey
}

// GetCAInfoResponse is the response from the GetCAInfo call
type GetCAInfoResponse struct {
	// CAName is the name of the CA
	CAName string
	// CAChain is the PEM-encoded bytes of the fabric-ca-server's CA chain.
	// The 1st element of the chain is the root CA cert
	CAChain []byte
	// Idemix issuer public key of the CA
	IssuerPublicKey []byte
	// Idemix issuer revocation public key of the CA
	IssuerRevocationPublicKey []byte
	// Version of the server
	Version string
}

// EnrollmentResponse is the response from Client.Enroll and Identity.Reenroll
type EnrollmentResponse struct {
	Identity *Identity
	CAInfo   GetCAInfoResponse
}

// Init initializes the client
func (c *Client) Init() error {
	if !c.initialized {
		cfg := c.Config
		log.Debugf("Initializing client with config: %+v", cfg)
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
			return errors.Wrap(err, "Failed to create keystore directory")
		}
		c.keyFile = path.Join(keyDir, "key.pem")

		// Cert directory and file
		certDir := path.Join(mspDir, "signcerts")
		err = os.MkdirAll(certDir, 0755)
		if err != nil {
			return errors.Wrap(err, "Failed to create signcerts directory")
		}
		c.certFile = path.Join(certDir, "cert.pem")

		// CA certs directory
		c.caCertsDir = path.Join(mspDir, "cacerts")
		err = os.MkdirAll(c.caCertsDir, 0755)
		if err != nil {
			return errors.Wrap(err, "Failed to create cacerts directory")
		}

		// CA's Idemix public key
		c.ipkFile = filepath.Join(mspDir, "IssuerPublicKey")

		// Idemix credentials directory
		c.idemixCredsDir = path.Join(mspDir, "user")
		err = os.MkdirAll(c.idemixCredsDir, 0755)
		if err != nil {
			return errors.Wrap(err, "Failed to create Idemix credentials directory 'user'")
		}
		c.idemixCredFile = path.Join(c.idemixCredsDir, "SignerConfig")

		// Initialize BCCSP (the crypto layer)
		c.csp, err = util.InitBCCSP(&cfg.CSP, mspDir, c.HomeDir)
		if err != nil {
			return err
		}
		// Create http.Client object and associate it with this client
		err = c.initHTTPClient()
		if err != nil {
			return err
		}

		// Successfully initialized the client
		c.initialized = true
	}
	return nil
}

func (c *Client) initHTTPClient() error {
	tr := new(http.Transport)
	if c.Config.TLS.Enabled {
		log.Info("TLS Enabled")

		err := tls.AbsTLSClient(&c.Config.TLS, c.HomeDir)
		if err != nil {
			return err
		}

		tlsConfig, err2 := tls.GetClientTLSConfig(&c.Config.TLS, c.csp)
		if err2 != nil {
			return fmt.Errorf("Failed to get client TLS config: %s", err2)
		}
		// set the default ciphers
		tlsConfig.CipherSuites = tls.DefaultCipherSuites
		tr.TLSClientConfig = tlsConfig
	}
	c.httpClient = &http.Client{Transport: tr}
	return nil
}

// GetCAInfo returns generic CA information
func (c *Client) GetCAInfo(req *api.GetCAInfoRequest) (*GetCAInfoResponse, error) {
	err := c.Init()
	if err != nil {
		return nil, err
	}
	body, err := util.Marshal(req, "GetCAInfo")
	if err != nil {
		return nil, err
	}
	cainforeq, err := c.newPost("cainfo", body)
	if err != nil {
		return nil, err
	}
	netSI := &common.CAInfoResponseNet{}
	err = c.SendReq(cainforeq, netSI)
	if err != nil {
		return nil, err
	}
	localSI := &GetCAInfoResponse{}
	err = c.net2LocalCAInfo(netSI, localSI)
	if err != nil {
		return nil, err
	}
	return localSI, nil
}

// GenCSR generates a CSR (Certificate Signing Request)
func (c *Client) GenCSR(req *api.CSRInfo, id string) ([]byte, bccsp.Key, error) {
	log.Debugf("GenCSR %+v", req)

	err := c.Init()
	if err != nil {
		return nil, nil, err
	}

	cr := c.newCertificateRequest(req)
	cr.CN = id

	if (cr.KeyRequest == nil) || (cr.KeyRequest.Size() == 0 && cr.KeyRequest.Algo() == "") {
		cr.KeyRequest = newCfsslBasicKeyRequest(api.NewBasicKeyRequest())
	}

	key, cspSigner, err := util.BCCSPKeyRequestGenerate(cr, c.csp)
	if err != nil {
		log.Debugf("failed generating BCCSP key: %s", err)
		return nil, nil, err
	}

	csrPEM, err := csr.Generate(cspSigner, cr)
	if err != nil {
		log.Debugf("failed generating CSR: %s", err)
		return nil, nil, err
	}

	return csrPEM, key, nil
}

// Enroll enrolls a new identity
// @param req The enrollment request
func (c *Client) Enroll(req *api.EnrollmentRequest) (*EnrollmentResponse, error) {
	log.Debugf("Enrolling %+v", req)

	err := c.Init()
	if err != nil {
		return nil, err
	}

	if strings.ToLower(req.Type) == "idemix" {
		return c.handleIdemixEnroll(req)
	}
	return c.handleX509Enroll(req)
}

// Convert from network to local CA information
func (c *Client) net2LocalCAInfo(net *common.CAInfoResponseNet, local *GetCAInfoResponse) error {
	caChain, err := util.B64Decode(net.CAChain)
	if err != nil {
		return errors.WithMessage(err, "Failed to decode CA chain")
	}
	if net.IssuerPublicKey != "" {
		ipk, err := util.B64Decode(net.IssuerPublicKey)
		if err != nil {
			return errors.WithMessage(err, "Failed to decode issuer public key")
		}
		local.IssuerPublicKey = ipk
	}
	if net.IssuerRevocationPublicKey != "" {
		rpk, err := util.B64Decode(net.IssuerRevocationPublicKey)
		if err != nil {
			return errors.WithMessage(err, "Failed to decode issuer revocation key")
		}
		local.IssuerRevocationPublicKey = rpk
	}
	local.CAName = net.CAName
	local.CAChain = caChain
	local.Version = net.Version
	return nil
}

func (c *Client) handleX509Enroll(req *api.EnrollmentRequest) (*EnrollmentResponse, error) {
	// Generate the CSR
	csrPEM, key, err := c.GenCSR(req.CSR, req.Name)
	if err != nil {
		return nil, errors.WithMessage(err, "Failure generating CSR")
	}

	reqNet := &api.EnrollmentRequestNet{
		CAName:   req.CAName,
		AttrReqs: req.AttrReqs,
	}

	if req.CSR != nil {
		reqNet.SignRequest.Hosts = req.CSR.Hosts
	}
	reqNet.SignRequest.Request = string(csrPEM)
	reqNet.SignRequest.Profile = req.Profile
	reqNet.SignRequest.Label = req.Label

	body, err := util.Marshal(reqNet, "SignRequest")
	if err != nil {
		return nil, err
	}

	// Send the CSR to the fabric-ca server with basic auth header
	post, err := c.newPost("enroll", body)
	if err != nil {
		return nil, err
	}
	post.SetBasicAuth(req.Name, req.Secret)
	var result common.EnrollmentResponseNet
	err = c.SendReq(post, &result)
	if err != nil {
		return nil, err
	}

	// Create the enrollment response
	return c.newEnrollmentResponse(&result, req.Name, key)
}

// Handles enrollment request for an Idemix credential
// 1. Sends a request with empty body to the /api/v1/idemix/credentail REST endpoint
//    of the server to get a Nonce from the CA
// 2. Constructs a credential request using the nonce, CA's idemix public key
// 3. Sends a request with the CredentialRequest object in the body to the
//    /api/v1/idemix/credentail REST endpoint to get a credential
func (c *Client) handleIdemixEnroll(req *api.EnrollmentRequest) (*EnrollmentResponse, error) {
	log.Debugf("Getting nonce from CA %s", req.CAName)
	reqNet := &api.IdemixEnrollmentRequestNet{
		CAName: req.CAName,
	}
	var identity *Identity

	// Get nonce from the CA
	body, err := util.Marshal(reqNet, "NonceRequest")
	if err != nil {
		return nil, errors.WithMessage(err, "Failed to marshal nonce request")
	}
	post, err := c.newPost("idemix/credential", body)
	if err != nil {
		return nil, errors.WithMessage(err, "Failed to create HTTP request for getting a nonce")
	}
	err = c.addAuthHeaderForIdemixEnroll(req, identity, body, post)
	if err != nil {
		return nil, errors.WithMessage(err,
			"Either username/password or X509 enrollment certificate is required to request an Idemix credential")
	}

	// Send the request and process the response
	var result common.IdemixEnrollmentResponseNet
	err = c.SendReq(post, &result)
	if err != nil {
		return nil, err
	}
	nonceBytes, err := util.B64Decode(result.Nonce)

	if err != nil {
		return nil, errors.WithMessage(err,
			fmt.Sprintf("Failed to decode nonce that was returned by CA %s", req.CAName))
	}
	nonce := fp256bn.FromBytes(nonceBytes)
	log.Infof("Successfully got nonce from CA %s", req.CAName)

	ipkBytes := []byte{}
	ipkBytes, err = util.B64Decode(result.CAInfo.IssuerPublicKey)
	if err != nil {
		return nil, errors.WithMessage(err, fmt.Sprintf("Failed to decode issuer public key that was returned by CA %s", req.CAName))
	}
	// Create credential request
	credReq, sk, err := c.newIdemixCredentialRequest(nonce, ipkBytes)
	if err != nil {
		return nil, errors.WithMessage(err, "Failed to create an Idemix credential request")
	}
	reqNet.CredRequest = credReq
	log.Info("Successfully created an Idemix credential request")

	body, err = util.Marshal(reqNet, "CredentialRequest")
	if err != nil {
		return nil, errors.WithMessage(err, "Failed to marshal Idemix credential request")
	}

	// Send the cred request to the CA
	post, err = c.newPost("idemix/credential", body)
	if err != nil {
		return nil, errors.WithMessage(err, "Failed to create HTTP request for getting Idemix credential")
	}
	err = c.addAuthHeaderForIdemixEnroll(req, identity, body, post)
	if err != nil {
		return nil, errors.WithMessage(err,
			"Either username/password or X509 enrollment certificate is required to request idemix credential")
	}
	err = c.SendReq(post, &result)
	if err != nil {
		return nil, err
	}
	log.Infof("Successfully received Idemix credential from CA %s", req.CAName)
	return c.newIdemixEnrollmentResponse(identity, &result, sk, req.Name)
}

// addAuthHeaderForIdemixEnroll adds authenticate header to the specified HTTP request
// It adds basic authentication header if userName and password are specified in the
// specified EnrollmentRequest object. Else, checks if a X509 credential in the client's
// MSP directory, if so, loads the identity, creates an oauth token based on the loaded
// identity's X509 credential, and adds the token to the HTTP request. The loaded
// identity is passed back to the caller.
func (c *Client) addAuthHeaderForIdemixEnroll(req *api.EnrollmentRequest, id *Identity,
	body []byte, post *http.Request) error {
	if req.Name != "" && req.Secret != "" {
		post.SetBasicAuth(req.Name, req.Secret)
		return nil
	}
	if id == nil {
		err := c.checkX509Enrollment()
		if err != nil {
			return err
		}
		id, err = c.LoadMyIdentity()
		if err != nil {
			return err
		}
	}
	err := id.addTokenAuthHdr(post, body)
	if err != nil {
		return err
	}
	return nil
}

// newEnrollmentResponse creates a client enrollment response from a network response
// @param result The result from server
// @param id Name of identity being enrolled or reenrolled
// @param key The private key which was used to sign the request
func (c *Client) newEnrollmentResponse(result *common.EnrollmentResponseNet, id string, key bccsp.Key) (*EnrollmentResponse, error) {
	log.Debugf("newEnrollmentResponse %s", id)
	certByte, err := util.B64Decode(result.Cert)
	if err != nil {
		return nil, errors.WithMessage(err, "Invalid response format from server")
	}
	signer, err := x509cred.NewSigner(key, certByte)
	if err != nil {
		return nil, err
	}
	x509Cred := x509cred.NewCredential(c.certFile, c.keyFile, c)
	err = x509Cred.SetVal(signer)
	if err != nil {
		return nil, err
	}
	resp := &EnrollmentResponse{
		Identity: NewIdentity(c, id, []credential.Credential{x509Cred}),
	}
	err = c.net2LocalCAInfo(&result.ServerInfo, &resp.CAInfo)
	if err != nil {
		return nil, err
	}
	return resp, nil
}

// newIdemixEnrollmentResponse creates a client idemix enrollment response from a network response
func (c *Client) newIdemixEnrollmentResponse(identity *Identity, result *common.IdemixEnrollmentResponseNet,
	sk *fp256bn.BIG, id string) (*EnrollmentResponse, error) {
	log.Debugf("newIdemixEnrollmentResponse %s", id)
	credBytes, err := util.B64Decode(result.Credential)
	if err != nil {
		return nil, errors.WithMessage(err, "Invalid response format from server")
	}

	criBytes, err := util.B64Decode(result.CRI)
	if err != nil {
		return nil, errors.WithMessage(err, "Invalid response format from server")
	}

	// Create SignerConfig object with credential bytes from the response
	// and secret key
	role, _ := result.Attrs["Role"].(int)
	ou, _ := result.Attrs["OU"].(string)
	enrollmentID, _ := result.Attrs["EnrollmentID"].(string)
	signerConfig := &idemixcred.SignerConfig{
		Cred:                            credBytes,
		Sk:                              idemix.BigToBytes(sk),
		Role:                            role,
		OrganizationalUnitIdentifier:    ou,
		EnrollmentID:                    enrollmentID,
		CredentialRevocationInformation: criBytes,
	}

	// Create IdemixCredential object
	cred := idemixcred.NewCredential(c.idemixCredFile, c)
	err = cred.SetVal(signerConfig)
	if err != nil {
		return nil, err
	}
	if identity == nil {
		identity = NewIdentity(c, id, []credential.Credential{cred})
	} else {
		identity.creds = append(identity.creds, cred)
	}

	resp := &EnrollmentResponse{
		Identity: identity,
	}
	err = c.net2LocalCAInfo(&result.CAInfo, &resp.CAInfo)
	if err != nil {
		return nil, err
	}
	log.Infof("Successfully processed response from the CA")
	return resp, nil
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
		cr.KeyRequest = newCfsslBasicKeyRequest(req.KeyRequest)
	}
	if req != nil {
		cr.CA = req.CA
		cr.SerialNumber = req.SerialNumber
	}
	return &cr
}

// newIdemixCredentialRequest returns CredentialRequest object, a secret key, and a random number used in
// the creation of credential request.
func (c *Client) newIdemixCredentialRequest(nonce *fp256bn.BIG, ipkBytes []byte) (*idemix.CredRequest, *fp256bn.BIG, error) {
	rng, err := idemix.GetRand()
	if err != nil {
		return nil, nil, err
	}
	sk := idemix.RandModOrder(rng)

	issuerPubKey, err := c.getIssuerPubKey(ipkBytes)
	if err != nil {
		return nil, nil, err
	}
	return idemix.NewCredRequest(sk, idemix.BigToBytes(nonce), issuerPubKey, rng), sk, nil
}

func (c *Client) getIssuerPubKey(ipkBytes []byte) (*idemix.IssuerPublicKey, error) {
	var err error
	if ipkBytes == nil || len(ipkBytes) == 0 {
		ipkBytes, err = ioutil.ReadFile(c.ipkFile)
		if err != nil {
			return nil, errors.Wrapf(err, "Error reading CA's Idemix public key at '%s'", c.ipkFile)
		}
	}
	pubKey := &idemix.IssuerPublicKey{}
	err = proto.Unmarshal(ipkBytes, pubKey)
	if err != nil {
		return nil, err
	}
	c.issuerPublicKey = pubKey
	return c.issuerPublicKey, nil
}

// LoadMyIdentity loads the client's identity from disk
func (c *Client) LoadMyIdentity() (*Identity, error) {
	err := c.Init()
	if err != nil {
		return nil, err
	}
	return c.LoadIdentity(c.keyFile, c.certFile, c.idemixCredFile)
}

// LoadIdentity loads an identity from disk
func (c *Client) LoadIdentity(keyFile, certFile, idemixCredFile string) (*Identity, error) {
	log.Debugf("Loading identity: keyFile=%s, certFile=%s", keyFile, certFile)
	err := c.Init()
	if err != nil {
		return nil, err
	}

	var creds []credential.Credential
	var x509Found, idemixFound bool
	x509Cred := x509cred.NewCredential(certFile, keyFile, c)
	err = x509Cred.Load()
	if err == nil {
		x509Found = true
		creds = append(creds, x509Cred)
	} else {
		log.Debugf("No X509 credential found at %s, %s", keyFile, certFile)
	}

	idemixCred := idemixcred.NewCredential(idemixCredFile, c)
	err = idemixCred.Load()
	if err == nil {
		idemixFound = true
		creds = append(creds, idemixCred)
	} else {
		log.Debugf("No Idemix credential found at %s", idemixCredFile)
	}

	if !x509Found && !idemixFound {
		return nil, errors.New("Identity does not posses any enrollment credentials")
	}

	return c.NewIdentity(creds)
}

// NewIdentity creates a new identity
func (c *Client) NewIdentity(creds []credential.Credential) (*Identity, error) {
	if len(creds) == 0 {
		return nil, errors.New("No credentials spcified. Atleast one credential must be specified")
	}
	name, err := creds[0].EnrollmentID()
	if err != nil {
		return nil, err
	}
	if len(creds) == 1 {
		return NewIdentity(c, name, creds), nil
	}

	//TODO: Get the enrollment ID from the creds...they all should return same value
	// for i := 1; i < len(creds); i++ {
	// 	localid, err := creds[i].EnrollmentID()
	// 	if err != nil {
	// 		return nil, err
	// 	}
	// 	if localid != name {
	// 		return nil, errors.New("Specified credentials belong to different identities, they should be long to same identity")
	// 	}
	// }
	return NewIdentity(c, name, creds), nil
}

// NewX509Identity creates a new identity
func (c *Client) NewX509Identity(name string, creds []credential.Credential) x509cred.Identity {
	return NewIdentity(c, name, creds)
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

// GetCertFilePath returns the path to the certificate file for this client
func (c *Client) GetCertFilePath() string {
	return c.certFile
}

// GetCSP returns BCCSP instance associated with this client
func (c *Client) GetCSP() bccsp.BCCSP {
	return c.csp
}

// GetIssuerPubKey returns issuer public key associated with this client
func (c *Client) GetIssuerPubKey() (*idemix.IssuerPublicKey, error) {
	if c.issuerPublicKey == nil {
		return c.getIssuerPubKey(nil)
	}
	return c.issuerPublicKey, nil
}

// newGet create a new GET request
func (c *Client) newGet(endpoint string) (*http.Request, error) {
	curl, err := c.getURL(endpoint)
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequest("GET", curl, bytes.NewReader([]byte{}))
	if err != nil {
		return nil, errors.Wrapf(err, "Failed creating GET request for %s", curl)
	}
	return req, nil
}

// newPut create a new PUT request
func (c *Client) newPut(endpoint string, reqBody []byte) (*http.Request, error) {
	curl, err := c.getURL(endpoint)
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequest("PUT", curl, bytes.NewReader(reqBody))
	if err != nil {
		return nil, errors.Wrapf(err, "Failed creating PUT request for %s", curl)
	}
	return req, nil
}

// newDelete create a new DELETE request
func (c *Client) newDelete(endpoint string) (*http.Request, error) {
	curl, err := c.getURL(endpoint)
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequest("DELETE", curl, bytes.NewReader([]byte{}))
	if err != nil {
		return nil, errors.Wrapf(err, "Failed creating DELETE request for %s", curl)
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
		return nil, errors.Wrapf(err, "Failed posting to %s", curl)
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

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return errors.Wrapf(err, "%s failure of request: %s", req.Method, reqStr)
	}
	var respBody []byte
	if resp.Body != nil {
		respBody, err = ioutil.ReadAll(resp.Body)
		defer func() {
			err := resp.Body.Close()
			if err != nil {
				log.Debugf("Failed to close the response body: %s", err.Error())
			}
		}()
		if err != nil {
			return errors.Wrapf(err, "Failed to read response of request: %s", reqStr)
		}
		log.Debugf("Received response\n%s", util.HTTPResponseToString(resp))
	}
	var body *cfsslapi.Response
	if respBody != nil && len(respBody) > 0 {
		body = new(cfsslapi.Response)
		err = json.Unmarshal(respBody, body)
		if err != nil {
			return errors.Wrapf(err, "Failed to parse response: %s", respBody)
		}
		if len(body.Errors) > 0 {
			var errorMsg string
			for _, err := range body.Errors {
				msg := fmt.Sprintf("Response from server: Error Code: %d - %s\n", err.Code, err.Message)
				if errorMsg == "" {
					errorMsg = msg
				} else {
					errorMsg = errorMsg + fmt.Sprintf("\n%s", msg)
				}
			}
			return errors.Errorf(errorMsg)
		}
	}
	scode := resp.StatusCode
	if scode >= 400 {
		return errors.Errorf("Failed with server status code %d for request:\n%s", scode, reqStr)
	}
	if body == nil {
		return errors.Errorf("Empty response body:\n%s", reqStr)
	}
	if !body.Success {
		return errors.Errorf("Server returned failure for request:\n%s", reqStr)
	}
	log.Debugf("Response body result: %+v", body.Result)
	if result != nil {
		return mapstructure.Decode(body.Result, result)
	}
	return nil
}

// StreamResponse reads the response as it comes back from the server
func (c *Client) StreamResponse(req *http.Request, stream string, cb func(*json.Decoder) error) (err error) {

	reqStr := util.HTTPRequestToString(req)
	log.Debugf("Sending request\n%s", reqStr)

	err = c.Init()
	if err != nil {
		return err
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return errors.Wrapf(err, "%s failure of request: %s", req.Method, reqStr)
	}
	defer resp.Body.Close()

	dec := json.NewDecoder(resp.Body)
	results, err := streamer.StreamJSONArray(dec, stream, cb)
	if err != nil {
		return err
	}
	if !results {
		fmt.Println("No results returned")
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
	var x509Enrollment, idemixEnrollment bool
	err = c.checkX509Enrollment()
	if err == nil {
		x509Enrollment = true
	}
	err = c.checkIdemixEnrollment()
	if err == nil {
		idemixEnrollment = true
	}
	if x509Enrollment || idemixEnrollment {
		return nil
	}
	log.Errorf("Enrollment check failed: %s", err.Error())
	return errors.New("Enrollment information does not exist. Please execute enroll command first. Example: fabric-ca-client enroll -u http://user:userpw@serverAddr:serverPort")
}

func (c *Client) checkX509Enrollment() error {
	keyFileExists := util.FileExists(c.keyFile)
	certFileExists := util.FileExists(c.certFile)
	if keyFileExists && certFileExists {
		return nil
	}
	// If key file does not exist, but certFile does, key file is probably
	// stored by bccsp, so check to see if this is the case
	if certFileExists {
		_, _, _, err := util.GetSignerFromCertFile(c.certFile, c.csp)
		if err == nil {
			// Yes, the key is stored by BCCSP
			return nil
		}
	}
	return errors.New("X509 enrollment information does not exist")
}

// checkIdemixEnrollment returns an error if CA's Idemix public key and user's
// Idemix credential does not exist and if they exist and credential verification
// fails. Returns nil if the credential verification suucceeds
func (c *Client) checkIdemixEnrollment() error {
	log.Debugf("CheckIdemixEnrollment - ipkFile: %s, idemixCredFrile: %s", c.ipkFile, c.idemixCredFile)

	idemixIssuerPubKeyExists := util.FileExists(c.ipkFile)
	idemixCredExists := util.FileExists(c.idemixCredFile)
	if idemixIssuerPubKeyExists && idemixCredExists {
		err := c.verifyIdemixCredential()
		if err != nil {
			return errors.WithMessage(err, "Idemix enrollment check failed")
		}
		return nil
	}
	return errors.New("Idemix enrollment information does not exist")
}

func (c *Client) verifyIdemixCredential() error {
	ipk, err := c.getIssuerPubKey(nil)
	if err != nil {
		return err
	}
	credfileBytes, err := util.ReadFile(c.idemixCredFile)
	if err != nil {
		return errors.Wrapf(err, "Failed to read %s", c.idemixCredFile)
	}
	signerConfig := &idemixcred.SignerConfig{}
	err = json.Unmarshal(credfileBytes, signerConfig)
	if err != nil {
		return errors.Wrapf(err, "Failed to unmarshal signer config from %s", c.idemixCredFile)
	}

	cred := new(idemix.Credential)
	err = proto.Unmarshal(signerConfig.GetCred(), cred)
	if err != nil {
		return errors.Wrap(err, "Failed to unmarshal Idemix credential from signer config")
	}
	sk := fp256bn.FromBytes(signerConfig.GetSk())

	// Verify that the credential is cryptographically valid
	err = cred.Ver(sk, ipk)
	if err != nil {
		return errors.Wrap(err, "Idemix credential is not cryptographically valid")
	}
	return nil
}

func newCfsslBasicKeyRequest(bkr *api.BasicKeyRequest) *csr.BasicKeyRequest {
	return &csr.BasicKeyRequest{A: bkr.Algo, S: bkr.Size}
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
