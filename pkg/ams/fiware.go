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

// entities and attributes in clonemap tenancy:
//
// clonemap
//     info: schmemas.CloneMAP
//     mascounter: int
// mas<id>
//     config: schemas.MASConfig
//     status: schemas.Status
//     imcounter: int
//     agentcounter: int
// mas<id>im<id>
//     config: schemas.ImageGroupConfig
//     agencycounter: int
// mas<id>im<id>agency<id>
//     info: schemas.AgencyInfo
// mas<id>agent<id>
//     info schemas.AgentInfo

package ams

import (
	"encoding/json"
	"errors"
	"log"
	"strconv"
	"time"

	"git.rwth-aachen.de/acs/public/cloud/fiware/gofiware/pkg/orion"
	"git.rwth-aachen.de/acs/public/cloud/mas/clonemap/pkg/schemas"
	"git.rwth-aachen.de/acs/public/cloud/mas/clonemap/pkg/status"
)

// fiwareStorage
type fiwareStorage struct {
	cli      *orion.Client // fiware orion client
	logError *log.Logger   // logger for error logging
}

// getCloneMAPInfo returns stored info about clonemap
func (stor *fiwareStorage) getCloneMAPInfo() (ret schemas.CloneMAP, err error) {
	var attr orion.Attribute
	attr, err = stor.cli.GetAttribute("clonemap", "info", "clonemap")
	if err != nil {
		return
	}
	err = extractAttributeValue(attr, &ret)
	return
}

// setCloneMAPInfo sets info specific to running clonemap instance
func (stor *fiwareStorage) setCloneMAPInfo(cloneMAP schemas.CloneMAP) (err error) {
	attr := orion.Attribute{
		Value: cloneMAP,
		Type:  "cloneMAP",
	}
	attrList := orion.AttributeList{Attributes: make(map[string]orion.Attribute)}
	attrList.Attributes["info"] = attr
	err = stor.cli.UpdateAttributes("clonemap", attrList, "clonemap")
	return
}

// getMASsShort returns specs of all MAS
func (stor *fiwareStorage) getMASsShort() (ret []schemas.MASInfoShort, err error) {
	var mascounter int
	mascounter, err = stor.getMasCounter()
	for i := 0; i < mascounter; i++ {
		var masInfo schemas.MASInfoShort
		masInfo, err = stor.getMASInfoShort(i)
		if err != nil {
			return
		}
		ret = append(ret, masInfo)
	}
	return
}

// getMASs returns specs of all MAS
func (stor *fiwareStorage) getMASs() (ret schemas.MASs, err error) {
	ret.Counter, err = stor.getMasCounter()
	for i := 0; i < ret.Counter; i++ {
		var masInfo schemas.MASInfo
		masInfo, err = stor.getMASInfo(i)
		if err != nil {
			return
		}
		ret.Inst = append(ret.Inst, masInfo)
	}
	return
}

// getMASInfoShort returns info of one MAS
func (stor *fiwareStorage) getMASInfoShort(masID int) (ret schemas.MASInfoShort, err error) {
	// check if mas exists
	var masExist bool
	masExist, err = stor.masExists(masID)
	if err != nil {
		return
	}
	if !masExist {
		err = errors.New("MAS does not exist")
		return
	}
	ret.ID = masID
	// get mas entity
	var masEnt orion.Entity
	masEnt, err = stor.cli.GetEntity("mas"+strconv.Itoa(masID), "clonemap")
	if err != nil {
		return
	}
	// config
	configAttr, ok := masEnt.Attributes["config"]
	if !ok {
		err = errors.New("missing config attribute")
		return
	}
	var masConfig schemas.MASConfig
	err = extractAttributeValue(configAttr, &masConfig)
	if err != nil {
		return
	}
	ret.Config = masConfig
	// status
	statusAttr, ok := masEnt.Attributes["status"]
	if !ok {
		err = errors.New("missing status attribute")
		return
	}
	var masStatus schemas.Status
	err = extractAttributeValue(statusAttr, &masStatus)
	if err != nil {
		return
	}
	ret.Status = masStatus

	// agents
	var agents schemas.Agents
	agents, err = stor.getAgents(masID)
	if err != nil {
		return
	}
	ret.NumAgents = agents.Counter

	return
}

