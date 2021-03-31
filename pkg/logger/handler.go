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

package logger

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"net/http"
	"strconv"
	"time"

	"git.rwth-aachen.de/acs/public/cloud/mas/clonemap/pkg/common/httpreply"
	"git.rwth-aachen.de/acs/public/cloud/mas/clonemap/pkg/schemas"
	"github.com/gorilla/mux"
)

// handleAlive is the handler for requests to path /api/alive
func (logger *Logger) handleAlive(w http.ResponseWriter, r *http.Request) {
	var httpErr error
	httpErr = httpreply.Alive(w, nil)
	logger.logErrors(r.URL.Path, nil, httpErr)
	return
}

// handlePostLogMsg is the handler for post requests to path
// /api/logging/{masid}/{agentid}/{topic}
func (logger *Logger) handlePostLogMsg(w http.ResponseWriter, r *http.Request) {
	var cmapErr, httpErr error
	// create new log message entry
	var body []byte
	body, cmapErr = ioutil.ReadAll(r.Body)
	if cmapErr != nil {
		httpErr = httpreply.InvalidBodyError(w)
		logger.logErrors(r.URL.Path, cmapErr, httpErr)
		return
	}
	var logmsg schemas.LogMessage
	cmapErr = json.Unmarshal(body, &logmsg)
	if cmapErr != nil {
		httpErr = httpreply.JSONUnmarshalError(w)
		logger.logErrors(r.URL.Path, cmapErr, httpErr)
		return
	}
	go logger.addAgentLogMessage(logmsg)
	httpErr = httpreply.Created(w, nil, "text/plain", []byte("Ressource Created"))
	logger.logErrors(r.URL.Path, cmapErr, httpErr)
	return
}

// handleGetCommunication is the handler for get requests to path
// /api/logging/{masid}/{agentid}/comm
func (logger *Logger) handleGetCommunication(w http.ResponseWriter, r *http.Request) {
	var cmapErr, httpErr error
	masID, agentID, cmapErr := getAgentID(r)
	if cmapErr != nil {
		httpErr = httpreply.NotFoundError(w)
		logger.logErrors(r.URL.Path, cmapErr, httpErr)
		return
	}
	var comm []schemas.Communication
	comm, cmapErr = logger.getCommunication(masID, agentID)
	if cmapErr != nil {
		httpErr = httpreply.CMAPError(w, cmapErr.Error())
		logger.logErrors(r.URL.Path, cmapErr, httpErr)
		return
	}
	httpErr = httpreply.Resource(w, comm, cmapErr)
	logger.logErrors(r.URL.Path, cmapErr, httpErr)
	return
}

// handlePutCommunication is the handler for put requests to path
// /api/logging/{masid}/{agentid}/comm
func (logger *Logger) handlePutCommunication(w http.ResponseWriter, r *http.Request) {
	var cmapErr, httpErr error
	masID, agentID, cmapErr := getAgentID(r)
	if cmapErr != nil {
		httpErr = httpreply.NotFoundError(w)
		logger.logErrors(r.URL.Path, cmapErr, httpErr)
		return
	}
	// update communication data
	var body []byte
	body, cmapErr = ioutil.ReadAll(r.Body)
	if cmapErr != nil {
		httpErr = httpreply.InvalidBodyError(w)
		logger.logErrors(r.URL.Path, cmapErr, httpErr)
		return
	}
	var comm []schemas.Communication
	cmapErr = json.Unmarshal(body, &comm)
	if cmapErr != nil {
		httpErr = httpreply.JSONUnmarshalError(w)
		logger.logErrors(r.URL.Path, cmapErr, httpErr)
		return
	}
	go logger.updateCommunication(masID, agentID, comm)
	httpErr = httpreply.Updated(w, nil)
	logger.logErrors(r.URL.Path, cmapErr, httpErr)
	return
}

