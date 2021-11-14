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
	"time"

	"github.com/RWTH-ACS/clonemap/pkg/schemas"
	mqtt "github.com/eclipse/paho.mqtt.golang"
)

// Behavior defines execution of a certain behavior
type Behavior interface {
	Start()
	Stop()
}

// aclProtocolBehavior describes how messages with a certain protocol should be handled
type aclProtocolBehavior struct {
	ag                 *Agent                                 // agent
	protocol           int                                    // indicates for which protocol handler should be used
	handlePerformative map[int]func(schemas.ACLMessage) error // handler functions for single performative acts
	handleDefault      func(schemas.ACLMessage) error         // default handler if no handler for performative is registered
	msgIn              chan schemas.ACLMessage                // msg inbox
	ctrl               chan int                               // control signals
	logInfo            *log.Logger
}

// NewMessageBehavior creates a new handler for messages of the specified protocol
func (agent *Agent) NewMessageBehavior(protocol int,
	handlePerformative map[int]func(schemas.ACLMessage) error,
	handleDefault func(schemas.ACLMessage) error) (behavior Behavior, err error) {
	if handleDefault == nil {
		err = errors.New("illegal default handler")
		return
	}
	protBehavior := &aclProtocolBehavior{
		ag:                 agent,
		protocol:           protocol,
		handlePerformative: handlePerformative,
		handleDefault:      handleDefault,
		msgIn:              make(chan schemas.ACLMessage, 1000),
		ctrl:               make(chan int, 10),
		logInfo:            agent.logInfo,
	}
	behavior = protBehavior
	return
}

// Start initiates the handling of messages
func (protBehavior *aclProtocolBehavior) Start() {
	// log the start
	protBehavior.ag.Logger.NewLog("beh", "protocol behavior starts", "")
	// register protocol handler
	protBehavior.ag.ACL.registerProtocolChannel(protBehavior.protocol, protBehavior.msgIn)
	// execute
	go protBehavior.task()
}

// task performs the multiplexing of messages with different performative to handler functions
func (protBehavior *aclProtocolBehavior) task() {
	protBehavior.logInfo.Println("Starting acl behavior for agent ", protBehavior.ag.GetAgentID(),
		" and protocol ", protBehavior.protocol)
	for {
		protBehavior.ag.mutex.Lock()
		act := protBehavior.ag.active
		protBehavior.ag.mutex.Unlock()
		if !act {
			protBehavior.Stop()
		}
		select {
		case msg := <-protBehavior.msgIn:
			if handle, ok := protBehavior.handlePerformative[msg.Performative]; ok {
				start := time.Now()
				handle(msg)
				end := time.Now()
				protBehavior.ag.Logger.NewBehStats(start, end, "protocol")
				protBehavior.ag.Logger.NewLog("beh", "Protocol behavior task", "start: "+start.String()+
					";end: "+end.String()+";duration:"+end.Sub(start).String())
			} else {
				start := time.Now()
				protBehavior.handleDefault(msg)
				end := time.Now()
				protBehavior.ag.Logger.NewBehStats(start, end, "protocol")
				protBehavior.ag.Logger.NewLog("beh", "Protocol behavior task", "start: "+start.String()+
					";end: "+end.String()+";duration:"+end.Sub(start).String())
			}
		case command := <-protBehavior.ctrl:
			switch command {
			case -1:
				protBehavior.logInfo.Println("Terminating acl behavior for agent ",
					protBehavior.ag.GetAgentID(), " and protocol ", protBehavior.protocol)
				return
			}
		}
	}
}

// Stop terminates the message handling
func (protBehavior *aclProtocolBehavior) Stop() {
	// log the stop
	protBehavior.ag.Logger.NewLog("beh", "protocol behavior stops", "")
	// deregister handler
	protBehavior.ag.ACL.deregisterProtocolChannel(protBehavior.protocol)
	// stop message handling
	protBehavior.ctrl <- -1
}

