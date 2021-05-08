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
	"reflect"

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
	Flags   []*Flag
	opts    Options
	rootCmd *cobra.Command
}

type Option func(o *Options)

var (
	DefaultFlags = []Flag{
		{
			Name:    "server_address",
			EnvVars: []string{"ENGINE_SERVER_ADDRESS"},
			Usage:   "Bind address for the server. 127.0.0.1:8080",
		},
	}
)

func NewCmd(opts ...Option) Cmd {
	options := Options{}

	for _, o := range opts {
		o(&options)
	}

	if len(options.ShortDesc) == 0 {
		options.ShortDesc = defaultShortDescription
	}

	c := new(cmd)

	for _, f := range DefaultFlags {
		c.Flags = append(c.Flags, &f)
	}

	c.opts = options
	c.rootCmd = &cobra.Command{}

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

	global := pflag.CommandLine

	cleanFlagSet := pflag.NewFlagSet(c.opts.Name, pflag.ContinueOnError)
	cleanFlagSet.SetNormalizeFunc(wordSepNormalizeFunc)

	c.rootCmd.PersistentPreRun = func(cmd *cobra.Command, args []string) {
		//fmt.Println("1============")
		//printFlags(cmd.Flags())
		//fmt.Println("============")
	}

	c.rootCmd.PreRun = func(cmd *cobra.Command, args []string) {
		//fmt.Println("2============")
		//printFlags(cmd.Flags())
		//fmt.Println("============")
	}

	c.rootCmd.RunE = func(cmd *cobra.Command, args []string) error {

		// short-circuit on help
		help, err := cmd.Flags().GetBool("help")
		if err != nil {
			fmt.Println("help >>>>>>", err)
			return errors.Wrapf(err, "\"help\" flag is non-bool, programmer error, please correct")
		}
		if help {
			return cmd.Help()
		}

		//cfgFile, err := cmd.Flags().GetString("config")
		//if err != nil {
		//	return errors.Wrapf(err, "\"config\" flag is non-string, programmer error, please correct")
		//}

		fmt.Println("3============")
		printFlags(cmd.Flags())
		fmt.Println("============")

		for _, val := range c.Flags {
			var err error
			var v interface{}
			switch t:=val.Value.(type) {
			case int:
				v, err = cmd.Flags().GetInt(val.Name)
			case string:
				v, err = cmd.Flags().GetString(val.Name)
			case bool:
				v, err = cmd.Flags().GetBool(val.Name)
			default:
				if val.Value != nil {
					fmt.Println("t >>>>>",  reflect.Indirect(reflect.ValueOf(t).Elem()).Kind())
				}
				v, err = cmd.Flags().GetString(val.Name)
			}
			val.Changed = cmd.Flags().Changed(val.Name)
			fmt.Println("1@ >>>>>>>>>>>>>", val.Name, val.Value)
			if err != nil {
				return errors.Wrapf(err, `"`+val.Name+`" flag is non-`+getType(val)+`, programmer error, please correct`)
			} else if val.Changed && val.Value != nil {
				fmt.Println("2@ >>>>>>>>>>>>>")
				reflect.ValueOf(val.Value).Elem().Set(reflect.ValueOf(v))
				// check if any type is its default(zero) value
			} else if val.Value != nil && reflect.DeepEqual(val.Value, reflect.Zero(reflect.TypeOf(val.Value)).Interface()) {
				fmt.Println("3@ >>>>>>>>>>>>>")
				reflect.ValueOf(val.Value).Elem().Set(reflect.ValueOf(val.DefValue))
			}
		}

		return nil
	}

	c.rootCmd.AddCommand(c.versionCommand())

	global.BoolP("help", "h", false, fmt.Sprintf("Show help for command"))
	global.BoolP("debug", "d", false, "Enable debug mode")
	global.StringP("config", "c", "", "Set config filepath")

	AddGlobalFlags(cleanFlagSet)

	return c
}

func (c *cmd) Options() Options {
	return c.opts
}

//func (c *cmd) AddFlag(flag *Flag) {
//	c.flags = append(c.flags, flag)
//	c.rootCmd.Flags().StringP(flag.Name, flag.Shorthand, "", flag.Usage)
//}

func (c *cmd) Get() *cmd {
	return c
}

func (c *cmd) Execute() error {

	for _, val := range c.Flags {
		switch val.Value.(type) {
		case int:
			fmt.Println("1 :: >>>>>>", val.Name)
			c.rootCmd.Flags().IntP(val.Name, val.Shorthand, 0, val.Usage)
		case string:
			fmt.Println("2 :: >>>>>>", val.Name)
			c.rootCmd.Flags().StringP(val.Name, val.Shorthand, "", val.Usage)
		case bool:
			fmt.Println("3 :: >>>>>>", val.Name)
			c.rootCmd.Flags().BoolP(val.Name, val.Shorthand, false, val.Usage)
		default:
			fmt.Println("4 :: >>>>>>", val.Name)
			c.rootCmd.Flags().StringP(val.Name, val.Shorthand, "", val.Usage)
			continue
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

func getType(i interface{}) string {
	if t := reflect.TypeOf(i); t.Kind() == reflect.Ptr {
		return "*" + t.Elem().Name()
	} else {
		return t.Name()
	}
}
