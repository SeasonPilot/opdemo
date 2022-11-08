package controllers

import (
	appv1beta1 "github.com/SeasonPilot/opdemo/api/v1beta1"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func mutateDeploy(app appv1beta1.MyApp, deploy *appsv1.Deployment) {
	var (
		labels = map[string]string{"app": app.Name}
	)

	deploy.Spec = appsv1.DeploymentSpec{
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

func mutateService(app appv1beta1.MyApp, svc *corev1.Service) {
	var (
		labels = map[string]string{"app": app.Name}
	)

	// fixme:
	//  svc = &corev1.Service{}  这样写是错误的，这是给 svc 重新初始化，不是更新字段
	svc.Spec = corev1.ServiceSpec{
		Ports:    app.Spec.Ports,
		Selector: labels,
		Type:     corev1.ServiceTypeNodePort,
	}
}
