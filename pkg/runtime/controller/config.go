package controller

import (
	"context"
	"fmt"
	"github.com/caarlos0/env/v7"
	"github.com/fatih/structs"
	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/jedib0t/go-pretty/v6/text"
	"github.com/lastbackend/toolkit/pkg/runtime"
	"github.com/lastbackend/toolkit/pkg/runtime/logger"
	"github.com/lastbackend/toolkit/pkg/runtime/meta"
	"strings"
)

const ConfigPrefixSeparator = "_"

type configController struct {
	runtime.Config

	log     logger.Logger
	configs []interface{}
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

func (c *configController) Configs() []interface{} {
	return c.configs
}

func (c *configController) Provide(cfg interface{}) {
	c.configs = append(c.configs, cfg)
	return
}

func (c *configController) SetMeta(meta *meta.Meta) {
	c.prefix = meta.GetEnvPrefix()
}

func (c *configController) Parse(v interface{}, prefix string, opts ...env.Options) error {
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

	tw.SetCaption("(c) No one!")

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

func newConfigController(_ context.Context, log logger.Logger) runtime.Config {
	cfg := new(configController)
	cfg.log = log
	return cfg
}
