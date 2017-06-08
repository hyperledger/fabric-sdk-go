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

package util

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"

	"github.com/cloudflare/cfssl/log"
	"github.com/mitchellh/mapstructure"
	"github.com/op/go-logging"
	"github.com/spf13/cast"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

const (
	// TagDefault is the tag name for a default value of a field as recognized
	// by RegisterFlags.
	TagDefault = "def"
	// TagHelp is the tag name for a help message of a field as recognized
	// by RegisterFlags.
	TagHelp = "help"
	// TagOpt is the tag name for a one character option of a field as recognized
	// by RegisterFlags.  For example, a value of "d" reserves "-d" for the
	// command line argument.
	TagOpt = "opt"
	// TagSkip is the tag name which causes the field to be skipped by
	// RegisterFlags.
	TagSkip = "skip"
)

// RegisterFlags registers flags for all fields in an arbitrary 'config' object.
// This method recognizes the following field tags:
// "def" - the default value of the field;
// "opt" - the optional one character short name to use on the command line;
// "help" - the help message to display on the command line;
// "skip" - to skip the field.
func RegisterFlags(flags *pflag.FlagSet, config interface{}, tags map[string]string) error {
	fr := &flagRegistrar{flags: flags, tags: tags}
	return ParseObj(config, fr.Register)
}

type flagRegistrar struct {
	flags *pflag.FlagSet
	tags  map[string]string
}

func (fr *flagRegistrar) Register(f *Field) (err error) {
	// Don't register non-leaf fields
	if !f.Leaf {
		return nil
	}
	// Don't register fields with no address
	if f.Addr == nil {
		return fmt.Errorf("Field is not addressable: %s", f.Path)
	}
	skip := fr.getTag(f, TagSkip)
	if skip != "" {
		return nil
	}
	help := fr.getTag(f, TagHelp)
	opt := fr.getTag(f, TagOpt)
	def := fr.getTag(f, TagDefault)
	switch f.Kind {
	case reflect.String:
		if help == "" {
			return fmt.Errorf("Field is missing a help tag: %s", f.Path)
		}
		fr.flags.StringVarP(f.Addr.(*string), f.Path, opt, def, help)
	case reflect.Int:
		if help == "" {
			return fmt.Errorf("Field is missing a help tag: %s", f.Path)
		}
		var intDef int
		if def != "" {
			intDef, err = strconv.Atoi(def)
			if err != nil {
				return fmt.Errorf("Invalid integer value in 'def' tag of %s field", f.Path)
			}
		}
		fr.flags.IntVarP(f.Addr.(*int), f.Path, opt, intDef, help)
	case reflect.Bool:
		if help == "" {
			return fmt.Errorf("Field is missing a help tag: %s", f.Path)
		}
		var boolDef bool
		if def != "" {
			boolDef, err = strconv.ParseBool(def)
			if err != nil {
				return fmt.Errorf("Invalid boolean value in 'def' tag of %s field", f.Path)
			}
		}
		fr.flags.BoolVarP(f.Addr.(*bool), f.Path, opt, boolDef, help)
	case reflect.Slice:
		if f.Type.Elem().Kind() == reflect.String {
			if help == "" {
				return fmt.Errorf("Field is missing a help tag: %s", f.Path)
			}
			fr.flags.StringSliceVarP(f.Addr.(*[]string), f.Path, opt, nil, help)
		} else {
			return nil
		}
	default:
		log.Debugf("Not registering flag for '%s' because it is a currently unsupported type: %s\n",
			f.Path, f.Kind)
		return nil
	}
	bindFlag(fr.flags, f.Path)
	return nil
}

func (fr *flagRegistrar) getTag(f *Field, tagName string) string {
	var key, val string
	key = fmt.Sprintf("%s.%s", tagName, f.Path)
	if fr.tags != nil {
		val = fr.tags[key]
	}
	if val == "" {
		val = f.Tag.Get(tagName)
	}
	return val
}

// CmdRunBegin is called at the beginning of each cobra run function
func CmdRunBegin() {
	// If -d or --debug, set debug logging level
	if viper.GetBool("debug") {
		log.Level = log.LevelDebug

		logging.SetLevel(logging.INFO, "bccsp")
		logging.SetLevel(logging.INFO, "bccsp_p11")
		logging.SetLevel(logging.INFO, "bccsp_sw")
	}
}

// FlagString sets up a flag for a string, binding it to its name
func FlagString(flags *pflag.FlagSet, name, short string, def string, desc string) {
	flags.StringP(name, short, def, desc)
	bindFlag(flags, name)
}

// common binding function
func bindFlag(flags *pflag.FlagSet, name string) {
	flag := flags.Lookup(name)
	if flag == nil {
		panic(fmt.Errorf("failed to lookup '%s'", name))
	}
	viper.BindPFlag(name, flag)
}

// ViperUnmarshal is a work around for a bug in viper.Unmarshal
// This can be removed once https://github.com/spf13/viper/issues/327 is fixed
// and vendored.
func ViperUnmarshal(cfg interface{}, stringSliceFields []string, vp *viper.Viper) error {
	decoderConfig := &mapstructure.DecoderConfig{
		Metadata:         nil,
		Result:           cfg,
		WeaklyTypedInput: true,
		DecodeHook:       mapstructure.StringToTimeDurationHookFunc(),
	}
	decoder, err := mapstructure.NewDecoder(decoderConfig)
	if err != nil {
		return fmt.Errorf("Failed to create decoder: %s", err)
	}
	settings := vp.AllSettings()
	for _, field := range stringSliceFields {
		var ok bool
		path := strings.Split(field, ".")
		m := settings
		name := path[0]
		// If it is a top level option check to see if nil before continuing
		if _, ok = m[name]; !ok {
			continue
		}

		if len(path) > 1 {
			for _, field2 := range path[1:] {
				m = m[name].(map[string]interface{})
				name = field2

				// Inspect nested options to see if nil before proceeding with loop
				if _, ok = m[name]; !ok {
					break
				}
			}
		}
		// Only do casting if path was valid
		if ok {
			m[name] = cast.ToStringSlice(m[name])
		}
	}

	return decoder.Decode(settings)
}
