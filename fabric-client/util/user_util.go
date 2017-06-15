/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package util

import (
	"fmt"
	"io/ioutil"
	"path/filepath"

	"github.com/hyperledger/fabric-ca/util"
	"github.com/hyperledger/fabric-sdk-go/config"
	fabricCAClient "github.com/hyperledger/fabric-sdk-go/fabric-ca-client"
	fc "github.com/hyperledger/fabric-sdk-go/fabric-client"
)

// GetPreEnrolledUser ...
func GetPreEnrolledUser(c fc.Client, keyDir string, certDir string, username string) (fc.User, error) {

	privateKeyDir := filepath.Join(config.GetCryptoConfigPath(), keyDir)
	privateKeyPath, err := getFirstPathFromDir(privateKeyDir)
	if err != nil {
		return nil, fmt.Errorf("Error finding the private key path: %v", err)
	}
	privateKey, err := util.ImportBCCSPKeyFromPEM(privateKeyPath, c.GetCryptoSuite(), true)
	if err != nil {
		return nil, fmt.Errorf("Error importing private key: %v", err)
	}

	enrollmentCertDir := filepath.Join(config.GetCryptoConfigPath(), certDir)
	enrollmentCertPath, err := getFirstPathFromDir(enrollmentCertDir)
	if err != nil {
		return nil, fmt.Errorf("Error finding the enrollment cert path: %v", err)
	}
	enrollmentCert, err := ioutil.ReadFile(enrollmentCertPath)
	if err != nil {
		return nil, fmt.Errorf("Error reading from the enrollment cert path: %v", err)
	}

	user := fc.NewUser(username)
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
func GetMember(client fc.Client, name string, pwd string) (fc.User, error) {
	user, err := client.LoadUserFromStateStore(name)
	if err != nil {
		return nil, fmt.Errorf("Error loading user from store: %v", err)
	}
	if user == nil {
		fabricCAClient, err := fabricCAClient.NewFabricCAClient()
		if err != nil {
			return nil, fmt.Errorf("NewFabricCAClient return error: %v", err)
		}
		key, cert, err := fabricCAClient.Enroll(name, pwd)
		if err != nil {
			return nil, fmt.Errorf("Enroll return error: %v", err)
		}
		user := fc.NewUser(name)
		user.SetPrivateKey(key)
		user.SetEnrollmentCertificate(cert)
		err = client.SaveUserToStateStore(user, false)
		if err != nil {
			return nil, fmt.Errorf("client.SaveUserToStateStore return error: %v", err)
		}
	}
	return user, nil
}
