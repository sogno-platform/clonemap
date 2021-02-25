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

// handler for http requests

package ams

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
func (ams *AMS) handleAlive(w http.ResponseWriter, r *http.Request) {
	var httpErr error
	httpErr = httpreply.Alive(w, nil)
	ams.logErrors(r.URL.Path, nil, httpErr)
	return
}

// handleCloneMAP is the handler for requests to path /api/clonemap
func (ams *AMS) handleCloneMAP(w http.ResponseWriter, r *http.Request) {
	var cmapErr, httpErr error
	defer ams.logErrors(r.URL.Path, cmapErr, httpErr)
	// return info about running clonemap instance
	var cmapInfo schemas.CloneMAP
	cmapInfo, cmapErr = ams.getCloneMAPInfo()
	if cmapErr != nil {
		httpErr = httpreply.CMAPError(w, cmapErr.Error())
		return
	}
	httpErr = httpreply.Resource(w, cmapInfo, cmapErr)
	return
}

// handleGetMAS is the handler for get requests to path /api/clonemap/mas
func (ams *AMS) handleGetMAS(w http.ResponseWriter, r *http.Request) {
	var cmapErr, httpErr error
	defer ams.logErrors(r.URL.Path, cmapErr, httpErr)
	var mass []schemas.MASInfoShort
	mass, cmapErr = ams.getMASsShort()
	if cmapErr != nil {
		httpErr = httpreply.CMAPError(w, cmapErr.Error())
		return
	}
	httpErr = httpreply.Resource(w, mass, cmapErr)
	return
}

// handlePostMAS is the handler for post requests to path /api/clonemap/mas
func (ams *AMS) handlePostMAS(w http.ResponseWriter, r *http.Request) {
	var cmapErr, httpErr error
	defer ams.logErrors(r.URL.Path, cmapErr, httpErr)
	// create new MAS
	var body []byte
	body, cmapErr = ioutil.ReadAll(r.Body)
	if cmapErr != nil {
		httpErr = httpreply.InvalidBodyError(w)
		return
	}
	var masSpec schemas.MASSpec
	cmapErr = json.Unmarshal(body, &masSpec)
	if cmapErr != nil {
		httpErr = httpreply.JSONUnmarshalError(w)
		return
	}
	cmapErr = ams.createMAS(masSpec)
	if cmapErr != nil {
		httpErr = httpreply.CMAPError(w, cmapErr.Error())
		return
	}
	httpErr = httpreply.Created(w, cmapErr, "text/plain", []byte("Ressource Created"))
	return
}

// handleGetMASID is the handler for get requests to path /api/clonemap/mas/{masid}
func (ams *AMS) handleGetMASID(w http.ResponseWriter, r *http.Request) {
	var cmapErr, httpErr error
	defer ams.logErrors(r.URL.Path, cmapErr, httpErr)
	vars := mux.Vars(r)
	masID, cmapErr := strconv.Atoi(vars["masid"])
	if cmapErr != nil {
		httpErr = httpreply.NotFoundError(w)
		return
	}
	// return long information about specified MAS
	var masInfo schemas.MASInfo
	masInfo, cmapErr = ams.getMASInfo(masID)
	if cmapErr != nil {
		httpErr = httpreply.CMAPError(w, cmapErr.Error())
		return
	}
	httpErr = httpreply.Resource(w, masInfo, cmapErr)
	return
}

// handleDeleteMASID is the handler for delete requests to path /api/clonemap/mas/{masid}
func (ams *AMS) handleDeleteMASID(w http.ResponseWriter, r *http.Request) {
	var cmapErr, httpErr error
	defer ams.logErrors(r.URL.Path, cmapErr, httpErr)
	vars := mux.Vars(r)
	masID, cmapErr := strconv.Atoi(vars["masid"])
	if cmapErr != nil {
		httpErr = httpreply.NotFoundError(w)
		return
	}
	// delete specified MAS
	cmapErr = ams.removeMAS(masID)
	if cmapErr != nil {
		httpErr = httpreply.CMAPError(w, cmapErr.Error())
		return
	}
	httpErr = httpreply.Deleted(w, cmapErr)
	return
}

