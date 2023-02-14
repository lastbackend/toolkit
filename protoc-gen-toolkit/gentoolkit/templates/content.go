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

package templates

// ContentTpl is the content template used for new services.
var ContentTpl = `
// This is a compile-time assertion to ensure that this generated file
// is compatible with the toolkit package it is being compiled against and
// suppress "imported and not used" errors
var _ context.Context
var _ logger.Logger
var _ emptypb.Empty
var _ server.Server
var _ grpc.Client
var _ http.Handler
var _ errors.Err
var _ io.Reader
var _ json.Marshaler
var _ ws.Client
{{- if not .HasNotServer }}
var _ server.Server
{{ end }}

{{ template "services-content" . }}

{{ if not .HasNotServer }}
	{{ template "server-content" . }}
{{ end }}
`