// getMASInfo returns info of one MAS
func (stor *fiwareStorage) getMASInfo(masID int) (ret schemas.MASInfo, err error) {
	// check if mas exists
	var masExist bool
	masExist, err = stor.masExists(masID)
	if err != nil {
		return
	}
	if !masExist {
		err = errors.New("MAS does not exist")
		return
	}
	ret.ID = masID
	// get mas entity
	var masEnt orion.Entity
	masEnt, err = stor.cli.GetEntity("mas"+strconv.Itoa(masID), "clonemap")
	if err != nil {
		return
	}
	// config
	configAttr, ok := masEnt.Attributes["config"]
	if !ok {
		err = errors.New("missing config attribute")
		return
	}
	var masConfig schemas.MASConfig
	err = extractAttributeValue(configAttr, &masConfig)
	if err != nil {
		return
	}
	ret.Config = masConfig
	// status
	statusAttr, ok := masEnt.Attributes["status"]
	if !ok {
		err = errors.New("missing status attribute")
		return
	}
	var masStatus schemas.Status
	err = extractAttributeValue(statusAttr, &masStatus)
	if err != nil {
		return
	}
	ret.Status = masStatus
	// graph TODO
	// graphAttr, ok := masEnt.Attributes["graph"]
	// if !ok {
	// 	err = errors.New("missing graph attribute")
	// 	return
	// }
	// var masGraph schemas.Graph
	// err = extractAttributeValue(graphAttr, &masGraph)
	// if err != nil {
	// 	return
	// }
	// ret.Graph = masGraph

	// agents
	ret.Agents, err = stor.getAgents(masID)
	if err != nil {
		return
	}

	// imagegroups
	ret.ImageGroups, err = stor.getGroups(masID)
	if err != nil {
		return
	}

	return
}

// getGroups returns specs of all agents in MAS
func (stor *fiwareStorage) getGroups(masID int) (ret schemas.ImageGroups, err error) {
	// check if mas exists
	var masExist bool
	masExist, err = stor.masExists(masID)
	if err != nil {
		return
	}
	if !masExist {
		err = errors.New("MAS does not exist")
		return
	}

	// im counter
	var imCounter int
	imCounter, err = stor.getImCounter(masID)
	if err != nil {
		return
	}
	ret.Counter = imCounter

	// im instances
	for i := 0; i < imCounter; i++ {
		var imInfo schemas.ImageGroupInfo
		imInfo, err = stor.getGroupInfo(masID, i)
		if err != nil {
			return
		}
		ret.Inst = append(ret.Inst, imInfo)
	}

	return
}

// getGroupInfo returns info of one image group
func (stor *fiwareStorage) getGroupInfo(masID int, imID int) (ret schemas.ImageGroupInfo,
	err error) {
	// check if im exists
	var imExist bool
	imExist, err = stor.imExists(masID, imID)
	if err != nil {
		return
	}
	if !imExist {
		err = errors.New("im does not exist")
		return
	}
	ret.ID = imID

	// get im entity
	var imEnt orion.Entity
	imEnt, err = stor.cli.GetEntity("mas"+strconv.Itoa(masID)+"im"+strconv.Itoa(imID), "clonemap")
	if err != nil {
		return
	}
	// config
	configAttr, ok := imEnt.Attributes["config"]
	if !ok {
		err = errors.New("missing config attribute")
		return
	}
	var imConfig schemas.ImageGroupConfig
	err = extractAttributeValue(configAttr, &imConfig)
	if err != nil {
		return
	}
	ret.Config = imConfig
	// agencies
	ret.Agencies, err = stor.getAgenciesGroup(masID, imID)
	return
}

// getAgents returns specs of all agents in MAS
func (stor *fiwareStorage) getAgents(masID int) (ret schemas.Agents, err error) {
	// check if mas exists
	var masExist bool
	masExist, err = stor.masExists(masID)
	if err != nil {
		return
	}
	if !masExist {
		err = errors.New("MAS does not exist")
		return
	}

	// get all agent entities
	agentEntities, err := stor.cli.GetAllEntitiesPattern("^mas"+strconv.Itoa(masID)+"agent",
		"clonemap")
	for i := range agentEntities {
		ret.Counter++
		attr, ok := agentEntities[i].Attributes["info"]
		if !ok {
			err = errors.New("missing info attribute")
			return
		}
		var agentInfo schemas.AgentInfo
		err = extractAttributeValue(attr, &agentInfo)
		if err != nil {
			return
		}
		ret.Inst = append(ret.Inst, agentInfo)
	}

	return
}

