/*
Copyright [2014] - [2022] The Last.Backend authors.

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

type MetaInfo interface {
	SetVersion(version string) Meta
	SetEnvPrefix(prefix string) Meta
	SetShortDescription(desc string) Meta
	SetLongDescription(desc string) Meta
	GetName() string
	GetVersion() string
	GetEnvPrefix() string
	GetShortDescription() string
	GetLongDescription() string
}
type Meta interface {
	MetaInfo

	SetName(name string) Meta
}

type meta struct {
	Name            string
	EnvPrefix       string
	ShorDescription string
	LongDescription string
	Version         string
}

func (m *meta) SetName(s string) Meta {
	m.Name = s
	return m
}

func (m *meta) GetName() string {
	return m.Name
}

func (m *meta) SetVersion(s string) Meta {
	m.Version = s
	return m
}

func (m *meta) GetVersion() string {
	return m.Name
}

func (m *meta) SetEnvPrefix(s string) Meta {
	EnvPrefix = s
	m.EnvPrefix = s
	return m
}

func (m *meta) GetEnvPrefix() string {
	return m.EnvPrefix
}

func (m *meta) SetShortDescription(s string) Meta {
	m.ShorDescription = s
	return m
}

func (m *meta) GetShortDescription() string {
	return m.ShorDescription
}

func (m *meta) SetLongDescription(s string) Meta {
	m.LongDescription = s
	return m
}

func (m *meta) GetLongDescription() string {
	return m.LongDescription
}
