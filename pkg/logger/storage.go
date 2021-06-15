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

package logger

import (
	"errors"
	"math"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"git.rwth-aachen.de/acs/public/cloud/mas/clonemap/pkg/schemas"
)

// storage is the interface for logging and state storage (either local or in db)
type storage interface {
	// addAgentLogMessage adds an entry to specified logging entry
	addAgentLogMessage(log schemas.LogMessage) (err error)

	// addAgentLogSeries add the log series
	addAgentLogSeries(series schemas.LogSeries)

	// addAgentBehStats add the behavior stats
	addAgentBehStats(logBeh schemas.BehStats)

	// getLatestAgentLogMessages return the latest num log messages
	getLatestAgentLogMessages(masID int, agentID int, topic string,
		num int) (logs []schemas.LogMessage, err error)

	// getAgentLogMessagesInRange return the log messages in the specified time range
	getAgentLogMessagesInRange(masID int, agentID int, topic string, start time.Time,
		end time.Time) (logs []schemas.LogMessage, err error)

	// getAgentLogSeriesNames gets the name of the log series
	getAgentLogSeriesNames(masID int, agentID int) (names []string, err error)

	// getAgentLogSeries get the log series
	getAgentLogSeries(masID int, agentID int, name string, start time.Time, end time.Time) (series []schemas.LogSeries, err error)

	// getMsgHeatmap get the msg communication frequency
	getMsgHeatmap(masID int) (heatmap map[[2]int]int, err error)

	// getStats get the data of a certain behtype
	getStats(masID int, agentID int, behType string, start time.Time, end time.Time) (statsInfo schemas.StatsInfo, err error)

	// deleteAgentLogMessages deletes all log messages og an agent
	deleteAgentLogMessages(masID int, agentID int) (err error)

	// updateCommunication updates communication data
	updateCommunication(masID int, agentID int, commData []schemas.Communication) (err error)

	// getCommunication returns communication data
	getCommunication(masID int, agentID int) (commData []schemas.Communication, err error)

	// updateAgentState updates the agent status
	updateAgentState(masID int, agentID int, state schemas.State) (err error)

	// getAgentState return the latest agent status
	getAgentState(masID int, agentID int) (state schemas.State, err error)

	// deleteAgentState deletes the status of an agent
	deleteAgentState(masID int, agentID int) (err error)
}

// LogMessage contains the content of a single log message
// type LogMessage struct {
// 	Timestamp time.Time
// 	Message   string
// 	Data      string
// }

// IDs of agents and mas correspond to index in slices!
type localStorage struct {
	mas   []masStorage
	mutex *sync.Mutex
}

type masStorage struct {
	agents []agentStorage
	msgCnt map[[2]int]int
}

type agentStorage struct {
	errLogs   []schemas.LogMessage
	dbgLogs   []schemas.LogMessage
	msgLogs   []schemas.LogMessage
	statLogs  []schemas.LogMessage
	appLogs   []schemas.LogMessage
	behLogs   []schemas.LogMessage
	behsStats map[string][]schemas.BehStats
	logSeries map[string][]schemas.LogSeries
	state     schemas.State
	commData  []schemas.Communication
}

// addAgentLogMessage adds an entry to specified logging entry
func (stor *localStorage) addAgentLogMessage(log schemas.LogMessage) (err error) {
	stor.mutex.Lock()
	numMAS := len(stor.mas)
	if numMAS <= log.MASID {
		for i := 0; i < log.MASID-numMAS+1; i++ {
			stor.mas = append(stor.mas, masStorage{})
		}
	}
	numAgents := len(stor.mas[log.MASID].agents)
	if numAgents <= log.AgentID {
		for i := 0; i < log.AgentID-numAgents+1; i++ {
			stor.mas[log.MASID].agents = append(stor.mas[log.MASID].agents, agentStorage{})
		}
	}
	switch log.Topic {
	case "error":
		stor.mas[log.MASID].agents[log.AgentID].errLogs = append(stor.mas[log.MASID].agents[log.AgentID].errLogs,
			log)
	case "debug":
		stor.mas[log.MASID].agents[log.AgentID].dbgLogs = append(stor.mas[log.MASID].agents[log.AgentID].dbgLogs,
			log)
	case "msg":
		stor.mas[log.MASID].agents[log.AgentID].msgLogs = append(stor.mas[log.MASID].agents[log.AgentID].msgLogs,
			log)
		if log.Message == "ACL send" {
			if stor.mas[log.MASID].msgCnt == nil {
				stor.mas[log.MASID].msgCnt = make(map[[2]int]int)
			}
			recvStr := strings.Split(log.AdditionalData, ";")[1]
			rec, _ := strconv.Atoi(strings.Split(recvStr, " ")[1])
			idx := [2]int{log.AgentID, rec}
			stor.mas[log.MASID].msgCnt[idx] += 1
		}
	case "status":
		stor.mas[log.MASID].agents[log.AgentID].statLogs = append(stor.mas[log.MASID].agents[log.AgentID].statLogs,
			log)
	case "app":
		stor.mas[log.MASID].agents[log.AgentID].appLogs = append(stor.mas[log.MASID].agents[log.AgentID].appLogs,
			log)
	case "beh":
		stor.mas[log.MASID].agents[log.AgentID].behLogs = append(stor.mas[log.MASID].agents[log.AgentID].behLogs,
			log)
	default:
		err = errors.New("wrong topic")
	}
	stor.mutex.Unlock()
	return
}

