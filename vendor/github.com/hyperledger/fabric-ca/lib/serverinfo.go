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
	"net/http"

	cfapi "github.com/cloudflare/cfssl/api"
	"github.com/cloudflare/cfssl/log"
)

// infoHandler handles the GET /info request
type infoHandler struct {
	server *Server
}

// newInfoHandler is the constructor for the infoHandler
func newInfoHandler(server *Server) (h http.Handler, err error) {
	return &cfapi.HTTPHandler{
		Handler: &infoHandler{server: server},
		Methods: []string{"GET"},
	}, nil
}

// The response to the GET /info request
type serverInfoResponseNet struct {
	// CAName is a unique name associated with fabric-ca-server's CA
	CAName string
	// Base64 encoding of PEM-encoded certificate chain
	CAChain string
}

// Handle is the handler for the GET /info request
func (ih *infoHandler) Handle(w http.ResponseWriter, r *http.Request) error {
	log.Debug("Received request for server info")
	resp := &serverInfoResponseNet{}
	err := ih.server.fillServerInfo(resp)
	if err != nil {
		return err
	}
	return cfapi.SendResponse(w, resp)
}
