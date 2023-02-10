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

package converter

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"io"
	"io/ioutil"
	"net"
	"net/http"
	"net/textproto"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/gorilla/mux"
	"github.com/pkg/errors"
	"google.golang.org/genproto/protobuf/field_mask"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/reflect/protoregistry"
	"google.golang.org/protobuf/types/known/durationpb"
	"google.golang.org/protobuf/types/known/timestamppb"
	"google.golang.org/protobuf/types/known/wrapperspb"
)

var (
	valuesKeyRegexp = regexp.MustCompile(`^(.*)\[(.*)\]$`)
)

const (
	MetadataPrefix             = "router-"
	MetadataHeaderPrefix       = "Router-Metadata-"
	metadataHeaderBinarySuffix = "-Bin"
	xHeaderPrefix              = "X-"
	xForwardedFor              = "X-Forwarded-For"
	xForwardedHostHeader       = "X-Forwarded-Host"
)

func NewReader(r io.Reader) (io.Reader, error) {
	b, err := io.ReadAll(r)
	if err != nil {
		return nil, err
	}
	return bytes.NewReader(b), nil
}

func PrepareHeaderFromRequest(req *http.Request) (map[string]string, error) {
	var headers = make(map[string]string, 0)

	for k, v := range req.Header {
		k = textproto.CanonicalMIMEHeaderKey(k)
		for _, val := range v {
			if h, ok := headerMatcher(k); ok {
				if strings.HasSuffix(k, metadataHeaderBinarySuffix) {
					b, err := decodeBinHeader(val)
					if err != nil {
						return nil, status.Errorf(codes.InvalidArgument, "invalid binary header %s: %s", k, err)
					}
					val = string(b)
				}
				headers[h] = val
			}
		}
	}

	if host := req.Header.Get(xForwardedHostHeader); host != "" {
		headers[strings.ToLower(xForwardedHostHeader)] = host
	} else if req.Host != "" {
		headers[strings.ToLower(xForwardedHostHeader)] = req.Host
	}

	if addr := req.RemoteAddr; addr != "" {
		if remoteIP, _, err := net.SplitHostPort(addr); err == nil {
			if forward := req.Header.Get(xForwardedFor); forward == "" {
				headers[strings.ToLower(xForwardedFor)] = remoteIP
			} else {
				headers[strings.ToLower(xForwardedFor)] = fmt.Sprintf("%s, %s", forward, remoteIP)
			}
		}
	}

	return headers, nil
}

func SetRawBodyToProto(r *http.Request, message proto.Message, param string) error {
	raw, err := ioutil.ReadAll(r.Body)
	if err != nil {
		return err
	}
	paramPath := strings.Split(param, ".")
	return matchFieldValue(message.ProtoReflect(), paramPath, []string{base64.StdEncoding.EncodeToString(raw)})
}

func ParseRequestUrlParametersToProto(r *http.Request, message proto.Message, param string) error {
	paramPath := strings.Split(param, ".")
	return matchFieldValue(message.ProtoReflect(), paramPath, []string{mux.Vars(r)[param]})
}

func ParseRequestQueryParametersToProto(message proto.Message, values url.Values) error {
	for k, v := range values {
		match := valuesKeyRegexp.FindStringSubmatch(k)
		if len(match) == 3 {
			k = match[1]
			v = append([]string{match[2]}, v...)
		}
		paramPath := strings.Split(k, ".")
		if err := matchFieldValue(message.ProtoReflect(), paramPath, v); err != nil {
			return err
		}
	}
	return nil
}

func headerMatcher(key string) (string, bool) {
	key = textproto.CanonicalMIMEHeaderKey(key)
	if strings.HasPrefix(key, xHeaderPrefix) {
		return key, true
	} else if strings.HasPrefix(key, MetadataHeaderPrefix) {
		return key[len(MetadataHeaderPrefix):], true
	} else {
		return MetadataPrefix + key, true
	}
}

func decodeBinHeader(v string) ([]byte, error) {
	if len(v)%4 == 0 {
		return base64.StdEncoding.DecodeString(v)
	}
	return base64.RawStdEncoding.DecodeString(v)
}

