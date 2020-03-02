/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

// Package msp enables creation and update of users on a Fabric network.
// Msp client supports the following actions:
// Enroll, Reenroll, Register,  Revoke and GetSigningIdentity.
//
//  Basic Flow:
//  1) Prepare client context
//  2) Create msp client
//  3) Register user
//  4) Enroll user
package msp

import (
	"fmt"

	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/fab"

	"strings"

	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/context"
	mspctx "github.com/hyperledger/fabric-sdk-go/pkg/common/providers/msp"
	"github.com/hyperledger/fabric-sdk-go/pkg/msp"
	mspapi "github.com/hyperledger/fabric-sdk-go/pkg/msp/api"
	"github.com/pkg/errors"
)

// Client enables access to Client services
type Client struct {
	orgName string
	// ID of the Fabbric CA configuration entry associated with this client (optional)
	// Client will direct Any Fabric CA calls to the URL specified in this configuration entry
	caID string
	// CA name (optional). CA within the Fabric CA server instance at the URL defined at caID.
	// If not present, all calls will be handled by the default CA of the Fabric CA server instance.
	caName string
	ctx    context.Client
}

// New creates a new Client instance
func New(clientProvider context.ClientProvider, opts ...ClientOption) (*Client, error) {

	ctx, c, err := initClientFromOptions(clientProvider, opts...)
	if err != nil {
		return nil, err
	}

	if c.orgName == "" {
		c.orgName = ctx.IdentityConfig().Client().Organization
	}
	if c.orgName == "" {
		return nil, errors.New("organization is not provided")
	}

	networkConfig := ctx.EndpointConfig().NetworkConfig()
	org, ok := networkConfig.Organizations[strings.ToLower(c.orgName)]
	if !ok {
		return nil, fmt.Errorf("non-existent organization: '%s'", c.orgName)
	}

	if c.caID == "" && len(org.CertificateAuthorities) > 0 {
		// Default to the first CA in org, if it is defined
		c.caID = org.CertificateAuthorities[0]
	}
	if c.caID != "" {
		ca, ok := ctx.IdentityConfig().CAConfig(c.caID)
		if !ok {
			return nil, fmt.Errorf("non-existent CA: '%s'", c.caID)
		}
		c.caName = ca.CAName
	}

	err = veiifyOrgCA(org, c.caID)
	if err != nil {
		return nil, err
	}

	return c, nil
}

func initClientFromOptions(clientProvider context.ClientProvider, opts ...ClientOption) (fab.ClientContext, *Client, error) {
	o := clientOptions{}
	for _, param := range opts {
		err := param(&o)
		if err != nil {
			return nil, nil, errors.WithMessage(err, "failed to create Client")
		}
	}

	ctx, err := clientProvider()
	if err != nil {
		return nil, nil, errors.WithMessage(err, "failed to create Client")
	}

	c := Client{
		ctx:     ctx,
		orgName: o.orgName,
		caID:    o.caID,
	}

	return ctx, &c, nil
}

func veiifyOrgCA(org fab.OrganizationConfig, caID string) error {
	if caID == "" {
		return nil
	}
	for _, name := range org.CertificateAuthorities {
		if caID == name {
			return nil
		}
	}
	return fmt.Errorf("ca: '%s' doesn't belong to organization: '%s'", caID, org.MSPID)
}

func newCAClient(ctx context.Client, orgName string, caID string) (mspapi.CAClient, error) {

	caClient, err := msp.NewCAClient(orgName, ctx, msp.WithCAInstance(caID))
	if err != nil {
		return nil, errors.WithMessage(err, "failed to create CA Client")
	}

	return caClient, nil
}

