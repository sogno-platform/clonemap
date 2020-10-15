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

package benchmark

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"strconv"
	"time"

	"git.rwth-aachen.de/acs/public/cloud/mas/clonemap/pkg/agency"
	"git.rwth-aachen.de/acs/public/cloud/mas/clonemap/pkg/schemas"
)

// CustomAgentData is the custom agent configuration
type CustomAgentData struct {
	NumAgents   int       // number of agents in mas
	BenchmarkID int       // id of task function to be executed
	PeerID      int       // id if agent to do ping pong with
	Start       bool      // indicates if agent starts ping pong
	StartTime   time.Time // start time of benchmark
	T           int       // cpu load period in ms
	Tr          float32   // part of T that agent is busy
}

// Task switches the right benchmark task
func Task(ag *agency.Agent) (err error) {
	var config CustomAgentData
	err = json.Unmarshal([]byte(ag.GetCustomData()), &config)
	if err != nil {
		fmt.Println(err)
		return
	}
	switch config.BenchmarkID {
	case 0:
		err = pingPong(ag, config)
	case 1:
		err = cpuLoad(ag, config)
	case 2:
		err = dfBench(ag, config)
	case 3:
		err = stateTest(ag, config)
	case 4:
		err = pingPongft(ag, config)
	default:
		fmt.Println("unknown task function")
	}
	if err != nil {
		fmt.Println(err)
		return
	}
	return
}

// pingPong Task function
func pingPong(ag *agency.Agent, config CustomAgentData) (err error) {
	// d := time.Until(config.StartTime)
	err = ag.Logger.NewLog("status", "Starting PingPong Behavior; Peer: "+
		strconv.Itoa(config.PeerID)+", Start: "+strconv.FormatBool(config.Start), "")
	if err != nil {
		fmt.Println(err)
		return
	}
	// time.Sleep(d)
	time.Sleep(time.Second * 40)
	if config.Start {
		var rtts []int
		var msg schemas.ACLMessage
		msg, err = ag.ACL.NewMessage(config.PeerID, schemas.FIPAProtQuery, schemas.FIPAPerfInform,
			"test msg")
		if err != nil {
			fmt.Println(err)
		}
		// send 1000 messages during start up time
		for i := 0; i < 1000; i++ {
			msg.Receiver = config.PeerID
			msg.Sender = ag.GetAgentID()
			err = ag.ACL.SendMessage(msg)
			if err != nil {
				fmt.Println(err)
			}
			msg, err = ag.ACL.RecvMessageWait()
			if err != nil {
				fmt.Println(err)
			}
		}
		// send 1000 messages during benchmark time
		for i := 0; i < 1000; i++ {
			msg.Receiver = config.PeerID
			msg.Sender = ag.GetAgentID()
			t := time.Now()
			err = ag.ACL.SendMessage(msg)
			if err != nil {
				fmt.Println(err)
			}
			msg, err = ag.ACL.RecvMessageWait()
			rtt := time.Since(t).Nanoseconds()
			if err != nil {
				fmt.Println(err)
			}
			rtts = append(rtts, int(rtt/1000))
		}
		// store rtts
		max := 0
		min := 1000000
		sum := 0
		avg := 0
		for i := 0; i < len(rtts); i++ {
			if max < rtts[i] {
				max = rtts[i]
			}
			if min > rtts[i] {
				min = rtts[i]
			}
			sum += rtts[i]
		}
		avg = sum / len(rtts)
		js, _ := json.Marshal(rtts)
		err = ag.Logger.NewLog("status", "RTT in µs: min: "+strconv.Itoa(min)+", max: "+
			strconv.Itoa(max)+", avg: "+strconv.Itoa(avg), string(js))
		if err != nil {
			fmt.Println(err)
		}
		// send 8000 messages during end time
		for i := 0; i < 8000; i++ {
			// for {
			msg.Receiver = config.PeerID
			msg.Sender = ag.GetAgentID()
			err = ag.ACL.SendMessage(msg)
			if err != nil {
				fmt.Println(err)
			}
			msg, err = ag.ACL.RecvMessageWait()
			if err != nil {
				fmt.Println(err)
			}
		}
	} else {
		// wait for messages and reply
		var msg schemas.ACLMessage
		for {
			msg, err = ag.ACL.RecvMessageWait()
			if err != nil {
				fmt.Println(err)
			}
			msg.Receiver = msg.Sender
			msg.Sender = ag.GetAgentID()
			err = ag.ACL.SendMessage(msg)
			if err != nil {
				fmt.Println(err)
			}
		}
	}
	return
}

