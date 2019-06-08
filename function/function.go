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
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// GetDetails function is provided to the reconciler function. In tests, it's provided by the test.
// The returned result is either a client.Object or client.ObjectList
type GetDetails func(query Query) runtime.Object

type ObjectToQuery func(obj client.Object) Query

// Effects specify the changes intended as results of the reconciler function. Persists is the list of objects to
// persist: create or update, depending on whether the object has a UID (which is supposed to be assigned by the
// Kubernetes API). Deletes is the list of objects to delete
type Effects struct {
	Persists []client.Object // Create | Update
	Deletes  []client.Object
}

// Query is a generalized API query - for either Get or List. The Type field is required, and MUST be an empty
// instance of a client.Object (if Name is non-empty) or client.ObjectList (Selector is an optional
// filter for the list).
type Query struct {
	Type      runtime.Object
	Namespace string
	Name      string
	Selector  labels.Selector
}
