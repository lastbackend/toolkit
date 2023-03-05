package controller

import (
	"os"
	"text/template"
)

const templateHelpText string = `
NAME:
	{{ .Name }} - {{ .Desc }}

USAGE:
  {{ .Name }}  [global options] command [command options] [arguments...]

COMMANDS:
   help, h  Shows a list of commands or help for one command

GLOBAL OPTIONS:
   --help, -h  					show help
   --all       					show all available variables. by default show only required
   --yaml      					show environment variables as a yaml, table by default
   --without-comments		disable printing envs comments

ENVIRONMENT VARIABLES

{{ .Envs }}

`

func (c *controller) help() bool {

	var (
		help, all, yaml, nocomments bool
	)

	for _, arg := range os.Args {
		switch arg {
		case "-h":
			help = true
		case "-help":
			help = true
		case "h":
			help = true
		case "help":
			help = true
		case "--all":
			all = true
		case "--yaml":
			yaml = true
		case "--without-comments":
			nocomments = true
		}

	}

	if !help {
		return false
	}

	type data struct {
		Name string
		Desc string
		Envs string
	}

	tpl, err := template.New("help").Parse(templateHelpText)
	if err != nil {
		panic(err)
	}

	var envs = c.Config().PrintTable(all, nocomments)

	if yaml {
		envs = c.Config().PrintYaml(all, nocomments)
	}

	if err := tpl.Execute(os.Stdout, data{
		Name: c.meta.GetName(),
		Desc: c.meta.GetDescription(),
		Envs: envs,
	}); err != nil {
		panic(err)
	}

	return true
}
