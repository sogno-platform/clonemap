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
	"strings"

	"git.rwth-aachen.de/acs/public/cloud/mas/clonemap/pkg/common/httpreply"
	"git.rwth-aachen.de/acs/public/cloud/mas/clonemap/pkg/schemas"
)

// handleAPI is the global handler for requests to path /api
func (agency *Agency) handleAPI(w http.ResponseWriter, r *http.Request) {
	var cmapErr, httpErr error
	agency.logInfo.Println("Received Request: ", r.Method, " ", r.URL.EscapedPath())
	// determine which ressource is requested and call corresponding handler
	respath := strings.Split(r.URL.EscapedPath(), "/")
	resvalid := false

	switch len(respath) {
	case 3:
		if respath[2] == "agency" {
			cmapErr, httpErr = agency.handleAgency(w, r)
			resvalid = true
		}
	case 4:
		if respath[2] == "agency" && respath[3] == "agents" {
			cmapErr, httpErr = agency.handleAgent(w, r)
			resvalid = true
		} else if respath[2] == "agency" && respath[3] == "msgs" {
			cmapErr, httpErr = agency.handleMsgs(w, r)
			resvalid = true
		} else if respath[2] == "agency" && respath[3] == "msgundeliv" {
			cmapErr, httpErr = agency.handleMsgs(w, r)
			resvalid = true
		}
	case 5:
		var id int
		id, cmapErr = strconv.Atoi(respath[4])
		if respath[2] == "agency" && respath[3] == "agents" && cmapErr == nil {
			cmapErr, httpErr = agency.handleAgentID(id, w, r)
			resvalid = true
		}
	case 6:
		var agentID int
		agentID, cmapErr = strconv.Atoi(respath[4])
		if respath[2] == "agency" && respath[3] == "agents" && cmapErr == nil {
			if respath[5] == "status" {
				cmapErr, httpErr = agency.handleAgentStatus(agentID, w, r)
				resvalid = true
			}
		}
	default:
		cmapErr = errors.New("Resource not found")
	}

	if !resvalid {
		httpErr = httpreply.NotFoundError(w)
		cmapErr = errors.New("Resource not found")
	}
	if cmapErr != nil {
		agency.logError.Println(respath, cmapErr)
	}
	if httpErr != nil {
		agency.logError.Println(respath, httpErr)
	}
}

// handleAgency is the handler for requests to path /api/agency
func (agency *Agency) handleAgency(w http.ResponseWriter, r *http.Request) (cmapErr,
	httpErr error) {
	if r.Method == "GET" {
		// return info about agency
		var agencyInfo schemas.AgencyInfo
		agencyInfo, cmapErr = agency.getAgencyInfo()
		httpErr = httpreply.Resource(w, agencyInfo, cmapErr)
	} else {
		httpErr = httpreply.MethodNotAllowed(w)
		cmapErr = errors.New("Error: Method not allowed on path /api/agency")
	}
	return
}

// handleAgent is the handler for requests to path /api/agency/agents
func (agency *Agency) handleAgent(w http.ResponseWriter, r *http.Request) (cmapErr, httpErr error) {
	if r.Method == "POST" {
		// create new agent in agency
		var body []byte
		body, cmapErr = ioutil.ReadAll(r.Body)
		if cmapErr == nil {
			var agentInfo schemas.AgentInfo
			cmapErr = json.Unmarshal(body, &agentInfo)
			if cmapErr == nil {
				go agency.createAgent(agentInfo)
				httpErr = httpreply.Created(w, nil, "text/plain", []byte("Ressource Created"))
			} else {
				httpErr = httpreply.JSONUnmarshalError(w)
			}
		} else {
			httpErr = httpreply.InvalidBodyError(w)
		}
	} else {
		httpErr = httpreply.MethodNotAllowed(w)
		cmapErr = errors.New("Error: Method not allowed on path /api/agency/agents")
	}
	return
}

// handleMsgs is the handler for requests to path /api/agency/msgs
func (agency *Agency) handleMsgs(w http.ResponseWriter, r *http.Request) (cmapErr, httpErr error) {
	if r.Method == "POST" {
		var body []byte
		body, cmapErr = ioutil.ReadAll(r.Body)
		if cmapErr == nil {
			var msgs []schemas.ACLMessage
			cmapErr = json.Unmarshal(body, &msgs)
			if cmapErr == nil {
				agency.msgIn <- msgs
				httpErr = httpreply.Created(w, cmapErr, "text/plain", []byte("Ressource Created"))
			} else {
				httpErr = httpreply.JSONUnmarshalError(w)
			}
		} else {
			httpErr = httpreply.InvalidBodyError(w)
		}
	} else {
		httpErr = httpreply.MethodNotAllowed(w)
		cmapErr = errors.New("Error: Method not allowed on path /api/agency/msgs")
	}
	return
}

// handleUndeliverableMsg is the handler for requests to path /api/agency/msgundeliv
func (agency *Agency) handleUndeliverableMsg(w http.ResponseWriter,
	r *http.Request) (cmapErr, httpErr error) {
	if r.Method == "POST" {
		var body []byte
		body, cmapErr = ioutil.ReadAll(r.Body)
		if cmapErr == nil {
			var msg schemas.ACLMessage
			cmapErr = json.Unmarshal(body, &msg)
			if cmapErr == nil {
				go agency.resendUndeliverableMsg(msg)
				httpErr = httpreply.Created(w, cmapErr, "text/plain", []byte("Ressource Created"))
			} else {
				httpErr = httpreply.JSONUnmarshalError(w)
			}
		} else {
			httpErr = httpreply.InvalidBodyError(w)
		}
	} else {
		httpErr = httpreply.MethodNotAllowed(w)
		cmapErr = errors.New("Error: Method not allowed on path /api/agency/msgundeliv")
	}
	return
}

// handleAgentID is the handler for requests to path /api/agency/agents/{agent-id}
func (agency *Agency) handleAgentID(agid int, w http.ResponseWriter, r *http.Request) (cmapErr,
	httpErr error) {
	if r.Method == "DELETE" {
		// delete specified agent
		go agency.removeAgent(agid)
		httpErr = httpreply.Deleted(w, nil)

	} else {
		httpErr = httpreply.MethodNotAllowed(w)
		cmapErr = errors.New("Error: Method not allowed on path /api/agency/agents/{agent-id}")
	}
	return
}

// handleAgentStatus is the handler for requests to path /api/agency/agents/{agent-id}/status
func (agency *Agency) handleAgentStatus(agid int, w http.ResponseWriter,
	r *http.Request) (cmapErr, httpErr error) {
	if r.Method == "GET" {
		// return status of specified agent
		var agentStatus schemas.Status
		agentStatus, cmapErr = agency.getAgentStatus(agid)
		httpErr = httpreply.Resource(w, agentStatus, cmapErr)
	} else {
		httpErr = httpreply.MethodNotAllowed(w)
		cmapErr = errors.New("Error: Method not allowed on path /api/agency/agents/{agent-id}/" +
			"status")
	}
	return
}

// listen opens a http server listening and serving request
func (agency *Agency) listen() (err error) {
	mux := http.NewServeMux()
	mux.HandleFunc("/api/", agency.handleAPI)
	s := &http.Server{
		Addr:    ":10000",
		Handler: mux,
	}
	err = s.ListenAndServe()
	return
}
