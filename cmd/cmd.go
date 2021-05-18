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
	"github.com/pkg/errors"
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
	Options() Options
	Execute() error
	Get() *cmd
}

type cmd struct {
	opts    Options
	Flags   *flags.Flags
	rootCmd *cobra.Command
}

type Option func(o *Options)

//var (
//	DefaultFlags = []flags.Flag{
//		&flags.StringFlag{
//			Name:    "server_address",
//			EnvVars: []string{"ENGINE_SERVER_ADDRESS"},
//			Usage:   "Bind address for the server. 127.0.0.1:8080",
//		},
//	}
//)

func NewCmd(opts ...Option) Cmd {
	options := Options{}

	for _, o := range opts {
		o(&options)
	}

	if len(options.ShortDesc) == 0 {
		options.ShortDesc = defaultShortDescription
	}

	c := new(cmd)
	c.opts = options
	c.Flags = flags.New()
	c.rootCmd = &cobra.Command{}

	c.rootCmd.SetGlobalNormalizationFunc(wordSepNormalizeFunc)

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

		if err := cmd.ParseFlags(args); err != nil {
			return errors.Wrapf(err, "Failed to parse flags")
		}

		// short-circuit on help
		help, err := cmd.Flags().GetBool("help")
		if err != nil {
			return errors.Wrapf(err, "\"help\" flag is non-bool, programmer error, please correct")
		}
		if help {
			return cmd.Help()
		}

		printFlags(cmd.Flags())

		return nil
	}

	c.rootCmd.AddCommand(c.versionCommand())

	global := pflag.CommandLine
	global.BoolP("help", "h", false, fmt.Sprintf("Show help for command"))
	global.BoolP("debug", "d", false, "Enable debug mode")
	global.StringP("config", "c", "", "Set config filepath")

	AddGlobalFlags(global)

	return c
}

func (c *cmd) Options() Options {
	return c.opts
}

func (c *cmd) Get() *cmd {
	return c
}

func (c *cmd) Execute() error {

	for _, flag := range *c.Flags {
		if err := flag.Apply(c.rootCmd.Flags()); err != nil {
			return err
		}
	}

	return c.rootCmd.Execute()
}

// wordSepNormalizeFunc normalizes cli flags
func wordSepNormalizeFunc(f *pflag.FlagSet, name string) pflag.NormalizedName {
	if strings.Contains(name, "_") {
		return pflag.NormalizedName(strings.Replace(name, "_", "-", -1))
	}
	return pflag.NormalizedName(name)
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

// printFlags logs the flags in the flagset
func printFlags(flags *pflag.FlagSet) {
	flags.VisitAll(func(flag *pflag.Flag) {
		fmt.Println(fmt.Sprintf("FLAG: --%s=%q", flag.Name, flag.Value))
	})
}