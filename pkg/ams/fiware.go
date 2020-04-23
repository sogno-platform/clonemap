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

import (
	"encoding/json"
	"errors"
	"log"

	"git.rwth-aachen.de/acs/public/cloud/fiware/gofiware/pkg/orion"
	"git.rwth-aachen.de/acs/public/cloud/mas/clonemap/pkg/schemas"
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

// getMASs returns specs of all MAS
func (stor *fiwareStorage) getMASs() (ret schemas.MASs, err error) {

	return
}

// getMASInfo returns info of one MAS
func (stor *fiwareStorage) getMASInfo(masID int) (ret schemas.MASInfo, err error) {

	return
}

// getGroupInfo returns info of one image group
func (stor *fiwareStorage) getGroupInfo(masID int, imID int) (ret schemas.ImageGroupInfo,
	err error) {
	return
}

// getAgents returns specs of all agents in MAS
func (stor *fiwareStorage) getAgents(masID int) (ret schemas.Agents, err error) {

	return
}

// getAgentInfo returns info of one agent
func (stor *fiwareStorage) getAgentInfo(masID int, agentID int) (ret schemas.AgentInfo, err error) {

	return
}

// getAgentAddress returns address of one agent
func (stor *fiwareStorage) getAgentAddress(masID int, agentID int) (ret schemas.Address,
	err error) {

	return
}

// setAgentAddress sets address of agent
func (stor *fiwareStorage) setAgentAddress(masID int, agentID int,
	address schemas.Address) (err error) {

	return
}

// getAgencies returns specs of all agencies in MAS
func (stor *fiwareStorage) getAgencies(masID int) (ret schemas.Agencies, err error) {

	return
}

// getAgencyInfoFull returns status of one agency
func (stor *fiwareStorage) getAgencyInfoFull(masID int, imID int,
	agencyID int) (ret schemas.AgencyInfoFull, err error) {
	return
}

// registerMAS registers a new MAS with the storage and returns its ID
func (stor *fiwareStorage) registerMAS() (masID int, err error) {

	return
}

// storeMAS stores MAS specs
func (stor *fiwareStorage) storeMAS(masID int, masInfo schemas.MASInfo) (err error) {

	return
}

// deleteMAS deletes MAS with specified ID
func (stor *fiwareStorage) deleteMAS(masID int) (err error) {

	return
}

// registerImageGroup registers a new image group with the storage and returns its ID
func (stor *fiwareStorage) registerImageGroup(masID int,
	config schemas.ImageGroupConfig) (newGroup bool, imID int, err error) {
	return
}

// registerAgent registers a new agent with the storage and returns its ID
func (stor *fiwareStorage) registerAgent(masID int, imID int, spec schemas.AgentSpec) (agentID int,
	err error) {
	return
}

// addAgent adds an agent to an exsiting MAS
func (stor *fiwareStorage) addAgent(masID int, imID int,
	agentSpec schemas.AgentSpec) (newAgency bool, agentID int, agencyID int, err error) {
	return
}

func (stor *fiwareStorage) agentExists(masID int, agentID int) (exists bool, err error) {

	return
}

func (stor *fiwareStorage) agencyExists(masID int, agencyID int) (exists bool, err error) {

	return
}

// getMasCounter returns the mas counter
func (stor *fiwareStorage) getMasCounter() (counter int, err error) {

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

	return
}

// subscribe subscribes to fiware contex broker
func (stor *fiwareStorage) subscribe() (err error) {

	return
}
