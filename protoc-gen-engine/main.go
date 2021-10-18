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

package main

import (
	"flag"
	"fmt"
	"github.com/lastbackend/engine/protoc-gen-engine/generator"

	"os"
)

func main() {
	showVersion := flag.Bool("version", false, "print the version and exit")

	flag.Parse()
	generator.ParseFlag()

	if *showVersion {
		fmt.Printf("protoc-gen-engine %v\n", generator.DefaultVersion)
		os.Exit(0)
	}

	g := generator.Init(
		generator.DebugEnv("ENGINE_DEBUG"),
	)
	if err := g.Run(); err != nil {
		os.Exit(1)
	}
}
