/*
Copyright 2022 season.

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

	appv1beta1 "github.com/SeasonPilot/opdemo/api/v1beta1"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

// MyAppReconciler reconciles a MyApp object
type MyAppReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=app.season.io,resources=myapps,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=app.season.io,resources=myapps/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=app.season.io,resources=myapps/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the MyApp object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.13.0/pkg/reconcile
func (r *MyAppReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	l := log.FromContext(ctx)
	l.Info("Reconciling MyApp")

	// Fetch the MyApp instance
	myApp := appv1beta1.MyApp{}
	err := r.Get(ctx, req.NamespacedName, &myApp)
	if err != nil {
		l.Error(err, "Get MyApp instance ERR")

		// MyApp was deleted, Ignore
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	deploy := &appsv1.Deployment{}
	deploy.Name = myApp.Name
	deploy.Namespace = myApp.Namespace
	or, err := ctrl.CreateOrUpdate(ctx, r.Client, deploy, func() error {
		mutateDeploy(myApp, deploy)
		return ctrl.SetControllerReference(&myApp, deploy, r.Scheme)
	})
	if err != nil {
		return ctrl.Result{}, err
	}

	l.Info("OperationResult", "deployment", or)

	svc := &corev1.Service{}
	svc.Name = myApp.Name
	svc.Namespace = myApp.Namespace
	or, err = ctrl.CreateOrUpdate(ctx, r.Client, svc, func() error {
		mutateService(myApp, svc)
		return ctrl.SetControllerReference(&myApp, svc, r.Scheme)
	})
	if err != nil {
		return ctrl.Result{}, err
	}

	l.Info("OperationResult", "svc", or)

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *MyAppReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&appv1beta1.MyApp{}).
		Owns(&appsv1.Deployment{}).
		Owns(&corev1.Service{}).
		Complete(r)
}
