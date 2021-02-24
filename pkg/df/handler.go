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

package df

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"net/http"
	"strconv"

	"git.rwth-aachen.de/acs/public/cloud/mas/clonemap/pkg/common/httpreply"
	"git.rwth-aachen.de/acs/public/cloud/mas/clonemap/pkg/schemas"
	"github.com/gorilla/mux"
)

// handleAlive is the handler for requests to path /api/alive
func (df *DF) handleAlive(w http.ResponseWriter, r *http.Request) {
	var httpErr error
	httpErr = httpreply.Alive(w, nil)
	df.logErrors(r.URL.Path, nil, httpErr)
	return
}

// handleGetMASService is the handler for get requests to path /api/df/{masid}/svc
func (df *DF) handleGetMASService(w http.ResponseWriter, r *http.Request) {
	var cmapErr, httpErr error
	defer df.logErrors(r.URL.Path, cmapErr, httpErr)
	vars := mux.Vars(r)
	masID, cmapErr := strconv.Atoi(vars["masid"])
	if cmapErr != nil {
		httpErr = httpreply.NotFoundError(w)
		return
	}
	var svc []schemas.Service
	svc, cmapErr = df.stor.searchServices(masID, "")
	if cmapErr != nil {
		httpErr = httpreply.CMAPError(w, cmapErr.Error())
		return
	}
	httpErr = httpreply.Resource(w, svc, cmapErr)
	return
}

// handlePostMASService is the handler for post requests to path /api/df/{masid}/svc
func (df *DF) handlePostMASService(w http.ResponseWriter, r *http.Request) {
	var cmapErr, httpErr error
	defer df.logErrors(r.URL.Path, cmapErr, httpErr)
	var body []byte
	body, cmapErr = ioutil.ReadAll(r.Body)
	if cmapErr != nil {
		httpErr = httpreply.InvalidBodyError(w)
		return
	}
	var svc schemas.Service
	cmapErr = json.Unmarshal(body, &svc)
	if cmapErr != nil {
		httpErr = httpreply.JSONUnmarshalError(w)
		return
	}
	var id string
	id, cmapErr = df.stor.registerService(svc)
	svc.GUID = id
	var res []byte
	res, cmapErr = json.Marshal(svc)
	if cmapErr != nil {
		httpErr = httpreply.CMAPError(w, cmapErr.Error())
		return
	}
	httpErr = httpreply.Created(w, cmapErr, "application/json", res)
	return
}

// handleGetMASGraph is the handler for get requests to path /api/df/{masid}/graph
func (df *DF) handleGetMASGraph(w http.ResponseWriter, r *http.Request) {
	var cmapErr, httpErr error
	defer df.logErrors(r.URL.Path, cmapErr, httpErr)
	vars := mux.Vars(r)
	masID, cmapErr := strconv.Atoi(vars["masid"])
	if cmapErr != nil {
		httpErr = httpreply.NotFoundError(w)
		return
	}
	var gr schemas.Graph
	gr, cmapErr = df.stor.getGraph(masID)
	if cmapErr != nil {
		httpErr = httpreply.CMAPError(w, cmapErr.Error())
		return
	}
	httpErr = httpreply.Resource(w, gr, cmapErr)
	return
}

// handlePostMASGraph is the handler for post and put requests to path /api/df/{masid}/graph
func (df *DF) handlePostMASGraph(w http.ResponseWriter, r *http.Request) {
	var cmapErr, httpErr error
	defer df.logErrors(r.URL.Path, cmapErr, httpErr)
	vars := mux.Vars(r)
	masID, cmapErr := strconv.Atoi(vars["masid"])
	if cmapErr != nil {
		httpErr = httpreply.NotFoundError(w)
		return
	}
	var body []byte
	body, cmapErr = ioutil.ReadAll(r.Body)
	if cmapErr != nil {
		httpErr = httpreply.InvalidBodyError(w)
		return
	}
	var gr schemas.Graph
	cmapErr = json.Unmarshal(body, &gr)
	if cmapErr != nil {
		httpErr = httpreply.JSONUnmarshalError(w)
		return
	}
	cmapErr = df.stor.updateGraph(masID, gr)
	if cmapErr != nil {
		httpErr = httpreply.CMAPError(w, cmapErr.Error())
		return
	}
	httpErr = httpreply.Created(w, cmapErr, "text/plain", []byte("Ressource Created"))
	return
}