// handlePostLogMsgList is the handler for post requests to path /api/logging/{masid}/list
func (logger *Logger) handlePostLogMsgList(w http.ResponseWriter, r *http.Request) {
	var cmapErr, httpErr error
	// create new log message entry
	var body []byte
	body, cmapErr = ioutil.ReadAll(r.Body)
	if cmapErr != nil {
		httpErr = httpreply.NotFoundError(w)
		logger.logErrors(r.URL.Path, cmapErr, httpErr)
		return
	}
	var logmsg []schemas.LogMessage
	cmapErr = json.Unmarshal(body, &logmsg)
	if cmapErr != nil {
		httpErr = httpreply.JSONUnmarshalError(w)
		logger.logErrors(r.URL.Path, cmapErr, httpErr)
		return
	}
	go logger.addAgentLogMessageList(logmsg)
	httpErr = httpreply.Created(w, nil, "text/plain", []byte("Ressource Created"))
	logger.logErrors(r.URL.Path, cmapErr, httpErr)
	return
}

// handleGetAllLatestLogMessages is the handler for request to path
// /api/logging/{masid}/latest/{num}
func (logger *Logger) handleGetAllLatestLogMessages(w http.ResponseWriter, r *http.Request) {
	var cmapErr, httpErr error
	masID, num, cmapErr := getMasAndNum(r)
	if cmapErr != nil {
		httpErr = httpreply.NotFoundError(w)
		logger.logErrors(r.URL.Path, cmapErr, httpErr)
		return
	}
	var logMsgs []schemas.LogMessage
	logMsgs, cmapErr = logger.getAllLatestLogMessages(masID, num)
	if cmapErr != nil {
		httpErr = httpreply.CMAPError(w, cmapErr.Error())
		logger.logErrors(r.URL.Path, cmapErr, httpErr)
		return
	}
	httpErr = httpreply.Resource(w, logMsgs, cmapErr)
	logger.logErrors(r.URL.Path, cmapErr, httpErr)
	return

}

// handleGetLogsLatest is the handler for requests to path
// /api/logging/{masid}/{agentid}/{topic}/latest/{num}
func (logger *Logger) handleGetLogsLatest(w http.ResponseWriter, r *http.Request) {
	var cmapErr, httpErr error
	masID, agentID, cmapErr := getAgentID(r)
	if cmapErr != nil {
		httpErr = httpreply.NotFoundError(w)
		logger.logErrors(r.URL.Path, cmapErr, httpErr)
		return
	}
	vars := mux.Vars(r)
	topic := vars["topic"]
	num, cmapErr := strconv.Atoi(vars["num"])
	if cmapErr != nil {
		httpErr = httpreply.InvalidBodyError(w)
		logger.logErrors(r.URL.Path, cmapErr, httpErr)
		return
	}
	var logMsg []schemas.LogMessage
	logMsg, cmapErr = logger.getLatestAgentLogMessages(masID, agentID, topic, num)
	if cmapErr != nil {
		httpErr = httpreply.CMAPError(w, cmapErr.Error())
		logger.logErrors(r.URL.Path, cmapErr, httpErr)
		return
	}
	httpErr = httpreply.Resource(w, logMsg, cmapErr)
	logger.logErrors(r.URL.Path, cmapErr, httpErr)
	return
}

// handleGetLogsTime is the handler for get requests to path
// /api/logging/{masid}/{agentid}/{topic}/time/{start}/{end}
func (logger *Logger) handleGetLogsTime(w http.ResponseWriter, r *http.Request) {
	var cmapErr, httpErr error
	masID, agentID, cmapErr := getAgentID(r)
	if cmapErr != nil {
		httpErr = httpreply.NotFoundError(w)
		logger.logErrors(r.URL.Path, cmapErr, httpErr)
		return
	}
	vars := mux.Vars(r)
	topic := vars["topic"]
	// start, cmapErr := time.Parse(time.RFC3339, vars["start"])
	start, cmapErr := time.Parse("20060102150405", vars["start"])
	if cmapErr != nil {
		httpErr = httpreply.NotFoundError(w)
		logger.logErrors(r.URL.Path, cmapErr, httpErr)
		return
	}
	end, cmapErr := time.Parse("20060102150405", vars["end"])
	if cmapErr != nil {
		httpErr = httpreply.NotFoundError(w)
		logger.logErrors(r.URL.Path, cmapErr, httpErr)
		return
	}
	var logMsg []schemas.LogMessage
	logMsg, cmapErr = logger.getAgentLogMessagesInRange(masID, agentID, topic, start, end)
	if cmapErr != nil {
		httpErr = httpreply.CMAPError(w, cmapErr.Error())
		logger.logErrors(r.URL.Path, cmapErr, httpErr)
		return
	}
	httpErr = httpreply.Resource(w, logMsg, cmapErr)
	logger.logErrors(r.URL.Path, cmapErr, httpErr)
	return
}

