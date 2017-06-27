/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package integration

import (
	"fmt"
	"io/ioutil"
	"math/rand"
	"path/filepath"
	"time"

	api "github.com/hyperledger/fabric-sdk-go/api"
	defaultImpl "github.com/hyperledger/fabric-sdk-go/fabric-txn/defaultImpl"
)

// GetOrdererAdmin ...
func GetOrdererAdmin(c api.FabricClient) (api.User, error) {
	keyDir := "ordererOrganizations/example.com/users/Admin@example.com/keystore"
	certDir := "ordererOrganizations/example.com/users/Admin@example.com/signcerts"
	return getDefaultImplPreEnrolledUser(c, keyDir, certDir, "ordererAdmin")
}

// GetAdmin ...
func GetAdmin(c api.FabricClient, userOrg string) (api.User, error) {
	keyDir := fmt.Sprintf("peerOrganizations/%s.example.com/users/Admin@%s.example.com/keystore", userOrg, userOrg)
	certDir := fmt.Sprintf("peerOrganizations/%s.example.com/users/Admin@%s.example.com/signcerts", userOrg, userOrg)
	username := fmt.Sprintf("peer%sAdmin", userOrg)
	return getDefaultImplPreEnrolledUser(c, keyDir, certDir, username)
}

// GenerateRandomID generates random ID
func GenerateRandomID() string {
	rand.Seed(time.Now().UnixNano())
	return randomString(10)
}

// Utility to create random string of strlen length
func randomString(strlen int) string {
	const chars = "abcdefghijklmnopqrstuvwxyz0123456789"
	result := make([]byte, strlen)
	for i := 0; i < strlen; i++ {
		result[i] = chars[rand.Intn(len(chars))]
	}
	return string(result)
}

// GetDefaultImplPreEnrolledUser ...
func getDefaultImplPreEnrolledUser(client api.FabricClient, keyDir string, certDir string, username string) (api.User, error) {
	privateKeyDir := filepath.Join(client.GetConfig().GetCryptoConfigPath(), keyDir)
	privateKeyPath, err := getFirstPathFromDir(privateKeyDir)
	if err != nil {
		return nil, fmt.Errorf("Error finding the private key path: %v", err)
	}

	enrollmentCertDir := filepath.Join(client.GetConfig().GetCryptoConfigPath(), certDir)
	enrollmentCertPath, err := getFirstPathFromDir(enrollmentCertDir)
	if err != nil {
		return nil, fmt.Errorf("Error finding the enrollment cert path: %v", err)
	}

	return defaultImpl.NewPreEnrolledUser(client, privateKeyPath, enrollmentCertPath, username)
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
		fmt.Printf("Reading file %s", fullName)
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

// GetUser ...
func GetUser(c api.FabricClient, userOrg string) (api.User, error) {
	keyDir := fmt.Sprintf("peerOrganizations/%s.example.com/users/User1@%s.example.com/keystore", userOrg, userOrg)
	certDir := fmt.Sprintf("peerOrganizations/%s.example.com/users/User1@%s.example.com/signcerts", userOrg, userOrg)
	username := fmt.Sprintf("peer%sUser1", userOrg)
	return getDefaultImplPreEnrolledUser(c, keyDir, certDir, username)
}
