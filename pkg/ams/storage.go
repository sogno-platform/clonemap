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
// stores the state of the AMS. This is MAS configuration and MAS and agent status

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

	// getAgencyInfoFull returns status of one agency
	getAgencyInfoFull(masID int, imID int, agencyID int) (ret schemas.AgencyInfoFull, err error)

	// registerMAS registers a new MAS with the storage and returns its ID
	registerMAS() (masID int, err error)

	// storeMAS stores MAS specs
	storeMAS(masID int, masInfo schemas.MASInfo) (err error)

	// deleteMAS deletes MAS with specified ID
	deleteMAS(masID int) (err error)

	// registerImageGroup registers a new image group with the storage and returns its ID
	registerImageGroup(masID int, config schemas.ImageGroupConfig) (imID int, err error)

	// registerAgent registers a new agent with the storage and returns its ID
	registerAgent(masID int, imID int, agencyID int, spec schemas.AgentSpec) (agentID int,
		err error)

	// addAgent adds an agent to an existing MAS
	addAgent(masID int, agentSpec schemas.AgentSpec) (err error)
}

// CommData helper struct for communication data
type CommData struct {
	ID         int // id of other agent
	NumMsgSent int // number of messages sent to this agent
	NumMsgRecv int // number of messages received from this agent
}

// information storage for local use of clonemap
// note: id should always equal slice index!

// represents local storage
type localStorage struct {
	cloneMAP   schemas.CloneMAP
	masCounter int               // counter for mas
	mas        []schemas.MASInfo // list of all running MAS
	mutex      *sync.Mutex
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
	ret.Inst = make([]schemas.MASInfo, len(stor.mas), len(stor.mas))
	ret.Counter = stor.masCounter
	for i := 0; i < len(stor.mas); i++ {
		ret.Inst[i] = stor.mas[i]
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
	ret = stor.mas[masID]
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
	ret = stor.mas[masID].Agents
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
	if len(stor.mas[masID].Agents.Inst)-1 < agentID {
		err = errors.New("Agent does not exist")
		return
	}
	ret = stor.mas[masID].Agents.Inst[agentID]
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
	if len(stor.mas[masID].Agents.Inst)-1 < agentID {
		stor.mutex.Unlock()
		err = errors.New("Agent does not exist")
		return
	}
	ret = stor.mas[masID].Agents.Inst[agentID].Address
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
	if len(stor.mas[masID].Agents.Inst)-1 < agentID {
		stor.mutex.Unlock()
		err = errors.New("Agent does not exist")
		return
	}
	stor.mas[masID].Agents.Inst[agentID].Address = address
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
	ret.Counter = 0
	for i := range stor.mas[masID].ImageGroups.Inst {
		ret.Inst = append(ret.Inst,
			stor.mas[masID].ImageGroups.Inst[i].Agencies.Inst...)
		ret.Counter += len(stor.mas[masID].ImageGroups.Inst[i].Agencies.Inst)
	}
	stor.mutex.Unlock()
	return
}

// getAgencyInfoFull returns status of one agency
func (stor *localStorage) getAgencyInfoFull(masID int, imID int,
	agencyID int) (ret schemas.AgencyInfoFull, err error) {
	stor.mutex.Lock()
	if len(stor.mas)-1 < masID {
		stor.mutex.Unlock()
		err = errors.New("Agency does not exist")
		return
	}
	if len(stor.mas[masID].ImageGroups.Inst)-1 < imID {
		stor.mutex.Unlock()
		err = errors.New("Agency does not exist")
		return
	}
	if len(stor.mas[masID].ImageGroups.Inst[imID].Agencies.Inst)-1 < agencyID {
		stor.mutex.Unlock()
		err = errors.New("Agency does not exist")
		return
	}
	ret.MASID = masID
	ret.Name = stor.mas[masID].ImageGroups.Inst[imID].Agencies.Inst[agencyID].Name
	ret.ID = agencyID
	ret.ImageGroupID = imID
	ret.Logger = stor.mas[masID].ImageGroups.Inst[imID].Agencies.Inst[agencyID].Logger
	ret.Status = stor.mas[masID].ImageGroups.Inst[imID].Agencies.Inst[agencyID].Status
	ret.Agents = make([]schemas.AgentInfo,
		len(stor.mas[masID].ImageGroups.Inst[imID].Agencies.Inst[agencyID].Agents),
		len(stor.mas[masID].ImageGroups.Inst[imID].Agencies.Inst[agencyID].Agents))
	for i := 0; i < len(ret.Agents); i++ {
		var temp schemas.AgentInfo
		temp, err = stor.getAgentInfoNolock(masID,
			stor.mas[masID].ImageGroups.Inst[imID].Agencies.Inst[agencyID].Agents[i])
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
			stor.mas = append(stor.mas, schemas.MASInfo{})
		}
	} else {
		// check if mas stor is already populated
		if stor.mas[masID].ID == masID {
			err = errors.New("MAS already exists")
			return
		}
	}
	stor.mas[masID] = newMAS
	stor.mutex.Unlock()
	return
}