// getLatestAgentLogMessages return the latest num log messages
func (stor *localStorage) getLatestAgentLogMessages(masID int, agentID int, topic string,
	num int) (logs []schemas.LogMessage, err error) {
	stor.mutex.Lock()
	if masID < len(stor.mas) {
		if agentID < len(stor.mas[masID].agents) {
			switch topic {
			case "error":
				length := len(stor.mas[masID].agents[agentID].errLogs)
				if length < num {
					num = length
				}
				logs = make([]schemas.LogMessage, num)
				copy(logs, stor.mas[masID].agents[agentID].errLogs[length-num:length])
			case "debug":
				length := len(stor.mas[masID].agents[agentID].dbgLogs)
				if length < num {
					num = length
				}
				logs = make([]schemas.LogMessage, num)
				copy(logs, stor.mas[masID].agents[agentID].dbgLogs[length-num:length])
			case "msg":
				length := len(stor.mas[masID].agents[agentID].msgLogs)
				if length < num {
					num = length
				}
				logs = make([]schemas.LogMessage, num)
				copy(logs, stor.mas[masID].agents[agentID].msgLogs[length-num:length])
			case "status":
				length := len(stor.mas[masID].agents[agentID].statLogs)
				if length < num {
					num = length
				}
				logs = make([]schemas.LogMessage, num)
				copy(logs, stor.mas[masID].agents[agentID].statLogs[length-num:length])
			case "app":
				length := len(stor.mas[masID].agents[agentID].appLogs)
				if length < num {
					num = length
				}
				logs = make([]schemas.LogMessage, num)
				copy(logs, stor.mas[masID].agents[agentID].appLogs[length-num:length])
			case "beh":
				length := len(stor.mas[masID].agents[agentID].behLogs)
				if length < num {
					num = length
				}
				logs = make([]schemas.LogMessage, num)
				copy(logs, stor.mas[masID].agents[agentID].behLogs[length-num:length])
			default:
				err = errors.New("wrong topic")
			}
		}
	}
	stor.mutex.Unlock()
	return
}