// CreateIdentity creates a new identity with the Fabric CA server. An enrollment secret is returned which can then be used,
// along with the enrollment ID, to enroll a new identity.
//  Parameters:
//  request holds info about identity
//
//  Returns:
//  Return identity info including the secret
func (c *Client) CreateIdentity(request *IdentityRequest) (*IdentityResponse, error) {

	ca, err := newCAClient(c.ctx, c.orgName, c.caID)
	if err != nil {
		return nil, err
	}

	var attrs []mspapi.Attribute
	for i := range request.Attributes {
		attrs = append(attrs, mspapi.Attribute{Name: request.Attributes[i].Name, Value: request.Attributes[i].Value, ECert: request.Attributes[i].ECert})
	}

	req := &mspapi.IdentityRequest{
		ID:             request.ID,
		Type:           request.Type,
		MaxEnrollments: request.MaxEnrollments,
		Affiliation:    request.Affiliation,
		Attributes:     attrs,
		CAName:         request.CAName,
		Secret:         request.Secret,
	}

	response, err := ca.CreateIdentity(req)
	if err != nil {
		return nil, err
	}

	return getIdentityResponse(response), nil
}

// ModifyIdentity modifies identity with the Fabric CA server.
//  Parameters:
//  request holds info about identity
//
//  Returns:
//  Return updated identity info
func (c *Client) ModifyIdentity(request *IdentityRequest) (*IdentityResponse, error) {

	ca, err := newCAClient(c.ctx, c.orgName, c.caID)
	if err != nil {
		return nil, err
	}

	var attrs []mspapi.Attribute
	for i := range request.Attributes {
		attrs = append(attrs, mspapi.Attribute{Name: request.Attributes[i].Name, Value: request.Attributes[i].Value, ECert: request.Attributes[i].ECert})
	}

	req := &mspapi.IdentityRequest{
		ID:             request.ID,
		Type:           request.Type,
		MaxEnrollments: request.MaxEnrollments,
		Affiliation:    request.Affiliation,
		Attributes:     attrs,
		CAName:         request.CAName,
		Secret:         request.Secret,
	}

	response, err := ca.ModifyIdentity(req)
	if err != nil {
		return nil, err
	}

	return getIdentityResponse(response), nil
}

// RemoveIdentity removes identity with the Fabric CA server.
//  Parameters:
//  request holds info about identity to be removed
//
//  Returns:
//  Return removed identity info
func (c *Client) RemoveIdentity(request *RemoveIdentityRequest) (*IdentityResponse, error) {

	ca, err := newCAClient(c.ctx, c.orgName, c.caID)
	if err != nil {
		return nil, err
	}

	req := &mspapi.RemoveIdentityRequest{
		ID:     request.ID,
		Force:  request.Force,
		CAName: request.CAName,
	}

	response, err := ca.RemoveIdentity(req)
	if err != nil {
		return nil, err
	}

	return getIdentityResponse(response), nil
}

// GetAllIdentities returns all identities that the caller is authorized to see
//  Parameters:
//  options holds optional request options
//  Returns:
//  Response containing identities
func (c *Client) GetAllIdentities(opts ...RequestOption) ([]*IdentityResponse, error) {

	o, err := c.prepareRequestOptsFromOptions(opts...)
	if err != nil {
		return nil, err
	}

	ca, err := newCAClient(c.ctx, c.orgName, c.caID)
	if err != nil {
		return nil, err
	}

	responses, err := ca.GetAllIdentities(o.caName)
	if err != nil {
		return nil, err
	}

	return getIdentityResponses(responses), nil

}

// GetIdentity retrieves identity information.
//  Parameters:
//  ID is required identity ID
//  options holds optional request options
//
//  Returns:
//  Response containing identity information
func (c *Client) GetIdentity(ID string, opts ...RequestOption) (*IdentityResponse, error) {

	o, err := c.prepareRequestOptsFromOptions(opts...)
	if err != nil {
		return nil, err
	}

	ca, err := newCAClient(c.ctx, c.orgName, c.caID)
	if err != nil {
		return nil, err
	}

	response, err := ca.GetIdentity(ID, o.caName)
	if err != nil {
		return nil, err
	}

	return getIdentityResponse(response), nil

}

func getIdentityResponse(response *mspapi.IdentityResponse) *IdentityResponse {

	var attributes []Attribute
	for i := range response.Attributes {
		attributes = append(attributes, Attribute{Name: response.Attributes[i].Name, Value: response.Attributes[i].Value, ECert: response.Attributes[i].ECert})
	}

	res := &IdentityResponse{ID: response.ID,
		Affiliation:    response.Affiliation,
		Type:           response.Type,
		Attributes:     attributes,
		MaxEnrollments: response.MaxEnrollments,
		Secret:         response.Secret,
		CAName:         response.CAName,
	}

	return res
}

