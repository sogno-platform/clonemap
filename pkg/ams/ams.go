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

// Package ams provides functionality for the Agent Management System. It provides an API for user
// interaction as well as for other MAP components. It also takes care of interacting with the
// underlying cluster (local or Kubernetes) and stores MAP related information (local or etcd)
package ams

import (
	"errors"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	agcli "git.rwth-aachen.de/acs/public/cloud/mas/clonemap/pkg/agency/client"
	dfcli "git.rwth-aachen.de/acs/public/cloud/mas/clonemap/pkg/df/client"
	"git.rwth-aachen.de/acs/public/cloud/mas/clonemap/pkg/schemas"
)

// AMS contains storage and deployment object
type AMS struct {
	stor     storage     // interface for local or distributed storage
	depl     deployment  // interface for local or cloud deployment
	logInfo  *log.Logger // logger for info logging
	logError *log.Logger // logger for error logging
}

// StartAMS starts an AMS instance. It initializes the cluster and storage object and starts API
// server.
func StartAMS() (err error) {
	ams := &AMS{logError: log.New(os.Stderr, "[ERROR] ", log.LstdFlags)}
	// create storage and deployment object according to specified deployment type
	err = ams.init()
	if err != nil {
		return
	}
	cmap := schemas.CloneMAP{
		Version: "v0.1",
		Uptime:  time.Now(),
	}
	ams.stor.setCloneMAPInfo(cmap)
	// start to listen and serve requests
	err = ams.listen()
	return
}

// init initializes deployment and storage. The deployment type is read from an environment
// variable. It also determines the address suffix that is used to reach other clonemap apps.
func (ams *AMS) init() (err error) {
	//fmt.Println("Getting deployment type")
	// ams.addrSuffix = os.Getenv("CLONEMAP_SUFFIX")
	logType := os.Getenv("CLONEMAP_LOG_LEVEL")
	switch logType {
	case "info":
		ams.logInfo = log.New(os.Stdout, "[INFO] ", log.LstdFlags)
	case "error":
		ams.logInfo = log.New(ioutil.Discard, "", log.LstdFlags)
	default:
		err = errors.New("Wrong log type: " + logType)
		return
	}
	ams.logInfo.Println("Starting AMS")

	deplType := os.Getenv("CLONEMAP_DEPLOYMENT_TYPE")
	switch deplType {
	case "local":
		ams.logInfo.Println("Local deployment")
		ams.depl, err = newLocalDeployment()
	case "minikube":
		ams.logInfo.Println("Kubernetes deployment")
		ams.depl, err = newKubeDeployment(deplType)
	case "production":
		ams.logInfo.Println("Kubernetes deployment")
		ams.depl, err = newKubeDeployment(deplType)
	default:
		err = errors.New("Wrong deployment type: " + deplType)
		return
	}
	storType := os.Getenv("CLONEMAP_STORAGE_TYPE")
	switch storType {
	case "local":
		ams.logInfo.Println("Local storage")
		ams.stor = newLocalStorage()
	case "etcd":
		if deplType == "local" {
			err = errors.New("etcd storage can not be used with local deployment")
			return
		}
		ams.logInfo.Println("ectd storage")
		ams.stor, err = newEtcdStorage(ams.logError)
	case "fiware":
		ams.logInfo.Println("FiWare storage")
		ams.stor, err = newFiwareStorage(ams.logError)
	default:
		err = errors.New("Wrong storage type: " + storType)
	}
	return
}

// getCloneMAPInfo returns info about clonemap
func (ams *AMS) getCloneMAPInfo() (ret schemas.CloneMAP, err error) {
	ret, err = ams.stor.getCloneMAPInfo()
	return
}

// getMASs returns specs of all MAS
func (ams *AMS) getMASs() (ret schemas.MASs, err error) {
	ret, err = ams.stor.getMASs()
	return
}

// getMASInfo returns info of one MAS
func (ams *AMS) getMASInfo(masID int) (ret schemas.MASInfo, err error) {
	ret, err = ams.stor.getMASInfo(masID)
	return
}

// getAgents returns specs of all agents in MAS
func (ams *AMS) getAgents(masID int) (ret schemas.Agents, err error) {
	ret, err = ams.stor.getAgents(masID)
	return
}

// getAgentInfo returns info of one or multiple agents
func (ams *AMS) getAgentInfo(masID int, agentID int) (ret schemas.AgentInfo, err error) {
	ret, err = ams.stor.getAgentInfo(masID, agentID)
	return
}

// getAgentAddress returns address of one agent
func (ams *AMS) getAgentAddress(masID int, agentID int) (ret schemas.Address, err error) {
	ret, err = ams.stor.getAgentAddress(masID, agentID)
	return
}

// updateAgentAddress sets address of agent
func (ams *AMS) updateAgentAddress(masID int, agentID int, address schemas.Address) (err error) {
	err = ams.stor.setAgentAddress(masID, agentID, address)
	return
}

