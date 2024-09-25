/*
Copyright 2024.

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

	"github.com/miekg/dns"
	networkingv1 "k8s.io/api/networking/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/log"

	policyv1 "github.com/icefed/domain-egress-policy/api/v1"
	rdns "github.com/icefed/domain-egress-policy/internal/dns"
)

// DomainEgressPolicyReconciler reconciles a DomainEgressPolicy object
type DomainEgressPolicyReconciler struct {
	client.Client
	Scheme *runtime.Scheme

	DNSResolver *rdns.Resolver
}

// +kubebuilder:rbac:groups=policy.icefed.io,resources=domainegresspolicies,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=policy.icefed.io,resources=domainegresspolicies/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=policy.icefed.io,resources=domainegresspolicies/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the DomainEgressPolicy object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.18.4/pkg/reconcile
func (r *DomainEgressPolicyReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	_ = log.FromContext(ctx)

	egressPolicy := &policyv1.DomainEgressPolicy{}
	if err := r.Get(ctx, req.NamespacedName, egressPolicy); err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	// if the DomainEgressPolicy is being deleted, remove the finalizer
	if egressPolicy.DeletionTimestamp != nil {
		controllerutil.RemoveFinalizer(egressPolicy, DomainEgressPolicyFinalizer)
		if err := r.Update(ctx, egressPolicy); err != nil {
			return ctrl.Result{Requeue: true}, err
		}
		return ctrl.Result{}, nil
	}

	// if the DomainEgressPolicy doesn't have the finalizer, add it
	if !controllerutil.ContainsFinalizer(egressPolicy, DomainEgressPolicyFinalizer) {
		controllerutil.AddFinalizer(egressPolicy, DomainEgressPolicyFinalizer)
		if err := r.Update(ctx, egressPolicy); err != nil {
			return ctrl.Result{Requeue: true}, err
		}
	}
	// create the NetworkPolicy if it doesn't exist, or update it if needed
	var networkPolicy networkingv1.NetworkPolicy
	networkPolicyNamespacedName := types.NamespacedName{Name: generateNetworkPolicyName(egressPolicy), Namespace: egressPolicy.Namespace}
	if err := r.Get(ctx, networkPolicyNamespacedName, &networkPolicy); err != nil {
		if errors.IsNotFound(err) {
			if err := r.Create(ctx, r.buildNetworkPolicy(egressPolicy)); err != nil {
				return ctrl.Result{Requeue: true}, err
			}
		}
		return ctrl.Result{Requeue: true}, err
	}
	// TODO: update
	updatedNetworkPolicy := r.buildNetworkPolicy(egressPolicy)
	if err := r.Update(ctx, updatedNetworkPolicy); err != nil {
		return ctrl.Result{Requeue: true}, err
	}

	return ctrl.Result{}, nil
}

const (
	// DomainEgressPolicyFinalizer is the finalizer for the DomainEgressPolicy object
	DomainEgressPolicyFinalizer = "policy.icefed.io/domainegresspolicy-finalizer"

	// LabelKeyDomainEgressPolicy is the bool value label for the NetworkPolicy object, that is managed by DomainEgressPolicy controller.
	LabelKeyDomainEgressPolicy = "policy.icefed.io/domainegresspolicy"
	// AnnoKeyDomainEgressPolicyRef is the annotation for the NetworkPolicy object, which points to the DomainEgressPolicy object
	AnnoKeyDomainEgressPolicyRef = "policy.icefed.io/domainegresspolicy-ref"
)

func (r *DomainEgressPolicyReconciler) buildNetworkPolicy(egressPolicy *policyv1.DomainEgressPolicy) *networkingv1.NetworkPolicy {
	policy := &networkingv1.NetworkPolicy{
		ObjectMeta: metav1.ObjectMeta{
			Name:      generateNetworkPolicyName(egressPolicy),
			Namespace: egressPolicy.Namespace,
			Labels: map[string]string{
				LabelKeyDomainEgressPolicy: "true",
			},
			Annotations: map[string]string{
				AnnoKeyDomainEgressPolicyRef: egressPolicy.Name,
			},
			OwnerReferences: []metav1.OwnerReference{
				*metav1.NewControllerRef(egressPolicy, policyv1.GroupVersion.WithKind("DomainEgressPolicy")),
			},
		},
		Spec: networkingv1.NetworkPolicySpec{
			PodSelector: egressPolicy.Spec.PodSelector,
			PolicyTypes: []networkingv1.PolicyType{
				networkingv1.PolicyTypeEgress,
			},
		},
	}

	for _, rule := range egressPolicy.Spec.DomainEgress {
		egress := networkingv1.NetworkPolicyEgressRule{
			Ports: rule.Ports,
			To: []networkingv1.NetworkPolicyPeer{
				// add loopback to avoid peers empty and matches all destinations
				{
					IPBlock: &networkingv1.IPBlock{
						CIDR: "127.0.0.1/32",
					},
				},
			},
		}
		// add ips from domain resolver
		for _, domain := range rule.Domains {
			rrs, err := r.DNSResolver.Resolve(dns.TypeA, domain)
			if err != nil {
				log.Log.Error(err, "failed to resolve domain", "domain", domain)
			}
			for _, rr := range rrs {
				switch v := rr.(type) {
				case *dns.A:
					egress.To = append(egress.To, networkingv1.NetworkPolicyPeer{
						IPBlock: &networkingv1.IPBlock{
							CIDR: v.A.String() + "/32",
						},
					})
				case *dns.CNAME:
				}
			}
		}
		policy.Spec.Egress = append(policy.Spec.Egress, egress)
	}

	return policy
}

// generateNetworkPolicyName generates the name of the network policy, based on the name of the DomainEgressPolicy.
// egressPolicy.Name + "-dep", egressPolicy name length should be less than 60.
func generateNetworkPolicyName(egressPolicy *policyv1.DomainEgressPolicy) string {
	return egressPolicy.Name + "-dep"
}

// SetupWithManager sets up the controller with the Manager.
func (r *DomainEgressPolicyReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&policyv1.DomainEgressPolicy{}).
		Complete(r)
}