// handleGetSvcDesc is the handler for get requests to path /api/df/{masid}/svc/desc/{desc}
func (df *DF) handleGetSvcDesc(w http.ResponseWriter, r *http.Request) {
	var cmapErr, httpErr error
	defer df.logErrors(r.URL.Path, cmapErr, httpErr)
	vars := mux.Vars(r)
	masID, cmapErr := strconv.Atoi(vars["masid"])
	if cmapErr != nil {
		httpErr = httpreply.NotFoundError(w)
		return
	}
	desc := vars["desc"]
	var svc []schemas.Service
	svc, cmapErr = df.stor.searchServices(masID, desc)
	if cmapErr != nil {
		httpErr = httpreply.CMAPError(w, cmapErr.Error())
		return
	}
	httpErr = httpreply.Resource(w, svc, cmapErr)
	return
}

// handleGetSvcNode is the handler for get requests to path
// /api/df/{masid}/svc/desc/{desc}/node/{nodeid}/dist/{dist}
func (df *DF) handleGetSvcNodeDist(w http.ResponseWriter, r *http.Request) {
	var cmapErr, httpErr error
	defer df.logErrors(r.URL.Path, cmapErr, httpErr)
	vars := mux.Vars(r)
	masID, cmapErr := strconv.Atoi(vars["masid"])
	if cmapErr != nil {
		httpErr = httpreply.NotFoundError(w)
		return
	}
	desc := vars["desc"]
	nodeID, cmapErr := strconv.Atoi(vars["nodeid"])
	if cmapErr != nil {
		httpErr = httpreply.NotFoundError(w)
		return
	}
	dist, cmapErr := strconv.ParseFloat(vars["dist"], 64)
	if cmapErr != nil {
		httpErr = httpreply.NotFoundError(w)
		return
	}
	var svc []schemas.Service
	svc, cmapErr = df.stor.searchLocalServices(masID, nodeID, dist, desc)
	if cmapErr != nil {
		httpErr = httpreply.CMAPError(w, cmapErr.Error())
		return
	}
	httpErr = httpreply.Resource(w, svc, cmapErr)
	return
}

// handleGetSvcID is the handler for get requests to path /api/df/{masid}/svc/id/{svcid}
func (df *DF) handleGetSvcID(w http.ResponseWriter, r *http.Request) {
	var cmapErr, httpErr error
	defer df.logErrors(r.URL.Path, cmapErr, httpErr)
	vars := mux.Vars(r)
	masID, cmapErr := strconv.Atoi(vars["masid"])
	if cmapErr != nil {
		httpErr = httpreply.NotFoundError(w)
		return
	}
	svcID := vars["svcid"]
	var svc schemas.Service
	svc, cmapErr = df.stor.getService(masID, svcID)
	if cmapErr != nil {
		httpErr = httpreply.CMAPError(w, cmapErr.Error())
		return
	}
	httpErr = httpreply.Resource(w, svc, cmapErr)
	return
}

// handleDeleteSvcID is the handler for delete requests to path /api/df/{masid}/svc/id/{svcid}
func (df *DF) handleDeleteSvcID(w http.ResponseWriter, r *http.Request) {
	var cmapErr, httpErr error
	defer df.logErrors(r.URL.Path, cmapErr, httpErr)
	vars := mux.Vars(r)
	masID, cmapErr := strconv.Atoi(vars["masid"])
	if cmapErr != nil {
		httpErr = httpreply.NotFoundError(w)
		return
	}
	svcID := vars["svcid"]
	cmapErr = df.stor.deregisterService(masID, svcID)
	if cmapErr != nil {
		httpErr = httpreply.CMAPError(w, cmapErr.Error())
		return
	}
	httpErr = httpreply.Deleted(w, cmapErr)
	return
}

