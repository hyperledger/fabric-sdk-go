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
/*
Notice: This file has been modified for Hyperledger Fabric SDK Go usage.
Please review third_party pinning scripts and patches for more details.
*/

// StreamJSONArray scans the JSON stream associated with 'decoder' to find
// an array value associated with the json element at 'pathToArray'.
// It then calls the 'cb' callback function so that it can decode one element
// in the stream at a time.

package streamer

import (
	"encoding/json"
	"fmt"
	"os"
	"reflect"
	"strings"

	"github.com/cloudflare/cfssl/api"
	log "github.com/hyperledger/fabric-sdk-go/internal/github.com/hyperledger/fabric-ca/sdkpatch/logbridge"
	"github.com/pkg/errors"
)

// SearchElement defines the JSON arrays for which to search
type SearchElement struct {
	Path string
	CB   func(*json.Decoder) error
}

// StreamJSONArray searches the JSON stream for an array matching 'path'.
// For each element of this array, it streams one element at a time.
func StreamJSONArray(decoder *json.Decoder, path string, cb func(*json.Decoder) error) (bool, error) {
	ses := []SearchElement{
		SearchElement{Path: path, CB: cb},
		SearchElement{Path: "errors", CB: errCB},
	}
	return StreamJSON(decoder, ses)
}

// StreamJSON searches the JSON stream for arrays matching a search element.
// For each array that it finds, it streams them one element at a time.
func StreamJSON(decoder *json.Decoder, search []SearchElement) (bool, error) {
	js := &jsonStream{decoder: decoder, search: search, stack: []string{}}
	err := js.stream()
	return js.gotResults, err
}

type jsonStream struct {
	decoder    *json.Decoder
	search     []SearchElement
	stack      []string
	gotResults bool
}

func (js *jsonStream) stream() error {
	t, err := js.getToken()
	if err != nil {
		return err
	}
	if _, ok := t.(json.Delim); !ok {
		return nil
	}
	path := strings.Join(js.stack, ".")
	se := js.getSearchElement(path)
	d := fmt.Sprintf("%s", t)
	switch d {
	case "[":
		if se != nil {
			for js.decoder.More() {
				err = se.CB(js.decoder)
				if err != nil {
					return err
				}
				js.gotResults = true
			}
		}
		err = js.skipToDelim("]")
		if err != nil {
			return err
		}
	case "]":
		return errors.Errorf("Unexpected '%s'", d)
	case "{":
		if se != nil {
			return errors.Errorf("Expecting array for value of '%s'", path)
		}
		for {
			name, err := js.getNextName()
			if err != nil {
				return err
			}
			if name == "" {
				return nil
			}
			stack := js.stack
			js.stack = append(stack, name)
			err = js.stream()
			if err != nil {
				return err
			}
			js.stack = stack
		}
	case "}":
		return errors.Errorf("Unexpected '%s'", d)
	default:
		return errors.Errorf("unknown JSON delimiter: '%s'", d)
	}
	return nil
}

// Find a search element named 'path'
func (js *jsonStream) getSearchElement(path string) *SearchElement {
	for _, ele := range js.search {
		if ele.Path == path {
			return &ele
		}
	}
	return nil
}

// Skip over tokens until we hit the delimiter
func (js *jsonStream) skipToDelim(delim string) error {
	for {
		t, err := js.getToken()
		if err != nil {
			return err
		}
		// Skip anything that isn't a delimiter
		if _, ok := t.(json.Delim); !ok {
			continue
		}
		// It is a delimiter
		d := fmt.Sprintf("%s", t)
		if d == delim {
			return nil
		}
		switch d {
		case "[":
			err = js.skipToDelim("]")
		case "]":
			err = errors.Errorf("Expecting '%s' but found '%s'", delim, d)
		case "{":
			err = js.skipToDelim("}")
		case "}":
			err = errors.Errorf("Expecting '%s' but found '%s'", delim, d)
		default:
			err = errors.Errorf("unknown JSON delimiter: '%s'", d)
		}
		if err != nil {
			return err
		}
	}
}

func (js *jsonStream) getNextName() (string, error) {
	token, err := js.getToken()
	if err != nil {
		return "", err
	}
	switch v := token.(type) {
	case string:
		return v, nil
	case json.Delim:
		d := fmt.Sprintf("%s", v)
		if d == "}" {
			return "", nil
		}
		return "", errors.Errorf("Expecting '}' delimiter but found '%s'", d)
	default:
		return "", errors.Errorf("Expecting string or delimiter but found '%s'", v)
	}
}

func (js *jsonStream) getToken() (interface{}, error) {
	token, err := js.decoder.Token()
	if os.Getenv("FABRIC_CA_JSON_STREAM_DEBUG") != "" {
		log.Debugf("TOKEN: type=%s, %+v\n", reflect.TypeOf(token), token)
	}
	return token, err
}

func errCB(decoder *json.Decoder) error {
	errMsg := &api.ResponseMessage{}
	err := decoder.Decode(errMsg)
	if err != nil {
		return errors.Errorf("Invalid JSON error format: %s", err)
	}
	return errors.Errorf("%+v", errMsg)
}
