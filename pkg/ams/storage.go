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

package ams

// gives an interface for interaction with the storage (local or etcd). The Storage
// stores the state of the AMS. This is MAS configuartion and MAS and agent status

import (
	"errors"
	"strconv"
	"sync"
	"time"

	"git.rwth-aachen.de/acs/public/cloud/mas/clonemap/pkg/schemas"
	"git.rwth-aachen.de/acs/public/cloud/mas/clonemap/pkg/status"
)

// storage interface for interaction with storage
type storage interface {
	// getCloneMAPInfo returns stored info about clonemap
	getCloneMAPInfo() (ret schemas.CloneMAP, err error)

	// setCloneMAPInfo sets info specific to running clonemap instance
	setCloneMAPInfo(cloneMAP schemas.CloneMAP) (err error)

	// getMASs returns specs of all MAS
	getMASs() (ret schemas.MASs, err error)

	// getMASInfo returns info of one MAS
	getMASInfo(masID int) (ret schemas.MASInfo, err error)

	// getAgents returns specs of all agents in MAS
	getAgents(masID int) (ret schemas.Agents, err error)

	// getAgentInfo returns info of one agent
	getAgentInfo(masID int, agentID int) (ret schemas.AgentInfo, err error)

	// getAgentAddress returns address of one agent
	getAgentAddress(masID int, agentID int) (ret schemas.Address, err error)

	// setAgentAddress sets address of agent
	setAgentAddress(masID int, agentID int, address schemas.Address) (err error)

	// getAgencies returns specs of all agencies in MAS
	getAgencies(masID int) (ret schemas.Agencies, err error)

	// getAgencyConfig returns status of one agency
	getAgencyConfig(masID int, agencyID int) (ret schemas.AgencyConfig, err error)

	// registerMAS registers a new MAS with the storage and returns its ID
	registerMAS() (masID int, err error)

	// storeMAS stores MAS specs
	storeMAS(masID int, masInfo schemas.MASInfo) (err error)

	// deleteMAS deletes MAS with specified ID
	deleteMAS(masID int) (err error)

	// addAgent adds an agent to an exsiting MAS
	addAgent(masID int, agentSpec schemas.AgentSpec) (err error)
}

// CommData helper struct for communication data
type CommData struct {
	ID         int // id of other agent
	NumMsgSent int // number of messages sent to this agent
	NumMsgRecv int // number of messages recived from this agent
}

// information storage for local use of clonemap
// note: id should always equal slice index!

// represents local storage
type localStorage struct {
	cloneMAP   schemas.CloneMAP
	masCounter int          // counter for mas
	mas        []masStorage // list of all running MAS
	mutex      *sync.Mutex
}

// mas storage
type masStorage struct {
	spec          schemas.MASSpec
	status        schemas.Status
	agentCounter  int                 // counter for agents
	agents        []schemas.AgentInfo // configuration of agents
	agencyCounter int                 // counter for agencies
	agencies      []schemas.AgencyInfo
	graph         schemas.Graph
}

// getCloneMAPInfo returns stored info about clonemap
func (stor *localStorage) getCloneMAPInfo() (ret schemas.CloneMAP, err error) {
	stor.mutex.Lock()
	ret = stor.cloneMAP
	stor.mutex.Unlock()
	return
}

// setCloneMAPInfo sets info specific to running clonemap instance
func (stor *localStorage) setCloneMAPInfo(cloneMAP schemas.CloneMAP) (err error) {
	stor.mutex.Lock()
	stor.cloneMAP = cloneMAP
	stor.mutex.Unlock()
	return
}

// getMASs returns specs of all MAS
func (stor *localStorage) getMASs() (ret schemas.MASs, err error) {
	stor.mutex.Lock()
	ret.Instances = make([]schemas.MASSpec, len(stor.mas), len(stor.mas))
	ret.Counter = stor.masCounter
	for i := 0; i < len(stor.mas); i++ {
		ret.Instances[i] = stor.mas[i].spec
	}
	stor.mutex.Unlock()
	return
}

