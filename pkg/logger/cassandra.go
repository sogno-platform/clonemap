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
	"encoding/json"
	"errors"
	"git.rwth-aachen.de/acs/public/cloud/mas/clonemap/pkg/schemas"
	"github.com/gocql/gocql"
	"time"
)

// cassStorage stores information regarding connection to cassandra cluster
type cassStorage struct {
	cluster *gocql.ClusterConfig
	session *gocql.Session
}

// addAgentLogMessage adds an entry to specified logging entry
func (stor *cassStorage) addAgentLogMessage(masID int, agentID int, logType string,
	log LogMessage) (err error) {
	var js []byte
	js, err = json.Marshal(log)
	switch logType {
	case "error":
		err = stor.session.Query("INSERT INTO logging_error (masid, agentid, t, log) "+
			"VALUES (?, ?, ?, ?)", masID, agentID, log.Timestamp, js).Exec()
	case "debug":
		err = stor.session.Query("INSERT INTO logging_debug (masid, agentid, t, log) "+
			"VALUES (?, ?, ?, ?)", masID, agentID, log.Timestamp, js).Exec()
	case "msg":
		err = stor.session.Query("INSERT INTO logging_msg (masid, agentid, t, log) "+
			"VALUES (?, ?, ?, ?)", masID, agentID, log.Timestamp, js).Exec()
	case "status":
		err = stor.session.Query("INSERT INTO logging_status (masid, agentid, t, log) "+
			"VALUES (?, ?, ?, ?)", masID, agentID, log.Timestamp, js).Exec()
	case "app":
		err = stor.session.Query("INSERT INTO logging_app (masid, agentid, t, log) "+
			"VALUES (?, ?, ?, ?)", masID, agentID, log.Timestamp, js).Exec()
	default:
		err = errors.New("WrongLogType")
	}
	return
}

// getLatestAgentLogMessages return the latest num log messages
func (stor *cassStorage) getLatestAgentLogMessages(masID int, agentID int, logtype string,
	num int) (logs []LogMessage, err error) {
	var iter *gocql.Iter
	switch logtype {
	case "error":
		iter = stor.session.Query("SELECT log FROM logging_error WHERE masid = ? AND "+
			"agentid = ? LIMIT ?", masID, agentID, num).Iter()
	case "debug":
		iter = stor.session.Query("SELECT log FROM logging_debug WHERE masid = ? AND "+
			"agentid = ? LIMIT ?", masID, agentID, num).Iter()
	case "msg":
		iter = stor.session.Query("SELECT log FROM logging_msg WHERE masid = ? AND "+
			"agentid = ? LIMIT ?", masID, agentID, num).Iter()
	case "status":
		iter = stor.session.Query("SELECT log FROM logging_status WHERE masid = ? AND "+
			"agentid = ? LIMIT ?", masID, agentID, num).Iter()
	case "app":
		iter = stor.session.Query("SELECT log FROM logging_app WHERE masid = ? AND "+
			"agentid = ? LIMIT ?", masID, agentID, num).Iter()
	default:
		err = errors.New("WrongLogType")
	}
	if err != nil {
		return
	}
	var js []byte
	for iter.Scan(&js) {
		var logmsg LogMessage
		err = json.Unmarshal(js, &logmsg)
		if err != nil {
			return
		}
		logs = append(logs, logmsg)
	}
	return
}

