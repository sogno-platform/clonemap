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

// information storage in etcd for use of clonemap in kubernetes
//
// etcd paths:
//
// ams/data: schemas.CloneMAP
// ams/mas/counter: int (masCounter)
// ams/mas/<masID>/config schemas.MASConfig
// ams/mas/<masID>/status schemas.Status
// ams/mas/<masID>/imcounter int (imCounter)
// ams/mas/<masID>/im/<imID>/config ImageGroupConfig
// ams/mas/<masID>/im/<imID>/agencycounter int(agencyCounter)
// ams/mas/<masID>/im/<imID>/agency/<agencyID>: schemas.AgencyInfo
// ams/mas/<masID>/agentcounter int (agentCounter)
// ams/mas/<masID>/agent/<agentID>: schemas.AgentInfo
//
// df/graph/<masID>: schemas.Graph

package ams

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"strconv"
	"strings"
	"sync"
	"time"

	"git.rwth-aachen.de/acs/public/cloud/mas/clonemap/pkg/schemas"
	"git.rwth-aachen.de/acs/public/cloud/mas/clonemap/pkg/status"
	"go.etcd.io/etcd/clientv3"
	"go.etcd.io/etcd/clientv3/concurrency"
	"go.etcd.io/etcd/mvcc/mvccpb"
)

// etcd storage
type etcdStorage struct {
	config        clientv3.Config  // configuration of client
	client        *clientv3.Client // client
	localStorage                   // local cache
	verMASCounter int
	verMAS        []masVersion // required to check if values received from etcd watcher are newer
	logError      *log.Logger  // logger for error logging
}

// version of mas keys in etcd
type masVersion struct {
	status       int
	config       int
	groupCounter int
	imGroups     []imGroupVersion
	agentCounter int
	agents       []int
	graph        int
}

// version of image group
type imGroupVersion struct {
	config        int
	agencyCounter int
	agencies      []int
}

// setCloneMAPInfo sets info specific to running clonemap instance
func (stor *etcdStorage) setCloneMAPInfo(cloneMAP schemas.CloneMAP) (err error) {
	err = stor.etcdPutResource("ams/data", cloneMAP)
	return
}

// setAgentAddress sets address of agent
func (stor *etcdStorage) setAgentAddress(masID int, agentID int,
	address schemas.Address) (err error) {
	err = stor.etcdPutResource("ams/mas/"+strconv.Itoa(masID)+"/agent/"+strconv.Itoa(agentID)+
		"/address", address)
	return
}

// registerMAS registers a new MAS with the storage and returns its ID
func (stor *etcdStorage) registerMAS() (masID int, err error) {
	// store new ams and determine ID
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	// use STM for atomic puts and retry in case values have been altered during function execution
	_, err = concurrency.NewSTMRepeatable(ctx, stor.client, func(s concurrency.STM) error {
		// get info about number of running mas
		var masCounter int
		err = json.Unmarshal([]byte(s.Get("ams/mas/counter")), &masCounter)
		if err != nil {
			return err
		}
		masID = masCounter
		masCounter++
		// update mas counter in etcd
		var res []byte
		res, err = json.Marshal(masCounter)
		if err != nil {
			return err
		}
		s.Put("ams/mas/counter", string(res))
		return err
	})
	cancel()

	return
}

// storeMAS stores MAS specs
func (stor *etcdStorage) storeMAS(masID int, masInfo schemas.MASInfo) (err error) {
	var tempConfig schemas.MASConfig
	_, err = stor.etcdGetResource("ams/mas/"+strconv.Itoa(masID)+"/config", &tempConfig)
	if err == nil {
		// resource already exists
		err = errors.New("MAS already exists")
		return
	}

	newMAS := createMASStorage(masID, masInfo)
	err = stor.etcdPutResource("ams/mas/"+strconv.Itoa(masID)+"/config", newMAS.Config)
	if err != nil {
		return
	}
	err = stor.etcdPutResource("ams/mas/"+strconv.Itoa(masID)+"/status", newMAS.Status)
	if err != nil {
		return
	}
	err = stor.etcdPutResource("ams/mas/"+strconv.Itoa(masID)+"/imcounter",
		newMAS.ImageGroups.Counter)
	if err != nil {
		return
	}
	err = stor.etcdPutResource("ams/mas/"+strconv.Itoa(masID)+"/agentcounter",
		newMAS.Agents.Counter)
	if err != nil {
		return
	}
	err = stor.etcdPutResource("df/graph/"+strconv.Itoa(masID),
		newMAS.Graph)
	if err != nil {
		return
	}

	err = stor.uploadAgentInfo(newMAS)
	if err != nil {
		return
	}

	err = stor.uploadImGroupInfo(newMAS)
	if err != nil {
		return
	}
	return
}

