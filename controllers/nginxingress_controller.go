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

package controllers

import (
	"context"
	"fmt"
	"time"

	"github.com/ingoxx/ingress-nginx-operator/controllers/internal"
	"github.com/ingoxx/ingress-nginx-operator/pkg/common"
	"github.com/ingoxx/ingress-nginx-operator/pkg/constants"
	"github.com/ingoxx/ingress-nginx-operator/pkg/operatorCli"
	v12 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/api/networking/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/tools/record"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/builder"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"
)

// NginxIngressReconciler reconciles a NginxIngress object
type NginxIngressReconciler struct {
	client.Client
	clientSet   common.K8sClientSet
	operatorCli common.OperatorClientSet
	Scheme      *runtime.Scheme
	recorder    record.EventRecorder
}

//+kubebuilder:rbac:groups=ingress.ingress-k8s.io,resources=nginxingresses,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=ingress.ingress-k8s.io,resources=nginxingresses/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=ingress.ingress-k8s.io,resources=nginxingresses/finalizers,verbs=update

//+kubebuilder:rbac:groups=cert-manager.io,resources=certificates,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=cert-manager.io,resources=issuers,verbs=get;list;watch;create;update;patch;delete

//+kubebuilder:rbac:groups=core,resources=services,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=core,resources=configmaps,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=core,resources=events,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=core,resources=endpoints,verbs=get;list;watch
//+kubebuilder:rbac:groups=core,resources=secrets,verbs=get;list;watch;create;update;patch;delete

// +kubebuilder:rbac:groups=apps,resources=deployments,verbs=get;list;watch;create;update;patch;delete

//+kubebuilder:rbac:groups=networking.k8s.io,resources=ingresses,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=networking.k8s.io,resources=ingresses/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=networking.k8s.io,resources=ingressclasses,verbs=get;list;watch;create;update;patch;delete

// Reconcile is part of the internal kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.13.0/pkg/reconcile
func (r *NginxIngressReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	fmt.Println(">>>> test github ci status")
	if err := internal.NewCrdNginxController(ctx, r.clientSet, r.operatorCli, r.recorder).Start(req); err != nil {
		return ctrl.Result{RequeueAfter: 15 * time.Second}, nil
	}

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *NginxIngressReconciler) SetupWithManager(mgr ctrl.Manager, clientSet common.K8sClientSet) error {
	r.clientSet = clientSet
	r.operatorCli = operatorCli.NewOperatorClientImp(mgr.GetClient())
	r.recorder = mgr.GetEventRecorderFor(constants.RecorderKey)

	enqueueIngress := handler.EnqueueRequestsFromMapFunc(func(obj client.Object) []reconcile.Request {
		ingList := &v1.IngressList{}
		listOpts := make([]client.ListOption, 0, 1)
		if ns := obj.GetNamespace(); ns != "" {
			listOpts = append(listOpts, client.InNamespace(ns))
		}

		if err := r.List(context.Background(), ingList, listOpts...); err != nil {
			return nil
		}

		reqs := make([]reconcile.Request, 0, len(ingList.Items))
		for _, ing := range ingList.Items {
			reqs = append(reqs, reconcile.Request{NamespacedName: client.ObjectKeyFromObject(&ing)})
		}

		return reqs
	})

	certObj := &unstructured.Unstructured{}
	certObj.SetGroupVersionKind(schema.GroupVersionKind{
		Group:   "cert-manager.io",
		Version: "v1",
		Kind:    "Certificate",
	})

	issuerObj := &unstructured.Unstructured{}
	issuerObj.SetGroupVersionKind(schema.GroupVersionKind{
		Group:   "cert-manager.io",
		Version: "v1",
		Kind:    "Issuer",
	})

	ctx := context.Background()
	if _, err := mgr.GetCache().GetInformer(ctx, &corev1.ConfigMap{}); err != nil {
		return fmt.Errorf("failed to start ConfigMap informer: %w", err)
	}
	if _, err := mgr.GetCache().GetInformer(ctx, &v12.Deployment{}); err != nil {
		return fmt.Errorf("failed to start Deployment informer: %w", err)
	}
	if _, err := mgr.GetCache().GetInformer(ctx, &corev1.Service{}); err != nil {
		return fmt.Errorf("failed to start Service informer: %w", err)
	}
	if _, err := mgr.GetCache().GetInformer(ctx, &corev1.Secret{}); err != nil {
		return fmt.Errorf("failed to start Secret informer: %w", err)
	}
	if _, err := mgr.GetCache().GetInformer(ctx, certObj); err != nil {
		return fmt.Errorf("failed to start Certificate informer: %w", err)
	}
	if _, err := mgr.GetCache().GetInformer(ctx, issuerObj); err != nil {
		return fmt.Errorf("failed to start Issuer informer: %w", err)
	}

	if err := mgr.GetCache().IndexField(ctx, &corev1.ConfigMap{}, "metadata.name", func(obj client.Object) []string { return []string{obj.GetName()} }); err != nil {
		return fmt.Errorf("failed to create ConfigMap index: %w", err)
	}

	if err := mgr.GetCache().IndexField(ctx, &v12.Deployment{}, "metadata.name", func(obj client.Object) []string { return []string{obj.GetName()} }); err != nil {
		return fmt.Errorf("failed to create Deployment index: %w", err)
	}
	if err := mgr.GetCache().IndexField(ctx, &corev1.Service{}, "metadata.name", func(obj client.Object) []string { return []string{obj.GetName()} }); err != nil {
		return fmt.Errorf("failed to create Service index: %w", err)
	}
	if err := mgr.GetCache().IndexField(ctx, &corev1.Secret{}, "metadata.name", func(obj client.Object) []string { return []string{obj.GetName()} }); err != nil {
		return fmt.Errorf("failed to create Secret index: %w", err)
	}

	return ctrl.NewControllerManagedBy(mgr).
		For(&v1.Ingress{}).
		Watches(&source.Kind{Type: &v12.Deployment{}}, enqueueIngress, builder.WithPredicates(predicate.ResourceVersionChangedPredicate{})).
		Watches(&source.Kind{Type: &corev1.Service{}}, enqueueIngress, builder.WithPredicates(predicate.ResourceVersionChangedPredicate{})).
		Watches(&source.Kind{Type: &corev1.ConfigMap{}}, enqueueIngress, builder.WithPredicates(predicate.ResourceVersionChangedPredicate{})).
		Watches(&source.Kind{Type: &corev1.Secret{}}, enqueueIngress, builder.WithPredicates(predicate.ResourceVersionChangedPredicate{})).
		Watches(&source.Kind{Type: certObj}, enqueueIngress, builder.WithPredicates(predicate.ResourceVersionChangedPredicate{})).
		Watches(&source.Kind{Type: issuerObj}, enqueueIngress, builder.WithPredicates(predicate.ResourceVersionChangedPredicate{})).
		Complete(r)
}
