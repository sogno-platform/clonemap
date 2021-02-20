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
	"strings"
	"time"

	"git.rwth-aachen.de/acs/public/cloud/mas/clonemap/pkg/common/httpreply"
	"git.rwth-aachen.de/acs/public/cloud/mas/clonemap/pkg/schemas"
	"github.com/gorilla/mux"
)

// handleAPI is the global handler for requests to path /api
func (logger *Logger) handleAPI(w http.ResponseWriter, r *http.Request) {
	var cmapErr, httpErr error
	logger.logInfo.Println("Received Request: ", r.Method, " ", r.URL.EscapedPath())
	// determine which ressource is requested and call corresponding handler
	respath := strings.Split(r.URL.EscapedPath(), "/")
	resvalid := false

	switch len(respath) {
	case 5:
		if respath[2] == "logging" {
			var masID int
			masID, cmapErr = strconv.Atoi(respath[3])
			if cmapErr == nil {
				if respath[4] == "list" {
					cmapErr, httpErr = logger.handleLoggerList(masID, w, r)
					resvalid = true
				}
			}
		} else if respath[2] == "state" {
			var masID, agentID int
			masID, cmapErr = strconv.Atoi(respath[3])
			if cmapErr == nil {
				if respath[4] == "list" {
					cmapErr, httpErr = logger.handleStateList(masID, w, r)
					resvalid = true
				} else {
					agentID, cmapErr = strconv.Atoi(respath[4])
					if cmapErr == nil {
						cmapErr, httpErr = logger.handleState(masID, agentID, w, r)
						resvalid = true
					}
				}
			}
		}
	case 6:
		if respath[2] == "logging" {
			var masID, agentID int
			masID, cmapErr = strconv.Atoi(respath[3])
			if cmapErr == nil {
				agentID, cmapErr = strconv.Atoi(respath[4])
				if cmapErr == nil {
					if respath[5] == "list" {
						cmapErr, httpErr = logger.handleLoggerList(masID, w, r)
					} else if respath[5] == "comm" {
						cmapErr, httpErr = logger.handleCommunication(masID, agentID, w, r)
					} else {
						logType := respath[5]
						cmapErr, httpErr = logger.handleLoggerNew(masID, agentID, logType, w, r)
					}
					resvalid = true
				}
			}
		}
	case 8:
		if respath[2] == "logging" && respath[6] == "latest" {
			var masID, agentID, num int
			masID, cmapErr = strconv.Atoi(respath[3])
			if cmapErr == nil {
				agentID, cmapErr = strconv.Atoi(respath[4])
				if cmapErr == nil {
					num, cmapErr = strconv.Atoi(respath[7])
					logType := respath[5]
					if cmapErr == nil {
						cmapErr, httpErr = logger.handleLogsLatest(masID, agentID, logType, num,
							w, r)
						resvalid = true
					}
				}
			}
		}
	case 9:
		if respath[2] == "logging" && respath[6] == "time" {
			var masID, agentID int
			var start, end time.Time
			masID, cmapErr = strconv.Atoi(respath[3])
			if cmapErr == nil {
				agentID, cmapErr = strconv.Atoi(respath[4])
				if cmapErr == nil {
					start, cmapErr = time.Parse(time.RFC3339, respath[7])
					if cmapErr == nil {
						end, cmapErr = time.Parse(time.RFC3339, respath[8])
						logType := respath[5]
						if cmapErr == nil {
							cmapErr, httpErr = logger.handleLogsTime(masID, agentID, logType, start,
								end, w, r)
							resvalid = true
						}
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
		logger.logError.Println(respath, cmapErr)
	}
	if httpErr != nil {
		logger.logError.Println(respath, httpErr)
	}
}

// handleAlive is the handler for requests to path /api/alive
func (logger *Logger) handleAlive(w http.ResponseWriter, r *http.Request) {
	logger.logInfo.Println("Received Request: ", r.Method, " ", r.URL.EscapedPath())
	var httpErr error
	httpErr = httpreply.Alive(w, nil)
	logger.logErrors(r.URL.Path, nil, httpErr)
	return
}

// handleLoggerNew is the handler for requests to path /api/logger/{mas-id}/{agent-id}/{logtype}
func (logger *Logger) handleLoggerNew(masID int, agentid int, logType string, w http.ResponseWriter,
	r *http.Request) (cmapErr, httpErr error) {
	if r.Method == "POST" {
		// create new log message entry
		var body []byte
		body, cmapErr = ioutil.ReadAll(r.Body)
		if cmapErr == nil {
			var logmsg schemas.LogMessage
			cmapErr = json.Unmarshal(body, &logmsg)
			if cmapErr == nil {
				go logger.addAgentLogMessage(logmsg)
				httpErr = httpreply.Created(w, nil, "text/plain", []byte("Ressource Created"))
			} else {
				httpErr = httpreply.JSONUnmarshalError(w)
			}
		} else {
			httpErr = httpreply.InvalidBodyError(w)
		}
	} else {
		httpErr = httpreply.MethodNotAllowed(w)
		cmapErr = errors.New("Error: Method not allowed on path /api/logger/{mas-id}/{agent-id}/" +
			"{logtype}")
	}
	return
}

// handleCommunication is the handler for requests to path /api/logger/{mas-id}/{agent-id}/comm
func (logger *Logger) handleCommunication(masID int, agentID int, w http.ResponseWriter,
	r *http.Request) (cmapErr, httpErr error) {
	if r.Method == "GET" {
		var comm []schemas.Communication
		comm, cmapErr = logger.getCommunication(masID, agentID)
		httpErr = httpreply.Resource(w, comm, cmapErr)
	} else if r.Method == "PUT" {
		// update communication data
		var body []byte
		body, cmapErr = ioutil.ReadAll(r.Body)
		if cmapErr == nil {
			var comm []schemas.Communication
			cmapErr = json.Unmarshal(body, &comm)
			if cmapErr == nil {
				go logger.updateCommunication(masID, agentID, comm)
				httpErr = httpreply.Updated(w, nil)
			} else {
				httpErr = httpreply.JSONUnmarshalError(w)
			}
		} else {
			httpErr = httpreply.InvalidBodyError(w)
		}
	} else {
		httpErr = httpreply.MethodNotAllowed(w)
		cmapErr = errors.New("Error: Method not allowed on path /api/logger/{mas-id}/{agent-id}/" +
			"{logtype}")
	}
	return
}

// handleLoggerList is the handler for requests to path /api/logger/{mas-id}/list
func (logger *Logger) handleLoggerList(masID int, w http.ResponseWriter, r *http.Request) (cmapErr,
	httpErr error) {
	if r.Method == "POST" {
		// create new log message entry
		var body []byte
		body, cmapErr = ioutil.ReadAll(r.Body)
		if cmapErr == nil {
			var logmsg []schemas.LogMessage
			cmapErr = json.Unmarshal(body, &logmsg)
			if cmapErr == nil {
				go logger.addAgentLogMessageList(logmsg)
				httpErr = httpreply.Created(w, nil, "text/plain", []byte("Ressource Created"))
			} else {
				httpErr = httpreply.JSONUnmarshalError(w)
			}
		} else {
			httpErr = httpreply.InvalidBodyError(w)
		}
	} else {
		httpErr = httpreply.MethodNotAllowed(w)
		cmapErr = errors.New("Error: Method not allowed on path /api/logger/{mas-id}/list}")
	}
	return
}

// handleLogsLatest is the handler for requests to path /api/logger/{mas-id}/{agent-id}/{logtype}/
// latest/{num}
func (logger *Logger) handleLogsLatest(masID int, agentid int, logType string, num int,
	w http.ResponseWriter, r *http.Request) (cmapErr, httpErr error) {
	if r.Method == "GET" {
		var logMsg []schemas.LogMessage
		logMsg, cmapErr = logger.getLatestAgentLogMessages(masID, agentid, logType, num)
		httpErr = httpreply.Resource(w, logMsg, cmapErr)
	} else {
		httpErr = httpreply.MethodNotAllowed(w)
		cmapErr = errors.New("Error: Method not allowed on path /api/logger/{mas-id}/{agent-id}/" +
			"{logtype}/latest/{num}")
	}
	return
}

// handleLogsTime is the handler for requests to path /api/logger/{mas-id}/{agent-id}/{logtype}/
// time/{start}/{end}
func (logger *Logger) handleLogsTime(masID int, agentid int, logType string, start time.Time,
	end time.Time, w http.ResponseWriter, r *http.Request) (cmapErr, httpErr error) {
	if r.Method == "GET" {
		var logMsg []schemas.LogMessage
		logMsg, cmapErr = logger.getAgentLogMessagesInRange(masID, agentid, logType, start, end)
		httpErr = httpreply.Resource(w, logMsg, cmapErr)
	} else {
		httpErr = httpreply.MethodNotAllowed(w)
		cmapErr = errors.New("Error: Method not allowed on path /api/logger/{mas-id}/{agent-id}/" +
			"{logtype}/time/{start}/{end}")
	}
	return
}

// handleState is the handler for requests to path /api/state/{mas-id}/{agent-id}
func (logger *Logger) handleState(masID int, agentid int, w http.ResponseWriter,
	r *http.Request) (cmapErr, httpErr error) {
	if r.Method == "GET" {
		var state schemas.State
		state, cmapErr = logger.getAgentState(masID, agentid)
		httpErr = httpreply.Resource(w, state, cmapErr)
	} else if r.Method == "PUT" {
		var body []byte
		body, cmapErr = ioutil.ReadAll(r.Body)
		if cmapErr == nil {
			var state schemas.State
			cmapErr = json.Unmarshal(body, &state)
			if cmapErr == nil {
				go logger.updateAgentState(masID, agentid, state)
				httpErr = httpreply.Updated(w, nil)
			} else {
				httpErr = httpreply.JSONUnmarshalError(w)
			}
		} else {
			httpErr = httpreply.InvalidBodyError(w)
		}
	} else {
		httpErr = httpreply.MethodNotAllowed(w)
		cmapErr = errors.New("Error: Method not allowed on path /api/state/{mas-id}/{agent-id}")
	}
	return
}

// handleStateList is the handler for requests to path /api/state/{mas-id}/list
func (logger *Logger) handleStateList(masID int, w http.ResponseWriter,
	r *http.Request) (cmapErr, httpErr error) {
	if r.Method == "PUT" {
		var body []byte
		body, cmapErr = ioutil.ReadAll(r.Body)
		if cmapErr == nil {
			var states []schemas.State
			cmapErr = json.Unmarshal(body, &states)
			if cmapErr == nil {
				go logger.updateAgentStatesList(masID, states)
				httpErr = httpreply.Updated(w, nil)
			} else {
				httpErr = httpreply.JSONUnmarshalError(w)
			}
		} else {
			httpErr = httpreply.InvalidBodyError(w)
		}
	} else {
		httpErr = httpreply.MethodNotAllowed(w)
		cmapErr = errors.New("Error: Method not allowed on path /api/state/{mas-id}/list")
	}
	return
}

// methodNotAllowed is the default handler for valid paths but invalid methods
func (logger *Logger) methodNotAllowed(w http.ResponseWriter, r *http.Request) {
	logger.logInfo.Println("Received Request: ", r.Method, " ", r.URL.EscapedPath())
	httpErr := httpreply.MethodNotAllowed(w)
	cmapErr := errors.New("Error: Method not allowed on path " + r.URL.Path)
	logger.logErrors(r.URL.Path, cmapErr, httpErr)
	return
}

// resourceNotFound is the default handler for invalid paths
func (logger *Logger) resourceNotFound(w http.ResponseWriter, r *http.Request) {
	logger.logInfo.Println("Received Request: ", r.Method, " ", r.URL.EscapedPath())
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

// server creates the logger server
func (logger *Logger) server(port int) (serv *http.Server) {
	r := mux.NewRouter()
	r.HandleFunc("/api/", logger.handleAPI)
	// s := r.PathPrefix("/api").Subrouter()
	// s.Path("/alive").Methods("GET").HandlerFunc(logger.handleAlive)

	// s.PathPrefix("").HandlerFunc(logger.resourceNotFound)

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
