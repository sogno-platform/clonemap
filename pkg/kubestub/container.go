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

// Package kubestub : interaction with containers (start and delete)
package kubestub

import (
	"errors"
	"os/exec"
	"strconv"
	"strings"
)

// createBridge creates a new docker bridge network for MAP parts to connect to
func (stub *LocalStub) createBridge() (err error) {
	com := "docker network create clonemap-net"
	cmd := exec.Command("bash", "-c", com)
	cmdOut, err := cmd.Output()
	if err != nil {
		err = errors.New(err.Error() + " " + string(cmdOut))
	}
	return
}

// deleteBridge deletes docker bridge network
func (stub *LocalStub) deleteBridge() (err error) {
	com := "docker network rm clonemap-net"
	cmd := exec.Command("bash", "-c", com)
	cmdOut, err := cmd.Output()
	if err != nil {
		err = errors.New(err.Error() + " " + string(cmdOut))
	}
	return
}

// createAMS starts a new AMS docker image
func (stub *LocalStub) createAMS() (err error) {
	com := "ip route show | grep docker0 | awk '{print $9}'"
	cmd := exec.Command("bash", "-c", com)
	cmdOut, err := cmd.Output()
	if err != nil {
		err = errors.New(err.Error() + " " + strings.Trim(string(cmdOut), "\n"))
	} else {
		ip := strings.Trim(string(cmdOut), "\n")
		com = "docker run -d"
		com += " --add-host=parent-host:" + ip
		com += " -p 30009:9000"
		com += " --name=ams" //.clonemap""
		com += " --hostname=ams"
		com += " --network=clonemap-net"
		com += " -e CLONEMAP_DEPLOYMENT_TYPE=\"local\""
		com += " -e CLONEMAP_STORAGE_TYPE=\"local\""
		com += " -e CLONEMAP_SUFFIX=\".clonemap\""
		com += " -e CLONEMAP_LOG_LEVEL=\"error\""
		com += " ams"
		cmd := exec.Command("bash", "-c", com)
		cmdOut, err := cmd.Output()
		if err != nil {
			err = errors.New(err.Error() + " " + string(cmdOut))
		}
	}
	return
}

// deleteAMS stops amd removes AMS docker image
func (stub *LocalStub) deleteAMS() (err error) {
	com := "docker stop ams" //.clonemap"
	cmd := exec.Command("bash", "-c", com)
	cmdOut, err := cmd.Output()
	if err != nil {
		err = errors.New(err.Error() + " " + string(cmdOut))
		return
	}
	com = "docker rm ams" //.clonemap"
	cmd = exec.Command("bash", "-c", com)
	cmdOut, err = cmd.Output()
	if err != nil {
		err = errors.New(err.Error() + " " + string(cmdOut))
	}
	return
}

// createLogger starts a new Logger docker image
func (stub *LocalStub) createLogger() (err error) {
	com := "docker run -d"
	com += " -p 30011:11000"
	com += " --name=logger" //.clonemap"
	com += " --hostname=logger"
	com += " --network=clonemap-net"
	com += " -e CLONEMAP_DEPLOYMENT_TYPE=\"local\""
	com += " -e CLONEMAP_LOG_LEVEL=\"error\""
	com += " logger"
	cmd := exec.Command("bash", "-c", com)
	cmdOut, err := cmd.Output()
	if err != nil {
		err = errors.New(err.Error() + " " + string(cmdOut))
	}
	return
}

// deleteLogger stops amd removes Logger docker image
func (stub *LocalStub) deleteLogger() (err error) {
	com := "docker stop logger" //.clonemap"
	cmd := exec.Command("bash", "-c", com)
	cmdOut, err := cmd.Output()
	if err != nil {
		err = errors.New(err.Error() + " " + string(cmdOut))
		return
	}
	com = "docker rm logger" //.clonemap"
	cmd = exec.Command("bash", "-c", com)
	cmdOut, err = cmd.Output()
	if err != nil {
		err = errors.New(err.Error() + " " + string(cmdOut))
	}
	return
}

// createDF starts a new DF docker image
func (stub *LocalStub) createDF() (err error) {
	com := "docker run -d"
	com += " -p 30012:12000"
	com += " --name=df" //.clonemap"
	com += " --hostname=df"
	com += " --network=clonemap-net"
	com += " -e CLONEMAP_DEPLOYMENT_TYPE=\"local\""
	com += " -e CLONEMAP_LOG_LEVEL=\"error\""
	com += " df"
	cmd := exec.Command("bash", "-c", com)
	cmdOut, err := cmd.Output()
	if err != nil {
		err = errors.New(err.Error() + " " + string(cmdOut))
	}
	return
}

