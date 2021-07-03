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

type IntFlag struct {
	flag

	Name        string // name as it appears on command line
	Shorthand   string // one-letter abbreviated flag
	Usage       string // help message
	EnvVar      string // environment values as set
	Value       int    // value as set
	Destination *int   // the return value is the address of a int variable that stores the value of the flag.
	Required    bool   // mark flag as required

	hasBeenSet bool
}

func (f *IntFlag) apply(set *pflag.FlagSet) error {
	val, ok := getEnvAsInt(f.EnvVar)
	f.Value = val
	f.hasBeenSet = ok

	if f.Destination == nil {
		set.IntP(f.Name, f.Shorthand, f.Value, f.Usage)
	} else {
		set.IntVarP(f.Destination, f.Name, f.Shorthand, f.Value, f.Usage)
	}

	if f.Required && !f.hasBeenSet {
		if err := cobra.MarkFlagRequired(set, f.Name); err != nil {
			return err
		}
	}

	return nil
}

func (f IntFlag) IsSet() bool {
	return f.hasBeenSet
}
