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
	"time"

	"git.rwth-aachen.de/acs/public/cloud/mas/clonemap/pkg/common/httpreply"
	"git.rwth-aachen.de/acs/public/cloud/mas/clonemap/pkg/schemas"
	"github.com/gorilla/mux"
)

type timeSlice []time.Time

func (s timeSlice) Less(i, j int) bool {
	return s[i].After(s[j])
}

func (s timeSlice) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}

func (s timeSlice) Len() int {
	return len(s)
}

// LogToSend contains data of a single agent log message
type LogToSend struct {
	Timestamp       time.Time `json:"timestamp"`       // time of message
	ScaledTimestamp float64   `json:"scaledTimestamp"` // scaled time for drawing
	Message         string    `json:"msg"`             // log message
	AdditionalData  string    `json:"data,omitempty"`  // additional information e.g in json
}

// handlePostLogs is the handler to /api/logging/{masid}/list
func (fe *Frontend) handlePostLogs(w http.ResponseWriter, r *http.Request) {
	var cmapErr, httpErr error
	vars := mux.Vars(r)
	masID, cmapErr := strconv.Atoi(vars["masid"])

	if cmapErr != nil {
		httpErr = httpreply.NotFoundError(w)
		fe.logErrors(r.URL.Path, cmapErr, httpErr)
		return
	}

	// create new log
	var body []byte
	body, cmapErr = ioutil.ReadAll(r.Body)
	if cmapErr != nil {
		httpErr = httpreply.InvalidBodyError(w)
		fe.logErrors(r.URL.Path, cmapErr, httpErr)
		return
	}
	var msgs []schemas.LogMessage
	cmapErr = json.Unmarshal(body, &msgs)
	if cmapErr != nil {
		httpErr = httpreply.JSONUnmarshalError(w)
		fe.logErrors(r.URL.Path, cmapErr, httpErr)
		return
	}
	_, cmapErr = fe.logClient.PostLogs(masID, msgs)
	if cmapErr != nil {
		httpErr = httpreply.CMAPError(w, cmapErr.Error())
		fe.logErrors(r.URL.Path, cmapErr, httpErr)
		return
	}
	httpErr = httpreply.Created(w, cmapErr, "text/plain", []byte("Resource Created"))
	fe.logErrors(r.URL.Path, cmapErr, httpErr)
	return
}

// handleGetLogs is the handler to /api/logging/{masid}/{agentid}/{topic}/latest/{num}
func (fe *Frontend) handleGetNLatestLogs(w http.ResponseWriter, r *http.Request) {
	var cmapErr, httpErr error
	masid, agentid, topic, num, cmapErr := getNLogs(r)
	if cmapErr != nil {
		httpErr = httpreply.NotFoundError(w)
		fe.logErrors(r.URL.Path, cmapErr, httpErr)
		return
	}

	var msgs []schemas.LogMessage
	msgs, _, cmapErr = fe.logClient.GetLatestLogs(masid, agentid, topic, num)
	if cmapErr != nil {
		httpErr = httpreply.CMAPError(w, cmapErr.Error())
		fe.logErrors(r.URL.Path, cmapErr, httpErr)
		return
	}
	httpErr = httpreply.Resource(w, msgs, cmapErr)
	fe.logErrors(r.URL.Path, cmapErr, httpErr)
	return
}

// handleGetLogsTime is the handler to /api/logging/{masid}/{agentid}/{topic}/time/{start}/{end}
func (fe *Frontend) handleGetLogsInRange(w http.ResponseWriter, r *http.Request) {
	var cmapErr, httpErr error
	masid, agentid, cmapErr := getAgentID(r)
	vars := mux.Vars(r)
	topic := vars["topic"]
	start := vars["start"]
	end := vars["end"]
	if cmapErr != nil {
		httpErr = httpreply.NotFoundError(w)
		fe.logErrors(r.URL.Path, cmapErr, httpErr)
		return
	}

	var msgs []schemas.LogMessage
	msgs, _, cmapErr = fe.logClient.GetLogsInRange(masid, agentid, topic, start, end)
	if cmapErr != nil {
		httpErr = httpreply.CMAPError(w, cmapErr.Error())
		fe.logErrors(r.URL.Path, cmapErr, httpErr)
		return
	}

	httpErr = httpreply.Resource(w, msgs, cmapErr)
	fe.logErrors(r.URL.Path, cmapErr, httpErr)
	return
}

// handleGetLogSeriesNames is the handler to api/logging/series/{masid}/{agentid}/names
func (fe *Frontend) handleGetLogSeriesNames(w http.ResponseWriter, r *http.Request) {
	var cmapErr, httpErr error
	masID, agentID, cmapErr := getAgentID(r)
	if cmapErr != nil {
		httpErr = httpreply.NotFoundError(w)
		fe.logErrors(r.URL.Path, cmapErr, httpErr)
		return
	}
	var names []string
	names, _, cmapErr = fe.logClient.GetLogSeriesNames(masID, agentID)
	if cmapErr != nil {
		httpErr = httpreply.CMAPError(w, cmapErr.Error())
		fe.logErrors(r.URL.Path, cmapErr, httpErr)
		return
	}
	httpErr = httpreply.Resource(w, names, cmapErr)
	fe.logErrors(r.URL.Path, cmapErr, httpErr)
	return

}

// handleGetLogSeriesByName is the handler to /api/logging/series/{masid}/{agentid}/{name}/time/{start}/{end}
func (fe *Frontend) handleGetLogSeriesByName(w http.ResponseWriter, r *http.Request) {
	var cmapErr, httpErr error
	masID, agentID, cmapErr := getAgentID(r)

	if cmapErr != nil {
		httpErr = httpreply.NotFoundError(w)
		fe.logErrors(r.URL.Path, cmapErr, httpErr)
		return
	}
	vars := mux.Vars(r)
	name := vars["name"]
	start := vars["start"]
	end := vars["end"]
	var series []schemas.LogSeries
	series, _, cmapErr = fe.logClient.GetLogSeriesByName(masID, agentID, name, start, end)
	if cmapErr != nil {
		httpErr = httpreply.CMAPError(w, cmapErr.Error())
		fe.logErrors(r.URL.Path, cmapErr, httpErr)
		return
	}
	httpErr = httpreply.Resource(w, series, cmapErr)
	fe.logErrors(r.URL.Path, cmapErr, httpErr)
	return
}