func getIdentityResponses(responses []*mspapi.IdentityResponse) []*IdentityResponse {

	ret := make([]*IdentityResponse, len(responses))
	for i, r := range responses {
		ret[i] = getIdentityResponse(r)
	}

	return ret
}

// Enroll enrolls a registered user in order to receive a signed X509 certificate.
// A new key pair is generated for the user. The private key and the
// enrollment certificate issued by the CA are stored in SDK stores.
// They can be retrieved by calling IdentityManager.GetSigningIdentity().
//  Parameters:
//  enrollmentID enrollment ID of a registered user
//  opts are optional enrollment options
//
//  Returns:
//  an error if enrollment fails
func (c *Client) Enroll(enrollmentID string, opts ...EnrollmentOption) error {

	eo := enrollmentOptions{}
	for _, param := range opts {
		err := param(&eo)
		if err != nil {
			return errors.WithMessage(err, "failed to enroll")
		}
	}

	ca, err := newCAClient(c.ctx, c.orgName, c.caID)
	if err != nil {
		return err
	}

	req := &mspapi.EnrollmentRequest{
		Name:    enrollmentID,
		Secret:  eo.secret,
		CAName:  c.caName,
		Profile: eo.profile,
		Type:    eo.typ,
		Label:   eo.label,
		CSR:     createCSRInfo(eo.csr),
	}

	if req.CAName == "" {
		req.CAName = c.caName
	}

	if len(eo.attrReqs) > 0 {
		attrs := make([]*mspapi.AttributeRequest, 0)
		for _, attr := range eo.attrReqs {
			attrs = append(attrs, &mspapi.AttributeRequest{Name: attr.Name, Optional: attr.Optional})
		}
		req.AttrReqs = attrs
	}

	return ca.Enroll(req)
}

// Reenroll reenrolls an enrolled user in order to obtain a new signed X509 certificate
//  Parameters:
//  enrollmentID enrollment ID of a registered user
//
//  Returns:
//  an error if re-enrollment fails
func (c *Client) Reenroll(enrollmentID string, opts ...EnrollmentOption) error {
	eo := enrollmentOptions{}
	for _, param := range opts {
		err := param(&eo)
		if err != nil {
			return errors.WithMessage(err, "failed to enroll")
		}
	}

	ca, err := newCAClient(c.ctx, c.orgName, c.caID)
	if err != nil {
		return err
	}

	req := &mspapi.ReenrollmentRequest{
		Name:    enrollmentID,
		Profile: eo.profile,
		Label:   eo.label,
		CAName:  c.caName,
		CSR:     createCSRInfo(eo.csr),
	}

	if req.CAName == "" {
		req.CAName = c.caName
	}
	if len(eo.attrReqs) > 0 {
		attrs := make([]*mspapi.AttributeRequest, 0)
		for _, attr := range eo.attrReqs {
			attrs = append(attrs, &mspapi.AttributeRequest{Name: attr.Name, Optional: attr.Optional})
		}
		req.AttrReqs = attrs
	}
	return ca.Reenroll(req)
}

// Register registers a User with the Fabric CA
//  Parameters:
//  request is registration request
//
//  Returns:
//  enrolment secret
func (c *Client) Register(request *RegistrationRequest) (string, error) {
	ca, err := newCAClient(c.ctx, c.orgName, c.caID)
	if err != nil {
		return "", err
	}

	var a []mspapi.Attribute
	for i := range request.Attributes {
		a = append(a, mspapi.Attribute{Name: request.Attributes[i].Name, Value: request.Attributes[i].Value, ECert: request.Attributes[i].ECert})
	}

	r := mspapi.RegistrationRequest{
		Name:           request.Name,
		Type:           request.Type,
		MaxEnrollments: request.MaxEnrollments,
		Affiliation:    request.Affiliation,
		Attributes:     a,
		CAName:         request.CAName,
		Secret:         request.Secret,
	}
	return ca.Register(&r)
}

