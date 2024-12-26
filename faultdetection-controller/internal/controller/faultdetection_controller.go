/*
Copyright 2024 DCN

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
    "time"

    "github.com/go-logr/logr"
    examplev1alpha1  "github.com/SSU-DCN/workflow-based-auto-recovery/faultdetection-controller/api/v1alpha1"
    metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
    "sigs.k8s.io/controller-runtime/pkg/client"
    ctrl "sigs.k8s.io/controller-runtime"
)

type FaultDetectionReconciler struct {
    client.Client
    Log logr.Logger
}

func (r *FaultDetectionReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
    log := r.Log.WithValues("faultdetection", req.NamespacedName)

    var fd examplev1alpha1.FaultDetection
    if err := r.Get(ctx, req.NamespacedName, &fd); err != nil {
        log.Error(err, "unable to fetch FaultDetection")
        return ctrl.Result{}, client.IgnoreNotFound(err)
    }

    // Detection logic
    detectedFaults := r.detectFaults(fd.Spec)
    fd.Status.Faults = detectedFaults
    fd.Status.LastChecked = metav1.Now()

    if err := r.Status().Update(ctx, &fd); err != nil {
        log.Error(err, "unable to update FaultDetection status")
        return ctrl.Result{}, err
    }

    return ctrl.Result{RequeueAfter: time.Minute * 5}, nil
}

func (r *FaultDetectionReconciler) detectFaults(spec examplev1alpha1.FaultDetectionSpec) []string {
    if len(spec.Rules) > 0 {
        return []string{"FaultA"}
    }
    return nil
}

func (r *FaultDetectionReconciler) SetupWithManager(mgr ctrl.Manager) error {
    return ctrl.NewControllerManagedBy(mgr).
        For(&examplev1alpha1.FaultDetection{}).
        Complete(r)
}
