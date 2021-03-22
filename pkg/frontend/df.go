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
	"encoding/json"
	"io/ioutil"
	"net/http"
	"strconv"

	"git.rwth-aachen.de/acs/public/cloud/mas/clonemap/pkg/common/httpreply"
	"git.rwth-aachen.de/acs/public/cloud/mas/clonemap/pkg/schemas"
	"github.com/gorilla/mux"
)

// handleGetSvcs is the handler of /api/df/{masid}/svc
func (fe *Frontend) handleGetSvcs(w http.ResponseWriter, r *http.Request) {
	var cmapErr, httpErr error
	var svcs []schemas.Service
	vars := mux.Vars(r)
	masID, cmapErr := strconv.Atoi(vars["masid"])
	if cmapErr != nil {
		httpErr = httpreply.NotFoundError(w)
		fe.logErrors(r.URL.Path, cmapErr, httpErr)
	}
	svcs, _, cmapErr = fe.dfClient.GetSvcs(masID)
	if cmapErr != nil {
		httpErr = httpreply.CMAPError(w, cmapErr.Error())
		fe.logErrors(r.URL.Path, cmapErr, httpErr)
		return
	}
	httpErr = httpreply.Resource(w, svcs, cmapErr)
	fe.logErrors(r.URL.Path, cmapErr, httpErr)
	return

}

// handleGetSvc is the handler of /api/df/{masid}/svc/desc/{desc}
func (fe *Frontend) handleGetSvc(w http.ResponseWriter, r *http.Request) {
	var cmapErr, httpErr error
	var svc []schemas.Service
	masID, desc, cmapErr := getDesc(r)
	if cmapErr != nil {
		httpErr = httpreply.NotFoundError(w)
		fe.logErrors(r.URL.Path, cmapErr, httpErr)
		return
	}
	svc, _, cmapErr = fe.dfClient.GetSvc(masID, desc)
	if cmapErr != nil {
		httpErr = httpreply.CMAPError(w, cmapErr.Error())
		fe.logErrors(r.URL.Path, cmapErr, httpErr)
		return
	}
	httpErr = httpreply.Resource(w, svc, cmapErr)
	fe.logErrors(r.URL.Path, cmapErr, httpErr)
	return
}

//handlePostSvc is the handler of /api/df/{masid}/svc
func (fe *Frontend) handlePostSvc(w http.ResponseWriter, r *http.Request) {
	var cmapErr, httpErr error
	vars := mux.Vars(r)
	masID, cmapErr := strconv.Atoi(vars["masid"])
	if cmapErr != nil {
		httpErr = httpreply.NotFoundError(w)
		fe.logErrors(r.URL.Path, cmapErr, httpErr)
		return
	}

	var body []byte
	body, cmapErr = ioutil.ReadAll(r.Body)
	if cmapErr != nil {
		httpErr = httpreply.InvalidBodyError(w)
		fe.logErrors(r.URL.Path, cmapErr, httpErr)
		return
	}

	var svc schemas.Service
	cmapErr = json.Unmarshal(body, &svc)
	if cmapErr != nil {
		httpErr = httpreply.JSONUnmarshalError(w)
		fe.logErrors(r.URL.Path, cmapErr, httpErr)
		return
	}

	_, _, cmapErr = fe.dfClient.PostSvc(masID, svc)
	if cmapErr != nil {
		httpErr = httpreply.CMAPError(w, cmapErr.Error())
		fe.logErrors(r.URL.Path, cmapErr, httpErr)
		return
	}
	httpErr = httpreply.Created(w, cmapErr, "text/plain", []byte("Resource Created"))
	fe.logErrors(r.URL.Path, cmapErr, httpErr)
	return
}

// handleSvcWithDist is the handler of /api/df/{masid}/svc/desc/{desc}/node/{nodeid}/dist/{dist}
func (fe *Frontend) handleSvcWithDist(w http.ResponseWriter, r *http.Request) {
	var cmapErr, httpErr error
	var svc []schemas.Service
	masID, desc, nodeID, dist, cmapErr := getDist(r)

	if cmapErr != nil {
		httpErr = httpreply.CMAPError(w, cmapErr.Error())
		fe.logErrors(r.URL.Path, cmapErr, httpErr)
		return
	}

	svc, _, cmapErr = fe.dfClient.GetLocalSvc(masID, desc, nodeID, dist)

	if cmapErr != nil {
		httpErr = httpreply.CMAPError(w, cmapErr.Error())
		fe.logErrors(r.URL.Path, cmapErr, httpErr)
		return
	}
	httpErr = httpreply.Resource(w, svc, cmapErr)
	fe.logErrors(r.URL.Path, cmapErr, httpErr)
	return
}
