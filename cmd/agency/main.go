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
	err := agency.StartAgency(task)
	if err != nil {
		fmt.Println(err)
	}
}

func task(ag *agency.Agent) (err error) {
	time.Sleep(2 * time.Second)
	id := ag.GetAgentID()
	recv := (id + 1) % 2
	msg, _ := ag.ACL.NewMessage(recv, 0, 0, "test message")
	ag.ACL.SendMessage(msg)
	ag.Logger.NewLog("app", "This is agent "+strconv.Itoa(id), "")
	time.Sleep(2 * time.Second)
	ag.Logger.NewLog("beh", "This is the behavior of the agent"+strconv.Itoa(id), "")
	ag.Logger.NewLog("debug", "This is the debug of the agent"+strconv.Itoa(id), "")
	svc := schemas.Service{
		Desc: "agent" + strconv.Itoa(id),
	}
	_, err = ag.DF.RegisterService(svc)
	if err != nil {
		fmt.Println(err)
	}
	for i := 0; i < 5; i++ {
		time.Sleep(2 * time.Second)
		/* 		idx := rand.Intn(4) + 1 */
		for idx := 1; idx < 5; idx++ {
			ag.Logger.NewLogSeries("type"+strconv.Itoa(idx), rand.Float64())
		}
	}
	return
}
