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

package httpretry

import (
	"bytes"
	"fmt"

	//"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"time"
)

// Request sends a http request and retries in case of an error
func Request(req *http.Request, numRetries int, delay time.Duration) (resp *http.Response,
	err error) {
	client := &http.Client{}
	resp = &http.Response{}
	resp, err = client.Do(req)
	if err != nil {
		for i := 0; i <= numRetries; i++ {
			time.Sleep(delay)
			resp, err = client.Do(req)
			if err == nil {
				break
			} else {
				fmt.Println("http error: " + err.Error())
			}
		}
	}
	return
}

//Post sends a post request and retries in case of an error
func Post(client *http.Client, url string, contentType string, content []byte,
	delay time.Duration, numRetries int) (body []byte, httpStatus int, err error) {
	resp := &http.Response{}
	resp, err = client.Post(url, contentType, bytes.NewReader(content))
	if err != nil {
		// fmt.Println("http erorr: " + err.Error())
		for i := 0; i <= numRetries; i++ {
			time.Sleep(delay)
			resp, err = client.Post(url, contentType, bytes.NewReader(content))
			if err == nil {
				break
			} else {
				// fmt.Println("http erorr: " + err.Error())
			}
		}
	}
	if err == nil {
		httpStatus = resp.StatusCode
		body, err = ioutil.ReadAll(resp.Body)
		resp.Body.Close()
	}
	return
}

//Get sends a get request and retries in case of an error
func Get(client *http.Client, url string, delay time.Duration, numRetries int) (body []byte,
	httpStatus int, err error) {
	resp := &http.Response{}
	resp, err = client.Get(url)
	if err != nil {
		// fmt.Println("http erorr: " + err.Error())
		for i := 0; i <= numRetries; i++ {
			time.Sleep(delay)
			resp, err = client.Get(url)
			if err == nil {
				break
			} else {
				// fmt.Println("http erorr: " + err.Error())
			}
		}
	}
	if err == nil {
		httpStatus = resp.StatusCode
		body, err = ioutil.ReadAll(resp.Body)
		resp.Body.Close()
	}
	return
}

//Delete sends a delete request and retries in case of an error
func Delete(client *http.Client, url string, body io.Reader, delay time.Duration,
	numRetries int) (httpStatus int, err error) {
	request := &http.Request{}
	request, err = http.NewRequest("DELETE", url, nil)
	resp := &http.Response{}
	resp, err = client.Do(request)
	if err != nil {
		// fmt.Println("http erorr: " + err.Error())
		for i := 0; i <= numRetries; i++ {
			time.Sleep(delay)
			request, err = http.NewRequest("DELETE", url, nil)
			resp, err = client.Do(request)
			if err == nil {
				break
			} else {
				// fmt.Println("http erorr: " + err.Error())
			}
		}
	}
	if err == nil {
		httpStatus = resp.StatusCode
		io.Copy(ioutil.Discard, resp.Body)
		resp.Body.Close()
	}
	return
}

//Put sends a post request and retries in case of an error
func Put(client *http.Client, url string, content []byte, delay time.Duration,
	numRetries int) (body []byte, httpStatus int, err error) {
	request := &http.Request{}
	request, err = http.NewRequest("PUT", url, bytes.NewReader(content))
	resp := &http.Response{}
	resp, err = client.Do(request)
	if err != nil {
		// fmt.Println("http erorr: " + err.Error())
		for i := 0; i <= numRetries; i++ {
			time.Sleep(delay)
			request, err = http.NewRequest("PUT", url, bytes.NewReader(content))
			resp, err = client.Do(request)
			if err == nil {
				break
			} else {
				//fmt.Println("http error: " + err.Error())
			}
		}
	}
	if err == nil {
		httpStatus = resp.StatusCode
		body, err = ioutil.ReadAll(resp.Body)
		resp.Body.Close()
	}
	return
}
