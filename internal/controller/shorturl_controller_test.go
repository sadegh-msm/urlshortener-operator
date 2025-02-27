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
	"fmt"
	"net/http"
	"net/http/httptest"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	urlshortenerv1 "urlshortener-operator/api/v1"
)

var _ = Describe("ShortURL Controller", func() {
	Context("When reconciling a resource", func() {
		const resourceName = "test-resource"
		ctx := context.Background()

		typeNamespacedName := types.NamespacedName{
			Name:      resourceName,
			Namespace: "default",
		}
		shorturl := &urlshortenerv1.ShortURL{}

		BeforeEach(func() {
			By("creating the custom resource for the Kind ShortURL")
			err := k8sClient.Get(ctx, typeNamespacedName, shorturl)
			if err != nil && apierrors.IsNotFound(err) {
				resource := &urlshortenerv1.ShortURL{
					ObjectMeta: metav1.ObjectMeta{
						Name:      resourceName,
						Namespace: "default",
					},
					Spec: urlshortenerv1.ShortURLSpec{
						TargetURL: "http://google.com",
					},
				}
				Expect(k8sClient.Create(ctx, resource)).To(Succeed())
			}
		})

		AfterEach(func() {
			resource := &urlshortenerv1.ShortURL{}
			err := k8sClient.Get(ctx, typeNamespacedName, resource)
			Expect(err).NotTo(HaveOccurred())

			By("cleaning up the ShortURL resource")
			Expect(k8sClient.Delete(ctx, resource)).To(Succeed())
		})

		It("should successfully reconcile the resource and update status", func() {
			fakeServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				switch r.URL.Path {
				case "/valid/testShort":
					fmt.Fprintf(w, `{"is_valid": "true"}`)
				case "/count/testShort":
					fmt.Fprintf(w, `{"click_count": 5}`)
				default:
					http.NotFound(w, r)
				}
			}))
			defer fakeServer.Close()
			ShortenerServiceURL = fakeServer.URL

			By("triggering reconciliation")
			controllerReconciler := &ShortURLReconciler{
				Client: k8sClient,
				Scheme: k8sClient.Scheme(),
			}
			_, err := controllerReconciler.Reconcile(ctx, reconcile.Request{
				NamespacedName: typeNamespacedName,
			})
			Expect(err).NotTo(HaveOccurred())

			time.Sleep(15 * time.Second)

			updatedShortURL := &urlshortenerv1.ShortURL{}
			Expect(k8sClient.Get(ctx, typeNamespacedName, updatedShortURL)).To(Succeed())
			Expect(updatedShortURL.Status.ShortPath).To(Equal("testShort"))
			Expect(updatedShortURL.Status.ClickCount).To(Equal(5))
			Expect(updatedShortURL.Status.IsValid).To(Equal("true"))
		})
	})
})
