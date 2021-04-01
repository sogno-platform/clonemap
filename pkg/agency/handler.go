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

// handler function for interaction with agents and ams

package agency

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

// handleGetAgency is the handler for get requests to path /api/agency
func (agency *Agency) handleGetAgency(w http.ResponseWriter, r *http.Request) {
	var cmapErr, httpErr error
	// return info about agency
	var agencyInfo schemas.AgencyInfo
	agencyInfo, cmapErr = agency.getAgencyInfo()
	if cmapErr != nil {
		httpErr = httpreply.CMAPError(w, cmapErr.Error())
		agency.logErrors(r.URL.Path, cmapErr, httpErr)
		return
	}
	httpErr = httpreply.Resource(w, agencyInfo, cmapErr)
	agency.logErrors(r.URL.Path, cmapErr, httpErr)
}

// handlePostAgent is the handler for post requests to path /api/agency/agents
func (agency *Agency) handlePostAgent(w http.ResponseWriter, r *http.Request) {
	var cmapErr, httpErr error
	// create new agent in agency
	var body []byte
	body, cmapErr = ioutil.ReadAll(r.Body)
	if cmapErr != nil {
		httpErr = httpreply.InvalidBodyError(w)
		agency.logErrors(r.URL.Path, cmapErr, httpErr)
		return
	}
	var agentInfo schemas.AgentInfo
	cmapErr = json.Unmarshal(body, &agentInfo)
	if cmapErr != nil {
		httpErr = httpreply.JSONUnmarshalError(w)
		agency.logErrors(r.URL.Path, cmapErr, httpErr)
		return
	}
	go agency.createAgent(agentInfo)
	httpErr = httpreply.Created(w, nil, "text/plain", []byte("Ressource Created"))
	agency.logErrors(r.URL.Path, cmapErr, httpErr)
}

// handlePostMsgs is the handler for post requests to path /api/agency/msgs
func (agency *Agency) handlePostMsgs(w http.ResponseWriter, r *http.Request) {
	var cmapErr, httpErr error
	var body []byte
	body, cmapErr = ioutil.ReadAll(r.Body)
	if cmapErr != nil {
		httpErr = httpreply.InvalidBodyError(w)
		agency.logErrors(r.URL.Path, cmapErr, httpErr)
		return
	}
	var msgs []schemas.ACLMessage
	cmapErr = json.Unmarshal(body, &msgs)
	if cmapErr != nil {
		httpErr = httpreply.JSONUnmarshalError(w)
		agency.logErrors(r.URL.Path, cmapErr, httpErr)
		return
	}
	agency.msgIn <- msgs
	httpErr = httpreply.Created(w, cmapErr, "text/plain", []byte("Ressource Created"))
	agency.logErrors(r.URL.Path, cmapErr, httpErr)
}

// handlePostUndeliverableMsg is the handler for post requests to path /api/agency/msgundeliv
func (agency *Agency) handlePostUndeliverableMsg(w http.ResponseWriter, r *http.Request) {
	var cmapErr, httpErr error
	var body []byte
	body, cmapErr = ioutil.ReadAll(r.Body)
	if cmapErr != nil {
		httpErr = httpreply.InvalidBodyError(w)
		agency.logErrors(r.URL.Path, cmapErr, httpErr)
		return
	}
	var msg schemas.ACLMessage
	cmapErr = json.Unmarshal(body, &msg)
	if cmapErr != nil {
		httpErr = httpreply.JSONUnmarshalError(w)
		agency.logErrors(r.URL.Path, cmapErr, httpErr)
		return
	}
	go agency.resendUndeliverableMsg(msg)
	httpErr = httpreply.Created(w, cmapErr, "text/plain", []byte("Ressource Created"))
	agency.logErrors(r.URL.Path, cmapErr, httpErr)
}

// handleDeleteAgentID is the handler for delete requests to path /api/agency/agents/{agentid}
func (agency *Agency) handleDeleteAgentID(w http.ResponseWriter, r *http.Request) {
	var cmapErr, httpErr error
	vars := mux.Vars(r)
	agentID, cmapErr := strconv.Atoi(vars["agentid"])
	if cmapErr != nil {
		httpErr = httpreply.NotFoundError(w)
		agency.logErrors(r.URL.Path, cmapErr, httpErr)
		return
	}
	// delete specified agent
	cmapErr = agency.removeAgent(agentID)
	httpErr = httpreply.Deleted(w, cmapErr)
	agency.logErrors(r.URL.Path, cmapErr, httpErr)
}

// handleGetAgentStatus is the handler for get requests to path /api/agency/agents/{agentid}/status
func (agency *Agency) handleGetAgentStatus(w http.ResponseWriter, r *http.Request) {
	var cmapErr, httpErr error
	vars := mux.Vars(r)
	agentID, cmapErr := strconv.Atoi(vars["agentid"])
	if cmapErr != nil {
		httpErr = httpreply.NotFoundError(w)
		agency.logErrors(r.URL.Path, cmapErr, httpErr)
		return
	}
	// return status of specified agent
	var agentStatus schemas.Status
	agentStatus, cmapErr = agency.getAgentStatus(agentID)
	if cmapErr != nil {
		httpErr = httpreply.CMAPError(w, cmapErr.Error())
		agency.logErrors(r.URL.Path, cmapErr, httpErr)
		return
	}
	httpErr = httpreply.Resource(w, agentStatus, cmapErr)
	agency.logErrors(r.URL.Path, cmapErr, httpErr)
}

