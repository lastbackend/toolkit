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

package generator

import (
	"os"
	"path/filepath"
	"text/template"

	ustrings "github.com/lastbackend/toolkit/cli/pkg/util/strings"
)

type Generator interface {
	Generate([]File) error
	Options() Options
}

type generator struct {
	opts Options
}

type File struct {
	Path     string
	Template string
	FileMode os.FileMode
}

func (g *generator) Generate(files []File) error {
	for _, file := range files {
		fp := filepath.Join(g.opts.Workdir, file.Path)
		dir := filepath.Dir(fp)

		if file.Template == "" {
			dir = fp
		}

		if _, err := os.Stat(dir); os.IsNotExist(err) {
			if err := os.MkdirAll(dir, 0755); err != nil {
				return err
			}
		}

		if file.Template == "" {
			continue
		}

		fn := template.FuncMap{
			"title":       ustrings.Title,
			"lower":       ustrings.ToLower,
			"upper":       ustrings.ToUpper,
			"camel":       ustrings.ToCamel,
			"git":         ustrings.GitParse,
			"dehyphen":    ustrings.DeHyphenFunc,
			"lowerhyphen": ustrings.LowerHyphenFunc,
			"tohyphen":    ustrings.ToHyphen,
		}

		t, err := template.New(fp).Funcs(fn).Parse(file.Template)
		if err != nil {
			return err
		}

		f, err := os.Create(fp)
		if err != nil {
			return err
		}

		if file.FileMode != 0 {
			if err := f.Chmod(file.FileMode); err != nil {
				return err
			}
		}

		err = t.Execute(f, g.opts)
		if err != nil {
			return err
		}
	}

	return nil
}

func (g *generator) Options() Options {
	return g.opts
}

func New(opts ...Option) Generator {
	var options Options
	for _, o := range opts {
		o(&options)
	}
	return &generator{
		opts: options,
	}
}
