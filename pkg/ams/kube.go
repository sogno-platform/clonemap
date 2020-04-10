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

// interaction with kubernetes cluster

package ams

import (
	"errors"
	"os"
	"strconv"
	"time"

	"git.rwth-aachen.de/acs/public/cloud/mas/clonemap/pkg/schemas"
	apiappsv1 "k8s.io/api/apps/v1"
	apicorev1 "k8s.io/api/core/v1"
	resource "k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

// kubeDeployment implements the Cluster interface for a Kubernetes instance of the MAP
type kubeDeployment struct {
	deplType  string
	config    *rest.Config
	clientset *kubernetes.Clientset
	resLimit  bool
}

// newMAS triggers the cluster manager to start new agency containers
func (kube *kubeDeployment) newMAS(masID int, images schemas.ImageGroups, logging bool,
	mqtt bool, df bool) (err error) {
	var exist bool
	exist, err = kube.existStatefulSet(masID)
	if err == nil {
		if !exist {
			var loggingEnv, mqttEnv, dfEnv string
			if logging {
				loggingEnv = "ON"
			} else {
				loggingEnv = "OFF"
			}
			if mqtt {
				mqttEnv = "ON"
			} else {
				mqttEnv = "OFF"
			}
			if df {
				dfEnv = "ON"
			} else {
				dfEnv = "OFF"
			}

			err = kube.createHeadlessService(masID)
			if err != nil {
				return
			}
			for i := range images.Inst {
				err = kube.createStatefulSet(masID, i, images.Inst[i].Config.Image,
					images.Inst[i].Config.PullSecret,
					len(images.Inst[i].Agencies.Inst), loggingEnv, mqttEnv, dfEnv)
				if err != nil {
					return
				}
			}
		} else {
			// error
		}
	}
	time.Sleep(time.Second * 2)
	return
}

// scaleMAS triggers the cluster manager to start or delete agency containers
func (kube *kubeDeployment) scaleMAS(masID int, deltaAgencies int) (err error) {
	var exist bool
	exist, err = kube.existStatefulSet(masID)
	if err == nil {
		if exist {
			err = kube.scaleStatefulSet(masID, deltaAgencies)
		} else {
			// error
		}
	}
	return
}

// deleteMAS triggers the cluster manager to delete all agency containers
func (kube *kubeDeployment) deleteMAS(masID int) (err error) {
	var exist bool
	exist, err = kube.existStatefulSet(masID)
	if err == nil {
		if exist {
			statefulSetClient := kube.clientset.Apps().StatefulSets("clonemap")
			var statefulSetList *apiappsv1.StatefulSetList
			statefulSetList, err = statefulSetClient.List(metav1.ListOptions{})
			var statefulSet apiappsv1.StatefulSet
			for i := range statefulSetList.Items {
				l, ok := statefulSetList.Items[i].Spec.Template.ObjectMeta.Labels["app"]
				if !ok {
					continue
				}
				if l == "mas"+strconv.Itoa(masID)+"agencies" {
					statefulSet = statefulSetList.Items[i]
					err = statefulSetClient.Delete(statefulSet.GetName(), &metav1.DeleteOptions{})
					if err != nil {
						return
					}
				}

			}

			err = kube.deleteHeadlessService(masID)
		} else {
			err = errors.New("StatefulSet does not exist")
		}
	}
	return
}

// newKubeDeployment returns Deployment interface with Kubernetes type
func newKubeDeployment(deplType string) (depl deployment, err error) {
	var temp kubeDeployment
	temp.deplType = deplType
	temp.resLimit = false
	// creates the in-cluster config
	config, err := rest.InClusterConfig()
	if err == nil {
		// creates the clientset
		clientset, err := kubernetes.NewForConfig(config)
		if err == nil {
			temp.config = config
			temp.clientset = clientset
		}
	}
	resLim := os.Getenv("CLONEMAP_RESOURCE_LIMITATION")
	if resLim == "YES" {
		temp.resLimit = true
	}
	depl = &temp
	return
}

// existStatefulSet checks if stateful set already exists and returns boolean value
func (kube *kubeDeployment) existStatefulSet(masID int) (exist bool, err error) {
	exist = false
	servicesClient := kube.clientset.Core().Services("clonemap")
	var services *apicorev1.ServiceList
	services, err = servicesClient.List(metav1.ListOptions{})
	if err == nil {
		// check if headless service for mas is already running
		for i := range services.Items {
			if services.Items[i].GetObjectMeta().GetName() == "mas"+strconv.Itoa(masID)+"agencies" {
				exist = true
				break
			}
		}
	}
	return
}

// get CPUCapcity returns the number of available cpus
func (kube *kubeDeployment) getCPUCapacity() (cap int, err error) {
	// TODO get capacity from k8s cluster
	if kube.deplType == "minikube" {
		cap = 2
	} else {
		cap = 56
	}
	return
}

// createHeadlessService creates a new headless service for all agencies in a mas
func (kube *kubeDeployment) createHeadlessService(masID int) (err error) {
	servicesClient := kube.clientset.Core().Services("clonemap")
	serv := &apicorev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name: "mas" + strconv.Itoa(masID) + "agencies",
		},
		Spec: apicorev1.ServiceSpec{
			ClusterIP: "None",
			Selector: map[string]string{
				"app": "mas" + strconv.Itoa(masID) + "agencies",
			},
			Ports: []apicorev1.ServicePort{
				apicorev1.ServicePort{
					Port: 10000,
					Name: "mas" + strconv.Itoa(masID) + "agencies",
				},
			},
		},
	}
	_, err = servicesClient.Create(serv)
	return
}

