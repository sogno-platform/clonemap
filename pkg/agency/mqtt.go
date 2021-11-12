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
	"strconv"
	"sync"

	"github.com/RWTH-ACS/clonemap/pkg/client"
	"github.com/RWTH-ACS/clonemap/pkg/schemas"
	mqtt "github.com/eclipse/paho.mqtt.golang"
)

// mqttCollector is the agency client for mqtt
type mqttCollector struct {
	client       mqtt.Client              // mqtt client
	msgIn        chan schemas.MQTTMessage // mqtt message inbox
	name         string                   // agency name
	config       schemas.MQTTConfig       // indicates if mqtt is active (switch via env)
	mutex        *sync.Mutex              // mutex for message inbox map
	subscription map[string][]*AgentMQTT  // map for subscription topics to agents' mqtt object
	// numDeliverer int                      // number of go routines for delivery
	logError *log.Logger
	logInfo  *log.Logger
}

// newMQTTCollector creates a new mqtt agency client
func newMQTTCollector(config schemas.MQTTConfig, name string, logErr *log.Logger,
	logInf *log.Logger) (col *mqttCollector) {
	col = &mqttCollector{
		name:     name,
		mutex:    &sync.Mutex{},
		logError: logErr,
		logInfo:  logInf,
		config:   config,
	}
	col.msgIn = make(chan schemas.MQTTMessage, 1000)
	col.subscription = make(map[string][]*AgentMQTT)
	col.logInfo.Println("Created new MQTT client; status: ", col.config.Active)
	col.init()
	return
}

func (mqttCol *mqttCollector) init() (err error) {
	if mqttCol.config.Active {
		opts := mqtt.NewClientOptions().AddBroker("tcp://" + mqttCol.config.Host + ":" +
			strconv.Itoa(mqttCol.config.Port)).SetClientID(mqttCol.name)
		opts.SetDefaultPublishHandler(mqttCol.newIncomingMQTTMessage)
		mqttCol.client = mqtt.NewClient(opts)
		if token := mqttCol.client.Connect(); token.Wait() && token.Error() != nil {
			err = errors.New("MQTTInitError")
			return
		}
		for i := 0; i < 3; i++ {
			go mqttCol.deliverMsgs()
		}
		// cli.numDeliverer = 3
	}
	return
}

func (mqttCol *mqttCollector) close() (err error) {
	if mqttCol.config.Active {
		mqttCol.logInfo.Println("Disconnecting MQTT client")
		mqttCol.client.Disconnect(250)
	}
	err = nil
	return
}

// newIncomingMQTTMessage adds message to channel for incoming messages
func (mqttCol *mqttCollector) newIncomingMQTTMessage(client mqtt.Client, msg mqtt.Message) {
	var mqttMsg schemas.MQTTMessage
	mqttMsg.Content = msg.Payload()
	mqttMsg.Topic = msg.Topic()
	mqttCol.msgIn <- mqttMsg
}

// subscribe subscribes to specified topics
func (mqttCol *mqttCollector) subscribe(mq *AgentMQTT, topic string, qos int) (err error) {
	if !mqttCol.config.Active {
		return
	}
	mqttCol.mutex.Lock()
	ag, ok := mqttCol.subscription[topic]
	mqttCol.mutex.Unlock()
	if ok {
		subscribed := false
		for i := range ag {
			if ag[i].agentID == mq.agentID {
				subscribed = true
				break
			}
		}
		if !subscribed {
			ag = append(ag, mq)
			mqttCol.mutex.Lock()
			mqttCol.subscription[topic] = ag
			mqttCol.mutex.Unlock()
		}
	} else {
		mqttCol.mutex.Lock()
		token := mqttCol.client.Subscribe(topic, byte(qos), nil)
		mqttCol.mutex.Unlock()
		if token.Wait() && token.Error() != nil {
			err = token.Error()
			return
		}
		ag = make([]*AgentMQTT, 0)
		ag = append(ag, mq)
		mqttCol.mutex.Lock()
		mqttCol.subscription[topic] = ag
		mqttCol.mutex.Unlock()
	}
	// cli.mutex.Lock()
	// numDel := len(cli.subscription) / 25
	// if numDel > cli.numDeliverer {
	// 	for i := 0; i < numDel-cli.numDeliverer; i++ {
	// 		go cli.deliverMsgs()
	// 	}
	// 	cli.numDeliverer = numDel
	// }
	// cli.mutex.Unlock()

	return
}

