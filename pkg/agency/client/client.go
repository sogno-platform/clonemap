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

// Package client contains code for interaction with agency
package client

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"git.rwth-aachen.de/acs/public/cloud/mas/clonemap/pkg/common/httpretry"
	"git.rwth-aachen.de/acs/public/cloud/mas/clonemap/pkg/schemas"
	"git.rwth-aachen.de/acs/public/cloud/mas/clonemap/pkg/status"
)

// Client is the ams client
type Client struct {
	httpClient *http.Client  // http client
	Port       int           // ams port
	delay      time.Duration // delay between two retries
	numRetries int           // number of retries
}

// GetInfo requests the agency info
func (cli *Client) GetInfo(agency string) (agencyInfo schemas.AgencyInfo, httpStatus int,
	err error) {
	var body []byte
	body, httpStatus, err = httpretry.Get(cli.httpClient, cli.prefix(agency)+"/api/agency",
		time.Second*2, 2)
	if err != nil {
		return
	}
	err = json.Unmarshal(body, &agencyInfo)
	if err != nil {
		agencyInfo = schemas.AgencyInfo{}
	}
	return
}

// GetAgents requests the agents running in agency
func (cli *Client) GetAgents(agency string) (agentInfo []schemas.AgentInfo, httpStatus int,
	err error) {
	var body []byte
	body, httpStatus, err = httpretry.Get(cli.httpClient, cli.prefix(agency)+"/api/agency/agents",
		time.Second*2, 2)
	if err != nil {
		return
	}
	err = json.Unmarshal(body, &agentInfo)
	if err != nil {
		agentInfo = []schemas.AgentInfo{}
	}
	return
}

// GetAgent requests one agent running in agency
func (cli *Client) GetAgent(agency string, agentID int) (agentInfo schemas.AgentInfo,
	httpStatus int, err error) {
	var body []byte
	body, httpStatus, err = httpretry.Get(cli.httpClient, cli.prefix(agency)+"/api/agency/agents/"+
		strconv.Itoa(agentID), time.Second*2, 2)
	if err != nil {
		return
	}
	err = json.Unmarshal(body, &agentInfo)
	if err != nil {
		agentInfo = schemas.AgentInfo{}
	}
	return
}

// PostAgent post an agent to agency
func (cli *Client) PostAgent(agency string, agent schemas.AgentInfo) (httpStatus int, err error) {
	js, _ := json.Marshal(agent)
	_, httpStatus, err = httpretry.Post(cli.httpClient, cli.prefix(agency)+"/api/agency/agents",
		"application/json", js, time.Second*2, 2)
	return
}

// DeleteAgent requests an agent to terminate
func (cli *Client) DeleteAgent(agency string, agentID int) (httpStatus int, err error) {
	httpStatus, err = httpretry.Delete(cli.httpClient, cli.prefix(agency)+"/api/agency/agents/"+
		strconv.Itoa(agentID), nil, time.Second*2, 2)
	return
}

// GetAgentStatus requests status from agent and returns it
func (cli *Client) GetAgentStatus(agency string, agentID int) (agentStatus schemas.Status,
	httpStatus int, err error) {
	var temp schemas.Status
	var body []byte
	body, httpStatus, err = httpretry.Get(cli.httpClient, cli.prefix(agency)+"/api/agency/agents/"+
		strconv.Itoa(agentID)+"/status", time.Second*2, 2)
	if err != nil {
		agentStatus.Code = status.Error
		return
	}
	err = json.Unmarshal(body, &temp)
	if err != nil {
		agentStatus.Code = status.Error
		return
	}
	agentStatus = temp
	return
}

// PostMsgs post an agent message to the agent
func (cli *Client) PostMsgs(agency string, msgs []schemas.ACLMessage) (httpStatus int, err error) {
	js, _ := json.Marshal(msgs)
	_, httpStatus, err = httpretry.Post(cli.httpClient, cli.prefix(agency)+"/api/agency/msgs",
		"application/json", js, time.Second*2, 2)
	return
}

// ReturnMsg return undeliverable msg
func (cli *Client) ReturnMsg(agency string, msg schemas.ACLMessage) (httpStatus int, err error) {
	js, _ := json.Marshal(msg)
	_, httpStatus, err = httpretry.Post(cli.httpClient, cli.prefix(agency)+"/api/agency/msgundeliv",
		"application/json", js, time.Second*2, 2)
	return
}

func (cli *Client) prefix(agency string) (ret string) {
	ret = "http://" + agency + ":" + strconv.Itoa(cli.Port)
	return
}

// New creates a new AMS client
func New(timeout time.Duration, del time.Duration, numRet int) (cli *Client) {
	cli = &Client{
		httpClient: &http.Client{Timeout: timeout},
		Port:       10000,
		delay:      del,
		numRetries: numRet,
	}
	return
}
