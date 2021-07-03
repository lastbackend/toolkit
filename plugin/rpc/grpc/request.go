/*
Copyright [2014] - [2021] The Last.Backend authors.

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

import (
	"fmt"
	"strings"
)

func methodToGRPC(service, method string) string {
	if len(method) == 0 || method[0] == '/' {
		return method
	}
	mParts := strings.Split(method, ".")
	if len(mParts) != 2 {
		return method
	}
	if len(service) == 0 {
		return fmt.Sprintf("/%s/%s", mParts[0], mParts[1])
	}
	return fmt.Sprintf("/%s.%s/%s", service, mParts[0], mParts[1])
}