// uploadAgentInfo puts all AgentInfo of a newly created MAS to etcd
func (stor *etcdStorage) uploadAgentInfo(newMAS schemas.MASInfo) (err error) {
	agentIndex := 0
	for {
		numAgInTrans := 100
		if newMAS.Agents.Counter-agentIndex < numAgInTrans {
			numAgInTrans = newMAS.Agents.Counter - agentIndex
		}
		Ops := make([]clientv3.Op, numAgInTrans, numAgInTrans)
		// put all agent structs together
		for i := 0; i < numAgInTrans; i++ {
			var res []byte
			res, err = json.Marshal(newMAS.Agents.Inst[agentIndex])
			if err != nil {
				return
			}
			Ops[i] = clientv3.OpPut("ams/mas/"+strconv.Itoa(newMAS.ID)+"/agent/"+
				strconv.Itoa(agentIndex), string(res))
			agentIndex++
		}
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		cond := clientv3.Compare(clientv3.Version("ams/mas/"+strconv.Itoa(newMAS.ID)+
			"/agentcounter"), ">", 0)
		_, err = stor.client.Txn(ctx).If(cond).Then(Ops...).Commit()
		cancel()
		if err != nil {
			return
		}
		if agentIndex >= newMAS.Agents.Counter {
			break
		}
		time.Sleep(50 * time.Millisecond)
	}
	return
}

// uploadImGroupInfo puts all AgencyInfo of a newly created MAS to etcd
func (stor *etcdStorage) uploadImGroupInfo(newMAS schemas.MASInfo) (err error) {
	for i := range newMAS.ImageGroups.Inst {
		err = stor.etcdPutResource("ams/mas/"+strconv.Itoa(newMAS.ID)+"/im/"+strconv.Itoa(i)+
			"/config", newMAS.ImageGroups.Inst[i].Config)
		if err != nil {
			return
		}
		err = stor.etcdPutResource("ams/mas/"+strconv.Itoa(newMAS.ID)+"/im/"+strconv.Itoa(i)+
			"/agencycounter", newMAS.ImageGroups.Inst[i].Agencies.Counter)
		if err != nil {
			return
		}
		agencyIndex := 0
		for {
			numAgInTrans := 100
			if newMAS.ImageGroups.Inst[i].Agencies.Counter-agencyIndex < numAgInTrans {
				numAgInTrans = newMAS.ImageGroups.Inst[i].Agencies.Counter - agencyIndex
			}
			Ops := make([]clientv3.Op, numAgInTrans, numAgInTrans)
			// put all agencies structs together
			for j := 0; j < numAgInTrans; j++ {
				var res []byte
				res, err = json.Marshal(newMAS.ImageGroups.Inst[i].Agencies.Inst[agencyIndex])
				if err != nil {
					return
				}
				Ops[j] = clientv3.OpPut("ams/mas/"+strconv.Itoa(newMAS.ID)+"/im/"+strconv.Itoa(i)+"/agency/"+
					strconv.Itoa(agencyIndex), string(res))
				agencyIndex++
			}
			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			cond := clientv3.Compare(clientv3.Version("ams/mas/"+strconv.Itoa(newMAS.ID)+"/im/"+
				strconv.Itoa(i)+"/agencycounter"), ">", 0)
			_, err = stor.client.Txn(ctx).If(cond).Then(Ops...).Commit()
			cancel()
			if err != nil {
				return
			}
			if agencyIndex >= newMAS.ImageGroups.Inst[i].Agencies.Counter {
				break
			}
			time.Sleep(50 * time.Millisecond)
		}
	}
	return
}

// deleteMAS deletes MAS with specified ID
func (stor *etcdStorage) deleteMAS(masID int) (err error) {
	stat := schemas.Status{
		Code:       status.Terminated,
		LastUpdate: time.Now(),
	}
	err = stor.etcdPutResource("ams/mas/"+strconv.Itoa(masID)+"/status", stat)
	return
}

// registerImageGroup registers a new image group with the storage and returns its ID
func (stor *etcdStorage) registerImageGroup(masID int,
	config schemas.ImageGroupConfig) (newGroup bool, imID int, err error) {
	stor.mutex.Lock()
	if len(stor.mas)-1 < masID {
		stor.mutex.Unlock()
		err = errors.New("MAS does not exist")
		return
	}
	stor.mutex.Unlock()

	// store new image group and determine ID
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	// use STM for atomic puts and retry in case values have been altered during function execution
	_, err = concurrency.NewSTMRepeatable(ctx, stor.client, func(s concurrency.STM) error {
		// get info about number of running mas
		stor.mutex.Lock()
		for i := range stor.mas[masID].ImageGroups.Inst {
			if stor.mas[masID].ImageGroups.Inst[i].Config.Image == config.Image {
				stor.mutex.Unlock()
				imID = i
				newGroup = false
				ctx.Done()
				cancel()
				return err
			}
		}
		stor.mutex.Unlock()
		newGroup = true

		var imCounter int
		err = json.Unmarshal([]byte(s.Get("ams/mas/"+strconv.Itoa(masID)+"/imcounter")), &imCounter)
		if err != nil {
			return err
		}
		imID = imCounter
		imCounter++
		// update mas counter in etcd
		var res []byte
		res, err = json.Marshal(imCounter)
		if err != nil {
			return err
		}
		s.Put("ams/mas/"+strconv.Itoa(masID)+"/imcounter", string(res))
		return err
	})
	cancel()

	if err != nil {
		if err.Error() == "context canceled" {
			err = nil
		}
	}

	if newGroup {
		err = stor.etcdPutResource("ams/mas/"+strconv.Itoa(masID)+"/im/"+strconv.Itoa(imID)+
			"/config", config)
		if err != nil {
			return
		}
		err = stor.etcdPutResource("ams/mas/"+strconv.Itoa(masID)+"/im/"+strconv.Itoa(imID)+
			"/agencycounter", 0)
		if err != nil {
			return
		}
	}

	return
}