// handlePutAgentCustom is the handler for put requests to path /api/agency/agents/{agentid}/custom
func (agency *Agency) handlePutAgentCustom(w http.ResponseWriter, r *http.Request) {
	var cmapErr, httpErr error
	vars := mux.Vars(r)
	agentID, cmapErr := strconv.Atoi(vars["agentid"])
	if cmapErr != nil {
		httpErr = httpreply.NotFoundError(w)
		agency.logErrors(r.URL.Path, cmapErr, httpErr)
		return
	}
	// update custom of specified agent
	var body []byte
	body, cmapErr = ioutil.ReadAll(r.Body)
	if cmapErr != nil {
		httpErr = httpreply.InvalidBodyError(w)
		agency.logErrors(r.URL.Path, cmapErr, httpErr)
		return
	}
	custom := string(body)
	cmapErr = agency.updateAgentCustom(agentID, custom)
	if cmapErr != nil {
		httpErr = httpreply.CMAPError(w, cmapErr.Error())
		agency.logErrors(r.URL.Path, cmapErr, httpErr)
		return
	}
	httpErr = httpreply.Updated(w, cmapErr)
	agency.logErrors(r.URL.Path, cmapErr, httpErr)
}

// methodNotAllowed is the default handler for valid paths but invalid methods
func (agency *Agency) methodNotAllowed(w http.ResponseWriter, r *http.Request) {
	httpErr := httpreply.MethodNotAllowed(w)
	cmapErr := errors.New("Error: Method not allowed on path " + r.URL.Path)
	agency.logErrors(r.URL.Path, cmapErr, httpErr)
}

// resourceNotFound is the default handler for invalid paths
func (agency *Agency) resourceNotFound(w http.ResponseWriter, r *http.Request) {
	httpErr := httpreply.NotFoundError(w)
	cmapErr := errors.New("resource not found")
	agency.logErrors(r.URL.Path, cmapErr, httpErr)
}

// logErrors logs errors if any
func (agency *Agency) logErrors(path string, cmapErr error, httpErr error) {
	if cmapErr != nil {
		agency.logError.Println(path, cmapErr)
	}
	if httpErr != nil {
		agency.logError.Println(path, httpErr)
	}
}

// loggingMiddleware logs request before calling final handler
func (agency *Agency) loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		agency.logInfo.Println("Received Request: ", r.Method, " ", r.URL.EscapedPath())
		next.ServeHTTP(w, r)
	})
}

// server creates the fe server
func (agency *Agency) server(port int) (serv *http.Server) {
	r := mux.NewRouter()
	s := r.PathPrefix("/api").Subrouter()
	s.Path("/agency").Methods("GET").HandlerFunc(agency.handleGetAgency)
	s.Path("/agency").Methods("PUT", "POST", "DELETE").HandlerFunc(agency.methodNotAllowed)
	s.Path("/agency/agents").Methods("POST").HandlerFunc(agency.handlePostAgent)
	s.Path("/agency/agents").Methods("PUT", "GET", "DELETE").HandlerFunc(agency.methodNotAllowed)
	s.Path("/agency/msgs").Methods("POST").HandlerFunc(agency.handlePostMsgs)
	s.Path("/agency/msgs").Methods("PUT", "GET", "DELETE").HandlerFunc(agency.methodNotAllowed)
	s.Path("/agency/msgundeliv").Methods("POST").HandlerFunc(agency.handlePostUndeliverableMsg)
	s.Path("/agency/msgundeliv").Methods("PUT", "GET", "DELETE").
		HandlerFunc(agency.methodNotAllowed)
	s.Path("/agency/agents/{agentid}").Methods("DELETE").HandlerFunc(agency.handleDeleteAgentID)
	s.Path("/agency/agents/{agentid}").Methods("PUT", "GET", "POST").
		HandlerFunc(agency.methodNotAllowed)
	s.Path("/agency/agents/{agentid}/status").Methods("GET").
		HandlerFunc(agency.handleGetAgentStatus)
	s.Path("/agency/agents/{agentid}/status").Methods("PUT", "DELETE", "POST").
		HandlerFunc(agency.methodNotAllowed)
	s.Path("/agency/agents/{agentid}/custom").Methods("PUT").
		HandlerFunc(agency.handlePutAgentCustom)
	s.Path("/agency/agents/{agentid}/custom").Methods("GET", "DELETE", "POST").
		HandlerFunc(agency.methodNotAllowed)
	s.Use(agency.loggingMiddleware)
	r.PathPrefix("").HandlerFunc(agency.resourceNotFound)
	serv = &http.Server{
		Addr:    ":" + strconv.Itoa(port),
		Handler: r,
	}
	return
}

// listen opens a http server listening and serving request
func (agency *Agency) listen(serv *http.Server) (err error) {
	agency.logInfo.Println("Frontend listening on " + serv.Addr)
	err = serv.ListenAndServe()
	return
}
