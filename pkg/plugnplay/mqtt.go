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

package plugnplay

import (
	"encoding/json"
	"errors"
	"log"
	"strconv"
	"sync"

	amscli "git.rwth-aachen.de/acs/public/cloud/mas/clonemap/pkg/ams/client"
	"git.rwth-aachen.de/acs/public/cloud/mas/clonemap/pkg/schemas"
	mqtt "github.com/eclipse/paho.mqtt.golang"
)

// mqttClient is the PnP client for mqtt
type mqttClient struct {
	client     mqtt.Client // mqtt client
	brokerSvc  string      // name of the mqtt broker service
	brokerPort int         // port of the mqtt broker
	name       string      // agency name
	mutex      *sync.Mutex // mutex for message inbox map
	logError   *log.Logger
	logInfo    *log.Logger
	amsClient  *amscli.Client
}

// newMQTTClient creates a new mqtt agency client
func newMQTTClient(svc string, port int, name string, logErr *log.Logger,
	logInf *log.Logger, amsClient *amscli.Client) (cli *mqttClient) {
	cli = &mqttClient{
		brokerSvc:  svc,
		brokerPort: port,
		name:       name,
		mutex:      &sync.Mutex{},
		logError:   logErr,
		logInfo:    logInf,
		amsClient:  amsClient,
	}
	cli.logInfo.Println("Created MQTT client")
	return
}

func (cli *mqttClient) init() (err error) {
	opts := mqtt.NewClientOptions().AddBroker("tcp://" + cli.brokerSvc + ":" +
		strconv.Itoa(cli.brokerPort)).SetClientID(cli.name)
	opts.SetDefaultPublishHandler(cli.newIncomingMQTTMessage)
	cli.client = mqtt.NewClient(opts)
	if token := cli.client.Connect(); token.Wait() && token.Error() != nil {
		err = errors.New("MQTTInitError")
		return
	}
	return
}

func (cli *mqttClient) close() (err error) {
	cli.logInfo.Println("Disconnecting MQTT client")
	cli.client.Disconnect(250)
	err = nil
	return
}

// newIncomingMQTTMessage adds message to channel for incoming messages
func (cli *mqttClient) newIncomingMQTTMessage(client mqtt.Client, msg mqtt.Message) {
	var imSpec schemas.ImageGroupSpec
	var err error
	err = json.Unmarshal(msg.Payload(), &imSpec)
	if err != nil {
		cli.logError.Println(err)
		return
	}
	ags := []schemas.ImageGroupSpec{imSpec}
	_, err = cli.amsClient.PostAgents(0, ags)
	if err != nil {
		cli.logError.Println(err)
	}
}

// subscribe subscribes to specified topics
func (cli *mqttClient) subscribe(topic string, qos int) (err error) {
	token := cli.client.Subscribe(topic, byte(qos), nil)
	if token.Wait() && token.Error() != nil {
		err = token.Error()
		return
	}
	return
}
