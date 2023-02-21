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

package add

import (
	"fmt"
	"os"
	"path"
	"regexp"

	"github.com/lastbackend/toolkit/cli/pkg/scaffold/injector"
	"github.com/pkg/errors"
	"github.com/urfave/cli/v2"
)

var (
	exposeHTTPRegexp = regexp.MustCompile(`^(get|post|put|delete):\/([-a-zA-Z0-9@:%._\\+~#?&\/\/={}]*)$`)
	rpcProxyRegexp   = regexp.MustCompile(`^(.+):\/([a-zA-Z0-9._\/]*)$`)
	exposeWSSRegexp  = regexp.MustCompile(`^\/([a-zA-Z0-9._\/]*)$`)
)

var methodFlags = []cli.Flag{
	&cli.StringFlag{
		Name:  "path",
		Usage: "Path to proto directory",
		Value: "./proto",
	},
	&cli.StringFlag{
		Name:   "expose-http",
		Usage:  "Add http handler",
		Action: validateExposeHTTPFlag,
	},
	&cli.StringFlag{
		Name:   "expose-ws",
		Usage:  "Add method subscribe for connection to websocket",
		Action: validateExposeWSFlag,
	},
	&cli.BoolFlag{
		Name:   "subscribe-ws",
		Usage:  "Add method for proxy events from websocket to grpc method",
		Action: validateSubscribeWsFlag,
	},
	&cli.StringFlag{
		Name:   "proxy",
		Usage:  "Add proxy to grpc method from http handler",
		Action: validateProxyFlag,
	},
}

func AddMethod(ctx *cli.Context) error {
	arg := ctx.Args().First()
	if len(arg) == 0 {
		return cli.ShowSubcommandHelp(ctx)
	}

	workdir := ctx.String("path")
	if path.IsAbs(workdir) {
		fmt.Println("must provide a relative path as service name")
		return nil
	}

	if _, err := os.Stat(workdir); os.IsNotExist(err) {
		return fmt.Errorf("%s not exists", workdir)
	}

	methOpt, err := makeMethodOptions(ctx, arg)
	if err != nil {
		return err
	}

	i := injector.New(
		injector.Workdir(workdir),
		injector.Method(methOpt),
	)

	if err := i.Inject(); err != nil {
		return err
	}

	return nil
}

func makeMethodOptions(ctx *cli.Context, name string) (injector.MethodOption, error) {
	opts := injector.MethodOption{
		Name: name,
	}
	if val := ctx.String("expose-http"); val != "" {
		match := exposeHTTPRegexp.FindStringSubmatch(val)
		opts.ExposeHTTP = new(injector.ExposeHTTP)
		opts.ExposeHTTP.Method = match[1]
		opts.ExposeHTTP.Path = match[2]
		if val := ctx.String("proxy"); val != "" {
			match = rpcProxyRegexp.FindStringSubmatch(val)
			opts.RPCProxy = new(injector.RpcProxy)
			opts.RPCProxy.Service = match[1]
			opts.RPCProxy.Method = match[2]
		}
		return opts, nil
	}

	if val := ctx.String("expose-ws"); val != "" {
		match := exposeWSSRegexp.FindStringSubmatch(val)
		opts.ExposeWS = new(injector.ExposeWS)
		opts.ExposeWS.Method = "get"
		opts.ExposeWS.Path = match[1]
		return opts, nil
	}

	if ctx.Bool("subscribe-ws") {
		opts.SubscriptionWS = true
		if val := ctx.String("proxy"); val != "" {
			match := rpcProxyRegexp.FindStringSubmatch(val)
			opts.RPCProxy = new(injector.RpcProxy)
			opts.RPCProxy.Service = match[1]
			opts.RPCProxy.Method = match[2]
		}
		return opts, nil
	}

	return opts, nil
}

func validateExposeHTTPFlag(ctx *cli.Context, arg string) error {
	if !exposeHTTPRegexp.MatchString(arg) {
		return cli.ShowSubcommandHelp(ctx)
	}
	return nil
}

func validateExposeWSFlag(ctx *cli.Context, arg string) error {
	if !exposeWSSRegexp.MatchString(arg) {
		return cli.ShowSubcommandHelp(ctx)
	}
	return nil
}

func validateSubscribeWsFlag(ctx *cli.Context, arg bool) error {
	if (arg && len(ctx.String("expose-http")) != 0) || (arg && len(ctx.String("proxy")) == 0) {
		if err := cli.ShowSubcommandHelp(ctx); err != nil {
			return err
		}
		return errors.New("\nfailed: the subscribe-ws flag cannot be used together with a flag expose-http")
	}
	return nil
}

func validateProxyFlag(ctx *cli.Context, arg string) error {
	if (len(ctx.String("expose-http")) == 0 && !ctx.Bool("subscribe-ws")) ||
		(len(ctx.String("expose-http")) != 0 && ctx.Bool("subscribe-ws")) {
		if err := cli.ShowSubcommandHelp(ctx); err != nil {
			return err
		}
		return errors.New("\nfailed: the proxy flag must be use with one flag expose-http or expose-ws")
	}
	if !rpcProxyRegexp.MatchString(arg) {
		return cli.ShowSubcommandHelp(ctx)
	}
	return nil
}