// Revoke revokes a User with the Fabric CA
//  Parameters:
//  request is revocation request
//
//  Returns:
//  revocation response
func (c *Client) Revoke(request *RevocationRequest) (*RevocationResponse, error) {
	ca, err := newCAClient(c.ctx, c.orgName, c.caID)
	if err != nil {
		return nil, err
	}
	req := mspapi.RevocationRequest(*request)
	resp, err := ca.Revoke(&req)
	if err != nil {
		return nil, err
	}
	var revokedCerts []RevokedCert
	for i := range resp.RevokedCerts {
		revokedCerts = append(
			revokedCerts,
			RevokedCert{
				Serial: resp.RevokedCerts[i].Serial,
				AKI:    resp.RevokedCerts[i].AKI,
			})
	}

	return &RevocationResponse{
		RevokedCerts: revokedCerts,
		CRL:          resp.CRL,
	}, nil
}

// GetCAInfo returns generic CA information
func (c *Client) GetCAInfo() (*GetCAInfoResponse, error) {
	ca, err := newCAClient(c.ctx, c.orgName, c.caID)
	if err != nil {
		return nil, err
	}

	resp, err := ca.GetCAInfo()
	if err != nil {
		return nil, err
	}

	return &GetCAInfoResponse{CAName: resp.CAName, CAChain: resp.CAChain[:], IssuerPublicKey: resp.IssuerPublicKey[:], IssuerRevocationPublicKey: resp.IssuerRevocationPublicKey[:], Version: resp.Version}, nil
}

// GetSigningIdentity returns signing identity for id
//  Parameters:
//  id is user id
//
//  Returns:
//  signing identity
func (c *Client) GetSigningIdentity(id string) (mspctx.SigningIdentity, error) {
	im, _ := c.ctx.IdentityManager(c.orgName)
	si, err := im.GetSigningIdentity(id)
	if err != nil {
		if err == mspctx.ErrUserNotFound {
			return nil, ErrUserNotFound
		}
		return nil, err
	}
	return si, nil
}

// CreateSigningIdentity creates a signing identity with the given options
func (c *Client) CreateSigningIdentity(opts ...mspctx.SigningIdentityOption) (mspctx.SigningIdentity, error) {
	im, _ := c.ctx.IdentityManager(c.orgName)
	return im.CreateSigningIdentity(opts...)
}

//prepareRequestOptsFromOptions reads request options from Option array
func (c *Client) prepareRequestOptsFromOptions(opts ...RequestOption) (requestOptions, error) {
	o := requestOptions{}
	for _, option := range opts {
		err := option(&o)
		if err != nil {
			return o, errors.WithMessage(err, "Failed to read opts")
		}
	}
	if o.caName == "" {
		o.caName = c.caName
	}
	return o, nil
}

// GetAffiliation returns information about the requested affiliation
func (c *Client) GetAffiliation(affiliation string, opts ...RequestOption) (*AffiliationResponse, error) {

	// Read request options
	o, err := c.prepareRequestOptsFromOptions(opts...)
	if err != nil {
		return nil, err
	}

	ca, err := newCAClient(c.ctx, c.orgName, c.caID)
	if err != nil {
		return nil, err
	}

	r, err := ca.GetAffiliation(affiliation, o.caName)
	if err != nil {
		return nil, err
	}

	resp := &AffiliationResponse{CAName: r.CAName, AffiliationInfo: AffiliationInfo{}}
	err = fillAffiliationInfo(&resp.AffiliationInfo, r.Name, r.Affiliations, r.Identities)

	return resp, err
}

// GetAllAffiliations returns all affiliations that the caller is authorized to see
func (c *Client) GetAllAffiliations(opts ...RequestOption) (*AffiliationResponse, error) {
	// Read request options
	o, err := c.prepareRequestOptsFromOptions(opts...)
	if err != nil {
		return nil, err
	}

	ca, err := newCAClient(c.ctx, c.orgName, c.caID)
	if err != nil {
		return nil, err
	}

	r, err := ca.GetAllAffiliations(o.caName)
	if err != nil {
		return nil, err
	}

	resp := &AffiliationResponse{CAName: r.CAName, AffiliationInfo: AffiliationInfo{}}
	err = fillAffiliationInfo(&resp.AffiliationInfo, r.Name, r.Affiliations, r.Identities)

	return resp, err
}