// getAgentLogMessagesInRange return the log messages in the specified time range
func (stor *localStorage) getAgentLogMessagesInRange(masID int, agentID int, topic string,
	start time.Time, end time.Time) (logs []schemas.LogMessage, err error) {
	stor.mutex.Lock()
	if masID < len(stor.mas) {
		if agentID < len(stor.mas[masID].agents) {
			switch topic {
			case "error":
				length := len(stor.mas[masID].agents[agentID].errLogs)
				if length > 0 {
					startIndex := sort.Search(length,
						func(i int) bool {
							return stor.mas[masID].agents[agentID].errLogs[i].Timestamp.After(start)
						})
					endIndex := sort.Search(length,
						func(i int) bool {
							return stor.mas[masID].agents[agentID].errLogs[i].Timestamp.After(end)
						})
					if endIndex-startIndex >= 0 {
						logs = make([]schemas.LogMessage, endIndex-startIndex)
						copy(logs, stor.mas[masID].agents[agentID].errLogs[startIndex:endIndex])
					}
				}
			case "debug":
				length := len(stor.mas[masID].agents[agentID].dbgLogs)
				if length > 0 {
					startIndex := sort.Search(length,
						func(i int) bool {
							return stor.mas[masID].agents[agentID].dbgLogs[i].Timestamp.After(start)
						})
					endIndex := sort.Search(length,
						func(i int) bool {
							return stor.mas[masID].agents[agentID].dbgLogs[i].Timestamp.After(end)
						})
					if endIndex-startIndex >= 0 {
						logs = make([]schemas.LogMessage, endIndex-startIndex)
						copy(logs, stor.mas[masID].agents[agentID].dbgLogs[startIndex:endIndex])
					}
				}
			case "msg":
				length := len(stor.mas[masID].agents[agentID].msgLogs)
				if length > 0 {
					startIndex := sort.Search(length,
						func(i int) bool {
							return stor.mas[masID].agents[agentID].msgLogs[i].Timestamp.After(start)
						})
					endIndex := sort.Search(length,
						func(i int) bool {
							return stor.mas[masID].agents[agentID].msgLogs[i].Timestamp.After(end)
						})
					if endIndex-startIndex >= 0 {
						logs = make([]schemas.LogMessage, endIndex-startIndex)
						copy(logs, stor.mas[masID].agents[agentID].msgLogs[startIndex:endIndex])
					}
				}
			case "status":
				length := len(stor.mas[masID].agents[agentID].statLogs)
				if length > 0 {
					startIndex := sort.Search(length,
						func(i int) bool {
							return stor.mas[masID].agents[agentID].statLogs[i].Timestamp.
								After(start)
						})
					endIndex := sort.Search(length,
						func(i int) bool {
							return stor.mas[masID].agents[agentID].statLogs[i].Timestamp.After(end)
						})
					if endIndex-startIndex >= 0 {
						logs = make([]schemas.LogMessage, endIndex-startIndex)
						copy(logs, stor.mas[masID].agents[agentID].statLogs[startIndex:endIndex])
					}
				}
			case "app":
				length := len(stor.mas[masID].agents[agentID].appLogs)
				if length > 0 {
					startIndex := sort.Search(length,
						func(i int) bool {
							return stor.mas[masID].agents[agentID].appLogs[i].Timestamp.After(start)
						})
					endIndex := sort.Search(length,
						func(i int) bool {
							return stor.mas[masID].agents[agentID].appLogs[i].Timestamp.After(end)
						})
					if endIndex-startIndex >= 0 {
						logs = make([]schemas.LogMessage, endIndex-startIndex)
						copy(logs, stor.mas[masID].agents[agentID].appLogs[startIndex:endIndex])
					}
				}
			case "beh":
				length := len(stor.mas[masID].agents[agentID].behLogs)
				if length > 0 {
					startIndex := sort.Search(length,
						func(i int) bool {
							return stor.mas[masID].agents[agentID].behLogs[i].Timestamp.After(start)
						})
					endIndex := sort.Search(length,
						func(i int) bool {
							return stor.mas[masID].agents[agentID].behLogs[i].Timestamp.After(end)
						})
					if endIndex-startIndex >= 0 {
						logs = make([]schemas.LogMessage, endIndex-startIndex)
						copy(logs, stor.mas[masID].agents[agentID].behLogs[startIndex:endIndex])
					}
				}
			default:
				err = errors.New("wrong topic")
			}
		}
	}
	stor.mutex.Unlock()
	return
}

// addAgentLogSeries add the log series
func (stor *localStorage) addAgentLogSeries(series schemas.LogSeries) {
	stor.mutex.Lock()
	numMAS := len(stor.mas)
	if numMAS <= series.MASID {
		for i := 0; i < series.MASID-numMAS+1; i++ {
			stor.mas = append(stor.mas, masStorage{})
		}
	}
	numAgents := len(stor.mas[series.MASID].agents)
	if numAgents <= series.AgentID {
		for i := 0; i < series.AgentID-numAgents+1; i++ {
			stor.mas[series.MASID].agents = append(stor.mas[series.MASID].agents, agentStorage{})
		}
	}

	if stor.mas[series.MASID].agents[series.AgentID].logSeries == nil {
		stor.mas[series.MASID].agents[series.AgentID].logSeries = make(map[string][]schemas.LogSeries)
	}
	stor.mas[series.MASID].agents[series.AgentID].logSeries[series.Name] = append(stor.mas[series.MASID].agents[series.AgentID].logSeries[series.Name], series)
	stor.mutex.Unlock()
	return
}

