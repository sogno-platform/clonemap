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
	"fmt"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/RWTH-ACS/clonemap/pkg/common/httpretry"
	"github.com/RWTH-ACS/clonemap/pkg/schemas"
)

// deployment interface for interaction with storage
type deployment interface {
	newMAS(masID int, images schemas.ImageGroups, logging bool, mqtt bool,
		df bool) (err error)
	newImageGroup(masID int, imGroup schemas.ImageGroupInfo, logging bool, mqtt bool,
		df bool) (err error)
	scaleImageGroup(masID int, imID int, deltaAgencies int) (err error)
	deleteMAS(masID int) (err error)
}

// localDeployment implements the Cluster interface for a local instance of the MAP
type localDeployment struct {
	hostName   string
	containers []map[string]schemas.StubAgencyConfig
}

// newMAS triggers the cluster manager to start new agency containers
func (localdepl *localDeployment) newMAS(masID int, images schemas.ImageGroups,
	logging bool, mqtt bool, df bool) (err error) {
	if len(localdepl.containers) <= masID {
		for i := 0; i < masID-len(localdepl.containers)+1; i++ {
			localdepl.containers = append(localdepl.containers, make(map[string]schemas.StubAgencyConfig))
		}
	}
	for i := range images.Inst {
		for j := 0; j < len(images.Inst[i].Agencies.Inst); j++ {
			temp := schemas.StubAgencyConfig{
				MASID:        masID,
				AgencyID:     j,
				ImageGroupID: i,
				Image:        images.Inst[i].Config.Image,
				Logging:      logging,
				MQTT:         mqtt,
				DF:           df,
			}
			js, _ := json.Marshal(temp)
			var statusCode int
			httpClient := &http.Client{Timeout: time.Second * 10}
			_, statusCode, err = httpretry.Post(httpClient, "http://"+localdepl.hostName+
				":8000/api/container", " ", js, time.Second*2, 2)
			if err != nil {
				return
			}
			if statusCode != http.StatusCreated {
				err = errors.New("Cannot create agency " + fmt.Sprint(temp))
				return
			}
			localdepl.containers[masID]["mas-"+strconv.Itoa(masID)+"-im-"+strconv.Itoa(i)] = temp
		}
	}
	return
}

// newImageGroup starts a new image group in an existing mas
func (localdepl *localDeployment) newImageGroup(masID int,
	imGroup schemas.ImageGroupInfo, logging bool, mqtt bool, df bool) (err error) {
	for j := 0; j < len(imGroup.Agencies.Inst); j++ {
		temp := schemas.StubAgencyConfig{
			MASID:        masID,
			AgencyID:     j,
			ImageGroupID: imGroup.ID,
			Image:        imGroup.Config.Image,
			Logging:      logging,
			MQTT:         mqtt,
			DF:           df,
		}
		js, _ := json.Marshal(temp)
		var statusCode int
		httpClient := &http.Client{Timeout: time.Second * 10}
		_, statusCode, err = httpretry.Post(httpClient, "http://"+localdepl.hostName+
			":8000/api/container", " ", js, time.Second*2, 2)
		if err == nil {
			if statusCode != http.StatusCreated {
				err = errors.New("cannot create agency")
				return
			}
		}
		localdepl.containers[masID]["mas-"+strconv.Itoa(masID)+"-im-"+strconv.Itoa(imGroup.ID)] =
			temp
	}
	return
}

// scaleImageGroup triggers the cluster manager to start or delete agency containers
func (localdepl *localDeployment) scaleImageGroup(masID int, imID int,
	deltaAgencies int) (err error) {
	temp := localdepl.containers[masID]["mas-"+strconv.Itoa(masID)+"-im-"+strconv.Itoa(imID)]
	for i := 0; i < deltaAgencies; i++ {
		temp.AgencyID++
		js, _ := json.Marshal(temp)
		var statusCode int
		httpClient := &http.Client{Timeout: time.Second * 10}
		_, statusCode, err = httpretry.Post(httpClient, "http://"+localdepl.hostName+
			":8000/api/container", " ", js, time.Second*2, 2)
		if err == nil {
			if statusCode != http.StatusCreated {
				err = errors.New("cannot create agency")
				return
			}
		}
	}
	localdepl.containers[masID]["mas-"+strconv.Itoa(masID)+"-im-"+strconv.Itoa(imID)] = temp
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