// getAgentInfo returns info of one agent
func (stor *fiwareStorage) getAgentInfo(masID int, agentID int) (ret schemas.AgentInfo, err error) {
	// check if agent exists
	var agentExist bool
	agentExist, err = stor.agentExists(masID, agentID)
	if err != nil {
		return
	}
	if !agentExist {
		err = errors.New("Agent does not exist")
		return
	}

	attr, err := stor.cli.GetAttribute("mas"+strconv.Itoa(masID)+"agent"+strconv.Itoa(agentID),
		"info", "clonemap")
	if err != nil {
		return
	}
	err = extractAttributeValue(attr, &ret)

	return
}

// getAgentAddress returns address of one agent
func (stor *fiwareStorage) getAgentAddress(masID int, agentID int) (ret schemas.Address,
	err error) {
	var agentInfo schemas.AgentInfo
	agentInfo, err = stor.getAgentInfo(masID, agentID)
	if err != nil {
		return
	}
	ret = agentInfo.Address
	return
}

// setAgentAddress sets address of agent
func (stor *fiwareStorage) setAgentAddress(masID int, agentID int,
	address schemas.Address) (err error) {

	return
}

// setAgentCustom sets custom config of agent
func (stor *fiwareStorage) setAgentCustom(masID int, agentID int, custom string) (err error) {

	return
}

// getAgencies returns specs of all agencies in MAS
func (stor *fiwareStorage) getAgencies(masID int) (ret schemas.Agencies, err error) {
	// check if mas exists
	var masExist bool
	masExist, err = stor.masExists(masID)
	if err != nil {
		return
	}
	if !masExist {
		err = errors.New("MAS does not exist")
		return
	}

	var groups schemas.ImageGroups
	groups, err = stor.getGroups(masID)
	if err != nil {
		return
	}
	ret.Counter = 0
	for i := range groups.Inst {
		ret.Counter += groups.Inst[i].Agencies.Counter
		ret.Inst = append(ret.Inst, groups.Inst[i].Agencies.Inst...)
	}
	return
}

// getAgenciesGroup returns specs of all agencies in a group of MAS
func (stor *fiwareStorage) getAgenciesGroup(masID int, imID int) (ret schemas.Agencies, err error) {
	// check if group exists
	var imExist bool
	imExist, err = stor.imExists(masID, imID)
	if err != nil {
		return
	}
	if !imExist {
		err = errors.New("Group does not exist")
		return
	}

	// get all agency entities
	agencyEntities, err := stor.cli.GetAllEntitiesPattern("^mas"+strconv.Itoa(masID)+"im"+
		strconv.Itoa(imID)+"agency", "clonemap")
	for i := range agencyEntities {
		ret.Counter++
		attr, ok := agencyEntities[i].Attributes["info"]
		if !ok {
			err = errors.New("missing info attribute")
			return
		}
		var agencyInfo schemas.AgencyInfo
		err = extractAttributeValue(attr, &agencyInfo)
		if err != nil {
			return
		}
		ret.Inst = append(ret.Inst, agencyInfo)
	}
	return
}

// getAgencyInfoFull returns status of one agency
func (stor *fiwareStorage) getAgencyInfoFull(masID int, imID int,
	agencyID int) (ret schemas.AgencyInfoFull, err error) {
	// check if agency exists
	var agencyExist bool
	agencyExist, err = stor.agencyExists(masID, imID, agencyID)
	if err != nil {
		return
	}
	if !agencyExist {
		err = errors.New("Agency does not exist")
		return
	}

	var agencyInfo schemas.AgencyInfo
	attr, err := stor.cli.GetAttribute("mas"+strconv.Itoa(masID)+"im"+strconv.Itoa(imID)+"agency"+
		strconv.Itoa(agencyID), "info", "clonemap")
	if err != nil {
		return
	}
	err = extractAttributeValue(attr, &agencyInfo)

	ret.MASID = masID
	ret.Name = agencyInfo.Name
	ret.ID = agencyID
	ret.ImageGroupID = imID
	// ret.Logger = agencyInfo.Logger
	ret.Status = agencyInfo.Status
	ret.Agents = make([]schemas.AgentInfo, len(agencyInfo.Agents), len(agencyInfo.Agents))
	for i := 0; i < len(ret.Agents); i++ {
		var temp schemas.AgentInfo
		temp, err = stor.getAgentInfo(masID, agencyInfo.Agents[i])
		if err != nil {
			return
		}
		ret.Agents[i] = temp
	}

	return
}

