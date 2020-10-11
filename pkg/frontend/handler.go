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

package frontend

import (
	"errors"
	"net/http"
	"strings"

	"git.rwth-aachen.de/acs/public/cloud/mas/clonemap/pkg/common/httpreply"
)

// handleAPI is the global handler for requests to path /api
func (fe *Frontend) handleAPI(w http.ResponseWriter, r *http.Request) {
	var cmapErr, httpErr error
	// ams.logInfo.Println("Received Request: ", r.Method, " ", r.URL.EscapedPath())
	// determine which ressource is requested and call corresponding handler
	respath := strings.Split(r.URL.EscapedPath(), "/")
	resvalid := false

	if len(respath) > 2 {
		switch respath[2] {
		case "pf":
			resvalid, cmapErr, httpErr = fe.handlePlatform(w, r, respath)
		case "ams":
			resvalid, cmapErr, httpErr = fe.handleAMS(w, r, respath)
		case "df":
			resvalid, cmapErr, httpErr = fe.handleDF(w, r, respath)
		case "logger":
			resvalid, cmapErr, httpErr = fe.handleLogging(w, r, respath)
		case "overview":
			resvalid = true
			cmapErr, httpErr = fe.handleOverview(w, r)
		default:
			cmapErr = errors.New("Resource not found")
		}
	}

	if !resvalid {
		httpErr = httpreply.NotFoundError(w)
		cmapErr = errors.New("Resource not found")
	}
	if cmapErr != nil {
		// ams.logError.Println(respath, cmapErr)
	}
	if httpErr != nil {
		// ams.logError.Println(respath, httpErr)
	}
}

// handleOverview is the handler for requests to path /api/overview
func (fe *Frontend) handleOverview(w http.ResponseWriter, r *http.Request) (cmapErr, httpErr error) {
	if r.Method == "GET" {
		cmapErr, httpErr = fe.handleMAS(w, r)
	} else {
		httpErr = httpreply.MethodNotAllowed(w)
		cmapErr = errors.New("Error: Method not allowed on path /api/overview")
	}
	return
}

// listen opens a http server listening and serving request
func (fe *Frontend) listen() (err error) {
	mux := http.NewServeMux()
	mux.HandleFunc("/api/", fe.handleAPI)
	mux.HandleFunc("/", http.FileServer(http.Dir("./web")).ServeHTTP)
	s := &http.Server{
		Addr:    ":13000",
		Handler: mux,
	}
	err = s.ListenAndServe()
	return
}
