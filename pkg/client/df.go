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

package client

import (

	//"fmt"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"strconv"
	"sync"
	"time"

	"git.rwth-aachen.de/acs/public/cloud/mas/clonemap/pkg/common/httpretry"
	"git.rwth-aachen.de/acs/public/cloud/mas/clonemap/pkg/schemas"
)

// DFClient is the ams client
type DFClient struct {
	httpClient *http.Client  // http client
	host       string        // ams host name
	port       int           // ams port
	delay      time.Duration // delay between two retries
	numRetries int           // number of retries
}

// Alive tests if alive
func (cli *DFClient) Alive() (alive bool) {
	alive = false
	_, httpStatus, err := httpretry.Get(cli.httpClient, cli.prefix()+"/api/alive", time.Second*2, 2)
	if err == nil && httpStatus == http.StatusOK {
		alive = true
	}
	return
}

// PostSvc post an mas
func (cli *DFClient) PostSvc(masID int, svc schemas.Service) (retSvc schemas.Service, httpStatus int,
	err error) {
	var body []byte
	js, _ := json.Marshal(svc)
	body, httpStatus, err = httpretry.Post(cli.httpClient, cli.prefix()+"/api/df/"+
		strconv.Itoa(masID)+"/svc", "application/json", js, time.Second*2, 2)
	if err != nil {
		return
	}
	err = json.Unmarshal(body, &retSvc)
	return
}

// GetSvc requests mas information
func (cli *DFClient) GetSvc(masID int, desc string) (svc []schemas.Service, httpStatus int,
	err error) {
	var body []byte
	body, httpStatus, err = httpretry.Get(cli.httpClient, cli.prefix()+"/api/df/"+
		strconv.Itoa(masID)+"/svc/desc/"+desc, time.Second*2, 2)
	if err != nil {
		return
	}
	err = json.Unmarshal(body, &svc)
	return
}

// GetLocalSvc requests mas information
func (cli *DFClient) GetLocalSvc(masID int, desc string, nodeID int,
	dist float64) (svc []schemas.Service, httpStatus int, err error) {
	var body []byte
	body, httpStatus, err = httpretry.Get(cli.httpClient, cli.prefix()+"/api/df/"+
		strconv.Itoa(masID)+"/svc/desc/"+desc+"/node/"+strconv.Itoa(nodeID)+"/dist/"+
		fmt.Sprintf("%f", dist), time.Second*2, 2)
	if err != nil {
		return
	}
	err = json.Unmarshal(body, &svc)
	return
}

// DeleteSvc removes service from df
func (cli *DFClient) DeleteSvc(masID int, svcID string) (httpStatus int, err error) {
	httpStatus, err = httpretry.Delete(cli.httpClient, cli.prefix()+"/api/df/"+strconv.Itoa(masID)+
		"/svc/id/"+svcID, nil, time.Second*2, 2)
	return
}

// PostGraph  post the graph of a mas
func (cli *DFClient) PostGraph(masID int, gr schemas.Graph) (httpStatus int, err error) {
	js, _ := json.Marshal(gr)
	_, httpStatus, err = httpretry.Post(cli.httpClient, cli.prefix()+"/api/df/"+
		strconv.Itoa(masID)+"/graph", "application/json", js, time.Second*2, 2)
	return
}

// GetGraph returns graph of mas
func (cli *DFClient) GetGraph(masID int) (graph schemas.Graph, httpStatus int, err error) {
	var body []byte
	body, httpStatus, err = httpretry.Get(cli.httpClient, cli.prefix()+"/api/df/"+
		strconv.Itoa(masID)+"/graph", time.Second*2, 2)
	if err != nil {
		return
	}
	err = json.Unmarshal(body, &graph)
	return
}

func (cli *DFClient) getIP() (ret string) {
	for {
		ips, err := net.LookupHost(cli.host)
		if len(ips) > 0 && err == nil {
			ret = ips[0]
			break
		}
	}
	return
}

func (cli *DFClient) prefix() (ret string) {
	ret = "http://" + cli.host + ":" + strconv.Itoa(cli.port)
	return
}

// NewDFClient creates a new AMS client
func NewDFClient(host string, port int, timeout time.Duration, del time.Duration,
	numRet int) (cli *DFClient) {
	cli = &DFClient{
		httpClient: &http.Client{Timeout: timeout},
		host:       host,
		port:       port,
		delay:      del,
		numRetries: numRet,
	}
	return
}

