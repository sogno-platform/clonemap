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

// Package client contains code for interaction with ams
package client

import (
	"encoding/json"
	//"fmt"
	"net"
	"net/http"
	"strconv"
	"time"

	"git.rwth-aachen.de/acs/public/cloud/mas/clonemap/pkg/common/httpretry"
	"git.rwth-aachen.de/acs/public/cloud/mas/clonemap/pkg/schemas"
)

// Host contains the host name of ams (IP or k8s service name)
var Host = "ams"

// Port contains the port on which ams is listening
var Port = 9000

var httpClient = &http.Client{Timeout: time.Second * 60}
var delay = time.Second * 1
var numRetries = 4

// GetCloneMAP requests CloneMAP information
func GetCloneMAP() (cmap schemas.CloneMAP, httpStatus int, err error) {
	var body []byte
	body, httpStatus, err = httpretry.Get(httpClient, "http://"+Host+":"+strconv.Itoa(Port)+
		"/api/clonemap", time.Second*2, 2)
	if err != nil {
		return
	}
	err = json.Unmarshal(body, &cmap)
	if err != nil {
		cmap = schemas.CloneMAP{}
	}
	return
}

// GetMASs requests mas information
func GetMASs() (mass schemas.MASs, httpStatus int, err error) {
	var body []byte
	body, httpStatus, err = httpretry.Get(httpClient, "http://"+Host+":"+strconv.Itoa(Port)+
		"/api/clonemap/mas", time.Second*2, 2)
	if err != nil {
		return
	}
	err = json.Unmarshal(body, &mass)
	if err != nil {
		mass = schemas.MASs{}
	}
	return
}

// PostMAS post an mas
func PostMAS(mas schemas.MASSpec) (httpStatus int, err error) {
	js, _ := json.Marshal(mas)
	_, httpStatus, err = httpretry.Post(httpClient, "http://"+Host+":"+strconv.Itoa(Port)+
		"/api/clonemap/mas", "application/json", js, time.Second*2, 2)
	return
}

// GetMAS requests mas information
func GetMAS(masID int) (mas schemas.MASInfo, httpStatus int, err error) {
	var body []byte
	body, httpStatus, err = httpretry.Get(httpClient, "http://"+Host+":"+strconv.Itoa(Port)+
		"/api/clonemap/mas/"+strconv.Itoa(masID), time.Second*2, 2)
	if err != nil {
		return
	}
	err = json.Unmarshal(body, &mas)
	if err != nil {
		mas = schemas.MASInfo{}
	}
	return
}

// DeleteMAS deletes a MAS
func DeleteMAS(masID int) (httpStatus int, err error) {
	httpStatus, err = httpretry.Delete(httpClient, "http://"+Host+":"+strconv.Itoa(Port)+
		"/api/clonemap/mas/"+strconv.Itoa(masID), nil,
		time.Second*2, 2)
	return
}

// GetAgents requests agent information
func GetAgents(masID int) (agents schemas.Agents, httpStatus int, err error) {
	var body []byte
	body, httpStatus, err = httpretry.Get(httpClient, "http://"+Host+":"+strconv.Itoa(Port)+
		"/api/clonemap/mas/"+strconv.Itoa(masID)+"/agents", time.Second*2, 2)
	if err != nil {
		return
	}
	err = json.Unmarshal(body, &agents)
	if err != nil {
		agents = schemas.Agents{}
	}
	return
}

// GetAgent requests agent information
func GetAgent(masID int, agentID int) (agent schemas.AgentInfo, httpStatus int, err error) {
	var body []byte
	body, httpStatus, err = httpretry.Get(httpClient, "http://"+Host+":"+strconv.Itoa(Port)+
		"/api/clonemap/mas/"+strconv.Itoa(masID)+"/agents/"+strconv.Itoa(agentID), time.Second*2, 2)
	if err != nil {
		return
	}
	err = json.Unmarshal(body, &agent)
	if err != nil {
		agent = schemas.AgentInfo{}
	}
	return
}

// GetAgentAddress requests agent address
func GetAgentAddress(masID int, agentID int) (address schemas.Address, httpStatus int, err error) {
	var body []byte
	ip := getIP()
	body, httpStatus, err = httpretry.Get(httpClient, "http://"+ip+":"+strconv.Itoa(Port)+
		"/api/clonemap/mas/"+strconv.Itoa(masID)+"/agents/"+strconv.Itoa(agentID)+"/address",
		time.Second*2, 2)
	if err != nil {
		return
	}
	err = json.Unmarshal(body, &address)
	if err != nil {
		address = schemas.Address{}
	}
	return
}

// GetAgencies requests agency information
func GetAgencies(masID int) (agencies schemas.Agencies, httpStatus int, err error) {
	var body []byte
	body, httpStatus, err = httpretry.Get(httpClient, "http://"+Host+":"+strconv.Itoa(Port)+
		"/api/clonemap/mas/"+strconv.Itoa(masID)+"/agencies", time.Second*2, 2)
	if err != nil {
		return
	}
	err = json.Unmarshal(body, &agencies)
	if err != nil {
		agencies = schemas.Agencies{}
	}
	return
}

// GetAgencyInfo requests agency information
func GetAgencyInfo(masID int, imID int, agencyID int) (agency schemas.AgencyInfoFull,
	httpStatus int, err error) {
	var body []byte
	body, httpStatus, err = httpretry.Get(httpClient, "http://"+Host+":"+strconv.Itoa(Port)+
		"/api/clonemap/mas/"+strconv.Itoa(masID)+"/imgroup/"+strconv.Itoa(imID)+"/agency/"+
		strconv.Itoa(agencyID), time.Second*2, 2)
	if err != nil {
		return
	}
	//fmt.Println(string(body))
	err = json.Unmarshal(body, &agency)
	if err != nil {
		agency = schemas.AgencyInfoFull{}
	}
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
