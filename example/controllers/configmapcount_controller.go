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

	"github.com/go-logr/logr"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/source"

	"github.com/imikushin/controllers-af/function"
	"github.com/imikushin/controllers-af/reconciler"

	sillyv1alpha1 "github.com/imikushin/controllers-af/example/api/v1alpha1"
)

// ConfigMapCountReconciler reconciles a ConfigMapCount object
type ConfigMapCountReconciler struct {
	Client client.Client
	Log    logr.Logger
	Scheme *runtime.Scheme
}

// +kubebuilder:rbac:groups=silly.example.org,resources=configmapcounts,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=silly.example.org,resources=configmapcounts/status,verbs=get;update;patch

func (r *ConfigMapCountReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&sillyv1alpha1.ConfigMapCount{}).
		Watches(&source.Kind{Type: &corev1.ConfigMap{}}, reconciler.EnqueueRequestsForQuery(r.Client, r.Log, configMapCountsInTheSameNS)).
		Complete(reconciler.New(r.Client, r.Log, &sillyv1alpha1.ConfigMapCount{}, Reconcile))
}

func configMapCountsInTheSameNS(obj client.Object) function.Query {
	return function.Query{
		Namespace: obj.GetNamespace(),
		Type:      &sillyv1alpha1.ConfigMapCountList{},
	}
}

func Reconcile(_ context.Context, object client.Object, getDetails function.GetDetails) (*function.Effects, error) {
	cmc := object.(*sillyv1alpha1.ConfigMapCount)

	cmSelector, err := labelSelector(cmc)
	if err != nil {
		return nil, err
	}

	cmList := getDetails(function.Query{
		Namespace: cmc.Namespace,
		Type:      &corev1.ConfigMapList{},
		Selector:  cmSelector,
	}).(*corev1.ConfigMapList)

	cmCount := len(cmList.Items)

	for cmList.Continue != "" {
		cmList = getDetails(function.Query{
			Namespace: cmc.Namespace,
			Type:      &corev1.ConfigMapList{},
			Selector:  cmSelector,
			Options:   []client.ListOption{client.Continue(cmList.Continue)},
		}).(*corev1.ConfigMapList)

		cmCount += len(cmList.Items)
	}

	cmc.Status = sillyv1alpha1.ConfigMapCountStatus{
		ConfigMaps: cmCount,
	}

	return &function.Effects{Persists: []client.Object{cmc}}, nil
}

func labelSelector(cmcInputObject *sillyv1alpha1.ConfigMapCount) (labels.Selector, error) {
	if cmcInputObject.Spec.Selector == nil {
		return labels.Everything(), nil
	}
	return metav1.LabelSelectorAsSelector(cmcInputObject.Spec.Selector)
}
