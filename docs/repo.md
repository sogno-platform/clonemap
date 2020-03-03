# Repository structure

The structure of the repository follows the [Golang standard project layout](https://github.com/golang-standards/project-layout).

## `api` - API documentation

This folder contains documentation of the http API implemented by the different components of cloneMAP.

## `build` - Dockerfiles

This folder contains the Dockerfiles used to build Docker images of all cloneMAP components.

## `cmd` - main files

This folder contains the `main.go` files for all cloneMAP components.

## `deployments` - Kubernetes

This folder contains the yaml file required to start cloneMAP on a Kubernetes cluster.

## `docs` - documentation

This folder contains documentation for getting started with cloneMAP.

## `examples` - benchmark behavior

This folder contains the implementation of agent behavior used for the benchmarking of cloneMAP.

## `pkg` - cloneMAP packages

This folder contains the Go packages which implement the functionality of the different cloneMAP components. There exists one package for each component as well as client packages and some commonly used packages.