// pingPong fault tolerance Task function
func pingPongft(ag *agency.Agent, config CustomAgentData) (err error) {
	// d := time.Until(config.StartTime)
	err = ag.Logger.NewLog("status", "Starting fault-tolerant PingPong Behavior; Peer: "+
		strconv.Itoa(config.PeerID)+", Start: "+strconv.FormatBool(config.Start), "")
	if err != nil {
		fmt.Println(err)
		return
	}
	// time.Sleep(d)
	time.Sleep(time.Second * 40)
	if config.Start {
		state := 0
		var rtts []int
		var msg schemas.ACLMessage
		msg, err = ag.ACL.NewMessage(config.PeerID, schemas.FIPAProtQuery, schemas.FIPAPerfInform,
			"test msg")
		if err != nil {
			fmt.Println(err)
		}
		// send 1000 messages during start up time
		for i := 0; i < 1000; i++ {
			msg.Receiver = config.PeerID
			msg.Sender = ag.GetAgentID()
			err = ag.ACL.SendMessage(msg)
			if err != nil {
				fmt.Println(err)
			}
			state++
			err = ag.Logger.UpdateState(strconv.Itoa(state))
			if err != nil {
				fmt.Println(err)
			}
			msg, err = ag.ACL.RecvMessageWait()
			if err != nil {
				fmt.Println(err)
			}
		}
		// send 1000 messages during benchmark time
		for i := 0; i < 1000; i++ {
			msg.Receiver = config.PeerID
			msg.Sender = ag.GetAgentID()
			t := time.Now()
			err = ag.ACL.SendMessage(msg)
			if err != nil {
				fmt.Println(err)
			}
			state++
			err = ag.Logger.UpdateState(strconv.Itoa(state))
			if err != nil {
				fmt.Println(err)
			}
			msg, err = ag.ACL.RecvMessageWait()
			rtt := time.Since(t).Nanoseconds()
			if err != nil {
				fmt.Println(err)
			}
			rtts = append(rtts, int(rtt/1000))
		}
		// store rtts
		max := 0
		min := 1000000
		sum := 0
		avg := 0
		for i := 0; i < len(rtts); i++ {
			if max < rtts[i] {
				max = rtts[i]
			}
			if min > rtts[i] {
				min = rtts[i]
			}
			sum += rtts[i]
		}
		avg = sum / len(rtts)
		js, _ := json.Marshal(rtts)
		err = ag.Logger.NewLog("status", "RTT in µs: min: "+strconv.Itoa(min)+", max: "+
			strconv.Itoa(max)+", avg: "+strconv.Itoa(avg), string(js))
		if err != nil {
			fmt.Println(err)
		}
		// send 8000 messages during end time
		for i := 0; i < 8000; i++ {
			// for {
			msg.Receiver = config.PeerID
			msg.Sender = ag.GetAgentID()
			err = ag.ACL.SendMessage(msg)
			if err != nil {
				fmt.Println(err)
			}
			state++
			err = ag.Logger.UpdateState(strconv.Itoa(state))
			if err != nil {
				fmt.Println(err)
			}
			msg, err = ag.ACL.RecvMessageWait()
			if err != nil {
				fmt.Println(err)
			}
		}
	} else {
		// wait for messages and reply
		var msg schemas.ACLMessage
		for {
			msg, err = ag.ACL.RecvMessageWait()
			if err != nil {
				fmt.Println(err)
			}
			msg.Receiver = msg.Sender
			msg.Sender = ag.GetAgentID()
			err = ag.ACL.SendMessage(msg)
			if err != nil {
				fmt.Println(err)
			}
		}
	}
	return
}