// AddAffiliation adds a new affiliation to the server
func (c *Client) AddAffiliation(request *AffiliationRequest) (*AffiliationResponse, error) {
	ca, err := newCAClient(c.ctx, c.orgName, c.caID)
	if err != nil {
		return nil, err
	}

	req := &mspapi.AffiliationRequest{
		Name:   request.Name,
		Force:  request.Force,
		CAName: request.CAName,
	}

	r, err := ca.AddAffiliation(req)
	if err != nil {
		return nil, err
	}

	resp := &AffiliationResponse{CAName: r.CAName, AffiliationInfo: AffiliationInfo{}}
	err = fillAffiliationInfo(&resp.AffiliationInfo, r.Name, r.Affiliations, r.Identities)

	return resp, err
}

// ModifyAffiliation renames an existing affiliation on the server
func (c *Client) ModifyAffiliation(request *ModifyAffiliationRequest) (*AffiliationResponse, error) {
	ca, err := newCAClient(c.ctx, c.orgName, c.caID)
	if err != nil {
		return nil, err
	}

	req := &mspapi.ModifyAffiliationRequest{
		NewName: request.NewName,
		AffiliationRequest: mspapi.AffiliationRequest{
			Name:   request.Name,
			Force:  request.Force,
			CAName: request.CAName,
		},
	}

	r, err := ca.ModifyAffiliation(req)
	if err != nil {
		return nil, err
	}

	resp := &AffiliationResponse{CAName: r.CAName, AffiliationInfo: AffiliationInfo{}}
	err = fillAffiliationInfo(&resp.AffiliationInfo, r.Name, r.Affiliations, r.Identities)

	return resp, err
}

// RemoveAffiliation removes an existing affiliation from the server
func (c *Client) RemoveAffiliation(request *AffiliationRequest) (*AffiliationResponse, error) {
	ca, err := newCAClient(c.ctx, c.orgName, c.caID)
	if err != nil {
		return nil, err
	}

	req := &mspapi.AffiliationRequest{
		Name:   request.Name,
		Force:  request.Force,
		CAName: request.CAName,
	}

	r, err := ca.RemoveAffiliation(req)
	if err != nil {
		return nil, err
	}

	resp := &AffiliationResponse{CAName: r.CAName, AffiliationInfo: AffiliationInfo{}}
	err = fillAffiliationInfo(&resp.AffiliationInfo, r.Name, r.Affiliations, r.Identities)

	return resp, err
}

func fillAffiliationInfo(info *AffiliationInfo, name string, affiliations []mspapi.AffiliationInfo, identities []mspapi.IdentityInfo) error {
	info.Name = name

	// Add identities which have this affiliation
	idents := []IdentityInfo{}
	for _, identity := range identities {
		idents = append(idents, IdentityInfo{ID: identity.ID, Type: identity.Type, Affiliation: identity.Affiliation, Attributes: getAllAttributes(identity.Attributes), MaxEnrollments: identity.MaxEnrollments})
	}
	if len(idents) > 0 {
		info.Identities = idents
	}

	// Create child affiliations (if any)
	children := []AffiliationInfo{}
	for _, aff := range affiliations {
		childAff := AffiliationInfo{Name: aff.Name}
		err := fillAffiliationInfo(&childAff, aff.Name, aff.Affiliations, aff.Identities)
		if err != nil {
			return err
		}
		children = append(children, childAff)
	}
	if len(children) > 0 {
		info.Affiliations = children
	}
	return nil
}

func createCSRInfo(csr *CSRInfo) *mspapi.CSRInfo {
	if csr == nil {
		// csr is not obrigatory, so we can return nil
		return nil
	}

	return &mspapi.CSRInfo{
		CN:    csr.CN,
		Hosts: csr.Hosts,
	}
}

func getAllAttributes(attrs []mspapi.Attribute) []Attribute {
	attriburtes := []Attribute{}
	for _, attr := range attrs {
		attriburtes = append(attriburtes, Attribute{Name: attr.Name, Value: attr.Value, ECert: attr.ECert})
	}

	return attriburtes
}
