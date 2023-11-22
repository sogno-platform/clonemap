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

// Package agency is used for running a single agency:
// management of agents located in same container / pod
package agency

import (
	"errors"
	"io/ioutil"
	"log"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/RWTH-ACS/clonemap/pkg/client"
	"github.com/RWTH-ACS/clonemap/pkg/schemas"
	"github.com/RWTH-ACS/clonemap/pkg/status"
)

// Agency contains information about agents located in agency
type Agency struct {
	info         schemas.AgencyInfo // configuration of agency
	loggerConfig schemas.LoggerConfig
	dfConfig     schemas.DFConfig
	mqttConfig   schemas.MQTTConfig
	masName      string
	masCustom    string
	// agents    []schemas.AgentInfo // list of agents in agency
	localAgents    map[int]*Agent
	remoteAgents   map[int]*Agent
	remoteAgencies map[string]*remoteAgency
	mutex          *sync.Mutex // mutex to protect agents from concurrent reads and writes
	agentTask      func(*Agent) error
	msgIn          chan []schemas.ACLMessage
	logCollector   *client.LogCollector
	mqttCollector  *mqttCollector
	dfClient       *client.DFClient
	amsClient      *client.AMSClient
	agencyClient   *client.AgencyClient
	logInfo        *log.Logger // logger for info logging
	logError       *log.Logger // logger for error logging
	errChan        chan error
}

// StartAgency is the entrance function of agency
func StartAgency(task func(*Agent) error) (err error) {
	agency := &Agency{
		mutex:          &sync.Mutex{},
		agentTask:      task,
		localAgents:    make(map[int]*Agent),
		remoteAgents:   make(map[int]*Agent),
		remoteAgencies: make(map[string]*remoteAgency),
		msgIn:          make(chan []schemas.ACLMessage, 1000),
		amsClient:      client.NewAMSClient(time.Second*60, time.Second*1, 4),
		agencyClient:   client.NewAgencyClient(time.Second*60, time.Second*1, 4),
		logError:       log.New(os.Stderr, "[ERROR] ", log.LstdFlags),
	}
	err = agency.init()
	if err != nil {
		agency.logError.Println(err)
		return
	}
	go agency.receiveMsgs()

	// catch kill signal in order to terminate agency and agents before exiting
	var gracefulStop = make(chan os.Signal, 10)
	signal.Notify(gracefulStop, syscall.SIGTERM)
	signal.Notify(gracefulStop, syscall.SIGINT)
	go agency.terminate(gracefulStop)

	// catch runtime errors from agents
	go agency.catchAgentErr()

	serv := agency.server(10000)
	if err != nil {
		agency.logError.Println(err)
		return
	}
	err = agency.listen(serv)
	if err != nil {
		agency.logError.Println(err)
	}
	return
}

// init determines the ID of agency from the container hostname and address suffix from env
func (agency *Agency) init() (err error) {
	logType := os.Getenv("CLONEMAP_LOG_LEVEL")
	switch logType {
	case "info":
		agency.logInfo = log.New(os.Stdout, "[INFO] ", log.LstdFlags)
	case "error":
		agency.logInfo = log.New(ioutil.Discard, "", log.LstdFlags)
	default:
		err = errors.New("Wrong log type: " + logType)
		return
	}
	// agency ID is extracted from hostname mas-{id}-agency-{id}
	//fmt.Println("Getting hostname")
	var temp string
	temp, err = os.Hostname()
	if err != nil {
		return
	}
	agency.logInfo.Println("Starting agency ", temp)
	hostname := strings.Split(temp, "-")
	agency.mutex.Lock()
	if len(hostname) < 6 {
		err = errors.New("incorrect hostname")
		agency.mutex.Unlock()
		return
	}
	agency.info.MASID, err = strconv.Atoi(hostname[1])
	if err != nil {
		agency.mutex.Unlock()
		return
	}
	agency.info.ImageGroupID, err = strconv.Atoi(hostname[3])
	if err != nil {
		agency.mutex.Unlock()
		return
	}
	agency.info.ID, err = strconv.Atoi(hostname[5])
	if err != nil {
		agency.mutex.Unlock()
		return
	}
	agency.info.Name = temp + ".mas" + hostname[1] + "agencies"
	agency.mutex.Unlock()

	// request configuration
	var agencyInfoFull schemas.AgencyInfoFull
	agencyInfoFull, _, err = agency.amsClient.GetAgencyInfo(agency.info.MASID,
		agency.info.ImageGroupID, agency.info.ID)
	agency.mutex.Lock()
	agency.info.ID = agencyInfoFull.ID
	agency.loggerConfig = agencyInfoFull.Logger
	agency.dfConfig = agencyInfoFull.DF
	agency.mqttConfig = agencyInfoFull.MQTT
	agency.masName = agencyInfoFull.MASName
	agency.masCustom = agencyInfoFull.MASCustom
	agency.mutex.Unlock()
	if err != nil {
		agency.info.Status = schemas.Status{
			Code:       status.Error,
			LastUpdate: time.Now(),
		}
		return
	}

	agency.mutex.Lock()
	agency.logCollector = client.NewLogCollector(agency.info.MASID, agency.loggerConfig,
		agency.logError, agency.logInfo)
	agency.dfClient = client.NewDFClient(agency.dfConfig.Host, agency.dfConfig.Port,
		time.Second*60, time.Second*1, 4)
	agency.mqttCollector = newMQTTCollector(agency.mqttConfig, agency.info.Name, agency.logError,
		agency.logInfo)
	agency.mutex.Unlock()

	go agency.startAgents(agencyInfoFull)
	return
}

