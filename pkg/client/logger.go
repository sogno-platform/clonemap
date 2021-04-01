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

// Package client contains code for interaction with agent
package client

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"strconv"
	"sync"
	"time"

	"git.rwth-aachen.de/acs/public/cloud/mas/clonemap/pkg/common/httpretry"
	"git.rwth-aachen.de/acs/public/cloud/mas/clonemap/pkg/schemas"
)

// LoggerClient is the ams client
type LoggerClient struct {
	httpClient *http.Client  // http client
	host       string        // ams host name
	port       int           // ams port
	delay      time.Duration // delay between two retries
	numRetries int           // number of retries
}

// Alive tests if alive
func (cli *LoggerClient) Alive() (alive bool) {
	alive = false
	_, httpStatus, err := httpretry.Get(cli.httpClient, cli.prefix()+"/api/alive", time.Second*2, 2)
	if err == nil && httpStatus == http.StatusOK {
		alive = true
	}
	return
}

// PostLogs posts new log messages to logger
func (cli *LoggerClient) PostLogs(masID int, logs []schemas.LogMessage) (httpStatus int, err error) {
	js, _ := json.Marshal(logs)
	_, httpStatus, err = httpretry.Post(cli.httpClient, cli.prefix()+"/api/logging/"+
		strconv.Itoa(masID)+"/list", "application/json", js, time.Second*2, 4)
	return
}

// GetLatestLogs gets log messages
func (cli *LoggerClient) GetLatestLogs(masID int, agentID int, topic string,
	num int) (msgs []schemas.LogMessage, httpStatus int, err error) {
	var body []byte
	body, httpStatus, err = httpretry.Get(cli.httpClient, cli.prefix()+"/api/logging/"+
		strconv.Itoa(masID)+"/"+strconv.Itoa(agentID)+"/"+topic+"/latest/"+
		strconv.Itoa(num), time.Second*2, 4)
	if err != nil {
		return
	}
	err = json.Unmarshal(body, &msgs)
	if err != nil {
		msgs = []schemas.LogMessage{}
	}
	return
}

// PutState updates the state
func (cli *LoggerClient) PutState(state schemas.State) (httpStatus int, err error) {
	js, _ := json.Marshal(state)
	_, httpStatus, err = httpretry.Put(cli.httpClient, cli.prefix()+"/api/state/"+
		strconv.Itoa(state.MASID)+"/"+strconv.Itoa(state.AgentID), js,
		time.Second*2, 4)
	return
}

// UpdateStates updates the state
func (cli *LoggerClient) UpdateStates(masID int, states []schemas.State) (httpStatus int, err error) {
	js, _ := json.Marshal(states)
	_, httpStatus, err = httpretry.Put(cli.httpClient, cli.prefix()+"/api/state/"+
		strconv.Itoa(masID)+"/list", js, time.Second*2, 4)
	return
}

// GetState requests state from logger
func (cli *LoggerClient) GetState(masID int, agentID int) (state schemas.State, httpStatus int,
	err error) {
	var body []byte
	body, httpStatus, err = httpretry.Get(cli.httpClient, cli.prefix()+"/api/state/"+
		strconv.Itoa(masID)+"/"+strconv.Itoa(agentID), time.Second*2, 4)
	if err != nil {
		return
	}
	err = json.Unmarshal(body, &state)
	if err != nil {
		state = schemas.State{}
	}
	return
}

func (cli *LoggerClient) prefix() (ret string) {
	ret = "http://" + cli.host + ":" + strconv.Itoa(cli.port)
	return
}

// NewLoggerClient creates a new Logger client
func NewLoggerClient(host string, port int, timeout time.Duration, del time.Duration,
	numRet int) (cli *LoggerClient, err error) {
	cli = &LoggerClient{
		httpClient: &http.Client{Timeout: timeout},
		host:       host,
		port:       port,
		delay:      del,
		numRetries: numRet,
	}
	if !cli.Alive() {
		err = errors.New("Logger Module is not running on " + host + ":" + strconv.Itoa(port))
	}
	return
}

// LogCollector collects logs and states and sends them to the Logger service
// one LogCollector per agency is used; each agent obtains one AgentLogger
type LogCollector struct {
	masID    int
	logIn    chan schemas.LogMessage // logging inbox
	stateIn  chan schemas.State
	client   *LoggerClient
	config   schemas.LoggerConfig
	logError *log.Logger
	logInfo  *log.Logger
}

// storeLogs periodically requests the logging service to store log messages
func (logCol *LogCollector) storeLogs() (err error) {
	if logCol.config.Active {
		for {
			if len(logCol.logIn) > 0 {
				numMsg := len(logCol.logIn)
				logMsgs := make([]schemas.LogMessage, numMsg)
				index := 0
				for i := 0; i < numMsg; i++ {
					logMsg := <-logCol.logIn
					if (logMsg.Topic == "msg" && !logCol.config.TopicMsg) ||
						(logMsg.Topic == "app" && !logCol.config.TopicApp) ||
						(logMsg.Topic == "debug" && !logCol.config.TopicDebug) ||
						(logMsg.Topic == "status" && !logCol.config.TopicStatus) {
						continue
					}
					logMsgs[index] = logMsg
					logMsgs[index].MASID = logCol.masID
					index++
				}
				logMsgs = logMsgs[:index]
				_, err = logCol.client.PostLogs(logCol.masID, logMsgs)
				if err != nil {
					logCol.logError.Println(err)
					for i := range logMsgs {
						logCol.logIn <- logMsgs[i]
					}
					continue
				}
			}
			tempTime := time.Now()
			for {
				time.Sleep(100 * time.Millisecond)
				if time.Since(tempTime).Seconds() > 15 || len(logCol.logIn) > 50 {
					break
				}
			}
		}
	} else {
		for {
			// print messages to stdout if logger is turned off
			logMsg := <-logCol.logIn
			if (logMsg.Topic == "msg" && !logCol.config.TopicMsg) ||
				(logMsg.Topic == "app" && !logCol.config.TopicApp) ||
				(logMsg.Topic == "debug" && !logCol.config.TopicDebug) ||
				(logMsg.Topic == "status" && !logCol.config.TopicStatus) {
				continue
			}
			logCol.logInfo.Println(logMsg)
		}
	}
}

