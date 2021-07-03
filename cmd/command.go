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
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

type Cmd struct {
	command
	FlagSet

	Use       string
	ShortDesc string
	LongDesc  string
	Example   string
	Run       func(cmd Command, args []string) error

	flags Flags
	cmd   *cobra.Command
}

func (c *Cmd) Flags() *pflag.FlagSet {
	return c.cmd.Flags()
}

func (c *Cmd) AddStringFlag(name string, shorthand string, value string, dest *string, envVar string, required bool, usage string) {
	c.flagsInit()
	c.flags.AddStringFlag(name, shorthand, value, dest, envVar, required, usage)
}

func (c *Cmd) AddIntFlag(name string, shorthand string, value int, dest *int, envVar string, required bool, usage string) {
	c.flagsInit()
	c.flags.AddIntFlag(name, shorthand, value, dest, envVar, required, usage)
}

func (c *Cmd) AddBoolFlag(name string, shorthand string, value bool, dest *bool, envVar string, required bool, usage string) {
	c.flagsInit()
	c.flags.AddBoolFlag(name, shorthand, value, dest, envVar, required, usage)
}

func (c *Cmd) AddStringSliceFlag(name string, shorthand string, value []string, dest *[]string, envVar string, required bool, usage string) {
	c.flagsInit()
	c.flags.AddStringSliceFlag(name, shorthand, value, dest, envVar, required, usage)
}

func (c *Cmd) apply(set CommandSet) error {
	c.cmd = &cobra.Command{
		Use:     c.Use,
		Short:   c.ShortDesc,
		Long:    c.LongDesc,
		Example: c.Example,
	}

	for _, f := range c.flags {
		if err := (f.(flag)).apply(c.cmd.Flags()); err != nil {
			return err
		}
	}

	if c.Run != nil {
		c.cmd.RunE = func(cmd *cobra.Command, args []string) error {
			return c.Run(cmd, args)
		}
	}

	set.AddCommand(c.cmd)

	return nil
}

func (c *Cmd) flagsInit() {
	if c.flags == nil {
		c.flags = make(Flags, 0)
	}
}
