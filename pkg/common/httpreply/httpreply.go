/*
Copyright 2020 Institute for Automation of Complex Power Systems,
E.ON Energy Research Center, RWTH Aachen University

This project is licensed under either of
- Apache License, Version 2.0
- MIT License
at your option.

Apache License, Version 2.0:

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.

MIT License:

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in
all copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
THE SOFTWARE.
*/

// Package httpreply implements common replies for http requests
package httpreply

import (
	"encoding/json"
	"net/http"
)

// MethodNotAllowed writes standard response if requested method is not allowed
func MethodNotAllowed(w http.ResponseWriter) (err error) {
	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(http.StatusMethodNotAllowed)
	_, err = w.Write([]byte("Method Not Allowed"))
	return
}

// Created writes standard response for ressource creation
func Created(w http.ResponseWriter, cmaperr error, content string, answer []byte) (err error) {
	if cmaperr == nil {
		w.Header().Set("Content-Type", content)
		w.WriteHeader(http.StatusCreated)
		_, err = w.Write(answer)
	} else {
		err = CMAPError(w, cmaperr.Error())
	}
	return
}

// Deleted writes standard response for ressource deleteion
func Deleted(w http.ResponseWriter, cmaperr error) (err error) {
	if cmaperr == nil {
		w.Header().Set("Content-Type", "text/plain")
		w.WriteHeader(http.StatusOK)
		_, err = w.Write([]byte("Ressource deleted"))
	} else {
		err = CMAPError(w, cmaperr.Error())
	}
	return
}

// Updated writes standard response for ressource update
func Updated(w http.ResponseWriter, cmaperr error) (err error) {
	if cmaperr == nil {
		w.Header().Set("Content-Type", "text/plain")
		w.WriteHeader(http.StatusOK)
		_, err = w.Write([]byte("Ressource updated"))
	} else {
		err = CMAPError(w, cmaperr.Error())
	}
	return
}

// Resource writes standard response for ressource get
func Resource(w http.ResponseWriter, v interface{}, cmaperr error) (err error) {
	if cmaperr == nil {
		var res []byte
		res, err = json.Marshal(v)
		if err == nil {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			_, err = w.Write(res)
		} else {
			err = JSONMarshalError(w)
		}
	} else {
		err = CMAPError(w, cmaperr.Error())
	}
	return
}

// JSONMarshalError writes standard response for JSON Marshal Error
func JSONMarshalError(w http.ResponseWriter) (err error) {
	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(http.StatusInternalServerError)
	_, err = w.Write([]byte("JSON Marshal Error"))
	return
}

// JSONUnmarshalError writes standard response for JSON Unmarshal Error
func JSONUnmarshalError(w http.ResponseWriter) (err error) {
	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(http.StatusBadRequest)
	_, err = w.Write([]byte("JSON Unmarshal Error"))
	return
}

// InvalidBodyError writes standard response for Invalid Body Error
func InvalidBodyError(w http.ResponseWriter) (err error) {
	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(http.StatusBadRequest)
	_, err = w.Write([]byte("Invalid Request Body"))
	return
}

// CMAPError writes standard response for cloneMAP Error
func CMAPError(w http.ResponseWriter, description string) (err error) {
	w.Header().Set("Content-Type", "text/plain")
	switch description {
	case "StorageError":
		w.WriteHeader(http.StatusInternalServerError)
		_, err = w.Write([]byte("CMAP Storage Error"))
	case "NotFoundError":
		err = NotFoundError(w)
	case "ClusterError":
		w.WriteHeader(http.StatusInternalServerError)
		_, err = w.Write([]byte("CMAP Cluster Error"))
	case "NotAllowedError":
		err = MethodNotAllowed(w)
	default:
		w.WriteHeader(http.StatusInternalServerError)
		_, err = w.Write([]byte("CMAP" + description))
	}
	return
}

// NotFoundError writes standard response for Ressource not found Error
func NotFoundError(w http.ResponseWriter) (err error) {
	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(http.StatusNotFound)
	_, err = w.Write([]byte("Ressource Not Found"))
	return
}