// handleGetMASName is the handler for get requests to path /api/clonemap/mas/name/{name}
func (ams *AMS) handleGetMASName(w http.ResponseWriter, r *http.Request) {
	var cmapErr, httpErr error
	defer ams.logErrors(r.URL.Path, cmapErr, httpErr)
	vars := mux.Vars(r)
	name := vars["name"]
	// search for MAS with matching name
	var ids []int
	ids, cmapErr = ams.getMASByName(name)
	if cmapErr != nil {
		httpErr = httpreply.CMAPError(w, cmapErr.Error())
		return
	}
	httpErr = httpreply.Resource(w, ids, cmapErr)
	return
}

// handleGetAgents is the handler for get requests to path /api/clonemap/mas/{masid}/agents
func (ams *AMS) handleGetAgents(w http.ResponseWriter, r *http.Request) {
	var cmapErr, httpErr error
	defer ams.logErrors(r.URL.Path, cmapErr, httpErr)
	vars := mux.Vars(r)
	masID, cmapErr := strconv.Atoi(vars["masid"])
	if cmapErr != nil {
		httpErr = httpreply.NotFoundError(w)
		return
	}
	// return short information of all agents in specified MAS
	var agents schemas.Agents
	agents, cmapErr = ams.getAgents(masID)
	if cmapErr != nil {
		httpErr = httpreply.CMAPError(w, cmapErr.Error())
		return
	}
	httpErr = httpreply.Resource(w, agents, cmapErr)
	return
}

// handlePostAgent is the handler for post requests to path /api/clonemap/mas/{masid}/agents
func (ams *AMS) handlePostAgent(w http.ResponseWriter, r *http.Request) {
	var cmapErr, httpErr error
	defer ams.logErrors(r.URL.Path, cmapErr, httpErr)
	vars := mux.Vars(r)
	masID, cmapErr := strconv.Atoi(vars["masid"])
	if cmapErr != nil {
		httpErr = httpreply.NotFoundError(w)
		return
	}
	// create new agent in MAS
	var body []byte
	body, cmapErr = ioutil.ReadAll(r.Body)
	if cmapErr != nil {
		httpErr = httpreply.InvalidBodyError(w)
		return
	}
	var groupSpecs []schemas.ImageGroupSpec
	cmapErr = json.Unmarshal(body, &groupSpecs)
	if cmapErr != nil {
		httpErr = httpreply.JSONUnmarshalError(w)
		return
	}
	cmapErr = ams.createAgents(masID, groupSpecs)
	if cmapErr != nil {
		httpErr = httpreply.CMAPError(w, cmapErr.Error())
		return
	}
	httpErr = httpreply.Created(w, cmapErr, "text/plain", []byte("Ressource Created"))
	return
}

// handleGetAgentID is the handler for get requests to path
// /api/clonemap/mas/{masid}/agents/{agentid}
func (ams *AMS) handleGetAgentID(w http.ResponseWriter, r *http.Request) {
	var cmapErr, httpErr error
	defer ams.logErrors(r.URL.Path, cmapErr, httpErr)
	masID, agentID, cmapErr := getAgentID(r)
	if cmapErr != nil {
		httpErr = httpreply.NotFoundError(w)
		return
	}
	var agentInfo schemas.AgentInfo
	agentInfo, cmapErr = ams.getAgentInfo(masID, agentID)
	if cmapErr != nil {
		httpErr = httpreply.CMAPError(w, cmapErr.Error())
		return
	}
	httpErr = httpreply.Resource(w, agentInfo, cmapErr)
	return
}

// handleDeleteAgentID is the handler for delete requests to path
// /api/clonemap/mas/{masid}/agents/{agentid}
func (ams *AMS) handleDeleteAgentID(w http.ResponseWriter, r *http.Request) {
	var cmapErr, httpErr error
	defer ams.logErrors(r.URL.Path, cmapErr, httpErr)
	masID, agentID, cmapErr := getAgentID(r)
	if cmapErr != nil {
		httpErr = httpreply.NotFoundError(w)
		return
	}
	// delete specified agent
	cmapErr = ams.removeAgent(masID, agentID)
	if cmapErr != nil {
		httpErr = httpreply.CMAPError(w, cmapErr.Error())
		return
	}
	httpErr = httpreply.Deleted(w, cmapErr)
	return
}

