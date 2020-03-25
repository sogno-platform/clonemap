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
	"net/http"
	"os"
	"strconv"
	"testing"
	"time"

	amsclient "git.rwth-aachen.de/acs/public/cloud/mas/clonemap/pkg/ams/client"
	"git.rwth-aachen.de/acs/public/cloud/mas/clonemap/pkg/common/httpreply"
	"git.rwth-aachen.de/acs/public/cloud/mas/clonemap/pkg/schemas"
)

func TestAMS(t *testing.T) {
	// start stub server
	go stubListen()

	os.Setenv("CLONEMAP_DEPLOYMENT_TYPE", "local")
	os.Setenv("CLONEMAP_STORAGE_TYPE", "local")
	os.Setenv("CLONEMAP_STUB_HOSTNAME", "localhost")
	os.Setenv("CLONEMAP_LOG_LEVEL", "error")
	ams := &AMS{}
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
	mux := http.NewServeMux()
	mux.HandleFunc("/api/", ams.handleAPI)
	s := &http.Server{
		Addr:    ":10000",
		Handler: mux,
	}

	// start dummy client
	go dummyClient(s, t)

	// start ams server
	err = s.ListenAndServe()
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
	return
}

// dummyClient makes requests to ams and terminates ams server at end
func dummyClient(s *http.Server, t *testing.T) {
	time.Sleep(time.Second * 1)
	amsclient.Host = "localhost"
	amsclient.Port = 10000

	var err error
	var httpStatus int
	_, httpStatus, err = amsclient.GetCloneMAP()
	if err != nil {
		t.Error(err)
	}
	if httpStatus != http.StatusOK {
		t.Error("Error GetCloneMAP " + strconv.Itoa(httpStatus))
	}
	mas := schemas.MASConfig{
		Spec: schemas.MASSpec{
			Name:               "test",
			AgencyImage:        "agent",
			ImagePullSecret:    "",
			NumAgentsPerAgency: 10,
			Logging:            false,
			MQTT:               false,
			DF:                 false,
		},
		Agents: []schemas.AgentSpec{
			schemas.AgentSpec{
				Name:  "test1",
				AType: "test",
			},
			schemas.AgentSpec{
				Name:  "test2",
				AType: "test",
			},
		},
	}
	httpStatus, err = amsclient.PostMAS(mas)
	if err != nil {
		t.Error(err)
	}
	if httpStatus != http.StatusCreated {
		t.Error("Error PostMAS " + strconv.Itoa(httpStatus))
	}

	_, httpStatus, err = amsclient.GetMASs()
	if err != nil {
		t.Error(err)
	}
	if httpStatus != http.StatusOK {
		t.Error("Error GetMASs " + strconv.Itoa(httpStatus))
	}

	_, httpStatus, err = amsclient.GetMAS(0)
	if err != nil {
		t.Error(err)
	}
	if httpStatus != http.StatusOK {
		t.Error("Error GetMAS " + strconv.Itoa(httpStatus))
	}

	_, httpStatus, err = amsclient.GetAgents(0)
	if err != nil {
		t.Error(err)
	}
	if httpStatus != http.StatusOK {
		t.Error("Error GetAgents " + strconv.Itoa(httpStatus))
	}

	_, httpStatus, err = amsclient.GetAgent(0, 0)
	if err != nil {
		t.Error(err)
	}
	if httpStatus != http.StatusOK {
		t.Error("Error GetAgent " + strconv.Itoa(httpStatus))
	}

	_, httpStatus, err = amsclient.GetAgentAddress(0, 0)
	if err != nil {
		t.Error(err)
	}
	if httpStatus != http.StatusOK {
		t.Error("Error GetAgentAddress " + strconv.Itoa(httpStatus))
	}

	_, httpStatus, err = amsclient.GetAgencies(0)
	if err != nil {
		t.Error(err)
	}
	if httpStatus != http.StatusOK {
		t.Error("Error GetAgencies " + strconv.Itoa(httpStatus))
	}

	_, httpStatus, err = amsclient.GetAgencyConfig(0, 0)
	if err != nil {
		t.Error(err)
	}
	if httpStatus != http.StatusOK {
		t.Error("Error GetAgencyConfig " + strconv.Itoa(httpStatus))
	}

	s.Shutdown(nil)
	return
}