// unsubscribe to a topic
func (mqttCol *mqttCollector) unsubscribe(mq *AgentMQTT, topic string) (err error) {
	if !mqttCol.config.Active {
		return
	}
	mqttCol.mutex.Lock()
	ag, ok := mqttCol.subscription[topic]
	mqttCol.mutex.Unlock()
	if !ok {
		return
	}
	index := -1
	for i := range ag {
		if mq.agentID == ag[i].agentID {
			index = i
			break
		}
	}
	if index == -1 {
		return
	}
	if index == 0 && len(ag) == 1 {
		// agent is the only one who has subscribed -> unsubscribe
		delete(mqttCol.subscription, topic)
		token := mqttCol.client.Unsubscribe(topic)
		if token.Wait() && token.Error() != nil {
			err = token.Error()
			return
		}
	} else {
		// remove agent from list of subscribed agents
		ag[index] = ag[len(ag)-1]
		ag[len(ag)-1] = nil
		ag = ag[:len(ag)-1]
		mqttCol.mutex.Lock()
		mqttCol.subscription[topic] = ag
		mqttCol.mutex.Unlock()
	}
	return
}

// publish sends a message
func (mqttCol *mqttCollector) publish(msg schemas.MQTTMessage, qos int) (err error) {
	if mqttCol.config.Active {
		token := mqttCol.client.Publish(msg.Topic, byte(qos), false, msg.Content)
		token.Wait()
	}
	return
}

// deliverMsg delivers incoming messages to agents according to their topic
func (mqttCol *mqttCollector) deliverMsgs() {
	var msg schemas.MQTTMessage
	for {
		msg = <-mqttCol.msgIn
		mqttCol.mutex.Lock()
		ag, ok := mqttCol.subscription[msg.Topic]
		mqttCol.mutex.Unlock()
		if ok {
			for i := range ag {
				ag[i].newIncomingMQTTMessage(msg)
			}
		}
	}
}

// AgentMQTT provides functions to subscribe and publish via mqtt
type AgentMQTT struct {
	collector  *mqttCollector
	mutex      *sync.Mutex                         // mutex for message inbox map
	subTopic   map[string]interface{}              // subscribed topics
	msgInTopic map[string]chan schemas.MQTTMessage // message inbox for messages with specified topic
	msgIn      chan schemas.MQTTMessage            // mqtt message inbox
	agentID    int
	logger     *client.AgentLogger
	logError   *log.Logger
	logInfo    *log.Logger
	active     bool
}

// newAgentMQTT returns a new pubsub connector of type mqtt
func (mqttCol *mqttCollector) newAgentMQTT(agentID int, cmaplog *client.AgentLogger,
	logErr *log.Logger, logInf *log.Logger) (mq *AgentMQTT) {
	mq = &AgentMQTT{
		collector: mqttCol,
		mutex:     &sync.Mutex{},
		agentID:   agentID,
		logger:    cmaplog,
		logError:  logErr,
		logInfo:   logInf,
		active:    mqttCol.config.Active,
	}
	mq.subTopic = make(map[string]interface{})
	mq.msgInTopic = make(map[string]chan schemas.MQTTMessage)
	mq.msgIn = make(chan schemas.MQTTMessage)
	return
}

// close closes the mqtt
func (mq *AgentMQTT) close() {
	for t := range mq.subTopic {
		mq.Unsubscribe(t)
	}
	mq.mutex.Lock()
	mq.logInfo.Println("Closing MQTT of agent ", mq.agentID)
	mq.active = false
	mq.mutex.Unlock()
}

