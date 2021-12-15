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
	cmd.configure()
	c.cobraCommand.SetHelpCommand(&cmd.cobraCommand)
}

func (c *Command) AddCommand(cmd *Command) {
	cmd.configure()
	c.cobraCommand.AddCommand(&cmd.cobraCommand)
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

func (c *Command) configure() {
	c.cobraCommand.Use = c.Use
	c.cobraCommand.Aliases = c.Aliases
	c.cobraCommand.Short = c.ShortDesc
	c.cobraCommand.Long = c.Desc
	c.cobraCommand.Example = c.Example
	c.cobraCommand.ArgAliases = c.ArgAliases
	c.cobraCommand.Deprecated = c.Deprecated
	c.cobraCommand.Annotations = c.Annotations
	c.cobraCommand.Version = c.Version
	c.cobraCommand.Hidden = c.Hidden

	if c.Args != nil {
		c.cobraCommand.Args = func(cbCmd *cobra.Command, args []string) error {
			return c.Args(convertCobraCommandToCommand(cbCmd), args)
		}
	}

	if c.PersistentPreRun != nil {
		c.cobraCommand.PersistentPreRunE = func(cbCmd *cobra.Command, args []string) error {
			return c.PersistentPreRun(convertCobraCommandToCommand(cbCmd), args)
		}
	}
	if c.PreRun != nil {
		c.cobraCommand.PreRunE = func(cbCmd *cobra.Command, args []string) error {
			return c.PreRun(convertCobraCommandToCommand(cbCmd), args)
		}
	}
	if c.Run != nil {
		c.cobraCommand.RunE = func(cbCmd *cobra.Command, args []string) error {
			return c.Run(convertCobraCommandToCommand(cbCmd), args)
		}
	}
	if c.PostRun != nil {
		c.cobraCommand.PostRunE = func(cbCmd *cobra.Command, args []string) error {
			return c.PostRun(convertCobraCommandToCommand(cbCmd), args)
		}
	}
	if c.PersistentPostRun != nil {
		c.cobraCommand.PersistentPostRunE = func(cbCmd *cobra.Command, args []string) error {
			return c.PersistentPostRun(convertCobraCommandToCommand(cbCmd), args)
		}
	}
}

func convertCobraCommandToCommand(cmd *cobra.Command) *Command {
	c := new(Command)
	c.Use = cmd.Use
	c.Aliases = cmd.Aliases
	c.ShortDesc = cmd.Short
	c.Desc = cmd.Long
	c.Example = cmd.Example
	c.ArgAliases = cmd.ArgAliases
	c.Deprecated = cmd.Deprecated
	c.Annotations = cmd.Annotations
	c.Version = cmd.Version
	c.Hidden = cmd.Hidden
	return c
}
