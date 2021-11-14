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

// etcd paths:
//
// df/mascounter
// df/mas/<masID>/svc/<svcID>
// df/graph/<masID>: schemas.Graph

package df

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/RWTH-ACS/clonemap/pkg/schemas"
	"go.etcd.io/etcd/clientv3"
	"go.etcd.io/etcd/mvcc/mvccpb"
)

// etcd storage
type etcdStorage struct {
	config       clientv3.Config  // configuration of client
	client       *clientv3.Client // client
	localStorage                  // local cache
	verMAS       []masVersion     // required to check if values received from etcd watcher are newer
	logError     *log.Logger      // logger for error logging
}

// version of mas keys in etcd
type masVersion struct {
	service map[string]int
	graph   int
}

// registerService stores a service and registers it with the specified types
func (stor *etcdStorage) registerService(svc schemas.Service) (svcID string, err error) {
	// check if service already exists
	stor.mutex.Lock()
	numMAS := len(stor.mas)
	stor.mutex.Unlock()
	if numMAS > svc.MASID {
		stor.mutex.Lock()
		agSvc, ok := stor.mas[svc.MASID].agentService[svc.AgentID]
		stor.mutex.Unlock()
		if ok {
			for i := range agSvc {
				if agSvc[i].Desc == svc.Desc {
					err = errors.New("service already exists")
					return
				}
			}
		}
	}

	svcID = nextGUID()
	svc.GUID = svcID

	err = stor.etcdPutResource("df/mas/"+strconv.Itoa(svc.MASID)+"/svc/"+svcID, svc)

	return
}

// deregisterService deletes a service
func (stor *etcdStorage) deregisterService(masID int, svcID string) (err error) {
	var svc schemas.Service
	var ok bool
	// check if mas and service exist
	stor.mutex.Lock()
	numMAS := len(stor.mas)
	stor.mutex.Unlock()
	if numMAS <= masID {
		err = errors.New("service does not exists")
		return
	}
	stor.mutex.Lock()
	svc, ok = stor.mas[masID].service[svcID]
	stor.mutex.Unlock()
	if !ok {
		err = errors.New("service does not exists")
		return
	}
	// check if service has not been deleted yet
	if svc.GUID == "-1" {
		err = errors.New("service does not exists")
		return
	}
	svc.GUID = "-1"
	err = stor.etcdPutResource("df/mas/"+strconv.Itoa(svc.MASID)+"/svc/"+svcID, svc)
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
	go temp.handleDFEvents()
	stor = &temp
	return
}

// initEtcd sets the clodmap version and uptime if not already present
func (stor *etcdStorage) initEtcd() (err error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	req := clientv3.OpPut("df/mascounter", strconv.Itoa(0))
	cond := clientv3.Compare(clientv3.Version("df/mascounter"), "=", 0)
	_, err = stor.client.Txn(ctx).If(cond).Then(req).Commit()
	cancel()
	return
}

// initCache initializes the cached local storage
func (stor *etcdStorage) initCache() (err error) {
	stor.mutex = &sync.Mutex{}
	// get info of all svcs and loop through
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	var resp *clientv3.GetResponse
	resp, err = stor.client.Get(ctx, "df/mas")
	if err == nil {
		for i := range resp.Kvs {
			temp := strings.Split(string(resp.Kvs[i].Key), "/")
			if len(temp) != 5 {
				continue
			}
			var masID int
			masID, err = strconv.Atoi(temp[2])
			if err != nil {
				stor.logError.Println(err)
				continue
			}
			// create masStorage object, if number of stored mas is lower than masID
			stor.mutex.Lock()
			numMas := len(stor.mas)
			stor.mutex.Unlock()
			if numMas <= masID {
				stor.mutex.Lock()
				for i := 0; i < masID-numMas+1; i++ {
					stor.mas = append(stor.mas, createMASStorage())
					stor.verMAS = append(stor.verMAS, masVersion{service: make(map[string]int)})
				}
				stor.mutex.Unlock()
			}

			svcID := temp[4]
			var svc schemas.Service
			err = json.Unmarshal(resp.Kvs[i].Value, &svc)
			if err != nil {
				continue
			}
			stor.mutex.Lock()
			agSvc := stor.mas[svc.MASID].agentService[svc.AgentID]
			descSvc := stor.mas[svc.MASID].descService[svc.Desc]
			stor.mutex.Unlock()
			// add service to agent and desc maps
			agSvc = append(agSvc, svc)
			descSvc = append(descSvc, svc)
			stor.mutex.Lock()
			stor.mas[masID].service[svcID] = svc
			stor.verMAS[masID].service[svcID] = int(resp.Kvs[i].Version)
			stor.mas[svc.MASID].agentService[svc.AgentID] = agSvc
			stor.mas[svc.MASID].descService[svc.Desc] = descSvc
			stor.mutex.Unlock()
		}
	}
	cancel()

	// get graphs
	ctx, cancel = context.WithTimeout(context.Background(), 10*time.Second)
	resp, err = stor.client.Get(ctx, "df/graph")
	if err != nil {
		for i := range resp.Kvs {
			temp := strings.Split(string(resp.Kvs[i].Key), "/")
			if len(temp) != 3 {
				continue
			}
			var masID int
			masID, err = strconv.Atoi(temp[2])
			if err != nil {
				continue
			}
			var gr schemas.Graph
			err = json.Unmarshal(resp.Kvs[i].Value, &gr)
			if err != nil {
				continue
			}
			stor.updateGraph(masID, gr)
		}
	}
	cancel()
	return
}