// handleGetAgentAddress is the handler for get requests to path
// /api/clonemap/mas/{masid}/agents/{agentid}/address
func (ams *AMS) handleGetAgentAddress(w http.ResponseWriter, r *http.Request) {
	var cmapErr, httpErr error
	defer ams.logErrors(r.URL.Path, cmapErr, httpErr)
	masID, agentID, cmapErr := getAgentID(r)
	if cmapErr != nil {
		httpErr = httpreply.NotFoundError(w)
		return
	}
	// return address of specified agent
	var agentAddr schemas.Address
	agentAddr, cmapErr = ams.getAgentAddress(masID, agentID)
	httpErr = httpreply.Resource(w, agentAddr, cmapErr)
	return
}

// handlePutAgentAddress is the handler for put requests to path
// /api/clonemap/mas/{masid}/agents/{agentid}/address
func (ams *AMS) handlePutAgentAddress(w http.ResponseWriter, r *http.Request) {
	var cmapErr, httpErr error
	defer ams.logErrors(r.URL.Path, cmapErr, httpErr)
	masID, agentID, cmapErr := getAgentID(r)
	if cmapErr != nil {
		httpErr = httpreply.NotFoundError(w)
		return
	}
	// update address of specified agent
	var body []byte
	body, cmapErr = ioutil.ReadAll(r.Body)
	if cmapErr != nil {
		httpErr = httpreply.InvalidBodyError(w)
		return
	}
	var agentAddr schemas.Address
	cmapErr = json.Unmarshal(body, &agentAddr)
	if cmapErr != nil {
		httpErr = httpreply.JSONUnmarshalError(w)
		return
	}
	cmapErr = ams.updateAgentAddress(masID, agentID, agentAddr)
	if cmapErr != nil {
		httpErr = httpreply.CMAPError(w, cmapErr.Error())
		return
	}
	httpErr = httpreply.Updated(w, cmapErr)
	return
}

// handlePutAgentCustom is the put handler for requests to path
// /api/clonemap/mas/{masid}/agents/{agentid}/custom
func (ams *AMS) handlePutAgentCustom(w http.ResponseWriter, r *http.Request) {
	var cmapErr, httpErr error
	defer ams.logErrors(r.URL.Path, cmapErr, httpErr)
	masID, agentID, cmapErr := getAgentID(r)
	if cmapErr != nil {
		httpErr = httpreply.NotFoundError(w)
		return
	}
	// update custom of specified agent
	var body []byte
	body, cmapErr = ioutil.ReadAll(r.Body)
	if cmapErr != nil {
		httpErr = httpreply.InvalidBodyError(w)
		return
	}
	custom := string(body)
	cmapErr = ams.updateAgentCustom(masID, agentID, custom)
	if cmapErr != nil {
		httpErr = httpreply.CMAPError(w, cmapErr.Error())
		return
	}
	httpErr = httpreply.Updated(w, cmapErr)
	return
}

// handleGetAgentName is the handler for get requests to path
// /api/clonemap/mas/{masid}/agents/name/{name}
func (ams *AMS) handleGetAgentName(w http.ResponseWriter, r *http.Request) {
	var cmapErr, httpErr error
	defer ams.logErrors(r.URL.Path, cmapErr, httpErr)
	vars := mux.Vars(r)
	masID, cmapErr := strconv.Atoi(vars["masid"])
	if cmapErr != nil {
		httpErr = httpreply.NotFoundError(w)
		return
	}
	name := vars["name"]
	// search for agents with matching name
	var ids []int
	ids, cmapErr = ams.getAgentsByName(masID, name)
	if cmapErr != nil {
		httpErr = httpreply.CMAPError(w, cmapErr.Error())
		return
	}
	httpErr = httpreply.Resource(w, ids, cmapErr)
	return
}

