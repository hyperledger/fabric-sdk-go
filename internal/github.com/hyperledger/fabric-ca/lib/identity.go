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
/*
Notice: This file has been modified for Hyperledger Fabric SDK Go usage.
Please review third_party pinning scripts and patches for more details.
*/

package lib

import (
	"net/http"

	"github.com/hyperledger/fabric-sdk-go/api/apicryptosuite"
	"github.com/hyperledger/fabric-sdk-go/pkg/errors"

	"github.com/hyperledger/fabric-sdk-go/internal/github.com/hyperledger/fabric-ca/api"
	log "github.com/hyperledger/fabric-sdk-go/internal/github.com/hyperledger/fabric-ca/sdkpatch/logbridge"
	"github.com/hyperledger/fabric-sdk-go/internal/github.com/hyperledger/fabric-ca/util"
)

func newIdentity(client *Client, name string, key apicryptosuite.Key, cert []byte) *Identity {
	id := new(Identity)
	id.name = name
	id.ecert = newSigner(key, cert, id)
	id.client = client
	if client != nil {
		id.CSP = client.csp
	} else {
		id.CSP = nil
	}
	return id
}

// Identity is fabric-ca's implementation of an identity
type Identity struct {
	name   string
	ecert  *Signer
	client *Client
	CSP    apicryptosuite.CryptoSuite
}

// GetName returns the identity name
func (i *Identity) GetName() string {
	return i.name
}

// GetECert returns the enrollment certificate signer for this identity
func (i *Identity) GetECert() *Signer {
	return i.ecert
}

// Register registers a new identity
// @param req The registration request
func (i *Identity) Register(req *api.RegistrationRequest) (rr *api.RegistrationResponse, err error) {
	log.Debugf("Register %+v", req)
	if req.Name == "" {
		return nil, errors.New("Register was called without a Name set")
	}

	reqBody, err := util.Marshal(req, "RegistrationRequest")
	if err != nil {
		return nil, err
	}

	// Send a post to the "register" endpoint with req as body
	resp := &api.RegistrationResponse{}
	err = i.Post("register", reqBody, resp)
	if err != nil {
		return nil, err
	}

	log.Debug("The register request completely successfully")
	return resp, nil
}

// Reenroll reenrolls an existing Identity and returns a new Identity
// @param req The reenrollment request
func (i *Identity) Reenroll(req *api.ReenrollmentRequest) (*EnrollmentResponse, error) {
	log.Debugf("Reenrolling %s", util.StructToString(req))

	csrPEM, key, err := i.client.GenCSR(req.CSR, i.GetName())
	if err != nil {
		return nil, err
	}

	reqNet := &api.ReenrollmentRequestNet{
		CAName:   req.CAName,
		AttrReqs: req.AttrReqs,
	}

	// Get the body of the request
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
	var result enrollmentResponseNet
	err = i.Post("reenroll", body, &result)
	if err != nil {
		return nil, err
	}
	return i.client.newEnrollmentResponse(&result, i.GetName(), key)
}

// Revoke the identity associated with 'id'
func (i *Identity) Revoke(req *api.RevocationRequest) error {
	log.Debugf("Entering identity.Revoke %+v", req)
	reqBody, err := util.Marshal(req, "RevocationRequest")
	if err != nil {
		return err
	}
	err = i.Post("revoke", reqBody, nil)
	if err != nil {
		return err
	}
	log.Debugf("Successfully revoked %+v", req)
	return nil
}

// Post sends arbtrary request body (reqBody) to an endpoint.
// This adds an authorization header which contains the signature
// of this identity over the body and non-signature part of the authorization header.
// The return value is the body of the response.
func (i *Identity) Post(endpoint string, reqBody []byte, result interface{}) error {
	req, err := i.client.newPost(endpoint, reqBody)
	if err != nil {
		return err
	}
	err = i.addTokenAuthHdr(req, reqBody)
	if err != nil {
		return err
	}
	return i.client.SendReq(req, result)
}

func (i *Identity) addTokenAuthHdr(req *http.Request, body []byte) error {
	log.Debug("adding token-based authorization header")
	cert := i.ecert.cert
	key := i.ecert.key
	token, err := util.CreateToken(i.CSP, cert, key, body)
	if err != nil {
		return errors.WithMessage(err, "Failed to add token authorization header")
	}
	req.Header.Set("authorization", token)
	return nil
}
