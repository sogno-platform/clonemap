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
	"strconv"

	"git.rwth-aachen.de/acs/public/cloud/mas/clonemap/pkg/common/httpreply"
	"git.rwth-aachen.de/acs/public/cloud/mas/clonemap/pkg/schemas"
	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
)

// handleGetModules is the handler for get requests to path /api/pf/modules
func (fe *Frontend) handleGetModules(w http.ResponseWriter, r *http.Request) {
	var cmapErr, httpErr error
	var mods schemas.ModuleStatus
	mods, cmapErr = fe.getModuleStatus()
	if cmapErr != nil {
		httpErr = httpreply.CMAPError(w, cmapErr.Error())
		fe.logErrors(r.URL.Path, cmapErr, httpErr)
		return
	}
	httpErr = httpreply.Resource(w, mods, cmapErr)
	fe.logErrors(r.URL.Path, cmapErr, httpErr)
	return
}

// methodNotAllowed is the default handler for valid paths but invalid methods
func (fe *Frontend) methodNotAllowed(w http.ResponseWriter, r *http.Request) {
	httpErr := httpreply.MethodNotAllowed(w)
	cmapErr := errors.New("Error: Method not allowed on path " + r.URL.Path)
	fe.logErrors(r.URL.Path, cmapErr, httpErr)
	return
}

// resourceNotFound is the default handler for invalid paths
func (fe *Frontend) resourceNotFound(w http.ResponseWriter, r *http.Request) {
	httpErr := httpreply.NotFoundError(w)
	cmapErr := errors.New("Resource not found")
	fe.logErrors(r.URL.Path, cmapErr, httpErr)
	return
}

// logErrors logs errors if any
func (fe *Frontend) logErrors(path string, cmapErr error, httpErr error) {
	if cmapErr != nil {
		fe.logError.Println(path, cmapErr)
	}
	if httpErr != nil {
		fe.logError.Println(path, httpErr)
	}
	return
}

// getAgentID returns the masID and agentID from the path
func getAgentID(r *http.Request) (masID int, agentID int, err error) {
	vars := mux.Vars(r)
	masID, err = strconv.Atoi(vars["masid"])
	if err != nil {
		return
	}
	agentID, err = strconv.Atoi(vars["agentid"])
	if err != nil {
		return
	}
	return
}

/****************************** Handler part for the df ********************/
// getDesc return the masID and description from the path
func getDesc(r *http.Request) (masID int, desc string, err error) {
	vars := mux.Vars(r)
	masID, err = strconv.Atoi(vars["masid"])
	if err != nil {
		return
	}
	desc = vars["desc"]
	return
}

func getDist(r *http.Request) (masID int, desc string, nodeID int, dist float64, err error) {
	vars := mux.Vars(r)
	masID, err = strconv.Atoi(vars["masid"])
	if err != nil {
		return
	}
	desc = vars["desc"]
	nodeID, err = strconv.Atoi(vars["nodeid"])
	if err != nil {
		return
	}
	dist, err = strconv.ParseFloat(vars["dist"], 64)
	if err != nil {
		return
	}
	return
}

// loggingMiddleware logs request before calling final handler
func (fe *Frontend) loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fe.logInfo.Println("Received Request: ", r.Method, " ", r.URL.EscapedPath())
		next.ServeHTTP(w, r)
	})
}

// server creates the fe server
func (fe *Frontend) server(port int) (serv *http.Server) {
	r := mux.NewRouter()
	s := r.PathPrefix("/api").Subrouter()
	s.Path("/overview").Methods("GET").HandlerFunc(fe.handleGetMASs)
	s.Path("/overview").Methods("POST").HandlerFunc(fe.handlePostMAS)
	s.Path("/overview").Methods("PUT", "DELETE").HandlerFunc(fe.methodNotAllowed)

	// api for mas
	s.Path("/ams/mas").Methods("GET").HandlerFunc(fe.handleGetMASs)
	s.Path("/ams/mas").Methods("POST").HandlerFunc(fe.handlePostMAS)
	s.Path("/ams/mas").Methods("PUT", "DELETE").HandlerFunc(fe.methodNotAllowed)
	s.Path("/ams/mas/{masid}").Methods("GET").HandlerFunc(fe.handleGetMASID)
	s.Path("/ams/mas/{masid}").Methods("DELETE").HandlerFunc(fe.handleDeleteMASID)
	s.Path("/ams/mas/{masid}").Methods("PUT", "POST").HandlerFunc(fe.methodNotAllowed)
	s.Path("/ams/mas/{masid}/agents").Methods("POST").HandlerFunc(fe.handlePostAgent)
	s.Path("/ams/mas/{masid}/agents").Methods("PUT", "GET", "DELETE").HandlerFunc(fe.methodNotAllowed)
	s.Path("/ams/mas/{masid}/agents/{agentid}").Methods("GET").HandlerFunc(fe.handleGetAgentID)
	s.Path("/ams/mas/{masid}/agents/{agentid}").Methods("DELETE").
		HandlerFunc(fe.handleDeleteAgentID)
	s.Path("/ams/mas/{masid}/agents/{agentid}").Methods("PUT", "POST").
		HandlerFunc(fe.methodNotAllowed)
	s.Path("/pf/modules").Methods("GET").HandlerFunc(fe.handleGetModules)
	s.Path("/pf/modules").Methods("POST", "PUT", "POST").HandlerFunc(fe.methodNotAllowed)

	// api for df
	s.Path("/df/{masid}/svc").Methods("GET").HandlerFunc(fe.handleGetSvcs)
	s.Path("/df/{masid}/svc").Methods("POST").HandlerFunc(fe.handlePostSvc)
	s.Path("/df/{masid}/svc/desc/{desc}").Methods("GET").HandlerFunc(fe.handleGetSvc)
	s.Path("/df/{masid}/svc/desc/{desc}/node/{nodeid}/dist/{dist}").Methods("Get").HandlerFunc(fe.handleSvcWithDist)

	// api for logger
	s.Path("/df/{masid}/svc/desc/{desc}").Methods("Get").HandlerFunc(fe.handleGetSvc)

	s.PathPrefix("").HandlerFunc(fe.resourceNotFound)
	s.Use(fe.loggingMiddleware)
	r.PathPrefix("/").Handler(http.StripPrefix("/", http.FileServer(http.Dir("./web"))))
	//r.HandleFunc("/", http.FileServer(http.Dir("./web")).ServeHTTP)
	// r.HandleFunc("/css/", http.FileServer(http.Dir("./web/css")).ServeHTTP)
	// r.HandleFunc("/js/", http.FileServer了(http.Dir("./web/js")).ServeHTTP)

	headersOk := handlers.AllowedHeaders([]string{"X-Request-With"})
	originsOk := handlers.AllowedOrigins([]string{"*"})
	methodsOk := handlers.AllowedMethods([]string{"GET", "HEAD", "POST", "PUT", "OPTIONS"})

	serv = &http.Server{
		Addr:    ":" + strconv.Itoa(port),
		Handler: handlers.CORS(originsOk, headersOk, methodsOk)(r),
	}

	return
}

// listen opens a http server listening and serving request
func (fe *Frontend) listen(serv *http.Server) (err error) {
	fe.logInfo.Println("Frontend listening on " + serv.Addr)
	err = http.ListenAndServe(serv.Addr, serv.Handler)
	return
}