// handleGetAgencies is the handler for get requests to path /api/cloumap/mas/{masid}/agencies
func (ams *AMS) handleGetAgencies(w http.ResponseWriter, r *http.Request) {
	var cmapErr, httpErr error
	defer ams.logErrors(r.URL.Path, cmapErr, httpErr)
	vars := mux.Vars(r)
	masID, cmapErr := strconv.Atoi(vars["masid"])
	if cmapErr != nil {
		httpErr = httpreply.NotFoundError(w)
		return
	}
	// return information of specified agency
	var agencies schemas.Agencies
	agencies, cmapErr = ams.getAgencies(masID)
	if cmapErr != nil {
		httpErr = httpreply.CMAPError(w, cmapErr.Error())
		return
	}
	httpErr = httpreply.Resource(w, agencies, cmapErr)
	return
}

// handleGetAgencyID is the handler for get requests to path
// /api/clonemap/mas/{masid}/imgroup/{imid}/agencies/{agencyid}
func (ams *AMS) handleGetAgencyID(w http.ResponseWriter, r *http.Request) {
	var cmapErr, httpErr error
	defer ams.logErrors(r.URL.Path, cmapErr, httpErr)
	vars := mux.Vars(r)
	masID, cmapErr := strconv.Atoi(vars["masid"])
	if cmapErr != nil {
		httpErr = httpreply.NotFoundError(w)
		return
	}
	imID, cmapErr := strconv.Atoi(vars["imid"])
	if cmapErr != nil {
		httpErr = httpreply.NotFoundError(w)
		return
	}
	agencyID, cmapErr := strconv.Atoi(vars["agencyid"])
	if cmapErr != nil {
		httpErr = httpreply.NotFoundError(w)
		return
	}
	var agencySpec schemas.AgencyInfoFull
	agencySpec, cmapErr = ams.getAgencyInfoFull(masID, imID, agencyID)
	if cmapErr != nil {
		httpErr = httpreply.CMAPError(w, cmapErr.Error())
		return
	}
	httpErr = httpreply.Resource(w, agencySpec, cmapErr)
	return
}

// methodNotAllowed is the default handler for valid paths but invalid methods
func (ams *AMS) methodNotAllowed(w http.ResponseWriter, r *http.Request) {
	httpErr := httpreply.MethodNotAllowed(w)
	cmapErr := errors.New("Error: Method not allowed on path " + r.URL.Path)
	ams.logErrors(r.URL.Path, cmapErr, httpErr)
	return
}

// resourceNotFound is the default handler for invalid paths
func (ams *AMS) resourceNotFound(w http.ResponseWriter, r *http.Request) {
	httpErr := httpreply.NotFoundError(w)
	cmapErr := errors.New("Resource not found")
	ams.logErrors(r.URL.Path, cmapErr, httpErr)
	return
}

// logErrors logs errors if any
func (ams *AMS) logErrors(path string, cmapErr error, httpErr error) {
	if cmapErr != nil {
		ams.logError.Println(path, cmapErr)
	}
	if httpErr != nil {
		ams.logError.Println(path, httpErr)
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

// loggingMiddleware logs request before calling final handler
func (ams *AMS) loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ams.logInfo.Println("Received Request: ", r.Method, " ", r.URL.EscapedPath())
		next.ServeHTTP(w, r)
	})
}

