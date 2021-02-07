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

// Package kubestub simulates the behavior of the Kubernetes API for local execution of clonemap
// It provides an API to be used by the MAS. Functionalities are: Startup of AMS at beginning,
// start of agencies and proper termination of all MAP-parts
package kubestub

import (
	"fmt"
	"os"
	"os/signal"
	"strconv"
	"syscall"

	"git.rwth-aachen.de/acs/public/cloud/mas/clonemap/pkg/schemas"
)

// LocalStub holds context information such as a list of all started agencies
type LocalStub struct {
	// address of the local stub
	address string
	// list of all agencies; necessary in order to prevent from starting agencies with same name
	// and to stop all agencies upon termination
	agencies     []schemas.StubAgencyConfig
	startModules bool
	fiware       bool
	mqtt         bool
	logger       bool
	df           bool
	pnp          bool
	frontend     bool
	logLevel     string
}

// StartLocalStub starts the local stub. The AMS is started and a server for AMS interaction is
// created
func StartLocalStub() {
	var err error
	// initialization
	cntxt := &LocalStub{}
	_, cntxt.startModules = os.LookupEnv("CLONEMAP_START_MODULES")
	_, cntxt.mqtt = os.LookupEnv("CLONEMAP_MODULE_MQTT")
	_, cntxt.fiware = os.LookupEnv("CLONEMAP_MODULE_FIWARE")
	_, cntxt.logger = os.LookupEnv("CLONEMAP_MODULE_LOGGER")
	_, cntxt.df = os.LookupEnv("CLONEMAP_MODULE_DF")
	_, cntxt.pnp = os.LookupEnv("CLONEMAP_MODULE_PNP")
	_, cntxt.frontend = os.LookupEnv("CLONEMAP_MODULE_FRONTEND")
	cntxt.logLevel, _ = os.LookupEnv("CLONEMAP_LOG_LEVEL")
	if cntxt.logLevel == "" {
		cntxt.logLevel = "error"
	}

	if cntxt.startModules {
		fmt.Println("Create Bridge Network")
		err = cntxt.createBridge()
		if err != nil {
			fmt.Println(err)
			return
		}
		fmt.Println("Create AMS Container")
		err = cntxt.createAMS()
		if err != nil {
			fmt.Println(err)
			return
		}
		if cntxt.mqtt || cntxt.fiware {
			fmt.Println("Create MQTT Broker Container")
			err = cntxt.createMQTT()
			if err != nil {
				fmt.Println(err)
				return
			}
		}
		if cntxt.fiware {
			fmt.Println("Ceate Fiware Containers")
			err = cntxt.createFiware()
			if err != nil {
				fmt.Println(err)
				return
			}
		}
		if cntxt.logger {
			fmt.Println("Create Logger Container")
			err = cntxt.createLogger()
			if err != nil {
				fmt.Println(err)
				return
			}
		}
		if cntxt.df {
			fmt.Println("Create DF Container")
			err = cntxt.createDF()
			if err != nil {
				fmt.Println(err)
				return
			}
		}
		if cntxt.pnp {
			fmt.Println("Create Plugnplay Container")
			err = cntxt.createPnP()
			if err != nil {
				fmt.Println(err)
				return
			}
		}
		if cntxt.frontend {
			fmt.Println("Create Frontend Container")
			err = cntxt.createFrontend()
			if err != nil {
				fmt.Println(err)
				return
			}
		}
	}
	fmt.Println("Ready")

	// catch kill signal in order to terminate MAP parts before exiting
	var gracefulStop = make(chan os.Signal)
	signal.Notify(gracefulStop, syscall.SIGTERM)
	signal.Notify(gracefulStop, syscall.SIGINT)
	go cntxt.terminate(gracefulStop)

	// start API server
	err = cntxt.listen()
	if err != nil {
		fmt.Println(err)
	}
}

// terminate takes care of terminating all parts of the MAP before exiting. It is to be called as a
// goroutine and waits until an OS signal is inserted into the channel gracefulStop
func (stub *LocalStub) terminate(gracefulStop chan os.Signal) {
	var err error
	sig := <-gracefulStop
	fmt.Printf("Caught sig: %+v\n", sig)
	for i := range stub.agencies {
		agencyName := "mas-" + strconv.Itoa(stub.agencies[i].MASID) + "-im-" +
			strconv.Itoa(stub.agencies[i].ImageGroupID) + "-agency-" +
			strconv.Itoa(stub.agencies[i].AgencyID)
		fmt.Println("Stop Agency Container " + agencyName)
		err = stub.deleteAgency(stub.agencies[i].MASID, stub.agencies[i].ImageGroupID,
			stub.agencies[i].AgencyID)
		if err != nil {
			fmt.Println(err)
			// os.Exit(0)
		}
	}

	if stub.startModules {
		fmt.Println("Stop AMS Container")
		err = stub.deleteAMS()
		if err != nil {
			fmt.Println(err)
			os.Exit(0)
		}
		if stub.logger {
			fmt.Println("Stop Logger Container")
			err = stub.deleteLogger()
			if err != nil {
				fmt.Println(err)
				os.Exit(0)
			}
		}
		if stub.df {
			fmt.Println("Stop DF Container")
			err = stub.deleteDF()
			if err != nil {
				fmt.Println(err)
				os.Exit(0)
			}
		}
		if stub.pnp {
			fmt.Println("Stop Plugnplay Container")
			err = stub.deletePnP()
			if err != nil {
				fmt.Println(err)
				os.Exit(0)
			}
		}
		if stub.frontend {
			fmt.Println("Stop Frontend Container")
			err = stub.deleteFrontend()
			if err != nil {
				fmt.Println(err)
				os.Exit(0)
			}
		}
		if stub.mqtt || stub.fiware {
			fmt.Println("Stop MQTT Broker Container")
			err = stub.deleteMQTT()
			if err != nil {
				fmt.Println(err)
				os.Exit(0)
			}
		}
		if stub.fiware {
			fmt.Println("Stop FIWARE Containers")
			err = stub.deleteFiware()
			if err != nil {
				fmt.Println(err)
				os.Exit(0)
			}
		}

		fmt.Println("Delete Bridge Network")
		err = stub.deleteBridge()
		if err != nil {
			fmt.Println(err)
		}
	}
	os.Exit(0)
}
