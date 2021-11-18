# User Guide

## Step-by-step example

In the following we provide you with a step-by-step example for agent behavior implementation, MAS deployment and log analysis.
It is assumed that you have a running cloneMAP instance, either locally or on a Kubernetes cluster.
If not, please have a look at the [Admin Guide](administration_guide.md).
In the example two agents will be started which send each other one message and indicate the receipt with a log message.

### Step 1 Behavior implementation

All agents are executed in agencies.
An agency is a single container pod in a StatefulSet.
One agency can host multiple agents.
For each agent one task function is executed in a seperate go-routine.
This task function is defined by the MAS developer.

If you want to use cloneMAP for MAS application development the only thing you have to implement is the agent behavior.
In order to implement a certain agent behavior you have to use the agency package (pkg/agency).
This package defines the *Agent* structure which can be used to access platform functionalities like messaging or interaction with the DF and the other modules.

Implement a task function that takes a pointer to an *Agent* object as parameter and returns an *error*.
Start the agency using that task function.
Every started agent will execute this task function as seperate go-routine.

In the base directory of your project create your Go module file `go.mod` and add clonemap as dependency

```bash
go mod init example
go get github.com/RWTH-ACS/clonemap/pkg/agency@develop
```

The following Go code is used for the implementation of the behavior

```Go
package main

import (
    "fmt"
    "strconv"

    "git.rwth-aachen.de/acs/public/cloud/mas/clonemap/pkg/agency"
)

func main() {
    err := agency.StartAgency(task)
    if err != nil {
        fmt.Println(err)
    }
}

func task(ag *agency.Agent) (err error) {
    id := ag.GetAgentID()
    ag.Logger.NewLog("app", "This is agent "+strconv.Itoa(id), nil)
    msg, _ := ag.ACL.NewMessage((id+1)%2, 0, 0, "Message from agent "+strconv.Itoa(id))
    ag.ACL.SendMessage(msg)
    msg, _ = ag.ACL.RecvMessageWait()
    ag.Logger.NewLog("app", msg.Content, nil)
    return
}
```

We define the agent behavior in the method `task`.
In the main function we start the agency with the task function as parameter.
Save the code in the file `cmd/main.go`.

#### Using other programming languages

Components in cloneMAP interact with each other using a REST API. This is also true for the agency.
Hence, you don't have to use Go as a programming language for your agent behavior.
In order to use any other language you have to make sure that the agency implements the API.
The API specification for the agency can be found in [api/agency](../api/agency/openapi.yaml) directory.
For interaction with other components, e.g. the DF, you have to implement clients that make use of the corresponding API.
All REST APIs are specified using the openapi 3 format.

A Python package for the implementation of agents in Python is already available [here](https://github.com/RWTH-ACS/clonemapy).

### Step 2 Building the Docker image

Create the following file `build/Dockerfile`

```Docker
FROM golang:1.13.8 AS agency_builder

WORKDIR /example
COPY go.mod .
RUN go mod download
COPY cmd cmd
ENV PATH="/example:${PATH}"
RUN cd cmd; CGO_ENABLED=0 GOOS=linux go build -ldflags '-s' -o agency; cp agency /example/

FROM alpine:latest

WORKDIR /root/
#RUN apk add --update netbase ca-certificates
COPY --from=agency_builder /example/agency .
ENV PATH="/root:${PATH}"
EXPOSE 10000
CMD ["./agency"]
```

Build the docker image, tag it and push it to a registry. Replace the name of the image and the registry with your data

```bash
docker build -f build/Dockerfile -t <imagename> .
docker tag <imagename> <registryname>
docker push <registryname>
```

### Step 3 Scenario creation

Besides specifying the agent behavior you also have to define the MAS configuration in order to start a MAS application. The MAS configuration consists of a list of agents you want to execute as well as individual configuration information for each agent. Take a look at the [API specification of the AMS](../api/ams/openapi.yaml) for a description of the MAS configuration. In the configuration you also have to specify the docker image to be used for the agencies. This corresponds to the image you created in the previous step.

Create the file `scenario.json` with following content.

```json
{
    "config":{
        "name":"test",
        "agentsperagency":1,
        "mqtt":{
            "active":true
        },
        "df":{
            "active":true
        },
        "logger":{
            "active":true,
            "msg":true,
            "app":true,
            "status":true,
            "debug":true
        }
    },
    "imagegroups":[
        {
            "config":{
                "image":"<image>",
                "secret":"<pull-secret>"
            },
            "agents":[
                {
                    "nodeid":0,
                    "name":"ExampleAgent",
                    "type":"example",
                    "custom":""
                },
                {
                    "nodeid":0,
                    "name":"ExampleAgent",
                    "type":"example",
                    "custom":""
                }
            ]
        }
    ],
    "graph":{
        "node":null,
        "edge":null
    }
}
```

Replace `<image>` and `<pull-secret>` with the image name and the name of the Kubernetes pull secret for the registry (in case it is a private registry).
The given scenario defines a MAS with two agents.
One agency contains one agent which will lead to the creation of two agency pods.
The previously created Docker image will be used for the agencies.

### Step 4 MAS execution

In order to execute a MAS you have to post the previously created `scenario.json` file to the AMS.
Look at the [api specification](../api/ams/openapi.yaml) for the correct path and method to be used.
Subsequently the AMS will start all agencies which then will start the single agents.
Depending on the size the creation of a MAS might take a few seconds.

```bash
curl -X "POST" -d @scenario.yaml <ip-address>:30009/api/clonemap/mas
```

The AMS is made available to the outside world via NodePort on port 30009.
Replace `<ip-address>` with the IP address of one of your Kubernetes machines.
The AMS should answer with http code 201.
It starts the two agencies as StatefulSet which automatically execute one agent each.

### Step 5 Analysis

Use the logger module to request logged messages

```bash
curl -X "GET" <ip-address>:30011/api/logging/0/0/app/latest/10
```

### Step 6 MAS termination

Terminate the MAS by sending the following request to the AMS

```bash
curl -X DELETE <ip-address>:30009/api/clonemap/mas/0
```
