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

// Package plugnplay implements a component which automatically triggers the creation of an agent
// in case a new IoT device is connected to the platform. For this purpose it subscribes to the
// MQTT topic "register" and invokes the AMS API every time a message is received
package plugnplay

import (
	"errors"
	"io/ioutil"
	"log"
	"os"
)

// PnP implements the plug and play mechanism
type PnP struct {
	logInfo  *log.Logger // logger for info logging
	logError *log.Logger // logger for error logging
	mqttCli  *mqttClient
}

// StartPnP starts an PnP instance. It initializes the storage object and starts the API server.
func StartPnP() (err error) {
	pnp := &PnP{logError: log.New(os.Stderr, "[ERROR] ", log.LstdFlags)}
	// create storage and deployment object according to specified deployment type
	err = pnp.init()
	if err != nil {
		return
	}
	// start to listen and serve requests
	// err = pnp.listen()
	return
}

// init initializes the storage.
func (pnp *PnP) init() (err error) {
	logType := os.Getenv("CLONEMAP_LOG_LEVEL")
	switch logType {
	case "info":
		pnp.logInfo = log.New(os.Stdout, "[INFO] ", log.LstdFlags)
	case "error":
		pnp.logInfo = log.New(ioutil.Discard, "", log.LstdFlags)
	default:
		err = errors.New("Wrong log type: " + logType)
		return
	}
	pnp.logInfo.Println("Starting Plug&Play")

	pnp.mqttCli = newMQTTClient("mqtt", 1883, "pnp", pnp.logError, pnp.logInfo)
	pnp.mqttCli.init()
	err = pnp.mqttCli.subscribe("register", 0)

	// storType := os.Getenv("CLONEMAP_STORAGE_TYPE")
	// switch storType {
	// case "local":
	// 	pnp.logInfo.Println("Local storage")
	// 	pnp.stor = newLocalStorage()
	// case "etcd":
	// 	if deplType == "local" {
	// 		err = errors.New("etcd storage can not be used with local deployment")
	// 		return
	// 	}
	// 	pnp.logInfo.Println("ectd storage")
	// 	pnp.stor, err = newEtcdStorage(pnp.logError)
	// case "fiware":
	// 	pnp.logInfo.Println("FiWare storage")
	// 	pnp.stor, err = newFiwareStorage(pnp.logError)
	// default:
	// 	err = errors.New("Wrong storage type: " + storType)
	// }
	return
}