// handleDFEvents is the handler function for events of the df/... path
func (stor *etcdStorage) handleDFEvents() {
	var err error
	watchChan := stor.client.Watch(context.Background(), "df", clientv3.WithPrefix())
	for {
		watchResp := <-watchChan
		for _, event := range watchResp.Events {
			key := string(event.Kv.Key)
			path := strings.Split(key, "/")
			if len(path) >= 3 {
				var masID int
				masID, err = strconv.Atoi(path[2])
				if err != nil {
					stor.logError.Println(err)
					continue
				}
				// create masStorage object, if number of stored mas is lower than masID
				stor.mutex.Lock()
				numMas := len(stor.mas)
				stor.mutex.Unlock()
				if numMas <= masID {
					stor.mutex.Lock()
					for i := 0; i < masID-numMas+1; i++ {
						stor.mas = append(stor.mas, createMASStorage())
						stor.verMAS = append(stor.verMAS, masVersion{service: make(map[string]int)})
					}
					stor.mutex.Unlock()
				}
				if len(path) == 3 {
					if path[1] == "graph" {
						err = stor.handleGraphEvents(masID, event.Kv)
						if err != nil {
							stor.logError.Println(err)
							continue
						}
					}
				} else if len(path) == 5 {
					if path[1] == "mas" && path[3] == "svc" {
						svcID := path[4]
						err = stor.handleSvcEvents(masID, svcID, event.Kv)
						if err != nil {
							stor.logError.Println(err)
							continue
						}
					}
				}
			}
		}
	}
}

// handleGraphEvents is the handler function for events of the df/graph/{masID}
func (stor *etcdStorage) handleGraphEvents(masID int, kv *mvccpb.KeyValue) (err error) {
	stor.mutex.Lock()
	ver := stor.verMAS[masID].graph
	stor.mutex.Unlock()
	if ver < int(kv.Version) {
		var g schemas.Graph
		err = json.Unmarshal(kv.Value, &g)
		if err != nil {
			return
		}
		err = stor.updateGraph(masID, g)
		if err != nil {
			return
		}
		stor.mutex.Lock()
		stor.verMAS[masID].graph = int(kv.Version)
		stor.mutex.Unlock()
	}
	return
}

// handleSvcEvents is the handler function for events of the df/mas/{masID}/svc/{svcID}
func (stor *etcdStorage) handleSvcEvents(masID int, svcID string, kv *mvccpb.KeyValue) (err error) {
	stor.mutex.Lock()
	ver := stor.verMAS[masID].service[svcID]
	stor.mutex.Unlock()
	if ver < int(kv.Version) {
		var svc schemas.Service
		err = json.Unmarshal(kv.Value, &svc)
		if err != nil {
			return
		}
		stor.mutex.Lock()
		elem, ok := stor.mas[masID].service[svcID]
		stor.mutex.Unlock()
		if !ok {
			// svc is new and not initialized -> add to agent and desc map
			stor.mutex.Lock()
			agSvc, _ := stor.mas[masID].agentService[svc.AgentID]
			descSvc := stor.mas[masID].descService[svc.Desc]
			stor.mutex.Unlock()
			agSvc = append(agSvc, svc)
			descSvc = append(descSvc, svc)
			stor.mutex.Lock()
			stor.mas[masID].agentService[svc.AgentID] = agSvc
			stor.mas[masID].descService[svc.Desc] = descSvc
			stor.mas[masID].service[svcID] = svc
			stor.mutex.Unlock()
		} else {
			// svc is deregistered -> delete from agent and desc map
			elem.GUID = "-1"
			stor.mutex.Lock()
			stor.mas[masID].service[svcID] = elem
			agSvc := stor.mas[masID].agentService[svc.AgentID]
			descSvc := stor.mas[masID].descService[svc.Desc]
			stor.mutex.Unlock()
			svcIndex := -1
			// search for service in agent map
			for i := range agSvc {
				if agSvc[i].Desc == svc.Desc {
					svcIndex = i
					break
				}
			}
			if svcIndex == -1 {
				err = errors.New("service does not exists")
				return
			}
			// delete service from slice of corresponding agent
			if len(agSvc) == 1 {
				stor.mutex.Lock()
				delete(stor.mas[masID].agentService, svc.AgentID)
				stor.mutex.Unlock()
			} else {
				copy(agSvc[svcIndex:], agSvc[svcIndex+1:])
				agSvc[len(agSvc)-1] = schemas.Service{}
				agSvc = agSvc[:len(agSvc)-1]
				stor.mutex.Lock()
				stor.mas[masID].agentService[svc.AgentID] = agSvc
				stor.mutex.Unlock()
			}

			// search service in desc map
			svcIndex = -1
			for i := range descSvc {
				if descSvc[i].AgentID == svc.AgentID {
					svcIndex = i
					break
				}
			}
			if svcIndex == -1 {
				err = errors.New("service does not exists")
				return
			}
			// delete service from slice of corresponding desc entry
			if len(descSvc) == 1 {
				stor.mutex.Lock()
				delete(stor.mas[masID].descService, svc.Desc)
				stor.mutex.Unlock()
			} else {
				copy(descSvc[svcIndex:], descSvc[svcIndex+1:])
				descSvc[len(descSvc)-1] = schemas.Service{}
				descSvc = descSvc[:len(descSvc)-1]
				stor.mutex.Lock()
				stor.mas[masID].descService[svc.Desc] = descSvc
				stor.mutex.Unlock()
			}
		}
		stor.mutex.Lock()
		stor.verMAS[masID].service[svcID] = int(kv.Version)
		stor.mutex.Unlock()
	}
	return
}

// etcdGetResource requests resourcefrom etcd and unmarshalls it
func (stor *etcdStorage) etcdGetResource(key string, v interface{}) (ver int, err error) {
	ver = 0
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	var resp *clientv3.GetResponse
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
