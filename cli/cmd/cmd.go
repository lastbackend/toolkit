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
	"os"
	"runtime/debug"

	"github.com/urfave/cli/v2"
)

type Option func(o *Options)

type Options struct {
	Name        string
	Description string
	Version     string
}

type CLI interface {
	App() *cli.App
	Options() Options
	Run() error
}

type cmd struct {
	app  *cli.App
	opts Options
}

func (c *cmd) App() *cli.App {
	return c.app
}

func (c *cmd) Options() Options {
	return c.opts
}

func (c *cmd) Run() error {
	return c.app.Run(os.Args)
}

func NewCLI(opts ...Option) CLI {
	options := Options{}

	for _, o := range opts {
		o(&options)
	}

	if len(options.Name) == 0 {
		options.Name = name
	}
	if len(options.Description) == 0 {
		options.Description = description
	}
	if len(options.Version) == 0 {
		if bi, ok := debug.ReadBuildInfo(); ok {
			options.Version = bi.Main.Version
		} else {
			options.Version = version
		}
	}

	c := new(cmd)
	c.app = cli.NewApp()
	c.opts = options
	c.app.Name = c.opts.Name
	c.app.Usage = c.opts.Description
	c.app.Version = c.opts.Version
	c.app.EnableBashCompletion = true
	c.app.HideVersion = len(options.Version) == 0

	return c
}
