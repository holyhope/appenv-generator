package tests

type Volume struct {
	// +appenv:mount:name=the-secret
	// +appenv:mount:path=/the/secret
	// +appenv:mount:kind=secret
	Secret string `json:"secret"`

	// +appenv:mount:name=the-certificate
	// +appenv:mount:path=/etc/certificates
	// +appenv:mount:kind=secret
	// +appenv:mount:items={"tls.key":"private_key.pem"}
	Certificate string `json:"certificate"`

	// +appenv:key=FULL_SECRET_MOUNT_LOCATION
	// +appenv:mount:name=full-secret
	// +appenv:mount:path=/the/full-secret-mount
	// +appenv:mount:kind=secret
	// +appenv:mount:items={"tls.key":"private_key.pem"}
	FullSecret string `json:"fullSecret"`
}
