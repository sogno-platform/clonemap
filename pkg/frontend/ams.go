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

// handleGetMASs is the handler for get requests to path /api/ams/mas
func (fe *Frontend) handleGetMASs(w http.ResponseWriter, r *http.Request) {
	var cmapErr, httpErr error
	// return short info of all MAS
	var mass []schemas.MASInfoShort
	mass, _, cmapErr = fe.amsClient.GetMASsShort()
	if cmapErr != nil {
		httpErr = httpreply.CMAPError(w, cmapErr.Error())
		fe.logErrors(r.URL.Path, cmapErr, httpErr)
		return
	}
	httpErr = httpreply.Resource(w, mass, cmapErr)
	fe.logErrors(r.URL.Path, cmapErr, httpErr)
	return
}

// handlePostMASs is the handler for post requests to path /api/ams/mas
func (fe *Frontend) handlePostMAS(w http.ResponseWriter, r *http.Request) {
	var cmapErr, httpErr error
	var body []byte
	body, cmapErr = ioutil.ReadAll(r.Body)
	if cmapErr != nil {
		httpErr = httpreply.InvalidBodyError(w)
		fe.logErrors(r.URL.Path, cmapErr, httpErr)
		return
	}
	var masSpec schemas.MASSpec
	cmapErr = json.Unmarshal(body, &masSpec)
	if cmapErr != nil {
		httpErr = httpreply.JSONUnmarshalError(w)
		fe.logErrors(r.URL.Path, cmapErr, httpErr)
		return
	}
	_, httpErr = fe.amsClient.PostMAS(masSpec)
	if httpErr != nil {
		httpErr = httpreply.CMAPError(w, cmapErr.Error())
		fe.logErrors(r.URL.Path, cmapErr, httpErr)
		return
	}
	httpErr = httpreply.Created(w, cmapErr, "text/plain", []byte("Ressource Created"))
	fe.logErrors(r.URL.Path, cmapErr, httpErr)
	return
}

// handleGetMASID is the handler for get requests to path /api/ams/mas/{masid}
func (fe *Frontend) handleGetMASID(w http.ResponseWriter, r *http.Request) {
	var cmapErr, httpErr error
	vars := mux.Vars(r)
	masID, cmapErr := strconv.Atoi(vars["masid"])
	if cmapErr != nil {
		httpErr = httpreply.NotFoundError(w)
		fe.logErrors(r.URL.Path, cmapErr, httpErr)
		return
	}
	// return long information about specified MAS
	var masInfo schemas.MASInfo
	masInfo, _, cmapErr = fe.amsClient.GetMAS(masID)
	if cmapErr != nil {
		httpErr = httpreply.CMAPError(w, cmapErr.Error())
		fe.logErrors(r.URL.Path, cmapErr, httpErr)
		return
	}
	httpErr = httpreply.Resource(w, masInfo, cmapErr)
	fe.logErrors(r.URL.Path, cmapErr, httpErr)
	return
}

// handleDeleteMASID is the handler for delete requests to path /api/ams/mas/{masid}
func (fe *Frontend) handleDeleteMASID(w http.ResponseWriter, r *http.Request) {
	return
}

// handlePostAgent is the handler for post requests to path /api/clonemap/mas/{masid}/agents
func (fe *Frontend) handlePostAgent(w http.ResponseWriter, r *http.Request) {
	var cmapErr, httpErr error
	vars := mux.Vars(r)
	masID, cmapErr := strconv.Atoi(vars["masid"])
	if cmapErr != nil {
		httpErr = httpreply.NotFoundError(w)
		fe.logErrors(r.URL.Path, cmapErr, httpErr)
		return
	}
	// create new agent in MAS
	var body []byte
	body, cmapErr = ioutil.ReadAll(r.Body)
	if cmapErr != nil {
		httpErr = httpreply.InvalidBodyError(w)
		fe.logErrors(r.URL.Path, cmapErr, httpErr)
		return
	}
	var groupSpecs []schemas.ImageGroupSpec
	cmapErr = json.Unmarshal(body, &groupSpecs)
	if cmapErr != nil {
		httpErr = httpreply.JSONUnmarshalError(w)
		fe.logErrors(r.URL.Path, cmapErr, httpErr)
		return
	}
	_, cmapErr = fe.amsClient.PostAgents(masID, groupSpecs)
	if cmapErr != nil {
		httpErr = httpreply.CMAPError(w, cmapErr.Error())
		fe.logErrors(r.URL.Path, cmapErr, httpErr)
		return
	}
	httpErr = httpreply.Created(w, cmapErr, "text/plain", []byte("Ressource Created"))
	fe.logErrors(r.URL.Path, cmapErr, httpErr)
	return
}

// handleGetAgentID is the handler for get requests to path /api/ams/mas/{masid}/agents/{agentid}
func (fe *Frontend) handleGetAgentID(w http.ResponseWriter, r *http.Request) {
	var cmapErr, httpErr error
	masID, agentID, cmapErr := getAgentID(r)
	if cmapErr != nil {
		httpErr = httpreply.NotFoundError(w)
		fe.logErrors(r.URL.Path, cmapErr, httpErr)
		return
	}
	// return long information of specified agent
	var agentInfo schemas.AgentInfo
	agentInfo, _, cmapErr = fe.amsClient.GetAgent(masID, agentID)
	if cmapErr != nil {
		httpErr = httpreply.CMAPError(w, cmapErr.Error())
		fe.logErrors(r.URL.Path, cmapErr, httpErr)
		return
	}
	httpErr = httpreply.Resource(w, agentInfo, cmapErr)
	fe.logErrors(r.URL.Path, cmapErr, httpErr)
	return
}

// handleDeleteAgentID is the handler for delete requests to path
// /api/ams/mas/{masid}/agents/{agentid}
func (fe *Frontend) handleDeleteAgentID(w http.ResponseWriter, r *http.Request) {
	var cmapErr, httpErr error
	masID, agentID, cmapErr := getAgentID(r)
	if cmapErr != nil {
		httpErr = httpreply.NotFoundError(w)
		fe.logErrors(r.URL.Path, cmapErr, httpErr)
		return
	}
	// delete specified agent
	_, cmapErr = fe.amsClient.DeleteAgent(masID, agentID)
	if cmapErr != nil {
		httpErr = httpreply.CMAPError(w, cmapErr.Error())
		fe.logErrors(r.URL.Path, cmapErr, httpErr)
		return
	}
	httpErr = httpreply.Deleted(w, cmapErr)
	fe.logErrors(r.URL.Path, cmapErr, httpErr)
	return
}
