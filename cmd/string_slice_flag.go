/*
Copyright [2014] - [2021] The Last.Backend authors.

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

type StringSliceFlag struct {
	flag

	name        string    // name as it appears on command line
	shorthand   string    // one-letter abbreviated flag
	usage       string    // help message
	envVar      string    // environment values as set
	value       []string  // value as set
	destination *[]string // the return value is the address of a string variable that stores the value of the flag.
	required    bool      // mark flag as required
	hasBeenSet  bool
}

func (f *StringSliceFlag) Short(short string) *StringSliceFlag {
	f.shorthand = short
	return f
}

func (f *StringSliceFlag) Env(env string) *StringSliceFlag {
	f.envVar = env
	return f
}

func (f *StringSliceFlag) Usage(usage string) *StringSliceFlag {
	f.usage = usage
	return f
}

func (f *StringSliceFlag) Default(value []string) *StringSliceFlag {
	f.value = value
	return f
}

func (f *StringSliceFlag) Required() *StringSliceFlag {
	f.required = true
	return f
}

func (f *StringSliceFlag) apply(set *pflag.FlagSet) error {
	if val, ok := getEnvAsSlice(f.envVar, ";"); ok {
		f.value = val
		f.hasBeenSet = ok
	}

	if f.destination == nil {
		set.StringSliceP(f.name, f.shorthand, f.value, f.usage)
	} else {
		set.StringSliceVarP(f.destination, f.name, f.shorthand, f.value, f.usage)
	}

	if f.required && !f.hasBeenSet {
		if err := cobra.MarkFlagRequired(set, f.name); err != nil {
			return err
		}
	}

	return nil
}

func (f StringSliceFlag) IsSet() bool {
	return f.hasBeenSet
}
