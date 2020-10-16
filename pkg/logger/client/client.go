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

// Package client contains code for interaction with agent
package client

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"git.rwth-aachen.de/acs/public/cloud/mas/clonemap/pkg/common/httpretry"
	"git.rwth-aachen.de/acs/public/cloud/mas/clonemap/pkg/schemas"
)

// Host contains the host name of logger (IP or k8s service name)
var Host = "logger"

// Port contains the port on which logger is listening
var Port = 11000

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

// PostLogs posts new log messages to logger
func PostLogs(masID int, logs []schemas.LogMessage) (httpStatus int, err error) {
	js, _ := json.Marshal(logs)
	_, httpStatus, err = httpretry.Post(httpClient, "http://"+Host+":"+strconv.Itoa(Port)+
		"/api/logging/"+strconv.Itoa(masID)+"/list", "application/json", js, time.Second*2, 4)
	return
}

// GetLatestLogs gets log messages
func GetLatestLogs(masID int, agentID int, topic string, num int) (msgs []schemas.LogMessage,
	httpStatus int, err error) {
	var body []byte
	body, httpStatus, err = httpretry.Get(httpClient, "http://"+Host+":"+strconv.Itoa(Port)+
		"/api/logging/"+strconv.Itoa(masID)+"/"+strconv.Itoa(agentID)+"/"+topic+"/latest/"+
		strconv.Itoa(num), time.Second*2, 4)
	if err != nil {
		return
	}
	err = json.Unmarshal(body, &msgs)
	if err != nil {
		msgs = []schemas.LogMessage{}
	}
	return
}

// PutState updates the state
func PutState(state schemas.State) (httpStatus int, err error) {
	js, _ := json.Marshal(state)
	_, httpStatus, err = httpretry.Put(httpClient, "http://"+Host+":"+strconv.Itoa(Port)+
		"/api/state/"+strconv.Itoa(state.MASID)+"/"+strconv.Itoa(state.AgentID), js,
		time.Second*2, 4)
	return
}

// UpdateStates updates the state
func UpdateStates(masID int, states []schemas.State) (httpStatus int, err error) {
	js, _ := json.Marshal(states)
	_, httpStatus, err = httpretry.Put(httpClient, "http://"+Host+":"+strconv.Itoa(Port)+
		"/api/state/"+strconv.Itoa(masID)+"/list", js, time.Second*2, 4)
	return
}

// GetState requests state from logger
func GetState(masID int, agentID int) (state schemas.State, httpStatus int, err error) {
	var body []byte
	body, httpStatus, err = httpretry.Get(httpClient, "http://"+Host+":"+strconv.Itoa(Port)+
		"/api/state/"+strconv.Itoa(masID)+"/"+strconv.Itoa(agentID), time.Second*2, 4)
	if err != nil {
		return
	}
	err = json.Unmarshal(body, &state)
	if err != nil {
		state = schemas.State{}
	}
	return
}