// addAgent adds an agent to an existing MAS
func (stor *etcdStorage) addAgent(masID int, imID int,
	agentSpec schemas.AgentSpec) (newAgency bool, agentID int, agencyID int, err error) {
	stor.mutex.Lock()
	if len(stor.mas)-1 < masID {
		stor.mutex.Unlock()
		err = errors.New("MAS does not exist")
		return
	}

	if len(stor.mas[masID].ImageGroups.Inst)-1 < imID {
		stor.mutex.Unlock()
		err = errors.New("ImageGroup does not exist")
		return
	}
	numAgentsPerAgency := stor.mas[masID].Config.NumAgentsPerAgency
	stor.mutex.Unlock()

	// register agent
	agentID, err = stor.registerAgent(masID, imID, agentSpec)
	if err != nil {
		return
	}
	// store new image group and determine ID
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	// use STM for atomic puts and retry in case values have been altered during function execution
	_, err = concurrency.NewSTMRepeatable(ctx, stor.client, func(s concurrency.STM) error {
		newAgency = true
		var agencyCounter int
		err = json.Unmarshal([]byte(s.Get("ams/mas/"+strconv.Itoa(masID)+"/im/"+strconv.Itoa(imID)+
			"/agencycounter")), &agencyCounter)
		if err != nil {
			return err
		}
		for i := 0; i < agencyCounter; i++ {
			var agencyInfo schemas.AgencyInfo
			err = json.Unmarshal([]byte(s.Get("ams/mas/"+strconv.Itoa(masID)+"/im/"+
				strconv.Itoa(imID)+"/agency/"+strconv.Itoa(i))), &agencyInfo)
			if err != nil {
				return err
			}
			if len(agencyInfo.Agents) < numAgentsPerAgency {
				// there exists an agency with space left
				agencyInfo.Agents = append(agencyInfo.Agents, agentID)
				var res []byte
				res, err = json.Marshal(agencyInfo)
				if err != nil {
					return err
				}
				s.Put("ams/mas/"+strconv.Itoa(masID)+"/im/"+strconv.Itoa(imID)+"/agency/"+
					strconv.Itoa(i), string(res))
				agencyID = i
				newAgency = false
				break
			}
		}
		if newAgency {
			// new agency has to be created
			agencyID = agencyCounter
			agencyCounter++
			// update mas counter in etcd
			var res []byte
			res, err = json.Marshal(agencyCounter)
			if err != nil {
				return err
			}
			s.Put("ams/mas/"+strconv.Itoa(masID)+"/im/"+strconv.Itoa(imID)+"/agencycounter",
				string(res))

			agencyInfo := schemas.AgencyInfo{
				MASID:        masID,
				ImageGroupID: imID,
				ID:           agencyID,
				Name: "mas-" + strconv.Itoa(masID) + "-im-" + strconv.Itoa(imID) +
					"-agency-" + strconv.Itoa(agencyID) + ".mas" + strconv.Itoa(masID) + "agencies",
				Logger: stor.mas[masID].Config.Logger,
				Agents: []int{agentID},
			}

			res, err = json.Marshal(agencyInfo)
			if err != nil {
				return err
			}
			s.Put("ams/mas/"+strconv.Itoa(masID)+"/im/"+strconv.Itoa(imID)+"/agency/"+
				strconv.Itoa(agencyID), string(res))
		}

		return err
	})
	cancel()

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

	err = stor.etcdPutResource("ams/mas/"+strconv.Itoa(masID)+"/agent/"+strconv.Itoa(agentID), info)
	if err != nil {
		return
	}

	return
}

