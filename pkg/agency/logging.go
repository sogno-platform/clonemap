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

package agency

import (
	"errors"
	"fmt"
	"log"
	"os"
	"sync"
	"time"

	"git.rwth-aachen.de/acs/public/cloud/mas/clonemap/pkg/logger/client"
	"git.rwth-aachen.de/acs/public/cloud/mas/clonemap/pkg/schemas"
)

// Logger logs data to logging service
type Logger struct {
	agentID  int
	client   *loggerClient
	mutex    *sync.Mutex
	config   schemas.LogConfig
	logError *log.Logger
	logInfo  *log.Logger
	active   bool
}

// NewLog sends a new logging message to the logging service
func (log *Logger) NewLog(logType string, message string, data string) (err error) {
	log.mutex.Lock()
	if !log.active {
		log.mutex.Unlock()
		return errors.New("log not active")
	}
	log.mutex.Unlock()
	if logType != "error" && logType != "debug" && logType != "status" && logType != "msg" &&
		logType != "app" {
		err = errors.New("UnknownLogType")
		return
	}
	log.mutex.Lock()
	if (logType == "msg" && !log.config.Msg) || (logType == "app" && !log.config.App) ||
		(logType == "debug" && !log.config.Debug) || (logType == "status" && !log.config.Status) {
		log.mutex.Unlock()
		return
	}
	time.Sleep(time.Millisecond * 5)
	tStamp := time.Now()
	log.mutex.Unlock()
	msg := schemas.LogMessage{
		AgentID:        log.agentID,
		Timestamp:      tStamp,
		LogType:        logType,
		Message:        message,
		AdditionalData: data}
	log.client.logIn <- msg

	return
}

// UpdateState overrides the state stored in database
func (log *Logger) UpdateState(state string) (err error) {
	log.mutex.Lock()
	if !log.active {
		log.mutex.Unlock()
		return errors.New("log not active")
	}
	log.mutex.Unlock()
	agState := schemas.State{
		MASID:     log.client.masID,
		AgentID:   log.agentID,
		Timestamp: time.Now(),
		State:     state}
	_, err = client.PutState(agState)
	return
}

// RestoreState loads state saved in database and return it
func (log *Logger) RestoreState() (state string, err error) {
	log.mutex.Lock()
	if !log.active {
		log.mutex.Unlock()
		err = errors.New("log not active")
		return
	}
	log.mutex.Unlock()
	var agState schemas.State
	agState, _, err = client.GetState(log.client.masID, log.agentID)
	state = agState.State
	return
}

// newLogger craetes a new object of type Logger
func newLogger(agentID int, client *loggerClient, config schemas.LogConfig, logErr *log.Logger,
	logInf *log.Logger) (log *Logger) {
	log = &Logger{
		agentID:  agentID,
		client:   client,
		mutex:    &sync.Mutex{},
		config:   config,
		logError: logErr,
		logInfo:  logInf,
		active:   true,
	}
	return
}

// close closes the logger
func (log *Logger) close() {
	log.mutex.Lock()
	log.logInfo.Println("Closing Logger of agent ", log.agentID)
	log.active = false
	log.mutex.Unlock()
	return
}

// loggerClient is the agency client for the logger
type loggerClient struct {
	masID    int
	logIn    chan schemas.LogMessage // logging inbox
	active   bool                    // indicates if logging is active (switch via env)
	logError *log.Logger
	logInfo  *log.Logger
}

// storeLogs periodically requests the logging service to store log messages
func (log *loggerClient) storeLogs() (err error) {
	if log.active {
		for {
			if len(log.logIn) > 0 {
				numMsg := len(log.logIn)
				logMsgs := make([]schemas.LogMessage, numMsg, numMsg)
				for i := 0; i < numMsg; i++ {
					logMsgs[i] = <-log.logIn
					logMsgs[i].MASID = log.masID
				}
				_, err = client.PostLogs(log.masID, logMsgs)
				if err != nil {
					log.logError.Println(err)
					for i := range logMsgs {
						log.logIn <- logMsgs[i]
					}
					continue
				}
			}
			tempTime := time.Now()
			for {
				time.Sleep(100 * time.Millisecond)
				if time.Since(tempTime).Seconds() > 15 || len(log.logIn) > 50 {
					break
				}
			}
		}
	} else {
		for {
			// print messages to stdout if logger is turned off
			logMsg := <-log.logIn
			fmt.Println(logMsg)
		}
	}
}

// newLoggerClient creates an agency logger client
func newLoggerClient(masID int, logErr *log.Logger, logInf *log.Logger) (log *loggerClient) {
	log = &loggerClient{
		masID:    masID,
		active:   false,
		logError: logErr,
		logInfo:  logInf,
	}
	temp := os.Getenv("CLONEMAP_LOGGING")
	if temp == "ON" {
		log.active = true
	}
	log.logIn = make(chan schemas.LogMessage, 10000)
	go log.storeLogs()
	log.logInfo.Println("Created new logger client; status: ", log.active)
	return
}
