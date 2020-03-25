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

// Package schemas defines the data structures used for the REST APIs
package schemas

import (
	"strconv"
	"time"
)

// CloneMAP contains information about clonemap
type CloneMAP struct {
	Version string    `json:"version,omitempty"` // version of clonemap
	Uptime  time.Time `json:"uptime,omitempty"`  // uptime of clonemap instance
}

// MASInfo contains info about MAS spec, agents and agencies in MAS
type MASInfo struct {
	Spec     MASSpec  `json:"spec"`
	ID       int      `json:"id"`
	Agents   Agents   `json:"agents"`
	Agencies Agencies `json:"agencies"`
	Status   Status   `json:"status"`
	Graph    Graph    `json:"graph"`
}

// MASSpec contains information about a MAS running in clonemap
type MASSpec struct {
	// ID                 int    `json:"id"`                        // unique ID of MAS
	Name               string `json:"name,omitempty"`            // name/description of MAS
	NumAgentsPerAgency int    `json:"agentsperagency,omitempty"` // number of agents per agency
	Logging            bool   `json:"logging"`                   // switch for logging module
	// Analysis           bool      `json:"analysis"`                  // switch for analysis logging
	MQTT   bool        `json:"mqtt"` //switch for mqtt
	DF     bool        `json:"df"`   //switch for df
	Logger LogConfig   `json:"log"`  // logger configuration
	Uptime time.Time   `json:"uptime"`
	Agents []AgentSpec `json:"agents"`
	Graph  Graph       `json:"graph"`
}

// AgentInfo contains information about agent spec, address, communication, mqtt and status
type AgentInfo struct {
	Spec     AgentSpec `json:"spec"`
	MASID    int       `json:"masid"`    // ID of MAS
	AgencyID int       `json:"agencyid"` // name of the agency
	ImageID  int       `json:"imid"`     // ID of agency image
	ID       int       `json:"id"`       // ID of agent
	Address  Address   `json:"address"`
	Status   Status    `json:"status"`
}

// AgentSpec contains information about a agent running in a MAS
type AgentSpec struct {
	// MASID           int    `json:"masid"`             // ID of MAS
	// AgencyID int `json:"agencyid"` // name of the agency
	NodeID int `json:"nodeid"` // id of the node the agent is attached to
	// ID              int    `json:"id"`                // unique ID of agent
	AgencyImage     string `json:"image"`             // docker image to be used for agencies
	ImagePullSecret string `json:"secret,omitempty"`  // image pull secret
	Name            string `json:"name,omitempty"`    // name/description of agent
	AType           string `json:"type,omitempty"`    // type of agent (application dependent)
	ASubtype        string `json:"subtype,omitempty"` // subtype of agent (application dependent)
	Custom          string `json:"custom,omitempty"`  // custom configuration data
}

// Address holds the address information of an agent
type Address struct {
	Agency string `json:"agency"`
}

// Status contains information about an agent's or agency's status
type Status struct {
	Code       int       `json:"code"`       // status code
	LastUpdate time.Time `json:"lastupdate"` // time of last update
}

// AgencyInfo contains information about agency spec and status
type AgencyInfo struct {
	Spec   AgencySpec `json:"spec"`
	Status Status     `json:"status"`
}

// AgencySpec contains information about agency
type AgencySpec struct {
	MASID  int         `json:"masid"` // ID of MAS
	Name   string      `json:"name"`  // name of agency (hostname of pod given by kubernetes)
	ID     int         `json:"id"`    // unique ID (contained in name)
	Logger LogConfig   `json:"log"`   // logger configuration
	Agents []AgentInfo `json:"agents"`
}

// MASs contains informaton about how many MASs are running
type MASs struct {
	Counter   int       `json:"counter"`   // number of running mas
	Instances []MASSpec `json:"instances"` // mas ids
}

// Agents contains information about how many agents are running
type Agents struct {
	Counter   int         `json:"counter"`   // counter for agents
	Instances []AgentInfo `json:"instances"` // agent ids
}

// Agencies contains information about how many agencies are running
type Agencies struct {
	Counter   int          `json:"counter"`   // counter for agents
	Instances []AgencyInfo `json:"instances"` // agencies
}

// AgentStatus contains status of agency
type AgentStatus struct {
	ID     int    `json:"id"`     // unique ID
	Status Status `json:"status"` // statuscode
}

// AgencyStatus contains status of agent
type AgencyStatus struct {
	Status Status        `json:"status"` // statuscode
	Agents []AgentStatus `json:"agents"` // status of all agents in agency
}

