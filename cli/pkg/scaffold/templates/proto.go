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

// Import is the Import template used for new services.
var Import = `//+toolkit:import
{{ range $i := . }}import "{{$i.Path}}";{{end -}}
`

// Plugin is the Plugin template used for template.
var Plugin = `//+toolkit:plugin
option (toolkit.plugins) = {
  plugin: "{{.Name}}"
	prefix: "{{.Prefix}}"
};
`

// ProtoService is the .proto file template used for new service services.
var ProtoService = `syntax = "proto3";

package {{dehyphen .Service}};

option go_package = "{{ if not .Vendor }}./gen;{{else}}{{.Vendor}}{{.Service}}/gen;{{end}}{{dehyphen .Service}}";
{{if .AnnotationImport}}
import "github.com/lastbackend/toolkit/protoc-gen-toolkit/toolkit/options/annotations.proto";
{{end}}
//+toolkit:import
import "validate/validate.proto";


// Install plugins

//+toolkit:plugin

{{- if .RedisPlugin}}
option (toolkit.plugins) = {
  prefix: "redis"
  plugin: "redis"
};
{{end}}
{{- if .RabbitMQPlugin}}
option (toolkit.plugins) = {
  prefix: "amqp"
  plugin: "rabbitmq"
};
{{end}}
{{- if .PostgresPGPlugin}}
option (toolkit.plugins) = {
  prefix: "pgsql"
  plugin: "postgres_pg"
};
{{end}}
{{- if .PostgresGORMPlugin}}
option (toolkit.plugins) = {
  prefix: "pgsql"
  plugin: "postgres_gorm"
};
{{end}}
{{- if .PostgresPlugin}}
option (toolkit.plugins) = {
  prefix: "pgsql"
  plugin: "postgres"
};
{{end}}
{{- if .CentrifugePlugin}}
option (toolkit.plugins) = {
  prefix: "centrifuge"
  plugin: "centrifuge"
};
{{end}}

// RPC methods

service {{title .Service}} {
	//+toolkit:method
  rpc HelloWorld(HelloWorldRequest) returns (HelloWorldResponse) {}
};

// RPC request/response messages

//+toolkit:message

message HelloWorldRequest {
  string name = 1 [(validate.rules).string.min_len = 1];
}

message HelloWorldResponse {
  string name = 1;
}
`

// Method is the MethodRPC template used for new services.
var Method = `//+toolkit:method
{{- if .ExposeHTTP}}
	rpc {{.Name | camel}}({{.Name | camel}}Request) returns ({{.Name | camel}}Response) {
		option (google.api.http) = {
			{{.ExposeHTTP.Method}}: "/{{.ExposeHTTP.Path}}"
		};
		{{- if .RPCProxy}}
		option (toolkit.server).http_proxy = {
			service: "{{.RPCProxy.Service}}"
			method: "/{{.RPCProxy.Method}}"
		};		
		{{- end}}
	}
	{{- else}}{{if .ExposeWS}}
	rpc {{.Name | camel}}({{.Name | camel}}Request) returns ({{.Name | camel}}Response) {
		option (toolkit.server).websocket = true;
		option (google.api.http) = {
			get: "/{{.ExposeWS.Path}}"
		};
	}
	{{- else}}{{if .SubscriptionWS}}
	rpc {{.Name | camel}}({{.Name | camel}}Request) returns ({{.Name | camel}}Response) {
		option (toolkit.server).websocket_proxy = {
			service: "{{.RPCProxy.Service}}"
			method: "/{{.RPCProxy.Method}}"
		};
	}
	{{- else}}
	rpc {{.Name | camel}}({{.Name | camel}}Request) returns ({{.Name | camel}}Response) { }
	{{- end}}{{- end}}
{{end -}}
`

// Message is the MessageRPC template used for new services.
var Message = `//+toolkit:message
message {{.Name | camel}} {}
`
