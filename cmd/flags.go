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
	"strconv"
	"strings"
	"syscall"
	"time"
)

type Flags []Flag

func (f *Flags) AddStringFlag(name string, shorthand string, value string, dest *string, envVar string, required bool, usage string) {
	*f = append(*f, &StringFlag{
		Name:        name,
		Shorthand:   shorthand,
		Value:       value,
		Usage:       usage,
		EnvVar:      envVar,
		Required:    required,
		Destination: dest,
	})
}

func (f *Flags) AddIntFlag(name string, shorthand string, value int, dest *int, envVar string, required bool, usage string) {
	*f = append(*f, &IntFlag{
		Name:        name,
		Shorthand:   shorthand,
		Value:       value,
		Usage:       usage,
		EnvVar:      envVar,
		Destination: dest,
	})
}

func (f *Flags) AddInt32Flag(name string, shorthand string, value int32, dest *int32, envVar string, required bool, usage string) {
	*f = append(*f, &Int32Flag{
		Name:        name,
		Shorthand:   shorthand,
		Value:       value,
		Usage:       usage,
		EnvVar:      envVar,
		Destination: dest,
	})
}

func (f *Flags) AddBoolFlag(name string, shorthand string, value bool, dest *bool, envVar string, required bool, usage string) {
	*f = append(*f, &BoolFlag{
		Name:        name,
		Shorthand:   shorthand,
		Value:       value,
		Usage:       usage,
		EnvVar:      envVar,
		Destination: dest,
	})
}

func (f *Flags) AddStringSliceFlag(name string, shorthand string, value []string, dest *[]string, envVar string, required bool, usage string) {
	*f = append(*f, &StringSliceFlag{
		Name:        name,
		Shorthand:   shorthand,
		Value:       value,
		Usage:       usage,
		EnvVar:      envVar,
		Destination: dest,
	})
}

func (f *Flags) AddDurationFlag(name string, shorthand string, value time.Duration, dest *time.Duration, envVar string, required bool, usage string) {
	*f = append(*f, &DurationFlag{
		Name:        name,
		Shorthand:   shorthand,
		Value:       value,
		Usage:       usage,
		EnvVar:      envVar,
		Destination: dest,
	})
}

func getEnv(envVar string) (string, bool) {
	envVar = strings.TrimSpace(envVar)
	envVar = envName(envVar)
	return syscall.Getenv(envVar)
}

func getEnvAsInt(name string) (int, bool) {
	valueStr, ok := getEnv(name)
	if value, err := strconv.Atoi(valueStr); err == nil {
		return value, ok
	}
	return 0, false
}

func getEnvAsInt32(name string) (int32, bool) {
	valueStr, ok := getEnv(name)
	if value, err := strconv.ParseInt(valueStr, 10, 32); err == nil {
		return int32(value), ok
	}
	return 0, false
}

func getEnvAsBool(name string) (bool, bool) {
	valStr, ok := getEnv(name)
	if val, err := strconv.ParseBool(valStr); err == nil {
		return val, ok
	}
	return false, false
}

func getEnvAsDuration(name string) (time.Duration, bool) {
	valStr, ok := getEnv(name)
	if val, err := strconv.ParseInt(valStr, 0, 64); err == nil {
		return time.Duration(val) * time.Millisecond, ok
	}
	return time.Duration(0) * time.Millisecond, false
}

func getEnvAsSlice(name string, sep string) ([]string, bool) {
	valStr, ok := getEnv(name)
	if valStr == "" {
		return make([]string, 0), false
	}
	val := strings.Split(valStr, sep)
	return val, ok
}

func envName(name string) string {
	return fmt.Sprintf("%s_%s", EnvPrefix, strings.Replace(strings.ToUpper(name), "-", "_", -1))
}
