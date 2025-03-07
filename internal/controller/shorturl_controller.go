/*
Copyright 2025.

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

package controller

import (
	"context"
	"log"
	"time"

	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	urlshortenerv1 "urlshortener-operator/api/v1"
)

// ShortURLReconciler reconciles a ShortURL object
type ShortURLReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

var ShortenerServiceURL = "http://urlshortener-api.urlshortener-operator-system.svc.cluster.local:8080"

// +kubebuilder:rbac:groups=urlshortener.shortener.io,resources=shorturls,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=apps,resources=deployments,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups="",resources=services,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=urlshortener.shortener.io,resources=shorturls/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=urlshortener.shortener.io,resources=shorturls/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the ShortURL object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.20.2/pkg/reconcile
func (r *ShortURLReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log.Println("Started reconcilation loop")
	err := r.ensureShortenerDeployment(ctx)
	if err != nil {
		return ctrl.Result{}, err
	}
	err = r.ensureShortenerService(ctx)
	if err != nil {
		return ctrl.Result{}, err
	}

	var shortURL urlshortenerv1.ShortURL
	if err := r.Get(ctx, req.NamespacedName, &shortURL); err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	if shortURL.Status.ShortPath == "" {
		shortenPath, err := shortenURL(shortURL.Spec.TargetURL, shortURL.Spec.ExpireAt)
		if err != nil {
			return ctrl.Result{}, err
		}

		shortURL.Status.ShortPath = shortenPath
		shortURL.Status.ClickCount = 0
		shortURL.Status.IsValid = "unknown"
		if err := r.Status().Update(ctx, &shortURL); err != nil {
			return ctrl.Result{}, err
		}
	}

	clickCnt, err := getClickCount(shortURL.Status.ShortPath)
	if err != nil {
		return ctrl.Result{}, err
	}

	shortURL.Status.ClickCount = clickCnt

	valid, err := checkURLValidity(shortURL.Status.ShortPath)
	if err != nil {
		return ctrl.Result{}, err
	}
	shortURL.Status.IsValid = valid

	if err := r.Status().Update(ctx, &shortURL); err != nil {
		return ctrl.Result{}, err
	}

	return ctrl.Result{RequeueAfter: 10 * time.Second}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *ShortURLReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&urlshortenerv1.ShortURL{}).
		Named("shorturl").
		Complete(r)
}