// getMASInfo returns info of one MAS
func (stor *localStorage) getMASInfo(masID int) (ret schemas.MASInfo, err error) {
	stor.mutex.Lock()
	if len(stor.mas)-1 < masID {
		stor.mutex.Unlock()
		err = errors.New("MAS does not exist")
		return
	}
	ret.Spec = stor.mas[masID].spec
	ret.Status = stor.mas[masID].status
	ret.Graph = stor.mas[masID].graph
	ret.Agencies.Instances = make([]schemas.AgencySpec, len(stor.mas[masID].agencies),
		len(stor.mas[masID].agencies))
	for i := 0; i < len(stor.mas[masID].agencies); i++ {
		ret.Agencies.Instances[i] = stor.mas[masID].agencies[i].Spec
	}
	ret.Agencies.Counter = stor.mas[masID].agencyCounter
	ret.Agents.Instances = make([]schemas.AgentSpec, len(stor.mas[masID].agents),
		len(stor.mas[masID].agents))
	for i := 0; i < len(stor.mas[masID].agents); i++ {
		ret.Agents.Instances[i] = stor.mas[masID].agents[i].Spec
	}
	ret.Agents.Counter = stor.mas[masID].agentCounter
	stor.mutex.Unlock()
	return
}

// getAgents returns specs of all agents in MAS
func (stor *localStorage) getAgents(masID int) (ret schemas.Agents, err error) {
	stor.mutex.Lock()
	if len(stor.mas)-1 < masID {
		stor.mutex.Unlock()
		err = errors.New("MAS does not exist")
		return
	}
	ret.Instances = make([]schemas.AgentSpec, len(stor.mas[masID].agents),
		len(stor.mas[masID].agents))
	for i := 0; i < len(stor.mas[masID].agents); i++ {
		ret.Instances[i] = stor.mas[masID].agents[i].Spec
	}
	ret.Counter = stor.mas[masID].agentCounter
	stor.mutex.Unlock()
	return
}

// getAgentInfo returns info of one agent
func (stor *localStorage) getAgentInfo(masID int, agentID int) (ret schemas.AgentInfo, err error) {
	stor.mutex.Lock()
	ret, err = stor.getAgentInfoNolock(masID, agentID)
	stor.mutex.Unlock()
	return
}

// getAgentInfo returns info of one agent
func (stor *localStorage) getAgentInfoNolock(masID int,
	agentID int) (ret schemas.AgentInfo, err error) {
	if len(stor.mas)-1 < masID {
		err = errors.New("Agent does not exist")
		return
	}
	if len(stor.mas[masID].agents)-1 < agentID {
		err = errors.New("Agent does not exist")
		return
	}
	ret = stor.mas[masID].agents[agentID]
	return
}

// getAgentAddress returns address of one agent
func (stor *localStorage) getAgentAddress(masID int, agentID int) (ret schemas.Address, err error) {
	stor.mutex.Lock()
	if len(stor.mas)-1 < masID {
		stor.mutex.Unlock()
		err = errors.New("Agent does not exist")
		return
	}
	if len(stor.mas[masID].agents)-1 < agentID {
		stor.mutex.Unlock()
		err = errors.New("Agent does not exist")
		return
	}
	ret = stor.mas[masID].agents[agentID].Address
	stor.mutex.Unlock()
	return
}

// setAgentAddress sets address of agent
func (stor *localStorage) setAgentAddress(masID int, agentID int,
	address schemas.Address) (err error) {
	stor.mutex.Lock()
	if len(stor.mas)-1 < masID {
		stor.mutex.Unlock()
		err = errors.New("Agent does not exist")
		return
	}
	if len(stor.mas[masID].agents)-1 < agentID {
		stor.mutex.Unlock()
		err = errors.New("Agent does not exist")
		return
	}
	stor.mas[masID].agents[agentID].Address = address
	stor.mutex.Unlock()
	return
}

// getAgencies returns specs of all agencies in MAS
func (stor *localStorage) getAgencies(masID int) (ret schemas.Agencies, err error) {
	stor.mutex.Lock()
	if len(stor.mas)-1 < masID {
		stor.mutex.Unlock()
		err = errors.New("Agency does not exist")
		return
	}
	ret.Instances = make([]schemas.AgencySpec, len(stor.mas[masID].agencies),
		len(stor.mas[masID].agencies))
	for i := 0; i < len(stor.mas[masID].agencies); i++ {
		ret.Instances[i] = stor.mas[masID].agencies[i].Spec
	}
	ret.Counter = stor.mas[masID].agencyCounter
	stor.mutex.Unlock()
	return
}