// methodNotAllowed is the default handler for valid paths but invalid methods
func (df *DF) methodNotAllowed(w http.ResponseWriter, r *http.Request) {
	httpErr := httpreply.MethodNotAllowed(w)
	cmapErr := errors.New("Error: Method not allowed on path " + r.URL.Path)
	df.logErrors(r.URL.Path, cmapErr, httpErr)
	return
}

// resourceNotFound is the default handler for invalid paths
func (df *DF) resourceNotFound(w http.ResponseWriter, r *http.Request) {
	httpErr := httpreply.NotFoundError(w)
	cmapErr := errors.New("Resource not found")
	df.logErrors(r.URL.Path, cmapErr, httpErr)
	return
}

// logErrors logs errors if any
func (df *DF) logErrors(path string, cmapErr error, httpErr error) {
	if cmapErr != nil {
		df.logError.Println(path, cmapErr)
	}
	if httpErr != nil {
		df.logError.Println(path, httpErr)
	}
	return
}

// loggingMiddleware logs request before calling final handler
func (df *DF) loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		df.logInfo.Println("Received Request: ", r.Method, " ", r.URL.EscapedPath())
		next.ServeHTTP(w, r)
	})
}

// server creates the df server
func (df *DF) server(port int) (serv *http.Server) {
	r := mux.NewRouter()
	// r.HandleFunc("/api/", logger.handleAPI)
	s := r.PathPrefix("/api").Subrouter()
	s.Path("/alive").Methods("GET").HandlerFunc(df.handleAlive)
	s.Path("/alive").Methods("POST", "PUT", "DELETE").HandlerFunc(df.methodNotAllowed)
	s.Path("/df/{masid}/svc").Methods("GET").HandlerFunc(df.handleGetMASService)
	s.Path("/df/{masid}/svc").Methods("POST").HandlerFunc(df.handlePostMASService)
	s.Path("/df/{masid}/svc").Methods("PUT", "DELETE").HandlerFunc(df.methodNotAllowed)
	s.Path("/df/{masid}/graph").Methods("GET").HandlerFunc(df.handleGetMASGraph)
	s.Path("/df/{masid}/graph").Methods("POST", "PUT").HandlerFunc(df.handlePostMASGraph)
	s.Path("/df/{masid}/graph").Methods("DELETE").HandlerFunc(df.methodNotAllowed)
	s.Path("/df/{masid}/svc/desc/{desc}").Methods("GET").HandlerFunc(df.handleGetSvcDesc)
	s.Path("/df/{masid}/svc/desc/{desc}").Methods("POST", "PUT", "DELETE").
		HandlerFunc(df.methodNotAllowed)
	s.Path("/df/{masid}/svc/desc/{desc}/node/{nodeid}/dist/{dist}").Methods("GET").
		HandlerFunc(df.handleGetSvcNodeDist)
	s.Path("/df/{masid}/svc/desc/{desc}/node/{nodeid}/dist/{dist}").
		Methods("POST", "PUT", "DELETE").HandlerFunc(df.methodNotAllowed)
	s.Path("/df/{masid}/svc/id/{svcid}").Methods("GET").HandlerFunc(df.handleGetSvcID)
	s.Path("/df/{masid}/svc/id/{svcid}").Methods("DELETE").HandlerFunc(df.handleDeleteSvcID)
	s.Path("/df/{masid}/svc/id/{svcid}").Methods("POST", "PUT").HandlerFunc(df.methodNotAllowed)
	s.PathPrefix("").HandlerFunc(df.resourceNotFound)
	s.Use(df.loggingMiddleware)
	serv = &http.Server{
		Addr:    ":" + strconv.Itoa(port),
		Handler: r,
	}
	return
}

// listen opens a http server listening and serving request
func (df *DF) listen(serv *http.Server) (err error) {
	df.logInfo.Println("DF listening on " + serv.Addr)
	err = serv.ListenAndServe()
	return
}
