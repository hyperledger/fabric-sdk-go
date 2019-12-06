/*
Copyright IBM Corp. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/
/*
Notice: This file has been modified for Hyperledger Fabric SDK Go usage.
Please review third_party pinning scripts and patches for more details.
*/

package lib

import (
	"encoding/hex"
	"strings"

	"github.com/cloudflare/cfssl/log"
	"github.com/hyperledger/fabric-ca/api"
	"github.com/hyperledger/fabric-ca/lib/caerrors"
	"github.com/hyperledger/fabric-ca/lib/server/db"
	"github.com/hyperledger/fabric-ca/util"
)

type revocationResponseNet struct {
	RevokedCerts []api.RevokedCert
	CRL          string
}

// CertificateStatus represents status of an enrollment certificate
type CertificateStatus string

const (
	// Revoked is the status of a revoked certificate
	Revoked CertificateStatus = "revoked"
	// Good is the status of a active certificate
	Good = "good"
)

func newRevokeEndpoint(s *Server) *serverEndpoint {
	return &serverEndpoint{
		Path:    "revoke",
		Methods: []string{"POST"},
		Handler: revokeHandler,
		Server:  s,
	}
}

// Handle an revoke request
func revokeHandler(ctx *serverRequestContextImpl) (interface{}, error) {
	// Parse revoke request body
	var req api.RevocationRequestNet
	err := ctx.ReadBody(&req)
	if err != nil {
		return nil, err
	}
	// Authentication
	caller, err := ctx.TokenAuthentication()
	if err != nil {
		return nil, err
	}
	// Get targeted CA
	ca, err := ctx.GetCA()
	if err != nil {
		return nil, err
	}

	req.AKI = parseInput(req.AKI)
	req.Serial = parseInput(req.Serial)

	certDBAccessor := ca.certDBAccessor
	registry := ca.registry
	reason := util.RevocationReasonCodes[req.Reason]

	result := &revocationResponseNet{}
	if req.Serial != "" && req.AKI != "" {
		calleraki := strings.ToLower(strings.TrimLeft(hex.EncodeToString(ctx.enrollmentCert.AuthorityKeyId), "0"))
		callerserial := strings.ToLower(strings.TrimLeft(util.GetSerialAsHex(ctx.enrollmentCert.SerialNumber), "0"))

		certificate, err := certDBAccessor.GetCertificateWithID(req.Serial, req.AKI)
		if err != nil {
			return nil, caerrors.NewHTTPErr(404, caerrors.ErrRevCertNotFound, "Certificate with serial %s and AKI %s was not found: %s",
				req.Serial, req.AKI, err)
		}

		// Authorization
		err = checkAuth(caller, certificate.ID, ca)
		if err != nil {
			return nil, err
		}

		if certificate.Status == string(Revoked) {
			return nil, caerrors.NewHTTPErr(404, caerrors.ErrCertAlreadyRevoked, "Certificate with serial %s and AKI %s was already revoked",
				req.Serial, req.AKI)
		}

		if req.Name != "" && req.Name != certificate.ID {
			return nil, caerrors.NewHTTPErr(400, caerrors.ErrCertWrongOwner, "Certificate with serial %s and AKI %s is not owned by %s",
				req.Serial, req.AKI, req.Name)
		}

		userInfo, err := registry.GetUser(certificate.ID, nil)
		if err != nil {
			return nil, caerrors.NewHTTPErr(404, caerrors.ErrRevokeIDNotFound, "Identity %s was not found: %s", certificate.ID, err)
		}

		if !((req.AKI == calleraki) && (req.Serial == callerserial)) {
			err = ctx.CanManageUser(userInfo)
			if err != nil {
				return nil, err
			}
		}

		err = certDBAccessor.RevokeCertificate(req.Serial, req.AKI, reason)
		if err != nil {
			return nil, caerrors.NewHTTPErr(500, caerrors.ErrRevokeFailure, "Revoke of certificate <%s,%s> failed: %s", req.Serial, req.AKI, err)
		}
		result.RevokedCerts = append(result.RevokedCerts, api.RevokedCert{Serial: req.Serial, AKI: req.AKI})
	} else if req.Name != "" {
		// Authorization
		err = checkAuth(caller, req.Name, ca)
		if err != nil {
			return nil, err
		}

		user, err := registry.GetUser(req.Name, nil)
		if err != nil {
			return nil, caerrors.NewHTTPErr(404, caerrors.ErrRevokeIDNotFound, "Identity %s was not found: %s", req.Name, err)
		}

		// Set user state to -1 for revoked user
		if user != nil {
			caller, err := ctx.GetCaller()
			if err != nil {
				return nil, err
			}

			if caller.GetName() != user.GetName() {
				err = ctx.CanManageUser(user)
				if err != nil {
					return nil, err
				}
			}

			err = user.Revoke()
			if err != nil {
				return nil, caerrors.NewHTTPErr(500, caerrors.ErrRevokeUpdateUser, "Failed to revoke user: %s", err)
			}
		}

		var recs []db.CertRecord
		recs, err = certDBAccessor.RevokeCertificatesByID(req.Name, reason)
		if err != nil {
			return nil, caerrors.NewHTTPErr(500, caerrors.ErrNoCertsRevoked, "Failed to revoke certificates for '%s': %s",
				req.Name, err)
		}

		if len(recs) == 0 {
			log.Warningf("No certificates were revoked for '%s' but the ID was disabled", req.Name)
		} else {
			log.Debugf("Revoked the following certificates owned by '%s': %+v", req.Name, recs)
			for _, certRec := range recs {
				result.RevokedCerts = append(result.RevokedCerts, api.RevokedCert{AKI: certRec.AKI, Serial: certRec.Serial})
			}
		}
	} else {
		return nil, caerrors.NewHTTPErr(400, caerrors.ErrMissingRevokeArgs, "Either Name or Serial and AKI are required for a revoke request")
	}

	log.Debugf("Revoke was successful: %+v", req)

	if req.GenCRL && len(result.RevokedCerts) > 0 {
		log.Debugf("Generating CRL")
		crl, err := genCRL(ca, api.GenCRLRequest{CAName: ca.Config.CA.Name})
		if err != nil {
			return nil, err
		}
		result.CRL = util.B64Encode(crl)
	}
	return result, nil
}

func parseInput(input string) string {
	return strings.Replace(strings.TrimLeft(strings.ToLower(input), "0"), ":", "", -1)
}

func checkAuth(callerName, revokeUserName string, ca *CA) error {
	if callerName != revokeUserName {
		// Make sure that the caller has the "hf.Revoker" attribute.
		err := ca.attributeIsTrue(callerName, "hf.Revoker")
		if err != nil {
			return caerrors.NewAuthorizationErr(caerrors.ErrNotRevoker, "Caller does not have authority to revoke")
		}
	}
	return nil
}
