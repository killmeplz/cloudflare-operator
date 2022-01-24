/*
Copyright 2022.

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
	"github.com/cloudflare/cloudflare-go"
	"k8s.io/apimachinery/pkg/api/errors"
	"time"

	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	ctrllog "sigs.k8s.io/controller-runtime/pkg/log"

	cfv1alpha1 "github.com/containeroo/cloudflare-operator/api/v1alpha1"
)

// ZoneReconciler reconciles a Zone object
type ZoneReconciler struct {
	client.Client
	Scheme *runtime.Scheme
	Cf     *cloudflare.API
}

//+kubebuilder:rbac:groups=cf.containeroo.ch,resources=zones,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=cf.containeroo.ch,resources=zones/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=cf.containeroo.ch,resources=zones/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the Zone object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.10.0/pkg/reconcile
func (r *ZoneReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := ctrllog.FromContext(ctx)

	// Fetch the Zone instance
	instance := &cfv1alpha1.Zone{}
	err := r.Get(ctx, req.NamespacedName, instance)
	if err != nil {
		if errors.IsNotFound(err) {
			log.Info("Zone resource not found. Ignoring since object must be deleted.")
			return ctrl.Result{}, nil
		}
		log.Error(err, "Failed to get Zone resource")
		return ctrl.Result{}, err
	}

	if r.Cf.APIKey == "" {
		log.Info("Cloudflare account not ready. Retrying in 5 seconds")
		return ctrl.Result{RequeueAfter: 5 * time.Second}, err
	}

	_, err = r.Cf.ZoneDetails(ctx, instance.Spec.ID)
	if err != nil {
		instance.Status.Phase = "Failed"
		instance.Status.Message = err.Error()
		err = r.Status().Update(ctx, instance)
		if err != nil {
			log.Error(err, "Failed to update Zone status")
			return ctrl.Result{}, err
		}
		return ctrl.Result{}, err
	}

	if instance.Status.Phase != "Active" {
		instance.Status.Phase = "Active"
		instance.Status.Message = ""
		err = r.Status().Update(ctx, instance)
		if err != nil {
			log.Error(err, "Failed to update Zone status")
			return ctrl.Result{}, err
		}
	}

	// TODO: Implement logic if default settings have changed

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *ZoneReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&cfv1alpha1.Zone{}).
		Complete(r)
}