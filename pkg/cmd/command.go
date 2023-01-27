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
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

type Commands []*Command

type Command struct {
	// Use is the one-line usage message.
	// Recommended syntax is as follow:
	//   [ ] identifies an optional argument. Arguments that are not enclosed in brackets are required.
	//   ... indicates that you can specify multiple values for the previous argument.
	//   |   indicates mutually exclusive information. You can use the argument to the left of the separator or the
	//       argument to the right of the separator. You cannot use both arguments in a single use of the command.
	//   { } delimits a set of mutually exclusive arguments when one of the arguments is required. If the arguments are
	//       optional, they are enclosed in brackets ([ ]).
	// Example: add [-F file | -D dir]... [-f format] profile
	Use string

	// Aliases is an array of aliases that can be used instead of the first word in Use.
	Aliases []string

	// ShortDesc is the short description shown in the 'help' output.
	ShortDesc string

	// Desc is the long message shown in the 'help <this-command>' output.
	Desc string

	// Example is examples of how to use the command.s
	Example string

	// Expected arguments
	Args PositionalArgs

	// ArgAliases is List of aliases for ValidArgs.
	// These are not suggested to the user in the shell completion,
	// but accepted if entered manually.
	ArgAliases []string

	// Deprecated defines, if this command is deprecated and should print this string when used.
	Deprecated string

	// Annotations are key/value pairs that can be used by applications to identify or
	// group commands.
	Annotations map[string]string

	// Version defines the version for this command. If this value is non-empty and the command does not
	// define a "version" flag, a "version" boolean flag will be added to the command and, if specified,
	// will print content of the "Version" variable. A shorthand "v" flag will also be added if the
	// command does not define one.
	Version string

	// The *Run functions are executed in the following order:
	//   * PersistentPreRun()
	//   * PreRun()
	//   * Run()
	//   * PostRun()
	//   * PersistentPostRun()
	// All functions get the same args, the arguments after the command name.
	//
	// PersistentPreRun: children of this command will inherit and execute.
	PersistentPreRun func(cmd *Command, args []string) error
	// PreRun: children of this command will not inherit.
	PreRun func(cmd *Command, args []string) error
	// Run: Typically the actual work function. Most commands will only implement this.
	Run func(cmd *Command, args []string) error
	// PostRun: run after the Run command.
	PostRun func(cmd *Command, args []string) error
	// PersistentPostRun: children of this command will inherit and execute after PostRun.
	PersistentPostRun func(cmd *Command, args []string) error

	// Hidden defines, if this command is hidden and should NOT show up in the list of available commands.
	Hidden bool

	// Flags for applies
	flags Flags

	// Relation on cobra.Command
	cobraCommand cobra.Command
}

type PositionalArgs func(cmd *Command, args []string) error

func (c *Command) SetGlobalNormalizationFunc(fn func(f *pflag.FlagSet, name string) pflag.NormalizedName) {
	c.cobraCommand.SetGlobalNormalizationFunc(fn)
}

func (c *Command) InitDefaultHelpFlag() {
	c.cobraCommand.InitDefaultHelpFlag()
}

func (c *Command) SetHelpCommand(cmd *Command) {
	c.cobraCommand.SetHelpCommand(cmd.convertCommandToCobraCommand())
}

func (c *Command) AddCommand(cmd *Command) {
	c.cobraCommand.AddCommand(cmd.convertCommandToCobraCommand())
}

func (c *Command) PersistentFlags() *pflag.FlagSet {
	return c.cobraCommand.PersistentFlags()
}

func (c *Command) Flags() *pflag.FlagSet {
	return c.cobraCommand.Flags()
}

func (c *Command) AddStringFlag(name string, dest *string) *StringFlag {
	c.flagsInit()
	return c.flags.AddStringFlag(name, dest)
}

func (c *Command) AddIntFlag(name string, dest *int) *IntFlag {
	c.flagsInit()
	return c.flags.AddIntFlag(name, dest)
}

func (c *Command) AddInt32Flag(name string, dest *int32) *Int32Flag {
	c.flagsInit()
	return c.flags.AddInt32Flag(name, dest)
}

func (c *Command) AddBoolFlag(name string, dest *bool) *BoolFlag {
	c.flagsInit()
	return c.flags.AddBoolFlag(name, dest)
}

func (c *Command) AddStringSliceFlag(name string, dest *[]string) *StringSliceFlag {
	c.flagsInit()
	return c.flags.AddStringSliceFlag(name, dest)
}

func (c *Command) flagsInit() {
	if c.flags == nil {
		c.flags = make(Flags, 0)
	}
}

func (c *Command) convertCommandToCobraCommand() *cobra.Command {
	cmd := new(cobra.Command)
	cmd.Use = c.Use
	cmd.Aliases = c.Aliases
	cmd.Short = c.ShortDesc
	cmd.Long = c.Desc
	cmd.Example = c.Example
	cmd.ArgAliases = c.ArgAliases
	cmd.Deprecated = c.Deprecated
	cmd.Annotations = c.Annotations
	cmd.Version = c.Version
	cmd.Hidden = c.Hidden

	if c.Args != nil {
		cmd.Args = func(cbCmd *cobra.Command, args []string) error {
			return c.Args(c, args)
		}
	}

	if c.PersistentPreRun != nil {
		cmd.PersistentPreRunE = func(cbCmd *cobra.Command, args []string) error {
			return c.PersistentPreRun(c, args)
		}
	}
	if c.PreRun != nil {
		cmd.PreRunE = func(cbCmd *cobra.Command, args []string) error {
			return c.PreRun(c, args)
		}
	}
	if c.Run != nil {
		cmd.RunE = func(cbCmd *cobra.Command, args []string) error {
			return c.Run(c, args)
		}
	}
	if c.PostRun != nil {
		cmd.PostRunE = func(cbCmd *cobra.Command, args []string) error {
			return c.PostRun(c, args)
		}
	}
	if c.PersistentPostRun != nil {
		cmd.PersistentPostRunE = func(cbCmd *cobra.Command, args []string) error {
			return c.PersistentPostRun(c, args)
		}
	}

	for _, f := range c.flags {
		if err := (f.(flag)).apply(cmd.Flags()); err != nil {
			panic(err)
		}
	}

	c.cobraCommand = *cmd

	return cmd
}
