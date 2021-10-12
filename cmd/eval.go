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
package cmd

import (
	"log"

	"github.com/echoboomer/paranoidaf/pkg/eval"
	"github.com/echoboomer/paranoidaf/pkg/kubetools"
	"github.com/spf13/cobra"
	"k8s.io/client-go/tools/clientcmd"
)

// evalOptions holds configuration options to pass into the eval package
type evalOptions struct {
	namespace string
}

// evalOpts holds default and customizable values from the command line
var evalOpts *evalOptions = &evalOptions{}

// evalCmd represents the eval command
var evalCmd = &cobra.Command{
	Use:   "eval",
	Short: "Evaluate a Kubernetes cluster's configuration.",
	Long: `Evaluate a Kubernetes cluster's configuration.
	
This command looks specifically at the resiliency of your applications and
assesses their behavior during disruptive events like cluster upgrades or
Node scaling.`,
	Run: func(cmd *cobra.Command, args []string) {
		// Initiate kubeconfig
		config, clientset, _ := kubetools.CreateKubeConfig(false)
		clientconfig, err := clientcmd.NewDefaultClientConfigLoadingRules().Load()
		if err != nil {
			log.Fatalf("Error loading client config: %s", err)
		}

		// Format options
		evalupgradeOpts := &eval.UGPrepOptions{
			ClusterName: clientconfig.Contexts[clientconfig.CurrentContext].Cluster,
			Namespace:   evalOpts.namespace,
		}

		// Start
		eval.Check(config, clientset, evalupgradeOpts)
	},
}

func init() {
	rootCmd.AddCommand(evalCmd)
	// Flags for evalupgrade
	evalCmd.Flags().StringVar(&evalOpts.namespace, "namespace", evalOpts.namespace, "Namespace to check. By default, all Namespaces (except for ones filtered out) are checked.")
}
