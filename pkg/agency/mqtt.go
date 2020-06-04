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
	"os"
	"strconv"
	"sync"

	"git.rwth-aachen.de/acs/public/cloud/mas/clonemap/pkg/schemas"
	mqtt "github.com/eclipse/paho.mqtt.golang"
)

// MQTT provides functions to subscribe and publish via mqtt
type MQTT struct {
	client     *mqttClient
	mutex      *sync.Mutex                         // mutex for message inbox map
	subTopic   map[string]interface{}              // subscribed topics
	msgInTopic map[string]chan schemas.MQTTMessage // message inbox for messages with specified topic
	msgIn      chan schemas.MQTTMessage            // mqtt message inbox
	agentID    int
	cmapLogger *Logger
	logError   *log.Logger
	logInfo    *log.Logger
	active     bool
}

// newMQTT returns a new pubsub connector of type mqtt
func newMQTT(agentID int, cli *mqttClient, cmaplog *Logger, logErr *log.Logger,
	logInf *log.Logger) (mq *MQTT) {
	mq = &MQTT{
		client:     cli,
		mutex:      &sync.Mutex{},
		agentID:    agentID,
		cmapLogger: cmaplog,
		logError:   logErr,
		logInfo:    logInf,
		active:     true,
	}
	mq.subTopic = make(map[string]interface{})
	mq.msgInTopic = make(map[string]chan schemas.MQTTMessage)
	mq.msgIn = make(chan schemas.MQTTMessage)
	return
}

// close closes the mqtt
func (mq *MQTT) close() {
	for t := range mq.subTopic {
		mq.Unsubscribe(t)
	}
	mq.mutex.Lock()
	mq.logInfo.Println("Closing MQTT of agent ", mq.agentID)
	mq.active = false
	mq.mutex.Unlock()
	return
}

// Subscribe subscribes to a topic
func (mq *MQTT) Subscribe(topic string, qos int) (err error) {
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
	err = mq.client.subscribe(mq, topic, qos)
	return
}

// Unsubscribe unsubscribes a topic
func (mq *MQTT) Unsubscribe(topic string) (err error) {
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
	err = mq.client.unsubscribe(mq, topic)
	return
}

// SendMessage sends a message
func (mq *MQTT) SendMessage(msg schemas.MQTTMessage, qos int) (err error) {
	mq.mutex.Lock()
	if !mq.active {
		mq.mutex.Unlock()
		err = errors.New("mqtt not active")
		return
	}
	mq.mutex.Unlock()
	err = mq.client.publish(msg, qos)
	if err != nil {
		return
	}
	err = mq.cmapLogger.NewLog("msg", "Sent MQTT message: "+string(msg.Content), string(msg.Content))
	return
}

// NewMessage returns a new initiaized message
func (mq *MQTT) NewMessage(topic string, content []byte) (msg schemas.MQTTMessage, err error) {
	msg.Topic = topic
	msg.Content = content
	err = nil
	return
}

