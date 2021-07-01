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

package reconciler

import (
	"context"
	"reflect"

	"github.com/go-logr/logr"
	"github.com/pkg/errors"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	"github.com/imikushin/controllers-af/function"
)

// Function is your reconciler function producing effects for the object being reconciled.
type Function func(ctx context.Context, object client.Object, getDetails function.GetDetails) (*function.Effects, error)

// New creates a reconcile.Reconciler for your object type and Function.
// `objType` should be an empty client.Object instance.
func New(cl client.Client, logger logr.Logger, objType client.Object, f Function) *reconciler {
	return &reconciler{
		client:  cl,
		logger:  logger,
		objType: objType.DeepCopyObject().(client.Object),
		f:       f,
	}
}

type reconciler struct {
	client  client.Client
	logger  logr.Logger
	objType client.Object
	f       Function
}

// EnqueueRequestsForQuery allows to create a handler.EventHandler by providing a function.ObjectToQuery function.
// It is kind of like a handler.EnqueueRequestsFromMapFunc, but without the boring parts :)
func EnqueueRequestsForQuery(c client.Client, log logr.Logger, toQuery function.ObjectToQuery) handler.EventHandler {
	return handler.EnqueueRequestsFromMapFunc(func(object client.Object) []reconcile.Request {
		query := toQuery(object)

		cache := cache{}
		if _, err := runQuery(context.Background(), c, cache, query); err != nil {
			log.Error(err, "running query to enqueue objects", "query", query)
			return nil
		}

		result := make([]reconcile.Request, 0, len(cache))
		for _, v := range cache {
			result = append(result, reconcile.Request{NamespacedName: types.NamespacedName{
				Namespace: v.GetNamespace(),
				Name:      v.GetName(),
			}})
		}
		return result
	})
}

func (r *reconciler) Reconcile(ctx context.Context, request reconcile.Request) (retRes reconcile.Result, retErr error) {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	obj := r.objType.DeepCopyObject().(client.Object)
	if err := r.client.Get(ctx, request.NamespacedName, obj); err != nil {
		if apierrors.IsNotFound(err) {
			return reconcile.Result{}, nil
		}
		return reconcile.Result{}, err
	}

	cache := cache{obj.GetUID(): obj.DeepCopyObject().(client.Object)}

	defer func() {
		retErr = panicErr(recover(), retErr)
	}()
	effects, err := r.f(ctx, obj, r.getDetails(ctx, cache)) // r.getDetails() panic-wraps an error
	if err != nil || effects == nil {
		return reconcile.Result{}, err
	}

	if err := r.persistObjects(ctx, cache, effects.Persists); err != nil {
		return reconcile.Result{}, err
	}
	if err := r.deleteObjects(ctx, effects.Deletes); err != nil {
		return reconcile.Result{}, err
	}

	return reconcile.Result{}, nil
}

type cache map[types.UID]client.Object

func panicErr(v interface{}, orig error) error {
	if v != nil {
		if err, isErr := v.(error); isErr {
			return err
		}
		panic(v) // v is not an error, re-throw :)
	}
	return orig
}

func (r *reconciler) getDetails(ctx context.Context, cache cache) function.GetDetails {
	return func(query function.Query) runtime.Object {
		result, err := runQuery(ctx, r.client, cache, query)
		if err != nil {
			panic(err)
		}
		return result
	}
}

func (r *reconciler) persistObjects(ctx context.Context, cache cache, objects []client.Object) error {
	created := make(created, len(objects))

	for _, object := range objects {
		object := object

		if object.GetUID() == "" {
			if err := r.create(ctx, created, object); err != nil {
				return err
			}
			continue
		}

		if err := r.patch(ctx, cache, object); err != nil {
			return err
		}
	}

	return nil
}

func (r *reconciler) deleteObjects(ctx context.Context, objects []client.Object) error {
	for _, object := range objects {
		object := object
		if err := r.client.Delete(ctx, object); err != nil {
			return err
		}
	}
	return nil
}

