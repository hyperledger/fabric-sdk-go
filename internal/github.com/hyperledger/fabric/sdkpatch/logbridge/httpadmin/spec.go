/*
Copyright IBM Corp. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/
/*
Notice: This file has been modified for Hyperledger Fabric SDK Go usage.
Please review third_party pinning scripts and patches for more details.
*/

package httpadmin

import (
	"encoding/json"
	"fmt"
	"net/http"

	flogging "github.com/hyperledger/fabric-sdk-go/internal/github.com/hyperledger/fabric/sdkpatch/logbridge"
)

//go:generate counterfeiter -o fakes/logging.go -fake-name Logging . Logging

type Logging interface {
	ActivateSpec(spec string) error
	Spec() string
}

type LogSpec struct {
	Spec string `json:"spec,omitempty"`
}

type ErrorResponse struct {
	Error string `json:"error"`
}

func NewSpecHandler() *SpecHandler {
	return &SpecHandler{
		Logger: flogging.MustGetLogger("flogging.httpadmin"),
	}
}

type SpecHandler struct {
	Logging Logging
	Logger  *flogging.Logger
}

func (h *SpecHandler) ServeHTTP(resp http.ResponseWriter, req *http.Request) {
	switch req.Method {
	case http.MethodPut:
		var logSpec LogSpec
		decoder := json.NewDecoder(req.Body)
		if err := decoder.Decode(&logSpec); err != nil {
			h.sendResponse(resp, http.StatusBadRequest, err)
			return
		}
		req.Body.Close()

		if err := h.Logging.ActivateSpec(logSpec.Spec); err != nil {
			h.sendResponse(resp, http.StatusBadRequest, err)
			return
		}
		resp.WriteHeader(http.StatusNoContent)

	case http.MethodGet:
		h.sendResponse(resp, http.StatusOK, &LogSpec{Spec: h.Logging.Spec()})

	default:
		err := fmt.Errorf("invalid request method: %s", req.Method)
		h.sendResponse(resp, http.StatusBadRequest, err)
	}
}

func (h *SpecHandler) sendResponse(resp http.ResponseWriter, code int, payload interface{}) {
	encoder := json.NewEncoder(resp)
	if err, ok := payload.(error); ok {
		payload = &ErrorResponse{Error: err.Error()}
	}

	resp.WriteHeader(code)

	resp.Header().Set("Content-Type", "application/json")
	if err := encoder.Encode(payload); err != nil {
		h.Logger.Errorf("[error] failed to encode payload", err)
	}
}