// addAgentBehStats add the behavior stats
func (stor *localStorage) addAgentBehStats(behStats schemas.BehStats) {
	stor.mutex.Lock()
	numMAS := len(stor.mas)
	if numMAS <= behStats.MASID {
		for i := 0; i < behStats.MASID-numMAS+1; i++ {
			stor.mas = append(stor.mas, masStorage{})
		}
	}
	numAgents := len(stor.mas[behStats.MASID].agents)
	if numAgents <= behStats.AgentID {
		for i := 0; i < behStats.AgentID-numAgents+1; i++ {
			stor.mas[behStats.MASID].agents = append(stor.mas[behStats.MASID].agents, agentStorage{})
		}
	}

	if stor.mas[behStats.MASID].agents[behStats.AgentID].behsStats == nil {
		stor.mas[behStats.MASID].agents[behStats.AgentID].behsStats = make(map[string][]schemas.BehStats)
	}
	stor.mas[behStats.MASID].agents[behStats.AgentID].behsStats[behStats.BehType] = append(stor.mas[behStats.MASID].agents[behStats.AgentID].behsStats[behStats.BehType], behStats)
	stor.mutex.Unlock()
	return
}

// getAgentLogSeriesNames gets the name of the log series
func (stor *localStorage) getAgentLogSeriesNames(masID int, agentID int) (names []string, err error) {
	stor.mutex.Lock()
	if masID < len(stor.mas) {
		if agentID < len(stor.mas[masID].agents) {
			if stor.mas[masID].agents[agentID].logSeries != nil {
				names = make([]string, 0, len(stor.mas[masID].agents[agentID].logSeries))
				for k := range stor.mas[masID].agents[agentID].logSeries {
					names = append(names, k)
				}
			}
		}
	}
	stor.mutex.Unlock()
	return
}

// getAgentLogSeries return the log series
func (stor *localStorage) getAgentLogSeries(masID int, agentID int, name string, start time.Time, end time.Time) (series []schemas.LogSeries, err error) {
	stor.mutex.Lock()
	if masID < len(stor.mas) {
		if agentID < len(stor.mas[masID].agents) {
			length := len(stor.mas[masID].agents[agentID].logSeries[name])
			if length > 0 {
				startIndex := sort.Search(length,
					func(i int) bool {
						return stor.mas[masID].agents[agentID].logSeries[name][i].Timestamp.After(start)
					})
				endIndex := sort.Search(length,
					func(i int) bool {
						return stor.mas[masID].agents[agentID].logSeries[name][i].Timestamp.After(end)
					})
				if endIndex-startIndex >= 0 {
					series = make([]schemas.LogSeries, endIndex-startIndex)
					copy(series, stor.mas[masID].agents[agentID].logSeries[name][startIndex:endIndex])
				}
			}
		}
	}
	stor.mutex.Unlock()
	return
}

// getMsgHeatmap return the msg communication frequency
func (stor *localStorage) getMsgHeatmap(masID int) (heatmap map[[2]int]int, err error) {
	stor.mutex.Lock()
	if masID < len(stor.mas) {
		heatmap = stor.mas[masID].msgCnt
	}
	stor.mutex.Unlock()
	return
}

func getDuration(behsStats []schemas.BehStats) (duration []int) {
	for _, behStats := range behsStats {
		duration = append(duration, behStats.Duration)
	}
	return
}

func getMax(duration []int) (maxVal int) {
	maxVal = 0
	for _, v := range duration {
		if v > maxVal {
			maxVal = v
		}
	}
	return
}

func getMin(duration []int) (minVal int) {
	minVal = math.MaxInt32
	for _, v := range duration {
		if v < minVal {
			minVal = v
		}
	}
	return
}

func getAverage(duration []int) (averageVal float32) {
	sum := 0
	for _, v := range duration {
		sum += v
	}
	return float32(sum) / float32(len(duration))
}

