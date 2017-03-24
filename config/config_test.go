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

package config

import (
	"fmt"
	"os"
	"testing"

	"github.com/spf13/viper"
)

func TestGetPeersConfig(t *testing.T) {
	pc := GetPeersConfig()

	for _, value := range pc {
		if value.Host == "" {
			t.Fatalf("Host value is empty")
		}
		if value.Port == "" {
			t.Fatalf("Port value is empty")
		}
		if value.Port == "" {
			t.Fatalf("EventHost value is empty")
		}
		if value.Port == "" {
			t.Fatalf("EventPort value is empty")
		}

	}

}

// Test case to create a new viper instance to prevent conflict with existing
// viper instances in applications that use the SDK
func TestMultipleVipers(t *testing.T) {
	viper.SetConfigFile("./test.yaml")
	err := viper.ReadInConfig()
	if err != nil {
		fmt.Println(err.Error())
	}
	testValue1 := viper.GetString("test.testKey")
	// Read initial value from test.yaml
	if testValue1 != "testvalue" {
		t.Fatalf("Expected testValue before config initialization got: %s", testValue1)
	}
	// initialize go sdk
	err = InitConfig("../integration_test/test_resources/config/config_test.yaml")
	if err != nil {
		fmt.Println(err.Error())
	}
	// Make sure initial value is unaffected
	testValue2 := viper.GetString("test.testKey")
	if testValue2 != "testvalue" {
		t.Fatalf("Expected testvalue after config initialization")
	}
	// Make sure Go SDK config is unaffected
	testValue3 := myViper.GetString("client.peers.peer1.host")
	if testValue3 != "localhost" {
		t.Fatalf("Expected existing config value to remain unchanged")
	}
}

func TestMain(m *testing.M) {
	err := InitConfig("../integration_test/test_resources/config/config_test.yaml")
	if err != nil {
		fmt.Println(err.Error())
	}
	os.Exit(m.Run())
}
