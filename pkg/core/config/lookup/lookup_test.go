/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package lookup

import (
	"testing"

	"os"

	"time"

	"strings"

	"reflect"

	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/fab"
	"github.com/hyperledger/fabric-sdk-go/pkg/core/mocks"
	"github.com/mitchellh/mapstructure"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
)

var sampleConfigFile = "../testdata/config_test.yaml"

const orgChannelID = "orgchannel"

var backend *mocks.MockConfigBackend

func TestMain(m *testing.M) {
	setupCustomBackend()
	r := m.Run()
	os.Exit(r)
}

func TestGetBool(t *testing.T) {
	testLookup := New(backend)
	assert.True(t, testLookup.GetBool("key.bool.true"), "expected lookup to return true")
	assert.False(t, testLookup.GetBool("key.bool.false"), "expected lookup to return false")
	assert.False(t, testLookup.GetBool("key.bool.invalid"), "expected lookup to return false for invalid value")
	assert.False(t, testLookup.GetBool("key.bool.notexisting"), "expected lookup to return false for not existing value")
}

func TestGetInt(t *testing.T) {
	testLookup := New(backend)
	assert.True(t, testLookup.GetInt("key.int.positive") == 5, "expected lookup to return valid positive value")
	assert.True(t, testLookup.GetInt("key.int.negative") == -5, "expected lookup to return valid negative value")
	assert.True(t, testLookup.GetInt("key.int.invalid") == 0, "expected lookup to return 0")
	assert.True(t, testLookup.GetInt("key.int.not.existing") == 0, "expected lookup to return 0")
}

func TestGetString(t *testing.T) {
	testLookup := New(backend)
	assert.True(t, testLookup.GetString("key.string.valid") == "valid-string", "expected lookup to return valid string value")
	assert.True(t, testLookup.GetString("key.string.valid.lower.case") == "valid-string", "expected lookup to return valid string value")
	assert.True(t, testLookup.GetString("key.string.valid.upper.case") == "VALID-STRING", "expected lookup to return valid string value")
	assert.True(t, testLookup.GetString("key.string.valid.mixed.case") == "VaLiD-StRiNg", "expected lookup to return valid string value")
	assert.True(t, testLookup.GetString("key.string.empty") == "", "expected lookup to return empty string value")
	assert.True(t, testLookup.GetString("key.string.nil") == "", "expected lookup to return empty string value")
	assert.True(t, testLookup.GetString("key.string.number") == "1234", "expected lookup to return valid string value")
	assert.True(t, testLookup.GetString("key.string.not existing") == "", "expected lookup to return empty string value")
}

func TestGetLowerString(t *testing.T) {
	testLookup := New(backend)
	assert.True(t, testLookup.GetLowerString("key.string.valid") == "valid-string", "expected lookup to return valid lowercase string value")
	assert.True(t, testLookup.GetLowerString("key.string.valid.lower.case") == "valid-string", "expected lookup to return valid lowercase string value")
	assert.True(t, testLookup.GetLowerString("key.string.valid.upper.case") == "valid-string", "expected lookup to return valid lowercase string value")
	assert.True(t, testLookup.GetLowerString("key.string.valid.mixed.case") == "valid-string", "expected lookup to return valid lowercase string value")
	assert.True(t, testLookup.GetLowerString("key.string.empty") == "", "expected lookup to return empty string value")
	assert.True(t, testLookup.GetLowerString("key.string.nil") == "", "expected lookup to return empty string value")
	assert.True(t, testLookup.GetLowerString("key.string.number") == "1234", "expected lookup to return valid string value")
	assert.True(t, testLookup.GetLowerString("key.string.not existing") == "", "expected lookup to return empty string value")
}

