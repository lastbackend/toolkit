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

package formpb

import (
	"github.com/lastbackend/toolkit/pkg/server/http/marshaler"
	"github.com/lastbackend/toolkit/pkg/util/converter"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"

	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/url"
	"reflect"
)

var (
	protoMessageType = reflect.TypeOf((*proto.Message)(nil)).Elem()
)

type protoEnum interface {
	fmt.Stringer
	EnumDescriptor() ([]byte, []int)
}

type FORMPb struct {
	protojson.MarshalOptions
	protojson.UnmarshalOptions
}

// ContentType always returns "application/x-www-form-urlencoded".
func (*FORMPb) ContentType() string {
	return "application/x-www-form-urlencoded"
}

// Marshal marshals "v" into JSON
// Currently it can marshal only proto.Message.
func (j *FORMPb) Marshal(v interface{}) ([]byte, error) {
	if _, ok := v.(proto.Message); !ok {
		return j.marshalNonProtoField(v)
	}

	var buf bytes.Buffer
	if err := j.marshalTo(&buf, v); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

// Unmarshal not realized for the moment
func (j *FORMPb) Unmarshal(data []byte, v interface{}) error {
	return nil
}

// NewDecoder returns a Decoder which reads form data stream from "r".
func (j *FORMPb) NewDecoder(r io.Reader) marshaler.Decoder {
	return marshaler.DecoderFunc(func(v interface{}) error { return decodeFORMPb(r, v) })
}

// NewEncoder returns an Encoder which writes JSON stream into "w".
func (j *FORMPb) NewEncoder(w io.Writer) marshaler.Encoder {
	return marshaler.EncoderFunc(func(v interface{}) error { return j.marshalTo(w, v) })
}

func decodeFORMPb(d io.Reader, v interface{}) error {
	msg, ok := v.(proto.Message)
	if !ok {
		return fmt.Errorf("not proto message")
	}

	formData, err := io.ReadAll(d)
	if err != nil {
		return err
	}

	values, err := url.ParseQuery(string(formData))
	if err != nil {
		return err
	}

	if err := converter.ParseRequestQueryParametersToProto(msg, values); err != nil {
		return err
	}

	return nil
}

func (j *FORMPb) marshalTo(w io.Writer, v interface{}) error {
	p, ok := v.(proto.Message)
	if !ok {
		buf, err := j.marshalNonProtoField(v)
		if err != nil {
			return err
		}
		_, err = w.Write(buf)
		return err
	}

	b, err := j.MarshalOptions.Marshal(p)
	if err != nil {
		return err
	}

	_, err = w.Write(b)

	return err
}

func (j *FORMPb) marshalNonProtoField(v interface{}) ([]byte, error) {
	if v == nil {
		return []byte("null"), nil
	}

	rv := reflect.ValueOf(v)

	for rv.Kind() == reflect.Ptr {
		if rv.IsNil() {
			return []byte("null"), nil
		}
		rv = rv.Elem()
	}

	if rv.Kind() == reflect.Slice {
		if rv.IsNil() {
			if j.EmitUnpopulated {
				return []byte("[]"), nil
			}
			return []byte("null"), nil
		}

		if rv.Type().Elem().Implements(protoMessageType) {
			var buf bytes.Buffer
			err := buf.WriteByte('[')
			if err != nil {
				return nil, err
			}
			for i := 0; i < rv.Len(); i++ {
				if i != 0 {
					err = buf.WriteByte(',')
					if err != nil {
						return nil, err
					}
				}
				if err = j.marshalTo(&buf, rv.Index(i).Interface().(proto.Message)); err != nil {
					return nil, err
				}
			}
			err = buf.WriteByte(']')
			if err != nil {
				return nil, err
			}

			return buf.Bytes(), nil
		}
	}

	if rv.Kind() == reflect.Map {
		m := make(map[string]*json.RawMessage)
		for _, k := range rv.MapKeys() {
			buf, err := j.Marshal(rv.MapIndex(k).Interface())
			if err != nil {
				return nil, err
			}
			m[fmt.Sprintf("%v", k.Interface())] = (*json.RawMessage)(&buf)
		}
		if j.Indent != "" {
			return json.MarshalIndent(m, "", j.Indent)
		}
		return json.Marshal(m)
	}

	if enum, ok := rv.Interface().(protoEnum); ok && !j.UseEnumNumbers {
		return json.Marshal(enum.String())
	}

	return json.Marshal(rv.Interface())
}