// getStats get the data of a certain behavior type
func (stor *localStorage) getStats(masID int, agentID int, behType string, start time.Time, end time.Time) (statsInfo schemas.StatsInfo, err error) {
	stor.mutex.Lock()
	if masID < len(stor.mas) {
		if agentID < len(stor.mas[masID].agents) {
			res, ok := stor.mas[masID].agents[agentID].behsStats[behType]
			if ok {
				length := len(res)
				startIndex := sort.Search(length,
					func(i int) bool {
						return res[i].Start.After(start)
					})
				endIndex := sort.Search(length,
					func(i int) bool {
						return res[i].Start.After(end)
					})
				if endIndex-startIndex >= 0 {
					logs := make([]schemas.BehStats, endIndex-startIndex)
					copy(logs, stor.mas[masID].agents[agentID].behsStats[behType][startIndex:endIndex])
					statsInfo.List = logs
					statsInfo.Max = getMax(getDuration(logs))
					statsInfo.Min = getMin(getDuration(logs))
					statsInfo.Count = len(logs)
					statsInfo.Average = getAverage((getDuration(logs)))

				}
			}
		}
	}
	stor.mutex.Unlock()
	return
}

// deleteAgentLogMessages deletes all log messages og an agent
func (stor *localStorage) deleteAgentLogMessages(masID int, agentID int) (err error) {
	stor.mutex.Lock()
	if masID < len(stor.mas) {
		if agentID < len(stor.mas[masID].agents) {
			stor.mas[masID].agents[agentID].errLogs = nil
			stor.mas[masID].agents[agentID].dbgLogs = nil
			stor.mas[masID].agents[agentID].msgLogs = nil
			stor.mas[masID].agents[agentID].statLogs = nil
			stor.mas[masID].agents[agentID].appLogs = nil
			stor.mas[masID].agents[agentID].behLogs = nil
		}
	}
	stor.mutex.Unlock()
	return
}

// updateCommunication updates communication data
func (stor *localStorage) updateCommunication(masID int, agentID int,
	commData []schemas.Communication) (err error) {
	stor.mutex.Lock()
	numMAS := len(stor.mas)
	if numMAS <= masID {
		for i := 0; i < masID-numMAS+1; i++ {
			stor.mas = append(stor.mas, masStorage{})
		}
	}
	numAgents := len(stor.mas[masID].agents)
	if numAgents <= agentID {
		for i := 0; i < agentID-numAgents+1; i++ {
			stor.mas[i].agents = append(stor.mas[i].agents, agentStorage{})
		}
	}
	stor.mas[masID].agents[agentID].commData = commData
	stor.mutex.Unlock()
	return
}

// getCommunication returns communication data
func (stor *localStorage) getCommunication(masID int,
	agentID int) (commData []schemas.Communication, err error) {
	stor.mutex.Lock()
	if masID < len(stor.mas) {
		if agentID < len(stor.mas[masID].agents) {
			commData = stor.mas[masID].agents[agentID].commData
		}
	}
	stor.mutex.Unlock()
	return
}

// updateAgentState updates the agent status
func (stor *localStorage) updateAgentState(masID int, agentID int, state schemas.State) (err error) {
	stor.mutex.Lock()
	numMAS := len(stor.mas)
	if numMAS <= masID {
		for i := 0; i < masID-numMAS+1; i++ {
			stor.mas = append(stor.mas, masStorage{})
		}
	}
	numAgents := len(stor.mas[masID].agents)
	if numAgents <= agentID {
		for i := 0; i < agentID-numAgents+1; i++ {
			stor.mas[i].agents = append(stor.mas[i].agents, agentStorage{})
		}
	}
	stor.mas[masID].agents[agentID].state = state
	stor.mutex.Unlock()
	return
}

// getAgentState return the latest agent status
func (stor *localStorage) getAgentState(masID int, agentID int) (state schemas.State, err error) {
	stor.mutex.Lock()
	if masID < len(stor.mas) {
		if agentID < len(stor.mas[masID].agents) {
			state = stor.mas[masID].agents[agentID].state
		}
	}
	stor.mutex.Unlock()
	return
}

// deleteAgentState deletes the status of an agent
func (stor *localStorage) deleteAgentState(masID int, agentID int) (err error) {
	stor.mutex.Lock()
	if masID < len(stor.mas) {
		if agentID < len(stor.mas[masID].agents) {
			stor.mas[masID].agents[agentID].state = schemas.State{}
		}
	}
	stor.mutex.Unlock()
	return
}

// newLocalStorage returns Storage interface with localStorage type
func newLocalStorage() storage {
	var temp localStorage
	temp.mutex = &sync.Mutex{}
	return &temp
}
