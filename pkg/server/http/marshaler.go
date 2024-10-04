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

package http

import (
	"github.com/lastbackend/toolkit/pkg/server/http/marshaler"
	"github.com/lastbackend/toolkit/pkg/server/http/marshaler/formpb"
	"github.com/lastbackend/toolkit/pkg/server/http/marshaler/jsonpb"
	"google.golang.org/protobuf/encoding/protojson"
)

const (
	ContentTypeForm = "application/x-www-form-urlencoded"
	ContentTypeJSON = "application/json"
)

var (
	DefaultMarshaler = &jsonpb.JSONPb{
		MarshalOptions:   defaultMarshalOptions,
		UnmarshalOptions: defaultUnmarshalOptions,
	}
)

var (
	defaultMarshalOptions = protojson.MarshalOptions{
		EmitUnpopulated: true, // Include fields with default values in the JSON output
		UseProtoNames:   true, // Use original proto field names (e.g., 'start_date') instead of CamelCase (e.g., 'StartDate')
	}
	defaultUnmarshalOptions = protojson.UnmarshalOptions{
		DiscardUnknown: true, // Ignores unknown fields during deserialization
	}
)

func GetMarshalerMap() map[string]marshaler.Marshaler {
	return map[string]marshaler.Marshaler{
		ContentTypeJSON: &jsonpb.JSONPb{
			MarshalOptions:   defaultMarshalOptions,
			UnmarshalOptions: defaultUnmarshalOptions,
		},
		ContentTypeForm: &formpb.FORMPb{
			MarshalOptions:   defaultMarshalOptions,
			UnmarshalOptions: defaultUnmarshalOptions,
		},
	}
}