// getAgencies returns specs of all agencies in MAS
func (ams *AMS) getAgencies(masID int) (ret schemas.Agencies, err error) {
	ret, err = ams.stor.getAgencies(masID)
	return
}

// getAgencyInfoFull returns status of one agency
func (ams *AMS) getAgencyInfoFull(masID int, imID int, agencyID int) (ret schemas.AgencyInfoFull,
	err error) {
	ret, err = ams.stor.getAgencyInfoFull(masID, imID, agencyID)
	return
}

// createMAS creates a new mas according to masconfig
func (ams *AMS) createMAS(masSpec schemas.MASSpec) (err error) {
	// fill masInfo
	var masInfo schemas.MASInfo
	var numAgencies []int
	masInfo, numAgencies, err = ams.configureMAS(masSpec)
	if err != nil {
		return
	}

	// safe mas in storage and get ID
	var masID int
	masID, err = ams.stor.registerMAS()
	ams.logInfo.Println("Create new MAS with ID ", masID)
	if err != nil {
		return
	}

	go ams.startMAS(masID, masInfo, numAgencies)

	return
}

// startMAS starts the MAS
func (ams *AMS) startMAS(masID int, masInfo schemas.MASInfo, numAgencies []int) (err error) {
	err = ams.stor.storeMAS(masID, masInfo)
	if err != nil {
		ams.logError.Println(err.Error())
		return
	}
	ams.logInfo.Println("Stored MAS data")
	if os.Getenv("CLONEMAP_DEPLOYMENT_TYPE") == "local" {
		_, err = dfcli.PostGraph(masID, masInfo.Graph)
		if err != nil {
			ams.logError.Println(err.Error())
			return
		}
	}

	// deploy containers
	err = ams.depl.newMAS(masID, masInfo.ImageGroups, masInfo.Config.Logging,
		masInfo.Config.MQTT, masInfo.Config.DF)
	if err != nil {
		ams.logError.Println(err.Error())
		return
	}
	ams.logInfo.Println("Started agencies")

	return
}

// configureMAS fills the missing configuration as agencies, agent ids and addresses
func (ams *AMS) configureMAS(masSpec schemas.MASSpec) (masInfo schemas.MASInfo,
	numAgencies []int, err error) {
	// extract all image groups
	imageTemp := make(map[string]interface{})
	for i := range masSpec.ImageGroups {
		if _, ok := imageTemp[masSpec.ImageGroups[i].Config.Image]; ok {
			// image already exists
			err = errors.New("invalid masSpec; two or more groups with same image")
			return
		}
		imGroupInfo := schemas.ImageGroupInfo{
			Config: masSpec.ImageGroups[i].Config,
			ID:     i,
		}
		masInfo.ImageGroups.Inst = append(masInfo.ImageGroups.Inst, imGroupInfo)
		masInfo.ImageGroups.Counter++
		imageTemp[masSpec.ImageGroups[i].Config.Image] = nil
	}

	// MAS configuration
	masInfo.Config = masSpec.Config
	masInfo.Graph = masSpec.Graph

	// total number of agents and total number of agencies
	masInfo.Agents.Counter = 0
	numAgencies = make([]int, masInfo.ImageGroups.Counter, masInfo.ImageGroups.Counter)
	for i := range masSpec.ImageGroups {
		masInfo.Agents.Counter += len(masSpec.ImageGroups[i].Agents)
		num := len(masSpec.ImageGroups[i].Agents) / masSpec.Config.NumAgentsPerAgency
		if len(masSpec.ImageGroups[i].Agents)%masSpec.Config.NumAgentsPerAgency > 0 {
			num++
		}
		masInfo.ImageGroups.Inst[i].Agencies.Inst = make([]schemas.AgencyInfo, num,
			num)
		masInfo.ImageGroups.Inst[i].Agencies.Counter = num
		numAgencies[i] = num
	}
	masInfo.Agents.Inst = make([]schemas.AgentInfo, masInfo.Agents.Counter,
		masInfo.Agents.Counter)

	// empty graph?
	if len(masInfo.Graph.Node) == 0 {
		masInfo.Graph.Node = append(masInfo.Graph.Node, schemas.Node{ID: 0})
	}

	// agent configuration
	agentID := 0
	for i := range masSpec.ImageGroups {
		for j := range masSpec.ImageGroups[i].Agents {
			masInfo.Agents.Inst[agentID].Spec = masSpec.ImageGroups[i].Agents[j]
			masInfo.Agents.Inst[agentID].ID = agentID
			masInfo.Agents.Inst[agentID].AgencyID = j / masSpec.Config.NumAgentsPerAgency
			masInfo.Agents.Inst[agentID].ImageGroupID = i
			masInfo.Agents.Inst[agentID].Address.Agency = "-im-" + strconv.Itoa(i) + "-agency-" +
				strconv.Itoa(j/masSpec.Config.NumAgentsPerAgency)
			for j := range masInfo.Graph.Node {
				if masInfo.Graph.Node[j].ID == masInfo.Agents.Inst[i].Spec.NodeID {
					masInfo.Graph.Node[j].Agent = append(masInfo.Graph.Node[j].Agent,
						masInfo.Agents.Inst[i].ID)
					break
				}
			}
			agentID++
		}
	}

	// agency configuration
	agentCounterTot := 0
	for i := range masSpec.ImageGroups {
		agentCounter := 0
		for j := 0; j < numAgencies[i]; j++ {
			agencyInfo := schemas.AgencyInfo{
				ImageGroupID: i,
				ID:           j,
				Logger:       masSpec.Config.Logger,
				Name:         "-im-" + strconv.Itoa(i) + "-agency-" + strconv.Itoa(j),
			}
			for k := 0; k < masSpec.Config.NumAgentsPerAgency; k++ {
				if agentCounter >= len(masSpec.ImageGroups[i].Agents) {
					break
				}
				agencyInfo.Agents = append(agencyInfo.Agents, agentCounterTot)
				agentCounter++
				agentCounterTot++
			}
			masInfo.ImageGroups.Inst[i].Agencies.Inst[j] = agencyInfo
		}
	}
	return
}

