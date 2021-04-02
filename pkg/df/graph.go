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

import (
	"errors"
	"fmt"
	"strconv"

	"git.rwth-aachen.de/acs/public/cloud/mas/clonemap/pkg/schemas"
)

// graph contains the topology of a mas
type graph struct {
	masID int
	node  map[int]*node
	edge  []*edge
}

// node is one node in the topology to which agents can be attached
type node struct {
	id       int
	neighbor map[int]*edge
	agent    []int
}

// edge connects to nodes
type edge struct {
	node1  *node
	node2  *node
	weight float64
}

// graphFromSchema converts a graph in schema format to actual graph format
func graphFromSchema(masID int, gIn schemas.Graph) (gOut graph, err error) {
	gOut = graph{}
	gTemp := graph{
		masID: masID,
	}
	gTemp.node = make(map[int]*node)
	for i := range gIn.Node {
		if _, ok := gTemp.node[gIn.Node[i].ID]; ok {
			err = errors.New("invalid graph")
			return
		}
		gTemp.node[gIn.Node[i].ID] = &node{
			id:       gIn.Node[i].ID,
			agent:    gIn.Node[i].Agent,
			neighbor: make(map[int]*edge),
		}
	}
	for i := range gIn.Edge {
		e := &edge{weight: gIn.Edge[i].Weight}
		if _, ok := gTemp.node[gIn.Edge[i].Node1]; !ok {
			err = errors.New("invalid graph")
			return
		}
		if _, ok := gTemp.node[gIn.Edge[i].Node2]; !ok {
			err = errors.New("invalid graph")
			return
		}
		e.node1 = gTemp.node[gIn.Edge[i].Node1]
		e.node2 = gTemp.node[gIn.Edge[i].Node2]
		if _, ok := e.node1.neighbor[e.node2.id]; ok {
			err = errors.New("invalid graph")
			return
		}
		if _, ok := e.node2.neighbor[e.node1.id]; ok {
			err = errors.New("invalid graph")
			return
		}
		e.node1.neighbor[e.node2.id] = e
		e.node2.neighbor[e.node1.id] = e
		gTemp.edge = append(gTemp.edge, e)
	}
	gOut = gTemp
	return
}

// toSchema converts a graph to schema format
func (gIn *graph) toSchema() (gOut schemas.Graph, err error) {
	for i := range gIn.node {
		n := schemas.Node{
			ID:    gIn.node[i].id,
			Agent: gIn.node[i].agent,
		}
		gOut.Node = append(gOut.Node, n)
	}
	for i := range gIn.edge {
		e := schemas.Edge{
			Weight: gIn.edge[i].weight,
			Node1:  gIn.edge[i].node1.id,
			Node2:  gIn.edge[i].node2.id,
		}
		gOut.Edge = append(gOut.Edge, e)
	}
	return
}

// String prints the graph
func (gIn graph) String() (ret string) {
	ret = "MAS " + strconv.Itoa(gIn.masID) + "\n"
	for i := range gIn.node {
		ret += "Node " + strconv.Itoa(gIn.node[i].id) + ", Neighbors: "
		for j := range gIn.node[i].neighbor {
			ret += strconv.Itoa(j) + ", "
		}
		ret += "Agents: "
		for j := range gIn.node[i].agent {
			ret += strconv.Itoa(gIn.node[i].agent[j]) + ", "
		}
		ret += "\n"
	}
	for i := range gIn.edge {
		ret += "Edge (" + strconv.Itoa(gIn.edge[i].node1.id) + ", " +
			strconv.Itoa(gIn.edge[i].node2.id) + " weight: " +
			fmt.Sprintf("%f", gIn.edge[i].weight) + "\n"
	}
	return
}

// getNodesInRange returns all node ids within specified range around one node
func (gIn *graph) getNodesInRange(nodeID int, dist float64) (ret map[int]float64, err error) {
	distMap := make(map[int]float64)
	ret = make(map[int]float64)
	distMap[nodeID] = 0
	startNode, ok := gIn.node[nodeID]
	if !ok {
		err = errors.New("unknown node")
		return
	}
	for i := range startNode.neighbor {
		nNeighbor := startNode.neighbor[i].node1
		if nNeighbor.id == startNode.id {
			nNeighbor = startNode.neighbor[i].node2
		}
		d := startNode.neighbor[i].weight
		distMap[nNeighbor.id] = d
		if d < dist {
			err = gIn.visitNode(nNeighbor, &distMap, dist)
		}
	}
	for i := range distMap {
		if distMap[i] < dist+0.001 {
			ret[i] = distMap[i]
		}
	}
	return
}

func (gIn *graph) visitNode(n *node, distMap *map[int]float64, maxDist float64) (err error) {
	dist, ok := (*distMap)[n.id]
	if !ok {
		err = errors.New("Error")
		return
	}
	for i := range n.neighbor {
		nNeighbor := n.neighbor[i].node1
		if nNeighbor.id == n.id {
			nNeighbor = n.neighbor[i].node2
		}
		var dCur float64
		d := dist + n.neighbor[i].weight
		if dCur, ok = (*distMap)[nNeighbor.id]; ok {
			if d < dCur {
				(*distMap)[nNeighbor.id] = d
				if d < maxDist+0.001 {
					err = gIn.visitNode(nNeighbor, distMap, maxDist)
				}
			}
		} else {
			(*distMap)[nNeighbor.id] = d
			if d < maxDist+0.001 {
				err = gIn.visitNode(nNeighbor, distMap, maxDist)
			}
		}
	}
	return
}