// registerMAS registers a new MAS with the storage and returns its ID
func (stor *fiwareStorage) registerMAS() (masID int, err error) {
	// get mascounter
	mascounter, err := stor.getMasCounter()
	if err != nil {
		return
	}
	masID = mascounter
	mascounter++
	attrList := orion.AttributeList{Attributes: make(map[string]orion.Attribute)}
	attrList.Attributes["mascounter"] = orion.Attribute{Value: mascounter, Type: "Integer"}
	err = stor.cli.UpdateAttributes("clonemap", attrList, "clonemap")

	return
}

// storeMAS stores MAS specs
func (stor *fiwareStorage) storeMAS(masID int, masInfo schemas.MASInfo) (err error) {
	masInfo = createMASStorage(masID, masInfo)
	// mas entity
	masEntity := orion.Entity{
		ID:         "mas" + strconv.Itoa(masID),
		Type:       "MAS",
		Attributes: make(map[string]orion.Attribute),
	}
	masConfig := masInfo.Config
	masStatus := schemas.Status{Code: status.Running, LastUpdate: time.Now()}
	masEntity.Attributes["config"] = orion.Attribute{Value: masConfig, Type: "MASConfig"}
	masEntity.Attributes["status"] = orion.Attribute{Value: masStatus, Type: "Status"}
	masEntity.Attributes["agentcounter"] = orion.Attribute{Value: masInfo.Agents.Counter,
		Type: "Integer"}
	masEntity.Attributes["imcounter"] = orion.Attribute{Value: masInfo.ImageGroups.Counter,
		Type: "Integer"}
	err = stor.cli.PostEntity(masEntity, "clonemap")
	if err != nil {
		return
	}

	// im entities
	for i := range masInfo.ImageGroups.Inst {
		imEntity := orion.Entity{
			ID:         "mas" + strconv.Itoa(masID) + "im" + strconv.Itoa(i),
			Type:       "ImageGroup",
			Attributes: make(map[string]orion.Attribute),
		}
		imEntity.Attributes["config"] = orion.Attribute{Value: masInfo.ImageGroups.Inst[i].Config,
			Type: "ImageGroupConfig"}
		imEntity.Attributes["agencycounter"] =
			orion.Attribute{Value: masInfo.ImageGroups.Inst[i].Agencies.Counter, Type: "Integer"}
		err = stor.cli.PostEntity(imEntity, "clonemap")
		if err != nil {
			return
		}
		// agency entities
		for j := range masInfo.ImageGroups.Inst[i].Agencies.Inst {
			agencyEntity := orion.Entity{
				ID: "mas" + strconv.Itoa(masID) + "im" + strconv.Itoa(i) + "agency" +
					strconv.Itoa(j),
				Type:       "Agency",
				Attributes: make(map[string]orion.Attribute),
			}
			agencyEntity.Attributes["info"] = orion.Attribute{
				Value: masInfo.ImageGroups.Inst[i].Agencies.Inst[j], Type: "AgencyInfo"}
			err = stor.cli.PostEntity(agencyEntity, "clonemap")
			if err != nil {
				return
			}
		}
	}

	// agent entities
	for i := range masInfo.Agents.Inst {
		agentEntity := orion.Entity{
			ID:         "mas" + strconv.Itoa(masID) + "agent" + strconv.Itoa(i),
			Type:       "Agent",
			Attributes: make(map[string]orion.Attribute),
		}
		agentEntity.Attributes["info"] = orion.Attribute{Value: masInfo.Agents.Inst[i],
			Type: "AgentInfo"}
		err = stor.cli.PostEntity(agentEntity, "clonemap")
		if err != nil {
			return
		}
	}

	return
}

// deleteMAS deletes MAS with specified ID
func (stor *fiwareStorage) deleteMAS(masID int) (err error) {

	return
}