// removeMAS removes specified mas if it exists
func (ams *AMS) removeMAS(masID int) (err error) {
	err = ams.depl.deleteMAS(masID)
	if err != nil {
		return
	}
	err = ams.stor.deleteMAS(masID)
	return
}

// createAgents creates new agents and adds them to an existing mas
func (ams *AMS) createAgents(masID int, groupSpecs []schemas.ImageGroupSpec) (err error) {
	for i := range groupSpecs {
		var newGroup bool
		var imID int
		newGroup, imID, err = ams.stor.registerImageGroup(masID, groupSpecs[i].Config)
		if err != nil {
			return
		}
		var newAgencies []int
		for j := range groupSpecs[i].Agents {
			var newAgency bool
			var agentID int
			var agencyID int
			newAgency, agentID, agencyID, err = ams.stor.addAgent(masID, imID,
				groupSpecs[i].Agents[j])
			if err != nil {
				return
			}
			if newGroup {
				// continue if group is new group
				continue
			} else if newAgency {
				newAgencies = append(newAgencies, agencyID)
			} else {
				// post agent to running agency if agency is not new
				var agentInfo schemas.AgentInfo
				agentInfo, err = ams.stor.getAgentInfo(masID, agentID)
				if err != nil {
					return
				}
				for k := range newAgencies {
					if agentInfo.AgencyID == newAgencies[k] {
						newAgency = true
					}
				}
				if newAgency {
					// continue if agency is new agency
					continue
				}
				err = ams.postAgentToAgency(agentInfo)
				if err != nil {
					return
				}
			}
		}
		if newGroup {
			var groupInfo schemas.ImageGroupInfo
			groupInfo, err = ams.stor.getGroupInfo(masID, imID)
			if err != nil {
				return
			}
			var masInfo schemas.MASInfo
			masInfo, err = ams.stor.getMASInfo(masID)
			if err != nil {
				return
			}
			err = ams.depl.newImageGroup(masID, groupInfo, masInfo.Config.Logging,
				masInfo.Config.MQTT, masInfo.Config.DF)
			if err != nil {
				return
			}
		} else {
			numNewAgencies := len(newAgencies)
			err = ams.depl.scaleImageGroup(masID, imID, numNewAgencies)
			if err != nil {
				return
			}
		}
	}
	return
}

// createAgent creates a new agent and adds it to an existing mas
func (ams *AMS) createAgent(masID int, agentSpec schemas.AgentSpec) (err error) {
	return
}

// removeAgent removes an agent from the MAS
func (ams *AMS) removeAgent(masID int, agentID int) (err error) {
	err = ams.stor.deleteAgent(masID, agentID)
	if err != nil {
		return
	}
	// TODO delete agent in agency
	return
}

// postAgentToAgency sends a post request to agency with info about agent to start
func (ams *AMS) postAgentToAgency(agentInfo schemas.AgentInfo) (err error) {
	var httpStatus int
	httpStatus, err = agcli.PostAgent(agentInfo.Address.Agency, agentInfo)
	if err != nil {
		return
	}
	if httpStatus == http.StatusCreated {
		// stat := schemas.Status{
		// 	Code:       status.Starting,
		// 	LastUpdate: time.Now(),
		// }
		// ams.updateAgentStatus(agentInfo.Spec.MASID, agentInfo.Spec.ID, stat)
	} else {
		err = errors.New("error posting to agency")
	}
	return
}