func TestGetDuration(t *testing.T) {
	testLookup := New(backend)
	assert.True(t, testLookup.GetDuration("key.duration.valid.hour").String() == (24*time.Hour).String(), "expected valid time value")
	assert.True(t, testLookup.GetDuration("key.duration.valid.minute").String() == (24*time.Minute).String(), "expected valid time value")
	assert.True(t, testLookup.GetDuration("key.duration.valid.second").String() == (24*time.Second).String(), "expected valid time value")
	assert.True(t, testLookup.GetDuration("key.duration.valid.millisecond").String() == (24*time.Millisecond).String(), "expected valid time value")
	assert.True(t, testLookup.GetDuration("key.duration.valid.microsecond").String() == (24*time.Microsecond).String(), "expected valid time value")
	assert.True(t, testLookup.GetDuration("key.duration.valid.nanosecond").String() == (24*time.Nanosecond).String(), "expected valid time value")
	//default value tests
	assert.True(t, testLookup.GetDuration("key.duration.valid.not.existing").String() == (0*time.Second).String(), "expected valid default time value")
	assert.True(t, testLookup.GetDuration("key.duration.valid.invalid").String() == (0*time.Second).String(), "expected valid  default time value")
	assert.True(t, testLookup.GetDuration("key.duration.valid.nil").String() == (0*time.Second).String(), "expected valid  default time value")
	assert.True(t, testLookup.GetDuration("key.duration.valid.empty").String() == (0*time.Second).String(), "expected valid  default time value")
	//default when no time unit provided
	assert.True(t, testLookup.GetDuration("key.duration.valid.no.unit").String() == (12*time.Nanosecond).String(), "expected valid default time value with default unit")
}

func TestUnmarshal(t *testing.T) {
	testLookup := New(backend)

	//output struct
	networkConfig := fab.NetworkConfig{}
	testLookup.UnmarshalKey("channels", &networkConfig.Channels)

	assert.True(t, len(networkConfig.Channels) == 3)
	assert.True(t, len(networkConfig.Channels["mychannel"].Peers) == 1)
	assert.True(t, networkConfig.Channels["mychannel"].Policies.QueryChannelConfig.MinResponses == 1)
	assert.True(t, networkConfig.Channels["mychannel"].Policies.QueryChannelConfig.MaxTargets == 1)
	assert.True(t, networkConfig.Channels["mychannel"].Policies.QueryChannelConfig.RetryOpts.MaxBackoff.String() == (5*time.Second).String())
	assert.True(t, networkConfig.Channels["mychannel"].Policies.QueryChannelConfig.RetryOpts.InitialBackoff.String() == (500*time.Millisecond).String())
	assert.True(t, networkConfig.Channels["mychannel"].Policies.QueryChannelConfig.RetryOpts.BackoffFactor == 2.0)

}

