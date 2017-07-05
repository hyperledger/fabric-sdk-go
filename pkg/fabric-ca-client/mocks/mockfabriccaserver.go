/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package mocks

import (
	"fmt"
	"net/http"

	cfapi "github.com/cloudflare/cfssl/api"
	cfsslapi "github.com/cloudflare/cfssl/api"
	"github.com/hyperledger/fabric-ca/api"
	"github.com/hyperledger/fabric-ca/util"
)

var ecert = `-----BEGIN CERTIFICATE-----
MIICEjCCAbigAwIBAgIQPjb63mDL4e062MPjtcA1CDAKBggqhkjOPQQDAjBgMQsw
CQYDVQQGEwJVUzETMBEGA1UECBMKQ2FsaWZvcm5pYTEWMBQGA1UEBxMNU2FuIEZy
YW5jaXNjbzERMA8GA1UEChMIcGVlck9yZzExETAPBgNVBAMTCHBlZXJPcmcxMB4X
DTE3MDMwMTE3MzY0MVoXDTI3MDIyNzE3MzY0MVowUjELMAkGA1UEBhMCVVMxEzAR
BgNVBAgTCkNhbGlmb3JuaWExFjAUBgNVBAcTDVNhbiBGcmFuY2lzY28xFjAUBgNV
BAMTDXBlZXJPcmcxUGVlcjEwWTATBgcqhkjOPQIBBggqhkjOPQMBBwNCAAS0hO8C
8ph+PiFkYikdVAK/zCd2ckxb6m5bTOq54VtWR7wbdPuu9djICTaROTUmfeoAHF60
ol/Z/penR/G6chqKo2IwYDAOBgNVHQ8BAf8EBAMCBaAwEwYDVR0lBAwwCgYIKwYB
BQUHAwEwDAYDVR0TAQH/BAIwADArBgNVHSMEJDAigCDYpbPKwbgh9uS0h86vH9I5
zc/DEIlBUJCLkPBekXlVajAKBggqhkjOPQQDAgNIADBFAiEAmGS3LTaqCkWV+myl
lhg9ovtLJABuxQLnajMJYQOXURgCIHLVNrDbEF0KpEmFwXIBYMFdsKGRAF0kC43M
bpq87UJq
-----END CERTIFICATE-----
`

// The enrollment response from the server
type enrollmentResponseNet struct {
	// Base64 encoded PEM-encoded ECert
	Cert string
	// The server information
	ServerInfo serverInfoResponseNet
}

// The response to the GET /info request
type serverInfoResponseNet struct {
	// CAName is a unique name associated with fabric-ca-server's CA
	CAName string
	// Base64 encoding of PEM-encoded certificate chain
	CAChain string
}

// StartFabricCAMockServer Start fabric ca mock server
func StartFabricCAMockServer(address string) error {

	// Register request handlers
	http.HandleFunc("/register", Register)
	http.HandleFunc("/enroll", Enroll)
	http.HandleFunc("/reenroll", Enroll)

	server := &http.Server{
		Addr:      address,
		TLSConfig: nil,
	}

	err := server.ListenAndServe()

	if err != nil {
		return fmt.Errorf("HTTP Server: Failed to start %v ", err.Error())
	}
	fmt.Println("HTTP Server started on :" + address)
	return nil

}

// Register user
func Register(w http.ResponseWriter, req *http.Request) {
	resp := &api.RegistrationResponseNet{RegistrationResponse: api.RegistrationResponse{Secret: "mockSecretValue"}}
	cfsslapi.SendResponse(w, resp)
}

// Enroll user
func Enroll(w http.ResponseWriter, req *http.Request) {
	resp := &enrollmentResponseNet{Cert: util.B64Encode([]byte(ecert))}
	fillCAInfo(&resp.ServerInfo)
	cfapi.SendResponse(w, resp)
}

// Reenroll user
func Reenroll(w http.ResponseWriter, req *http.Request) {
	resp := &enrollmentResponseNet{Cert: util.B64Encode([]byte(ecert))}
	fillCAInfo(&resp.ServerInfo)
	cfapi.SendResponse(w, resp)
}

// Fill the CA info structure appropriately
func fillCAInfo(info *serverInfoResponseNet) {
	info.CAName = "MockCAName"
	info.CAChain = util.B64Encode([]byte("MockCAChain"))
}