// Subscribe subscribes to a topic
func (mq *AgentMQTT) Subscribe(topic string, qos int) (err error) {
	mq.mutex.Lock()
	if !mq.active {
		mq.mutex.Unlock()
		err = errors.New("mqtt not active")
		return
	}
	_, ok := mq.subTopic[topic]
	mq.mutex.Unlock()
	if ok {
		return
	}
	mq.mutex.Lock()
	mq.subTopic[topic] = nil
	mq.mutex.Unlock()
	err = mq.collector.subscribe(mq, topic, qos)
	return
}

// Unsubscribe unsubscribes a topic
func (mq *AgentMQTT) Unsubscribe(topic string) (err error) {
	mq.mutex.Lock()
	if !mq.active {
		mq.mutex.Unlock()
		err = errors.New("mqtt not active")
		return
	}
	_, ok := mq.subTopic[topic]
	mq.mutex.Unlock()
	if !ok {
		return
	}
	mq.mutex.Lock()
	delete(mq.subTopic, topic)
	mq.mutex.Unlock()
	err = mq.collector.unsubscribe(mq, topic)
	return
}

// SendMessage sends a message
func (mq *AgentMQTT) SendMessage(msg schemas.MQTTMessage, qos int) (err error) {
	mq.mutex.Lock()
	if !mq.active {
		mq.mutex.Unlock()
		err = errors.New("mqtt not active")
		return
	}
	mq.mutex.Unlock()
	err = mq.collector.publish(msg, qos)
	if err != nil {
		return
	}
	err = mq.logger.NewLog("msg", "MQTT publish", msg.String())
	return
}

// NewMessage returns a new initiaized message
func (mq *AgentMQTT) NewMessage(topic string, content []byte) (msg schemas.MQTTMessage, err error) {
	msg.Topic = topic
	msg.Content = content
	err = nil
	return
}

// RecvMessages retrieves all messages since last call of this function
func (mq *AgentMQTT) RecvMessages() (num int, msgs []schemas.MQTTMessage, err error) {
	mq.mutex.Lock()
	if !mq.active {
		mq.mutex.Unlock()
		err = errors.New("mqtt not active")
		return
	}
	mq.mutex.Unlock()
	num = 0
	err = nil
	for {
		select {
		case msgtemp := <-mq.msgIn:
			msgs = append(msgs, msgtemp)
			num++
		default:
			return
		}
	}
}

// RecvMessageWait retrieves next message and blocks if no message is available
func (mq *AgentMQTT) RecvMessageWait() (msg schemas.MQTTMessage, err error) {
	mq.mutex.Lock()
	if !mq.active {
		mq.mutex.Unlock()
		err = errors.New("mqtt not active")
		return
	}
	mq.mutex.Unlock()
	err = nil
	msg = <-mq.msgIn
	return
}

// newIncomingMQTTMessage adds message to channel for incoming messages
func (mq *AgentMQTT) newIncomingMQTTMessage(msg schemas.MQTTMessage) {
	mq.mutex.Lock()
	if !mq.active {
		mq.mutex.Unlock()
		return
	}
	mq.mutex.Unlock()
	mq.logger.NewLog("msg", "MQTT receive", msg.String())
	mq.mutex.Lock()
	inbox, ok := mq.msgInTopic[msg.Topic]
	mq.mutex.Unlock()
	if ok {
		inbox <- msg
	} else {
		mq.msgIn <- msg
	}
}

func (mq *AgentMQTT) registerTopicChannel(topic string,
	topicChan chan schemas.MQTTMessage) (err error) {
	mq.mutex.Lock()
	if !mq.active {
		mq.mutex.Unlock()
		err = errors.New("mqtt not active")
		return
	}

	if _, ok := mq.msgInTopic[topic]; !ok {
		mq.msgInTopic[topic] = topicChan
	} else {
		err = errors.New("topic is already handled")
	}
	mq.mutex.Unlock()
	return
}

func (mq *AgentMQTT) deregisterTopicChannel(topic string) (err error) {
	mq.mutex.Lock()
	if _, ok := mq.msgInTopic[topic]; ok {
		delete(mq.msgInTopic, topic)
	} else {
		err = errors.New("topic is not handled")
	}
	mq.mutex.Unlock()
	return
}
