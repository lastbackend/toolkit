/*
Copyright [2014] - [2023] The Last.Backend authors.

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

package cmd

import (
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

type Int32Flag struct {
	flag

	name        string // name as it appears on command line
	shorthand   string // one-letter abbreviated flag
	usage       string // help message
	envVar      string // environment values as set
	value       int32  // value as set
	destination *int32 // the return value is the address of a int variable that stores the value of the flag.
	required    bool   // mark flag as required

	hasBeenSet bool
}

func (f *Int32Flag) Short(short string) *Int32Flag {
	f.shorthand = short
	return f
}

func (f *Int32Flag) Env(env string) *Int32Flag {
	f.envVar = env
	return f
}

func (f *Int32Flag) Usage(usage string) *Int32Flag {
	f.usage = usage
	return f
}

func (f *Int32Flag) Default(value int32) *Int32Flag {
	f.value = value
	return f
}
func (f *Int32Flag) Required() *Int32Flag {
	f.required = true
	return f
}

func (f *Int32Flag) apply(set *pflag.FlagSet) error {
	if val, ok := getEnvAsInt32(f.envVar); ok {
		f.value = val
		f.hasBeenSet = ok
	}

	if f.destination == nil {
		set.Int32P(f.name, f.shorthand, f.value, f.usage)
	} else {
		set.Int32VarP(f.destination, f.name, f.shorthand, f.value, f.usage)
	}

	if f.required && !f.hasBeenSet {
		if err := cobra.MarkFlagRequired(set, f.name); err != nil {
			return err
		}
	}

	return nil
}

func (f Int32Flag) IsSet() bool {
	return f.hasBeenSet
}
