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

package logger

type Option func(*Options)

func WithConsoleLevel(opt Level) Option {
	return func(args *Options) {
		args.ConsoleLevel = opt
	}
}

func WithEnableConsole(opt bool) Option {
	return func(args *Options) {
		args.EnableConsole = opt
	}
}

func WithConsoleJSONFormat(opt bool) Option {
	return func(args *Options) {
		args.ConsoleJSONFormat = opt
	}
}

func WithEnableFile(opt bool) Option {
	return func(args *Options) {
		args.EnableFile = opt
	}
}

func WithFileJSONFormat(opt bool) Option {
	return func(args *Options) {
		args.FileJSONFormat = opt
	}
}

func WithFileLevel(opt Level) Option {
	return func(args *Options) {
		args.FileLevel = opt
	}
}

func WithFileLocation(opt string) Option {
	return func(args *Options) {
		args.FileLocation = opt
	}
}

type Options struct {
	ConsoleLevel      Level
	EnableConsole     bool
	ConsoleJSONFormat bool
	EnableFile        bool
	FileJSONFormat    bool
	FileLevel         Level
	FileLocation      string
}
