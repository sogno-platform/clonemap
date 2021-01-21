# Adminitration Guide

This administration guide provides the knowledge necessary to start, maintain and stop cloneMAP, either locally or on a Kubernetes Cluster.

## Local deployment

Although cloneMAP is designed to be executed within a Kubernetes cluster it can also be executed locally.
This is ment as an easier way to test your MAS application.
Note that microservices such as the AMS, DF and Logger cannot be scaled horizontally with this method.
Moreover, the platform is not fault-tolerant when deployed locally and cannot be spread over several machines.

In order to start cloneMAP locally you need a docker installation.
How to get started with docker is explained [here](https://docs.docker.com/get-started/).
An installation guide for Ubuntu can be found [here](https://docs.docker.com/install/linux/docker-ce/ubuntu/).

After installing Docker you can start cloneMAP by running the following command:

```bash
docker run -p 8000:8000 -v /var/run/docker.sock:/var/run/docker.sock --name=kubestub registry.git.rwth-aachen.de/acs/public/cloud/mas/clonemap/clonemap_local
```

This command will start a Docker container of the image `registry.git.rwth-aachen.de/acs/public/cloud/mas/clonemap/clonemap_local` and assign the name `kubestub` to it.
The kubestub container will create a custom Docker network called `clonemap-net` and attach itself to this network.
Subsequently it will start the AMS in a seperate container on the host.
This is enabled by mounting the directory `/var/run/docker.sock` of the host to the same directory in the kubestub container.
As a result, all docker commands executed inside of the kubestub container will be send to the Docker daemon of the host.

The output of starting the kubestub container should look like this:

```bash
docker run -p 8000:8000 -v /var/run/docker.sock:/var/run/docker.sock --name=kubestub registry.git.rwth-aachen.de/acs/public/cloud/mas/clonemap/clonemap_local
>Create Bridge Network
>Create AMS Container
>Ready
```

You can check the success of starting cloneMAP with the `docker ps` command.
The output should look similar to this:

```bash
docker ps
>CONTAINER ID   IMAGE                                                                      COMMAND                  CREATED          STATUS          PORTS                                 NAMES
>951af55605ef   registry.git.rwth-aachen.de/acs/public/cloud/mas/clonemap/ams              "./ams"                  14 seconds ago   Up 12 seconds   0.0.0.0:30009->9000/tcp               ams
>00fd5d077b8a   registry.git.rwth-aachen.de/acs/public/cloud/mas/clonemap/clonemap_local   "docker-entrypoint.s…"   19 seconds ago   Up 16 seconds   0.0.0.0:8000->8000/tcp                kubestub
```

If you want use other modules you can start them as well by setting corresponding environment variables when starting the kubestub container.
For example, if you would like to start a MQTT broker the command would look like this:

```bash
docker run -p 8000:8000 -v /var/run/docker.sock:/var/run/docker.sock -e CLONEMAP_MODULE_MQTT=true --name=kubestub registry.git.rwth-aachen.de/acs/public/cloud/mas/clonemap/clonemap_local
```

The following environment variables can be used to start further modules:

* CLONEMAP_MODULE_MQTT: Mosquitto MQTT broker
* CLONEMAP_MODULE_DF: cloneMAP DF module
* CLONEMAP_MODULE_LOGGER: cloneMAP Logging module
* CLONEMAP_MODULE_PNP: cloneMAP PnP module
* CLONEMAP_MODULE_FRONTEND: cloneMAP WebUI module
* CLONEMAP_MODULE_FIWARE: FiWare Orion Broker ans MongoDB

Moreover, you can set the log level to either:

* CLONEMAP_LOG_LEVEL=info
* CLONEMAP_LOG_LEVEL=error (default)

The kubestub container is started in foreground mode.
It will give a message when a new container is started and show *Ready* once all specified modules are started.
You can terminate cloneMAP terminating the kubestub container.
In the concole where you started the kubestub container press `ctrl`+`c`.
This process might take a few seconds since the kubestub container cleans up before terminating.
That means that all containers started by the kubestub container are terminated.
A corresponding output shows the progress.
Subsequently it will detach itself from the clonemap-net network, delete the network and then terminate itself.
Starting the platform locally and then terminating it should produce the following output:

```bash
docker run -p 8000:8000 -v /var/run/docker.sock:/var/run/docker.sock --name=kubestub registry.git.rwth-aachen.de/acs/public/cloud/mas/clonemap/clonemap_local
>Create Bridge Network
>Create AMS Container
>Ready
>^CCaught sig: interrupt
>Stop AMS Container
>Delete Bridge Network
```

## Kubernetes deployment

Kubernetes can be installed using [kubeadm](https://kubernetes.io/docs/setup/production-environment/tools/kubeadm/create-cluster-kubeadm/).
If you use a public cloud provider you typically can create preconfigured Kubernetes clusters.

For testing purposes it might be more convenient to use a local installation of Kubernetes instead of a real production cluster.
This can be achieved by using Minikube.
Minikube starts a single virtual machine on your local machine and runs all necessary parts of Kubernetes in this VM.
Minikube can be installed as described [here](https://kubernetes.io/docs/tasks/tools/install-minikube/).

To get started with Kubernetes, please have a look [here](https://kubernetes.io/docs/home/).

Once you have a running Kubernets cluster you can deploy cloneMAP by applying the yaml file in `deployments` directory:

```bash
kubectl apply -f deployments/k8s.yaml
```

This will create a new namespace called *clonemap* and deploy all cloneMAP components to that namespace.
The process will take a few seconds.

During execution you can manage cloneMAP components by using the Kubernetes dashboard or `kubectl` commands.
For example you could scale microservice horizontally to cope with a high load.

cloneMAP is terminated by deleting all its resources:

```bash
kubectl delete -f deployments/k8s.yaml
```
