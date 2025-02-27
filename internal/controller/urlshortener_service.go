package controller

import (
	"context"

	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// ensureShortenerService creates the Service for the shortener API if it does not exist.
func (r *ShortURLReconciler) ensureShortenerService(ctx context.Context) error {
	service := &corev1.Service{}
	err := r.Get(ctx, client.ObjectKey{Name: "urlshortener-api", Namespace: "urlshortener-operator-system"}, service)
	if err != nil && apierrors.IsNotFound(err) {
		service = &corev1.Service{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "urlshortener-api",
				Namespace: "urlshortener-operator-system",
			},
			Spec: corev1.ServiceSpec{
				Selector: map[string]string{"app": "urlshortener-api"},
				Ports: []corev1.ServicePort{
					{
						Port:       8080,
						TargetPort: intstr.FromInt(8080),
						Protocol:   corev1.ProtocolTCP,
					},
				},
				Type: corev1.ServiceTypeClusterIP,
			},
		}
		if err := r.Create(ctx, service); err != nil {
			return err
		}
	} else if err != nil {
		return err
	}
	return nil
}
