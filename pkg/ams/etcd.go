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
// ams/mas/<masID>/spec schemas.MASSpec
// ams/mas/<masID>/status schemas.Status
// ams/mas/<masID>/agentcounter int (agentCounter)
// // ams/mas/<masID>/agents/data schemas.Agents
// ams/mas/<masID>/agents/<agentID>: schemas.AgentInfo
// ams/mas/<masID>/agencycounter: int (agencyCounter)
// ams/mas/<masID>/agencies/<agencyID>: AgencyInfo
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
	status        int
	spec          int
	agentCounter  int
	agents        []int
	agencyCounter int
	agencies      []int
	graph         int
}

// setCloneMAPInfo sets info specific to running clonemap instance
func (stor *etcdStorage) setCloneMAPInfo(cloneMAP schemas.CloneMAP) (err error) {
	err = stor.etcdPutResource("ams/data", cloneMAP)
	return
}

// setAgentAddress sets address of agent
func (stor *etcdStorage) setAgentAddress(masID int, agentID int,
	address schemas.Address) (err error) {
	err = stor.etcdPutResource("ams/mas/"+strconv.Itoa(masID)+"/agents/"+strconv.Itoa(agentID)+
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
	var tempSpec schemas.MASSpec
	_, err = stor.etcdGetResource("ams/mas/"+strconv.Itoa(masID)+"/spec", &tempSpec)
	if err == nil {
		// resource already exists
		err = errors.New("MAS already exists")
		return
	}

	newMAS := createMASStorage(masID, masInfo)
	err = stor.etcdPutResource("ams/mas/"+strconv.Itoa(masID)+"/spec", newMAS.Spec)
	if err != nil {
		return
	}
	err = stor.etcdPutResource("ams/mas/"+strconv.Itoa(masID)+"/status", newMAS.Status)
	if err != nil {
		return
	}
	err = stor.etcdPutResource("ams/mas/"+strconv.Itoa(masID)+"/agentcounter",
		newMAS.Agents.Counter)
	if err != nil {
		return
	}
	err = stor.etcdPutResource("ams/mas/"+strconv.Itoa(masID)+"/agencycounter",
		newMAS.Agencies.Counter)
	if err != nil {
		return
	}
	err = stor.etcdPutResource("df/graph/"+strconv.Itoa(masID),
		newMAS.Spec.Graph)
	if err != nil {
		return
	}

	err = stor.uploadAgentInfo(newMAS)
	if err != nil {
		return
	}

	err = stor.uploadAgencyInfo(newMAS)
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
			res, err = json.Marshal(newMAS.Agents.Instances[agentIndex])
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

// uploadAgencyInfo puts all AgencyInfo of a newly created MAS to etcd
func (stor *etcdStorage) uploadAgencyInfo(newMAS schemas.MASInfo) (err error) {
	agencyIndex := 0
	for {
		numAgInTrans := 100
		if newMAS.Agencies.Counter-agencyIndex < numAgInTrans {
			numAgInTrans = newMAS.Agencies.Counter - agencyIndex
		}
		Ops := make([]clientv3.Op, numAgInTrans, numAgInTrans)
		// put all agencies structs together
		for i := 0; i < numAgInTrans; i++ {
			var res []byte
			res, err = json.Marshal(newMAS.Agencies.Instances[agencyIndex])
			if err != nil {
				return
			}
			Ops[i] = clientv3.OpPut("ams/mas/"+strconv.Itoa(newMAS.ID)+"/agency/"+
				strconv.Itoa(agencyIndex), string(res))
			agencyIndex++
		}
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		cond := clientv3.Compare(clientv3.Version("ams/mas/"+strconv.Itoa(newMAS.ID)+
			"/agencycounter"), ">", 0)
		_, err = stor.client.Txn(ctx).If(cond).Then(Ops...).Commit()
		cancel()
		if err != nil {
			return
		}
		if agencyIndex >= newMAS.Agencies.Counter {
			break
		}
		time.Sleep(50 * time.Millisecond)
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

// addAgent adds an agent to an existing MAS
func (stor *etcdStorage) addAgent(masID int, agentSpec schemas.AgentSpec) (err error) {
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
		// MAS spec
		stor.verMAS[i].spec, err = stor.etcdGetResource("ams/mas/"+strconv.Itoa(i)+"/spec",
			&stor.mas[i].Spec)
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
			&stor.mas[i].Spec.Graph)
		if err != nil {
			return
		}
		// MAS agents
		err = stor.initMASAgents(i)
		if err != nil {
			return
		}
		// MAS agencies
		err = stor.initMASAgencies(i)
		if err != nil {
			return
		}
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
	stor.mas[masID].Agents.Instances = make([]schemas.AgentInfo, stor.mas[masID].Agents.Counter,
		stor.mas[masID].Agents.Counter)
	stor.verMAS[masID].agents = make([]int, stor.mas[masID].Agents.Counter,
		stor.mas[masID].Agents.Counter)

	// get info of all agents and loop through
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	resp := &clientv3.GetResponse{}
	resp, err = stor.client.Get(ctx, "ams/mas/"+strconv.Itoa(masID)+"/agent")
	if err == nil {
		for i := range resp.Kvs {
			temp := strings.Split(string(resp.Kvs[i].Key), "/")
			if len(temp) != 5 {
				continue
			}
			var agentID int
			agentID, err = strconv.Atoi(temp[4])
			if err != nil {
				continue
			}
			if agentID >= stor.mas[masID].Agents.Counter {
				continue
			}
			err = json.Unmarshal(resp.Kvs[i].Value, &stor.mas[masID].Agents.Instances[agentID])
			if err != nil {
				continue
			}
			stor.verMAS[masID].agents[agentID] = int(resp.Kvs[i].Version)
		}
	}
	cancel()
	return
}

// initMASAgencies retrieves data of agencies in a mas from etcd
func (stor *etcdStorage) initMASAgencies(masID int) (err error) {
	// MAS agency counter
	stor.verMAS[masID].agencyCounter, err = stor.etcdGetResource("ams/mas/"+strconv.Itoa(masID)+
		"/agencycounter", &stor.mas[masID].Agencies.Counter)
	if err != nil {
		return
	}
	stor.mas[masID].Agencies.Instances = make([]schemas.AgencyInfo, stor.mas[masID].Agencies.Counter,
		stor.mas[masID].Agencies.Counter)
	stor.verMAS[masID].agencies = make([]int, stor.mas[masID].Agencies.Counter,
		stor.mas[masID].Agencies.Counter)

	// get info of all agencies and loop through
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	resp := &clientv3.GetResponse{}
	resp, err = stor.client.Get(ctx, "ams/mas/"+strconv.Itoa(masID)+"/agency")
	if err == nil {
		for i := range resp.Kvs {
			temp := strings.Split(string(resp.Kvs[i].Key), "/")
			if len(temp) != 5 {
				continue
			}
			var agencyID int
			agencyID, err = strconv.Atoi(temp[4])
			if err != nil {
				continue
			}
			if agencyID >= stor.mas[masID].Agencies.Counter {
				continue
			}
			err = json.Unmarshal(resp.Kvs[i].Value, &stor.mas[masID].Agencies.Instances[agencyID])
			if err != nil {
				continue
			}
			stor.verMAS[masID].agencies[agencyID] = int(resp.Kvs[i].Version)
		}
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
				if len(path) == 4 {
					err = stor.handleMASEvents(event.Kv, masID, path[3])
				} else if len(path) == 5 {
					if strings.HasPrefix(key, "ams/mas/") {
						if path[3] == "agent" {
							var agentID int
							agentID, err = strconv.Atoi(path[4])
							if err != nil {
								stor.logError.Println(err)
								continue
							}
							err = stor.handleAgentEvents(event.Kv, masID, agentID)
						} else if path[3] == "agency" {
							var agencyID int
							agencyID, err = strconv.Atoi(path[4])
							if err != nil {
								stor.logError.Println(err)
								continue
							}
							err = stor.handleAgencyEvents(event.Kv, masID, agencyID)
						}
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
		err = json.Unmarshal(kv.Value, &stor.masCounter)
		if err != nil {
			return
		}
		stor.verMASCounter = int(kv.Version)
	}
	return
}

// handleMASEvents is the handler function for events of the ams/mas/<id>/... path
func (stor *etcdStorage) handleMASEvents(kv *mvccpb.KeyValue, masID int, key string) (err error) {
	switch key {
	case "spec":
		if stor.verMAS[masID].spec < int(kv.Version) {
			err = json.Unmarshal(kv.Value, &stor.mas[masID].Spec)
			if err != nil {
				return
			}
			stor.verMAS[masID].spec = int(kv.Version)
		}
	case "status":
		if stor.verMAS[masID].status < int(kv.Version) {
			err = json.Unmarshal(kv.Value, &stor.mas[masID].Status)
			if err != nil {
				return
			}
			stor.verMAS[masID].status = int(kv.Version)
		}
	case "agentcounter":
		if stor.verMAS[masID].agentCounter < int(kv.Version) {
			err = json.Unmarshal(kv.Value, &stor.mas[masID].Agents.Counter)
			if err != nil {
				return
			}
			stor.verMAS[masID].agentCounter = int(kv.Version)
		}
	case "agencycounter":
		if stor.verMAS[masID].agencyCounter < int(kv.Version) {
			err = json.Unmarshal(kv.Value, &stor.mas[masID].Agencies.Counter)
			if err != nil {
				return
			}
			stor.verMAS[masID].agencyCounter = int(kv.Version)
		}
	}
	return
}

// handleAgentEvents is the handler function for events of the ams/mas/<id>/agent/... path
func (stor *etcdStorage) handleAgentEvents(kv *mvccpb.KeyValue, masID int,
	agentID int) (err error) {
	// create agent storage object, if number of stored agents is lower than agentID
	stor.mutex.Lock()
	if len(stor.mas[masID].Agents.Instances) <= agentID {
		for i := 0; i < agentID-len(stor.mas[masID].Agents.Instances)+1; i++ {
			stor.mas[masID].Agents.Instances = append(stor.mas[masID].Agents.Instances,
				schemas.AgentInfo{})
		}
	}
	stor.mutex.Unlock()
	if len(stor.verMAS[masID].agents) <= agentID {
		for i := 0; i < agentID-len(stor.verMAS[masID].agents)+1; i++ {
			stor.verMAS[masID].agents = append(stor.verMAS[masID].agents, 0)
		}
	}

	if stor.verMAS[masID].agents[agentID] < int(kv.Version) {
		err = json.Unmarshal(kv.Value, &stor.mas[masID].Agents.Instances[agentID])
		if err != nil {
			return
		}
		stor.verMAS[masID].agents[agentID] = int(kv.Version)
	}
	return
}

// handleAgencyEvents is the handler function for events of the ams/mas/<id>/agency/... path
func (stor *etcdStorage) handleAgencyEvents(kv *mvccpb.KeyValue, masID int,
	agencyID int) (err error) {
	// create agency storage object, if number of stored agencies is lower than agencyID
	stor.mutex.Lock()
	if len(stor.mas[masID].Agencies.Instances) <= agencyID {
		for i := 0; i < agencyID-len(stor.mas[masID].Agencies.Instances)+1; i++ {
			stor.mas[masID].Agencies.Instances = append(stor.mas[masID].Agencies.Instances,
				schemas.AgencyInfo{})
		}
	}
	stor.mutex.Unlock()
	if len(stor.verMAS[masID].agencies) <= agencyID {
		for i := 0; i < agencyID-len(stor.verMAS[masID].agencies)+1; i++ {
			stor.verMAS[masID].agencies = append(stor.verMAS[masID].agencies, 0)
		}
	}

	if stor.verMAS[masID].agencies[agencyID] < int(kv.Version) {
		err = json.Unmarshal(kv.Value, &stor.mas[masID].Agencies.Instances[agencyID])
		if err != nil {
			return
		}
		stor.verMAS[masID].agencies[agencyID] = int(kv.Version)
	}
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
					err = json.Unmarshal(event.Kv.Value, &stor.mas[masID].Spec.Graph)
					if err != nil {
						return
					}
					stor.verMAS[masID].graph = int(event.Kv.Version)
				}
			}
		}
	}
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
