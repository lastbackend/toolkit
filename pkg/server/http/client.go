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
	"errors"
	"net/http"
	"strconv"
)

func HandleGRPCResponse(w http.ResponseWriter, r *http.Request, headers map[string]string) (bool, error) {

	var statusCode, ok = headers["x-http-status-code"]
	if !ok {
		w.WriteHeader(http.StatusOK)
		return true, nil
	}

	i, err := strconv.Atoi(statusCode)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return false, err
	}

	if i < 200 || i > 505 {
		w.WriteHeader(http.StatusInternalServerError)
		return false, errors.New("invalid status code")
	}

	if i == 302 {
		if redirectUrl, ok := headers["x-http-redirect-uri"]; ok {
			http.Redirect(w, r, redirectUrl, http.StatusFound)
			return false, nil
		}
	}

	w.WriteHeader(i)
	return true, nil
}