// registerImageGroup registers a new image group with the storage and returns its ID
func (stor *fiwareStorage) registerImageGroup(masID int,
	config schemas.ImageGroupConfig) (newGroup bool, imID int, err error) {
	// check if mas exists
	var masExist bool
	masExist, err = stor.masExists(masID)
	if err != nil {
		return
	}
	if !masExist {
		err = errors.New("MAS does not exist")
		return
	}
	newGroup = true

	var groups schemas.ImageGroups
	groups, err = stor.getGroups(masID)
	if err != nil {
		return
	}

	for i := range groups.Inst {
		if groups.Inst[i].Config.Image == config.Image {
			newGroup = false
			imID = groups.Inst[i].ID
			break
		}
	}

	if !newGroup {
		return
	}

	var imCounter int
	imCounter, err = stor.getImCounter(masID)
	if err != nil {
		return
	}
	info := schemas.ImageGroupInfo{
		Config: config,
		ID:     imCounter,
	}
	imCounter++
	attrList := orion.AttributeList{Attributes: make(map[string]orion.Attribute)}
	attrList.Attributes["imcounter"] = orion.Attribute{Value: imCounter, Type: "Integer"}
	err = stor.cli.UpdateAttributes("mas"+strconv.Itoa(masID), attrList, "clonemap")
	if err != nil {
		return
	}
	imEntity := orion.Entity{
		ID:         "mas" + strconv.Itoa(masID) + "im" + strconv.Itoa(info.ID),
		Type:       "ImageGroup",
		Attributes: make(map[string]orion.Attribute),
	}
	imEntity.Attributes["config"] = orion.Attribute{Value: info.Config, Type: "ImageGroupConfig"}
	imEntity.Attributes["agencycounter"] = orion.Attribute{Value: 0, Type: "Integer"}
	err = stor.cli.PostEntity(imEntity, "clonemap")
	if err != nil {
		return
	}

	imID = info.ID
	return
}

// registerAgent registers a new agent with the storage and returns its ID
func (stor *fiwareStorage) registerAgent(masID int, imID int, spec schemas.AgentSpec) (agentID int,
	err error) {
	// check if mas exists
	var masExist bool
	masExist, err = stor.masExists(masID)
	if err != nil {
		return
	}
	if !masExist {
		err = errors.New("MAS does not exist")
		return
	}

	var agentCounter int
	agentCounter, err = stor.getAgentCounter(masID)
	if err != nil {
		return
	}
	agentID = agentCounter

	agentCounter++
	attrList := orion.AttributeList{Attributes: make(map[string]orion.Attribute)}
	attrList.Attributes["agentcounter"] = orion.Attribute{Value: agentCounter, Type: "Integer"}
	err = stor.cli.UpdateAttributes("mas"+strconv.Itoa(masID), attrList, "clonemap")

	return
}

