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

package engine

type meta struct {
	Name            string
	EnvPrefix       string
	ShorDescription string
	LongDescription string
	Version         string
}

func (m *meta) SetVersion(s string) *meta {
	m.Version = s
	return m
}

func (m *meta) SetEnvPrefix(s string) *meta {
	m.EnvPrefix = s
	return m
}

func (m *meta) SetShortDescription(s string) *meta {
	m.ShorDescription = s
	return m
}

func (m *meta) SetLongDescription(s string) *meta {
	m.LongDescription = s
	return m
}
