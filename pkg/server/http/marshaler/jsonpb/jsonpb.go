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

package jsonpb

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/lastbackend/toolkit/pkg/server/http/marshaler"
	"github.com/lastbackend/toolkit/pkg/server/http/marshaler/util"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
	"io"
	"reflect"
	"regexp"
)

var (
	protoMessageType = reflect.TypeOf((*proto.Message)(nil)).Elem()
	typeProtoMessage = reflect.TypeOf((*proto.Message)(nil)).Elem()
	convFromType     = map[reflect.Kind]reflect.Value{
		reflect.String:  reflect.ValueOf(util.String),
		reflect.Bool:    reflect.ValueOf(util.Bool),
		reflect.Float64: reflect.ValueOf(util.Float64),
		reflect.Float32: reflect.ValueOf(util.Float32),
		reflect.Int64:   reflect.ValueOf(util.Int64),
		reflect.Int32:   reflect.ValueOf(util.Int32),
		reflect.Uint64:  reflect.ValueOf(util.Uint64),
		reflect.Uint32:  reflect.ValueOf(util.Uint32),
		reflect.Slice:   reflect.ValueOf(util.Bytes),
	}
)

type protoEnum interface {
	fmt.Stringer
	EnumDescriptor() ([]byte, []int)
}

type JSONPb struct {
	protojson.MarshalOptions
	protojson.UnmarshalOptions
}

func (*JSONPb) ContentType() string {
	return "application/json"
}

func (j *JSONPb) Marshal(v interface{}) ([]byte, error) {
	if _, ok := v.(proto.Message); !ok {
		return j.marshalNonProtoField(v)
	}

	var buf bytes.Buffer

	if err := j.marshalTo(&buf, v); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

func (j *JSONPb) Unmarshal(data []byte, v interface{}) error {
	return unmarshalJSONPb(data, j.UnmarshalOptions, v)
}

func (j *JSONPb) NewDecoder(r io.Reader) marshaler.Decoder {
	d := json.NewDecoder(r)
	return DecoderWrapper{
		Decoder:          d,
		UnmarshalOptions: j.UnmarshalOptions,
	}
}

func (j *JSONPb) Delimiter() []byte {
	return []byte("\n")
}

func (j *JSONPb) NewEncoder(w io.Writer) marshaler.Encoder {
	return marshaler.EncoderFunc(func(v interface{}) error {
		if err := j.marshalTo(w, v); err != nil {
			return err
		}
		_, err := w.Write(j.Delimiter())
		return err
	})
}

func (j *JSONPb) marshalTo(w io.Writer, v interface{}) error {
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

func (j *JSONPb) marshalNonProtoField(v interface{}) ([]byte, error) {
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

type DecoderWrapper struct {
	*json.Decoder
	protojson.UnmarshalOptions
}

func (d DecoderWrapper) Decode(v interface{}) error {
	return decodeJSONPb(d.Decoder, d.UnmarshalOptions, v)
}

func unmarshalJSONPb(data []byte, unmarshaler protojson.UnmarshalOptions, v interface{}) error {
	d := json.NewDecoder(bytes.NewReader(data))
	return decodeJSONPb(d, unmarshaler, v)
}

func decodeJSONPb(d *json.Decoder, unmarshaler protojson.UnmarshalOptions, v interface{}) error {
	p, ok := v.(proto.Message)
	if !ok {
		return decodeNonProtoField(d, unmarshaler, v)
	}
	var b json.RawMessage
	err := d.Decode(&b)
	if err != nil {
		return err
	}

	err = unmarshaler.Unmarshal(b, p)
	if err != nil {
		message := err.Error()

		re := regexp.MustCompile(`proto:\s*\(line.*\):\s*(.*)$`)
		match := re.FindStringSubmatch(message)

		if len(match) > 1 {
			message = match[1]
		}

		return errors.New(message)
	}

	return err
}

func decodeNonProtoField(d *json.Decoder, unmarshaler protojson.UnmarshalOptions, v interface{}) error {
	rv := reflect.ValueOf(v)

	if rv.Kind() != reflect.Ptr {
		return fmt.Errorf("%T is not a pointer", v)
	}

	for rv.Kind() == reflect.Ptr {
		if rv.IsNil() {
			rv.Set(reflect.New(rv.Type().Elem()))
		}
		if rv.Type().ConvertibleTo(typeProtoMessage) {
			var b json.RawMessage
			err := d.Decode(&b)
			if err != nil {
				return err
			}

			return unmarshaler.Unmarshal(b, rv.Interface().(proto.Message))
		}
		rv = rv.Elem()
	}

	if rv.Kind() == reflect.Map {
		if rv.IsNil() {
			rv.Set(reflect.MakeMap(rv.Type()))
		}
		conv, ok := convFromType[rv.Type().Key().Kind()]
		if !ok {
			return fmt.Errorf("unsupported type of map field key: %v", rv.Type().Key())
		}

		m := make(map[string]*json.RawMessage)
		if err := d.Decode(&m); err != nil {
			return err
		}
		for k, v := range m {
			result := conv.Call([]reflect.Value{reflect.ValueOf(k)})
			if err := result[1].Interface(); err != nil {
				return err.(error)
			}
			bk := result[0]
			bv := reflect.New(rv.Type().Elem())
			if v == nil {
				null := json.RawMessage("null")
				v = &null
			}
			if err := unmarshalJSONPb(*v, unmarshaler, bv.Interface()); err != nil {
				return err
			}
			rv.SetMapIndex(bk, bv.Elem())
		}
		return nil
	}

	if rv.Kind() == reflect.Slice {
		var sl []json.RawMessage
		if err := d.Decode(&sl); err != nil {
			return err
		}
		if sl != nil {
			rv.Set(reflect.MakeSlice(rv.Type(), 0, 0))
		}
		for _, item := range sl {
			bv := reflect.New(rv.Type().Elem())
			if err := unmarshalJSONPb(item, unmarshaler, bv.Interface()); err != nil {
				return err
			}
			rv.Set(reflect.Append(rv, bv.Elem()))
		}
		return nil
	}

	if _, ok := rv.Interface().(protoEnum); ok {
		var data interface{}
		if err := d.Decode(&data); err != nil {
			return err
		}
		switch v := data.(type) {
		case string:
			return fmt.Errorf("unmarshaling of symbolic enum %q not supported: %T", data, rv.Interface())
		case float64:
			rv.Set(reflect.ValueOf(int32(v)).Convert(rv.Type()))
			return nil
		default:
			return fmt.Errorf("cannot assign %#v into Go type %T", data, rv.Interface())
		}
	}

	return d.Decode(v)
}