// addAgent adds an agent to an exsiting MAS
func (stor *fiwareStorage) addAgent(masID int, imID int,
	agentSpec schemas.AgentSpec) (newAgency bool, agentID int, agencyID int, err error) {
	// check if group exists
	var imExist bool
	imExist, err = stor.imExists(masID, imID)
	if err != nil {
		return
	}
	if !imExist {
		err = errors.New("Group does not exist")
		return
	}

	// register agent
	agentID, err = stor.registerAgent(masID, imID, agentSpec)
	if err != nil {
		return
	}

	// schedule agent to agency
	var masConfig schemas.MASConfig
	var attr orion.Attribute
	attr, err = stor.cli.GetAttribute("mas"+strconv.Itoa(masID), "config", "clonemap")
	if err != nil {
		return
	}
	err = extractAttributeValue(attr, &masConfig)
	if err != nil {
		return
	}
	newAgency = true
	var im schemas.ImageGroupInfo
	im, err = stor.getGroupInfo(masID, imID)
	if err != nil {
		return
	}
	for i := range im.Agencies.Inst {
		if len(im.Agencies.Inst[i].Agents) < masConfig.NumAgentsPerAgency {
			im.Agencies.Inst[i].Agents = append(im.Agencies.Inst[i].Agents, agentID)
			agencyID = i
			newAgency = false
			attrList := orion.AttributeList{Attributes: make(map[string]orion.Attribute)}
			attrList.Attributes["agencycounter"] = orion.Attribute{Value: im.Agencies.Inst[i],
				Type: "AgencyInfo"}
			err = stor.cli.UpdateAttributes("mas"+strconv.Itoa(masID)+"im"+strconv.Itoa(imID)+
				"agency"+strconv.Itoa(agencyID), attrList, "clonemap")
			if err != nil {
				return
			}
			break
		}
	}
	if newAgency {
		// new agency has to be created
		var agencyCounter int
		agencyCounter, err = stor.getAgencyCounter(masID, imID)
		if err != nil {
			return
		}
		agencyID = agencyCounter
		agencyCounter++
		attrList := orion.AttributeList{Attributes: make(map[string]orion.Attribute)}
		attrList.Attributes["agencycounter"] = orion.Attribute{Value: agencyCounter,
			Type: "Integer"}
		err = stor.cli.UpdateAttributes("mas"+strconv.Itoa(masID)+"im"+strconv.Itoa(imID), attrList,
			"clonemap")
		if err != nil {
			return
		}
		agencyInfo := schemas.AgencyInfo{
			MASID:        masID,
			ImageGroupID: imID,
			ID:           agencyID,
			Name: "mas-" + strconv.Itoa(masID) + "-im-" + strconv.Itoa(imID) +
				"-agency-" + strconv.Itoa(agencyID) + ".mas" + strconv.Itoa(masID) + "agencies",
			// Logger: masConfig.Logger,
			Agents: []int{agentID},
		}
		agencyEntity := orion.Entity{
			ID: "mas" + strconv.Itoa(masID) + "im" + strconv.Itoa(imID) + "agency" +
				strconv.Itoa(agencyID),
			Type:       "Agency",
			Attributes: make(map[string]orion.Attribute),
		}
		agencyEntity.Attributes["info"] = orion.Attribute{Value: agencyInfo, Type: "AgencyInfo"}
		err = stor.cli.PostEntity(agencyEntity, "clonemap")
		if err != nil {
			return
		}
	}

	// store agent info
	info := schemas.AgentInfo{
		Spec:         agentSpec,
		MASID:        masID,
		ImageGroupID: imID,
		AgencyID:     agencyID,
		ID:           agentID,
		Address: schemas.Address{
			Agency: "mas-" + strconv.Itoa(masID) + "-im-" + strconv.Itoa(imID) + "-agency-" +
				strconv.Itoa(agencyID) + ".mas" + strconv.Itoa(masID) + "agencies",
		},
	}
	agentEntity := orion.Entity{
		ID:         "mas" + strconv.Itoa(masID) + "agent" + strconv.Itoa(agentID),
		Type:       "Agent",
		Attributes: make(map[string]orion.Attribute),
	}
	agentEntity.Attributes["info"] = orion.Attribute{Value: info, Type: "AgentInfo"}
	err = stor.cli.PostEntity(agentEntity, "clonemap")

	return
}

// deleteAgent deletes an agent
func (stor *fiwareStorage) deleteAgent(masID int, agentID int) (err error) {
	var agentExist bool
	agentExist, err = stor.agentExists(masID, agentID)
	if err != nil {
		return
	}
	if !agentExist {
		err = errors.New("Agent does not exist")
		return
	}
	var info schemas.AgentInfo
	info, err = stor.getAgentInfo(masID, agentID)
	info.Address.Agency = ""
	info.Status.Code = status.Terminated
	info.Status.LastUpdate = time.Now()

	attrList := orion.AttributeList{Attributes: make(map[string]orion.Attribute)}
	attrList.Attributes["info"] = orion.Attribute{Value: info, Type: "AgentInfo"}
	err = stor.cli.UpdateAttributes("mas"+strconv.Itoa(masID)+"agent"+strconv.Itoa(agentID),
		attrList, "clonemap")

	return
}

func (stor *fiwareStorage) masExists(masID int) (exists bool, err error) {
	exists = false
	var masCounter int
	masCounter, err = stor.getMasCounter()
	if err != nil {
		return
	}
	if masID >= masCounter {
		return
	}
	exists = true
	return
}

func (stor *fiwareStorage) agentExists(masID int, agentID int) (exists bool, err error) {
	exists = false
	var agentCounter int
	agentCounter, err = stor.getAgentCounter(masID)
	if err != nil {
		if err.Error() == "notexist" {
			err = nil
		}
		return
	}
	if agentID >= agentCounter {
		return
	}
	exists = true
	return
}

func (stor *fiwareStorage) imExists(masID int, imID int) (exists bool, err error) {
	exists = false
	var imCounter int
	imCounter, err = stor.getImCounter(masID)
	if err != nil {
		if err.Error() == "notexist" {
			err = nil
		}
		return
	}
	if imID >= imCounter {
		return
	}
	exists = true
	return
}

func (stor *fiwareStorage) agencyExists(masID int, imID int, agencyID int) (exists bool, err error) {
	exists = false
	var agencyCounter int
	agencyCounter, err = stor.getAgencyCounter(masID, imID)
	if err != nil {
		if err.Error() == "notexist" {
			err = nil
		}
		return
	}
	if agencyID >= agencyCounter {
		return
	}
	exists = true
	return
}

