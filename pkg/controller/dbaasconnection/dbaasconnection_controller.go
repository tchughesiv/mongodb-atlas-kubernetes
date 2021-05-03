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

package dbaasconnection

import (
	"context"
	"fmt"

	"go.uber.org/zap"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/rest"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/source"

	mdbv1 "github.com/mongodb/mongodb-atlas-kubernetes/pkg/api/v1"
	"github.com/mongodb/mongodb-atlas-kubernetes/pkg/api/v1/status"
	"github.com/mongodb/mongodb-atlas-kubernetes/pkg/controller/watch"
)

// DbaaSConnectionReconciler reconciles a DbaaSConnection object
type DbaaSConnectionReconciler struct {
	Client     client.Client
	RestClient rest.RESTClient
	watch.ResourceWatcher
	Log         *zap.SugaredLogger
	Scheme      *runtime.Scheme
	AtlasDomain string
	OperatorPod client.ObjectKey
}

// Dev note: duplicate the permissions in both sections below to generate both Role and ClusterRoles

// +kubebuilder:rbac:groups=atlas.mongodb.com,resources=DbaaSConnections,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=atlas.mongodb.com,resources=DbaaSConnections/status,verbs=get;update;patch
// +kubebuilder:rbac:groups="",resources=secrets,verbs=get;list;watch

// +kubebuilder:rbac:groups=atlas.mongodb.com,namespace=default,resources=DbaaSConnections,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=atlas.mongodb.com,namespace=default,resources=DbaaSConnections/status,verbs=get;update;patch
// +kubebuilder:rbac:groups="",namespace=default,resources=secrets,verbs=get;list;watch

func (r *DbaaSConnectionReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	//log := r.Log.With("DbaaSConnection", req.NamespacedName)

	connection := &mdbv1.DbaaSConnection{}
	if err := r.Client.Get(context.Background(), req.NamespacedName, connection); err != nil {
		return ctrl.Result{}, err
	}

	clusters := &mdbv1.AtlasClusterList{}
	if err := r.Client.List(context.Background(), clusters, &client.ListOptions{Namespace: req.Namespace}); err != nil {
		return ctrl.Result{}, err
	}
	found := false
	groupName := ""
	for _, db := range clusters.Items {
		if db.Spec.Name == connection.Spec.Cluster {
			r.Log.Info("Processing database connection", "vendor", connection.Spec.Vendor, "database", connection.Spec.Cluster)
			connection.Status.ConnectionString = db.Status.ConnectionStrings.Standard
			groupName = db.Spec.Project.Name
			found = true
			break
		}
	}

	if !found {
		r.Log.Info("Failed to find database ", "vendor ", connection.Spec.Vendor, " database ", connection.Spec.Cluster)
		return ctrl.Result{}, fmt.Errorf("Failed to find database %s with vendor %s", connection.Spec.Cluster, connection.Spec.Vendor)
	}

	dbUsers := &mdbv1.AtlasDatabaseUserList{}
	if err := r.Client.List(context.Background(), dbUsers, &client.ListOptions{Namespace: req.Namespace}); err != nil {
		return ctrl.Result{}, err
	}
	dbUser := ""
	dbUserSecret := ""
	found = false
	for _, user := range dbUsers.Items {
		if user.Spec.Project.Name == groupName {
			r.Log.Info("Processing database user", "user", user.Spec.Username)
			dbUser = user.Spec.Username
			dbUserSecret = user.Spec.PasswordSecret.Name
			found = true
			break
		}
	}

	if !found {
		r.Log.Info("Failed to find database user", "vendor", connection.Spec.Vendor, "database", connection.Spec.Cluster)
		return ctrl.Result{}, fmt.Errorf("Failed to find database user %s with vendor %s", connection.Spec.Cluster, connection.Spec.Vendor)
	}
	passwordSecret := &corev1.Secret{}
	if err := r.Client.Get(context.Background(), types.NamespacedName{Namespace: connection.Namespace, Name: dbUserSecret}, passwordSecret); err != nil {
		r.Log.Info("Failed to retrieve password from secret", "secret name", dbUserSecret)
		return ctrl.Result{}, fmt.Errorf("Failed to retrieve password from secret %s", dbUserSecret)
	}
	dbpassword := string(passwordSecret.Data["password"])
	connection.Status.DBUserSecret = fmt.Sprintf("%s-%s-%s", connection.Spec.Vendor, connection.Spec.Cluster, "creds")
	secret := &corev1.Secret{}
	isCreate := false
	if err := r.Client.Get(context.Background(), types.NamespacedName{Namespace: connection.Namespace, Name: connection.Status.DBUserSecret}, secret); err != nil {
		if !apierrors.IsNotFound(err) {
			return ctrl.Result{}, err
		}
		secret = &corev1.Secret{
			ObjectMeta: ctrl.ObjectMeta{
				Namespace: connection.Namespace,
				Name:      connection.Status.DBUserSecret,
			},
			Data: map[string][]byte{
				"db.user":     []byte(dbUser),
				"db.password": []byte(dbpassword),
			},
		}
		if err := r.Client.Create(context.TODO(), secret); err != nil {
			return ctrl.Result{}, err
		}
		isCreate = true
	}
	if !isCreate {
		secret.Data = map[string][]byte{
			"db.user":     []byte(dbUser),
			"db.password": []byte(dbpassword),
		}

		if err := r.Client.Update(context.TODO(), secret); err != nil {
			return ctrl.Result{}, err
		}
	}

	SetCondition("Ready", "True", connection)

	if err := r.Client.Status().Update(context.TODO(), connection); err != nil {
		return ctrl.Result{}, err
	}

	return ctrl.Result{}, nil
}

// Delete is a no-op
func (r *DbaaSConnectionReconciler) Delete(e event.DeleteEvent) error {
	return nil
}

func (r *DbaaSConnectionReconciler) SetupWithManager(mgr ctrl.Manager) error {
	c, err := controller.New("DbaaSConnection", mgr, controller.Options{Reconciler: r})
	if err != nil {
		return err
	}

	// Watch for changes to primary resource DbaaSConnection & handle delete separately
	err = c.Watch(&source.Kind{Type: &mdbv1.DbaaSConnection{}}, &watch.EventHandlerWithDelete{Controller: r}, watch.CommonPredicates())
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

func SetCondition(conditionType status.ConditionType, conditionStatus corev1.ConditionStatus, connection *mdbv1.DbaaSConnection) {
	now := metav1.Now()
	conditions := connection.Status.Conditions
	for i, cond := range conditions {
		if cond.Type == conditionType {
			if conditions[i].Status != conditionStatus {
				cond.LastTransitionTime = now
			}
			cond.Status = conditionStatus
			connection.Status.Conditions[i] = cond
			return
		}
	}

	// If the condition does not exist,
	// initialize the lastTransitionTime
	c := status.Condition{
		LastTransitionTime: now,
		Type:               conditionType,
		Status:             conditionStatus,
	}
	connection.Status.Conditions = append(connection.Status.Conditions, c)
}
