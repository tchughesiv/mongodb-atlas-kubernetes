/*
Copyright 2020 MongoDB.

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

package atlasservice

import (
	"context"

	"go.uber.org/zap"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/source"

	mdbv1 "github.com/mongodb/mongodb-atlas-kubernetes/pkg/api/v1"
	"github.com/mongodb/mongodb-atlas-kubernetes/pkg/api/v1/status"
	"github.com/mongodb/mongodb-atlas-kubernetes/pkg/controller/atlas"
	"github.com/mongodb/mongodb-atlas-kubernetes/pkg/controller/customresource"
	"github.com/mongodb/mongodb-atlas-kubernetes/pkg/controller/statushandler"
	"github.com/mongodb/mongodb-atlas-kubernetes/pkg/controller/watch"
	"github.com/mongodb/mongodb-atlas-kubernetes/pkg/controller/workflow"
)

// AtlasserviceReconciler reconciles a AtlasService object
type AtlasServiceReconciler struct {
	Client client.Client
	watch.ResourceWatcher
	Log         *zap.SugaredLogger
	Scheme      *runtime.Scheme
	AtlasDomain string
	OperatorPod client.ObjectKey
}

// Dev note: duplicate the permissions in both sections below to generate both Role and ClusterRoles

// +kubebuilder:rbac:groups=atlas.mongodb.com,resources=atlasservices,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=atlas.mongodb.com,resources=atlasservices/status,verbs=get;update;patch
// +kubebuilder:rbac:groups="",resources=secrets,verbs=get;list;watch

// +kubebuilder:rbac:groups=atlas.mongodb.com,namespace=default,resources=atlasservices,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=atlas.mongodb.com,namespace=default,resources=atlasservices/status,verbs=get;update;patch
// +kubebuilder:rbac:groups="",namespace=default,resources=secrets,verbs=get;list;watch

func (r *AtlasServiceReconciler) Reconcile(context context.Context, req ctrl.Request) (ctrl.Result, error) {
	_ = context
	log := r.Log.With("atlasservice", req.NamespacedName)

	service := &mdbv1.AtlasService{}
	result := customresource.PrepareResource(r.Client, req, service, log)
	if !result.IsOk() {
		return result.ReconcileResult(), nil
	}
	if service.ConnectionSecretObjectKey() != nil {
		// Note, that we are not watching the global connection secret - seems there is no point in reconciling all
		// the services once that secret is changed
		r.EnsureResourcesAreWatched(req.NamespacedName, "Secret", log, *service.ConnectionSecretObjectKey())
	}
	ctx := customresource.MarkReconciliationStarted(r.Client, service, log)

	log.Infow("-> Starting AtlasService reconciliation", "spec", service.Spec)

	// This update will make sure the status is always updated in case of any errors or successful result
	defer statushandler.Update(ctx, r.Client, service)

	connection, err := atlas.ReadConnection(log, r.Client, r.OperatorPod, service.ConnectionSecretObjectKey())
	if err != nil {
		result := workflow.Terminate(workflow.AtlasCredentialsNotProvided, err.Error())
		ctx.SetConditionFromResult(status.ReadyType, result)
		return result.ReconcileResult(), nil
	}
	ctx.Connection = connection

	atlasClient, err := atlas.Client(r.AtlasDomain, connection, log)
	if err != nil {
		ctx.SetConditionFromResult(status.ReadyType, workflow.Terminate(workflow.Internal, err.Error()))
		return result.ReconcileResult(), nil
	}
	ctx.Client = atlasClient

	svcList, result := r.discoverServices(ctx)
	if !result.IsOk() {
		ctx.SetConditionFromResult(status.ReadyType, result)
		return result.ReconcileResult(), nil
	}

	ctx.
		SetConditionTrue(status.ReadyType).
		EnsureStatusOption(status.AtlasProjectServiceListOption(svcList))

	return ctrl.Result{}, nil
}

// Delete is a no-op
func (r *AtlasServiceReconciler) Delete(e event.DeleteEvent) error {
	return nil
}

func (r *AtlasServiceReconciler) SetupWithManager(mgr ctrl.Manager) error {
	c, err := controller.New("AtlasService", mgr, controller.Options{Reconciler: r})
	if err != nil {
		return err
	}

	// Watch for changes to primary resource AtlasService & handle delete separately
	err = c.Watch(&source.Kind{Type: &mdbv1.AtlasService{}}, &watch.EventHandlerWithDelete{Controller: r}, watch.CommonPredicates())
	if err != nil {
		return err
	}

	// Watch for Connection Secrets
	err = c.Watch(&source.Kind{Type: &corev1.Secret{}}, watch.NewSecretHandler(r.WatchedResources))
	if err != nil {
		return err
	}
	return nil
}
