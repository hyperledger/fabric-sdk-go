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
	"crypto/x509"
	"encoding/asn1"
	"encoding/pem"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"

	cfapi "github.com/cloudflare/cfssl/api"
	"github.com/cloudflare/cfssl/csr"
	cferr "github.com/cloudflare/cfssl/errors"
	"github.com/cloudflare/cfssl/log"
	"github.com/cloudflare/cfssl/signer"
	"github.com/hyperledger/fabric-ca/api"
	"github.com/hyperledger/fabric-ca/util"
)

var (
	// The X.509 BasicConstraints object identifier (RFC 5280, 4.2.1.9)
	basicConstraintsOID = asn1.ObjectIdentifier{2, 5, 29, 19}
)

// newEnrollHandler is the constructor for the enroll handler
func newEnrollHandler(server *Server) (h http.Handler, err error) {
	return newSignHandler(server, "enroll")
}

// newReenrollHandler is the constructor for the reenroll handler
func newReenrollHandler(server *Server) (h http.Handler, err error) {
	return newSignHandler(server, "reenroll")
}

// signHandler for enroll or reenroll requests
type signHandler struct {
	server *Server
	// "enroll" or "reenroll"
	endpoint string
}

// The enrollment response from the server
type enrollmentResponseNet struct {
	// Base64 encoded PEM-encoded ECert
	Cert string
	// The server information
	ServerInfo serverInfoResponseNet
}

// newSignHandler is the constructor for an enroll or reenroll handler
func newSignHandler(server *Server, endpoint string) (h http.Handler, err error) {
	// NewHandler is constructor for register handler
	return &cfapi.HTTPHandler{
		Handler: &signHandler{server: server, endpoint: endpoint},
		Methods: []string{"POST"},
	}, nil
}

// Handle an enroll or reenroll request.
// Authentication has already occurred for both enroll and reenroll prior
// to calling this function in auth.go.
func (sh *signHandler) Handle(w http.ResponseWriter, r *http.Request) error {

	log.Debugf("Received request for endpoint %s", sh.endpoint)

	// Read the request's body
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		return err
	}
	r.Body.Close()

	var req api.EnrollmentRequestNet

	err = util.Unmarshal(body, &req, sh.endpoint)
	if err != nil {
		return err
	}

	log.Debugf("Enrollment request: %+v\n", req)

	caname := r.Header.Get(caHdrName)

	// Make any authorization checks needed, depending on the contents
	// of the CSR (Certificate Signing Request)
	err = sh.csrAuthCheck(&req.SignRequest, r)
	if err != nil {
		return err
	}

	// Sign the certificate
	cert, err := sh.server.caMap[caname].enrollSigner.Sign(req.SignRequest)
	if err != nil {
		err = fmt.Errorf("Failed signing for endpoint %s: %s", sh.endpoint, err)
		log.Error(err.Error())
		return err
	}

	// Send the response with the cert and the server info
	resp := &enrollmentResponseNet{Cert: util.B64Encode(cert)}
	err = sh.server.caMap[caname].fillCAInfo(&resp.ServerInfo)
	if err != nil {
		return err
	}

	return cfapi.SendResponse(w, resp)
}

// Make any authorization checks needed, depending on the contents
// of the CSR (Certificate Signing Request).
// In particular, if the request is for an intermediate CA certificate,
// the caller must have the "hf.IntermediateCA" attribute.
func (sh *signHandler) csrAuthCheck(req *signer.SignRequest, r *http.Request) error {
	// Decode and parse the request into a CSR so we can make checks
	caname := r.Header.Get(caHdrName)
	block, _ := pem.Decode([]byte(req.Request))
	if block == nil {
		return cferr.New(cferr.CSRError, cferr.DecodeFailed)
	}
	if block.Type != "NEW CERTIFICATE REQUEST" && block.Type != "CERTIFICATE REQUEST" {
		return cferr.Wrap(cferr.CSRError,
			cferr.BadRequest, errors.New("not a certificate or csr"))
	}
	csrReq, err := x509.ParseCertificateRequest(block.Bytes)
	if err != nil {
		return err
	}
	// Check the CSR for the X.509 BasicConstraints extension (RFC 5280, 4.2.1.9)
	for _, val := range csrReq.Extensions {
		if val.Id.Equal(basicConstraintsOID) {
			var constraints csr.BasicConstraints
			var rest []byte
			if rest, err = asn1.Unmarshal(val.Value, &constraints); err != nil {
				return cferr.Wrap(cferr.CSRError, cferr.ParseFailed, err)
			} else if len(rest) != 0 {
				return cferr.Wrap(cferr.CSRError, cferr.ParseFailed, errors.New("x509: trailing data after X.509 BasicConstraints"))
			}
			if constraints.IsCA {
				log.Debug("CSR request received for an intermediate CA")
				// This is a request for a CA certificate, so make sure the caller
				// has the 'hf.IntermediateCA' attribute
				return sh.server.caMap[caname].userHasAttribute(r.Header.Get(enrollmentIDHdrName), "hf.IntermediateCA")
			}
		}
	}
	log.Debug("CSR request received")
	return nil
}