func TestLookupUnmarshalAgainstViperUnmarshal(t *testing.T) {

	//new lookup
	testLookup := New(backend)
	//setup viper
	sampleViper := newViper()
	//viper network config
	networkConfigViper := fab.NetworkConfig{}
	//lookup network config
	networkConfig := fab.NetworkConfig{}

	/*
		TEST NETWORK CONFIG CLIENT
	*/
	//get client through backend lookup
	err := testLookup.UnmarshalKey("client", &networkConfig.Client)
	if err != nil {
		t.Fatal(err)
	}
	//get client from viper
	sampleViper.UnmarshalKey("client", &networkConfigViper.Client)
	//now compare
	assert.True(t, reflect.DeepEqual(&networkConfig.Client, &networkConfigViper.Client), "unmarshalled value from config lookup supposed to match unmarshalled value from viper")

	/*
		TEST NETWORK CONFIG ORDERERS
	*/
	//get orderers through backend lookup
	err = testLookup.UnmarshalKey("orderers", &networkConfig.Orderers)
	if err != nil {
		t.Fatal(err)
	}
	//get orderers from viper
	sampleViper.UnmarshalKey("orderers", &networkConfigViper.Orderers)
	//now compare
	assert.True(t, reflect.DeepEqual(&networkConfig.Orderers, &networkConfigViper.Orderers), "unmarshalled value from config lookup supposed to match unmarshalled value from viper")

	/*
		TEST NETWORK CONFIG CERTIFICATE AUTHORITIES
	*/
	//get certificate authorities through backend lookup
	err = testLookup.UnmarshalKey("certificateAuthorities", &networkConfig.CertificateAuthorities)
	if err != nil {
		t.Fatal(err)
	}
	//get certificate authorities from viper
	sampleViper.UnmarshalKey("certificateAuthorities", &networkConfigViper.CertificateAuthorities)
	//now compare
	assert.True(t, reflect.DeepEqual(&networkConfig.CertificateAuthorities, &networkConfigViper.CertificateAuthorities), "unmarshalled value from config lookup supposed to match unmarshalled value from viper")

	/*
		TEST NETWORK CONFIG ENTITY MATCHERS
	*/
	//get entityMatchers backend lookup
	err = testLookup.UnmarshalKey("entityMatchers", &networkConfig.EntityMatchers)
	if err != nil {
		t.Fatal(err)
	}
	//get entityMatchers from viper
	sampleViper.UnmarshalKey("entityMatchers", &networkConfigViper.EntityMatchers)
	//now compare
	assert.True(t, reflect.DeepEqual(&networkConfig.EntityMatchers, &networkConfigViper.EntityMatchers), "unmarshalled value from config lookup supposed to match unmarshalled value from viper")

	/*
		TEST NETWORK CONFIG ORGANIZATIONS
	*/
	//get organizations through backend lookup
	err = testLookup.UnmarshalKey("organizations", &networkConfig.Organizations)
	if err != nil {
		t.Fatal(err)
	}
	//get organizations from viper
	sampleViper.UnmarshalKey("organizations", &networkConfigViper.Organizations)
	//now compare
	assert.True(t, reflect.DeepEqual(&networkConfig.Organizations, &networkConfigViper.Organizations), "unmarshalled value from config lookup supposed to match unmarshalled value from viper")

	/*
		TEST NETWORK CONFIG CHANNELS
	*/
	//get channels through backend lookup
	err = testLookup.UnmarshalKey("channels", &networkConfig.Channels)
	if err != nil {
		t.Fatal(err)
	}
	//get channels from viper
	sampleViper.UnmarshalKey("channels", &networkConfigViper.Channels)
	//now compare
	assert.True(t, reflect.DeepEqual(&networkConfig.Channels, &networkConfigViper.Channels), "unmarshalled value from config lookup supposed to match unmarshalled value from viper")

	/*
		TEST NETWORK CONFIG PEERS
	*/
	//get peers through backend lookup
	err = testLookup.UnmarshalKey("peers", &networkConfig.Peers)
	if err != nil {
		t.Fatal(err)
	}
	//get peers from viper
	sampleViper.UnmarshalKey("peers", &networkConfigViper.Peers)
	//now compare
	assert.True(t, reflect.DeepEqual(&networkConfig.Peers, &networkConfigViper.Peers), "unmarshalled value from config lookup supposed to match unmarshalled value from viper")

	//Just to make sure that empty values are not being compared
	assert.True(t, len(networkConfigViper.Channels) > 0, "expected to get valid unmarshalled value")
	assert.True(t, len(networkConfigViper.Organizations) > 0, "expected to get valid unmarshalled value")
	assert.True(t, len(networkConfigViper.Orderers) > 0, "expected to get valid unmarshalled value")
	assert.True(t, len(networkConfigViper.Peers) > 0, "expected to get valid unmarshalled value")
	assert.True(t, len(networkConfigViper.EntityMatchers) > 0, "expected to get valid unmarshalled value")
	assert.True(t, networkConfigViper.Client.Organization != "", "expected to get valid unmarshalled value")

}

func setupCustomBackend() {
	backendMap := make(map[string]interface{})

	backendMap["key.bool.true"] = true
	backendMap["key.bool.false"] = false
	backendMap["key.bool.invalid"] = "INVALID"

	backendMap["key.int.positive"] = 5
	backendMap["key.int.negative"] = -5
	backendMap["key.int.invalid"] = "INVALID"

	backendMap["key.string.valid"] = "valid-string"
	backendMap["key.string.valid.mixed.case"] = "VaLiD-StRiNg"
	backendMap["key.string.valid.lower.case"] = "valid-string"
	backendMap["key.string.valid.upper.case"] = "VALID-STRING"
	backendMap["key.string.empty"] = ""
	backendMap["key.string.nil"] = nil
	backendMap["key.string.number"] = 1234

	backendMap["key.duration.valid.hour"] = "24h"
	backendMap["key.duration.valid.minute"] = "24m"
	backendMap["key.duration.valid.second"] = "24s"
	backendMap["key.duration.valid.millisecond"] = "24ms"
	backendMap["key.duration.valid.microsecond"] = "24µs"
	backendMap["key.duration.valid.nanosecond"] = "24ns"
	backendMap["key.duration.valid.no.unit"] = "12"
	backendMap["key.duration.invalid"] = "24XYZ"
	backendMap["key.duration.nil"] = nil
	backendMap["key.duration.empty"] = ""

	//test fab network config
	sampleViper := newViper()
	backendMap["client"] = sampleViper.Get("client")
	backendMap["channels"] = sampleViper.Get("channels")
	backendMap["certificateAuthorities"] = sampleViper.Get("certificateAuthorities")
	backendMap["entityMatchers"] = sampleViper.Get("entityMatchers")
	backendMap["organizations"] = sampleViper.Get("organizations")
	backendMap["orderers"] = sampleViper.Get("orderers")
	backendMap["peers"] = sampleViper.Get("peers")

	backend = &mocks.MockConfigBackend{KeyValueMap: backendMap}
}

