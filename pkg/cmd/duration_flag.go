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

	"time"
)

type DurationFlag struct {
	flag

	name        string         // name as it appears on command line
	shorthand   string         // one-letter abbreviated flag
	usage       string         // help message
	value       time.Duration  // value as set
	destination *time.Duration // the return value is the address of a int variable that stores the value of the flag.
	envVar      string         // environment values as set
	required    bool           // mark flag as required

	hasBeenSet bool
}

func (f *DurationFlag) Short(short string) *DurationFlag {
	f.shorthand = short
	return f
}

func (f *DurationFlag) Env(env string) *DurationFlag {
	f.envVar = env
	return f
}

func (f *DurationFlag) Usage(usage string) *DurationFlag {
	f.usage = usage
	return f
}

func (f *DurationFlag) Default(value time.Duration) *DurationFlag {
	f.value = value
	return f
}

func (f *DurationFlag) Required() *DurationFlag {
	f.required = true
	return f
}

func (f *DurationFlag) apply(set *pflag.FlagSet) error {
	if val, ok := getEnvAsDuration(f.envVar); ok {
		f.value = val
		f.hasBeenSet = ok
	}

	if f.destination == nil {
		set.DurationP(f.name, f.shorthand, f.value, f.usage)
	} else {
		set.DurationVarP(f.destination, f.name, f.shorthand, f.value, f.usage)
	}

	if f.required && !f.hasBeenSet {
		if err := cobra.MarkFlagRequired(set, f.name); err != nil {
			return err
		}
	}

	return nil
}

func (f DurationFlag) IsSet() bool {
	return f.hasBeenSet
}