// getAgencyConfig returns status of one agency
func (stor *localStorage) getAgencyConfig(masID int,
	agencyID int) (ret schemas.AgencyConfig, err error) {
	stor.mutex.Lock()
	if len(stor.mas)-1 < masID {
		stor.mutex.Unlock()
		err = errors.New("Agency does not exist")
		return
	}
	if len(stor.mas[masID].agencies)-1 < agencyID {
		stor.mutex.Unlock()
		err = errors.New("Agency does not exist")
		return
	}
	ret.Spec = stor.mas[masID].agencies[agencyID].Spec
	ret.Agents = make([]schemas.AgentInfo, len(ret.Spec.Agents), len(ret.Spec.Agents))
	for i := 0; i < len(ret.Spec.Agents); i++ {
		var temp schemas.AgentInfo
		temp, err = stor.getAgentInfoNolock(masID, ret.Spec.Agents[i])
		if err != nil {
			stor.mutex.Unlock()
			return
		}
		ret.Agents[i] = temp
	}
	stor.mutex.Unlock()
	return
}

// registerMAS registers a new MAS with the storage and returns its ID
func (stor *localStorage) registerMAS() (masID int, err error) {
	stor.mutex.Lock()
	masID = stor.masCounter
	stor.masCounter++
	stor.mutex.Unlock()
	return
}

// storeMAS stores MAS specs
func (stor *localStorage) storeMAS(masID int, masInfo schemas.MASInfo) (err error) {
	newMAS := createMASStorage(masID, masInfo)
	stor.mutex.Lock()
	numMAS := len(stor.mas)
	if numMAS <= masID {
		for i := 0; i < masID-numMAS+1; i++ {
			stor.mas = append(stor.mas, masStorage{})
		}
	} else {
		// check if mas stor is already populated
		if stor.mas[masID].spec.ID == masID {
			err = errors.New("MAS already exists")
			return
		}
	}
	stor.mas[masID] = newMAS
	stor.mutex.Unlock()
	return
}

// createMASStorage returns a filled masStorage object
func createMASStorage(masID int, masInfo schemas.MASInfo) (ret masStorage) {
	ret.status.Code = status.Running
	ret.status.LastUpdate = time.Now()
	ret.spec = masInfo.Spec
	ret.agentCounter = masInfo.Agents.Counter
	ret.agencyCounter = masInfo.Agencies.Counter
	ret.graph = masInfo.Graph

	ret.spec.ID = masID
	ret.agents = make([]schemas.AgentInfo, ret.agentCounter, ret.agentCounter)
	for i := 0; i < ret.agentCounter; i++ {
		ret.agents[i].Spec = masInfo.Agents.Instances[i]
		ret.agents[i].Spec.MASID = masID
		ret.agents[i].Address.Agency = "mas-" + strconv.Itoa(masID) + "-agency-" +
			strconv.Itoa(ret.agents[i].Spec.AgencyID) + ".mas" + strconv.Itoa(masID) + "agencies"
	}
	ret.agencies = make([]schemas.AgencyInfo, ret.agencyCounter, ret.agencyCounter)
	for i := 0; i < ret.agencyCounter; i++ {
		ret.agencies[i].Spec = masInfo.Agencies.Instances[i]
		ret.agencies[i].Spec.MASID = masID
		ret.agencies[i].Spec.Name = "mas-" + strconv.Itoa(masID) + "-agency-" + strconv.Itoa(i) +
			".mas" + strconv.Itoa(masID) + "agencies"
	}
	return
}

// deleteMAS deletes MAS with specified ID
func (stor *localStorage) deleteMAS(masID int) (err error) {
	stor.mutex.Lock()
	if len(stor.mas)-1 < masID {
		stor.mutex.Unlock()
		err = errors.New("MAS does not exist")
		return
	}
	stor.mas[masID].status.Code = status.Terminated
	stor.mas[masID].status.LastUpdate = time.Now()
	// copy(stor.mas[masID:], stor.mas[masID+1:])
	// stor.mas[len(stor.mas)-1] = masStorage{}
	// stor.mas = stor.mas[:len(stor.mas)-1]
	stor.mutex.Unlock()
	return
}

// addAgent adds an agent to an exsiting MAS
func (stor *localStorage) addAgent(masID int, agentSpec schemas.AgentSpec) (err error) {
	return
}

// newLocalStorage returns Storage interface with localStorage type
func newLocalStorage() storage {
	var temp localStorage
	temp.mutex = &sync.Mutex{}
	temp.cloneMAP.Uptime = time.Now()
	return &temp
}