// registerAgent registers a new agent with the storage and returns its ID
func (stor *etcdStorage) registerAgent(masID int, imID int, spec schemas.AgentSpec) (agentID int,
	err error) {
	stor.mutex.Lock()
	if len(stor.mas)-1 < masID {
		stor.mutex.Unlock()
		err = errors.New("MAS does not exist")
		return
	}
	stor.mutex.Unlock()

	// store new agent and determine ID
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	// use STM for atomic puts and retry in case values have been altered during function execution
	_, err = concurrency.NewSTMRepeatable(ctx, stor.client, func(s concurrency.STM) error {
		// get info about number of running agents

		var agentCounter int
		err = json.Unmarshal([]byte(s.Get("ams/mas/"+strconv.Itoa(masID)+"/agentcounter")), &agentCounter)
		if err != nil {
			return err
		}
		agentID = agentCounter
		agentCounter++
		// update mas counter in etcd
		var res []byte
		res, err = json.Marshal(agentCounter)
		if err != nil {
			return err
		}
		s.Put("ams/mas/"+strconv.Itoa(masID)+"/agentcounter", string(res))
		return err
	})
	cancel()

	return
}

// newEtcdStorage returns Storage interface with etcdStorage type
func newEtcdStorage(logErr *log.Logger) (stor storage, err error) {
	temp := etcdStorage{logError: logErr}
	temp.config.Endpoints = []string{"http://etcd-cluster-client:2379"}
	temp.config.DialTimeout = 10 * time.Second
	temp.client, err = clientv3.New(temp.config)
	if err != nil {
		return
	}
	err = temp.initEtcd()
	if err != nil {
		return
	}
	err = temp.initCache()
	if err != nil {
		return
	}
	go temp.handleAMSEvents()
	go temp.handleGraphEvents()
	stor = &temp
	return
}

// initEtcd sets the clonemap version and uptime if not already present
func (stor *etcdStorage) initEtcd() (err error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	cloneMAP := schemas.CloneMAP{Version: "v0.1", Uptime: time.Now()}
	var res []byte
	res, err = json.Marshal(cloneMAP)
	req := clientv3.OpPut("ams/data", string(res))
	cond := clientv3.Compare(clientv3.Version("ams/data"), "=", 0)
	_, err = stor.client.Txn(ctx).If(cond).Then(req).Commit()
	if err == nil {
		req = clientv3.OpPut("ams/mas/counter", strconv.Itoa(0))
		cond = clientv3.Compare(clientv3.Version("ams/mas/counter"), "=", 0)
		_, err = stor.client.Txn(ctx).If(cond).Then(req).Commit()
	}
	cancel()
	return
}

// initCache initializes the cached local storage
func (stor *etcdStorage) initCache() (err error) {
	stor.mutex = &sync.Mutex{}
	// get cloneMAP info and mas counter
	_, err = stor.etcdGetResource("ams/data", &stor.cloneMAP)
	if err != nil {
		return
	}
	stor.verMASCounter, err = stor.etcdGetResource("ams/mas/counter", &stor.masCounter)
	if err != nil {
		return
	}
	stor.mas = make([]schemas.MASInfo, stor.masCounter, stor.masCounter)
	stor.verMAS = make([]masVersion, stor.masCounter, stor.masCounter)
	// get data for each mas
	for i := 0; i < stor.masCounter; i++ {
		// MAS config
		stor.verMAS[i].config, err = stor.etcdGetResource("ams/mas/"+strconv.Itoa(i)+"/config",
			&stor.mas[i].Config)
		if err != nil {
			return
		}
		// MAS status
		stor.verMAS[i].status, err = stor.etcdGetResource("ams/mas/"+strconv.Itoa(i)+"/status",
			&stor.mas[i].Status)
		if err != nil {
			return
		}
		// MAS graph
		stor.verMAS[i].graph, err = stor.etcdGetResource("df/graph/"+strconv.Itoa(i),
			&stor.mas[i].Graph)
		if err != nil {
			return
		}
		// MAS agencies
		err = stor.initMASImGroups(i)
		if err != nil {
			return
		}
		// MAS agents
		err = stor.initMASAgents(i)
		if err != nil {
			return
		}
	}
	return
}

