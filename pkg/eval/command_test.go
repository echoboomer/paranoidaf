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
	"reflect"
	"testing"

	appsv1 "k8s.io/api/apps/v1"
	autoscalingv1 "k8s.io/api/autoscaling/v1"
	policyv1 "k8s.io/api/policy/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/fake"
)

func Test_returnEligibleDeployments(t *testing.T) {
	var replicas int32 = 1

	type args struct {
		clientset kubernetes.Interface
		nsList    []string
	}
	tests := []struct {
		name string
		args args
		want []appsv1.Deployment
	}{
		{
			name: "Matched Deployment objects should be returned",
			args: args{
				clientset: fake.NewSimpleClientset(
					&appsv1.Deployment{
						ObjectMeta: metav1.ObjectMeta{
							Name:      "foo",
							Namespace: "default",
							Labels: map[string]string{
								"app":                        "foo",
								"app.kubernetes.io/instance": "foo",
							},
						},
						Spec: appsv1.DeploymentSpec{
							Replicas: &replicas,
						},
					},
				),
				nsList: []string{"default"},
			},
			want: []appsv1.Deployment{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "foo",
						Namespace: "default",
						Labels: map[string]string{
							"app":                        "foo",
							"app.kubernetes.io/instance": "foo",
						},
					},
					Spec: appsv1.DeploymentSpec{
						Replicas: &replicas,
					},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := returnEligibleDeployments(tt.args.clientset, tt.args.nsList); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("returnEligibleDeployments() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_returnHorizontalPodAutoscalers(t *testing.T) {
	var replicas int32 = 1

	type args struct {
		clientset      kubernetes.Interface
		application    string
		ns             string
		labelsAsString string
	}
	tests := []struct {
		name string
		args args
		want *hpaDescription
	}{
		{
			name: "Matched HorizontalPodAutoscaler objects should be returned",
			args: args{
				clientset: fake.NewSimpleClientset(
					&autoscalingv1.HorizontalPodAutoscaler{
						ObjectMeta: metav1.ObjectMeta{
							Name:      "foo",
							Namespace: "default",
							Labels: map[string]string{
								"app":                        "foo",
								"app.kubernetes.io/instance": "foo",
							},
						},
						Spec: autoscalingv1.HorizontalPodAutoscalerSpec{
							MinReplicas: &replicas,
							MaxReplicas: 5,
						},
					},
				),
				application:    "foo",
				ns:             "default",
				labelsAsString: "app=foo, app.kubernetes.io/instance=foo",
			},
			want: &hpaDescription{
				application: "foo",
				max:         5,
				min:         1,
				name:        "foo",
				namespace:   "default",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := returnHorizontalPodAutoscalers(tt.args.clientset, tt.args.application, tt.args.ns, tt.args.labelsAsString); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("returnHorizontalPodAutoscalers() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_returnPodDisruptionBudgets(t *testing.T) {
	var maxUnavailable intstr.IntOrString = intstr.IntOrString{IntVal: 1}
	var minAvailable intstr.IntOrString = intstr.IntOrString{IntVal: 1}

	type args struct {
		clientset      kubernetes.Interface
		application    string
		ns             string
		labelsAsString string
	}
	tests := []struct {
		name string
		args args
		want *pdbDescription
	}{
		{
			name: "Matched PodDisruptionBudget objects should be returned",
			args: args{
				clientset: fake.NewSimpleClientset(
					&policyv1.PodDisruptionBudget{
						ObjectMeta: metav1.ObjectMeta{
							Name:      "foo",
							Namespace: "default",
							Labels: map[string]string{
								"app":                        "foo",
								"app.kubernetes.io/instance": "foo",
							},
						},
						Spec: policyv1.PodDisruptionBudgetSpec{
							MaxUnavailable: &maxUnavailable,
							MinAvailable:   &minAvailable,
						},
					},
				),
				application:    "foo",
				ns:             "default",
				labelsAsString: "app=foo, app.kubernetes.io/instance=foo",
			},
			want: &pdbDescription{
				application: "foo",
				availabilityConfig: map[string]int32{
					"maxUnavailable": 1,
					"minAvailable":   1,
				},
				name:      "foo",
				namespace: "default",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := returnPodDisruptionBudgets(tt.args.clientset, tt.args.application, tt.args.ns, tt.args.labelsAsString); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("returnPodDisruptionBudgets() = %v, want %v", got, tt.want)
			}
		})
	}
}
