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

package injector

import (
	"bytes"
	"fmt"
	"github.com/jhump/protoreflect/desc/protoparse"
	"github.com/lastbackend/toolkit/cli/pkg/scaffold/templates"
	ustrings "github.com/lastbackend/toolkit/cli/pkg/util/strings"
	"os"
	"strings"
	"text/template"

	"github.com/lastbackend/toolkit/cli/pkg/util/filesystem"
	"github.com/pkg/errors"
	"google.golang.org/protobuf/types/descriptorpb"
)

type Injector interface {
	Inject() error
	Options() Options
}

type injector struct {
	opts Options
}

type descriptor struct {
	fileMap map[string]*File
}

type File struct {
	*descriptorpb.FileDescriptorProto
}

func (i *injector) Inject() error {
	files, err := filesystem.WalkMatch(i.opts.Workdir, "*.proto")
	if err != nil {
		return err
	}

	parser := protoparse.Parser{}
	fdesc, err := parser.ParseFilesButDoNotLink(files...)
	if err != nil {
		return err
	}

	desc := new(descriptor)
	desc.fileMap = make(map[string]*File)
	for _, d := range fdesc {
		desc.fileMap[d.GetName()] = &File{d}
	}

	for _, file := range files {
		fDesc := desc.fileMap[file]
		imports := make([]ImportOption, 0)

		methodSlice := make([]string, 0)
		messageSlice := make([]string, 0)
		for _, service := range fDesc.Service {
			for _, meth := range service.Method {
				methodSlice = append(methodSlice, meth.GetName())
				messageSlice = append(messageSlice, meth.GetInputType())
				messageSlice = append(messageSlice, meth.GetOutputType())
			}
		}

		read, err := os.ReadFile(file)
		if err != nil {
			return err
		}

		newContents := string(read)
		messages := make([]MessageOption, 0)

		// Inject methods
		if i.opts.Methods != nil {

			methT, err := template.New("method").Funcs(tplFn).Parse(templates.Method)
			if err != nil {
				return err
			}

			for _, meth := range i.opts.Methods {
				if meth.ExposeHTTP != nil || meth.ExposeWS != nil {

					exists := contains(fDesc.Dependency, "google/api/annotations.proto")
					if !exists {
						imports = append(imports, []ImportOption{
							{Path: "google/api/annotations.proto"},
						}...)
					}
				}

				if exists := contains(methodSlice, meth.Name); exists {
					return fmt.Errorf("failed: duplicate method %s name", meth.Name)
				}

				messageRequest := meth.Name + "Request"
				messageResponse := meth.Name + "Response"

				if exists := contains(messageSlice, messageRequest); exists {
					return fmt.Errorf("failed: duplicate message %s name", messageRequest)
				}
				if exists := contains(messageSlice, messageResponse); exists {
					return fmt.Errorf("failed: duplicate message %s name", messageResponse)
				}

				messages = append(messages, MessageOption{Name: messageRequest})
				messages = append(messages, MessageOption{Name: messageResponse})

				var b bytes.Buffer
				if err := methT.Execute(&b, meth); err != nil {
					return errors.Wrap(err, "Failed to execute method template")
				}
				newContents = strings.Replace(newContents, "//+toolkit:method", b.String(), -1)
			}
		}

		// Inject messages
		msgT, err := template.New("method").Funcs(tplFn).Parse(templates.Message)
		if err != nil {
			return err
		}

		for _, msg := range messages {
			var b bytes.Buffer
			if err := msgT.Execute(&b, msg); err != nil {
				return errors.Wrap(err, "Failed to execute method template")
			}
			newContents = strings.Replace(newContents, "//+toolkit:message", b.String(), -1)
		}

		// Inject plugins
		plgT, err := template.New("plugin").Funcs(tplFn).Parse(templates.Plugin)
		if err != nil {
			return err
		}

		for _, p := range i.opts.Plugins {
			var b bytes.Buffer
			if err := plgT.Execute(&b, p); err != nil {
				return errors.Wrap(err, "Failed to execute method template")
			}
			newContents = strings.Replace(newContents, "//+toolkit:plugin", b.String(), -1)
		}

		// Inject imports
		if len(imports) != 0 {
			impT, err := template.New("import").Funcs(tplFn).Parse(templates.Import)
			if err != nil {
				return err
			}
			var b bytes.Buffer
			if err := impT.Execute(&b, imports); err != nil {
				return errors.Wrap(err, "Failed to execute import template")
			}
			newContents = strings.Replace(newContents, "//+toolkit:import", b.String(), -1)
		}

		// Write new content
		err = os.WriteFile(file, []byte(newContents), 0)
		if err != nil {
			return err
		}
	}

	return nil
}

func (i *injector) Options() Options {
	return i.opts
}

func New(opts ...Option) Injector {
	var options Options
	for _, o := range opts {
		o(&options)
	}
	return &injector{
		opts: options,
	}
}

func contains(s []string, str string) bool {
	for _, v := range s {
		if v == str {
			return true
		}
	}

	return false
}

var tplFn = template.FuncMap{
	"dehyphen":    ustrings.DeHyphenFunc,
	"lowerhyphen": ustrings.LowerHyphenFunc,
	"tohyphen":    ustrings.ToHyphen,
	"title":       ustrings.Title,
	"lower":       ustrings.ToLower,
	"upper":       ustrings.ToUpper,
	"camel":       ustrings.ToCamel,
}