// createStatefulSet creates a new headless service and a statefulset for agencies if it has not
// been created yet
func (kube *kubeDeployment) createStatefulSet(masID int, imID int, image string, pullSecret string,
	numAgencies int, loggingEnv string, mqttEnv string,
	dfEnv string) (err error) {
	// Pod Spec
	podSpec := apicorev1.PodSpec{
		Containers: []apicorev1.Container{
			apicorev1.Container{
				Name:            "mas" + strconv.Itoa(masID) + "agencies",
				Image:           image,
				ImagePullPolicy: "Always",
				Ports: []apicorev1.ContainerPort{
					apicorev1.ContainerPort{
						ContainerPort: 10000,
						Name:          "mas" + strconv.Itoa(masID) + "agencies",
					},
				},
				Env: []apicorev1.EnvVar{
					apicorev1.EnvVar{
						Name:  "CLONEMAP_LOGGING",
						Value: loggingEnv,
					},
					apicorev1.EnvVar{
						Name:  "CLONEMAP_MQTT",
						Value: mqttEnv,
					},
					apicorev1.EnvVar{
						Name:  "CLONEMAP_DF",
						Value: dfEnv,
					},
					apicorev1.EnvVar{
						Name:  "CLONEMAP_LOG_LEVEL",
						Value: os.Getenv("CLONEMAP_LOG_LEVEL"),
					},
				},
				LivenessProbe: &apicorev1.Probe{
					Handler: apicorev1.Handler{
						HTTPGet: &apicorev1.HTTPGetAction{
							Path: "/api/agency",
							Port: intstr.IntOrString{
								IntVal: 10000,
							},
						},
					},
					InitialDelaySeconds: 20,
					TimeoutSeconds:      30,
				},
			},
		},
	}
	if kube.resLimit {
		// determine cpu requests
		var cpureq resource.Quantity
		var cpulim resource.Quantity
		var cpureqDez int
		ftemp, _ := kube.getCPUCapacity()
		cpureqDez = int((float32(ftemp) - 8.0) / float32(numAgencies) * 1000)
		if cpureqDez > 3000 {
			cpureqDez = 3000
		}
		cpureq, err = resource.ParseQuantity(strconv.Itoa(cpureqDez) + "m")
		if err != nil {
			return
		}
		cpulim, err = resource.ParseQuantity(strconv.Itoa(int(2*float32(cpureqDez))) + "m")
		if err != nil {
			return
		}
		podSpec.Containers[0].Resources = apicorev1.ResourceRequirements{
			Requests: apicorev1.ResourceList{
				apicorev1.ResourceCPU: cpureq,
			},
			Limits: apicorev1.ResourceList{
				apicorev1.ResourceCPU: cpulim,
			},
		}
	}
	if pullSecret != "" {
		podSpec.ImagePullSecrets = []apicorev1.LocalObjectReference{
			apicorev1.LocalObjectReference{
				Name: pullSecret,
			},
		}
	}

	// statefulset
	if err == nil {
		replicas := int32(numAgencies)
		statefulset := &apiappsv1.StatefulSet{
			ObjectMeta: metav1.ObjectMeta{
				Name: "mas-" + strconv.Itoa(masID) + "-im-" + strconv.Itoa(imID) + "-agency",
			},
			Spec: apiappsv1.StatefulSetSpec{
				Replicas: &replicas,
				Selector: &metav1.LabelSelector{
					MatchLabels: map[string]string{
						"app": "mas" + strconv.Itoa(masID) + "agencies",
					},
				},
				ServiceName: "mas" + strconv.Itoa(masID) + "agencies",
				Template: apicorev1.PodTemplateSpec{
					ObjectMeta: metav1.ObjectMeta{
						Labels: map[string]string{
							"app": "mas" + strconv.Itoa(masID) + "agencies",
						},
					},
					Spec: podSpec,
				},
				PodManagementPolicy: apiappsv1.ParallelPodManagement,
			},
		}
		statefulsetclient := kube.clientset.Apps().StatefulSets("clonemap")
		_, err = statefulsetclient.Create(statefulset)
	}
	return
}

// scaleStatefulSet updates an existing stateful set with the given number of replicas
func (kube *kubeDeployment) scaleStatefulSet(masID int, replicasDelta int) (err error) {
	statefulSetClient := kube.clientset.Apps().StatefulSets("clonemap")
	var statefulSetList *apiappsv1.StatefulSetList
	statefulSetList, err = statefulSetClient.List(metav1.ListOptions{})
	var statefulSet apiappsv1.StatefulSet
	for i := range statefulSetList.Items {
		if statefulSetList.Items[i].GetName() == "mas-"+strconv.Itoa(masID)+"-agency" {
			statefulSet = statefulSetList.Items[i]
			replicas := *statefulSet.Spec.Replicas + int32(replicasDelta)
			statefulSet.Spec.Replicas = &replicas
			_, err = statefulSetClient.Update(&statefulSet)
			break
		}
	}
	return
}

// deleteHeadlessService deletes the headless service for a mas
func (kube *kubeDeployment) deleteHeadlessService(masID int) (err error) {
	servicesClient := kube.clientset.Core().Services("clonemap")
	err = servicesClient.Delete("mas"+strconv.Itoa(masID)+"agencies", &metav1.DeleteOptions{})
	return
}

// deleteStatefulSet deletes the statefulset and corresponding headless service
func (kube *kubeDeployment) deleteStatefulSet(masID int) (err error) {
	statefulsetclient := kube.clientset.Apps().StatefulSets("clonemap")
	err = statefulsetclient.Delete("mas-"+strconv.Itoa(masID)+"-agency", &metav1.DeleteOptions{})
	return
}
