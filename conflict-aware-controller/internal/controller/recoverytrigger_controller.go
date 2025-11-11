/*
Copyright 2025.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
either express or implied. See the License for the specific language
governing permissions and limitations under the License.
*/

package controller

import (
	"context"
	"fmt"
	"time"

	argov1alpha1 "github.com/argoproj/argo-workflows/v3/pkg/apis/workflow/v1alpha1"
	recoveryv1alpha1 "github.com/phuongbac/conflictawareworkflowcontroller/api/v1alpha1"

	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// RecoveryTriggerReconciler reconciles RecoveryTrigger CRs
type RecoveryTriggerReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

// Reconcile executes conflict detection and workflow submission
func (r *RecoveryTriggerReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	var trigger recoveryv1alpha1.RecoveryTrigger
	if err := r.Get(ctx, req.NamespacedName, &trigger); err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	// Fetch all triggers in the same namespace
	var triggerList recoveryv1alpha1.RecoveryTriggerList
	if err := r.List(ctx, &triggerList, client.InNamespace(req.Namespace)); err != nil {
		return ctrl.Result{}, err
	}

	// Detect conflicts
	conflict := detectConflicts(&trigger, triggerList.Items)

	originalState := trigger.Status.State

	switch conflict {
	case "None":
		// If workflow not yet created → submit to Argo
		if trigger.Status.WorkflowName == "" {
			wfName, err := r.submitWorkflow(ctx, &trigger)
			if err != nil {
				return ctrl.Result{}, err
			}
			trigger.Status.State = "Running"
			trigger.Status.Reason = "No conflicts, workflow started"
			trigger.Status.StartedAt = &metav1.Time{Time: time.Now()}
			trigger.Status.WorkflowName = wfName

			fmt.Printf("[Controller] Submitted workflow %s using template %s\n", wfName, trigger.Spec.WorkflowTemplate)
		}

	case "ResourceConflict":
		trigger.Status.State = "Suspended"
		trigger.Status.Reason = "Resource conflict detected"

	case "DependencyConflict":
		trigger.Status.State = "Delayed"
		trigger.Status.Reason = "Dependency conflict detected"

	default:
		trigger.Status.State = "Discarded"
		trigger.Status.Reason = "Workflow discarded as unnecessary"
	}

	// Update status only if changed
	if trigger.Status.State != originalState {
		if err := r.Status().Update(ctx, &trigger); err != nil {
			return ctrl.Result{}, err
		}
	}

	return ctrl.Result{}, nil
}

// submitWorkflow creates an Argo Workflow from WorkflowTemplateRef
func (r *RecoveryTriggerReconciler) submitWorkflow(ctx context.Context, trigger *recoveryv1alpha1.RecoveryTrigger) (string, error) {
	// Build workflow object
	wf := &argov1alpha1.Workflow{
		ObjectMeta: metav1.ObjectMeta{
			GenerateName: fmt.Sprintf("%s-", trigger.Name),
			Namespace:    trigger.Namespace,
		},
		Spec: argov1alpha1.WorkflowSpec{
			WorkflowTemplateRef: &argov1alpha1.WorkflowTemplateRef{
				Name: trigger.Spec.WorkflowTemplate,
			},
		},
	}

	// Create workflow in cluster
	if err := r.Create(ctx, wf); err != nil {
		// if already exists, return existing workflow name
		if apierrors.IsAlreadyExists(err) {
			return wf.Name, nil
		}
		return "", err
	}
	return wf.Name, nil
}

// detectConflicts checks if new trigger overlaps with running ones
func detectConflicts(new *recoveryv1alpha1.RecoveryTrigger, running []recoveryv1alpha1.RecoveryTrigger) string {
	for _, t := range running {
		if t.Status.State == "Running" {
			// Check same resource conflict
			for _, obj1 := range new.Spec.TargetObjects {
				for _, obj2 := range t.Spec.TargetObjects {
					if obj1.Kind == obj2.Kind && obj1.Name == obj2.Name {
						return "ResourceConflict"
					}
				}
			}
			// Check dependency conflict (same failure type)
			if new.Spec.FailureType == t.Spec.FailureType {
				return "DependencyConflict"
			}
		}
	}
	return "None"
}

// SetupWithManager registers controller with manager
func (r *RecoveryTriggerReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&recoveryv1alpha1.RecoveryTrigger{}).
		Complete(r)
}
