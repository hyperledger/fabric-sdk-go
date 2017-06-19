/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package util

import (
	"fmt"
	"io/ioutil"
	"path/filepath"

	api "github.com/hyperledger/fabric-sdk-go/api"
	sdkUser "github.com/hyperledger/fabric-sdk-go/pkg/fabric-client/user"

	"github.com/hyperledger/fabric-ca/util"
	fabricCAClient "github.com/hyperledger/fabric-sdk-go/pkg/fabric-ca-client"
)

// GetPreEnrolledUser ...
func GetPreEnrolledUser(c api.FabricClient, keyDir string, certDir string, username string) (api.User, error) {

	privateKeyDir := filepath.Join(c.GetConfig().GetCryptoConfigPath(), keyDir)
	privateKeyPath, err := getFirstPathFromDir(privateKeyDir)
	if err != nil {
		return nil, fmt.Errorf("Error finding the private key path: %v", err)
	}
	privateKey, err := util.ImportBCCSPKeyFromPEM(privateKeyPath, c.GetCryptoSuite(), true)
	if err != nil {
		return nil, fmt.Errorf("Error importing private key: %v", err)
	}

	enrollmentCertDir := filepath.Join(c.GetConfig().GetCryptoConfigPath(), certDir)
	enrollmentCertPath, err := getFirstPathFromDir(enrollmentCertDir)
	if err != nil {
		return nil, fmt.Errorf("Error finding the enrollment cert path: %v", err)
	}
	enrollmentCert, err := ioutil.ReadFile(enrollmentCertPath)
	if err != nil {
		return nil, fmt.Errorf("Error reading from the enrollment cert path: %v", err)
	}

	user := sdkUser.NewUser(username)
	user.SetEnrollmentCertificate(enrollmentCert)
	user.SetPrivateKey(privateKey)

	return user, nil
}

// Gets the first path from the dir directory
func getFirstPathFromDir(dir string) (string, error) {

	files, err := ioutil.ReadDir(dir)
	if err != nil {
		return "", fmt.Errorf("Could not read directory %s, err %s", err, dir)
	}

	for _, p := range files {
		if p.IsDir() {
			continue
		}

		fullName := filepath.Join(dir, string(filepath.Separator), p.Name())
		logger.Debugf("Reading file %s", fullName)
	}

	for _, f := range files {
		if f.IsDir() {
			continue
		}

		fullName := filepath.Join(dir, string(filepath.Separator), f.Name())
		return fullName, nil
	}

	return "", fmt.Errorf("No paths found in directory: %s", dir)
}

// GetMember ...
func GetMember(client api.FabricClient, name string, pwd string) (api.User, error) {
	user, err := client.LoadUserFromStateStore(name)
	if err != nil {
		return nil, fmt.Errorf("Error loading user from store: %v", err)
	}
	if user == nil {
		fabricCAClient, err := fabricCAClient.NewFabricCAClient(client.GetConfig())
		if err != nil {
			return nil, fmt.Errorf("NewFabricCAClient return error: %v", err)
		}
		key, cert, err := fabricCAClient.Enroll(name, pwd)
		if err != nil {
			return nil, fmt.Errorf("Enroll return error: %v", err)
		}
		user := sdkUser.NewUser(name)
		user.SetPrivateKey(key)
		user.SetEnrollmentCertificate(cert)
		err = client.SaveUserToStateStore(user, false)
		if err != nil {
			return nil, fmt.Errorf("client.SaveUserToStateStore return error: %v", err)
		}
	}
	return user, nil
}