// getMasCounter returns the mas counter
func (stor *fiwareStorage) getMasCounter() (counter int, err error) {
	var attr orion.Attribute
	attr, err = stor.cli.GetAttribute("clonemap", "mascounter", "clonemap")
	if err != nil {
		return
	}
	mascounter, ok := attr.Value.(int)
	if !ok {
		// temp, ok := attr.Value.(float64)
		// if !ok {
		err = errors.New("unknown attribute value")
		return
		// }
		// mascounter = int(temp)
	}
	counter = mascounter

	return
}

// getAgentCounter returns the agent counter
func (stor *fiwareStorage) getAgentCounter(masID int) (counter int, err error) {
	// check if mas exists
	var masExist bool
	masExist, err = stor.masExists(masID)
	if err != nil {
		return
	}
	if !masExist {
		err = errors.New("notexist")
		return
	}

	var attr orion.Attribute
	attr, err = stor.cli.GetAttribute("mas"+strconv.Itoa(masID), "agentcounter", "clonemap")
	if err != nil {
		return
	}
	var agentCounter int
	agentCounter, ok := attr.Value.(int)
	if !ok {
		err = errors.New("unknown attribute value")
		return
	}
	counter = agentCounter

	return
}

// getImCounter returns the im counter
func (stor *fiwareStorage) getImCounter(masID int) (counter int, err error) {
	// check if mas exists
	var masExist bool
	masExist, err = stor.masExists(masID)
	if err != nil {
		return
	}
	if !masExist {
		err = errors.New("notexist")
		return
	}

	var attr orion.Attribute
	attr, err = stor.cli.GetAttribute("mas"+strconv.Itoa(masID), "imcounter", "clonemap")
	if err != nil {
		return
	}
	var imCounter int
	imCounter, ok := attr.Value.(int)
	if !ok {
		err = errors.New("unknown attribute value")
		return
	}
	counter = imCounter

	return
}

// getAgencyCounter returns the agency counter
func (stor *fiwareStorage) getAgencyCounter(masID int, imID int) (counter int, err error) {
	// check if im and mas exists
	var imExist bool
	imExist, err = stor.imExists(masID, imID)
	if err != nil {
		return
	}
	if !imExist {
		err = errors.New("notexist")
		return
	}

	var attr orion.Attribute
	attr, err = stor.cli.GetAttribute("mas"+strconv.Itoa(masID)+"im"+strconv.Itoa(imID),
		"agencycounter", "clonemap")
	if err != nil {
		return
	}
	var agencyCounter int
	agencyCounter, ok := attr.Value.(int)
	if !ok {
		err = errors.New("unknown attribute value")
		return
	}
	counter = agencyCounter

	return
}

// extractAttributeValue extracts the value of an orion attribute
func extractAttributeValue(attr orion.Attribute, t interface{}) (err error) {
	b, ok := attr.Value.(json.RawMessage)
	if !ok {
		err = errors.New("unknown attribute value")
		return
	}
	err = json.Unmarshal(b, t)
	if err != nil {
		return
	}
	return
}

// newFiwareStorage returns Storage interface with fiwareStorage type
func newFiwareStorage(logErr *log.Logger) (stor storage, err error) {
	temp := fiwareStorage{logError: logErr}
	temp.cli, err = orion.NewClient("http://orion:1026")
	if err != nil {
		return
	}
	err = temp.initFiware()
	if err != nil {
		return
	}
	stor = &temp
	return
}

// initFiware sets the clonemap version and uptime if not already present
func (stor *fiwareStorage) initFiware() (err error) {
	_, err = stor.cli.GetEntity("clonemap", "clonemap")
	if err != nil {
		cloneMAP := schemas.CloneMAP{Version: "v0.1", Uptime: time.Now()}
		ent := orion.Entity{ID: "clonemap", Type: "instance", Attributes: make(map[string]orion.Attribute)}
		ent.Attributes["info"] = orion.Attribute{Type: "CloneMAP", Value: cloneMAP}
		ent.Attributes["mascounter"] = orion.Attribute{Type: "Integer", Value: 0}
		err = stor.cli.PostEntity(ent, "clonemap")
		if err != nil {
			return
		}
	}

	return
}

// subscribe subscribes to fiware contex broker
func (stor *fiwareStorage) subscribe() (err error) {

	return
}
