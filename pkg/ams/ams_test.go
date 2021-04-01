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
	"context"
	"log"
	"net/http"
	"os"
	"strconv"
	"testing"
	"time"

	agclient "git.rwth-aachen.de/acs/public/cloud/mas/clonemap/pkg/agency/client"
	"git.rwth-aachen.de/acs/public/cloud/mas/clonemap/pkg/client"
	"git.rwth-aachen.de/acs/public/cloud/mas/clonemap/pkg/common/httpreply"
	dfclient "git.rwth-aachen.de/acs/public/cloud/mas/clonemap/pkg/df/client"
	"git.rwth-aachen.de/acs/public/cloud/mas/clonemap/pkg/schemas"
)

func TestAMS(t *testing.T) {
	// start stub server
	go stubListen()

	os.Setenv("CLONEMAP_DEPLOYMENT_TYPE", "local")
	os.Setenv("CLONEMAP_STORAGE_TYPE", "local")
	os.Setenv("CLONEMAP_STUB_HOSTNAME", "localhost")
	os.Setenv("CLONEMAP_LOG_LEVEL", "error")
	ams := &AMS{
		logError:     log.New(os.Stderr, "[ERROR] ", log.LstdFlags),
		logInfo:      log.New(os.Stdout, "[INFO] ", log.LstdFlags),
		agencyClient: agclient.New(time.Second*60, time.Second*1, 4),
		dfClient:     dfclient.New(time.Second*60, time.Second*1, 4),
	}
	// create storage and deployment object according to specified deployment type
	err := ams.init()
	if err != nil {
		t.Error(err)
	}
	cmap := schemas.CloneMAP{
		Version: "v0.1",
		Uptime:  time.Now(),
	}
	ams.stor.setCloneMAPInfo(cmap)
	// start to listen and serve requests
	serv := ams.server(10001)
	// r := mux.NewRouter()
	// r.PathPrefix("/api/").HandlerFunc(ams.handleAPI)

	// mux := http.NewServeMux()
	// mux.HandleFunc("/api/", ams.handleAPI)

	// serv := &http.Server{
	// 	Addr:    ":10001",
	// 	Handler: r,
	// }

	// start dummy client
	go dummyClient(serv, t)

	// start ams server
	err = serv.ListenAndServe()
	if err != nil && err != http.ErrServerClosed {
		t.Error(err)
	}
}

// listen opens a http server listening and serving request
func stubListen() (err error) {
	mux := http.NewServeMux()
	mux.HandleFunc("/api/", stubHandler)
	s := &http.Server{
		Addr:    ":8000",
		Handler: mux,
	}
	err = s.ListenAndServe()
	return
}

// stubHandler answers with created
func stubHandler(w http.ResponseWriter, r *http.Request) {
	httpreply.Created(w, nil, "text/plain", []byte("Ressource Created"))
}

// dummyClient makes requests to ams and terminates ams server at end
func dummyClient(s *http.Server, t *testing.T) {
	time.Sleep(time.Second * 1)
	amsClient := client.NewAMSClient(time.Second*60, time.Second*1, 4)
	amsClient.Host = "localhost"
	amsClient.Port = 10001

	var err error
	var httpStatus int
	_, httpStatus, err = amsClient.GetCloneMAP()
	if err != nil {
		t.Error(err)
	}
	if httpStatus != http.StatusOK {
		t.Error("Error GetCloneMAP " + strconv.Itoa(httpStatus))
	}
	mas := schemas.MASSpec{
		Config: schemas.MASConfig{
			Name:               "test",
			NumAgentsPerAgency: 10,
			Logger: schemas.LoggerConfig{
				Active: false,
			},
			MQTT: schemas.MQTTConfig{
				Active: false,
			},
			DF: schemas.DFConfig{
				Active: false,
			},
		},
		ImageGroups: []schemas.ImageGroupSpec{
			{
				Config: schemas.ImageGroupConfig{
					Image:      "agent",
					PullSecret: "",
				},
				Agents: []schemas.AgentSpec{
					{
						Name:  "test1",
						AType: "test",
					},
					{
						Name:  "test2",
						AType: "test",
					},
				},
			},
		},
	}
	httpStatus, err = amsClient.PostMAS(mas)
	if err != nil {
		t.Error(err)
	}
	if httpStatus != http.StatusCreated {
		t.Error("Error PostMAS " + strconv.Itoa(httpStatus))
	}

	_, httpStatus, err = amsClient.GetMASsShort()
	if err != nil {
		t.Error(err)
	}
	if httpStatus != http.StatusOK {
		t.Error("Error GetMASs " + strconv.Itoa(httpStatus))
	}

	_, httpStatus, err = amsClient.GetMAS(0)
	if err != nil {
		t.Error(err)
	}
	if httpStatus != http.StatusOK {
		t.Error("Error GetMAS " + strconv.Itoa(httpStatus))
	}

	_, httpStatus, err = amsClient.GetAgents(0)
	if err != nil {
		t.Error(err)
	}
	if httpStatus != http.StatusOK {
		t.Error("Error GetAgents " + strconv.Itoa(httpStatus))
	}

	_, httpStatus, err = amsClient.GetAgent(0, 0)
	if err != nil {
		t.Error(err)
	}
	if httpStatus != http.StatusOK {
		t.Error("Error GetAgent " + strconv.Itoa(httpStatus))
	}

	_, httpStatus, err = amsClient.GetAgentAddress(0, 0)
	if err != nil {
		t.Error(err)
	}
	if httpStatus != http.StatusOK {
		t.Error("Error GetAgentAddress " + strconv.Itoa(httpStatus))
	}

	_, httpStatus, err = amsClient.GetAgencies(0)
	if err != nil {
		t.Error(err)
	}
	if httpStatus != http.StatusOK {
		t.Error("Error GetAgencies " + strconv.Itoa(httpStatus))
	}

	_, httpStatus, err = amsClient.GetAgencyInfo(0, 0, 0)
	if err != nil {
		t.Error(err)
	}
	if httpStatus != http.StatusOK {
		t.Error("Error GetAgencyConfig " + strconv.Itoa(httpStatus))
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	s.Shutdown(ctx)
}
