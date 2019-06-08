# Controllers AF

A Go library for building Kubernetes Controllers - As Functions

### Why

Building Kubernetes controllers is hard. One of the reasons is: it is very hard to write unit-tests for controllers. 
Why is that? Controllers need to talk to the Kubernetes API all the time, and it takes enormous discipline to mock out 
these interactions. So, what ends up happening: we don't do it, and our controllers become branching spaghetti monsters.

Enter Controllers AF. It is a tiny library (or micro-framework) built on top of [controller-runtime](https://github.com/kubernetes-sigs/controller-runtime)
and is designed to do one thing: separate controller logic from interactions with Kubernetes API. Usage section below illustrates how to use it. We also provide a fully functioning [example](example).

### Usage

```go
package yourcontroller

import (
	"context"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/source"

	"github.com/imikushin/controllers-af/function"
	"github.com/imikushin/controllers-af/reconciler"

	yourapiv1alpha1 "github.com/your/awesome-controller/api/v1alpha1"
)

func SetupWithManager(mgr ctrl.Manager) error {
	theClient := mgr.GetClient()
	log := ctrl.Log.WithName("controllers").WithName("YourObject")

	yourReconciler := reconciler.New(theClient, log, &yourapiv1alpha1.YourObject{}, ReconcileFun)
	
	return ctrl.NewControllerManagedBy(mgr).For(&yourapiv1alpha1.YourObject{}).Complete(yourReconciler)
}

func ReconcileFun(_ context.Context, object client.Object, getDetails function.GetDetails) (*function.Effects, error) {
	yourObject := object.(*yourapiv1alpha1.YourObject)
	
	yourOtherObjects := getDetails(function.Query{
		Namespace: yourObject.Namespace,
		Type:      &yourapiv1alpha1.YourOtherObjectList{},
		Selector:  metav1.LabelSelectorAsSelector(yourObject.Spec.YourOtherObjectSelector),
	}).(*YourOtherObjectList)
	
	// calculate your effects - no need to talk to the API!

	return &function.Effects{
		Persists: []client.Object{create, or, update, these},
		Deletes: []client.Object{things, toDelete},
	}, nil
}
```
