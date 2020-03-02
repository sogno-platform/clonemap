# Installation requirements

## Golang

cloneMAP is written in Go. Therefore, in order to get started you first need a running Go installation on your machine. How to setup Go is described [here](https://golang.org/doc/install). The used Go version is Go1.13.8.
The code documentation of all components follows the [godoc](https://blog.golang.org/godoc-documenting-go-code) specification.
Dependency management is handled with [Go modules](https://github.com/golang/go/wiki/Modules) which has been included as official Go tool since Go1.11.

## Docker

Single components of the MAP are deployed as docker-containers. How to get started with docker is explained [here](https://docs.docker.com/get-started/). An installation guide for Ubuntu can be found [here](https://docs.docker.com/install/linux/docker-ce/ubuntu/).

## Kubernetes

Kubernetes can be installed using [kubeadm](https://kubernetes.io/docs/setup/production-environment/tools/kubeadm/create-cluster-kubeadm/).
If you use a public cloud provider you typically can create preconfigured Kubernetes clusters.

For testing purposes it might be more convenient to use a local installation of Kubernetes instead of a real production cluster. This can be achieved by using Minikube. Minikube starts a single virtual machine on your local machine and runs all necessary parts of Kubernetes in this VM. Minikube can be installed as described [here](https://kubernetes.io/docs/tasks/tools/install-minikube/).
