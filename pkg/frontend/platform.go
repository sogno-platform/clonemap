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

package frontend

import (
	"errors"
	"net/http"

	amsclient "git.rwth-aachen.de/acs/public/cloud/mas/clonemap/pkg/ams/client"
	"git.rwth-aachen.de/acs/public/cloud/mas/clonemap/pkg/common/httpreply"
	dfclient "git.rwth-aachen.de/acs/public/cloud/mas/clonemap/pkg/df/client"
	logclient "git.rwth-aachen.de/acs/public/cloud/mas/clonemap/pkg/logger/client"
	"git.rwth-aachen.de/acs/public/cloud/mas/clonemap/pkg/schemas"
)

// handlePlatform handles requests to /api/pf/...
func (fe *Frontend) handlePlatform(w http.ResponseWriter, r *http.Request,
	respath []string) (resvalid bool, cmapErr error, httpErr error) {
	resvalid = false
	switch len(respath) {
	case 4:
		if respath[3] == "modules" {
			cmapErr, httpErr = fe.handleModules(w, r)
			resvalid = true
		}
	default:
		cmapErr = errors.New("Resource not found")
	}
	return
}

// handleModules is the handler for requests to path /api/pf/modules
func (fe *Frontend) handleModules(w http.ResponseWriter, r *http.Request) (cmapErr, httpErr error) {
	if r.Method == "GET" {
		// return short info of all MAS
		var mods schemas.ModuleStatus
		mods, cmapErr = getModuleStatus()
		if cmapErr == nil {
			httpErr = httpreply.Resource(w, mods, cmapErr)
		} else {
			httpErr = httpreply.CMAPError(w, cmapErr.Error())
		}
	} else {
		httpErr = httpreply.MethodNotAllowed(w)
		cmapErr = errors.New("Error: Method not allowed on path /api/pf/modules")
	}
	return
}

// getModuleStatus returns the on/off status of all modules
func getModuleStatus() (mods schemas.ModuleStatus, err error) {
	mods.Logging = logclient.Alive()
	mods.Core = amsclient.Alive()
	mods.DF = dfclient.Alive()
	return
}
