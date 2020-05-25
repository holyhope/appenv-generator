package tests

type FromField struct {
	// +appenv:key=THE_SECRET
	// +appenv:from:kind=secret
	// +appenv:from:field=the-secret-entry
	SecretName string `json:"secretName"`

	// +appenv:from:kind=secret
	FullSecretName string `json:"fullSecretName"`

	// +appenv:key=THE_CONFIG
	// +appenv:from:kind=configMap
	// +appenv:from:field=the-config-entry
	ConfigName string `json:"configName"`

	// +appenv:from:kind=configMap
	FullConfigName string `json:"fullConfigName"`
}

type FromField2 struct {
	// +appenv:key=THE_SECRET
	// +appenv:from:kind=secret
	// +appenv:from:field=the-secret-entry
	SecretName *string `json:"secretName"`

	// +appenv:from:kind=configMap
	FullConfigName *string `json:"fullConfigName"`
}
