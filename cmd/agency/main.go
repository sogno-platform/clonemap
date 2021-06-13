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

package main

import (
	"fmt"
	"math/rand"
	"strconv"
	"time"

	"git.rwth-aachen.de/acs/public/cloud/mas/clonemap/pkg/agency"
	"git.rwth-aachen.de/acs/public/cloud/mas/clonemap/pkg/schemas"
	// "git.rwth-aachen.de/acs/public/cloud/mas/clonemap/pkg/schemas"
)

func main() {
	err := agency.StartAgency(task_test)
	if err != nil {
		fmt.Println(err)
	}
}

func display(msg schemas.MQTTMessage) (err error) {
	time.Sleep(10 * time.Second)
	fmt.Println(string(msg.Content))
	return
}

func handleDefault(msg schemas.ACLMessage) (err error) {
	fmt.Println(string(msg.Content))
	return
}

func task_test(ag *agency.Agent) (err error) {
	id := ag.GetAgentID()

	// agent 0 subsribes the topic1
	if id == 0 {
		ag.MQTT.Subscribe("topic1", 1)
		behMQTT, err := ag.NewMQTTTopicBehavior("topic1", display)

		if err == nil {
			behMQTT.Start()
		}
	}

	// agent 1 publishes the topic1
	if id == 1 {
		for i := 0; i < 20; i++ {
			time.Sleep(5 * time.Second)
			msg := "test message" + strconv.Itoa(i)
			MQTTMsg, err := ag.MQTT.NewMessage("topic1", []byte(msg))
			if err == nil {
				ag.MQTT.SendMessage(MQTTMsg, 1)
			}
		}
	}

	// new protocol behavior
	/* 	handlePerformative := make(map[int]func(schemas.ACLMessage) error)
	   	handlePerformative[0] = handleDefault
	   	behPro, err := ag.NewMessageBehavior(0, handlePerformative, handleDefault)
	   	if err != nil {
	   		fmt.Println("protocol started with error")
	   		ag.Logger.NewLog("beh", "protocol error", "")
	   	} else {
	   		behPro.Start()
	   		time.Sleep(30 * time.Second)
	   	} */
	return
}

func task(ag *agency.Agent) (err error) {
	time.Sleep(10 * time.Second)
	id := ag.GetAgentID()

	// sends 40 messages randomly to other agents
	for i := 0; i < 40; i++ {
		interval := rand.Intn(5)
		time.Sleep(time.Duration(interval) * time.Second)
		recv := rand.Intn(20)
		if recv == id {
			continue
		}
		msg, _ := ag.ACL.NewMessage(recv, 0, 0, "test message")
		ag.ACL.SendMessage(msg)
	}

	// app logs
	cnt := rand.Intn(10)
	for i := 0; i < cnt; i++ {
		ag.Logger.NewLog("app", "This is agent "+strconv.Itoa(id), "")
		time.Sleep(2 * time.Second)
	}

	// service
	svc := schemas.Service{
		Desc: "agent" + strconv.Itoa(id),
	}
	_, err = ag.DF.RegisterService(svc)
	if err != nil {
		fmt.Println(err)
	}
	for i := 0; i < 5; i++ {
		time.Sleep(2 * time.Second)
		for idx := 1; idx < 5; idx++ {
			ag.Logger.NewLogSeries("type"+strconv.Itoa(idx), rand.Float64())
		}
	}

	return
}