// handleGetState is the handler for get requests to path /api/state/{masid}/{agentid}
func (logger *Logger) handleGetState(w http.ResponseWriter, r *http.Request) {
	var cmapErr, httpErr error
	masID, agentID, cmapErr := getAgentID(r)
	if cmapErr != nil {
		httpErr = httpreply.NotFoundError(w)
		logger.logErrors(r.URL.Path, cmapErr, httpErr)
		return
	}
	var state schemas.State
	state, cmapErr = logger.getAgentState(masID, agentID)
	if cmapErr != nil {
		httpErr = httpreply.CMAPError(w, cmapErr.Error())
		logger.logErrors(r.URL.Path, cmapErr, httpErr)
		return
	}
	httpErr = httpreply.Resource(w, state, cmapErr)
	logger.logErrors(r.URL.Path, cmapErr, httpErr)
	return
}

// handlePutState is the handler for put requests to path /api/state/{masid}/{agentid}
func (logger *Logger) handlePutState(w http.ResponseWriter, r *http.Request) {
	var cmapErr, httpErr error
	masID, agentID, cmapErr := getAgentID(r)
	if cmapErr != nil {
		httpErr = httpreply.NotFoundError(w)
		logger.logErrors(r.URL.Path, cmapErr, httpErr)
		return
	}
	var body []byte
	body, cmapErr = ioutil.ReadAll(r.Body)
	if cmapErr != nil {
		httpErr = httpreply.InvalidBodyError(w)
		logger.logErrors(r.URL.Path, cmapErr, httpErr)
		return
	}
	var state schemas.State
	cmapErr = json.Unmarshal(body, &state)
	if cmapErr != nil {
		httpErr = httpreply.JSONUnmarshalError(w)
		logger.logErrors(r.URL.Path, cmapErr, httpErr)
		return
	}
	go logger.updateAgentState(masID, agentID, state)
	httpErr = httpreply.Updated(w, nil)
	logger.logErrors(r.URL.Path, cmapErr, httpErr)
	return
}

// handlePutStateList is the handler for requests to path /api/state/{masid}/list
func (logger *Logger) handlePutStateList(w http.ResponseWriter, r *http.Request) {
	var cmapErr, httpErr error
	vars := mux.Vars(r)
	masID, cmapErr := strconv.Atoi(vars["masid"])
	if cmapErr != nil {
		httpErr = httpreply.NotFoundError(w)
		logger.logErrors(r.URL.Path, cmapErr, httpErr)
		return
	}
	var body []byte
	body, cmapErr = ioutil.ReadAll(r.Body)
	if cmapErr != nil {
		httpErr = httpreply.InvalidBodyError(w)
		logger.logErrors(r.URL.Path, cmapErr, httpErr)
		return
	}
	var states []schemas.State
	cmapErr = json.Unmarshal(body, &states)
	if cmapErr != nil {
		httpErr = httpreply.JSONUnmarshalError(w)
		logger.logErrors(r.URL.Path, cmapErr, httpErr)
		return
	}
	go logger.updateAgentStatesList(masID, states)
	httpErr = httpreply.Updated(w, nil)
	logger.logErrors(r.URL.Path, cmapErr, httpErr)
	return
}

// methodNotAllowed is the default handler for valid paths but invalid methods
func (logger *Logger) methodNotAllowed(w http.ResponseWriter, r *http.Request) {
	httpErr := httpreply.MethodNotAllowed(w)
	cmapErr := errors.New("Error: Method not allowed on path " + r.URL.Path)
	logger.logErrors(r.URL.Path, cmapErr, httpErr)
	return
}

// resourceNotFound is the default handler for invalid paths
func (logger *Logger) resourceNotFound(w http.ResponseWriter, r *http.Request) {
	httpErr := httpreply.NotFoundError(w)
	cmapErr := errors.New("Resource not found")
	logger.logErrors(r.URL.Path, cmapErr, httpErr)
	return
}

