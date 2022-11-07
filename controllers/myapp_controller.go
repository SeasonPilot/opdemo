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
	"reflect"

	appv1beta1 "github.com/SeasonPilot/opdemo/api/v1beta1"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/json"
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

	annoKey := "data"

	// Fetch the MyApp instance
	myApp := appv1beta1.MyApp{}
	err := r.Get(ctx, req.NamespacedName, &myApp)
	if err != nil {
		l.Error(err, "Get MyApp instance ERR")

		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	// 没有获取到 deployment
	deploy := &appsv1.Deployment{}
	err = r.Get(ctx, req.NamespacedName, deploy) // fixme: 这里应该要判断 err!=nil, err==nil 说明已经获取到了 deployment
	if err != nil && errors.IsNotFound(err) {    // 其他 err 情况不处理
		l.Info("deployment Not Found")

		// 创建关联资源
		// 1. 创建 Deploy
		deploy = newDeploy(myApp)
		err = r.Create(ctx, deploy)
		if err != nil {
			l.Error(err, "Create deployment err")

			return ctrl.Result{}, err
		}

		// 2. 创建 service
		service := newService(myApp)
		err = r.Create(ctx, service)
		if err != nil {
			return ctrl.Result{}, err
		}

		// 3. 关联 annotation  // fixme: 是更新 CR 的 annotation，不是 deployment。将 CR 的 spec 写入到 annotation
		data, err := json.Marshal(myApp.Spec)
		if err != nil {
			return ctrl.Result{}, err
		}

		myApp.Annotations = map[string]string{annoKey: string(data)} // 不管 annotation map 是否为 nil 都赋值
		err = r.Update(ctx, &myApp)
		if err != nil {
			return ctrl.Result{}, err
		}

		// fixme: 这里要返回，不然就往下执行了。每个分支都需要返回
		return ctrl.Result{}, nil
	}

	// 获取到 deployment
	// 获取 cr 的 annotation
	oldMyAppSpec := appv1beta1.MyAppSpec{}
	err = json.Unmarshal([]byte(myApp.Annotations[annoKey]), &oldMyAppSpec)
	if err != nil {
		return ctrl.Result{}, err
	}

	l.Info("获取到 deployment")

	// 对比 annotation 查看 CR 是否有更新
	if !reflect.DeepEqual(myApp.Spec, oldMyAppSpec) { // fixme: 应该通过结构体来对比，不是对比字符串
		l.Info("CR 有更新")

		// 更新关联资源
		deploy.Spec = newDeploy(myApp).Spec
		err = r.Update(ctx, deploy) // fixme: 不是创建新的资源，应该是更新旧的对象
		if err != nil {
			return ctrl.Result{}, err
		}

		l.Info("更新 deployment 成功")

		oldSvc := &corev1.Service{}
		err = r.Get(ctx, req.NamespacedName, oldSvc)
		if err != nil {
			return ctrl.Result{}, err
		}
		// fixme: 需要指定 ClusterIP 为之前的，不然更新会报错。
		//   试了下，更新的时候好像也没有报错.  暂不修改
		oldSvc.Spec = newService(myApp).Spec
		err = r.Update(ctx, oldSvc)
		if err != nil {
			return ctrl.Result{}, err
		}

		l.Info("更新 service 成功")

		return ctrl.Result{}, nil
	}

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
