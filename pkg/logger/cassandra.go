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
	"fmt"
	"time"

	"github.com/RWTH-ACS/clonemap/pkg/schemas"
	"github.com/gocql/gocql"
)

// logEnvelope contains log and agentid masid
// type logEnvelope struct {
// 	masID   int
// 	agentID int
// 	log     schemas.LogMessage
// }

// cassStorage stores information regarding connection to cassandra cluster
type cassStorage struct {
	cluster     *gocql.ClusterConfig
	session     *gocql.Session
	logStatusIn chan schemas.LogMessage // logging inbox
	logAppIn    chan schemas.LogMessage // logging inbox
	logErrorIn  chan schemas.LogMessage // logging inbox
	logDebugIn  chan schemas.LogMessage // logging inbox
	logMsgIn    chan schemas.LogMessage // logging inbox
	stateIn     chan schemas.State      // state inbox
}

// addAgentLogMessage adds an entry to specified logging entry
func (stor *cassStorage) addAgentLogMessage(log schemas.LogMessage) (err error) {

	// var js []byte
	// js, err = json.Marshal(log)
	switch log.Topic {
	case "error":
		stor.logErrorIn <- log
		// err = stor.session.Query("INSERT INTO logging_error (masid, agentid, t, log) "+
		// 	"VALUES (?, ?, ?, ?)", masID, agentID, log.Timestamp, js).Exec()
	case "debug":
		stor.logDebugIn <- log
		// err = stor.session.Query("INSERT INTO logging_debug (masid, agentid, t, log) "+
		// 	"VALUES (?, ?, ?, ?)", masID, agentID, log.Timestamp, js).Exec()
	case "msg":
		stor.logMsgIn <- log
		// err = stor.session.Query("INSERT INTO logging_msg (masid, agentid, t, log) "+
		// 	"VALUES (?, ?, ?, ?)", masID, agentID, log.Timestamp, js).Exec()
	case "status":
		stor.logStatusIn <- log
		// err = stor.session.Query("INSERT INTO logging_status (masid, agentid, t, log) "+
		// 	"VALUES (?, ?, ?, ?)", masID, agentID, log.Timestamp, js).Exec()
	case "app":
		stor.logAppIn <- log
		// err = stor.session.Query("INSERT INTO logging_app (masid, agentid, t, log) "+
		// 	"VALUES (?, ?, ?, ?)", masID, agentID, log.Timestamp, js).Exec()
	default:
		err = errors.New("wrong topic")
	}
	return
}

// getLatestAgentLogMessages return the latest num log messages
func (stor *cassStorage) getLatestAgentLogMessages(masID int, agentID int, topic string,
	num int) (logs []schemas.LogMessage, err error) {
	var iter *gocql.Iter
	switch topic {
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
		err = errors.New("wrong topic")
	}
	if err != nil {
		return
	}
	var js []byte
	for iter.Scan(&js) {
		var logmsg schemas.LogMessage
		err = json.Unmarshal(js, &logmsg)
		if err != nil {
			return
		}
		logs = append(logs, logmsg)
	}
	return
}

// getAgentLogMessagesInRange return the log messages in the specified time range
func (stor *cassStorage) getAgentLogMessagesInRange(masID int, agentID int, topic string,
	start time.Time, end time.Time) (logs []schemas.LogMessage, err error) {
	var iter *gocql.Iter
	switch topic {
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
		err = errors.New("wrong topic")
	}
	if err != nil {
		return
	}
	var js []byte
	for iter.Scan(&js) {
		var logmsg schemas.LogMessage
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
	iter := stor.session.Query("SELECT comm FROM communication WHERE masid = ? AND agentid = ?",
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
	stor.stateIn <- state
	// err = stor.session.Query("INSERT INTO state (masid, agentid, state) "+
	// 	"VALUES (?, ?, ?)", masID, agentID, state).Exec()
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

// storeLogs stores the logs in a batch operation
func (stor *cassStorage) storeLogs(topic string) {
	var logIn chan schemas.LogMessage
	var err error
	stmt := "INSERT INTO logging_" + topic + " (masid, agentid, t, log) VALUES (?, ?, ?, ?)"
	if topic == "status" {
		logIn = stor.logStatusIn
	} else if topic == "app" {
		logIn = stor.logAppIn
	} else if topic == "error" {
		logIn = stor.logErrorIn
	} else if topic == "debug" {
		logIn = stor.logDebugIn
	} else if topic == "msg" {
		logIn = stor.logMsgIn
	} else {
		return
	}

	for {
		batch := gocql.NewBatch(gocql.UnloggedBatch)
		log := <-logIn
		var js []byte
		js, err = json.Marshal(log)
		if err != nil {
			fmt.Println(err)
		}
		batch.Query(stmt, log.MASID, log.AgentID, log.Timestamp, js)
		size := len(js)
		for i := 0; i < 9; i++ {
			// maximum of 10 operations in batch
			if size > 25000 {
				break
			}
			empty := false
			select {
			case log = <-logIn:
				js, err = json.Marshal(log)
				if err != nil {
					fmt.Println(err)
				}
				batch.Query(stmt, log.MASID, log.AgentID, log.Timestamp, js)
				size += len(js)
			default:
				empty = true
			}
			if empty {
				break
			}
		}
		err = stor.session.ExecuteBatch(batch)
		if err != nil {
			fmt.Println(err)
		}
	}
}

// storeState stores the state in a batch operation
func (stor *cassStorage) storeState() {
	var err error
	stmt := "INSERT INTO state (masid, agentid, state) VALUES (?, ?, ?)"

	for {
		batch := gocql.NewBatch(gocql.UnloggedBatch)
		state := <-stor.stateIn
		var js []byte
		js, err = json.Marshal(state)
		if err != nil {
			fmt.Println(err)
		}
		batch.Query(stmt, state.MASID, state.AgentID, js)
		size := len(js)
		for i := 0; i < 9; i++ {
			// maximum of 10 operations in batch
			if size > 25000 {
				break
			}
			select {
			case state = <-stor.stateIn:
				js, err = json.Marshal(state)
				if err != nil {
					fmt.Println(err)
				}
				batch.Query(stmt, state.MASID, state.AgentID, js)
				size += len(js)
			default:
				break
			}
		}
		err = stor.session.ExecuteBatch(batch)
		if err != nil {
			fmt.Println(err)
		}
	}
}

func (stor *cassStorage) disconnect() {
	stor.session.Close()
}

// newCassandraStorage returns Storage interface with cassStorage type
func newCassandraStorage(ip []string, user string, pass string) (stor storage, err error) {
	var temp cassStorage
	temp.cluster = gocql.NewCluster(ip...)
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
	temp.logStatusIn = make(chan schemas.LogMessage, 10000)
	temp.logAppIn = make(chan schemas.LogMessage, 10000)
	temp.logDebugIn = make(chan schemas.LogMessage, 10000)
	temp.logErrorIn = make(chan schemas.LogMessage, 10000)
	temp.logMsgIn = make(chan schemas.LogMessage, 10000)
	temp.stateIn = make(chan schemas.State, 10000)

	for i := 0; i < 3; i++ {
		go temp.storeLogs("status")
		go temp.storeLogs("app")
		go temp.storeLogs("error")
		go temp.storeLogs("debug")
		go temp.storeLogs("msg")
		go temp.storeState()
	}
	stor = &temp
	return
}
