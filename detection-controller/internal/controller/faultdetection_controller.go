/*
Copyright 2025.

Licensed under the the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing,
software distributed under the License is distributed on an
"AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
either express or implied. See the License for the specific
language governing permissions and limitations under the License.
*/

package controller

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	detectv1 "github.com/phuongbac/detection-controller/api/v1alpha1"
)

// FaultDetectionReconciler reconciles a FaultDetection object
type FaultDetectionReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=detect.failure-recovery.io,resources=faultdetections,verbs=get;list;watch;update;patch
//+kubebuilder:rbac:groups=detect.failure-recovery.io,resources=faultdetections/status,verbs=update;patch
//+kubebuilder:rbac:groups=detect.failure-recovery.io,resources=detectiontemplates,verbs=get;list;watch
//+kubebuilder:rbac:groups="",resources=nodes,verbs=get;list;watch

func (r *FaultDetectionReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := log.FromContext(ctx)

	// 1. Get FaultDetection CR
	var fd detectv1.FaultDetection
	if err := r.Get(ctx, req.NamespacedName, &fd); err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	// 2. Get DetectionTemplate
	var tmpl detectv1.DetectionTemplate
	if err := r.Get(ctx, client.ObjectKey{Name: fd.Spec.TemplateRef}, &tmpl); err != nil {
		logger.Error(err, "unable to fetch DetectionTemplate", "template", fd.Spec.TemplateRef)
		return ctrl.Result{}, nil
	}

	results := []detectv1.Result{}
	anomaly := false
	reason := ""

	// reset NodeResults each reconcile
	fd.Status.NodeResults = []detectv1.NodeResult{}

	// --- Option B: API-based detection ---
	if tmpl.Spec.APIVersion != "" && tmpl.Spec.Kind != "" && tmpl.Spec.FieldPath != "" {
		if tmpl.Spec.Scope == "Node" && fd.Spec.Target.Name == "" {
			var nodeList unstructured.UnstructuredList
			nodeList.SetAPIVersion("v1")
			nodeList.SetKind("NodeList")

			if err := r.List(ctx, &nodeList); err != nil {
				anomaly = true
				reason = "Failed to list nodes"
			} else {
				for _, node := range nodeList.Items {
					// Special handling for Node Ready
					if tmpl.Spec.FieldPath == "status.conditions[Ready].status" {
						conditions, found, _ := unstructured.NestedSlice(node.Object, "status", "conditions")
						if !found {
							anomaly = true
							fd.Status.NodeResults = append(fd.Status.NodeResults, detectv1.NodeResult{
								NodeName: node.GetName(),
								Ok:       false,
								Message:  "status.conditions not found",
							})
							continue
						}

						readyStatus := ""
						for _, c := range conditions {
							if cond, ok := c.(map[string]interface{}); ok {
								if cond["type"] == "Ready" {
									if s, ok := cond["status"].(string); ok {
										readyStatus = s
									}
								}
							}
						}

						if readyStatus == "" {
							anomaly = true
							fd.Status.NodeResults = append(fd.Status.NodeResults, detectv1.NodeResult{
								NodeName: node.GetName(),
								Ok:       false,
								Message:  "Ready condition not found",
							})
						} else if readyStatus != tmpl.Spec.Expected {
							anomaly = true
							reason = fmt.Sprintf("Node %s Expected Ready=%s but got %s", node.GetName(), tmpl.Spec.Expected, readyStatus)
							fd.Status.NodeResults = append(fd.Status.NodeResults, detectv1.NodeResult{
								NodeName: node.GetName(),
								Ok:       false,
								Message:  reason,
							})
						} else {
							fd.Status.NodeResults = append(fd.Status.NodeResults, detectv1.NodeResult{
								NodeName: node.GetName(),
								Ok:       true,
								Message:  "Node Ready",
							})
						}
					} else {
						// fallback generic path
						parts := strings.Split(tmpl.Spec.FieldPath, ".")
						actual, found, _ := unstructured.NestedString(node.Object, parts...)
						if !found {
							anomaly = true
							msg := fmt.Sprintf("Field %s not found", tmpl.Spec.FieldPath)
							fd.Status.NodeResults = append(fd.Status.NodeResults, detectv1.NodeResult{
								NodeName: node.GetName(),
								Ok:       false,
								Message:  msg,
							})
							continue
						}
						if actual != tmpl.Spec.Expected {
							anomaly = true
							msg := fmt.Sprintf("Expected %s=%s but got %s", tmpl.Spec.FieldPath, tmpl.Spec.Expected, actual)
							fd.Status.NodeResults = append(fd.Status.NodeResults, detectv1.NodeResult{
								NodeName: node.GetName(),
								Ok:       false,
								Message:  msg,
							})
						} else {
							fd.Status.NodeResults = append(fd.Status.NodeResults, detectv1.NodeResult{
								NodeName: node.GetName(),
								Ok:       true,
								Message:  "OK",
							})
						}
					}
				}
			}
		} else {
			// Single target object
			u := &unstructured.Unstructured{}
			u.SetAPIVersion(tmpl.Spec.APIVersion)
			u.SetKind(tmpl.Spec.Kind)

			key := client.ObjectKey{
				Name:      fd.Spec.Target.Name,
				Namespace: fd.Spec.Target.Namespace,
			}

			if err := r.Get(ctx, key, u); err != nil {
				anomaly = true
				reason = "Target resource not found or unreachable"
			} else {
				parts := strings.Split(tmpl.Spec.FieldPath, ".")
				actual, found, _ := unstructured.NestedString(u.Object, parts...)
				if !found {
					anomaly = true
					reason = fmt.Sprintf("Field %s not found in resource", tmpl.Spec.FieldPath)
				} else if actual != tmpl.Spec.Expected {
					anomaly = true
					reason = fmt.Sprintf("Expected %s=%s but got %s", tmpl.Spec.FieldPath, tmpl.Spec.Expected, actual)
				}
			}
		}
		// --- Option A: Prometheus-based detection ---
	} else if tmpl.Spec.PrometheusAPI != "" && len(tmpl.Spec.Queries) > 0 {
		for _, q := range tmpl.Spec.Queries {
			value, err := queryPrometheus(tmpl.Spec.PrometheusAPI, q.Query)
			if err != nil {
				logger.Error(err, "failed querying prometheus", "metric", q.Metric)
				continue
			}

			results = append(results, detectv1.Result{
				Metric: q.Metric,
				Value:  fmt.Sprintf("%f", value),
			})

			if tmpl.Spec.Rule != "" && strings.Contains(tmpl.Spec.Rule, q.Metric) {
				threshold := parseThreshold(tmpl.Spec.Rule)
				if value > threshold {
					anomaly = true
					reason = fmt.Sprintf("metric %s value %f exceeded threshold %f", q.Metric, value, threshold)
				}
			}
		}
	}

	// 4. Optional ML check
	if tmpl.Spec.ML != nil && tmpl.Spec.ML.Endpoint != "" {
		mlResult, err := callMLModel(tmpl.Spec.ML.Endpoint, results)
		if err != nil {
			logger.Error(err, "failed calling ML model")
		} else if mlResult {
			anomaly = true
			reason = "ML model detected anomaly"
		}
	}

	// 5. Update status
	now := metav1.Now()
	fd.Status.LastRun = &now
	fd.Status.Results = results
	fd.Status.Anomalous = anomaly
	fd.Status.Reason = reason

	if anomaly {
		fd.Status.Triggered = true
		fd.Status.TriggerMsg = "Anomaly detected - printing instead of triggering"
		logger.Info("Anomaly detected!", "reason", reason, "nodes", fd.Status.NodeResults)
	}

	if err := r.Status().Update(ctx, &fd); err != nil {
		return ctrl.Result{}, err
	}

	// 6. Requeue
	return ctrl.Result{RequeueAfter: tmpl.Spec.Interval.Duration}, nil
}

