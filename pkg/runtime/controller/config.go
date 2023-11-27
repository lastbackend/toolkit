package controller

import (
	"context"
	"fmt"
	"github.com/caarlos0/env/v7"
	"github.com/fatih/structs"
	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/jedib0t/go-pretty/v6/text"
	"github.com/lastbackend/toolkit/pkg/runtime"
	"github.com/lastbackend/toolkit/pkg/runtime/meta"
	"strings"
)

const ConfigPrefixSeparator = "_"

type configController struct {
	runtime.Config

	runtime runtime.Runtime
	configs []any
	parsed  map[string]interface{}
	prefix  string
}

func (c *configController) buildPrefix(prefix string) string {

	pt := make([]string, 0)
	if c.prefix != "" {
		pt = append(pt, c.prefix)
	}

	if prefix != "" {
		pt = append(pt, prefix)
	}

	if len(pt) > 0 {
		pt = append(pt, "")
		return strings.ToUpper(strings.Join(pt, ConfigPrefixSeparator))
	}

	return ""
}

func (c *configController) Configs() []any {
	return c.configs
}

func (c *configController) Provide(cfgs ...any) error {

	if len(cfgs) == 0 {
		return nil
	}

	for _, cfg := range cfgs {
		if err := c.Parse(cfg, ""); err != nil {
			c.runtime.Log().V(5).Errorf("can not parse config: %s", err.Error())
			return err
		}
		c.configs = append(c.configs, cfg)
	}

	return nil
}

func (c *configController) SetMeta(meta *meta.Meta) {
	c.prefix = meta.GetEnvPrefix()
}

func (c *configController) Parse(v interface{}, prefix string, opts ...env.Options) error {
	c.parsed[prefix] = v
	opts = append(opts, env.Options{Prefix: c.buildPrefix(prefix)})
	return env.Parse(v, opts...)
}

func (c *configController) Print(v interface{}, prefix string) {

	tw := table.NewWriter()
	tw.AppendHeader(table.Row{"ENVIRONMENT", "DEFAULT VALUE", "REQUIRED", "DESCRIPTION"})
	tw.Style().Options.DrawBorder = true
	tw.Style().Options.SeparateColumns = true
	tw.Style().Options.SeparateFooter = false
	tw.Style().Options.SeparateHeader = true
	tw.Style().Options.SeparateRows = true

	fields := structs.Fields(v)
	for _, field := range fields {

		tagE := field.Tag("env")
		tagD := field.Tag("envDefault")
		tagR := field.Tag("required")
		tagC := field.Tag("comment")

		if tagE != "" {
			tw.AppendRows([]table.Row{
				{fmt.Sprintf("%s%s", c.buildPrefix(prefix), tagE), tagD, tagR, text.WrapText(tagC, 120)},
			})
		}

	}

	fmt.Println(tw.Render())
}

func (c *configController) PrintTable(all, nocomments bool) string {

	tw := table.NewWriter()

	if nocomments {
		tw.AppendHeader(table.Row{"ENVIRONMENT", "DEFAULT VALUE"})
	} else {
		tw.AppendHeader(table.Row{"ENVIRONMENT", "DEFAULT VALUE", "REQUIRED", "DESCRIPTION"})
	}

	tw.Style().Options.DrawBorder = true
	tw.Style().Options.SeparateColumns = true
	tw.Style().Options.SeparateFooter = false
	tw.Style().Options.SeparateHeader = true
	tw.Style().Options.SeparateRows = true

	for prefix, v := range c.parsed {
		fields := structs.Fields(v)
		for _, field := range fields {

			tagE := field.Tag("env")
			tagD := field.Tag("envDefault")
			tagR := field.Tag("required")
			tagC := field.Tag("comment")

			if tagE != "" {

				if tagD == "" && tagR == "" && !all {
					continue
				}

				if nocomments {
					tw.AppendRows([]table.Row{
						{fmt.Sprintf("%s%s", c.buildPrefix(prefix), tagE), tagD},
					})
				} else {
					tw.AppendRows([]table.Row{
						{fmt.Sprintf("%s%s", c.buildPrefix(prefix), tagE), tagD, tagR, text.WrapText(tagC, 120)},
					})
				}

			}

		}
	}

	return tw.Render()
}

func (c *configController) PrintYaml(all, nocomments bool) string {

	yamlStr := "---"

	for prefix, v := range c.parsed {
		fields := structs.Fields(v)
		for _, field := range fields {

			tagE := field.Tag("env")
			tagD := field.Tag("envDefault")
			tagR := field.Tag("required")
			tagC := field.Tag("comment")

			if tagE != "" {

				if tagD == "" && tagR == "" && !all {
					continue
				}
				if tagC != "" && !nocomments {
					yamlStr = fmt.Sprintf("%s\n// %s", yamlStr, tagC)
				}
				if tagD != "" {
					yamlStr = fmt.Sprintf("%s\n%s: '%s'", yamlStr, fmt.Sprintf("%s%s", c.buildPrefix(prefix), tagE), tagD)
				} else {
					yamlStr = fmt.Sprintf("%s\n%s:", yamlStr, fmt.Sprintf("%s%s", c.buildPrefix(prefix), tagE))
				}

			}

		}
	}

	return yamlStr
}

func newConfigController(_ context.Context, runtime runtime.Runtime) runtime.Config {
	cfg := new(configController)
	cfg.runtime = runtime

	cfg.configs = make([]interface{}, 0)
	cfg.parsed = make(map[string]interface{}, 0)

	return cfg
}
