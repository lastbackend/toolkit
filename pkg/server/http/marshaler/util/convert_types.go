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

package util

import (
	"encoding/base64"
	"fmt"
	"strconv"
	"strings"

	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/types/known/durationpb"
	"google.golang.org/protobuf/types/known/timestamppb"
	"google.golang.org/protobuf/types/known/wrapperspb"
)

func String(val string) (string, error) {
	return val, nil
}

func StringSlice(val, sep string) ([]string, error) {
	return strings.Split(val, sep), nil
}

func Bool(val string) (bool, error) {
	return strconv.ParseBool(val)
}

func BoolSlice(val, sep string) ([]bool, error) {
	s := strings.Split(val, sep)
	values := make([]bool, len(s))
	for i, v := range s {
		value, err := Bool(v)
		if err != nil {
			return values, err
		}
		values[i] = value
	}
	return values, nil
}

func Float64(val string) (float64, error) {
	return strconv.ParseFloat(val, 64)
}

func Float64Slice(val, sep string) ([]float64, error) {
	s := strings.Split(val, sep)
	values := make([]float64, len(s))
	for i, v := range s {
		value, err := Float64(v)
		if err != nil {
			return values, err
		}
		values[i] = value
	}
	return values, nil
}

func Float32(val string) (float32, error) {
	f, err := strconv.ParseFloat(val, 32)
	if err != nil {
		return 0, err
	}
	return float32(f), nil
}

func Float32Slice(val, sep string) ([]float32, error) {
	s := strings.Split(val, sep)
	values := make([]float32, len(s))
	for i, v := range s {
		value, err := Float32(v)
		if err != nil {
			return values, err
		}
		values[i] = value
	}
	return values, nil
}

// Int64 converts the given string representation of an integer into int64.
func Int64(val string) (int64, error) {
	return strconv.ParseInt(val, 0, 64)
}

func Int64Slice(val, sep string) ([]int64, error) {
	s := strings.Split(val, sep)
	values := make([]int64, len(s))
	for i, v := range s {
		value, err := Int64(v)
		if err != nil {
			return values, err
		}
		values[i] = value
	}
	return values, nil
}

func Int32(val string) (int32, error) {
	i, err := strconv.ParseInt(val, 0, 32)
	if err != nil {
		return 0, err
	}
	return int32(i), nil
}

func Int32Slice(val, sep string) ([]int32, error) {
	s := strings.Split(val, sep)
	values := make([]int32, len(s))
	for i, v := range s {
		value, err := Int32(v)
		if err != nil {
			return values, err
		}
		values[i] = value
	}
	return values, nil
}

func Uint64(val string) (uint64, error) {
	return strconv.ParseUint(val, 0, 64)
}

// Uint64Slice converts 'val' where individual integers are separated by
// 'sep' into a uint64 slice.
func Uint64Slice(val, sep string) ([]uint64, error) {
	s := strings.Split(val, sep)
	values := make([]uint64, len(s))
	for i, v := range s {
		value, err := Uint64(v)
		if err != nil {
			return values, err
		}
		values[i] = value
	}
	return values, nil
}

// Uint32 converts the given string representation of an integer into uint32.
func Uint32(val string) (uint32, error) {
	i, err := strconv.ParseUint(val, 0, 32)
	if err != nil {
		return 0, err
	}
	return uint32(i), nil
}

func Uint32Slice(val, sep string) ([]uint32, error) {
	s := strings.Split(val, sep)
	values := make([]uint32, len(s))
	for i, v := range s {
		value, err := Uint32(v)
		if err != nil {
			return values, err
		}
		values[i] = value
	}
	return values, nil
}

func Bytes(val string) ([]byte, error) {
	b, err := base64.StdEncoding.DecodeString(val)
	if err != nil {
		b, err = base64.URLEncoding.DecodeString(val)
		if err != nil {
			return nil, err
		}
	}
	return b, nil
}

func BytesSlice(val, sep string) ([][]byte, error) {
	s := strings.Split(val, sep)
	values := make([][]byte, len(s))
	for i, v := range s {
		value, err := Bytes(v)
		if err != nil {
			return values, err
		}
		values[i] = value
	}
	return values, nil
}

func Timestamp(val string) (*timestamppb.Timestamp, error) {
	var r timestamppb.Timestamp
	unmarshaler := &protojson.UnmarshalOptions{}
	err := unmarshaler.Unmarshal([]byte(val), &r)
	if err != nil {
		return nil, err
	}
	return &r, nil
}

func Duration(val string) (*durationpb.Duration, error) {
	var r durationpb.Duration
	unmarshaler := &protojson.UnmarshalOptions{}
	err := unmarshaler.Unmarshal([]byte(val), &r)
	if err != nil {
		return nil, err
	}
	return &r, nil
}

func Enum(val string, enumValMap map[string]int32) (int32, error) {
	e, ok := enumValMap[val]
	if ok {
		return e, nil
	}

	i, err := Int32(val)
	if err != nil {
		return 0, fmt.Errorf("%s is not valid", val)
	}
	for _, v := range enumValMap {
		if v == i {
			return i, nil
		}
	}
	return 0, fmt.Errorf("%s is not valid", val)
}

func EnumSlice(val, sep string, enumValMap map[string]int32) ([]int32, error) {
	s := strings.Split(val, sep)
	values := make([]int32, len(s))
	for i, v := range s {
		value, err := Enum(v, enumValMap)
		if err != nil {
			return values, err
		}
		values[i] = value
	}
	return values, nil
}

// ===========================================================
// Support for google.protobuf.wrappers of primitive types
// ===========================================================

func StringValue(val string) (*wrapperspb.StringValue, error) {
	return &wrapperspb.StringValue{Value: val}, nil
}

func FloatValue(val string) (*wrapperspb.FloatValue, error) {
	parsedVal, err := Float32(val)
	return &wrapperspb.FloatValue{Value: parsedVal}, err
}

func DoubleValue(val string) (*wrapperspb.DoubleValue, error) {
	parsedVal, err := Float64(val)
	return &wrapperspb.DoubleValue{Value: parsedVal}, err
}

func BoolValue(val string) (*wrapperspb.BoolValue, error) {
	parsedVal, err := Bool(val)
	return &wrapperspb.BoolValue{Value: parsedVal}, err
}

func Int32Value(val string) (*wrapperspb.Int32Value, error) {
	parsedVal, err := Int32(val)
	return &wrapperspb.Int32Value{Value: parsedVal}, err
}

func UInt32Value(val string) (*wrapperspb.UInt32Value, error) {
	parsedVal, err := Uint32(val)
	return &wrapperspb.UInt32Value{Value: parsedVal}, err
}

func Int64Value(val string) (*wrapperspb.Int64Value, error) {
	parsedVal, err := Int64(val)
	return &wrapperspb.Int64Value{Value: parsedVal}, err
}

func UInt64Value(val string) (*wrapperspb.UInt64Value, error) {
	parsedVal, err := Uint64(val)
	return &wrapperspb.UInt64Value{Value: parsedVal}, err
}

func BytesValue(val string) (*wrapperspb.BytesValue, error) {
	parsedVal, err := Bytes(val)
	return &wrapperspb.BytesValue{Value: parsedVal}, err
}
