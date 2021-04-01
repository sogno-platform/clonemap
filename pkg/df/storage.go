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

package df

// implements storage interface for different types (local, etcd)

import (
	"errors"
	"sync"

	"github.com/rs/xid"

	"git.rwth-aachen.de/acs/public/cloud/mas/clonemap/pkg/schemas"
)

// storage interface for interaction with storage
type storage interface {
	registerService(svc schemas.Service) (svcID string, err error)
	deregisterService(masID int, svcID string) (err error)
	searchServices(masID int, desc string) (svc []schemas.Service, err error)
	searchLocalServices(masID int, nodeID int, dist float64, desc string) (svc []schemas.Service,
		err error)
	getService(masID int, svcID string) (svc schemas.Service, err error)
	updateGraph(masID int, g schemas.Graph) (err error)
	getGraph(masID int) (g schemas.Graph, err error)
}

// represents local storage
type localStorage struct {
	mas   []masStorage
	mutex *sync.Mutex
}

// mas storage
type masStorage struct {
	service      map[string]schemas.Service // index == ID
	agentService map[int][]schemas.Service
	descService  map[string][]schemas.Service
	graph        graph
}

// registerService stores a service and registers it with the specified types
func (stor *localStorage) registerService(svc schemas.Service) (svcID string, err error) {
	// create mas storages if necessary (masID == index)
	stor.mutex.Lock()
	numMAS := len(stor.mas)
	stor.mutex.Unlock()
	if numMAS <= svc.MASID {
		stor.mutex.Lock()
		for i := 0; i < svc.MASID-numMAS+1; i++ {
			stor.mas = append(stor.mas, createMASStorage())
		}
		stor.mutex.Unlock()
	}
	// check if service already exists
	stor.mutex.Lock()
	agSvc, ok := stor.mas[svc.MASID].agentService[svc.AgentID]
	descSvc := stor.mas[svc.MASID].descService[svc.Desc]
	stor.mutex.Unlock()
	if ok {
		for i := range agSvc {
			if agSvc[i].Desc == svc.Desc {
				err = errors.New("service already exists")
				return
			}
		}
	}
	// determine service id and add service to slices and maps
	svcID = nextGUID()
	svc.GUID = svcID
	agSvc = append(agSvc, svc)
	descSvc = append(descSvc, svc)
	stor.mutex.Lock()
	stor.mas[svc.MASID].agentService[svc.AgentID] = agSvc
	stor.mas[svc.MASID].descService[svc.Desc] = descSvc
	stor.mas[svc.MASID].service[svcID] = svc
	stor.mutex.Unlock()

	return
}

// createMASStorage returns a filled masStorage object
func createMASStorage() (ret masStorage) {
	ret.service = make(map[string]schemas.Service)
	ret.agentService = make(map[int][]schemas.Service)
	ret.descService = make(map[string][]schemas.Service)
	return
}

// nextGUID returns a glabally unique identifier
func nextGUID() (ret string) {
	ret = xid.New().String()
	return
}

// deregisterService deletes a service
func (stor *localStorage) deregisterService(masID int, svcID string) (err error) {
	var svc schemas.Service
	var ok bool
	stor.mutex.Lock()
	numMAS := len(stor.mas)
	stor.mutex.Unlock()
	// check if mas and service exist
	if numMAS <= masID {
		err = errors.New("service does not exists")
		return
	}
	stor.mutex.Lock()
	svc, ok = stor.mas[masID].service[svcID]
	stor.mutex.Unlock()
	// check if service has not been deleted yet
	if !ok || svc.GUID == "-1" {
		err = errors.New("service does not exists")
		return
	}
	// mark service as deleted
	svc.GUID = "-1"
	stor.mutex.Lock()
	stor.mas[masID].service[svcID] = svc
	agSvc := stor.mas[masID].agentService[svc.AgentID]
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
	stor.mutex.Lock()
	descSvc := stor.mas[masID].descService[svc.Desc]
	stor.mutex.Unlock()
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
	return
}

// searchServices searches for all services with specified description
func (stor *localStorage) searchServices(masID int, desc string) (svc []schemas.Service,
	err error) {
	stor.mutex.Lock()
	numMAS := len(stor.mas)
	stor.mutex.Unlock()
	if numMAS <= masID {
		// err = errors.New("MAS does not exists")
		return
	}
	if desc == "" {
		stor.mutex.Lock()
		for i := range stor.mas[masID].agentService {
			svc = append(svc, stor.mas[masID].agentService[i]...)
		}
		stor.mutex.Unlock()
	} else {
		stor.mutex.Lock()
		svc = stor.mas[masID].descService[desc]
		stor.mutex.Unlock()
	}
	return
}

// searchLocalServices searches all services within a certain range from specified node
func (stor *localStorage) searchLocalServices(masID int, nodeID int, dist float64,
	desc string) (svc []schemas.Service, err error) {
	var nodes map[int]float64
	stor.mutex.Lock()
	numMAS := len(stor.mas)
	stor.mutex.Unlock()
	if numMAS <= masID {
		err = errors.New("MAS does not exists")
		return
	}
	stor.mutex.Lock()
	nodes, err = stor.mas[masID].graph.getNodesInRange(nodeID, dist)
	if err != nil {
		return
	}
	svcAll := stor.mas[masID].descService[desc]
	stor.mutex.Unlock()
	for i := range svcAll {
		for j := range nodes {
			if svcAll[i].NodeID == j {
				svcAll[i].Dist = nodes[j]
				svc = append(svc, svcAll[i])
			}
		}
	}
	return
}

// getService returns a service with a certain id
func (stor *localStorage) getService(masID int, svcID string) (svc schemas.Service, err error) {
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
	return
}

// updateGraph updates the stored graph for a mas
func (stor *localStorage) updateGraph(masID int, g schemas.Graph) (err error) {
	var gr graph
	gr, err = graphFromSchema(masID, g)
	if err != nil {
		return
	}
	// create mas storages if necessary (masID == index)
	stor.mutex.Lock()
	numMAS := len(stor.mas)
	stor.mutex.Unlock()
	if numMAS <= masID {
		stor.mutex.Lock()
		for i := 0; i < masID-numMAS+1; i++ {
			stor.mas = append(stor.mas, createMASStorage())
		}
		stor.mutex.Unlock()
	}
	stor.mutex.Lock()
	stor.mas[masID].graph = gr
	stor.mutex.Unlock()
	return
}

func (stor *localStorage) getGraph(masID int) (g schemas.Graph, err error) {
	stor.mutex.Lock()
	numMAS := len(stor.mas)
	stor.mutex.Unlock()
	if numMAS <= masID {
		err = errors.New("graph does not exists")
		return
	}
	stor.mutex.Lock()
	g, err = stor.mas[masID].graph.toSchema()
	stor.mutex.Unlock()
	return
}

// newLocalStorage returns Storage interface with localStorage type
func newLocalStorage() storage {
	var temp localStorage
	temp.mutex = &sync.Mutex{}
	return &temp
}