// RecvMessages retrieves all messages since last call of this function
func (mq *MQTT) RecvMessages() (num int, msgs []schemas.MQTTMessage, err error) {
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
func (mq *MQTT) RecvMessageWait() (msg schemas.MQTTMessage, err error) {
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
func (mq *MQTT) newIncomingMQTTMessage(msg schemas.MQTTMessage) {
	mq.mutex.Lock()
	if !mq.active {
		mq.mutex.Unlock()
		return
	}
	mq.mutex.Unlock()
	mq.cmapLogger.NewLog("msg", "Received MQTT message: "+string(msg.Content), string(msg.Content))
	mq.mutex.Lock()
	inbox, ok := mq.msgInTopic[msg.Topic]
	mq.mutex.Unlock()
	if ok {
		inbox <- msg
	} else {
		mq.msgIn <- msg
	}
	return
}

func (mq *MQTT) registerTopicChannel(topic string, topicChan chan schemas.MQTTMessage) (err error) {
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

func (mq *MQTT) deregisterTopicChannel(topic string) (err error) {
	mq.mutex.Lock()
	if _, ok := mq.msgInTopic[topic]; ok {
		delete(mq.msgInTopic, topic)
	} else {
		err = errors.New("topic is not handled")
	}
	mq.mutex.Unlock()
	return
}

// mqttClient is the agency client for mqtt
type mqttClient struct {
	client       mqtt.Client              // mqtt client
	brokerSvc    string                   // name of the mqtt broker service
	brokerPort   int                      // port of the mqtt broker
	msgIn        chan schemas.MQTTMessage // mqtt message inbox
	name         string                   // agency name
	active       bool                     // indicates if mqtt is active (switch via env)
	mutex        *sync.Mutex              // mutex for message inbox map
	subscription map[string][]*MQTT       // map for subscription topics to agents' mqtt object
	// numDeliverer int                      // number of go routines for delivery
	logError *log.Logger
	logInfo  *log.Logger
}

// newMQTTClient creates a new mqtt agency client
func newMQTTClient(svc string, port int, name string, logErr *log.Logger,
	logInf *log.Logger) (cli *mqttClient) {
	cli = &mqttClient{
		brokerSvc:  svc,
		brokerPort: port,
		name:       name,
		mutex:      &sync.Mutex{},
		active:     false,
		logError:   logErr,
		logInfo:    logInf,
	}
	act := os.Getenv("CLONEMAP_MQTT")
	if act == "ON" {
		cli.active = true
	}
	cli.msgIn = make(chan schemas.MQTTMessage, 1000)
	cli.subscription = make(map[string][]*MQTT)
	cli.logInfo.Println("Created new MQTT client; status: ", cli.active)
	return
}

func (cli *mqttClient) init() (err error) {
	if cli.active {
		opts := mqtt.NewClientOptions().AddBroker("tcp://" + cli.brokerSvc + ":" +
			strconv.Itoa(cli.brokerPort)).SetClientID(cli.name)
		opts.SetDefaultPublishHandler(cli.newIncomingMQTTMessage)
		cli.client = mqtt.NewClient(opts)
		if token := cli.client.Connect(); token.Wait() && token.Error() != nil {
			err = errors.New("MQTTInitError")
			return
		}
		for i := 0; i < 3; i++ {
			go cli.deliverMsgs()
		}
		// cli.numDeliverer = 3
	}
	return
}

func (cli *mqttClient) close() (err error) {
	if cli.active {
		cli.logInfo.Println("Disconnecting MQTT client")
		cli.client.Disconnect(250)
	}
	err = nil
	return
}

// newIncomingMQTTMessage adds message to channel for incoming messages
func (cli *mqttClient) newIncomingMQTTMessage(client mqtt.Client, msg mqtt.Message) {
	var mqttMsg schemas.MQTTMessage
	mqttMsg.Content = msg.Payload()
	mqttMsg.Topic = msg.Topic()
	cli.msgIn <- mqttMsg
	return
}

// subscribe subscribes to specified topics
func (cli *mqttClient) subscribe(mq *MQTT, topic string, qos int) (err error) {
	if !cli.active {
		return
	}
	cli.mutex.Lock()
	ag, ok := cli.subscription[topic]
	cli.mutex.Unlock()
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
			cli.mutex.Lock()
			cli.subscription[topic] = ag
			cli.mutex.Unlock()
		}
	} else {
		cli.mutex.Lock()
		token := cli.client.Subscribe(topic, byte(qos), nil)
		cli.mutex.Unlock()
		if token.Wait() && token.Error() != nil {
			err = token.Error()
			return
		}
		ag = make([]*MQTT, 0)
		ag = append(ag, mq)
		cli.mutex.Lock()
		cli.subscription[topic] = ag
		cli.mutex.Unlock()
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
func (cli *mqttClient) unsubscribe(mq *MQTT, topic string) (err error) {
	if !cli.active {
		return
	}
	cli.mutex.Lock()
	ag, ok := cli.subscription[topic]
	cli.mutex.Unlock()
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
		delete(cli.subscription, topic)
		token := cli.client.Unsubscribe(topic)
		if token.Wait() && token.Error() != nil {
			err = token.Error()
			return
		}
	} else {
		// remove agent from list of subscribed agents
		ag[index] = ag[len(ag)-1]
		ag[len(ag)-1] = nil
		ag = ag[:len(ag)-1]
		cli.mutex.Lock()
		cli.subscription[topic] = ag
		cli.mutex.Unlock()
	}
	return
}

// publish sends a message
func (cli *mqttClient) publish(msg schemas.MQTTMessage, qos int) (err error) {
	if cli.active {
		token := cli.client.Publish(msg.Topic, byte(qos), false, msg.Content)
		token.Wait()
	}
	return
}

// deliverMsg delivers incoming messages to agents according to their topic
func (cli *mqttClient) deliverMsgs() {
	var msg schemas.MQTTMessage
	for {
		msg = <-cli.msgIn
		cli.mutex.Lock()
		ag, ok := cli.subscription[msg.Topic]
		cli.mutex.Unlock()
		if ok {
			for i := range ag {
				ag[i].newIncomingMQTTMessage(msg)
			}
		}
	}
}