// terminate takes care of terminating all parts of the MAP before exiting. It is to be called as a
// goroutine and waits until an OS signal is inserted into the channel gracefulStop
func (agency *Agency) terminate(gracefulStop chan os.Signal) {
	<-gracefulStop
	agency.logInfo.Println("Terminating agency")
	agency.mutex.Lock()
	for i := range agency.localAgents {
		agency.localAgents[i].Terminate()
	}
	agency.mutex.Unlock()
	agency.mqttCollector.close()
	time.Sleep(time.Second * 2)
	os.Exit(0)
}

// catchAgentErr takes care of terminating all parts of the MAP before exiting. It is to be called as a
// goroutine and waits until a runtime error occurs in an agent and is sent to the agency's channel errChan
func (agency *Agency) catchAgentErr() {
	err := <-agency.errChan
	agency.logError.Fatal("Caught error: `" + err.Error() + "`. Terminating Agency") // TODO log AgentID
	agency.mutex.Lock()
	for i := range agency.localAgents {
		agency.localAgents[i].Terminate()
	}
	agency.mutex.Unlock()
	agency.mqttCollector.close()
	time.Sleep(time.Second * 2)
	os.Exit(0)
}

// startAgents starts all the agents
func (agency *Agency) startAgents(agencyInfoFull schemas.AgencyInfoFull) (err error) {
	agency.logInfo.Println("Starting agents")
	for i := 0; i < len(agencyInfoFull.Agents); i++ {
		err = agency.createAgent(agencyInfoFull.Agents[i])
		if err != nil {
			agency.mutex.Lock()
			agency.info.Status = schemas.Status{
				Code:       status.Error,
				LastUpdate: time.Now(),
			}
			agency.mutex.Unlock()
			return
		}
	}
	return
}

// createAgent creates a new agent according to agInfo
func (agency *Agency) createAgent(agentInfo schemas.AgentInfo) (err error) {
	// check if agent does not exist
	agency.mutex.Lock()
	_, agExist := agency.localAgents[agentInfo.ID]
	agency.mutex.Unlock()
	if agExist {
		err = errors.New("NotAllowedError")
		return
	}
	// allocate port for agent
	agentInfo.Status.Code = status.Starting
	msgIn := make(chan schemas.ACLMessage, 1000)
	agency.mutex.Lock()
	ag := newAgent(agentInfo, agency.masName, agency.masCustom, msgIn, agency.aclLookup,
		agency.logCollector, agency.loggerConfig, agency.mqttCollector, agency.dfConfig.Active,
		agency.dfClient, agency.logError, agency.logInfo)
	agency.localAgents[agentInfo.ID] = ag
	agency.mutex.Unlock()
	ag.startAgent(agency.agentTask, agency.errChan)
	return
}

// getAgentStatus returns status of agent
func (agency *Agency) getAgentStatus(agentID int) (ret schemas.Status, err error) {

	return
}

// removeAgent terminates and removes the agent with the given ID
func (agency *Agency) removeAgent(agentID int) (err error) {
	agency.mutex.Lock()
	ag, ok := agency.localAgents[agentID]
	agency.mutex.Unlock()
	if !ok {
		return
	}
	ag.Terminate()
	agency.mutex.Lock()
	delete(agency.localAgents, agentID)
	agency.mutex.Unlock()
	return
}

// getAgencyInfo returns configuration of agency
func (agency *Agency) getAgencyInfo() (agencyInfo schemas.AgencyInfo, err error) {
	agency.mutex.Lock()
	agencyInfo = agency.info
	agency.mutex.Unlock()
	err = nil
	return
}

// updateAgentCustom updates the custom agent config
func (agency *Agency) updateAgentCustom(agentID int, custom string) (err error) {
	agentExist := false
	agency.mutex.Lock()
	for i := range agency.localAgents {
		if i == agentID {
			agentExist = true
			agency.localAgents[i].updateCustomData(custom)
			break
		}
	}
	agency.mutex.Unlock()
	if !agentExist {
		err = errors.New("agent does not exist")
	}
	return
}
