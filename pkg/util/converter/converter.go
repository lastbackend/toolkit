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
	"encoding/base64"
	"fmt"
	"math"
	"reflect"
	"strconv"
	"strings"
	"time"
)

func StringToInt64(s string) int64 {
	i, _ := strconv.ParseInt(s, 10, 64)
	return i
}

func StringToInt(s string) int {
	i, _ := strconv.Atoi(s)
	return i
}

func StringToFloat(s string) float64 {
	i, _ := strconv.ParseFloat(s, 64)
	return i
}

func IntToString(i int) string {
	return strconv.Itoa(i)
}

func StringToBool(s string) bool {
	s = strings.ToLower(s)
	if s == "true" || s == "1" || s == "t" {
		return true
	}
	return false
}

func StringToUint(s string) uint {
	i, _ := strconv.ParseUint(s, 10, 64)
	return uint(i)
}

func BoolToString(str string) (bool, error) {
	switch str {
	case "":
		return false, nil
	case "1", "t", "T", "true", "TRUE", "True":
		return true, nil
	case "0", "f", "F", "false", "FALSE", "False":
		return false, nil
	}
	return false, fmt.Errorf("parse bool string: %s", str)
}

func Int64ToInt(i int64) int {
	return StringToInt(strconv.FormatInt(i, 10))
}

func DecodeBase64(s string) string {
	buf, _ := base64.StdEncoding.DecodeString(s)
	return string(buf)
}

func EncodeBase64(s string) string {
	return base64.StdEncoding.EncodeToString([]byte(s))
}

func HumanizeTimeDuration(duration time.Duration) string {
	hours := int64(math.Mod(duration.Hours(), 24))
	minutes := int64(math.Mod(duration.Minutes(), 60))
	seconds := int64(math.Mod(duration.Seconds(), 60))

	chunks := []struct {
		amount int64
	}{
		{hours},
		{minutes},
		{seconds},
	}

	parts := make([]string, 0)

	for _, chunk := range chunks {
		switch chunk.amount {
		case 0:
			parts = append(parts, "00")
			continue
		default:
			if chunk.amount >= 10 {
				parts = append(parts, fmt.Sprintf("%d", chunk.amount))
			} else {
				parts = append(parts, fmt.Sprintf("0%d", chunk.amount))
			}
		}
	}

	return strings.Join(parts, ":")
}

func EnforcePtr(obj interface{}) (reflect.Value, error) {
	v := reflect.ValueOf(obj)
	if v.Kind() != reflect.Ptr {
		if v.Kind() == reflect.Invalid {
			return reflect.Value{}, fmt.Errorf("expected pointer, but got invalid kind")
		}
		return reflect.Value{}, fmt.Errorf("expected pointer, but got %v type", v.Type())
	}
	if v.IsNil() {
		return reflect.Value{}, fmt.Errorf("expected pointer, but got nil")
	}
	return v.Elem(), nil
}

func ToBoolPointer(v bool) *bool {
	return &v
}

func ToStringPointer(v string) *string {
	return &v
}

func ToIntPointer(v int) *int {
	return &v
}

func ToInt64Pointer(v int64) *int64 {
	return &v
}

func ToDurationPointer(v time.Duration) *time.Duration {
	return &v
}
