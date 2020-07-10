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

// Package client implements a df client
package client

import (

	//"fmt"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"strconv"
	"time"

	"git.rwth-aachen.de/acs/public/cloud/mas/clonemap/pkg/common/httpretry"
	"git.rwth-aachen.de/acs/public/cloud/mas/clonemap/pkg/schemas"
)

// Host contains the host name of df (IP or k8s service name)
var Host = "df"

// Port contains the port on which ams is listening
var Port = 12000

var httpClient = &http.Client{Timeout: time.Second * 60}
var delay = time.Second * 1
var numRetries = 4

// Alive tests if alive
func Alive() (alive bool) {
	alive = false
	_, httpStatus, err := httpretry.Get(httpClient, "http://"+Host+":"+strconv.Itoa(Port)+
		"/api/alive", time.Second*2, 2)
	if err == nil && httpStatus == http.StatusOK {
		alive = true
	}
	return
}

// PostSvc post an mas
func PostSvc(masID int, svc schemas.Service) (retSvc schemas.Service, httpStatus int, err error) {
	var body []byte
	js, _ := json.Marshal(svc)
	body, httpStatus, err = httpretry.Post(httpClient, "http://"+Host+":"+strconv.Itoa(Port)+
		"/api/df/"+strconv.Itoa(masID)+"/svc", "application/json", js, time.Second*2, 2)
	if err != nil {
		return
	}
	err = json.Unmarshal(body, &retSvc)
	return
}

// GetSvc requests mas information
func GetSvc(masID int, desc string) (svc []schemas.Service, httpStatus int, err error) {
	var body []byte
	body, httpStatus, err = httpretry.Get(httpClient, "http://"+Host+":"+strconv.Itoa(Port)+
		"/api/df/"+strconv.Itoa(masID)+"/svc/desc/"+desc, time.Second*2, 2)
	if err != nil {
		return
	}
	err = json.Unmarshal(body, &svc)
	return
}

// GetLocalSvc requests mas information
func GetLocalSvc(masID int, desc string, nodeID int, dist float64) (svc []schemas.Service,
	httpStatus int, err error) {
	var body []byte
	body, httpStatus, err = httpretry.Get(httpClient, "http://"+Host+":"+strconv.Itoa(Port)+
		"/api/df/"+strconv.Itoa(masID)+"/svc/desc/"+desc+"/node/"+strconv.Itoa(nodeID)+"/dist/"+
		fmt.Sprintf("%f", dist), time.Second*2, 2)
	if err != nil {
		return
	}
	err = json.Unmarshal(body, &svc)
	return
}

// DeleteSvc removes service from df
func DeleteSvc(masID int, svcID string) (httpStatus int, err error) {
	httpStatus, err = httpretry.Delete(httpClient, "http://"+Host+":"+strconv.Itoa(Port)+
		"/api/df/"+strconv.Itoa(masID)+"/svc/id/"+svcID, nil, time.Second*2, 2)
	return
}

// PostGraph  post the graph of a mas
func PostGraph(masID int, gr schemas.Graph) (httpStatus int, err error) {
	js, _ := json.Marshal(gr)
	_, httpStatus, err = httpretry.Post(httpClient, "http://"+Host+":"+strconv.Itoa(Port)+
		"/api/df/"+strconv.Itoa(masID)+"/graph", "application/json", js, time.Second*2, 2)
	return
}

// GetGraph returns graph of mas
func GetGraph(masID int) (graph schemas.Graph, httpStatus int, err error) {
	var body []byte
	body, httpStatus, err = httpretry.Get(httpClient, "http://"+Host+":"+strconv.Itoa(Port)+
		"/api/df/"+strconv.Itoa(masID)+"/graph", time.Second*2, 2)
	if err != nil {
		return
	}
	err = json.Unmarshal(body, &graph)
	return
}

// Init initializes the client
func Init(timeout time.Duration, del time.Duration, numRet int) {
	httpClient.Timeout = timeout
	delay = del
	numRetries = numRet
}

func getIP() (ret string) {
	for {
		ips, err := net.LookupHost(Host)
		if len(ips) > 0 && err == nil {
			ret = ips[0]
			break
		}
	}
	return
}
