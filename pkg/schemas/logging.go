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

package schemas

import (
	"time"
)

// LoggerConfig contains configuration of logger service
type LoggerConfig struct {
	Active      bool   `json:"active"`           // indicates if logger is active/usable
	TopicMsg    bool   `json:"msg,omitempty"`    // activation of msg log topic
	TopicApp    bool   `json:"app,omitempty"`    // activation of app log topic
	TopicStatus bool   `json:"status,omitempty"` // activation of status log topic
	TopicDebug  bool   `json:"debug,omitempty"`  // activation of debug log topic
	TopicBeh    bool   `json:"beh,omitempty"`    // activation of beh log topic
	Host        string `json:"host,omitempty"`   // hostname of Logger
	Port        int    `json:"port,omitempty"`   // port of Logger
}

// LogMessage contains data of a single agent log message
type LogMessage struct {
	MASID          int       `json:"masid"`          // ID of MAS agent runs in
	AgentID        int       `json:"agentid"`        // ID of agent
	Timestamp      time.Time `json:"timestamp"`      // time of message
	Topic          string    `json:"topic"`          // log type (error, debug, msg, status, app, beh)
	Message        string    `json:"msg"`            // log message
	AdditionalData string    `json:"data,omitempty"` // additional information e.g in json
}

// LogSeries contains series data of a single agent
type LogSeries struct {
	MASID     int       `json:"masid"`   // ID of MAS agent runs in
	AgentID   int       `json:"agentid"` // ID of agent
	Name      string    `json:"name"`
	Timestamp time.Time `json:"timestamp"` // time of the logSeries
	Value     float64   `json:"value"`     // value of the series item
}

// State contains the state of an agent as byte array (json)
type State struct {
	MASID     int       `json:"masid"`     // ID of MAS agent runs in
	AgentID   int       `json:"agentid"`   // ID of agent
	Timestamp time.Time `json:"timestamp"` // time of state
	State     string    `json:"state"`     // State
}

// Communication contains information regarding communication with another agent
type Communication struct {
	ID         int `json:"id"`      // id of other agent
	NumMsgSent int `json:"numsent"` // number of messages sent to this agent
	NumMsgRecv int `json:"numrecv"` // number of messages received from this agent
}

type Heatmap struct {
	Sender   int `json:"sender"`
	Receiver int `json:"receiver"`
}