// mqttTopicBehavior describes how mqtt messages with a certain topic should be handled
type mqttTopicBehavior struct {
	ag      *Agent                   // agent
	topic   string                   // indicates for which protocol handler should be used
	handle  func(mqtt.Message) error // handler function
	msgIn   chan mqtt.Message        // msg inbox
	ctrl    chan int                 // control signals
	logInfo *log.Logger
}

// NewMQTTTopicBehavior creates a new handler for messages of the specified topic
func (agent *Agent) NewMQTTTopicBehavior(topic string,
	handle func(mqtt.Message) error) (behavior Behavior, err error) {
	if handle == nil {
		err = errors.New("illegal handler")
		return
	}
	mqttBehavior := &mqttTopicBehavior{
		ag:      agent,
		topic:   topic,
		handle:  handle,
		msgIn:   make(chan mqtt.Message, 100),
		ctrl:    make(chan int, 10),
		logInfo: agent.logInfo,
	}
	behavior = mqttBehavior
	return
}

// Start initiates the handling of messages
func (mqttBehavior *mqttTopicBehavior) Start() {
	// log the start
	mqttBehavior.ag.Logger.NewLog("beh", "mqtt topic behavior starts; Topic:"+mqttBehavior.topic, "")
	// register protocol handle
	mqttBehavior.ag.MQTT.registerTopicChannel(mqttBehavior.topic, mqttBehavior.msgIn)
	// execute
	go mqttBehavior.task()
}

// task performs the execution of the handle function
func (mqttBehavior *mqttTopicBehavior) task() {
	mqttBehavior.ag.Logger.NewLog("app", "executing mqtt task......", "")
	for {
		mqttBehavior.ag.mutex.Lock()
		act := mqttBehavior.ag.active
		mqttBehavior.ag.mutex.Unlock()
		if !act {
			mqttBehavior.Stop()
		}
		select {
		case msg := <-mqttBehavior.msgIn:
			start := time.Now()
			mqttBehavior.handle(msg)
			end := time.Now()
			mqttBehavior.ag.Logger.NewBehStats(start, end, "mqtt")
			mqttBehavior.ag.Logger.NewLog("beh", "mqtt topic behavior task", "start: "+start.String()+
				";end: "+end.String()+";duration:"+end.Sub(start).String()+";"+mqttMsgToString(msg))
		case command := <-mqttBehavior.ctrl:
			switch command {
			case -1:
				mqttBehavior.logInfo.Println("Terminating mqtt behavior for agent ",
					mqttBehavior.ag.GetAgentID())
				return
			}
		}
	}
}

// Stop terminates the message handling
func (mqttBehavior *mqttTopicBehavior) Stop() {
	// log the stop
	mqttBehavior.ag.Logger.NewLog("beh", "mqtt topic behavior stops", "")
	// deregister handler
	mqttBehavior.ag.MQTT.deregisterTopicChannel(mqttBehavior.topic)
	// stop message handling
	mqttBehavior.ctrl <- -1
}

// periodicBehavior describes an action that should be performed periodically
type periodicBehavior struct {
	ag      *Agent        // agent
	period  time.Duration // duration between two executions
	handle  func() error  // handler function
	ctrl    chan int      // control signals
	logInfo *log.Logger
}

// NewPeriodicBehavior creates a new handler for periodic actions
func (agent *Agent) NewPeriodicBehavior(period time.Duration,
	handle func() error) (behavior Behavior, err error) {
	if handle == nil {
		err = errors.New("illegal handler")
		return
	}
	periodBehavior := &periodicBehavior{
		ag:      agent,
		period:  period,
		handle:  handle,
		ctrl:    make(chan int, 10),
		logInfo: agent.logInfo,
	}
	behavior = periodBehavior
	return
}

// Start initiates the handling of messages
func (periodBehavior *periodicBehavior) Start() {
	// log the start
	periodBehavior.ag.Logger.NewLog("beh", "periodic behavior starts", "")
	// execute
	go periodBehavior.task()
}

