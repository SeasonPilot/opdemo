package controllers

import (
	appv1beta1 "github.com/SeasonPilot/opdemo/api/v1beta1"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

func newDeploy(app appv1beta1.MyApp) *appsv1.Deployment {
	var (
		labels = map[string]string{"app": app.Name}
	)

	return &appsv1.Deployment{
		TypeMeta: metav1.TypeMeta{
			APIVersion: appsv1.SchemeGroupVersion.String(),
			Kind:       "Deployment",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      app.Name,
			Namespace: app.Namespace,
			Labels:    labels,
			OwnerReferences: []metav1.OwnerReference{
				*metav1.NewControllerRef(&app, schema.GroupVersionKind{ // fixme: 可以使用 NewControllerRef 函数
					Group:   appv1beta1.GroupVersion.Group,
					Version: appv1beta1.GroupVersion.Version,
					Kind:    app.Kind,
				}),
			},
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: app.Spec.Size,
			Selector: &metav1.LabelSelector{
				MatchLabels: labels,
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: labels,
				},
				Spec: corev1.PodSpec{
					Containers: newContainers(app),
				},
			},
		},
	}
}

func newContainers(app appv1beta1.MyApp) (containers []corev1.Container) {
	var containerPorts []corev1.ContainerPort
	for _, port := range app.Spec.Ports {
		var containerPort corev1.ContainerPort
		containerPort.ContainerPort = port.TargetPort.IntVal
		containerPorts = append(containerPorts, containerPort)
	}
	return []corev1.Container{
		{
			Name:            app.Name,
			Image:           app.Spec.Image,
			Resources:       app.Spec.Resources,
			Ports:           containerPorts,
			ImagePullPolicy: corev1.PullIfNotPresent,
			Env:             app.Spec.Envs,
		},
	}
}

func newService(app appv1beta1.MyApp) *corev1.Service {
	var (
		labels = map[string]string{"app": app.Name}
	)

	return &corev1.Service{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Service", // fixme: 大写开头,小写开头也可以正常创建出资源
			APIVersion: corev1.SchemeGroupVersion.String(),
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      app.Name,
			Namespace: app.Namespace,
			OwnerReferences: []metav1.OwnerReference{
				*metav1.NewControllerRef(&app, schema.GroupVersionKind{
					Group:   appv1beta1.GroupVersion.Group,
					Version: appv1beta1.GroupVersion.Version,
					Kind:    app.Kind,
				})},
		},
		Spec: corev1.ServiceSpec{
			Ports:    app.Spec.Ports,
			Selector: labels,
			Type:     corev1.ServiceTypeNodePort,
		},
	}
}