// initMASImGroups retrieves data of agencies in a mas from etcd
func (stor *etcdStorage) initMASImGroups(masID int) (err error) {
	// MAS im group counter
	stor.verMAS[masID].groupCounter, err = stor.etcdGetResource("ams/mas/"+strconv.Itoa(masID)+"/imcounter",
		&stor.mas[masID].ImageGroups.Counter)
	if err != nil {
		return
	}
	stor.mas[masID].ImageGroups.Inst = make([]schemas.ImageGroupInfo,
		stor.mas[masID].ImageGroups.Counter, stor.mas[masID].ImageGroups.Counter)
	stor.verMAS[masID].imGroups = make([]imGroupVersion, stor.mas[masID].ImageGroups.Counter,
		stor.mas[masID].ImageGroups.Counter)

	for i := 0; i < stor.mas[masID].ImageGroups.Counter; i++ {
		stor.verMAS[masID].imGroups[i].agencyCounter, err = stor.etcdGetResource("ams/mas/"+
			strconv.Itoa(masID)+"/im/"+strconv.Itoa(i)+"/agencycounter",
			&stor.mas[masID].ImageGroups.Inst[i].Agencies.Counter)
		if err != nil {
			return
		}
		stor.verMAS[masID].imGroups[i].config, err = stor.etcdGetResource("ams/mas/"+
			strconv.Itoa(masID)+"/im/"+strconv.Itoa(i)+"/config",
			&stor.mas[masID].ImageGroups.Inst[i].Config)
		if err != nil {
			return
		}
		stor.mas[masID].ImageGroups.Inst[i].Agencies.Inst = make([]schemas.AgencyInfo,
			stor.mas[masID].ImageGroups.Inst[i].Agencies.Counter,
			stor.mas[masID].ImageGroups.Inst[i].Agencies.Counter)
		stor.verMAS[masID].imGroups[i].agencies = make([]int,
			stor.mas[masID].ImageGroups.Inst[i].Agencies.Counter,
			stor.mas[masID].ImageGroups.Inst[i].Agencies.Counter)

		// get info of all agencies and loop through
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		resp := &clientv3.GetResponse{}
		resp, err = stor.client.Get(ctx, "ams/mas/"+strconv.Itoa(masID)+"/im/"+strconv.Itoa(i)+
			"/agency", clientv3.WithPrefix())
		if err != nil {
			cancel()
			return
		}
		for j := range resp.Kvs {
			temp := strings.Split(string(resp.Kvs[j].Key), "/")
			if len(temp) != 7 {
				continue
			}
			var agencyID int
			agencyID, err = strconv.Atoi(temp[6])
			if err != nil {
				cancel()
				return
			}
			if agencyID >= stor.mas[masID].ImageGroups.Inst[i].Agencies.Counter {
				continue
			}
			err = json.Unmarshal(resp.Kvs[j].Value,
				&stor.mas[masID].ImageGroups.Inst[i].Agencies.Inst[agencyID])
			if err != nil {
				cancel()
				return
			}
			stor.verMAS[masID].imGroups[i].agencies[agencyID] = int(resp.Kvs[j].Version)
			// stor.mas[masID].ImageGroups.Inst[i].Agencies.Inst[agencyID].MASID = masID
			// stor.mas[masID].ImageGroups.Inst[i].Agencies.Inst[agencyID].ID = agencyID
			// stor.mas[masID].ImageGroups.Inst[i].Agencies.Inst[agencyID].ImageGroupID =
			// 	i
			// stor.mas[masID].ImageGroups.Inst[i].Agencies.Inst[agencyID].Logger =
			// 	stor.mas[masID].Config.Logger
			// stor.mas[masID].ImageGroups.Inst[i].Agencies.Inst[agencyID].Name =
			// 	"mas-" + strconv.Itoa(masID) + "-im-" + strconv.Itoa(i) +
			// 		"-agency-" + strconv.Itoa(agencyID) + ".mas" + strconv.Itoa(masID) +
			// 		"agencies"
		}
		cancel()
	}
	return
}

// initMASAgents retrieves data of agents in a mas from etcd
func (stor *etcdStorage) initMASAgents(masID int) (err error) {
	// MAS agent counter
	stor.verMAS[masID].agentCounter, err = stor.etcdGetResource("ams/mas/"+strconv.Itoa(masID)+
		"/agentcounter", &stor.mas[masID].Agents.Counter)
	if err != nil {
		return
	}
	stor.mas[masID].Agents.Inst = make([]schemas.AgentInfo, stor.mas[masID].Agents.Counter,
		stor.mas[masID].Agents.Counter)
	stor.verMAS[masID].agents = make([]int, stor.mas[masID].Agents.Counter,
		stor.mas[masID].Agents.Counter)

	// get info of all agents and loop through
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	resp := &clientv3.GetResponse{}
	resp, err = stor.client.Get(ctx, "ams/mas/"+strconv.Itoa(masID)+"/agent", clientv3.WithPrefix())
	if err != nil {
		cancel()
		return
	}
	for i := range resp.Kvs {
		temp := strings.Split(string(resp.Kvs[i].Key), "/")
		if len(temp) != 5 {
			continue
		}
		var agentID int
		agentID, err = strconv.Atoi(temp[4])
		if err != nil {
			cancel()
			return
		}
		if agentID >= stor.mas[masID].Agents.Counter {
			continue
		}
		err = json.Unmarshal(resp.Kvs[i].Value, &stor.mas[masID].Agents.Inst[agentID])
		if err != nil {
			cancel()
			return
		}
		stor.verMAS[masID].agents[agentID] = int(resp.Kvs[i].Version)
		imID := stor.mas[masID].Agents.Inst[agentID].ImageGroupID
		agencyID := stor.mas[masID].Agents.Inst[agentID].AgencyID
		stor.adjustAgencyStorage(masID, imID, agencyID)
		// stor.mas[masID].ImageGroups.Inst[imID].Agencies.Inst[agencyID].Agents =
		// 	append(stor.mas[masID].ImageGroups.Inst[imID].Agencies.Inst[agencyID].Agents, agentID)
	}
	cancel()
	return
}

