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

package marshaler

import (
	"io"
)

type Marshaler interface {
	Marshal(v interface{}) ([]byte, error)
	Unmarshal(data []byte, v interface{}) error
	NewDecoder(r io.Reader) Decoder
	NewEncoder(w io.Writer) Encoder
	ContentType() string
}

type Decoder interface {
	Decode(v interface{}) error
}

type Encoder interface {
	Encode(v interface{}) error
}

type DecoderFunc func(v interface{}) error

func (f DecoderFunc) Decode(v interface{}) error { return f(v) }

type EncoderFunc func(v interface{}) error

func (f EncoderFunc) Encode(v interface{}) error { return f(v) }

type Delimited interface {
	Delimiter() []byte
}
