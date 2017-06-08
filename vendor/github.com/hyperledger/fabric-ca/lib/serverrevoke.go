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

	// Parse revoke request body
	var req api.RevocationRequestNet
	err = json.Unmarshal(body, &req)
	if err != nil {
		return badRequest(w, err)
	}

	log.Debugf("Revoke request: %+v", req)

	caname := r.Header.Get(caHdrName)

	cert, err := util.VerifyToken(h.server.caMap[caname].csp, authHdr, body)
	if err != nil {
		return authErr(w, err)
	}

	// Make sure that the user has the "hf.Revoker" attribute in order to be authorized
	// to revoke a certificate.  This attribute comes from the user registry, which
	// is either in the DB if LDAP is not configured, or comes from LDAP if LDAP is
	// configured.
	err = h.server.caMap[caname].userHasAttribute(cert.Subject.CommonName, "hf.Revoker")
	if err != nil {
		return authErr(w, err)
	}

	req.AKI = strings.TrimLeft(strings.ToLower(req.AKI), "0")
	req.Serial = strings.TrimLeft(strings.ToLower(req.Serial), "0")

	certDBAccessor := h.server.caMap[caname].certDBAccessor
	registry := h.server.caMap[caname].registry
	reason := util.RevocationReasonCodes[req.Reason]

	if req.Serial != "" && req.AKI != "" {
		certificate, err := certDBAccessor.GetCertificateWithID(req.Serial, req.AKI)
		if err != nil {
			msg := fmt.Sprintf("Failed to retrieve certificate for the provided serial number and AKI: %s", err)
			log.Errorf(msg)
			return notFound(w, errors.New(msg))
		}

		if req.Name != "" && req.Name != certificate.ID {
			err = fmt.Errorf("The serial number %s and the AKI %s do not belong to the Enrollment ID %s",
				req.Serial, req.AKI, req.Name)
			return badRequest(w, err)
		}

		userInfo, err2 := registry.GetUserInfo(certificate.ID)
		if err2 != nil {
			msg := fmt.Sprintf("Failed to find user: %s", err2)
			log.Errorf(msg)
			return dbErr(w, errors.New(msg))
		}

		err2 = h.checkAffiliations(cert.Subject.CommonName, userInfo, caname)
		if err2 != nil {
			log.Error(err2)
			return authErr(w, err2)
		}

		err = certDBAccessor.RevokeCertificate(req.Serial, req.AKI, reason)
		if err != nil {
			msg := fmt.Sprintf("Failed to revoke certificate: %s", err)
			log.Error(msg)
			return notFound(w, errors.New(msg))
		}
	} else if req.Name != "" {

		user, err := registry.GetUser(req.Name, nil)
		if err != nil {
			err = fmt.Errorf("Failed to get identity %s: %s", req.Name, err)
			return notFound(w, err)
		}

		// Set user state to -1 for revoked user
		if user != nil {
			var userInfo spi.UserInfo
			userInfo, err = registry.GetUserInfo(req.Name)
			if err != nil {
				err = fmt.Errorf("Failed to get identity info %s: %s", req.Name, err)
				return notFound(w, err)
			}

			err = h.checkAffiliations(cert.Subject.CommonName, userInfo, caname)
			if err != nil {
				log.Error(err)
				return authErr(w, err)
			}

			userInfo.State = -1

			err = registry.UpdateUser(userInfo)
			if err != nil {
				log.Warningf("Revoke failed: %s", err)
				return dbErr(w, err)
			}
		}

		var recs []CertRecord
		recs, err = certDBAccessor.RevokeCertificatesByID(req.Name, reason)
		if err != nil {
			log.Warningf("No certificates were revoked for '%s' but the ID was disabled: %s", req.Name, err)
			return dbErr(w, err)
		}

		if len(recs) == 0 {
			log.Warningf("No certificates were revoked for '%s' but the ID was disabled", req.Name)
		}

		log.Debugf("Revoked the following certificates owned by '%s': %+v", req.Name, recs)

	} else {
		return badRequest(w, errors.New("Either Name or Serial and AKI are required for a revoke request"))
	}

	log.Debugf("Revoke was successful: %+v", req)

	result := map[string]string{}
	return cfsslapi.SendResponse(w, result)
}

func (h *revokeHandler) checkAffiliations(revoker string, revoking spi.UserInfo, caname string) error {
	log.Debugf("Check to see if revoker %s has affiliations to revoke: %s", revoker, revoking.Name)
	userAffiliation, err := h.server.caMap[caname].getUserAffiliation(revoker)
	if err != nil {
		return err
	}

	log.Debugf("Affiliation of revoker: %s, affiliation of identity being revoked: %s", userAffiliation, revoking.Affiliation)

	// Revoking user has root affiliation thus has ability to revoke
	if userAffiliation == "" {
		log.Debug("Identity with root affiliation revoking")
		return nil
	}

	revokingAffiliation := strings.Split(revoking.Affiliation, ".")
	revokerAffiliation := strings.Split(userAffiliation, ".")
	for i := range revokerAffiliation {
		if revokerAffiliation[i] != revokingAffiliation[i] {
			return fmt.Errorf("Revoker %s does not have proper affiliation to revoke identity %s", revoker, revoking.Name)
		}

	}

	return nil
}
