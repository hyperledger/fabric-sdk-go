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

// registerHandler for register requests
type registerHandler struct {
	server *Server
}

// newRegisterHandler is constructor for register handler
func newRegisterHandler(server *Server) (h http.Handler, err error) {
	// NewHandler is constructor for register handler
	return &cfsslapi.HTTPHandler{
		Handler: &registerHandler{server: server},
		Methods: []string{"POST"},
	}, nil
}

// Handle a register request
func (h *registerHandler) Handle(w http.ResponseWriter, r *http.Request) error {
	log.Debug("Register request received")

	reqBody, err := ioutil.ReadAll(r.Body)
	if err != nil {
		return err
	}
	r.Body.Close()

	// Parse request body
	var req api.RegistrationRequestNet
	err = json.Unmarshal(reqBody, &req)
	if err != nil {
		return err
	}

	caname := r.Header.Get(caHdrName)

	// Register User
	callerID := r.Header.Get(enrollmentIDHdrName)
	secret, err := h.RegisterUser(&req, callerID, caname)
	if err != nil {
		return err
	}

	resp := &api.RegistrationResponseNet{RegistrationResponse: api.RegistrationResponse{Secret: secret}}

	log.Debugf("Registration completed - sending response %+v", &resp)
	return cfsslapi.SendResponse(w, resp)
}

// RegisterUser will register a user
func (h *registerHandler) RegisterUser(req *api.RegistrationRequestNet, registrar, caname string) (string, error) {

	secret := req.Secret
	req.Secret = "<<user-specified>>"
	log.Debugf("Received registration request from %s: %+v", registrar, req)
	req.Secret = secret

	var err error

	if registrar != "" {
		// Check the permissions of member named 'registrar' to perform this registration
		err = h.canRegister(registrar, req.Type, caname)
		if err != nil {
			log.Debugf("Registration of '%s' failed: %s", req.Name, err)
			return "", err
		}
	}

	err = h.validateID(req, caname)
	if err != nil {
		log.Debugf("Registration of '%s' failed: %s", req.Name, err)
		return "", err
	}

	secret, err = h.registerUserID(req, caname)

	if err != nil {
		log.Debugf("Registration of '%s' failed: %s", req.Name, err)
		return "", err
	}

	return secret, nil
}

func (h *registerHandler) validateID(req *api.RegistrationRequestNet, caname string) error {
	log.Debug("Validate ID")
	// Check whether the affiliation is required for the current user.
	if h.requireAffiliation(req.Type) {
		// If yes, is the affiliation valid
		err := h.isValidAffiliation(req.Affiliation, caname)
		if err != nil {
			return err
		}
	}
	return nil
}

// registerUserID registers a new user and its enrollmentID, role and state
func (h *registerHandler) registerUserID(req *api.RegistrationRequestNet, caname string) (string, error) {
	log.Debugf("Registering user id: %s\n", req.Name)
	var err error

	if req.Secret == "" {
		req.Secret = util.RandomString(12)
	}

	caMaxEnrollments := h.server.caMap[caname].Config.Registry.MaxEnrollments

	req.MaxEnrollments, err = getMaxEnrollments(req.MaxEnrollments, caMaxEnrollments)
	if err != nil {
		return "", err
	}

	// Make sure delegateRoles is not larger than roles
	roles := GetAttrValue(req.Attributes, attrRoles)
	delegateRoles := GetAttrValue(req.Attributes, attrDelegateRoles)
	err = util.IsSubsetOf(delegateRoles, roles)
	if err != nil {
		return "", fmt.Errorf("delegateRoles is superset of roles: %s", err)
	}

	insert := spi.UserInfo{
		Name:           req.Name,
		Pass:           req.Secret,
		Type:           req.Type,
		Affiliation:    req.Affiliation,
		Attributes:     req.Attributes,
		MaxEnrollments: req.MaxEnrollments,
	}

	registry := h.server.caMap[caname].registry

	_, err = registry.GetUser(req.Name, nil)
	if err == nil {
		return "", fmt.Errorf("Identity '%s' is already registered", req.Name)
	}

	err = registry.InsertUser(insert)
	if err != nil {
		return "", err
	}

	return req.Secret, nil
}

func (h *registerHandler) isValidAffiliation(affiliation string, caname string) error {
	log.Debug("Validating affiliation: " + affiliation)

	_, err := h.server.caMap[caname].registry.GetAffiliation(affiliation)
	if err != nil {
		return fmt.Errorf("Failed getting affiliation '%s': %s", affiliation, err)
	}

	return nil
}

func (h *registerHandler) requireAffiliation(idType string) bool {
	log.Debugf("An affiliation is required for identity type %s", idType)
	// Require an affiliation for all identity types
	return true
}

func (h *registerHandler) canRegister(registrar string, userType string, caname string) error {
	log.Debugf("canRegister - Check to see if user %s can register", registrar)

	user, err := h.server.caMap[caname].registry.GetUser(registrar, nil)
	if err != nil {
		return fmt.Errorf("Registrar does not exist: %s", err)
	}

	var roles []string
	rolesStr := user.GetAttribute("hf.Registrar.Roles")
	if rolesStr != "" {
		roles = strings.Split(rolesStr, ",")
	} else {
		roles = make([]string, 0)
	}
	if userType != "" {
		if !util.StrContained(userType, roles) {
			return fmt.Errorf("Identity '%s' may not register type '%s'", registrar, userType)
		}
	} else {
		return errors.New("No identity type provided. Please provide identity type")
	}

	return nil
}
