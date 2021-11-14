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

// Package logger implements the frontend for the logging module
package logger

import (
	"errors"
	"io/ioutil"
	"log"
	"os"
	"time"

	"github.com/RWTH-ACS/clonemap/pkg/schemas"
)

// Logger stores information regarding logging
type Logger struct {
	stor     storage
	logInfo  *log.Logger // logger for info logging
	logError *log.Logger // logger for error logging
}

// StartLogger starts the logger app
func StartLogger() {
	log := &Logger{logError: log.New(os.Stderr, "[ERROR] ", log.LstdFlags)}
	err := log.init()
	if err != nil {
		log.logError.Println(err)
		return
	}
	serv := log.server(11000)
	if err != nil {
		log.logError.Println(err)
		return
	}
	err = log.listen(serv)
	if err != nil {
		log.logError.Println(err)
	}
}

func (logger *Logger) init() (err error) {
	logType := os.Getenv("CLONEMAP_LOG_LEVEL")
	switch logType {
	case "info":
		logger.logInfo = log.New(os.Stdout, "[INFO] ", log.LstdFlags)
	case "error":
		logger.logInfo = log.New(ioutil.Discard, "", log.LstdFlags)
	default:
		err = errors.New("wrong log type: " + logType)
		return
	}
	logger.logInfo.Println("Starting Logger")

	//fmt.Println("Getting deployment type")
	deplType := os.Getenv("CLONEMAP_DEPLOYMENT_TYPE")
	switch deplType {
	case "local":
		logger.logInfo.Println("Local storage")
		logger.stor = newLocalStorage()
	case "minikube":
		logger.logInfo.Println("Local storage")
		logger.stor = newLocalStorage()
	case "production":
		logger.logInfo.Println("Cassandra storage")
		logger.stor, err = newCassandraStorage([]string{"cass-ssset-0.cassandra", "cass-ssset-1.cassandra", "cass-ssset-2.cassandra"}, "cassandra", "cassandra")
	default:
		err = errors.New("Wrong deployment type: " + deplType)
	}
	return
}

// addAgentLogMessage adds an entry to specified logging entry
func (logger *Logger) addAgentLogMessage(logmsg schemas.LogMessage) (err error) {
	err = logger.stor.addAgentLogMessage(logmsg)
	return
}

// addAgentLogMessageList
func (logger *Logger) addAgentLogMessageList(logmsg []schemas.LogMessage) (err error) {
	for i := 0; i < len(logmsg); i++ {
		err = logger.addAgentLogMessage(logmsg[i])
		if err != nil {
			return
		}
	}
	return
}

// addAgentLogSeries add log series
func (logger *Logger) addAgentLogSeries(logseries []schemas.LogSeries) {
	for i := 0; i < len(logseries); i++ {
		logger.stor.addAgentLogSeries(logseries[i])
	}
}

// addAgentBehsStats add behavior log
func (logger *Logger) addAgentBehsStats(behsStats []schemas.BehStats) {
	for i := 0; i < len(behsStats); i++ {
		logger.stor.addAgentBehStats(behsStats[i])
	}
}

// getLatestAgentLogMessages return the latest num log messages
func (logger *Logger) getLatestAgentLogMessages(masID int, agentID int, topic string,
	num int) (logs []schemas.LogMessage, err error) {
	logs, err = logger.stor.getLatestAgentLogMessages(masID, agentID, topic, num)
	return
}

// getAgentLogMessagesInRange return the log messages in the specified time range
func (logger *Logger) getAgentLogMessagesInRange(masID int, agentID int, topic string,
	start time.Time, end time.Time) (logs []schemas.LogMessage, err error) {
	logs, err = logger.stor.getAgentLogMessagesInRange(masID, agentID, topic, start, end)
	return
}

// getLogSeriesByName return all the log series of a specific name
func (logger *Logger) getAgentLogSeries(masID int, agentID int, name string, start time.Time, end time.Time) (series []schemas.LogSeries, err error) {
	series, err = logger.stor.getAgentLogSeries(masID, agentID, name, start, end)
	return
}

// getAgentLogSeriesNames return log series names
func (logger *Logger) getAgentLogSeriesNames(masID int, agentID int) (names []string, err error) {
	names, err = logger.stor.getAgentLogSeriesNames(masID, agentID)
	return
}

// getMsgHeatmap return the frequency of message communication

func (logger *Logger) getMsgHeatmap(masID int, start time.Time, end time.Time) (heatmap map[[2]int]int, err error) {
	heatmap, err = logger.stor.getMsgHeatmap(masID, start, end)
	return
}

// getStats get the statistical data of a certain behType
func (logger *Logger) getStats(masID int, agentID int, behType string, start time.Time, end time.Time) (statsInfo schemas.StatsInfo, err error) {
	statsInfo, err = logger.stor.getStats(masID, agentID, behType, start, end)
	return
}

// updateCommunication updates communication data of agent
func (logger *Logger) updateCommunication(masID int, agentID int,
	comm []schemas.Communication) (err error) {
	err = logger.stor.updateCommunication(masID, agentID, comm)
	return
}

// getCommunication returns communication data of agent
func (logger *Logger) getCommunication(masID int, agentID int) (comm []schemas.Communication,
	err error) {
	return
}

// getAgentState returns the latest agent state
func (logger *Logger) getAgentState(masID int, agentID int) (agState schemas.State, err error) {
	agState, err = logger.stor.getAgentState(masID, agentID)
	return
}

// updateAgentState updates agent state
func (logger *Logger) updateAgentState(masID int, agentID int, agState schemas.State) (err error) {
	err = logger.stor.updateAgentState(masID, agentID, agState)
	return
}

// addAgentLogMessageList
func (logger *Logger) updateAgentStatesList(masID int, states []schemas.State) (err error) {
	for i := 0; i < len(states); i++ {
		err = logger.updateAgentState(masID, states[i].AgentID, states[i])
		if err != nil {
			return
		}
	}
	return
}
