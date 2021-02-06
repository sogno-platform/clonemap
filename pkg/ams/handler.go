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
	"strings"

	"git.rwth-aachen.de/acs/public/cloud/mas/clonemap/pkg/common/httpreply"
	"git.rwth-aachen.de/acs/public/cloud/mas/clonemap/pkg/schemas"
)

// handleAPI is the global handler for requests to path /api
func (ams *AMS) handleAPI(w http.ResponseWriter, r *http.Request) {
	var cmapErr, httpErr error
	ams.logInfo.Println("Received Request: ", r.Method, " ", r.URL.EscapedPath())
	// determine which ressource is requested and call corresponding handler
	respath := strings.Split(r.URL.EscapedPath(), "/")
	resvalid := false

	switch len(respath) {
	case 3:
		if respath[2] == "clonemap" {
			cmapErr, httpErr = ams.handleCloneMAP(w, r)
			resvalid = true
		} else if respath[2] == "alive" {
			cmapErr, httpErr = ams.handleAlive(w, r)
			resvalid = true
		}
	case 4:
		if respath[2] == "clonemap" && respath[3] == "mas" {
			cmapErr, httpErr = ams.handleMAS(w, r)
			resvalid = true
		}
	case 5:
		var masID int
		masID, cmapErr = strconv.Atoi(respath[4])
		if respath[2] == "clonemap" && respath[3] == "mas" && cmapErr == nil {
			cmapErr, httpErr = ams.handlemasID(masID, w, r)
			resvalid = true
		}
	case 6:
		if respath[2] == "clonemap" && respath[3] == "mas" && cmapErr == nil {
			if respath[4] == "name" {
				cmapErr, httpErr = ams.handleMASName(respath[5], w, r)
				resvalid = true
			} else {
				var masID int
				masID, cmapErr = strconv.Atoi(respath[4])
				if cmapErr == nil {
					if respath[5] == "agents" {
						cmapErr, httpErr = ams.handleAgent(masID, w, r)
						resvalid = true
					} else if respath[5] == "agencies" {
						cmapErr, httpErr = ams.handleAgency(masID, w, r)
						resvalid = true
					}
				}
			}
		}
	case 7:
		var masID int
		masID, cmapErr = strconv.Atoi(respath[4])
		if respath[2] == "clonemap" && respath[3] == "mas" && cmapErr == nil {
			if respath[5] == "agents" {
				var agentID int
				agentID, cmapErr = strconv.Atoi(respath[6])
				if cmapErr == nil {
					cmapErr, httpErr = ams.handleAgentID(masID, agentID, w, r)
					resvalid = true
				}
			}
		}
	case 8:
		var masID int
		masID, cmapErr = strconv.Atoi(respath[4])
		if respath[2] == "clonemap" && respath[3] == "mas" && cmapErr == nil {
			if respath[5] == "agents" {
				if respath[6] == "name" {
					cmapErr, httpErr = ams.handleAgentName(masID, respath[7], w, r)
					resvalid = true
				} else {
					var agentID int
					agentID, cmapErr = strconv.Atoi(respath[6])
					if cmapErr == nil {
						// if respath[7] == "status" {
						// 	cmapErr, httpErr = ams.handleAgentStatus(masID, agentID, w, r)
						// 	resvalid = true
						// } else
						if respath[7] == "address" {
							cmapErr, httpErr = ams.handleAgentAddress(masID, agentID, w, r)
							resvalid = true
						} else if respath[7] == "custom" {
							cmapErr, httpErr = ams.handleAgentCustom(masID, agentID, w, r)
							resvalid = true
						}
					}
				}
			}
		}
	case 9:
		var masID int
		masID, cmapErr = strconv.Atoi(respath[4])
		if respath[2] == "clonemap" && respath[3] == "mas" && cmapErr == nil {
			if respath[5] == "imgroup" && respath[7] == "agency" {
				var imID, agencyID int
				imID, cmapErr = strconv.Atoi(respath[6])
				if cmapErr == nil {
					agencyID, cmapErr = strconv.Atoi(respath[8])
					if cmapErr == nil {
						cmapErr, httpErr = ams.handleAgencyID(masID, imID, agencyID, w, r)
						resvalid = true
					}
				}
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
		ams.logError.Println(respath, cmapErr)
	}
	if httpErr != nil {
		ams.logError.Println(respath, httpErr)
	}
}

// handleAlive is the handler for requests to path /api/alive
func (ams *AMS) handleAlive(w http.ResponseWriter, r *http.Request) (cmapErr, httpErr error) {
	if r.Method == "GET" {
		httpErr = httpreply.Alive(w, nil)
	} else {
		httpErr = httpreply.MethodNotAllowed(w)
		cmapErr = errors.New("Error: Method not allowed on path /api/alive")
	}
	return
}

// handleCloneMAP is the handler for requests to path /api/clonemap
func (ams *AMS) handleCloneMAP(w http.ResponseWriter, r *http.Request) (cmapErr, httpErr error) {
	if r.Method == "GET" {
		// return info about running clonemap instance
		var cmapInfo schemas.CloneMAP
		cmapInfo, cmapErr = ams.getCloneMAPInfo()
		httpErr = httpreply.Resource(w, cmapInfo, cmapErr)
	} else {
		httpErr = httpreply.MethodNotAllowed(w)
		cmapErr = errors.New("Error: Method not allowed on path /api/clonemap")
	}
	return
}

// handleMAS is the handler for requests to path /api/clonemap/mas
func (ams *AMS) handleMAS(w http.ResponseWriter, r *http.Request) (cmapErr, httpErr error) {
	if r.Method == "GET" {
		// return short info of all MAS
		var mass []schemas.MASInfoShort
		mass, cmapErr = ams.getMASsShort()
		httpErr = httpreply.Resource(w, mass, cmapErr)
	} else if r.Method == "POST" {
		// create new MAS
		var body []byte
		body, cmapErr = ioutil.ReadAll(r.Body)
		if cmapErr == nil {
			var masSpec schemas.MASSpec
			cmapErr = json.Unmarshal(body, &masSpec)
			if cmapErr == nil {
				cmapErr = ams.createMAS(masSpec)
				if cmapErr == nil {
					httpErr = httpreply.Created(w, cmapErr, "text/plain",
						[]byte("Ressource Created"))
				} else {
					httpErr = httpreply.CMAPError(w, cmapErr.Error())
				}
			} else {
				httpErr = httpreply.JSONUnmarshalError(w)
			}
		} else {
			httpErr = httpreply.InvalidBodyError(w)
		}
	} else {
		httpErr = httpreply.MethodNotAllowed(w)
		cmapErr = errors.New("Error: Method not allowed on path /api/clonemap/mas")
	}
	return
}

// handlemasID is the handler for requests to path /api/clonemap/mas/{mas-id}
func (ams *AMS) handlemasID(masID int, w http.ResponseWriter, r *http.Request) (cmapErr,
	httpErr error) {
	if r.Method == "GET" {
		// return long information about specified MAS
		var masInfo schemas.MASInfo
		masInfo, cmapErr = ams.getMASInfo(masID)
		httpErr = httpreply.Resource(w, masInfo, cmapErr)
	} else if r.Method == "DELETE" {
		// delete specified MAS
		cmapErr = ams.removeMAS(masID)
		httpErr = httpreply.Deleted(w, cmapErr)
	} else {
		httpErr = httpreply.MethodNotAllowed(w)
		cmapErr = errors.New("Error: Method not allowed on path /api/clonemap/mas/{mas-id}")
	}
	return
}

// handleMASName is the handler for requests to path /api/clonemap/mas/name/{name}
func (ams *AMS) handleMASName(name string, w http.ResponseWriter, r *http.Request) (cmapErr,
	httpErr error) {
	if r.Method == "GET" {
		// search for MAS with matching name
		// ToDo
	} else {
		httpErr = httpreply.MethodNotAllowed(w)
		cmapErr = errors.New("Error: Method not allowed on path /api/clonemap/name/{name}")
	}
	return
}

// handleAgent is the handler for requests to path /api/clonemap/mas/{mas-id}/agents
func (ams *AMS) handleAgent(masID int, w http.ResponseWriter, r *http.Request) (cmapErr,
	httpErr error) {
	if r.Method == "GET" {
		// return short information of all agents in specified MAS
		var agents schemas.Agents
		agents, cmapErr = ams.getAgents(masID)
		httpErr = httpreply.Resource(w, agents, cmapErr)
	} else if r.Method == "POST" {
		// create new agent in MAS
		var body []byte
		body, cmapErr = ioutil.ReadAll(r.Body)
		if cmapErr == nil {
			var groupSpecs []schemas.ImageGroupSpec
			cmapErr = json.Unmarshal(body, &groupSpecs)
			if cmapErr == nil {
				cmapErr = ams.createAgents(masID, groupSpecs)
				httpErr = httpreply.Created(w, cmapErr, "text/plain", []byte("Ressource Created"))
			} else {
				httpErr = httpreply.JSONUnmarshalError(w)
			}
		} else {
			httpErr = httpreply.InvalidBodyError(w)
		}
	} else {
		httpErr = httpreply.MethodNotAllowed(w)
		cmapErr = errors.New("Error: Method not allowed on path /api/clonemap/mas/{mas-id}/agents")
	}
	return
}

// handleAgentID is the handler for requests to path /api/clonemap/mas/{mas-id}/agents/{agent-id}
func (ams *AMS) handleAgentID(masID int, agentid int, w http.ResponseWriter,
	r *http.Request) (cmapErr, httpErr error) {
	if r.Method == "GET" {
		// return long information of specified agent
		var agentInfo schemas.AgentInfo
		agentInfo, cmapErr = ams.getAgentInfo(masID, agentid)
		httpErr = httpreply.Resource(w, agentInfo, cmapErr)
	} else if r.Method == "DELETE" {
		// delete specified agent
		cmapErr = ams.removeAgent(masID, agentid)
		httpErr = httpreply.Deleted(w, cmapErr)

	} else {
		httpErr = httpreply.MethodNotAllowed(w)
		cmapErr = errors.New("Error: Method not allowed on path /api/clonemap/mas/{mas-id}/agents/" +
			"{agent-id}")
	}
	return
}

// handleAgentAddress is the handler for requests to path
// /api/clonemap/mas/{mas-id}/agents/{agent-id}/address
func (ams *AMS) handleAgentAddress(masID int, agentid int, w http.ResponseWriter,
	r *http.Request) (cmapErr, httpErr error) {
	if r.Method == "GET" {
		// return address of specified agent
		var agentAddr schemas.Address
		agentAddr, cmapErr = ams.getAgentAddress(masID, agentid)
		httpErr = httpreply.Resource(w, agentAddr, cmapErr)
	} else if r.Method == "PUT" {
		// update address of specified agent
		var body []byte
		body, cmapErr = ioutil.ReadAll(r.Body)
		if cmapErr == nil {
			var agentAddr schemas.Address
			cmapErr := json.Unmarshal(body, &agentAddr)
			if cmapErr == nil {
				cmapErr = ams.updateAgentAddress(masID, agentid, agentAddr)
				httpErr = httpreply.Updated(w, cmapErr)
			} else {
				httpErr = httpreply.JSONUnmarshalError(w)
			}
		} else {
			httpErr = httpreply.InvalidBodyError(w)
		}
	} else {
		httpErr = httpreply.MethodNotAllowed(w)
		cmapErr = errors.New("Error: Method not allowed on path /api/clonemap/mas/{mas-id}/agent/" +
			"{agent-id}/address")
	}
	return
}

// handleAgentAddress is the handler for requests to path
// /api/clonemap/mas/{mas-id}/agents/{agent-id}/custom
func (ams *AMS) handleAgentCustom(masID int, agentid int, w http.ResponseWriter,
	r *http.Request) (cmapErr, httpErr error) {
	if r.Method == "PUT" {
		// update custom of specified agent
		var body []byte
		body, cmapErr = ioutil.ReadAll(r.Body)
		if cmapErr == nil {
			custom := string(body)
			cmapErr = ams.updateAgentCustom(masID, agentid, custom)
			httpErr = httpreply.Updated(w, cmapErr)
		} else {
			httpErr = httpreply.InvalidBodyError(w)
		}
	} else {
		httpErr = httpreply.MethodNotAllowed(w)
		cmapErr = errors.New("Error: Method not allowed on path /api/clonemap/mas/{mas-id}/agent/" +
			"{agent-id}/custom")
	}
	return
}

// handleAgentName is the handler for requests to path /api/clonemap/mas/{masid}/agents/name/{name}
func (ams *AMS) handleAgentName(masID int, name string, w http.ResponseWriter,
	r *http.Request) (cmapErr, httpErr error) {
	if r.Method == "GET" {
		// search for agents with matching name
		// ToDo
	} else {
		httpErr = httpreply.MethodNotAllowed(w)
		cmapErr = errors.New("Error: Method not allowed on path /api/clonemap/mas/{masid}/agents/" +
			"name/{name}")
	}
	return
}

// handleAgency is the handler for requests to path /api/cloumap/mas/{mas-id}/agencies
func (ams *AMS) handleAgency(masID int, w http.ResponseWriter, r *http.Request) (cmapErr,
	httpErr error) {
	if r.Method == "GET" {
		// return information of specified agency
		var agencies schemas.Agencies
		agencies, cmapErr = ams.getAgencies(masID)
		httpErr = httpreply.Resource(w, agencies, cmapErr)
	} else {
		httpreply.MethodNotAllowed(w)
		cmapErr = errors.New("Error: Method not allowed on path /api/clonemap/mas/{mas-id}/" +
			"agencies/")
	}
	return
}

// handleAgencyID is the handler for requests to path /api/clonemap/mas/{mas-id}/imgroup/{imID}/agencies/{agency-id}
func (ams *AMS) handleAgencyID(masID int, imID int, agencyid int, w http.ResponseWriter,
	r *http.Request) (cmapErr, httpErr error) {
	if r.Method == "GET" {
		var agencySpec schemas.AgencyInfoFull
		agencySpec, cmapErr = ams.getAgencyInfoFull(masID, imID, agencyid)
		httpErr = httpreply.Resource(w, agencySpec, cmapErr)
	} else {
		httpErr = httpreply.MethodNotAllowed(w)
		cmapErr = errors.New("Error: Method not allowed on path /api/clonemap/mas/{mas-id}/" +
			"agencies/{agency-id}")
	}
	return
}

// listen opens a http server listening and serving request
func (ams *AMS) listen() (err error) {
	mux := http.NewServeMux()
	mux.HandleFunc("/api/", ams.handleAPI)
	s := &http.Server{
		Addr:    ":9000",
		Handler: mux,
	}
	ams.logInfo.Println("AMS listening on port 9000")
	err = s.ListenAndServe()
	return
}
