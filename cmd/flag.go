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
	"github.com/spf13/pflag"
	"time"
)

var EnvPrefix = "ENGINE"

type Flag interface {
	IsSet() bool
}

type flag interface {
	apply(*pflag.FlagSet) error
}

type FlagSet interface {
	AddStringFlag(name string, shorthand string, value string, dest *string, envVars string, required bool, usage string)
	AddIntFlag(name string, shorthand string, value int, dest *int, envVars string, required bool, usage string)
	AddBoolFlag(name string, shorthand string, value bool, dest *bool, envVars string, required bool, usage string)
	AddStringSliceFlag(name string, shorthand string, value []string, dest *[]string, envVars string, required bool, usage string)
	AddDurationFlag(name string, shorthand string, value time.Duration, dest *time.Duration, envVars string, required bool, usage string)
}
