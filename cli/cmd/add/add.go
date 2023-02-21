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

package add

import (
	tcli "github.com/lastbackend/toolkit/cli/cmd"
	"github.com/urfave/cli/v2"
)

func init() {
	tcli.Register(&cli.Command{
		Name:  "add",
		Usage: "Add new entity to service template",
		Subcommands: []*cli.Command{
			{
				Name: "method",
				Usage: `Add a new method to service template
					Ex:
						` + tcli.App().Name + ` add method --expose-http=get:/demo --proxy=helloworld:/helloworld.Greeter/SayHello Example
						` + tcli.App().Name + ` add method --expose-ws=/demo Example
						` + tcli.App().Name + ` add method --subscribe-ws --proxy=prime:/helloworld.Greeter/SayHello Example
						` + tcli.App().Name + ` add method Example 			
				`,
				Action: AddMethod,
				Flags:  methodFlags,
			},
			{
				Name: "plugin",
				Usage: `Add a new plugin to service template
					Ex:
						` + tcli.App().Name + ` add plugin --prefix=redis redis
				`,
				Action: AddPlugin,
				Flags:  pluginFlags,
			},
		},
	})
}