// getAgentLogMessagesInRange return the log messages in the specified time range
func (stor *cassStorage) getAgentLogMessagesInRange(masID int, agentID int, logtype string,
	start time.Time, end time.Time) (logs []LogMessage, err error) {
	var iter *gocql.Iter
	switch logtype {
	case "error":
		iter = stor.session.Query("SELECT log FROM logging_error WHERE masid = ? AND "+
			"agentid = ? AND t > ? AND t < ?", masID, agentID, start, end).Iter()
	case "debug":
		iter = stor.session.Query("SELECT log FROM logging_debug WHERE masid = ? AND "+
			"agentid = ? AND t > ? AND t < ?", masID, agentID, start, end).Iter()
	case "msg":
		iter = stor.session.Query("SELECT log FROM logging_msg WHERE masid = ? AND "+
			"agentid = ? AND t > ? AND t < ?", masID, agentID, start, end).Iter()
	case "status":
		iter = stor.session.Query("SELECT log FROM logging_status WHERE masid = ? AND "+
			"agentid = ? AND t > ? AND t < ?", masID, agentID, start, end).Iter()
	case "app":
		iter = stor.session.Query("SELECT log FROM logging_app WHERE masid = ? AND "+
			"agentid = ? AND t > ? AND t < ?", masID, agentID, start, end).Iter()
	default:
		err = errors.New("WrongLogType")
	}
	if err != nil {
		return
	}
	var js []byte
	for iter.Scan(&js) {
		var logmsg LogMessage
		err = json.Unmarshal(js, &logmsg)
		if err != nil {
			return
		}
		logs = append(logs, logmsg)
	}
	iter.Close()
	return
}

// deleteAgentLogMessages deletes all log messages og an agent
func (stor *cassStorage) deleteAgentLogMessages(masID int, agentID int) (err error) {

	return
}

// updateCommunication updates communication data
func (stor *cassStorage) updateCommunication(masID int, agentID int,
	commData []schemas.Communication) (err error) {
	var js []byte
	js, err = json.Marshal(commData)
	if err != nil {
		return
	}
	err = stor.session.Query("INSERT INTO communication (masid, agentid, comm) "+
		"VALUES (?, ?, ?)", masID, agentID, js).Exec()
	return
}

// getCommunication returns communication data
func (stor *cassStorage) getCommunication(masID int,
	agentID int) (commData []schemas.Communication, err error) {
	var iter *gocql.Iter
	iter = stor.session.Query("SELECT comm FROM communication WHERE masid = ? AND agentid = ?",
		masID, agentID).Iter()
	if iter.NumRows() == 1 {
		var js []byte
		iter.Scan(&js)
		err = json.Unmarshal(js, &commData)
	}
	iter.Close()
	return
}

// updateAgentState updates the agent status
func (stor *cassStorage) updateAgentState(masID int, agentID int, state schemas.State) (err error) {
	var js []byte
	js, err = json.Marshal(state)
	if err != nil {
		return
	}
	err = stor.session.Query("INSERT INTO state (masid, agentid, state) "+
		"VALUES (?, ?, ?)", masID, agentID, js).Exec()
	return
}

// getAgentState return the latest agent status
func (stor *cassStorage) getAgentState(masID int, agentID int) (state schemas.State, err error) {
	var iter *gocql.Iter
	var js []byte
	iter = stor.session.Query("SELECT state FROM state WHERE masid = ? AND agentid = ?", masID,
		agentID).Iter()
	if iter.NumRows() == 1 {
		iter.Scan(&js)
		err = json.Unmarshal(js, &state)
	}
	iter.Close()
	return
}

// deleteAgentState deletes the status of an agent
func (stor *cassStorage) deleteAgentState(masID int, agentID int) (err error) {

	return
}

func (stor *cassStorage) disconnect() {
	stor.session.Close()
}

// newCassandraStorage returns Storage interface with cassStorage type
func newCassandraStorage(ip string, user string, pass string) (stor storage, err error) {
	var temp cassStorage
	temp.cluster = gocql.NewCluster(ip)
	temp.cluster.Authenticator = gocql.PasswordAuthenticator{
		Username: user,
		Password: pass,
	}
	temp.cluster.Keyspace = "clonemap"
	temp.cluster.Timeout = 10 * time.Second
	temp.cluster.ProtoVersion = 4
	temp.session, err = temp.cluster.CreateSession()
	if err != nil {
		return
	}
	stor = &temp
	return
}
