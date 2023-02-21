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

package init

import (
	"bytes"
	"fmt"
	"os"
	"path"
	"text/template"

	tcli "github.com/lastbackend/toolkit/cli/cmd"
	"github.com/lastbackend/toolkit/cli/pkg/scaffold/generator"
	"github.com/lastbackend/toolkit/cli/pkg/scaffold/templates"
	"github.com/lastbackend/toolkit/cli/pkg/util/parser"
	"github.com/pkg/errors"
	"github.com/urfave/cli/v2"
)

var flags = []cli.Flag{
	&cli.BoolFlag{
		Name:  "redis",
		Usage: "Add Redis plugin",
	},
	&cli.BoolFlag{
		Name:  "rabbitmq",
		Usage: "Add RabbitMQ plugin",
	},
	&cli.BoolFlag{
		Name:  "postgres-pg",
		Usage: "Add Postgres PG plugin",
	},
	&cli.BoolFlag{
		Name:  "postgres-gorm",
		Usage: "Add Postgres GORM plugin",
	},
	&cli.BoolFlag{
		Name:  "postgres",
		Usage: "Add Postgres plugin",
	},
	&cli.BoolFlag{
		Name:  "centrifuge",
		Usage: "Add Centrifuge plugin",
	},
}

func init() {
	tcli.Register(&cli.Command{
		Name:  "init",
		Usage: "Create a new service template",
		Subcommands: []*cli.Command{
			{
				Name:   "service",
				Usage:  "Create a new service template, e.g. " + tcli.App().Name + " new service github.com/lastbackend/example",
				Action: CreateService,
				Flags:  flags,
			},
		},
	})
}

func CreateService(ctx *cli.Context) error {
	arg := ctx.Args().First()
	if len(arg) == 0 {
		return cli.ShowSubcommandHelp(ctx)
	}

	name, vendor := parser.ParseNameAndVendor(arg)

	workdir := name
	if path.IsAbs(workdir) {
		fmt.Println("must provide a relative path as service name")
		return nil
	}

	if _, err := os.Stat(workdir); !os.IsNotExist(err) {
		return fmt.Errorf("%s already exists", workdir)
	}

	g := generator.New(
		generator.Service(name),
		generator.Vendor(vendor),
		generator.Workdir(workdir),
		generator.RedisPlugin(ctx.Bool("redis")),
		generator.RabbitMQPlugin(ctx.Bool("rabbitmq")),
		generator.PostgresPGPlugin(ctx.Bool("postgres-pg")),
		generator.PostgresGORMPlugin(ctx.Bool("postgres-gorm")),
		generator.PostgresPlugin(ctx.Bool("postgres")),
		generator.CentrifugePlugin(ctx.Bool("centrifuge")),
	)

	files := []generator.File{
		{Path: "scripts/bootstrap.sh", Template: templates.BootstrapScript, FileMode: 0755},
		{Path: "scripts/generate.sh", Template: templates.GenerateScript, FileMode: 0755},
		{Path: "proto/" + name + ".proto", Template: templates.ProtoService},
		{Path: "config/config.go", Template: templates.ServiceConfig},
		{Path: "internal/server/server.go", Template: templates.ServiceServer},
		{Path: ".dockerignore", Template: templates.DockerIgnore},
		{Path: ".gitignore", Template: templates.GitIgnore},
		{Path: "Dockerfile", Template: templates.Dockerfile},
		{Path: "Makefile", Template: templates.Makefile},
		{Path: "main.go", Template: templates.Main},
		{Path: "go.mod", Template: templates.Module},
	}

	if err := g.Generate(files); err != nil {
		return err
	}

	usage, err := getUsage(name, workdir)
	if err != nil {
		return err
	}

	fmt.Println(usage)

	return nil
}

func getUsage(name, path string) (string, error) {
	tmp := `Next steps:
Install requirements and compile the proto file {{ .Name }}.proto and install dependencies:
 - cd {{ .Path }}
 - make init proto update tidy`

	t, err := template.New("comments").Parse(tmp)
	if err != nil {
		return "", errors.Wrap(err, "Failed to parse comments template")
	}

	var b bytes.Buffer
	if err := t.Execute(&b, map[string]interface{}{
		"Name": name,
		"Path": path,
	}); err != nil {
		return "", errors.Wrap(err, "Failed to execute proto comments template")
	}

	return b.String(), nil
}
