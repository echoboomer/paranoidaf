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
package kubetools

import (
	"os"
	"path/filepath"

	"github.com/echoboomer/paranoidaf/pkg/common"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/afero"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
)

var fs afero.Fs = afero.NewOsFs()

// CreateKubeConfig Instantiates an object to reference the local machine's kubeconfig
// and provides two ways to run against a target Kubernetes cluster -
// inCluster=true uses the ServiceAccount credentials provided to a Pod where paranoidaf may run -
// inCluster=false (default) uses ~/.kube/config -
// returns config, clientset, and a string identifying the location of kubeconfig
func CreateKubeConfig(inCluster bool) (*rest.Config, kubernetes.Interface, string) {
	// inCluster is either true or false
	// If it's true, we pull Kubernetes API authentication from Pod SA
	// If it's false, we use local machine settings
	if inCluster {
		config, err := rest.InClusterConfig()
		if err != nil {
			panic(err.Error())
		}

		clientset, err := kubernetes.NewForConfig(config)
		if err != nil {
			panic(err.Error())
		}

		return config, clientset, "in-cluster"
	}

	// Set path to kubeconfig
	kubeconfig := ReturnKubeConfigPath()

	// Check to make sure kubeconfig actually exists
	if common.FileExists(fs, kubeconfig) {
		log.Infof("kubeconfig exists at %s", kubeconfig)
	} else {
		log.Fatalf("kubeconfig doesn't exist, we looked here: %s", kubeconfig)
	}

	// Only proceed if kubeconfig exists
	// Show what path was set for kubeconfig
	log.Infof("Setting kubeconfig to: %s", kubeconfig)

	// Build configuration instance from the provided config file
	config, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
	if err != nil {
		log.Fatalf("Unable to locate kubeconfig file - checked path: %s", kubeconfig)
	}

	// Create clientset, which is used to run operations against the API
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		panic(err)
	}

	return config, clientset, kubeconfig
}

// ReturnKubeConfigPath generates the path in the filesystem to kubeconfig
func ReturnKubeConfigPath() string {
	var kubeconfig string
	// We expect kubeconfig to be available at ~/.kube/config
	// However, sometimes some people may use the env var $KUBECONFIG
	// to set the path to the active one - we will switch on that here
	if os.Getenv("KUBECONFIG") != "" {
		kubeconfig = os.Getenv("KUBECONFIG")
	} else {
		kubeconfig = filepath.Join(homedir.HomeDir(), ".kube", "config")
	}
	return kubeconfig
}
