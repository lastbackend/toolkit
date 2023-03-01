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

package config

import (
	"fmt"
	"github.com/caarlos0/env/v7"
	"github.com/fatih/structs"
	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/jedib0t/go-pretty/v6/text"
	"strings"
)

type Config struct {
	Debug         bool
	ServerAddress string
}

func Parse(v interface{}, prefix string, opts ...env.Options) error {
	opts = append(opts, env.Options{Prefix: strings.ToUpper(prefix)})
	return env.Parse(v, opts...)
}

func Print(v interface{}, prefix string) {

	tw := table.NewWriter()
	tw.AppendHeader(table.Row{"ENVIRONMENT", "DEFAULT VALUE", "REQUIRED", "DESCRIPTION"})
	tw.Style().Options.DrawBorder = true
	tw.Style().Options.SeparateColumns = true
	tw.Style().Options.SeparateFooter = false
	tw.Style().Options.SeparateHeader = true
	tw.Style().Options.SeparateRows = true

	tw.SetCaption("(c) No one!")

	fields := structs.Fields(v)
	for _, field := range fields {

		tagE := field.Tag("env")
		tagD := field.Tag("envDefault")
		tagR := field.Tag("required")
		tagC := field.Tag("comment")

		if tagE != "" {
			tw.AppendRows([]table.Row{
				{fmt.Sprintf("%s_%s", strings.ToUpper(prefix), tagE), tagD, tagR, text.WrapText(tagC, 120)},
			})
		}
	}

	fmt.Println(tw.Render())
}