// AgentDF provides access to the functionality of the DF
type AgentDF struct {
	agentID            int
	masID              int
	nodeID             int
	mutex              *sync.Mutex
	registeredServices map[string]schemas.Service
	active             bool // indicates if df is active (switch via env)
	dfClient           *DFClient
	logError           *log.Logger
	logInfo            *log.Logger
}

// RegisterService registers a new service with the DF
func (df *AgentDF) RegisterService(svc schemas.Service) (id string, err error) {
	df.mutex.Lock()
	if !df.active {
		df.mutex.Unlock()
		return
	}
	df.mutex.Unlock()
	id = "-1"
	if svc.Desc == "" {
		err = errors.New("empty description not allowed")
		return
	}
	df.mutex.Lock()
	_, ok := df.registeredServices[svc.Desc]
	df.mutex.Unlock()
	if ok {
		err = errors.New("service already registered")
		return
	}
	df.mutex.Lock()
	masID := df.masID
	agentID := df.agentID
	nodeID := df.nodeID
	df.mutex.Unlock()
	svc.MASID = masID
	svc.AgentID = agentID
	svc.NodeID = nodeID
	svc.CreatedAt = time.Now()
	svc.ChangedAt = svc.CreatedAt
	svc, _, err = df.dfClient.PostSvc(masID, svc)
	id = svc.GUID
	if err != nil {
		return
	}
	df.mutex.Lock()
	df.registeredServices[svc.Desc] = svc
	df.mutex.Unlock()
	return
}

// SearchForService search for a service with given description
func (df *AgentDF) SearchForService(desc string) (svc []schemas.Service, err error) {
	df.mutex.Lock()
	if !df.active {
		df.mutex.Unlock()
		return
	}
	masID := df.masID
	df.mutex.Unlock()
	var temp []schemas.Service
	temp, _, err = df.dfClient.GetSvc(masID, desc)
	if err != nil {
		return
	}
	for i := range temp {
		if temp[i].AgentID != df.agentID {
			svc = append(svc, temp[i])
		}
	}
	return
}

// SearchForLocalService search for a service with given description
func (df *AgentDF) SearchForLocalService(desc string, dist float64) (svc []schemas.Service, err error) {
	df.mutex.Lock()
	if !df.active {
		df.mutex.Unlock()
		return
	}
	masID := df.masID
	nodeID := df.nodeID
	df.mutex.Unlock()
	var temp []schemas.Service
	temp, _, err = df.dfClient.GetLocalSvc(masID, desc, nodeID, dist)
	if err != nil {
		return
	}
	for i := range temp {
		if temp[i].AgentID != df.agentID {
			svc = append(svc, temp[i])
		}
	}
	return
}

// DeregisterService registers a new service with the DF
func (df *AgentDF) DeregisterService(svcID string) (err error) {
	df.mutex.Lock()
	if !df.active {
		df.mutex.Unlock()
		return
	}
	df.mutex.Unlock()
	desc := ""
	df.mutex.Lock()
	masID := df.masID
	for i := range df.registeredServices {
		if df.registeredServices[i].GUID == svcID {
			desc = i
			break
		}
	}
	df.mutex.Unlock()
	if desc == "" {
		err = errors.New("no such service")
		return
	}
	df.mutex.Lock()
	delete(df.registeredServices, desc)
	df.mutex.Unlock()
	_, err = df.dfClient.DeleteSvc(masID, svcID)
	return
}

// NewAgentDF creates a new DF object
func NewAgentDF(masID int, agentID int, nodeID int, dfCli *DFClient, logErr *log.Logger,
	logInf *log.Logger) (df *AgentDF) {
	df = &AgentDF{
		agentID:  agentID,
		masID:    masID,
		nodeID:   nodeID,
		mutex:    &sync.Mutex{},
		active:   false,
		logError: logErr,
		logInfo:  logInf,
	}
	act := os.Getenv("CLONEMAP_DF")
	if act == "ON" {
		df.active = true
	}
	df.registeredServices = make(map[string]schemas.Service)
	return
}

// close closes the DF module
func (df *AgentDF) Close() {
	for d := range df.registeredServices {
		svc := df.registeredServices[d]
		df.DeregisterService(svc.GUID)
	}
	df.mutex.Lock()
	df.logInfo.Println("Closing DF of agent ", df.agentID)
	df.active = false
	df.mutex.Unlock()
}
