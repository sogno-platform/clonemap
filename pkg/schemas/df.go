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

package schemas

import "time"

// DFConfig contains the host and port configuration of the DF and indicates if it is active
type DFConfig struct {
	Active bool   `json:"active"`         // indicates if DF is active/usable
	Host   string `json:"host,omitempty"` // hostname of DF
	Port   int    `json:"port,omitempty"` // port of DF
}

// Service holds information about an agent service that can be registered and searched with the DF
type Service struct {
	GUID      string    `json:"id"`        // unique svc id
	AgentID   int       `json:"agentid"`   // ID of agent who registered service
	NodeID    int       `json:"nodeid"`    // ID of node agent is located at
	MASID     int       `json:"masid"`     // ID of MAS agent lives in
	CreatedAt time.Time `json:"createdat"` // time of service creation
	ChangedAt time.Time `json:"changedat"` // time of last change
	Desc      string    `json:"desc"`      // description
	Dist      float64   `json:"dist"`      // distance (only if local search was executed)
}

// Graph stores one mas graph for topological search
type Graph struct {
	Node []Node `json:"node"` // list of graph nodes
	Edge []Edge `json:"edge"` // list of graph edges
}

// Node is one graph node
type Node struct {
	ID    int   `json:"id"`               // unique ID of node
	Agent []int `json:"agents,omitempty"` // list of agents attached to node
}

// Edge is one dge of graph
type Edge struct {
	Node1  int     `json:"n1"`     // id of node 1
	Node2  int     `json:"n2"`     // id of node 2
	Weight float64 `json:"weight"` // weight of edge
}