// cpuLoad is the task function for a benchmark with high cpu load
func cpuLoad(ag *agency.Agent, config CustomAgentData) (err error) {
	err = ag.Logger.NewLog("status", "Starting CPU Load Behavior; Peer: "+
		strconv.Itoa(config.PeerID)+", Start: "+strconv.FormatBool(config.Start), "")
	if err != nil {
		fmt.Println(err)
		return
	}
	T := float32(config.T)
	Tr := config.Tr
	a := make([]float64, 100, 100)
	b := make([]float64, 100, 100)
	c := make([]float64, 100, 100)
	for i := 0; i < 100; i++ {
		a[i] = rand.Float64()
		b[i] = rand.Float64()
		c[i] = rand.Float64()
	}
	time.Sleep(time.Second * 30)
	go communication(ag, config)
	for {
		tStart := time.Now()
		for {
			for i := 0; i < 100; i++ {
				a[i] = b[i] + c[i]
			}
			if time.Since(tStart) > time.Microsecond*time.Duration(int(1000*T*Tr)) {
				break
			}
		}
		time.Sleep(time.Microsecond * time.Duration(int(1000*T*(1-Tr))))
	}
}

func communication(ag *agency.Agent, config CustomAgentData) (err error) {
	var msg schemas.ACLMessage
	msg, err = ag.ACL.NewMessage(config.PeerID, schemas.FIPAProtQuery, schemas.FIPAPerfInform,
		"test msg")
	if err != nil {
		fmt.Println(err)
	}
	for {
		err = ag.ACL.SendMessage(msg)
		if err != nil {
			fmt.Println(err)
		}
		ag.ACL.RecvMessages()
		time.Sleep(time.Second * 1)
	}
}

func dfBench(ag *agency.Agent, config CustomAgentData) (err error) {
	err = ag.Logger.NewLog("status", "Starting DF Behavior; Peer: "+
		strconv.Itoa(config.PeerID)+", Start: "+strconv.FormatBool(config.Start), "")
	if err != nil {
		return
	}
	svc := schemas.Service{
		Desc: "svc" + strconv.Itoa(rand.Int()%8),
	}
	var svcID string
	var rtts []int
	time.Sleep(time.Second * 30)
	for i := 0; i < 100; i++ {
		_, err = ag.DF.SearchForService("svc" + strconv.Itoa(rand.Int()%8))
		if err != nil {
			return
		}
	}
	tStart := time.Now()

	t := time.Now()
	svcID, err = ag.DF.RegisterService(svc)
	if err != nil {
		return
	}
	rtts = append(rtts, int(time.Since(t).Nanoseconds())/1000)

	for i := 0; i < 8; i++ {
		t = time.Now()
		_, err = ag.DF.SearchForService("svc" + strconv.Itoa(rand.Int()%8))
		if err != nil {
			return
		}
		rtts = append(rtts, int(time.Since(t).Nanoseconds())/1000)
	}

	t = time.Now()
	err = ag.DF.DeregisterService(svcID)
	if err != nil {
		return
	}
	rtts = append(rtts, int(time.Since(t).Nanoseconds())/1000)

	rtts = append(rtts, int(time.Since(tStart).Nanoseconds())/1000)

	for i := 0; i < 100; i++ {
		_, err = ag.DF.SearchForService("svc" + strconv.Itoa(rand.Int()%8))
		if err != nil {
			return
		}
	}

	js, _ := json.Marshal(rtts)
	err = ag.Logger.NewLog("status", "Total time in µs: "+strconv.Itoa(rtts[10]), string(js))

	return
}

func stateTest(ag *agency.Agent, config CustomAgentData) (err error) {
	time.Sleep(time.Second * 10)
	err = ag.Logger.NewLog("status", "Starting state test Behavior", "")
	if err != nil {
		return
	}
	var state string
	state, err = ag.Logger.RestoreState()
	if err != nil {
		return
	}
	if state == "" {
		err = ag.Logger.NewLog("status", "No previous state", "")
		if err != nil {
			return
		}
	}
	state = "test state"
	err = ag.Logger.UpdateState(state)
	if err != nil {
		return
	}
	state, err = ag.Logger.RestoreState()
	if err != nil {
		return
	}
	err = ag.Logger.NewLog("status", "State: "+string(state), "")
	return
}