type created map[corev1.ObjectReference]client.Object

func (created created) add(object client.Object) {
	apiVersion, kind := object.GetObjectKind().GroupVersionKind().ToAPIVersionAndKind()
	created[corev1.ObjectReference{
		APIVersion: apiVersion,
		Kind:       kind,
		Namespace:  object.GetNamespace(),
		Name:       object.GetName(),
	}] = object
}

func (r *reconciler) create(ctx context.Context, created created, object client.Object) error {
	if err := fixOwnerRefUIDs(object, created); err != nil {
		return err
	}
	if err := r.client.Create(ctx, object); err != nil {
		return err
	}
	created.add(object)
	return nil
}

func (r *reconciler) patch(ctx context.Context, cache cache, object client.Object) error {
	cached := cache[object.GetUID()]
	if reflect.DeepEqual(cached, object) {
		return nil
	}
	patch := client.MergeFromWithOptions(cached, client.MergeFromWithOptimisticLock{})
	status := object.DeepCopyObject().(client.Object)
	if err := r.client.Patch(ctx, object, patch); err != nil {
		return err
	}
	if err := r.client.Status().Patch(ctx, status, patch); err != nil {
		if !apierrors.IsNotFound(err) {
			return err
		}
	}
	return nil
}

func runQuery(ctx context.Context, c client.Client, cache cache, query function.Query) (runtime.Object, error) {
	if query.Name == "" {
		// get a list
		list, castOK := query.Type.DeepCopyObject().(client.ObjectList)
		if !castOK {
			return nil, errors.Errorf("casting %v to client.ObjectList type", query.Type)
		}
		opts := append(query.Options, client.InNamespace(query.Namespace), selectorOpt(query.Selector))
		if err := c.List(ctx, list, opts...); err != nil {
			return nil, err
		}
		addListToCache(cache, list)
		return list, nil
	}

	// get an object
	object, castOK := query.Type.DeepCopyObject().(client.Object)
	if !castOK {
		return nil, errors.Errorf("casting %v to client.Object type", query.Type)
	}
	if err := c.Get(ctx, client.ObjectKey{Namespace: query.Namespace, Name: query.Name}, object); err != nil {
		if apierrors.IsNotFound(err) {
			return nil, nil
		}
		return nil, err
	}
	if _, exists := cache[object.GetUID()]; !exists {
		cache[object.GetUID()] = object
	}
	return object, nil
}

func selectorOpt(selector labels.Selector) client.ListOption {
	if selector == nil {
		return noopListOption{}
	}
	return client.MatchingLabelsSelector{Selector: selector}
}

type noopListOption struct{}

func (noopListOption) ApplyToList(*client.ListOptions) {}

func addListToCache(cache cache, list client.ObjectList) {
	items := reflect.ValueOf(list).Elem().FieldByName("Items")
	for i := 0; i < items.Len(); i += 1 {
		objValue := items.Index(i)
		object := objValue.Addr().Interface().(client.Object)
		if _, exists := cache[object.GetUID()]; !exists {
			cache[object.GetUID()] = object
		}
	}
}

func fixOwnerRefUIDs(object client.Object, created created) error {
	ownerRefs := object.GetOwnerReferences()
	for i, ownerRef := range ownerRefs {
		if ownerRef.UID == "" {
			createdObject, exists := created[createdKey(ownerRef, object.GetNamespace())]
			if !exists {
				return errors.Errorf("creating %s %s/%s, cannot find ownerRef %+v", object.GetObjectKind().GroupVersionKind(), object.GetNamespace(), object.GetName(), ownerRef)
			}
			ownerRefs[i].UID = createdObject.GetUID()
		}
	}
	object.SetOwnerReferences(ownerRefs)
	return nil
}

func createdKey(ownerRef v1.OwnerReference, namespace string) corev1.ObjectReference {
	return corev1.ObjectReference{
		APIVersion: ownerRef.APIVersion,
		Kind:       ownerRef.Kind,
		Namespace:  namespace,
		Name:       ownerRef.Name,
	}
}
