package status

// DbaaSConnectionStatus defines the observed state of DbaaSConnection.
type DbaaSConnectionStatus struct {
	Common `json:",inline"`

	// ConnectionString is a connection string that your applications use to connect to this cluster.
	ConnectionString string `json:"connectionString,omitempty"`

	// DBUserSecret is database user secret name
	DBUserSecret string `json:"dbUserSecret,omitempty"`
}

// +k8s:deepcopy-gen=false

// DbaaSConnectionStatusOption is the option that is applied to Atlas Cluster Status.
type DbaaSConnectionStatusOption func(s *DbaaSConnectionStatus)

func DbaaSConnectionConnectionStringOption(connectionString string) DbaaSConnectionStatusOption {
	return func(s *DbaaSConnectionStatus) {
		s.ConnectionString = connectionString
	}
}

func DbaaSConnectionDBUserSecretOption(secretName string) DbaaSConnectionStatusOption {
	return func(s *DbaaSConnectionStatus) {
		s.DBUserSecret = secretName
	}
}
