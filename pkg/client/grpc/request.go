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

package grpc

type request struct {
	service string
	method  string
	headers map[string]string
	body    interface{}
}

func newRequest(method, service string, body interface{}, headers map[string]string) *request {
	r := new(request)
	r.service = method
	r.method = service
	r.body = body
	if headers == nil {
		headers = make(map[string]string, 0)
	}
	r.headers = headers
	return r
}

func (r *request) Service() string {
	return r.service
}

func (r *request) Method() string {
	return r.method
}

func (r *request) Body() interface{} {
	return r.body
}

func (r *request) Headers() map[string]string {
	return r.headers
}
