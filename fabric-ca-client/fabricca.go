/*
Copyright SecureKey Technologies Inc. All Rights Reserved.


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

package fabricca

import (
	"fmt"
	"os"

	"github.com/hyperledger/fabric-ca/api"
	fabric_ca "github.com/hyperledger/fabric-ca/lib"
	"github.com/hyperledger/fabric-sdk-go/config"

	"github.com/op/go-logging"
)

var logger = logging.MustGetLogger("fabric_sdk_go")

// Services ...
type Services interface {
	Enroll(enrollmentID string, enrollmentSecret string) ([]byte, []byte, error)
}

type services struct {
	fabricCAClient *fabric_ca.Client
}

// NewFabricCAClient ...
/**
 * @param {string} clientConfigFile for fabric-ca services"
 */
func NewFabricCAClient() (Services, error) {
	configPath, err := config.GetFabricCAClientPath()
	if err != nil {
		return nil, fmt.Errorf("error setting up fabric-ca configurations: %s", err.Error())
	}
	//Remove temporary config file after setup
	defer os.Remove(configPath)
	// Create new Fabric-ca client with configs
	c, err := fabric_ca.NewClient(configPath)
	if err != nil {
		return nil, fmt.Errorf("New fabricCAClient failed: %s", err)
	}

	fabricCAClient := &services{fabricCAClient: c}
	logger.Infof("Constructed fabricCAClient instance: %v", fabricCAClient)

	return fabricCAClient, nil
}

// Enroll ...
/**
 * Enroll a registered user in order to receive a signed X509 certificate
 * @param {string} enrollmentID The registered ID to use for enrollment
 * @param {string} enrollmentSecret The secret associated with the enrollment ID
 * @returns {[]byte} X509 certificate
 * @returns {[]byte} private key
 */
func (fabricCAServices *services) Enroll(enrollmentID string, enrollmentSecret string) ([]byte, []byte, error) {
	if enrollmentID == "" {
		return nil, nil, fmt.Errorf("enrollmentID is empty")
	}
	if enrollmentSecret == "" {
		return nil, nil, fmt.Errorf("enrollmentSecret is empty")
	}
	req := &api.EnrollmentRequest{
		Name:   enrollmentID,
		Secret: enrollmentSecret,
	}
	id, err := fabricCAServices.fabricCAClient.Enroll(req)
	if err != nil {
		return nil, nil, fmt.Errorf("Enroll failed: %s", err)
	}
	return id.GetECert().Key(), id.GetECert().Cert(), nil
}
