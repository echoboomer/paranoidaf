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

	"github.com/echoboomer/paranoidaf/pkg/common"
	log "github.com/sirupsen/logrus"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

// Ignored Kubernetes Namespaces
var nsFilter = []string{"kube-system", "kube-node-lease", "kube-public"}

// UGPrepOptions acts as a container to hold information passed into the process
type UGPrepOptions struct {
	ClusterName string
	Namespace   string
}

// Check carries out various processes related to the check
func Check(config *rest.Config, clientset kubernetes.Interface, o *UGPrepOptions) {
	// Friendly info
	log.Infof("Checking cluster %s...", o.ClusterName)

	// If a Namespace is passed in, we only check that one
	// Otherwise, we check all non-filtered Namespaces
	var nsList []string
	if o.Namespace != "" {
		nsList = append(nsList, o.Namespace)
	} else {
		// Build a list of Namespaces
		namespaces, err := clientset.CoreV1().Namespaces().List(context.TODO(), metav1.ListOptions{})
		if err != nil {
			log.Errorf("Error building list of Namespaces: %s", err)
		}

		// Build a list of eligible Namespaces
		for _, ns := range namespaces.Items {
			nsList = append(nsList, ns.Name)
		}

		// Filter out unwanted Namespaces
		// kube-public, kube-system, kube-node-lease
		for _, filteredNS := range nsFilter {
			nsList = common.DeleteFromSlice(nsList, filteredNS)
		}
	}

	// Establish qualifying Deployments as the basis for the check
	deployments := returnEligibleDeployments(clientset, nsList)
	if len(deployments) == 0 {
		log.Infof("Didn't find any Deployments in these Namespaces: %s", nsList)
	}

	// Run it
	checkDeployments(clientset, deployments)
}
