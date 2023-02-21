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
	"text/tabwriter"

	"github.com/jhump/protoreflect/desc/protoparse"
	tcli "github.com/lastbackend/toolkit/cli/cmd"
	"github.com/lastbackend/toolkit/cli/pkg/util/filesystem"
	"github.com/urfave/cli/v2"
)

var flags = []cli.Flag{
	&cli.StringFlag{
		Name:  "path",
		Usage: "Path to proto dir",
		Value: "./proto",
	},
}

func init() {
	tcli.Register(&cli.Command{
		Name:   "describe",
		Usage:  "Get describe service info",
		Action: Describe,
		Flags:  flags,
	})
}

func Describe(ctx *cli.Context) error {
	files, err := filesystem.WalkMatch(ctx.String("path"), "*.proto")
	if err != nil {
		return err
	}

	parser := protoparse.Parser{}
	desc, err := parser.ParseFilesButDoNotLink(files...)
	if err != nil {
		return err
	}

	w := new(tabwriter.Writer)
	w.Init(os.Stdout, 8, 8, 0, '\t', 0)
	defer w.Flush()

	for _, d := range desc {
		for _, svc := range d.Service {
			fmt.Fprintf(w, "\n Service: %s\n", svc.GetName())
			for index, meth := range svc.Method {
				if !meth.GetClientStreaming() && !meth.GetClientStreaming() {
					fmt.Fprintf(w, "\n %d. rpc %s(ctx context.Context, req *%s) returns (resp *%s, err error) \t", index+1, meth.GetName(), meth.GetInputType(), meth.GetOutputType())
				} else if meth.GetClientStreaming() {
					fmt.Fprintf(w, "\n %d. rpc %s(req *%s) returns (stream *%s, error) \t", index+1, meth.GetName(), meth.GetInputType(), meth.GetOutputType())
				} else {
					fmt.Fprintf(w, "\n %d. rpc %s(stream *%s) returns error \t", index+1, meth.GetName(), meth.GetInputType())
				}
			}
			fmt.Fprintf(w, "\n")
		}
	}
	fmt.Fprintf(w, "\n")

	return nil
}
