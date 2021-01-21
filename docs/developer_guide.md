# Developer Guide

This developer guide aims at getting cloneMAP developers started.
It describes important implementation details of the existing modules and provides a description of how to add further modules to cloneMAP.
Please also have a look at the [repo structure](repo.md).

## General considerations

cloneMAP follows a cloud-native application design.
That means that the platform is composed by microservices which can be deployed seperately and offer a REST API for interaction with other microservices.
Microservices are implemented stateless.
State information is handled by distributed storages, i.e., an etcd store and a Cassandra DB.
Single microservices of the MAP are deployed as docker-containers.

cloneMAP can be deployed locally or on a Kubernetes Cluster (see [Admin Guide](administration_guide.md)).
All cloneMAP micorservices are deployed as Docker containers.
cloneMAP microservices are started on Kubernetes as Deployments with a corresponding Service.
These Deployments can be scaled horizontally.
Distributed storages are started as StatefulSets.

All existing modules are written in Go.
How to setup Go is described [here](https://golang.org/doc/install). The used Go version is Go1.13.8.
The code documentation of all components follows the [godoc](https://blog.golang.org/godoc-documenting-go-code) specification.
Dependency management is handled with [Go modules](https://github.com/golang/go/wiki/Modules) which has been included as official Go tool since Go1.11.
New modules can also be implemented in any other language since interaction between modules happens via a REST API.
However, it might be necessary to provide a Go client for the API in order to enable the usage of the new module from within other modules.

## Existing modules

TBD

### Core

#### AMS

#### Agencies

### DF

### Logging

### IoT

### Plug&Play

### WebUI

#### Backend

#### Frontend

## Implementing new modules

TBD