// task performs the execution of the handle function
func (periodBehavior *periodicBehavior) task() {
	periodBehavior.logInfo.Println("Starting periodic behavior for agent ",
		periodBehavior.ag.GetAgentID(), " and period ", periodBehavior.period)
	for {
		periodBehavior.ag.mutex.Lock()
		act := periodBehavior.ag.active
		periodBehavior.ag.mutex.Unlock()
		if !act {
			periodBehavior.Stop()
		}
		time.Sleep(periodBehavior.period)
		select {
		case command := <-periodBehavior.ctrl:
			switch command {
			case -1:
				periodBehavior.logInfo.Println("Terminating periodic behavior for agent ",
					periodBehavior.ag.GetAgentID())
				return
			}
		default:
			start := time.Now()
			periodBehavior.handle()
			end := time.Now()
			periodBehavior.ag.Logger.NewBehStats(start, end, "period")
			periodBehavior.ag.Logger.NewLog("beh", "peroidic behavior task", "start: "+start.String()+
				";end: "+end.String()+";duration:"+end.Sub(start).String())
		}
	}
}

// Stop terminates the message handling
func (periodBehavior *periodicBehavior) Stop() {
	// log the stop
	periodBehavior.ag.Logger.NewLog("beh", "periodic behavior stops", "")
	// stop message handling
	periodBehavior.ctrl <- -1
}

// customUpdateBehavior describes an action that should be performed when the custom configuration
// is updated
type customUpdateBehavior struct {
	ag       *Agent                    // agent
	handle   func(custom string) error // handler function
	ctrl     chan int                  // control signals
	customIn chan string               // custom config inbox
	logInfo  *log.Logger
}

// NewCustomUpdateBehavior creates a new handler for custom config update actions
func (agent *Agent) NewCustomUpdateBehavior(
	handle func(custom string) error) (behavior Behavior, err error) {
	if handle == nil {
		err = errors.New("illegal handler")
		return
	}
	custUpBehavior := &customUpdateBehavior{
		ag:       agent,
		handle:   handle,
		ctrl:     make(chan int, 10),
		customIn: make(chan string, 10),
		logInfo:  agent.logInfo,
	}
	behavior = custUpBehavior
	return
}

// Start initiates the handling of messages
func (custUpBehavior *customUpdateBehavior) Start() {
	// log the start
	custUpBehavior.ag.Logger.NewLog("beh", "custom update behavior starts", "")
	custUpBehavior.ag.registerCustomUpdateChannel(custUpBehavior.customIn)
	// execute
	go custUpBehavior.task()
}

// task performs the execution of the handle function
func (custUpBehavior *customUpdateBehavior) task() {
	custUpBehavior.logInfo.Println("Starting custom configuration update behavior for agent ",
		custUpBehavior.ag.GetAgentID())
	for {
		custUpBehavior.ag.mutex.Lock()
		act := custUpBehavior.ag.active
		custUpBehavior.ag.mutex.Unlock()
		if !act {
			custUpBehavior.Stop()
		}
		select {
		case custom := <-custUpBehavior.customIn:
			start := time.Now()
			custUpBehavior.handle(custom)
			end := time.Now()
			custUpBehavior.ag.Logger.NewBehStats(start, end, "custom")
			custUpBehavior.ag.Logger.NewLog("beh", "custom behavior task", "start: "+start.String()+
				";end: "+end.String()+";duration:"+end.Sub(start).String())
		case command := <-custUpBehavior.ctrl:
			switch command {
			case -1:
				custUpBehavior.logInfo.Println("Terminating custom configuration update ",
					"behavior for agent ", custUpBehavior.ag.GetAgentID())
				return
			}
		}
	}
}

// Stop terminates the behavior
func (custUpBehavior *customUpdateBehavior) Stop() {
	// log the stop
	custUpBehavior.ag.Logger.NewLog("beh", "custom update behavior ends", "")
	custUpBehavior.ag.deregisterCustomUpdateChannel()
	// stop behavior
	custUpBehavior.ctrl <- -1
}