// createMASStorage returns a filled masStorage object
func createMASStorage(masID int, masInfo schemas.MASInfo) (ret schemas.MASInfo) {
	ret = masInfo
	ret.Status.Code = status.Running
	ret.Status.LastUpdate = time.Now()

	ret.ID = masID
	for i := 0; i < ret.Agents.Counter; i++ {
		ret.Agents.Inst[i].MASID = masID
		ret.Agents.Inst[i].Address.Agency = "mas-" + strconv.Itoa(masID) +
			ret.Agents.Inst[i].Address.Agency + ".mas" + strconv.Itoa(masID) + "agencies"
	}
	for i := 0; i < ret.ImageGroups.Counter; i++ {
		for j := 0; j < ret.ImageGroups.Inst[i].Agencies.Counter; j++ {
			ret.ImageGroups.Inst[i].Agencies.Inst[j].MASID = masID
			ret.ImageGroups.Inst[i].Agencies.Inst[j].Name = "mas-" + strconv.Itoa(masID) +
				ret.ImageGroups.Inst[i].Agencies.Inst[j].Name + ".mas" +
				strconv.Itoa(masID) + "agencies"
		}
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
	stor.mas[masID].Status.Code = status.Terminated
	stor.mas[masID].Status.LastUpdate = time.Now()
	// copy(stor.mas[masID:], stor.mas[masID+1:])
	// stor.mas[len(stor.mas)-1] = masStorage{}
	// stor.mas = stor.mas[:len(stor.mas)-1]
	stor.mutex.Unlock()
	return
}

// registerImageGroup registers a new image group with the storage and returns its ID
func (stor *localStorage) registerImageGroup(masID int, config schemas.ImageGroupConfig) (imID int,
	err error) {
	stor.mutex.Lock()
	if len(stor.mas)-1 < masID {
		stor.mutex.Unlock()
		err = errors.New("MAS does not exist")
		return
	}

	for i := range stor.mas[masID].ImageGroups.Inst {
		if stor.mas[masID].ImageGroups.Inst[i].Config.Image == config.Image {
			stor.mutex.Unlock()
			err = errors.New("ImageGroup already exists")
			return
		}
	}

	info := schemas.ImageGroupInfo{
		Config: config,
		ID:     stor.mas[masID].ImageGroups.Counter,
	}
	stor.mas[masID].ImageGroups.Counter++
	stor.mas[masID].ImageGroups.Inst = append(stor.mas[masID].ImageGroups.Inst,
		info)
	stor.mutex.Unlock()
	imID = info.ID
	return
}

// addAgent adds an agent to an existing MAS
func (stor *localStorage) addAgent(masID int, agentSpec schemas.AgentSpec) (err error) {
	return
}

// registerAgent registers a new agent with the storage and returns its ID
func (stor *localStorage) registerAgent(masID int, imID int, agencyID int,
	spec schemas.AgentSpec) (agentID int, err error) {
	stor.mutex.Lock()
	if len(stor.mas)-1 < masID {
		stor.mutex.Unlock()
		err = errors.New("MAS does not exist")
		return
	}

	agentID = stor.mas[masID].Agents.Counter
	stor.mas[masID].Agents.Counter++

	info := schemas.AgentInfo{
		Spec:         spec,
		MASID:        masID,
		ImageGroupID: imID,
		AgencyID:     agencyID,
		ID:           agentID,
		Address: schemas.Address{
			Agency: "mas-" + strconv.Itoa(masID) + "-im-" + strconv.Itoa(imID) + "-agency-" +
				strconv.Itoa(agencyID) + ".mas" + strconv.Itoa(masID) + "agencies",
		},
	}

	stor.mas[masID].Agents.Inst = append(stor.mas[masID].Agents.Inst, info)
	stor.mutex.Unlock()
	return
}

// newLocalStorage returns Storage interface with localStorage type
func newLocalStorage() storage {
	var temp localStorage
	temp.mutex = &sync.Mutex{}
	temp.cloneMAP.Uptime = time.Now()
	return &temp
}
