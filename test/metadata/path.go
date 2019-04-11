/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package metadata

import (
	"errors"
	"go/build"
	"os"
	"path/filepath"
	"sync"
)

var projectPathOnce sync.Once

// GetProjectPath returns the path to the source of fabric-sdk-go. More specifically, this function searches for the
// directory containing the project's go.mod file. This function must only be called from tests.
func GetProjectPath() string {
	projectPathOnce.Do(func() {
		ProjectPath = getProjectPath()
	})
	return ProjectPath
}

func getProjectPath() string {
	if len(ProjectPath) > 0 {
		return filepath.Clean(ProjectPath)
	}

	pwd, err := os.Getwd()
	if err != nil {
		goPath := goPath()
		return filepath.Join(goPath, "src", Project)
	}
	pwd = filepath.Clean(pwd)

	modDir, err := findParentModule(pwd)
	if err != nil {
		return pwd
	}

	return modDir
}

func findParentModule(wd string) (string, error) {
	for {
		modPath := filepath.Join(wd, "go.mod")
		modExists, err := fileExists(modPath)
		if err != nil {
			return "", err
		}

		if modExists {
			return wd, nil
		}

		pd := filepath.Dir(wd)
		if wd == pd {
			break
		}
		wd = pd
	}
	return "", errors.New("project module was not found")
}

func fileExists(path string) (bool, error) {
	fi, err := os.Stat(path)
	if os.IsNotExist(err) {
		return false, nil
	} else if err != nil {
		return false, err
	}

	if fi.IsDir() {
		return false, nil
	}

	return true, nil
}

// goPath returns the current GOPATH. If the system
// has multiple GOPATHs then the first is used.
func goPath() string {
	gpDefault := build.Default.GOPATH
	gps := filepath.SplitList(gpDefault)

	return gps[0]
}