// StubAgencyConfig holds configuration of agency to be started or terminated
type StubAgencyConfig struct {
	MASID     int    `json:"masid"`
	AgencyID  int    `json:"agencyid"`
	NumAgents int    `json:"numagents"`
	Image     string `json:"image"`
	Logging   bool   `json:"logging"` // switch for logging module
	MQTT      bool   `json:"mqtt"`    //switch for mqtt
	DF        bool   `json:"df"`      //switch for df
}

// ACLMessage struct representing agent message
type ACLMessage struct {
	Timestamp      time.Time `json:"ts"`                // sending time
	Performative   int       `json:"perf"`              // Denotes the type of the communicative act of the ACL message
	Sender         int       `json:"sender"`            // Denotes the identity of the sender of the message
	AgencySender   string    `json:"agencys"`           // denotes the name of the sender agency
	Receiver       int       `json:"receiver"`          // Denotes the identity of the intended recipients of the message
	AgencyReceiver string    `json:"agencyr"`           // denotes the name of the receiver agency
	ReplyTo        int       `json:"repto,omitempty"`   // This parameter indicates that subsequent messages in this conversation thread are to be directed to the agent named in the reply-to parameter, instead of to the agent named in the sender parameter
	Content        string    `json:"content"`           // Denotes the content of the message
	Language       string    `json:"lang,omitempty"`    // Denotes the language in which the content parameter is expressed
	Encoding       string    `json:"enc,omitempty"`     // Denotes the specific encoding of the content language expression
	Ontology       string    `json:"ont,omitempty"`     // Denotes the ontology(s) used to give a meaning to the symbols in the content expression
	Protocol       int       `json:"prot"`              // Denotes the interaction protocol that the sending agent is employing with this ACL message
	ConversationID int       `json:"convid,omitempty"`  // Introduces an expression which is used to identify the ongoing sequence of communicative acts that together form a conversation
	ReplyWith      string    `json:"repwith,omitempty"` // Introduces an expression that will be used by the responding agent to identify this message
	InReplyTo      int       `json:"inrepto,omitempty"` // Denotes an expression that references an earlier action to which this message is a reply
	ReplyBy        time.Time `json:"repby,omitempty"`   // Denotes a time and/or date expression which indicates the latest time by which the sending agent would like to receive a reply
}

// String outputs message
func (msg ACLMessage) String() (ret string) {
	ret = "Sender: " + strconv.Itoa(msg.Sender) + "; Receiver: " + strconv.Itoa(msg.Receiver) +
		"; Timestamp: " + msg.Timestamp.String() + "; "
	switch msg.Protocol {
	case FIPAProtNone:
		ret += "Protocol: None; "
	case FIPAProtRequest:
		ret += "Protocol: Request; "
	case FIPAProtQuery:
		ret += "Protocol: Query; "
	case FIPAProtRequestWhen:
		ret += "Protocol: Request When; "
	case FIPAProtContractNet:
		ret += "Protocol: Contract-Net; "
	case FIPAProtIteratedContractNet:
		ret += "Protocol: Iterated Contract-Net; "
	case FIPAProtEnglishAuction:
		ret += "Protocol: English Auction; "
	case FIPAProtDutchAuction:
		ret += "Protocol: Dutch Auction; "
	case FIPAProtBrokering:
		ret += "Protocol: Brokering; "
	case FIPAProtRecruiting:
		ret += "Protocol: Recruiting; "
	case FIPAProtSubscribe:
		ret += "Protocol: Subscribe; "
	case FIPAProtPropose:
		ret += "Protocol: Propose; "
	default:
		ret += "Protocol: Unknown(" + strconv.Itoa(msg.Protocol) + "); "
	}
	switch msg.Performative {
	case FIPAPerfNone:
		ret += "Performative: None; "
	case FIPAPerfAcceptProposal:
		ret += "Performative: Accept Proposal; "
	case FIPAPerfAgree:
		ret += "Performative: Agree; "
	case FIPAPerfCancel:
		ret += "Performative: Cancel; "
	case FIPAPerfCallForProposal:
		ret += "Performative: Call For Proposal; "
	case FIPAPerfConfirm:
		ret += "Performative: Confirm; "
	case FIPAPerfDisconfirm:
		ret += "Performative: Disconfirm; "
	case FIPAPerfFailure:
		ret += "Performative: Failure; "
	case FIPAPerfInform:
		ret += "Performative: Inform; "
	case FIPAPerfInformIf:
		ret += "Performative: Inform If; "
	case FIPAPerfInformRef:
		ret += "Performative: Inform Ref; "
	case FIPAPerfNotUnderstood:
		ret += "Performative: Not Understood; "
	case FIPAPerfPropagate:
		ret += "Performative: Propagate; "
	case FIPAPerfPropose:
		ret += "Performative: Propose; "
	case FIPAPerfProxy:
		ret += "Performative: Proxy; "
	case FIPAPerfQueryIf:
		ret += "Performative: Query If; "
	case FIPAPerfQueryRef:
		ret += "Performative: Query Ref; "
	case FIPAPerfRefuse:
		ret += "Performative: Refuse; "
	case FIPAPerfRejectProposal:
		ret += "Performative: Reject Proposal; "
	case FIPAPerfRequest:
		ret += "Performative: Request; "
	case FIPAPerfRequestWhen:
		ret += "Performative: Request When; "
	case FIPAPerfRequestWhenever:
		ret += "Performative: Request Whenever; "
	case FIPAPerfSubscribe:
		ret += "Performative: Subscribe; "
	default:
		ret += "Performative: Unknown(" + strconv.Itoa(msg.Performative) + "); "
	}
	ret += "Content: " + msg.Content
	return
}