func TestUnmarshalWithHookFunc(t *testing.T) {
	testLookup := New(backend)
	tamperPeerChannelConfig(backend)
	//output struct
	networkConfig := fab.NetworkConfig{}
	testLookup.UnmarshalKey("channels", &networkConfig.Channels, WithUnmarshalHookFunction(setTrueDefaultForPeerChannelConfig()))

	//Test if mandatory hook func is working as expected
	assert.True(t, len(networkConfig.Channels) == 3)
	assert.True(t, len(networkConfig.Channels["mychannel"].Peers) == 1)
	assert.True(t, networkConfig.Channels["mychannel"].Policies.QueryChannelConfig.MinResponses == 1)
	assert.True(t, networkConfig.Channels["mychannel"].Policies.QueryChannelConfig.MaxTargets == 1)
	assert.True(t, networkConfig.Channels["mychannel"].Policies.QueryChannelConfig.RetryOpts.MaxBackoff.String() == (5*time.Second).String())
	assert.True(t, networkConfig.Channels["mychannel"].Policies.QueryChannelConfig.RetryOpts.InitialBackoff.String() == (500*time.Millisecond).String())
	assert.True(t, networkConfig.Channels["mychannel"].Policies.QueryChannelConfig.RetryOpts.BackoffFactor == 2.0)

	//Test if custom hook func is working
	assert.True(t, len(networkConfig.Channels[orgChannelID].Peers) == 2)
	//test orgchannel peer1 (EndorsingPeer should be true as set, remaining should be default = true)
	orgChannelPeer1 := networkConfig.Channels[orgChannelID].Peers["peer0.org1.example.com"]
	assert.True(t, orgChannelPeer1.EndorsingPeer)
	assert.True(t, orgChannelPeer1.LedgerQuery)
	assert.True(t, orgChannelPeer1.EventSource)
	assert.True(t, orgChannelPeer1.ChaincodeQuery)

	//test orgchannel peer1 (EndorsingPeer should be false as set, remaining should be default = true)
	orgChannelPeer2 := networkConfig.Channels[orgChannelID].Peers["peer0.org2.example.com"]
	assert.False(t, orgChannelPeer2.EndorsingPeer)
	assert.True(t, orgChannelPeer2.LedgerQuery)
	assert.True(t, orgChannelPeer2.EventSource)
	assert.True(t, orgChannelPeer2.ChaincodeQuery)
}

func newViper() *viper.Viper {
	myViper := viper.New()
	replacer := strings.NewReplacer(".", "_")
	myViper.SetEnvKeyReplacer(replacer)
	myViper.SetConfigFile(sampleConfigFile)
	err := myViper.MergeInConfig()
	if err != nil {
		panic(err)
	}
	return myViper
}

func tamperPeerChannelConfig(backend *mocks.MockConfigBackend) {
	channelsMap := backend.KeyValueMap["channels"]
	orgChannel := map[string]interface{}{
		"orderers": []string{"orderer.example.com"},
		"peers": map[string]interface{}{
			"peer0.org1.example.com": map[string]interface{}{"endorsingpeer": true},
			"peer0.org2.example.com": map[string]interface{}{"endorsingpeer": false},
		},
	}
	(channelsMap.(map[string]interface{}))[orgChannelID] = orgChannel
}

func setTrueDefaultForPeerChannelConfig() mapstructure.DecodeHookFunc {
	return func(
		f reflect.Type,
		t reflect.Type,
		data interface{}) (interface{}, error) {

		//If target is of type 'fab.PeerChannelConfig', then only hook should work
		if t == reflect.TypeOf(fab.PeerChannelConfig{}) {
			dataMap, ok := data.(map[string]interface{})
			if ok {
				setDefault(dataMap, "endorsingpeer", true)
				setDefault(dataMap, "chaincodequery", true)
				setDefault(dataMap, "ledgerquery", true)
				setDefault(dataMap, "eventsource", true)

				return dataMap, nil
			}
		}
		return data, nil
	}
}

//setDefault sets default value provided to map if given key not found
func setDefault(dataMap map[string]interface{}, key string, defaultVal bool) {
	_, ok := dataMap[key]
	if !ok {
		dataMap[key] = true
	}
}