// handleAMSEvents is the handler function for events of the ams/mas/... path
func (stor *etcdStorage) handleAMSEvents() {
	var err error
	watchChan := stor.client.Watch(context.Background(), "ams/mas", clientv3.WithPrefix())
	for {
		watchResp := <-watchChan
		for _, event := range watchResp.Events {
			key := string(event.Kv.Key)
			path := strings.Split(key, "/")
			if len(path) == 3 {
				if key == "ams/mas/counter" {
					err = stor.handleMASCounterEvents(event.Kv)
				}
			} else {
				var masID int
				masID, err = strconv.Atoi(path[2])
				if err != nil {
					stor.logError.Println(err)
					continue
				}

				if len(path) == 4 {
					err = stor.handleMASEvents(event.Kv, masID, path[3])
				} else if len(path) == 5 && path[3] == "agent" {
					var agentID int
					agentID, err = strconv.Atoi(path[4])
					if err != nil {
						stor.logError.Println(err)
						continue
					}
					err = stor.handleAgentEvents(event.Kv, masID, agentID)
				} else if len(path) >= 6 && path[3] == "im" {
					var imID int
					imID, err = strconv.Atoi(path[4])
					if err != nil {
						stor.logError.Println(err)
						continue
					}
					if len(path) == 6 {
						err = stor.handleImGroupEvents(event.Kv, masID, imID, path[5])
					} else if len(path) == 7 {
						var agencyID int
						agencyID, err = strconv.Atoi(path[6])
						if err != nil {
							stor.logError.Println(err)
							continue
						}
						err = stor.handleAgencyEvents(event.Kv, masID, imID, agencyID)
					}
				}
			}
			if err != nil {
				stor.logError.Println(err)
				err = nil
				continue
			}
		}
	}
}

// handleMASCounterEvents is the handler function for events of the ams/mas/counter path
func (stor *etcdStorage) handleMASCounterEvents(kv *mvccpb.KeyValue) (err error) {
	if stor.verMASCounter < int(kv.Version) {
		stor.mutex.Lock()
		err = json.Unmarshal(kv.Value, &stor.masCounter)
		stor.mutex.Unlock()
		if err != nil {
			return
		}
		stor.verMASCounter = int(kv.Version)
	}
	return
}

// handleMASEvents is the handler function for events of the ams/mas/<id>/... path
func (stor *etcdStorage) handleMASEvents(kv *mvccpb.KeyValue, masID int, key string) (err error) {
	stor.adjustMASStorage(masID)

	switch key {
	case "config":
		stor.mutex.Lock()
		if stor.verMAS[masID].config < int(kv.Version) {
			err = json.Unmarshal(kv.Value, &stor.mas[masID].Config)
			if err != nil {
				stor.mutex.Unlock()
				return
			}
			stor.verMAS[masID].config = int(kv.Version)
		}
		stor.mutex.Unlock()
	case "status":
		stor.mutex.Lock()
		if stor.verMAS[masID].status < int(kv.Version) {
			err = json.Unmarshal(kv.Value, &stor.mas[masID].Status)
			if err != nil {
				stor.mutex.Unlock()
				return
			}
			stor.verMAS[masID].status = int(kv.Version)
		}
		stor.mutex.Unlock()
	case "imcounter":
		stor.mutex.Lock()
		if stor.verMAS[masID].groupCounter < int(kv.Version) {
			err = json.Unmarshal(kv.Value, &stor.mas[masID].ImageGroups.Counter)
			if err != nil {
				stor.mutex.Unlock()
				return
			}
			stor.verMAS[masID].groupCounter = int(kv.Version)
		}
		stor.mutex.Unlock()
	case "agentcounter":
		stor.mutex.Lock()
		if stor.verMAS[masID].agentCounter < int(kv.Version) {
			err = json.Unmarshal(kv.Value, &stor.mas[masID].Agents.Counter)
			if err != nil {
				stor.mutex.Unlock()
				return
			}
			stor.verMAS[masID].agentCounter = int(kv.Version)
		}
		stor.mutex.Unlock()
	}
	return
}

// handleAgentEvents is the handler function for events of the ams/mas/<id>/agent/... path
func (stor *etcdStorage) handleAgentEvents(kv *mvccpb.KeyValue, masID int,
	agentID int) (err error) {
	// create agent storage object, if number of stored agents is lower than agentID
	stor.adjustAgentStorage(masID, agentID)

	stor.mutex.Lock()
	if stor.verMAS[masID].agents[agentID] < int(kv.Version) {
		err = json.Unmarshal(kv.Value, &stor.mas[masID].Agents.Inst[agentID])
		if err != nil {
			stor.mutex.Unlock()
			return
		}
		stor.verMAS[masID].agents[agentID] = int(kv.Version)
		if stor.verMAS[masID].agents[agentID] == 1 {
			// add agent to agency's list of agents
			imID := stor.mas[masID].Agents.Inst[agentID].ImageGroupID
			agencyID := stor.mas[masID].Agents.Inst[agentID].AgencyID
			stor.mutex.Unlock()
			stor.adjustAgencyStorage(masID, imID, agencyID)
			stor.mutex.Lock()
			// stor.mas[masID].ImageGroups.Inst[imID].Agencies.Inst[agencyID].Agents =
			// 	append(stor.mas[masID].ImageGroups.Inst[imID].Agencies.Inst[agencyID].Agents, agentID)
		}
	}
	stor.mutex.Unlock()
	return
}

