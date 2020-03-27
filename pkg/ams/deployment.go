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

// provides an interface for interaction with cluster (local or Kubernetes)
// Interacting with the cluster allows to start and delete agencies

import (
	"encoding/json"
	"errors"
	"net/http"
	"os"
	"strconv"
	"time"

	"git.rwth-aachen.de/acs/public/cloud/mas/clonemap/pkg/common/httpretry"
	"git.rwth-aachen.de/acs/public/cloud/mas/clonemap/pkg/schemas"
)

// deployment interface for interaction with storage
type deployment interface {
	newMAS(masID int, image string, pullSecret string, numAgencies map[int]int, logging bool, mqtt bool,
		df bool) (err error)
	scaleMAS(masID int, deltaAgencies int) (err error)
	deleteMAS(masID int) (err error)
}

// localDeployment implements the Cluster interface for a local instance of the MAP
type localDeployment struct {
	hostName string
}

// newMAS triggers the cluster manager to start new agency containers
func (localdepl *localDeployment) newMAS(masID int, image string, pullSecret string,
	numAgencies map[int]int, logging bool, mqtt bool, df bool) (err error) {
	for i := 0; i < numAgencies[0]; i++ {
		temp := schemas.StubAgencyConfig{
			MASID:     masID,
			AgencyID:  i,
			NumAgents: 10,
			Image:     image,
			Logging:   logging,
			MQTT:      mqtt,
			DF:        df,
		}
		js, _ := json.Marshal(temp)
		var statusCode int
		httpClient := &http.Client{Timeout: time.Second * 10}
		_, statusCode, err = httpretry.Post(httpClient, "http://"+localdepl.hostName+
			":8000/api/container", " ", js, time.Second*2, 2)
		if err == nil {
			if statusCode != http.StatusCreated {
				err = errors.New("Cannot create agency")
			}
		}
	}
	return
}

// scaleMAS triggers the cluster manager to start or delete agency containers
func (localdepl *localDeployment) scaleMAS(masID int, deltaAgencies int) (err error) {
	// ToDo
	return
}

// deleteMAS triggers the cluster manager to delete all agency containers
func (localdepl *localDeployment) deleteMAS(masID int) (err error) {
	httpClient := &http.Client{Timeout: time.Second * 10}
	_, err = httpretry.Delete(httpClient, "http://"+localdepl.hostName+
		":8000/api/container/"+strconv.Itoa(masID), nil,
		time.Second*2, 2)
	return
}

// newLocalDeployment returns Deployment interface with localCluster type
func newLocalDeployment() (depl deployment, err error) {
	var temp localDeployment
	if val, ok := os.LookupEnv("CLONEMAP_STUB_HOSTNAME"); ok {
		temp.hostName = val
	} else {
		temp.hostName = "parent-host"
	}
	depl = &temp
	err = nil
	return
}
