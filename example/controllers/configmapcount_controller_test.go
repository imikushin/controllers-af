/*
Copyright 2021 Ivan Mikushin

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package controllers

import (
	"context"
	"testing"

	"github.com/go-logr/logr"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/envtest/printer"
	// +kubebuilder:scaffold:imports

	"github.com/imikushin/controllers-af/example/api/v1alpha1"
	"github.com/imikushin/controllers-af/function"
)

func TestAPIs(t *testing.T) {
	RegisterFailHandler(Fail)

	RunSpecsWithDefaultAndCustomReporters(t,
		"Controller Suite",
		[]Reporter{printer.NewlineReporter{}})
}

var _ = Describe("Reconcile", func() {
	var cmcReconciler *ConfigMapCountReconciler

	BeforeEach(func() {
		cmcReconciler = &ConfigMapCountReconciler{Log: logr.Discard()}
	})

	When("no ConfigMaps exist in the same namespace", func() {
		expectedCMs := &corev1.ConfigMapList{
			Items: []corev1.ConfigMap{{}, {}}, // len() == 2
		}

		getDetails := func(query function.Query) runtime.Object {
			return expectedCMs
		}

		It("should set .status.configMaps to 0", func() {
			inputCMC := &v1alpha1.ConfigMapCount{}

			effects, err := cmcReconciler.Reconcile(context.TODO(), inputCMC, getDetails)
			Expect(err).ToNot(HaveOccurred())

			Expect(effects).ToNot(BeNil())
			Expect(effects.Persists).To(HaveLen(1))
			Expect(effects.Persists[0].(*v1alpha1.ConfigMapCount).Status.ConfigMaps).To(Equal(len(expectedCMs.Items)))
		})
	})
})
