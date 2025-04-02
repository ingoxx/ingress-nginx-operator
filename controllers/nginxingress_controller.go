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
	"github.com/ingoxx/ingress-nginx-operator/pkg/interfaces"
	v1 "k8s.io/api/networking/v1"
	"k8s.io/klog/v2"

	ingressv1 "github.com/ingoxx/ingress-nginx-operator/api/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

// NginxIngressReconciler reconciles a NginxIngress object
type NginxIngressReconciler struct {
	client.Client
	clientSet interfaces.K8sClientSet
	Scheme    *runtime.Scheme
}

//+kubebuilder:rbac:groups=ingress.ingress-k8s.io,resources=nginxingresses,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=ingress.ingress-k8s.io,resources=nginxingresses/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=ingress.ingress-k8s.io,resources=nginxingresses/finalizers,verbs=update
//+kubebuilder:rbac:groups=cert-manager.io,resources=certificates,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=cert-manager.io,resources=issuers,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=core,resources=services,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=core,resources=secrets,verbs=get;list;watch;create;update;patch;delete

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the NginxIngress object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.13.0/pkg/reconcile
func (r *NginxIngressReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	_ = log.FromContext(ctx)

	// TODO(user): your logic here
	var ic = new(v1.Ingress)
	if err := r.Get(ctx, req.NamespacedName, ic); err != nil {
		klog.Infof("No ingress resource with name '%s' was found in the namespace '%s'", req.NamespacedName.Namespace, req.NamespacedName.Name)
		return ctrl.Result{}, nil
	}

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *NginxIngressReconciler) SetupWithManager(mgr ctrl.Manager, clientSet interfaces.K8sClientSet) error {
	r.clientSet = clientSet
	return ctrl.NewControllerManagedBy(mgr).
		For(&ingressv1.NginxIngress{}).
		For(&v1.Ingress{}).
		Complete(r)
}
