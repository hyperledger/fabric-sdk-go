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
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"

	cfsslapi "github.com/cloudflare/cfssl/api"
	"github.com/cloudflare/cfssl/log"

	"github.com/hyperledger/fabric-ca/api"
	"github.com/hyperledger/fabric-ca/lib/spi"
	"github.com/hyperledger/fabric-ca/util"
)

// NewRevokeHandler is constructor for revoke handler
func NewRevokeHandler() (h http.Handler, err error) {
	// NewHandler is constructor for register handler
	return &cfsslapi.HTTPHandler{
		Handler: &revokeHandler{},
		Methods: []string{"POST"}}, nil
}

// revokeHandler for revoke requests
type revokeHandler struct {
}

// Handle an revoke request
func (h *revokeHandler) Handle(w http.ResponseWriter, r *http.Request) error {

	log.Debug("Revoke request received")

	authHdr := r.Header.Get("authorization")
	if authHdr == "" {
		return authErr(w, errors.New("no authorization header"))
	}

	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		return badRequest(w, err)
	}
	r.Body.Close()

	cert, err := util.VerifyToken(MyCSP, authHdr, body)
	if err != nil {
		return authErr(w, err)
	}

	// Make sure that the user has the "hf.Revoker" attribute in order to be authorized
	// to revoke a certificate.  This attribute comes from the user registry, which
	// is either in the DB if LDAP is not configured, or comes from LDAP if LDAP is
	// configured.
	err = userHasAttribute(cert.Subject.CommonName, "hf.Revoker")
	if err != nil {
		return authErr(w, err)
	}

	// Parse revoke request body
	var req api.RevocationRequestNet
	err = json.Unmarshal(body, &req)
	if err != nil {
		return badRequest(w, err)
	}

	log.Debugf("Revoke request: %+v", req)

	if req.Serial != "" && req.AKI != "" {
		err = MyCertDBAccessor.RevokeCertificate(req.Serial, req.AKI, req.Reason)
		if err != nil {
			return notFound(w, err)
		}
	} else if req.Name != "" {

		user, err := UserRegistry.GetUser(req.Name, nil)
		if err != nil {
			err = fmt.Errorf("Failed to get user %s: %s", req.Name, err)
			return notFound(w, err)
		}

		// Set user state to -1 for revoked user
		if user != nil {
			var userInfo spi.UserInfo
			userInfo, err = UserRegistry.GetUserInfo(req.Name)
			if err != nil {
				err = fmt.Errorf("Failed to get user info %s: %s", req.Name, err)
				return notFound(w, err)
			}

			userInfo.State = -1

			err = UserRegistry.UpdateUser(userInfo)
			if err != nil {
				log.Warningf("Revoke failed: %s", err)
				return dbErr(w, err)
			}
		}

		var recs []CertRecord
		recs, err = MyCertDBAccessor.RevokeCertificatesByID(req.Name, req.Reason)
		if err != nil {
			log.Warningf("No certificates were revoked for '%s' but the ID was disabled: %s", req.Name, err)
			return dbErr(w, err)
		}
		log.Debugf("Revoked the following certificates owned by '%s': %+v", req.Name, recs)

	} else {
		return badRequest(w, errors.New("Either Name or Serial and AKI are required for a revoke request"))
	}

	log.Debugf("Revoke was successful: %+v", req)

	result := map[string]string{}
	return cfsslapi.SendResponse(w, result)
}