// storeState requests the logging service to store state
func (logCol *LogCollector) storeState() (err error) {
	if logCol.config.Active {
		for {
			var states []schemas.State
			state := <-logCol.stateIn
			states = append(states, state)
			for i := 0; i < 24; i++ {
				// maximum of 25 states
				select {
				case state = <-logCol.stateIn:
					states = append(states, state)
				default:
					break
				}
			}
			_, err = logCol.client.UpdateStates(states[0].MASID, states)
			if err != nil {
				logCol.logError.Println(err)
				for i := range states {
					logCol.stateIn <- states[i]
				}
				continue
			}
		}
	}
	return
}

// NewLogCollector creates an agency logger client
func NewLogCollector(masID int, config schemas.LoggerConfig, logErr *log.Logger,
	logInf *log.Logger) (logCol *LogCollector, err error) {
	logCol = &LogCollector{
		masID:    masID,
		logError: logErr,
		logInfo:  logInf,
		config:   config,
	}
	if logCol.config.Active {
		logCol.client, err = NewLoggerClient(config.Host, config.Port, time.Second*60,
			time.Second*1, 4)
		if err != nil {
			logCol.config.Active = false
		}
	}
	logCol.logIn = make(chan schemas.LogMessage, 10000)
	logCol.stateIn = make(chan schemas.State, 10000)
	go logCol.storeLogs()
	go logCol.storeState()
	logCol.logInfo.Println("Created new logger client; status: ", logCol.config.Active)
	return
}

// AgentLogger is the endpoint for agents
type AgentLogger struct {
	agentID  int
	masID    int
	client   *LoggerClient
	logOut   chan schemas.LogMessage // logging inbox
	stateOut chan schemas.State
	mutex    *sync.Mutex
	logError *log.Logger
	logInfo  *log.Logger
	active   bool
}

// NewLog sends a new logging message to the logging service
func (agLog *AgentLogger) NewLog(topic string, message string, data string) (err error) {
	agLog.mutex.Lock()
	if !agLog.active {
		agLog.mutex.Unlock()
		return errors.New("logger not active")
	}
	agLog.mutex.Unlock()
	if topic != "error" && topic != "debug" && topic != "status" && topic != "msg" &&
		topic != "app" {
		err = errors.New("unknown topic")
		return
	}
	agLog.mutex.Lock()
	time.Sleep(time.Millisecond * 5)
	tStamp := time.Now()
	agLog.mutex.Unlock()
	msg := schemas.LogMessage{
		AgentID:        agLog.agentID,
		Timestamp:      tStamp,
		Topic:          topic,
		Message:        message,
		AdditionalData: data}
	agLog.logOut <- msg
	return
}

// UpdateState overrides the state stored in database
func (agLog *AgentLogger) UpdateState(state string) (err error) {
	agLog.mutex.Lock()
	if !agLog.active {
		agLog.mutex.Unlock()
		return errors.New("agLog not active")
	}
	agLog.mutex.Unlock()
	agState := schemas.State{
		MASID:     agLog.masID,
		AgentID:   agLog.agentID,
		Timestamp: time.Now(),
		State:     state}
	agLog.stateOut <- agState
	return
}

// RestoreState loads state saved in database and return it
func (agLog *AgentLogger) RestoreState() (state string, err error) {
	agLog.mutex.Lock()
	if !agLog.active {
		agLog.mutex.Unlock()
		err = errors.New("agLog not active")
		return
	}
	agLog.mutex.Unlock()
	var agState schemas.State
	agState, _, err = agLog.client.GetState(agLog.masID, agLog.agentID)
	state = agState.State
	return
}

// NewAgentLogger craetes a new object of type AgentLogger
func (logCol *LogCollector) NewAgentLogger(agentID int, config schemas.LoggerConfig,
	logErr *log.Logger, logInf *log.Logger) (agLog *AgentLogger) {
	agLog = &AgentLogger{
		agentID:  agentID,
		masID:    logCol.masID,
		client:   logCol.client,
		logOut:   logCol.logIn,
		stateOut: logCol.stateIn,
		mutex:    &sync.Mutex{},
		logError: logErr,
		logInfo:  logInf,
		active:   true,
	}
	return
}

// close closes the logger
func (agLog *AgentLogger) Close() {
	agLog.mutex.Lock()
	agLog.logInfo.Println("Closing Logger of agent ", agLog.agentID)
	agLog.active = false
	agLog.mutex.Unlock()
}
