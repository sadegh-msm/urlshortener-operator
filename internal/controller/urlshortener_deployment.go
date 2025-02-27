package controller

import (
	"context"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/utils/pointer"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// ensureShortenerDeployment creates the Deployment for the shortener API if it does not exist.
func (r *ShortURLReconciler) ensureShortenerDeployment(ctx context.Context) error {
	deployment := &appsv1.Deployment{}
	err := r.Get(ctx, client.ObjectKey{Name: "urlshortener-api", Namespace: "urlshortener-operator-system"}, deployment)
	if err != nil && apierrors.IsNotFound(err) {
		deployment = &appsv1.Deployment{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "urlshortener-api",
				Namespace: "urlshortener-operator-system",
				Labels:    map[string]string{"app": "urlshortener-api"},
			},
			Spec: appsv1.DeploymentSpec{
				Replicas: pointer.Int32Ptr(1),
				Selector: &metav1.LabelSelector{
					MatchLabels: map[string]string{"app": "urlshortener-api"},
				},
				Template: corev1.PodTemplateSpec{
					ObjectMeta: metav1.ObjectMeta{
						Labels: map[string]string{"app": "urlshortener-api"},
					},
					Spec: corev1.PodSpec{
						Containers: []corev1.Container{
							{
								Name:  "urlshortener-api",
								Image: "docker.io/sadegh81/url-shortener:v2",
								Ports: []corev1.ContainerPort{
									{
										ContainerPort: 8080,
									},
								},
							},
						},
					},
				},
			},
		}
		if err := r.Create(ctx, deployment); err != nil {
			return err
		}
	} else if err != nil {
		return err
	}
	return nil
}
