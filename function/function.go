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

package function

import (
	"context"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client" // TODO decouple and remove!
)

type Reconcile func(ctx context.Context, object Object, getDetails GetDetails) (*Effects, error)

type Object interface {
	metav1.Object
	runtime.Object
}

type ObjectList interface {
	metav1.ListInterface
	runtime.Object
}

// GetDetails function is provided to the reconciler function. In tests, it's provided by the test.
// The returned result is either a Object or ObjectList
type GetDetails func(query Query) runtime.Object

type ObjectToQuery func(obj Object) Query

// Effects specify the changes intended as results of the reconciler function.
type Effects struct {
	// Persists lists objects to persist: create or update.
	//
	// If an object is being updated, its UID field is recommended to be set (otherwise, the update costs an extra GET
	// request).
	//
	// OwnerReferences of persisted objects should have UID field set. The only exception is when an OwnerReference is to
	// an object being persisted in the same Persists list (and its UID is, obviously, not yet known). In such cases, the
	// owner object should come earlier in the Persists list, as it is processed sequentially.
	Persists []Object

	// Deletes lists objects to delete. Deletes are handled after Persists to allow for necessary preparation, like
	// removing finalizers. It is also processed sequentially.
	Deletes []Object
}

// Query is a generalized API query - for either Get or List. The Type field is required, and MUST be an empty
// instance of a Object (if Name is non-empty) or ObjectList (Selector is an optional
// filter for the list). Options field is an optional list of []client.ListOption. Namespace and Selector values
// override those set by Options.
type Query struct {
	Type      runtime.Object
	Namespace string
	Name      string
	Selector  labels.Selector
	Options   []client.ListOption
}
