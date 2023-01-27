/*
Copyright [2014] - [2022] The Last.Backend authors.

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

	"fmt"
	"os"
	"strings"
)

func AddGlobalFlags(fs *pflag.FlagSet) {

	// lookup flags in global flag set and re-register the values with our flagset
	global := pflag.CommandLine
	local := pflag.NewFlagSet(os.Args[0], pflag.ExitOnError)

	pflagRegister(global, local, "debug", "d")

	fs.AddFlagSet(local)
}

func pflagRegister(global, local *pflag.FlagSet, globalName string, shaortName string) {
	if f := global.Lookup(globalName); f != nil {
		f.Name = normalize(f.Name)
		f.Shorthand = shaortName
		local.AddFlag(f)
	} else {
		panic(fmt.Sprintf("failed to find flag in global flagset (pflag): %s", globalName))
	}
}

func normalize(s string) string {
	return strings.Replace(s, "_", "-", -1)
}