// handleImGroupEvents is the handler function for events of the ams/mas/<id>/im/... path
func (stor *etcdStorage) handleImGroupEvents(kv *mvccpb.KeyValue, masID int,
	imID int, key string) (err error) {
	stor.adjustImGroupStorage(masID, imID)

	switch key {
	case "config":
		stor.mutex.Lock()
		if stor.verMAS[masID].imGroups[imID].config < int(kv.Version) {
			err = json.Unmarshal(kv.Value, &stor.mas[masID].ImageGroups.Inst[imID].Config)
			if err != nil {
				stor.mutex.Unlock()
				return
			}
			stor.verMAS[masID].imGroups[imID].config = int(kv.Version)
		}
		stor.mutex.Unlock()
	case "agencycounter":
		stor.mutex.Lock()
		if stor.verMAS[masID].imGroups[imID].agencyCounter < int(kv.Version) {
			err = json.Unmarshal(kv.Value, &stor.mas[masID].ImageGroups.Inst[imID].Agencies.Counter)
			if err != nil {
				stor.mutex.Unlock()
				return
			}
			stor.verMAS[masID].imGroups[imID].agencyCounter = int(kv.Version)
		}
		stor.mutex.Unlock()
	}
	return
}

// handleAgencyEvents is the handler function for events of the ams/mas/<id>/agency/... path
func (stor *etcdStorage) handleAgencyEvents(kv *mvccpb.KeyValue, masID int, imID int,
	agencyID int) (err error) {
	// create agency storage object, if number of stored agencies is lower than agencyID
	stor.adjustAgencyStorage(masID, imID, agencyID)

	stor.mutex.Lock()
	if stor.verMAS[masID].imGroups[imID].agencies[agencyID] < int(kv.Version) {
		err = json.Unmarshal(kv.Value,
			&stor.mas[masID].ImageGroups.Inst[imID].Agencies.Inst[agencyID])
		if err != nil {
			stor.mutex.Unlock()
			return
		}
		stor.verMAS[masID].imGroups[imID].agencies[agencyID] = int(kv.Version)
		// if stor.verMAS[masID].imGroups[imID].agencies[agencyID] == 1 {
		// 	// agency is new -> fill agency info fields
		// 	stor.mas[masID].ImageGroups.Inst[imID].Agencies.Inst[agencyID].MASID = masID
		// 	stor.mas[masID].ImageGroups.Inst[imID].Agencies.Inst[agencyID].ID = agencyID
		// 	stor.mas[masID].ImageGroups.Inst[imID].Agencies.Inst[agencyID].ImageGroupID =
		// 		imID
		// 	stor.mas[masID].ImageGroups.Inst[imID].Agencies.Inst[agencyID].Logger =
		// 		stor.mas[masID].Config.Logger
		// 	stor.mas[masID].ImageGroups.Inst[imID].Agencies.Inst[agencyID].Name =
		// 		"mas-" + strconv.Itoa(masID) + "-im-" + strconv.Itoa(imID) + "-agency-" +
		// 			strconv.Itoa(agencyID) + ".mas" + strconv.Itoa(masID) + "agencies"
		// }
	}
	stor.mutex.Unlock()
	return
}

// handleGraphEvents handles events on all mas graphs
func (stor *etcdStorage) handleGraphEvents() {
	var err error
	watchChan := stor.client.Watch(context.Background(), "df/graph", clientv3.WithPrefix())
	for {
		watchResp := <-watchChan
		for _, event := range watchResp.Events {
			key := string(event.Kv.Key)
			path := strings.Split(key, "/")
			if len(path) == 3 && path[0] == "df" && path[1] == "graph" {
				var masID int
				masID, err = strconv.Atoi(path[2])
				if err != nil {
					stor.logError.Println(err)
					continue
				}
				// create masStorage object, if number of stored mas is lower than masID
				stor.mutex.Lock()
				if len(stor.mas) <= masID {
					for i := 0; i < masID-len(stor.mas)+1; i++ {
						stor.mas = append(stor.mas, schemas.MASInfo{})
					}
				}
				stor.mutex.Unlock()
				if len(stor.verMAS) <= masID {
					for i := 0; i < masID-len(stor.verMAS)+1; i++ {
						stor.verMAS = append(stor.verMAS, masVersion{})
					}
				}
				if stor.verMAS[masID].graph < int(event.Kv.Version) {
					err = json.Unmarshal(event.Kv.Value, &stor.mas[masID].Graph)
					if err != nil {
						return
					}
					stor.verMAS[masID].graph = int(event.Kv.Version)
				}
			}
		}
	}
}

