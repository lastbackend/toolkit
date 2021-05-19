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
	"gitlab.com/lastbackend/engine/cmd/flags"

	"fmt"
	"os"
	"strings"
)

const (
	defaultShortDescription = "a engine service"
)

type Cmd interface {
	flags.FlagSet

	SetName(string)
	SetVersion(string)
	SetShortDescription(string)
	SetLongDescription(string)
	SetEnvPrefix(string)
	Execute() error
}

type cmd struct {
	flags.Flags

	opts    Options
	rootCmd *cobra.Command
}

type Option func(o *Options)

func New(opts ...Option) Cmd {
	options := Options{}

	for _, o := range opts {
		o(&options)
	}

	if len(options.ShortDesc) == 0 {
		options.ShortDesc = defaultShortDescription
	}

	c := new(cmd)
	c.opts = options

	return c
}

func (c *cmd) SetName(s string) {
	c.opts.Name = s
}

func (c *cmd) SetVersion(s string) {
	c.opts.Version = s
}

func (c *cmd) SetShortDescription(s string) {
	c.opts.ShortDesc = s
}

func (c *cmd) SetLongDescription(s string) {
	c.opts.LongDesc = s
}

func (c *cmd) SetEnvPrefix(s string) {
	flags.EnvPrefix = s
}

func (c *cmd) Execute() error {

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
		printFlags(cmd.Flags())
		return nil
	}

	for _, flag := range c.Flags {
		if err := flag.Apply(c.rootCmd.Flags()); err != nil {
			return err
		}
	}

	return c.rootCmd.Execute()
}

func (c *cmd) versionCommand() *cobra.Command {
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
