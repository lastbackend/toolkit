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
	"github.com/lastbackend/engine/logger"
	"github.com/lastbackend/engine/network/resolver"
	"github.com/lastbackend/engine/network/resolver/consul"
	"github.com/lastbackend/engine/network/resolver/local"
	"github.com/lastbackend/engine/network/resolver/route"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"

	"fmt"
	"os"
	"regexp"
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
	Run(f func() error) error
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

	// Set default services
	resolver.DefaultResolver = local.NewResolver()

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

func (c *cli) Run(fn func() error) error {

	c.rootCmd = &cobra.Command{}
	c.rootCmd.SetGlobalNormalizationFunc(wordSepNormalizeFunc)
	c.rootCmd.InitDefaultHelpFlag()
	c.rootCmd.SetHelpCommand(&cobra.Command{Use: "no-help", Hidden: true})
	c.rootCmd.AddCommand(c.versionCommand())

	c.rootCmd.PersistentFlags().BoolP("debug", "d", false, "Enable debug mode")
	c.rootCmd.Flags().StringP("resolver", "", resolver.LocalResolver, "Sets the value for initial resolver (default: local)")
	c.rootCmd.Flags().StringP("resolver-endpoint", "", resolver.LocalResolver, "Sets the value for initial resolver endpoint")

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

	c.rootCmd.PreRunE = func(cmd *cobra.Command, args []string) error {

		resolverFlag, err := cmd.Flags().GetString("resolver")
		if err != nil {
			return err
		}

		resolverEndpointFlag, err := cmd.Flags().GetString("resolver-endpoint")
		if err != nil {
			return err
		}

		switch resolverFlag {
		case resolver.LocalResolver:
			resolver.DefaultResolver = local.NewResolver()
			addresses := strings.Split(resolverEndpointFlag, ",")
			for _, addr := range addresses {
				re := regexp.MustCompile("([\\w]+):(.*)")
				match := re.FindStringSubmatch(addr)
				if len(match) > 0 {
					resolver.DefaultResolver.Table().Create(route.Route{
						Service: match[1],
						Address: match[2],
					})
				}
			}
		case resolver.ConsulResolver:
			resolver.DefaultResolver = consul.NewResolver(consul.WithEndpoint(resolverEndpointFlag))
		}
		return nil
	}

	c.rootCmd.RunE = func(cmd *cobra.Command, args []string) error {
		debugFlag, err := cmd.Flags().GetBool("debug")
		if err != nil {
			return err
		}

		if debugFlag {
			logger.DefaultLogger.Init(logger.Options{
				Level: logger.DebugLevel,
			})
			printFlags(cmd.Flags())
		}

		return fn()
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

// printFlags logs the flags in the flagSet
func printFlags(flags *pflag.FlagSet) {
	flags.VisitAll(func(flag *pflag.Flag) {
		fmt.Println(fmt.Sprintf("FLAG: --%s=%q", flag.Name, flag.Value))
	})
}
