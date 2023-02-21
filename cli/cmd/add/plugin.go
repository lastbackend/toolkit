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
	"fmt"
	"os"
	"path"
	"strings"

	"github.com/lastbackend/toolkit/cli/pkg/scaffold/injector"
	"github.com/urfave/cli/v2"
)

var pluginFlags = []cli.Flag{
	&cli.StringFlag{
		Name:  "path",
		Usage: "Path to proto directory",
		Value: "./proto",
	},
	&cli.StringFlag{
		Name:  "prefix",
		Usage: "Set prefix for plugin",
	},
}

func AddPlugin(ctx *cli.Context) error {
	arg := ctx.Args().First()
	if len(arg) == 0 {
		return cli.ShowSubcommandHelp(ctx)
	}

	workdir := ctx.String("path")
	if path.IsAbs(workdir) {
		fmt.Println("must provide a relative path as service name")
		return nil
	}

	if _, err := os.Stat(workdir); os.IsNotExist(err) {
		return fmt.Errorf("%s not exists", workdir)
	}

	po := injector.PluginOption{
		Prefix: ctx.String("prefix"),
		Name:   strings.ToLower(arg),
	}

	i := injector.New(
		injector.Workdir(workdir),
		injector.Plugin(po),
	)

	if err := i.Inject(); err != nil {
		return err
	}

	return nil
}