// logErrors logs errors if any
func (logger *Logger) logErrors(path string, cmapErr error, httpErr error) {
	if cmapErr != nil {
		logger.logError.Println(path, cmapErr)
	}
	if httpErr != nil {
		logger.logError.Println(path, httpErr)
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

// getMasAnd returns the masID and num of logs from the path
func getMasAndNum(r *http.Request) (masID int, num int, err error) {
	vars := mux.Vars(r)
	masID, err = strconv.Atoi(vars["masid"])
	if err != nil {
		return
	}
	num, err = strconv.Atoi(vars["num"])
	if err != nil {
		return
	}
	return
}

// loggingMiddleware logs request before calling final handler
func (logger *Logger) loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		logger.logInfo.Println("Received Request: ", r.Method, " ", r.URL.EscapedPath())
		next.ServeHTTP(w, r)
	})
}

// server creates the logger server
func (logger *Logger) server(port int) (serv *http.Server) {
	r := mux.NewRouter()
	// r.HandleFunc("/api/", logger.handleAPI)
	s := r.PathPrefix("/api").Subrouter()
	s.Path("/alive").Methods("GET").HandlerFunc(logger.handleAlive)
	s.Path("/logging/{masid}/latest/{num}").Methods("GET").
		HandlerFunc(logger.handleGetAllLatestLogMessages)
	s.Path("/logging/{masid}/{agentid}/{topic}").Methods("POST").HandlerFunc(logger.handlePostLogMsg)
	s.Path("/logging/{masid}/{agentid}/{topic}").Methods("PUT", "GET", "DELETE").
		HandlerFunc(logger.methodNotAllowed)
	s.Path("/logging/{masid}/{agentid}/comm").Methods("GET").
		HandlerFunc(logger.handleGetCommunication)
	s.Path("/logging/{masid}/{agentid}/comm").Methods("PUT").
		HandlerFunc(logger.handlePutCommunication)
	s.Path("/logging/{masid}/{agentid}/comm").Methods("POST", "DELETE").
		HandlerFunc(logger.methodNotAllowed)
	s.Path("/logging/{masid}/list").Methods("POST").HandlerFunc(logger.handlePostLogMsgList)
	s.Path("/logging/{masid}/list").Methods("PUT", "GET", "DELETE").
		HandlerFunc(logger.methodNotAllowed)

	s.Path("/logging/{masid}/{agentid}/{topic}/latest/{num}").Methods("GET").
		HandlerFunc(logger.handleGetLogsLatest)
	s.Path("/logging/{masid}/{agentid}/{topic}/latest/{num}").Methods("POST", "PUT", "DELETE").
		HandlerFunc(logger.methodNotAllowed)
	s.Path("/logging/{masid}/{agentid}/{topic}/time/{start}/{end}").Methods("GET").
		HandlerFunc(logger.handleGetLogsTime)
	s.Path("/logging/{masid}/{agentid}/{topic}/time/{start}/{end}").
		Methods("POST", "PUT", "DELETE").HandlerFunc(logger.methodNotAllowed)
	s.Path("/state/{masid}/{agentid}").Methods("GET").HandlerFunc(logger.handleGetState)
	s.Path("/state/{masid}/{agentid}").Methods("PUT").HandlerFunc(logger.handlePutState)
	s.Path("/state/{masid}/{agentid}").Methods("POST", "DELETE").
		HandlerFunc(logger.methodNotAllowed)
	s.Path("/state/{masid}/list").Methods("PUT").HandlerFunc(logger.handlePutStateList)
	s.Path("/state/{masid}/list").Methods("POST", "DELETE", "GET").
		HandlerFunc(logger.methodNotAllowed)
	s.PathPrefix("").HandlerFunc(logger.resourceNotFound)
	s.Use(logger.loggingMiddleware)
	serv = &http.Server{
		Addr:    ":" + strconv.Itoa(port),
		Handler: r,
	}
	return
}

// listen opens a http server listening and serving request
func (logger *Logger) listen(serv *http.Server) (err error) {
	logger.logInfo.Println("Logger listening on " + serv.Addr)
	err = serv.ListenAndServe()
	return
}
