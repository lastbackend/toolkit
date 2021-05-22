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
	"fmt"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"os"
	"strings"
)

const (
	defaultShortDescription = "a engine service"
)

type CLI interface {
	FlagSet

	SetName(string)
	SetVersion(string)
	SetShortDescription(string)
	SetLongDescription(string)
	SetEnvPrefix(string)
	AddFlags(...Flag)
	AddCommands(...Command)
	Execute() error
}

type cli struct {
	Flags
	Commands

	opts    Options
	rootCmd *cobra.Command
}

type Option func(o *Options)

func New(opts ...Option) CLI {
	options := Options{}

	for _, o := range opts {
		o(&options)
	}

	if len(options.ShortDesc) == 0 {
		options.ShortDesc = defaultShortDescription
	}

	c := new(cli)
	c.opts = options

	return c
}

func (c *cli) SetName(s string) {
	c.opts.Name = s
}

func (c *cli) SetVersion(s string) {
	c.opts.Version = s
}

func (c *cli) SetShortDescription(s string) {
	c.opts.ShortDesc = s
}

func (c *cli) SetLongDescription(s string) {
	c.opts.LongDesc = s
}

func (c *cli) SetEnvPrefix(s string) {
	EnvPrefix = s
}

func (c *cli) AddFlags(flags ...Flag) {
	if len(flags) == 0 {
		return
	}
	c.Flags = append(c.Flags, flags...)
}

func (c *cli) AddCommands(commands ...Command) {
	if len(commands) == 0 {
		return
	}
	c.Commands = append(c.Commands, commands...)
}

func (c *cli) Execute() error {

	c.rootCmd = &cobra.Command{}
	c.rootCmd.SetGlobalNormalizationFunc(wordSepNormalizeFunc)
	c.rootCmd.InitDefaultHelpFlag()
	c.rootCmd.SetHelpCommand(&cobra.Command{Use: "no-help", Hidden: true})
	c.rootCmd.PersistentFlags().BoolP("debug", "d", false, "Enable debug mode")
	c.rootCmd.AddCommand(c.versionCommand())

	if len(c.opts.Name) > 0 {
		c.rootCmd.Use = c.opts.Name
	}
	if len(c.opts.ShortDesc) > 0 {
		c.rootCmd.Short = c.opts.ShortDesc
	}
	if len(c.opts.ShortDesc) > 0 {
		c.rootCmd.Long = c.opts.LongDesc
	}
	if len(c.opts.Version) > 0 {
		c.rootCmd.Version = c.opts.Version
	}

	c.rootCmd.RunE = func(cmd *cobra.Command, args []string) error {

		debug, err := cmd.Flags().GetBool("debug")
		if err != nil {
			return err
		}

		if debug {
			printFlags(cmd.Flags())
		}

		return nil
	}

	for _, f := range c.Flags {
		if err := (f.(flag)).apply(c.rootCmd.Flags()); err != nil {
			return err
		}
	}

	for _, cmd := range c.Commands {
		if err := (cmd.(command)).apply(c.rootCmd); err != nil {
			return err
		}
	}

	return c.rootCmd.Execute()
}

func (c *cli) versionCommand() *cobra.Command {
	return &cobra.Command{
		Use:     "version",
		Aliases: []string{"v"},
		Short:   fmt.Sprintf("Print the version number of %s", c.opts.Name),
		Long:    fmt.Sprintf(`All software has versions. This is %s's`, c.opts.Name),
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Println(fmt.Sprintf("version: %s", c.opts.Version))
			os.Exit(0)
			return nil
		},
	}
}

// wordSepNormalizeFunc normalizes cli flags
func wordSepNormalizeFunc(f *pflag.FlagSet, name string) pflag.NormalizedName {
	if strings.Contains(name, "_") {
		return pflag.NormalizedName(strings.Replace(name, "_", "-", -1))
	}
	return pflag.NormalizedName(name)
}

// printFlags logs the flags in the flagset
func printFlags(flags *pflag.FlagSet) {
	flags.VisitAll(func(flag *pflag.Flag) {
		fmt.Println(fmt.Sprintf("FLAG: --%s=%q", flag.Name, flag.Value))
	})
}
