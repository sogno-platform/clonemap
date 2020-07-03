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
	"errors"
	"io/ioutil"
	"net/http"

	amsclient "git.rwth-aachen.de/acs/public/cloud/mas/clonemap/pkg/ams/client"
	"git.rwth-aachen.de/acs/public/cloud/mas/clonemap/pkg/common/httpreply"
	"git.rwth-aachen.de/acs/public/cloud/mas/clonemap/pkg/schemas"
)

// handleMAS is the handler for requests to path /api/ams/mas
func (fe *Frontend) handleMAS(w http.ResponseWriter, r *http.Request) (cmapErr, httpErr error) {
	if r.Method == "GET" {
		// return short info of all MAS
		var mass []schemas.MASInfoShort
		mass, _, cmapErr = amsclient.GetMASsShort()
		if cmapErr == nil {
			httpErr = httpreply.Resource(w, mass, cmapErr)
		} else {
			httpErr = httpreply.CMAPError(w, cmapErr.Error())
		}
	} else if r.Method == "POST" {

	} else {
		httpErr = httpreply.MethodNotAllowed(w)
		cmapErr = errors.New("Error: Method not allowed on path /api/ams/mas")
	}
	return
}

// handlemasID is the handler for requests to path /api/ams/mas/{mas-id}
func (fe *Frontend) handlemasID(masID int, w http.ResponseWriter, r *http.Request) (cmapErr,
	httpErr error) {
	if r.Method == "GET" {
		// return long information about specified MAS
		var masInfo schemas.MASInfo
		masInfo, _, cmapErr = amsclient.GetMAS(masID)
		httpErr = httpreply.Resource(w, masInfo, cmapErr)
	} else if r.Method == "DELETE" {
		// delete specified MAS

	} else {
		httpErr = httpreply.MethodNotAllowed(w)
		cmapErr = errors.New("Error: Method not allowed on path /api/ams/mas/{mas-id}")
	}
	return
}

// handleAgent is the handler for requests to path /api/clonemap/mas/{mas-id}/agents
func (fe *Frontend) handleAgent(masID int, w http.ResponseWriter, r *http.Request) (cmapErr,
	httpErr error) {
	if r.Method == "POST" {
		// create new agent in MAS
		var body []byte
		body, cmapErr = ioutil.ReadAll(r.Body)
		if cmapErr == nil {
			var groupSpecs []schemas.ImageGroupSpec
			cmapErr = json.Unmarshal(body, &groupSpecs)
			if cmapErr == nil {
				_, cmapErr = amsclient.PostAgents(masID, groupSpecs)
				httpErr = httpreply.Created(w, cmapErr, "text/plain", []byte("Ressource Created"))
			} else {
				httpErr = httpreply.JSONUnmarshalError(w)
			}
		} else {
			httpErr = httpreply.InvalidBodyError(w)
		}
	} else {
		httpErr = httpreply.MethodNotAllowed(w)
		cmapErr = errors.New("Error: Method not allowed on path /api/ams/mas/{mas-id}/agents")
	}
	return
}

// handleAgentID is the handler for requests to path /api/ams/mas/{mas-id}/agents/{agent-id}
func (fe *Frontend) handleAgentID(masID int, agentID int, w http.ResponseWriter,
	r *http.Request) (cmapErr, httpErr error) {
	if r.Method == "GET" {
		// return long information of specified agent
		var agentInfo schemas.AgentInfo
		agentInfo, _, cmapErr = amsclient.GetAgent(masID, agentID)
		httpErr = httpreply.Resource(w, agentInfo, cmapErr)
	} else if r.Method == "DELETE" {
		// delete specified agent
		_, cmapErr = amsclient.DeleteAgent(masID, agentID)
		httpErr = httpreply.Deleted(w, cmapErr)
	} else {
		httpErr = httpreply.MethodNotAllowed(w)
		cmapErr = errors.New("Error: Method not allowed on path /api/ams/mas/{mas-id}/agents/" +
			"{agent-id}")
	}
	return
}