func matchFieldValue(msgValue protoreflect.Message, paramPath []string, values []string) error {
	if len(paramPath) < 1 {
		return errors.New("no param path")
	}
	if len(values) < 1 {
		return errors.New("no value provided")
	}

	var fieldDescriptor protoreflect.FieldDescriptor
	for index, paramName := range paramPath {
		fields := msgValue.Descriptor().Fields()
		fieldDescriptor = fields.ByName(protoreflect.Name(paramName))
		if fieldDescriptor == nil {
			fieldDescriptor = fields.ByJSONName(paramName)
			if fieldDescriptor == nil {
				fmt.Println(fmt.Sprintf("field not found in %q: %q", msgValue.Descriptor().FullName(), strings.Join(paramPath, ".")))
				return nil
			}
		}
		if index == len(paramPath)-1 {
			break
		}
		if fieldDescriptor.Message() == nil || fieldDescriptor.Cardinality() == protoreflect.Repeated {
			return fmt.Errorf("invalid path: %q is not a message", paramName)
		}
		msgValue = msgValue.Mutable(fieldDescriptor).Message()
	}
	if of := fieldDescriptor.ContainingOneof(); of != nil {
		if f := msgValue.WhichOneof(of); f != nil {
			return fmt.Errorf("field already set for oneof %q", of.FullName().Name())
		}
	}
	switch {
	case fieldDescriptor.IsList():
		return matchListField(fieldDescriptor, msgValue.Mutable(fieldDescriptor).List(), values)
	case fieldDescriptor.IsMap():
		return matchMapField(fieldDescriptor, msgValue.Mutable(fieldDescriptor).Map(), values)
	}
	if len(values) > 1 {
		return fmt.Errorf("too many values for field %q: %s", fieldDescriptor.FullName().Name(), strings.Join(values, ", "))
	}
	return matchField(fieldDescriptor, msgValue, values[0])
}

func matchField(fieldDescriptor protoreflect.FieldDescriptor, message protoreflect.Message, value string) error {
	v, err := parseField(fieldDescriptor, value)
	if err != nil {
		return fmt.Errorf("parsing field %q: %w", fieldDescriptor.FullName().Name(), err)
	}
	message.Set(fieldDescriptor, v)
	return nil
}

func matchListField(fieldDescriptor protoreflect.FieldDescriptor, list protoreflect.List, values []string) error {
	for _, v := range values {
		v, err := parseField(fieldDescriptor, v)
		if err != nil {
			return fmt.Errorf("parsing list %q: %w", fieldDescriptor.FullName().Name(), err)
		}
		list.Append(v)
	}
	return nil
}

func matchMapField(fieldDescriptor protoreflect.FieldDescriptor, items protoreflect.Map, values []string) error {
	if len(values) != 2 {
		return fmt.Errorf("more than one value provided for key %q in map %q", values[0], fieldDescriptor.FullName())
	}
	key, err := parseField(fieldDescriptor.MapKey(), values[0])
	if err != nil {
		return fmt.Errorf("parsing map key %q: %w", fieldDescriptor.FullName().Name(), err)
	}
	value, err := parseField(fieldDescriptor.MapValue(), values[1])
	if err != nil {
		return fmt.Errorf("parsing map value %q: %w", fieldDescriptor.FullName().Name(), err)
	}
	items.Set(key.MapKey(), value)
	return nil
}

func parseField(fieldDescriptor protoreflect.FieldDescriptor, value string) (protoreflect.Value, error) {
	switch fieldDescriptor.Kind() {
	case protoreflect.StringKind:
		return protoreflect.ValueOfString(value), nil
	case protoreflect.BytesKind:
		v, err := base64.StdEncoding.DecodeString(value)
		if err != nil {
			return protoreflect.Value{}, err
		}
		return protoreflect.ValueOfBytes(v), nil
	case protoreflect.BoolKind:
		v, err := strconv.ParseBool(value)
		if err != nil {
			return protoreflect.Value{}, err
		}
		return protoreflect.ValueOfBool(v), nil
	case protoreflect.Int32Kind, protoreflect.Sint32Kind, protoreflect.Sfixed32Kind:
		v, err := strconv.ParseInt(value, 10, 32)
		if err != nil {
			return protoreflect.Value{}, err
		}
		return protoreflect.ValueOfInt32(int32(v)), nil
	case protoreflect.Int64Kind, protoreflect.Sint64Kind, protoreflect.Sfixed64Kind:
		v, err := strconv.ParseInt(value, 10, 64)
		if err != nil {
			return protoreflect.Value{}, err
		}
		return protoreflect.ValueOfInt64(v), nil
	case protoreflect.Uint32Kind, protoreflect.Fixed32Kind:
		v, err := strconv.ParseUint(value, 10, 32)
		if err != nil {
			return protoreflect.Value{}, err
		}
		return protoreflect.ValueOfUint32(uint32(v)), nil
	case protoreflect.Uint64Kind, protoreflect.Fixed64Kind:
		v, err := strconv.ParseUint(value, 10, 64)
		if err != nil {
			return protoreflect.Value{}, err
		}
		return protoreflect.ValueOfUint64(v), nil
	case protoreflect.FloatKind:
		v, err := strconv.ParseFloat(value, 32)
		if err != nil {
			return protoreflect.Value{}, err
		}
		return protoreflect.ValueOfFloat32(float32(v)), nil
	case protoreflect.DoubleKind:
		v, err := strconv.ParseFloat(value, 64)
		if err != nil {
			return protoreflect.Value{}, err
		}
		return protoreflect.ValueOfFloat64(v), nil
	case protoreflect.MessageKind, protoreflect.GroupKind:
		return parseMessage(fieldDescriptor.Message(), value)
	case protoreflect.EnumKind:
		enum, err := protoregistry.GlobalTypes.FindEnumByName(fieldDescriptor.Enum().FullName())
		switch {
		case err == protoregistry.NotFound:
			return protoreflect.Value{}, fmt.Errorf("enum %q is not registered", fieldDescriptor.Enum().FullName())
		case err != nil:
			return protoreflect.Value{}, fmt.Errorf("failed to look up enum: %w", err)
		}
		v := enum.Descriptor().Values().ByName(protoreflect.Name(value))
		if v == nil {
			i, err := strconv.Atoi(value)
			if err != nil {
				return protoreflect.Value{}, fmt.Errorf("%q is not a valid value", value)
			}
			v = enum.Descriptor().Values().ByNumber(protoreflect.EnumNumber(i))
			if v == nil {
				return protoreflect.Value{}, fmt.Errorf("%q is not a valid value", value)
			}
		}
		return protoreflect.ValueOfEnum(v.Number()), nil
	default:
		panic(fmt.Sprintf("unknown field kind: %v", fieldDescriptor.Kind()))
	}
}

