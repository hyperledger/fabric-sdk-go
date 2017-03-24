/*
Copyright IBM Corp. 2017 All Rights Reserved.

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
	"net/url"

	"github.com/hyperledger/fabric-ca/api"
	"github.com/hyperledger/fabric-ca/lib/tls"
)

// ClientConfig is the fabric-ca client's config
type ClientConfig struct {
	Debug      bool   `def:"false" opt:"d" help:"Enable debug level logging"`
	URL        string `def:"http://localhost:7054" opt:"u" help:"URL of fabric-ca-server"`
	MSPDir     string `def:"msp" opt:"M" help:"Membership Service Provider directory"`
	TLS        tls.ClientTLSConfig
	Enrollment api.EnrollmentRequest
	CSR        api.CSRInfo
	ID         api.RegistrationRequest
}

// Enroll a client given the server's URL and the client's home directory.
// The URL may be of the form: http://user:pass@host:port where user and pass
// are the enrollment ID and secret, respectively.
func (c *ClientConfig) Enroll(rawurl, home string) (*EnrollmentResponse, error) {
	purl, err := url.Parse(rawurl)
	if err != nil {
		return nil, err
	}
	if purl.User != nil {
		name := purl.User.Username()
		secret, _ := purl.User.Password()
		c.Enrollment.Name = name
		c.Enrollment.Secret = secret
		purl.User = nil
	}
	c.URL = purl.String()
	c.TLS.Enabled = purl.Scheme == "https"
	c.Enrollment.CSR = &c.CSR
	client := &Client{HomeDir: home, Config: c}
	return client.Enroll(&c.Enrollment)
}
