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
	"log"
	"net"
	"net/http"
	"strconv"

	agencyclient "git.rwth-aachen.de/acs/public/cloud/mas/clonemap/pkg/agency/client"
	"git.rwth-aachen.de/acs/public/cloud/mas/clonemap/pkg/schemas"
)

// remoteAgency holds the channel used for sending messages to remot agency
type remoteAgency struct {
	msgIn        chan schemas.ACLMessage // ACL message inbox
	agencyClient *agencyclient.Client
	// agents map[int]*agent.Agent
}

// aclLookup provides the correct acl object for an agent
func (agency *Agency) aclLookup(agentID int) (acl *ACL, err error) {
	var ag *Agent
	var ok bool
	agency.mutex.Lock()
	ag, ok = agency.localAgents[agentID]
	agency.mutex.Unlock()
	// check if agent is local agent
	if ok {
		acl = ag.ACL
		return
	}
	agency.mutex.Lock()
	ag, ok = agency.remoteAgents[agentID]
	agency.mutex.Unlock()
	// check if remote agent is already known
	if ok {
		acl = ag.ACL
		return
	}

	// request address of unknown agent
	var address schemas.Address
	address, err = agency.requestAgentAddress(agentID)
	if err != nil {
		return
	}
	agency.logInfo.Println("Request address of unknown agent")
	agentInfo := schemas.AgentInfo{}
	agency.mutex.Lock()
	agencyName := agency.info.Name
	agency.mutex.Unlock()
	if address.Agency == agencyName {
		err = errors.New("MassiveError")
		return
	}
	if address.Agency == "" {
		err = errors.New("receiver is not active")
		return
	}
	var remAgency *remoteAgency
	agency.mutex.Lock()
	remAgency, ok = agency.remoteAgencies[address.Agency]
	agency.mutex.Unlock()
	// check if remote agency is already known
	if ok {
		agency.logInfo.Println("New remote agent ", agentID, " in known agency ", address.Agency)
		ag = newAgent(agentInfo, remAgency.msgIn, nil, nil, schemas.LoggerConfig{}, nil, nil,
			agency.logError, agency.logInfo)
	} else {
		agency.logInfo.Println("New remote agent ", agentID, " in unknown agency ", address.Agency)
		// create new remote agency
		remAgency = &remoteAgency{
			msgIn:        make(chan schemas.ACLMessage, 1000),
			agencyClient: agency.agencyClient,
		}
		agency.mutex.Lock()
		agency.remoteAgencies[address.Agency] = remAgency
		numRemAgencies := len(agency.remoteAgencies)
		numLocalAgs := len(agency.localAgents)
		agency.mutex.Unlock()
		// start go routine for sending to new remote agency
		go remAgency.sendMsgs(address.Agency, agencyName, agency.logError)
		agency.logInfo.Println("Started go routine for sending to agency ", address.Agency)
		// start additional go routine for incoming messages
		if numRemAgencies > 1 &&
			numRemAgencies < numLocalAgs {
			go agency.receiveMsgs()
		}
		ag = newAgent(agentInfo, remAgency.msgIn, nil, nil, schemas.LoggerConfig{}, nil, nil,
			agency.logError, agency.logInfo)
	}
	agency.mutex.Lock()
	agency.remoteAgents[agentID] = ag
	agency.mutex.Unlock()

	acl = ag.ACL
	return
}

// requestAgentAddress requests the address of an agent from ams
func (agency *Agency) requestAgentAddress(agentID int) (address schemas.Address, err error) {
	agency.mutex.Lock()
	masID := agency.info.MASID
	agency.mutex.Unlock()
	address, _, err = agency.amsClient.GetAgentAddress(masID, agentID)
	return
}

// sendMsgs is to be executed as go routine. It sends msgs to remote agency
func (remAgency *remoteAgency) sendMsgs(remName string, localName string, logErr *log.Logger) {
	var err error
	var ip string
	var stat int
	ip, err = getIP(remName)
	if err != nil {
		logErr.Println(err)
		return
	}
	for {
		msg := <-remAgency.msgIn
		num := len(remAgency.msgIn)
		if num > 99 {
			num = 99
		}
		msgs := make([]schemas.ACLMessage, num+1)
		msgs[0] = msg
		msgs[0].AgencySender = localName
		msgs[0].AgencyReceiver = remName
		for i := 0; i < num; i++ {
			msgs[i+1] = <-remAgency.msgIn
			msgs[i+1].AgencySender = localName
			msgs[i+1].AgencyReceiver = remName
		}
		stat, err = remAgency.agencyClient.PostMsgs(ip, msgs)
		if err != nil || stat != http.StatusCreated {
			ip, err = getIP(remName)
			if err != nil {
				logErr.Println(err)
				return
			}
			stat, err = remAgency.agencyClient.PostMsgs(ip, msgs)
			if err != nil {
				logErr.Println(err)
			}
			if stat != http.StatusCreated {
				logErr.Println("Wrong http code: " + strconv.Itoa(stat))
			}
		}
		// fmt.Println(time.Now().String() + " sent " + strconv.Itoa(len(msgs)) + " messages to agency " + msgs[0].AgencyReceiver)
	}
}

// getIP requests IT from DNS
func getIP(dnsName string) (ip string, err error) {
	for i := 0; i < 5; i++ {
		var ips []string
		ips, err = net.LookupHost(dnsName)
		if len(ips) > 0 && err == nil {
			ip = ips[0]
			break
		}
	}
	return
}

// receiveMsgs is to be executed as go routine. It delivers incoming messages to local agents
func (agency *Agency) receiveMsgs() {
	agency.logInfo.Println("Started go routine for message receiving")
	var msgs []schemas.ACLMessage
	for {
		msgs = <-agency.msgIn
		for i := range msgs {
			var ok bool
			var ag *Agent
			agency.mutex.Lock()
			ag, ok = agency.localAgents[msgs[i].Receiver]
			agency.mutex.Unlock()
			if ok {
				err := ag.ACL.newIncomingMessage(msgs[i])
				if err != nil {
					_, err = agency.agencyClient.ReturnMsg(msgs[i].AgencySender, msgs[i])
					if err != nil {
						agency.logError.Println(err)
						return
					}
				}
			} else {
				_, err := agency.agencyClient.ReturnMsg(msgs[i].AgencySender, msgs[i])
				if err != nil {
					agency.logError.Println(err)
					return
				}
			}
		}
	}
}

// resendUndeliverableMsg resends a messages that has been returned as undeliverable by a
// remote agency
func (agency *Agency) resendUndeliverableMsg(msg schemas.ACLMessage) (err error) {
	var remAg, locAg *Agent
	var ok bool
	agency.mutex.Lock()
	remAg, ok = agency.remoteAgents[msg.Receiver]
	agency.mutex.Unlock()
	if ok {
		agency.mutex.Lock()
		locAg, ok = agency.localAgents[msg.Sender]
		agency.mutex.Unlock()
		if ok {
			remAg.ACL.close()
			agency.mutex.Lock()
			delete(agency.remoteAgents, msg.Receiver)
			agency.mutex.Unlock()
			locAg.ACL.SendMessage(msg)
		}
	}
	return
}