func parseMessage(msgDescriptor protoreflect.MessageDescriptor, value string) (protoreflect.Value, error) {
	var protoMessage proto.Message
	switch msgDescriptor.FullName() {
	case "google.protobuf.BoolValue":
		v, err := strconv.ParseBool(value)
		if err != nil {
			return protoreflect.Value{}, err
		}
		protoMessage = &wrapperspb.BoolValue{Value: v}
	case "google.protobuf.StringValue":
		protoMessage = &wrapperspb.StringValue{Value: value}
	case "google.protobuf.FloatValue":
		v, err := strconv.ParseFloat(value, 32)
		if err != nil {
			return protoreflect.Value{}, err
		}
		protoMessage = &wrapperspb.FloatValue{Value: float32(v)}
	case "google.protobuf.Int64Value":
		v, err := strconv.ParseInt(value, 10, 64)
		if err != nil {
			return protoreflect.Value{}, err
		}
		protoMessage = &wrapperspb.Int64Value{Value: v}
	case "google.protobuf.Int32Value":
		v, err := strconv.ParseInt(value, 10, 32)
		if err != nil {
			return protoreflect.Value{}, err
		}
		protoMessage = &wrapperspb.Int32Value{Value: int32(v)}
	case "google.protobuf.UInt64Value":
		v, err := strconv.ParseUint(value, 10, 64)
		if err != nil {
			return protoreflect.Value{}, err
		}
		protoMessage = &wrapperspb.UInt64Value{Value: v}
	case "google.protobuf.UInt32Value":
		v, err := strconv.ParseUint(value, 10, 32)
		if err != nil {
			return protoreflect.Value{}, err
		}
		protoMessage = &wrapperspb.UInt32Value{Value: uint32(v)}
	case "google.protobuf.BytesValue":
		v, err := base64.StdEncoding.DecodeString(value)
		if err != nil {
			return protoreflect.Value{}, err
		}
		protoMessage = &wrapperspb.BytesValue{Value: v}
	case "google.protobuf.DoubleValue":
		v, err := strconv.ParseFloat(value, 64)
		if err != nil {
			return protoreflect.Value{}, err
		}
		protoMessage = &wrapperspb.DoubleValue{Value: v}
	case "google.protobuf.FieldMask":
		fm := &field_mask.FieldMask{}
		fm.Paths = append(fm.Paths, strings.Split(value, ",")...)
		protoMessage = fm
	case "google.protobuf.Duration":
		if value == "null" {
			break
		}
		d, err := time.ParseDuration(value)
		if err != nil {
			return protoreflect.Value{}, err
		}
		protoMessage = durationpb.New(d)
	case "google.protobuf.Timestamp":
		if value == "null" {
			break
		}
		t, err := time.Parse(time.RFC3339Nano, value)
		if err != nil {
			return protoreflect.Value{}, err
		}
		protoMessage = timestamppb.New(t)
	default:
		return protoreflect.Value{}, fmt.Errorf("unsupported message type: %q", string(msgDescriptor.FullName()))
	}

	return protoreflect.ValueOfMessage(protoMessage.ProtoReflect()), nil
}