// FIPA performatives
const (
	// initialization performative
	FIPAPerfNone = iota
	// The action of accepting a previously submitted proposal to perform an action
	FIPAPerfAcceptProposal = iota
	// The action of agreeing to perform some action, possibly in the future
	FIPAPerfAgree = iota
	// The action of one agent informing another agent that the first agent no longer has the
	// intention that the second agent perform some action
	FIPAPerfCancel = iota
	// The action of calling for proposals to perform a given action
	FIPAPerfCallForProposal = iota
	// The sender informs the receiver that a given proposition is true, where the receiver is
	// known to be uncertain about the proposition
	FIPAPerfConfirm = iota
	// The sender informs the receiver that a given proposition is false, where the receiver is
	// known to believe, or believe it likely that, the proposition is true
	FIPAPerfDisconfirm = iota
	// The action of telling another agent that an action was attempted but the attempt failed
	FIPAPerfFailure = iota
	// The sender informs the receiver that a given proposition is true
	FIPAPerfInform = iota
	// A macro action for the agent of the action to inform the recipient whether or not a
	// proposition is true
	FIPAPerfInformIf = iota
	// A macro action for sender to inform the receiver the object which corresponds to a
	// descriptor, for example, a name
	FIPAPerfInformRef = iota
	// The sender of the act (for example, i) informs the receiver (for example, j) that it
	// perceived that j performed some action, but that i did not understand what j just did.
	// A particular core case is that i tells j that i did not understand the message that j has
	// just sent to i
	FIPAPerfNotUnderstood = iota
	// The sender intends that the receiver treat the embedded message as sent directly to the
	// receiver, and wants the receiver to identify the agents denoted by the given descriptor and
	// send the received propagate message to them
	FIPAPerfPropagate = iota
	// The action of submitting a proposal to perform a certain action, given certain preconditions
	FIPAPerfPropose = iota
	// The  sender wants the receiver to select target agents denoted by a given description and
	// to send an embedded message to them
	FIPAPerfProxy = iota
	// The action of asking another agent whether or not a given proposition is true
	FIPAPerfQueryIf = iota
	// The action of asking another agent for the object referred to by a referential expression
	FIPAPerfQueryRef = iota
	// The action of refusing to perform a given action, and explaining the reason for the refusal
	FIPAPerfRefuse = iota
	// The action of rejecting a proposal to perform some action during a negotiation
	FIPAPerfRejectProposal = iota
	// The sender requests the receiver to perform some action. One important class of uses of the
	// request act is to request the receiver to perform another communicative act
	FIPAPerfRequest = iota
	// The sender wants the receiver to perform some action when some given proposition becomes true
	FIPAPerfRequestWhen = iota
	// The sender wants the receiver to perform some action as soon as some proposition becomes
	// true and thereafter each time the proposition becomes true again
	FIPAPerfRequestWhenever = iota
	// The act of requesting a persistent intention to notify the sender of the value of a
	// reference, and to notify again whenever the object identified by the reference changes
	FIPAPerfSubscribe = iota
)

// FIPA protocols
const (
	// initialization protocol
	FIPAProtNone                = iota
	FIPAProtRequest             = iota
	FIPAProtQuery               = iota
	FIPAProtRequestWhen         = iota
	FIPAProtContractNet         = iota
	FIPAProtIteratedContractNet = iota
	FIPAProtEnglishAuction      = iota
	FIPAProtDutchAuction        = iota
	FIPAProtBrokering           = iota
	FIPAProtRecruiting          = iota
	FIPAProtSubscribe           = iota
	FIPAProtPropose             = iota
)
