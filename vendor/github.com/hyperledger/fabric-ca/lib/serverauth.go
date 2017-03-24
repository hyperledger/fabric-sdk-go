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
	"errors"
	"io/ioutil"
	"net/http"

	"github.com/cloudflare/cfssl/api"
	cerr "github.com/cloudflare/cfssl/errors"
	"github.com/cloudflare/cfssl/log"
	"github.com/cloudflare/cfssl/revoke"
	"github.com/hyperledger/fabric-ca/util"
)

const (
	enrollmentIDHdrName = "__eid__"
)

// AuthType is the enum for authentication types: basic and token
type authType int

const (
	noAuth authType = iota
	basic           // basic = 1
	token           // token = 2
)

// Fabric CA authentication handler
type fcaAuthHandler struct {
	server   *Server
	authType authType
	next     http.Handler
}

var authError = cerr.NewBadRequest(errors.New("Authorization failure"))

func (ah *fcaAuthHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	err := ah.serveHTTP(w, r)
	if err != nil {
		api.HandleError(w, err)
	} else {
		ah.next.ServeHTTP(w, r)
	}
}

// Handle performs authentication
func (ah *fcaAuthHandler) serveHTTP(w http.ResponseWriter, r *http.Request) error {
	log.Debugf("Received request\n%s", util.HTTPRequestToString(r))
	authHdr := r.Header.Get("authorization")
	switch ah.authType {
	case noAuth:
		// No authentication required
		return nil
	case basic:
		if authHdr == "" {
			log.Debug("No authorization header")
			return errNoAuthHdr
		}
		user, pwd, ok := r.BasicAuth()
		if ok {
			if ah.authType != basic {
				log.Debugf("Basic auth is not allowed; found %s", authHdr)
				return errBasicAuthNotAllowed
			}
			u, err := ah.server.registry.GetUser(user, nil)
			if err != nil {
				log.Debugf("Failed to get user '%s': %s", user, err)
				return authError
			}
			err = u.Login(pwd)
			if err != nil {
				log.Debugf("Failed to login '%s': %s", user, err)
				return authError
			}
			log.Debug("User/Pass was correct")
			r.Header.Set(enrollmentIDHdrName, user)
			return nil
		}
		return authError
	case token:
		// read body
		body, err := ioutil.ReadAll(r.Body)
		if err != nil {
			log.Debugf("Failed to read body: %s", err)
			return authError
		}
		r.Body = ioutil.NopCloser(bytes.NewReader(body))
		// verify token
		cert, err2 := util.VerifyToken(ah.server.csp, authHdr, body)
		if err2 != nil {
			log.Debugf("Failed to verify token: %s", err2)
			return authError
		}
		id := util.GetEnrollmentIDFromX509Certificate(cert)
		log.Debugf("Checking for revocation/expiration of certificate owned by '%s'", id)
		// Check for certificate revocation and expiration
		revokedOrExpired, checked := revoke.VerifyCertificate(cert)
		if revokedOrExpired {
			log.Debugf("Certificate was either revoked or has expired owned by '%s'", id)
			return authError
		}
		if !checked {
			log.Debug("A failure occurred while checking for revocation and expiration")
			return authError
		}
		log.Debugf("Successful authentication of '%s'", id)
		r.Header.Set(enrollmentIDHdrName, util.GetEnrollmentIDFromX509Certificate(cert))
		return nil
	default: // control should never reach here
		log.Errorf("No handler for the authentication type: %d", ah.authType)
		return authError
	}
}

func wrappedPath(path string) string {
	return "/api/v1/cfssl/" + path
}
