/*
Copyright 2021 MongoDB.

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

package dbaasregistration

import (
	"context"
	"encoding/json"

	"github.com/go-logr/logr"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/yaml"
	"k8s.io/client-go/kubernetes"
)

const (
	name      = "mongodb-atlas-registration"
	namespace = "dbaas-operator" // namespace where the configmap should be stored
	data      = `
apiVersion: v1
kind: ConfigMap
metadata:
  name: mongodb-atlas-registration
  labels:
    related-to: dbaas-operator
    type: dbaas-provider-registration
data:
  provider: Red Hat DBaaS / MongoDB Atlas
  display_name: MongoDB Atlas Cloud Database Service
  logo: |
    data: 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
    mediatype: image/png
  inventory_kind: MongoDBAtlasInventory
  connection_kind: MongoDBAtlasConnection
  credentials_fields: |
    - key: orgId
      display_name: Organization ID
      type: string
      required: true
    - key: publicApiKey
      display_name: Public API Key
      type: string
      required: true
    - key: privateApiKey
      display_name: Private API Key
      type: maskedstring
      required: true`
)

// +kubebuilder:rbac:groups="",resources=namespaces,verbs=get;list
// +kubebuilder:rbac:groups="",resources=configmaps,resourceNames=mongodb-atlas-registration,verbs=get;list;create;update;patch;delete;watch
// +kubebuilder:rbac:groups="",namespace=default,resources=namespaces,verbs=get;list
// +kubebuilder:rbac:groups="",namespace=default,resources=configmaps,resourceNames=mongodb-atlas-registration,verbs=get;list;create;update;patch;delete;watch

// CreateAtlasRegistrationConfigMap creates a configmap for registration with Red Hat DBaaS Operator
func CreateAtlasRegistrationConfigMap(clientSet *kubernetes.Clientset, log logr.Logger) error {
	_, err := clientSet.CoreV1().Namespaces().Get(context.Background(), namespace, metav1.GetOptions{})
	if err != nil && errors.IsNotFound(err) {
		// if dbaas-operator namespace does not exist, no registration is done
		log.Info("DBaaS registration configmap creation is skipped as dbaas-operator namespace does not exist")
		return nil
	}
	_, err = clientSet.CoreV1().ConfigMaps(namespace).Get(context.TODO(), name, metav1.GetOptions{})
	if err != nil && errors.IsNotFound(err) {
		// the registration configmap does not exist, create one in dbaas-operator namespace
		jsonData, err := yaml.ToJSON([]byte(data))
		if err != nil {
			return err
		}
		cm := &corev1.ConfigMap{}
		err = json.Unmarshal(jsonData, cm)
		if err != nil {
			return err
		}
		_, err = clientSet.CoreV1().ConfigMaps(namespace).Create(context.Background(), cm, metav1.CreateOptions{})
		if err != nil {
			// failed to create the configmap
			return err
		}
		log.Info("DBaaS registration configmap is created in dbaas-operator namespace")
		return nil
	}
	if err == nil {
		log.Info("DBaaS registration configmap already exists in dbaas-operator namespace")
	}
	return err
}

// CleanupAtlasRegistrationConfigMap cleans up a configmap for registration with Red Hat DBaaS Operator
func CleanupAtlasRegistrationConfigMap(clientSet *kubernetes.Clientset, log logr.Logger) error {
	_, err := clientSet.CoreV1().ConfigMaps(namespace).Get(context.TODO(), name, metav1.GetOptions{})
	if err != nil && errors.IsNotFound(err) {
		// The registration configmap no longer exists
		return nil
	}

	// Delete the registration configmap
	err = clientSet.CoreV1().ConfigMaps(namespace).Delete(context.Background(), name, metav1.DeleteOptions{})
	if err != nil {
		// failed to delete the configmap
		log.Error(err, "unable to delete Provider Registration ConfigMap", "configMap.name", name, "configMap.Namespace", namespace)
		return err
	}

	return nil
}