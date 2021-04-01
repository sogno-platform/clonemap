# Copyright 2020 Institute for Automation of Complex Power Systems,
# E.ON Energy Research Center, RWTH Aachen University
#
# This project is licensed under either of
# - Apache License, Version 2.0
# - MIT License
# at your option.
#
# Apache License, Version 2.0:
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.
#
# MIT License:
#
# Permission is hereby granted, free of charge, to any person obtaining a copy
# of this software and associated documentation files (the "Software"), to deal
# in the Software without restriction, including without limitation the rights
# to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
# copies of the Software, and to permit persons to whom the Software is
# furnished to do so, subject to the following conditions:
#
# The above copyright notice and this permission notice shall be included in
# all copies or substantial portions of the Software.
#
# THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
# IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
# FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
# AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
# LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
# OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
# THE SOFTWARE.

FROM golang:1.15.8 AS agency_builder

WORKDIR /clonemap
COPY go.mod .
COPY go.sum .
RUN go mod download
COPY cmd/agency cmd/agency
COPY pkg/agency pkg/agency
COPY pkg/client pkg/client
COPY pkg/schemas pkg/schemas
COPY pkg/status pkg/status
COPY pkg/common pkg/common
ENV PATH="/clonemap:${PATH}"
RUN cd cmd/agency; CGO_ENABLED=0 GOOS=linux go build -ldflags '-s' -o agency; cp agency /clonemap/

FROM alpine:latest

WORKDIR /root/
#RUN apk add --update netbase ca-certificates
COPY --from=agency_builder /clonemap/agency .
ENV PATH="/root:${PATH}"
EXPOSE 10000
CMD ["./agency"]
