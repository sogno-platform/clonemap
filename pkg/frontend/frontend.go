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

package frontend

import (
	"errors"
	"io/ioutil"
	"log"
	"os"
	"time"

	amsclient "git.rwth-aachen.de/acs/public/cloud/mas/clonemap/pkg/ams/client"
	dfclient "git.rwth-aachen.de/acs/public/cloud/mas/clonemap/pkg/df/client"
	logclient "git.rwth-aachen.de/acs/public/cloud/mas/clonemap/pkg/logger/client"
	"git.rwth-aachen.de/acs/public/cloud/mas/clonemap/pkg/schemas"
)

// Frontend frontend
type Frontend struct {
	amsClient *amsclient.Client
	dfClient  *dfclient.Client
	logClient *logclient.Client
	logInfo   *log.Logger // logger for info logging
	logError  *log.Logger // logger for error logging
}

// StartFrontend start
func StartFrontend() (err error) {
	fe := &Frontend{
		amsClient: amsclient.New(time.Second*60, time.Second*1, 4),
		dfClient:  dfclient.New(time.Second*60, time.Second*1, 4),
		logClient: logclient.New(time.Second*60, time.Second*1, 4),
		logError:  log.New(os.Stderr, "[ERROR] ", log.LstdFlags),
	}
	logType := os.Getenv("CLONEMAP_LOG_LEVEL")
	switch logType {
	case "info":
		fe.logInfo = log.New(os.Stdout, "[INFO] ", log.LstdFlags)
	case "error":
		fe.logInfo = log.New(ioutil.Discard, "", log.LstdFlags)
	default:
		err = errors.New("Wrong log type: " + logType)
		return
	}
	fe.logInfo.Println("Starting DF")
	serv := fe.server(13000)
	if err != nil {
		fe.logError.Println(err)
		return
	}
	err = fe.listen(serv)
	if err != nil {
		fe.logError.Println(err)
	}
	return
}

// getModuleStatus returns the on/off status of all modules
func (fe *Frontend) getModuleStatus() (mods schemas.ModuleStatus, err error) {
	mods.Logger = fe.logClient.Alive()
	mods.Core = fe.amsClient.Alive()
	mods.DF = fe.dfClient.Alive()
	return
}