// deleteDF stops amd removes DF docker image
func (stub *LocalStub) deleteDF() (err error) {
	com := "docker stop df" //.clonemap"
	cmd := exec.Command("bash", "-c", com)
	cmdOut, err := cmd.Output()
	if err != nil {
		err = errors.New(err.Error() + " " + string(cmdOut))
		return
	}
	com = "docker rm df" //.clonemap"
	cmd = exec.Command("bash", "-c", com)
	cmdOut, err = cmd.Output()
	if err != nil {
		err = errors.New(err.Error() + " " + string(cmdOut))
	}
	return
}

// createMQTT starts a new MQTT Broker docker image
func (stub *LocalStub) createMQTT() (err error) {
	com := "docker run -d"
	com += " -p 30883:1883"
	com += " --name=mqtt" //.clonemap"
	com += " --hostname=mqtt"
	com += " --network=clonemap-net"
	com += " -e CLONEMAP_DEPLOYMENT_TYPE=\"local\""
	com += " eclipse-mosquitto"
	cmd := exec.Command("bash", "-c", com)
	cmdOut, err := cmd.Output()
	if err != nil {
		err = errors.New(err.Error() + " " + string(cmdOut))
	}
	return
}

// deleteMQTT stops amd removes MQTT Broker docker image
func (stub *LocalStub) deleteMQTT() (err error) {
	com := "docker stop mqtt" //.clonemap"
	cmd := exec.Command("bash", "-c", com)
	cmdOut, err := cmd.Output()
	if err != nil {
		err = errors.New(err.Error() + " " + string(cmdOut))
		return
	}
	com = "docker rm mqtt" //.clonemap"
	cmd = exec.Command("bash", "-c", com)
	cmdOut, err = cmd.Output()
	if err != nil {
		err = errors.New(err.Error() + " " + string(cmdOut))
	}
	return
}

// createAgency starts a new agency docker image
func (stub *LocalStub) createAgency(image string, agents int, masID int, agencyID int, logging bool,
	mqtt bool, df bool) (err error) {
	com := "docker run -d"
	com += " --name=mas-" + strconv.Itoa(masID) + "-agency-" + strconv.Itoa(agencyID) + ".mas" +
		strconv.Itoa(masID) + "agencies" //.clonemap"
	com += " --hostname=mas-" + strconv.Itoa(masID) + "-agency-" + strconv.Itoa(agencyID)
	com += " --network=clonemap-net"
	//com += " -e CLONEMAP_SUFFIX=\".clonemap\" "
	if logging {
		com += " -e CLONEMAP_LOGGING=\"ON\" "
	} else {
		com += " -e CLONEMAP_LOGGING=\"OFF\" "
	}
	if mqtt {
		com += " -e CLONEMAP_MQTT=\"ON\" "
	} else {
		com += " -e CLONEMAP_MQTT=\"OFF\" "
	}
	if df {
		com += " -e CLONEMAP_DF=\"ON\" "
	} else {
		com += " -e CLONEMAP_DF=\"OFF\" "
	}
	com += " -e CLONEMAP_LOG_LEVEL=\"error\" "

	com += image
	cmd := exec.Command("bash", "-c", com)
	cmdOut, err := cmd.Output()
	if err != nil {
		err = errors.New(err.Error() + " " + string(cmdOut))
	}
	return
}

// deleteAgency stops and removes agency docker image
func (stub *LocalStub) deleteAgency(masID int, agencyID int) (err error) {
	com := "docker stop "
	com += "mas-" + strconv.Itoa(masID) + "-agency-" + strconv.Itoa(agencyID) + ".mas" +
		strconv.Itoa(masID) + "agencies" //.clonemap"
	cmd := exec.Command("bash", "-c", com)
	cmdOut, err := cmd.Output()
	if err != nil {
		err = errors.New(err.Error() + " " + string(cmdOut))
		return
	}
	com = "docker rm "
	com += "mas-" + strconv.Itoa(masID) + "-agency-" + strconv.Itoa(agencyID) + ".mas" +
		strconv.Itoa(masID) + "agencies" //.clonemap"
	cmd = exec.Command("bash", "-c", com)
	cmdOut, err = cmd.Output()
	if err != nil {
		err = errors.New(err.Error() + " " + string(cmdOut))
	}
	return
}
