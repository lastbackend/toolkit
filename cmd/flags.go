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
	"fmt"
	"strings"
	"syscall"
)

type Flags []Flag

func (f *Flags) AddStringFlag(name string, shorthand string, value string, dest *string, envVars []string, required bool, usage string) {
	*f = append(*f, &StringFlag{
		Name:        name,
		Shorthand:   shorthand,
		Value:       value,
		Usage:       usage,
		EnvVars:     envVars,
		Required:    required,
		Destination: dest,
	})
}

func (f *Flags) AddIntFlag(name string, shorthand string, value int, dest *int, envVars []string, required bool, usage string) {
	*f = append(*f, &IntFlag{
		Name:        name,
		Shorthand:   shorthand,
		Value:       value,
		Usage:       usage,
		EnvVars:     envVars,
		Destination: dest,
	})
}

func (f *Flags) AddBoolFlag(name string, shorthand string, value bool, dest *bool, envVars []string, required bool, usage string) {
	*f = append(*f, &BoolFlag{
		Name:        name,
		Shorthand:   shorthand,
		Value:       value,
		Usage:       usage,
		EnvVars:     envVars,
		Destination: dest,
	})
}

func flagFromEnv(envVars []string) (val string, ok bool) {
	for _, envVar := range envVars {
		envVar = strings.TrimSpace(envVar)
		envVar = envName(envVar)
		if val, ok := syscall.Getenv(envVar); ok {
			return val, true
		}
	}
	return "", false
}

func envName(name string) string {
	return fmt.Sprintf("%s_%s", EnvPrefix, strings.Replace(strings.ToUpper(name), "-", "_", -1))
}