// server creates the ams server
func (ams *AMS) server(port int) (serv *http.Server) {
	r := mux.NewRouter()
	// r.HandleFunc("/api/", ams.handleAPI)
	s := r.PathPrefix("/api").Subrouter()
	s.Path("/alive").Methods("GET").HandlerFunc(ams.handleAlive)
	s.Path("/alive").Methods("POST", "PUT", "DELETE").HandlerFunc(ams.methodNotAllowed)
	s.Path("/clonemap").Methods("GET").HandlerFunc(ams.handleCloneMAP)
	s.Path("/clonemap").Methods("POST", "PUT", "DELETE").HandlerFunc(ams.methodNotAllowed)
	s.Path("/clonemap/mas").Methods("GET").HandlerFunc(ams.handleGetMAS)
	s.Path("/clonemap/mas").Methods("POST").HandlerFunc(ams.handlePostMAS)
	s.Path("/clonemap/mas").Methods("PUT", "DELETE").HandlerFunc(ams.methodNotAllowed)
	s.Path("/clonemap/mas/{masid}").Methods("GET").HandlerFunc(ams.handleGetMASID)
	s.Path("/clonemap/mas/{masid}").Methods("DELETE").HandlerFunc(ams.handleDeleteMASID)
	s.Path("/clonemap/mas/{masid}").Methods("PUT", "POST").HandlerFunc(ams.methodNotAllowed)
	s.Path("/clonemap/mas/name/{name}").Methods("GET").HandlerFunc(ams.handleGetMASName)
	s.Path("/clonemap/mas/name/{name}").Methods("PUT", "POST", "DELETE").
		HandlerFunc(ams.methodNotAllowed)
	s.Path("/clonemap/mas/{masid}/agents").Methods("GET").HandlerFunc(ams.handleGetAgents)
	s.Path("/clonemap/mas/{masid}/agents").Methods("POST").HandlerFunc(ams.handlePostAgent)
	s.Path("/clonemap/mas/{masid}/agents").Methods("PUT", "DELETE").
		HandlerFunc(ams.methodNotAllowed)
	s.Path("/clonemap/mas/{masid}/agents/{agentid}").Methods("GET").
		HandlerFunc(ams.handleGetAgentID)
	s.Path("/clonemap/mas/{masid}/agents/{agentid}").Methods("DELETE").
		HandlerFunc(ams.handleDeleteAgentID)
	s.Path("/clonemap/mas/{masid}/agents/{agentid}").Methods("PUT", "POST").
		HandlerFunc(ams.methodNotAllowed)
	s.Path("/clonemap/mas/{masid}/agents/{agentid}/address").Methods("GET").
		HandlerFunc(ams.handleGetAgentAddress)
	s.Path("/clonemap/mas/{masid}/agents/{agentid}/address").Methods("PUT").
		HandlerFunc(ams.handlePutAgentAddress)
	s.Path("/clonemap/mas/{masid}/agents/{agentid}/address").Methods("DELETE", "POST").
		HandlerFunc(ams.methodNotAllowed)
	s.Path("/clonemap/mas/{masid}/agents/{agentid}/custom").Methods("PUT").
		HandlerFunc(ams.handlePutAgentCustom)
	s.Path("/clonemap/mas/{masid}/agents/{agentid}/custom").Methods("DELETE", "POST", "GET").
		HandlerFunc(ams.methodNotAllowed)
	s.Path("/clonemap/mas/{masid}/agents/name/{name}").Methods("GET").
		HandlerFunc(ams.handleGetAgentName)
	s.Path("/clonemap/mas/{masid}/agents/name/{name}").Methods("DELETE", "POST", "PUT").
		HandlerFunc(ams.methodNotAllowed)
	s.Path("/clonemap/mas/{masid}/agencies").Methods("GET").HandlerFunc(ams.handleGetAgencies)
	s.Path("/clonemap/mas/{masid}/agencies").Methods("PUT", "DELETE", "POST").
		HandlerFunc(ams.methodNotAllowed)
	s.Path("/clonemap/mas/{masid}/imgroup/{imid}/agency/{agencyid}").Methods("GET").
		HandlerFunc(ams.handleGetAgencyID)
	s.Path("/clonemap/mas/{masid}/imgroup/{imid}/agency/{agencyid}").
		Methods("PUT", "DELETE", "POST").HandlerFunc(ams.methodNotAllowed)
	s.PathPrefix("").HandlerFunc(ams.resourceNotFound)
	s.Use(ams.loggingMiddleware)
	serv = &http.Server{
		Addr:    ":" + strconv.Itoa(port),
		Handler: r,
	}
	return
}

// listen opens a http server listening and serving request
func (ams *AMS) listen(serv *http.Server) (err error) {
	ams.logInfo.Println("AMS listening on " + serv.Addr)
	err = serv.ListenAndServe()
	return
}