// adjustMASStorage adjusts the size of the MAS storage to the masID
func (stor *etcdStorage) adjustMASStorage(masID int) {
	// create masStorage object, if number of stored mas is lower than masID
	stor.mutex.Lock()
	if len(stor.mas) <= masID {
		for i := 0; i < masID-len(stor.mas)+1; i++ {
			stor.mas = append(stor.mas, schemas.MASInfo{})
		}
	}
	if len(stor.verMAS) <= masID {
		for i := 0; i < masID-len(stor.verMAS)+1; i++ {
			stor.verMAS = append(stor.verMAS, masVersion{})
		}
	}
	stor.mutex.Unlock()
	return
}

// adjustImGroupStorage adjusts the size of the image group storage to the imID
func (stor *etcdStorage) adjustImGroupStorage(masID int, imID int) {
	stor.adjustMASStorage(masID)
	// create imGroup storage object, if number of stored im groups is lower than imID
	stor.mutex.Lock()
	numImGroups := len(stor.mas[masID].ImageGroups.Inst)
	if numImGroups <= imID {
		for i := 0; i < imID-numImGroups+1; i++ {
			stor.mas[masID].ImageGroups.Inst = append(stor.mas[masID].ImageGroups.Inst,
				schemas.ImageGroupInfo{ID: numImGroups + i})
		}
	}
	if len(stor.verMAS[masID].imGroups) <= imID {
		for i := 0; i < imID-len(stor.verMAS[masID].imGroups)+1; i++ {
			stor.verMAS[masID].imGroups = append(stor.verMAS[masID].imGroups, imGroupVersion{})
		}
	}
	stor.mutex.Unlock()
	return
}

// adjustAgencyStorage adjusts the size of the agency storage to the agencyID
func (stor *etcdStorage) adjustAgencyStorage(masID int, imID int, agencyID int) {
	stor.adjustImGroupStorage(masID, imID)
	// create agency storage object, if number of stored agencies is lower than agencyID
	stor.mutex.Lock()
	if len(stor.mas[masID].ImageGroups.Inst[imID].Agencies.Inst) <= agencyID {
		for i := 0; i < agencyID-len(stor.mas[masID].ImageGroups.Inst[imID].Agencies.Inst)+1; i++ {
			stor.mas[masID].ImageGroups.Inst[imID].Agencies.Inst = append(
				stor.mas[masID].ImageGroups.Inst[imID].Agencies.Inst,
				schemas.AgencyInfo{})
		}
	}
	if len(stor.verMAS[masID].imGroups[imID].agencies) <= agencyID {
		for i := 0; i < agencyID-len(stor.verMAS[masID].imGroups[imID].agencies)+1; i++ {
			stor.verMAS[masID].imGroups[imID].agencies = append(
				stor.verMAS[masID].imGroups[imID].agencies, 0)
		}
	}
	stor.mutex.Unlock()
	return
}

// adjustAgentStorage adjusts the size of the agent storage to the agentID
func (stor *etcdStorage) adjustAgentStorage(masID int, agentID int) {
	stor.adjustMASStorage(masID)
	// create agent storage object, if number of stored agents is lower than agentID
	stor.mutex.Lock()
	if len(stor.mas[masID].Agents.Inst) <= agentID {
		for i := 0; i < agentID-len(stor.mas[masID].Agents.Inst)+1; i++ {
			stor.mas[masID].Agents.Inst = append(stor.mas[masID].Agents.Inst,
				schemas.AgentInfo{})
		}
	}
	if len(stor.verMAS[masID].agents) <= agentID {
		for i := 0; i < agentID-len(stor.verMAS[masID].agents)+1; i++ {
			stor.verMAS[masID].agents = append(stor.verMAS[masID].agents, 0)
		}
	}
	stor.mutex.Unlock()
	return
}

// etcdGetResource requests resourcefrom etcd and unmarshalls it
func (stor *etcdStorage) etcdGetResource(key string, v interface{}) (ver int, err error) {
	ver = 0
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	resp := &clientv3.GetResponse{}
	resp, err = stor.client.Get(ctx, key)
	if err == nil {
		if len(resp.Kvs) > 0 {
			err = json.Unmarshal(resp.Kvs[0].Value, v)
			ver = int(resp.Kvs[0].Version)
		} else {
			err = errors.New("NotFoundError")
		}
	}
	cancel()
	return
}

// etcdPutResource marshalls resource and puts it to etcd
func (stor *etcdStorage) etcdPutResource(key string, v interface{}) (err error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	var res []byte
	res, err = json.Marshal(v)
	if err == nil {
		_, err = stor.client.Put(ctx, key, string(res))
	}
	cancel()
	return
}

// etcdDeleteResourceRecursively deletes all keys and values recursively
func (stor *etcdStorage) etcdDeleteResourceRecursively(key string) (err error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	_, err = stor.client.Delete(ctx, key, clientv3.WithPrefix())
	cancel()
	return
}
