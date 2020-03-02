# User Guide

## Step-by-step example

In the following we provide you with a step-by-step example for agent behavior implementation, cloneMAP deployment, MAS deployment and log analysis. It is assumed that you have followed the instruction for the [installation requirements](installation.md). In the example two agents will be started which send each other one message and indicate the receipt with a log message.

### Step 1 Behavior implementation

All agents are executed in agencies. An agency is a single container pod in a StatefulSet. One agency can host multiple agents. For each agent one task function is executed in a seperate go-routine. This task function is defined by the MAS developer.

If you want to use cloneMAP for MAS application development the only thing you have to implement is the agent behavior. In order to implement a certain agent behavior you have to use the agency package (pkg/agency). This package defines the *Agent* structure which can be used to access platform functionalities like messaging or interaction with the DF and the other modules.

Implement a task function that takes a pointer to an *Agent* object as parameter and returns an *error*. Start the agency using that task function. Every started agent will execute this task function as seperate go-routine.

In the base directory create your your Go module file ```go.mod```

```Go
module example

require (
    git.rwth-aachen.de/acs/public/cloud/mas/clonemap v0.0.0-20200109090525-86fd277abf43
)
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
    ag.Logger.NewLog("app", "This is agent " + strconv.Itoa(ag.GetAgentID()), nil)
    return
}
```

We define the agent behavior in the method ```task```. In the main function we start the agency with the task function as parameter. Save the code in the file ```cmd/main.go```.

#### Using other programming languages

Components in cloneMAP interact with each other using a REST API. This is also true for the agency. Hence, you don't have to use Go as a programming language for your agent behavior. In order to use any other language you have make sure that the agency implements the API. The API specification for the agency can be found in [api/agency](../api/agency/openapi.yaml) directory. For interaction with other components, e.g. the DF, you have to implement clients that make use of the corresponding API. All REST APIs are specified using the openapi 3 format.

### Step 2 Building the Docker image

Create the following file ```build/Dockerfile```

```Docker
FROM golang:1.11.2 AS agency_builder

WORKDIR /example
COPY cmd cmd
COPY go.mod .
RUN go mod download
ENV PATH="/example:${PATH}"
RUN cd cmd; GO111MODULE=on CGO_ENABLED=0 GOOS=linux go build -ldflags '-s' -o agency; cp agency /example/

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

Create the file ```scenario.yaml``` with following content.

```yaml
{"spec":{"id":0,"name":"example","registry":"<registry>","image":"<image>","secret":"<pull-secret>","agentsperagency":1,"logging":true,"mqtt":true,"df":true,"log":{"msg":true,"app":true,"status":true,"debug":true},"uptime":"0001-01-01T00:00:00Z"},"agents":[{"masid":0,"agencyid":0,"nodeid":0,"id":0,"name":"ExampleAgent","type":"example","custom":""},{"masid":0,"agencyid":0,"nodeid":0,"id":0,"name":"ExampleAgent","type":"example","custom":""}],"graph":{"node":null,"edge":null}}
```

Replace ```<registry>```, ```<image>``` and ```<pull-secret>``` with the registry name, the image name and the name of the Kubernetes pull secret for the registry (in case it is a private registry). The given scenario defines a MAS with two agents. One agency contains one agent which will lead to the creation of two agency pods. The previously created Docker image will be used for the agencies.

### Step 4 cloneMAP deployment

It is assumed that you have a running Kubernetes cluster. Deploy cloneMAP by applying the yaml file in ```deployments``` directory:

```bash
kubectl apply -f deployments/clonemap.yaml
```

This will create a new namespace called ```clonemap``` and deploy all cloneMAP components to that namespace. The process will take a few seconds.

#### Local deployment

Although cloneMAP is designed to be executed within a cloud-hosted container orchestration it can also be executed locally. This is ment as an easier way to test your MAS application. Note that components such as AMS, DF and Logger cannot be scaled horizontally with this method. In order to use the local version you have to build all necessary Docker images locally (AMS, DF, Logger). Afterwards go to cmd/kubestub directory and execute the stub:

```bash
cd cmd/kubestub
go run main.go
```

The stub starts AMS, Logger, DF and a MQTT broker as Docker containers.

### Step 5 MAS execution

In order to execute a MAS you have to post the previously created ```scenario.yaml``` file to the AMS. Look at the [api specification](../api/ams/openapi.yaml) for the correct path and method to be used. Subsequently the AMS will start all agencies which then will start the single agents. Depending on the size the creation of a MAS might take a few seconds.

```bash
curl -X "POST" -d @scenarios.yaml http://<ip-address>:30009/api/clonemap/mas
```

The AMS is made available to the outside world via NodePort on port 30009. Replace ```<ip-address>``` with the IP address of one of your Kubernetes machines. The AMS should answer with http code 201. It starts the two agencies as StatefulSet which automatically execute one agent each.

### Step 6 Analysis

Use the logger module to request logged messages

```bash
curl -X "GET" http://<ip-address>:30011/api/logging/0/0/app/latest/10
```
