/*
Copyright © 2021 Scott Hawkins <scott@echoboomer.net>

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
package eval

import (
	"context"
	"fmt"
	"strings"

	"github.com/kyokomi/emoji/v2"
	log "github.com/sirupsen/logrus"
	appsv1 "k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

// deploymentDescription returns information regarding each Deployment
type deploymentDescription struct {
	labels    string
	name      string
	namespace string
	replicas  int32
	selectors map[string]string
}

// hpaDescription returns information regarding a given HorizontalPodAutoscaler
type hpaDescription struct {
	application string
	max         int32
	min         int32
	name        string
	namespace   string
}

// pdbDescription returns information regarding a given PodDisruptionBudget
type pdbDescription struct {
	application        string
	availabilityConfig map[string]int32
	name               string
	namespace          string
}

// buildDeploymentDescription returns a struct with information regarding each Deployment
// with information that can be used to calculate risk
func buildDeploymentDescription(d appsv1.Deployment) *deploymentDescription {
	// Make the labels a string
	var container []string
	var stringsAsLabels string
	for k, v := range d.Spec.Selector.MatchLabels {
		concatLabel := fmt.Sprintf("%s=%s", k, v)
		container = append(container, concatLabel)
		stringsAsLabels = strings.Join(container, ",")
	}
	return &deploymentDescription{
		labels:    stringsAsLabels,
		name:      d.Name,
		namespace: d.Namespace,
		replicas:  *d.Spec.Replicas,
		selectors: d.Spec.Selector.MatchLabels,
	}
}

// checkDeployments procsses a list of Deployments and verifies their configurations as they
// relate to high availability and resiliency
func checkDeployments(clientset kubernetes.Interface, deployments []appsv1.Deployment) {
	fmt.Println()
	for _, d := range deployments {
		// Build a struct for each Deployment
		dep := buildDeploymentDescription(d)
		_, err := emoji.Printf(":package: %s\n", dep.name)
		if err != nil {
			log.Fatal(err)
		}
		fmt.Printf("----------------------------------------------------------------------\n")
		_, err = emoji.Printf(":information_source:	Current replicas: %v\n", dep.replicas)
		if err != nil {
			log.Fatal(err)
		}
		_, err = emoji.Printf(":information_source:	Matching resources using labels: %v\n", dep.labels)
		if err != nil {
			log.Fatal(err)
		}

		// Check for HorizontalPodAutoscaler
		hpa := returnHorizontalPodAutoscalers(clientset, dep.name, dep.namespace, dep.labels)
		if hpa.name == "" {
			_, err = emoji.Printf(":warning:	Could not find a HorizontalPodAutoscaler using labels %s. Double check the labels. The Deployment replica count is likely static. Read more here: https://kubernetes.io/docs/tasks/run-application/horizontal-pod-autoscale/\n", dep.labels)
			if err != nil {
				log.Fatal(err)
			}
			if dep.replicas != 0 {
				if dep.replicas < 2 {
					_, err = emoji.Printf(":point_right:	Suggestion - verify that the minimum replica count is not set for a single replica, enable a HorizontalPodAutoscaler, and set minReplicas to at least 2.\n")
					if err != nil {
						log.Fatal(err)
					}
					_, err = emoji.Printf(":point_right:	Suggestion - add and enable a PodDisruptionBudget with at least a maxUnavailable less than configured min replicas.\n")
					if err != nil {
						log.Fatal(err)
					}
				} else {
					_, err = emoji.Printf(":white_check_mark:	Current replica count is at least 2. This helps keep this application up during events like rollouts and upgrades.\n")
					if err != nil {
						log.Fatal(err)
					}
				}
			} else {
				_, err = emoji.Printf(":warning:	Couldn't figure out spec.replicas.")
				if err != nil {
					log.Fatal(err)
				}
			}
		} else {
			_, err = emoji.Printf(":white_check_mark:	This app has a HorizontalPodAutoscaler with %v min replicas and %v max replicas.\n", hpa.min, hpa.max)
			if err != nil {
				log.Fatal(err)
			}
		}

		// Check for PodDisruptionBudget
		pdb := returnPodDisruptionBudgets(clientset, dep.name, dep.namespace, dep.labels)
		if pdb.name == "" {
			_, err = emoji.Printf(":warning:	This app does not have a PodDisruptionBudget. This application could experience interruptions during rollouts, upgrades, etc. Read more here: https://kubernetes.io/docs/concepts/workloads/pods/disruptions/\n")
			if err != nil {
				log.Fatal(err)
			}
			_, err = emoji.Printf(":point_right:	Suggestion - enable a PodDisruptionBudget.\n")
			if err != nil {
				log.Fatal(err)
			}
		} else {
			_, err = emoji.Printf(":white_check_mark:	This app has a PodDisruptionBudget configured with: %v\n", pdb.availabilityConfig)
			if err != nil {
				log.Fatal(err)
			}
		}
		fmt.Println()
	}
}

// returnEligibleDeployments accepts a []string containing target Namespaces
// and returns a []v1.Deployment with related Deployment spec
func returnEligibleDeployments(clientset kubernetes.Interface, nsList []string) []appsv1.Deployment {
	// Restrict which Pods to return
	// We should return all that we own - i.e. not kube-system, etc.
	deploymentListOptions := metav1.ListOptions{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Deployment",
			APIVersion: "v1",
		},
	}

	// Add each v1.Deployment to the returned list
	var deploymentList []appsv1.Deployment
	for _, ns := range nsList {
		deployments, err := clientset.AppsV1().Deployments(ns).List(context.TODO(), deploymentListOptions)
		if err != nil {
			log.Errorf("Error: %s", err)
		}
		deploymentList = append(deploymentList, deployments.Items...)
	}
	return deploymentList
}

// returnHorizontalPodAutoscalers returns a list of HorizontalPodAutoscalers for a given
// application
func returnHorizontalPodAutoscalers(clientset kubernetes.Interface, application string, ns string, labelsAsString string) *hpaDescription {
	hpas, err := clientset.AutoscalingV1().HorizontalPodAutoscalers(ns).List(context.TODO(), metav1.ListOptions{
		TypeMeta: metav1.TypeMeta{
			Kind:       "HorizontalPodAutoscaler",
			APIVersion: "autoscaling/v1",
		},
		LabelSelector: labelsAsString,
	})
	if err != nil {
		log.Errorf("Error: %s", err)
	}

	if len(hpas.Items) > 0 {
		hpa := hpas.Items[0]
		return &hpaDescription{
			application: application,
			max:         hpa.Spec.MaxReplicas,
			min:         *hpa.Spec.MinReplicas,
			name:        hpa.Name,
			namespace:   hpa.Namespace,
		}
	}
	return &hpaDescription{}
}

// returnPodDisruptionBudgets returns a list of PodDisruptionBudgets for a given
// application
func returnPodDisruptionBudgets(clientset kubernetes.Interface, application string, ns string, labelsAsString string) *pdbDescription {
	pdbs, err := clientset.PolicyV1().PodDisruptionBudgets(ns).List(context.TODO(), metav1.ListOptions{
		TypeMeta: metav1.TypeMeta{
			Kind:       "PodDisruptionBudget",
			APIVersion: "policy/v1",
		},
		LabelSelector: labelsAsString,
	})
	if err != nil {
		log.Errorf("Error: %s", err)
	}

	if len(pdbs.Items) > 0 {
		pdb := pdbs.Items[0]
		var avcfg map[string]int32
		if pdb.Spec.MaxUnavailable != nil && pdb.Spec.MinAvailable == nil {
			avcfg = map[string]int32{
				"maxUnavailable": pdb.Spec.MaxUnavailable.IntVal,
			}
		} else if pdb.Spec.MinAvailable != nil && pdb.Spec.MaxUnavailable == nil {
			avcfg = map[string]int32{
				"minAvailable": pdb.Spec.MinAvailable.IntVal,
			}
		} else {
			avcfg = map[string]int32{
				"maxUnavailable": pdb.Spec.MaxUnavailable.IntVal,
				"minAvailable":   pdb.Spec.MinAvailable.IntVal,
			}
		}
		return &pdbDescription{
			application:        application,
			availabilityConfig: avcfg,
			name:               pdb.Name,
			namespace:          pdb.Namespace,
		}
	}
	return &pdbDescription{}
}
