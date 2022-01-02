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

package kubestub

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"

	"github.com/RWTH-ACS/clonemap/pkg/common/httpreply"
	"github.com/RWTH-ACS/clonemap/pkg/schemas"
)

// handleAPI is the global handler for requests to path /api
func (stub *LocalStub) handleAPI(w http.ResponseWriter, r *http.Request) {
	var err error
	// determine which ressource is requested and handle request
	respath := strings.Split(r.URL.EscapedPath(), "/")
	resvalid := false

	switch len(respath) {
	case 1:
		err = errors.New("error - wrong path: /")
	case 2:
		err = errors.New("error - wrong path: /api")
	case 3:
		if respath[2] == "container" {
			if r.Method == "GET" {
				// return list of all agency configurations
				fmt.Println("Received Request: GET /api/container")
				err = httpreply.Resource(w, stub.agencies, nil)
			} else if r.Method == "POST" {
				// check if post request is valid and create new agency container
				fmt.Println("Received Request: POST /api/container")
				var body []byte
				body, err = ioutil.ReadAll(r.Body)
				if err == nil {
					var agconfig schemas.StubAgencyConfig
					err = json.Unmarshal(body, &agconfig)
					if err == nil {
						agexist := false
						for i := range stub.agencies {
							if stub.agencies[i].AgencyID == agconfig.AgencyID &&
								stub.agencies[i].MASID == agconfig.MASID &&
								stub.agencies[i].ImageGroupID == agconfig.ImageGroupID {
								agexist = true
								break
							}
						}
						if !agexist {
							err = stub.createAgency(agconfig.Image, agconfig.MASID, agconfig.ImageGroupID,
								agconfig.AgencyID, agconfig.Logging, agconfig.MQTT, agconfig.DF)
							if err == nil {
								stub.agencies = append(stub.agencies, agconfig)
								err = httpreply.Created(w, nil, "text/plain", []byte("Ressource Created"))
							} else {
								fmt.Println(err)
								err = httpreply.CMAPError(w, err.Error())
							}
						} else {
							fmt.Println(err)
							err = httpreply.MethodNotAllowed(w)
						}
					} else {
						fmt.Println(err)
						err = httpreply.JSONUnmarshalError(w)
					}
				} else {
					fmt.Println(err)
					err = httpreply.InvalidBodyError(w)
				}
			} else {
				fmt.Println("Received invalid request " + r.Method + " : /api/container")
				err = httpreply.MethodNotAllowed(w)
			}
			resvalid = true
		}
	case 4:
		if respath[2] == "container" {
			if r.Method == "DELETE" {
				// check if mas exists and delete containers
				fmt.Println("Received Request: DELETE /api/container/" + respath[3])
				var masID int
				// var agencies []int
				masID, err = strconv.Atoi(respath[3])
				if err == nil {
					for i := range stub.agencies {
						if stub.agencies[i].MASID == masID {
							go stub.deleteAgency(masID, stub.agencies[i].ImageGroupID, stub.agencies[i].AgencyID)
						}
					}
				}
				err = httpreply.Deleted(w, err)
			} else {
				fmt.Println("Received invalid request " + r.Method + " : /api/container" +
					respath[3])
				err = httpreply.MethodNotAllowed(w)
			}
			resvalid = true
		}
	default:
		err = errors.New("Error - wrong path: " + r.URL.EscapedPath())
	}

	if !resvalid {
		err = httpreply.MethodNotAllowed(w)
	}
	if err != nil {
		fmt.Println(err)
	}
}

// listen opens a http server listening and serving request
func (stub *LocalStub) listen() (err error) {
	mux := http.NewServeMux()
	mux.HandleFunc("/api/", stub.handleAPI)
	s := &http.Server{
		Addr:    ":8000",
		Handler: mux,
	}
	err = s.ListenAndServe()
	return
}
