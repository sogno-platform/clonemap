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

// provides a struct containing fields of a FIPA ACL message and functions to manipulate
// messages

package agency

import (
	"errors"
	"log"
	"sync"
	"time"

	"github.com/RWTH-ACS/clonemap/pkg/client"
	"github.com/RWTH-ACS/clonemap/pkg/schemas"
)

// ACL provides functionality for agent messaging
type ACL struct {
	msgIn         chan schemas.ACLMessage         // ACL message inbox
	msgInProtocol map[int]chan schemas.ACLMessage // registered handlers for message protocol
	addrBook      map[int]*ACL                    // ACL address book of other agents
	mutex         *sync.Mutex                     // mutex for address book
	// commIn        chan int                        // ID of agents that have sent messages
	// commOut       chan int                        // ID of agents that messages have been sent to
	agentID   int
	active    bool
	aclLookup func(int) (*ACL, error)
	logger    *client.AgentLogger
	logError  *log.Logger
	logInfo   *log.Logger
}

// commData stores data about communication with other agent
// type commData struct {
// 	numMsgSent int // number of messages sent to this agent
// 	numMsgRecv int // number of messages received from this agent
// }

// newACL creates a new ACL object
func newACL(agentID int, msgIn chan schemas.ACLMessage,
	aclLookup func(int) (*ACL, error), cmaplog *client.AgentLogger,
	logErr *log.Logger, logInf *log.Logger) (acl *ACL) {
	acl = &ACL{
		mutex:         &sync.Mutex{},
		msgIn:         msgIn,
		msgInProtocol: make(map[int]chan schemas.ACLMessage),
		// commIn:        make(chan int, 5000),
		// commOut:       make(chan int, 5000),
		addrBook:  make(map[int]*ACL),
		agentID:   agentID,
		active:    true,
		aclLookup: aclLookup,
		logger:    cmaplog,
		logError:  logErr,
		logInfo:   logInf,
	}
	return
}

// // getCommDataChannels returns channels for comm data analysis
// func (acl *ACL) getCommDataChannels() (in chan int, out chan int) {
// 	in = acl.commIn
// 	out = acl.commOut
// 	return
// }

// close closes the acl
func (acl *ACL) close() {
	acl.mutex.Lock()
	acl.logInfo.Println("Closing ACL of agent ", acl.agentID)
	acl.active = false
	acl.mutex.Unlock()
}

// NewMessage returns a new initiaized message
func (acl *ACL) NewMessage(receiver int, prot int, perf int,
	content string) (msg schemas.ACLMessage, err error) {
	msg.Sender = acl.agentID
	msg.Receiver = receiver
	if perf < schemas.FIPAPerfNone || perf > schemas.FIPAPerfSubscribe {
		err = errors.New("non fipa-conform performative")
	}
	msg.Protocol = prot
	msg.Performative = perf
	msg.Content = content
	// msg.Timestamp = time.Now()
	return
}

// RecvMessages retrieves all messages since last call of this function
func (acl *ACL) RecvMessages() (num int, msgs []schemas.ACLMessage, err error) {
	acl.mutex.Lock()
	if !acl.active {
		acl.mutex.Unlock()
		err = errors.New("acl not active")
		return
	}
	acl.mutex.Unlock()
	num = 0
	err = nil
	for {
		select {
		case msgtemp := <-acl.msgIn:
			msgs = append(msgs, msgtemp)
			num++
		default:
			return
		}
	}
}

// RecvMessageWait retrieves next message and blocks if no message is available
func (acl *ACL) RecvMessageWait() (msg schemas.ACLMessage, err error) {
	acl.mutex.Lock()
	if !acl.active {
		acl.mutex.Unlock()
		err = errors.New("acl not active")
		return
	}
	acl.mutex.Unlock()
	err = nil
	msg = <-acl.msgIn
	return
}

// SendMessage sends a message
func (acl *ACL) SendMessage(msg schemas.ACLMessage) (err error) {
	var aclRecv *ACL
	var ok bool
	msg.Timestamp = time.Now()

	acl.mutex.Lock()
	if !acl.active {
		acl.mutex.Unlock()
		return errors.New("acl not active")
	}

	msg.Sender = acl.agentID
	aclRecv, ok = acl.addrBook[msg.Receiver]
	acl.mutex.Unlock()
	if ok {
		err = aclRecv.newIncomingMessage(msg)
		if err != nil {
			acl.mutex.Lock()
			delete(acl.addrBook, msg.Receiver)
			acl.mutex.Unlock()
			aclRecv, err = acl.aclLookup(msg.Receiver)
			if err != nil {
				return
			}
			acl.mutex.Lock()
			acl.addrBook[msg.Receiver] = aclRecv
			acl.mutex.Unlock()
			err = aclRecv.newIncomingMessage(msg)
		}
	} else {
		aclRecv, err = acl.aclLookup(msg.Receiver)
		if err != nil {
			return
		}
		acl.mutex.Lock()
		acl.addrBook[msg.Receiver] = aclRecv
		acl.mutex.Unlock()
		err = aclRecv.newIncomingMessage(msg)
	}
	if err != nil {
		return
	}
	err = acl.logger.NewLog("msg", "ACL send", msg.String())
	// acl.mutex.Lock()
	// if acl.analysis {
	// 	acl.commOut <- msg.Receiver
	// }
	// acl.mutex.Unlock()
	return
}

// newIncomingMessage adds message to channel for incoming messages
func (acl *ACL) newIncomingMessage(msg schemas.ACLMessage) (err error) {
	acl.mutex.Lock()
	if !acl.active {
		acl.mutex.Unlock()
		return errors.New("acl not active")
	}
	acl.mutex.Unlock()
	acl.logInfo.Println("New message for agent ", msg.Receiver)
	acl.mutex.Lock()
	inbox, ok := acl.msgInProtocol[msg.Protocol]
	acl.mutex.Unlock()
	if ok {
		inbox <- msg
	} else {
		acl.msgIn <- msg
	}
	err = acl.logger.NewLog("msg", "ACL receive", msg.String())
	// acl.mutex.Lock()
	// if acl.analysis {
	// 	acl.commIn <- msg.Sender
	// }
	// acl.mutex.Unlock()
	return
}

// registerProtocolChannel registers the protocol channel with the messaging service
func (acl *ACL) registerProtocolChannel(prot int,
	protChannel chan schemas.ACLMessage) (err error) {
	acl.mutex.Lock()
	if !acl.active {
		acl.mutex.Unlock()
		return errors.New("acl not active")
	}
	_, ok := acl.msgInProtocol[prot]
	acl.mutex.Unlock()
	if !ok {
		acl.msgInProtocol[prot] = protChannel
	} else {
		err = errors.New("protocol is already handled")
	}
	return
}

// deregisterProtocolChannel deregisters the protocol channel with the messaging service
func (acl *ACL) deregisterProtocolChannel(prot int) (err error) {
	acl.mutex.Lock()
	_, ok := acl.msgInProtocol[prot]
	acl.mutex.Unlock()
	if ok {
		delete(acl.msgInProtocol, prot)
	} else {
		err = errors.New("protocol is not handled")
	}
	return
}
