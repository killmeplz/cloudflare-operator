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
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"strings"
	"time"

	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	ctrllog "sigs.k8s.io/controller-runtime/pkg/log"

	cfv1alpha1 "github.com/containeroo/cloudflare-operator/api/v1alpha1"
)

// AccountReconciler reconciles a Account object
type AccountReconciler struct {
	client.Client
	Scheme *runtime.Scheme
	Cf     *cloudflare.API
}

//+kubebuilder:rbac:groups=cf.containeroo.ch,resources=accounts,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=cf.containeroo.ch,resources=accounts/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=cf.containeroo.ch,resources=accounts/finalizers,verbs=update
//+kubebuilder:rbac:groups=core,resources=secrets,verbs=get;list

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the Account object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.10.0/pkg/reconcile
func (r *AccountReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := ctrllog.FromContext(ctx)

	// Fetch the Account instance
	instance := &cfv1alpha1.Account{}
	err := r.Get(ctx, req.NamespacedName, instance)
	if err != nil {
		if errors.IsNotFound(err) {
			log.Info("Account resource not found. Ignoring since object must be deleted.")
			return ctrl.Result{}, nil
		}
		log.Error(err, "Failed to get Account resource")
		return ctrl.Result{}, err
	}

	// Fetch the secret
	secret := &v1.Secret{}
	err = r.Get(ctx, client.ObjectKey{Namespace: instance.Spec.GlobalApiKey.SecretRef.Namespace, Name: instance.Spec.GlobalApiKey.SecretRef.Name}, secret)
	if err != nil {
		log.Error(err, "Failed to get Secret resource", "Secret.Namespace", instance.Spec.GlobalApiKey.SecretRef.Namespace, "Secret.Name", instance.Spec.GlobalApiKey.SecretRef.Name)
		return ctrl.Result{}, err
	}
	apiKey := string(secret.Data["apiKey"])
	cf, err := cloudflare.New(apiKey, instance.Spec.Email)
	if err != nil {
		log.Error(err, "Failed to create Cloudflare client")
		instance.Status.Phase = "Failed"
		instance.Status.Message = err.Error()
		err = r.Status().Update(ctx, instance)
		if err != nil {
			log.Error(err, "Failed to update Account status")
			return ctrl.Result{}, err
		}
		return ctrl.Result{}, err
	}
	*r.Cf = *cf

	zones, err := r.Cf.ListZones(ctx)
	if err != nil {
		log.Error(err, "Failed to create Cloudflare client. Retrying in 30 seconds")
		instance.Status.Phase = "Failed"
		instance.Status.Message = err.Error()
		err = r.Status().Update(ctx, instance)
		if err != nil {
			log.Error(err, "Failed to update Account status")
			return ctrl.Result{}, err
		}
		return ctrl.Result{RequeueAfter: 30 * time.Second}, err
	}

	if instance.Status.Phase != "Ready" || instance.Status.Message != "" {
		instance.Status.Phase = "Ready"
		instance.Status.Message = ""
		err = r.Status().Update(ctx, instance)
		if err != nil {
			log.Error(err, "Failed to update Account status")
			return ctrl.Result{}, err
		}
	}

	// Fetch all Zone objects
	zonesList := &cfv1alpha1.ZoneList{}
	err = r.List(ctx, zonesList, client.InNamespace(instance.Namespace))
	if err != nil {
		log.Error(err, "Failed to list Zone resources")
		return ctrl.Result{}, err
	}

	// Check if all zones are present
	for _, zone := range zones {
		found := false
		for _, z := range zonesList.Items {
			if z.Spec.ID == zone.ID {
				found = true
				break
			}
		}
		if !found {
			log.Info("Zone not found. Creating", "Zone.Name", zone.Name, "Zone.ID", zone.ID)
			trueVar := true
			z := &cfv1alpha1.Zone{
				ObjectMeta: metav1.ObjectMeta{
					Name: strings.ReplaceAll(zone.Name, ".", "-"),
					Labels: map[string]string{
						"app.kubernetes.io/managed-by": "cloudflare-operator",
						"app.kubernetes.io/created-by": "cloudflare-operator",
					},
					OwnerReferences: []metav1.OwnerReference{
						{
							APIVersion:         "cf.containeroo.ch/v1alpha1",
							Kind:               "Account",
							Name:               instance.Name,
							UID:                instance.UID,
							Controller:         &trueVar,
							BlockOwnerDeletion: &trueVar,
						},
					},
				},
				Spec: cfv1alpha1.ZoneSpec{
					Name:            zone.Name,
					ID:              zone.ID,
					DefaultSettings: instance.Spec.DefaultSettings,
				},
				Status: cfv1alpha1.ZoneStatus{
					Phase: "Pending",
				},
			}
			err = r.Create(ctx, z)
			if err != nil {
				log.Error(err, "Failed to create Zone resource", "Zone.Name", zone.Name, "Zone.ID", zone.ID)
				return ctrl.Result{}, err
			}
		}
	}

	for _, zone := range zones {
		found := false
		for _, z := range instance.Status.Zones {
			if z.ID == zone.ID {
				found = true
				break
			}
		}
		if !found {
			instance.Status.Zones = append(instance.Status.Zones, cfv1alpha1.AccountStatusZones{
				ID:   zone.ID,
				Name: zone.Name,
			})
			err := r.Status().Update(ctx, instance)
			if err != nil {
				log.Error(err, "Failed to update Account status")
				return ctrl.Result{}, err
			}
		}
	}

	// TODO: Cleanup Zones absent in Cloudflare

	// TODO: Cleanup DNSRecords absent in Cloudflare

	// TODO: Implement logic if default settings have changed

	return ctrl.Result{RequeueAfter: instance.Spec.Interval.Duration}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *AccountReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&cfv1alpha1.Account{}).
		Complete(r)
}