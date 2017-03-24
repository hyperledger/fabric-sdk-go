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
	"strings"

	cfsslapi "github.com/cloudflare/cfssl/api"
	"github.com/cloudflare/cfssl/log"

	"github.com/hyperledger/fabric-ca/api"
	"github.com/hyperledger/fabric-ca/lib/spi"
	"github.com/hyperledger/fabric-ca/util"
)

// newRevokeHandler is constructor for revoke handler
func newRevokeHandler(server *Server) (h http.Handler, err error) {
	return &cfsslapi.HTTPHandler{
		Handler: &revokeHandler{server: server},
		Methods: []string{"POST"}}, nil
}

// revokeHandler for revoke requests
type revokeHandler struct {
	server *Server
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

	cert, err := util.VerifyToken(h.server.csp, authHdr, body)
	if err != nil {
		return authErr(w, err)
	}

	// Make sure that the user has the "hf.Revoker" attribute in order to be authorized
	// to revoke a certificate.  This attribute comes from the user registry, which
	// is either in the DB if LDAP is not configured, or comes from LDAP if LDAP is
	// configured.
	err = h.server.userHasAttribute(cert.Subject.CommonName, "hf.Revoker")
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

	req.AKI = strings.ToLower(req.AKI)
	req.Serial = strings.ToLower(req.Serial)

	certDBAccessor := h.server.certDBAccessor
	registry := h.server.registry

	if req.Serial != "" && req.AKI != "" {
		certificate, err := certDBAccessor.GetCertificateWithID(req.Serial, req.AKI)
		if err != nil {
			log.Error(notFound(w, err))
			return notFound(w, err)
		}

		userInfo, err2 := registry.GetUserInfo(certificate.ID)
		if err2 != nil {
			return err2
		}

		err2 = h.checkAffiliations(cert.Subject.CommonName, userInfo.Affiliation)
		if err2 != nil {
			return err2
		}

		err = certDBAccessor.RevokeCertificate(req.Serial, req.AKI, req.Reason)
		if err != nil {
			log.Error(notFound(w, err))
			return notFound(w, err)
		}
	} else if req.Name != "" {

		user, err := registry.GetUser(req.Name, nil)
		if err != nil {
			err = fmt.Errorf("Failed to get user %s: %s", req.Name, err)
			return notFound(w, err)
		}

		// Set user state to -1 for revoked user
		if user != nil {
			var userInfo spi.UserInfo
			userInfo, err = registry.GetUserInfo(req.Name)
			if err != nil {
				err = fmt.Errorf("Failed to get user info %s: %s", req.Name, err)
				return notFound(w, err)
			}

			err = h.checkAffiliations(cert.Subject.CommonName, userInfo.Affiliation)
			if err != nil {
				return err
			}

			userInfo.State = -1

			err = registry.UpdateUser(userInfo)
			if err != nil {
				log.Warningf("Revoke failed: %s", err)
				return dbErr(w, err)
			}
		}

		var recs []CertRecord
		recs, err = certDBAccessor.RevokeCertificatesByID(req.Name, req.Reason)
		if err != nil {
			log.Warningf("No certificates were revoked for '%s' but the ID was disabled: %s", req.Name, err)
			return dbErr(w, err)
		}

		if len(recs) == 0 {
			log.Warningf("No certificates were revoked for '%s' but the ID was disabled: %s", req.Name)
		}

		log.Debugf("Revoked the following certificates owned by '%s': %+v", req.Name, recs)

	} else {
		return badRequest(w, errors.New("Either Name or Serial and AKI are required for a revoke request"))
	}

	log.Debugf("Revoke was successful: %+v", req)

	result := map[string]string{}
	return cfsslapi.SendResponse(w, result)
}

func (h *revokeHandler) checkAffiliations(revoker string, affiliation string) error {
	log.Debugf("Check to see if revoker %s has affiliations to revoke: %s", revoker, affiliation)
	revokerAffiliation, err := h.server.getUserAffiliation(revoker)
	if err != nil {
		return err
	}

	log.Debugf("Affiliation of revoker: %s, affiliation of user being revoked: %s", revokerAffiliation, affiliation)

	if !strings.HasPrefix(affiliation, revokerAffiliation) {
		return fmt.Errorf("Revoker %s does not have proper affiliation to revoke user", revoker)
	}

	return nil
}
