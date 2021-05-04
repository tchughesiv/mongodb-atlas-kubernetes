package status

import (
	"github.com/mongodb/mongodb-atlas-kubernetes/pkg/util/compat"
)

// +k8s:deepcopy-gen=false

// AtlasServiceStatusOption is the option that is applied to Atlas Cluster Status.
type AtlasServiceStatusOption func(s *AtlasServiceStatus)

func AtlasProjectServiceListOption(serviceList []AtlasProjectService) AtlasServiceStatusOption {
	return func(s *AtlasServiceStatus) {
		sl := []AtlasProjectService{}
		err := compat.JSONCopy(&sl, serviceList)
		if err != nil {
			return
		}
		s.AtlasProjectServiceList = sl
	}
}

// AtlasProjectService defines the observed state of AtlasProjectService
type AtlasProjectService struct {
	// The ID of the Atlas Project
	// +optional
	ID string `json:"id,omitempty"`

	// The Name of the Atlas Project
	// +optional
	Name string `json:"name,omitempty"`

	// DBUserList is the list of DB users associated with the Atlas project
	// +optional
	DBUserList []string `json:"dbUsers,omitempty"`

	// ClusterList is the list of DB clusters associated with the Atlas project
	// +optional
	ClusterList []AtlasClusterService `json:"clusters,omitempty"`
}

// AtlasClusterService defines the observed state of AtlasClusterService
type AtlasClusterService struct {
	// The ID of the Atlas Project
	// +optional
	ID string `json:"id,omitempty"`

	// The Name of the Atlas Project
	// +optional
	Name string `json:"name,omitempty"`

	// InstanceSizeName is the db flavor indicating the instance size
	// +optional
	InstanceSizeName string `json:"instanceSizeName,omitempty"`

	// ProviderName is the cloud provider name
	// +optional
	ProviderName string `json:"providerName,omitempty"`

	// RegionName is the region name
	// +optional
	RegionName string `json:"regionName,omitempty"`

	// ConnectionString is the mongodb+srv://<host> connection URL for the cluster
	// +optional
	ConnectionString string `json:"conectionString,omitempty"`
}

// AtlasServiceStatus defines the observed state of AtlasService
type AtlasServiceStatus struct {
	Common `json:",inline"`

	// AtlasProjectServiceList is the list of projects associated with the user
	AtlasProjectServiceList []AtlasProjectService `json:"projects,omitempty"`
}