func (r *FaultDetectionReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&detectv1.FaultDetection{}).
		Complete(r)
}

// -------------------- Helper Functions --------------------

func queryPrometheus(api, query string) (float64, error) {
	url := fmt.Sprintf("%s/api/v1/query?query=%s", api, query)
	resp, err := http.Get(url)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()

	body, _ := ioutil.ReadAll(resp.Body)
	var data map[string]interface{}
	if err := json.Unmarshal(body, &data); err != nil {
		return 0, err
	}

	results := data["data"].(map[string]interface{})["result"].([]interface{})
	if len(results) == 0 {
		return 0, fmt.Errorf("no data returned for query %s", query)
	}
	valStr := results[0].(map[string]interface{})["value"].([]interface{})[1].(string)
	return strconv.ParseFloat(valStr, 64)
}

func parseThreshold(rule string) float64 {
	parts := strings.Split(rule, ">")
	if len(parts) != 2 {
		return 0
	}
	val, _ := strconv.ParseFloat(strings.TrimSpace(parts[1]), 64)
	return val
}

func callMLModel(endpoint string, results []detectv1.Result) (bool, error) {
	jsonBody, _ := json.Marshal(results)
	resp, err := http.Post(endpoint, "application/json", strings.NewReader(string(jsonBody)))
	if err != nil {
		return false, err
	}
	defer resp.Body.Close()

	body, _ := ioutil.ReadAll(resp.Body)
	var respData map[string]interface{}
	if err := json.Unmarshal(body, &respData); err != nil {
		return false, err
	}
	if anomaly, ok := respData["anomaly"].(bool); ok {
		return anomaly, nil
	}
	return false, nil
}
