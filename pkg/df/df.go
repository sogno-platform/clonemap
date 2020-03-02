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

// Package df implements the Directory Facilitator
package df

import (
	"errors"
	"io/ioutil"
	"log"
	"os"
)

// DF contains storage of directory facilitator
type DF struct {
	stor     storage     // interface for local or distributed storage
	logInfo  *log.Logger // logger for info logging
	logError *log.Logger // logger for error logging
}

// StartDF starts the df
func StartDF() (err error) {
	df := &DF{logError: log.New(os.Stderr, "[ERROR] ", log.LstdFlags)}
	// create storage and deployment object according to specified deployment type
	err = df.init()
	if err != nil {
		return
	}
	// start to listen and serve requests
	err = df.listen()

	return
}

// init initializes the storage.
func (df *DF) init() (err error) {
	logType := os.Getenv("CLONEMAP_LOG_LEVEL")
	switch logType {
	case "info":
		df.logInfo = log.New(os.Stdout, "[INFO] ", log.LstdFlags)
	case "error":
		df.logInfo = log.New(ioutil.Discard, "", log.LstdFlags)
	default:
		err = errors.New("Wrong log type: " + logType)
		return
	}
	df.logInfo.Println("Starting DF")

	deplType := os.Getenv("CLONEMAP_DEPLOYMENT_TYPE")
	switch deplType {
	case "local":
		df.logInfo.Println("Local storage")
		df.stor = newLocalStorage()
	case "minikube":
		df.logInfo.Println("etcd storage")
		df.stor, err = newEtcdStorage(df.logError)
	case "production":
		df.logInfo.Println("etcd storage")
		df.stor, err = newEtcdStorage(df.logError)
	default:
		err = errors.New("Wrong deployment type: " + deplType)
	}
	return
}
