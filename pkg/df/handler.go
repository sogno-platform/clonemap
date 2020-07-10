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

package df

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
func (df *DF) handleAPI(w http.ResponseWriter, r *http.Request) {
	var cmapErr, httpErr error
	df.logInfo.Println("Received Request: ", r.Method, " ", r.URL.EscapedPath())
	// determine which ressource is requested and call corresponding handler
	respath := strings.Split(r.URL.EscapedPath(), "/")
	resvalid := false

	switch len(respath) {
	case 3:
		if respath[2] == "alive" {
			cmapErr, httpErr = df.handleAlive(w, r)
			resvalid = true
		}
	case 5:
		var masID int
		masID, cmapErr = strconv.Atoi(respath[3])
		if respath[2] == "df" && cmapErr == nil {
			if respath[4] == "svc" {
				cmapErr, httpErr = df.handlemasID(masID, w, r)
				resvalid = true
			} else if respath[4] == "graph" {
				cmapErr, httpErr = df.handleMASGraph(masID, w, r)
				resvalid = true
			}
		}
	case 7:
		var masID int
		masID, cmapErr = strconv.Atoi(respath[3])
		if respath[2] == "df" && respath[4] == "svc" && cmapErr == nil {
			if respath[5] == "desc" {
				cmapErr, httpErr = df.handleSvcDesc(masID, respath[6], w, r)
				resvalid = true
			} else if respath[5] == "id" {
				svcID := respath[6]
				if cmapErr == nil {
					cmapErr, httpErr = df.handleSvcID(masID, svcID, w, r)
					resvalid = true
				}
			}
		}
	case 11:
		var masID int
		masID, cmapErr = strconv.Atoi(respath[3])
		if respath[2] == "df" && respath[4] == "svc" && respath[5] == "desc" &&
			respath[7] == "node" && respath[9] == "dist" && cmapErr == nil {
			var nodeID int
			nodeID, cmapErr = strconv.Atoi(respath[8])
			if cmapErr == nil {
				var dist float64
				dist, cmapErr = strconv.ParseFloat(respath[10], 64)
				if cmapErr == nil {
					cmapErr, httpErr = df.handleSvcNode(masID, respath[6], nodeID, dist, w, r)
					resvalid = true
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
		df.logError.Println(respath, cmapErr)
	}
	if httpErr != nil {
		df.logError.Println(respath, httpErr)
	}
}

// handleAlive is the handler for requests to path /api/alive
func (df *DF) handleAlive(w http.ResponseWriter, r *http.Request) (cmapErr, httpErr error) {
	if r.Method == "GET" {
		httpErr = httpreply.Alive(w, nil)
	} else {
		httpErr = httpreply.MethodNotAllowed(w)
		cmapErr = errors.New("Error: Method not allowed on path /api/alive")
	}
	return
}

// handlemasID is the handler for requests to path /api/df/{mas-id}/svc
func (df *DF) handlemasID(masID int, w http.ResponseWriter,
	r *http.Request) (cmapErr, httpErr error) {
	if r.Method == "GET" {
		var svc []schemas.Service
		svc, cmapErr = df.stor.searchServices(masID, "")
		httpErr = httpreply.Resource(w, svc, cmapErr)
	} else if r.Method == "POST" {
		var body []byte
		body, cmapErr = ioutil.ReadAll(r.Body)
		if cmapErr == nil {
			var svc schemas.Service
			cmapErr = json.Unmarshal(body, &svc)
			if cmapErr == nil {
				var id string
				id, cmapErr = df.stor.registerService(svc)
				svc.GUID = id
				var res []byte
				res, cmapErr = json.Marshal(svc)
				if cmapErr == nil {
					httpErr = httpreply.Created(w, cmapErr, "application/json", res)
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
		cmapErr = errors.New("Error: Method not allowed on path /api/mas/{mas-id}/svc")
	}
	return
}

// handleMASGraph is the handler for requests to path /api/df/{mas-id}/graph
func (df *DF) handleMASGraph(masID int, w http.ResponseWriter,
	r *http.Request) (cmapErr, httpErr error) {
	if r.Method == "GET" {
		var gr schemas.Graph
		gr, cmapErr = df.stor.getGraph(masID)
		httpErr = httpreply.Resource(w, gr, cmapErr)
	} else if r.Method == "POST" || r.Method == "PUT" {
		var body []byte
		body, cmapErr = ioutil.ReadAll(r.Body)
		if cmapErr == nil {
			var gr schemas.Graph
			cmapErr = json.Unmarshal(body, &gr)
			if cmapErr == nil {
				cmapErr = df.stor.updateGraph(masID, gr)
				httpErr = httpreply.Created(w, cmapErr, "text/plain", []byte("Ressource Created"))
			} else {
				httpErr = httpreply.JSONUnmarshalError(w)
			}
		} else {
			httpErr = httpreply.InvalidBodyError(w)
		}
	} else {
		httpErr = httpreply.MethodNotAllowed(w)
		cmapErr = errors.New("Error: Method not allowed on path /api/mas/{mas-id}/graph")
	}
	return
}

// handleSvcDesc is the handler for requests to path /api/df/{mas-id}/svc/desc/{desc}
func (df *DF) handleSvcDesc(masID int, desc string, w http.ResponseWriter,
	r *http.Request) (cmapErr, httpErr error) {
	if r.Method == "GET" {
		var svc []schemas.Service
		svc, cmapErr = df.stor.searchServices(masID, desc)
		httpErr = httpreply.Resource(w, svc, cmapErr)
	} else {
		httpErr = httpreply.MethodNotAllowed(w)
		cmapErr = errors.New("Error: Method not allowed on path /api/mas/{mas-id}/svc/desc/{desc}")
	}
	return
}

// handleSvcNode is the handler for requests to path /api/df/{mas-id}/svc/desc/{desc}/node/
// {nodeid}/dist/{dist}
func (df *DF) handleSvcNode(masID int, desc string, nodeID int, dist float64, w http.ResponseWriter,
	r *http.Request) (cmapErr, httpErr error) {
	if r.Method == "GET" {
		var svc []schemas.Service
		svc, cmapErr = df.stor.searchLocalServices(masID, nodeID, dist, desc)
		httpErr = httpreply.Resource(w, svc, cmapErr)
	} else {
		httpErr = httpreply.MethodNotAllowed(w)
		cmapErr = errors.New("Error: Method not allowed on path /api/mas/{mas-id}/svc/desc/{desc}")
	}
	return
}

// handleSvcID is the handler for requests to path /api/df/{mas-id}/svc/id/{svcID}
func (df *DF) handleSvcID(masID int, svcID string, w http.ResponseWriter,
	r *http.Request) (cmapErr, httpErr error) {
	if r.Method == "GET" {
		var svc schemas.Service
		svc, cmapErr = df.stor.getService(masID, svcID)
		httpErr = httpreply.Resource(w, svc, cmapErr)
	} else if r.Method == "DELETE" {
		cmapErr = df.stor.deregisterService(masID, svcID)
		httpErr = httpreply.Deleted(w, cmapErr)
	} else {
		httpErr = httpreply.MethodNotAllowed(w)
		cmapErr = errors.New("Error: Method not allowed on path /api/mas/{mas-id}/svc/id/{svcID}")
	}
	return
}

// listen opens a http server listening and serving request
func (df *DF) listen() (err error) {
	mux := http.NewServeMux()
	mux.HandleFunc("/api/", df.handleAPI)
	s := &http.Server{
		Addr:    ":12000",
		Handler: mux,
	}
	err = s.ListenAndServe()
	return
}
